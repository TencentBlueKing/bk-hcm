/*
 * TencentBlueKing is pleased to support the open source community by making
 * 蓝鲸智云 - 混合云管理平台 (BlueKing - Hybrid Cloud Management System) available.
 * Copyright (C) 2024 THL A29 Limited,
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

package billadjustment

import (
	"fmt"
	"time"

	"hcm/cmd/account-server/logics/bill/export"
	"hcm/pkg/api/account-server/bill"
	"hcm/pkg/api/core"
	accountset "hcm/pkg/api/core/account-set"
	billcore "hcm/pkg/api/core/bill"
	dsbillapi "hcm/pkg/api/data-service/bill"
	"hcm/pkg/criteria/enumor"
	"hcm/pkg/criteria/errf"
	"hcm/pkg/dal/dao/tools"
	"hcm/pkg/iam/meta"
	"hcm/pkg/kit"
	"hcm/pkg/logs"
	"hcm/pkg/rest"
	"hcm/pkg/runtime/filter"
	"hcm/pkg/thirdparty/api-gateway/cmdb"
	"hcm/pkg/tools/converter"
	"hcm/pkg/tools/slice"

	"github.com/TencentBlueKing/gopkg/conv"
)

const (
	defaultExportFilename = "bill_adjustment_item-%s.csv"
)

// ExportBillAdjustmentItem 查询调账明细
func (b *billAdjustmentSvc) ExportBillAdjustmentItem(cts *rest.Contexts) (any, error) {

	req := new(bill.AdjustmentItemExportReq)
	if err := cts.DecodeInto(req); err != nil {
		return nil, err
	}
	if err := req.Validate(); err != nil {
		return nil, errf.NewFromErr(errf.InvalidParameter, err)
	}

	err := b.authorizer.AuthorizeWithPerm(cts.Kit,
		meta.ResourceAttribute{Basic: &meta.Basic{Type: meta.AccountBill, Action: meta.Find}})
	if err != nil {
		return nil, err
	}

	result, err := b.fetchBillAdjustmentItem(cts.Kit, req)
	if err != nil {
		logs.Errorf("fetch bill adjustment item failed, req: %v, err: %v, rid: %s", req, err, cts.Kit.Rid)
		return nil, err
	}

	bizIDMap := make(map[int64]struct{})
	mainAccountIDMap := make(map[string]struct{})
	for _, detail := range result {
		bizIDMap[detail.BkBizID] = struct{}{}
		mainAccountIDMap[detail.MainAccountID] = struct{}{}
	}
	bizIDs := converter.MapKeyToSlice(bizIDMap)
	mainAccountIDs := converter.MapKeyToSlice(mainAccountIDMap)

	mainAccountMap, err := b.listMainAccount(cts.Kit, mainAccountIDs)
	if err != nil {
		logs.Errorf("list main account failed, err: %v, rid: %s", err, cts.Kit.Rid)
		return nil, err
	}
	bizMap, err := b.listBiz(cts.Kit, bizIDs)
	if err != nil {
		logs.Errorf("list biz failed, bizIDs: %v, err: %v, rid: %s", bizIDs, err, cts.Kit.Rid)
		return nil, err
	}

	filename, filepath, writer, closeFunc, err := export.CreateWriterByFileName(cts.Kit, generateFilename())
	defer func() {
		if closeFunc != nil {
			closeFunc()
		}
	}()
	if err != nil {
		logs.Errorf("create writer failed: %v, rid: %s", err, cts.Kit.Rid)
		return nil, err
	}

	if err := writer.Write(export.BillAdjustmentTableHeader); err != nil {
		logs.Errorf("write header failed: %v, rid: %s", err, cts.Kit.Rid)
		return nil, err
	}

	table, err := toRawData(cts.Kit, result, mainAccountMap, bizMap)
	if err != nil {
		logs.Errorf("convert to raw data error: %v, rid: %s", err, cts.Kit.Rid)
		return nil, err
	}
	if err := writer.WriteAll(table); err != nil {
		logs.Errorf("write data failed: %v, rid: %s", err, cts.Kit.Rid)
		return nil, err
	}

	return &bill.FileDownloadResp{
		ContentTypeStr:        "application/octet-stream",
		ContentDispositionStr: fmt.Sprintf(`attachment; filename="%s"`, filename),
		FilePath:              filepath,
	}, nil
}

func generateFilename() string {
	return fmt.Sprintf(defaultExportFilename, time.Now().Format("2006-01-02-15_04_05"))
}

func (b *billAdjustmentSvc) fetchBillAdjustmentItem(kt *kit.Kit, req *bill.AdjustmentItemExportReq) (
	[]*billcore.AdjustmentItem, error) {

	var expression = tools.ExpressionAnd(
		tools.RuleEqual("bill_year", req.BillYear),
		tools.RuleEqual("bill_month", req.BillMonth),
	)
	if req.Filter != nil {
		var err error
		expression, err = tools.And(req.Filter, expression)
		if err != nil {
			logs.Errorf("build filter expression failed, error: %v, rid: %s", err, kt.Rid)
			return nil, err
		}
	}
	totalCount, err := b.fetchBillAdjustmentItemCount(kt, expression)
	if err != nil {
		logs.Errorf("fetch bill adjustment item count failed, err: %v, rid: %s", err, kt.Rid)
		return nil, err
	}
	exportLimit := min(totalCount, req.ExportLimit)

	result := make([]*billcore.AdjustmentItem, 0, exportLimit)
	for offset := uint64(0); offset < exportLimit; offset = offset + uint64(core.DefaultMaxPageLimit) {
		left := exportLimit - offset
		listReq := &dsbillapi.BillAdjustmentItemListReq{
			Filter: expression,
			Page: &core.BasePage{
				Start: uint32(offset),
				Limit: min(uint(left), core.DefaultMaxPageLimit),
			},
		}
		tmpResult, err := b.client.DataService().Global.Bill.ListBillAdjustmentItem(kt, listReq)
		if err != nil {
			logs.Errorf("list bill adjustment item failed, err: %v, rid: %s", err, kt.Rid)
			return nil, err
		}
		result = append(result, tmpResult.Details...)
	}

	return result, nil
}

func (b *billAdjustmentSvc) fetchBillAdjustmentItemCount(kt *kit.Kit,
	expression *filter.Expression) (uint64, error) {

	listReq := &dsbillapi.BillAdjustmentItemListReq{
		Filter: expression,
		Page:   core.NewCountPage(),
	}
	details, err := b.client.DataService().Global.Bill.ListBillAdjustmentItem(kt, listReq)
	if err != nil {
		return 0, err
	}
	return details.Count, nil
}

func toRawData(kt *kit.Kit, details []*billcore.AdjustmentItem, mainAccountMap map[string]*accountset.BaseMainAccount,
	bizMap map[int64]string) ([][]string, error) {

	data := make([][]string, 0, len(details))
	for _, detail := range details {
		bizName, ok := bizMap[detail.BkBizID]
		if !ok {
			logs.Warnf("biz(%d) not found", detail.BkBizID)
		}
		mainAccount, ok := mainAccountMap[detail.MainAccountID]
		if !ok {
			return nil, fmt.Errorf("main account(%s) not found", detail.MainAccountID)
		}

		table := export.BillAdjustmentTable{
			UpdateTime:      detail.UpdatedAt,
			BillID:          detail.ID,
			BKBizID:         conv.ToString(detail.BkBizID),
			BKBizName:       bizName,
			MainAccountName: mainAccount.Name,
			AdjustType:      enumor.BillAdjustmentTypeNameMap[detail.Type],
			Operator:        detail.Operator,
			Cost:            detail.Cost.String(),
			Currency:        string(detail.Currency),
			AdjustStatus:    enumor.BillAdjustmentStateNameMap[detail.State],
		}
		values, err := table.GetHeaderValues()
		if err != nil {
			logs.Errorf("get header fields failed, table: %v, error: %v, rid: %s", table, err, kt.Rid)
			return nil, err
		}
		data = append(data, values)
	}
	return data, nil
}

func (b *billAdjustmentSvc) listBiz(kt *kit.Kit, ids []int64) (map[int64]string, error) {
	ids = slice.Unique(ids)
	if len(ids) == 0 {
		return nil, nil
	}

	data := make(map[int64]string)
	for _, split := range slice.Split(ids, int(filter.DefaultMaxInLimit)) {
		rules := []cmdb.Rule{
			&cmdb.AtomRule{
				Field:    "bk_biz_id",
				Operator: cmdb.OperatorIn,
				Value:    split,
			},
		}
		expression := &cmdb.QueryFilter{Rule: &cmdb.CombinedRule{Condition: "AND", Rules: rules}}

		params := &cmdb.SearchBizParams{
			BizPropertyFilter: expression,
			Fields:            []string{"bk_biz_id", "bk_biz_name"},
		}
		resp, err := b.cmdbCli.SearchBusiness(kt, params)
		if err != nil {
			logs.Errorf("call cmdb search business api failed, err: %v, rid: %s", err, kt.Rid)
			return nil, fmt.Errorf("call cmdb search business api failed, err: %v", err)
		}

		infos := resp.Info
		for _, biz := range infos {
			data[biz.BizID] = biz.BizName
		}
	}
	return data, nil
}

func (b *billAdjustmentSvc) listMainAccount(kt *kit.Kit, ids []string) (map[string]*accountset.BaseMainAccount, error) {
	ids = slice.Unique(ids)
	if len(ids) == 0 {
		return nil, nil
	}

	result := make(map[string]*accountset.BaseMainAccount)
	for _, split := range slice.Split(ids, int(filter.DefaultMaxInLimit)) {
		listReq := &core.ListReq{
			Filter: tools.ExpressionAnd(tools.RuleIn("id", split)),
			Page:   core.NewDefaultBasePage(),
		}
		tmpResult, err := b.client.DataService().Global.MainAccount.List(kt, listReq)
		if err != nil {
			logs.Errorf("list main account failed, id: %v,err: %v, rid: %s", split, err, kt.Rid)
			return nil, err
		}
		for _, item := range tmpResult.Details {
			result[item.ID] = item
		}
	}

	return result, nil
}
