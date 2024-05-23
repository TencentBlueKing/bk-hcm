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

// Package audit ...
package audit

import (
	"hcm/pkg/api/data-service/cloud"
	"hcm/pkg/criteria/enumor"
	"hcm/pkg/dal/table/audit"
	"hcm/pkg/dal/table/cloud/load-balancer"
)

// Audit define audit.
type Audit struct {
	ID                   uint64                   `json:"id"`
	ResID                string                   `json:"res_id"`
	CloudResID           string                   `json:"cloud_res_id"`
	ResName              string                   `json:"res_name"`
	ResType              enumor.AuditResourceType `json:"res_type"`
	AssociatedResID      string                   `json:"associated_res_id"`
	AssociatedCloudResID string                   `json:"associated_cloud_res_id"`
	AssociatedResName    string                   `json:"associated_res_name"`
	AssociatedResType    string                   `json:"associated_res_type"`
	Action               enumor.AuditAction       `json:"action"`
	BkBizID              int64                    `json:"bk_biz_id"`
	Vendor               enumor.Vendor            `json:"vendor"`
	AccountID            string                   `json:"account_id"`
	Operator             string                   `json:"operator"`
	Source               enumor.RequestSourceType `json:"source"`
	Rid                  string                   `json:"rid"`
	AppCode              string                   `json:"app_code"`
	Detail               any                      `json:"detail,omitempty"` // Detail list接口该字段默认不返回
	CreatedAt            string                   `json:"created_at"`
}

// RawAudit define audit.
type RawAudit struct {
	ID                   uint64                   `json:"id"`
	ResID                string                   `json:"res_id"`
	CloudResID           string                   `json:"cloud_res_id"`
	ResName              string                   `json:"res_name"`
	ResType              enumor.AuditResourceType `json:"res_type"`
	AssociatedResID      string                   `json:"associated_res_id"`
	AssociatedCloudResID string                   `json:"associated_cloud_res_id"`
	AssociatedResName    string                   `json:"associated_res_name"`
	AssociatedResType    string                   `json:"associated_res_type"`
	Action               enumor.AuditAction       `json:"action"`
	BkBizID              int64                    `json:"bk_biz_id"`
	Vendor               enumor.Vendor            `json:"vendor"`
	AccountID            string                   `json:"account_id"`
	Operator             string                   `json:"operator"`
	Source               enumor.RequestSourceType `json:"source"`
	Rid                  string                   `json:"rid"`
	AppCode              string                   `json:"app_code"`
	Detail               *audit.BasicDetailRaw    `json:"detail,omitempty"` // Detail list接口该字段默认不返回
	CreatedAt            string                   `json:"created_at"`
}

// TargetGroupAsyncAuditDetail 目标组异步任务操作详情
type TargetGroupAsyncAuditDetail struct {
	LoadBalancer tablelb.LoadBalancerTable `json:"load_balancer"`
	ResFlow      *cloud.ResFlowLockReq     `json:"res_flow"`
}
