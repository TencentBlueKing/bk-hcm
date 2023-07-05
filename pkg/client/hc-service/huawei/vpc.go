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

package huawei

import (
	"context"
	"net/http"

	"hcm/pkg/api/core"
	"hcm/pkg/api/hc-service/sync"
	hcservice "hcm/pkg/api/hc-service/vpc"
	"hcm/pkg/criteria/errf"
	"hcm/pkg/rest"
)

// VpcClient is hc service huawei vpc api client.
type VpcClient struct {
	client rest.ClientInterface
}

// NewVpcClient create a new vpc api client.
func NewVpcClient(client rest.ClientInterface) *VpcClient {
	return &VpcClient{
		client: client,
	}
}

// Create vpc.
func (v *VpcClient) Create(ctx context.Context, h http.Header,
	req *hcservice.VpcCreateReq[hcservice.HuaWeiVpcCreateExt]) (*core.CreateResult, error) {

	resp := new(core.CreateResp)

	err := v.client.Post().
		WithContext(ctx).
		Body(req).
		SubResourcef("/vpcs/create").
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

// Update vpc.
func (v *VpcClient) Update(ctx context.Context, h http.Header, id string, req *hcservice.VpcUpdateReq) error {
	resp := new(rest.BaseResp)

	err := v.client.Patch().
		WithContext(ctx).
		Body(req).
		SubResourcef("/vpcs/%s", id).
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

// Delete vpc.
func (v *VpcClient) Delete(ctx context.Context, h http.Header, id string) error {
	resp := new(rest.BaseResp)

	err := v.client.Delete().
		WithContext(ctx).
		Body(nil).
		SubResourcef("/vpcs/%s", id).
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

// SyncVpc sync huawei vpc.
func (v *VpcClient) SyncVpc(ctx context.Context, h http.Header, req *sync.HuaWeiSyncReq) error {
	resp := new(rest.BaseResp)

	err := v.client.Post().
		WithContext(ctx).
		Body(req).
		SubResourcef("/vpcs/sync").
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
