/*
 *
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
	"hcm/pkg/api/core"
	"hcm/pkg/api/core/cloud/clb"
	dataproto "hcm/pkg/api/data-service/cloud"
	"hcm/pkg/client/common"
	"hcm/pkg/kit"
	"hcm/pkg/rest"
)

type LoadBalancerClient struct {
	client rest.ClientInterface
}

func NewLoadBalancerClient(client rest.ClientInterface) *LoadBalancerClient {
	return &LoadBalancerClient{client: client}
}

// BatchCreateTCloudClb 批量创建腾讯云CLB
func (cli *LoadBalancerClient) BatchCreateTCloudClb(kt *kit.Kit, req *dataproto.TCloudCLBCreateReq) (
	*core.BatchCreateResult, error) {

	return common.Request[dataproto.TCloudCLBCreateReq, core.BatchCreateResult](
		cli.client, rest.POST, kt, req, "/clbs/batch/create")
}

// Get 获取clb 详情
func (cli *LoadBalancerClient) Get(kt *kit.Kit, id string) (*clb.Clb[clb.TCloudClbExtension], error) {

	return common.Request[common.Empty, clb.Clb[clb.TCloudClbExtension]](
		cli.client, rest.GET, kt, nil, "/clbs/%s", id)
}

// GetListener 获取监听器详情
func (cli *LoadBalancerClient) GetListener(kt *kit.Kit, id string) (*clb.BaseListener, error) {
	return common.Request[common.Empty, clb.BaseListener](cli.client, rest.GET, kt, nil, "/listeners/%s", id)
}
