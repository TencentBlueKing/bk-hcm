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

// Package dataorgtopo org topo data service
package dataorgtopo

import (
	"hcm/pkg/api/core"
	"hcm/pkg/criteria/errf"
	"hcm/pkg/criteria/validator"
	table "hcm/pkg/dal/table/org-topo"
)

// ListReq ...
type ListReq struct {
	core.ListReq `json:",inline"`
}

// Validate ListReq
func (req *ListReq) Validate() error {
	if err := validator.Validate.Struct(req); err != nil {
		return err
	}

	if err := req.ListReq.Validate(); err != nil {
		return err
	}

	return nil
}

// ListResp ...
type ListResp core.ListResultT[table.OrgTopo]

// ListByDeptIDsReq ...
type ListByDeptIDsReq struct {
	DeptIDs []string `json:"dept_ids" validate:"required,min=1"`
}

// Validate ListReq
func (req *ListByDeptIDsReq) Validate() error {
	if err := validator.Validate.Struct(req); err != nil {
		return err
	}
	return nil
}

// BatchCreateOrgTopoReq batch create request
type BatchCreateOrgTopoReq struct {
	OrgTopos []table.OrgTopo `json:"org_topos" validate:"required,max=500"`
}

// Validate ...
func (c *BatchCreateOrgTopoReq) Validate() error {
	if len(c.OrgTopos) == 0 || len(c.OrgTopos) > 500 {
		return errf.Newf(errf.InvalidParameter, "org_topos count should between 1 and 500")
	}
	return validator.Validate.Struct(c)
}

// BatchUpdateOrgTopoReq batch update request
type BatchUpdateOrgTopoReq struct {
	OrgTopos []table.OrgTopo `json:"org_topos" validate:"required"`
}

// Validate ...
func (c *BatchUpdateOrgTopoReq) Validate() error {
	if len(c.OrgTopos) == 0 || len(c.OrgTopos) > 500 {
		return errf.Newf(errf.InvalidParameter, "org_topos count should between 1 and 500")
	}
	return validator.Validate.Struct(c)
}

// BatchDeleteOrgTopoReq ...
type BatchDeleteOrgTopoReq struct {
	core.BatchDeleteReq `json:",inline"`
}

// Validate BatchCreateReq
func (req *BatchDeleteOrgTopoReq) Validate() error {
	if err := validator.Validate.Struct(req); err != nil {
		return err
	}

	if err := req.BatchDeleteReq.Validate(); err != nil {
		return err
	}

	return nil
}

// BatchUpsertOrgTopoReq batch upsert request
type BatchUpsertOrgTopoReq struct {
	AddOrgTopos    []table.OrgTopo `json:"add_org_topos" validate:"omitempty"`
	UpdateOrgTopos []table.OrgTopo `json:"update_org_topos" validate:"omitempty"`
}

// Validate ...
func (c *BatchUpsertOrgTopoReq) Validate() error {
	if len(c.AddOrgTopos) == 0 && len(c.UpdateOrgTopos) == 0 {
		return errf.Newf(errf.InvalidParameter, "add_org_topos and update_org_topos cannot be empty at the same time")
	}
	return validator.Validate.Struct(c)
}
