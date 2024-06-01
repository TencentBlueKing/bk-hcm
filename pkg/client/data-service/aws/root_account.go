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
	protocore "hcm/pkg/api/core/account-set"
	dataproto "hcm/pkg/api/data-service/account-set"
	"hcm/pkg/criteria/errf"
	"hcm/pkg/rest"
)

// RootAccountClient defines the client for RootAccount
type RootAccountClient struct {
	client rest.ClientInterface
}

// NewRootAccountClient ...
func NewRootAccountClient(client rest.ClientInterface) *RootAccountClient {
	return &RootAccountClient{
		client: client,
	}
}

// Create ...
func (a *RootAccountClient) Create(ctx context.Context, h http.Header,
	request *dataproto.RootAccountCreateReq[dataproto.AwsRootAccountExtensionCreateReq]) (
	*core.CreateResult, error,
) {
	resp := new(core.CreateResp)

	err := a.client.Post().
		WithContext(ctx).
		Body(request).
		SubResourcef("/root_accounts/create").
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

// Get aws account detail.
func (a *RootAccountClient) Get(ctx context.Context, h http.Header, accountID string) (
	*dataproto.RootAccountGetResult[protocore.AwsRootAccountExtension], error,
) {

	resp := new(dataproto.RootAccountGetResp[protocore.AwsRootAccountExtension])

	err := a.client.Get().
		WithContext(ctx).
		SubResourcef("/root_accounts/%s", accountID).
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
func (a *RootAccountClient) Update(ctx context.Context, h http.Header, accountID string,
	request *dataproto.RootAccountUpdateReq[dataproto.AwsRootAccountExtensionUpdateReq]) (
	interface{}, error,
) {
	resp := new(core.UpdateResp)

	err := a.client.Patch().
		WithContext(ctx).
		Body(request).
		SubResourcef("/root_accounts/%s", accountID).
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
