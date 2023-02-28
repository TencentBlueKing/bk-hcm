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
	dataproto "hcm/pkg/api/data-service/cloud/eip"
	"hcm/pkg/criteria/constant"
	"hcm/pkg/criteria/validator"
	"hcm/pkg/rest"
	"hcm/pkg/runtime/filter"
)

// EipCvmRelBatchCreateReq ...
type EipCvmRelBatchCreateReq struct {
	Rels []EipCvmRelCreateReq `json:"rels" validate:"required"`
}

// Validate ...
func (req *EipCvmRelBatchCreateReq) Validate() error {
	if len(req.Rels) > constant.BatchOperationMaxLimit {
		return fmt.Errorf("rels count should <= %d", constant.BatchOperationMaxLimit)
	}

	return validator.Validate.Struct(req)
}

// EipCvmRelCreateReq ...
type EipCvmRelCreateReq struct {
	EipID string `json:"eip_id" validate:"required"`
	CvmID string `json:"cvm_id" validate:"required"`
}

// EipCvmRelListReq ...
type EipCvmRelListReq struct {
	Filter *filter.Expression `json:"filter" validate:"required"`
	Page   *core.BasePage     `json:"page" validate:"required"`
	Fields []string           `json:"fields" validate:"omitempty"`
}

// Validate ...
func (req *EipCvmRelListReq) Validate() error {
	return validator.Validate.Struct(req)
}

// EipCvmRelDeleteReq ...
type EipCvmRelDeleteReq struct {
	Filter *filter.Expression `json:"filter" validate:"required"`
}

// Validate ...
func (req *EipCvmRelDeleteReq) Validate() error {
	return validator.Validate.Struct(req)
}

// EipCvmRelListResult ...
type EipCvmRelListResult struct {
	Count   *uint64            `json:"count,omitempty"`
	Details []*EipCvmRelResult `json:"details"`
}

// EipCvmRelResult ...
type EipCvmRelResult struct {
	ID        uint64     `json:"id,omitempty"`
	EipID     string     `json:"eip_id,omitempty"`
	CvmID     string     `json:"cvm_id,omitempty"`
	Creator   string     `json:"creator,omitempty"`
	CreatedAt *time.Time `json:"created_at,omitempty"`
}

// EipCvmRelListResp ...
type EipCvmRelListResp struct {
	rest.BaseResp `json:",inline"`
	Data          *EipCvmRelListResult `json:"data"`
}

// EipCvmRelWithEipListReq ...
type EipCvmRelWithEipListReq struct {
	CvmIDs []string `json:"cvm_ids" validate:"required"`
}

// Validate ....
func (req *EipCvmRelWithEipListReq) Validate() error {
	return validator.Validate.Struct(req)
}

// EipCvmRelWithEipListResp ...
type EipCvmRelWithEipListResp struct {
	rest.BaseResp `json:",inline"`
	Data          []*EipWithCvmID `json:"data"`
}

// EipWithCvmID ...
type EipWithCvmID struct {
	dataproto.EipResult `json:",inline"`
	CvmID               string     `json:"cvm_id"`
	RelCreator          string     `json:"rel_creator"`
	RelCreatedAt        *time.Time `json:"rel_created_at"`
}

// EipCvmRelWithEipExtListReq ...
type EipCvmRelWithEipExtListReq struct {
	CvmIDs []string `json:"cvm_ids" validate:"required"`
}

// Validate ....
func (req *EipCvmRelWithEipExtListReq) Validate() error {
	return validator.Validate.Struct(req)
}

// EipCvmRelWithEipExtListResp ...
type EipCvmRelWithEipExtListResp[T dataproto.EipExtensionResult] struct {
	rest.BaseResp `json:",inline"`
	Data          []*EipExtWithCvmID[T] `json:"data"`
}

// EipExtWithCvmID ...
type EipExtWithCvmID[T dataproto.EipExtensionResult] struct {
	dataproto.EipExtResult[T] `json:",inline"`
	CvmID                     string     `json:"cvm_id"`
	RelCreator                string     `json:"rel_creator"`
	RelCreatedAt              *time.Time `json:"rel_created_at"`
}
