// Copyright 2023 The frp Authors
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
	"context"
	"fmt"
	"os"
	"strings"
	"time"

	"github.com/rodaine/table"
	"github.com/spf13/cobra"

	"github.com/fatedier/frp/pkg/config"
	v1 "github.com/fatedier/frp/pkg/config/v1"
	clientsdk "github.com/fatedier/frp/pkg/sdk/client"
)

var adminAPITimeout = 30 * time.Second

func init() {
	commands := []struct {
		name        string
		description string
		handler     func(*v1.ClientCommonConfig) error
	}{
		{"reload", "重载 ME Frp 客户端配置", ReloadHandler},
		{"status", "查看 ME Frp 隧道状态", StatusHandler},
		{"stop", "停止 ME Frp 客户端", StopHandler},
	}

	for _, cmdConfig := range commands {
		cmd := NewAdminCommand(cmdConfig.name, cmdConfig.description, cmdConfig.handler)
		cmd.Flags().DurationVar(&adminAPITimeout, "api-timeout", adminAPITimeout, "Timeout for admin API calls")
		rootCmd.AddCommand(cmd)
	}
}

func NewAdminCommand(name, short string, handler func(*v1.ClientCommonConfig) error) *cobra.Command {
	return &cobra.Command{
		Use:   name,
		Short: short,
		Run: func(cmd *cobra.Command, args []string) {
			cfg, _, _, _, err := config.LoadClientConfig(cfgFile, strictConfigMode)
			if err != nil {
				fmt.Println(err)
				os.Exit(1)
			}
			if cfg.WebServer.Port <= 0 {
				fmt.Println("如果需要使用此功能, 请设置 WebServer 端口")
				os.Exit(1)
			}

			if err := handler(cfg); err != nil {
				fmt.Println(err)
				os.Exit(1)
			}
		},
	}
}

func ReloadHandler(clientCfg *v1.ClientCommonConfig) error {
	client := clientsdk.New(clientCfg.WebServer.Addr, clientCfg.WebServer.Port)
	client.SetAuth(clientCfg.WebServer.User, clientCfg.WebServer.Password)
	ctx, cancel := context.WithTimeout(context.Background(), adminAPITimeout)
	defer cancel()
	if err := client.Reload(ctx, strictConfigMode); err != nil {
		return err
	}
	fmt.Println("重载成功")
	return nil
}

func StatusHandler(clientCfg *v1.ClientCommonConfig) error {
	client := clientsdk.New(clientCfg.WebServer.Addr, clientCfg.WebServer.Port)
	client.SetAuth(clientCfg.WebServer.User, clientCfg.WebServer.Password)
	ctx, cancel := context.WithTimeout(context.Background(), adminAPITimeout)
	defer cancel()
	res, err := client.GetAllProxyStatus(ctx)
	if err != nil {
		return err
	}

	fmt.Printf("ME Frp 隧道状态...\n\n")
	for _, typ := range proxyTypes {
		arrs := res[string(typ)]
		if len(arrs) == 0 {
			continue
		}

		fmt.Println(strings.ToUpper(string(typ)))
		tbl := table.New("Name", "Status", "LocalAddr", "Plugin", "RemoteAddr", "Error")
		for _, ps := range arrs {
			tbl.AddRow(ps.Name, ps.Status, ps.LocalAddr, ps.Plugin, ps.RemoteAddr, ps.Err)
		}
		tbl.Print()
		fmt.Println("")
	}
	return nil
}

func StopHandler(clientCfg *v1.ClientCommonConfig) error {
	client := clientsdk.New(clientCfg.WebServer.Addr, clientCfg.WebServer.Port)
	client.SetAuth(clientCfg.WebServer.User, clientCfg.WebServer.Password)
	ctx, cancel := context.WithTimeout(context.Background(), adminAPITimeout)
	defer cancel()
	if err := client.Stop(ctx); err != nil {
		return err
	}
	fmt.Println("关闭成功")
	return nil
}
