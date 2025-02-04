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

package config

import (
	"fmt"
	"strconv"
	"strings"

	"github.com/spf13/cobra"
	"github.com/spf13/pflag"

	"github.com/fatedier/frp/pkg/config/types"
	v1 "github.com/fatedier/frp/pkg/config/v1"
	"github.com/fatedier/frp/pkg/config/v1/validation"
)

// WordSepNormalizeFunc changes all flags that contain "_" separators
func WordSepNormalizeFunc(f *pflag.FlagSet, name string) pflag.NormalizedName {
	if strings.Contains(name, "_") {
		return pflag.NormalizedName(strings.ReplaceAll(name, "_", "-"))
	}
	return pflag.NormalizedName(name)
}

type RegisterFlagOption func(*registerFlagOptions)

type registerFlagOptions struct {
	sshMode bool
}

func WithSSHMode() RegisterFlagOption {
	return func(o *registerFlagOptions) {
		o.sshMode = true
	}
}

type BandwidthQuantityFlag struct {
	V *types.BandwidthQuantity
}

func (f *BandwidthQuantityFlag) Set(s string) error {
	return f.V.UnmarshalString(s)
}

func (f *BandwidthQuantityFlag) String() string {
	return f.V.String()
}

func (f *BandwidthQuantityFlag) Type() string {
	return "string"
}

func RegisterProxyFlags(cmd *cobra.Command, c v1.ProxyConfigurer, opts ...RegisterFlagOption) {
	registerProxyBaseConfigFlags(cmd, c.GetBaseConfig(), opts...)

	switch cc := c.(type) {
	case *v1.TCPProxyConfig:
		cmd.Flags().IntVarP(&cc.RemotePort, "remote_port", "r", 0, "远程端口")
	case *v1.UDPProxyConfig:
		cmd.Flags().IntVarP(&cc.RemotePort, "remote_port", "r", 0, "远程端口")
	case *v1.HTTPProxyConfig:
		registerProxyDomainConfigFlags(cmd, &cc.DomainConfig)
		cmd.Flags().StringSliceVarP(&cc.Locations, "locations", "", []string{}, "Locations")
		cmd.Flags().StringVarP(&cc.HTTPUser, "http_user", "", "", "HTTP Auth User")
		cmd.Flags().StringVarP(&cc.HTTPPassword, "http_pwd", "", "", "HTTP Auth Password")
		cmd.Flags().StringVarP(&cc.HostHeaderRewrite, "host_header_rewrite", "", "", "Host-Header-Rewrite")
	case *v1.HTTPSProxyConfig:
		registerProxyDomainConfigFlags(cmd, &cc.DomainConfig)
	case *v1.TCPMuxProxyConfig:
		registerProxyDomainConfigFlags(cmd, &cc.DomainConfig)
		cmd.Flags().StringVarP(&cc.Multiplexer, "mux", "", "", "Multiplexer")
		cmd.Flags().StringVarP(&cc.HTTPUser, "http_user", "", "", "HTTP Auth User")
		cmd.Flags().StringVarP(&cc.HTTPPassword, "http_pwd", "", "", "HTTP Auth Password")
	case *v1.STCPProxyConfig:
		cmd.Flags().StringVarP(&cc.Secretkey, "sk", "", "", "访问密钥")
		cmd.Flags().StringSliceVarP(&cc.AllowUsers, "allow_users", "", []string{}, "允许访问者用户")
	case *v1.SUDPProxyConfig:
		cmd.Flags().StringVarP(&cc.Secretkey, "sk", "", "", "访问密钥")
		cmd.Flags().StringSliceVarP(&cc.AllowUsers, "allow_users", "", []string{}, "允许访问者用户")
	case *v1.XTCPProxyConfig:
		cmd.Flags().StringVarP(&cc.Secretkey, "sk", "", "", "访问密钥")
		cmd.Flags().StringSliceVarP(&cc.AllowUsers, "allow_users", "", []string{}, "允许访问者用户")
	}
}

func registerProxyBaseConfigFlags(cmd *cobra.Command, c *v1.ProxyBaseConfig, opts ...RegisterFlagOption) {
	if c == nil {
		return
	}
	options := &registerFlagOptions{}
	for _, opt := range opts {
		opt(options)
	}

	cmd.Flags().StringVarP(&c.Name, "proxy_name", "n", "", "隧道名称")
	cmd.Flags().StringToStringVarP(&c.Metadatas, "metadatas", "", nil, "元数据键值对 (比如 key1=value1,key2=value2)")
	cmd.Flags().StringToStringVarP(&c.Annotations, "annotations", "", nil, "注释键值对 (比如 key1=value1,key2=value2)")

	if !options.sshMode {
		cmd.Flags().StringVarP(&c.LocalIP, "local_ip", "i", "127.0.0.1", "本地IP")
		cmd.Flags().IntVarP(&c.LocalPort, "local_port", "l", 0, "本地端口")
		cmd.Flags().BoolVarP(&c.Transport.UseEncryption, "ue", "", false, "使用数据加密")
		cmd.Flags().BoolVarP(&c.Transport.UseCompression, "uc", "", false, "使用数据压缩")
		cmd.Flags().StringVarP(&c.Transport.BandwidthLimitMode, "bandwidth_limit_mode", "", types.BandwidthLimitModeClient, "带宽限制模式")
		cmd.Flags().VarP(&BandwidthQuantityFlag{V: &c.Transport.BandwidthLimit}, "bandwidth_limit", "", "带宽限制 (比如 100KB or 1MB)")
	}
}

func registerProxyDomainConfigFlags(cmd *cobra.Command, c *v1.DomainConfig) {
	if c == nil {
		return
	}
	cmd.Flags().StringSliceVarP(&c.CustomDomains, "custom_domain", "d", []string{}, "自定义域名")
	cmd.Flags().StringVarP(&c.SubDomain, "sd", "", "", "自定义子域名")
}

func RegisterVisitorFlags(cmd *cobra.Command, c v1.VisitorConfigurer, opts ...RegisterFlagOption) {
	registerVisitorBaseConfigFlags(cmd, c.GetBaseConfig(), opts...)

	// add visitor flags if exist
}

func registerVisitorBaseConfigFlags(cmd *cobra.Command, c *v1.VisitorBaseConfig, _ ...RegisterFlagOption) {
	if c == nil {
		return
	}
	cmd.Flags().StringVarP(&c.Name, "visitor_name", "n", "", "访问者名称")
	cmd.Flags().BoolVarP(&c.Transport.UseEncryption, "ue", "", false, "使用数据加密")
	cmd.Flags().BoolVarP(&c.Transport.UseCompression, "uc", "", false, "使用数据压缩")
	cmd.Flags().StringVarP(&c.SecretKey, "sk", "", "", "访问密钥")
	cmd.Flags().StringVarP(&c.ServerName, "server_name", "", "", "服务端名称")
	cmd.Flags().StringVarP(&c.ServerUser, "server-user", "", "", "服务端用户")
	cmd.Flags().StringVarP(&c.BindAddr, "bind_addr", "", "", "绑定地址")
	cmd.Flags().IntVarP(&c.BindPort, "bind_port", "", 0, "绑定端口")
}

func RegisterClientCommonConfigFlags(cmd *cobra.Command, c *v1.ClientCommonConfig, opts ...RegisterFlagOption) {
	options := &registerFlagOptions{}
	for _, opt := range opts {
		opt(options)
	}

	if !options.sshMode {
		cmd.PersistentFlags().StringVarP(&c.ServerAddr, "server_addr", "s", "127.0.0.1", "ME Frp 服务端地址")
		cmd.PersistentFlags().IntVarP(&c.ServerPort, "server_port", "P", 7000, "ME Frp 服务端端口")
		cmd.PersistentFlags().StringVarP(&c.Transport.Protocol, "protocol", "p", "tcp",
			fmt.Sprintf("可选值为 %v", validation.SupportedTransportProtocols))
		cmd.PersistentFlags().StringVarP(&c.Log.Level, "log_level", "", "info", "日志级别")
		cmd.PersistentFlags().StringVarP(&c.Log.To, "log_file", "", "console", "日志保存位置: console 或 文件路径")
		cmd.PersistentFlags().Int64VarP(&c.Log.MaxDays, "log_max_days", "", 3, "日志文件保留天数")
		cmd.PersistentFlags().BoolVarP(&c.Log.DisablePrintColor, "disable_log_color", "", false, "禁用日志颜色")
		cmd.PersistentFlags().StringVarP(&c.Transport.TLS.ServerName, "tls_server_name", "", "", "指定 TLS 证书的自定义服务器名称")
		cmd.PersistentFlags().StringVarP(&c.DNSServer, "dns_server", "", "", "指定 DNS 服务器, 而不使用系统默认的 DNS 服务器")
		c.Transport.TLS.Enable = cmd.PersistentFlags().BoolP("tls_enable", "", true, "启用 TLS 连接")
	}
	cmd.PersistentFlags().StringVarP(&c.User, "user", "u", "", "用户")
	cmd.PersistentFlags().StringVarP(&c.Auth.Token, "token", "t", "", "Token")
}

type PortsRangeSliceFlag struct {
	V *[]types.PortsRange
}

func (f *PortsRangeSliceFlag) String() string {
	if f.V == nil {
		return ""
	}
	return types.PortsRangeSlice(*f.V).String()
}

func (f *PortsRangeSliceFlag) Set(s string) error {
	slice, err := types.NewPortsRangeSliceFromString(s)
	if err != nil {
		return err
	}
	*f.V = slice
	return nil
}

func (f *PortsRangeSliceFlag) Type() string {
	return "string"
}

type BoolFuncFlag struct {
	TrueFunc  func()
	FalseFunc func()

	v bool
}

func (f *BoolFuncFlag) String() string {
	return strconv.FormatBool(f.v)
}

func (f *BoolFuncFlag) Set(s string) error {
	f.v = strconv.FormatBool(f.v) == "true"

	if !f.v {
		if f.FalseFunc != nil {
			f.FalseFunc()
		}
		return nil
	}

	if f.TrueFunc != nil {
		f.TrueFunc()
	}
	return nil
}

func (f *BoolFuncFlag) Type() string {
	return "bool"
}

func RegisterServerConfigFlags(cmd *cobra.Command, c *v1.ServerConfig, opts ...RegisterFlagOption) {
	cmd.PersistentFlags().StringVarP(&c.BindAddr, "bind_addr", "", "0.0.0.0", "绑定地址")
	cmd.PersistentFlags().IntVarP(&c.BindPort, "bind_port", "p", 7000, "绑定端口")
	cmd.PersistentFlags().IntVarP(&c.KCPBindPort, "kcp_bind_port", "", 0, "kcp 绑定 udp 端口")
	cmd.PersistentFlags().IntVarP(&c.QUICBindPort, "quic_bind_port", "", 0, "QUIC 绑定 UDP 端口")
	cmd.PersistentFlags().StringVarP(&c.ProxyBindAddr, "proxy_bind_addr", "", "0.0.0.0", "隧道绑定地址")
	cmd.PersistentFlags().IntVarP(&c.VhostHTTPPort, "vhost_http_port", "", 0, "vhost http 端口")
	cmd.PersistentFlags().IntVarP(&c.VhostHTTPSPort, "vhost_https_port", "", 0, "vhost https 端口")
	cmd.PersistentFlags().Int64VarP(&c.VhostHTTPTimeout, "vhost_http_timeout", "", 60, "vhost http 响应头超时")
	cmd.PersistentFlags().StringVarP(&c.WebServer.Addr, "dashboard_addr", "", "0.0.0.0", "WebServer 地址")
	cmd.PersistentFlags().IntVarP(&c.WebServer.Port, "dashboard_port", "", 0, "WebServer 端口")
	cmd.PersistentFlags().StringVarP(&c.WebServer.User, "dashboard_user", "", "admin", "WebServer 用户")
	cmd.PersistentFlags().StringVarP(&c.WebServer.Password, "dashboard_pwd", "", "admin", "WebServer 密码")
	cmd.PersistentFlags().BoolVarP(&c.EnablePrometheus, "enable_prometheus", "", false, "启用 Prometheus 监控面板")
	cmd.PersistentFlags().StringVarP(&c.Log.To, "log_file", "", "console", "日志文件")
	cmd.PersistentFlags().StringVarP(&c.Log.Level, "log_level", "", "info", "日志级别")
	cmd.PersistentFlags().Int64VarP(&c.Log.MaxDays, "log_max_days", "", 3, "日志文件保留天数")
	cmd.PersistentFlags().BoolVarP(&c.Log.DisablePrintColor, "disable_log_color", "", false, "禁用日志颜色")
	cmd.PersistentFlags().StringVarP(&c.Auth.Token, "token", "t", "", "Token")
	cmd.PersistentFlags().StringVarP(&c.SubDomainHost, "subdomain_host", "", "", "子域名主机")
	cmd.PersistentFlags().VarP(&PortsRangeSliceFlag{V: &c.AllowPorts}, "allow_ports", "", "允许端口")
	cmd.PersistentFlags().Int64VarP(&c.MaxPortsPerClient, "max_ports_per_client", "", 0, "每个客户端允许的最大端口数")
	cmd.PersistentFlags().BoolVarP(&c.Transport.TLS.Force, "tls_only", "", false, "服务端仅允许 TLS 连接")

	webServerTLS := v1.TLSConfig{}
	cmd.PersistentFlags().StringVarP(&webServerTLS.CertFile, "dashboard_tls_cert_file", "", "", "WebServer TLS 证书文件")
	cmd.PersistentFlags().StringVarP(&webServerTLS.KeyFile, "dashboard_tls_key_file", "", "", "WebServer TLS 密钥文件")
	cmd.PersistentFlags().VarP(&BoolFuncFlag{
		TrueFunc: func() { c.WebServer.TLS = &webServerTLS },
	}, "dashboard_tls_mode", "", "启用 WebServer TLS 连接")
}
