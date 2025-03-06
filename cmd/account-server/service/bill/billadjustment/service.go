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

// Package billadjustment ...
package billadjustment

import (
	"hcm/cmd/account-server/logics/audit"
	"hcm/cmd/account-server/service/capability"
	"hcm/pkg/client"
	"hcm/pkg/iam/auth"
	"hcm/pkg/rest"
	"hcm/pkg/thirdparty/api-gateway/cmdb"
)

// InitBillAdjustmentService 注册账单调调整服务
func InitBillAdjustmentService(c *capability.Capability) {
	svc := &billAdjustmentSvc{
		client:     c.ApiClient,
		authorizer: c.Authorizer,
		audit:      c.Audit,
		cmdbCli:    c.CmdbClient,
	}

	h := rest.NewHandler()

	h.Add("ListBillAdjustmentItem", "POST",
		"/bills/adjustment_items/list", svc.ListBillAdjustmentItem)
	h.Add("CreateBillAdjustmentItem", "POST",
		"/bills/adjustment_items/create", svc.CreateBillAdjustmentItem)
	h.Add("ImportBillAdjustments", "POST",
		"/bills/adjustment_items/import", svc.ImportBillAdjustment)
	h.Add("UpdateBillAdjustmentItem", "PATCH",
		"/bills/adjustment_items/{id}", svc.UpdateBillAdjustmentItem)
	h.Add("DeleteBillAdjustmentItem", "DELETE",
		"/bills/adjustment_items/{id}", svc.DeleteBillAdjustmentItem)
	h.Add("BatchDeleteBillAdjustmentItem", "DELETE",
		"/bills/adjustment_items/batch", svc.BatchDeleteBillAdjustmentItem)
	h.Add("BatchConfirmBillAdjustmentItem", "POST",
		"/bills/adjustment_items/confirm", svc.BatchConfirmBillAdjustmentItem)
	h.Add("SumBillAdjustmentItem", "POST",
		"/bills/adjustment_items/sum", svc.SumBillAdjustmentItem)
	h.Add("ListBillAdjustmentItem", "POST",
		"/bills/adjustment_items/export", svc.ExportBillAdjustmentItem)

	h.Load(c.WebService)
}

// 账单明细
type billAdjustmentSvc struct {
	client     *client.ClientSet
	authorizer auth.Authorizer
	audit      audit.Interface
	cmdbCli    cmdb.Client
}
