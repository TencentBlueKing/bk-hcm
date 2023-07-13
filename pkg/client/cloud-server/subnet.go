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

package cloudserver

import (
	"context"
	"net/http"

	proto "hcm/pkg/api/cloud-server"
	"hcm/pkg/api/core"
	"hcm/pkg/criteria/errf"
	"hcm/pkg/rest"
)

// SubnetClient is data service subnet api client.
type SubnetClient struct {
	client rest.ClientInterface
}

// NewSubnetClient create a new subnet api client.
func NewSubnetClient(client rest.ClientInterface) *SubnetClient {
	return &SubnetClient{
		client: client,
	}
}

// ListInRes subnets.
func (v *SubnetClient) ListInRes(ctx context.Context, h http.Header, req *core.ListReq) (
	*proto.SubnetListResult, error) {

	resp := new(proto.SubnetListResp)

	err := v.client.Post().
		WithContext(ctx).
		Body(req).
		SubResourcef("/subnets/list").
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

// ListInBiz subnets.
func (v *SubnetClient) ListInBiz(ctx context.Context, h http.Header, bizID int64, req *core.ListReq) (
	*proto.SubnetListResult, error) {

	resp := new(proto.SubnetListResp)

	err := v.client.Post().
		WithContext(ctx).
		Body(req).
		SubResourcef("/bizs/%d/subnets/list", bizID).
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

// ListCountIPInBiz 获取子网可用IP.
func (v *SubnetClient) ListCountIPInBiz(ctx context.Context, h http.Header, bizID int64,
	req *proto.ListSubnetCountIPReq) (map[string]proto.SubnetCountIPResult, error) {

	resp := new(proto.ListSubnetCountIPResp)

	err := v.client.Post().
		WithContext(ctx).
		Body(req).
		SubResourcef("/bizs/%d/subnets/ips/count/list", bizID).
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

// ListCountIPInRes 获取子网可用IP.
func (v *SubnetClient) ListCountIPInRes(ctx context.Context, h http.Header, req *proto.ListSubnetCountIPReq) (
	map[string]proto.SubnetCountIPResult, error) {

	resp := new(proto.ListSubnetCountIPResp)

	err := v.client.Post().
		WithContext(ctx).
		Body(req).
		SubResourcef("/subnets/ips/count/list").
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
