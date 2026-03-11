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

// Package resplansubticket ...
package resplansubticket

import (
	"fmt"

	rpproto "hcm/pkg/api/data-service/resource-plan"
	"hcm/pkg/criteria/errf"
	"hcm/pkg/dal/dao/orm"
	"hcm/pkg/dal/dao/tools"
	tablers "hcm/pkg/dal/table/resource-plan/res-plan-sub-ticket"
	ttypes "hcm/pkg/dal/table/types"
	"hcm/pkg/kit"
	"hcm/pkg/logs"
	"hcm/pkg/rest"
	"hcm/pkg/runtime/filter"

	"github.com/jmoiron/sqlx"
)

// BatchUpdateResPlanSubTicket update resource plan sub ticket
func (svc *service) BatchUpdateResPlanSubTicket(cts *rest.Contexts) (interface{}, error) {
	req := new(rpproto.ResPlanSubTicketBatchUpdateReq)

	if err := cts.DecodeInto(req); err != nil {
		return nil, errf.NewFromErr(errf.DecodeRequestFailed, err)
	}

	if err := req.Validate(); err != nil {
		return nil, errf.NewFromErr(errf.InvalidParameter, err)
	}

	_, err := svc.dao.Txn().AutoTxn(cts.Kit, func(txn *sqlx.Tx, opt *orm.TxnOption) (interface{}, error) {
		_, err := svc.batchUpdateResPlanSubTicketWithTx(cts.Kit, txn, req.SubTickets)
		if err != nil {
			logs.Errorf("failed to batch update res plan sub ticket with tx, err: %v, rid: %v", err, cts.Kit.Rid)
			return nil, err
		}
		return nil, nil
	})
	if err != nil {
		logs.Errorf("update res plan sub ticket failed, err: %v, rid: %v", err, cts.Kit.Rid)
		return nil, err
	}

	return nil, nil
}

func (svc *service) batchUpdateResPlanSubTicketWithTx(kt *kit.Kit, txn *sqlx.Tx,
	updateReqs []rpproto.ResPlanSubTicketUpdateReq) ([]string, error) {

	for _, updateReq := range updateReqs {
		record := &tablers.ResPlanSubTicketTable{
			SubType:             updateReq.SubType,
			BkBizID:             updateReq.BkBizID,
			BkBizName:           updateReq.BkBizName,
			OpProductID:         updateReq.OpProductID,
			OpProductName:       updateReq.OpProductName,
			PlanProductID:       updateReq.PlanProductID,
			PlanProductName:     updateReq.PlanProductName,
			VirtualDeptID:       updateReq.VirtualDeptID,
			VirtualDeptName:     updateReq.VirtualDeptName,
			Status:              updateReq.Status,
			Message:             updateReq.Message,
			Stage:               updateReq.Stage,
			AdminAuditStatus:    updateReq.AdminAuditStatus,
			AdminAuditOperator:  updateReq.AdminAuditOperator,
			AdminAuditAt:        updateReq.AdminAuditAt,
			OperateInfo:         updateReq.OperateInfo,
			CrpSN:               updateReq.CrpSN,
			CrpURL:              updateReq.CrpURL,
			SubOriginalOS:       updateReq.SubOriginalOS,
			SubOriginalCPUCore:  updateReq.SubOriginalCPUCore,
			SubOriginalMemory:   updateReq.SubOriginalMemory,
			SubOriginalDiskSize: updateReq.SubOriginalDiskSize,
			SubUpdatedOS:        updateReq.SubUpdatedOS,
			SubUpdatedCPUCore:   updateReq.SubUpdatedCPUCore,
			SubUpdatedMemory:    updateReq.SubUpdatedMemory,
			SubUpdatedDiskSize:  updateReq.SubUpdatedDiskSize,
			Reviser:             kt.User,
		}
		if updateReq.SubDemands != nil {
			record.SubDemands = ttypes.JsonField(*updateReq.SubDemands)
		}

		if _, err := svc.dao.ResPlanSubTicket().UpdateWithTx(kt, txn,
			tools.EqualExpression("id", updateReq.ID), record); err != nil {
			logs.Errorf("update res plan sub ticket failed, err: %v, rid: %s", err, kt.Rid)
			return nil, err
		}
	}
	return nil, nil
}

// UpdateResPlanSubTicketStatusCAS updates res plan sub ticket status with cas.
func (svc *service) UpdateResPlanSubTicketStatusCAS(cts *rest.Contexts) (interface{}, error) {
	req := new(rpproto.ResPlanSubTicketStatusUpdateReq)

	if err := cts.DecodeInto(req); err != nil {
		return nil, errf.NewFromErr(errf.DecodeRequestFailed, err)
	}

	if err := req.Validate(); err != nil {
		return nil, errf.NewFromErr(errf.InvalidParameter, err)
	}

	effected, err := svc.dao.Txn().AutoTxn(cts.Kit, func(txn *sqlx.Tx, opt *orm.TxnOption) (interface{}, error) {
		rules := make([]*filter.AtomRule, 0)
		rules = append(rules, tools.RuleEqual("ticket_id", req.TicketID))
		rules = append(rules, tools.RuleEqual("status", req.Source))
		if len(req.IDs) > 0 {
			rules = append(rules, tools.RuleIn("id", req.IDs))
		}
		updateFilter := tools.ExpressionAnd(rules...)

		record := &tablers.ResPlanSubTicketTable{
			Status:  req.Target,
			Message: req.Message,
			Reviser: cts.Kit.User,
		}

		effected, err := svc.dao.ResPlanSubTicket().UpdateWithTx(cts.Kit, txn, updateFilter, record)
		if err != nil {
			logs.Errorf("update res plan sub ticket status failed, err: %v, rid: %s", err, cts.Kit.Rid)
			return nil, err
		}
		return effected, nil
	})
	if err != nil {
		logs.Errorf("update res plan sub ticket status failed, err: %v, rid: %v", err, cts.Kit.Rid)
		return nil, err
	}

	// 入参指定ID时，需确保更新的行数和提供的ID一致
	if len(req.IDs) > 0 && effected != int64(len(req.IDs)) {
		logs.Errorf("update res plan sub ticket status failed, expected row count: %d, actual row count: %d, rid: %s",
			len(req.IDs), effected, cts.Kit.Rid)
		return nil, fmt.Errorf("update res plan sub ticket status failed, effected rows: %d", effected)
	}

	return nil, nil
}
