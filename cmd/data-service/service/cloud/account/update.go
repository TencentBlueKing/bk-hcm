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
	protocloud "hcm/pkg/api/data-service/cloud"
	"hcm/pkg/criteria/enumor"
	"hcm/pkg/criteria/errf"
	"hcm/pkg/dal/dao/tools"
	"hcm/pkg/dal/dao/types"
	tablecloud "hcm/pkg/dal/table/cloud"
	tabletype "hcm/pkg/dal/table/types"
	"hcm/pkg/logs"
	"hcm/pkg/rest"
	"hcm/pkg/tools/json"
)

// UpdateAccount account with filter.
func (svc *service) UpdateAccount(cts *rest.Contexts) (interface{}, error) {
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
	case enumor.Other:
		return updateAccount[protocloud.OtherAccountExtensionUpdateReq](accountID, svc, cts)
	}

	return nil, nil
}

func getAccountFromTable(accountID string, svc *service, cts *rest.Contexts) (*tablecloud.AccountTable, error) {
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

func updateAccount[T protocloud.AccountExtensionUpdateReq, PT protocloud.SecretEncryptor[T]](accountID string,
	svc *service, cts *rest.Contexts) (interface{}, error) {

	req := new(protocloud.AccountUpdateReq[T])

	if err := cts.DecodeInto(req); err != nil {
		return nil, errf.NewFromErr(errf.DecodeRequestFailed, err)
	}

	if err := req.Validate(); err != nil {
		return nil, errf.NewFromErr(errf.InvalidParameter, err)
	}

	account := &tablecloud.AccountTable{
		Name:               req.Name,
		Managers:           req.Managers,
		Price:              req.Price,
		PriceUnit:          req.PriceUnit,
		Memo:               req.Memo,
		RecycleReserveTime: req.RecycleReserveTime,
		BkBizID:            req.BkBizID,
		Reviser:            cts.Kit.User,
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
