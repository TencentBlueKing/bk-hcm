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
	"fmt"
	"time"

	"hcm/pkg/api/core"
	dataproto "hcm/pkg/api/data-service/cloud/disk"
	"hcm/pkg/criteria/constant"
	"hcm/pkg/criteria/validator"
	"hcm/pkg/rest"
	"hcm/pkg/runtime/filter"
)

// DiskCvmRelBatchCreateReq ...
type DiskCvmRelBatchCreateReq struct {
	Rels []DiskCvmRelCreateReq `json:"rels" validate:"required"`
}

// Validate ...
func (req *DiskCvmRelBatchCreateReq) Validate() error {
	if len(req.Rels) > constant.BatchOperationMaxLimit {
		return fmt.Errorf("rels count should <= %d", constant.BatchOperationMaxLimit)
	}

	return validator.Validate.Struct(req)
}

// DiskCvmRelCreateReq ...
type DiskCvmRelCreateReq struct {
	DiskID string `json:"disk_id" validate:"required"`
	CvmID  string `json:"cvm_id" validate:"required"`
}

// Validate ...
func (req *DiskCvmRelCreateReq) Validate() error {
	return validator.Validate.Struct(req)
}

// DiskCvmRelListReq ...
type DiskCvmRelListReq struct {
	Filter *filter.Expression `json:"filter" validate:"required"`
	Page   *core.BasePage     `json:"page" validate:"required"`
	Fields []string           `json:"fields" validate:"omitempty"`
}

// Validate ...
func (req *DiskCvmRelListReq) Validate() error {
	return validator.Validate.Struct(req)
}

// DiskCvmRelDeleteReq ...
type DiskCvmRelDeleteReq struct {
	Filter *filter.Expression `json:"filter" validate:"required"`
}

// Validate ...
func (req *DiskCvmRelDeleteReq) Validate() error {
	return validator.Validate.Struct(req)
}

// DiskCvmRelListResult ...
type DiskCvmRelListResult struct {
	Count   *uint64             `json:"count,omitempty"`
	Details []*DiskCvmRelResult `json:"details"`
}

// DiskCvmRelResult ...
type DiskCvmRelResult struct {
	ID        uint64     `json:"id,omitempty"`
	DiskID    string     `json:"disk_id,omitempty"`
	CvmID     string     `json:"cvm_id,omitempty"`
	Creator   string     `json:"creator,omitempty"`
	CreatedAt *time.Time `json:"created_at,omitempty"`
}

// DiskCvmRelListResp ...
type DiskCvmRelListResp struct {
	rest.BaseResp `json:",inline"`
	Data          *DiskCvmRelListResult `json:"data"`
}

// DiskCvmRelWithDiskListReq ...
type DiskCvmRelWithDiskListReq struct {
	CvmIDs []string `json:"cvm_ids" validate:"required"`
}

// Validate ....
func (req *DiskCvmRelWithDiskListReq) Validate() error {
	return validator.Validate.Struct(req)
}

// DiskCvmRelWithDiskListResp ...
type DiskCvmRelWithDiskListResp struct {
	rest.BaseResp `json:",inline"`
	Data          []*DiskWithCvmID `json:"data"`
}

// DiskWithCvmID ...
type DiskWithCvmID struct {
	dataproto.DiskResult `json:",inline"`
	CvmID                string     `json:"cvm_id"`
	RelCreator           string     `json:"rel_creator"`
	RelCreatedAt         *time.Time `json:"rel_created_at"`
}
