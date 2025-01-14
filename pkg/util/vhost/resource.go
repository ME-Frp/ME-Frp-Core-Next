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
<html>
<head>
<meta charset="UTF-8">
<meta name="viewport" content="width=device-width,user-scalable=no,initial-scale=1.0,maximum-scale=1.0,minimum-scale=1.0">
<title>服务不可用 | ME Frp</title>
<pnk rel="icon" href="https://mefrp-preview.lxhtt.cn/assets/img/logo.svg" />
</head>
<body>
<div class="container">
<h1>⚠ 503 - 服务不可用</h1>
<p>如果您是隧道所有者，造成无法访问的原因可能有：</p>
<ul>
<p>您访问的网站使用了镜缘映射，但客户端不在线。</p>
<p>该网站或隧道已被管理员封禁。</p>
<p>域名解析 未生效 或 解析错误。</p>
</ul></br>
<p>如果您是普通访问者，您可以：</p>
<ul>
<p>稍等一段时间后再次尝试访问此站点。</p>
<p>尝试与该网站的所有者取得联系。</p>
<p>刷新您的 DNS 缓存或在其他网络环境访问。</p>
</ul>
<p class="powered-by">&copy; <a target="_blank" href="https://www.mefrp.com/">ME Frp</a> 项目组 2021-2025. </br>Frp 内网穿透联盟统一识别编码：<a target="_blank" href="https://内网穿透.中国/">AZWB66WB</a></p>
</div>
</body>
<style>
* {
margin: 0;
padding: 0;
font-family: 'Microsoft YaHei', Arial, sans-serif;
}
.container {
height: 100vh;
display: flex;
justify-content: center;
flex-direction: column;
align-items: left;
padding-left: 10%;
}
.container h1 {
font-weight: 400;
margin-bottom: 1rem;
}
.container p {
color: gray;
}
.container ul {
margin-top: 1rem;
}
.container .powered-by {
margin-top: 2rem;
}
.container .powered-by a {
color: rgb(21, 129, 218);
text-decoration: none;
transition: 0.3s;
}
.container .powered-by a:hover {
color: rgb(24, 144, 243);
}
</style>
<style>
@media (prefers-color-scheme: dark) {
html {
background-color: rgb(41, 41, 41);
}
.container h1 {
color: white;
}
}
</style>
</html>`
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
