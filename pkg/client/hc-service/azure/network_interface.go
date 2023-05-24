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

	"hcm/pkg/api/hc-service/sync"
	"hcm/pkg/criteria/errf"
	"hcm/pkg/rest"
)

// NetworkInterfaceClient is hc service huawei network interface api client.
type NetworkInterfaceClient struct {
	client rest.ClientInterface
}

// NewNetworkInterfaceClient create a new network interface api client.
func NewNetworkInterfaceClient(client rest.ClientInterface) *NetworkInterfaceClient {
	return &NetworkInterfaceClient{
		client: client,
	}
}

// SyncNetworkInterface huawei network interface.
func (v *NetworkInterfaceClient) SyncNetworkInterface(ctx context.Context, h http.Header,
	req *sync.AzureSyncReq) error {

	resp := new(rest.BaseResp)

	err := v.client.Post().
		WithContext(ctx).
		Body(req).
		SubResourcef("/network_interfaces/sync").
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
