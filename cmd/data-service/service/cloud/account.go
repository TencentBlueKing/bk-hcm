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
	tableaudit "hcm/pkg/dal/table/audit"
	tablecloud "hcm/pkg/dal/table/cloud"
	tabletype "hcm/pkg/dal/table/types"
	"hcm/pkg/kit"
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
	h.Add("DeleteValidate", "POST", "/accounts/{account_id}/delete/validate", svc.DeleteValidate)

	h.Add("UpdateAccountBizRel", "PUT", "/account_biz_rels/accounts/{account_id}", svc.UpdateAccountBizRel)
	h.Add("ListAccountBizRel", "POST", "/account_biz_rels/list", svc.ListAccountBizRel)
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
			Vendor:    string(vendor),
			Name:      req.Name,
			Managers:  req.Managers,
			Type:      string(req.Type),
			Site:      string(req.Site),
			Memo:      req.Memo,
			Extension: tabletype.JsonField(extensionJson),
			Creator:   cts.Kit.User,
			Reviser:   cts.Kit.User,
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
		Name:      req.Name,
		Managers:  req.Managers,
		Price:     req.Price,
		PriceUnit: req.PriceUnit,
		Memo:      req.Memo,
		Reviser:   cts.Kit.User,
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
		ID:        dbAccount.ID,
		Vendor:    enumor.Vendor(dbAccount.Vendor),
		Name:      dbAccount.Name,
		Managers:  dbAccount.Managers,
		Type:      enumor.AccountType(dbAccount.Type),
		Site:      enumor.AccountSiteType(dbAccount.Site),
		Price:     dbAccount.Price,
		PriceUnit: dbAccount.PriceUnit,
		Memo:      dbAccount.Memo,
		BkBizIDs:  bizIDs,
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
			ID:        account.ID,
			Vendor:    enumor.Vendor(account.Vendor),
			Name:      account.Name,
			Managers:  account.Managers,
			Type:      enumor.AccountType(account.Type),
			Site:      enumor.AccountSiteType(account.Site),
			Price:     account.Price,
			PriceUnit: account.PriceUnit,
			Memo:      account.Memo,
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

// ListAccountWithBiz 查询账号列表带业务ID列表，但extension中没有密钥相关信息。
func (a *accountSvc) ListAccountWithBiz(kt *kit.Kit, ids []string) ([]types.Account, error) {

	if len(ids) > int(core.DefaultMaxPageLimit) {
		return nil, fmt.Errorf("ids shuold <= %d", core.DefaultMaxPageLimit)
	}

	listOpt := &types.ListOption{
		Filter: tools.ContainersExpression("id", ids),
		Page:   core.DefaultBasePage,
	}
	result, err := a.dao.Account().List(kt, listOpt)
	if err != nil {
		logs.Errorf("list account failed, err: %v, rid: %s", err, kt.Rid)
		return nil, err
	}

	listBizOpt := &types.ListOption{
		Filter: tools.ContainersExpression("account_id", ids),
		Page:   core.DefaultBasePage,
	}
	bizResult, err := a.dao.AccountBizRel().List(kt, listBizOpt)
	if err != nil {
		logs.Errorf("list account biz rel failed, err: %v, rid: %s", err, kt.Rid)
		return nil, err
	}

	accountBizMap := make(map[string][]int64, 0)
	for _, one := range bizResult.Details {
		if _, exist := accountBizMap[one.AccountID]; !exist {
			accountBizMap[one.AccountID] = make([]int64, 0)
		}

		accountBizMap[one.AccountID] = append(accountBizMap[one.AccountID], one.BkBizID)
	}

	accounts := make([]types.Account, 0, len(result.Details))
	for _, one := range result.Details {
		bizIDs := make([]int64, 0)
		if list, exist := accountBizMap[one.ID]; exist {
			bizIDs = list
		}

		extension := tools.AccountExtensionRemoveSecretKey(string(one.Extension))

		accounts = append(accounts, types.Account{
			AccountTable: tablecloud.AccountTable{
				ID:        one.ID,
				Name:      one.Name,
				Vendor:    one.Vendor,
				Managers:  one.Managers,
				Type:      one.Type,
				Site:      one.Site,
				Price:     one.Price,
				PriceUnit: one.PriceUnit,
				Extension: tabletype.JsonField(extension),
				TenantID:  one.TenantID,
				Creator:   one.Creator,
				Reviser:   one.Reviser,
				CreatedAt: one.CreatedAt,
				UpdatedAt: one.UpdatedAt,
				Memo:      one.Memo,
			},
			BkBizIDs: bizIDs,
		})
	}

	return accounts, nil
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
		// 校验账号下是否还有资源存在
		_, err = a.dao.Account().DeleteValidate(cts.Kit, one.ID)
		if err != nil {
			return nil, err
		}

		delAccountIDs[index] = one.ID
	}

	_, err = a.dao.Txn().AutoTxn(cts.Kit, func(txn *sqlx.Tx, opt *orm.TxnOption) (interface{}, error) {
		accounts, err := a.ListAccountWithBiz(cts.Kit, delAccountIDs)
		if err != nil {
			return nil, err
		}

		delAccountFilter := tools.ContainersExpression("id", delAccountIDs)
		if err := a.dao.Account().DeleteWithTx(cts.Kit, txn, delAccountFilter); err != nil {
			return nil, err
		}

		delAccountBizRelFilter := tools.ContainersExpression("account_id", delAccountIDs)
		if err := a.dao.AccountBizRel().DeleteWithTx(cts.Kit, txn, delAccountBizRelFilter); err != nil {
			return nil, err
		}

		// create audit
		audits := make([]*tableaudit.AuditTable, 0, len(accounts))
		for _, one := range accounts {
			extension := tools.AccountExtensionRemoveSecretKey(string(one.Extension))
			one.Extension = tabletype.JsonField(extension)

			audits = append(audits, &tableaudit.AuditTable{
				ResID:      one.ID,
				CloudResID: "",
				ResName:    one.Name,
				ResType:    enumor.AccountAuditResType,
				Action:     enumor.Delete,
				BkBizID:    0,
				Vendor:     enumor.Vendor(one.Vendor),
				AccountID:  one.ID,
				Operator:   cts.Kit.User,
				Source:     cts.Kit.GetRequestSource(),
				Rid:        cts.Kit.Rid,
				AppCode:    cts.Kit.AppCode,
				Detail: &tableaudit.BasicDetail{
					Data: one,
				},
			})
		}
		if err = a.dao.Audit().BatchCreate(cts.Kit, audits); err != nil {
			logs.Errorf("batch create audit failed, err: %v, rid: %s", err, cts.Kit.Rid)
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
				ID:        account.ID,
				Vendor:    enumor.Vendor(account.Vendor),
				Name:      account.Name,
				Managers:  account.Managers,
				Type:      enumor.AccountType(account.Type),
				Site:      enumor.AccountSiteType(account.Site),
				Price:     account.Price,
				PriceUnit: account.PriceUnit,
				Memo:      account.Memo,
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

// DeleteValidate account delete validate.
func (a *accountSvc) DeleteValidate(cts *rest.Contexts) (interface{}, error) {
	accountID := cts.PathParameter("account_id").String()
	if len(accountID) == 0 {
		return nil, errf.New(errf.InvalidParameter, "account_id is required")
	}

	validateResult, err := a.dao.Account().DeleteValidate(cts.Kit, accountID)
	if err != nil {
		return validateResult, err
	}

	return nil, nil
}
