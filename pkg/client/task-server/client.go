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
	"fmt"

	"hcm/pkg/api/core"
	"hcm/pkg/api/core/task"
	apits "hcm/pkg/api/task-server"
	"hcm/pkg/criteria/errf"
	"hcm/pkg/kit"
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

// AddAsyncFlow add async flow.
func (c *Client) AddAsyncFlow(kt *kit.Kit, request *apits.AddFlowReq) (string, error) {
	resp := new(apits.FlowAddResp)

	err := c.client.Post().
		WithContext(kt.Ctx).
		Body(request).
		SubResourcef("/async/flows/tpls/add").
		WithHeaders(kt.Header()).
		Do().
		Into(resp)

	if resp.Code != errf.OK {
		return "", errf.New(resp.Code, resp.Message)
	}

	return resp.FlowID, err
}

// ListAsyncFlow list async flow.
func (c *Client) ListAsyncFlow(kt *kit.Kit, req *core.ListReq) (*apits.FlowListResult, error) {
	resp := new(apits.FlowListResp)

	err := c.client.Post().
		WithContext(kt.Ctx).
		Body(req).
		SubResourcef("/async/flows/list").
		WithHeaders(kt.Header()).
		Do().
		Into(resp)

	if resp.Code != errf.OK {
		return nil, errf.New(resp.Code, resp.Message)
	}

	return resp.Data, err
}

// GetAsyncFlow get async flow.
func (c *Client) GetAsyncFlow(kt *kit.Kit, flowID string) (*task.AsyncFlow, error) {
	resp := new(apits.FlowResp)

	err := c.client.Get().
		WithContext(kt.Ctx).
		SubResourcef("/async/flows/%s", flowID).
		WithHeaders(kt.Header()).
		Do().
		Into(resp)
	if err != nil {
		return nil, err
	}

	if resp.Code != errf.OK {
		return nil, errf.New(resp.Code, resp.Message)
	}

	return resp.Data, nil
}
