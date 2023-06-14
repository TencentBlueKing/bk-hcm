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
	dataservice "hcm/pkg/api/data-service"
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

// List network interface.
func (n *NetworkInterfaceClient) List(ctx context.Context, h http.Header, req *core.ListReq) (
	*datacloudniproto.NetworkInterfaceListResult, error) {

	resp := new(datacloudniproto.NetworkInterfaceListResp)

	err := n.client.Post().
		WithContext(ctx).
		Body(req).
		SubResourcef("/network_interfaces/list").
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

// ListAssociate list network interface associate.
func (n *NetworkInterfaceClient) ListAssociate(ctx context.Context, h http.Header,
	req *datacloudniproto.NetworkInterfaceListReq) (*datacloudniproto.NetworkInterfaceAssociateListResult, error) {

	resp := new(datacloudniproto.NetworkInterfaceAssociateListResp)

	err := n.client.Post().
		WithContext(ctx).
		Body(req).
		SubResourcef("/network_interfaces/associate/list").
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

// BatchDelete batch delete network interface.
func (n *NetworkInterfaceClient) BatchDelete(ctx context.Context, h http.Header,
	req *dataservice.BatchDeleteReq) error {

	resp := new(rest.BaseResp)

	err := n.client.Delete().
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

// BatchUpdateNICommonInfo batch update network interface common info.
func (n *NetworkInterfaceClient) BatchUpdateNICommonInfo(ctx context.Context, h http.Header,
	request *datacloudniproto.NetworkInterfaceCommonInfoBatchUpdateReq) error {

	resp := new(rest.BaseResp)

	err := n.client.Patch().
		WithContext(ctx).
		Body(request).
		SubResourcef("/network_interfaces/common/info/batch/update").
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
