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
	"strconv"

	"hcm/cmd/data-service/service/capability"
	"hcm/pkg/api/core"
	protocore "hcm/pkg/api/core/cloud"
	protocloud "hcm/pkg/api/data-service/cloud"
	"hcm/pkg/criteria/enumor"
	"hcm/pkg/criteria/errf"
	"hcm/pkg/dal/dao/orm"
	"hcm/pkg/dal/dao/types"
	daotypes "hcm/pkg/dal/dao/types"
	tablecloud "hcm/pkg/dal/table/cloud"
	tabletype "hcm/pkg/dal/table/types"
	"hcm/pkg/logs"
	"hcm/pkg/rest"
	"hcm/pkg/runtime/filter"
	"hcm/pkg/tools/json"

	"hcm/pkg/dal/dao"

	"github.com/jmoiron/sqlx"
)

// InitAccountService initial the account service
func InitAccountService(cap *capability.Capability) {
	svc := &accountSvc{
		dao: cap.Dao,
	}

	h := rest.NewHandler()

	h.Add("CreateAccount", "POST", "/vendor/{vendor}/account/create", svc.CreateAccount)
	h.Add("UpdateAccount", "PATCH", "/vendor/{vendor}/account/{account_id}", svc.UpdateAccount)
	h.Add("ListAccount", "POST", "/account/list", svc.ListAccount)
	h.Add("DeleteAccount", "DELETE", "/account", svc.DeleteAccount)

	h.Load(cap.WebService)
}

// TODO 考虑废弃 accountSvc 模式
type accountSvc struct {
	dao dao.Set
}

// CreateAccount account with options
func (svc *accountSvc) CreateAccount(cts *rest.Contexts) (interface{}, error) {
	vendor := enumor.Vendor(cts.Request.PathParameter("vendor"))
	if err := vendor.Validate(); err != nil {
		return nil, errf.Newf(errf.InvalidParameter, err.Error())
	}
	switch vendor {
	case enumor.TCloud:
		return createAccount[protocloud.CreateTCloudAccountExtensionReq](vendor, svc, cts)
	case enumor.AWS:
		return createAccount[protocloud.CreateAwsAccountExtensionReq](vendor, svc, cts)
	case enumor.HuaWei:
		return createAccount[protocloud.CreateHuaWeiAccountExtensionReq](vendor, svc, cts)
	case enumor.GCP:
		return createAccount[protocloud.CreateGcpAccountExtensionReq](vendor, svc, cts)
	case enumor.Azure:
		return createAccount[protocloud.CreateAzureAccountExtensionReq](vendor, svc, cts)
	}

	return nil, nil
}

func createAccount[T protocloud.CreateAccountExtensionReq](vendor enumor.Vendor, svc *accountSvc, cts *rest.Contexts) (interface{}, error) {
	req := new(protocloud.CreateAccountReq[T])
	if err := cts.DecodeInto(req); err != nil {
		return nil, errf.New(errf.DecodeRequestFailed, err.Error())
	}

	if err := req.Validate(); err != nil {
		return nil, errf.Newf(errf.InvalidParameter, err.Error())
	}

	accountID, err := svc.dao.Txn().AutoTxn(cts.Kit, func(txn *sqlx.Tx, opt *orm.TxnOption) (interface{}, error) {
		extensionJson, err := json.MarshalToString(req.Extension)
		if err != nil {
			return nil, errf.Newf(errf.InvalidParameter, err.Error())
		}

		account := &tablecloud.AccountTable{
			Vendor:       string(vendor),
			Name:         req.Spec.Name,
			Managers:     req.Spec.Managers,
			DepartmentID: req.Spec.DepartmentID,
			Type:         string(req.Spec.Type),
			Site:         string(req.Spec.Site),
			SyncStatus:   enumor.NotStart,
			Memo:         req.Spec.Memo,
			Extension:    tabletype.JsonField(extensionJson),
			Creator:      cts.Kit.User,
			Reviser:      cts.Kit.User,
		}

		accountID, err := svc.dao.Account().CreateWithTx(cts.Kit, txn, account)
		if err != nil {
			return nil, fmt.Errorf("create account failed, err: %v", err)
		}

		rels := make([]*tablecloud.AccountBizRelTable, len(req.Attachment.BkBizIDs))
		for index, bizID := range req.Attachment.BkBizIDs {
			rels[index] = &tablecloud.AccountBizRelTable{
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
	// TODO: Vendor和ID从Path 获取后并校验，可以通用化
	vendor := enumor.Vendor(cts.Request.PathParameter("vendor"))
	if err := vendor.Validate(); err != nil {
		return nil, errf.Newf(errf.InvalidParameter, err.Error())
	}

	accountID, err := strconv.ParseUint(cts.Request.PathParameter("account_id"), 10, 64)
	if err != nil {
		return nil, errf.Newf(errf.InvalidParameter, err.Error())
	}

	switch vendor {
	case enumor.TCloud:
		return updateAccount[protocloud.UpdateTCloudAccountExtensionReq](accountID, svc, cts)
	case enumor.AWS:
		return updateAccount[protocloud.UpdateAwsAccountExtensionReq](accountID, svc, cts)
	case enumor.HuaWei:
		return updateAccount[protocloud.UpdateHuaWeiAccountExtensionReq](accountID, svc, cts)
	case enumor.GCP:
		return updateAccount[protocloud.UpdateGcpAccountExtensionReq](accountID, svc, cts)
	case enumor.Azure:
		return updateAccount[protocloud.UpdateAzureAccountExtensionReq](accountID, svc, cts)
	}

	return nil, nil
}

func updateAccount[T protocloud.UpdateAccountExtensionReq](accountID uint64, svc *accountSvc, cts *rest.Contexts) (interface{}, error) {
	req := new(protocloud.UpdateAccountReq[T])

	if err := cts.DecodeInto(req); err != nil {
		return nil, errf.New(errf.DecodeRequestFailed, err.Error())
	}

	if err := req.Validate(); err != nil {
		return nil, errf.Newf(errf.InvalidParameter, err.Error())
	}

	// TODO: 这个ID条件比较通用，可以单独函数
	// 更新和查询的过滤条件：id=xxx
	idCondition := &filter.Expression{
		Op: filter.And,
		Rules: []filter.RuleFactory{
			filter.AtomRule{Field: "id", Op: filter.Equal.Factory(), Value: accountID},
		},
	}

	account := &tablecloud.AccountTable{
		Name:         req.Spec.Name,
		Managers:     req.Spec.Managers,
		DepartmentID: req.Spec.DepartmentID,
		SyncStatus:   req.Spec.SyncStatus,
		Price:        req.Spec.Price,
		PriceUnit:    req.Spec.PriceUnit,
		Memo:         req.Spec.Memo,
		Reviser:      cts.Kit.User,
	}

	// 只有提供了Extension才进行更新
	if req.Extension != nil {
		// TODO: 单独查询Extension逻辑是否封装为一个函数
		// 对于Extension，由于是Json值，需要取出来，对比是否变化了，变化了则更新
		opt := &types.ListOption{
			Filter: idCondition,
			Page:   &daotypes.BasePage{Count: false, Start: 0, Limit: 1},
		}
		listAccountDetails, err := svc.dao.Account().List(cts.Kit, opt)
		if err != nil {
			logs.Errorf("list account failed, err: %v, rid: %s", cts.Kit.Rid)
			return nil, fmt.Errorf("list account failed, err: %v", err)
		}
		details := listAccountDetails.Details
		if len(details) != 1 {
			return nil, fmt.Errorf("list account failed, account(id=%d) don't exist", accountID)
		}

		// 将新的Extension转为json数据
		extensionJson, err := json.MarshalToString(req.Extension)
		if err != nil {
			return nil, fmt.Errorf("MarshalToString req extension failed, err: %v", err)
		}
		// 合并覆盖dbExtension
		updatedExtension, err := json.UpdateMerge(extensionJson, string(details[0].Extension))
		if err != nil {
			return nil, fmt.Errorf("json UpdateMerge extension failed, err: %v", err)
		}
		
		account.Extension = tabletype.JsonField(updatedExtension)
	}

	err := svc.dao.Account().Update(cts.Kit, idCondition, account)
	if err != nil {
		logs.Errorf("update account failed, err: %v, rid: %s", cts.Kit.Rid)
		return nil, fmt.Errorf("update account failed, err: %v", err)
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
	daoAccountResp, err := svc.dao.Account().List(cts.Kit, opt)
	if err != nil {
		logs.Errorf("list account failed, err: %v, rid: %s", cts.Kit.Rid)
		return nil, fmt.Errorf("list account failed, err: %v", err)
	}
	if req.Page.Count {
		return &protocloud.ListAccountResult{Count: daoAccountResp.Count}, nil
	}

	details := make([]*protocloud.ListBaseAccountReq, 0, len(daoAccountResp.Details))
	for _, account := range daoAccountResp.Details {
		details = append(details, &protocloud.ListBaseAccountReq{
			ID:     account.ID,
			Vendor: enumor.Vendor(account.Vendor),
			Spec: &protocore.AccountSpec{
				Name:         account.Name,
				Managers:     account.Managers,
				DepartmentID: account.DepartmentID,
				Type:         enumor.AccountType(account.Type),
				Site:         enumor.AccountSiteType(account.Site),
				SyncStatus:   enumor.AccountSyncStatus(account.SyncStatus),
				Price:        account.Price,
				PriceUnit:    account.PriceUnit,
				Memo:         account.Memo,
			},
		})

	}

	return &protocloud.ListAccountResult{Details: details}, nil
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
					Op:    filter.In.Factory(),
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
