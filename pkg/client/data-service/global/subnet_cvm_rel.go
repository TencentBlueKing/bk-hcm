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

// NewSubnetCvmRelClient create a new subnet & cvm relation api client.
func NewSubnetCvmRelClient(client rest.ClientInterface) *SubnetCvmRelClient {
	return &SubnetCvmRelClient{
		client: client,
	}
}

// SubnetCvmRelClient is data service subnet cvm rel api client.
type SubnetCvmRelClient struct {
	client rest.ClientInterface
}

// BatchCreate subnet cvm relations.
func (cli *SubnetCvmRelClient) BatchCreate(ctx context.Context, h http.Header,
	request *protocloud.SubnetCvmRelBatchCreateReq) error {

	resp := new(rest.BaseResp)

	err := cli.client.Post().
		WithContext(ctx).
		Body(request).
		SubResourcef("/subnet_cvm_rels/batch/create").
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

// BatchDelete subnet cvm relations.
func (cli *SubnetCvmRelClient) BatchDelete(ctx context.Context, h http.Header, request *proto.BatchDeleteReq) error {

	resp := new(rest.BaseResp)

	err := cli.client.Delete().
		WithContext(ctx).
		Body(request).
		SubResourcef("/subnet_cvm_rels/batch").
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

// List subnet cvm relations.
func (cli *SubnetCvmRelClient) List(ctx context.Context, h http.Header, request *core.ListReq) (
	*protocloud.SubnetCvmRelListResult, error) {

	resp := new(protocloud.SubnetCvmRelListResp)

	err := cli.client.Post().
		WithContext(ctx).
		Body(request).
		SubResourcef("/subnet_cvm_rels/list").
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

// ListWithSubnet list subnet cvm relations with subnet details.
func (cli *SubnetCvmRelClient) ListWithSubnet(ctx context.Context, h http.Header,
	req *protocloud.SubnetCvmRelWithSubnetListReq) ([]corecloud.SubnetCvmRelWithBaseSubnet, error) {

	resp := new(protocloud.SubnetCvmRelWithSubnetListResp)

	err := cli.client.Post().
		WithContext(ctx).
		Body(req).
		SubResourcef("/subnet_cvm_rels/with/subnet/list").
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
