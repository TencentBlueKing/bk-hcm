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

package user

import (
	"fmt"
	"reflect"

	"github.com/jmoiron/sqlx"

	"hcm/pkg/api/core"
	coreuser "hcm/pkg/api/core/user"
	dataservice "hcm/pkg/api/data-service"
	dssubaccount "hcm/pkg/api/data-service/cloud/sub-account"
	dsuser "hcm/pkg/api/data-service/user"
	"hcm/pkg/criteria/errf"
	"hcm/pkg/dal/dao/orm"
	"hcm/pkg/dal/dao/tools"
	"hcm/pkg/dal/dao/types"
	tableuser "hcm/pkg/dal/table/user"
	"hcm/pkg/logs"
	"hcm/pkg/rest"
)

// CreateUserCollection create user collection.
func (svc *service) CreateUserCollection(cts *rest.Contexts) (interface{}, error) {

	req := new(dsuser.UserCollectionCreateReq)
	if err := cts.DecodeInto(req); err != nil {
		return nil, errf.NewFromErr(errf.DecodeRequestFailed, err)
	}

	if err := req.Validate(); err != nil {
		return nil, errf.NewFromErr(errf.InvalidParameter, err)
	}

	result, err := svc.dao.Txn().AutoTxn(cts.Kit, func(txn *sqlx.Tx, opt *orm.TxnOption) (interface{}, error) {
		model := &tableuser.UserCollTable{
			User:    cts.Kit.User,
			ResType: req.ResType,
			ResID:   req.ResID,
			Creator: cts.Kit.User,
		}
		id, err := svc.dao.UserCollection().CreateWithTx(cts.Kit, txn, model)
		if err != nil {
			logs.Errorf("create user collection failed, err: %v, rid: %s", err, cts.Kit.Rid)
			return nil, fmt.Errorf("create user collection failed, err: %v", err)
		}

		return id, nil
	})
	if err != nil {
		return nil, err
	}

	id, ok := result.(string)
	if !ok {
		return nil, fmt.Errorf("create user collection but return id type not string, id type: %v",
			reflect.TypeOf(id).String())
	}

	return &core.CreateResult{ID: id}, nil
}

// BatchDeleteUserCollection delete user collection.
func (svc *service) BatchDeleteUserCollection(cts *rest.Contexts) (interface{}, error) {

	req := new(dataservice.BatchDeleteReq)
	if err := cts.DecodeInto(req); err != nil {
		return nil, err
	}

	if err := req.Validate(); err != nil {
		return nil, errf.NewFromErr(errf.InvalidParameter, err)
	}

	opt := &types.ListOption{
		Filter: req.Filter,
		Page:   core.NewDefaultBasePage(),
	}
	listResp, err := svc.dao.UserCollection().List(cts.Kit, opt)
	if err != nil {
		logs.Errorf("list user collection failed, err: %v, rid: %s", err, cts.Kit.Rid)
		return nil, fmt.Errorf("list sub account failed, err: %v", err)
	}

	if len(listResp.Details) == 0 {
		return nil, nil
	}

	delIDs := make([]string, len(listResp.Details))
	for index, one := range listResp.Details {
		delIDs[index] = one.ID
	}

	_, err = svc.dao.Txn().AutoTxn(cts.Kit, func(txn *sqlx.Tx, opt *orm.TxnOption) (interface{}, error) {
		flt := tools.ContainersExpression("id", delIDs)
		if err = svc.dao.UserCollection().DeleteWithTx(cts.Kit, txn, flt); err != nil {
			return nil, err
		}

		return nil, nil
	})
	if err != nil {
		logs.Errorf("delete user collection failed, err: %v, rid: %s", err, cts.Kit.Rid)
		return nil, err
	}

	return nil, nil
}

// ListUserCollection list user collection.
func (svc *service) ListUserCollection(cts *rest.Contexts) (interface{}, error) {
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
	result, err := svc.dao.UserCollection().List(cts.Kit, opt)
	if err != nil {
		logs.Errorf("list user collection failed, err: %v, rid: %s", err, cts.Kit.Rid)
		return nil, fmt.Errorf("list user collection failed, err: %v", err)
	}
	if req.Page.Count {
		return &dssubaccount.ListResult{Count: result.Count}, nil
	}

	details := make([]coreuser.UserCollection, 0, len(result.Details))
	for _, one := range result.Details {
		details = append(details, coreuser.UserCollection{
			ID:      one.User,
			User:    one.User,
			ResType: one.ResType,
			ResID:   one.ResID,
			CreatedRevision: core.CreatedRevision{
				Creator:   one.Creator,
				CreatedAt: one.CreatedAt.String(),
			},
		})
	}

	return &dsuser.UserCollectionListResult{Details: details}, nil
}
