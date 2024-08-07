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

package gcp

import (
	"context"
	"net/http"

	"hcm/pkg/api/core"
	billcore "hcm/pkg/api/core/bill"
	protobill "hcm/pkg/api/data-service/bill"
	"hcm/pkg/criteria/errf"
	"hcm/pkg/rest"
)

// RootAccountBillConfigClient is data service bill api client.
type RootAccountBillConfigClient struct {
	client rest.ClientInterface
}

// NewRootAccountBillConfigClient create a new bill api client.
func NewRootAccountBillConfigClient(client rest.ClientInterface) *RootAccountBillConfigClient {
	return &RootAccountBillConfigClient{
		client: client,
	}
}

// BatchCreate batch create aws bill.
func (v *RootAccountBillConfigClient) BatchCreate(ctx context.Context, h http.Header,
	req *protobill.RootAccountBillConfigBatchCreateReq[billcore.GcpBillConfigExtension]) (
	*core.BatchCreateResult, error) {

	resp := new(core.BatchCreateResp)

	err := v.client.Post().
		WithContext(ctx).
		Body(req).
		SubResourcef("/bills/root_account_config/batch/create").
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
func (v *RootAccountBillConfigClient) BatchUpdate(ctx context.Context, h http.Header,
	req *protobill.RootAccountBillConfigBatchUpdateReq[billcore.GcpBillConfigExtension]) error {

	resp := new(rest.BaseResp)

	err := v.client.Patch().
		WithContext(ctx).
		Body(req).
		SubResourcef("/bills/root_account_config/batch").
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
func (v *RootAccountBillConfigClient) List(ctx context.Context, h http.Header, req *core.ListReq) (
	*protobill.RootAccountBillConfigExtListResult[billcore.GcpBillConfigExtension], error) {

	resp := new(protobill.RootAccountBillConfigExtListResp[billcore.GcpBillConfigExtension])

	err := v.client.Post().
		WithContext(ctx).
		Body(req).
		SubResourcef("/bills/root_account_config/list").
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
