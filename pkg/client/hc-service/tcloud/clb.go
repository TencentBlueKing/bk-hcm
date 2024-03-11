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

	protoclb "hcm/pkg/api/hc-service/clb"
	"hcm/pkg/client/common"
	"hcm/pkg/kit"
	"hcm/pkg/rest"

	tclb "github.com/tencentcloud/tencentcloud-sdk-go/tencentcloud/clb/v20180317"
)

// NewClbClient create a new clb api client.
func NewClbClient(client rest.ClientInterface) *ClbClient {
	return &ClbClient{
		client: client,
	}
}

// ClbClient is hc service clb api client.
type ClbClient struct {
	client rest.ClientInterface
}

// DescribeResources ...
func (c *ClbClient) DescribeResources(kt *kit.Kit, req *protoclb.TCloudDescribeResourcesOption) (
	*tclb.DescribeResourcesResponseParams, error) {

	return common.Request[protoclb.TCloudDescribeResourcesOption, tclb.DescribeResourcesResponseParams](
		c.client, http.MethodPost, kt, req, "clbs/resources/describe")
}

// BatchCreate ...
func (c *ClbClient) BatchCreate(kt *kit.Kit, req *protoclb.TCloudBatchCreateReq) (*protoclb.BatchCreateResult, error) {
	return common.Request[protoclb.TCloudBatchCreateReq, protoclb.BatchCreateResult](
		c.client, http.MethodPost, kt, req, "clbs/batch/create")
}

// Update ...
func (c *ClbClient) Update(kt *kit.Kit, id string, req *protoclb.TCloudUpdateReq) error {
	return common.RequestNoResp[protoclb.TCloudUpdateReq](c.client, http.MethodPatch, kt, req, "clbs/%s", id)
}
