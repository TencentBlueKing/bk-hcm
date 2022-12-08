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
	"reflect"

	"hcm/cmd/data-service/service/capability"
	"hcm/pkg/api/core"
	"hcm/pkg/api/core/cloud"
	protocloud "hcm/pkg/api/data-service/cloud"
	"hcm/pkg/criteria/enumor"
	"hcm/pkg/criteria/errf"
	"hcm/pkg/dal/dao/orm"
	"hcm/pkg/dal/dao/types"
	"hcm/pkg/logs"
	"hcm/pkg/rest"
	"hcm/pkg/runtime/filter"

	"github.com/jmoiron/sqlx"

	"hcm/pkg/dal/dao"
)

// InitAccountService initial the account service
func InitAccountService(cap *capability.Capability) {
	svc := &accountSvc{
		dao: cap.Dao,
	}

	h := rest.NewHandler()

	h.Add("CreateAccount", "POST", "/account/create", svc.CreateAccount)
	h.Add("ListAccount", "POST", "/account/list", svc.ListAccount)
	h.Add("UpdateAccount", "PATCH", "/account", svc.UpdateAccount)
	h.Add("DeleteAccount", "DELETE", "/account", svc.DeleteAccount)

	h.Load(cap.WebService)
}

// TODO 考虑废弃 accountSvc 模式
type accountSvc struct {
	dao dao.Set
}

// CreateAccount account with options
func (svc *accountSvc) CreateAccount(cts *rest.Contexts) (interface{}, error) {
	req := new(protocloud.CreateAccountReq)

	if err := cts.DecodeInto(req); err != nil {
		return nil, errf.New(errf.DecodeRequestFailed, err.Error())
	}

	if err := req.Validate(); err != nil {
		return nil, errf.Newf(errf.InvalidParameter, err.Error())
	}

	accountID, err := svc.dao.Txn().AutoTxn(cts.Kit, func(txn *sqlx.Tx, opt *orm.TxnOption) (interface{}, error) {
		account := &cloud.Account{
			Vendor: req.Vendor,
			Spec: &cloud.AccountSpec{
				Name:         req.Spec.Name,
				Managers:     req.Spec.Managers,
				DepartmentID: req.Spec.DepartmentID,
				Type:         req.Spec.Type,
				Site:         req.Spec.Site,
				SyncStatus:   enumor.NotStart,
				Memo:         req.Spec.Memo,
			},
			Extension: req.Extension,
			Revision: &core.Revision{
				Creator: cts.Kit.User,
				Reviser: cts.Kit.User,
			},
		}

		accountID, err := svc.dao.Account().CreateWithTx(cts.Kit, txn, account)
		if err != nil {
			return nil, fmt.Errorf("create account failed, err: %v", err)
		}

		rels := make([]*cloud.AccountBizRel, len(req.Attachment.BkBizIDs))
		for index, bizID := range req.Attachment.BkBizIDs {
			rels[index] = &cloud.AccountBizRel{
				BkBizID:   bizID,
				AccountID: accountID,
				Creator:   cts.Kit.User,
			}
		}
		_, err = svc.dao.AccountBizRel().BatchCreateWithTx(cts.Kit, txn, rels)
		if err != nil {
			return nil, fmt.Errorf("batch create account_biz_rels failed, err: %v", err)
		}

		return accountID, nil
	})
	if err != nil {
		return nil, err
	}

	id, ok := accountID.(uint64)
	if !ok {
		return nil, fmt.Errorf("create account but return id type not uint64, id type: %v",
			reflect.TypeOf(accountID).String())
	}

	return &core.CreateResult{ID: id}, nil
}

// UpdateAccount account with filter.
func (svc *accountSvc) UpdateAccount(cts *rest.Contexts) (interface{}, error) {
	req := new(protocloud.UpdateAccountReq)

	if err := cts.DecodeInto(req); err != nil {
		return nil, errf.New(errf.DecodeRequestFailed, err.Error())
	}

	if err := req.Validate(); err != nil {
		return nil, errf.Newf(errf.InvalidParameter, err.Error())
	}

	account := &cloud.Account{
		Spec: req.Spec,
		Revision: &core.Revision{
			Reviser: cts.Kit.User,
		},
	}

	err := svc.dao.Account().Update(cts.Kit, req.Filter, account)
	if err != nil {
		logs.Errorf("update account failed, err: %v, rid: %s", cts.Kit.Rid)
		return nil, fmt.Errorf("create account failed, err: %v", err)
	}

	return nil, nil
}

// ListAccount accounts with filter
func (svc *accountSvc) ListAccount(cts *rest.Contexts) (interface{}, error) {
	req := new(protocloud.ListAccountReq)
	if err := cts.DecodeInto(req); err != nil {
		return nil, err
	}

	if err := req.Validate(); err != nil {
		return nil, errf.Newf(errf.InvalidParameter, err.Error())
	}

	opt := &types.ListOption{
		Filter: req.Filter,
		Page:   req.Page,
	}
	accountResp, err := svc.dao.Account().List(cts.Kit, opt)
	if err != nil {
		logs.Errorf("list account failed, err: %v, rid: %s", cts.Kit.Rid)
		return nil, fmt.Errorf("list account failed, err: %v", err)
	}
	if req.Page.Count {
		return &protocloud.ListAccountResult{Count: accountResp.Count}, nil
	}

	// TODO：改为批量查
	for index, one := range accountResp.Details {
		opt := &types.ListOption{
			Filter: &filter.Expression{
				Op: filter.And,
				Rules: []filter.RuleFactory{
					&filter.AtomRule{
						Field: "account_id",
						Op:    filter.Equal.Factory(),
						Value: one.ID,
					},
				},
			},
			// TODO：支持查询全量的Page
			Page: &types.BasePage{
				Start: 0,
				Limit: types.DefaultMaxPageLimit,
			},
		}

		relResp, err := svc.dao.AccountBizRel().List(cts.Kit, opt)
		if err != nil {
			return nil, err
		}

		bizIDs := make([]int64, len(relResp.Details))
		for idx, rel := range relResp.Details {
			bizIDs[idx] = rel.BkBizID
		}

		accountResp.Details[index].Attachment = &cloud.AccountAttachment{
			BkBizIDs: bizIDs,
		}
	}

	return &protocloud.ListAccountResult{Details: accountResp.Details}, nil
}

// DeleteAccount account with filter.
func (svc *accountSvc) DeleteAccount(cts *rest.Contexts) (interface{}, error) {
	req := new(protocloud.DeleteAccountReq)
	if err := cts.DecodeInto(req); err != nil {
		return nil, err
	}

	if err := req.Validate(); err != nil {
		return nil, errf.Newf(errf.InvalidParameter, err.Error())
	}

	opt := &types.ListOption{
		Filter: req.Filter,
		Page: &types.BasePage{
			Start: 0,
			Limit: types.DefaultMaxPageLimit,
		},
	}
	listResp, err := svc.dao.Account().List(cts.Kit, opt)
	if err != nil {
		logs.Errorf("list account failed, err: %v, rid: %s", cts.Kit.Rid)
		return nil, fmt.Errorf("list account failed, err: %v", err)
	}

	if len(listResp.Details) == 0 {
		return nil, nil
	}

	delAccountIDs := make([]uint64, len(listResp.Details))
	for index, one := range listResp.Details {
		delAccountIDs[index] = one.ID
	}

	_, err = svc.dao.Txn().AutoTxn(cts.Kit, func(txn *sqlx.Tx, opt *orm.TxnOption) (interface{}, error) {
		if err := svc.dao.Account().DeleteWithTx(cts.Kit, txn, req.Filter); err != nil {
			return nil, err
		}

		ftr := &filter.Expression{
			Op: filter.And,
			Rules: []filter.RuleFactory{
				&filter.AtomRule{
					Field: "account_id",
					Op:    filter.Equal.Factory(),
					Value: delAccountIDs,
				},
			},
		}
		if err := svc.dao.AccountBizRel().DeleteWithTx(cts.Kit, txn, ftr); err != nil {
			return nil, err
		}

		return nil, nil
	})
	if err != nil {
		logs.Errorf("delete account failed, err: %v, rid: %s", err, cts.Kit.Rid)
		return nil, err
	}

	return nil, nil
}
