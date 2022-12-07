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

package cloud

import (
	"context"
	"net/http"

	"hcm/pkg/api/protocol/base"
	protocloud "hcm/pkg/api/protocol/data-service/cloud"
	"hcm/pkg/criteria/errf"
	"hcm/pkg/rest"
)

type AccountBizRelClient struct {
	client rest.ClientInterface
}

// NewCloudAccountClient create a new account api client.
func NewAccountBizRelClient(client rest.ClientInterface) *AccountBizRelClient {
	return &AccountBizRelClient{
		client: client,
	}
}

// Create ...
func (a *AccountBizRelClient) Create(ctx context.Context, h http.Header, request *protocloud.CreateAccountBizRelReq) (
	*base.CreateResult, error,
) {
	resp := new(base.CreateResp)

	err := a.client.Post().
		WithContext(ctx).
		Body(request).
		SubResourcef("/cloud/account_biz_rels/create/").
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

// Update ...
func (a *AccountBizRelClient) Update(ctx context.Context, h http.Header, request *protocloud.UpdateAccountBizRelsReq) (
	interface{}, error,
) {
	resp := new(base.UpdateResp)

	err := a.client.Post().
		WithContext(ctx).
		Body(request).
		SubResourcef("/cloud/account_biz_rels/update/").
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

// List ...
func (a *AccountBizRelClient) List(ctx context.Context, h http.Header, request *protocloud.ListAccountBizRelsReq) (
	*protocloud.ListAccountBizRelsResult, error,
) {
	resp := new(protocloud.ListAccountBizRelsResp)

	err := a.client.Post().
		WithContext(ctx).
		Body(request).
		SubResourcef("/cloud/account_biz_rels/list/").
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

// Delete ...
func (a *AccountBizRelClient) Delete(ctx context.Context, h http.Header, request *protocloud.DeleteAccountBizRelsReq) (
	interface{}, error,
) {
	resp := new(base.DeleteResp)

	err := a.client.Post().
		WithContext(ctx).
		Body(request).
		SubResourcef("/cloud/account_biz_rels/delete/").
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
