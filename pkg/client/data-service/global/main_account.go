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
	"hcm/pkg/api/core"
	dataproto "hcm/pkg/api/data-service/account-set"
	"hcm/pkg/client/common"
	"hcm/pkg/kit"
	"hcm/pkg/rest"
)

// MainAccountClient is data service account api client.
type MainAccountClient struct {
	client rest.ClientInterface
}

// NewMainAccountClient create a new account api client.
func NewMainAccountClient(client rest.ClientInterface) *MainAccountClient {
	return &MainAccountClient{
		client: client,
	}
}

// GetBasicInfo ...
func (a *MainAccountClient) GetBasicInfo(kt *kit.Kit, accountID string) (
	*dataproto.MainAccountGetBaseResult, error,
) {

	return common.Request[common.Empty, dataproto.MainAccountGetBaseResult](
		a.client, rest.GET, kt, nil, "/main_accounts/basic_info/%s", accountID,
	)
}

// List ...
func (a *MainAccountClient) List(kt *kit.Kit, request *core.ListReq) (
	*dataproto.MainAccountListResult, error,
) {

	return common.Request[core.ListReq, dataproto.MainAccountListResult](
		a.client, rest.POST, kt, request, "/main_accounts/list")
}

// Update ...
func (a *MainAccountClient) Update(kt *kit.Kit, accountID string,
	request *dataproto.MainAccountUpdateReq) (
	interface{}, error,
) {

	return common.Request[dataproto.MainAccountUpdateReq, interface{}](
		a.client, rest.PATCH, kt, request, "/main_accounts/%s", accountID,
	)
}
