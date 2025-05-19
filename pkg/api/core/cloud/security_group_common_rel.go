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

package cloud

import (
	"hcm/pkg/api/core/cloud/cvm"
	"hcm/pkg/api/core/cloud/load-balancer"
	"hcm/pkg/criteria/enumor"
)

// SecurityGroupCommonRel define security group common rel.
type SecurityGroupCommonRel struct {
	ID              uint64                   `json:"id"`
	ResVendor       enumor.Vendor            `json:"res_vendor"`
	ResID           string                   `json:"res_id"`
	ResType         enumor.CloudResourceType `json:"res_type"`
	Priority        int64                    `json:"priority"`
	SecurityGroupID string                   `json:"security_group_id"`
	Creator         string                   `json:"creator"`
	Reviser         string                   `json:"reviser"`
	CreatedAt       string                   `json:"created_at"`
	UpdatedAt       string                   `json:"updated_at"`
}

// SGCommonRelWithBaseSecurityGroup define security group with common id.
type SGCommonRelWithBaseSecurityGroup struct {
	BaseSecurityGroup `json:",inline"`
	ResID             string                   `json:"res_id"`
	ResType           enumor.CloudResourceType `json:"res_type"`
	Priority          int64                    `json:"priority"`
	RelCreator        string                   `db:"rel_creator" json:"rel_creator"`
	RelCreatedAt      string                   `db:"rel_created_at" json:"rel_created_at"`
}

// SGCommonRelWithCVMSummary define security group with cvm summary.
type SGCommonRelWithCVMSummary struct {
	cvm.SummaryCVM  `json:",inline"`
	SecurityGroupId string `db:"security_group_id" json:"security_group_id"`
	RelCreator      string `db:"rel_creator" json:"rel_creator"`
	RelCreatedAt    string `db:"rel_created_at" json:"rel_created_at"`
}

// SGCommonRelWithLBSummary define security group with lb summary.
type SGCommonRelWithLBSummary struct {
	loadbalancer.SummaryBalancer `json:",inline"`
	SecurityGroupId              string `db:"security_group_id" json:"security_group_id"`
	RelCreator                   string `db:"rel_creator" json:"rel_creator"`
	RelCreatedAt                 string `db:"rel_created_at" json:"rel_created_at"`
}
