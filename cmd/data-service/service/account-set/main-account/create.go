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
	"reflect"

	"hcm/pkg/api/core"
	dataproto "hcm/pkg/api/data-service/account-set"
	"hcm/pkg/criteria/enumor"
	"hcm/pkg/criteria/errf"
	"hcm/pkg/dal/dao/orm"
	table "hcm/pkg/dal/table/account-set"
	"hcm/pkg/dal/table/types"
	"hcm/pkg/logs"
	"hcm/pkg/rest"
	"hcm/pkg/tools/json"

	"github.com/jmoiron/sqlx"
)

// CreateMainAccount create main account with options
func (svc *service) CreateMainAccount(cts *rest.Contexts) (interface{}, error) {
	vendor := enumor.Vendor(cts.Request.PathParameter("vendor"))
	if err := vendor.Validate(); err != nil {
		return nil, errf.NewFromErr(errf.InvalidParameter, err)
	}
	var (
		result interface{}
		err    error
	)
	switch vendor {
	case enumor.Aws:
		result, err = createAccount[dataproto.AwsMainAccountExtensionCreateReq](vendor, svc, cts)
	case enumor.Gcp:
		result, err = createAccount[dataproto.GcpMainAccountExtensionCreateReq](vendor, svc, cts)
	case enumor.HuaWei:
		result, err = createAccount[dataproto.HuaWeiMainAccountExtensionCreateReq](vendor, svc, cts)
	case enumor.Azure:
		result, err = createAccount[dataproto.AzureMainAccountExtensionCreateReq](vendor, svc, cts)
	case enumor.Zenlayer:
		result, err = createAccount[dataproto.ZenlayerMainAccountExtensionCreateReq](vendor, svc, cts)
	case enumor.Kaopu:
		result, err = createAccount[dataproto.KaopuMainAccountExtensionCreateReq](vendor, svc, cts)
	default:
		return nil, fmt.Errorf("unsupport %s vendor for now", vendor)
	}

	if err != nil {
		logs.Errorf("create [%s] main account failed, err: %s, rid: %s", vendor, err, cts.Kit.Rid)
		return nil, err
	}
	return result, nil
}

func createAccount[T dataproto.MainAccountExtensionCreateReq, PT dataproto.SecretEncryptor[T]](vendor enumor.Vendor,
	svc *service, cts *rest.Contexts) (interface{}, error) {
	req := new(dataproto.MainAccountCreateReq[T])
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

		account := &table.MainAccountTable{
			Vendor:            string(vendor),
			CloudID:           req.CloudID,
			Email:             req.Email,
			Managers:          req.Managers,
			BakManagers:       req.BakManagers,
			Site:              string(req.Site),
			BusinessType:      string(req.BusinessType),
			Status:            string(req.Status),
			ParentAccountName: req.ParentAccountName,
			ParentAccountID:   req.ParentAccountID,
			DeptID:            req.DeptID,
			BkBizID:           req.BkBizID,
			OpProductID:       req.OpProductID,
			Memo:              req.Memo,
			Extension:         types.JsonField(extensionJson),
			Creator:           cts.Kit.User,
			Reviser:           cts.Kit.User,
		}
		accountID, err := svc.dao.MainAccount().CreateWithTx(cts.Kit, txn, account)
		if err != nil {
			return nil, fmt.Errorf("create account failed, err: %v", err)
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
