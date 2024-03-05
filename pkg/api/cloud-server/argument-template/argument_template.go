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

package csargstpl

import (
	"errors"
	"fmt"

	"hcm/pkg/criteria/constant"
	"hcm/pkg/criteria/validator"
)

// AssignArgsTplToBizReq define assign argument template to biz req.
type AssignArgsTplToBizReq struct {
	BkBizID     int64    `json:"bk_biz_id" validate:"required"`
	TemplateIDs []string `json:"template_ids" validate:"required"`
}

// Validate assign argument template to biz request.
func (req *AssignArgsTplToBizReq) Validate() error {
	if err := validator.Validate.Struct(req); err != nil {
		return err
	}

	if req.BkBizID <= 0 {
		return errors.New("bk_biz_id should >= 0")
	}

	if len(req.TemplateIDs) == 0 {
		return errors.New("template ids is required")
	}

	if len(req.TemplateIDs) > constant.BatchOperationMaxLimit {
		return fmt.Errorf("template ids should <= %d", constant.BatchOperationMaxLimit)
	}

	return nil
}

// BindArgsTplInstanceRuleResp define bind argument template instance rule req.
type BindArgsTplInstanceRuleResp struct {
	ID          string `json:"id"`
	InstanceNum int64  `json:"instance_num"`
	RuleNum     int64  `json:"rule_num"`
}

// -------------------------- Delete --------------------------

// ArgsTplBatchIDsReq argument template batch ids request.
type ArgsTplBatchIDsReq struct {
	IDs []string `json:"ids" validate:"required,min=1"`
}

// Validate argument template batch ids validate.
func (req *ArgsTplBatchIDsReq) Validate() error {
	if len(req.IDs) > constant.BatchOperationMaxLimit {
		return fmt.Errorf("ids should <= %d", constant.BatchOperationMaxLimit)
	}

	return validator.Validate.Struct(req)
}
