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
// hc-service client 给到 cloud-server 使用
package tcloud

import (
	"net/http"

	typescfs "hcm/pkg/adaptor/types/cfs"
	corecfs "hcm/pkg/api/core/cloud/cfs"
	protocloud "hcm/pkg/api/data-service/cloud"
	protocfs "hcm/pkg/api/hc-service/cfs"
	"hcm/pkg/client/common"
	"hcm/pkg/kit"
	"hcm/pkg/rest"
)

// NewCfsClient create a new cfs api client.
func NewCfsClient(client rest.ClientInterface) *CfsClient {
	return &CfsClient{
		client: client,
	}
}

// CfsClient is hc service cfs api client.
type CfsClient struct {
	client rest.ClientInterface
}

// CreateCfsStorage 创建cfs
func (c *CfsClient) CreateCfsStorage(kt *kit.Kit, req *protocfs.TCloudCreateCfsReq) (
	*corecfs.Cfs[corecfs.TCloudCfsExtension], error) {
	return common.Request[protocfs.TCloudCreateCfsReq, corecfs.Cfs[corecfs.TCloudCfsExtension]](c.client,
		http.MethodPost, kt, req, "/cfs/storage/create")
}

// DeleteCfsStorage 删除cfs
func (c *CfsClient) DeleteCfsStorage(kt *kit.Kit, req *protocfs.TCloudDeleteCfsReq) (
	*typescfs.TCloudDeleteStorageResult, error) {
	return common.Request[protocfs.TCloudDeleteCfsReq, typescfs.TCloudDeleteStorageResult](c.client,
		http.MethodDelete, kt, req, "/cfs/storage/delete")
}

// ListCfsStorage 查询cfs
func (c *CfsClient) ListCfsStorage(kt *kit.Kit, req *protocfs.TCloudListCfsReq,
) (*protocloud.CfsExtListResult[corecfs.TCloudCfsExtension],
	error) {
	return common.Request[protocfs.TCloudListCfsReq, protocloud.CfsExtListResult[corecfs.TCloudCfsExtension]](c.client,
		http.MethodPost, kt, req, "/cfs/storage/list")
}

// GetCfsStorage 查询cfs
func (c *CfsClient) GetCfsStorage(kt *kit.Kit, req *protocfs.TCloudGetCfsReq) (*corecfs.Cfs[corecfs.TCloudCfsExtension],
	error) {
	return common.Request[protocfs.TCloudGetCfsReq, corecfs.Cfs[corecfs.TCloudCfsExtension]](c.client, http.MethodPost,
		kt, req, "/cfs/storage/get")
}
