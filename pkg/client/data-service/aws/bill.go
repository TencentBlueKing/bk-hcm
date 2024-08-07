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

package aws

import (
	"context"
	"net/http"

	"hcm/pkg/api/core"
	"hcm/pkg/api/core/cloud"
	billproto "hcm/pkg/api/data-service/bill"
	protobill "hcm/pkg/api/data-service/cloud/bill"
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

// BatchCreate batch create aws bill.
func (v *BillClient) BatchCreate(ctx context.Context, h http.Header,
	req *protobill.AccountBillConfigBatchCreateReq[cloud.AwsBillConfigExtension]) (*core.BatchCreateResult, error) {

	resp := new(core.BatchCreateResp)

	err := v.client.Post().
		WithContext(ctx).
		Body(req).
		SubResourcef("/bills/config/batch/create").
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

// BatchUpdate aws bill update.
func (v *BillClient) BatchUpdate(ctx context.Context, h http.Header,
	req *protobill.AccountBillConfigBatchUpdateReq[cloud.AwsBillConfigExtension]) error {

	resp := new(rest.BaseResp)

	err := v.client.Patch().
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

// List list bill.
func (v *BillClient) List(ctx context.Context, h http.Header, req *core.ListReq) (
	*protobill.AccountBillConfigExtListResult[cloud.AwsBillConfigExtension], error) {

	resp := new(protobill.AccountBillConfigExtListResp[cloud.AwsBillConfigExtension])

	err := v.client.Post().
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

// ListBillItem list bill item
func (b *BillClient) ListBillItem(kt *kit.Kit, req *billproto.BillItemListReq) (
	*billproto.AwsBillItemListResult, error) {

	return common.Request[billproto.BillItemListReq, billproto.AwsBillItemListResult](
		b.client, rest.POST, kt, req, "/bills/items/list")
}
