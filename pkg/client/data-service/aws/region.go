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

package aws

import (
	"context"
	"net/http"

	"hcm/pkg/api/core"
	dataservice "hcm/pkg/api/data-service"
	protoregion "hcm/pkg/api/data-service/cloud/region"
	"hcm/pkg/criteria/errf"
	"hcm/pkg/rest"
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

// BatchCreate batch create aws region.
func (v *RegionClient) BatchCreate(ctx context.Context, h http.Header,
	req *protoregion.AwsRegionCreateReq) (*core.BatchCreateResult, error) {

	resp := new(core.BatchCreateResp)

	err := v.client.Post().
		WithContext(ctx).
		Body(req).
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

// BatchUpdate batch update aws region.
func (v *RegionClient) BatchUpdate(ctx context.Context, h http.Header, req *protoregion.AwsRegionBatchUpdateReq) error {

	resp := new(rest.BaseResp)

	err := v.client.Patch().
		WithContext(ctx).
		Body(req).
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

// BatchForbiddenRegionState batch forbidden aws region state.
func (v *RegionClient) BatchForbiddenRegionState(ctx context.Context, h http.Header,
	req *protoregion.AwsRegionBatchUpdateReq) error {

	resp := new(rest.BaseResp)

	err := v.client.Patch().
		WithContext(ctx).
		Body(req).
		SubResourcef("/regions/batch/state").
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

// BatchDelete batch delete aws region.
func (v *RegionClient) BatchDelete(ctx context.Context, h http.Header, req *dataservice.BatchDeleteReq) error {

	resp := new(rest.BaseResp)

	err := v.client.Delete().
		WithContext(ctx).
		Body(req).
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

// ListRegion get aws region list.
func (v *RegionClient) ListRegion(ctx context.Context, h http.Header, req *core.ListReq) (
	*protoregion.AwsRegionListResult, error) {

	resp := new(protoregion.AwsRegionListResp)

	err := v.client.Post().
		WithContext(ctx).
		Body(req).
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
