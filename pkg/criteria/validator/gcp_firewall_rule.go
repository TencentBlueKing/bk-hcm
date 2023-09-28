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
)

// ValidateGcpName 长度63，名称必须以小写字母开头，后面最多可跟 62 个小写字母、数字或连字符，但不能以连字符结尾
func ValidateGcpName(name string) error {
	if len(name) == 0 || len(name) > 63 {
		return errors.New("gcp firewall rule name should 1-63")
	}

	if !gcpNameRegexp.MatchString(name) {
		return fmt.Errorf("invalid name: %s, only allows include lowercase english、numbers、hyphen (-), and must "+
			"start and end with an english、numbers", name)
	}

	return nil
}
