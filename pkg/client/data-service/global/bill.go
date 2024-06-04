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
	"net/http"

	"hcm/pkg/api/core"
	dataservice "hcm/pkg/api/data-service"
	billproto "hcm/pkg/api/data-service/bill"
	datacloudbillproto "hcm/pkg/api/data-service/cloud/bill"
	"hcm/pkg/client/common"
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
		b.client, rest.POST, kt, req, "/bill/adjustmentitems")
}

// BatchDeleteBillAdjustmentItem delete bill adjustment item
func (b *BillClient) BatchDeleteBillAdjustmentItem(kt *kit.Kit, req *dataservice.BatchDeleteReq) error {
	return common.RequestNoResp[dataservice.BatchDeleteReq](
		b.client, rest.DELETE, kt, req, "/bill/adjustmentitems")
}

// UpdateBillAdjustmentItem update bill adjustment item
func (b *BillClient) UpdateBillAdjustmentItem(kt *kit.Kit, req *billproto.BillAdjustmentItemUpdateReq) error {
	return common.RequestNoResp[billproto.BillAdjustmentItemUpdateReq](
		b.client, rest.PATCH, kt, req, "/bill/adjustmentitems")
}

// ListBillAdjustmentItem list bill adjustment item
func (b *BillClient) ListBillAdjustmentItem(kt *kit.Kit, req *billproto.BillAdjustmentItemListReq) (
	*billproto.BillAdjustmentItemListResult, error) {
	return common.Request[billproto.BillAdjustmentItemListReq, billproto.BillAdjustmentItemListResult](
		b.client, rest.GET, kt, req, "/bill/adjustmentitems")
}

// --- bill item ---

// BatchCreateBillItem create bill item
func (b *BillClient) BatchCreateBillItem(kt *kit.Kit, req *billproto.BatchBillItemCreateReq) (
	*core.BatchCreateResult, error) {
	return common.Request[billproto.BatchBillItemCreateReq, core.BatchCreateResult](
		b.client, rest.POST, kt, req, "/bill/items")
}

// BatchDeleteBillItem delete bill item
func (b *BillClient) BatchDeleteBillItem(kt *kit.Kit, req *dataservice.BatchDeleteReq) error {
	return common.RequestNoResp[dataservice.BatchDeleteReq](
		b.client, rest.DELETE, kt, req, "/bill/items")
}

// UpdateBillItem update bill item
func (b *BillClient) UpdateBillItem(kt *kit.Kit, req *billproto.BillItemUpdateReq) error {
	return common.RequestNoResp[billproto.BillItemUpdateReq](
		b.client, rest.PATCH, kt, req, "/bill/items")
}

// ListBillItem list bill item
func (b *BillClient) ListBillItem(kt *kit.Kit, req *billproto.BillItemListReq) (
	*billproto.BillItemListResult, error) {
	return common.Request[billproto.BillItemListReq, billproto.BillItemListResult](
		b.client, rest.GET, kt, req, "/bill/items")
}

// --- bill daily pull task ---

// CreateBillDailyPullTask create bill daily pull task
func (b *BillClient) CreateBillDailyPullTask(kt *kit.Kit, req *billproto.BillDailyPullTaskCreateReq) (
	*core.CreateResult, error) {
	return common.Request[billproto.BillDailyPullTaskCreateReq, core.CreateResult](
		b.client, rest.POST, kt, req, "/bill/dailypulltasks")
}

// BatchDeleteBillDailyPullTask delete bill daily pull task
func (b *BillClient) BatchDeleteBillDailyPullTask(kt *kit.Kit, req *dataservice.BatchDeleteReq) error {
	return common.RequestNoResp[dataservice.BatchDeleteReq](
		b.client, rest.DELETE, kt, req, "/bill/dailypulltasks")
}

// UpdateBillDailyPullTask update bill daily pull task
func (b *BillClient) UpdateBillDailyPullTask(kt *kit.Kit, req *billproto.BillDailyPullTaskUpdateReq) error {
	return common.RequestNoResp[billproto.BillDailyPullTaskUpdateReq](
		b.client, rest.PATCH, kt, req, "/bill/dailypulltasks")
}

// ListBillDailyPullTask list bill daily pull task
func (b *BillClient) ListBillDailyPullTask(kt *kit.Kit, req *billproto.BillDailyPullTaskListReq) (
	*billproto.BillDailyPullTaskListResult, error) {
	return common.Request[billproto.BillDailyPullTaskListReq, billproto.BillDailyPullTaskListResult](
		b.client, rest.GET, kt, req, "/bill/dailypulltasks")
}

// --- bill puller ---

// CreateBillPuller create bill puller
func (b *BillClient) CreateBillPuller(kt *kit.Kit, req *billproto.BillPullerCreateReq) (
	*core.CreateResult, error) {
	return common.Request[billproto.BillPullerCreateReq, core.CreateResult](
		b.client, rest.POST, kt, req, "/bill/pullers")
}

// BatchDeleteBillPuller delete bill puller
func (b *BillClient) BatchDeleteBillPuller(kt *kit.Kit, req *dataservice.BatchDeleteReq) error {
	return common.RequestNoResp[dataservice.BatchDeleteReq](
		b.client, rest.DELETE, kt, req, "/bill/pullers")
}

// UpdateBillPuller update bill puller
func (b *BillClient) UpdateBillPuller(kt *kit.Kit, req *billproto.BillPullerUpdateReq) error {
	return common.RequestNoResp[billproto.BillPullerUpdateReq](
		b.client, rest.PATCH, kt, req, "/bill/pullers")
}

// ListBillPuller list bill puller
func (b *BillClient) ListBillPuller(kt *kit.Kit, req *billproto.BillPullerListReq) (
	*billproto.BillPullerListResult, error) {
	return common.Request[billproto.BillPullerListReq, billproto.BillPullerListResult](
		b.client, rest.GET, kt, req, "/bill/pullers")
}

// --- bill summary ---

// CreateBillSummary create bill summary
func (b *BillClient) CreateBillSummary(kt *kit.Kit, req *billproto.BillSummaryCreateReq) (
	*core.CreateResult, error) {
	return common.Request[billproto.BillSummaryCreateReq, core.CreateResult](
		b.client, rest.POST, kt, req, "/bill/summarys")
}

// BatchDeleteBillSummary delete bill summary
func (b *BillClient) BatchDeleteBillSummary(kt *kit.Kit, req *dataservice.BatchDeleteReq) error {
	return common.RequestNoResp[dataservice.BatchDeleteReq](
		b.client, rest.DELETE, kt, req, "/bill/summarys")
}

// UpdateBillSummary update bill summary
func (b *BillClient) UpdateBillSummary(kt *kit.Kit, req *billproto.BillSummaryUpdateReq) error {
	return common.RequestNoResp[billproto.BillSummaryUpdateReq](
		b.client, rest.PATCH, kt, req, "/bill/summarys")
}

// ListBillSummary list bill summary
func (b *BillClient) ListBillSummary(kt *kit.Kit, req *billproto.BillSummaryListReq) (
	*billproto.BillSummaryListResult, error) {
	return common.Request[billproto.BillSummaryListReq, billproto.BillSummaryListResult](
		b.client, rest.GET, kt, req, "/bill/summarys")
}

// --- bill summary daily ---

// CreateBillSummaryDaily create bill summary daily
func (b *BillClient) CreateBillSummaryDaily(kt *kit.Kit, req *billproto.BillSummaryDailyCreateReq) (
	*core.CreateResult, error) {
	return common.Request[billproto.BillSummaryDailyCreateReq, core.CreateResult](
		b.client, rest.POST, kt, req, "/bill/summarydailys")
}

// BatchDeleteBillSummaryDaily delete bill summary daily
func (b *BillClient) BatchDeleteBillSummaryDaily(kt *kit.Kit, req *dataservice.BatchDeleteReq) error {
	return common.RequestNoResp[dataservice.BatchDeleteReq](
		b.client, rest.DELETE, kt, req, "/bill/summarydailys")
}

// UpdateBillSummaryDaily update bill summary daily
func (b *BillClient) UpdateBillSummaryDaily(kt *kit.Kit, req *billproto.BillSummaryDailyUpdateReq) error {
	return common.RequestNoResp[billproto.BillSummaryDailyUpdateReq](
		b.client, rest.PATCH, kt, req, "/bill/summarydailys")
}

// ListBillSummaryDaily list bill summary daily
func (b *BillClient) ListBillSummaryDaily(kt *kit.Kit, req *billproto.BillSummaryDailyListReq) (
	*billproto.BillSummaryDailyListResult, error) {
	return common.Request[billproto.BillSummaryDailyListReq, billproto.BillSummaryDailyListResult](
		b.client, rest.GET, kt, req, "/bill/summarydailys")
}

// --- bill summary version ---

// CreateBillSummaryVersion create bill summary version
func (b *BillClient) CreateBillSummaryVersion(kt *kit.Kit, req *billproto.BillSummaryVersionCreateReq) (
	*core.CreateResult, error) {
	return common.Request[billproto.BillSummaryVersionCreateReq, core.CreateResult](
		b.client, rest.POST, kt, req, "/bill/summaryversions")
}

// BatchDeleteBillSummaryVersion delete bill summary version
func (b *BillClient) BatchDeleteBillSummaryVersion(kt *kit.Kit, req *dataservice.BatchDeleteReq) error {
	return common.RequestNoResp[dataservice.BatchDeleteReq](
		b.client, rest.DELETE, kt, req, "/bill/summaryversions")
}

// UpdateBillSummaryVersion update bill summary version
func (b *BillClient) UpdateBillSummaryVersion(kt *kit.Kit, req *billproto.BillSummaryVersionUpdateReq) error {
	return common.RequestNoResp[billproto.BillSummaryVersionUpdateReq](
		b.client, rest.PATCH, kt, req, "/bill/summaryversions")
}

// ListBillSummaryVersion list bill summary version
func (b *BillClient) ListBillSummaryVersion(kt *kit.Kit, req *billproto.BillSummaryVersionListReq) (
	*billproto.BillSummaryVersionListResult, error) {
	return common.Request[billproto.BillSummaryVersionListReq, billproto.BillSummaryVersionListResult](
		b.client, rest.GET, kt, req, "/bill/summaryversions")
}

// --  raw bill ---

// CreateRawBill create raw bill
func (b *BillClient) CreateRawBill(kt *kit.Kit, req *billproto.RawBillCreateReq) (*core.CreateResult, error) {
	return common.Request[billproto.RawBillCreateReq, core.CreateResult](
		b.client, rest.POST, kt, req, "/bill/rawbills")
}
