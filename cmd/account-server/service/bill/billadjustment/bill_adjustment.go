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
	"io"
	"strings"

	"hcm/pkg/api/account-server/bill"
	"hcm/pkg/api/core"
	accountset "hcm/pkg/api/core/account-set"
	dataservice "hcm/pkg/api/data-service"
	dsbill "hcm/pkg/api/data-service/bill"
	"hcm/pkg/criteria/constant"
	"hcm/pkg/criteria/enumor"
	"hcm/pkg/criteria/errf"
	"hcm/pkg/dal/dao/tools"
	"hcm/pkg/iam/meta"
	"hcm/pkg/kit"
	"hcm/pkg/logs"
	"hcm/pkg/rest"
	cvt "hcm/pkg/tools/converter"

	"github.com/xuri/excelize/v2"
)

// CreateBillAdjustmentItem 手动创建调账明细
func (b *billAdjustmentSvc) CreateBillAdjustmentItem(cts *rest.Contexts) (any, error) {

	req := new(bill.BatchBillAdjustmentItemCreateReq)
	if err := cts.DecodeInto(req); err != nil {
		return nil, errf.NewFromErr(errf.DecodeRequestFailed, err)
	}

	if err := req.Validate(); err != nil {
		return nil, errf.NewFromErr(errf.InvalidParameter, err)
	}

	err := b.authorizer.AuthorizeWithPerm(cts.Kit,
		meta.ResourceAttribute{Basic: &meta.Basic{Type: meta.AccountBill, Action: meta.Create}})
	if err != nil {
		return nil, err
	}

	// 1. 校验一级账号和二级账号是否存在并匹配
	rootAccountInfo, err := b.client.DataService().Global.RootAccount.GetBasicInfo(cts.Kit, req.RootAccountID)
	if err != nil {
		return nil, err
	}

	filledItems, err := b.convBillAdjustmentCreate(cts.Kit, rootAccountInfo.ID, req)
	if err != nil {
		logs.Errorf("fail to check main account: err: %v, rid: %s", err, cts.Kit.Rid)
		return nil, err
	}

	dataReq := &dsbill.BatchBillAdjustmentItemCreateReq{Items: filledItems}
	result, err := b.client.DataService().Global.Bill.BatchCreateBillAdjustmentItem(cts.Kit, dataReq)
	if err != nil {
		logs.Errorf("fail create bill adjustment item, err: %v, rid: %s", err, cts.Kit.Rid)
		return nil, err
	}
	return result, nil
}

func (b *billAdjustmentSvc) convBillAdjustmentCreate(kt *kit.Kit, rootAccountID string,
	req *bill.BatchBillAdjustmentItemCreateReq) ([]dsbill.BillAdjustmentItemCreateReq, error) {

	dsReq := make([]dsbill.BillAdjustmentItemCreateReq, 0, len(req.Items))
	mainAccountIdMap := make(map[string]struct{})
	for _, item := range req.Items {
		if item.RootAccountID == "" {
			item.RootAccountID = rootAccountID
		}
		if item.RootAccountID != rootAccountID {
			return nil, fmt.Errorf("root account id does not match, want: %s, given: %s",
				rootAccountID, item.RootAccountID)
		}
		dsReq = append(dsReq, dsbill.BillAdjustmentItemCreateReq{
			RootAccountID: item.RootAccountID,
			MainAccountID: item.MainAccountID,
			Vendor:        req.Vendor,
			ProductID:     item.ProductID,
			BkBizID:       item.BkBizID,
			BillYear:      item.BillYear,
			BillMonth:     item.BillMonth,
			BillDay:       1,
			State:         enumor.BillAdjustmentStateUnconfirmed,
			Type:          item.Type,
			Operator:      kt.User,
			Currency:      item.Currency,
			Cost:          item.Cost,
			RMBCost:       item.RmbCost,
			Memo:          item.Memo,
		})
		mainAccountIdMap[item.MainAccountID] = struct{}{}
	}
	mainAccountReq := &core.ListReq{
		Fields: []string{"id"},
		Filter: tools.ExpressionAnd(
			tools.RuleEqual("parent_account_id", rootAccountID),
			tools.RuleIn("id", cvt.MapKeyToStringSlice(mainAccountIdMap)),
		),
		Page: core.NewDefaultBasePage(),
	}
	mainAccountList, err := b.client.DataService().Global.MainAccount.List(kt, mainAccountReq)
	if err != nil {
		logs.Errorf("fail to list main account for create bill adjustment item, err: %v, rid: %s", err, kt.Rid)
		return nil, err
	}

	for _, item := range mainAccountList.Details {
		delete(mainAccountIdMap, item.ID)
	}
	if len(mainAccountIdMap) > 0 {
		return nil, fmt.Errorf("main account id not found: %v", cvt.MapKeyToSlice(mainAccountIdMap))
	}
	return dsReq, nil
}

// ListBillAdjustmentItem 查询调账明细
func (b *billAdjustmentSvc) ListBillAdjustmentItem(cts *rest.Contexts) (any, error) {

	req := new(bill.ListBillAdjustmentReq)
	if err := cts.DecodeInto(req); err != nil {
		return nil, errf.NewFromErr(errf.DecodeRequestFailed, err)
	}

	if err := req.Validate(); err != nil {
		return nil, errf.NewFromErr(errf.InvalidParameter, err)
	}

	err := b.authorizer.AuthorizeWithPerm(cts.Kit,
		meta.ResourceAttribute{Basic: &meta.Basic{Type: meta.AccountBill, Action: meta.Find}})
	if err != nil {
		return nil, err
	}

	dsItems, err := b.client.DataService().Global.Bill.ListBillAdjustmentItem(cts.Kit, req)
	if err != nil {
		logs.Errorf("fail to call data service to list bill adjustment item, err: %v, rid: %s", err, cts.Kit.Rid)
		return nil, err
	}
	resp := core.ListResultT[bill.AdjustmentItemResult]{
		Count:   dsItems.Count,
		Details: make([]bill.AdjustmentItemResult, len(dsItems.Details)),
	}
	if len(dsItems.Details) == 0 {
		return dsItems, nil
	}
	mainAccIDCloudIDMap := make(map[string]*accountset.BaseMainAccount, len(dsItems.Details))
	// collect main account id for list cloud id
	for i, adjustmentItem := range dsItems.Details {
		resp.Details[i].AdjustmentItem = adjustmentItem
		mainAccIDCloudIDMap[adjustmentItem.MainAccountID] = nil
	}
	// list for cloud id
	mainAccReq := &core.ListReq{
		Filter: tools.ContainersExpression("id", cvt.MapKeyToSlice(mainAccIDCloudIDMap)),
		Page:   req.Page,
		Fields: []string{"id", "email", "cloud_id"},
	}
	mainAccountListResult, err := b.client.DataService().Global.MainAccount.List(cts.Kit, mainAccReq)
	if err != nil {
		logs.Errorf("fail to query main account for bill adjustment item, err: %v, rid: %s", err, cts.Kit.Rid)
		return nil, err
	}
	for _, mainAccount := range mainAccountListResult.Details {
		mainAccIDCloudIDMap[mainAccount.ID] = mainAccount
	}
	// 填充主账号云id 和 email
	for i, adjustmentItem := range resp.Details {
		mainAccount := mainAccIDCloudIDMap[adjustmentItem.MainAccountID]
		if mainAccount == nil {
			// Skip not found account info
			logs.Warnf("main account of bill adjustment not found, main account id: %s, adjustment id: %s, rid: %s",
				adjustmentItem.MainAccountID, adjustmentItem.ID, cts.Kit.Rid)
			continue
		}
		resp.Details[i].MainAccountCloudID = mainAccount.CloudID
		resp.Details[i].MainAccountEmail = mainAccount.Email
	}
	return resp, nil
}

// UpdateBillAdjustmentItem 更新调账明细
func (b *billAdjustmentSvc) UpdateBillAdjustmentItem(cts *rest.Contexts) (any, error) {
	id := cts.PathParameter("id").String()
	if len(id) == 0 {
		return nil, errf.New(errf.InvalidParameter, "id is required")
	}
	req := new(bill.BillAdjustmentItemUpdateReq)
	if err := cts.DecodeInto(req); err != nil {
		return nil, errf.NewFromErr(errf.DecodeRequestFailed, err)
	}

	if err := req.Validate(); err != nil {
		return nil, errf.NewFromErr(errf.InvalidParameter, err)
	}

	err := b.authorizer.AuthorizeWithPerm(cts.Kit,
		meta.ResourceAttribute{Basic: &meta.Basic{Type: meta.AccountBill, Action: meta.Update}})
	if err != nil {
		return nil, err
	}

	if err := b.checkAdjustmentUnconfirmed(cts, []string{id}); err != nil {
		return nil, err
	}

	dsReq := &dsbill.BillAdjustmentItemUpdateReq{
		ID:            id,
		MainAccountID: req.MainAccountID,
		ProductID:     req.ProductID,
		BkBizID:       req.BkBizID,
		Type:          req.Type,
		Memo:          req.Memo,
		Currency:      req.Currency,
		Cost:          req.Cost,
	}

	err = b.client.DataService().Global.Bill.UpdateBillAdjustmentItem(cts.Kit, dsReq)
	if err != nil {
		logs.Errorf("fail to update bill adjustment item, err: %v, rid: %s", err, cts.Kit.Rid)
		return nil, err
	}
	return nil, nil
}

// BatchConfirmBillAdjustmentItem 批量确认调账明细
func (b *billAdjustmentSvc) BatchConfirmBillAdjustmentItem(cts *rest.Contexts) (any, error) {
	req := new(core.BatchDeleteReq)
	if err := cts.DecodeInto(req); err != nil {
		return nil, errf.NewFromErr(errf.DecodeRequestFailed, err)
	}

	if err := req.Validate(); err != nil {
		return nil, errf.NewFromErr(errf.InvalidParameter, err)
	}

	err := b.authorizer.AuthorizeWithPerm(cts.Kit,
		meta.ResourceAttribute{Basic: &meta.Basic{Type: meta.AccountBill, Action: meta.Update}})
	if err != nil {
		return nil, err
	}

	if err := b.checkAdjustmentUnconfirmed(cts, req.IDs); err != nil {
		return nil, err
	}

	err = b.client.DataService().Global.Bill.BatchConfirmBillAdjustmentItem(cts.Kit, req)
	if err != nil {
		logs.Errorf("fail to update bill adjustment item, err: %v, rid: %s", err, cts.Kit.Rid)
		return nil, err
	}
	return nil, nil
}

// DeleteBillAdjustmentItem ...
func (b *billAdjustmentSvc) DeleteBillAdjustmentItem(cts *rest.Contexts) (any, error) {
	id := cts.PathParameter("id").String()
	if len(id) == 0 {
		return nil, errf.New(errf.InvalidParameter, "id is required")
	}

	err := b.authorizer.AuthorizeWithPerm(cts.Kit,
		meta.ResourceAttribute{Basic: &meta.Basic{Type: meta.AccountBill, Action: meta.Delete}})
	if err != nil {
		return nil, err
	}

	if err := b.checkAdjustmentUnconfirmed(cts, []string{id}); err != nil {
		return nil, err
	}

	delReq := &dataservice.BatchDeleteReq{
		Filter: tools.EqualExpression("id", id),
	}
	err = b.client.DataService().Global.Bill.BatchDeleteBillAdjustmentItem(cts.Kit, delReq)
	if err != nil {
		logs.Errorf("fail to delete bill adjustment item by id %s, err: %v, rid: %s", id, err, cts.Kit.Rid)
		return nil, err
	}
	return nil, nil
}

// BatchDeleteBillAdjustmentItem ...
func (b *billAdjustmentSvc) BatchDeleteBillAdjustmentItem(cts *rest.Contexts) (any, error) {

	req := new(bill.BatchDeleteReq)
	if err := cts.DecodeInto(req); err != nil {
		return nil, errf.NewFromErr(errf.DecodeRequestFailed, err)
	}

	if err := req.Validate(); err != nil {
		return nil, errf.NewFromErr(errf.InvalidParameter, err)
	}

	err := b.authorizer.AuthorizeWithPerm(cts.Kit,
		meta.ResourceAttribute{Basic: &meta.Basic{Type: meta.AccountBill, Action: meta.Delete}})
	if err != nil {
		return nil, err
	}

	if err := b.checkAdjustmentUnconfirmed(cts, req.Ids); err != nil {
		return nil, err
	}

	delReq := &dataservice.BatchDeleteReq{
		Filter: tools.ContainersExpression("id", req.Ids),
	}
	err = b.client.DataService().Global.Bill.BatchDeleteBillAdjustmentItem(cts.Kit, delReq)
	if err != nil {
		logs.Errorf("fail to batch delete bill adjustment item, err: %v, ids: %v rid: %s", err, req.Ids, cts.Kit.Rid)
		return nil, err
	}
	return nil, nil
}

// 检查给定的调整明细是否都是未确认调账条目，如果存在已确定条目会返回错误
func (b *billAdjustmentSvc) checkAdjustmentUnconfirmed(cts *rest.Contexts, ids []string) error {
	// 检查是否已确认调账明细
	listReq := &core.ListReq{
		Filter: tools.ContainersExpression("id", ids),
		Page:   core.NewDefaultBasePage(),
		Fields: []string{"id", "state"},
	}
	itemResp, err := b.client.DataService().Global.Bill.ListBillAdjustmentItem(cts.Kit, listReq)
	if err != nil {
		logs.Errorf("fail to query bill adjustment for check unconfirmed, err: %v, ids: %v, rid: %s",
			err, ids, cts.Kit.Rid)
		return err
	}

	if len(itemResp.Details) != len(ids) {
		return errf.New(errf.RecordNotFound, "item not found")
	}
	confirmed := make([]string, 0)
	for _, detail := range itemResp.Details {
		if detail.State == enumor.BillAdjustmentStateConfirmed {
			confirmed = append(confirmed, detail.ID)
		}
	}
	if len(confirmed) > 0 {
		return errf.New(errf.InvalidParameter, "confirmed items can not be modified, ids: "+strings.Join(confirmed,
			","))
	}
	return nil
}

// ImportBillAdjustment 导入账单明细
func (b *billAdjustmentSvc) ImportBillAdjustment(cts *rest.Contexts) (any, error) {

	err := b.authorizer.AuthorizeWithPerm(cts.Kit,
		meta.ResourceAttribute{Basic: &meta.Basic{Type: meta.AccountBill, Action: meta.Create}})
	if err != nil {
		return nil, err
	}

	req := new(bill.ImportBillAdjustmentReq)
	if err := cts.DecodeInto(req); err != nil {
		return nil, errf.NewFromErr(errf.DecodeRequestFailed, err)
	}

	if err := req.Validate(); err != nil {
		return nil, errf.NewFromErr(errf.InvalidParameter, err)
	}

	err = excelRowsIterator(cts.Kit, cts.Request.Request.Body, 0, constant.BatchOperationMaxLimit,
		func(rows [][]string, err error) error {
			// 组装然后调用接口创建
			if err != nil {
				logs.Errorf("fail to read excel, err: %v, rid: %s", err, cts.Kit.Rid)
				return err
			}
			if len(rows) == 0 {
				return nil
			}
			// TODO 确认Excel 格式

			return nil
		})
	if err != nil {
		logs.Errorf("fail parase excel file, err: %v, rid: %s", err, cts.Kit.Rid)
		return nil, err
	}

	return nil, nil
}

// excelRowsIterator traverse each row in Excel file by given operation
func excelRowsIterator(kt *kit.Kit, reader io.Reader, sheetIdx, batchSize int,
	opFunc func([][]string, error) error) error {

	excel, err := excelize.OpenReader(reader)
	if err != nil {
		logs.Errorf("fialed to create excel reader, err: %v, rid: %s", err, kt.Rid)
		return err
	}
	defer excel.Close()

	sheetName := excel.GetSheetName(sheetIdx)

	rows, err := excel.Rows(sheetName)
	if err != nil {
		logs.Errorf("fail to read rows from sheet(%s), err: %v, rid: %s", sheetName, err, kt.Rid)
		return err
	}
	defer rows.Close()

	rowBatch := make([][]string, 0, batchSize)
	// traverse all rows
	for rows.Next() {
		columns, err := rows.Columns()
		if err != nil {
			return opFunc(nil, err)
		}
		rowBatch = append(rowBatch, columns)
		if len(rowBatch) < batchSize {
			continue
		}
		if err := opFunc(rowBatch, nil); err != nil {
			return err
		}
		rowBatch = rowBatch[:0]

	}
	return nil
}
