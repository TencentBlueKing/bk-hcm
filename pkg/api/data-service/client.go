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

// Package dataservice defines data-service api client.
package dataservice

import (
	"fmt"

	"hcm/pkg/rest"
	"hcm/pkg/rest/client"
)

// Client is data-service api client.
type Client struct {
	client rest.ClientInterface
}

// NewClient create a new data-service api client.
func NewClient(c *client.Capability, version string) *Client {
	base := fmt.Sprintf("/api/%s/data", version)
	return &Client{
		client: rest.NewClient(c, base),
	}
}

// Account get account client.
func (c *Client) Account() *AccountClient {
	return NewAccountClient(c.client)
}

// Auth get api client for authorize use.
func (c *Client) Auth() *AuthClient {
	return NewAuthClient(c.client)
}
