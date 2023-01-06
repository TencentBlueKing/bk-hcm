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
	protocore "hcm/pkg/api/core/cloud"
	protocloud "hcm/pkg/api/data-service/cloud"
	"hcm/pkg/criteria/enumor"
	"hcm/pkg/criteria/errf"
	"hcm/pkg/dal/dao"
	"hcm/pkg/dal/dao/orm"
	"hcm/pkg/dal/dao/tools"
	"hcm/pkg/dal/dao/types"
	daotypes "hcm/pkg/dal/dao/types"
	tablecloud "hcm/pkg/dal/table/cloud"
	tabletype "hcm/pkg/dal/table/types"
	"hcm/pkg/logs"
	"hcm/pkg/rest"
	"hcm/pkg/tools/json"

	"github.com/jmoiron/sqlx"
)

// InitAccountService initial the account service
func InitAccountService(cap *capability.Capability) {
	svc := &accountSvc{
		dao: cap.Dao,
	}

	h := rest.NewHandler()

	h.Add("CreateAccount", "POST", "/vendors/{vendor}/accounts/create", svc.CreateAccount)
	h.Add("UpdateAccount", "PATCH", "/vendors/{vendor}/accounts/{account_id}", svc.UpdateAccount)
	h.Add("GetAccount", "GET", "/vendors/{vendor}/accounts/{account_id}", svc.GetAccount)
	h.Add("ListAccount", "POST", "/accounts/list", svc.ListAccount)
	h.Add("DeleteAccount", "DELETE", "/accounts", svc.DeleteAccount)
	h.Add("UpdateAccountBizRel", "PUT", "/account_biz_rels/accounts/{account_id}", svc.UpdateAccountBizRel)

	h.Load(cap.WebService)
}

type accountSvc struct {
	dao dao.Set
}

// CreateAccount account with options
func (svc *accountSvc) CreateAccount(cts *rest.Contexts) (interface{}, error) {
	vendor := enumor.Vendor(cts.Request.PathParameter("vendor"))
	if err := vendor.Validate(); err != nil {
		return nil, errf.NewFromErr(errf.InvalidParameter, err)
	}
	switch vendor {
	case enumor.TCloud:
		return createAccount[protocloud.TCloudAccountExtensionCreateReq](vendor, svc, cts)
	case enumor.Aws:
		return createAccount[protocloud.AwsAccountExtensionCreateReq](vendor, svc, cts)
	case enumor.HuaWei:
		return createAccount[protocloud.HuaWeiAccountExtensionCreateReq](vendor, svc, cts)
	case enumor.Gcp:
		return createAccount[protocloud.GcpAccountExtensionCreateReq](vendor, svc, cts)
	case enumor.Azure:
		return createAccount[protocloud.AzureAccountExtensionCreateReq](vendor, svc, cts)
	}

	return nil, nil
}

func createAccount[T protocloud.AccountExtensionCreateReq](vendor enumor.Vendor, svc *accountSvc, cts *rest.Contexts) (interface{}, error) {
	req := new(protocloud.AccountCreateReq[T])
	if err := cts.DecodeInto(req); err != nil {
		return nil, errf.NewFromErr(errf.DecodeRequestFailed, err)
	}

	if err := req.Validate(); err != nil {
		return nil, errf.NewFromErr(errf.InvalidParameter, err)
	}

	accountID, err := svc.dao.Txn().AutoTxn(cts.Kit, func(txn *sqlx.Tx, opt *orm.TxnOption) (interface{}, error) {
		extensionJson, err := json.MarshalToString(req.Extension)
		if err != nil {
			return nil, errf.NewFromErr(errf.InvalidParameter, err)
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
		err = svc.dao.AccountBizRel().BatchCreateWithTx(cts.Kit, txn, rels)
		if err != nil {
			return nil, fmt.Errorf("batch create account_biz_rels failed, err: %v", err)
		}

		return accountID, nil
	})
	if err != nil {
		return nil, err
	}

	id, ok := accountID.(string)
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
		return nil, errf.NewFromErr(errf.InvalidParameter, err)
	}

	accountID := cts.PathParameter("account_id").String()

	switch vendor {
	case enumor.TCloud:
		return updateAccount[protocloud.TCloudAccountExtensionUpdateReq](accountID, svc, cts)
	case enumor.Aws:
		return updateAccount[protocloud.AwsAccountExtensionUpdateReq](accountID, svc, cts)
	case enumor.HuaWei:
		return updateAccount[protocloud.HuaWeiAccountExtensionUpdateReq](accountID, svc, cts)
	case enumor.Gcp:
		return updateAccount[protocloud.GcpAccountExtensionUpdateReq](accountID, svc, cts)
	case enumor.Azure:
		return updateAccount[protocloud.AzureAccountExtensionUpdateReq](accountID, svc, cts)
	}

	return nil, nil
}

func getAccountFromTable(accountID string, svc *accountSvc, cts *rest.Contexts) (*tablecloud.AccountTable, error) {
	opt := &types.ListOption{
		Filter: tools.EqualExpression("id", accountID),
		Page:   &daotypes.BasePage{Count: false, Start: 0, Limit: 1},
	}
	listAccountDetails, err := svc.dao.Account().List(cts.Kit, opt)
	if err != nil {
		logs.Errorf("list account failed, err: %v, rid: %s", cts.Kit.Rid)
		return nil, fmt.Errorf("list account failed, err: %v", err)
	}
	details := listAccountDetails.Details
	if len(details) != 1 {
		return nil, fmt.Errorf("list account failed, account(id=%s) don't exist", accountID)
	}

	return details[0], nil
}

func updateAccount[T protocloud.AccountExtensionUpdateReq](accountID string, svc *accountSvc, cts *rest.Contexts) (interface{}, error) {
	req := new(protocloud.AccountUpdateReq[T])

	if err := cts.DecodeInto(req); err != nil {
		return nil, errf.NewFromErr(errf.DecodeRequestFailed, err)
	}

	if err := req.Validate(); err != nil {
		return nil, errf.NewFromErr(errf.InvalidParameter, err)
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
		// 查询账号
		dbAccount, err := getAccountFromTable(accountID, svc, cts)
		if err != nil {
			return nil, err
		}

		// 合并覆盖dbExtension
		updatedExtension, err := json.UpdateMerge(req.Extension, string(dbAccount.Extension))
		if err != nil {
			return nil, fmt.Errorf("json UpdateMerge extension failed, err: %v", err)
		}

		account.Extension = tabletype.JsonField(updatedExtension)
	}

	err := svc.dao.Account().Update(cts.Kit, tools.EqualExpression("id", accountID), account)
	if err != nil {
		logs.Errorf("update account failed, err: %v, rid: %s", cts.Kit.Rid)
		return nil, fmt.Errorf("update account failed, err: %v", err)
	}

	return nil, nil
}

func convertToAccountResult[T protocloud.AccountExtensionGetResp](baseAccount *protocore.BaseAccount, dbExtension tabletype.JsonField) (*protocloud.AccountGetResult[T], error) {
	extension := new(T)
	err := json.UnmarshalFromString(string(dbExtension), extension)
	if err != nil {
		return nil, fmt.Errorf("UnmarshalFromString db extension failed, err: %v", err)
	}
	return &protocloud.AccountGetResult[T]{
		BaseAccount: *baseAccount,
		Extension:   extension,
	}, nil
}

// GetAccount accounts with detail
func (svc *accountSvc) GetAccount(cts *rest.Contexts) (interface{}, error) {
	// TODO: Vendor和ID从Path 获取后并校验，可以通用化
	vendor := enumor.Vendor(cts.Request.PathParameter("vendor"))
	if err := vendor.Validate(); err != nil {
		return nil, errf.NewFromErr(errf.InvalidParameter, err)
	}

	accountID := cts.PathParameter("account_id").String()

	// 查询账号信息
	dbAccount, err := getAccountFromTable(accountID, svc, cts)
	if err != nil {
		return nil, err
	}

	// 查询账号关联信息，这里只有业务
	opt := &types.ListOption{
		Filter: tools.EqualExpression("account_id", accountID),
		// TODO：支持查询全量的Page
		Page: &types.BasePage{Start: 0, Limit: types.DefaultMaxPageLimit},
	}
	relResp, err := svc.dao.AccountBizRel().List(cts.Kit, opt)
	if err != nil {
		return nil, err
	}
	bizIDs := make([]int64, 0, len(relResp.Details))
	for _, rel := range relResp.Details {
		bizIDs = append(bizIDs, rel.BkBizID)
	}

	// 组装响应数据 - 账号基本信息
	baseAccount := &protocore.BaseAccount{
		ID:     dbAccount.ID,
		Vendor: enumor.Vendor(dbAccount.Vendor),
		Spec: &protocore.AccountSpec{
			Name:         dbAccount.Name,
			Managers:     dbAccount.Managers,
			DepartmentID: dbAccount.DepartmentID,
			Type:         enumor.AccountType(dbAccount.Type),
			Site:         enumor.AccountSiteType(dbAccount.Site),
			SyncStatus:   enumor.AccountSyncStatus(dbAccount.SyncStatus),
			Price:        dbAccount.Price,
			PriceUnit:    dbAccount.PriceUnit,
			Memo:         dbAccount.Memo,
		},
		Attachment: &protocore.AccountAttachment{
			BkBizIDs: bizIDs,
		},
		Revision: &core.Revision{
			Creator:   dbAccount.Creator,
			Reviser:   dbAccount.Reviser,
			CreatedAt: dbAccount.CreatedAt,
			UpdatedAt: dbAccount.UpdatedAt,
		},
	}

	// 转换为最终的数据结构
	var account interface{}
	switch enumor.Vendor(dbAccount.Vendor) {
	case enumor.TCloud:
		account, err = convertToAccountResult[protocore.TCloudAccountExtension](baseAccount, dbAccount.Extension)
	case enumor.Aws:
		account, err = convertToAccountResult[protocore.AwsAccountExtension](baseAccount, dbAccount.Extension)
	case enumor.HuaWei:
		account, err = convertToAccountResult[protocore.HuaWeiAccountExtension](baseAccount, dbAccount.Extension)
	case enumor.Gcp:
		account, err = convertToAccountResult[protocore.GcpAccountExtension](baseAccount, dbAccount.Extension)
	case enumor.Azure:
		account, err = convertToAccountResult[protocore.AzureAccountExtension](baseAccount, dbAccount.Extension)
	}

	if err != nil {
		return nil, err
	}

	return account, nil
}

// ListAccount accounts with filter
func (svc *accountSvc) ListAccount(cts *rest.Contexts) (interface{}, error) {
	req := new(protocloud.AccountListReq)
	if err := cts.DecodeInto(req); err != nil {
		return nil, err
	}

	if err := req.Validate(); err != nil {
		return nil, errf.NewFromErr(errf.InvalidParameter, err)
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
		return &protocloud.AccountListResult{Count: daoAccountResp.Count}, nil
	}

	details := make([]*protocloud.BaseAccountListResp, 0, len(daoAccountResp.Details))
	for _, account := range daoAccountResp.Details {
		details = append(details, &protocloud.BaseAccountListResp{
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
			Revision: &core.Revision{
				Creator:   account.Creator,
				Reviser:   account.Reviser,
				CreatedAt: account.CreatedAt,
				UpdatedAt: account.UpdatedAt,
			},
		})

	}

	return &protocloud.AccountListResult{Details: details}, nil
}

// DeleteAccount account with filter.
func (svc *accountSvc) DeleteAccount(cts *rest.Contexts) (interface{}, error) {
	req := new(protocloud.AccountDeleteReq)
	if err := cts.DecodeInto(req); err != nil {
		return nil, err
	}

	if err := req.Validate(); err != nil {
		return nil, errf.NewFromErr(errf.InvalidParameter, err)
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

	delAccountIDs := make([]string, len(listResp.Details))
	for index, one := range listResp.Details {
		delAccountIDs[index] = one.ID
	}

	_, err = svc.dao.Txn().AutoTxn(cts.Kit, func(txn *sqlx.Tx, opt *orm.TxnOption) (interface{}, error) {
		delAccountFilter := tools.ContainersExpression("id", delAccountIDs)
		if err := svc.dao.Account().DeleteWithTx(cts.Kit, txn, delAccountFilter); err != nil {
			return nil, err
		}

		delAccountBizRelFilter := tools.ContainersExpression("account_id", delAccountIDs)
		if err := svc.dao.AccountBizRel().DeleteWithTx(cts.Kit, txn, delAccountBizRelFilter); err != nil {
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

// UpdateAccountBizRel update account biz rel.
func (svc *accountSvc) UpdateAccountBizRel(cts *rest.Contexts) (interface{}, error) {
	accountID := cts.PathParameter("account_id").String()

	req := new(protocloud.AccountBizRelUpdateReq)
	if err := cts.DecodeInto(req); err != nil {
		return nil, err
	}

	if err := req.Validate(); err != nil {
		return nil, errf.NewFromErr(errf.InvalidParameter, err)
	}

	_, err := svc.dao.Txn().AutoTxn(cts.Kit, func(txn *sqlx.Tx, opt *orm.TxnOption) (interface{}, error) {
		ftr := tools.EqualExpression("account_id", accountID)
		if err := svc.dao.AccountBizRel().DeleteWithTx(cts.Kit, txn, ftr); err != nil {
			return nil, fmt.Errorf("delete account_biz_rels failed, err: %v", err)
		}

		rels := make([]*tablecloud.AccountBizRelTable, len(req.BkBizIDs))
		for index, bizID := range req.BkBizIDs {
			rels[index] = &tablecloud.AccountBizRelTable{
				BkBizID:   bizID,
				AccountID: accountID,
				Creator:   cts.Kit.User,
			}
		}
		if err := svc.dao.AccountBizRel().BatchCreateWithTx(cts.Kit, txn, rels); err != nil {
			return nil, fmt.Errorf("batch create account_biz_rels failed, err: %v", err)
		}

		return nil, nil
	})
	if err != nil {
		return nil, err
	}

	return nil, err
}
