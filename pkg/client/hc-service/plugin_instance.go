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

package hcservice

import (
	"hcm/pkg/api/core"
	coreplugin "hcm/pkg/api/core/plugin"
	hcplugin "hcm/pkg/api/hc-service/plugin"
	"hcm/pkg/client/common"
	"hcm/pkg/kit"
	"hcm/pkg/rest"
)

// PluginInstanceClient hc-service plugin_instance API 客户端。
type PluginInstanceClient struct {
	client rest.ClientInterface
}

// NewPluginInstanceClient 创建 PluginInstanceClient。
func NewPluginInstanceClient(client rest.ClientInterface) *PluginInstanceClient {
	return &PluginInstanceClient{
		client: client,
	}
}

// Create 创建插件实例。
func (c *PluginInstanceClient) Create(kt *kit.Kit, req *hcplugin.PluginInstanceCreateReq) (
	*core.CreateResult, error) {

	return common.Request[hcplugin.PluginInstanceCreateReq, core.CreateResult](
		c.client, rest.POST, kt, req, "/plugin_instances/create")
}

// Update 按 ID 更新插件实例。
func (c *PluginInstanceClient) Update(kt *kit.Kit, id string, req *hcplugin.PluginInstanceUpdateReq) error {
	return common.RequestNoResp[hcplugin.PluginInstanceUpdateReq](
		c.client, rest.PATCH, kt, req, "/plugin_instances/%s", id)
}

// Delete 按 ID 删除插件实例。
func (c *PluginInstanceClient) Delete(kt *kit.Kit, id string) error {
	return common.RequestNoResp[common.Empty](
		c.client, rest.DELETE, kt, common.NoData, "/plugin_instances/%s", id)
}

// Get 按 ID 获取单条插件实例详情。
func (c *PluginInstanceClient) Get(kt *kit.Kit, id string) (*coreplugin.PluginInstance, error) {
	return common.Request[common.Empty, coreplugin.PluginInstance](
		c.client, rest.GET, kt, common.NoData, "/plugin_instances/%s", id)
}

// List 列表查询插件实例。
func (c *PluginInstanceClient) List(kt *kit.Kit, req *hcplugin.PluginInstanceListReq) (
	*hcplugin.PluginInstanceListResult, error) {

	return common.Request[hcplugin.PluginInstanceListReq, hcplugin.PluginInstanceListResult](
		c.client, rest.POST, kt, req, "/plugin_instances/list")
}
