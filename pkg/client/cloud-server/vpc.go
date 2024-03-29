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
	"hcm/pkg/client/common"
	"hcm/pkg/criteria/enumor"
	"hcm/pkg/criteria/errf"
	"hcm/pkg/kit"
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

// Assign 分配vpc到业务下
func (v *VpcClient) Assign(kt *kit.Kit, req *csvpc.AssignVpcToBizReq) error {
	resp := new(rest.BaseResp)

	err := v.client.Post().
		WithContext(kt.Ctx).
		Body(req).
		SubResourcef("/vpcs/assign/bizs").
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

// Update vpc
func (v *VpcClient) Update(kt *kit.Kit, id string, req *csvpc.VpcUpdateReq) error {
	resp := new(rest.BaseResp)

	err := v.client.Patch().
		WithContext(kt.Ctx).
		Body(req).
		SubResourcef("/vpcs/%s", id).
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

// UpdateInBiz update vpc in business
func (v *VpcClient) UpdateInBiz(kt *kit.Kit, bizID int64, id string, req *csvpc.VpcUpdateReq) error {
	resp := new(rest.BaseResp)

	err := v.client.Patch().
		WithContext(kt.Ctx).
		Body(req).
		SubResourcef("/bizs/%d/vpcs/%s", bizID, id).
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

// ListInRes vpcs.
func (v *VpcClient) ListInRes(kt *kit.Kit, req *core.ListReq) (
	*csvpc.VpcListResult, error) {

	resp := new(csvpc.VpcListResp)

	err := v.client.Post().
		WithContext(kt.Ctx).
		Body(req).
		SubResourcef("/vpcs/list").
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

// ListInBiz vpcs.
func (v *VpcClient) ListInBiz(kt *kit.Kit, bizID int64, req *core.ListReq) (
	*csvpc.VpcListResult, error) {

	resp := new(csvpc.VpcListResp)

	err := v.client.Post().
		WithContext(kt.Ctx).
		Body(req).
		SubResourcef("/bizs/%d/vpcs/list", bizID).
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

// GetInBiz 获取业务下vpc详情
func (v *VpcClient) GetInBiz(kt *kit.Kit, bizID int, vpcID string) (*corecloud.BaseVpc, error) {

	resp := new(core.BaseResp[*corecloud.BaseVpc])

	err := v.client.Get().
		WithContext(kt.Ctx).
		Body(struct{}{}).
		SubResourcef("/bizs/%d/vpcs/%s", bizID, vpcID).
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

// DeleteInBiz delete vpc in business
func (v *VpcClient) DeleteInBiz(kt *kit.Kit, bizID int, vpcID string) error {

	resp := new(rest.BaseResp)
	err := v.client.Delete().
		WithContext(kt.Ctx).
		Body(struct{}{}).
		SubResourcef("/bizs/%d/vpcs/%s", bizID, vpcID).
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

// Delete in res
func (v *VpcClient) Delete(kt *kit.Kit, vpcID string) error {

	return common.RequestNoResp[common.Empty](v.client, rest.DELETE, kt, common.NoData, "/vpcs/%s", vpcID)
}

// CreateTCloudVpc ...
func (v *VpcClient) CreateTCloudVpc(kt *kit.Kit, req *csvpc.TCloudVpcCreateReq) (*core.CreateResult, error) {
	resp := new(core.CreateResp)

	err := v.client.Post().
		WithContext(kt.Ctx).
		Body(req).
		SubResourcef("/vpcs/create").
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
