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

// Crypto 定义了需要实现的三组加解密方法，分别是以字节、字符串、Base64字符串格式的方法
type Crypto interface {
	Encrypt(plaintext []byte) []byte
	Decrypt(encryptedText []byte) ([]byte, error)

	EncryptToString(plaintext []byte) string
	DecryptString(encryptedText string) ([]byte, error)

	EncryptToBase64(plaintext string) string
	DecryptFromBase64(encryptedTextB64 string) (string, error)
}
