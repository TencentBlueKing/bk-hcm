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

package csdisk

import (
	"errors"

	"hcm/pkg/criteria/constant"
	"hcm/pkg/criteria/errf"
	"hcm/pkg/criteria/validator"
	"hcm/pkg/tools/assert"
)

// AzureDiskAttachReq ...
type AzureDiskAttachReq struct {
	DiskID      string `json:"disk_id" validate:"required"`
	CvmID       string `json:"cvm_id" validate:"required"`
	CachingType string `json:"caching_type" validate:"required,eq=None|eq=ReadOnly|eq=ReadWrite"`
}

// Validate ...
func (req *AzureDiskAttachReq) Validate() error {
	return validator.Validate.Struct(req)
}

// AzureDiskCreateReq ...
type AzureDiskCreateReq struct {
	AccountID         string  `json:"account_id" validate:"required"`
	BkBizID           int64   `json:"bk_biz_id" validate:"omitempty"`
	DiskName          string  `json:"disk_name" validate:"required,lowercase"`
	ResourceGroupName string  `json:"resource_group_name" validate:"required,lowercase"`
	Region            string  `json:"region" validate:"required,lowercase"`
	Zone              string  `json:"zone" validate:"required,lowercase"`
	DiskType          string  `json:"disk_type" validate:"required"`
	DiskSize          int32   `json:"disk_size" validate:"required"`
	DiskCount         int32   `json:"disk_count" validate:"required"`
	Memo              *string `json:"memo" validate:"omitempty"`
}

// Validate ...
func (req *AzureDiskCreateReq) Validate(bizRequired bool) error {
	if err := validator.Validate.Struct(req); err != nil {
		return err
	}

	if bizRequired && req.BkBizID == 0 {
		return errors.New("bk_biz_id is required")
	}

	if req.DiskCount > constant.BatchOperationMaxLimit {
		return errors.New("disk count should <= 100")
	}

	// region can be no space lowercase
	if !assert.IsSameCaseNoSpaceString(req.Region) {
		return errf.New(errf.InvalidParameter, "region can only be lowercase")
	}

	// zone can be no space lowercase
	if !assert.IsSameCaseNoSpaceString(req.Zone) {
		return errf.New(errf.InvalidParameter, "zone can only be lowercase")
	}

	return nil
}
