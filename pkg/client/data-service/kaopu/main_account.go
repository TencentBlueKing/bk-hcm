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

package kaopu

import (
	"context"
	"net/http"

	"hcm/pkg/api/core"
	protocore "hcm/pkg/api/core/account-set"
	dataproto "hcm/pkg/api/data-service/account-set"
	"hcm/pkg/criteria/errf"
	"hcm/pkg/rest"
)

// MainAccountClient defines the client for MainAccount
type MainAccountClient struct {
	client rest.ClientInterface
}

// NewMainAccountClient ...
func NewMainAccountClient(client rest.ClientInterface) *MainAccountClient {
	return &MainAccountClient{
		client: client,
	}
}

// Create ...
func (a *MainAccountClient) Create(ctx context.Context, h http.Header,
	request *dataproto.MainAccountCreateReq[dataproto.KaopuMainAccountExtensionCreateReq]) (
	*core.CreateResult, error,
) {
	resp := new(core.CreateResp)

	err := a.client.Post().
		WithContext(ctx).
		Body(request).
		SubResourcef("/main_accounts/create").
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

// Get kaopu account detail.
func (a *MainAccountClient) Get(ctx context.Context, h http.Header, accountID string) (
	*dataproto.MainAccountGetResult[protocore.KaopuMainAccountExtension], error,
) {

	resp := new(dataproto.MainAccountGetResp[protocore.KaopuMainAccountExtension])

	err := a.client.Get().
		WithContext(ctx).
		SubResourcef("/main_accounts/%s", accountID).
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
