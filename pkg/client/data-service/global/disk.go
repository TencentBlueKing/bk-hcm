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
	dataproto "hcm/pkg/api/data-service/cloud/disk"
	"hcm/pkg/criteria/errf"
)

// ListDisk 查询云盘列表
func (rc *restClient) ListDisk(ctx context.Context,
	h http.Header,
	request *dataproto.DiskListReq,
) (*dataproto.DiskListResult, error) {
	resp := new(dataproto.DiskListResp)
	err := rc.client.Post().WithContext(ctx).Body(request).SubResourcef("/disks/list").WithHeaders(h).Do().Into(resp)
	if err != nil {
		return nil, err
	}

	if resp.Code != errf.OK {
		return nil, errf.New(resp.Code, resp.Message)
	}

	return resp.Data, nil
}

// DeleteDisk 删除云盘记录
func (rc *restClient) DeleteDisk(
	ctx context.Context,
	h http.Header,
	request *dataproto.DiskDeleteReq,
) (interface{}, error) {
	resp := new(core.DeleteResp)
	err := rc.client.Delete().
		WithContext(ctx).
		Body(request).
		SubResourcef("/disks/batch").
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

// BatchUpdateDisk 批量更新云盘的基础信息
func (rc *restClient) BatchUpdateDisk(ctx context.Context,
	h http.Header,
	request *dataproto.DiskBatchUpdateReq,
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
