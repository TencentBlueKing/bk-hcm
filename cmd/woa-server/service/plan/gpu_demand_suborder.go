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
	"hcm/pkg/criteria/enumor"
	"hcm/pkg/criteria/errf"
	"hcm/pkg/dal/dao/tools"
	gpusuborder "hcm/pkg/dal/table/resource-plan/res-plan-demand-gpu-suborder"
	gputemplate "hcm/pkg/dal/table/resource-plan/res-plan-demand-gpu-template"
	tabletypes "hcm/pkg/dal/table/types"
	"hcm/pkg/iam/meta"
	"hcm/pkg/kit"
	"hcm/pkg/logs"
	"hcm/pkg/rest"
	"hcm/pkg/runtime/filter"
	cvt "hcm/pkg/tools/converter"
)

// ListResPlanDemandGpuSubOrder list resource-side GPU demand suborders.
func (s *service) ListResPlanDemandGpuSubOrder(cts *rest.Contexts) (interface{}, error) {
	req := new(core.ListReq)
	if err := cts.DecodeInto(req); err != nil {
		logs.Errorf("failed to decode list gpu demand suborders request, err: %v, rid: %s", err, cts.Kit.Rid)
		return nil, errf.NewFromErr(errf.DecodeRequestFailed, err)
	}

	if err := req.Validate(); err != nil {
		logs.Errorf("failed to validate list gpu demand suborders request, err: %v, rid: %s", err, cts.Kit.Rid)
		return nil, errf.NewFromErr(errf.InvalidParameter, err)
	}

	authRes := meta.ResourceAttribute{Basic: &meta.Basic{Type: meta.ZiYanResPlanGPUDemands, Action: meta.Find}}
	if err := s.authorizer.AuthorizeWithPerm(cts.Kit, authRes); err != nil {
		return nil, err
	}

	return s.listResPlanDemandGpuSubOrder(cts.Kit, req)
}

// ListBizResPlanDemandGpuSubOrder list biz-side GPU demand suborders.
func (s *service) ListBizResPlanDemandGpuSubOrder(cts *rest.Contexts) (interface{}, error) {
	bizID, err := cts.PathParameter("bk_biz_id").Int64()
	if err != nil {
		return nil, errf.NewFromErr(errf.InvalidParameter, err)
	}

	req := new(core.ListReq)
	if err = cts.DecodeInto(req); err != nil {
		logs.Errorf("failed to decode list biz gpu demand suborders request, err: %v, rid: %s", err, cts.Kit.Rid)
		return nil, errf.NewFromErr(errf.DecodeRequestFailed, err)
	}

	if err = req.Validate(); err != nil {
		logs.Errorf("failed to validate list biz gpu demand suborders request, err: %v, rid: %s", err, cts.Kit.Rid)
		return nil, errf.NewFromErr(errf.InvalidParameter, err)
	}

	authRes := meta.ResourceAttribute{Basic: &meta.Basic{Type: meta.Biz, Action: meta.Access}, BizID: bizID}
	if err = s.authorizer.AuthorizeWithPerm(cts.Kit, authRes); err != nil {
		return nil, err
	}

	req.Filter, err = tools.And(req.Filter, tools.RuleEqual("bk_biz_id", bizID))
	if err != nil {
		logs.Errorf("failed to build biz gpu demand suborders filter, err: %v, biz: %d, rid: %s",
			err, bizID, cts.Kit.Rid)
		return nil, errf.NewFromErr(errf.InvalidParameter, err)
	}

	return s.listResPlanDemandGpuSubOrder(cts.Kit, req)
}

// BatchUpdateBizResPlanDemandGpuSubOrder batch update biz GPU demand suborders.
func (s *service) BatchUpdateBizResPlanDemandGpuSubOrder(cts *rest.Contexts) (interface{}, error) {
	bizID, err := cts.PathParameter("bk_biz_id").Int64()
	if err != nil {
		return nil, errf.NewFromErr(errf.InvalidParameter, err)
	}

	req := new(ptypes.BatchUpdateResPlanDemandGpuSubOrderReq)
	if err = cts.DecodeInto(req); err != nil {
		logs.Errorf("failed to decode batch update biz gpu demand suborders request, err: %v, rid: %s",
			err, cts.Kit.Rid)
		return nil, errf.NewFromErr(errf.DecodeRequestFailed, err)
	}
	if err = req.ValidateBiz(); err != nil {
		logs.Errorf("failed to validate batch update biz gpu demand suborders request, err: %v, rid: %s",
			err, cts.Kit.Rid)
		return nil, errf.NewFromErr(errf.InvalidParameter, err)
	}

	authRes := meta.ResourceAttribute{Basic: &meta.Basic{Type: meta.Biz, Action: meta.Access}, BizID: bizID}
	if err = s.authorizer.AuthorizeWithPerm(cts.Kit, authRes); err != nil {
		return nil, err
	}

	subOrderMap, err := s.listGpuSubOrdersByUpdateItems(cts.Kit, req.SubOrderData, &bizID,
		[]string{"id", "status", "order_id"})
	if err != nil {
		logs.Errorf("failed to list biz gpu demand suborders before update, err: %v, biz: %d, rid: %s",
			err, bizID, cts.Kit.Rid)
		return nil, errf.NewFromErr(errf.Aborted, err)
	}

	updateReq, affectedOrderIDs, updateAuditLogs, err := s.buildBizGpuSubOrderUpdateReqAndAudit(
		req.SubOrderData, subOrderMap)
	if err != nil {
		return nil, err
	}

	if err = s.planController.AuditGpuSubOrderUpdates(cts.Kit, updateAuditLogs); err != nil {
		logs.Errorf("audit biz batch update gpu demand suborders failed, err: %v, biz: %d, rid: %s",
			err, bizID, cts.Kit.Rid)
		return nil, errf.NewFromErr(errf.Aborted, err)
	}

	if err = s.client.DataService().Global.ResourcePlan.BatchUpdateResPlanDemandGpuSubOrder(
		cts.Kit, updateReq); err != nil {
		logs.Errorf("failed to batch update biz gpu demand suborders, err: %v, biz: %d, rid: %s",
			err, bizID, cts.Kit.Rid)
		return nil, errf.NewFromErr(errf.Aborted, err)
	}

	if err = s.planController.RefreshGpuOrderStatusAfterBizEdit(cts.Kit, affectedOrderIDs); err != nil {
		logs.Errorf("failed to refresh gpu demand orders status after biz edit, err: %v, biz: %d, rid: %s",
			err, bizID, cts.Kit.Rid)
		return nil, errf.NewFromErr(errf.Aborted, err)
	}

	return nil, nil
}

// buildBizGpuSubOrderUpdateReqAndAudit 遍历业务侧子单修改项，构建子单更新请求、受影响主单ID列表和审计日志。
func (s *service) buildBizGpuSubOrderUpdateReqAndAudit(items []ptypes.UpdateResPlanDemandGpuSubOrderItem,
	subOrderMap map[string]gpusuborder.ResPlanDemandGpuSubOrderTable) (
	*rpproto.ResPlanDemandGpuSubOrderBatchUpdateReq, []string, []protoaudit.CloudResourceUpdateInfo, error) {

	updateReq := &rpproto.ResPlanDemandGpuSubOrderBatchUpdateReq{
		SubOrders: make([]rpproto.ResPlanDemandGpuSubOrderUpdateReq, 0, len(items)),
	}
	affectedOrderIDs := make([]string, 0)
	affectedOrderIDSet := make(map[string]struct{})
	auditLogs := make([]protoaudit.CloudResourceUpdateInfo, 0, len(items))

	for _, item := range items {
		subOrder, exists := subOrderMap[item.SubOrderID]
		if !exists {
			return nil, nil, nil, errf.NewFromErr(errf.InvalidParameter,
				fmt.Errorf("suborder(%s) not found", item.SubOrderID))
		}
		extension, err := jsonFieldPtrFromArray(item.Extension)
		if err != nil {
			return nil, nil, nil, errf.NewFromErr(errf.InvalidParameter, err)
		}
		updateItem := rpproto.ResPlanDemandGpuSubOrderUpdateReq{
			ID:        subOrder.ID,
			Extension: extension,
		}
		updateFields := map[string]interface{}{"extension": item.Extension}
		setDemandFields(&updateItem, item, updateFields)

		switch subOrder.Status {
		case enumor.RPDemandGPUSubOrderStatusInit:
			// INIT 只更新记录本身，不修改状态。
		case enumor.RPDemandGPUSubOrderStatusReject:
			updateItem.Status = enumor.RPDemandGPUSubOrderStatusPending
			updateFields["status"] = enumor.RPDemandGPUSubOrderStatusPending
			if _, exists := affectedOrderIDSet[subOrder.OrderID]; !exists {
				affectedOrderIDSet[subOrder.OrderID] = struct{}{}
				affectedOrderIDs = append(affectedOrderIDs, subOrder.OrderID)
			}
		default:
			return nil, nil, nil, errf.NewFromErr(errf.InvalidParameter,
				fmt.Errorf("suborder(%s) status(%s) is not allowed to edit", subOrder.ID, subOrder.Status))
		}
		updateReq.SubOrders = append(updateReq.SubOrders, updateItem)
		auditLogs = append(auditLogs, protoaudit.CloudResourceUpdateInfo{
			ResType:      enumor.ResPlanGPUDemandsSuborderAuditResType,
			ResID:        subOrder.ID,
			UpdateFields: updateFields,
		})
	}

	return updateReq, affectedOrderIDs, auditLogs, nil
}

// BatchUpdateResPlanDemandGpuSubOrder batch update/review GPU demand suborders in scr.
func (s *service) BatchUpdateResPlanDemandGpuSubOrder(cts *rest.Contexts) (interface{}, error) {
	req := new(ptypes.BatchUpdateResPlanDemandGpuSubOrderReq)
	if err := cts.DecodeInto(req); err != nil {
		logs.Errorf("failed to decode batch update gpu demand suborders request, err: %v, rid: %s",
			err, cts.Kit.Rid)
		return nil, errf.NewFromErr(errf.DecodeRequestFailed, err)
	}
	if err := req.ValidateResource(); err != nil {
		logs.Errorf("failed to validate batch update gpu demand suborders request, err: %v, rid: %s",
			err, cts.Kit.Rid)
		return nil, errf.NewFromErr(errf.InvalidParameter, err)
	}

	authRes := meta.ResourceAttribute{Basic: &meta.Basic{Type: meta.ZiYanResPlanGPUDemands, Action: meta.Update}}
	if err := s.authorizer.AuthorizeWithPerm(cts.Kit, authRes); err != nil {
		return nil, err
	}

	subOrderMap, err := s.listGpuSubOrdersByUpdateItems(cts.Kit, req.SubOrderData, nil,
		[]string{"id", "status", "order_id"})
	if err != nil {
		logs.Errorf("failed to list gpu demand suborders before scr update, err: %v, rid: %s", err, cts.Kit.Rid)
		return nil, errf.NewFromErr(errf.Aborted, err)
	}

	updateReq, reviewOrderIDs, updateAuditLogs, err := s.buildScrGpuSubOrderUpdateReqAndAudit(
		req.SubOrderData, subOrderMap)
	if err != nil {
		return nil, err
	}

	if err = s.planController.AuditGpuSubOrderUpdates(cts.Kit, updateAuditLogs); err != nil {
		logs.Errorf("audit scr batch update gpu demand suborders failed, err: %v, rid: %s", err, cts.Kit.Rid)
		return nil, errf.NewFromErr(errf.Aborted, err)
	}

	if err = s.client.DataService().Global.ResourcePlan.BatchUpdateResPlanDemandGpuSubOrder(
		cts.Kit, updateReq); err != nil {
		logs.Errorf("failed to batch update gpu demand suborders in scr, err: %v, rid: %s", err, cts.Kit.Rid)
		return nil, errf.NewFromErr(errf.Aborted, err)
	}

	if err = s.planController.RefreshGpuOrderStatusAfterReview(cts.Kit, reviewOrderIDs); err != nil {
		logs.Errorf("failed to refresh gpu demand order status after review, err: %v, rid: %s", err, cts.Kit.Rid)
		return nil, errf.NewFromErr(errf.Aborted, err)
	}
	return nil, nil
}

// buildScrGpuSubOrderUpdateReqAndAudit 遍历资源侧子单修改/评审项，构建子单更新请求、受影响主单ID列表和审计日志。
func (s *service) buildScrGpuSubOrderUpdateReqAndAudit(items []ptypes.UpdateResPlanDemandGpuSubOrderItem,
	subOrderMap map[string]gpusuborder.ResPlanDemandGpuSubOrderTable) (
	*rpproto.ResPlanDemandGpuSubOrderBatchUpdateReq, []string, []protoaudit.CloudResourceUpdateInfo, error) {

	updateReq := &rpproto.ResPlanDemandGpuSubOrderBatchUpdateReq{
		SubOrders: make([]rpproto.ResPlanDemandGpuSubOrderUpdateReq, 0, len(items)),
	}
	reviewOrderIDs := make([]string, 0)
	reviewOrderIDSet := make(map[string]struct{})
	auditLogs := make([]protoaudit.CloudResourceUpdateInfo, 0, len(items))

	for _, item := range items {
		subOrder, exists := subOrderMap[item.SubOrderID]
		if !exists {
			return nil, nil, nil, errf.NewFromErr(errf.InvalidParameter,
				fmt.Errorf("suborder(%s) not found", item.SubOrderID))
		}
		updateItem := rpproto.ResPlanDemandGpuSubOrderUpdateReq{
			ID: subOrder.ID,
		}

		// 子单状态必须是"评审中"
		if subOrder.Status != enumor.RPDemandGPUSubOrderStatusPending {
			return nil, nil, nil, errf.NewFromErr(errf.InvalidParameter,
				fmt.Errorf("suborder(%s) status(%s) is not allowed to review or edit", subOrder.ID, subOrder.Status))
		}
		if item.Extension != nil {
			extension, err := jsonFieldPtrFromArray(item.Extension)
			if err != nil {
				return nil, nil, nil, errf.NewFromErr(errf.InvalidParameter, err)
			}
			updateItem.Extension = extension
			updateFields := map[string]interface{}{"extension": item.Extension}
			setDemandFields(&updateItem, item, updateFields)
			updateReq.SubOrders = append(updateReq.SubOrders, updateItem)
			auditLogs = append(auditLogs, protoaudit.CloudResourceUpdateInfo{
				ResType:      enumor.ResPlanGPUDemandsSuborderAuditResType,
				ResID:        subOrder.ID,
				UpdateFields: updateFields,
			})
			continue
		}

		comment, err := jsonFieldPtrFromArray(item.Comment)
		if err != nil {
			return nil, nil, nil, errf.NewFromErr(errf.InvalidParameter, err)
		}

		updateItem.Status = item.Status
		updateItem.Comment = comment
		updateReq.SubOrders = append(updateReq.SubOrders, updateItem)
		auditLogs = append(auditLogs, protoaudit.CloudResourceUpdateInfo{
			ResType:      enumor.ResPlanGPUDemandsSuborderAuditResType,
			ResID:        subOrder.ID,
			UpdateFields: map[string]interface{}{"status": item.Status, "comment": item.Comment},
		})
		if _, exists = reviewOrderIDSet[subOrder.OrderID]; !exists {
			reviewOrderIDSet[subOrder.OrderID] = struct{}{}
			reviewOrderIDs = append(reviewOrderIDs, subOrder.OrderID)
		}
	}

	return updateReq, reviewOrderIDs, auditLogs, nil
}

// BatchUpdateStatusResPlanDemandGpuSubOrder batch review GPU demand suborders in scr.
func (s *service) BatchUpdateStatusResPlanDemandGpuSubOrder(cts *rest.Contexts) (interface{}, error) {
	req := new(ptypes.BatchUpdateStatusResPlanDemandGpuSubOrderReq)
	if err := cts.DecodeInto(req); err != nil {
		logs.Errorf("failed to decode batch update gpu demand suborders status request, err: %v, rid: %s",
			err, cts.Kit.Rid)
		return nil, errf.NewFromErr(errf.DecodeRequestFailed, err)
	}
	if err := req.Validate(); err != nil {
		logs.Errorf("failed to validate batch update gpu demand suborders status request, err: %v, rid: %s",
			err, cts.Kit.Rid)
		return nil, errf.NewFromErr(errf.InvalidParameter, err)
	}

	authRes := meta.ResourceAttribute{Basic: &meta.Basic{Type: meta.ZiYanResPlanGPUDemands, Action: meta.Update}}
	if err := s.authorizer.AuthorizeWithPerm(cts.Kit, authRes); err != nil {
		return nil, err
	}

	if err := s.planController.BatchReviewGpuSubOrders(cts.Kit, req); err != nil {
		return nil, err
	}

	return nil, nil
}

// BatchTerminateBizResPlanDemandGpuSubOrder batch terminate biz GPU demand suborders.
func (s *service) BatchTerminateBizResPlanDemandGpuSubOrder(cts *rest.Contexts) (interface{}, error) {
	bizID, err := cts.PathParameter("bk_biz_id").Int64()
	if err != nil {
		return nil, errf.NewFromErr(errf.InvalidParameter, err)
	}

	req := new(ptypes.BatchTerminateResPlanDemandGpuSubOrderReq)
	if err = cts.DecodeInto(req); err != nil {
		logs.Errorf("failed to decode batch terminate biz gpu demand suborders request, err: %v, rid: %s",
			err, cts.Kit.Rid)
		return nil, errf.NewFromErr(errf.DecodeRequestFailed, err)
	}

	if err = req.Validate(); err != nil {
		logs.Errorf("failed to validate batch terminate biz gpu demand suborders request, err: %v, rid: %s",
			err, cts.Kit.Rid)
		return nil, errf.NewFromErr(errf.InvalidParameter, err)
	}

	authRes := meta.ResourceAttribute{Basic: &meta.Basic{Type: meta.Biz, Action: meta.Access}, BizID: bizID}
	if err = s.authorizer.AuthorizeWithPerm(cts.Kit, authRes); err != nil {
		return nil, err
	}

	if err = s.planController.TerminateBizGpuSubOrders(cts.Kit, bizID, req.SubOrderIDs); err != nil {
		return nil, errf.NewFromErr(errf.Aborted, err)
	}

	return nil, nil
}

func (s *service) listResPlanDemandGpuSubOrder(kt *kit.Kit, req *core.ListReq) (interface{}, error) {
	resp, err := s.client.DataService().Global.ResourcePlan.ListResPlanDemandGpuSubOrder(kt,
		&rpproto.ResPlanDemandGpuSubOrderListReq{ListReq: cvt.PtrToVal(req)})
	if err != nil {
		logs.Errorf("failed to list gpu demand suborders, err: %v, filter: %s, rid: %s",
			err, req.Filter.LogMarshal(), kt.Rid)
		return nil, errf.NewFromErr(errf.InvalidParameter, err)
	}

	result := &ptypes.ListResPlanDemandGpuSubOrderResp{
		Count:     resp.Count,
		Details:   make([]gpusuborder.ResPlanDemandGpuSubOrderTable, 0),
		TplConfig: make([]ptypes.ResPlanDemandGpuTplConfig, 0),
	}
	if req.Page.Count || len(resp.Details) == 0 {
		return result, nil
	}

	result.Details = resp.Details
	tplConfig, err := s.listDemandGpuSubOrderTplConfig(kt, resp.Details)
	if err != nil {
		logs.Errorf("failed to list gpu demand suborder tpl config, err: %v, rid: %s", err, kt.Rid)
		return nil, errf.NewFromErr(errf.InvalidParameter, err)
	}
	result.TplConfig = tplConfig

	return result, nil
}

// listDemandGpuSubOrderTplConfig 列出GPU需求子单的模板配置
func (s *service) listDemandGpuSubOrderTplConfig(kt *kit.Kit,
	subOrders []gpusuborder.ResPlanDemandGpuSubOrderTable) ([]ptypes.ResPlanDemandGpuTplConfig, error) {

	templateIDs := uniqueGpuDemandSubOrderTemplateIDs(subOrders)
	if len(templateIDs) == 0 {
		return make([]ptypes.ResPlanDemandGpuTplConfig, 0), nil
	}

	templateResp, err := s.client.DataService().Global.ResourcePlan.ListDemandGpuTemplate(kt,
		&rpproto.DemandGpuTemplateListReq{
			ListReq: core.ListReq{
				Filter: tools.ContainersExpression("id", templateIDs),
				Page:   &core.BasePage{Start: 0, Limit: uint(len(templateIDs))},
				Fields: []string{"id", "tpl_schema"},
			},
		})
	if err != nil {
		return nil, err
	}

	templateMap := make(map[string]gputemplate.ResPlanDemandGpuTemplateTable, len(templateResp.Details))
	for _, detail := range templateResp.Details {
		templateMap[detail.ID] = detail
	}

	orderedTemplates := make([]gputemplate.ResPlanDemandGpuTemplateTable, 0, len(templateIDs))
	for _, templateID := range templateIDs {
		tpl, exists := templateMap[templateID]
		if !exists {
			logs.Warnf("template(%s) not found, skipping tpl_config, rid: %s", templateID, kt.Rid)
			continue
		}
		orderedTemplates = append(orderedTemplates, tpl)
	}

	return buildGpuDemandTplConfig(orderedTemplates)
}

func uniqueGpuDemandSubOrderTemplateIDs(subOrders []gpusuborder.ResPlanDemandGpuSubOrderTable) []string {
	seen := make(map[string]struct{}, len(subOrders))
	result := make([]string, 0, len(subOrders))
	for _, detail := range subOrders {
		if _, exists := seen[detail.TemplateID]; exists || detail.TemplateID == "" {
			continue
		}

		seen[detail.TemplateID] = struct{}{}
		result = append(result, detail.TemplateID)
	}

	return result
}

func buildGpuDemandTplConfig(templates []gputemplate.ResPlanDemandGpuTemplateTable) (
	[]ptypes.ResPlanDemandGpuTplConfig, error) {

	result := make([]ptypes.ResPlanDemandGpuTplConfig, 0, len(templates))
	for _, tpl := range templates {
		sheets, err := extractGpuDemandTplSheets(tpl.TplSchema)
		if err != nil {
			return nil, fmt.Errorf("failed to extract tpl_schema sheets for template(%s): %w", tpl.ID, err)
		}
		result = append(result, ptypes.ResPlanDemandGpuTplConfig{
			ID:     tpl.ID,
			Sheets: sheets,
		})
	}

	return result, nil
}

// extractGpuDemandTplSheets 从 tpl_schema JSON 中提取 sheets 数组，
// tpl_schema 的存储格式为 {"sheets": [...]}, 这里只取内层的数组部分。
func extractGpuDemandTplSheets(schema tabletypes.JsonField) (json.RawMessage, error) {
	if schema.IsEmpty() {
		return nil, fmt.Errorf("tpl_schema is empty")
	}

	var wrapper struct {
		Sheets json.RawMessage `json:"sheets"`
	}
	if err := json.Unmarshal([]byte(schema), &wrapper); err != nil {
		return nil, fmt.Errorf("invalid tpl_schema JSON: %w", err)
	}
	if wrapper.Sheets == nil {
		return nil, fmt.Errorf("tpl_schema missing 'sheets' field")
	}

	return wrapper.Sheets, nil
}

func jsonFieldPtrFromArray(data []json.RawMessage) (*tabletypes.JsonField, error) {
	if data == nil {
		return nil, nil
	}

	field, err := tabletypes.NewJsonField(data)
	if err != nil {
		return nil, err
	}

	return &field, nil
}

func (s *service) listGpuSubOrdersByUpdateItems(kt *kit.Kit, items []ptypes.UpdateResPlanDemandGpuSubOrderItem,
	bizID *int64, fields []string) (map[string]gpusuborder.ResPlanDemandGpuSubOrderTable, error) {

	subOrderIDs, err := uniqueGpuSubOrderIDs(items)
	if err != nil {
		return nil, err
	}

	rules := []*filter.AtomRule{tools.RuleIn("id", subOrderIDs)}
	if bizID != nil {
		rules = append(rules, tools.RuleEqual("bk_biz_id", *bizID))
	}

	listResp, err := s.client.DataService().Global.ResourcePlan.ListResPlanDemandGpuSubOrder(kt,
		&rpproto.ResPlanDemandGpuSubOrderListReq{
			ListReq: core.ListReq{
				Filter: tools.ExpressionAnd(rules...),
				Page:   &core.BasePage{Start: 0, Limit: uint(len(subOrderIDs))},
				Fields: fields,
			},
		})
	if err != nil {
		return nil, err
	}

	if len(listResp.Details) != len(subOrderIDs) {
		return nil, fmt.Errorf("some suborder_ids are invalid")
	}

	subOrderMap := make(map[string]gpusuborder.ResPlanDemandGpuSubOrderTable, len(listResp.Details))
	for _, detail := range listResp.Details {
		subOrderMap[detail.ID] = detail
	}

	return subOrderMap, nil
}

func uniqueGpuSubOrderIDs(items []ptypes.UpdateResPlanDemandGpuSubOrderItem) ([]string, error) {
	seen := make(map[string]struct{}, len(items))
	subOrderIDs := make([]string, 0, len(items))
	for _, item := range items {
		if _, exists := seen[item.SubOrderID]; exists {
			return nil, fmt.Errorf("duplicate suborder_id: %s", item.SubOrderID)
		}

		seen[item.SubOrderID] = struct{}{}
		subOrderIDs = append(subOrderIDs, item.SubOrderID)
	}

	return subOrderIDs, nil
}

// setDemandFields 将请求中的需求字段设置到更新请求和审计字段中。
func setDemandFields(updateItem *rpproto.ResPlanDemandGpuSubOrderUpdateReq,
	item ptypes.UpdateResPlanDemandGpuSubOrderItem, updateFields map[string]interface{}) {

	if item.DemandYear != nil {
		updateItem.DemandYear = cvt.PtrToVal(item.DemandYear)
		updateFields["demand_year"] = cvt.PtrToVal(item.DemandYear)
	}

	if item.DemandMonth != nil {
		updateItem.DemandMonth = cvt.PtrToVal(item.DemandMonth)
		updateFields["demand_month"] = cvt.PtrToVal(item.DemandMonth)
	}

	if item.GPUNum != nil {
		updateItem.GPUNum = cvt.PtrToVal(item.GPUNum)
		updateFields["gpu_num"] = cvt.PtrToVal(item.GPUNum)
	}

	if item.QpmMax != nil {
		updateItem.QpmMax = cvt.PtrToVal(item.QpmMax)
		updateFields["qpm_max"] = cvt.PtrToVal(item.QpmMax)
	}
}
