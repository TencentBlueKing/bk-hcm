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
	"reflect"

	"hcm/pkg/api/core"
	dataproto "hcm/pkg/api/data-service/account-set"
	"hcm/pkg/criteria/enumor"
	"hcm/pkg/criteria/errf"
	"hcm/pkg/dal/dao/orm"
	table "hcm/pkg/dal/table/account-set"
	tabletype "hcm/pkg/dal/table/types"
	"hcm/pkg/logs"
	"hcm/pkg/rest"
	"hcm/pkg/tools/json"

	"github.com/jmoiron/sqlx"
)

// CreateRootAccount account with options
func (svc *service) CreateRootAccount(cts *rest.Contexts) (interface{}, error) {
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
		result, err = createAccount[dataproto.AwsRootAccountExtensionCreateReq](vendor, svc, cts)
	case enumor.Gcp:
		result, err = createAccount[dataproto.GcpRootAccountExtensionCreateReq](vendor, svc, cts)
	case enumor.HuaWei:
		result, err = createAccount[dataproto.HuaWeiRootAccountExtensionCreateReq](vendor, svc, cts)
	case enumor.Azure:
		result, err = createAccount[dataproto.AzureRootAccountExtensionCreateReq](vendor, svc, cts)
	case enumor.Zenlayer:
		result, err = createAccount[dataproto.ZenlayerRootAccountExtensionCreateReq](vendor, svc, cts)
	case enumor.Kaopu:
		result, err = createAccount[dataproto.KaopuRootAccountExtensionCreateReq](vendor, svc, cts)
	default:
		return nil, fmt.Errorf("unsupport %s vendor for now", vendor)
	}

	if err != nil {
		logs.Errorf("create [%s] root account failed, err: %s, cid: %s", vendor, err, cts.Kit.Rid)
		return nil, err
	}
	return result, nil
}

func createAccount[T dataproto.RootAccountExtensionCreateReq, PT dataproto.RootSecretEncryptor[T]](vendor enumor.Vendor,
	svc *service, cts *rest.Contexts) (interface{}, error) {
	req := new(dataproto.RootAccountCreateReq[T])
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

		account := &table.RootAccountTable{
			Vendor:      string(vendor),
			Name:        req.Name,
			CloudID:     req.CloudID,
			Email:       req.Email,
			Managers:    req.Managers,
			BakManagers: req.BakManagers,
			Site:        string(req.Site),
			DeptID:      req.DeptID,
			Memo:        req.Memo,
			Extension:   tabletype.JsonField(extensionJson),
			Creator:     cts.Kit.User,
			Reviser:     cts.Kit.User,
		}
		accountID, err := svc.dao.RootAccount().CreateWithTx(cts.Kit, txn, account)
		if err != nil {
			return nil, fmt.Errorf("create root account failed, err: %v", err)
		}

		return accountID, nil
	})
	if err != nil {
		return nil, err
	}

	id, ok := accountID.(string)
	if !ok {
		return nil, fmt.Errorf("create root account but return id type not string, id type: %v",
			reflect.TypeOf(accountID).String())
	}
	return &core.CreateResult{ID: id}, nil
}
