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
	"fmt"
	"io"
	"net/http"
	"strings"

	"hcm/pkg/api/core"
	protoaudit "hcm/pkg/api/data-service/audit"
	rpproto "hcm/pkg/api/data-service/resource-plan"
	woaapi "hcm/pkg/api/woa-server"
	"hcm/pkg/criteria/constant"
	"hcm/pkg/criteria/enumor"
	"hcm/pkg/criteria/errf"
	"hcm/pkg/dal/dao/tools"
	rpgpuorder "hcm/pkg/dal/table/resource-plan/res-plan-demand-gpu-order"
	"hcm/pkg/iam/meta"
	"hcm/pkg/kit"
	"hcm/pkg/logs"
	"hcm/pkg/rest"
	"hcm/pkg/tools/converter"
	toolexcel "hcm/pkg/tools/excel"
	toolsjson "hcm/pkg/tools/json"

	"github.com/xuri/excelize/v2"
)

// ExcelImportGpuDemand GPU需求Excel导入预览
func (s *service) ExcelImportGpuDemand(cts *rest.Contexts) (interface{}, error) {
	bkBizID, err := cts.PathParameter("bk_biz_id").Int64()
	if err != nil {
		return nil, errf.NewFromErr(errf.DecodeRequestFailed, err)
	}

	authRes := meta.ResourceAttribute{
		Basic: &meta.Basic{Type: meta.Biz, Action: meta.Access},
		BizID: bkBizID,
	}
	if err = s.authorizer.AuthorizeWithPerm(cts.Kit, authRes); err != nil {
		return nil, err
	}

	// 设置允许传入的最大excel文件为50MB
	cts.Request.Request.Body = http.MaxBytesReader(nil, cts.Request.Request.Body, constant.MaxExcelFileSize)
	file, _, err := cts.Request.Request.FormFile("file")
	if err != nil {
		logs.Errorf("failed to get upload file, err: %v, rid: %s", err, cts.Kit.Rid)
		return nil, errf.NewFromErr(errf.InvalidParameter, err)
	}
	defer file.Close()

	result, err := s.parseGpuDemandExcel(cts.Kit, file)
	if err != nil {
		logs.Errorf("failed to parse gpu demand excel, err: %v, rid: %s", err, cts.Kit.Rid)
		return nil, errf.NewFromErr(errf.Aborted, err)
	}

	return result, nil
}

// parseGpuDemandExcel 解析GPU需求Excel文件，校验表结构并返回解析结果。
func (s *service) parseGpuDemandExcel(kt *kit.Kit, reader io.Reader) (*woaapi.GpuDemandExcelImportResp, error) {
	schema, err := s.latestGpuTplSchema(kt)
	if err != nil {
		return nil, err
	}

	excelFile, err := excelize.OpenReader(reader)
	if err != nil {
		logs.Errorf("failed to open excel file, err: %v, rid: %s", err, kt.Rid)
		return nil, fmt.Errorf("invalid excel file format: %w", err)
	}
	defer excelFile.Close()

	if err = toolexcel.ValidateFileIntegrity(excelFile, schema); err != nil {
		logs.Errorf("excel file integrity validation failed, err: %v, rid: %s", err, kt.Rid)
		return nil, err
	}

	var allDetails []woaapi.GpuDemandExcelImportDetail
	for i := range schema.Sheets {
		rows, formulaErrs, parseErr := toolexcel.ParseSheetRowsAndFormulas(excelFile, &schema.Sheets[i])
		if parseErr != nil {
			logs.Errorf("failed to parse sheet[%s] data, err: %v, rid: %s",
				schema.Sheets[i].Name, parseErr, kt.Rid)
			return nil, parseErr
		}

		details := buildDetails(&schema.Sheets[i], rows, formulaErrs)
		allDetails = append(allDetails, details...)
	}
	if len(allDetails) == 0 {
		logs.Errorf("the excel file is empty, rid: %s", kt.Rid)
		return nil, fmt.Errorf("the excel file is empty, must have at least one sheet and one row")
	}

	return &woaapi.GpuDemandExcelImportResp{
		Sheets:  schema.Sheets,
		Details: allDetails,
	}, nil
}

// latestGpuTplSchema 获取最新GPU模版的Schema（不返回模版ID）。
func (s *service) latestGpuTplSchema(kt *kit.Kit) (*toolexcel.Schema, error) {
	_, schema, err := s.latestGpuTemplate(kt)
	if err != nil {
		logs.Errorf("failed to get latest gpu template schema, err: %v, rid: %s", err, kt.Rid)
		return nil, err
	}

	return schema, nil
}

// buildDetails 将解析后的行数据构建为 GpuDemandExcelImportDetail 列表。
// formulaErrs 与 rows 按索引对应，为 ParseSheetRowsAndFormulas 返回的每行公式校验错误。
func buildDetails(sheet *toolexcel.Sheet, rows [][]string, formulaErrs [][]string,
) []woaapi.GpuDemandExcelImportDetail {

	details := make([]woaapi.GpuDemandExcelImportDetail, 0, len(rows))
	allExcelHeaders := sheet.AllExcelHeaders()

	for rowIdx, row := range rows {
		rawData := make([]interface{}, 0, len(row))
		validateResult := make([]string, 0)

		for i, h := range allExcelHeaders {
			val := ""
			if i < len(row) {
				val = row[i]
			}

			if !h.Hidden {
				rawData = append(rawData, toolexcel.ConvertCellValue(val, h))
			}

			if errs := toolexcel.ValidateCellValue(val, h); len(errs) > 0 {
				validateResult = append(validateResult, errs...)
			}
		}

		validateResult = append(validateResult, formulaErrs[rowIdx]...)

		details = append(details, woaapi.GpuDemandExcelImportDetail{
			Name:           sheet.Name,
			RawData:        rawData,
			ValidateResult: validateResult,
		})
	}

	return details
}

// CreateGpuDemandOrder 创建GPU需求主单及子单
func (s *service) CreateGpuDemandOrder(cts *rest.Contexts) (interface{}, error) {
	bkBizID, err := cts.PathParameter("bk_biz_id").Int64()
	if err != nil {
		return nil, errf.NewFromErr(errf.DecodeRequestFailed, err)
	}

	req := new(woaapi.CreateGpuDemandOrderReq)
	if err = cts.DecodeInto(req); err != nil {
		logs.Errorf("failed to decode create gpu demand order request, err: %v, rid: %s",
			err, cts.Kit.Rid)
		return nil, errf.NewFromErr(errf.DecodeRequestFailed, err)
	}

	if err = req.Validate(); err != nil {
		logs.Errorf("failed to validate create gpu demand order request, err: %v, rid: %s",
			err, cts.Kit.Rid)
		return nil, errf.NewFromErr(errf.InvalidParameter, err)
	}

	authRes := meta.ResourceAttribute{
		Basic: &meta.Basic{Type: meta.Biz, Action: meta.Access},
		BizID: bkBizID,
	}
	if err = s.authorizer.AuthorizeWithPerm(cts.Kit, authRes); err != nil {
		return nil, err
	}

	orderID, err := s.createGpuDemandOrder(cts.Kit, bkBizID, req)
	if err != nil {
		logs.Errorf("failed to create gpu demand order, bkBizID: %d, err: %v, rid: %s",
			bkBizID, err, cts.Kit.Rid)
		return nil, errf.NewFromErr(errf.Aborted, err)
	}

	return map[string]interface{}{"id": orderID}, nil
}

// OverwriteGpuDemandOrder 覆盖上传已驳回的GPU需求提报主单
func (s *service) OverwriteGpuDemandOrder(cts *rest.Contexts) (interface{}, error) {
	bkBizID, err := cts.PathParameter("bk_biz_id").Int64()
	if err != nil {
		return nil, errf.NewFromErr(errf.DecodeRequestFailed, err)
	}

	req := new(woaapi.OverwriteGpuDemandOrderReq)
	if err = cts.DecodeInto(req); err != nil {
		logs.Errorf("failed to decode overwrite gpu demand order request, err: %v, rid: %s",
			err, cts.Kit.Rid)
		return nil, errf.NewFromErr(errf.DecodeRequestFailed, err)
	}

	if err = req.Validate(); err != nil {
		logs.Errorf("failed to validate overwrite gpu demand order request, err: %v, rid: %s",
			err, cts.Kit.Rid)
		return nil, errf.NewFromErr(errf.InvalidParameter, err)
	}

	authRes := meta.ResourceAttribute{
		Basic: &meta.Basic{Type: meta.Biz, Action: meta.Access},
		BizID: bkBizID,
	}
	if err = s.authorizer.AuthorizeWithPerm(cts.Kit, authRes); err != nil {
		return nil, err
	}

	if err = s.overwriteGpuDemandOrder(cts.Kit, bkBizID, req); err != nil {
		logs.Errorf("failed to overwrite gpu demand order, orderID: %s, bkBizID: %d, err: %v, rid: %s",
			req.OrderID, bkBizID, err, cts.Kit.Rid)
		return nil, errf.NewFromErr(errf.Aborted, err)
	}

	return nil, nil
}

// createGpuDemandOrder 创建GPU需求主单及子单，校验扩展字段后批量写入。
func (s *service) createGpuDemandOrder(kt *kit.Kit, bkBizID int64, req *woaapi.CreateGpuDemandOrderReq,
) (string, error) {

	templateID, schema, err := s.latestGpuTemplate(kt)
	if err != nil {
		return "", err
	}

	if err = validateOrderDetails(req.Details, schema); err != nil {
		logs.Errorf("failed to validate order details for create, err: %v, rid: %s", err, kt.Rid)
		return "", err
	}

	orderID, err := s.batchCreateGpuMainOrder(kt, bkBizID, req, templateID)
	if err != nil {
		return "", err
	}

	params := gpuOrderParams{
		OrderID:       orderID,
		OpProductID:   req.OpProductID,
		OpProductName: req.OpProductName,
		BkBizID:       bkBizID,
		TemplateID:    templateID,
	}
	if err = s.batchCreateGpuSubOrders(kt, params, req.Details); err != nil {
		return "", err
	}

	return orderID, nil
}

// overwriteGpuDemandOrder 覆盖已驳回的GPU需求主单，删除旧子单后重建。
func (s *service) overwriteGpuDemandOrder(kt *kit.Kit, bkBizID int64, req *woaapi.OverwriteGpuDemandOrderReq) error {
	order, err := s.validateOrderForOverwrite(kt, req.OrderID, bkBizID)
	if err != nil {
		return err
	}

	if err = s.validateSubOrderStatuses(kt, req.OrderID, order); err != nil {
		return err
	}

	templateID, schema, err := s.latestGpuTemplate(kt)
	if err != nil {
		return err
	}

	if err = validateOrderDetails(req.Details, schema); err != nil {
		logs.Errorf("failed to validate order details for overwrite, orderID: %s, err: %v, rid: %s",
			req.OrderID, err, kt.Rid)
		return err
	}

	params := gpuOrderParams{
		OrderID:       req.OrderID,
		OpProductID:   order.OpProductID,
		OpProductName: order.OpProductName,
		BkBizID:       bkBizID,
		TemplateID:    templateID,
	}
	if err = s.overwriteGpuSubOrders(kt, params, req.Details); err != nil {
		return err
	}

	s.auditGpuOrderUpdate(kt, req.OrderID, map[string]interface{}{
		"status": string(order.Status),
	})

	if err = s.resetGpuOrderStatus(kt, req.OrderID); err != nil {
		return err
	}

	return nil
}

// latestGpuTemplate 获取最新GPU模版ID及解析后的Schema。
func (s *service) latestGpuTemplate(kt *kit.Kit) (string, *toolexcel.Schema, error) {
	listReq := &rpproto.DemandGpuTemplateListReq{
		ListReq: core.ListReq{
			Filter: tools.AllExpression(),
			Page: &core.BasePage{
				Start: 0,
				Limit: 1,
				Sort:  "id",
				Order: core.Descending,
			},
		},
	}

	result, err := s.client.DataService().Global.ResourcePlan.ListDemandGpuTemplate(kt, listReq)
	if err != nil {
		logs.Errorf("failed to list gpu demand templates, err: %v, rid: %s", err, kt.Rid)
		return "", nil, fmt.Errorf("failed to query gpu template: %w", err)
	}

	if len(result.Details) == 0 {
		logs.Errorf("no gpu demand template found, rid: %s", kt.Rid)
		return "", nil, fmt.Errorf("no gpu demand template found, please configure template first")
	}

	tpl := result.Details[0]
	schema := new(toolexcel.Schema)
	if err = toolsjson.Unmarshal([]byte(tpl.TplSchema), schema); err != nil {
		logs.Errorf("failed to unmarshal tpl_schema, err: %v, rid: %s", err, kt.Rid)
		return "", nil, fmt.Errorf("invalid template schema format: %w", err)
	}

	return tpl.ID, schema, nil
}

// gpuOrderParams 创建GPU子单时需要的主单上下文参数。
type gpuOrderParams struct {
	OrderID       string
	OpProductID   int64
	OpProductName string
	BkBizID       int64
	TemplateID    string
}

// batchCreateGpuMainOrder 通过data-service创建GPU需求主单记录。
func (s *service) batchCreateGpuMainOrder(kt *kit.Kit, bkBizID int64, req *woaapi.CreateGpuDemandOrderReq,
	templateID string,
) (string, error) {

	createReq := &rpproto.ResPlanDemandGpuOrderBatchCreateReq{
		Items: []rpproto.ResPlanDemandGpuOrderCreateReq{{
			BkBizID:       bkBizID,
			OpProductID:   req.OpProductID,
			OpProductName: req.OpProductName,
			TemplateID:    templateID,
			Status:        enumor.ResPlanDemandGpuOrderStatusInit,
		}},
	}

	result, err := s.client.DataService().Global.ResourcePlan.BatchCreateResPlanDemandGpuOrder(kt, createReq)
	if err != nil {
		logs.Errorf("failed to create gpu demand main order, err: %v, rid: %s", err, kt.Rid)
		return "", fmt.Errorf("failed to create gpu demand main order: %w", err)
	}

	if len(result.IDs) == 0 {
		logs.Errorf("gpu demand main order created but no ID returned, rid: %s", kt.Rid)
		return "", fmt.Errorf("gpu demand main order created but no ID returned")
	}

	return result.IDs[0], nil
}

// batchCreateGpuSubOrders 批量创建GPU需求子单。
func (s *service) batchCreateGpuSubOrders(kt *kit.Kit, params gpuOrderParams,
	details []woaapi.CreateGpuDemandOrderDetail,
) error {

	subOrders := make([]rpproto.ResPlanDemandGpuSubOrderCreateReq, 0, len(details))
	for _, d := range details {
		subOrders = append(subOrders, rpproto.ResPlanDemandGpuSubOrderCreateReq{
			OrderID:       params.OrderID,
			OpProductID:   params.OpProductID,
			OpProductName: params.OpProductName,
			BkBizID:       params.BkBizID,
			DemandType:    d.DemandType,
			TemplateID:    params.TemplateID,
			DemandYear:    d.DemandYear,
			DemandMonth:   d.DemandMonth,
			GPUNum:        d.GPUNum,
			QpmMax:        d.QpmMax,
			Status:        enumor.RPDemandGPUSubOrderStatusInit,
			Extension:     d.Extension,
		})
	}

	createReq := &rpproto.ResPlanDemandGpuSubOrderBatchCreateReq{SubOrders: subOrders}
	if _, err := s.client.DataService().Global.ResourcePlan.BatchCreateResPlanDemandGpuSubOrder(
		kt, createReq); err != nil {
		logs.Errorf("failed to batch create gpu demand sub orders, orderID: %s, err: %v, rid: %s",
			params.OrderID, err, kt.Rid)
		return fmt.Errorf("failed to create gpu demand sub orders: %w", err)
	}

	return nil
}

// overwriteGpuSubOrders atomically replaces all sub orders of the given order via a single data-service call.
func (s *service) overwriteGpuSubOrders(kt *kit.Kit, params gpuOrderParams, details []woaapi.CreateGpuDemandOrderDetail,
) error {

	subOrders := make([]rpproto.ResPlanDemandGpuSubOrderCreateReq, 0, len(details))
	for _, d := range details {
		subOrders = append(subOrders, rpproto.ResPlanDemandGpuSubOrderCreateReq{
			OrderID:       params.OrderID,
			OpProductID:   params.OpProductID,
			OpProductName: params.OpProductName,
			BkBizID:       params.BkBizID,
			DemandType:    d.DemandType,
			TemplateID:    params.TemplateID,
			DemandYear:    d.DemandYear,
			DemandMonth:   d.DemandMonth,
			GPUNum:        d.GPUNum,
			QpmMax:        d.QpmMax,
			Status:        enumor.RPDemandGPUSubOrderStatusInit,
			Extension:     d.Extension,
		})
	}

	overwriteReq := &rpproto.ResPlanDemandGpuSubOrderOverwriteReq{
		OrderID:   params.OrderID,
		SubOrders: subOrders,
	}
	if _, err := s.client.DataService().Global.ResourcePlan.OverwriteResPlanDemandGpuSubOrders(
		kt, overwriteReq); err != nil {
		logs.Errorf("failed to overwrite gpu demand sub orders, orderID: %s, err: %v, rid: %s",
			params.OrderID, err, kt.Rid)
		return fmt.Errorf("failed to overwrite gpu demand sub orders: %w", err)
	}

	return nil
}

// resetGpuOrderStatus 将GPU需求主单状态重置为INIT。
func (s *service) resetGpuOrderStatus(kt *kit.Kit, orderID string) error {
	updateReq := &rpproto.ResPlanDemandGpuOrderBatchUpdateReq{
		Items: []rpproto.ResPlanDemandGpuOrderUpdateReq{{
			ID:     orderID,
			Status: enumor.ResPlanDemandGpuOrderStatusInit,
		}},
	}
	if err := s.client.DataService().Global.ResourcePlan.BatchUpdateResPlanDemandGpuOrder(
		kt, updateReq); err != nil {
		logs.Errorf("failed to reset gpu demand order status, orderID: %s, err: %v, rid: %s",
			orderID, err, kt.Rid)
		return fmt.Errorf("failed to reset gpu demand order status: %w", err)
	}

	return nil
}

// auditGpuOrderUpdate 记录GPU需求主单更新审计，审计失败仅记日志不阻断业务。
func (s *service) auditGpuOrderUpdate(kt *kit.Kit, orderID string, fields map[string]interface{}) {
	auditReq := &protoaudit.CloudResourceUpdateAuditReq{
		Updates: []protoaudit.CloudResourceUpdateInfo{{
			ResType:      enumor.ResPlanGPUDemandsOrderAuditResType,
			ResID:        orderID,
			UpdateFields: fields,
		}},
	}
	if err := s.client.DataService().Global.Audit.CloudResourceUpdateAudit(
		kt.Ctx, kt.Header(), auditReq); err != nil {
		logs.Errorf("failed to create gpu order update audit, orderID: %s, err: %v, rid: %s",
			orderID, err, kt.Rid)
	}
}

// validateOrderForOverwrite 校验主单存在性、业务归属及状态是否允许覆盖。
func (s *service) validateOrderForOverwrite(kt *kit.Kit, orderID string, bkBizID int64) (
	*rpgpuorder.ResPlanDemandGpuOrderTable, error,
) {

	listReq := &rpproto.ResPlanDemandGpuOrderListReq{
		ListReq: core.ListReq{
			Filter: tools.EqualExpression("id", orderID),
			Page:   &core.BasePage{Start: 0, Limit: 1},
		},
	}
	result, err := s.client.DataService().Global.ResourcePlan.ListResPlanDemandGpuOrder(kt, listReq)
	if err != nil {
		logs.Errorf("failed to get gpu demand order, orderID: %s, err: %v, rid: %s",
			orderID, err, kt.Rid)
		return nil, fmt.Errorf("failed to get gpu demand order: %w", err)
	}

	if len(result.Details) == 0 {
		logs.Errorf("gpu demand order %s not found, rid: %s", orderID, kt.Rid)
		return nil, fmt.Errorf("gpu demand order %s not found", orderID)
	}

	order := result.Details[0]
	if order.BkBizID != bkBizID {
		logs.Errorf("gpu demand order %s does not belong to biz %d, actual biz: %d, rid: %s",
			orderID, bkBizID, order.BkBizID, kt.Rid)
		return nil, fmt.Errorf("gpu demand order %s does not belong to biz %d", orderID, bkBizID)
	}

	if order.Status != enumor.ResPlanDemandGpuOrderStatusRejectAll &&
		order.Status != enumor.ResPlanDemandGpuOrderStatusInit {
		logs.Errorf("gpu demand order %s status is %s, not allowed to overwrite, rid: %s",
			orderID, order.Status, kt.Rid)
		return nil, fmt.Errorf(
			"gpu demand order %s status is %s, only REJECT_ALL and INIT status orders can be overwritten",
			orderID, order.Status)
	}

	return &order, nil
}

// validateSubOrderStatuses 校验主单下所有子单状态均符合覆盖条件（分页拉取全部子单）。
func (s *service) validateSubOrderStatuses(kt *kit.Kit, orderID string, order *rpgpuorder.ResPlanDemandGpuOrderTable,
) error {

	listReq := &rpproto.ResPlanDemandGpuSubOrderListReq{
		ListReq: core.ListReq{
			Filter: tools.EqualExpression("order_id", orderID),
			Page:   core.NewDefaultBasePage(),
		},
	}

	for {
		result, err := s.client.DataService().Global.ResourcePlan.ListResPlanDemandGpuSubOrder(kt, listReq)
		if err != nil {
			logs.Errorf("failed to list gpu demand sub orders, orderID: %s, err: %v, rid: %s",
				orderID, err, kt.Rid)
			return fmt.Errorf("failed to list gpu demand sub orders: %w", err)
		}

		for _, sub := range result.Details {
			switch order.Status {
			case enumor.ResPlanDemandGpuOrderStatusInit:
				if sub.Status != enumor.RPDemandGPUSubOrderStatusInit {
					logs.Errorf("sub order %s status is %s, expected INIT when order is INIT, "+
						"orderID: %s, rid: %s", sub.ID, sub.Status, orderID, kt.Rid)
					return fmt.Errorf(
						"sub order %s status is %s, only INIT sub orders can be overwritten "+
							"when order status is INIT",
						sub.ID, sub.Status)
				}
			case enumor.ResPlanDemandGpuOrderStatusRejectAll:
				if sub.Status != enumor.RPDemandGPUSubOrderStatusReject &&
					sub.Status != enumor.RPDemandGPUSubOrderStatusTerminate {
					logs.Errorf("sub order %s status is %s, expected REJECT or TERMINATE "+
						"when order is REJECT_ALL, orderID: %s, rid: %s",
						sub.ID, sub.Status, orderID, kt.Rid)
					return fmt.Errorf(
						"sub order %s status is %s, only REJECT or TERMINATE sub orders "+
							"can be overwritten when order status is REJECT_ALL",
						sub.ID, sub.Status)
				}
			default:
				logs.Errorf("unexpected order status %s for overwrite, orderID: %s, rid: %s",
					order.Status, orderID, kt.Rid)
				return fmt.Errorf(
					"order status is %s, only INIT or REJECT_ALL orders can be overwritten",
					order.Status)
			}
		}

		if len(result.Details) < int(listReq.Page.Limit) {
			break
		}
		listReq.Page.Start += uint32(listReq.Page.Limit)
	}

	return nil
}

// validateOrderDetails 遍历子单列表，按demand_type匹配schema中的sheet，
// 依次对固定字段和动态扩展字段进行校验，汇总所有错误。
func validateOrderDetails(details []woaapi.CreateGpuDemandOrderDetail, schema *toolexcel.Schema) error {
	var allErrs []string
	for i, d := range details {
		sheet := schema.FindSheet(d.DemandType)
		if sheet == nil {
			allErrs = append(allErrs,
				fmt.Sprintf("details[%d]: unknown demand_type %q", i, d.DemandType))
			continue
		}

		errs := validateDetailFixedFields(d, sheet.FixedHeaders)
		errs = append(errs, validateDetailExtension(d, sheet.Headers)...)

		for _, e := range errs {
			allErrs = append(allErrs, fmt.Sprintf("details[%d]: %s", i, e))
		}
	}

	if len(allErrs) > 0 {
		return fmt.Errorf("detail validation failed: %s", strings.Join(allErrs, "; "))
	}

	return nil
}

// validateDetailFixedFields 将detail序列化为map后按fixed_headers进行校验。
func validateDetailFixedFields(detail woaapi.CreateGpuDemandOrderDetail, headers []toolexcel.Header) []string {
	fieldMap, err := converter.StructToMap(detail)
	if err != nil {
		return []string{fmt.Sprintf("failed to convert detail to map: %v", err)}
	}

	return toolexcel.ValidateFixedFields(fieldMap, headers)
}

// validateDetailExtension 解析extension为值数组后按headers进行校验。
func validateDetailExtension(detail woaapi.CreateGpuDemandOrderDetail, headers []toolexcel.Header) []string {
	var values []interface{}
	if err := toolsjson.Unmarshal([]byte(detail.Extension), &values); err != nil {
		return []string{fmt.Sprintf("invalid extension format: %v", err)}
	}

	return toolexcel.ValidateExtension(values, headers)
}
