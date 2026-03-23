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

package plan

import (
	"encoding/json"
	"fmt"

	ptypes "hcm/cmd/woa-server/types/plan"
	"hcm/pkg/api/core"
	protoaudit "hcm/pkg/api/data-service/audit"
	rpproto "hcm/pkg/api/data-service/resource-plan"
	"hcm/pkg/criteria/constant"
	"hcm/pkg/criteria/enumor"
	"hcm/pkg/criteria/errf"
	"hcm/pkg/dal/dao/tools"
	resplandemandgpuorder "hcm/pkg/dal/table/resource-plan/res-plan-demand-gpu-order"
	gpusuborder "hcm/pkg/dal/table/resource-plan/res-plan-demand-gpu-suborder"
	tabletypes "hcm/pkg/dal/table/types"
	"hcm/pkg/kit"
	"hcm/pkg/logs"
	"hcm/pkg/runtime/filter"
	cvt "hcm/pkg/tools/converter"
	"hcm/pkg/tools/slice"
)

// TerminateBizGpuSubOrders 终止业务侧 GPU 需求子单的核心逻辑，包括校验、更新子单状态、刷新主单状态。
func (c *Controller) TerminateBizGpuSubOrders(kt *kit.Kit, bizID int64, subOrderIDs []string) error {
	subOrders, orderIDs, err := c.listTerminableBizGpuSubOrders(kt, bizID, subOrderIDs)
	if err != nil {
		logs.Errorf("failed to prepare biz gpu demand suborders before terminate, err: %v, biz: %d, rid: %s",
			err, bizID, kt.Rid)
		return err
	}

	suborderUpdateReq, terminateLogs := c.genTerminateGpuSubOrderReqAndAudit(subOrders)
	if err = c.AuditGpuSubOrderUpdates(kt, terminateLogs); err != nil {
		logs.Errorf("audit biz batch terminate gpu demand suborders failed, err: %v, biz: %d, rid: %s",
			err, bizID, kt.Rid)
		return err
	}

	if err = c.client.DataService().Global.ResourcePlan.BatchUpdateResPlanDemandGpuSubOrder(
		kt, suborderUpdateReq); err != nil {
		logs.Errorf("failed to batch terminate biz gpu demand suborders, err: %v, biz: %d, rid: %s",
			err, bizID, kt.Rid)
		return err
	}

	if err = c.updateTerminatedGpuOrders(kt, orderIDs); err != nil {
		logs.Errorf("failed to update gpu demand orders after suborder terminate, err: %v, biz: %d, rid: %s",
			err, bizID, kt.Rid)
		return err
	}

	return nil
}

// BatchReviewGpuSubOrders 资源侧批量评审 GPU 需求子单，包括校验、审计、更新子单状态和刷新主单状态。
func (c *Controller) BatchReviewGpuSubOrders(kt *kit.Kit,
	req *ptypes.BatchUpdateStatusResPlanDemandGpuSubOrderReq) error {

	subOrderMap, err := c.listGpuSubOrdersByIDs(kt, req.SubOrderIDs, nil, []string{"id", "status", "order_id"})
	if err != nil {
		logs.Errorf("failed to list gpu demand suborders before batch review, err: %v, rid: %s", err, kt.Rid)
		return errf.NewFromErr(errf.Aborted, err)
	}

	comment, err := buildJsonFieldPtrFromArray(req.Comment)
	if err != nil {
		logs.Errorf("failed to parse batch review gpu demand suborders comment, err: %v, rid: %s", err, kt.Rid)
		return errf.NewFromErr(errf.InvalidParameter, err)
	}

	reviewOrderIDs, auditLogs, err := c.buildGpuSubOrderBatchReviewAudit(
		req.SubOrderIDs, req.Status, req.Comment, subOrderMap)
	if err != nil {
		return err
	}

	if err = c.AuditGpuSubOrderUpdates(kt, auditLogs); err != nil {
		logs.Errorf("audit batch review gpu demand suborders failed, err: %v, rid: %s", err, kt.Rid)
		return errf.NewFromErr(errf.Aborted, err)
	}

	for _, ids := range slice.Split(req.SubOrderIDs, constant.BatchOperationMaxLimit) {
		updateReq := &rpproto.ResPlanDemandGpuSubOrderBatchUpdateStatusReq{
			IDs:     ids,
			Status:  req.Status,
			Comment: comment,
		}
		if err = c.client.DataService().Global.ResourcePlan.BatchUpdateStatusResPlanDemandGpuSubOrder(
			kt, updateReq); err != nil {
			logs.Errorf("failed to batch update gpu demand suborders status in scr, err: %v, rid: %s", err, kt.Rid)
			return errf.NewFromErr(errf.Aborted, err)
		}
	}

	if err = c.RefreshGpuOrderStatusAfterReview(kt, reviewOrderIDs); err != nil {
		logs.Errorf("failed to refresh gpu demand order status after batch review, err: %v, rid: %s", err, kt.Rid)
		return errf.NewFromErr(errf.Aborted, err)
	}

	return nil
}

func (c *Controller) listTerminableBizGpuSubOrders(kt *kit.Kit, bizID int64, subOrderIDs []string) (
	[]gpusuborder.ResPlanDemandGpuSubOrderTable, []string, error) {

	resp, err := c.client.DataService().Global.ResourcePlan.ListResPlanDemandGpuSubOrder(kt,
		&rpproto.ResPlanDemandGpuSubOrderListReq{
			ListReq: core.ListReq{
				Filter: tools.ExpressionAnd(tools.RuleIn("id", subOrderIDs), tools.RuleEqual("bk_biz_id", bizID)),
				Page:   &core.BasePage{Start: 0, Limit: uint(len(subOrderIDs))},
				Fields: []string{"id", "status", "order_id"},
			},
		})
	if err != nil {
		return nil, nil, err
	}
	if len(resp.Details) != len(subOrderIDs) {
		return nil, nil, fmt.Errorf("some suborder_ids are invalid or not in biz: %d", bizID)
	}

	orderIDs := make([]string, 0, len(resp.Details))
	orderIDSet := make(map[string]struct{}, len(resp.Details))
	for _, subOrder := range resp.Details {
		if subOrder.Status != enumor.RPDemandGPUSubOrderStatusReject &&
			subOrder.Status != enumor.RPDemandGPUSubOrderStatusInit {
			return nil, nil, fmt.Errorf("suborder(%s) status(%s) is not allowed to terminate",
				subOrder.ID, subOrder.Status)
		}
		if _, exists := orderIDSet[subOrder.OrderID]; !exists && subOrder.OrderID != "" {
			orderIDSet[subOrder.OrderID] = struct{}{}
			orderIDs = append(orderIDs, subOrder.OrderID)
		}
	}

	return resp.Details, orderIDs, nil
}

func (c *Controller) buildGpuSubOrderBatchReviewAudit(subOrderIDs []string, status enumor.RPDemandGPUSubOrderStatus,
	comment []json.RawMessage, subOrderMap map[string]gpusuborder.ResPlanDemandGpuSubOrderTable) (
	[]string, []protoaudit.CloudResourceUpdateInfo, error) {

	reviewOrderIDs := make([]string, 0)
	reviewOrderIDSet := make(map[string]struct{})
	auditLogs := make([]protoaudit.CloudResourceUpdateInfo, 0, len(subOrderIDs))

	for _, subOrderID := range subOrderIDs {
		subOrder, exists := subOrderMap[subOrderID]
		if !exists {
			return nil, nil, errf.NewFromErr(errf.InvalidParameter, fmt.Errorf("suborder(%s) not found", subOrderID))
		}
		if subOrder.Status != enumor.RPDemandGPUSubOrderStatusPending {
			return nil, nil, errf.NewFromErr(errf.InvalidParameter,
				fmt.Errorf("suborder(%s) status(%s) is not allowed to review", subOrder.ID, subOrder.Status))
		}

		updateFields := map[string]interface{}{"status": status}
		if comment != nil {
			updateFields["comment"] = comment
		}
		auditLogs = append(auditLogs, protoaudit.CloudResourceUpdateInfo{
			ResType:      enumor.ResPlanGPUDemandsSuborderAuditResType,
			ResID:        subOrder.ID,
			UpdateFields: updateFields,
		})
		if _, exists = reviewOrderIDSet[subOrder.OrderID]; !exists {
			reviewOrderIDSet[subOrder.OrderID] = struct{}{}
			reviewOrderIDs = append(reviewOrderIDs, subOrder.OrderID)
		}
	}

	return reviewOrderIDs, auditLogs, nil
}

// genTerminateGpuSubOrderReqAndAudit 一次遍历同时生成子单终止更新请求和审计日志。
func (c *Controller) genTerminateGpuSubOrderReqAndAudit(subOrders []gpusuborder.ResPlanDemandGpuSubOrderTable) (
	*rpproto.ResPlanDemandGpuSubOrderBatchUpdateReq, []protoaudit.CloudResourceUpdateInfo) {

	req := &rpproto.ResPlanDemandGpuSubOrderBatchUpdateReq{
		SubOrders: make([]rpproto.ResPlanDemandGpuSubOrderUpdateReq, 0, len(subOrders)),
	}
	auditLogs := make([]protoaudit.CloudResourceUpdateInfo, 0, len(subOrders))
	for _, subOrder := range subOrders {
		req.SubOrders = append(req.SubOrders, rpproto.ResPlanDemandGpuSubOrderUpdateReq{
			ID:     subOrder.ID,
			Status: enumor.RPDemandGPUSubOrderStatusTerminate,
		})
		auditLogs = append(auditLogs, protoaudit.CloudResourceUpdateInfo{
			ResType:      enumor.ResPlanGPUDemandsSuborderAuditResType,
			ResID:        subOrder.ID,
			UpdateFields: map[string]interface{}{"status": enumor.RPDemandGPUSubOrderStatusTerminate},
		})
	}

	return req, auditLogs
}

// updateTerminatedGpuOrders 子单终止后刷新关联主单的状态：
// 1. 所有子单都已终止 → 主单变为"已终止"
// 2. 其余情况复用 RefreshGpuOrderStatusAfterReview 刷新主单状态
func (c *Controller) updateTerminatedGpuOrders(kt *kit.Kit, orderIDs []string) error {
	if len(orderIDs) == 0 {
		return nil
	}

	// 先筛选出所有子单都已终止的主单，直接置为"已终止"
	terminateReq := &rpproto.ResPlanDemandGpuOrderBatchUpdateReq{
		Items: make([]rpproto.ResPlanDemandGpuOrderUpdateReq, 0),
	}
	remainingOrderIDs := make([]string, 0, len(orderIDs))

	for _, orderID := range orderIDs {
		nonTerminateCount, err := c.countGpuSubOrders(kt, tools.ExpressionAnd(
			tools.RuleEqual("order_id", orderID),
			tools.RuleNotEqual("status", enumor.RPDemandGPUSubOrderStatusTerminate),
		))
		if err != nil {
			return err
		}

		if nonTerminateCount == 0 {
			// 所有子单都已终止，主单变为"已终止"
			terminateReq.Items = append(terminateReq.Items, rpproto.ResPlanDemandGpuOrderUpdateReq{
				ID:     orderID,
				Status: enumor.ResPlanDemandGpuOrderStatusTerminate,
			})
		} else {
			remainingOrderIDs = append(remainingOrderIDs, orderID)
		}
	}

	if len(terminateReq.Items) > 0 {
		if err := c.client.DataService().Global.ResourcePlan.BatchUpdateResPlanDemandGpuOrder(
			kt, terminateReq); err != nil {
			return err
		}
	}

	// 非全部终止的主单，复用评审后的状态刷新逻辑
	return c.RefreshGpuOrderStatusAfterReview(kt, remainingOrderIDs)
}

// RefreshGpuOrderStatusAfterBizEdit 业务侧修改驳回子单后，根据子单状态重新聚合主单状态。
// 仅对当前处于"已驳回"或"全部驳回"的主单生效，复用 resolveGpuOrderStatusBySubOrders 统一计算目标状态。
func (c *Controller) RefreshGpuOrderStatusAfterBizEdit(kt *kit.Kit, orderIDs []string) error {
	if len(orderIDs) == 0 {
		return nil
	}

	orderMap, err := c.listGpuOrdersByIDs(kt, orderIDs, []string{"id", "status"})
	if err != nil {
		return err
	}

	updateReq := &rpproto.ResPlanDemandGpuOrderBatchUpdateReq{
		Items: make([]rpproto.ResPlanDemandGpuOrderUpdateReq, 0, len(orderIDs)),
	}
	for _, orderID := range orderIDs {
		order, exists := orderMap[orderID]
		if !exists {
			return fmt.Errorf("gpu demand order(%s) not found", orderID)
		}

		// 只有"已驳回"或"全部驳回"的主单才需要更新
		if order.Status != enumor.ResPlanDemandGpuOrderStatusReject &&
			order.Status != enumor.ResPlanDemandGpuOrderStatusRejectAll {
			continue
		}

		// 统一复用子单聚合逻辑，正确处理部分驳回行被修改后的混合态
		newStatus, err := c.resolveGpuOrderStatusBySubOrders(kt, orderID)
		if err != nil {
			return err
		}
		if newStatus == order.Status {
			continue
		}
		updateReq.Items = append(updateReq.Items, rpproto.ResPlanDemandGpuOrderUpdateReq{
			ID:     orderID,
			Status: newStatus,
		})
	}

	if len(updateReq.Items) == 0 {
		return nil
	}

	return c.client.DataService().Global.ResourcePlan.BatchUpdateResPlanDemandGpuOrder(kt, updateReq)
}

// RefreshGpuOrderStatusAfterReview 评审后根据子单状态聚合刷新主单状态。
func (c *Controller) RefreshGpuOrderStatusAfterReview(kt *kit.Kit, orderIDs []string) error {
	if len(orderIDs) == 0 {
		return nil
	}

	// 先查询主单当前状态，避免对 DONE / TERMINATE / INIT 等不应变更的主单进行无效更新
	orderMap, err := c.listGpuOrdersByIDs(kt, orderIDs, []string{"id", "status"})
	if err != nil {
		return err
	}

	updateReq := &rpproto.ResPlanDemandGpuOrderBatchUpdateReq{
		Items: make([]rpproto.ResPlanDemandGpuOrderUpdateReq, 0, len(orderIDs)),
	}
	for _, orderID := range orderIDs {
		order, exists := orderMap[orderID]
		if !exists {
			return fmt.Errorf("gpu demand order(%s) not found", orderID)
		}

		// 已终止/已评审/待评审的主单不应被本函数变更
		if order.Status == enumor.ResPlanDemandGpuOrderStatusTerminate ||
			order.Status == enumor.ResPlanDemandGpuOrderStatusDone ||
			order.Status == enumor.ResPlanDemandGpuOrderStatusInit {
			continue
		}

		newStatus, err := c.resolveGpuOrderStatusBySubOrders(kt, orderID)
		if err != nil {
			return err
		}
		if newStatus == order.Status {
			continue
		}
		updateReq.Items = append(updateReq.Items, rpproto.ResPlanDemandGpuOrderUpdateReq{
			ID:     orderID,
			Status: newStatus,
		})
	}

	if len(updateReq.Items) == 0 {
		return nil
	}

	return c.client.DataService().Global.ResourcePlan.BatchUpdateResPlanDemandGpuOrder(kt, updateReq)
}

// AuditGpuSubOrderUpdates 统一提交 GPU 需求子单变更审计
func (c *Controller) AuditGpuSubOrderUpdates(kt *kit.Kit, updates []protoaudit.CloudResourceUpdateInfo) error {
	if len(updates) == 0 {
		return nil
	}

	auditReq := protoaudit.CloudResourceUpdateAuditReq{Updates: updates}
	if err := c.client.DataService().Global.Audit.CloudResourceUpdateAudit(kt.Ctx, kt.Header(), &auditReq); err != nil {
		logs.Errorf("audit gpu demand suborder update failed, err: %v, rid: %s", err, kt.Rid)
		return err
	}

	return nil
}

// resolveGpuOrderStatusBySubOrders 根据子单状态统计，计算主单应当变更到的目标状态。
func (c *Controller) resolveGpuOrderStatusBySubOrders(kt *kit.Kit, orderID string) (
	enumor.ResPlanDemandGpuOrderStatus, error) {

	rejectCount, err := c.countGpuSubOrders(kt, tools.ExpressionAnd(
		tools.RuleEqual("order_id", orderID),
		tools.RuleEqual("status", enumor.RPDemandGPUSubOrderStatusReject),
	))
	if err != nil {
		return "", err
	}

	// 有驳回子单 → "部分已驳回"或"全部已驳回"
	if rejectCount > 0 {
		nonRejectCount, err := c.countGpuSubOrders(kt, tools.ExpressionAnd(
			tools.RuleEqual("order_id", orderID),
			tools.RuleNotEqual("status", enumor.RPDemandGPUSubOrderStatusReject),
		))
		if err != nil {
			return "", err
		}
		if nonRejectCount == 0 {
			return enumor.ResPlanDemandGpuOrderStatusRejectAll, nil
		}
		return enumor.ResPlanDemandGpuOrderStatusReject, nil
	}

	// 无驳回子单：DONE / TERMINATE 视为已结束，仅 INIT / PENDING 阻止主单完成
	nonDoneCount, err := c.countGpuSubOrders(kt, tools.ExpressionAnd(
		tools.RuleEqual("order_id", orderID),
		tools.RuleNotIn("status", []enumor.RPDemandGPUSubOrderStatus{
			enumor.RPDemandGPUSubOrderStatusDone,
			enumor.RPDemandGPUSubOrderStatusTerminate,
		}),
	))
	if err != nil {
		return "", err
	}

	if nonDoneCount == 0 {
		return enumor.ResPlanDemandGpuOrderStatusDone, nil
	}
	return enumor.ResPlanDemandGpuOrderStatusPending, nil
}

func (c *Controller) listGpuOrdersByIDs(kt *kit.Kit, orderIDs []string, fields []string) (
	map[string]resplandemandgpuorder.ResPlanDemandGpuOrderTable, error) {

	resp, err := c.client.DataService().Global.ResourcePlan.ListResPlanDemandGpuOrder(kt,
		&rpproto.ResPlanDemandGpuOrderListReq{
			ListReq: core.ListReq{
				Filter: tools.ContainersExpression("id", orderIDs),
				Page:   &core.BasePage{Start: 0, Limit: uint(len(orderIDs))},
				Fields: fields,
			},
		})
	if err != nil {
		return nil, err
	}

	orderMap := make(map[string]resplandemandgpuorder.ResPlanDemandGpuOrderTable, len(resp.Details))
	for _, detail := range resp.Details {
		orderMap[detail.ID] = detail
	}

	return orderMap, nil
}

func (c *Controller) listGpuSubOrdersByIDs(kt *kit.Kit, subOrderIDs []string, bizID *int64, fields []string) (
	map[string]gpusuborder.ResPlanDemandGpuSubOrderTable, error) {

	subOrderMap := make(map[string]gpusuborder.ResPlanDemandGpuSubOrderTable, len(subOrderIDs))
	for _, ids := range slice.Split(subOrderIDs, int(core.DefaultMaxPageLimit)) {
		rules := []*filter.AtomRule{tools.RuleIn("id", ids)}
		if bizID != nil {
			rules = append(rules, tools.RuleEqual("bk_biz_id", cvt.PtrToVal(bizID)))
		}

		listResp, err := c.client.DataService().Global.ResourcePlan.ListResPlanDemandGpuSubOrder(kt,
			&rpproto.ResPlanDemandGpuSubOrderListReq{
				ListReq: core.ListReq{
					Filter: tools.ExpressionAnd(rules...),
					Page:   &core.BasePage{Start: 0, Limit: uint(len(ids))},
					Fields: fields,
				},
			})
		if err != nil {
			return nil, err
		}

		for _, detail := range listResp.Details {
			subOrderMap[detail.ID] = detail
		}
	}

	if len(subOrderMap) != len(subOrderIDs) {
		return nil, fmt.Errorf("some suborder_ids are invalid")
	}

	return subOrderMap, nil
}

func (c *Controller) countGpuSubOrders(kt *kit.Kit, expr *filter.Expression) (uint64, error) {
	resp, err := c.client.DataService().Global.ResourcePlan.ListResPlanDemandGpuSubOrder(kt,
		&rpproto.ResPlanDemandGpuSubOrderListReq{ListReq: core.ListReq{Filter: expr, Page: core.NewCountPage()}})
	if err != nil {
		return 0, err
	}

	return resp.Count, nil
}

func buildJsonFieldPtrFromArray(data []json.RawMessage) (*tabletypes.JsonField, error) {
	if data == nil {
		return nil, nil
	}

	field, err := tabletypes.NewJsonField(data)
	if err != nil {
		return nil, err
	}

	return &field, nil
}
