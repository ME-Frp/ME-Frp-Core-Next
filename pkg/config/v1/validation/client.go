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

package validation

import (
	"fmt"
	"os"
	"path/filepath"
	"slices"

	"github.com/samber/lo"

	v1 "github.com/fatedier/frp/pkg/config/v1"
)

func ValidateClientCommonConfig(c *v1.ClientCommonConfig) (Warning, error) {
	var (
		warnings Warning
		errs     error
	)
	if !slices.Contains(SupportedAuthMethods, c.Auth.Method) {
		errs = AppendError(errs, fmt.Errorf("验证方法无效, 可选值为 %v", SupportedAuthMethods))
	}
	if !lo.Every(SupportedAuthAdditionalScopes, c.Auth.AdditionalScopes) {
		errs = AppendError(errs, fmt.Errorf("验证附加范围无效, 可选值为 %v", SupportedAuthAdditionalScopes))
	}

	if err := validateLogConfig(&c.Log); err != nil {
		errs = AppendError(errs, err)
	}

	if err := validateWebServerConfig(&c.WebServer); err != nil {
		errs = AppendError(errs, err)
	}

	if c.Transport.HeartbeatTimeout > 0 && c.Transport.HeartbeatInterval > 0 {
		if c.Transport.HeartbeatTimeout < c.Transport.HeartbeatInterval {
			errs = AppendError(errs, fmt.Errorf("心跳包超时时间小于心跳包间隔时间"))
		}
	}

	if !lo.FromPtr(c.Transport.TLS.Enable) {
		checkTLSConfig := func(name string, value string) Warning {
			if value != "" {
				return fmt.Errorf("%s 在 transport.tls.enable 为 false 时无效", name)
			}
			return nil
		}

		warnings = AppendError(warnings, checkTLSConfig("transport.tls.certFile", c.Transport.TLS.CertFile))
		warnings = AppendError(warnings, checkTLSConfig("transport.tls.keyFile", c.Transport.TLS.KeyFile))
		warnings = AppendError(warnings, checkTLSConfig("transport.tls.trustedCaFile", c.Transport.TLS.TrustedCaFile))
	}

	if !slices.Contains(SupportedTransportProtocols, c.Transport.Protocol) {
		errs = AppendError(errs, fmt.Errorf("隧道协议无效, 可选值为 %v", SupportedTransportProtocols))
	}

	for _, f := range c.IncludeConfigFiles {
		absDir, err := filepath.Abs(filepath.Dir(f))
		if err != nil {
			errs = AppendError(errs, fmt.Errorf("错误的 Include: 解析 %s 的目录失败: %v", f, err))
			continue
		}
		if _, err := os.Stat(absDir); os.IsNotExist(err) {
			errs = AppendError(errs, fmt.Errorf("错误的 Include: %s 的目录不存在", f))
		}
	}
	return warnings, errs
}

func ValidateAllClientConfig(c *v1.ClientCommonConfig, proxyCfgs []v1.ProxyConfigurer, visitorCfgs []v1.VisitorConfigurer) (Warning, error) {
	var warnings Warning
	if c != nil {
		warning, err := ValidateClientCommonConfig(c)
		warnings = AppendError(warnings, warning)
		if err != nil {
			return warnings, err
		}
	}

	for _, c := range proxyCfgs {
		if err := ValidateProxyConfigurerForClient(c); err != nil {
			return warnings, fmt.Errorf("隧道 %s: %v", c.GetBaseConfig().Name, err)
		}
	}

	for _, c := range visitorCfgs {
		if err := ValidateVisitorConfigurer(c); err != nil {
			return warnings, fmt.Errorf("访问者 %s: %v", c.GetBaseConfig().Name, err)
		}
	}
	return warnings, nil
}
