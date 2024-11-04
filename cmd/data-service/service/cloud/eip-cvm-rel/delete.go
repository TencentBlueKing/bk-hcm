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

package eipcvmrel

import (
	"fmt"

	"hcm/pkg/api/core"
	datarelproto "hcm/pkg/api/data-service/cloud"
	"hcm/pkg/criteria/errf"
	"hcm/pkg/dal/dao/orm"
	"hcm/pkg/dal/dao/tools"
	"hcm/pkg/dal/dao/types"
	"hcm/pkg/rest"

	"github.com/jmoiron/sqlx"
)

// BatchDeleteEipCvmRels ...
func (svc *relSvc) BatchDeleteEipCvmRels(cts *rest.Contexts) (interface{}, error) {
	req := new(datarelproto.EipCvmRelDeleteReq)
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

	relResult, err := svc.dao.EipCvmRel().List(cts.Kit, opt)
	if err != nil {
		return nil, fmt.Errorf("list eip cvm rels failed, err: %v", err)
	}

	if len(relResult.Details) == 0 {
		return nil, nil
	}

	delIDs := make([]uint64, len(relResult.Details))
	for idx, rel := range relResult.Details {
		delIDs[idx] = rel.ID
	}

	_, err = svc.dao.Txn().AutoTxn(cts.Kit, func(txn *sqlx.Tx, opt *orm.TxnOption) (interface{}, error) {
		return nil, svc.dao.EipCvmRel().DeleteWithTx(cts.Kit, txn, tools.ContainersExpression("id", delIDs))
	})
	return nil, err
}
