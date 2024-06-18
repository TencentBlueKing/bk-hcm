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

	"hcm/pkg/api/account-server/bill"
	"hcm/pkg/api/core"
	dataservice "hcm/pkg/api/data-service"
	dsbill "hcm/pkg/api/data-service/bill"
	"hcm/pkg/criteria/errf"
	"hcm/pkg/dal/dao/tools"
	"hcm/pkg/iam/meta"
	"hcm/pkg/kit"
	"hcm/pkg/logs"
	"hcm/pkg/rest"
	cvt "hcm/pkg/tools/converter"

	excelize "github.com/xuri/excelize/v2"
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

	// 校验一级账号和二级账号是否存在并匹配

	rootAccountInfo, err := b.client.DataService().Global.RootAccount.GetBasicInfo(cts.Kit, req.RootAccountID)
	if err != nil {
		return nil, err
	}

	filledItems, err := b.checkMainAccountAndFillRoot(cts.Kit, rootAccountInfo.ID, req.Items)
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

func (b *billAdjustmentSvc) checkMainAccountAndFillRoot(kt *kit.Kit, rootAccountID string,
	items []dsbill.BillAdjustmentItemCreateReq) ([]dsbill.BillAdjustmentItemCreateReq, error) {

	mainAccountIdMap := make(map[string]struct{})
	for i, item := range items {
		if item.RootAccountID == "" {
			items[i].RootAccountID = rootAccountID
		}
		if item.RootAccountID != rootAccountID {
			return nil, fmt.Errorf("root account id does not match  want: %s, given: %s",
				rootAccountID, item.RootAccountID)
		}
		item.Operator = kt.User
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
	return items, nil
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

	return b.client.DataService().Global.Bill.ListBillAdjustmentItem(cts.Kit, req)
}

// UpdateBillAdjustmentItem 更新调账明细
func (b *billAdjustmentSvc) UpdateBillAdjustmentItem(cts *rest.Contexts) (any, error) {
	id := cts.PathParameter("id").String()
	if len(id) == 0 {
		return nil, errf.New(errf.InvalidParameter, "id is required")
	}
	req := new(dsbill.BillAdjustmentItemUpdateReq)
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
	req.ID = id
	err = b.client.DataService().Global.Bill.UpdateBillAdjustmentItem(cts.Kit, req)
	if err != nil {
		logs.Errorf("fail to update bill adjustment item, err: %v, rid: %s", err, cts.Kit.Rid)
		return nil, err
	}
	return nil, nil
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

	err = ExcelRowsIterator(cts.Kit, cts.Request.Request.Body, 0, func(columns []string, err error) error {
		// 组装然后调用接口创建
		if err != nil {
			return err
		}
		// TODO
		return nil
	})
	if err != nil {
		logs.Errorf("fail parase excel file, err: %v, rid: %s", err, cts.Kit.Rid)
		return nil, err
	}

	// TODO
	return nil, nil
}

// ExcelRowsIterator traverse each row in Excel file by given operation
func ExcelRowsIterator(kt *kit.Kit, reader io.Reader, sheetIdx int, opFunc func([]string, error) error) error {

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
	// traverse all rows
	for rows.Next() {
		columns, err := rows.Columns()
		if err != nil {
			return opFunc(nil, err)
		}

		if err := opFunc(columns, nil); err != nil {
			return err
		}
	}
	return nil
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
