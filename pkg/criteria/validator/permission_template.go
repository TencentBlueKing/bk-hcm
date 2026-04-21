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
	"fmt"
	"regexp"
)

// nameRegexp 权限模版命名正则表达式，只能包含英文字母、数字和-_
var nameRegexp = regexp.MustCompile(`^[a-zA-Z0-9_-]+$`)

// ValidatePermTmplName 验证权限模板名称
func ValidatePermTmplName(name string) error {
	if !nameRegexp.MatchString(name) {
		return fmt.Errorf("invalid name: %s, only allows english letters, numbers, underscore (_) and hyphen (-)", name)
	}
	return nil
}
