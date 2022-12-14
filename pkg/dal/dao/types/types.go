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

package types

import (
	"hcm/pkg/criteria/errf"
	"hcm/pkg/runtime/filter"
)

// ListOption defines options to list accounts.
type ListOption struct {
	Filter *filter.Expression
	Page   *BasePage
}

// Validate list option.
func (opt *ListOption) Validate(eo *filter.ExprOption, po *PageOption) error {
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

// DefaultSqlWhereOption is default sql where option.
var DefaultSqlWhereOption = &filter.SQLWhereOption{
	Priority: filter.Priority{"id"},
}

// DefaultIgnoredFields is default ignored field.
var DefaultIgnoredFields = []string{"id", "creator", "created_at"}

// DefaultPageSQLOption define default page sql option.
var DefaultPageSQLOption = &PageSQLOption{Sort: SortOption{Sort: "id", IfNotPresent: true}}
