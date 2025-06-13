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
	coretenant "hcm/pkg/api/core/tenant"
	"hcm/pkg/api/data-service/tenant"
	"hcm/pkg/client/common"
	"hcm/pkg/kit"
	"hcm/pkg/rest"
)

// TenantClient is data service tenant api client.
type TenantClient struct {
	client rest.ClientInterface
}

// NewTenantClient create a new tenant api client.
func NewTenantClient(client rest.ClientInterface) *TenantClient {
	return &TenantClient{
		client: client,
	}
}

// List tenant.
func (t *TenantClient) List(kt *kit.Kit, req *core.ListReq) (*core.ListResultT[coretenant.Tenant], error) {
	return common.Request[core.ListReq, core.ListResultT[coretenant.Tenant]](
		t.client, rest.POST, kt, req, "/tenants/list")
}

// Create tenant.
func (t *TenantClient) Create(kt *kit.Kit, req *tenant.CreateTenantReq) (*core.BatchCreateResult, error) {
	return common.Request[tenant.CreateTenantReq, core.BatchCreateResult](
		t.client, rest.POST, kt, req, "/tenants/create")
}

// Update update tenant.
func (t *TenantClient) Update(kt *kit.Kit, req *tenant.UpdateTenantReq) error {
	return common.RequestNoResp[tenant.UpdateTenantReq](t.client, rest.PATCH, kt, req, "/tenants/update")
}
