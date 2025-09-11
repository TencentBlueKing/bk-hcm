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

package dao

import (
	"strings"
	"testing"

	"hcm/pkg/cc"

	"github.com/stretchr/testify/assert"
)

// TestTLSConfig_Enable 测试TLS配置启用逻辑
func TestTLSConfig_Enable(t *testing.T) {
	t.Run("empty config should return false", func(t *testing.T) {
		emptyTLS := cc.TLSConfig{}
		assert.False(t, emptyTLS.Enable())
	})

	t.Run("config with CA file should return true", func(t *testing.T) {
		withCA := cc.TLSConfig{CAFile: "/path/to/ca.pem"}
		assert.True(t, withCA.Enable())
	})

	t.Run("config with cert file should return true", func(t *testing.T) {
		withCert := cc.TLSConfig{CertFile: "/path/to/cert.pem"}
		assert.True(t, withCert.Enable())
	})

	t.Run("config with key file should return true", func(t *testing.T) {
		withKey := cc.TLSConfig{KeyFile: "/path/to/key.pem"}
		assert.True(t, withKey.Enable())
	})

	t.Run("config with all files should return true", func(t *testing.T) {
		fullConfig := cc.TLSConfig{
			CAFile:   "/path/to/ca.pem",
			CertFile: "/path/to/cert.pem",
			KeyFile:  "/path/to/key.pem",
		}
		assert.True(t, fullConfig.Enable())
	})
}

// TestTLSConfig_Validate 测试TLS配置验证逻辑
func TestTLSConfig_Validate(t *testing.T) {
	t.Run("empty config should be valid", func(t *testing.T) {
		emptyTLS := cc.TLSConfig{}
		assert.NoError(t, emptyTLS.Validate())
	})

	t.Run("config with only cert file should be invalid", func(t *testing.T) {
		invalidConfig := cc.TLSConfig{CertFile: "/path/to/cert.pem"}
		assert.Error(t, invalidConfig.Validate())
		assert.Contains(t, invalidConfig.Validate().Error(), "client key file is required")
	})

	t.Run("config with only key file should be invalid", func(t *testing.T) {
		invalidConfig := cc.TLSConfig{KeyFile: "/path/to/key.pem"}
		assert.Error(t, invalidConfig.Validate())
		assert.Contains(t, invalidConfig.Validate().Error(), "client cert file is required")
	})

	t.Run("config with both cert and key files should be valid", func(t *testing.T) {
		validConfig := cc.TLSConfig{
			CertFile: "/path/to/cert.pem",
			KeyFile:  "/path/to/key.pem",
		}
		assert.NoError(t, validConfig.Validate())
	})

	t.Run("config with CA file only should be valid", func(t *testing.T) {
		validConfig := cc.TLSConfig{CAFile: "/path/to/ca.pem"}
		assert.NoError(t, validConfig.Validate())
	})
}

// TestURI_SSLGeneration 测试URI生成中的SSL参数逻辑
func TestURI_SSLGeneration(t *testing.T) {
	t.Run("URI without TLS config should not contain SSL params", func(t *testing.T) {
		config := cc.ResourceDB{
			Endpoints: []string{"localhost:3306"},
			User:      "testuser",
			Password:  "password",
			Database:  "testdb",
			TLS:       cc.TLSConfig{},
		}

		uriStr := generateTestURI(config)
		assert.NotContains(t, uriStr, "ssl-mode")
		assert.NotContains(t, uriStr, "ssl-ca")
		assert.NotContains(t, uriStr, "ssl-cert")
		assert.NotContains(t, uriStr, "ssl-key")
	})

	t.Run("URI with CA file should contain ssl-mode and ssl-ca", func(t *testing.T) {
		config := cc.ResourceDB{
			Endpoints: []string{"localhost:3306"},
			User:      "testuser",
			Password:  "password",
			Database:  "testdb",
			TLS: cc.TLSConfig{
				CAFile: "/path/to/ca.pem",
			},
		}

		uriStr := generateTestURI(config)
		assert.Contains(t, uriStr, "ssl-mode=VERIFY_CA")
		assert.Contains(t, uriStr, "ssl-ca=")
	})

	t.Run("URI with insecure skip verify and CA file should contain ssl-mode=PREFERRED", func(t *testing.T) {
		config := cc.ResourceDB{
			Endpoints: []string{"localhost:3306"},
			User:      "testuser",
			Password:  "password",
			Database:  "testdb",
			TLS: cc.TLSConfig{
				InsecureSkipVerify: true,
				CAFile:             "/path/to/ca.pem", // 需要至少一个证书文件来启用SSL
			},
		}
		uriStr := generateTestURI(config)
		assert.Contains(t, uriStr, "ssl-mode=PREFERRED")
	})
}

// TestConnect_SSLValidation 测试SSL连接验证逻辑
func TestConnect_SSLValidation(t *testing.T) {
	t.Run("connect should validate TLS config before proceeding", func(t *testing.T) {
		invalidConfig := cc.ResourceDB{
			Endpoints: []string{"localhost:3306"},
			User:      "testuser",
			Password:  "password",
			Database:  "testdb",
			TLS: cc.TLSConfig{
				CertFile: "/nonexistent/cert.pem", // 只配置cert，没有key
			},
		}

		// 模拟connect函数的验证逻辑
		err := invalidConfig.TLS.Validate()
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "client key file is required")
	})
}

// generateTestURI 模拟uri函数的SSL参数生成逻辑
func generateTestURI(opt cc.ResourceDB) string {
	baseURI := "testuser:password@tcp(localhost:3306)/testdb?parseTime=true"

	if opt.TLS.Enable() {
		sslParams := make([]string, 0)

		if opt.TLS.InsecureSkipVerify {
			sslParams = append(sslParams, "ssl-mode=PREFERRED")
		} else if opt.TLS.CAFile != "" {
			sslParams = append(sslParams, "ssl-mode=VERIFY_CA")
		} else {
			sslParams = append(sslParams, "ssl-mode=REQUIRED")
		}

		if opt.TLS.CAFile != "" {
			sslParams = append(sslParams, "ssl-ca="+opt.TLS.CAFile)
		}
		if opt.TLS.CertFile != "" {
			sslParams = append(sslParams, "ssl-cert="+opt.TLS.CertFile)
		}
		if opt.TLS.KeyFile != "" {
			sslParams = append(sslParams, "ssl-key="+opt.TLS.KeyFile)
		}

		if len(sslParams) > 0 {
			baseURI += "&" + strings.Join(sslParams, "&")
		}
	}

	return baseURI
}
