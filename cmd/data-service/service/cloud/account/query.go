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

package account

import (
	"fmt"

	"hcm/pkg/api/core"
	protocore "hcm/pkg/api/core/cloud"
	protocloud "hcm/pkg/api/data-service/cloud"
	"hcm/pkg/criteria/enumor"
	"hcm/pkg/criteria/errf"
	"hcm/pkg/dal/dao/tools"
	"hcm/pkg/dal/dao/types"
	tablecloud "hcm/pkg/dal/table/cloud"
	tabletype "hcm/pkg/dal/table/types"
	"hcm/pkg/kit"
	"hcm/pkg/logs"
	"hcm/pkg/rest"
	"hcm/pkg/runtime/filter"
	"hcm/pkg/tools/converter"
	"hcm/pkg/tools/json"
)

func convertToAccountResult[T protocloud.AccountExtensionGetResp, PT protocloud.SecretDecryptor[T]](
	baseAccount *protocore.BaseAccount, dbExtension tabletype.JsonField, svc *service,
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
func (svc *service) GetAccount(cts *rest.Contexts) (interface{}, error) {
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
		Page: &core.BasePage{Start: 0, Limit: core.DefaultMaxPageLimit},
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
		ID:                 dbAccount.ID,
		Vendor:             enumor.Vendor(dbAccount.Vendor),
		Name:               dbAccount.Name,
		Managers:           dbAccount.Managers,
		Type:               enumor.AccountType(dbAccount.Type),
		Site:               enumor.AccountSiteType(dbAccount.Site),
		Price:              dbAccount.Price,
		PriceUnit:          dbAccount.PriceUnit,
		Memo:               dbAccount.Memo,
		BkBizIDs:           bizIDs,
		RecycleReserveTime: dbAccount.RecycleReserveTime,
		Revision: core.Revision{
			Creator:   dbAccount.Creator,
			Reviser:   dbAccount.Reviser,
			CreatedAt: dbAccount.CreatedAt.String(),
			UpdatedAt: dbAccount.UpdatedAt.String(),
		},
	}

	// 转换为最终的数据结构
	var account interface{}
	switch enumor.Vendor(dbAccount.Vendor) {
	case enumor.TCloud:
		account, err = convertToAccountResult[protocore.TCloudAccountExtension](baseAccount, dbAccount.Extension, svc)
	case enumor.Aws:
		account, err = convertToAccountResult[protocore.AwsAccountExtension](baseAccount, dbAccount.Extension, svc)
	case enumor.HuaWei:
		account, err = convertToAccountResult[protocore.HuaWeiAccountExtension](baseAccount, dbAccount.Extension, svc)
	case enumor.Gcp:
		account, err = convertToAccountResult[protocore.GcpAccountExtension](baseAccount, dbAccount.Extension, svc)
	case enumor.Azure:
		account, err = convertToAccountResult[protocore.AzureAccountExtension](baseAccount, dbAccount.Extension, svc)
	case enumor.Other:
		account, err = convertToAccountResult[protocore.OtherAccountExtension](baseAccount, dbAccount.Extension, svc)

	}

	if err != nil {
		return nil, err
	}

	return account, nil
}

// ListAccount accounts with filter
func (svc *service) ListAccount(cts *rest.Contexts) (interface{}, error) {
	req := new(protocloud.AccountListReq)
	if err := cts.DecodeInto(req); err != nil {
		return nil, err
	}

	if err := req.Validate(); err != nil {
		return nil, errf.NewFromErr(errf.InvalidParameter, err)
	}

	opt := &types.ListOption{
		Fields: req.Fields,
		Filter: req.Filter,
		Page:   req.Page,
	}
	daoAccountResp, err := svc.dao.Account().List(cts.Kit, opt)
	if err != nil {
		logs.Errorf("list account failed, err: %v, rid: %s", err, cts.Kit.Rid)
		return nil, fmt.Errorf("list account failed, err: %v", err)
	}
	if req.Page.Count {
		return &protocloud.AccountListResult{Count: daoAccountResp.Count}, nil
	}

	ids := make([]string, 0, len(daoAccountResp.Details))
	details := make([]*protocore.BaseAccount, 0, len(daoAccountResp.Details))
	for _, account := range daoAccountResp.Details {
		ids = append(ids, account.ID)
		details = append(details, &protocore.BaseAccount{
			ID:                 account.ID,
			Vendor:             enumor.Vendor(account.Vendor),
			Name:               account.Name,
			Managers:           account.Managers,
			Type:               enumor.AccountType(account.Type),
			Site:               enumor.AccountSiteType(account.Site),
			Price:              account.Price,
			PriceUnit:          account.PriceUnit,
			Memo:               account.Memo,
			RecycleReserveTime: account.RecycleReserveTime,
			Revision: core.Revision{
				Creator:   account.Creator,
				Reviser:   account.Reviser,
				CreatedAt: account.CreatedAt.String(),
				UpdatedAt: account.UpdatedAt.String(),
			},
		})
	}

	// 查询账号业务信息，并赋值
	accountBizMap, err := svc.getAccountBizMap(cts.Kit, ids)
	if err != nil {
		return nil, err
	}

	for _, one := range details {
		one.BkBizIDs = accountBizMap[one.ID]
	}

	return &protocloud.AccountListResult{Details: details}, nil
}

// getAccountBizMap 获取账号和业务的映射关系
func (svc *service) getAccountBizMap(kt *kit.Kit, accountIDs []string) (map[string][]int64, error) {
	if len(accountIDs) == 0 {
		return make(map[string][]int64), nil
	}

	result := make(map[string][]int64)
	start := uint32(0)
	listOpt := &types.ListOption{
		Filter: &filter.Expression{
			Op: filter.And,
			Rules: []filter.RuleFactory{
				&filter.AtomRule{
					Field: "account_id",
					Op:    filter.In.Factory(),
					Value: accountIDs,
				},
			},
		},
		Page: &core.BasePage{
			Start: 0,
			Limit: core.DefaultMaxPageLimit,
		},
	}
	for {
		listOpt.Page.Start = start

		list, err := svc.dao.AccountBizRel().List(kt, listOpt)
		if err != nil {
			logs.Errorf("list account biz rel failed, err: %v, ids: %v, rid: %s", err, accountIDs, kt.Rid)
			return nil, err
		}

		for _, one := range list.Details {
			if _, exist := result[one.AccountID]; !exist {
				result[one.AccountID] = make([]int64, 0)
			}

			result[one.AccountID] = append(result[one.AccountID], one.BkBizID)
		}

		if len(list.Details) < int(core.DefaultMaxPageLimit) {
			break
		}

		start += uint32(core.DefaultMaxPageLimit)
	}

	return result, nil
}

// ListAccountWithBiz 查询账号列表带业务ID列表，但extension中没有密钥相关信息。
func (svc *service) ListAccountWithBiz(kt *kit.Kit, ids []string) ([]types.Account, error) {

	if len(ids) > int(core.DefaultMaxPageLimit) {
		return nil, fmt.Errorf("ids shuold <= %d", core.DefaultMaxPageLimit)
	}

	listOpt := &types.ListOption{
		Filter: tools.ContainersExpression("id", ids),
		Page:   core.NewDefaultBasePage(),
	}
	result, err := svc.dao.Account().List(kt, listOpt)
	if err != nil {
		logs.Errorf("list account failed, err: %v, rid: %s", err, kt.Rid)
		return nil, err
	}

	accountBizMap, err := svc.getAccountBizMap(kt, ids)
	if err != nil {
		return nil, err
	}

	accounts := make([]types.Account, 0, len(result.Details))
	for _, one := range result.Details {
		extension := tools.AccountExtensionRemoveSecretKey(string(one.Extension))

		accounts = append(accounts, types.Account{
			AccountTable: tablecloud.AccountTable{
				ID:                 one.ID,
				Name:               one.Name,
				Vendor:             one.Vendor,
				Managers:           one.Managers,
				Type:               one.Type,
				Site:               one.Site,
				Price:              one.Price,
				PriceUnit:          one.PriceUnit,
				Extension:          tabletype.JsonField(extension),
				TenantID:           one.TenantID,
				Creator:            one.Creator,
				Reviser:            one.Reviser,
				CreatedAt:          one.CreatedAt,
				UpdatedAt:          one.UpdatedAt,
				Memo:               one.Memo,
				RecycleReserveTime: one.RecycleReserveTime,
			},
			BkBizIDs: accountBizMap[one.ID],
		})
	}

	return accounts, nil
}

func convertToAccountExtension[T protocloud.AccountExtensionGetResp, PT protocloud.SecretDecryptor[T]](
	dbExtension tabletype.JsonField, svc *service) (map[string]interface{}, error) {

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
func (svc *service) ListAccountWithExtension(cts *rest.Contexts) (interface{}, error) {
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
		logs.Errorf("list account with extension failed, err: %v, rid: %s", err, cts.Kit.Rid)
		return nil, fmt.Errorf("list account with extension failed, err: %v", err)
	}
	if req.Page.Count {
		return &protocloud.AccountWithExtensionListResult{Count: daoAccountResp.Count}, nil
	}

	ids := make([]string, 0, len(daoAccountResp.Details))
	details := make([]*protocloud.BaseAccountWithExtensionListResp, 0, len(daoAccountResp.Details))
	for _, account := range daoAccountResp.Details {
		var extension map[string]interface{}
		switch enumor.Vendor(account.Vendor) {
		case enumor.TCloud:
			extension, err = convertToAccountExtension[protocore.TCloudAccountExtension](account.Extension, svc)
		case enumor.Aws:
			extension, err = convertToAccountExtension[protocore.AwsAccountExtension](account.Extension, svc)
		case enumor.HuaWei:
			extension, err = convertToAccountExtension[protocore.HuaWeiAccountExtension](account.Extension, svc)
		case enumor.Gcp:
			extension, err = convertToAccountExtension[protocore.GcpAccountExtension](account.Extension, svc)
		case enumor.Azure:
			extension, err = convertToAccountExtension[protocore.AzureAccountExtension](account.Extension, svc)
		}
		if err != nil {
			return nil, fmt.Errorf("json unmarshal extension to vendor extension failed, err: %v", err)
		}

		ids = append(ids, account.ID)
		details = append(details, &protocloud.BaseAccountWithExtensionListResp{
			BaseAccount: protocore.BaseAccount{
				ID:                 account.ID,
				Vendor:             enumor.Vendor(account.Vendor),
				Name:               account.Name,
				Managers:           account.Managers,
				Type:               enumor.AccountType(account.Type),
				Site:               enumor.AccountSiteType(account.Site),
				Price:              account.Price,
				PriceUnit:          account.PriceUnit,
				Memo:               account.Memo,
				RecycleReserveTime: account.RecycleReserveTime,
				Revision: core.Revision{
					Creator:   account.Creator,
					Reviser:   account.Reviser,
					CreatedAt: account.CreatedAt.String(),
					UpdatedAt: account.UpdatedAt.String(),
				},
			},
			Extension: extension,
		})
	}

	// 查询账号业务信息，并赋值
	accountBizMap, err := svc.getAccountBizMap(cts.Kit, ids)
	if err != nil {
		return nil, err
	}

	for _, one := range details {
		one.BkBizIDs = accountBizMap[one.ID]
	}

	return &protocloud.AccountWithExtensionListResult{Details: details}, nil
}
