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

package account

import (
	"hcm/pkg/criteria/enumor"
)

var (
	vendorMainAccountIDFieldMap = map[enumor.Vendor]string{
		enumor.TCloud: "cloud_main_account_id",
		enumor.Aws:    "cloud_account_id",
		enumor.HuaWei: "cloud_sub_account_id",
		enumor.Gcp:    "cloud_project_id",
		enumor.Azure:  "cloud_tenant_id",
	}

	accountTypNameMap = map[enumor.AccountType]string{
		enumor.RegistrationAccount:  "登记账号",
		enumor.ResourceAccount:      "资源账号",
		enumor.SecurityAuditAccount: "安全审计账号",
	}

	accountSiteTypeNameMap = map[enumor.AccountSiteType]string{
		enumor.InternationalSite: "国际站",
		enumor.ChinaSite:         "中国站",
	}
)
