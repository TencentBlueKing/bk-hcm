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
	coreni "hcm/pkg/api/core/cloud/network-interface"
	datacloudniproto "hcm/pkg/api/data-service/cloud/network-interface"
	"hcm/pkg/criteria/errf"
	"hcm/pkg/rest"
)

// NetworkInterfaceClient is data service network interface api client.
type NetworkInterfaceClient struct {
	client rest.ClientInterface
}

// NewNetworkInterfaceClient create a new network interface api client.
func NewNetworkInterfaceClient(client rest.ClientInterface) *NetworkInterfaceClient {
	return &NetworkInterfaceClient{
		client: client,
	}
}

// BatchCreate batch create huawei network interface.
func (v *NetworkInterfaceClient) BatchCreate(ctx context.Context, h http.Header,
	req *datacloudniproto.NetworkInterfaceBatchCreateReq[datacloudniproto.HuaWeiNICreateExt]) (
	*core.BatchCreateResult, error) {

	resp := new(core.BatchCreateResp)

	err := v.client.Post().
		WithContext(ctx).
		Body(req).
		SubResourcef("/network_interfaces/batch/create").
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

// BatchUpdate huawei network interface.
func (v *NetworkInterfaceClient) BatchUpdate(ctx context.Context, h http.Header,
	req *datacloudniproto.NetworkInterfaceBatchUpdateReq[datacloudniproto.HuaWeiNICreateExt]) error {

	resp := new(rest.BaseResp)

	err := v.client.Patch().
		WithContext(ctx).
		Body(req).
		SubResourcef("/network_interfaces/batch").
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

// Get huawei network interface.
func (v *NetworkInterfaceClient) Get(ctx context.Context, h http.Header, id string) (
	*coreni.NetworkInterface[coreni.HuaWeiNIExtension], error) {

	resp := new(datacloudniproto.NetworkInterfaceGetResp[coreni.HuaWeiNIExtension])

	err := v.client.Get().
		WithContext(ctx).
		SubResourcef("/network_interfaces/%s", id).
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
