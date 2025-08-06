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

package billsummaryroot

import (
	"fmt"
	"time"

	"hcm/cmd/account-server/logics/bill/export"
	asbillapi "hcm/pkg/api/account-server/bill"
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
	"hcm/pkg/tools/converter"

	"github.com/TencentBlueKing/gopkg/conv"
)

const (
	defaultExportFilename = "bill_summary_root-%s.csv"
)

// ExportRootAccountSummary export root account summary with options
func (s *service) ExportRootAccountSummary(cts *rest.Contexts) (interface{}, error) {
	req := new(asbillapi.RootAccountSummaryExportReq)
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

	result, err := s.fetchRootAccountSummary(cts, req)
	if err != nil {
		logs.Errorf("fetch root account summary failed, err: %v, rid: %s", err, cts.Kit.Rid)
		return nil, err
	}

	rootAccountIDMap := make(map[string]struct{})
	for _, detail := range result {
		rootAccountIDMap[detail.RootAccountID] = struct{}{}
	}
	rootAccountIDs := converter.MapKeyToSlice(rootAccountIDMap)
	rootAccountMap, err := s.listRootAccount(cts.Kit, rootAccountIDs)
	if err != nil {
		logs.Errorf("list root account error: %v, rid: %s", err, cts.Kit.Rid)
		return nil, err
	}

	filename, filepath, writer, closeFunc, err := export.CreateWriterByFileName(cts.Kit, generateFileName())
	defer func() {
		if closeFunc != nil {
			closeFunc()
		}
	}()
	if err != nil {
		logs.Errorf("create writer failed: %v, rid: %s", err, cts.Kit.Rid)
		return nil, err
	}

	for _, header := range export.BillSummaryRootTableHeaders {
		if err = writer.Write(header); err != nil {
			logs.Errorf("write header failed: %v, val: %v, rid: %s", err, header, cts.Kit.Rid)
			return nil, err
		}
	}

	table, err := toRawData(cts.Kit, result, rootAccountMap)
	if err != nil {
		logs.Errorf("convert to raw data failed, err: %v, rid: %s", err, cts.Kit.Rid)
		return nil, err
	}
	if err := writer.WriteAll(table); err != nil {
		logs.Errorf("write data failed: %v, rid: %s", err, cts.Kit.Rid)
		return nil, err
	}
	return &asbillapi.FileDownloadResp{
		ContentTypeStr:        "application/octet-stream",
		ContentDispositionStr: fmt.Sprintf(`attachment; filename="%s"`, filename),
		FilePath:              filepath,
	}, nil
}

func generateFileName() string {
	return fmt.Sprintf(defaultExportFilename, time.Now().Format("2006-01-02-15_04_05"))
}

func (s *service) fetchRootAccountSummary(cts *rest.Contexts, req *asbillapi.RootAccountSummaryExportReq) (
	[]*billcore.SummaryRoot, error) {

	var expression = tools.ExpressionAnd(
		tools.RuleEqual("bill_year", req.BillYear),
		tools.RuleEqual("bill_month", req.BillMonth),
	)
	if req.Filter != nil {
		var err error
		expression, err = tools.And(req.Filter, expression)
		if err != nil {
			logs.Errorf("build filter expression failed, error: %v, rid: %s", err, cts.Kit.Rid)
			return nil, err
		}
	}

	countReq := &dsbillapi.BillSummaryRootListReq{
		Filter: expression,
		Page:   core.NewCountPage(),
	}
	details, err := s.client.DataService().Global.Bill.ListBillSummaryRoot(cts.Kit, countReq)
	if err != nil {
		logs.Errorf("list bill summary root failed, err: %v, rid: %s", err, cts.Kit.Rid)
		return nil, err
	}

	exportLimit := min(*details.Count, req.ExportLimit)
	result := make([]*billcore.SummaryRoot, 0, exportLimit)
	for offset := uint64(0); offset < exportLimit; offset = offset + uint64(core.DefaultMaxPageLimit) {
		left := exportLimit - offset
		listReq := &dsbillapi.BillSummaryRootListReq{
			Filter: expression,
			Page: &core.BasePage{
				Start: uint32(offset),
				Limit: min(uint(left), core.DefaultMaxPageLimit),
			},
		}
		tmpResult, err := s.client.DataService().Global.Bill.ListBillSummaryRoot(cts.Kit, listReq)
		if err != nil {
			logs.Errorf("list bill summary root failed, err: %v, rid: %s", err, cts.Kit.Rid)
			return nil, err
		}
		result = append(result, tmpResult.Details...)
	}
	return result, nil
}

func toRawData(kt *kit.Kit, details []*billcore.SummaryRoot,
	accountMap map[string]*accountset.BaseRootAccount) ([][]string, error) {
	data := make([][]string, 0, len(details))
	for _, detail := range details {
		rootAccount, ok := accountMap[detail.RootAccountID]
		if !ok {
			return nil, fmt.Errorf("root account not found, id: %s", detail.RootAccountID)
		}
		table := export.BillSummaryRootTable{
			RootAccountID:             rootAccount.CloudID,
			RootAccountName:           rootAccount.Name,
			State:                     enumor.RootAccountBillSummaryStateMap[detail.State],
			CurrentMonthRMBCostSynced: detail.CurrentMonthRMBCostSynced.String(),
			LastMonthRMBCostSynced:    detail.LastMonthRMBCostSynced.String(),
			CurrentMonthCostSynced:    detail.CurrentMonthCostSynced.String(),
			LastMonthCostSynced:       detail.LastMonthCostSynced.String(),
			MonthOnMonthValue:         conv.ToString(detail.MonthOnMonthValue),
			CurrentMonthRMB:           detail.CurrentMonthRMBCost.String(),
			CurrentMonthCost:          detail.CurrentMonthCost.String(),
			AdjustRMBCost:             detail.AdjustmentRMBCost.String(),
			AdjustCost:                detail.AdjustmentCost.String(),
		}
		fields, err := table.GetValuesByHeader()
		if err != nil {
			logs.Errorf("get header fields failed: %v, rid: %s", err, kt.Rid)
			return nil, err
		}
		data = append(data, fields)
	}
	return data, nil
}
