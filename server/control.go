// Copyright 2017 fatedier, fatedier@gmail.com
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package server

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net"
	"net/http"
	"runtime/debug"
	"sync"
	"sync/atomic"
	"time"

	"github.com/samber/lo"

	"github.com/fatedier/frp/pkg/auth"
	"github.com/fatedier/frp/pkg/config"
	v1 "github.com/fatedier/frp/pkg/config/v1"
	pkgerr "github.com/fatedier/frp/pkg/errors"
	"github.com/fatedier/frp/pkg/msg"
	plugin "github.com/fatedier/frp/pkg/plugin/server"
	"github.com/fatedier/frp/pkg/transport"
	netpkg "github.com/fatedier/frp/pkg/util/net"
	"github.com/fatedier/frp/pkg/util/util"
	"github.com/fatedier/frp/pkg/util/version"
	"github.com/fatedier/frp/pkg/util/wait"
	"github.com/fatedier/frp/pkg/util/xlog"
	"github.com/fatedier/frp/server/controller"
	"github.com/fatedier/frp/server/metrics"
	"github.com/fatedier/frp/server/proxy"
)

type ControlManager struct {
	// controls indexed by run id
	ctlsByRunID map[string]*Control

	mu sync.RWMutex
}

func NewControlManager() *ControlManager {
	return &ControlManager{
		ctlsByRunID: make(map[string]*Control),
	}
}

func (cm *ControlManager) Add(runID string, ctl *Control) (old *Control) {
	cm.mu.Lock()
	defer cm.mu.Unlock()

	var ok bool
	old, ok = cm.ctlsByRunID[runID]
	if ok {
		old.Replaced(ctl)
	}
	cm.ctlsByRunID[runID] = ctl
	return
}

// we should make sure if it's the same control to prevent delete a new one
func (cm *ControlManager) Del(runID string, ctl *Control) {
	cm.mu.Lock()
	defer cm.mu.Unlock()
	if c, ok := cm.ctlsByRunID[runID]; ok && c == ctl {
		delete(cm.ctlsByRunID, runID)
	}
}

func (cm *ControlManager) GetByID(runID string) (ctl *Control, ok bool) {
	cm.mu.RLock()
	defer cm.mu.RUnlock()
	ctl, ok = cm.ctlsByRunID[runID]
	return
}

func (cm *ControlManager) Close() error {
	cm.mu.Lock()
	defer cm.mu.Unlock()
	for _, ctl := range cm.ctlsByRunID {
		ctl.Close()
	}
	cm.ctlsByRunID = make(map[string]*Control)
	return nil
}

type Control struct {
	// all resource managers and controllers
	rc *controller.ResourceController

	// proxy manager
	pxyManager *proxy.Manager

	// plugin manager
	pluginManager *plugin.Manager

	// verifies authentication based on selected method
	authVerifier auth.Verifier

	// other components can use this to communicate with client
	msgTransporter transport.MessageTransporter

	// msgDispatcher is a wrapper for control connection.
	// It provides a channel for sending messages, and you can register handlers to process messages based on their respective types.
	msgDispatcher *msg.Dispatcher

	// login message
	loginMsg *msg.Login

	// control connection
	conn net.Conn

	// work connections
	workConnCh chan net.Conn

	// proxies in one client
	proxies map[string]proxy.Proxy

	// pool count
	poolCount int

	// ports used, for limitations
	portsUsedNum int

	// last time got the Ping message
	lastPing atomic.Value

	// A new run id will be generated when a new client login.
	// If run id got from login message has same run id, it means it's the same client, so we can
	// replace old controller instantly.
	runID string

	mu sync.RWMutex

	// Server configuration information
	serverCfg *v1.ServerConfig

	xl     *xlog.Logger
	ctx    context.Context
	doneCh chan struct{}
}

// TODO(fatedier): Referencing the implementation of frpc, encapsulate the input parameters as SessionContext.
func NewControl(
	ctx context.Context,
	rc *controller.ResourceController,
	pxyManager *proxy.Manager,
	pluginManager *plugin.Manager,
	authVerifier auth.Verifier,
	ctlConn net.Conn,
	ctlConnEncrypted bool,
	loginMsg *msg.Login,
	serverCfg *v1.ServerConfig,
) (*Control, error) {
	poolCount := loginMsg.PoolCount
	if poolCount > int(serverCfg.Transport.MaxPoolCount) {
		poolCount = int(serverCfg.Transport.MaxPoolCount)
	}
	ctl := &Control{
		rc:            rc,
		pxyManager:    pxyManager,
		pluginManager: pluginManager,
		authVerifier:  authVerifier,
		conn:          ctlConn,
		loginMsg:      loginMsg,
		workConnCh:    make(chan net.Conn, poolCount+10),
		proxies:       make(map[string]proxy.Proxy),
		poolCount:     poolCount,
		portsUsedNum:  0,
		runID:         loginMsg.RunID,
		serverCfg:     serverCfg,
		xl:            xlog.FromContextSafe(ctx),
		ctx:           ctx,
		doneCh:        make(chan struct{}),
	}
	ctl.lastPing.Store(time.Now())

	if ctlConnEncrypted {
		cryptoRW, err := netpkg.NewCryptoReadWriter(ctl.conn, []byte(ctl.serverCfg.Auth.Token))
		if err != nil {
			return nil, err
		}
		ctl.msgDispatcher = msg.NewDispatcher(cryptoRW)
	} else {
		ctl.msgDispatcher = msg.NewDispatcher(ctl.conn)
	}
	ctl.registerMsgHandlers()
	ctl.msgTransporter = transport.NewMessageTransporter(ctl.msgDispatcher.SendChannel())
	return ctl, nil
}

// Start send a login success message to client and start working.
func (ctl *Control) Start() {
	loginRespMsg := &msg.LoginResp{
		Version: version.Full(),
		RunID:   ctl.runID,
		Error:   "",
	}
	_ = msg.WriteMsg(ctl.conn, loginRespMsg)

	go func() {
		for i := 0; i < ctl.poolCount; i++ {
			// ignore error here, that means that this control is closed
			_ = ctl.msgDispatcher.Send(&msg.ReqWorkConn{})
		}
	}()
	go ctl.worker()
}

func (ctl *Control) Close() error {
	ctl.conn.Close()
	return nil
}

func (ctl *Control) Replaced(newCtl *Control) {
	xl := ctl.xl
	xl.Infof("被客户端 [%s] 替换", newCtl.runID)
	ctl.runID = ""
	ctl.conn.Close()
}

func (ctl *Control) RegisterWorkConn(conn net.Conn) error {
	xl := ctl.xl
	defer func() {
		if err := recover(); err != nil {
			xl.Errorf("panic error: %v", err)
			xl.Errorf(string(debug.Stack()))
		}
	}()

	select {
	case ctl.workConnCh <- conn:
		xl.Debugf("新的工作连接已登记")
		return nil
	default:
		xl.Warnf("工作连接池已满, 丢弃新连接")
		return fmt.Errorf("工作连接池已满")
	}
}

// When frps get one user connection, we get one work connection from the pool and return it.
// If no workConn available in the pool, send message to frpc to get one or more
// and wait until it is available.
// return an error if wait timeout
func (ctl *Control) GetWorkConn() (workConn net.Conn, err error) {
	xl := ctl.xl
	defer func() {
		if err := recover(); err != nil {
			xl.Errorf("panic error: %v", err)
			xl.Errorf(string(debug.Stack()))
		}
	}()

	var ok bool
	// get a work connection from the pool
	select {
	case workConn, ok = <-ctl.workConnCh:
		if !ok {
			err = pkgerr.ErrCtlClosed
			return
		}
		xl.Debugf("从工作连接池获取工作连接")
	default:
		// no work connections available in the poll, send message to frpc to get more
		if err := ctl.msgDispatcher.Send(&msg.ReqWorkConn{}); err != nil {
			return nil, fmt.Errorf("控制已关闭")
		}

		select {
		case workConn, ok = <-ctl.workConnCh:
			if !ok {
				err = pkgerr.ErrCtlClosed
				xl.Warnf("没有可用的工作连接, %v", err)
				return
			}

		case <-time.After(time.Duration(ctl.serverCfg.UserConnTimeout) * time.Second):
			err = fmt.Errorf("尝试获取工作连接超时")
			xl.Warnf("%v", err)
			return
		}
	}

	// When we get a work connection from pool, replace it with a new one.
	_ = ctl.msgDispatcher.Send(&msg.ReqWorkConn{})
	return
}

func (ctl *Control) heartbeatWorker() {
	if ctl.serverCfg.Transport.HeartbeatTimeout <= 0 {
		return
	}

	xl := ctl.xl
	go wait.Until(func() {
		if time.Since(ctl.lastPing.Load().(time.Time)) > time.Duration(ctl.serverCfg.Transport.HeartbeatTimeout)*time.Second {
			xl.Warnf("心跳超时")
			ctl.conn.Close()
			return
		}
	}, time.Second, ctl.doneCh)
}

// block until Control closed
func (ctl *Control) WaitClosed() {
	<-ctl.doneCh
}

func (ctl *Control) worker() {
	xl := ctl.xl

	go ctl.heartbeatWorker()
	go ctl.msgDispatcher.Run()

	<-ctl.msgDispatcher.Done()
	ctl.conn.Close()

	ctl.mu.Lock()
	defer ctl.mu.Unlock()

	close(ctl.workConnCh)
	for workConn := range ctl.workConnCh {
		workConn.Close()
	}

	for _, pxy := range ctl.proxies {
		pxy.Close()
		reqBody := struct {
			ProxyName string `json:"proxyName"`
		}{
			ProxyName: pxy.GetName(),
		}
		reqBytes, _ := json.Marshal(reqBody)
		req, _ := http.NewRequest("POST", fmt.Sprintf("%s/api/frp/proxyClosed", ctl.serverCfg.MefrpApi.ApiUrl), bytes.NewBuffer(reqBytes))
		req.Header.Set("Content-Type", "application/json")
		req.Header.Set("Authorization", fmt.Sprintf("Bearer %s|%d", ctl.serverCfg.MefrpApi.Token, ctl.serverCfg.MefrpApi.NodeId))
		http.DefaultClient.Do(req)
		ctl.pxyManager.Del(pxy.GetName())
		metrics.Server.CloseProxy(pxy.GetName(), pxy.GetConfigurer().GetBaseConfig().Type)

		notifyContent := &plugin.CloseProxyContent{
			User: plugin.UserInfo{
				User:  ctl.loginMsg.User,
				Metas: ctl.loginMsg.Metas,
				RunID: ctl.loginMsg.RunID,
			},
			CloseProxy: msg.CloseProxy{
				ProxyName: pxy.GetName(),
			},
		}
		go func() {
			_ = ctl.pluginManager.CloseProxy(notifyContent)
		}()
	}

	metrics.Server.CloseClient()

	xl.Infof("客户端已成功断开连接")
	close(ctl.doneCh)
}

func (ctl *Control) registerMsgHandlers() {
	ctl.msgDispatcher.RegisterHandler(&msg.NewProxy{}, ctl.handleNewProxy)
	ctl.msgDispatcher.RegisterHandler(&msg.Ping{}, ctl.handlePing)
	ctl.msgDispatcher.RegisterHandler(&msg.NatHoleVisitor{}, msg.AsyncHandler(ctl.handleNatHoleVisitor))
	ctl.msgDispatcher.RegisterHandler(&msg.NatHoleClient{}, msg.AsyncHandler(ctl.handleNatHoleClient))
	ctl.msgDispatcher.RegisterHandler(&msg.NatHoleReport{}, msg.AsyncHandler(ctl.handleNatHoleReport))
	ctl.msgDispatcher.RegisterHandler(&msg.CloseProxy{}, ctl.handleCloseProxy)
}

func (ctl *Control) handleNewProxy(m msg.Message) {
	xl := ctl.xl
	inMsg := m.(*msg.NewProxy)

	content := &plugin.NewProxyContent{
		User: plugin.UserInfo{
			User:  ctl.loginMsg.User,
			Metas: ctl.loginMsg.Metas,
			RunID: ctl.loginMsg.RunID,
		},
		NewProxy: *inMsg,
	}
	var remoteAddr string
	if ctl.loginMsg.User == "" {
		xl.Warnf("用户 Token 为空")
		_ = ctl.msgDispatcher.Send(&msg.NewProxyResp{
			ProxyName: inMsg.ProxyName,
			Error:     "用户 Token 为空，请检查配置",
		})
		return
	}
	reqBody := struct {
		ProxyName string `json:"proxyName"`
		RunId     string `json:"runId"`
	}{
		ProxyName: inMsg.ProxyName,
		RunId:     ctl.runID,
	}
	reqBytes, _ := json.Marshal(reqBody)
	req, _ := http.NewRequest("POST", fmt.Sprintf("%s/api/frp/proxyStarted", ctl.serverCfg.MefrpApi.ApiUrl), bytes.NewBuffer(reqBytes))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", fmt.Sprintf("Bearer %s|%d", ctl.serverCfg.MefrpApi.Token, ctl.serverCfg.MefrpApi.NodeId))
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		xl.Infof("用户 [%s] 的隧道校验成功: %v", ctl.loginMsg.User, err.Error())
		_ = ctl.msgDispatcher.Send(&msg.NewProxyResp{
			ProxyName: inMsg.ProxyName,
			Error:     "隧道校验失败: " + err.Error(),
		})
		return
	} else if resp.StatusCode != 200 {
		var respData struct {
			Code    int    `json:"code"`
			Data    string `json:"data"`
			Message string `json:"message"`
		}
		body, _ := io.ReadAll(resp.Body)
		if err := json.Unmarshal(body, &respData); err != nil {
			xl.Errorf("解析响应失败: %v", err)
			_ = ctl.msgDispatcher.Send(&msg.NewProxyResp{
				ProxyName: inMsg.ProxyName,
				Error:     "隧道校验失败: " + string(body),
			})
			return
		}
		xl.Warnf("用户 [%s] 的隧道校验失败: %v", ctl.loginMsg.User, respData.Message)
		_ = ctl.msgDispatcher.Send(&msg.NewProxyResp{
			ProxyName: inMsg.ProxyName,
			Error:     "隧道校验失败: " + respData.Message,
		})
		return
	}

	var apiResp struct {
		Code    int    `json:"code"`
		Message string `json:"message"`
		Data    struct {
			OutBound int64 `json:"outBound"`
			InBound  int64 `json:"inBound"`
		} `json:"data"`
	}
	body, _ := io.ReadAll(resp.Body)
	if err := json.Unmarshal(body, &apiResp); err != nil {
		xl.Errorf("获取带宽限制失败: %v", err)
	} else {
		if apiResp.Data.OutBound > 0 || apiResp.Data.InBound > 0 {
			outBoundMB := apiResp.Data.OutBound / (1024 * 1024)
			inBoundMB := apiResp.Data.InBound / (1024 * 1024)
			if outBoundMB > 0 && inBoundMB > 0 {
				inMsg.BandwidthLimit = fmt.Sprintf("%dMB %dMB", inBoundMB, outBoundMB)
			} else if outBoundMB > 0 {
				inMsg.BandwidthLimit = fmt.Sprintf("%dMB", outBoundMB)
			} else if inBoundMB > 0 {
				inMsg.BandwidthLimit = fmt.Sprintf("%dMB", inBoundMB)
			}
			inMsg.BandwidthLimitMode = "client"
			_ = ctl.msgDispatcher.Send(&msg.GetProxyBandwidthLimitResp{
				ProxyName: inMsg.ProxyName,
				OutBound:  apiResp.Data.OutBound,
				InBound:   apiResp.Data.InBound,
			})
		}
	}

	retContent, err := ctl.pluginManager.NewProxy(content)
	if err == nil {
		inMsg = &retContent.NewProxy
		remoteAddr, err = ctl.RegisterProxy(inMsg)
	} else {
		xl.Warnf("登记隧道 [%s] 失败: %v", inMsg.ProxyName, err)
		_ = ctl.msgDispatcher.Send(&msg.NewProxyResp{
			ProxyName: inMsg.ProxyName,
			Error: util.GenerateResponseErrorString(fmt.Sprintf("登记隧道 [%s] 失败", inMsg.ProxyName),
				err, lo.FromPtr(ctl.serverCfg.DetailedErrorsToClient)),
		})
		return
	}

	if err != nil {
		xl.Warnf("登记 [%s] 隧道 [%s] 失败: %v", inMsg.ProxyType, inMsg.ProxyName, err)
		_ = ctl.msgDispatcher.Send(&msg.NewProxyResp{
			ProxyName: inMsg.ProxyName,
			Error: util.GenerateResponseErrorString(fmt.Sprintf("登记隧道 [%s] 失败", inMsg.ProxyName),
				err, lo.FromPtr(ctl.serverCfg.DetailedErrorsToClient)),
		})
		return
	}
	regResp := &msg.NewProxyResp{
		ProxyName: inMsg.ProxyName,
	}
	metrics.Server.NewProxy(inMsg.ProxyName, inMsg.ProxyType)
	regResp.RemoteAddr = remoteAddr
	if inMsg.ProxyType == "http" || inMsg.ProxyType == "https" {
		regResp.RemoteAddr = fmt.Sprintf("%s://%s", inMsg.ProxyType, remoteAddr)
	}
	xl.Infof("登记 [%s] 隧道 [%s] 成功", inMsg.ProxyType, inMsg.ProxyName)
	_ = ctl.msgDispatcher.Send(regResp)
}

func (ctl *Control) handlePing(m msg.Message) {
	xl := ctl.xl
	inMsg := m.(*msg.Ping)

	content := &plugin.PingContent{
		User: plugin.UserInfo{
			User:  ctl.loginMsg.User,
			Metas: ctl.loginMsg.Metas,
			RunID: ctl.loginMsg.RunID,
		},
		Ping: *inMsg,
	}
	retContent, err := ctl.pluginManager.Ping(content)
	if err == nil {
		inMsg = &retContent.Ping
		err = ctl.authVerifier.VerifyPing(inMsg)
	}
	if err != nil {
		xl.Warnf("收到无效的心跳: %v", err)
		_ = ctl.msgDispatcher.Send(&msg.Pong{
			Error: util.GenerateResponseErrorString("无效的心跳", err, lo.FromPtr(ctl.serverCfg.DetailedErrorsToClient)),
		})
		return
	}
	ctl.lastPing.Store(time.Now())
	xl.Debugf("收到心跳")
	_ = ctl.msgDispatcher.Send(&msg.Pong{})
}

func (ctl *Control) handleNatHoleVisitor(m msg.Message) {
	inMsg := m.(*msg.NatHoleVisitor)
	ctl.rc.NatHoleController.HandleVisitor(inMsg, ctl.msgTransporter, ctl.loginMsg.User)
}

func (ctl *Control) handleNatHoleClient(m msg.Message) {
	inMsg := m.(*msg.NatHoleClient)
	ctl.rc.NatHoleController.HandleClient(inMsg, ctl.msgTransporter)
}

func (ctl *Control) handleNatHoleReport(m msg.Message) {
	inMsg := m.(*msg.NatHoleReport)
	ctl.rc.NatHoleController.HandleReport(inMsg)
}

func (ctl *Control) handleCloseProxy(m msg.Message) {
	xl := ctl.xl
	inMsg := m.(*msg.CloseProxy)
	_ = ctl.CloseProxy(inMsg)
	xl.Infof("关闭隧道 [%s] 成功", inMsg.ProxyName)
}

func (ctl *Control) RegisterProxy(pxyMsg *msg.NewProxy) (remoteAddr string, err error) {
	var pxyConf v1.ProxyConfigurer
	// Load configures from NewProxy message and validate.
	pxyConf, err = config.NewProxyConfigurerFromMsg(pxyMsg, ctl.serverCfg)
	if err != nil {
		return
	}

	// User info
	userInfo := plugin.UserInfo{
		User:  ctl.loginMsg.User,
		Metas: ctl.loginMsg.Metas,
		RunID: ctl.runID,
	}

	// NewProxy will return an interface Proxy.
	// In fact, it creates different proxies based on the proxy type. We just call run() here.
	pxy, err := proxy.NewProxy(ctl.ctx, &proxy.Options{
		UserInfo:           userInfo,
		LoginMsg:           ctl.loginMsg,
		PoolCount:          ctl.poolCount,
		ResourceController: ctl.rc,
		GetWorkConnFn:      ctl.GetWorkConn,
		Configurer:         pxyConf,
		ServerCfg:          ctl.serverCfg,
	})
	if err != nil {
		return remoteAddr, err
	}

	// Check ports used number in each client
	if ctl.serverCfg.MaxPortsPerClient > 0 {
		ctl.mu.Lock()
		if ctl.portsUsedNum+pxy.GetUsedPortsNum() > int(ctl.serverCfg.MaxPortsPerClient) {
			ctl.mu.Unlock()
			err = fmt.Errorf("超过最大端口数")
			return
		}
		ctl.portsUsedNum += pxy.GetUsedPortsNum()
		ctl.mu.Unlock()

		defer func() {
			if err != nil {
				ctl.mu.Lock()
				ctl.portsUsedNum -= pxy.GetUsedPortsNum()
				ctl.mu.Unlock()
			}
		}()
	}

	if ctl.pxyManager.Exist(pxyMsg.ProxyName) {
		err = fmt.Errorf("隧道 [%s] 当前在线, 请尝试使用强制下线隧道功能", pxyMsg.ProxyName)
		return
	}

	remoteAddr, err = pxy.Run()
	if err != nil {
		return
	}
	defer func() {
		if err != nil {
			pxy.Close()
		}
	}()

	err = ctl.pxyManager.Add(pxyMsg.ProxyName, pxy)
	if err != nil {
		return
	}

	ctl.mu.Lock()
	ctl.proxies[pxy.GetName()] = pxy
	ctl.mu.Unlock()
	return
}

func (ctl *Control) CloseProxy(closeMsg *msg.CloseProxy) (err error) {
	ctl.mu.Lock()
	pxy, ok := ctl.proxies[closeMsg.ProxyName]
	if !ok {
		ctl.mu.Unlock()
		return
	}

	if ctl.serverCfg.MaxPortsPerClient > 0 {
		ctl.portsUsedNum -= pxy.GetUsedPortsNum()
	}
	pxy.Close()
	ctl.pxyManager.Del(pxy.GetName())
	delete(ctl.proxies, closeMsg.ProxyName)
	ctl.mu.Unlock()

	metrics.Server.CloseProxy(pxy.GetName(), pxy.GetConfigurer().GetBaseConfig().Type)

	notifyContent := &plugin.CloseProxyContent{
		User: plugin.UserInfo{
			User:  ctl.loginMsg.User,
			Metas: ctl.loginMsg.Metas,
			RunID: ctl.loginMsg.RunID,
		},
		CloseProxy: msg.CloseProxy{
			ProxyName: pxy.GetName(),
		},
	}
	go func() {
		_ = ctl.pluginManager.CloseProxy(notifyContent)
	}()
	return
}
