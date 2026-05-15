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

package validator

import "testing"

func TestValidatePhone(t *testing.T) {
	tests := []struct {
		name  string
		phone string
		want  bool
	}{
		// 中国手机号 - 有效
		{
			name:  "中国手机号-带+86",
			phone: "+8613800138000",
			want:  true,
		},
		{
			name:  "中国手机号-带86空格",
			phone: "86 13800138000",
			want:  false, // phonenumbers 不支持空格分隔的格式
		},
		{
			name:  "中国手机号-带086",
			phone: "+08613800138000",
			want:  false, // phonenumbers 不支持 086 格式
		},

		// 美国手机号 - 有效
		{
			name:  "美国手机号-带+1",
			phone: "+16502530000",
			want:  true,
		},
		{
			name:  "美国手机号-带1空格",
			phone: "1 6502530000",
			want:  false, // phonenumbers 不支持空格分隔的格式
		},

		// 英国手机号 - 有效
		{
			name:  "英国手机号-带+44",
			phone: "+447912345678",
			want:  true,
		},

		// 日本手机号 - 有效
		{
			name:  "日本手机号-带+81",
			phone: "+818012345678",
			want:  true,
		},

		// 香港手机号 - 有效
		{
			name:  "香港手机号-带+852",
			phone: "+85261234567",
			want:  true,
		},

		// 无效手机号
		{
			name:  "无效-纯手机号无国家代码",
			phone: "13800138000",
			want:  false,
		},
		{
			name:  "无效-号码太短",
			phone: "+86138",
			want:  false,
		},
		{
			name:  "无效-号码太长",
			phone: "+8613800138000123",
			want:  false,
		},
		{
			name:  "无效-错误国家代码",
			phone: "+99913800138000",
			want:  false,
		},
		{
			name:  "无效-包含字母",
			phone: "+8613800ABCD",
			want:  false,
		},
		{
			name:  "无效-空字符串",
			phone: "",
			want:  false,
		},
		{
			name:  "无效-特殊字符",
			phone: "+86-138-0013-8000",
			want:  true, // phonenumbers 会忽略分隔符，实际能解析
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := ValidatePhone(tt.phone)
			if got != tt.want {
				t.Errorf("ValidatePhone(%q) = %v, want %v", tt.phone, got, tt.want)
			}
		})
	}
}

func TestValidatePhoneWithRegion(t *testing.T) {
	tests := []struct {
		name   string
		phone  string
		region string
		want   bool
	}{
		// 中国手机号
		{
			name:   "中国手机号-CN地区",
			phone:  "13800138000",
			region: "CN",
			want:   true,
		},
		{
			name:   "中国手机号-带+86-CN地区",
			phone:  "+8613800138000",
			region: "CN",
			want:   true,
		},
		{
			name:   "中国手机号-错误地区-US",
			phone:  "13800138000",
			region: "US",
			want:   false,
		},

		// 美国手机号
		{
			name:   "美国手机号-US地区",
			phone:  "6502530000",
			region: "US",
			want:   true,
		},
		{
			name:   "美国手机号-带+1-US地区",
			phone:  "+16502530000",
			region: "US",
			want:   true,
		},
		{
			name:   "美国手机号-错误地区-CN",
			phone:  "6502530000",
			region: "CN",
			want:   false,
		},

		// 英国手机号
		{
			name:   "英国手机号-GB地区",
			phone:  "7912345678",
			region: "GB",
			want:   true,
		},
		{
			name:   "英国手机号-带+44-GB地区",
			phone:  "+447912345678",
			region: "GB",
			want:   true,
		},

		// 日本手机号
		{
			name:   "日本手机号-JP地区",
			phone:  "08012345678",
			region: "JP",
			want:   true,
		},

		// 香港手机号
		{
			name:   "香港手机号-HK地区",
			phone:  "61234567",
			region: "HK",
			want:   true,
		},

		// 无效场景
		{
			name:   "无效-空手机号",
			phone:  "",
			region: "CN",
			want:   false,
		},
		{
			name:   "无效-空地区代码",
			phone:  "13800138000",
			region: "",
			want:   false,
		},
		{
			name:   "无效-错误地区代码",
			phone:  "13800138000",
			region: "XX",
			want:   false,
		},
		{
			name:   "无效-号码格式错误",
			phone:  "123",
			region: "CN",
			want:   false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := ValidatePhoneWithRegion(tt.phone, tt.region)
			if got != tt.want {
				t.Errorf("ValidatePhoneWithRegion(%q, %q) = %v, want %v", tt.phone, tt.region, got, tt.want)
			}
		})
	}
}

func TestValidatePhoneWithCountryCode(t *testing.T) {
	tests := []struct {
		name        string
		countryCode string
		phoneNum    string
		want        bool
	}{
		// 中国手机号
		{
			name:        "中国手机号-+86",
			countryCode: "+86",
			phoneNum:    "13800138000",
			want:        true,
		},
		{
			name:        "中国手机号-86不带加号",
			countryCode: "86",
			phoneNum:    "13800138000",
			want:        false, // phonenumbers 解析 "8613800138000" 时无法识别国家代码
		},
		{
			name:        "中国手机号-手机号太短",
			countryCode: "+86",
			phoneNum:    "13800138",
			want:        false,
		},
		{
			name:        "中国手机号-手机号太长",
			countryCode: "+86",
			phoneNum:    "13800138000123",
			want:        false,
		},

		// 美国手机号
		{
			name:        "美国手机号-+1",
			countryCode: "+1",
			phoneNum:    "6502530000",
			want:        true,
		},
		{
			name:        "美国手机号-1不带加号",
			countryCode: "1",
			phoneNum:    "6502530000",
			want:        false, // phonenumbers 解析 "16502530000" 时无法识别国家代码
		},

		// 英国手机号
		{
			name:        "英国手机号-+44",
			countryCode: "+44",
			phoneNum:    "7912345678",
			want:        true,
		},

		// 日本手机号
		{
			name:        "日本手机号-+81",
			countryCode: "+81",
			phoneNum:    "8012345678",
			want:        true,
		},

		// 香港手机号
		{
			name:        "香港手机号-+852",
			countryCode: "+852",
			phoneNum:    "61234567",
			want:        true,
		},

		// 国家代码与手机号不匹配
		{
			name:        "不匹配-美国国家代码配中国手机号",
			countryCode: "+1",
			phoneNum:    "13800138000",
			want:        false,
		},
		{
			name:        "不匹配-中国国家代码配美国手机号",
			countryCode: "+86",
			phoneNum:    "6502530000",
			want:        false,
		},

		// 无效场景
		{
			name:        "无效-空国家代码-空手机号",
			countryCode: "",
			phoneNum:    "",
			want:        false,
		},
		{
			name:        "无效-错误的国家代码",
			countryCode: "+999",
			phoneNum:    "13800138000",
			want:        false,
		},
		{
			name:        "无效-手机号包含字母",
			countryCode: "+86",
			phoneNum:    "13800ABCD",
			want:        false,
		},
		{
			name:        "没有国家代码",
			countryCode: "",
			phoneNum:    "15602727481",
			want:        false,
		},
		{
			name:        "无效-特殊字符",
			countryCode: "+86",
			phoneNum:    "138-0013-8000",
			want:        true, // phonenumbers 会忽略分隔符，实际能解析
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := ValidatePhoneWithCountryCode(tt.countryCode, tt.phoneNum)
			if got != tt.want {
				t.Errorf("ValidatePhoneWithCountryCode(%q, %q) = %v, want %v", tt.countryCode, tt.phoneNum, got, tt.want)
			}
		})
	}
}
