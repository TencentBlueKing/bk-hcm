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

// Package mainaccount ...
package mainaccount

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

// GetMainAccountBasicInfo ...
func (svc *service) GetMainAccountBasicInfo(cts *rest.Contexts) (interface{}, error) {
	accountID := cts.PathParameter("account_id").String()

	dbAccount, err := getMainAccountFromTable(accountID, svc, cts)
	if err != nil {
		logs.Errorf("GetMainAccountBasicInfo getMainAccountFromTable accountID:%s, error: %s, rid: %s", accountID, err.Error(), cts.Kit.Rid)
		return nil, err
	}

	baseAccount := &protocore.BaseMainAccount{
		ID:                dbAccount.ID,
		Name:              dbAccount.Name,
		Vendor:            enumor.Vendor(dbAccount.Vendor),
		CloudID:           dbAccount.CloudID,
		Email:             dbAccount.Email,
		Managers:          dbAccount.Managers,
		BakManagers:       dbAccount.BakManagers,
		Site:              enumor.MainAccountSiteType(dbAccount.Site),
		BusinessType:      enumor.MainAccountBusinessType(dbAccount.BusinessType),
		Status:            enumor.MainAccountStatus(dbAccount.Status),
		ParentAccountName: dbAccount.ParentAccountName,
		ParentAccountID:   dbAccount.ParentAccountID,
		DeptID:            dbAccount.DeptID,
		BkBizID:           dbAccount.BkBizID,
		OpProductID:       dbAccount.OpProductID,
		Memo:              dbAccount.Memo,
		Revision: core.Revision{
			Creator:   dbAccount.Creator,
			Reviser:   dbAccount.Reviser,
			CreatedAt: dbAccount.CreatedAt.String(),
			UpdatedAt: dbAccount.UpdatedAt.String(),
		},
	}

	return dataproto.MainAccountGetBaseResult{
		BaseMainAccount: *baseAccount,
	}, nil
}

// GetMainAccount get main account
func (svc *service) GetMainAccount(cts *rest.Contexts) (interface{}, error) {
	vendor := enumor.Vendor(cts.Request.PathParameter("vendor"))
	if err := vendor.Validate(); err != nil {
		return nil, errf.NewFromErr(errf.InvalidParameter, err)
	}
	accountID := cts.PathParameter("account_id").String()

	dbAccount, err := getMainAccountFromTable(accountID, svc, cts)
	if err != nil {
		logs.Errorf("GetMainAccount getMainAccountFromTable accountID: %s, error: %s, rid: %s", accountID, err.Error(), cts.Kit.Rid)
		return nil, err
	}

	baseAccount := &protocore.BaseMainAccount{
		ID:                dbAccount.ID,
		Name:              dbAccount.Name,
		Vendor:            enumor.Vendor(dbAccount.Vendor),
		CloudID:           dbAccount.CloudID,
		Email:             dbAccount.Email,
		Managers:          dbAccount.Managers,
		BakManagers:       dbAccount.BakManagers,
		Site:              enumor.MainAccountSiteType(dbAccount.Site),
		BusinessType:      enumor.MainAccountBusinessType(dbAccount.BusinessType),
		Status:            enumor.MainAccountStatus(dbAccount.Status),
		ParentAccountName: dbAccount.ParentAccountName,
		ParentAccountID:   dbAccount.ParentAccountID,
		DeptID:            dbAccount.DeptID,
		BkBizID:           dbAccount.BkBizID,
		OpProductID:       dbAccount.OpProductID,
		Memo:              dbAccount.Memo,
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
		account, err = convertToMainAccountResult[protocore.AwsMainAccountExtension](baseAccount, dbAccount.Extension, svc)
	case enumor.Gcp:
		account, err = convertToMainAccountResult[protocore.GcpMainAccountExtension](baseAccount, dbAccount.Extension, svc)
	case enumor.HuaWei:
		account, err = convertToMainAccountResult[protocore.HuaWeiMainAccountExtension](baseAccount, dbAccount.Extension, svc)
	case enumor.Azure:
		account, err = convertToMainAccountResult[protocore.AzureMainAccountExtension](baseAccount, dbAccount.Extension, svc)
	case enumor.Zenlayer:
		account, err = convertToMainAccountResult[protocore.ZenlayerMainAccountExtension](baseAccount, dbAccount.Extension, svc)
	case enumor.Kaopu:
		account, err = convertToMainAccountResult[protocore.KaopuMainAccountExtension](baseAccount, dbAccount.Extension, svc)
	}

	if err != nil {
		return nil, err
	}

	return account, nil
}

func getMainAccountFromTable(accountID string, svc *service, cts *rest.Contexts) (*table.MainAccountTable, error) {
	opt := &types.ListOption{
		Filter: tools.EqualExpression("id", accountID),
		Page:   &core.BasePage{Count: false, Start: 0, Limit: 1},
	}
	listMainAccountDetails, err := svc.dao.MainAccount().List(cts.Kit, opt)
	if err != nil {
		logs.Errorf("list main account failed, account id: %s, err: %v, rid: %s", accountID, err, cts.Kit.Rid)
		return nil, fmt.Errorf("list main account failed, err: %v", err)
	}

	details := listMainAccountDetails.Details
	if len(details) != 1 {
		return nil, fmt.Errorf("list account failed, account(id=%s) don't exist", accountID)
	}

	return details[0], nil
}

func convertToMainAccountResult[T dataproto.MainAccountExtensionGetResp, PT dataproto.SecretDecryptor[T]](
	baseMainAccount *protocore.BaseMainAccount, dbExtension tabletype.JsonField, svc *service,
) (*dataproto.MainAccountGetResult[T], error) {
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

	return &dataproto.MainAccountGetResult[T]{
		BaseMainAccount: *baseMainAccount,
		Extension:       extension,
	}, nil
}
