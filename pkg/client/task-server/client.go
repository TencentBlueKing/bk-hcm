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

package taskserver

import (
	"context"
	"fmt"
	"net/http"

	taskserver "hcm/pkg/api/task-server"
	"hcm/pkg/criteria/errf"
	"hcm/pkg/rest"
	"hcm/pkg/rest/client"
)

// Client is task-server api client.
type Client struct {
	client rest.ClientInterface
}

// NewClient create a new task-server api client.
func NewClient(c *client.Capability, version string) *Client {
	base := fmt.Sprintf("/api/%s/task", version)
	return &Client{
		client: rest.NewClient(c, base),
	}
}

// AsyncTaskCreate create async task.
func (c *Client) AsyncTaskCreate(ctx context.Context, h http.Header, request *taskserver.AsyncTask) error {
	resp := new(rest.BaseResp)

	err := c.client.Post().
		WithContext(ctx).
		Body(request).
		SubResourcef("/async/tasks/create").
		WithHeaders(h).
		Do().
		Into(resp)

	if resp.Code != errf.OK {
		return errf.New(resp.Code, resp.Message)
	}

	return err
}
