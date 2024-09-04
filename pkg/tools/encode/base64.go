/*
 * TencentBlueKing is pleased to support the open source community by making
 * 蓝鲸智云 - 混合云管理平台 (BlueKing - Hybrid Cloud Management System) available.
 * Copyright (C) 2024 THL A29 Limited,
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

package encode

import (
	"bytes"
	"encoding/base64"
	"fmt"
	"io"
	"strings"
)

// Base64StrToReader converts a base64 string to an io.Reader.
func Base64StrToReader(str string) io.Reader {
	return base64.NewDecoder(base64.StdEncoding, strings.NewReader(str))
}

// ReaderToBase64Str converts an io.Reader to a base64 string.
func ReaderToBase64Str(reader io.Reader) (string, error) {
	// 创建一个base64编码器，它实现了io.Writer接口
	var b64 bytes.Buffer
	encoder := base64.NewEncoder(base64.StdEncoding, &b64)

	// 创建一个管道，将reader的数据写入到encoder中
	// 这里使用io.Copy进行数据的复制操作，它会在遇到EOF时停止
	if _, err := io.Copy(encoder, reader); err != nil {
		return "", fmt.Errorf("failed to copy data to base64 encoder: %v", err)
	}

	// 必须关闭编码器以确保所有数据都被写入
	if err := encoder.Close(); err != nil {
		return "", err
	}

	// 此时b64中包含了base64编码后的数据
	return b64.String(), nil
}
