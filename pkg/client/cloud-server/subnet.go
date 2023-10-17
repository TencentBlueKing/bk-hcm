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
	"hcm/pkg/client/common"
	"hcm/pkg/criteria/errf"
	"hcm/pkg/kit"
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
func (v *SubnetClient) ListInRes(kt *kit.Kit, req *core.ListReq) (
	*proto.SubnetListResult, error) {

	resp := new(proto.SubnetListResp)

	err := v.client.Post().
		WithContext(kt.Ctx).
		Body(req).
		SubResourcef("/subnets/list").
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

// ListInBiz subnets.
func (v *SubnetClient) ListInBiz(kt *kit.Kit, bizID int64, req *core.ListReq) (
	*proto.SubnetListResult, error) {

	resp := new(proto.SubnetListResp)

	err := v.client.Post().
		WithContext(kt.Ctx).
		Body(req).
		SubResourcef("/bizs/%d/subnets/list", bizID).
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

// Assign subnet to business
func (v *SubnetClient) Assign(kt *kit.Kit, req *proto.AssignSubnetToBizReq) error {
	resp := new(rest.BaseResp)

	err := v.client.Post().
		WithContext(kt.Ctx).
		Body(req).
		SubResourcef("/subnets/assign/bizs").
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

// Create subnet
func (v *SubnetClient) Create(kt *kit.Kit, req *proto.TCloudSubnetCreateReq) (*core.CreateResult,
	error) {
	resp := new(core.BaseResp[*core.CreateResult])

	err := v.client.Post().
		WithContext(kt.Ctx).
		Body(req).
		SubResourcef("/subnets/create").
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

// CreateInBiz create subnet in business
func (v *SubnetClient) CreateInBiz(kt *kit.Kit, bizID int64, req *proto.TCloudSubnetCreateReq) (*core.CreateResult,
	error) {
	resp := new(core.BaseResp[*core.CreateResult])

	err := v.client.Post().
		WithContext(kt.Ctx).
		Body(req).
		SubResourcef("/bizs/%d/subnets/create", bizID).
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

// Update subnet
func (v *SubnetClient) Update(kt *kit.Kit, subnetID string, req *proto.SubnetUpdateReq) error {
	return common.RequestNoResp[proto.SubnetUpdateReq](v.client, rest.PATCH, kt, req,
		"/subnets/%s", subnetID)
}

// UpdateInBiz update subnet in biz
func (v *SubnetClient) UpdateInBiz(kt *kit.Kit, bizID int64, subnetID string, req *proto.SubnetUpdateReq) error {
	return common.RequestNoResp[proto.SubnetUpdateReq](v.client, rest.PATCH, kt, req,
		"/bizs/%d/subnets/%s", bizID, subnetID)
}

// Delete subnet
func (v *SubnetClient) Delete(kt *kit.Kit, req *proto.BatchDeleteReq) error {
	return common.RequestNoResp[proto.BatchDeleteReq](v.client, rest.DELETE, kt, req,
		"/subnets/batch")
}

// DeleteInBiz delete subnet in biz
func (v *SubnetClient) DeleteInBiz(kt *kit.Kit, bizID int64, req *proto.BatchDeleteReq) error {
	return common.RequestNoResp[proto.BatchDeleteReq](v.client, rest.DELETE, kt, req,
		"/bizs/%d/subnets/batch", bizID)
}
