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

package azure

import (
	"context"
	"net/http"

	"hcm/pkg/api/core"
	corecloud "hcm/pkg/api/core/cloud"
	protocloud "hcm/pkg/api/data-service/cloud"
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

// BatchCreate batch create azure vpc.
func (v *VpcClient) BatchCreate(ctx context.Context, h http.Header,
	req *protocloud.VpcBatchCreateReq[protocloud.AzureVpcCreateExt]) (*core.BatchCreateResult, error) {

	resp := new(core.BatchCreateResp)

	err := v.client.Post().
		WithContext(ctx).
		Body(req).
		SubResourcef("/vpcs/batch/create").
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

// Get azure vpc.
func (v *VpcClient) Get(kt *kit.Kit, id string) (*corecloud.Vpc[corecloud.AzureVpcExtension],
	error) {

	resp := new(protocloud.VpcGetResp[corecloud.AzureVpcExtension])

	err := v.client.Get().
		WithContext(kt.Ctx).
		SubResourcef("/vpcs/%s", id).
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

// BatchUpdate azure vpc.
func (v *VpcClient) BatchUpdate(ctx context.Context, h http.Header,
	req *protocloud.VpcBatchUpdateReq[protocloud.AzureVpcUpdateExt]) error {

	resp := new(rest.BaseResp)

	err := v.client.Patch().
		WithContext(ctx).
		Body(req).
		SubResourcef("/vpcs/batch").
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

// ListVpcExt list vpc with extension.
func (v *VpcClient) ListVpcExt(ctx context.Context, h http.Header, req *core.ListReq) (
	*protocloud.VpcExtListResult[corecloud.AzureVpcExtension], error) {

	resp := new(protocloud.VpcExtListResp[corecloud.AzureVpcExtension])

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
