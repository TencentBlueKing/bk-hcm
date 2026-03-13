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

// Package resplandemandgputemplate ...
package resplandemandgputemplate

import (
	rpproto "hcm/pkg/api/data-service/resource-plan"
	"hcm/pkg/criteria/errf"
	"hcm/pkg/dal/dao/orm"
	"hcm/pkg/dal/dao/tools"
	tablegputpl "hcm/pkg/dal/table/resource-plan/res-plan-demand-gpu-template"
	"hcm/pkg/kit"
	"hcm/pkg/logs"
	"hcm/pkg/rest"
	cvt "hcm/pkg/tools/converter"

	"github.com/jmoiron/sqlx"
)

// BatchUpdateDemandGpuTemplate update demand gpu template.
func (svc *service) BatchUpdateDemandGpuTemplate(cts *rest.Contexts) (interface{}, error) {
	req := new(rpproto.DemandGpuTemplateBatchUpdateReq)

	if err := cts.DecodeInto(req); err != nil {
		return nil, errf.NewFromErr(errf.DecodeRequestFailed, err)
	}

	if err := req.Validate(); err != nil {
		return nil, errf.NewFromErr(errf.InvalidParameter, err)
	}

	_, err := svc.dao.Txn().AutoTxn(cts.Kit, func(txn *sqlx.Tx, opt *orm.TxnOption) (interface{}, error) {
		if err := svc.batchUpdateDemandGpuTemplateWithTx(cts.Kit, txn, req.Templates); err != nil {
			logs.Errorf("failed to batch update demand gpu template with tx, err: %v, rid: %v",
				err, cts.Kit.Rid)
			return nil, err
		}
		return nil, nil
	})
	if err != nil {
		logs.Errorf("update demand gpu template failed, err: %v, rid: %v", err, cts.Kit.Rid)
		return nil, err
	}

	return nil, nil
}

func (svc *service) batchUpdateDemandGpuTemplateWithTx(kt *kit.Kit, txn *sqlx.Tx,
	updateReqs []rpproto.DemandGpuTemplateUpdateReq) error {

	for _, updateReq := range updateReqs {
		record := &tablegputpl.ResPlanDemandGpuTemplateTable{
			TplSchema: updateReq.TplSchema,
			Reviser:   updateReq.Reviser,
			Remark:    cvt.PtrToVal(updateReq.Remark),
		}
		if err := svc.dao.ResPlanDemandGpuTemplate().UpdateWithTx(kt, txn,
			tools.EqualExpression("id", updateReq.ID), record); err != nil {
			logs.Errorf("update demand gpu template failed, err: %v, rid: %s", err, kt.Rid)
			return err
		}
	}

	return nil
}
