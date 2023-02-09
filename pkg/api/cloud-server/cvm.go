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

package cloudserver

import (
	"errors"
	"fmt"

	"hcm/pkg/criteria/constant"
	"hcm/pkg/criteria/validator"
)

// AssignCvmToBizReq define assign cvm to biz req.
type AssignCvmToBizReq struct {
	BkBizID int64    `json:"bk_biz_id" validate:"required"`
	CvmIDs  []string `json:"cvm_ids" validate:"required"`
}

// Validate assign cvm to biz request.
func (req *AssignCvmToBizReq) Validate() error {
	if err := validator.Validate.Struct(req); err != nil {
		return err
	}

	if req.BkBizID <= 0 {
		return errors.New("bk_biz_id should >= 0")
	}

	if len(req.CvmIDs) == 0 {
		return errors.New("cvm ids is required")
	}

	if len(req.CvmIDs) > constant.BatchOperationMaxLimit {
		return fmt.Errorf("cvm ids should <= %d", constant.BatchOperationMaxLimit)
	}

	return nil
}

// BatchStartCvmReq batch start cvm req.
type BatchStartCvmReq struct {
	IDs []string `json:"ids" validate:"required"`
}

// Validate batch start cvm request.
func (req *BatchStartCvmReq) Validate() error {
	if len(req.IDs) > constant.BatchOperationMaxLimit {
		return fmt.Errorf("cvm ids should <= %d", constant.BatchOperationMaxLimit)
	}

	return validator.Validate.Struct(req)
}

// BatchStopCvmReq batch stop cvm req.
type BatchStopCvmReq struct {
	IDs []string `json:"ids" validate:"required"`
}

// Validate batch stop cvm request.
func (req *BatchStopCvmReq) Validate() error {
	if len(req.IDs) > constant.BatchOperationMaxLimit {
		return fmt.Errorf("cvm ids should <= %d", constant.BatchOperationMaxLimit)
	}

	return validator.Validate.Struct(req)
}

// BatchRebootCvmReq batch reboot cvm req.
type BatchRebootCvmReq struct {
	IDs []string `json:"ids" validate:"required"`
}

// Validate batch reboot cvm request.
func (req *BatchRebootCvmReq) Validate() error {
	if len(req.IDs) > constant.BatchOperationMaxLimit {
		return fmt.Errorf("cvm ids should <= %d", constant.BatchOperationMaxLimit)
	}

	return validator.Validate.Struct(req)
}
