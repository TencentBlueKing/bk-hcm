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

// Package mask ...
package mask

import (
	"strings"
)

const (
	// maskChar 用于脱敏的替换字符
	maskChar = "*"
	// minMaskLen 最小脱敏长度，低于此长度不进行脱敏
	minMaskLen = 2
)

// MaskEmail 邮箱脱敏
// 例如：example@gmail.com -> ex***@gmail.com
// 规则：
//  1. 如果用户名为空或长度小于2，返回原邮箱
//  2. 用户名保留前2位，后面用***替换
//  3. @后面的域名部分保持原样，不进行脱敏
//
// 参数 email 为待脱敏的邮箱地址
// 返回脱敏后的邮箱，如果输入为空或格式不正确则返回原字符串
func MaskEmail(email string) string {
	if email == "" {
		return email
	}

	// 去除可能的首尾空格
	email = strings.TrimSpace(email)

	// 查找@符号位置
	atIndex := strings.Index(email, "@")
	if atIndex == -1 || atIndex == 0 {
		// 没有@符号或@在开头，返回原字符串
		return email
	}

	// 用户名部分
	username := email[:atIndex]
	if len(username) < minMaskLen {
		// 用户名太短，不脱敏
		return email
	}

	// 域名部分（保持原样，不脱敏）
	domain := email[atIndex:] // 包含@符号及后面的域名

	// 用户名保留前2位
	maskedUsername := username[:2] + strings.Repeat(maskChar, 3)

	return maskedUsername + domain
}

// MaskPhone 手机号脱敏
// 例如：13800138000 -> 138****8000
// 规则：
//  1. 保留前3位和后4位
//  2. 中间用****替换
//  3. 如果长度小于7位，不进行脱敏，直接返回原号码
//  4. 自动去除首尾空格
//
// 参数 phone 为待脱敏的手机号码
// 返回脱敏后的手机号，如果输入为空或长度不足则返回原字符串
func MaskPhone(phone string) string {
	if phone == "" {
		return phone
	}

	// 去除可能的首尾空格
	phone = strings.TrimSpace(phone)

	// 手机号长度至少7位才能脱敏（前3位 + 后4位）
	if len(phone) < 7 {
		return phone
	}

	// 保留前3位和后4位
	prefix := phone[:3]
	suffix := phone[len(phone)-4:]

	return prefix + "****" + suffix
}
