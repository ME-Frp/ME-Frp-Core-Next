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
<title>镜缘映射 · ME Frp</title>
<style>
body{font-family:Arial,sans-serif;background-color:#f4f4f4;margin:0;line-height:1.6;padding:0}
.container{width:100%;max-width:600px;margin:0 auto;background-color:#fff;padding:20px;box-shadow:0 0 10px rgba(0,0,0,.1);font-size:16px}
.header{text-align:center;padding:10px 0;display:flex;align-items:center;justify-content:center}
.header img{height:48px;width:48px;margin-right:10px}
.header h1{margin:0;font-size:30px}
.content{margin:20px 0}
.footer{text-align:center;color:#888;font-size:12px;margin-top:20px}
</style>
</head>
<body>
<div class="container">
<div class="header">
<img src="https://img.ltyears.com/editor/23904/20241227/0d0d374e13bb5ea22e17e73a6c49f5e5.svg?t=1735314090679" alt="ME Frp Logo">
<h1>镜缘映射 · ME Frp</h1>
</div>
<div class="content">
<h2>404 - 页面未找到</h2>
<p>抱歉, 您访问的页面不存在。</p>
<p>请检查您输入的网址是否正确, 或返回首页。</p>
</div>
<div class="footer">
<p>此页面由系统自动生成</p>
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
			log.Warnf("读取自定义 404 页面错误: %v", err)
			buf = []byte(NotFound)
		}
	} else {
		buf = []byte(NotFound)
	}
	return buf
}

func NotFoundResponse() *http.Response {
	header := make(http.Header)
	header.Set("server", "frp/"+version.Full())
	header.Set("Content-Type", "text/html")

	content := getNotFoundPageContent()
	res := &http.Response{
		Status:        "Not Found",
		StatusCode:    404,
		Proto:         "HTTP/1.1",
		ProtoMajor:    1,
		ProtoMinor:    1,
		Header:        header,
		Body:          io.NopCloser(bytes.NewReader(content)),
		ContentLength: int64(len(content)),
	}
	return res
}
