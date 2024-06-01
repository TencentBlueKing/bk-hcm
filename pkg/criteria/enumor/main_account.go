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

package enumor

import "fmt"

// MainAccountBusinessType is main account type
type MainAccountBusinessType string

// Validate the main account type
func (a MainAccountBusinessType) Validate() error {
	switch a {
	case InternationalBusiness:
	case ChinaBusiness:
	default:
		return fmt.Errorf("invalid main account type: %s", a)
	}
	return nil
}

const (
	InternationalBusiness MainAccountBusinessType = "international"
	ChinaBusiness         MainAccountBusinessType = "china"
)

// AccountSiteType is site type.
type MainAccountSiteType string

// Validate the AccountSiteType is valid or not
func (a MainAccountSiteType) Validate() error {
	switch a {
	case MainAccountChinaSite:
	case MainAccountInternationalSite:
	default:
		return fmt.Errorf("unsupported main account site type: %s", a)

	}

	return nil
}

const (
	// MainAccountChinaSite is china site.
	MainAccountChinaSite MainAccountSiteType = "china"
	// MainAccountInternationalSite is international site.
	MainAccountInternationalSite MainAccountSiteType = "international"
)
