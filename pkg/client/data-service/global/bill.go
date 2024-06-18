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

package global

import (
	"context"
	rawjson "encoding/json"
	"fmt"
	"net/http"

	"hcm/pkg/api/core"
	"hcm/pkg/api/core/bill"
	dataservice "hcm/pkg/api/data-service"
	billproto "hcm/pkg/api/data-service/bill"
	datacloudbillproto "hcm/pkg/api/data-service/cloud/bill"
	"hcm/pkg/client/common"
	"hcm/pkg/criteria/enumor"
	"hcm/pkg/criteria/errf"
	"hcm/pkg/kit"
	"hcm/pkg/rest"
)

// BillClient is data service bill api client.
type BillClient struct {
	client rest.ClientInterface
}

// NewBillClient create a new bill api client.
func NewBillClient(client rest.ClientInterface) *BillClient {
	return &BillClient{
		client: client,
	}
}

// List list bill.
func (b *BillClient) List(ctx context.Context, h http.Header, req *core.ListReq) (
	*datacloudbillproto.AccountBillConfigListResult, error) {

	resp := new(datacloudbillproto.AccountBillConfigListResp)

	err := b.client.Post().
		WithContext(ctx).
		Body(req).
		SubResourcef("/bills/config/list").
		WithHeaders(h).
		Do().
		Into(resp)
	if err != nil {
		return nil, err
	}

	if resp.Code != errf.OK {
		return nil, errf.New(resp.Code, resp.Message)
	}

	return resp.Data, nil
}

// BatchDelete batch delete bill.
func (b *BillClient) BatchDelete(ctx context.Context, h http.Header, req *dataservice.BatchDeleteReq) error {
	resp := new(rest.BaseResp)

	err := b.client.Delete().
		WithContext(ctx).
		Body(req).
		SubResourcef("/bills/config/batch").
		WithHeaders(h).
		Do().
		Into(resp)
	if err != nil {
		return err
	}

	if resp.Code != errf.OK {
		return errf.New(resp.Code, resp.Message)
	}

	return nil
}

// --- bill adjustment item ---

// BatchCreateBillAdjustmentItem create bill adjustment item
func (b *BillClient) BatchCreateBillAdjustmentItem(kt *kit.Kit, req *billproto.BatchBillAdjustmentItemCreateReq) (
	*core.BatchCreateResult, error) {
	return common.Request[billproto.BatchBillAdjustmentItemCreateReq, core.BatchCreateResult](
		b.client, rest.POST, kt, req, "/bills/adjustmentitems")
}

// BatchDeleteBillAdjustmentItem delete bill adjustment item
func (b *BillClient) BatchDeleteBillAdjustmentItem(kt *kit.Kit, req *dataservice.BatchDeleteReq) error {
	return common.RequestNoResp[dataservice.BatchDeleteReq](
		b.client, rest.DELETE, kt, req, "/bills/adjustmentitems")
}

// UpdateBillAdjustmentItem update bill adjustment item
func (b *BillClient) UpdateBillAdjustmentItem(kt *kit.Kit, req *billproto.BillAdjustmentItemUpdateReq) error {
	return common.RequestNoResp[billproto.BillAdjustmentItemUpdateReq](
		b.client, rest.PUT, kt, req, "/bills/adjustmentitems")
}

// ListBillAdjustmentItem list bill adjustment item
func (b *BillClient) ListBillAdjustmentItem(kt *kit.Kit, req *billproto.BillAdjustmentItemListReq) (
	*billproto.BillAdjustmentItemListResult, error) {
	return common.Request[billproto.BillAdjustmentItemListReq, billproto.BillAdjustmentItemListResult](
		b.client, rest.GET, kt, req, "/bills/adjustmentitems")
}

// --- bill item ---

// BatchDeleteBillItem delete bill item
func (b *BillClient) BatchDeleteBillItem(kt *kit.Kit, req *dataservice.BatchDeleteReq) error {
	return common.RequestNoResp[dataservice.BatchDeleteReq](
		b.client, rest.DELETE, kt, req, "/bills/items")
}

// UpdateBillItem update bill item
func (b *BillClient) UpdateBillItem(kt *kit.Kit, req *billproto.BillItemUpdateReq) error {
	return common.RequestNoResp[billproto.BillItemUpdateReq](
		b.client, rest.PUT, kt, req, "/bills/items")
}

// ListBillItem list bill item
func (b *BillClient) ListBillItem(kt *kit.Kit, req *core.ListReq) (*billproto.BillItemBaseListResult, error) {

	return common.Request[core.ListReq, billproto.BillItemBaseListResult](
		b.client, rest.POST, kt, req, "/bills/items/list")
}

// ListBillItemRaw list with extension
func (b *BillClient) ListBillItemRaw(kt *kit.Kit, req *core.ListReq) (*core.ListResultT[*bill.BillItemRaw], error) {

	return common.Request[core.ListReq, core.ListResultT[*bill.BillItemRaw]](
		b.client, rest.POST, kt, req, "/bills/items/list_with_extension")
}

// --- bill daily pull task ---

// CreateBillDailyPullTask create bill daily pull task
func (b *BillClient) CreateBillDailyPullTask(kt *kit.Kit, req *billproto.BillDailyPullTaskCreateReq) (
	*core.CreateResult, error) {

	return common.Request[billproto.BillDailyPullTaskCreateReq, core.CreateResult](
		b.client, rest.POST, kt, req, "/bills/dailypulltasks")
}

// BatchDeleteBillDailyPullTask delete bill daily pull task
func (b *BillClient) BatchDeleteBillDailyPullTask(kt *kit.Kit, req *dataservice.BatchDeleteReq) error {
	return common.RequestNoResp[dataservice.BatchDeleteReq](
		b.client, rest.DELETE, kt, req, "/bills/dailypulltasks")
}

// UpdateBillDailyPullTask update bill daily pull task
func (b *BillClient) UpdateBillDailyPullTask(kt *kit.Kit, req *billproto.BillDailyPullTaskUpdateReq) error {
	return common.RequestNoResp[billproto.BillDailyPullTaskUpdateReq](
		b.client, rest.PUT, kt, req, "/bills/dailypulltasks")
}

// ListBillDailyPullTask list bill daily pull task
func (b *BillClient) ListBillDailyPullTask(kt *kit.Kit, req *billproto.BillDailyPullTaskListReq) (
	*billproto.BillDailyPullTaskListResult, error) {
	return common.Request[billproto.BillDailyPullTaskListReq, billproto.BillDailyPullTaskListResult](
		b.client, rest.GET, kt, req, "/bills/dailypulltasks")
}

// --- bill puller ---

// CreateBillPuller create bill puller
func (b *BillClient) CreateBillPuller(kt *kit.Kit, req *billproto.BillPullerCreateReq) (
	*core.CreateResult, error) {
	return common.Request[billproto.BillPullerCreateReq, core.CreateResult](
		b.client, rest.POST, kt, req, "/bills/pullers")
}

// BatchDeleteBillPuller delete bill puller
func (b *BillClient) BatchDeleteBillPuller(kt *kit.Kit, req *dataservice.BatchDeleteReq) error {
	return common.RequestNoResp[dataservice.BatchDeleteReq](
		b.client, rest.DELETE, kt, req, "/bills/pullers")
}

// UpdateBillPuller update bill puller
func (b *BillClient) UpdateBillPuller(kt *kit.Kit, req *billproto.BillPullerUpdateReq) error {
	return common.RequestNoResp[billproto.BillPullerUpdateReq](
		b.client, rest.PUT, kt, req, "/bills/pullers")
}

// ListBillPuller list bill puller
func (b *BillClient) ListBillPuller(kt *kit.Kit, req *billproto.BillPullerListReq) (
	*billproto.BillPullerListResult, error) {
	return common.Request[billproto.BillPullerListReq, billproto.BillPullerListResult](
		b.client, rest.GET, kt, req, "/bills/pullers")
}

// --- bill summary ---

// CreateBillSummaryMain create bill summary
func (b *BillClient) CreateBillSummaryMain(kt *kit.Kit, req *billproto.BillSummaryMainCreateReq) (
	*core.CreateResult, error) {
	return common.Request[billproto.BillSummaryMainCreateReq, core.CreateResult](
		b.client, rest.POST, kt, req, "/bills/summarymains")
}

// BatchDeleteBillSummaryMain delete bill summary
func (b *BillClient) BatchDeleteBillSummaryMain(kt *kit.Kit, req *dataservice.BatchDeleteReq) error {
	return common.RequestNoResp[dataservice.BatchDeleteReq](
		b.client, rest.DELETE, kt, req, "/bills/summarymains")
}

// UpdateBillSummaryMain update bill summary
func (b *BillClient) UpdateBillSummaryMain(kt *kit.Kit, req *billproto.BillSummaryMainUpdateReq) error {
	return common.RequestNoResp[billproto.BillSummaryMainUpdateReq](
		b.client, rest.PUT, kt, req, "/bills/summarymains")
}

// ListBillSummaryMain list bill summary
func (b *BillClient) ListBillSummaryMain(kt *kit.Kit, req *billproto.BillSummaryMainListReq) (
	*billproto.BillSummaryMainListResult, error) {
	return common.Request[billproto.BillSummaryMainListReq, billproto.BillSummaryMainListResult](
		b.client, rest.GET, kt, req, "/bills/summarymains")
}

// --- bill summary daily ---

// CreateBillSummaryDaily create bill summary daily
func (b *BillClient) CreateBillSummaryDaily(kt *kit.Kit, req *billproto.BillSummaryDailyCreateReq) (
	*core.CreateResult, error) {
	return common.Request[billproto.BillSummaryDailyCreateReq, core.CreateResult](
		b.client, rest.POST, kt, req, "/bills/summarydailys")
}

// BatchDeleteBillSummaryDaily delete bill summary daily
func (b *BillClient) BatchDeleteBillSummaryDaily(kt *kit.Kit, req *dataservice.BatchDeleteReq) error {
	return common.RequestNoResp[dataservice.BatchDeleteReq](
		b.client, rest.DELETE, kt, req, "/bills/summarydailys")
}

// UpdateBillSummaryDaily update bill summary daily
func (b *BillClient) UpdateBillSummaryDaily(kt *kit.Kit, req *billproto.BillSummaryDailyUpdateReq) error {
	return common.RequestNoResp[billproto.BillSummaryDailyUpdateReq](
		b.client, rest.PUT, kt, req, "/bills/summarydailys")
}

// ListBillSummaryDaily list bill summary daily
func (b *BillClient) ListBillSummaryDaily(kt *kit.Kit, req *billproto.BillSummaryDailyListReq) (
	*billproto.BillSummaryDailyListResult, error) {
	return common.Request[billproto.BillSummaryDailyListReq, billproto.BillSummaryDailyListResult](
		b.client, rest.GET, kt, req, "/bills/summarydailys")
}

// --- bill summary version ---

// CreateBillSummaryVersion create bill summary version
func (b *BillClient) CreateBillSummaryVersion(kt *kit.Kit, req *billproto.BillSummaryVersionCreateReq) (
	*core.CreateResult, error) {
	return common.Request[billproto.BillSummaryVersionCreateReq, core.CreateResult](
		b.client, rest.POST, kt, req, "/bills/summaryversions")
}

// BatchDeleteBillSummaryVersion delete bill summary version
func (b *BillClient) BatchDeleteBillSummaryVersion(kt *kit.Kit, req *dataservice.BatchDeleteReq) error {
	return common.RequestNoResp[dataservice.BatchDeleteReq](
		b.client, rest.DELETE, kt, req, "/bills/summaryversions")
}

// UpdateBillSummaryVersion update bill summary version
func (b *BillClient) UpdateBillSummaryVersion(kt *kit.Kit, req *billproto.BillSummaryVersionUpdateReq) error {
	return common.RequestNoResp[billproto.BillSummaryVersionUpdateReq](
		b.client, rest.PUT, kt, req, "/bills/summaryversions")
}

// ListBillSummaryVersion list bill summary version
func (b *BillClient) ListBillSummaryVersion(kt *kit.Kit, req *billproto.BillSummaryVersionListReq) (
	*billproto.BillSummaryVersionListResult, error) {
	return common.Request[billproto.BillSummaryVersionListReq, billproto.BillSummaryVersionListResult](
		b.client, rest.GET, kt, req, "/bills/summaryversions")
}

// --- bill summary root ---

// CreateBillSummaryRoot create bill summary version
func (b *BillClient) CreateBillSummaryRoot(kt *kit.Kit, req *billproto.BillSummaryRootCreateReq) (
	*core.CreateResult, error) {
	return common.Request[billproto.BillSummaryRootCreateReq, core.CreateResult](
		b.client, rest.POST, kt, req, "/bills/summaryroots")
}

// UpdateBillSummaryRoot update bill summary version
func (b *BillClient) UpdateBillSummaryRoot(kt *kit.Kit, req *billproto.BillSummaryRootUpdateReq) error {
	return common.RequestNoResp[billproto.BillSummaryRootUpdateReq](
		b.client, rest.PUT, kt, req, "/bills/summaryroots")
}

// ListBillSummaryRoot list bill summary version
func (b *BillClient) ListBillSummaryRoot(kt *kit.Kit, req *billproto.BillSummaryRootListReq) (
	*billproto.BillSummaryRootListResult, error) {
	return common.Request[billproto.BillSummaryRootListReq, billproto.BillSummaryRootListResult](
		b.client, rest.GET, kt, req, "/bills/summaryroots")
}

// --- raw bill ---

// CreateRawBill create raw bill
func (b *BillClient) CreateRawBill(kt *kit.Kit, req *billproto.RawBillCreateReq) (*core.CreateResult, error) {
	return common.Request[billproto.RawBillCreateReq, core.CreateResult](
		b.client, rest.POST, kt, req, "/bills/rawbills")
}

// ListRawBillFileNames list raw bill file names
func (b *BillClient) ListRawBillFileNames(kt *kit.Kit, req *billproto.RawBillItemNameListReq) (
	*billproto.RawBillItemNameListResult, error) {

	return common.Request[billproto.RawBillItemNameListReq, billproto.RawBillItemNameListResult](
		b.client, rest.GET, kt, nil, fmt.Sprintf("/bills/rawbills/%s/%s/%s/%s/%s/%s/%s",
			req.Vendor, req.FirstAccountID, req.AccountID,
			req.BillYear, req.BillMonth, req.Version, req.BillDate))
}

// QueryRawBillItems get rawl bill item
func (b *BillClient) QueryRawBillItems(kt *kit.Kit, req *billproto.RawBillItemQueryReq) (
	*billproto.RawBillItemQueryResult, error) {

	return common.Request[billproto.RawBillItemQueryReq, billproto.RawBillItemQueryResult](
		b.client, rest.GET, kt, nil, fmt.Sprintf("/bills/rawbills/%s/%s/%s/%s/%s/%s/%s/%s",
			req.Vendor, req.FirstAccountID, req.AccountID,
			req.BillYear, req.BillMonth, req.Version, req.BillDate, req.FileName))
}

// --- bill item ---

// BatchCreateBillItem create bill item
func (b *BillClient) BatchCreateBillItem(
	kt *kit.Kit, vendor enumor.Vendor, req *billproto.BatchBillItemCreateReq[rawjson.RawMessage]) (
	*core.BatchCreateResult, error) {

	return common.Request[billproto.BatchBillItemCreateReq[rawjson.RawMessage], core.BatchCreateResult](
		b.client, rest.POST, kt, req, fmt.Sprintf("/vendors/%s/bills/rawitems/create", vendor))
}
