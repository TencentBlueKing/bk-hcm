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
	"hcm/pkg/api/core"
	corecos "hcm/pkg/api/core/cloud/cos"
	"hcm/pkg/criteria/enumor"
	"hcm/pkg/criteria/validator"
	"hcm/pkg/dal/table/types"
	"hcm/pkg/runtime/filter"
)

// CosCreate cos create.
type CosCreate[Extension corecos.CosExtension] struct {
	CloudID   string        `validate:"required" json:"cloud_id"`
	Name      string        `validate:"required" json:"name"`
	Vendor    enumor.Vendor `validate:"required"  json:"vendor"`
	AccountID string        `validate:"required" json:"account_id"`
	BkBizID   int64         `validate:"omitempty" json:"bk_biz_id"`
	Region    string        `validate:"required" json:"region"`

	ACL                       string          `json:"acl"`
	GrantFullControl          string          `json:"grant_full_control"`
	GrantRead                 string          `json:"grant_read"`
	GrantWrite                string          `json:"grant_write"`
	GrantReadACP              string          `json:"grant_read_acp"`
	GrantWriteACP             string          `json:"grant_write_acp"`
	CreateBucketConfiguration interface{} `json:"create_bucket_configuration"`

	Domain           string          `json:"domain"`
	Status           string          `json:"status"`
	CloudCreatedTime string          `json:"cloud_created_time"`
	CloudStatusTime  string          `json:"cloud_status_time"`
	CloudExpiredTime string          `json:"cloud_expired_time"`
	SyncTime         string          `json:"sync_time"`
	Tags             types.StringMap `json:"tags"`

	Extension *Extension `json:"extension"`
}

// CosListResult define cos list result.
type CosListResult = core.ListResultT[corecos.BaseCos]

// CosExtListResult define cos with extension list result.
type CosExtListResult[T corecos.CosExtension] struct {
	Count   uint64           `json:"count,omitempty"`
	Details []corecos.Cos[T] `json:"details,omitempty"`
}

// CosBatchCreateReq cos batch create request.
type CosBatchCreateReq[Extension corecos.CosExtension] struct {
	Cos []CosCreate[Extension] `json:"cos" validate:"required,min=1"`
}

// Validate cos batch create request.
func (req *CosBatchCreateReq[Extension]) Validate() error {
	return validator.Validate.Struct(req)
}

// CosExtBatchUpdateReq cos batch update request.
type CosExtBatchUpdateReq[Extension corecos.CosExtension] struct {
	Cos []*CosExtUpdateReq[Extension] `json:"cos" validate:"min=1"`
}

// CosExtUpdateReq cos update request.
type CosExtUpdateReq[Extension corecos.CosExtension] struct {
	ID      string `validate:"required" json:"id"`
	Name    string `validate:"required" json:"name"`
	BkBizID int64  `validate:"omitempty" json:"bk_biz_id"`

	ACL                       string          `json:"acl"`
	GrantFullControl          string          `json:"grant_full_control"`
	GrantRead                 string          `json:"grant_read"`
	GrantWrite                string          `json:"grant_write"`
	GrantReadACP              string          `json:"grant_read_acp"`
	GrantWriteACP             string          `json:"grant_write_acp"`
	CreateBucketConfiguration types.JsonField `json:"create_bucket_configuration"`

	Domain           string          `json:"domain"`
	Status           string          `json:"status"`
	CloudCreatedTime string          `json:"cloud_created_time"`
	CloudStatusTime  string          `json:"cloud_status_time"`
	CloudExpiredTime string          `json:"cloud_expired_time"`
	SyncTime         string          `json:"sync_time"`
	Tags             types.StringMap `json:"tags"`

	Extension *Extension `json:"extension"`
}

// Validate cos batch update request.
func (req *CosExtBatchUpdateReq[Extension]) Validate() error {
	return validator.Validate.Struct(req)
}

// CosBatchDeleteReq cos batch delete request.
type CosBatchDeleteReq struct {
	Filter *filter.Expression `json:"filter" validate:"required"`
}

// Validate delete request.
func (req *CosBatchDeleteReq) Validate() error {
	return validator.Validate.Struct(req)
}

// TCloudCosCreateReq batch create cos
type TCloudCosBatchCreateReq = CosBatchCreateReq[corecos.TCloudCosExtension]

// TCloudCosCreate create cos
type TCloudCosCreate = CosCreate[corecos.TCloudCosExtension]

// TCloudCosBatchUpdateReq batch update cos
type TCloudCosBatchUpdateReq = CosExtBatchUpdateReq[corecos.TCloudCosExtension]