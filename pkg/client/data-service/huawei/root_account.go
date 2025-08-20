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
	"hcm/pkg/api/core"
	protocore "hcm/pkg/api/core/account-set"
	dataproto "hcm/pkg/api/data-service/account-set"
	"hcm/pkg/client/common"
	"hcm/pkg/kit"
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
func (a *RootAccountClient) Create(kt *kit.Kit,
	request *dataproto.RootAccountCreateReq[dataproto.HuaWeiRootAccountExtensionCreateReq]) (
	*core.CreateResult, error,
) {

	return common.Request[dataproto.RootAccountCreateReq[dataproto.HuaWeiRootAccountExtensionCreateReq],
		core.CreateResult](a.client, rest.POST, kt, request, "/root_accounts/create")
}

// Get huawei account detail.
func (a *RootAccountClient) Get(kt *kit.Kit, accountID string) (
	*dataproto.RootAccountGetResult[protocore.HuaWeiRootAccountExtension], error,
) {

	return common.Request[common.Empty, dataproto.RootAccountGetResult[protocore.HuaWeiRootAccountExtension]](
		a.client, rest.GET, kt, nil, "/root_accounts/%s", accountID)
}

// Update ...
func (a *RootAccountClient) Update(kt *kit.Kit, accountID string,
	request *dataproto.RootAccountUpdateReq[dataproto.HuaWeiRootAccountExtensionUpdateReq]) (
	interface{}, error,
) {

	return common.Request[dataproto.RootAccountUpdateReq[dataproto.HuaWeiRootAccountExtensionUpdateReq], interface{}](
		a.client, rest.PATCH, kt, request, "/root_accounts/%s", accountID)
}
