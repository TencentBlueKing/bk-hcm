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

// Package rootaccount ...
package rootaccount

import (
	"fmt"

	"hcm/pkg/api/core"
	protocore "hcm/pkg/api/core/account-set"
	dataproto "hcm/pkg/api/data-service/account-set"
	"hcm/pkg/criteria/enumor"
	"hcm/pkg/criteria/errf"
	"hcm/pkg/dal/dao/tools"
	"hcm/pkg/dal/dao/types"
	table "hcm/pkg/dal/table/account-set"
	tabletype "hcm/pkg/dal/table/types"
	"hcm/pkg/logs"
	"hcm/pkg/rest"
	"hcm/pkg/tools/json"
)

// GetRootAccountBasicInfo ...
func (svc *service) GetRootAccountBasicInfo(cts *rest.Contexts) (interface{}, error) {
	accountID := cts.PathParameter("account_id").String()

	dbAccount, err := getRootAccountFromTable(accountID, svc, cts)
	if err != nil {
		logs.Errorf("GetRootAccountBasicInfo getRootAccountFromTable accountID: %s, error: %s, rid: %s",
			accountID, err.Error(), cts.Kit.Rid)
		return nil, err
	}

	baseAccount := &protocore.BaseRootAccount{
		ID:          dbAccount.ID,
		Name:        dbAccount.Name,
		Vendor:      enumor.Vendor(dbAccount.Vendor),
		CloudID:     dbAccount.CloudID,
		Email:       dbAccount.Email,
		Managers:    dbAccount.Managers,
		BakManagers: dbAccount.BakManagers,
		Site:        enumor.RootAccountSiteType(dbAccount.Site),
		DeptID:      dbAccount.DeptID,
		Memo:        dbAccount.Memo,
		Revision: core.Revision{
			Creator:   dbAccount.Creator,
			Reviser:   dbAccount.Reviser,
			CreatedAt: dbAccount.CreatedAt.String(),
			UpdatedAt: dbAccount.UpdatedAt.String(),
		},
	}

	return dataproto.RootAccountGetBaseResult{
		BaseRootAccount: *baseAccount,
	}, nil
}

// GetRootAccount get root account
func (svc *service) GetRootAccount(cts *rest.Contexts) (interface{}, error) {
	vendor := enumor.Vendor(cts.Request.PathParameter("vendor"))
	if err := vendor.Validate(); err != nil {
		return nil, errf.NewFromErr(errf.InvalidParameter, err)
	}
	accountID := cts.PathParameter("account_id").String()

	dbAccount, err := getRootAccountFromTable(accountID, svc, cts)
	if err != nil {
		logs.Errorf("GetRootAccount getRootAccountFromTable accountID: %s, error: %s, rid: %s",
			accountID, err.Error(), cts.Kit.Rid)
		return nil, err
	}

	baseAccount := &protocore.BaseRootAccount{
		ID:          dbAccount.ID,
		Name:        dbAccount.Name,
		Vendor:      enumor.Vendor(dbAccount.Vendor),
		CloudID:     dbAccount.CloudID,
		Email:       dbAccount.Email,
		Managers:    dbAccount.Managers,
		BakManagers: dbAccount.BakManagers,
		Site:        enumor.RootAccountSiteType(dbAccount.Site),
		DeptID:      dbAccount.DeptID,
		Memo:        dbAccount.Memo,
		Revision: core.Revision{
			Creator:   dbAccount.Creator,
			Reviser:   dbAccount.Reviser,
			CreatedAt: dbAccount.CreatedAt.String(),
			UpdatedAt: dbAccount.UpdatedAt.String(),
		},
	}

	// 转换为最终的数据结构
	var account interface{}
	switch vendor {
	case enumor.Aws:
		account, err = convertToRootAccountResult[protocore.AwsRootAccountExtension](
			baseAccount, dbAccount.Extension, svc)
	case enumor.Gcp:
		account, err = convertToRootAccountResult[protocore.GcpRootAccountExtension](
			baseAccount, dbAccount.Extension, svc)
	case enumor.HuaWei:
		account, err = convertToRootAccountResult[protocore.HuaWeiRootAccountExtension](
			baseAccount, dbAccount.Extension, svc)
	case enumor.Azure:
		account, err = convertToRootAccountResult[protocore.AzureRootAccountExtension](
			baseAccount, dbAccount.Extension, svc)
	case enumor.Zenlayer:
		account, err = convertToRootAccountResult[protocore.ZenlayerRootAccountExtension](
			baseAccount, dbAccount.Extension, svc)
	case enumor.Kaopu:
		account, err = convertToRootAccountResult[protocore.KaopuRootAccountExtension](
			baseAccount, dbAccount.Extension, svc)
	}

	if err != nil {
		return nil, err
	}

	return account, nil
}

func getRootAccountFromTable(accountID string, svc *service, cts *rest.Contexts) (*table.RootAccountTable, error) {
	opt := &types.ListOption{
		Filter: tools.EqualExpression("id", accountID),
		Page:   &core.BasePage{Count: false, Start: 0, Limit: 1},
	}
	listRootAccountDetails, err := svc.dao.RootAccount().List(cts.Kit, opt)
	if err != nil {
		logs.Errorf("list root account failed, err: %v, rid: %s", err, cts.Kit.Rid)
		return nil, fmt.Errorf("list root account failed, err: %v", err)
	}

	details := listRootAccountDetails.Details
	if len(details) != 1 {
		return nil, fmt.Errorf("list account failed, account(id=%s) don't exist", accountID)
	}

	return details[0], nil
}

func convertToRootAccountResult[T dataproto.RootAccountExtensionGetResp, PT dataproto.RootSecretDecryptor[T]](
	baseRootAccount *protocore.BaseRootAccount, dbExtension tabletype.JsonField, svc *service,
) (*dataproto.RootAccountGetResult[T], error) {

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

	return &dataproto.RootAccountGetResult[T]{
		BaseRootAccount: *baseRootAccount,
		Extension:       extension,
	}, nil
}
