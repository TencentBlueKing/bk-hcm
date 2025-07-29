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

package billsummarybiz

import (
	"fmt"
	"time"

	"hcm/cmd/account-server/logics/bill/export"
	"hcm/pkg/api/account-server/bill"
	"hcm/pkg/api/core"
	billproto "hcm/pkg/api/data-service/bill"
	"hcm/pkg/criteria/errf"
	"hcm/pkg/dal/dao/tools"
	"hcm/pkg/iam/meta"
	"hcm/pkg/kit"
	"hcm/pkg/logs"
	"hcm/pkg/rest"
	"hcm/pkg/runtime/filter"
	"hcm/pkg/thirdparty/api-gateway/cmdb"
	"hcm/pkg/tools/slice"

	"github.com/TencentBlueKing/gopkg/conv"
)

const (
	defaultExportFilename = "bill_summary_biz-%s.csv"
)

// ExportBizSummary export biz summary with options
func (s *service) ExportBizSummary(cts *rest.Contexts) (interface{}, error) {
	req := new(bill.BizSummaryExportReq)
	if err := cts.DecodeInto(req); err != nil {
		return nil, err
	}
	if err := req.Validate(); err != nil {
		return nil, errf.NewFromErr(errf.InvalidParameter, err)
	}

	err := s.authorizer.AuthorizeWithPerm(cts.Kit,
		meta.ResourceAttribute{Basic: &meta.Basic{Type: meta.AccountBill, Action: meta.Find}})
	if err != nil {
		return nil, err
	}

	result, err := s.fetchBizSummary(cts, req)
	if err != nil {
		logs.Errorf("fetch biz summary failed, err: %v, rid: %s", err, cts.Kit.Rid)
		return nil, err
	}

	bkBizIDs := make([]int64, 0, len(result))
	for _, detail := range result {
		bkBizIDs = append(bkBizIDs, detail.BkBizID)
	}
	bizMap, err := s.listBiz(cts.Kit, bkBizIDs)
	if err != nil {
		logs.Errorf("list biz failed, err: %v, rid: %s", err, cts.Kit.Rid)
		return nil, err
	}

	filename, filepath, writer, closeFunc, err := export.CreateWriterByFileName(cts.Kit, generateFilename())
	defer func() {
		if closeFunc != nil {
			closeFunc()
		}
	}()
	if err != nil {
		logs.Errorf("create writer failed, err: %v, rid: %s", err, cts.Kit.Rid)
		return nil, err
	}

	for _, header := range export.BillSummaryBizTableHeader {
		if err = writer.Write(header); err != nil {
			logs.Errorf("write header failed, err: %v, header: %v, rid: %s", err, header, cts.Kit.Rid)
			return nil, err
		}
	}

	table, err := toRawData(cts.Kit, result, bizMap)
	if err != nil {
		logs.Errorf("convert to raw data failed, err: %v, rid: %s", err, cts.Kit.Rid)
		return nil, err
	}
	if err := writer.WriteAll(table); err != nil {
		logs.Errorf("write data failed, err: %v, rid: %s", err, cts.Kit.Rid)
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

func toRawData(kt *kit.Kit, details []*billproto.BillSummaryBizResult, bizMap map[int64]string) ([][]string, error) {
	result := make([][]string, 0, len(details))
	for _, detail := range details {
		table := export.BillSummaryBizTable{
			BkBizID:                   conv.ToString(detail.BkBizID),
			BkBizName:                 bizMap[detail.BkBizID],
			CurrentMonthRMBCostSynced: detail.CurrentMonthRMBCostSynced.String(),
			CurrentMonthCostSynced:    detail.CurrentMonthCostSynced.String(),
			CurrentMonthRMBCost:       detail.CurrentMonthRMBCost.String(),
			CurrentMonthCost:          detail.CurrentMonthCost.String(),
		}
		fields, err := table.GetValuesByHeader()
		if err != nil {
			logs.Errorf("get header fields failed, err: %v, rid: %s", err, kt.Rid)
			return nil, err
		}
		result = append(result, fields)
	}
	return result, nil
}

func (s *service) fetchBizSummary(cts *rest.Contexts, req *bill.BizSummaryExportReq) (
	[]*billproto.BillSummaryBizResult, error) {

	if len(req.BKBizIDs) == 0 {
		return s.fetchAllBizSummary(cts, req)
	}
	result := make([]*billproto.BillSummaryBizResult, 0)
	for _, bkBizIDs := range slice.Split(req.BKBizIDs, int(filter.DefaultMaxInLimit)) {
		expression := tools.ExpressionAnd(
			tools.RuleEqual("bill_year", req.BillYear),
			tools.RuleEqual("bill_month", req.BillMonth),
			tools.RuleIn("bk_biz_id", bkBizIDs),
		)
		listReq := &core.ListReq{
			Filter: expression,
			Page:   core.NewDefaultBasePage(),
		}
		tmpResult, err := s.client.DataService().Global.Bill.ListBillSummaryBiz(cts.Kit, listReq)
		if err != nil {
			return nil, err
		}
		result = append(result, tmpResult.Details...)
	}
	if len(result) > int(req.ExportLimit) {
		return result[:req.ExportLimit], nil
	}
	return result, nil
}

func (s *service) fetchAllBizSummary(cts *rest.Contexts, req *bill.BizSummaryExportReq) (
	[]*billproto.BillSummaryBizResult, error) {

	var expression = tools.ExpressionAnd(
		tools.RuleEqual("bill_year", req.BillYear),
		tools.RuleEqual("bill_month", req.BillMonth),
	)
	listReq := &core.ListReq{
		Filter: expression,
		Page:   core.NewCountPage(),
	}
	details, err := s.client.DataService().Global.Bill.ListBillSummaryBiz(cts.Kit, listReq)
	if err != nil {
		return nil, err
	}

	exportLimit := min(details.Count, req.ExportLimit)

	result := make([]*billproto.BillSummaryBizResult, 0, exportLimit)
	for offset := uint64(0); offset < exportLimit; offset = offset + uint64(core.DefaultMaxPageLimit) {
		left := exportLimit - offset
		listReq := &core.ListReq{
			Filter: expression,
			Page: &core.BasePage{
				Start: uint32(offset),
				Limit: min(uint(left), core.DefaultMaxPageLimit),
			},
		}
		tmpResult, err := s.client.DataService().Global.Bill.ListBillSummaryBiz(cts.Kit, listReq)
		if err != nil {
			return nil, err
		}
		result = append(result, tmpResult.Details...)
	}
	return result, nil
}

func (s *service) listBiz(kt *kit.Kit, ids []int64) (map[int64]string, error) {
	ids = slice.Unique(ids)
	if len(ids) == 0 {
		return nil, nil
	}
	rules := []cmdb.Rule{
		&cmdb.AtomRule{
			Field:    "bk_biz_id",
			Operator: cmdb.OperatorIn,
			Value:    ids,
		},
	}
	expression := &cmdb.QueryFilter{Rule: &cmdb.CombinedRule{Condition: "AND", Rules: rules}}

	params := &cmdb.SearchBizParams{
		BizPropertyFilter: expression,
		Fields:            []string{"bk_biz_id", "bk_biz_name"},
	}
	resp, err := s.cmdbCli.SearchBusiness(kt, params)
	if err != nil {
		logs.Errorf("call cmdb search business api failed, err: %v, rid: %s", err, kt.Rid)
		return nil, fmt.Errorf("call cmdb search business api failed, err: %v", err)
	}

	infos := resp.Info
	data := make(map[int64]string, len(infos))
	for _, biz := range infos {
		data[biz.BizID] = biz.BizName
	}

	return data, nil
}
