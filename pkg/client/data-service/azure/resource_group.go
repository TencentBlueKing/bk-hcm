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
	"hcm/pkg/api/core"
	protoregion "hcm/pkg/api/data-service/cloud/region"
	"hcm/pkg/criteria/errf"
	"hcm/pkg/rest"
	"net/http"
)

// ResourceGroupClient is data service ResourceGroup api client.
type ResourceGroupClient struct {
	client rest.ClientInterface
}

// ResourceGroupClient create a new ResourceGroup api client.
func NewResourceGroupClient(client rest.ClientInterface) *ResourceGroupClient {
	return &ResourceGroupClient{
		client: client,
	}
}

// ListResourceGroup list resourceGroup.
func (cli *ResourceGroupClient) ListResourceGroup(ctx context.Context, h http.Header,
	request *protoregion.AzureRGListReq) (*protoregion.AzureRGListResult, error) {

	resp := new(protoregion.AzureRGListResp)

	err := cli.client.Post().
		WithContext(ctx).
		Body(request).
		SubResourcef("/resource_groups/list").
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

// BatchDeleteResourceGroup delete ResourceGroup.
func (cli *ResourceGroupClient) BatchDeleteResourceGroup(ctx context.Context, h http.Header, request *protoregion.
	AzureRGBatchDeleteReq) error {

	resp := new(core.DeleteResp)

	err := cli.client.Delete().
		WithContext(ctx).
		Body(request).
		SubResourcef("/resource_groups/batch").
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

// BatchCreateResourceGroup batch create ResourceGroup.
func (cli *ResourceGroupClient) BatchCreateResourceGroup(ctx context.Context, h http.Header, request *protoregion.
	AzureRGBatchCreateReq) (*core.BatchCreateResult, error) {

	resp := new(core.BatchCreateResp)

	err := cli.client.Post().
		WithContext(ctx).
		Body(request).
		SubResourcef("/resource_groups/batch/create").
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

// BatchUpdateResourceGroup batch create ResourceGroup.
func (cli *ResourceGroupClient) BatchUpdateRG(ctx context.Context, h http.Header, request *protoregion.
	AzureRGBatchUpdateReq) error {

	resp := new(core.UpdateResp)

	err := cli.client.Put().
		WithContext(ctx).
		Body(request).
		SubResourcef("/resource_groups/batch/update").
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
