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
	coreasync "hcm/pkg/api/core/async"
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

// CreateTemplateFlow add template flow.
func (c *Client) CreateTemplateFlow(kt *kit.Kit, request *apits.AddTemplateFlowReq) (*core.CreateResult, error) {
	resp := new(core.CreateResp)

	err := c.client.Post().
		WithContext(kt.Ctx).
		Body(request).
		SubResourcef("/template_flows/create").
		WithHeaders(kt.Header()).
		Do().
		Into(resp)

	if resp.Code != errf.OK {
		return nil, errf.New(resp.Code, resp.Message)
	}

	return resp.Data, err
}

// CreateCustomFlow add custom flow.
func (c *Client) CreateCustomFlow(kt *kit.Kit, request *apits.AddCustomFlowReq) (*core.CreateResult, error) {
	resp := new(core.CreateResp)

	err := c.client.Post().
		WithContext(kt.Ctx).
		Body(request).
		SubResourcef("/custom_flows/create").
		WithHeaders(kt.Header()).
		Do().
		Into(resp)

	if resp.Code != errf.OK {
		return nil, errf.New(resp.Code, resp.Message)
	}

	return resp.Data, err
}

// ListFlow list flow.
func (c *Client) ListFlow(kt *kit.Kit, req *core.ListReq) (*apits.ListFlowResult, error) {
	resp := new(core.BaseResp[*apits.ListFlowResult])

	err := c.client.Post().
		WithContext(kt.Ctx).
		Body(req).
		SubResourcef("/flows/list").
		WithHeaders(kt.Header()).
		Do().
		Into(resp)

	if resp.Code != errf.OK {
		return nil, errf.New(resp.Code, resp.Message)
	}

	return resp.Data, err
}

// GetFlow get flow.
func (c *Client) GetFlow(kt *kit.Kit, id string) (*coreasync.AsyncFlow, error) {
	resp := new(core.BaseResp[*coreasync.AsyncFlow])

	err := c.client.Get().
		WithContext(kt.Ctx).
		SubResourcef("/flows/%s", id).
		WithHeaders(kt.Header()).
		Do().
		Into(resp)

	if resp.Code != errf.OK {
		return nil, errf.New(resp.Code, resp.Message)
	}

	return resp.Data, err
}

// ListTask list task.
func (c *Client) ListTask(kt *kit.Kit, req *core.ListReq) (*apits.ListTaskResult, error) {
	resp := new(core.BaseResp[*apits.ListTaskResult])

	err := c.client.Post().
		WithContext(kt.Ctx).
		Body(req).
		SubResourcef("/tasks/list").
		WithHeaders(kt.Header()).
		Do().
		Into(resp)

	if resp.Code != errf.OK {
		return nil, errf.New(resp.Code, resp.Message)
	}

	return resp.Data, err
}

// GetTask get task.
func (c *Client) GetTask(kt *kit.Kit, id string) (*coreasync.AsyncFlowTask, error) {
	resp := new(core.BaseResp[*coreasync.AsyncFlowTask])

	err := c.client.Get().
		WithContext(kt.Ctx).
		SubResourcef("/tasks/%s", id).
		WithHeaders(kt.Header()).
		Do().
		Into(resp)

	if resp.Code != errf.OK {
		return nil, errf.New(resp.Code, resp.Message)
	}

	return resp.Data, err
}
