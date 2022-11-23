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

package dataservice

import (
	"context"
	"net/http"

	"hcm/pkg/api/protocol/base"
	dataservice "hcm/pkg/api/protocol/data-service"
	"hcm/pkg/criteria/errf"
	"hcm/pkg/rest"
)

// AccountClient is data service account api client.
type AccountClient struct {
	client rest.ClientInterface
}

// NewAccountClient create a new account api client.
func NewAccountClient(client rest.ClientInterface) *AccountClient {
	return &AccountClient{
		client: client,
	}
}

// Create account.
func (a *AccountClient) Create(ctx context.Context, h http.Header, request *dataservice.CreateAccountReq) (
	*base.CreateResult, error) {

	resp := new(base.CreateResp)

	err := a.client.Post().
		WithContext(ctx).
		Body(request).
		SubResourcef("/create/account/account").
		WithHeaders(h).
		Do().
		Into(resp)

	if resp.Code != errf.OK {
		return nil, errf.New(resp.Code, resp.Message)
	}

	return resp.Data, err
}
