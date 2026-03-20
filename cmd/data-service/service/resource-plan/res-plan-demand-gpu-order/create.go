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

package resplandemandgpuorder

import (
	"fmt"

	rpaudit "hcm/cmd/data-service/service/audit/cloud/resource-plan"
	"hcm/pkg/api/core"
	rpproto "hcm/pkg/api/data-service/resource-plan"
	"hcm/pkg/criteria/errf"
	"hcm/pkg/dal/dao/orm"
	rpgpu "hcm/pkg/dal/table/resource-plan/res-plan-demand-gpu-order"
	"hcm/pkg/logs"
	"hcm/pkg/rest"
	"hcm/pkg/tools/util"

	"github.com/jmoiron/sqlx"
)

// BatchCreateResPlanDemandGpuOrder batch create res plan demand gpu order.
func (svc *service) BatchCreateResPlanDemandGpuOrder(cts *rest.Contexts) (interface{}, error) {
	req := new(rpproto.ResPlanDemandGpuOrderBatchCreateReq)
	if err := cts.DecodeInto(req); err != nil {
		return nil, errf.NewFromErr(errf.DecodeRequestFailed, err)
	}

	if err := req.Validate(); err != nil {
		return nil, errf.NewFromErr(errf.InvalidParameter, err)
	}

	ids, err := svc.dao.Txn().AutoTxn(cts.Kit, func(txn *sqlx.Tx, opt *orm.TxnOption) (interface{}, error) {
		models := make([]rpgpu.ResPlanDemandGpuOrderTable, len(req.Items))
		for idx, item := range req.Items {
			m := rpgpu.ResPlanDemandGpuOrderTable{
				BkBizID:       item.BkBizID,
				OpProductID:   item.OpProductID,
				OpProductName: item.OpProductName,
				TemplateID:    item.TemplateID,
				Status:        item.Status,
				Remark:        item.Remark,
				Creator:       cts.Kit.User,
				Reviser:       cts.Kit.User,
			}
			models[idx] = m
		}

		recordIDs, err := svc.dao.ResPlanDemandGpuOrder().CreateWithTx(cts.Kit, txn, models)
		if err != nil {
			logs.Errorf("batch create res plan demand gpu order failed, err: %v, rid: %s", err, cts.Kit.Rid)
			return nil, err
		}

		strIDs, err := util.GetStrSliceByInterface(recordIDs)
		if err != nil {
			return nil, fmt.Errorf("convert record ids to []string failed, err: %v", err)
		}
		for i := range models {
			models[i].ID = strIDs[i]
		}

		audits := rpaudit.GpuOrderCreateAudits(cts.Kit, models)
		if err = svc.dao.Audit().BatchCreateWithTx(cts.Kit, txn, audits); err != nil {
			logs.Errorf("batch create gpu order audit failed, err: %v, rid: %s", err, cts.Kit.Rid)
			return nil, err
		}

		return strIDs, nil
	})
	if err != nil {
		logs.Errorf("batch create res plan demand gpu order txn failed, err: %v, rid: %s", err, cts.Kit.Rid)
		return nil, err
	}

	strIDs, err := util.GetStrSliceByInterface(ids)
	if err != nil {
		return nil, fmt.Errorf("batch create res plan demand gpu order but return ids type not []string, err: %v", err)
	}

	return &core.BatchCreateResult{IDs: strIDs}, nil
}
