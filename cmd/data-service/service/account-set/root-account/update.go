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

	dataproto "hcm/pkg/api/data-service/account-set"
	"hcm/pkg/criteria/enumor"
	"hcm/pkg/criteria/errf"
	"hcm/pkg/dal/dao/tools"
	table "hcm/pkg/dal/table/account-set"
	tabletype "hcm/pkg/dal/table/types"
	"hcm/pkg/logs"
	"hcm/pkg/rest"
	"hcm/pkg/tools/json"
)

// UpdateRootAccount update root account with filter.
func (svc *service) UpdateRootAccount(cts *rest.Contexts) (interface{}, error) {
	vendor := enumor.Vendor(cts.Request.PathParameter("vendor"))
	if err := vendor.Validate(); err != nil {
		return nil, err
	}

	accountID := cts.PathParameter("account_id").String()

	switch vendor {
	case enumor.Aws:
		return updateRootAccount[dataproto.AwsRootAccountExtensionUpdateReq](accountID, svc, cts)
	case enumor.Gcp:
		return updateRootAccount[dataproto.GcpRootAccountExtensionUpdateReq](accountID, svc, cts)
	case enumor.HuaWei:
		return updateRootAccount[dataproto.HuaWeiRootAccountExtensionUpdateReq](accountID, svc, cts)
	case enumor.Azure:
		return updateRootAccount[dataproto.AzureRootAccountExtensionUpdateReq](accountID, svc, cts)
	case enumor.Zenlayer:
		return updateRootAccount[dataproto.ZenlayerRootAccountExtensionUpdateReq](accountID, svc, cts)
	case enumor.Kaopu:
		return updateRootAccount[dataproto.KaopuRootAccountExtensionUpdateReq](accountID, svc, cts)
	}
	return nil, nil
}

func updateRootAccount[T dataproto.RootAccountExtensionUpdateReq, PT dataproto.RootSecretEncryptor[T]](accountID string,
	svc *service, cts *rest.Contexts) (interface{}, error) {

	req := new(dataproto.RootAccountUpdateReq[T])
	if err := cts.DecodeInto(req); err != nil {
		return nil, errf.NewFromErr(errf.DecodeRequestFailed, err)
	}

	if err := req.Validate(); err != nil {
		return nil, errf.NewFromErr(errf.InvalidParameter, err)
	}

	account := &table.RootAccountTable{
		Name:        req.Name,
		Managers:    req.Managers,
		BakManagers: req.BakManagers,
		DeptID:      req.DeptID,
		Memo:        req.Memo,
		Reviser:     cts.Kit.User,
	}

	// 只有提供了Extension才进行更新
	if req.Extension != nil {
		// 将参数里的SecretKey加密
		p := PT(req.Extension)
		p.EncryptSecretKey(svc.cipher)

		// 查询账号
		dbAccount, err := getRootAccountFromTable(accountID, svc, cts)
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

	err := svc.dao.RootAccount().Update(cts.Kit, tools.EqualExpression("id", accountID), account)
	if err != nil {
		err = fmt.Errorf("update main account failed, accountID: %s, err: %v, rid: %s", accountID, err, cts.Kit.Rid)
		logs.Errorf(err.Error())
		return nil, err
	}

	return nil, nil
}
