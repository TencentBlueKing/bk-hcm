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

	"hcm/pkg/api/core"
	corert "hcm/pkg/api/core/cloud/route-table"
	dataservice "hcm/pkg/api/data-service"
	routetable "hcm/pkg/api/data-service/cloud/route-table"
	"hcm/pkg/criteria/errf"
	"hcm/pkg/rest"
)

// RouteTableClient is data service route table api client.
type RouteTableClient struct {
	client rest.ClientInterface
}

// NewRouteTableClient create a new route table api client.
func NewRouteTableClient(client rest.ClientInterface) *RouteTableClient {
	return &RouteTableClient{
		client: client,
	}
}

// BatchCreate batch create aws route table.
func (r *RouteTableClient) BatchCreate(ctx context.Context, h http.Header,
	req *routetable.RouteTableBatchCreateReq[routetable.AwsRouteTableCreateExt]) (*core.BatchCreateResult, error) {

	resp := new(core.BatchCreateResp)

	err := r.client.Post().
		WithContext(ctx).
		Body(req).
		SubResourcef("/route_tables/batch/create").
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

// Get aws route table.
func (r *RouteTableClient) Get(ctx context.Context, h http.Header, id string) (
	*corert.RouteTable[corert.AwsRouteTableExtension], error) {

	resp := new(routetable.RouteTableGetResp[corert.AwsRouteTableExtension])

	err := r.client.Get().
		WithContext(ctx).
		SubResourcef("/route_tables/%s", id).
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

// BatchCreateRoute batch create aws route.
func (r *RouteTableClient) BatchCreateRoute(ctx context.Context, h http.Header, routeTableID string,
	req *routetable.AwsRouteBatchCreateReq) (*core.BatchCreateResult, error) {

	resp := new(core.BatchCreateResp)

	err := r.client.Post().
		WithContext(ctx).
		Body(req).
		SubResourcef("/route_tables/%s/routes/batch/create", routeTableID).
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

// ListRoute list aws route.
func (r *RouteTableClient) ListRoute(ctx context.Context, h http.Header, routeTableID string, req *core.ListReq) (
	*routetable.AwsRouteListResult, error) {

	resp := new(routetable.AwsRouteListResp)

	err := r.client.Post().
		WithContext(ctx).
		Body(req).
		SubResourcef("/route_tables/%s/routes/list", routeTableID).
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

// BatchUpdateRoute batch update aws route.
func (r *RouteTableClient) BatchUpdateRoute(ctx context.Context, h http.Header, routeTableID string,
	req *routetable.AwsRouteBatchUpdateReq) error {

	resp := new(rest.BaseResp)

	err := r.client.Patch().
		WithContext(ctx).
		Body(req).
		SubResourcef("/route_tables/%s/routes/batch", routeTableID).
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

// BatchDeleteRoute batch delete aws route.
func (r *RouteTableClient) BatchDeleteRoute(ctx context.Context, h http.Header, routeTableID string,
	req *dataservice.BatchDeleteReq) error {

	resp := new(rest.BaseResp)

	err := r.client.Delete().
		WithContext(ctx).
		Body(req).
		SubResourcef("/route_tables/%s/routes/batch", routeTableID).
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
