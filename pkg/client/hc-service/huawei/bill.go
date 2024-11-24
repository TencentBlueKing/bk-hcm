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

package huawei

import (
	"context"
	"net/http"

	hcbillservice "hcm/pkg/api/hc-service/bill"
	"hcm/pkg/client/common"
	"hcm/pkg/criteria/errf"
	"hcm/pkg/kit"
	"hcm/pkg/rest"
)

// BillClient is hc service bill api client.
type BillClient struct {
	client rest.ClientInterface
}

// NewBillClient create a new bill api client.
func NewBillClient(client rest.ClientInterface) *BillClient {
	return &BillClient{
		client: client,
	}
}

// List bill.
func (v *BillClient) List(ctx context.Context, h http.Header, req *hcbillservice.HuaWeiBillListReq) (
	*hcbillservice.HuaWeiBillListResult, error) {

	resp := new(hcbillservice.HuaWeiBillListResp)

	err := v.client.Post().
		WithContext(ctx).
		Body(req).
		SubResourcef("/bills/list").
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

// ListFeeRecord list fee record
func (v *BillClient) ListFeeRecord(kt *kit.Kit, req *hcbillservice.HuaWeiFeeRecordListReq) (
	*hcbillservice.HuaWeiRootBillListResult, error) {

	return common.Request[hcbillservice.HuaWeiFeeRecordListReq, hcbillservice.HuaWeiRootBillListResult](v.client,
		rest.POST, kt, req, "/feerecords/list")
}
