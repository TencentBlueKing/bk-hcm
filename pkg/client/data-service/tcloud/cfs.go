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

// Package tcloud 如下:
// data-service client
package tcloud

import (
	"context"

	"hcm/pkg/api/core"
	corecfs "hcm/pkg/api/core/cloud/cfs"
	protocloud "hcm/pkg/api/data-service/cloud"
	"hcm/pkg/criteria/errf"
	"hcm/pkg/rest"
	"net/http"
)

// NewCfsClient create a new cfs api client.
func NewCfsClient(client rest.ClientInterface) *CfsClient {
	return &CfsClient{
		client: client,
	}
}

// CfsClient is data service cfs api client.
type CfsClient struct {
	client rest.ClientInterface
}

// CreateCfs create cfs rule.
func (cli *CfsClient) CreateCfs(ctx context.Context, h http.Header,
	request *protocloud.CfsCreateReq[corecfs.TCloudCfsExtension]) (*core.BatchCreateResult, error) {
	resp := new(core.BatchCreateResp)

	err := cli.client.Post().
		WithContext(ctx).
		Body(request).
		SubResourcef("/cfs/create").
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

// BatchDeleteCfs batch delete cfs rule.
func (cli *CfsClient) BatchDeleteCfs(ctx context.Context, h http.Header, request *protocloud.CfsBatchDeleteReq) error {
	resp := new(rest.BaseResp)

	err := cli.client.Delete().
		WithContext(ctx).
		Body(request).
		SubResourcef("/cfs/delete").
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

// ListCfsExt list cfs with extension.
func (cli *CfsClient) ListCfsExt(ctx context.Context, h http.Header, request *protocloud.CfsListReq) (
	*protocloud.CfsExtListResult[corecfs.TCloudCfsExtension], error) {
	resp := new(protocloud.CfsExtListResp[corecfs.TCloudCfsExtension])

	err := cli.client.Post().
		WithContext(ctx).
		Body(request).
		SubResourcef("/cfs/list").
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

// GetCfs get cfs.
func (cli *CfsClient) GetCfs(ctx context.Context, h http.Header, id string) (
	*corecfs.Cfs[corecfs.TCloudCfsExtension], error) {
	resp := new(protocloud.CfsGetResp[corecfs.TCloudCfsExtension])

	err := cli.client.Get().
		WithContext(ctx).
		SubResourcef("/cfs/%s", id).
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
