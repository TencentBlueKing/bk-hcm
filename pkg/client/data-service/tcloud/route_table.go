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

package tcloud

import (
	"context"
	"net/http"

	"hcm/pkg/api/core"
	corert "hcm/pkg/api/core/cloud/route-table"
	dataservice "hcm/pkg/api/data-service"
	routetable "hcm/pkg/api/data-service/cloud/route-table"
	"hcm/pkg/criteria/errf"
	"hcm/pkg/kit"
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

// BatchCreate batch create tcloud route table.
func (r *RouteTableClient) BatchCreate(ctx context.Context, h http.Header,
	req *routetable.RouteTableBatchCreateReq[routetable.TCloudRouteTableCreateExt]) (*core.BatchCreateResult, error) {

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

// Get tcloud route table.
func (r *RouteTableClient) Get(ctx context.Context, h http.Header, id string) (
	*corert.RouteTable[corert.TCloudRouteTableExtension], error) {

	resp := new(routetable.RouteTableGetResp[corert.TCloudRouteTableExtension])

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

// BatchCreateRoute batch create tcloud route.
func (r *RouteTableClient) BatchCreateRoute(kt *kit.Kit, routeTableID string,
	req *routetable.TCloudRouteBatchCreateReq) (*core.BatchCreateResult, error) {

	resp := new(core.BatchCreateResp)

	err := r.client.Post().
		WithContext(kt.Ctx).
		Body(req).
		SubResourcef("/route_tables/%s/routes/batch/create", routeTableID).
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

// ListRoute list tcloud route.
func (r *RouteTableClient) ListRoute(ctx context.Context, h http.Header, routeTableID string, req *core.ListReq) (
	*routetable.TCloudRouteListResult, error) {

	resp := new(routetable.TCloudRouteListResp)

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

// ListAllRoute list tcloud route.
func (r *RouteTableClient) ListAllRoute(ctx context.Context, h http.Header, req *routetable.TCloudRouteListReq) (
	*routetable.TCloudRouteListResult, error) {

	resp := new(routetable.TCloudRouteListResp)

	err := r.client.Post().
		WithContext(ctx).
		Body(req).
		SubResourcef("/route_tables/route_id/routes/list/all").
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

// BatchUpdateRoute batch update tcloud route.
func (r *RouteTableClient) BatchUpdateRoute(kt *kit.Kit, routeTableID string,
	req *routetable.TCloudRouteBatchUpdateReq) error {

	resp := new(rest.BaseResp)

	err := r.client.Patch().
		WithContext(kt.Ctx).
		Body(req).
		SubResourcef("/route_tables/%s/routes/batch", routeTableID).
		WithHeaders(kt.Header()).
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

// BatchDeleteRoute batch delete tcloud route.
func (r *RouteTableClient) BatchDeleteRoute(kt *kit.Kit, routeTableID string, req *dataservice.BatchDeleteReq) error {

	resp := new(rest.BaseResp)

	err := r.client.Delete().
		WithContext(kt.Ctx).
		Body(req).
		SubResourcef("/route_tables/%s/routes/batch", routeTableID).
		WithHeaders(kt.Header()).
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

// ListRouteTableWithExt list route tables extension.
func (r *RouteTableClient) ListRouteTableWithExt(ctx context.Context, h http.Header, req *core.ListReq) (
	[]*corert.RouteTable[corert.TCloudRouteTableExtension], error) {

	resp := new(routetable.RouteTableListExtResp[corert.TCloudRouteTableExtension])

	err := r.client.Post().
		WithContext(ctx).
		Body(req).
		SubResourcef("/route_tables/list").
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
