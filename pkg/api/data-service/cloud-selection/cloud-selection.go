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

package dsselection

import (
	"errors"

	"hcm/pkg/criteria/enumor"
	"hcm/pkg/criteria/validator"
	"hcm/pkg/dal/table/types"
)

// SchemeUpdateReq define scheme update request.
type SchemeUpdateReq struct {
	Name    string `json:"name" validate:"omitempty"`
	BkBizID int64  `json:"bk_biz_id" validate:"omitempty"`
}

// Validate SchemeUpdateReq.
func (req SchemeUpdateReq) Validate() error {
	if len(req.Name) == 0 && req.BkBizID == 0 {
		return errors.New("not found update field")
	}

	return validator.Validate.Struct(req)
}

// SchemeCreateReq define scheme create request.
type SchemeCreateReq struct {
	BkBizID                int64                     `json:"bk_biz_id" validate:"required"`
	Name                   string                    `json:"name" validate:"required,max=255"`
	BizType                string                    `json:"biz_type" validate:"max=64"`
	Vendors                types.StringArray         `json:"vendors" validate:"required"`
	DeploymentArchitecture []enumor.SchemeDeployArch `json:"deployment_architecture" validate:"required"`
	CoverPing              float64                   `json:"cover_ping" validate:"required"`
	CompositeScore         float64                   `json:"composite_score" validate:"required"`
	NetScore               float64                   `json:"net_score" validate:"required"`
	CostScore              float64                   `json:"cost_score" validate:"required"`
	CoverRate              float64                   `json:"cover_rate" validate:"min=0"`
	UserDistribution       types.AreaInfos           `json:"user_distribution" validate:"required"`
	ResultIdcIDs           types.StringArray         `json:"result_idc_ids" validate:"required"`
}

// Validate SchemeCreateReq.
func (req SchemeCreateReq) Validate() error {
	return validator.Validate.Struct(req)
}
