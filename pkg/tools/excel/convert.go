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
	"strconv"
	"strings"
)

// ConvertCellValue 根据Header类型定义将字符串单元格值转换为对应Go类型。
// 空白值原样返回；转换失败时回退返回原始字符串。
func ConvertCellValue(val string, header Header) interface{} {
	trimmed := strings.TrimSpace(val)
	if trimmed == "" {
		return val
	}

	switch {
	case header.Type == "int":
		if v, err := strconv.ParseInt(trimmed, 10, 64); err == nil {
			return v
		}
	case strings.HasPrefix(header.Type, "float"):
		if v, err := strconv.ParseFloat(trimmed, 64); err == nil {
			return v
		}
	case header.Type == "enum":
		return convertEnumValue(trimmed, header.Value)
	}

	return val
}

// convertEnumValue 从枚举值定义推断目标类型并转换单元格字符串。
// JSON反序列化后数值类型为float64，优先尝试int解析以保留整数语义。
func convertEnumValue(val string, enumValues []interface{}) interface{} {
	if len(enumValues) == 0 {
		return val
	}

	switch enumValues[0].(type) {
	case float64:
		if v, err := strconv.ParseInt(val, 10, 64); err == nil {
			return v
		}
		if v, err := strconv.ParseFloat(val, 64); err == nil {
			return v
		}
	}

	return val
}
