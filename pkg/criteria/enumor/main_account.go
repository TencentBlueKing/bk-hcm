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

var (
	MainAccountBusinessTypeNameMap = map[MainAccountBusinessType]string{
		InternationalBusiness: "国际业务",
		ChinaBusiness:         "中国业务",
	}
)

// MainAccountSiteType is site type of main account, should be same as its root account
type MainAccountSiteType = RootAccountSiteType

const (
	// MainAccountChinaSite is china site.
	MainAccountChinaSite = RootAccountChinaSite
	// MainAccountInternationalSite is international site.
	MainAccountInternationalSite = RootAccountInternationalSite
)

var (
	// MainAccountSiteTypeNameMap is the map of main account site type name
	MainAccountSiteTypeNameMap = map[MainAccountSiteType]string{
		MainAccountChinaSite:         "中国站",
		MainAccountInternationalSite: "国际站",
	}
)

// GetMainAccountSiteTypeName get the main account site type name
func (a MainAccountSiteType) GetMainAccountSiteTypeName() string {
	return MainAccountSiteTypeNameMap[a]
}

// MainAccountCommonFields is the common fields of main account
type MainAccountCommonFields struct {
	AccountName  string
	AccountID    string
	InitPassword string
}

// MainAccountNameFieldNameMap is the map of main account fields name, only use for main account management
var MainAccountNameFieldNameMap = map[Vendor]MainAccountCommonFields{
	Aws: {
		AccountName:  "cloud_main_account_name",
		AccountID:    "cloud_main_account_id",
		InitPassword: "cloud_init_password",
	},
	Gcp: {
		AccountName: "cloud_project_name",
		AccountID:   "cloud_project_id",
	},
	HuaWei: {
		AccountName:  "cloud_main_account_name",
		AccountID:    "cloud_main_account_id",
		InitPassword: "cloud_init_password",
	},
	Azure: {
		AccountName:  "cloud_subscription_name",
		AccountID:    "cloud_subscription_id",
		InitPassword: "cloud_init_password",
	},
	Zenlayer: {
		AccountName:  "cloud_main_account_name",
		AccountID:    "cloud_main_account_id",
		InitPassword: "cloud_init_password",
	},
	Kaopu: {
		AccountName:  "cloud_main_account_name",
		AccountID:    "cloud_main_account_id",
		InitPassword: "cloud_init_password",
	},
}

// GetMainAccountNameFieldName get the main account name field name
func (v Vendor) GetMainAccountNameFieldName() string {
	return MainAccountNameFieldNameMap[v].AccountName
}

// GetMainAccountIDFieldName get the main account id field name
func (v Vendor) GetMainAccountIDFieldName() string {
	return MainAccountNameFieldNameMap[v].AccountID
}

// GetMainAccountInitPasswordFieldName get the main account init password field name
func (v Vendor) GetMainAccountInitPasswordFieldName() string {
	return MainAccountNameFieldNameMap[v].InitPassword
}

// MainAccountStatus is main account status, 状态为后续功能预留
type MainAccountStatus string

const (
	// MainAccountStatusRUNNING is main account running status
	MainAccountStatusRUNNING MainAccountStatus = "RUNNING"
	// MainAccountStatusDELETED is main account deleted status
	MainAccountStatusDELETED MainAccountStatus = "DELETED"
	// MainAccountStatusSUSPEND is main account suspend status
	MainAccountStatusSUSPEND MainAccountStatus = "SUSPEND"
)

// Validate the AccountSiteType is valid or not
func (a MainAccountStatus) Validate() error {
	switch a {
	case MainAccountStatusRUNNING:
	case MainAccountStatusDELETED:
	case MainAccountStatusSUSPEND:
	default:
		return fmt.Errorf("unsupported main account status type: %s", a)

	}

	return nil
}
