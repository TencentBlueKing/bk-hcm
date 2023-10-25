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
	protorg "hcm/pkg/api/data-service/cloud/resource-group"
	"hcm/pkg/criteria/errf"
	"hcm/pkg/rest"
)

// ResourceGroupClient is data service ResourceGroupName api client.
type ResourceGroupClient struct {
	client rest.ClientInterface
}

// NewResourceGroupClient create a new ResourceGroupName api client.
func NewResourceGroupClient(client rest.ClientInterface) *ResourceGroupClient {
	return &ResourceGroupClient{
		client: client,
	}
}

// ListResourceGroup list resourceGroup.
func (cli *ResourceGroupClient) ListResourceGroup(ctx context.Context, h http.Header,
	request *protorg.AzureRGListReq) (*protorg.AzureRGListResult, error) {

	resp := new(protorg.AzureRGListResp)

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

// BatchDeleteResourceGroup delete ResourceGroupName.
func (cli *ResourceGroupClient) BatchDeleteResourceGroup(ctx context.Context, h http.Header, request *protorg.
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

// BatchCreateResourceGroup batch create ResourceGroupName.
func (cli *ResourceGroupClient) BatchCreateResourceGroup(ctx context.Context, h http.Header, request *protorg.
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

// BatchUpdateRG batch create ResourceGroupName.
func (cli *ResourceGroupClient) BatchUpdateRG(ctx context.Context, h http.Header, request *protorg.
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
