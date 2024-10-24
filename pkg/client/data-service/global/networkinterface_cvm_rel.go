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
	proto "hcm/pkg/api/data-service"
	protocloud "hcm/pkg/api/data-service/cloud"
	"hcm/pkg/criteria/errf"
	"hcm/pkg/kit"
	"hcm/pkg/rest"
)

// NewNetworkInterfaceCvmRelClient create a new networkinterface cvm rel api client.
func NewNetworkInterfaceCvmRelClient(client rest.ClientInterface) *NetworkInterfaceCvmRelClient {
	return &NetworkInterfaceCvmRelClient{
		client: client,
	}
}

// NetworkInterfaceCvmRelClient is data service networkinterface cvm rel api client.
type NetworkInterfaceCvmRelClient struct {
	client rest.ClientInterface
}

// BatchCreateNetworkCvmRels create networkinterface cvm rels.
func (cli *NetworkInterfaceCvmRelClient) BatchCreateNetworkCvmRels(ctx context.Context, h http.Header,
	request *protocloud.NetworkInterfaceCvmRelBatchCreateReq) error {

	resp := new(rest.BaseResp)

	err := cli.client.Post().
		WithContext(ctx).
		Body(request).
		SubResourcef("/network_cvm_rels/batch/create").
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

// BatchDeleteNetworkCvmRels delete networkinterface cvm rels.
func (cli *NetworkInterfaceCvmRelClient) BatchDeleteNetworkCvmRels(ctx context.Context, h http.Header,
	request *proto.BatchDeleteReq) error {

	resp := new(rest.BaseResp)

	err := cli.client.Delete().
		WithContext(ctx).
		Body(request).
		SubResourcef("/network_cvm_rels/batch").
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

// ListNetworkCvmRels list networkinterface cvm rels.
func (cli *NetworkInterfaceCvmRelClient) ListNetworkCvmRels(kt *kit.Kit, request *core.ListReq) (
	*protocloud.NetworkInterfaceCvmRelListResult, error) {

	resp := new(protocloud.NetworkInterfaceCvmRelListResp)

	err := cli.client.Post().
		WithContext(kt.Ctx).
		Body(request).
		SubResourcef("/network_cvm_rels/list").
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
