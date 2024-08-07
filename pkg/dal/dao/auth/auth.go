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

// Package auth ...
package auth

import (
	"fmt"

	"hcm/pkg/api/core"
	"hcm/pkg/criteria/errf"
	"hcm/pkg/dal/dao/orm"
	"hcm/pkg/dal/dao/types"
	"hcm/pkg/iam/client"
	"hcm/pkg/kit"
	"hcm/pkg/runtime/filter"
)

// Auth only used for auth.
type Auth interface {
	// ListInstances list instances with options.
	ListInstances(kt *kit.Kit, opts *types.ListInstancesOption) (*types.ListInstanceDetails, error)
}

var _ Auth = new(AuthDao)

// AuthDao auth dao.
type AuthDao struct {
	Orm orm.Interface
}

// ListInstances list instances with options.
func (r *AuthDao) ListInstances(kt *kit.Kit, opts *types.ListInstancesOption) (*types.ListInstanceDetails, error) {
	if opts == nil {
		return nil, errf.New(errf.InvalidParameter, "list instances options is null")
	}

	if opts.Priority == nil {
		opts.Priority = filter.Priority{"name"}
	}

	if opts.DisplayNameField == "" {
		opts.DisplayNameField = "name"
	}

	if opts.ResourceIDField == "" {
		opts.ResourceIDField = "id"
	}

	// enable unlimited query, because this is iam pull resource callback.
	po := &core.PageOption{MaxLimit: client.BkIAMMaxPageSize}
	if err := opts.Validate(po); err != nil {
		return nil, err
	}

	sqlOpt := &filter.SQLWhereOption{
		Priority: opts.Priority,
	}
	whereExpr, whereValue, err := opts.Filter.SQLWhereExpr(sqlOpt)
	if err != nil {
		return nil, err
	}

	var sql string
	if opts.Page.Count {
		// count instance data by whereExpr
		sql := fmt.Sprintf(`SELECT COUNT(*) FROM %s %s`, opts.TableName, whereExpr)
		count, err := r.Orm.Do().Count(kt.Ctx, sql, whereValue)
		if err != nil {
			return nil, err
		}

		return &types.ListInstanceDetails{Count: count, Details: make([]types.InstanceResource, 0)}, nil
	}

	// select instance data by whereExpr
	pageExpr, err := types.PageSQLExpr(opts.Page, &types.PageSQLOption{Sort: types.SortOption{Sort: "id",
		IfNotPresent: true}})
	if err != nil {
		return nil, err
	}

	sql = fmt.Sprintf(`SELECT %s as id, %s as name FROM %s %s %s`, opts.ResourceIDField, opts.DisplayNameField, opts.TableName, whereExpr, pageExpr)
	list := make([]types.InstanceResource, 0)
	err = r.Orm.Do().Select(kt.Ctx, &list, sql, whereValue)
	if err != nil {
		return nil, err
	}

	return &types.ListInstanceDetails{Count: 0, Details: list}, nil
}
