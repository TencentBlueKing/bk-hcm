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
	"reflect"

	"github.com/jmoiron/sqlx"

	"hcm/pkg/api/core"
	protocloud "hcm/pkg/api/data-service/cloud"
	"hcm/pkg/criteria/constant"
	"hcm/pkg/criteria/enumor"
	"hcm/pkg/criteria/errf"
	"hcm/pkg/dal/dao/orm"
	tablecloud "hcm/pkg/dal/table/cloud"
	tabletype "hcm/pkg/dal/table/types"
	"hcm/pkg/rest"
	"hcm/pkg/tools/json"
)

// CreateAccount account with options
func (svc *service) CreateAccount(cts *rest.Contexts) (interface{}, error) {
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
	case enumor.Other:
		return createAccount[protocloud.OtherAccountExtensionCreateReq](vendor, svc, cts)
	default:
		return nil, fmt.Errorf("unsupport %s vendor for now", vendor)
	}
}

func createAccount[T protocloud.AccountExtensionCreateReq, PT protocloud.SecretEncryptor[T]](vendor enumor.Vendor,
	svc *service, cts *rest.Contexts) (interface{}, error) {

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
			Vendor:             string(vendor),
			Name:               req.Name,
			Managers:           req.Managers,
			Type:               string(req.Type),
			Site:               string(req.Site),
			Memo:               req.Memo,
			Extension:          tabletype.JsonField(extensionJson),
			RecycleReserveTime: constant.UnsetRecycleTime,
			Creator:            cts.Kit.User,
			Reviser:            cts.Kit.User,
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
