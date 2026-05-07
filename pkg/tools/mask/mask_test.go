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

package mask

import (
	"testing"
)

func TestMaskEmail(t *testing.T) {
	tests := []struct {
		name     string
		email    string
		expected string
	}{
		{
			name:     "正常邮箱-标准格式",
			email:    "example@gmail.com",
			expected: "ex***@gmail.com",
		},
		{
			name:     "正常邮箱-短用户名",
			email:    "ab@qq.com",
			expected: "ab***@qq.com",
		},
		{
			name:     "正常邮箱-长用户名",
			email:    "longusername@163.com",
			expected: "lo***@163.com",
		},
		{
			name:     "正常邮箱-多级域名",
			email:    "user@mail.example.com",
			expected: "us***@mail.example.com",
		},
		{
			name:     "正常邮箱-.cn后缀",
			email:    "test@company.cn",
			expected: "te***@company.cn",
		},
		{
			name:     "空字符串",
			email:    "",
			expected: "",
		},
		{
			name:     "不带@的字符串",
			email:    "notanemail",
			expected: "notanemail",
		},
		{
			name:     "@在开头的邮箱",
			email:    "@gmail.com",
			expected: "@gmail.com",
		},
		{
			name:     "用户名太短-1位",
			email:    "a@gmail.com",
			expected: "a@gmail.com",
		},
		{
			name:     "域名没有点",
			email:    "user@localhost",
			expected: "us***@localhost",
		},
		{
			name:     "域名点在开头",
			email:    "user@.com",
			expected: "us***@.com",
		},
		{
			name:     "带特殊字符的邮箱",
			email:    "user.name+tag@gmail.com",
			expected: "us***@gmail.com",
		},
		{
			name:     "边界-用户名刚好2个字符",
			email:    "ab@gmail.com",
			expected: "ab***@gmail.com",
		},
		{
			name:     "边界-域名只有1个字符",
			email:    "user@a",
			expected: "us***@a",
		},
		{
			name:     "边界-用户名2字符域名1字符",
			email:    "ab@c",
			expected: "ab***@c",
		},
		{
			name:     "边界-域名主域名只有1字符",
			email:    "user@a.com",
			expected: "us***@a.com",
		},
		{
			name:     "边界-多级域名主域名2字符",
			email:    "user@ab.example.com",
			expected: "us***@ab.example.com",
		},
		{
			name:     "边界-@后面直接是点",
			email:    "user@.example.com",
			expected: "us***@.example.com",
		},
		{
			name:     "边界-只有@和域名后缀",
			email:    "user@.com",
			expected: "us***@.com",
		},
		{
			name:     "边界-用户名最短有效长度且域名正常",
			email:    "ab@xyz.com",
			expected: "ab***@xyz.com",
		},
		{
			name:     "多个@符号-只处理第一个",
			email:    "user@name@domain.com",
			expected: "us***@name@domain.com",
		},
		{
			name:     "全是空格",
			email:    "   ",
			expected: "",
		},
		{
			name:     "带空格的邮箱",
			email:    "  example@gmail.com  ",
			expected: "ex***@gmail.com",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := MaskEmail(tt.email)
			if result != tt.expected {
				t.Errorf("MaskEmail(%q) = %q, expected %q", tt.email, result, tt.expected)
			}
		})
	}
}

func TestMaskPhone(t *testing.T) {
	tests := []struct {
		name     string
		phone    string
		expected string
	}{
		{
			name:     "正常手机号-11位",
			phone:    "13800138000",
			expected: "138****8000",
		},
		{
			name:     "正常手机号-带空格",
			phone:    " 13800138000 ",
			expected: "138****8000",
		},
		{
			name:     "手机号-12位",
			phone:    "138001380001",
			expected: "138****0001",
		},
		{
			name:     "手机号-带国际区号",
			phone:    "+8613800138000",
			expected: "+86****8000",
		},
		{
			name:     "空字符串",
			phone:    "",
			expected: "",
		},
		{
			name:     "长度6位-不脱敏",
			phone:    "123456",
			expected: "123456",
		},
		{
			name:     "长度7位-刚好满足脱敏条件",
			phone:    "1234567",
			expected: "123****4567",
		},
		{
			name:     "长度8位",
			phone:    "12345678",
			expected: "123****5678",
		},
		{
			name:     "带横杠的手机号",
			phone:    "138-0013-8000",
			expected: "138****8000",
		},
		{
			name:     "带空格的手机号",
			phone:    "138 0013 8000",
			expected: "138****8000",
		},
		{
			name:     "边界-刚好7位手机号",
			phone:    "1234567",
			expected: "123****4567",
		},
		{
			name:     "边界-刚好7位带空格",
			phone:    " 1234567 ",
			expected: "123****4567",
		},
		{
			name:     "边界-8位手机号",
			phone:    "12345678",
			expected: "123****5678",
		},
		{
			name:     "边界-9位手机号",
			phone:    "123456789",
			expected: "123****6789",
		},
		{
			name:     "不脱敏-6位",
			phone:    "123456",
			expected: "123456",
		},
		{
			name:     "不脱敏-1位",
			phone:    "1",
			expected: "1",
		},
		{
			name:     "全是空格",
			phone:    "   ",
			expected: "",
		},
		{
			name:     "带特殊字符横杠",
			phone:    "138-0000-0000",
			expected: "138****0000",
		},
		{
			name:     "带括号等字符",
			phone:    "(020)13800000000",
			expected: "(02****0000",
		},
		{
			name:     "国际号码-带空格",
			phone:    "+86 138 0000 0000",
			expected: "+86****0000",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := MaskPhone(tt.phone)
			if result != tt.expected {
				t.Errorf("MaskPhone(%q) = %q, expected %q", tt.phone, result, tt.expected)
			}
		})
	}
}