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

// Package cloudserver defines cloud-server api client.
package cloudserver

import (
	"fmt"

	"hcm/pkg/rest"
	"hcm/pkg/rest/client"
)

// Client is cloud-server api client.
type Client struct {
	client rest.ClientInterface
}

// NewClient create a new cloud-server api client.
func NewClient(c *client.Capability, version string) *Client {
	base := fmt.Sprintf("/api/%s/cloud", version)
	return &Client{
		client: rest.NewClient(c, base),
	}
}

// Account get account client.
func (c *Client) Account() *AccountClient {
	return NewAccountClient(c.client)
}

// Cvm get cvm client.
func (c *Client) Cvm() *CvmClient {
	return NewCvmClient(c.client)
}
