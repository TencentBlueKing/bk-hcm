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

import (
	"github.com/nyaruka/phonenumbers"
)

// ValidatePhone 校验手机号是否有效（自动识别国家代码）
// 参数 phone 必须是带国际冠码(+号)的完整号码，如 "+8613800138000"
// 注意：不支持空格分隔的格式（如 "86 13800138000"），国家代码前必须带"+"号
// 返回 true 表示手机号有效，false 表示无效
func ValidatePhone(phone string) bool {
	num, err := phonenumbers.Parse(phone, "")
	if err != nil {
		return false
	}
	return phonenumbers.IsValidNumber(num)
}

// ValidatePhoneWithRegion 校验手机号是否对指定国家/地区有效
// phone: 手机号，可以带国家代码，如 "+8613800138000" 或 "13800138000"
// region: ISO 国家代码，如 "CN", "US", "JP" 等
// 返回 true 表示手机号对该地区有效，false 表示无效
func ValidatePhoneWithRegion(phone string, region string) bool {
	num, err := phonenumbers.Parse(phone, region)
	if err != nil {
		return false
	}
	return phonenumbers.IsValidNumberForRegion(num, region)
}

// ValidatePhoneWithCountryCode 校验国家代码和手机号是否匹配
// countryCode: 国家代码，必须以"+"开头，如 "+86"、"+1" 等（不支持"86"、"1"等不带+号的格式）
// phoneNum: 手机号，如 "13800138000"
// 返回 true 表示国家代码和手机号匹配且有效，false 表示不匹配或无效
// 注意：该函数会将 countryCode 和 phoneNum 拼接后解析，要求 countryCode 必须带"+"号
func ValidatePhoneWithCountryCode(countryCode, phoneNum string) bool {
	fullPhone := countryCode + phoneNum
	num, err := phonenumbers.Parse(fullPhone, "")
	if err != nil {
		return false
	}
	// 获取号码对应的国家代码
	region := phonenumbers.GetRegionCodeForNumber(num)
	if region == "" {
		return false
	}
	return phonenumbers.IsValidNumberForRegion(num, region)
}
