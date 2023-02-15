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
	"encoding/base64"

	"github.com/TencentBlueKing/gopkg/conv"
	gopkgcryptography "github.com/TencentBlueKing/gopkg/cryptography"
)

// AESGcm : 是对于gopkg里cryptography的扩展，提供支持Base64字符串的密文及明文加密为Base64字符串
type AESGcm struct {
	*gopkgcryptography.AESGcm
}

// NewAESGcm returns a new AES-GCM instance
func NewAESGcm(key []byte, nonce []byte) (*AESGcm, error) {
	gopkgAesGcm, err := gopkgcryptography.NewAESGcm(key, nonce)
	if err != nil {
		return nil, err
	}

	return &AESGcm{AESGcm: gopkgAesGcm}, nil

}

// EncryptToBase64 : 将字符串明文，使用AES Gcm算法加密后再转化为Base64格式的字符串
func (a *AESGcm) EncryptToBase64(plaintext string) string {
	plaintextBytes := conv.StringToBytes(plaintext)
	encryptedText := a.Encrypt(plaintextBytes)
	return base64.StdEncoding.EncodeToString(encryptedText)
}

// DecryptFromBase64 : 将Base64格式的AES Gcm密文，解密为明文字符串
func (a *AESGcm) DecryptFromBase64(encryptedTextB64 string) (plaintext string, err error) {
	var encryptedText []byte
	encryptedText, err = base64.StdEncoding.DecodeString(encryptedTextB64)
	if err != nil {
		return
	}

	var plaintextBytes []byte
	plaintextBytes, err = a.Decrypt(encryptedText)
	if err != nil {
		return
	}

	return conv.BytesToString(plaintextBytes), err
}
