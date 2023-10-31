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
	coredisk "hcm/pkg/api/core/cloud/disk"
	"hcm/pkg/criteria/validator"
	"hcm/pkg/runtime/filter"
)

// DiskExtCreateReq ...
type DiskExtCreateReq[T coredisk.Extension] struct {
	AccountID    string  `json:"account_id" validate:"required"`
	Name         string  `json:"name" validate:"required"`
	BkBizID      int64   `json:"bk_biz_id"`
	CloudID      string  `json:"cloud_id" validate:"required"`
	Region       string  `json:"region" validate:"required"`
	Zone         string  `json:"zone" validate:"omitempty"`
	DiskSize     uint64  `json:"disk_size" validate:"required"`
	DiskType     string  `json:"disk_type" validate:"required"`
	IsSystemDisk bool    `json:"is_system_disk" validate:"omitempty"`
	Status       string  `json:"status" validate:"required"`
	Memo         *string `json:"memo"`
	Extension    *T      `json:"extension"`
}

// Validate ...
func (req *DiskExtCreateReq[T]) Validate() error {
	return validator.Validate.Struct(req)
}

// DiskExtBatchCreateReq ...
type DiskExtBatchCreateReq[T coredisk.Extension] []*DiskExtCreateReq[T]

// Validate ...
func (req *DiskExtBatchCreateReq[T]) Validate() error {
	for _, r := range *req {
		if err := r.Validate(); err != nil {
			return err
		}
	}
	return nil
}

// DiskExtUpdateReq ...
type DiskExtUpdateReq[T coredisk.Extension] struct {
	ID           string  `json:"id" validate:"required"`
	Region       string  `json:"region"`
	Name         string  `json:"name"`
	BkBizID      uint64  `json:"bk_biz_id"`
	Status       string  `json:"status"`
	IsSystemDisk *bool   `json:"is_system_disk" validate:"omitempty"`
	Memo         *string `json:"memo"`
	Extension    *T      `json:"extension"`
}

// Validate ...
func (req *DiskExtUpdateReq[T]) Validate() error {
	return validator.Validate.Struct(req)
}

// DiskExtBatchUpdateReq ...
type DiskExtBatchUpdateReq[T coredisk.Extension] []*DiskExtUpdateReq[T]

// Validate ...
func (req *DiskExtBatchUpdateReq[T]) Validate() error {
	for _, r := range *req {
		if err := r.Validate(); err != nil {
			return err
		}
	}
	return nil
}

// DiskBatchUpdateReq ...
type DiskBatchUpdateReq struct {
	IDs     []string `json:"ids" validate:"required"`
	BkBizID uint64   `json:"bk_biz_id"`
	Status  string   `json:"status"`
	Memo    *string  `json:"memo"`
}

// Validate ...
func (req *DiskBatchUpdateReq) Validate() error {
	return validator.Validate.Struct(req)
}

// DiskDeleteReq ...
type DiskDeleteReq struct {
	Filter *filter.Expression `json:"filter" validate:"required"`
}

// Validate ...
func (req *DiskDeleteReq) Validate() error {
	return validator.Validate.Struct(req)
}
