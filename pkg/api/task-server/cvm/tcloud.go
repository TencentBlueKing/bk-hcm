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

// Package cvm ...
package cvm

import (
	"fmt"

	protocvm "hcm/pkg/api/hc-service/cvm"
	"hcm/pkg/criteria/enumor"
	"hcm/pkg/criteria/errf"
	"hcm/pkg/criteria/validator"
)

// BatchTaskCvmResetOption ...
type BatchTaskCvmResetOption struct {
	Vendor enumor.Vendor `json:"vendor" validate:"required"`
	// ManagementDetailIDs 对应的详情行id列表，需要和批量绑定的Targets参数长度对应
	ManagementDetailIDs []string                           `json:"management_detail_ids" validate:"required,min=1"`
	CvmResetList        []*protocvm.TCloudBatchResetCvmReq `json:"cvm_reset_list" validate:"required,min=1,dive"`
}

// Validate validate option.
func (opt BatchTaskCvmResetOption) Validate() error {
	switch opt.Vendor {
	case enumor.TCloud:
	default:
		return fmt.Errorf("unsupport vendor for batch reset cvm: %s", opt.Vendor)
	}

	if opt.CvmResetList == nil {
		return errf.New(errf.InvalidParameter, "cvm_reset_list is required")
	}
	if len(opt.ManagementDetailIDs) != len(opt.CvmResetList) {
		return errf.Newf(errf.InvalidParameter, "management_detail_ids and cvm_reset_list num not match, %d != %d",
			len(opt.ManagementDetailIDs), len(opt.CvmResetList))
	}
	return validator.Validate.Struct(opt)
}

// CvmOperationOption operation cvm option.
type CvmOperationOption struct {
	Vendor    enumor.Vendor `json:"vendor" validate:"required"`
	AccountID string        `json:"account_id" validate:"required"`
	Region    string        `json:"region" validate:"omitempty"`
	// IDs TCloud/HuaWei/Aws 支持批量操作，Azure/Gcp 仅支持单个操作
	IDs                 []string `json:"ids" validate:"required,min=1,max=100"`
	ManagementDetailIDs []string `json:"management_detail_ids" validate:"required,min=1,max=100"`
}

// Validate operation cvm option.
func (opt CvmOperationOption) Validate() error {

	switch opt.Vendor {
	case enumor.TCloud:
		if len(opt.Region) == 0 {
			return fmt.Errorf("vendor: %s region is required", opt.Vendor)
		}
	default:
		return fmt.Errorf("cvm operation option unsupported vendor: %s", opt.Vendor)
	}
	if len(opt.ManagementDetailIDs) != len(opt.IDs) {
		return errf.Newf(errf.InvalidParameter, "management_detail_ids and IDs length not match: %d! = %d",
			len(opt.ManagementDetailIDs), len(opt.IDs))
	}

	return validator.Validate.Struct(opt)
}
