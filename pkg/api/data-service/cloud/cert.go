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

	"hcm/pkg/api/core"
	corecert "hcm/pkg/api/core/cloud/cert"
	"hcm/pkg/criteria/constant"
	"hcm/pkg/criteria/enumor"
	"hcm/pkg/criteria/validator"
	"hcm/pkg/dal/table/types"
	"hcm/pkg/rest"
	"hcm/pkg/runtime/filter"
)

// -------------------------- Create --------------------------

// CertBatchCreateReq cert create req.
type CertBatchCreateReq[Extension corecert.Extension] struct {
	Certs []CertBatchCreate[Extension] `json:"certs" validate:"required"`
}

// CertBatchCreate define cert batch create.
type CertBatchCreate[Extension corecert.Extension] struct {
	CloudID          string          `json:"cloud_id" validate:"required"`
	Name             string          `json:"name"`
	Vendor           string          `json:"vendor" validate:"required"`
	AccountID        string          `json:"account_id" validate:"required"`
	BkBizID          int64           `json:"bk_biz_id" validate:"omitempty"`
	Domain           types.JsonField `json:"domain"`
	CertType         enumor.CertType `json:"cert_type"`
	CertStatus       string          `json:"cert_status"`
	EncryptAlgorithm string          `json:"encrypt_algorithm"`
	CloudCreatedTime string          `json:"cloud_created_time"`
	CloudExpiredTime string          `json:"cloud_expired_time"`
	Memo             *string         `json:"memo"`
	Extension        *Extension      `json:"extension"`
}

// Validate cert create request.
func (req *CertBatchCreateReq[T]) Validate() error {
	if len(req.Certs) > constant.BatchOperationMaxLimit {
		return fmt.Errorf("certs count should <= %d", constant.BatchOperationMaxLimit)
	}

	return validator.Validate.Struct(req)
}

// -------------------------- Update --------------------------

// CertExtUpdateReq ...
type CertExtUpdateReq[T corecert.Extension] struct {
	ID               string          `json:"id" validate:"required"`
	Name             string          `json:"name"`
	Vendor           string          `json:"vendor"`
	BkBizID          uint64          `json:"bk_biz_id"`
	AccountID        string          `json:"account_id"`
	Domain           types.JsonField `json:"domain"`
	CertType         enumor.CertType `json:"cert_type"`
	CertStatus       string          `json:"cert_status"`
	EncryptAlgorithm string          `json:"encrypt_algorithm"`
	CloudCreatedTime string          `json:"cloud_created_time"`
	CloudExpiredTime string          `json:"cloud_expired_time"`
	Memo             *string         `json:"memo"`
	*core.Revision   `json:",inline"`
	Extension        *T `json:"extension"`
}

// Validate ...
func (req *CertExtUpdateReq[T]) Validate() error {
	return validator.Validate.Struct(req)
}

// CertExtBatchUpdateReq ...
type CertExtBatchUpdateReq[T corecert.Extension] []*CertExtUpdateReq[T]

// Validate ...
func (req *CertExtBatchUpdateReq[T]) Validate() error {
	for _, r := range *req {
		if err := r.Validate(); err != nil {
			return err
		}
	}

	return nil
}

// -------------------------- UpdateExpr --------------------------

// CertBatchUpdateExprReq ...
type CertBatchUpdateExprReq struct {
	IDs              []string        `json:"ids" validate:"required"`
	BkBizID          int64           `json:"bk_biz_id"`
	Domain           types.JsonField `json:"domain"`
	CertType         enumor.CertType `json:"cert_type"`
	CertStatus       string          `json:"cert_status"`
	EncryptAlgorithm string          `json:"encrypt_algorithm"`
	CloudCreatedTime string          `json:"cloud_created_time"`
	CloudExpiredTime string          `json:"cloud_expired_time"`
}

// Validate ...
func (req *CertBatchUpdateExprReq) Validate() error {
	return validator.Validate.Struct(req)
}

// -------------------------- List --------------------------

// CertListReq list req.
type CertListReq struct {
	Field  []string           `json:"field" validate:"omitempty"`
	Filter *filter.Expression `json:"filter" validate:"required"`
	Page   *core.BasePage     `json:"page" validate:"required"`
}

// Validate list request.
func (req *CertListReq) Validate() error {
	return validator.Validate.Struct(req)
}

// CertListResult define cert list result.
type CertListResult struct {
	Count   uint64              `json:"count"`
	Details []corecert.BaseCert `json:"details"`
}

// CertListResp define list resp.
type CertListResp struct {
	rest.BaseResp `json:",inline"`
	Data          *CertListResult `json:"data"`
}

// CertExtListReq list req.
type CertExtListReq struct {
	Field  []string           `json:"field" validate:"omitempty"`
	Filter *filter.Expression `json:"filter" validate:"required"`
	Page   *core.BasePage     `json:"page" validate:"required"`
}

// Validate list request.
func (req *CertExtListReq) Validate() error {
	return validator.Validate.Struct(req)
}

// CertExtListResult define cert with extension list result.
type CertExtListResult[T corecert.Extension] struct {
	Count   uint64             `json:"count,omitempty"`
	Details []corecert.Cert[T] `json:"details,omitempty"`
}

// CertExtListResp define list resp.
type CertExtListResp[T corecert.Extension] struct {
	rest.BaseResp `json:",inline"`
	Data          *CertExtListResult[T] `json:"data"`
}

// CertListExtResp ...
type CertListExtResp[T corecert.Extension] struct {
	rest.BaseResp `json:",inline"`
	Data          *CertListExtResult[T] `json:"data"`
}

// CertListExtResult ...
type CertListExtResult[T corecert.Extension] struct {
	Count   uint64              `json:"count,omitempty"`
	Details []*corecert.Cert[T] `json:"details"`
}

// -------------------------- Delete --------------------------

// CertBatchDeleteReq delete request.
type CertBatchDeleteReq struct {
	Filter *filter.Expression `json:"filter" validate:"required"`
}

// Validate delete request.
func (req *CertBatchDeleteReq) Validate() error {
	return validator.Validate.Struct(req)
}
