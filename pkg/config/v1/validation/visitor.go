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
	"errors"
	"fmt"
	"slices"

	v1 "github.com/fatedier/frp/pkg/config/v1"
)

func ValidateVisitorConfigurer(c v1.VisitorConfigurer) error {
	base := c.GetBaseConfig()
	if err := validateVisitorBaseConfig(base); err != nil {
		return err
	}

	switch v := c.(type) {
	case *v1.STCPVisitorConfig:
	case *v1.SUDPVisitorConfig:
	case *v1.XTCPVisitorConfig:
		return validateXTCPVisitorConfig(v)
	default:
		return errors.New("未知访问者配置类型")
	}
	return nil
}

func validateVisitorBaseConfig(c *v1.VisitorBaseConfig) error {
	if c.Name == "" {
		return errors.New("访问者名不能为空")
	}

	if c.ServerName == "" {
		return errors.New("服务端名不能为空")
	}

	if c.BindPort == 0 {
		return errors.New("绑定端口不能为空")
	}
	return nil
}

func validateXTCPVisitorConfig(c *v1.XTCPVisitorConfig) error {
	if !slices.Contains([]string{"kcp", "quic"}, c.Protocol) {
		return fmt.Errorf("协议应为 [kcp] 或 [quic]")
	}
	return nil
}
