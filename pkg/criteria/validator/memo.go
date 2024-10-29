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
	"errors"
	"regexp"
	"unicode/utf8"
)

const (
	// qualifiedMemoFmt hcm resource's memo format.
	qualifiedMemoFmt        = "(" + chineseEnglishNumberFmt + qExtMemoFmt + "*)?" + chineseEnglishNumberFmt
	qExtMemoFmt      string = "[\u4E00-\u9FA5A-Za-z0-9-_\\s]"
)

// qualifiedMemoRegexp hcm resource's memo regexp.
var qualifiedMemoRegexp = regexp.MustCompile("^" + qualifiedMemoFmt + "$")

// ValidateMemo validate hcm resource memo's length and format.
func ValidateMemo(memo *string, required bool) error {
	// check data is nil and required.
	if required && (memo == nil || len(*memo) == 0) {
		return errors.New("memo is required, can not be empty")
	}

	if memo == nil || len(*memo) == 0 {
		return nil
	}

	m := *memo
	if utf8.RuneCountInString(m) > 255 {
		return errors.New("invalid memo, length should less than 255")
	}

	// 只有非登记账号才校验 备注格式
	if !qualifiedMemoRegexp.MatchString(m) {
		return errors.New("invalid memo, only allows include chinese、english、numbers、underscore (_)" +
			"、hyphen (-)、space, and must start and end with an chinese、english、numbers")
	}

	return nil
}
