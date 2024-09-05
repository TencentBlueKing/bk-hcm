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

package global

import (
	"hcm/pkg/api/data-service/cos"
	"hcm/pkg/client/common"
	"hcm/pkg/kit"
	"hcm/pkg/rest"
)

// CosClient is data service cos api client.
type CosClient struct {
	client rest.ClientInterface
}

// NewCosClient create a new cos api client.
func NewCosClient(client rest.ClientInterface) *CosClient {
	return &CosClient{
		client: client,
	}
}

// Upload ...
func (a *CosClient) Upload(kt *kit.Kit, req *cos.UploadFileReq) error {
	return common.RequestNoResp[cos.UploadFileReq](
		a.client, rest.POST, kt, req, "/cos/upload")
}

// GenerateTemporalUrl ...
func (a *CosClient) GenerateTemporalUrl(kt *kit.Kit, action string, req *cos.GenerateTemporalUrlReq) (
	*cos.GenerateTemporalUrlResult, error) {

	return common.Request[cos.GenerateTemporalUrlReq, cos.GenerateTemporalUrlResult](
		a.client, rest.POST, kt, req, "/cos/temporal_urls/%s/generate", action)
}
