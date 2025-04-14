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
	"net/http"

	typescos "hcm/pkg/adaptor/types/cos"
	protocos "hcm/pkg/api/hc-service/cos"
	"hcm/pkg/client/common"
	"hcm/pkg/kit"
	"hcm/pkg/rest"
)

// NewCosClient create a new cos api client.
func NewCosClient(client rest.ClientInterface) *CosClient {
	return &CosClient{
		client: client,
	}
}

// CosClient is hc service cos api client.
type CosClient struct {
	client rest.ClientInterface
}

// CreateCosBucket ....
func (c *CosClient) CreateCosBucket(kt *kit.Kit, req *protocos.TCloudCreateBucketReq) error {
	return common.RequestNoResp[protocos.TCloudCreateBucketReq](c.client, http.MethodPost, kt, req,
		"/cos/buckets/create")
}

// DeleteCosBucket ....
func (c *CosClient) DeleteCosBucket(kt *kit.Kit, req *protocos.TCloudDeleteBucketReq) error {
	return common.RequestNoResp[protocos.TCloudDeleteBucketReq](c.client, http.MethodDelete, kt, req,
		"/cos/buckets/delete")
}

// ListCosBucket ...
func (c *CosClient) ListCosBucket(kt *kit.Kit, req *protocos.TCloudBucketListReq) (
	*typescos.TCloudBucketListResult, error) {

	return common.Request[protocos.TCloudBucketListReq, typescos.TCloudBucketListResult](
		c.client, http.MethodPost, kt, req, "/cos/buckets/list")
}
