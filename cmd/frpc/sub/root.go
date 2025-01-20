// Copyright 2018 fatedier, fatedier@gmail.com
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

package sub

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io/fs"
	"net/http"
	"os"
	"os/signal"
	"path/filepath"
	"strings"
	"sync"
	"syscall"
	"time"

	"github.com/samber/lo"
	"github.com/spf13/cobra"

	"github.com/fatedier/frp/client"
	"github.com/fatedier/frp/pkg/config"
	v1 "github.com/fatedier/frp/pkg/config/v1"
	"github.com/fatedier/frp/pkg/config/v1/validation"
	"github.com/fatedier/frp/pkg/util/log"
	"github.com/fatedier/frp/pkg/util/version"
)

var (
	cfgFile          string
	cfgDir           string
	showVersion      bool
	strictConfigMode bool
	userToken        string
	proxyId          string
)

type ProxyConfigResp struct {
	ProxyId              int64  `json:"proxyId"`
	Username             string `json:"username"`
	ProxyName            string `json:"proxyName"`
	ProxyType            string `json:"proxyType"`
	IsBanned             bool   `json:"isBanned"`
	LocalIp              string `json:"localIp"`
	LocalPort            int32  `json:"localPort"`
	RemotePort           int32  `json:"remotePort"`
	RunId                string `json:"runId"`
	IsOnline             bool   `json:"isOnline"`
	Domain               string `json:"domain"`
	LastStartTime        int64  `json:"lastStartTime"`
	LastCloseTime        int64  `json:"lastCloseTime"`
	ClientVersion        string `json:"clientVersion"`
	ProxyProtocolVersion string `json:"proxyProtocolVersion"`
	UseEncryption        bool   `json:"useEncryption"`
	UseCompression       bool   `json:"useCompression"`
	Location             string `json:"location"`
	AccessKey            string `json:"accessKey"`
	HostHeaderRewrite    string `json:"hostHeaderRewrite"`
	HeaderXFromWhere     string `json:"headerXFromWhere"`
	NodeAddr             string `json:"nodeAddr"`
	NodePort             int32  `json:"nodePort"`
	NodeToken            string `json:"nodeToken"`
}

type APIResponse struct {
	Code    int             `json:"code"`
	Message string          `json:"message"`
	Data    ProxyConfigResp `json:"data"`
}

func init() {
	rootCmd.PersistentFlags().StringVarP(&cfgFile, "config", "c", "", "ME Frp 客户端配置文件")
	rootCmd.PersistentFlags().StringVarP(&cfgDir, "config_dir", "", "", "配置目录, 为每个配置文件运行一个 ME Frp 隧道")
	rootCmd.PersistentFlags().BoolVarP(&showVersion, "version", "v", false, "ME Frp 客户端版本")
	rootCmd.PersistentFlags().BoolVarP(&strictConfigMode, "strict_config", "", true, "严格配置解析模式, 未知字段将导致错误")
	rootCmd.PersistentFlags().StringVarP(&userToken, "token", "t", "", "快捷启动的用户 Token")
	rootCmd.PersistentFlags().StringVarP(&proxyId, "proxy", "p", "", "快捷启动的隧道 Id")
}

var rootCmd = &cobra.Command{
	Use:   "frpc",
	Short: "ME Frp 客户端",
	RunE: func(cmd *cobra.Command, args []string) error {
		if showVersion {
			fmt.Println(version.Full())
			return nil
		}

		if cfgFile != "" {
			err := runClient(cfgFile)
			if err != nil {
				fmt.Println(err)
				os.Exit(1)
			}
			return nil
		} else if userToken != "" && proxyId != "" {
			return runEasyStartup()
		} else if cfgDir != "" {
			return runMultipleClients(cfgDir)
		}

		return fmt.Errorf("请提供配置文件 (-c) 或快捷启动参数 (-t -p)")
	},
}

func runMultipleClients(cfgDir string) error {
	var wg sync.WaitGroup
	err := filepath.WalkDir(cfgDir, func(path string, d fs.DirEntry, err error) error {
		if err != nil || d.IsDir() {
			return nil
		}
		wg.Add(1)
		time.Sleep(time.Millisecond)
		go func() {
			defer wg.Done()
			err := runClient(path)
			if err != nil {
				fmt.Printf("配置文件 [%s] 的 ME Frp 隧道出错\n", path)
			}
		}()
		return nil
	})
	wg.Wait()
	return err
}

func Execute() {
	rootCmd.SetGlobalNormalizationFunc(config.WordSepNormalizeFunc)
	if err := rootCmd.Execute(); err != nil {
		os.Exit(1)
	}
}

func handleTermSignal(svr *client.Service) {
	ch := make(chan os.Signal, 1)
	signal.Notify(ch, syscall.SIGINT, syscall.SIGTERM)
	<-ch
	svr.GracefulClose(500 * time.Millisecond)
}

func runClient(cfgFilePath string) error {
	cfg, proxyCfgs, visitorCfgs, isLegacyFormat, err := config.LoadClientConfig(cfgFilePath, strictConfigMode)
	if err != nil {
		return err
	}
	if isLegacyFormat {
		fmt.Printf("警告: INI 格式已弃用, 将在未来版本中移除, 请使用 Yaml/JSON/Toml 格式!\n")
	}

	warning, err := validation.ValidateAllClientConfig(cfg, proxyCfgs, visitorCfgs)
	if warning != nil {
		fmt.Printf("警告: %v\n", warning)
	}
	if err != nil {
		return err
	}
	return startService(cfg, proxyCfgs, visitorCfgs, cfgFilePath)
}

func startService(
	cfg *v1.ClientCommonConfig,
	proxyCfgs []v1.ProxyConfigurer,
	visitorCfgs []v1.VisitorConfigurer,
	cfgFile string,
) error {
	log.InitLogger(cfg.Log.To, cfg.Log.Level, int(cfg.Log.MaxDays), cfg.Log.DisablePrintColor)

	if cfgFile != "" {
		log.Infof("开始启动 ME Frp 隧道, 配置文件 [%s]", cfgFile)
		defer log.Infof("ME Frp 客户端配置文件 [%s] 已停止", cfgFile)
	} else if cfg.User != "" {
		log.Infof("开始启动 ME Frp 隧道, 当前正在使用快捷启动")
		defer log.Infof("ME Frp 客户端已停止")
	}
	svr, err := client.NewService(client.ServiceOptions{
		Common:         cfg,
		ProxyCfgs:      proxyCfgs,
		VisitorCfgs:    visitorCfgs,
		ConfigFilePath: cfgFile,
	})
	if err != nil {
		return err
	}

	shouldGracefulClose := cfg.Transport.Protocol == "kcp" || cfg.Transport.Protocol == "quic"
	// Capture the exit signal if we use kcp or quic.
	if shouldGracefulClose {
		go handleTermSignal(svr)
	}
	return svr.Run(context.Background())
}

func runEasyStartup() error {
	if userToken == "" || proxyId == "" {
		return fmt.Errorf("使用快捷启动时，用户Token 和 隧道Id 都是必需的")
	}

	// 获取所有隧道配置
	proxyIds := strings.Split(proxyId, ",")
	var proxies []ProxyConfigResp

	for _, pid := range proxyIds {
		proxyConfig, err := fetchProxyConfig(pid, userToken)
		if err != nil {
			return fmt.Errorf("获取隧道 [%s] 配置失败: %v", pid, err)
		}
		proxies = append(proxies, proxyConfig)
	}

	if len(proxies) == 0 {
		return fmt.Errorf("没有获取到任何隧道配置")
	}

	// 创建通用配置
	cfg := &v1.ClientCommonConfig{
		Auth: v1.AuthClientConfig{
			Method: "token",
			Token:  proxies[0].NodeToken,
		},
		ServerAddr:    proxies[0].NodeAddr,
		ServerPort:    int(proxies[0].NodePort),
		User:          userToken,
		LoginFailExit: lo.ToPtr(true),
		Log: v1.LogConfig{
			To:      "console",
			Level:   "info",
			MaxDays: 3,
		},
	}

	// 创建所有代理配置
	var proxyCfgs []v1.ProxyConfigurer
	for _, proxy := range proxies {
		proxyCfg := createProxyConfig(&proxy, userToken)
		if proxyCfg == nil {
			return fmt.Errorf("不支持的隧道类型: %s (支持的类型: tcp, http, https)", proxy.ProxyType)
		}
		proxyCfgs = append(proxyCfgs, proxyCfg)
	}

	return startService(cfg, proxyCfgs, nil, "")
}

func createProxyConfig(proxy *ProxyConfigResp, userToken string) v1.ProxyConfigurer {
	baseConfig := v1.ProxyBaseConfig{
		Name: userToken + "." + proxy.ProxyName,
		Type: proxy.ProxyType,
		ProxyBackend: v1.ProxyBackend{
			LocalIP:   proxy.LocalIp,
			LocalPort: int(proxy.LocalPort),
		},
		Transport: v1.ProxyTransport{
			UseEncryption:        proxy.UseEncryption,
			UseCompression:       proxy.UseCompression,
			ProxyProtocolVersion: proxy.ProxyProtocolVersion,
		},
		Metadatas: map[string]string{
			"location":  proxy.Location,
			"accessKey": proxy.AccessKey,
		},
	}

	switch proxy.ProxyType {
	case "tcp":
		return &v1.TCPProxyConfig{
			ProxyBaseConfig: baseConfig,
			RemotePort:      int(proxy.RemotePort),
		}
	case "http":
		return &v1.HTTPProxyConfig{
			ProxyBaseConfig: baseConfig,
			DomainConfig: v1.DomainConfig{
				CustomDomains: []string{proxy.Domain},
			},
			HostHeaderRewrite: proxy.HostHeaderRewrite,
			RequestHeaders: v1.HeaderOperations{
				Set: map[string]string{
					"x-from-where": proxy.HeaderXFromWhere,
				},
			},
		}
	case "https":
		return &v1.HTTPSProxyConfig{
			ProxyBaseConfig: baseConfig,
			DomainConfig: v1.DomainConfig{
				CustomDomains: []string{proxy.Domain},
			},
		}
	default:
		return nil
	}
}

func fetchProxyConfig(proxyId string, userToken string) (ProxyConfigResp, error) {
	url := "http://127.0.0.1:8080/api/auth/easyStartup"
	jsonBody := []byte(fmt.Sprintf(`{"proxyId": %s}`, proxyId))
	req, err := http.NewRequest("POST", url, bytes.NewBuffer(jsonBody))
	if err != nil {
		return ProxyConfigResp{}, fmt.Errorf("创建请求失败: %v", err)
	}

	req.Header.Set("Authorization", "Bearer "+userToken)
	req.Header.Set("Content-Type", "application/json")

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return ProxyConfigResp{}, fmt.Errorf("发送请求失败: %v", err)
	}
	defer resp.Body.Close()

	var proxyResp APIResponse
	if err := json.NewDecoder(resp.Body).Decode(&proxyResp); err != nil {
		return ProxyConfigResp{}, fmt.Errorf("解析响应失败: %v", err)
	}

	if proxyResp.Code != 200 {
		return ProxyConfigResp{}, fmt.Errorf("API 错误: %s", proxyResp.Message)
	}

	if proxyResp.Data.IsBanned {
		return ProxyConfigResp{}, fmt.Errorf("隧道已被封禁")
	}

	if proxyResp.Data.ProxyType == "" {
		return ProxyConfigResp{}, fmt.Errorf("隧道类型为空，请检查隧道是否存在")
	}

	if proxyResp.Data.NodeAddr == "" {
		return ProxyConfigResp{}, fmt.Errorf("API 返回的节点地址为空")
	}

	if proxyResp.Data.NodePort == 0 {
		return ProxyConfigResp{}, fmt.Errorf("API 返回的节点端口为空")
	}

	return proxyResp.Data, nil
}
