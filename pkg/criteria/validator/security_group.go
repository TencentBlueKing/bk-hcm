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
	"fmt"
	"regexp"
	"strings"
)

const (
	qnameExtSGNameFmt  string = "[a-z0-9-]"
	qualifiedSGNameFmt        = "(" + lowEnglish + qnameExtSGNameFmt + "*)?" + lowEnglish
)

// qualifiedSGNameRegexp security group's name regexp.
var qualifiedSGNameRegexp = regexp.MustCompile("^" + qualifiedSGNameFmt + "$")

// ValidateSecurityGroupName validate security group name's length and format.
func ValidateSecurityGroupName(name string) error {
	if len(name) < 1 {
		return errors.New("invalid name, length should >= 1")
	}

	if len(name) > 60 {
		return errors.New("invalid name, length should <= 60")
	}

	if strings.HasPrefix(name, "sg-") {
		return errors.New("name can not start with 'sg-'")
	}

	if !qualifiedSGNameRegexp.MatchString(name) {
		return fmt.Errorf("invalid name: %s, only allows to include low english、numbers、hyphen (-), and must start "+
			"and end with an low english", name)
	}

	return nil
}

// ValidateSecurityGroupMemo validate security group memo's length and format.
func ValidateSecurityGroupMemo(memo *string) error {
	if memo == nil {
		return errors.New("memo is nil")
	}

	content := *memo
	if len(content) == 0 {
		return nil
	}

	if len(content) > 100 {
		return errors.New("invalid memo, length should <= 100")
	}

	if !qualifiedMemoRegexp.MatchString(content) {
		return errors.New("invalid memo, only allows include chinese、english、numbers、underscore (_)" +
			"、hyphen (-)、space, and must start and end with an chinese、english、numbers")
	}

	return nil
}
