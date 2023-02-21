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
	dataproto "hcm/pkg/api/data-service/cloud/disk"
	"hcm/pkg/criteria/errf"
)

// BatchCreateDisk 批量创建云盘
func (rc *restClient) BatchCreateDisk(ctx context.Context,
	h http.Header,
	request *dataproto.DiskExtBatchCreateReq[dataproto.AzureDiskExtensionCreateReq],
) (*core.BatchCreateResult, error) {
	resp := new(core.BatchCreateResp)
	err := rc.client.Post().
		WithContext(ctx).
		Body(request).
		SubResourcef("/disks/batch/create").
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

// RetrieveDisk 查询单个云盘详情
func (rc *restClient) RetrieveDisk(
	ctx context.Context,
	h http.Header,
	diskID string,
) (*dataproto.DiskExtResult[dataproto.AzureDiskExtensionResult], error) {
	resp := new(dataproto.DiskExtRetrieveResp[dataproto.AzureDiskExtensionResult])
	err := rc.client.Get().WithContext(ctx).SubResourcef("/disks/%s", diskID).WithHeaders(h).Do().Into(resp)
	if err != nil {
		return nil, err
	}

	if resp.Code != errf.OK {
		return nil, errf.New(resp.Code, resp.Message)
	}

	return resp.Data, nil
}

// ListDisk 查询云盘列表(带 extension 字段)
func (rc *restClient) ListDisk(
	ctx context.Context,
	h http.Header,
	request *dataproto.DiskListReq,
) (*dataproto.DiskExtListResult[dataproto.AzureDiskExtensionResult], error) {
	resp := new(dataproto.DiskExtListResp[dataproto.AzureDiskExtensionResult])
	err := rc.client.Post().WithContext(ctx).Body(request).SubResourcef("/disks/list").WithHeaders(h).Do().Into(resp)
	if err != nil {
		return nil, err
	}

	if resp.Code != errf.OK {
		return nil, errf.New(resp.Code, resp.Message)
	}

	return resp.Data, nil
}

// BatchUpdateDisk 批量更新云盘信息
func (rc *restClient) BatchUpdateDisk(
	ctx context.Context,
	h http.Header,
	request *dataproto.DiskExtBatchUpadteReq[dataproto.AzureDiskExtensionUpdateReq],
) (interface{}, error) {
	resp := new(core.UpdateResp)
	err := rc.client.Patch().WithContext(ctx).Body(request).SubResourcef("/disks").WithHeaders(h).Do().Into(resp)
	if err != nil {
		return nil, err
	}

	if resp.Code != errf.OK {
		return nil, errf.New(resp.Code, resp.Message)
	}

	return resp.Data, nil
}
