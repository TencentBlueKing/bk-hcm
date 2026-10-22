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

// IacPluginSchemaClient data-service iac_plugin_schema API 客户端。
type IacPluginSchemaClient struct {
	client rest.ClientInterface
}

// NewIacPluginSchemaClient 创建 IacPluginSchemaClient。
func NewIacPluginSchemaClient(client rest.ClientInterface) *IacPluginSchemaClient {
	return &IacPluginSchemaClient{
		client: client,
	}
}

// BatchCreate 批量创建 iac_plugin_schema 记录。
func (c *IacPluginSchemaClient) BatchCreate(kt *kit.Kit, req *dataplugin.IacPluginSchemaBatchCreateReq) (
	*core.BatchCreateResult, error) {

	return common.Request[dataplugin.IacPluginSchemaBatchCreateReq, core.BatchCreateResult](
		c.client, rest.POST, kt, req, "/iac_plugin_schema/batch/create")
}

// BatchUpdate 批量更新 iac_plugin_schema 记录。
func (c *IacPluginSchemaClient) BatchUpdate(kt *kit.Kit, req *dataplugin.IacPluginSchemaBatchUpdateReq) error {
	return common.RequestNoResp[dataplugin.IacPluginSchemaBatchUpdateReq](
		c.client, rest.PATCH, kt, req, "/iac_plugin_schema/batch/update")
}

// BatchDelete 批量删除 iac_plugin_schema 记录。
func (c *IacPluginSchemaClient) BatchDelete(kt *kit.Kit, req *dataplugin.IacPluginSchemaBatchDeleteReq) error {
	return common.RequestNoResp[dataplugin.IacPluginSchemaBatchDeleteReq](
		c.client, rest.DELETE, kt, req, "/iac_plugin_schema/batch")
}

// List 列表查询 iac_plugin_schema，支持 fields 投影。
func (c *IacPluginSchemaClient) List(kt *kit.Kit, req *dataplugin.IacPluginSchemaListReq) (
	*dataplugin.IacPluginSchemaListResult, error) {

	return common.Request[dataplugin.IacPluginSchemaListReq, dataplugin.IacPluginSchemaListResult](
		c.client, rest.POST, kt, req, "/iac_plugin_schema/list")
}
