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

package vhost

import (
	"bytes"
	"io"
	"net/http"
	"os"
	"strings"

	"github.com/fatedier/frp/pkg/util/log"
	"github.com/fatedier/frp/pkg/util/version"
)

var NotFoundPagePath = ""

const (
	NotFound = `<!DOCTYPE html>
<html lang="zh-CN">
<head>
<meta charset="UTF-8">
<meta name="viewport" content="width=device-width,initial-scale=1">
<title>503 Service Unavailable | ME Frp</title>
<style>
body{font-family:Arial,sans-serif;background-color:#f4f4f4;margin:0;line-height:1.6;padding:0}
.container{width:100%;max-width:600px;margin:0 auto;background-color:#fff;padding:20px;box-shadow:0 0 10px rgba(0,0,0,.1);font-size:16px}
.header{text-align:center;padding:10px 0;display:flex;align-items:center;justify-content:center}
.header h2{margin:0;font-size:30px}
.content{margin:20px 0}
.footer{text-align:center;color:#888;font-size:12px;margin-top:20px}
</style>
</head>
<body>
<div class="container">
<div class="header">
<h2>镜缘映射 · ME Frp</h2>
</div>
<div class="content">
<h3>503 - 服务不可用</h3>
<p>如果您是隧道所有者，造成无法访问的原因可能有：</p>
<li>您访问的网站使用了内网穿透，但是对应的客户端没有运行。</li>
<li>该网站或隧道已被管理员临时或永久禁止连接。</li>
<li>域名解析更改还未生效或解析错误，请检查设置是否正确。</li>
<li>隧道使用的必须为HTTP或HTTPS类型的隧道</li>
<p>如果您是普通访问者，您可以：</p>
<li>稍等一段时间后再次尝试访问此站点。</li>
<li>尝试与该网站的所有者取得联系。</li>
<li>刷新您的 DNS 缓存或在其他网络环境访问。</li>
</div>
<div class="footer">
<p>此页面由 ME Frp 服务端自动生成</p>
<p>&copy; ME Frp 项目组 2021-2025.</p>
<p>Frp 内网穿透联盟统一识别编码：AZWB66WB</p>
</div>
</div>
</body>
</html>
`
)

func getNotFoundPageContent() []byte {
	var (
		buf []byte
		err error
	)
	if NotFoundPagePath != "" {
		buf, err = os.ReadFile(NotFoundPagePath)
		if err != nil {
			log.Warnf("读取自定义 503 页面错误: %v", err)
			buf = []byte(NotFound)
		}
	} else {
		buf = []byte(NotFound)
	}
	return buf
}

func NotFoundResponse() *http.Response {
	header := make(http.Header)
	header.Set("Server", "MEFrp/"+strings.Split(version.Full(), "_")[1])
	header.Set("Content-Type", "text/html")

	content := getNotFoundPageContent()
	res := &http.Response{
		Status:        "Service Unavailable",
		StatusCode:    503,
		Proto:         "HTTP/1.1",
		ProtoMajor:    1,
		ProtoMinor:    1,
		Header:        header,
		Body:          io.NopCloser(bytes.NewReader(content)),
		ContentLength: int64(len(content)),
	}
	return res
}
