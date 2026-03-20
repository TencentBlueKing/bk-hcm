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
	"slices"

	ptypes "hcm/cmd/woa-server/types/plan"
	"hcm/pkg/api/core"
	protoaudit "hcm/pkg/api/data-service/audit"
	rpproto "hcm/pkg/api/data-service/resource-plan"
	"hcm/pkg/criteria/constant"
	"hcm/pkg/criteria/enumor"
	"hcm/pkg/criteria/errf"
	"hcm/pkg/dal/dao/tools"
	rpgpu "hcm/pkg/dal/table/resource-plan/res-plan-demand-gpu-order"
	"hcm/pkg/iam/meta"
	"hcm/pkg/kit"
	"hcm/pkg/logs"
	"hcm/pkg/rest"
	"hcm/pkg/runtime/filter"
	"hcm/pkg/tools/slice"
)

// BatchSetResPlanDemandGpuOrderPending 资源下批量将主单改为"评审中"状态
func (s *service) BatchSetResPlanDemandGpuOrderPending(cts *rest.Contexts) (interface{}, error) {
	req := new(ptypes.BatchGpuOrderStatusReq)
	if err := cts.DecodeInto(req); err != nil {
		return nil, errf.NewFromErr(errf.DecodeRequestFailed, err)
	}

	if err := req.Validate(); err != nil {
		return nil, errf.NewFromErr(errf.InvalidParameter, err)
	}

	authRes := meta.ResourceAttribute{Basic: &meta.Basic{Type: meta.ZiYanResPlanGPUDemands, Action: meta.Update}}
	if err := s.authorizer.AuthorizeWithPerm(cts.Kit, authRes); err != nil {
		return nil, err
	}

	return nil, s.batchUpdateGpuOrderStatus(cts.Kit, req.OrderIDs,
		[]enumor.ResPlanDemandGpuOrderStatus{enumor.ResPlanDemandGpuOrderStatusInit},
		enumor.ResPlanDemandGpuOrderStatusPending, enumor.RPDemandGPUSubOrderStatusPending)
}

// BatchRejectResPlanDemandGpuOrder 资源下批量整单驳回
func (s *service) BatchRejectResPlanDemandGpuOrder(cts *rest.Contexts) (interface{}, error) {
	req := new(ptypes.BatchGpuOrderStatusReq)
	if err := cts.DecodeInto(req); err != nil {
		return nil, errf.NewFromErr(errf.DecodeRequestFailed, err)
	}

	if err := req.Validate(); err != nil {
		return nil, errf.NewFromErr(errf.InvalidParameter, err)
	}

	authRes := meta.ResourceAttribute{Basic: &meta.Basic{Type: meta.ZiYanResPlanGPUDemands, Action: meta.Update}}
	if err := s.authorizer.AuthorizeWithPerm(cts.Kit, authRes); err != nil {
		return nil, err
	}

	return nil, s.batchUpdateGpuOrderStatus(cts.Kit, req.OrderIDs,
		[]enumor.ResPlanDemandGpuOrderStatus{enumor.ResPlanDemandGpuOrderStatusPending},
		enumor.ResPlanDemandGpuOrderStatusRejectAll, enumor.RPDemandGPUSubOrderStatusReject)
}

// BatchTerminateResPlanDemandGpuOrder 资源下批量终止主单
func (s *service) BatchTerminateResPlanDemandGpuOrder(cts *rest.Contexts) (interface{}, error) {
	req := new(ptypes.BatchGpuOrderStatusReq)
	if err := cts.DecodeInto(req); err != nil {
		return nil, errf.NewFromErr(errf.DecodeRequestFailed, err)
	}

	if err := req.Validate(); err != nil {
		return nil, errf.NewFromErr(errf.InvalidParameter, err)
	}

	authRes := meta.ResourceAttribute{Basic: &meta.Basic{Type: meta.ZiYanResPlanGPUDemands, Action: meta.Update}}
	if err := s.authorizer.AuthorizeWithPerm(cts.Kit, authRes); err != nil {
		return nil, err
	}

	return nil, s.batchUpdateGpuOrderStatus(cts.Kit, req.OrderIDs,
		[]enumor.ResPlanDemandGpuOrderStatus{enumor.ResPlanDemandGpuOrderStatusPending},
		enumor.ResPlanDemandGpuOrderStatusTerminate, enumor.RPDemandGPUSubOrderStatusTerminate)
}

// BatchTerminateBizResPlanDemandGpuOrder 业务下批量终止主单
func (s *service) BatchTerminateBizResPlanDemandGpuOrder(cts *rest.Contexts) (interface{}, error) {
	bkBizID, err := cts.PathParameter("bk_biz_id").Int64()
	if err != nil {
		return nil, errf.NewFromErr(errf.DecodeRequestFailed, err)
	}

	req := new(ptypes.BatchGpuOrderStatusReq)
	if err = cts.DecodeInto(req); err != nil {
		return nil, errf.NewFromErr(errf.DecodeRequestFailed, err)
	}

	if err = req.Validate(); err != nil {
		return nil, errf.NewFromErr(errf.InvalidParameter, err)
	}

	bkBizIDs, err := s.bizLogics.ListAuthorizedBiz(cts.Kit)
	if err != nil {
		logs.Errorf("list authorized biz failed, err: %v, rid: %s", err, cts.Kit.Rid)
		return nil, err
	}

	if !slice.IsItemInSlice(bkBizIDs, bkBizID) {
		return nil, errf.Newf(errf.PermissionDenied, "no permission to access biz %d", bkBizID)
	}

	if err = s.validateGpuOrdersBelongToBiz(cts.Kit, req.OrderIDs, bkBizID); err != nil {
		return nil, err
	}

	return nil, s.batchUpdateGpuOrderStatus(cts.Kit, req.OrderIDs,
		[]enumor.ResPlanDemandGpuOrderStatus{
			enumor.ResPlanDemandGpuOrderStatusInit, enumor.ResPlanDemandGpuOrderStatusRejectAll,
		},
		enumor.ResPlanDemandGpuOrderStatusTerminate, enumor.RPDemandGPUSubOrderStatusTerminate)
}

// batchUpdateGpuOrderStatus 批量变更主单及子单状态的公共逻辑
// 步骤：查主单→校验前置状态→写审计→更新主单→查子单→分批更新子单
func (s *service) batchUpdateGpuOrderStatus(kt *kit.Kit, orderIDs []string,
	allowedFromStatuses []enumor.ResPlanDemandGpuOrderStatus, targetOrderStatus enumor.ResPlanDemandGpuOrderStatus,
	targetSubOrderStatus enumor.RPDemandGPUSubOrderStatus) error {

	if err := s.validateGpuOrderStatuses(kt, orderIDs, allowedFromStatuses); err != nil {
		return err
	}

	if err := s.auditGpuOrderStatusUpdate(kt, orderIDs, targetOrderStatus); err != nil {
		return err
	}

	items := make([]rpproto.ResPlanDemandGpuOrderUpdateReq, 0, len(orderIDs))
	for _, id := range orderIDs {
		items = append(items, rpproto.ResPlanDemandGpuOrderUpdateReq{ID: id, Status: targetOrderStatus})
	}
	if err := s.client.DataService().Global.ResourcePlan.BatchUpdateResPlanDemandGpuOrder(
		kt, &rpproto.ResPlanDemandGpuOrderBatchUpdateReq{Items: items}); err != nil {
		logs.Errorf("batch update gpu demand order status failed, err: %v, rid: %s", err, kt.Rid)
		return err
	}

	return s.batchUpdateSubOrderStatuses(kt, orderIDs, targetSubOrderStatus)
}

// validateGpuOrderStatuses 查询主单并校验前置状态合法性
func (s *service) validateGpuOrderStatuses(kt *kit.Kit, orderIDs []string,
	allowedFromStatuses []enumor.ResPlanDemandGpuOrderStatus) error {

	listReq := &rpproto.ResPlanDemandGpuOrderListReq{
		ListReq: core.ListReq{
			Filter: tools.ContainersExpression("id", orderIDs),
			Page:   core.NewDefaultBasePage(),
		},
	}
	result, err := s.client.DataService().Global.ResourcePlan.ListResPlanDemandGpuOrder(kt, listReq)
	if err != nil {
		logs.Errorf("list gpu demand orders failed, err: %v, rid: %s", err, kt.Rid)
		return err
	}

	if len(result.Details) != len(orderIDs) {
		return errf.Newf(errf.InvalidParameter, "some gpu demand orders not found")
	}

	for _, order := range result.Details {
		if !slices.Contains(allowedFromStatuses, order.Status) {
			return errf.Newf(errf.InvalidParameter,
				"order %s status %s is not in allowed statuses %v", order.ID, order.Status, allowedFromStatuses)
		}
	}

	return nil
}

// auditGpuOrderStatusUpdate 写入主单状态变更审计记录
func (s *service) auditGpuOrderStatusUpdate(kt *kit.Kit, orderIDs []string,
	targetStatus enumor.ResPlanDemandGpuOrderStatus) error {

	updateLogs := make([]protoaudit.CloudResourceUpdateInfo, 0, len(orderIDs))
	for _, id := range orderIDs {
		updateLogs = append(updateLogs, protoaudit.CloudResourceUpdateInfo{
			ResType:      enumor.ResPlanGPUDemandsOrderAuditResType,
			ResID:        id,
			UpdateFields: map[string]interface{}{"status": targetStatus},
		})
	}

	auditReq := protoaudit.CloudResourceUpdateAuditReq{Updates: updateLogs}
	if err := s.client.DataService().Global.Audit.CloudResourceUpdateAudit(kt.Ctx, kt.Header(), &auditReq); err != nil {
		logs.Errorf("audit gpu demand order status update failed, err: %v, rid: %s", err, kt.Rid)
		return err
	}

	return nil
}

// validateGpuOrdersBelongToBiz 校验所有主单均属于指定业务
func (s *service) validateGpuOrdersBelongToBiz(kt *kit.Kit, orderIDs []string, bkBizID int64) error {
	listReq := &rpproto.ResPlanDemandGpuOrderListReq{
		ListReq: core.ListReq{
			Filter: &filter.Expression{
				Op: filter.And,
				Rules: []filter.RuleFactory{
					tools.ContainersExpression("id", orderIDs),
					tools.RuleEqual("bk_biz_id", bkBizID),
				},
			},
			Page: core.NewDefaultBasePage(),
		},
	}
	result, err := s.client.DataService().Global.ResourcePlan.ListResPlanDemandGpuOrder(kt, listReq)
	if err != nil {
		logs.Errorf("list gpu demand orders failed, err: %v, rid: %s", err, kt.Rid)
		return err
	}

	if len(result.Details) != len(orderIDs) {
		return errf.Newf(errf.InvalidParameter, "some gpu demand orders do not belong to biz %d", bkBizID)
	}

	return nil
}

// batchUpdateSubOrderStatuses 查子单并分批更新子单状态
func (s *service) batchUpdateSubOrderStatuses(kt *kit.Kit, orderIDs []string,
	targetStatus enumor.RPDemandGPUSubOrderStatus) error {

	listReq := &rpproto.ResPlanDemandGpuSubOrderListReq{
		ListReq: core.ListReq{
			Filter: tools.ContainersExpression("order_id", orderIDs),
			Page:   core.NewDefaultBasePage(),
			Fields: []string{"id"},
		},
	}
	for {
		result, err := s.client.DataService().Global.ResourcePlan.ListResPlanDemandGpuSubOrder(kt, listReq)
		if err != nil {
			logs.Errorf("list gpu demand sub orders failed, err: %v, rid: %s", err, kt.Rid)
			return err
		}

		for _, batch := range slice.Split(result.Details, constant.BatchOperationMaxLimit) {
			ids := make([]string, 0, len(batch))
			for _, item := range batch {
				ids = append(ids, item.ID)
			}
			updateReq := &rpproto.ResPlanDemandGpuSubOrderBatchUpdateStatusReq{
				IDs:    ids,
				Status: targetStatus,
			}
			err = s.client.DataService().Global.ResourcePlan.BatchUpdateStatusResPlanDemandGpuSubOrder(kt, updateReq)
			if err != nil {
				logs.Errorf("batch update gpu demand sub order status failed, err: %v, rid: %s", err, kt.Rid)
				return err
			}
		}

		if len(result.Details) < int(listReq.Page.Limit) {
			break
		}
		listReq.Page.Start += uint32(listReq.Page.Limit)
	}

	return nil
}

// ListResPlanDemandGpuOrder 资源视角查询GPU需求主单列表（含子单聚合字段）
func (s *service) ListResPlanDemandGpuOrder(cts *rest.Contexts) (interface{}, error) {
	req := new(ptypes.ListGpuDemandOrderReq)
	if err := cts.DecodeInto(req); err != nil {
		return nil, errf.NewFromErr(errf.DecodeRequestFailed, err)
	}

	if err := req.Validate(); err != nil {
		return nil, errf.NewFromErr(errf.InvalidParameter, err)
	}

	authRes := meta.ResourceAttribute{Basic: &meta.Basic{Type: meta.ZiYanResPlanGPUDemands, Action: meta.Find}}
	if err := s.authorizer.AuthorizeWithPerm(cts.Kit, authRes); err != nil {
		return nil, err
	}

	return s.listResPlanDemandGpuOrder(cts.Kit, req.Filter, req.Page)
}

// ListBizResPlanDemandGpuOrder 业务视角查询GPU需求主单列表（含子单聚合字段）
func (s *service) ListBizResPlanDemandGpuOrder(cts *rest.Contexts) (interface{}, error) {
	bkBizID, err := cts.PathParameter("bk_biz_id").Int64()
	if err != nil {
		return nil, errf.NewFromErr(errf.DecodeRequestFailed, err)
	}

	req := new(ptypes.ListGpuDemandOrderReq)
	if err = cts.DecodeInto(req); err != nil {
		return nil, errf.NewFromErr(errf.DecodeRequestFailed, err)
	}

	if err = req.Validate(); err != nil {
		return nil, errf.NewFromErr(errf.InvalidParameter, err)
	}

	bkBizIDs, err := s.bizLogics.ListAuthorizedBiz(cts.Kit)
	if err != nil {
		logs.Errorf("list authorized biz failed, err: %v, rid: %s", err, cts.Kit.Rid)
		return nil, err
	}

	if !slice.IsItemInSlice(bkBizIDs, bkBizID) {
		return nil, errf.Newf(errf.PermissionDenied, "no permission to access biz %d", bkBizID)
	}

	// 注入 bk_biz_id 过滤条件，确保只查该业务下的主单
	bizFilter := &filter.Expression{
		Op: filter.And,
		Rules: []filter.RuleFactory{
			req.Filter,
			tools.RuleEqual("bk_biz_id", bkBizID),
		},
	}

	return s.listResPlanDemandGpuOrder(cts.Kit, bizFilter, req.Page)
}

// listResPlanDemandGpuOrder 两阶段查询主单列表并附加子单聚合字段
func (s *service) listResPlanDemandGpuOrder(kt *kit.Kit, filterExpr *filter.Expression,
	page *core.BasePage) (*ptypes.ListGpuDemandOrderResult, error) {

	listReq := &rpproto.ResPlanDemandGpuOrderListReq{
		ListReq: core.ListReq{Filter: filterExpr, Page: page},
	}
	result, err := s.client.DataService().Global.ResourcePlan.ListResPlanDemandGpuOrder(kt, listReq)
	if err != nil {
		logs.Errorf("list gpu demand orders failed, err: %v, rid: %s", err, kt.Rid)
		return nil, err
	}

	// count 模式下直接返回，不查子单
	if page.Count {
		return &ptypes.ListGpuDemandOrderResult{Count: result.Count}, nil
	}

	if len(result.Details) == 0 {
		return &ptypes.ListGpuDemandOrderResult{Details: []ptypes.GpuDemandOrderItem{}}, nil
	}

	orderIDs := make([]string, 0, len(result.Details))
	for _, order := range result.Details {
		orderIDs = append(orderIDs, order.ID)
	}

	statsMap, err := s.fetchSubOrderStats(kt, orderIDs)
	if err != nil {
		return nil, err
	}

	return &ptypes.ListGpuDemandOrderResult{Details: assembleGpuOrderItems(result.Details, statsMap)}, nil
}

// subOrderStats 子单聚合数据
type subOrderStats struct {
	TotalGpuNum int64
	TotalQpmMax int64
}

// fetchSubOrderStats 批量查子单明细并按 order_id 聚合 gpu_num, qpm_max
func (s *service) fetchSubOrderStats(kt *kit.Kit, orderIDs []string) (map[string]*subOrderStats, error) {
	statsMap := make(map[string]*subOrderStats, len(orderIDs))

	listReq := &rpproto.ResPlanDemandGpuSubOrderListReq{
		ListReq: core.ListReq{
			Filter: tools.ContainersExpression("order_id", orderIDs),
			Page:   core.NewDefaultBasePage(),
		},
	}
	for {
		result, err := s.client.DataService().Global.ResourcePlan.ListResPlanDemandGpuSubOrder(kt, listReq)
		if err != nil {
			logs.Errorf("list gpu demand sub orders failed, err: %v, rid: %s", err, kt.Rid)
			return nil, err
		}

		for _, sub := range result.Details {
			if _, ok := statsMap[sub.OrderID]; !ok {
				statsMap[sub.OrderID] = &subOrderStats{}
			}
			statsMap[sub.OrderID].TotalGpuNum += sub.GPUNum
			statsMap[sub.OrderID].TotalQpmMax += sub.QpmMax
		}

		if len(result.Details) < int(listReq.Page.Limit) {
			break
		}
		listReq.Page.Start += uint32(listReq.Page.Limit)
	}

	return statsMap, nil
}

// assembleGpuOrderItems 将主单列表与子单聚合数据拼装为响应结构
func assembleGpuOrderItems(orders []rpgpu.ResPlanDemandGpuOrderTable,
	statsMap map[string]*subOrderStats) []ptypes.GpuDemandOrderItem {

	items := make([]ptypes.GpuDemandOrderItem, 0, len(orders))
	for _, order := range orders {
		item := ptypes.GpuDemandOrderItem{
			ID:            order.ID,
			BkBizID:       order.BkBizID,
			OpProductID:   order.OpProductID,
			OpProductName: order.OpProductName,
			TemplateID:    order.TemplateID,
			Status:        order.Status,
			Remark:        order.Remark,
			Creator:       order.Creator,
			Reviser:       order.Reviser,
			CreatedAt:     order.CreatedAt,
			UpdatedAt:     order.UpdatedAt,
		}
		if stats, ok := statsMap[order.ID]; ok {
			item.TotalGpuNum = stats.TotalGpuNum
			item.TotalQpmMax = stats.TotalQpmMax
		}
		items = append(items, item)
	}

	return items
}
