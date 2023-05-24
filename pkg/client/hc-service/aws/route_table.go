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

package aws

import (
	"context"
	"net/http"

	routetable "hcm/pkg/api/hc-service/route-table"
	"hcm/pkg/api/hc-service/sync"
	"hcm/pkg/criteria/errf"
	"hcm/pkg/rest"
)

// RouteTableClient is hc service aws route table api client.
type RouteTableClient struct {
	client rest.ClientInterface
}

// NewRouteTableClient create a new route table api client.
func NewRouteTableClient(client rest.ClientInterface) *RouteTableClient {
	return &RouteTableClient{
		client: client,
	}
}

// Update route table.
func (r *RouteTableClient) Update(ctx context.Context, h http.Header, id string,
	req *routetable.RouteTableUpdateReq) error {

	resp := new(rest.BaseResp)

	err := r.client.Patch().
		WithContext(ctx).
		Body(req).
		SubResourcef("/route_tables/%s", id).
		WithHeaders(h).
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

// Delete route table.
func (r *RouteTableClient) Delete(ctx context.Context, h http.Header, id string) error {
	resp := new(rest.BaseResp)

	err := r.client.Delete().
		WithContext(ctx).
		Body(nil).
		SubResourcef("/route_tables/%s", id).
		WithHeaders(h).
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

// UpdateRoute update route.
func (r *RouteTableClient) UpdateRoute(ctx context.Context, h http.Header, id string,
	req *routetable.RouteUpdateReq) error {

	resp := new(rest.BaseResp)

	err := r.client.Patch().
		WithContext(ctx).
		Body(req).
		SubResourcef("/routes/%s", id).
		WithHeaders(h).
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

// DeleteRoute delete route.
func (r *RouteTableClient) DeleteRoute(ctx context.Context, h http.Header, id string) error {
	resp := new(rest.BaseResp)

	err := r.client.Delete().
		WithContext(ctx).
		Body(nil).
		SubResourcef("/routes/%s", id).
		WithHeaders(h).
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

// SyncRouteTable route table.
func (r *RouteTableClient) SyncRouteTable(ctx context.Context, h http.Header, req *sync.AwsSyncReq) error {

	resp := new(rest.BaseResp)

	err := r.client.Post().
		WithContext(ctx).
		Body(req).
		SubResourcef("/route_tables/sync").
		WithHeaders(h).
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
