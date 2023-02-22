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
	corecloud "hcm/pkg/api/core/cloud"
	proto "hcm/pkg/api/data-service"
	protocloud "hcm/pkg/api/data-service/cloud"
	"hcm/pkg/criteria/errf"
	"hcm/pkg/rest"
)

// NewVpcCvmRelClient create a new vpc & cvm relation api client.
func NewVpcCvmRelClient(client rest.ClientInterface) *VpcCvmRelClient {
	return &VpcCvmRelClient{
		client: client,
	}
}

// VpcCvmRelClient is data service vpc cvm rel api client.
type VpcCvmRelClient struct {
	client rest.ClientInterface
}

// BatchCreate vpc cvm relations.
func (cli *VpcCvmRelClient) BatchCreate(ctx context.Context, h http.Header,
	request *protocloud.VpcCvmRelBatchCreateReq) error {

	resp := new(rest.BaseResp)

	err := cli.client.Post().
		WithContext(ctx).
		Body(request).
		SubResourcef("/vpc_cvm_rels/batch/create").
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

// BatchDelete vpc cvm relations.
func (cli *VpcCvmRelClient) BatchDelete(ctx context.Context, h http.Header, request *proto.BatchDeleteReq) error {

	resp := new(rest.BaseResp)

	err := cli.client.Delete().
		WithContext(ctx).
		Body(request).
		SubResourcef("/vpc_cvm_rels/batch").
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

// List vpc cvm relations.
func (cli *VpcCvmRelClient) List(ctx context.Context, h http.Header, request *core.ListReq) (
	*protocloud.VpcCvmRelListResult, error) {

	resp := new(protocloud.VpcCvmRelListResp)

	err := cli.client.Post().
		WithContext(ctx).
		Body(request).
		SubResourcef("/vpc_cvm_rels/list").
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

// ListWithVpc list vpc cvm relations with vpc details.
func (cli *VpcCvmRelClient) ListWithVpc(ctx context.Context, h http.Header, req *protocloud.VpcCvmRelWithVpcListReq) (
	[]corecloud.VpcCvmRelWithBaseVpc, error) {

	resp := new(protocloud.VpcCvmRelWithVpcListResp)

	err := cli.client.Post().
		WithContext(ctx).
		Body(req).
		SubResourcef("/vpc_cvm_rels/with/vpc/list").
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
