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

package tcloud

import (
	"context"
	"net/http"

	"hcm/pkg/api/core"
	dataproto "hcm/pkg/api/data-service/cloud/image"
	"hcm/pkg/criteria/errf"
)

// BatchCreateImage 批量创建公共镜像
func (rc *restClient) BatchCreateImage(ctx context.Context,
	h http.Header,
	request *dataproto.ImageExtBatchCreateReq[dataproto.TCloudImageExtensionCreateReq],
) (*core.BatchCreateResult, error) {
	resp := new(core.BatchCreateResp)
	err := rc.client.Post().
		WithContext(ctx).
		Body(request).
		SubResourcef("/images/batch/create").
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

// RetrieveImage 查询单个公共镜像详情
func (rc *restClient) RetrieveImage(
	ctx context.Context,
	h http.Header,
	imageID string,
) (*dataproto.ImageExtResult[dataproto.TCloudImageExtensionResult], error) {
	resp := new(dataproto.ImageExtRetrieveResp[dataproto.TCloudImageExtensionResult])
	err := rc.client.Get().WithContext(ctx).SubResourcef("/images/%s", imageID).WithHeaders(h).Do().Into(resp)
	if err != nil {
		return nil, err
	}

	if resp.Code != errf.OK {
		return nil, errf.New(resp.Code, resp.Message)
	}

	return resp.Data, nil
}

// ListImage 查询公共镜像列表(带 extension 字段)
func (rc *restClient) ListImage(
	ctx context.Context,
	h http.Header,
	request *dataproto.ImageListReq,
) (*dataproto.ImageExtListResult[dataproto.TCloudImageExtensionResult], error) {
	resp := new(dataproto.ImageExtListResp[dataproto.TCloudImageExtensionResult])
	err := rc.client.Post().
		WithContext(ctx).
		Body(request).
		SubResourcef("/images/list").
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

// BatchUpdateImage 更新公共镜像(带 extension 字段)
func (rc *restClient) BatchUpdateImage(
	ctx context.Context,
	h http.Header,
	request *dataproto.ImageExtBatchUpdateReq[dataproto.TCloudImageExtensionUpdateReq],
) (interface{}, error) {
	resp := new(core.UpdateResp)
	err := rc.client.Patch().
		WithContext(ctx).
		Body(request).
		SubResourcef("/images").
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
