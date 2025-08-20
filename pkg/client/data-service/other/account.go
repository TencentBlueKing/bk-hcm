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

package other

import (
	"hcm/pkg/api/core"
	"hcm/pkg/api/core/cloud"
	protocloud "hcm/pkg/api/data-service/cloud"
	"hcm/pkg/client/common"
	"hcm/pkg/kit"
	"hcm/pkg/rest"
)

// AccountClient defines the client for RootAccount
type AccountClient struct {
	client rest.ClientInterface
}

// NewAccountClient ...
func NewAccountClient(client rest.ClientInterface) *AccountClient {
	return &AccountClient{
		client: client,
	}
}

// Create ...
func (a *AccountClient) Create(kt *kit.Kit,
	request *protocloud.AccountCreateReq[protocloud.OtherAccountExtensionCreateReq]) (
	*core.CreateResult, error) {

	return common.Request[protocloud.AccountCreateReq[protocloud.OtherAccountExtensionCreateReq], core.CreateResult](
		a.client, rest.POST, kt, request, "/accounts/create")
}

// Get other account detail.
func (a *AccountClient) Get(kt *kit.Kit, accountID string) (
	*protocloud.AccountGetResult[cloud.OtherAccountExtension], error) {

	return common.Request[common.Empty, protocloud.AccountGetResult[cloud.OtherAccountExtension]](
		a.client, rest.GET, kt, nil, "/accounts/%s", accountID)
}

// Update ...
func (a *AccountClient) Update(kt *kit.Kit, accountID string,
	request *protocloud.AccountUpdateReq[protocloud.OtherAccountExtensionUpdateReq]) (interface{}, error) {

	return common.Request[protocloud.AccountUpdateReq[protocloud.OtherAccountExtensionUpdateReq], interface{}](
		a.client, rest.PATCH, kt, request, "/accounts/%s", accountID)
}
