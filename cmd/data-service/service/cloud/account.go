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
	"hcm/pkg/cryptography"
	"hcm/pkg/dal/dao"
	"hcm/pkg/dal/dao/orm"
	"hcm/pkg/dal/dao/tools"
	"hcm/pkg/dal/dao/types"
	tablecloud "hcm/pkg/dal/table/cloud"
	tabletype "hcm/pkg/dal/table/types"
	"hcm/pkg/logs"
	"hcm/pkg/rest"
	"hcm/pkg/tools/converter"
	"hcm/pkg/tools/json"

	"github.com/jmoiron/sqlx"
)

// InitAccountService initial the account service
func InitAccountService(cap *capability.Capability) {
	svc := &accountSvc{
		dao:    cap.Dao,
		cipher: cap.Cipher,
	}

	h := rest.NewHandler()

	h.Add("CreateAccount", "POST", "/vendors/{vendor}/accounts/create", svc.CreateAccount)
	h.Add("UpdateAccount", "PATCH", "/vendors/{vendor}/accounts/{account_id}", svc.UpdateAccount)
	h.Add("GetAccount", "GET", "/vendors/{vendor}/accounts/{account_id}", svc.GetAccount)
	h.Add("ListAccount", "POST", "/accounts/list", svc.ListAccount)
	h.Add("ListAccountWithExtension", "POST", "/accounts/extensions/list", svc.ListAccountWithExtension)
	h.Add("DeleteAccount", "DELETE", "/accounts", svc.DeleteAccount)

	h.Add("UpdateAccountBizRel", "PUT", "/account_biz_rels/accounts/{account_id}", svc.UpdateAccountBizRel)
	h.Add("ListWithAccount", "POST", "/account_biz_rels/with/accounts/list", svc.ListWithAccount)

	h.Load(cap.WebService)
}

type accountSvc struct {
	dao    dao.Set
	cipher cryptography.Crypto
}

// CreateAccount account with options
func (a *accountSvc) CreateAccount(cts *rest.Contexts) (interface{}, error) {
	vendor := enumor.Vendor(cts.Request.PathParameter("vendor"))
	if err := vendor.Validate(); err != nil {
		return nil, errf.NewFromErr(errf.InvalidParameter, err)
	}
	switch vendor {
	case enumor.TCloud:
		return createAccount[protocloud.TCloudAccountExtensionCreateReq](vendor, a, cts)
	case enumor.Aws:
		return createAccount[protocloud.AwsAccountExtensionCreateReq](vendor, a, cts)
	case enumor.HuaWei:
		return createAccount[protocloud.HuaWeiAccountExtensionCreateReq](vendor, a, cts)
	case enumor.Gcp:
		return createAccount[protocloud.GcpAccountExtensionCreateReq](vendor, a, cts)
	case enumor.Azure:
		return createAccount[protocloud.AzureAccountExtensionCreateReq](vendor, a, cts)
	default:
		return nil, fmt.Errorf("unsupport %s vendor for now", vendor)
	}
}

func createAccount[T protocloud.AccountExtensionCreateReq, PT protocloud.SecretEncryptor[T]](vendor enumor.Vendor, svc *accountSvc, cts *rest.Contexts) (interface{}, error) {
	req := new(protocloud.AccountCreateReq[T])
	if err := cts.DecodeInto(req); err != nil {
		return nil, errf.NewFromErr(errf.DecodeRequestFailed, err)
	}

	if err := req.Validate(); err != nil {
		return nil, errf.NewFromErr(errf.InvalidParameter, err)
	}

	// 将参数里的SecretKey加密
	if req.Extension != nil {
		p := PT(req.Extension)
		// 加密密钥
		p.EncryptSecretKey(svc.cipher)
	}

	accountID, err := svc.dao.Txn().AutoTxn(cts.Kit, func(txn *sqlx.Tx, opt *orm.TxnOption) (interface{}, error) {
		extensionJson, err := json.MarshalToString(req.Extension)
		if err != nil {
			return nil, errf.NewFromErr(errf.InvalidParameter, err)
		}

		account := &tablecloud.AccountTable{
			Vendor:        string(vendor),
			Name:          req.Name,
			Managers:      req.Managers,
			DepartmentIDs: req.DepartmentIDs,
			Type:          string(req.Type),
			Site:          string(req.Site),
			SyncStatus:    enumor.NotStart,
			Memo:          req.Memo,
			Extension:     tabletype.JsonField(extensionJson),
			Creator:       cts.Kit.User,
			Reviser:       cts.Kit.User,
		}

		accountID, err := svc.dao.Account().CreateWithTx(cts.Kit, txn, account)
		if err != nil {
			return nil, fmt.Errorf("create account failed, err: %v", err)
		}

		rels := make([]*tablecloud.AccountBizRelTable, len(req.BkBizIDs))
		for index, bizID := range req.BkBizIDs {
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
		return nil, fmt.Errorf("create account but return id type not string, id type: %v",
			reflect.TypeOf(accountID).String())
	}

	return &core.CreateResult{ID: id}, nil
}

// UpdateAccount account with filter.
func (a *accountSvc) UpdateAccount(cts *rest.Contexts) (interface{}, error) {
	// TODO: Vendor和ID从Path 获取后并校验，可以通用化
	vendor := enumor.Vendor(cts.Request.PathParameter("vendor"))
	if err := vendor.Validate(); err != nil {
		return nil, errf.NewFromErr(errf.InvalidParameter, err)
	}

	accountID := cts.PathParameter("account_id").String()

	switch vendor {
	case enumor.TCloud:
		return updateAccount[protocloud.TCloudAccountExtensionUpdateReq](accountID, a, cts)
	case enumor.Aws:
		return updateAccount[protocloud.AwsAccountExtensionUpdateReq](accountID, a, cts)
	case enumor.HuaWei:
		return updateAccount[protocloud.HuaWeiAccountExtensionUpdateReq](accountID, a, cts)
	case enumor.Gcp:
		return updateAccount[protocloud.GcpAccountExtensionUpdateReq](accountID, a, cts)
	case enumor.Azure:
		return updateAccount[protocloud.AzureAccountExtensionUpdateReq](accountID, a, cts)
	}

	return nil, nil
}

func getAccountFromTable(accountID string, svc *accountSvc, cts *rest.Contexts) (*tablecloud.AccountTable, error) {
	opt := &types.ListOption{
		Filter: tools.EqualExpression("id", accountID),
		Page:   &core.BasePage{Count: false, Start: 0, Limit: 1},
	}
	listAccountDetails, err := svc.dao.Account().List(cts.Kit, opt)
	if err != nil {
		logs.Errorf("list account failed, err: %v, rid: %s", err, cts.Kit.Rid)
		return nil, fmt.Errorf("list account failed, err: %v", err)
	}
	details := listAccountDetails.Details
	if len(details) != 1 {
		return nil, fmt.Errorf("list account failed, account(id=%s) don't exist", accountID)
	}

	return details[0], nil
}

func updateAccount[T protocloud.AccountExtensionUpdateReq, PT protocloud.SecretEncryptor[T]](accountID string, svc *accountSvc, cts *rest.Contexts) (interface{}, error) {
	req := new(protocloud.AccountUpdateReq[T])

	if err := cts.DecodeInto(req); err != nil {
		return nil, errf.NewFromErr(errf.DecodeRequestFailed, err)
	}

	if err := req.Validate(); err != nil {
		return nil, errf.NewFromErr(errf.InvalidParameter, err)
	}

	account := &tablecloud.AccountTable{
		Name:          req.Name,
		Managers:      req.Managers,
		DepartmentIDs: req.DepartmentIDs,
		SyncStatus:    req.SyncStatus,
		Price:         req.Price,
		PriceUnit:     req.PriceUnit,
		Memo:          req.Memo,
		Reviser:       cts.Kit.User,
	}

	// 只有提供了Extension才进行更新
	if req.Extension != nil {
		// 将参数里的SecretKey加密
		p := PT(req.Extension)
		p.EncryptSecretKey(svc.cipher)

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
		logs.Errorf("update account failed, err: %v, rid: %s", err, cts.Kit.Rid)
		return nil, fmt.Errorf("update account failed, err: %v", err)
	}

	return nil, nil
}

func convertToAccountResult[T protocloud.AccountExtensionGetResp, PT protocloud.SecretDecryptor[T]](
	baseAccount *protocore.BaseAccount, dbExtension tabletype.JsonField, svc *accountSvc,
) (*protocloud.AccountGetResult[T], error) {
	extension := new(T)
	err := json.UnmarshalFromString(string(dbExtension), extension)
	if err != nil {
		return nil, fmt.Errorf("UnmarshalFromString db extension failed, err: %v", err)
	}

	// 解密密钥
	p := PT(extension)
	err = p.DecryptSecretKey(svc.cipher)
	if err != nil {
		return nil, fmt.Errorf("decrypt secret key of extension failed, err: %v", err)
	}

	return &protocloud.AccountGetResult[T]{
		BaseAccount: *baseAccount,
		Extension:   extension,
	}, nil
}

// GetAccount accounts with detail
func (a *accountSvc) GetAccount(cts *rest.Contexts) (interface{}, error) {
	// TODO: Vendor和ID从Path 获取后并校验，可以通用化
	vendor := enumor.Vendor(cts.Request.PathParameter("vendor"))
	if err := vendor.Validate(); err != nil {
		return nil, errf.NewFromErr(errf.InvalidParameter, err)
	}

	accountID := cts.PathParameter("account_id").String()

	// 查询账号信息
	dbAccount, err := getAccountFromTable(accountID, a, cts)
	if err != nil {
		return nil, err
	}

	// 查询账号关联信息，这里只有业务
	opt := &types.ListOption{
		Filter: tools.EqualExpression("account_id", accountID),
		// TODO：支持查询全量的Page
		Page: &core.BasePage{Start: 0, Limit: core.DefaultMaxPageLimit},
	}
	relResp, err := a.dao.AccountBizRel().List(cts.Kit, opt)
	if err != nil {
		return nil, err
	}
	bizIDs := make([]int64, 0, len(relResp.Details))
	for _, rel := range relResp.Details {
		bizIDs = append(bizIDs, rel.BkBizID)
	}

	// 组装响应数据 - 账号基本信息
	baseAccount := &protocore.BaseAccount{
		ID:            dbAccount.ID,
		Vendor:        enumor.Vendor(dbAccount.Vendor),
		Name:          dbAccount.Name,
		Managers:      dbAccount.Managers,
		DepartmentIDs: dbAccount.DepartmentIDs,
		Type:          enumor.AccountType(dbAccount.Type),
		Site:          enumor.AccountSiteType(dbAccount.Site),
		SyncStatus:    enumor.AccountSyncStatus(dbAccount.SyncStatus),
		Price:         dbAccount.Price,
		PriceUnit:     dbAccount.PriceUnit,
		Memo:          dbAccount.Memo,
		BkBizIDs:      bizIDs,
		Revision: core.Revision{
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
		account, err = convertToAccountResult[protocore.TCloudAccountExtension](baseAccount, dbAccount.Extension, a)
	case enumor.Aws:
		account, err = convertToAccountResult[protocore.AwsAccountExtension](baseAccount, dbAccount.Extension, a)
	case enumor.HuaWei:
		account, err = convertToAccountResult[protocore.HuaWeiAccountExtension](baseAccount, dbAccount.Extension, a)
	case enumor.Gcp:
		account, err = convertToAccountResult[protocore.GcpAccountExtension](baseAccount, dbAccount.Extension, a)
	case enumor.Azure:
		account, err = convertToAccountResult[protocore.AzureAccountExtension](baseAccount, dbAccount.Extension, a)
	}

	if err != nil {
		return nil, err
	}

	return account, nil
}

// ListAccount accounts with filter
func (a *accountSvc) ListAccount(cts *rest.Contexts) (interface{}, error) {
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
	daoAccountResp, err := a.dao.Account().List(cts.Kit, opt)
	if err != nil {
		logs.Errorf("list account failed, err: %v, rid: %s", err, cts.Kit.Rid)
		return nil, fmt.Errorf("list account failed, err: %v", err)
	}
	if req.Page.Count {
		return &protocloud.AccountListResult{Count: daoAccountResp.Count}, nil
	}

	details := make([]*protocloud.BaseAccountListResp, 0, len(daoAccountResp.Details))
	for _, account := range daoAccountResp.Details {
		details = append(details, &protocloud.BaseAccountListResp{
			ID:            account.ID,
			Vendor:        enumor.Vendor(account.Vendor),
			Name:          account.Name,
			Managers:      account.Managers,
			DepartmentIDs: account.DepartmentIDs,
			Type:          enumor.AccountType(account.Type),
			Site:          enumor.AccountSiteType(account.Site),
			SyncStatus:    enumor.AccountSyncStatus(account.SyncStatus),
			Price:         account.Price,
			PriceUnit:     account.PriceUnit,
			Memo:          account.Memo,
			Revision: core.Revision{
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
func (a *accountSvc) DeleteAccount(cts *rest.Contexts) (interface{}, error) {
	req := new(protocloud.AccountDeleteReq)
	if err := cts.DecodeInto(req); err != nil {
		return nil, err
	}

	if err := req.Validate(); err != nil {
		return nil, errf.NewFromErr(errf.InvalidParameter, err)
	}

	opt := &types.ListOption{
		Filter: req.Filter,
		Page:   core.DefaultBasePage,
	}
	listResp, err := a.dao.Account().List(cts.Kit, opt)
	if err != nil {
		logs.Errorf("list account failed, err: %v, rid: %s", err, cts.Kit.Rid)
		return nil, fmt.Errorf("list account failed, err: %v", err)
	}

	if len(listResp.Details) == 0 {
		return nil, nil
	}

	delAccountIDs := make([]string, len(listResp.Details))
	for index, one := range listResp.Details {
		delAccountIDs[index] = one.ID
	}

	_, err = a.dao.Txn().AutoTxn(cts.Kit, func(txn *sqlx.Tx, opt *orm.TxnOption) (interface{}, error) {
		delAccountFilter := tools.ContainersExpression("id", delAccountIDs)
		if err := a.dao.Account().DeleteWithTx(cts.Kit, txn, delAccountFilter); err != nil {
			return nil, err
		}

		delAccountBizRelFilter := tools.ContainersExpression("account_id", delAccountIDs)
		if err := a.dao.AccountBizRel().DeleteWithTx(cts.Kit, txn, delAccountBizRelFilter); err != nil {
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

func convertToAccountExtension[T protocloud.AccountExtensionGetResp, PT protocloud.SecretDecryptor[T]](
	dbExtension tabletype.JsonField, svc *accountSvc,
) (map[string]interface{}, error) {
	extension := new(T)
	err := json.UnmarshalFromString(string(dbExtension), extension)
	if err != nil {
		return nil, fmt.Errorf("UnmarshalFromString db extension failed, err: %v", err)
	}

	// 解密密钥
	p := PT(extension)
	err = p.DecryptSecretKey(svc.cipher)
	if err != nil {
		return nil, fmt.Errorf("decrypt secret key of extension failed, err: %v", err)
	}

	return converter.StructToMap(extension)
}

// ListAccountWithExtension accounts with extension by filter
func (a *accountSvc) ListAccountWithExtension(cts *rest.Contexts) (interface{}, error) {
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
	daoAccountResp, err := a.dao.Account().List(cts.Kit, opt)
	if err != nil {
		logs.Errorf("list account with extension failed, err: %v, rid: %s", err, cts.Kit.Rid)
		return nil, fmt.Errorf("list account with extension failed, err: %v", err)
	}
	if req.Page.Count {
		return &protocloud.AccountWithExtensionListResult{Count: daoAccountResp.Count}, nil
	}

	details := make([]*protocloud.BaseAccountWithExtensionListResp, 0, len(daoAccountResp.Details))
	for _, account := range daoAccountResp.Details {
		var extension map[string]interface{}
		switch enumor.Vendor(account.Vendor) {
		case enumor.TCloud:
			extension, err = convertToAccountExtension[protocore.TCloudAccountExtension](account.Extension, a)
		case enumor.Aws:
			extension, err = convertToAccountExtension[protocore.AwsAccountExtension](account.Extension, a)
		case enumor.HuaWei:
			extension, err = convertToAccountExtension[protocore.HuaWeiAccountExtension](account.Extension, a)
		case enumor.Gcp:
			extension, err = convertToAccountExtension[protocore.GcpAccountExtension](account.Extension, a)
		case enumor.Azure:
			extension, err = convertToAccountExtension[protocore.AzureAccountExtension](account.Extension, a)
		}
		if err != nil {
			return nil, fmt.Errorf("json unmarshal extension to vendor extension failed, err: %v", err)
		}

		details = append(details, &protocloud.BaseAccountWithExtensionListResp{
			BaseAccountListResp: protocloud.BaseAccountListResp{
				ID:            account.ID,
				Vendor:        enumor.Vendor(account.Vendor),
				Name:          account.Name,
				Managers:      account.Managers,
				DepartmentIDs: account.DepartmentIDs,
				Type:          enumor.AccountType(account.Type),
				Site:          enumor.AccountSiteType(account.Site),
				SyncStatus:    enumor.AccountSyncStatus(account.SyncStatus),
				Price:         account.Price,
				PriceUnit:     account.PriceUnit,
				Memo:          account.Memo,
				Revision: core.Revision{
					Creator:   account.Creator,
					Reviser:   account.Reviser,
					CreatedAt: account.CreatedAt,
					UpdatedAt: account.UpdatedAt,
				},
			},
			Extension: extension,
		})
	}

	return &protocloud.AccountWithExtensionListResult{Details: details}, nil
}
