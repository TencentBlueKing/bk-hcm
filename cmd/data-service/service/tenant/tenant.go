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

// Package tenant ...
package tenant

import (
	"fmt"
	"hcm/pkg/runtime/filter"
	"hcm/pkg/tools/slice"
	"reflect"

	"hcm/pkg/api/core"
	coretenant "hcm/pkg/api/core/tenant"
	"hcm/pkg/api/data-service/tenant"
	"hcm/pkg/criteria/errf"
	"hcm/pkg/dal/dao/orm"
	"hcm/pkg/dal/dao/tools"
	"hcm/pkg/dal/dao/types"
	tabletenant "hcm/pkg/dal/table/tenant"
	"hcm/pkg/logs"
	"hcm/pkg/rest"

	"github.com/jmoiron/sqlx"
)

// CreateTenant create tenant.
func (svc *service) CreateTenant(cts *rest.Contexts) (interface{}, error) {
	req := new(tenant.CreateTenantReq)
	if err := cts.DecodeInto(req); err != nil {
		return nil, errf.NewFromErr(errf.DecodeRequestFailed, err)
	}

	if err := req.Validate(); err != nil {
		return nil, errf.NewFromErr(errf.InvalidParameter, err)
	}

	tenantIDs, err := svc.dao.Txn().AutoTxn(cts.Kit, func(txn *sqlx.Tx, opt *orm.TxnOption) (interface{}, error) {
		models := make([]tabletenant.TenantTable, 0, len(req.Items))
		for _, item := range req.Items {
			model := tabletenant.TenantTable{
				TenantID: item.TenantID,
				Status:   item.Status,
				Creator:  cts.Kit.User,
				Reviser:  cts.Kit.User,
			}
			models = append(models, model)
		}
		ids, err := svc.dao.Tenant().CreateWithTx(cts.Kit, txn, models)
		if err != nil {
			return nil, fmt.Errorf("batch create tenant failed, err: %v", err)
		}

		return ids, nil
	})
	if err != nil {
		logs.Errorf("batch create tenant commit txn failed, err: %v, rid: %s", err, cts.Kit.Rid)
		return nil, err
	}

	ids, ok := tenantIDs.([]string)
	if !ok {
		return nil, fmt.Errorf("create tenant but return id type not string, id type: %v",
			reflect.TypeOf(tenantIDs).String())
	}

	return &core.BatchCreateResult{IDs: ids}, nil
}

// DeleteTenant delete tenant.
func (svc *service) DeleteTenant(cts *rest.Contexts) (interface{}, error) {
	req := new(tenant.DeleteTenantReq)
	if err := cts.DecodeInto(req); err != nil {
		return nil, err
	}

	if err := req.Validate(); err != nil {
		return nil, errf.NewFromErr(errf.InvalidParameter, err)
	}

	_, err := svc.dao.Txn().AutoTxn(cts.Kit, func(txn *sqlx.Tx, opt *orm.TxnOption) (interface{}, error) {
		// 循环分页查询出所有待删除记录的id然后再循环分页删除
		delIDs := make([]string, 0)
		page := core.NewDefaultBasePage()
		for {
			opt := &types.ListOption{
				Fields: []string{"id"},
				Filter: req.Filter,
				Page:   page,
			}
			listResp, err := svc.dao.Tenant().List(cts.Kit, opt)
			if err != nil {
				logs.Errorf("list tenant failed, err: %v, rid: %s", err, cts.Kit.Rid)
				return nil, fmt.Errorf("list tenant failed, err: %v", err)
			}

			for _, one := range listResp.Tenants {
				delIDs = append(delIDs, one.ID)
			}

			// 如果当前页数据不足一页，说明后面没有更多数据了
			if uint(len(listResp.Tenants)) < page.Limit {
				break
			}
			page.Start += uint32(page.Limit)
		}

		for _, chunk := range slice.Split(delIDs, int(filter.DefaultMaxInLimit)) {
			delFilter := tools.ContainersExpression("id", chunk)
			if err := svc.dao.Tenant().DeleteWithTx(cts.Kit, txn, delFilter); err != nil {
				logs.Errorf("delete tenant chunk failed, err: %s, chunk: %v, rid: %s", err, chunk, cts.Kit.Rid)
				return nil, err
			}
		}
		return nil, nil
	})
	if err != nil {
		logs.Errorf("delete tenant failed, err: %v, rid: %s", err, cts.Kit.Rid)
		return nil, err
	}
	return nil, nil
}

// UpdateTenant update tenant.
func (svc *service) UpdateTenant(cts *rest.Contexts) (interface{}, error) {
	req := new(tenant.UpdateTenantReq)
	if err := cts.DecodeInto(req); err != nil {
		return nil, errf.NewFromErr(errf.DecodeRequestFailed, err)
	}

	if err := req.Validate(); err != nil {
		return nil, errf.NewFromErr(errf.InvalidParameter, err)
	}

	_, err := svc.dao.Txn().AutoTxn(cts.Kit, func(txn *sqlx.Tx, opt *orm.TxnOption) (interface{}, error) {
		for _, one := range req.Items {
			tenants := &tabletenant.TenantTable{
				TenantID: one.TenantID,
				Status:   one.Status,
				Reviser:  cts.Kit.User,
			}

			flt := tools.EqualExpression("id", one.ID)
			if err := svc.dao.Tenant().UpdateWithTx(cts.Kit, txn, flt, tenants); err != nil {
				logs.Errorf("update tenant failed, err: %v, rid: %s, tenant_id: %s", err, cts.Kit.Rid, one.TenantID)
				return nil, fmt.Errorf("update tenant failed, err: %v", err)
			}
		}

		return nil, nil
	})
	if err != nil {
		return nil, err
	}

	return nil, nil
}

// ListTenant list tenant.
func (svc *service) ListTenant(cts *rest.Contexts) (interface{}, error) {
	req := new(core.ListReq)
	if err := cts.DecodeInto(req); err != nil {
		return nil, err
	}

	if err := req.Validate(); err != nil {
		return nil, errf.NewFromErr(errf.InvalidParameter, err)
	}

	opt := &types.ListOption{
		Filter: req.Filter,
		Page:   req.Page,
		Fields: req.Fields,
	}
	res, err := svc.dao.Tenant().List(cts.Kit, opt)
	if err != nil {
		logs.Errorf("list tenant failed, err: %v, rid: %s", err, cts.Kit.Rid)
		return nil, fmt.Errorf("list tenant failed, err: %v", err)
	}
	if req.Page.Count {
		return &tenant.ListTenantResult{Count: res.Count}, nil
	}

	tenants := make([]coretenant.Tenant, 0, len(res.Tenants))
	for _, one := range res.Tenants {
		tenants = append(tenants, coretenant.Tenant{
			ID:       one.ID,
			TenantID: one.TenantID,
			Status:   one.Status,
			Revision: core.Revision{
				Creator:   one.Creator,
				Reviser:   one.Reviser,
				CreatedAt: one.CreatedAt.String(),
				UpdatedAt: one.UpdatedAt.String(),
			},
		})
	}

	return &tenant.ListTenantResult{Details: tenants}, nil
}
