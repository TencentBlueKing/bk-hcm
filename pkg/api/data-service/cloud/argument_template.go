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
	"errors"
	"fmt"

	"hcm/pkg/api/core"
	coreargstpl "hcm/pkg/api/core/cloud/argument-template"
	"hcm/pkg/criteria/constant"
	"hcm/pkg/criteria/enumor"
	"hcm/pkg/criteria/validator"
	"hcm/pkg/dal/table/types"
	"hcm/pkg/rest"
	"hcm/pkg/runtime/filter"
)

// -------------------------- Create --------------------------

// ArgsTplBatchCreateReq argument template create req.
type ArgsTplBatchCreateReq[Extension coreargstpl.Extension] struct {
	ArgumentTemplates []ArgsTplBatchCreate[Extension] `json:"argument_templates" validate:"required"`
}

// ArgsTplBatchCreate define argument template batch create.
type ArgsTplBatchCreate[Extension coreargstpl.Extension] struct {
	CloudID        string              `json:"cloud_id" validate:"required"`
	Name           string              `json:"name" validate:"required"`
	Vendor         string              `json:"vendor" validate:"required"`
	AccountID      string              `json:"account_id" validate:"required"`
	BkBizID        int64               `json:"bk_biz_id" validate:"omitempty"`
	Type           enumor.TemplateType `json:"type"`
	Templates      types.JsonField     `json:"templates"`
	GroupTemplates types.JsonField     `json:"group_templates"`
	Memo           *string             `json:"memo"`
	Extension      *Extension          `json:"extension"`
}

// Validate argument template create request.
func (req *ArgsTplBatchCreateReq[T]) Validate() error {
	if len(req.ArgumentTemplates) > constant.BatchOperationMaxLimit {
		return fmt.Errorf("argument templates count should <= %d", constant.BatchOperationMaxLimit)
	}

	for _, item := range req.ArgumentTemplates {
		if err := validator.Validate.Struct(item); err != nil {
			return err
		}

		if (item.Type == enumor.AddressType || item.Type == enumor.ServiceType) && len(item.Templates) == 0 {
			return errors.New("templates is required")
		}

		if (item.Type == enumor.AddressGroupType || item.Type == enumor.ServiceGroupType) &&
			len(item.GroupTemplates) == 0 {
			return errors.New("group_templates is required")
		}
	}

	return validator.Validate.Struct(req)
}

// -------------------------- Update --------------------------

// ArgsTplExtUpdateReq ...
type ArgsTplExtUpdateReq[T coreargstpl.Extension] struct {
	ID             string              `json:"id" validate:"required"`
	Name           string              `json:"name"`
	Vendor         string              `json:"vendor"`
	BkBizID        uint64              `json:"bk_biz_id"`
	AccountID      string              `json:"account_id"`
	Type           enumor.TemplateType `json:"type"`
	Templates      types.JsonField     `json:"templates"`
	GroupTemplates types.JsonField     `json:"group_templates"`
	Memo           *string             `json:"memo"`
	*core.Revision `json:",inline"`
	Extension      *T `json:"extension"`
}

// Validate ...
func (req *ArgsTplExtUpdateReq[T]) Validate() error {
	return validator.Validate.Struct(req)
}

// ArgsTplExtBatchUpdateReq ...
type ArgsTplExtBatchUpdateReq[T coreargstpl.Extension] []*ArgsTplExtUpdateReq[T]

// Validate ...
func (req *ArgsTplExtBatchUpdateReq[T]) Validate() error {
	for _, r := range *req {
		if err := r.Validate(); err != nil {
			return err
		}
	}

	return nil
}

// -------------------------- UpdateExpr --------------------------

// ArgsTplBatchUpdateExprReq ...
type ArgsTplBatchUpdateExprReq struct {
	IDs            []string            `json:"ids" validate:"required"`
	BkBizID        int64               `json:"bk_biz_id"`
	Name           string              `json:"name"`
	Type           enumor.TemplateType `json:"type"`
	Templates      types.JsonField     `json:"templates"`
	GroupTemplates types.JsonField     `json:"group_templates"`
}

// Validate ...
func (req *ArgsTplBatchUpdateExprReq) Validate() error {
	return validator.Validate.Struct(req)
}

// -------------------------- List --------------------------

// ArgsTplListReq list req.
type ArgsTplListReq struct {
	Field  []string           `json:"field" validate:"omitempty"`
	Filter *filter.Expression `json:"filter" validate:"required"`
	Page   *core.BasePage     `json:"page" validate:"required"`
}

// Validate list request.
func (req *ArgsTplListReq) Validate() error {
	return validator.Validate.Struct(req)
}

// ArgsTplListResult define argument template list result.
type ArgsTplListResult struct {
	Count   uint64                    `json:"count"`
	Details []coreargstpl.BaseArgsTpl `json:"details"`
}

// ArgsTplListResp define list resp.
type ArgsTplListResp struct {
	rest.BaseResp `json:",inline"`
	Data          *ArgsTplListResult `json:"data"`
}

// ArgsTplExtListReq list req.
type ArgsTplExtListReq struct {
	Field  []string           `json:"field" validate:"omitempty"`
	Filter *filter.Expression `json:"filter" validate:"required"`
	Page   *core.BasePage     `json:"page" validate:"required"`
}

// Validate list request.
func (req *ArgsTplExtListReq) Validate() error {
	return validator.Validate.Struct(req)
}

// ArgsTplExtListResult define argument template with extension list result.
type ArgsTplExtListResult[T coreargstpl.Extension] struct {
	Count   uint64                    `json:"count,omitempty"`
	Details []*coreargstpl.ArgsTpl[T] `json:"details,omitempty"`
}

// ArgsTplExtListResp define list resp.
type ArgsTplExtListResp[T coreargstpl.Extension] struct {
	rest.BaseResp `json:",inline"`
	Data          *ArgsTplExtListResult[T] `json:"data"`
}

// ListExtResp ...
type ListExtResp[T coreargstpl.Extension] struct {
	rest.BaseResp `json:",inline"`
	Data          *ListExtResult[T] `json:"data"`
}

// ListExtResult ...
type ListExtResult[T coreargstpl.Extension] struct {
	Count   uint64                    `json:"count,omitempty"`
	Details []*coreargstpl.ArgsTpl[T] `json:"details"`
}

// -------------------------- Delete --------------------------

// ArgsTplBatchDeleteReq delete request.
type ArgsTplBatchDeleteReq struct {
	Filter *filter.Expression `json:"filter" validate:"required"`
}

// Validate delete request.
func (req *ArgsTplBatchDeleteReq) Validate() error {
	return validator.Validate.Struct(req)
}
