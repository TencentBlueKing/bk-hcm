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
	dataproto "hcm/pkg/api/data-service/account-set"
	"hcm/pkg/client/common"
	"hcm/pkg/criteria/errf"
	"hcm/pkg/kit"
	"hcm/pkg/rest"
)

// MainAccountClient is data service account api client.
type MainAccountClient struct {
	client rest.ClientInterface
}

// NewAccountClient create a new account api client.
func NewMainAccountClient(client rest.ClientInterface) *MainAccountClient {
	return &MainAccountClient{
		client: client,
	}
}

// GetBasicInfo ...
func (a *MainAccountClient) GetBasicInfo(kt *kit.Kit, h http.Header, accountID string) (
	*dataproto.MainAccountGetBaseResult, error,
) {
	resp := new(dataproto.MainAccountGetBaseResp)

	err := a.client.Get().
		WithContext(kt.Ctx).
		SubResourcef("/main_accounts/basic_info/%s", accountID).
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
func (a *MainAccountClient) List(kt *kit.Kit, request *core.ListWithoutFieldReq) (
	*dataproto.MainAccountListResult, error,
) {

	return common.Request[core.ListWithoutFieldReq, dataproto.MainAccountListResult](
		a.client, rest.POST, kt, request, "/main_accounts/list")
}

// Update ...
func (a *MainAccountClient) Update(ctx context.Context, h http.Header, accountID string,
	request *dataproto.MainAccountUpdateReq) (
	interface{}, error,
) {

	resp := new(core.UpdateResp)

	err := a.client.Patch().
		WithContext(ctx).
		Body(request).
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
