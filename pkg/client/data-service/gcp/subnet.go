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

package gcp

import (
	"context"
	"net/http"

	"hcm/pkg/api/core"
	corecloud "hcm/pkg/api/core/cloud"
	protocloud "hcm/pkg/api/data-service/cloud"
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

// BatchCreate batch create gcp subnet.
func (v *SubnetClient) BatchCreate(ctx context.Context, h http.Header,
<<<<<<< HEAD
	req *protocloud.SubnetBatchCreateReq[corecloud.GcpSubnetExtension]) (*core.BatchCreateResult, error) {
=======
	req *protocloud.SubnetBatchCreateReq[protocloud.GcpSubnetCreateExt]) (*core.BatchCreateResult, error) {
>>>>>>> 304144ec282c951c6c2127f39ca83cb7f1c70b41

	resp := new(core.BatchCreateResp)

	err := v.client.Post().
		WithContext(ctx).
		Body(req).
		SubResourcef("/subnets/batch/create").
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

// Get gcp subnet.
func (v *SubnetClient) Get(ctx context.Context, h http.Header, id string) (*corecloud.Subnet[corecloud.GcpSubnetExtension],
	error) {

	resp := new(protocloud.SubnetGetResp[corecloud.GcpSubnetExtension])

	err := v.client.Get().
		WithContext(ctx).
		SubResourcef("/subnets/%s", id).
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

// BatchUpdate gcp subnet.
func (v *SubnetClient) BatchUpdate(ctx context.Context, h http.Header,
	req *protocloud.SubnetBatchUpdateReq[protocloud.GcpSubnetUpdateExt]) error {

	resp := new(rest.BaseResp)

	err := v.client.Patch().
		WithContext(ctx).
		Body(req).
		SubResourcef("/subnets/batch").
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
