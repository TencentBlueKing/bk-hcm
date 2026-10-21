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

package cryptography

import (
	"bytes"
	"crypto/ecdsa"
	"crypto/elliptic"
	"crypto/rand"
	"crypto/rsa"
	"crypto/sha256"
	"crypto/x509"
	"encoding/base64"
	"encoding/pem"
	"testing"

	"github.com/stretchr/testify/assert"
)

func generateTestRSAKeyPEM(t *testing.T, bits int) (*rsa.PrivateKey, string) {
	t.Helper()
	priv, err := rsa.GenerateKey(rand.Reader, bits)
	assert.NoError(t, err)

	der, err := x509.MarshalPKIXPublicKey(&priv.PublicKey)
	assert.NoError(t, err)

	pubPEM := pem.EncodeToMemory(&pem.Block{
		Type:  "PUBLIC KEY",
		Bytes: der,
	})
	return priv, string(pubPEM)
}

func generateTestECDSAKeyPEM(t *testing.T) string {
	t.Helper()
	priv, err := ecdsa.GenerateKey(elliptic.P256(), rand.Reader)
	assert.NoError(t, err)

	der, err := x509.MarshalPKIXPublicKey(&priv.PublicKey)
	assert.NoError(t, err)

	pubPEM := pem.EncodeToMemory(&pem.Block{
		Type:  "PUBLIC KEY",
		Bytes: der,
	})
	return string(pubPEM)
}

func TestParseRSAPublicKey(t *testing.T) {
	priv3072, pub3072PEM := generateTestRSAKeyPEM(t, 3072)
	_, pub4096PEM := generateTestRSAKeyPEM(t, 4096)
	_, pub2048PEM := generateTestRSAKeyPEM(t, 2048)
	pkcs1PEM := string(pem.EncodeToMemory(&pem.Block{
		Type:  "RSA PUBLIC KEY",
		Bytes: x509.MarshalPKCS1PublicKey(&priv3072.PublicKey),
	}))
	ecdsaPEM := generateTestECDSAKeyPEM(t)

	testCases := []struct {
		name     string
		input    string
		wantBits int
		wantErr  bool
	}{
		{name: "valid 3072-bit SPKI public key", input: pub3072PEM, wantBits: 3072, wantErr: false},
		{name: "valid 4096-bit SPKI public key", input: pub4096PEM, wantBits: 4096, wantErr: false},
		{name: "empty PEM string", input: "   ", wantErr: true},
		{name: "not a PEM block", input: "not-a-pem-block", wantErr: true},
		{name: "trailing garbage after PEM block", input: pub3072PEM + "extra-trailing-data", wantErr: true},
		{name: "PKCS1 RSA PUBLIC KEY block rejected", input: pkcs1PEM, wantErr: true},
		{name: "non-RSA (ECDSA) SPKI public key rejected", input: ecdsaPEM, wantErr: true},
		{name: "2048-bit RSA public key rejected (< 3072)", input: pub2048PEM, wantErr: true},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			pub, err := ParseRSAPublicKey(tc.input)
			if tc.wantErr {
				assert.Error(t, err)
				assert.Nil(t, pub)
				return
			}
			assert.NoError(t, err)
			assert.NotNil(t, pub)
			assert.Equal(t, tc.wantBits, pub.N.BitLen())
		})
	}
}

func TestRSAOaep_EncryptToBase64(t *testing.T) {
	priv3072, pub3072PEM := generateTestRSAKeyPEM(t, 3072)
	rsaOaep, err := NewRSAOaepFromPEM(pub3072PEM)
	assert.NoError(t, err)
	assert.Equal(t, 318, rsaOaep.MaxPlaintextBytes())

	testCases := []struct {
		name      string
		plaintext []byte
		wantErr   bool
	}{
		{
			name:      "standard launch_config JSON",
			plaintext: []byte(`{"access_token":"pWqGXk3vEi5xGfE1DTqYV6mKdUyEwF","bk_tenant_id":"default"}`),
			wantErr:   false,
		},
		{
			name:      "exact max length 318 bytes for 3072-bit key",
			plaintext: bytes.Repeat([]byte("a"), 318),
			wantErr:   false,
		},
		{
			name:      "exceed max length 319 bytes for 3072-bit key",
			plaintext: bytes.Repeat([]byte("a"), 319),
			wantErr:   true,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			b64_1, err := rsaOaep.EncryptToBase64(tc.plaintext)
			if tc.wantErr {
				assert.Error(t, err)
				assert.Empty(t, b64_1)
				return
			}

			assert.NoError(t, err)
			// 3072 bits = 384 bytes -> 512 base64 characters.
			assert.Equal(t, 512, len(b64_1))

			// OAEP produces different ciphertexts for the same plaintext.
			b64_2, err := rsaOaep.EncryptToBase64(tc.plaintext)
			assert.NoError(t, err)
			assert.NotEqual(t, b64_1, b64_2)

			rawCipher, err := base64.StdEncoding.DecodeString(b64_1)
			assert.NoError(t, err)

			decrypted, err := rsa.DecryptOAEP(sha256.New(), rand.Reader, priv3072, rawCipher, nil)
			assert.NoError(t, err)
			assert.Equal(t, tc.plaintext, decrypted)
		})
	}
}
