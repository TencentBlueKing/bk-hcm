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
	"context"
	"net/http"

	"hcm/pkg/api/core"
	proto "hcm/pkg/api/data-service"
	"hcm/pkg/criteria/errf"
	"hcm/pkg/kit"
	"hcm/pkg/rest"
)

// ApplicationClient is data service application api client.
type ApplicationClient struct {
	client rest.ClientInterface
}

// NewApplicationClient create a new application api client.
func NewApplicationClient(client rest.ClientInterface) *ApplicationClient {
	return &ApplicationClient{
		client: client,
	}
}

// CreateApplication ...
func (a *ApplicationClient) CreateApplication(ctx context.Context, h http.Header, request *proto.ApplicationCreateReq) (
	*core.CreateResult, error,
) {
	resp := new(core.CreateResp)

	err := a.client.Post().
		WithContext(ctx).
		Body(request).
		SubResourcef("/applications/create").
		WithHeaders(h).
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

// UpdateApplication ...
func (a *ApplicationClient) UpdateApplication(kt *kit.Kit, id string, request *proto.ApplicationUpdateReq) (interface{}, error) {
	resp := new(core.UpdateResp)

	err := a.client.Patch().
		WithContext(kt.Ctx).
		Body(request).
		SubResourcef("/applications/%s", id).
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

// GetApplication ...
func (a *ApplicationClient) GetApplication(ctx context.Context, h http.Header, applicationID string) (
	*proto.ApplicationResp, error,
) {
	resp := new(proto.ApplicationGetResp)

	err := a.client.Get().
		WithContext(ctx).
		SubResourcef("/applications/%s", applicationID).
		WithHeaders(h).
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

// ListApplication ...
func (a *ApplicationClient) ListApplication(kt *kit.Kit, request *proto.ApplicationListReq) (
	*proto.ApplicationListResult, error,
) {
	resp := new(proto.ApplicationListResp)

	err := a.client.Post().
		WithContext(kt.Ctx).
		Body(request).
		SubResourcef("/applications/list").
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
