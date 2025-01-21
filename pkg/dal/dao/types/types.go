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

// Package types ...
package types

import (
	"hcm/pkg/api/core"
	"hcm/pkg/criteria/errf"
	"hcm/pkg/runtime/filter"
)

// ListResult define list result.
type ListResult[T any] struct {
	Count   uint64 `json:"count,omitempty"`
	Details []T    `json:"details,omitempty"`
}

// ListOption defines options to list resources.
type ListOption struct {
	Fields []string
	Filter *filter.Expression
	Page   *core.BasePage
}

// Validate list option.
func (opt ListOption) Validate(eo *filter.ExprOption, po *core.PageOption) error {
	if opt.Filter == nil {
		return errf.New(errf.InvalidParameter, "filter expr is required")
	}

	if opt.Page == nil {
		return errf.New(errf.InvalidParameter, "page is required")
	}

	if eo == nil {
		return errf.New(errf.InvalidParameter, "filter expr option is required")
	}

	if po == nil {
		return errf.New(errf.InvalidParameter, "page option is required")
	}

	if err := opt.Filter.Validate(eo); err != nil {
		return err
	}

	if err := opt.Page.Validate(po); err != nil {
		return err
	}

	return nil
}

// CountOption defines options to count resources.
type CountOption struct {
	Filter  *filter.Expression
	GroupBy string
}

// Validate list option.
func (opt *CountOption) Validate(eo *filter.ExprOption) error {
	if opt.Filter == nil {
		return errf.New(errf.InvalidParameter, "filter expr is required")
	}

	if eo == nil {
		return errf.New(errf.InvalidParameter, "filter expr option is required")
	}

	if err := opt.Filter.Validate(eo); err != nil {
		return err
	}

	return nil
}

// ValidateExcludeFilter validate list option, Filter is allowed to be empty.
func (opt ListOption) ValidateExcludeFilter(eo *filter.ExprOption, po *core.PageOption) error {
	if opt.Filter != nil {
		if eo == nil {
			return errf.New(errf.InvalidParameter, "filter expr option is required")
		}
		if err := opt.Filter.Validate(eo); err != nil {
			return err
		}
	}

	if opt.Page == nil {
		return errf.New(errf.InvalidParameter, "page is required")
	}

	if po == nil {
		return errf.New(errf.InvalidParameter, "page option is required")
	}

	if err := opt.Page.Validate(po); err != nil {
		return err
	}

	return nil
}

// CountResult defines count resources with group by options result.
type CountResult struct {
	GroupField string `db:"group_field" json:"group_field"`
	Count      uint64 `db:"count" json:"count"`
}

// DefaultIgnoredFields is default ignored field.
var DefaultIgnoredFields = []string{"id", "creator", "created_at", "tenant_id", "rel_created_at"}

// DefaultPageSQLOption define default page sql option.
var DefaultPageSQLOption = &PageSQLOption{Sort: SortOption{Sort: "id", IfNotPresent: true}}

// DefaultRelJoinWithoutField 因为rel表join时，id、creator、created_at 在两张表中都有，该字段需要手动设置。
var DefaultRelJoinWithoutField = []string{"id", "creator", "created_at"}
