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
	"hcm/pkg/api/hc-service/region"
	"hcm/pkg/criteria/errf"
	"hcm/pkg/rest"
)

// NewResourceGroupClient create a new rg api client.
func NewResourceGroupClient(client rest.ClientInterface) *ResourceGroupClient {
	return &ResourceGroupClient{
		client: client,
	}
}

// ResourceGroupClient is hc service rg api client.
type ResourceGroupClient struct {
	client rest.ClientInterface
}

// SyncResourceGroup sync zone.
func (cli *ResourceGroupClient) SyncResourceGroup(ctx context.Context, h http.Header,
	request *region.AzureRGSyncReq) error {

	resp := new(core.SyncResp)

	err := cli.client.Post().
		WithContext(ctx).
		Body(request).
		SubResourcef("/resource_groups/sync").
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
