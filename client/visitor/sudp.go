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

package visitor

import (
	"fmt"
	"io"
	"net"
	"strconv"
	"sync"
	"time"

	"github.com/fatedier/golib/errors"
	libio "github.com/fatedier/golib/io"

	v1 "github.com/fatedier/frp/pkg/config/v1"
	"github.com/fatedier/frp/pkg/msg"
	"github.com/fatedier/frp/pkg/proto/udp"
	netpkg "github.com/fatedier/frp/pkg/util/net"
	"github.com/fatedier/frp/pkg/util/util"
	"github.com/fatedier/frp/pkg/util/xlog"
)

type SUDPVisitor struct {
	*BaseVisitor

	checkCloseCh chan struct{}
	// udpConn is the listener of udp packet
	udpConn *net.UDPConn
	readCh  chan *msg.UDPPacket
	sendCh  chan *msg.UDPPacket

	cfg *v1.SUDPVisitorConfig
}

// SUDP Run start listen a udp port
func (sv *SUDPVisitor) Run() (err error) {
	xl := xlog.FromContextSafe(sv.ctx)

	addr, err := net.ResolveUDPAddr("udp", net.JoinHostPort(sv.cfg.BindAddr, strconv.Itoa(sv.cfg.BindPort)))
	if err != nil {
		return fmt.Errorf("SUDP ResolveUDPAddr 错误: %v", err)
	}

	sv.udpConn, err = net.ListenUDP("udp", addr)
	if err != nil {
		return fmt.Errorf("SUDP 监听 UDP 端口 %s 错误: %v", addr.String(), err)
	}

	sv.sendCh = make(chan *msg.UDPPacket, 1024)
	sv.readCh = make(chan *msg.UDPPacket, 1024)

	xl.Infof("SUDP 开始工作, 监听于 %s", addr)

	go sv.dispatcher()
	go udp.ForwardUserConn(sv.udpConn, sv.readCh, sv.sendCh, int(sv.clientCfg.UDPPacketSize))

	return
}

func (sv *SUDPVisitor) dispatcher() {
	xl := xlog.FromContextSafe(sv.ctx)

	var (
		visitorConn net.Conn
		err         error

		firstPacket *msg.UDPPacket
	)

	for {
		select {
		case firstPacket = <-sv.sendCh:
			if firstPacket == nil {
				xl.Infof("SUDP 用户隧道已关闭")
				return
			}
		case <-sv.checkCloseCh:
			xl.Infof("SUDP 用户隧道已关闭")
			return
		}

		visitorConn, err = sv.getNewVisitorConn()
		if err != nil {
			xl.Warnf("连接到 ME Frp 节点错误: %v, 尝试重新连接", err)
			continue
		}

		// visitorConn always be closed when worker done.
		sv.worker(visitorConn, firstPacket)

		select {
		case <-sv.checkCloseCh:
			return
		default:
		}
	}
}

func (sv *SUDPVisitor) worker(workConn net.Conn, firstPacket *msg.UDPPacket) {
	xl := xlog.FromContextSafe(sv.ctx)
	xl.Debugf("启动 SUDP 隧道工作")

	wg := &sync.WaitGroup{}
	wg.Add(2)
	closeCh := make(chan struct{})

	// udp service -> frpc -> frps -> frpc visitor -> user
	workConnReaderFn := func(conn net.Conn) {
		defer func() {
			conn.Close()
			close(closeCh)
			wg.Done()
		}()

		for {
			var (
				rawMsg msg.Message
				errRet error
			)

			// frpc will send heartbeat in workConn to frpc visitor for keeping alive
			_ = conn.SetReadDeadline(time.Now().Add(60 * time.Second))
			if rawMsg, errRet = msg.ReadMsg(conn); errRet != nil {
				xl.Warnf("从工作连接读取用户 UDP 连接错误: %v", errRet)
				return
			}

			_ = conn.SetReadDeadline(time.Time{})
			switch m := rawMsg.(type) {
			case *msg.Ping:
				xl.Debugf("SUDP 用户隧道收到 Ping 消息")
				continue
			case *msg.UDPPacket:
				if errRet := errors.PanicToError(func() {
					sv.readCh <- m
					xl.Tracef("SUDP 用户隧道从工作连接收到 UDP 包: %s", m.Content)
				}); errRet != nil {
					xl.Infof("SUDP 用户隧道工作连接读取器已关闭")
					return
				}
			}
		}
	}

	// udp service <- frpc <- frps <- frpc visitor <- user
	workConnSenderFn := func(conn net.Conn) {
		defer func() {
			conn.Close()
			wg.Done()
		}()

		var errRet error
		if firstPacket != nil {
			if errRet = msg.WriteMsg(conn, firstPacket); errRet != nil {
				xl.Warnf("SUDP 用户隧道工作连接发送器已关闭: %v", errRet)
				return
			}
			xl.Tracef("发送 UDP 包到工作连接: %s", firstPacket.Content)
		}

		for {
			select {
			case udpMsg, ok := <-sv.sendCh:
				if !ok {
					xl.Infof("SUDP 用户隧道工作连接发送器已关闭")
					return
				}

				if errRet = msg.WriteMsg(conn, udpMsg); errRet != nil {
					xl.Warnf("SUDP 用户隧道工作连接发送器已关闭: %v", errRet)
					return
				}
				xl.Tracef("发送 UDP 包到工作连接: %s", udpMsg.Content)
			case <-closeCh:
				return
			}
		}
	}

	go workConnReaderFn(workConn)
	go workConnSenderFn(workConn)

	wg.Wait()
	xl.Infof("SUDP 隧道工作已关闭")
}

func (sv *SUDPVisitor) getNewVisitorConn() (net.Conn, error) {
	xl := xlog.FromContextSafe(sv.ctx)
	visitorConn, err := sv.helper.ConnectServer()
	if err != nil {
		return nil, fmt.Errorf("SUDP 连接到 ME Frp 节点错误: %v", err)
	}

	now := time.Now().Unix()
	newVisitorConnMsg := &msg.NewVisitorConn{
		RunID:          sv.helper.RunID(),
		ProxyName:      sv.cfg.ServerName,
		SignKey:        util.GetAuthKey(sv.cfg.SecretKey, now),
		Timestamp:      now,
		UseEncryption:  sv.cfg.Transport.UseEncryption,
		UseCompression: sv.cfg.Transport.UseCompression,
	}
	err = msg.WriteMsg(visitorConn, newVisitorConnMsg)
	if err != nil {
		return nil, fmt.Errorf("SUDP 发送 newVisitorConnMsg 到 ME Frp 节点错误: %v", err)
	}

	var newVisitorConnRespMsg msg.NewVisitorConnResp
	_ = visitorConn.SetReadDeadline(time.Now().Add(10 * time.Second))
	err = msg.ReadMsgInto(visitorConn, &newVisitorConnRespMsg)
	if err != nil {
		return nil, fmt.Errorf("SUDP 读取 newVisitorConnRespMsg 错误: %v", err)
	}
	_ = visitorConn.SetReadDeadline(time.Time{})

	if newVisitorConnRespMsg.Error != "" {
		return nil, fmt.Errorf("启动新的用户连接错误: %s", newVisitorConnRespMsg.Error)
	}

	var remote io.ReadWriteCloser
	remote = visitorConn
	if sv.cfg.Transport.UseEncryption {
		remote, err = libio.WithEncryption(remote, []byte(sv.cfg.SecretKey))
		if err != nil {
			xl.Errorf("创建加密流错误: %v", err)
			return nil, err
		}
	}
	if sv.cfg.Transport.UseCompression {
		remote = libio.WithCompression(remote)
	}
	return netpkg.WrapReadWriteCloserToConn(remote, visitorConn), nil
}

func (sv *SUDPVisitor) Close() {
	sv.mu.Lock()
	defer sv.mu.Unlock()

	select {
	case <-sv.checkCloseCh:
		return
	default:
		close(sv.checkCloseCh)
	}
	sv.BaseVisitor.Close()
	if sv.udpConn != nil {
		sv.udpConn.Close()
	}
	close(sv.readCh)
	close(sv.sendCh)
}
