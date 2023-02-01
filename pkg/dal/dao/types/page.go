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
	"errors"
	"fmt"

	"hcm/pkg/api/core"
	"hcm/pkg/criteria/errf"
)

// PageSQLOption defines the options to generate a sql expression
// based on the BasePage.
type PageSQLOption struct {
	// Sort defines the field to do sort.
	// Note:
	// 1. If set, then user defined Sort field will be overlapped.
	// 2. Sort field should always be an indexed field in db.
	Sort SortOption `json:"sort"`
}

// SortOption defines how to set the order column when do the BasePage.SQLExpr
// operation.
type SortOption struct {
	// Sort defines the sorted column.
	Sort string `json:"sort"`
	// IfNotPresent means if the sort column is not defined by user, then
	// use this Sort column as default.
	IfNotPresent bool `json:"if_not_present"`
	// ForceOverlap means no matter what sort column defined, use this
	// Sort column overlapped.
	// Note: ForceOverlap option have more priority than IfNotPresent
	ForceOverlap bool `json:"force_overlap"`
}

// PageSQLExpr return the expression of the query clause based one the page options.
// Note:
//  1. do not call this, when it's a count request.
//  2. if sort is not set, use the default resource's identity 'id' as the sort key.
//  3. if Sort is set by the system(PageSQLOption.Sort), then use its Sort value
//     according to the various options.
//
// see the test case to get more returned example and learn the supported scenarios.
func PageSQLExpr(bp *core.BasePage, ps *PageSQLOption) (where string, err error) {
	defer func() {
		if err != nil {
			err = errf.NewFromErr(errf.InvalidParameter, err)
		}
	}()
	if bp == nil {
		return "", errors.New("page is nil")
	}
	if ps == nil {
		return "", errors.New("page sql option is nil")
	}
	if bp.Count {
		// this is a count query clause.
		return "", errors.New("page.count is enabled, do not support generate SQL expression")
	}
	if bp.Start == 0 && bp.Limit == 0 {
		// it means do not need to sort.
		return "", nil
	}
	var sort string
	if ps.Sort.ForceOverlap {
		// force overlapped user defined sort field.
		sort = ps.Sort.Sort
	} else {
		if ps.Sort.IfNotPresent && len(bp.Sort) == 0 {
			// user note defined sort, then use default sort.
			sort = ps.Sort.Sort
		} else {
			// use user defined sort column
			sort = bp.Sort
		}
	}
	if len(sort) == 0 {
		// if sort is not set, use the default resource's
		// identity id as the default sort column.
		sort = "id"
	}
	expr := fmt.Sprintf("ORDER BY %s", sort)
	if bp.Start == 0 && bp.Limit == 0 {
		// this is a special scenario, which means query all the resources at once.
		return fmt.Sprintf("%s %s", expr, bp.Order.Order()), nil
	}
	// if Start >=1, then Limit can not be 0.
	if bp.Limit == 0 {
		return "", errors.New("page.limit value should >= 1")
	}
	// bp.Limit is > 0, already validated upper.
	expr = fmt.Sprintf("%s %s LIMIT %d OFFSET %d", expr, bp.Order.Order(), bp.Limit, bp.Start)
	return expr, nil
}
