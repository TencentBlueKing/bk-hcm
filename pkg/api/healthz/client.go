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

// Package healthz defines health check client.
package healthz

import (
	"context"

	"hcm/pkg/criteria/errf"
	"hcm/pkg/rest"
	"hcm/pkg/rest/client"
)

// NewClient create a new health check api client.
func NewClient(c *client.Capability) *Client {
	return &Client{
		client: rest.NewClient(c, "/"),
	}
}

// Client is health check api client.
type Client struct {
	client rest.ClientInterface
}

// HealthCheck check if service is healthy, returns error if service is not.
func (c *Client) HealthCheck() error {
	resp := new(rest.BaseResp)

	err := c.client.Get().
		WithContext(context.Background()).
		SubResourcef("/healthz").
		Body(nil).
		Do().
		Into(resp)
	if err != nil {
		return err
	}

	if resp.Code != errf.OK {
		return errf.New(resp.Code, resp.Message)
	}

	return nil
}
