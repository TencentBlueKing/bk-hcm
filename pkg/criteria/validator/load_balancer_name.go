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
)

const (
	// loadBalancerNameExtFmt CLB名称中间部分字符类：支持中文、英文、数字、连字符(-)
	loadBalancerNameExtFmt string = "[\u4E00-\u9FA5A-Za-z0-9-]"
	// loadBalancerNameFmt CLB名称格式：支持中文、英文、数字、连字符(-)，且必须以中文、英文或数字开头和结尾
	loadBalancerNameFmt = chineseEnglishNumberFmt + "(" + loadBalancerNameExtFmt + "*" + chineseEnglishNumberFmt + ")?"
)

// loadBalancerNameRegexp CLB名称正则表达式
var loadBalancerNameRegexp = regexp.MustCompile("^" + loadBalancerNameFmt + "$")

// ValidateLoadBalancerName validate load balancer name.
func ValidateLoadBalancerName(name string) error {
	if len(name) == 0 || len(name) > 60 {
		return errors.New("load balancer name should be 1-60 characters")
	}

	if !loadBalancerNameRegexp.MatchString(name) {
		return fmt.Errorf("invalid name: %s, only allows to include chinese、english、numbers、hyphen (-), "+
			"and must start and end with chinese、english or numbers", name)
	}

	return nil
}
