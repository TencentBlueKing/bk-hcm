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

package billitem

import (
	"fmt"

	"hcm/pkg/api/core"
	dataproto "hcm/pkg/api/data-service/bill"
	"hcm/pkg/criteria/errf"
	"hcm/pkg/dal/dao/orm"
	"hcm/pkg/dal/dao/tools"
	"hcm/pkg/dal/dao/types"
	"hcm/pkg/logs"
	"hcm/pkg/rest"

	"github.com/jmoiron/sqlx"
)

// DeleteBillItem delete bill item with options
func (svc *service) DeleteBillItem(cts *rest.Contexts) (interface{}, error) {
	req := new(dataproto.BillItemDeleteReq)
	if err := cts.DecodeInto(req); err != nil {
		return nil, err
	}

	if err := req.Validate(); err != nil {
		return nil, errf.NewFromErr(errf.InvalidParameter, err)
	}
	opt := &types.ListOption{
		Filter: req.Filter,
		Page: &core.BasePage{
			Start: 0,
			Limit: core.DefaultMaxPageLimit,
		},
	}
	listResp, err := svc.dao.AccountBillItem().List(cts.Kit, req.ItemCommonOpt, opt)
	if err != nil {
		logs.Errorf("delete list account bill item failed, err: %v, rid: %s", err, cts.Kit.Rid)
		return nil, fmt.Errorf("delete list account bill item failed, err: %v", err)
	}
	if len(listResp.Details) == 0 {
		return nil, nil
	}
	delIDs := make([]string, len(listResp.Details))
	for index, one := range listResp.Details {
		delIDs[index] = one.ID
	}
	_, err = svc.dao.Txn().AutoTxn(cts.Kit, func(txn *sqlx.Tx, opt *orm.TxnOption) (interface{}, error) {
		delFilter := tools.ContainersExpression("id", delIDs)
		err = svc.dao.AccountBillItem().DeleteWithTx(cts.Kit, txn, req.ItemCommonOpt, delFilter)
		if err != nil {
			return nil, err
		}
		return nil, nil
	})
	if err != nil {
		logs.Errorf("delete account bill item failed, err: %v, rid: %s", err, cts.Kit.Rid)
		return nil, err
	}

	return nil, nil
}
