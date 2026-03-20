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

package resourceplan

import (
	"hcm/pkg/api/core"
	"hcm/pkg/criteria/validator"
	dtypes "hcm/pkg/dal/dao/types"
	tablegputpl "hcm/pkg/dal/table/resource-plan/res-plan-demand-gpu-template"
	"hcm/pkg/dal/table/types"
)

// DemandGpuTemplateBatchCreateReq batch create request for demand gpu template.
type DemandGpuTemplateBatchCreateReq struct {
	Templates []DemandGpuTemplateCreateReq `json:"templates" validate:"required,min=1,max=100"`
}

// Validate validate
func (r *DemandGpuTemplateBatchCreateReq) Validate() error {
	if err := validator.Validate.Struct(r); err != nil {
		return err
	}

	for _, c := range r.Templates {
		if err := c.Validate(); err != nil {
			return err
		}
	}

	return nil
}

// DemandGpuTemplateCreateReq create request for demand gpu template.
type DemandGpuTemplateCreateReq struct {
	TplSchema types.JsonField `json:"tpl_schema" validate:"required"`
	Remark    string          `json:"remark" validate:"max=255"`
	Creator   string          `json:"creator" validate:"required,max=64"`
}

// Validate validate
func (r *DemandGpuTemplateCreateReq) Validate() error {
	return validator.Validate.Struct(r)
}

// DemandGpuTemplateBatchUpdateReq batch update request for demand gpu template.
type DemandGpuTemplateBatchUpdateReq struct {
	Templates []DemandGpuTemplateUpdateReq `json:"templates" validate:"required,min=1,max=100"`
}

// Validate validate
func (r *DemandGpuTemplateBatchUpdateReq) Validate() error {
	if err := validator.Validate.Struct(r); err != nil {
		return err
	}

	for _, c := range r.Templates {
		if err := c.Validate(); err != nil {
			return err
		}
	}

	return nil
}

// DemandGpuTemplateUpdateReq update request for demand gpu template.
type DemandGpuTemplateUpdateReq struct {
	ID        string          `json:"id" validate:"required"`
	TplSchema types.JsonField `json:"tpl_schema"`
	Remark    *string         `json:"remark" validate:"omitempty,max=255"`
	Reviser   string          `json:"reviser" validate:"required,max=64"`
}

// Validate validate
func (r *DemandGpuTemplateUpdateReq) Validate() error {
	return validator.Validate.Struct(r)
}

// DemandGpuTemplateListResult list result for demand gpu template.
type DemandGpuTemplateListResult = dtypes.ListResult[tablegputpl.ResPlanDemandGpuTemplateTable]

// DemandGpuTemplateListReq list request for demand gpu template.
type DemandGpuTemplateListReq struct {
	core.ListReq `json:",inline"`
}

// Validate validate
func (r *DemandGpuTemplateListReq) Validate() error {
	return r.ListReq.Validate()
}
