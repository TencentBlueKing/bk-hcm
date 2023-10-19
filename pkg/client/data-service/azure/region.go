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

// RegionClient is data service region api client.
type RegionClient struct {
	client rest.ClientInterface
}

// NewRegionClient create a new region api client.
func NewRegionClient(client rest.ClientInterface) *RegionClient {
	return &RegionClient{
		client: client,
	}
}

// ListRegion list region.
func (cli *RegionClient) ListRegion(ctx context.Context, h http.Header,
	request *protoregion.AzureRegionListReq) (*protoregion.AzureRegionListResult, error) {

	resp := new(protoregion.AzureRegionListResp)

	err := cli.client.Post().
		WithContext(ctx).
		Body(request).
		SubResourcef("/regions/list").
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

// BatchDeleteRegion delete region.
func (cli *RegionClient) BatchDeleteRegion(ctx context.Context, h http.Header, request *protoregion.
	AzureRegionBatchDeleteReq) error {

	resp := new(core.DeleteResp)

	err := cli.client.Delete().
		WithContext(ctx).
		Body(request).
		SubResourcef("/regions/batch").
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

// BatchCreateRegion batch create region.
func (cli *RegionClient) BatchCreateRegion(ctx context.Context, h http.Header, request *protoregion.
	AzureRegionBatchCreateReq) (*core.BatchCreateResult, error) {

	resp := new(core.BatchCreateResp)

	err := cli.client.Post().
		WithContext(ctx).
		Body(request).
		SubResourcef("/regions/batch/create").
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

// BatchUpdateRegion batch update region.
func (cli *RegionClient) BatchUpdateRegion(ctx context.Context, h http.Header, request *protoregion.
	AzureRegionBatchUpdateReq) error {

	resp := new(core.UpdateResp)

	err := cli.client.Put().
		WithContext(ctx).
		Body(request).
		SubResourcef("/regions/batch/update").
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
