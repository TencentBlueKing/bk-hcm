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

package excel

import (
	"fmt"
	"strings"
)

const splitChar = "[::]"

// CombineFileNameAndSheet 将文件名和sheet名称合并为一个字符串
func CombineFileNameAndSheet(fileName, sheet string) string {
	return fileName + splitChar + sheet
}

// splitToFileNameAndSheet 将字符串分割为文件名和sheet名称
func splitToFileNameAndSheet(name string) (string, string, error) {
	parts := strings.Split(name, splitChar)
	if len(parts) != 2 {
		return "", "", fmt.Errorf("invalid name: %s", name)
	}
	return parts[0], parts[1], nil
}
