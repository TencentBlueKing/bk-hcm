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

	csvpc "hcm/pkg/api/cloud-server/vpc"
	"hcm/pkg/api/core"
	corecloud "hcm/pkg/api/core/cloud"
	protocloud "hcm/pkg/api/data-service/cloud"
	"hcm/pkg/criteria/enumor"
	"hcm/pkg/criteria/errf"
	"hcm/pkg/rest"
)

// VpcClient is data service vpc api client.
type VpcClient struct {
	client rest.ClientInterface
}

// NewVpcClient create a new vpc api client.
func NewVpcClient(client rest.ClientInterface) *VpcClient {
	return &VpcClient{
		client: client,
	}
}

// ListInRes vpcs.
func (v *VpcClient) ListInRes(ctx context.Context, h http.Header, req *core.ListReq) (
	*csvpc.VpcListResult, error) {

	resp := new(csvpc.VpcListResp)

	err := v.client.Post().
		WithContext(ctx).
		Body(req).
		SubResourcef("/vpcs/list").
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

// ListInBiz vpcs.
func (v *VpcClient) ListInBiz(ctx context.Context, h http.Header, bizID int64, req *core.ListReq) (
	*csvpc.VpcListResult, error) {

	resp := new(csvpc.VpcListResp)

	err := v.client.Post().
		WithContext(ctx).
		Body(req).
		SubResourcef("/bizs/%d/vpcs/list", bizID).
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

// TCloudListExtInBiz ...
func (v *VpcClient) TCloudListExtInBiz(ctx context.Context, h http.Header, bizID int64, req *core.ListReq) (
	*protocloud.VpcExtListResult[corecloud.TCloudVpcExtension], error) {
	return listVpcExt[corecloud.TCloudVpcExtension](ctx, h, v.client, bizID, enumor.TCloud, req)
}

// AwsListExtInBiz ...
func (v *VpcClient) AwsListExtInBiz(ctx context.Context, h http.Header, bizID int64, req *core.ListReq) (
	*protocloud.VpcExtListResult[corecloud.AwsVpcExtension], error) {
	return listVpcExt[corecloud.AwsVpcExtension](ctx, h, v.client, bizID, enumor.Aws, req)
}

// HuaWeiListExtInBiz ...
func (v *VpcClient) HuaWeiListExtInBiz(ctx context.Context, h http.Header, bizID int64, req *core.ListReq) (
	*protocloud.VpcExtListResult[corecloud.HuaWeiVpcExtension], error) {
	return listVpcExt[corecloud.HuaWeiVpcExtension](ctx, h, v.client, bizID, enumor.HuaWei, req)
}

// GcpListExtInBiz ...
func (v *VpcClient) GcpListExtInBiz(ctx context.Context, h http.Header, bizID int64, req *core.ListReq) (
	*protocloud.VpcExtListResult[corecloud.GcpVpcExtension], error) {
	return listVpcExt[corecloud.GcpVpcExtension](ctx, h, v.client, bizID, enumor.Gcp, req)
}

// AzureListExtInBiz ...
func (v *VpcClient) AzureListExtInBiz(ctx context.Context, h http.Header, bizID int64, req *core.ListReq) (
	*protocloud.VpcExtListResult[corecloud.AzureVpcExtension], error) {
	return listVpcExt[corecloud.AzureVpcExtension](ctx, h, v.client, bizID, enumor.Azure, req)
}

// listVpcExt list vpc with extension.
func listVpcExt[T corecloud.VpcExtension](ctx context.Context, h http.Header, cli rest.ClientInterface, bizID int64,
	vendor enumor.Vendor, req *core.ListReq) (*protocloud.VpcExtListResult[T], error) {

	resp := new(protocloud.VpcExtListResp[T])

	err := cli.Post().
		WithContext(ctx).
		Body(req).
		SubResourcef("/bizs/%d/vendors/%s/vpcs/list", bizID, vendor).
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

// TCloudListExtInRes ...
func (v *VpcClient) TCloudListExtInRes(ctx context.Context, h http.Header, req *core.ListReq) (
	*protocloud.VpcExtListResult[corecloud.TCloudVpcExtension], error) {
	return listVpcExtInRes[corecloud.TCloudVpcExtension](ctx, h, v.client, enumor.TCloud, req)
}

// AwsListExtInRes ...
func (v *VpcClient) AwsListExtInRes(ctx context.Context, h http.Header, req *core.ListReq) (
	*protocloud.VpcExtListResult[corecloud.AwsVpcExtension], error) {
	return listVpcExtInRes[corecloud.AwsVpcExtension](ctx, h, v.client, enumor.Aws, req)
}

// HuaWeiListExtInRes ...
func (v *VpcClient) HuaWeiListExtInRes(ctx context.Context, h http.Header, req *core.ListReq) (
	*protocloud.VpcExtListResult[corecloud.HuaWeiVpcExtension], error) {
	return listVpcExtInRes[corecloud.HuaWeiVpcExtension](ctx, h, v.client, enumor.HuaWei, req)
}

// GcpListExtInRes ...
func (v *VpcClient) GcpListExtInRes(ctx context.Context, h http.Header, req *core.ListReq) (
	*protocloud.VpcExtListResult[corecloud.GcpVpcExtension], error) {
	return listVpcExtInRes[corecloud.GcpVpcExtension](ctx, h, v.client, enumor.Gcp, req)
}

// AzureListExtInRes ...
func (v *VpcClient) AzureListExtInRes(ctx context.Context, h http.Header, req *core.ListReq) (
	*protocloud.VpcExtListResult[corecloud.AzureVpcExtension], error) {
	return listVpcExtInRes[corecloud.AzureVpcExtension](ctx, h, v.client, enumor.Azure, req)
}

// listVpcExtInRes list vpc with extension.
func listVpcExtInRes[T corecloud.VpcExtension](ctx context.Context, h http.Header, cli rest.ClientInterface,
	vendor enumor.Vendor, req *core.ListReq) (*protocloud.VpcExtListResult[T], error) {

	resp := new(protocloud.VpcExtListResp[T])

	err := cli.Post().
		WithContext(ctx).
		Body(req).
		SubResourcef("/vendors/%s/vpcs/list", vendor).
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
