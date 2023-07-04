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
	"hcm/pkg/api/core"
	rr "hcm/pkg/api/core/recycle-record"
	datarelproto "hcm/pkg/api/data-service/cloud"
	"hcm/pkg/criteria/validator"
	"hcm/pkg/runtime/filter"
)

// DiskListReq ...
type DiskListReq struct {
	Filter *filter.Expression `json:"filter" validate:"required"`
	Page   *core.BasePage     `json:"page" validate:"required"`
}

// Validate ...
func (req *DiskListReq) Validate() error {
	return validator.Validate.Struct(req)
}

// DiskAssignReq ...
type DiskAssignReq struct {
	IDs     []string `json:"disk_ids" validate:"required"`
	BkBizID uint64   `json:"bk_biz_id" validate:"required"`
}

// Validate ...
func (req *DiskAssignReq) Validate() error {
	return validator.Validate.Struct(req)
}

// DiskDetachReq ...
type DiskDetachReq struct {
	DiskID string `json:"disk_id" validate:"required"`
}

// Validate ...
func (req *DiskDetachReq) Validate() error {
	return validator.Validate.Struct(req)
}

// AccountReq ...
type AccountReq struct {
	AccountID string `json:"account_id" validate:"required"`
}

// Validate ...
func (req *AccountReq) Validate() error {
	return validator.Validate.Struct(req)
}

// DiskReq ...
type DiskReq struct {
	DiskID string `json:"disk_id" validate:"required"`
}

// Validate ...
func (req *DiskReq) Validate() error {
	return validator.Validate.Struct(req)
}

type DiskCvmRelListReq = datarelproto.DiskCvmRelListReq

// -------------------------- Recycle ------------------------

// DiskRecycleReq recycle disk request.
type DiskRecycleReq struct {
	Infos []DiskRecycleInfo `json:"infos" validate:"min=1,max=100"`
}

// DiskRecycleInfo defines recycle one disk info.
type DiskRecycleInfo struct {
	ID                     string `json:"id" validate:"required"`
	*rr.DiskRecycleOptions `json:",inline" validate:"omitempty"`
}

// Validate DiskRecycleReq
func (req DiskRecycleReq) Validate() error {
	return validator.Validate.Struct(req)
}

// -------------------------- Recover ------------------------

// DiskRecoverReq recover disk request.
type DiskRecoverReq struct {
	RecordIDs []string `json:"record_ids" validate:"min=1,max=100"`
}

// Validate DiskRecoverReq
func (req DiskRecoverReq) Validate() error {
	return validator.Validate.Struct(req)
}

// -------------------------- Recycle ------------------------

// DiskDeleteRecycleReq delete recycle disk request.
type DiskDeleteRecycleReq struct {
	RecordIDs []string `json:"record_ids" validate:"min=1,max=100"`
}

// Validate DiskDeleteRecycleReq
func (req DiskDeleteRecycleReq) Validate() error {
	return validator.Validate.Struct(req)
}
