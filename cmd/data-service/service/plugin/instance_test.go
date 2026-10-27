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
	"encoding/json"
	"reflect"
	"strings"
	"testing"

	"hcm/pkg/cryptography"
)

func newTestCipher(t *testing.T) cryptography.Crypto {
	t.Helper()
	key := []byte("0123456789abcdef0123456789abcdef") // 32 bytes
	nonce := []byte("0123456789ab")                   // 12 bytes
	c, err := cryptography.NewAESGcm(key, nonce)
	if err != nil {
		t.Fatalf("create test AESGcm failed: %v", err)
	}
	return c
}

func TestEncryptAndDecryptSecretAttrsRoundTrip(t *testing.T) {
	cipher := newTestCipher(t)

	input := map[string]any{
		"secret_key": "SK-top-secret",
		"pin_code":   float64(8848),
		"strict_tls": true,
		"secret_headers": map[string]any{
			"X-Token": "tk-99",
		},
		"auth": []any{
			map[string]any{}, // 稀疏数组下标 0 占位符
			map[string]any{
				"password": "s3cret-password-02",
			},
		},
	}

	encryptedJSON, err := encryptSecretAttrs(input, cipher)
	if err != nil {
		t.Fatalf("encryptSecretAttrs failed: %v", err)
	}

	rawStr := string(encryptedJSON)
	// 验证明文敏感值不会出现在密文 JSON 中，而嵌套结构 key 与空对象占位符保留
	for _, secretPlain := range []string{"SK-top-secret", "8848", "tk-99", "s3cret-password-02"} {
		if strings.Contains(rawStr, secretPlain) {
			t.Fatalf("expected encrypted JSON not to contain plaintext %q, got: %s", secretPlain, rawStr)
		}
	}
	for _, structuralKey := range []string{"secret_key", "secret_headers", "X-Token", "auth", "password", "{}"} {
		if !strings.Contains(rawStr, structuralKey) {
			t.Fatalf("expected encrypted JSON to preserve structural key/placeholder %q, got: %s",
				structuralKey, rawStr)
		}
	}

	var storedTree map[string]any
	if err := json.Unmarshal([]byte(encryptedJSON), &storedTree); err != nil {
		t.Fatalf("unmarshal storedTree failed: %v", err)
	}
	authSlice, ok := storedTree["auth"].([]any)
	if !ok || len(authSlice) != 2 {
		t.Fatalf("expected auth slice of len 2, got: %v", storedTree["auth"])
	}
	elem0, ok := authSlice[0].(map[string]any)
	if !ok || len(elem0) != 0 {
		t.Fatalf("expected index 0 placeholder {} to remain empty map, got: %v", authSlice[0])
	}

	// 解密还原并验证标量类型（string、float64、bool）完全一致
	decrypted, err := decryptSecretAttrs(encryptedJSON, cipher)
	if err != nil {
		t.Fatalf("decryptSecretAttrs failed: %v", err)
	}

	if !reflect.DeepEqual(decrypted, input) {
		t.Fatalf("round-trip mismatch:\ngot:  %#v\nwant: %#v", decrypted, input)
	}
}

func TestEncryptAndDecryptEmptySecretAttrs(t *testing.T) {
	cipher := newTestCipher(t)

	enc, err := encryptSecretAttrs(nil, cipher)
	if err != nil {
		t.Fatalf("encrypt nil secretAttrs failed: %v", err)
	}
	if string(enc) != "{}" {
		t.Fatalf("expected {}, got %s", enc)
	}

	dec, err := decryptSecretAttrs(enc, cipher)
	if err != nil {
		t.Fatalf("decrypt empty secretAttrs failed: %v", err)
	}
	if len(dec) != 0 {
		t.Fatalf("expected empty map, got %v", dec)
	}
}
