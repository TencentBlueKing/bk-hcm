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
	dataproto "hcm/pkg/api/hc-service/main-account"
	"hcm/pkg/client/common"
	"hcm/pkg/kit"
	"hcm/pkg/rest"
)

// NewMainAccountClient create a new image api client.
func NewMainAccountClient(client rest.ClientInterface) *MainAccountClient {
	return &MainAccountClient{
		client: client,
	}
}

// MainAccountClient main account client.
type MainAccountClient struct {
	client rest.ClientInterface
}

// Create ...
func (m *MainAccountClient) Create(kt *kit.Kit,
	request *dataproto.CreateAwsMainAccountReq) (*dataproto.CreateAwsMainAccountResp, error) {

	return common.Request[dataproto.CreateAwsMainAccountReq, dataproto.CreateAwsMainAccountResp](
		m.client, rest.POST, kt, request, "/main_accounts/create")
}
