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

package iacplugin

import (
	"errors"
	"fmt"
	"net/url"
	"strings"
	"unicode"

	coreplugin "hcm/pkg/api/core/plugin"
	"hcm/pkg/cryptography"
)

const (
	// CryptoTypeRSA http 插件加密类型：RSA。
	CryptoTypeRSA = "RSA"
	// MinRSAKeyBits RSA 公钥的最小长度。
	MinRSAKeyBits = cryptography.MinRSAPublicKeyBits
)

// ValidateAPIPath 校验 http 插件 api_path 的形状：必须是以 / 开头的非空路径，
// 不能带 query、fragment、反斜杠、空白与控制字符，各路径段不能为空、不能是 . 或 ..（含 URL 编码形式）。
func ValidateAPIPath(apiPath string) error {
	if len(apiPath) == 0 {
		return errors.New("api_path is required")
	}

	if apiPath[0] != '/' {
		return errors.New("api_path should start with /")
	}

	unescaped, err := url.PathUnescape(apiPath)
	if err != nil {
		return fmt.Errorf("api_path is invalid, err: %v", err)
	}

	for _, p := range []string{apiPath, unescaped} {
		for _, r := range p {
			if unicode.IsSpace(r) || unicode.IsControl(r) {
				return errors.New("api_path should not contain whitespace or control characters")
			}
		}
		if strings.ContainsAny(p, `?#\`) {
			return errors.New(`api_path should not contain ?, # or \`)
		}
		for _, seg := range strings.Split(p[1:], "/") {
			if seg == "" || seg == "." || seg == ".." {
				return errors.New("api_path should not contain empty, . or .. segment")
			}
		}
	}

	return nil
}

// ValidateCrypto 校验 http 插件的加密配置。目前只支持 RSA：公钥必须是 PEM 编码的 SPKI，类型为 RSA，长度不低于 3072 位。
func ValidateCrypto(cryptoType string, cfg *coreplugin.HTTPCryptoConfig) error {
	switch cryptoType {
	case CryptoTypeRSA:
	case "":
		return errors.New("crypto_type is required")
	default:
		return fmt.Errorf("unsupported crypto_type: %s, only %s is supported", cryptoType, CryptoTypeRSA)
	}

	if cfg == nil || len(cfg.PublicKey) == 0 {
		return errors.New("crypto_config.public_key is required")
	}

	_, err := cryptography.ParseRSAPublicKey(cfg.PublicKey)
	return err
}
