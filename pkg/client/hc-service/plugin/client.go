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

// Package plugin hc-service iac plugin api client.
package plugin

import (
	proto "hcm/pkg/api/hc-service/plugin"
	"hcm/pkg/client/common"
	"hcm/pkg/kit"
	"hcm/pkg/rest"
)

// Client is hc-service iac plugin api client.
type Client struct {
	client rest.ClientInterface
}

// NewClient create a new hc-service iac plugin api client.
func NewClient(client rest.ClientInterface) *Client {
	return &Client{client: client}
}

// RegisterHTTP 注册 http 插件。
func (c *Client) RegisterHTTP(kt *kit.Kit, req *proto.HTTPRegisterReq) (*proto.HTTPPluginResult, error) {
	return common.Request[proto.HTTPRegisterReq, proto.HTTPPluginResult](c.client, rest.POST, kt, req,
		"/iac_plugins/http/register")
}

// UpgradeHTTP 按插件名升级 http 插件，作用于该插件名下版本最高的记录。
func (c *Client) UpgradeHTTP(kt *kit.Kit, req *proto.HTTPUpgradeReq) (*proto.HTTPPluginResult, error) {
	return common.Request[proto.HTTPUpgradeReq, proto.HTTPPluginResult](c.client, rest.POST, kt, req,
		"/iac_plugins/http/upgrade")
}

// UpdateHTTP 按插件名修改 http 插件，作用于该插件名下版本最高的记录。
func (c *Client) UpdateHTTP(kt *kit.Kit, req *proto.HTTPUpdateReq) (*proto.HTTPPluginResult, error) {
	return common.Request[proto.HTTPUpdateReq, proto.HTTPPluginResult](c.client, rest.POST, kt, req,
		"/iac_plugins/http/update")
}
