/*
 * TencentBlueKing is pleased to support the open source community by making
 * 蓝鲸智云 - 混合云管理平台 (BlueKing - Hybrid Cloud Management System) available.
 * Copyright (C) 2024 THL A29 Limited,
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

// Package billitem ...
package billitem

import (
	"hcm/cmd/account-server/logics/audit"
	"hcm/cmd/account-server/service/capability"
	"hcm/pkg/client"
	"hcm/pkg/iam/auth"
	"hcm/pkg/rest"
)

// InitBillItemService 注册账单明细服务
func InitBillItemService(c *capability.Capability) {
	svc := &billItemSvc{
		client:     c.ApiClient,
		authorizer: c.Authorizer,
		audit:      c.Audit,
	}

	h := rest.NewHandler()

	h.Add("ListBillItems", "POST", "/vendors/{vendor}/bills/items/list", svc.ListBillItems)

	h.Add("ExportBillItems", "POST", "/vendors/{vendor}/bills/items/export", svc.ExportBillItems)
	h.Add("ImportBillItemsPreview", "POST",
		"/vendors/{vendor}/bills/items/import/preview", svc.ImportBillItemsPreview)
	h.Add("ImportBillItems",
		"POST", "/vendors/{vendor}/bills/items/import", svc.ImportBillItems)

	h.Add("PullBillItemForThirdParty", "POST",
		"/vendors/{vendor}/bills/items/pull", svc.PullBillItemForThirdParty)

	h.Load(c.WebService)
}

// 账单明细
type billItemSvc struct {
	client     *client.ClientSet
	authorizer auth.Authorizer
	audit      audit.Interface
}
