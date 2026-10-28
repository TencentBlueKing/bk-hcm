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
	dataplugin "hcm/pkg/api/data-service/plugin"
	"hcm/pkg/client/common"
	"hcm/pkg/kit"
	"hcm/pkg/rest"
)

// IaCPluginClient is data service iac plugin api client.
type IaCPluginClient struct {
	client rest.ClientInterface
}

// NewIaCPluginClient create a new iac plugin api client.
func NewIaCPluginClient(client rest.ClientInterface) *IaCPluginClient {
	return &IaCPluginClient{client: client}
}

// Create create iac plugin record.
func (c *IaCPluginClient) Create(kt *kit.Kit, req *dataplugin.IaCPluginCreateReq) (*core.BatchCreateResult, error) {
	return common.Request[dataplugin.IaCPluginCreateReq, core.BatchCreateResult](
		c.client, rest.POST, kt, req, "/iac_plugins/create")
}

// BatchDelete 按过滤条件删除 iac plugin 记录。
func (c *IaCPluginClient) BatchDelete(kt *kit.Kit, req *dataplugin.IaCPluginBatchDeleteReq) error {
	return common.RequestNoResp[dataplugin.IaCPluginBatchDeleteReq](
		c.client, rest.DELETE, kt, req, "/iac_plugins/batch")
}

// Update update iac plugin record, set req.ExpectState to make it a conditional update.
func (c *IaCPluginClient) Update(kt *kit.Kit, req *dataplugin.IaCPluginUpdateReq) error {
	return common.RequestNoResp[dataplugin.IaCPluginUpdateReq](
		c.client, rest.PATCH, kt, req, "/iac_plugins/update")
}

// BatchUpdateState batch update the state of the iac plugin by plugin name.
func (c *IaCPluginClient) BatchUpdateState(kt *kit.Kit, req *dataplugin.IaCPluginBatchUpdateStateReq) error {
	return common.RequestNoResp[dataplugin.IaCPluginBatchUpdateStateReq](
		c.client, rest.PATCH, kt, req, "/iac_plugins/batch/update_state")
}

// List list iac plugin records.
func (c *IaCPluginClient) List(kt *kit.Kit, req *core.ListReq) (*dataplugin.IaCPluginListResult, error) {
	return common.Request[core.ListReq, dataplugin.IaCPluginListResult](
		c.client, rest.POST, kt, req, "/iac_plugins/list")
}
