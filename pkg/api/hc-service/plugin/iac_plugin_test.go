/*
 * TencentBlueKing is pleased to support the open source community by making
 * 蓝鲸智云 - 混合云管理平台 (BlueKing - Hybrid Cloud Management System) available.
 * Copyright (C) 2022 THL A29 Limited,
 * a Tencent company. All rights reserved.
 * Licensed under the MIT License (the "License");
 * you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at http://opensource.org/licenses/MIT
 * Unless required by applicable law or agreed to in writing,
 * software distributed under the License is distributed on
 * an "AS IS" BASIS, WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND,
 * either express or implied. See the License for the
 * specific language governing permissions and limitations under the License.
 *
 * We undertake not to change the open source license (MIT license) applicable
 *
 * to the current version of the project delivered to anyone in the future.
 */

package plugin

import (
	"crypto/rand"
	"crypto/rsa"
	"crypto/x509"
	"encoding/pem"
	"strings"
	"testing"

	coreplugin "hcm/pkg/api/core/plugin"
	cvt "hcm/pkg/tools/converter"
)

func TestHTTPRegisterReqValidate(t *testing.T) {
	key, err := rsa.GenerateKey(rand.Reader, 3072)
	if err != nil {
		t.Fatalf("generate rsa key failed, err: %v", err)
	}
	der, err := x509.MarshalPKIXPublicKey(&key.PublicKey)
	if err != nil {
		t.Fatalf("marshal rsa key failed, err: %v", err)
	}
	publicKey := string(pem.EncodeToMemory(&pem.Block{Type: "PUBLIC KEY", Bytes: der}))

	valid := func() *HTTPRegisterReq {
		return &HTTPRegisterReq{PluginName: "bk-hcm-plugin.cmdb_1", APIPath: "/api/p/prod", CryptoType: "RSA",
			CryptoConfig: &coreplugin.HTTPCryptoConfig{PublicKey: publicKey}}
	}

	if err := valid().Validate(); err != nil {
		t.Errorf("valid request: unexpected err: %v", err)
	}

	cases := map[string]func(r *HTTPRegisterReq){
		"缺少插件名":   func(r *HTTPRegisterReq) { r.PluginName = "" },
		"插件名含斜杠":  func(r *HTTPRegisterReq) { r.PluginName = "a/b" },
		"插件名特殊开头": func(r *HTTPRegisterReq) { r.PluginName = "-a" },
		"插件名超长":   func(r *HTTPRegisterReq) { r.PluginName = strings.Repeat("a", 65) },
		"缺少地址":    func(r *HTTPRegisterReq) { r.APIPath = "" },
		"地址带主机":   func(r *HTTPRegisterReq) { r.APIPath = "//evil.example.com/x" },
		"缺少加密配置":  func(r *HTTPRegisterReq) { r.CryptoConfig = nil },
		"加密类型不支持": func(r *HTTPRegisterReq) { r.CryptoType = "AES" },
		"备注超长":    func(r *HTTPRegisterReq) { r.Memo = cvt.ValToPtr(strings.Repeat("a", 256)) },
	}
	for name, mutate := range cases {
		req := valid()
		mutate(req)
		if err := req.Validate(); err == nil {
			t.Errorf("%s: expect error", name)
		}
	}
}

func TestHTTPUpdateReqValidate(t *testing.T) {
	if err := (&HTTPUpdateReq{}).Validate(); err == nil {
		t.Errorf("empty update: expect error")
	}
	if err := (&HTTPUpdateReq{PluginName: "http-demo"}).Validate(); err == nil {
		t.Errorf("plugin name only: expect error")
	}
	if err := (&HTTPUpdateReq{PluginName: "http-demo", Memo: cvt.ValToPtr("m")}).Validate(); err != nil {
		t.Errorf("memo only: unexpected err: %v", err)
	}
	if err := (&HTTPUpdateReq{PluginName: "http-demo", APIPath: cvt.ValToPtr("/a/../b")}).Validate(); err == nil {
		t.Errorf("invalid api_path: expect error")
	}
}
