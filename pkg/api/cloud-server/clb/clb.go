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

package csclb

import (
	"errors"
	"fmt"

	"hcm/pkg/criteria/constant"
	"hcm/pkg/criteria/validator"
)

// BatchBindClbSecurityGroupReq batch bind clb security group req.
type BatchBindClbSecurityGroupReq struct {
	ClbID            string   `json:"clb_id" validate:"required"`
	SecurityGroupIDs []string `json:"security_group_ids" validate:"required,max=50"`
}

// Validate validate.
func (req *BatchBindClbSecurityGroupReq) Validate() error {
	if len(req.ClbID) == 0 {
		return errors.New("clb_id is required")
	}

	if len(req.SecurityGroupIDs) == 0 {
		return errors.New("security_group_ids is required")
	}

	if len(req.SecurityGroupIDs) > constant.LoadBalancerBindSecurityGroupMaxLimit {
		return fmt.Errorf("security_group_ids should <= %d", constant.LoadBalancerBindSecurityGroupMaxLimit)
	}

	return validator.Validate.Struct(req)
}

// UnBindClbSecurityGroupReq unbind clb security group req.
type UnBindClbSecurityGroupReq struct {
	ClbID           string `json:"clb_id" validate:"required"`
	SecurityGroupID string `json:"security_group_id" validate:"required"`
}

// Validate validate.
func (req *UnBindClbSecurityGroupReq) Validate() error {
	if len(req.ClbID) == 0 {
		return errors.New("clb_id is required")
	}

	if len(req.SecurityGroupID) == 0 {
		return errors.New("security_group_id is required")
	}

	return validator.Validate.Struct(req)
}

// -------------------------- Associate --------------------------

// SecurityGroupAssociateClbReq define security group associate clb option.
type SecurityGroupAssociateClbReq struct {
	ClbID            string   `json:"clb_id" validate:"required"`
	SecurityGroupIDs []string `json:"security_group_ids" validate:"required,max=50"`
}

// Validate security group associate clb request.
func (req *SecurityGroupAssociateClbReq) Validate() error {
	return validator.Validate.Struct(req)
}

// AssignClbToBizReq define assign clb to biz req.
type AssignClbToBizReq struct {
	BkBizID int64    `json:"bk_biz_id" validate:"required,min=0"`
	ClbIDs  []string `json:"clb_ids" validate:"required,min=1"`
}

// Validate assign clb to biz request.
func (req *AssignClbToBizReq) Validate() error {

	if req.BkBizID <= 0 {
		return errors.New("bk_biz_id should >= 0")
	}

	if len(req.ClbIDs) == 0 {
		return errors.New("clb ids is required")
	}

	if len(req.ClbIDs) > constant.BatchOperationMaxLimit {
		return fmt.Errorf("clb ids should <= %d", constant.BatchOperationMaxLimit)
	}

	return validator.Validate.Struct(req)
}
