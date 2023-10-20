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

	"github.com/jmoiron/sqlx"

	"hcm/pkg/api/core"
	protocloud "hcm/pkg/api/data-service/cloud"
	"hcm/pkg/criteria/enumor"
	"hcm/pkg/criteria/errf"
	"hcm/pkg/dal/dao/orm"
	"hcm/pkg/dal/dao/tools"
	"hcm/pkg/dal/dao/types"
	tableaudit "hcm/pkg/dal/table/audit"
	tabletype "hcm/pkg/dal/table/types"
	"hcm/pkg/kit"
	"hcm/pkg/logs"
	"hcm/pkg/rest"
)

// DeleteAccount account with filter.
func (svc *service) DeleteAccount(cts *rest.Contexts) (interface{}, error) {
	req := new(protocloud.AccountDeleteReq)
	if err := cts.DecodeInto(req); err != nil {
		return nil, err
	}

	if err := req.Validate(); err != nil {
		return nil, errf.NewFromErr(errf.InvalidParameter, err)
	}

	opt := &types.ListOption{
		Filter: req.Filter,
		Page:   core.NewDefaultBasePage(),
	}
	listResp, err := svc.dao.Account().List(cts.Kit, opt)
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
		_, err = svc.dao.Account().DeleteValidate(cts.Kit, one.ID)
		if err != nil {
			return nil, err
		}

		delAccountIDs[index] = one.ID
	}

	_, err = svc.dao.Txn().AutoTxn(cts.Kit, func(txn *sqlx.Tx, opt *orm.TxnOption) (interface{}, error) {
		accounts, err := svc.ListAccountWithBiz(cts.Kit, delAccountIDs)
		if err != nil {
			return nil, err
		}

		delAccountFilter := tools.ContainersExpression("id", delAccountIDs)
		if err := svc.dao.Account().DeleteWithTx(cts.Kit, txn, delAccountFilter); err != nil {
			return nil, err
		}

		delAccountBizRelFilter := tools.ContainersExpression("account_id", delAccountIDs)
		if err := svc.dao.AccountBizRel().DeleteWithTx(cts.Kit, txn, delAccountBizRelFilter); err != nil {
			return nil, err
		}

		// create audit
		if err = svc.createDeleteAudit(cts.Kit, accounts); err != nil {
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

func (svc *service) createDeleteAudit(kt *kit.Kit, accounts []types.Account) error {
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
			Operator:   kt.User,
			Source:     kt.GetRequestSource(),
			Rid:        kt.Rid,
			AppCode:    kt.AppCode,
			Detail: &tableaudit.BasicDetail{
				Data: one,
			},
		})
	}
	if err := svc.dao.Audit().BatchCreate(kt, audits); err != nil {
		logs.Errorf("batch create audit failed, err: %v, rid: %s", err, kt.Rid)
		return err
	}

	return nil
}

// DeleteValidate account delete validate.
func (svc *service) DeleteValidate(cts *rest.Contexts) (interface{}, error) {
	accountID := cts.PathParameter("account_id").String()
	if len(accountID) == 0 {
		return nil, errf.New(errf.InvalidParameter, "account_id is required")
	}

	validateResult, err := svc.dao.Account().DeleteValidate(cts.Kit, accountID)
	if err != nil {
		return validateResult, err
	}

	return nil, nil
}
