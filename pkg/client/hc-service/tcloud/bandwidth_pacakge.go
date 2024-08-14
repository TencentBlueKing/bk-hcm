/*
 * TencentBlueKing is pleased to support the open source community by making
 * 蓝鲸智云 - 混合云管理平台 (BlueKing - Hybrid Cloud Management System) available.
 * Copyright (C) 2024 THL A29 Limited,
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
	"net/http"

	"hcm/pkg/adaptor/types"
	hcbwpkg "hcm/pkg/api/hc-service/bandwidth-packages"
	"hcm/pkg/client/common"
	"hcm/pkg/kit"
	"hcm/pkg/rest"
)

// NewBandPkgClient create a new bandwidth package api client.
func NewBandPkgClient(client rest.ClientInterface) *BandwidthPackageClient {
	return &BandwidthPackageClient{
		client: client,
	}
}

// BandwidthPackageClient is hc service bandwidth package  api client.
type BandwidthPackageClient struct {
	client rest.ClientInterface
}

// ListBandwidthPackage 查询带宽包
func (c *BandwidthPackageClient) ListBandwidthPackage(kt *kit.Kit, req *hcbwpkg.ListTCloudBwPkgOption) (
	*types.TCloudListBwPkgResult, error) {

	return common.Request[hcbwpkg.ListTCloudBwPkgOption, types.TCloudListBwPkgResult](c.client, http.MethodPost, kt,
		req, "/bandwidth_packages/list")
}
