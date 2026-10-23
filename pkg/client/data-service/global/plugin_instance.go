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

// IacPluginInstanceClient data-service iac_plugin_instance API 客户端。
type IacPluginInstanceClient struct {
	client rest.ClientInterface
}

// NewIacPluginInstanceClient 创建 IacPluginInstanceClient。
func NewIacPluginInstanceClient(client rest.ClientInterface) *IacPluginInstanceClient {
	return &IacPluginInstanceClient{
		client: client,
	}
}

// BatchCreate 批量创建 iac_plugin_instance 记录。
func (c *IacPluginInstanceClient) BatchCreate(kt *kit.Kit, req *dataplugin.IacPluginInstanceBatchCreateReq) (
	*core.BatchCreateResult, error) {

	return common.Request[dataplugin.IacPluginInstanceBatchCreateReq, core.BatchCreateResult](
		c.client, rest.POST, kt, req, "/iac_plugin_instance/batch/create")
}

// BatchUpdate 批量更新 iac_plugin_instance 记录。
func (c *IacPluginInstanceClient) BatchUpdate(kt *kit.Kit, req *dataplugin.IacPluginInstanceBatchUpdateReq) error {
	return common.RequestNoResp[dataplugin.IacPluginInstanceBatchUpdateReq](
		c.client, rest.PATCH, kt, req, "/iac_plugin_instance/batch/update")
}

// BatchDelete 批量删除 iac_plugin_instance 记录。
func (c *IacPluginInstanceClient) BatchDelete(kt *kit.Kit, req *dataplugin.IacPluginInstanceBatchDeleteReq) error {
	return common.RequestNoResp[dataplugin.IacPluginInstanceBatchDeleteReq](
		c.client, rest.DELETE, kt, req, "/iac_plugin_instance/batch")
}

// List 列表查询 iac_plugin_instance 记录。
func (c *IacPluginInstanceClient) List(kt *kit.Kit, req *dataplugin.IacPluginInstanceListReq) (
	*dataplugin.IacPluginInstanceListResult, error) {

	return common.Request[dataplugin.IacPluginInstanceListReq, dataplugin.IacPluginInstanceListResult](
		c.client, rest.POST, kt, req, "/iac_plugin_instance/list")
}
