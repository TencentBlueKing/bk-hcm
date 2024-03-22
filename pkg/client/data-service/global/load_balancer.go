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
	"hcm/pkg/api/core"
	dataproto "hcm/pkg/api/data-service/cloud"
	"hcm/pkg/client/common"
	"hcm/pkg/kit"
	"hcm/pkg/rest"
)

// LoadBalancerClient is data service load balancer api client.
type LoadBalancerClient struct {
	client rest.ClientInterface
}

// NewLoadBalancerClient create a new load balancer api client.
func NewLoadBalancerClient(client rest.ClientInterface) *LoadBalancerClient {
	return &LoadBalancerClient{
		client: client,
	}
}

// ListLoadBalancer list load balancer.
func (cli *LoadBalancerClient) ListLoadBalancer(kt *kit.Kit, req *core.ListReq) (*dataproto.LbListResult, error) {
	return common.Request[core.ListReq, dataproto.LbListResult](cli.client, rest.POST, kt, req, "/load_balancers/list")
}

// BatchUpdateClbBizInfo update biz
func (cli *LoadBalancerClient) BatchUpdateClbBizInfo(kt *kit.Kit, req *dataproto.ClbBizBatchUpdateReq) error {
	return common.RequestNoResp[dataproto.ClbBizBatchUpdateReq](cli.client, rest.PATCH,
		kt, req, "/load_balancers/biz/batch/update")

}

// ListListener list listener.
func (cli *LoadBalancerClient) ListListener(kt *kit.Kit, req *core.ListReq) (
	*dataproto.ListenerListResult, error) {

	return common.Request[core.ListReq, dataproto.ListenerListResult](cli.client,
		rest.POST, kt, req, "/load_balancers/listeners/list")
}

// ListTarget list target.
func (cli *LoadBalancerClient) ListTarget(kt *kit.Kit, req *core.ListReq) (*dataproto.TargetListResult, error) {
	return common.Request[core.ListReq, dataproto.TargetListResult](
		cli.client, rest.POST, kt, req, "/load_balancers/targets/list")
}

// ListTargetGroup list target group.
func (cli *LoadBalancerClient) ListTargetGroup(kt *kit.Kit, req *core.ListReq) (*dataproto.TargetGroupListResult,
	error) {
	return common.Request[core.ListReq, dataproto.TargetGroupListResult](
		cli.client, rest.POST, kt, req, "/load_balancers/target_groups/list")
}

// BatchDelete 批量删除
func (cli *LoadBalancerClient) BatchDelete(kt *kit.Kit, req *dataproto.LoadBalancerBatchDeleteReq) error {
	return common.RequestNoResp[dataproto.LoadBalancerBatchDeleteReq](cli.client, rest.DELETE,
		kt, req, "/load_balancers/batch")
}

// ListTargetGroupListenerRel list target group listener rel.
func (cli *LoadBalancerClient) ListTargetGroupListenerRel(kt *kit.Kit, req *core.ListReq) (
	*dataproto.TargetListenerRuleRelListResult, error) {

	return common.Request[core.ListReq, dataproto.TargetListenerRuleRelListResult](
		cli.client, rest.POST, kt, req, "/target_group_listener_rels/list")
}

// DeleteTargetGroup delete target group.
func (cli *LoadBalancerClient) DeleteTargetGroup(kt *kit.Kit, req *core.ListReq) error {
	return common.RequestNoResp[core.ListReq](cli.client, rest.DELETE, kt, req, "/target_groups/batch")
}

// CreateTargetGroupListenerRel create target group listener rel.
func (cli *LoadBalancerClient) CreateTargetGroupListenerRel(kt *kit.Kit,
	req *dataproto.TargetGroupListenerRelCreateReq) (*core.BatchCreateResult, error) {

	return common.Request[dataproto.TargetGroupListenerRelCreateReq, core.BatchCreateResult](
		cli.client, rest.POST, kt, req, "/target_group_listener_rels/create")
}

// DeleteListener delete listener.
func (cli *LoadBalancerClient) DeleteListener(kt *kit.Kit, req *core.ListReq) error {
	return common.RequestNoResp[core.ListReq](cli.client, rest.DELETE, kt, req, "/listeners/batch")
}
