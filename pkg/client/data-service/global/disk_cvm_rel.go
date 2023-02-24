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

	dataproto "hcm/pkg/api/data-service/cloud"
	"hcm/pkg/criteria/errf"
	"hcm/pkg/rest"
)

// BatchCreateDiskCvmRel ...
func (rc *restClient) BatchCreateDiskCvmRel(ctx context.Context,
	h http.Header,
	request *dataproto.DiskCvmRelBatchCreateReq,
) error {
	resp := new(rest.BaseResp)

	err := rc.client.Post().
		WithContext(ctx).
		Body(request).
		SubResourcef("/disk_cvm_rels/batch/create").
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

// ListDiskCvmRel ...
func (rc *restClient) ListDiskCvmRel(ctx context.Context,
	h http.Header, request *dataproto.DiskCvmRelListReq,
) (*dataproto.DiskCvmRelListResult, error) {
	resp := new(dataproto.DiskCvmRelListResp)

	err := rc.client.Post().
		WithContext(ctx).
		Body(request).
		SubResourcef("/disk_cvm_rels/list").
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

// ListDiskCvmRelWithDisk ...
func (rc *restClient) ListDiskCvmRelWithDisk(ctx context.Context,
	h http.Header,
	request *dataproto.DiskCvmRelWithDiskListReq,
) ([]*dataproto.DiskWithCvmID, error) {
	resp := new(dataproto.DiskCvmRelWithDiskListResp)

	err := rc.client.Post().
		WithContext(ctx).
		Body(request).
		SubResourcef("/disk_cvm_rels/with/disks/list").
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

// DeleteDiskCvmRel ...
func (rc *restClient) DeleteDiskCvmRel(
	ctx context.Context,
	h http.Header,
	request *dataproto.DiskCvmRelDeleteReq,
) error {
	resp := new(rest.BaseResp)

	err := rc.client.Delete().
		WithContext(ctx).
		Body(request).
		SubResourcef("/disk_cvm_rels/batch").
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
