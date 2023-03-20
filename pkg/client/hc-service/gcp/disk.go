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

package gcp

import (
	"context"
	"net/http"

	"hcm/pkg/api/core"
	"hcm/pkg/api/hc-service/disk"
	"hcm/pkg/criteria/errf"
	"hcm/pkg/rest"
)

// NewCloudDiskClient create a new disk api client.
func NewCloudDiskClient(client rest.ClientInterface) *DiskClient {
	return &DiskClient{
		client: client,
	}
}

// DiskClient is hc service disk api client.
type DiskClient struct {
	client rest.ClientInterface
}

// SyncDisk sync disk.
func (cli *DiskClient) SyncDisk(ctx context.Context, h http.Header,
	request *disk.DiskSyncReq,
) error {
	resp := new(core.SyncResp)

	err := cli.client.Post().
		WithContext(ctx).
		Body(request).
		SubResourcef("/disks/sync").
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

// AttachDisk ...
func (cli *DiskClient) AttachDisk(ctx context.Context, h http.Header, req *disk.GcpDiskAttachReq) error {
	resp := new(rest.BaseResp)

	err := cli.client.Post().
		WithContext(ctx).
		Body(req).
		SubResourcef("/disks/attach").
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

// DetachDisk ...
func (cli *DiskClient) DetachDisk(ctx context.Context, h http.Header, req *disk.DiskDetachReq) error {
	resp := new(rest.BaseResp)

	err := cli.client.Post().
		WithContext(ctx).
		Body(req).
		SubResourcef("/disks/detach").
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

// DeleteDisk ...
func (cli *DiskClient) DeleteDisk(ctx context.Context, h http.Header, req *disk.DiskDeleteReq) error {
	resp := new(rest.BaseResp)

	err := cli.client.Delete().
		WithContext(ctx).
		Body(req).
		SubResourcef("/disks").
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

// CreateDisk ...
func (cli *DiskClient) CreateDisk(
	ctx context.Context,
	h http.Header,
	req *disk.GcpDiskCreateReq,
) (*core.BatchCreateResult, error) {
	resp := new(core.BatchCreateResp)

	err := cli.client.Post().
		WithContext(ctx).
		Body(req).
		SubResourcef("/disks/create").
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
