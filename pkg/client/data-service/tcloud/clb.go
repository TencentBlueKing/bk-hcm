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
	corelb "hcm/pkg/api/core/cloud/load-balancer"
	dataproto "hcm/pkg/api/data-service/cloud"
	"hcm/pkg/client/common"
	"hcm/pkg/kit"
	"hcm/pkg/rest"
)

// LoadBalancerClient ...
type LoadBalancerClient struct {
	client rest.ClientInterface
}

// NewLoadBalancerClient ...
func NewLoadBalancerClient(client rest.ClientInterface) *LoadBalancerClient {
	return &LoadBalancerClient{client: client}
}

// BatchCreateTCloudClb 批量创建腾讯云CLB
func (cli *LoadBalancerClient) BatchCreateTCloudClb(kt *kit.Kit, req *dataproto.TCloudCLBCreateReq) (
	*core.BatchCreateResult, error) {

	return common.Request[dataproto.TCloudCLBCreateReq, core.BatchCreateResult](
		cli.client, rest.POST, kt, req, "/load_balancers/batch/create")
}

// Get 获取clb 详情
func (cli *LoadBalancerClient) Get(kt *kit.Kit, id string) (*corelb.LoadBalancer[corelb.TCloudClbExtension],
	error) {

	return common.Request[common.Empty, corelb.LoadBalancer[corelb.TCloudClbExtension]](
		cli.client, rest.GET, kt, nil, "/load_balancers/%s", id)
}

// BatchUpdate 批量更新CLB
func (cli *LoadBalancerClient) BatchUpdate(kt *kit.Kit, req *dataproto.TCloudClbBatchUpdateReq) error {
	return common.RequestNoResp[dataproto.TCloudClbBatchUpdateReq](cli.client,
		rest.PATCH, kt, req, "/load_balancers/batch/update")
}

// GetListener 获取监听器详情
func (cli *LoadBalancerClient) GetListener(kt *kit.Kit, id string) (*corelb.BaseListener, error) {
	return common.Request[common.Empty, corelb.BaseListener](cli.client, rest.GET, kt, nil, "/listeners/%s", id)
}

// ListLoadBalancer list tcloud load balancer
func (cli *LoadBalancerClient) ListLoadBalancer(kt *kit.Kit, req *core.ListReq) (
	*core.ListResultT[corelb.TCloudLoadBalancer], error) {

	return common.Request[core.ListReq, core.ListResultT[corelb.TCloudLoadBalancer]](
		cli.client, rest.POST, kt, req, "/clbs/list")

}
