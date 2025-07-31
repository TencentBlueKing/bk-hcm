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

package eip

import (
	"hcm/pkg/api/core"
	"hcm/pkg/criteria/validator"
	"hcm/pkg/runtime/filter"
)

// EipExtBatchCreateReq ...
type EipExtBatchCreateReq[T EipExtensionCreateReq] []*EipExtCreateReq[T]

// Validate ...
func (req *EipExtBatchCreateReq[T]) Validate() error {
	for _, r := range *req {
		if err := r.Validate(); err != nil {
			return err
		}
	}
	return nil
}

// EipExtCreateReq ...
type EipExtCreateReq[T EipExtensionCreateReq] struct {
	AccountID  string  `json:"account_id" validate:"required"`
	Name       *string `json:"name"`
	CloudID    string  `json:"cloud_id" validate:"required"`
	Region     string  `json:"region" validate:"required"`
	InstanceId *string `json:"instance_id"`
	Status     string  `json:"status"`
	PublicIp   string  `json:"public_ip"`
	PrivateIp  string  `json:"private_ip"`
	BkBizID    int64   `json:"bk_biz_id"`
	Extension  *T      `json:"extension"`
}

// Validate ...
func (req *EipExtCreateReq[T]) Validate() error {
	return validator.Validate.Struct(req)
}

// EipExtensionCreateReq ...
type EipExtensionCreateReq interface {
	TCloudEipExtensionCreateReq |
		AwsEipExtensionCreateReq |
		AzureEipExtensionCreateReq |
		GcpEipExtensionCreateReq |
		HuaWeiEipExtensionCreateReq
}

// EipListReq ...
type EipListReq struct {
	Filter *filter.Expression `json:"filter" validate:"required"`
	Page   *core.BasePage     `json:"page" validate:"required"`
	Fields []string           `json:"fields" validate:"omitempty"`
}

// Validate ...
func (req *EipListReq) Validate() error {
	return validator.Validate.Struct(req)
}

// EipExtensionUpdateReq ...
type EipExtensionUpdateReq interface {
	TCloudEipExtensionUpdateReq |
		AwsEipExtensionUpdateReq |
		AzureEipExtensionUpdateReq |
		GcpEipExtensionUpdateReq |
		HuaWeiEipExtensionUpdateReq
}

// EipExtBatchUpdateReq ...
type EipExtBatchUpdateReq[T EipExtensionUpdateReq] []*EipExtUpdateReq[T]

// Validate ...
func (req *EipExtBatchUpdateReq[T]) Validate() error {
	for _, r := range *req {
		if err := r.Validate(); err != nil {
			return err
		}
	}
	return nil
}

// EipExtUpdateReq ...
type EipExtUpdateReq[T EipExtensionUpdateReq] struct {
	ID            string  `json:"id" validate:"required"`
	Name          *string `json:"name" validate:"omitempty"`
	BkBizID       uint64  `json:"bk_biz_id"`
	Status        string  `json:"status"`
	RecycleStatus string  `json:"recycle_status"`
	Extension     *T      `json:"extension"`
}

// Validate ...
func (req *EipExtUpdateReq[T]) Validate() error {
	return validator.Validate.Struct(req)
}

// EipBatchUpdateReq ...
type EipBatchUpdateReq struct {
	IDs           []string `json:"ids" validate:"required"`
	BkBizID       uint64   `json:"bk_biz_id"`
	Status        string   `json:"status"`
	InstanceId    *string  `json:"instance_id"`
	InstanceType  string   `json:"instance_type"`
	RecycleStatus string   `json:"recycle_status"`
}

// Validate ...
func (req *EipBatchUpdateReq) Validate() error {
	return validator.Validate.Struct(req)
}

// EipDeleteReq ...
type EipDeleteReq struct {
	Filter *filter.Expression `json:"filter" validate:"required"`
}

// Validate ...
func (req *EipDeleteReq) Validate() error {
	return validator.Validate.Struct(req)
}
