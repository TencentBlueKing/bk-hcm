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

package disk

import (
	"fmt"

	"hcm/pkg/criteria/constant"
	"hcm/pkg/criteria/validator"
	"hcm/pkg/rest"
)

// DiskBaseCreateReq 云盘基础请求数据
type DiskBaseCreateReq struct {
	AccountID string  `json:"account_id" validate:"required"`
	DiskName  *string `json:"disk_name"`
	Region    string  `json:"region" validate:"required"`
	Zone      string  `json:"zone" validate:"required"`
	DiskSize  uint64  `json:"disk_size" validate:"required"`
	DiskType  string  `json:"disk_type" validate:"required"`
	DiskCount uint32  `json:"disk_count" validate:"required"`
	Memo      *string `json:"memo"`
}

// DiskSyncReq disk sync request
type DiskSyncReq struct {
	AccountID         string   `json:"account_id" validate:"required"`
	Region            string   `json:"region" validate:"omitempty"`
	ResourceGroupName string   `json:"resource_group_name" validate:"omitempty"`
	Zone              string   `json:"zone" validate:"omitempty"`
	CloudIDs          []string `json:"cloud_ids" validate:"omitempty"`
	SelfLinks         []string `json:"self_links" validate:"omitempty"`
}

// Validate disk sync request.
func (req *DiskSyncReq) Validate() error {
	if len(req.CloudIDs) > constant.BatchOperationMaxLimit {
		return fmt.Errorf("operate sync count should <= %d", constant.BatchOperationMaxLimit)
	}

	return validator.Validate.Struct(req)
}

// DiskDeleteReq ...
type DiskDeleteReq struct {
	DiskID string `json:"disk_id" validate:"required"`
}

// Validate ...
func (req *DiskDeleteReq) Validate() error {
	return validator.Validate.Struct(req)
}

// DiskDetachReq ...
type DiskDetachReq struct {
	CvmID  string `json:"cvm_id" validate:"required"`
	DiskID string `json:"disk_id" validate:"required"`
}

// Validate ...
func (req *DiskDetachReq) Validate() error {
	return validator.Validate.Struct(req)
}

// BatchCreateResult ...
type BatchCreateResult struct {
	UnknownCloudIDs []string `json:"unknown_cloud_ids"`
	SuccessCloudIDs []string `json:"success_cloud_ids"`
	FailedCloudIDs  []string `json:"failed_cloud_ids"`
	FailedMessage   string   `json:"failed_message"`
}

// BatchCreateResp ...
type BatchCreateResp struct {
	rest.BaseResp `json:",inline"`
	Data          *BatchCreateResult `json:"data"`
}
