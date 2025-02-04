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
	"fmt"
	"os"

	"github.com/spf13/cobra"

	"github.com/fatedier/frp/pkg/config"
	v1 "github.com/fatedier/frp/pkg/config/v1"
	"github.com/fatedier/frp/pkg/nathole"
)

var (
	natHoleSTUNServer string
	natHoleLocalAddr  string
)

func init() {
	rootCmd.AddCommand(natholeCmd)
	natholeCmd.AddCommand(natholeDiscoveryCmd)

	natholeCmd.PersistentFlags().StringVarP(&natHoleSTUNServer, "nat_hole_stun_server", "", "", "STUN 服务器地址")
	natholeCmd.PersistentFlags().StringVarP(&natHoleLocalAddr, "nat_hole_local_addr", "l", "", "连接 STUN 服务器的本地地址")
}

var natholeCmd = &cobra.Command{
	Use:   "nathole",
	Short: "NAT 有关操作",
}

var natholeDiscoveryCmd = &cobra.Command{
	Use:   "discover",
	Short: "从 STUN 服务器获取 NAT 信息",
	RunE: func(cmd *cobra.Command, args []string) error {
		// ignore error here, because we can use command line pameters
		cfg, _, _, _, err := config.LoadClientConfig(cfgFile, strictConfigMode)
		if err != nil {
			cfg = &v1.ClientCommonConfig{}
			cfg.Complete()
		}
		if natHoleSTUNServer != "" {
			cfg.NatHoleSTUNServer = natHoleSTUNServer
		}

		if err := validateForNatHoleDiscovery(cfg); err != nil {
			fmt.Println(err)
			os.Exit(1)
		}

		addrs, localAddr, err := nathole.Discover([]string{cfg.NatHoleSTUNServer}, natHoleLocalAddr)
		if err != nil {
			fmt.Println("发现失败:", err)
			os.Exit(1)
		}
		if len(addrs) < 2 {
			fmt.Printf("发现失败: 无法获取足够的地址, 需要 2 个, 实际获取: %v\n", addrs)
			os.Exit(1)
		}

		localIPs, _ := nathole.ListLocalIPsForNatHole(10)

		natFeature, err := nathole.ClassifyNATFeature(addrs, localIPs)
		if err != nil {
			fmt.Println("分类 NAT 特征错误:", err)
			os.Exit(1)
		}
		fmt.Println("STUN 服务器:", cfg.NatHoleSTUNServer)
		fmt.Println("您的 NAT 类型是:", natFeature.NatType)
		fmt.Println("行为是:", natFeature.Behavior)
		fmt.Println("外部地址是:", addrs)
		fmt.Println("本地地址是:", localAddr.String())
		fmt.Println("公共网络:", natFeature.PublicNetwork)
		return nil
	},
}

func validateForNatHoleDiscovery(cfg *v1.ClientCommonConfig) error {
	if cfg.NatHoleSTUNServer == "" {
		return fmt.Errorf("nat_hole_stun_server 不能为空")
	}
	return nil
}
