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

	"hcm/pkg/criteria/enumor"
)

var gcpNameRegexp = regexp.MustCompile(`^([a-z0-9][a-z0-9-]*)[a-z0-9]$`)
var huaweiCvmNameRegexp = qualifiedNameRegexp
var azureCvmNameRegexp = regexp.MustCompile(`^([A-Za-z0-9][A-Za-z0-9-.]*)[A-Za-z0-9]$`)

// ValidateCvmName validate cvm name.
// vendor: 	length		regexp
// TCloud:  60
// Aws:		256
// HuaWei:  64			中文字符、英文字母、数字及“_”、“-”组成
// Gcp:		62			小写字母 (a-z)、数字和连字符组成
// Azure:				字母、数字、"."和"-"
func ValidateCvmName(vendor enumor.Vendor, name string) error {
	switch vendor {
	case enumor.TCloud:
		if len(name) == 0 || len(name) > 60 {
			return errors.New("aws cvm name should 1-60")
		}

	case enumor.Aws:
		if len(name) == 0 || len(name) > 256 {
			return errors.New("aws cvm name should 1-256")
		}

	case enumor.HuaWei:
		if len(name) == 0 || len(name) > 64 {
			return errors.New("aws cvm name should 1-64")
		}

		if !huaweiCvmNameRegexp.MatchString(name) {
			return fmt.Errorf("invalid name: %s, only allows to include chinese、english、numbers、underscore (_)"+
				"、hyphen (-), and must start and end with an chinese、english、numbers", name)
		}

	case enumor.Gcp:
		if len(name) == 0 || len(name) > 62 {
			return errors.New("gcp cvm name should 1-62")
		}

		if !gcpNameRegexp.MatchString(name) {
			return fmt.Errorf("invalid name: %s, only allows include lowercase english、numbers、hyphen (-), and must "+
				"start and end with an english、numbers", name)
		}

	case enumor.Azure:
		if len(name) == 0 {
			return errors.New("aws cvm name should > 0")
		}

		if !azureCvmNameRegexp.MatchString(name) {
			return fmt.Errorf("invalid name: %s, only allows include english、numbers、hyphen (-)、point (.), and must "+
				"start and end with an english、numbers", name)
		}

	default:
		return fmt.Errorf("vendor %s not support", vendor)
	}

	return nil
}
