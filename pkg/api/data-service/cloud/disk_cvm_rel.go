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
	"hcm/pkg/api/core/cloud/cvm"
	diskcvmrel "hcm/pkg/api/core/cloud/disk-cvm-rel"
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

// DiskCvmRelWithDiskExtListReq ...
type DiskCvmRelWithDiskExtListReq struct {
	CvmIDs []string `json:"cvm_ids" validate:"required"`
}

// Validate ....
func (req *DiskCvmRelWithDiskExtListReq) Validate() error {
	return validator.Validate.Struct(req)
}

// DiskCvmRelWithDiskExtListResp ...
type DiskCvmRelWithDiskExtListResp[T dataproto.DiskExtensionResult] struct {
	rest.BaseResp `json:",inline"`
	Data          []*DiskExtWithCvmID[T] `json:"data"`
}

// DiskExtWithCvmID ...
type DiskExtWithCvmID[T dataproto.DiskExtensionResult] struct {
	dataproto.DiskExtResult[T] `json:",inline"`
	CvmID                      string     `json:"cvm_id"`
	RelCreator                 string     `json:"rel_creator"`
	RelCreatedAt               *time.Time `json:"rel_created_at"`
}

// ListWithCvmReq ...
type ListWithCvmReq struct {
	Fields []string           `json:"fields" validate:"omitempty"`
	Filter *filter.Expression `json:"filter" validate:"required"`
	Page   *core.BasePage     `json:"page" validate:"required"`
	// NotEqualDiskID 关联关系表中 disk_id 不等于该值的查询条件。
	NotEqualDiskID string `json:"not_equal_disk_id" validate:"omitempty"`
}

// Validate ...
func (req *ListWithCvmReq) Validate() error {
	return validator.Validate.Struct(req)
}

// ListWithCvmResp ...
type ListWithCvmResp struct {
	rest.BaseResp `json:",inline"`
	Data          *ListCvmResult `json:"data"`
}

// ListCvmResult ...
type ListCvmResult struct {
	Count   uint64        `json:"count"`
	Details []cvm.BaseCvm `json:"details"`
}

// ListDiskWithoutCvmReq ...
type ListDiskWithoutCvmReq struct {
	Fields []string           `json:"fields" validate:"omitempty"`
	Filter *filter.Expression `json:"filter" validate:"omitempty"`
	Page   *core.BasePage     `json:"page" validate:"required"`
}

// Validate ...
func (req *ListDiskWithoutCvmReq) Validate() error {
	return validator.Validate.Struct(req)
}

// ListDiskWithoutCvmResult ...
type ListDiskWithoutCvmResult struct {
	Count   uint64                   `json:"count"`
	Details []diskcvmrel.RelWithDisk `json:"details"`
}

// ListDiskWithoutCvmResp ...
type ListDiskWithoutCvmResp struct {
	rest.BaseResp `json:",inline"`
	Data          *ListDiskWithoutCvmResult `json:"data"`
}
