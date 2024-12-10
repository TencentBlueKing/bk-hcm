/*
 * TencentBlueKing is pleased to support the open source community by making
 * 蓝鲸智云 - 混合云管理平台 (BlueKing - Hybrid Cloud Management System) available.
 * Copyright (C) 2024 THL A29 Limited,
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
	datagconf "hcm/pkg/api/data-service/global_config"
	"hcm/pkg/client/common"
	"hcm/pkg/kit"
	"hcm/pkg/rest"
)

// GlobalConfigsClient is data service global config api client.
type GlobalConfigsClient struct {
	client rest.ClientInterface
}

// NewGlobalConfigClient create a new global config api client.
func NewGlobalConfigClient(client rest.ClientInterface) *GlobalConfigsClient {
	return &GlobalConfigsClient{
		client: client,
	}
}

// List ...
func (g *GlobalConfigsClient) List(kt *kit.Kit, req *datagconf.ListReq) (
	*datagconf.ListResp, error) {

	return common.Request[datagconf.ListReq, datagconf.ListResp](
		g.client, rest.POST, kt, req, "/global_configs/list")
}

// BatchCreate ...
func (g *GlobalConfigsClient) BatchCreate(kt *kit.Kit, req *datagconf.BatchCreateReq) (
	*core.BatchCreateResult, error) {

	return common.Request[datagconf.BatchCreateReq, core.BatchCreateResult](
		g.client, rest.POST, kt, req, "/global_configs/batch/create")
}

// BatchUpdate ...
func (g *GlobalConfigsClient) BatchUpdate(kt *kit.Kit, req *datagconf.BatchUpdateReq) error {
	return common.RequestNoResp[datagconf.BatchUpdateReq](
		g.client, rest.PATCH, kt, req, "/global_configs/batch")
}

// BatchDelete ...
func (g *GlobalConfigsClient) BatchDelete(kt *kit.Kit, req *datagconf.BatchDeleteReq) error {
	return common.RequestNoResp[datagconf.BatchDeleteReq](
		g.client, rest.DELETE, kt, req, "/global_configs/batch")
}
