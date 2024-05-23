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

package sgcomrel

import (
	"fmt"

	"hcm/pkg/api/core"
	proto "hcm/pkg/api/data-service"
	"hcm/pkg/criteria/errf"
	"hcm/pkg/dal/dao/orm"
	"hcm/pkg/dal/dao/tools"
	"hcm/pkg/dal/dao/types"
	"hcm/pkg/logs"
	"hcm/pkg/rest"

	"github.com/jmoiron/sqlx"
)

// BatchDelete rels.
func (svc *sgComRelSvc) BatchDelete(cts *rest.Contexts) (interface{}, error) {
	req := new(proto.BatchDeleteReq)
	if err := cts.DecodeInto(req); err != nil {
		return nil, err
	}

	if err := req.Validate(); err != nil {
		return nil, errf.NewFromErr(errf.InvalidParameter, err)
	}

	opt := &types.ListOption{
		Fields: []string{"id"},
		Filter: req.Filter,
		Page:   core.NewDefaultBasePage(),
	}
	listResp, err := svc.dao.SGCommonRel().List(cts.Kit, opt)
	if err != nil {
		logs.Errorf("list security group common rels failed, err: %v, rid: %s", err, cts.Kit.Rid)
		return nil, fmt.Errorf("list security group common rels failed, err: %v", err)
	}

	if len(listResp.Details) == 0 {
		return nil, nil
	}

	delIDs := make([]uint64, len(listResp.Details))
	for index, one := range listResp.Details {
		delIDs[index] = one.ID
	}

	_, err = svc.dao.Txn().AutoTxn(cts.Kit, func(txn *sqlx.Tx, opt *orm.TxnOption) (interface{}, error) {
		if err = svc.dao.SGCommonRel().DeleteWithTx(
			cts.Kit, txn, tools.ContainersExpression("id", delIDs)); err != nil {
			return nil, err
		}

		return nil, nil
	})
	if err != nil {
		logs.Errorf("delete security group common rels failed, err: %v, rid: %s", err, cts.Kit.Rid)
		return nil, err
	}

	return nil, nil
}
