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
func (cli *LoadBalancerClient) GetListener(kt *kit.Kit, id string) (*dataproto.TCloudListenerDetailResult, error) {
	return common.Request[common.Empty, dataproto.TCloudListenerDetailResult](
		cli.client, rest.GET, kt, nil, "/listeners/%s", id)
}

// ListLoadBalancer list tcloud load balancer
func (cli *LoadBalancerClient) ListLoadBalancer(kt *kit.Kit, req *core.ListReq) (
	*core.ListResultT[corelb.TCloudLoadBalancer], error) {

	return common.Request[core.ListReq, core.ListResultT[corelb.TCloudLoadBalancer]](
		cli.client, rest.POST, kt, req, "/load_balancers/list")

}

// ListUrlRule list url rule.
func (cli *LoadBalancerClient) ListUrlRule(kt *kit.Kit, req *core.ListReq) (*dataproto.TCloudURLRuleListResult, error) {

	return common.Request[core.ListReq, dataproto.TCloudURLRuleListResult](
		cli.client, rest.POST, kt, req, "/load_balancers/url_rules/list")
}

// BatchCreateTCloudTargetGroup 批量创建腾讯云目标组
func (cli *LoadBalancerClient) BatchCreateTCloudTargetGroup(kt *kit.Kit, req *dataproto.TCloudTargetGroupCreateReq) (
	*core.BatchCreateResult, error) {

	return common.Request[dataproto.TCloudTargetGroupCreateReq, core.BatchCreateResult](
		cli.client, rest.POST, kt, req, "/target_groups/batch/create")
}

// BatchCreateTargetGroupWithRel 批量创建目标组 以及对应的 绑定关系
func (cli *LoadBalancerClient) BatchCreateTargetGroupWithRel(kt *kit.Kit,
	req *dataproto.TCloudBatchCreateTgWithRelReq) (*core.BatchCreateResult, error) {

	return common.Request[dataproto.TCloudBatchCreateTgWithRelReq, core.BatchCreateResult](
		cli.client, rest.POST, kt, req, "/target_groups/with/rels/batch/create")
}

// BatchUpdateTCloudTargetGroup 批量更新腾讯云目标组
func (cli *LoadBalancerClient) BatchUpdateTCloudTargetGroup(kt *kit.Kit, req *dataproto.TargetGroupUpdateReq) error {
	return common.RequestNoResp[dataproto.TargetGroupUpdateReq](
		cli.client, rest.PATCH, kt, req, "/target_groups")
}

// GetTargetGroup 获取目标组详情
func (cli *LoadBalancerClient) GetTargetGroup(kt *kit.Kit, id string) (
	*corelb.TargetGroup[corelb.TCloudTargetGroupExtension], error) {

	return common.Request[common.Empty, corelb.TargetGroup[corelb.TCloudTargetGroupExtension]](
		cli.client, rest.GET, kt, nil, "/target_groups/%s", id)
}

// BatchCreateTCloudUrlRule 批量创建腾讯云Url规则
func (cli *LoadBalancerClient) BatchCreateTCloudUrlRule(kt *kit.Kit, req *dataproto.TCloudUrlRuleBatchCreateReq) (
	*core.BatchCreateResult, error) {

	return common.Request[dataproto.TCloudUrlRuleBatchCreateReq, core.BatchCreateResult](
		cli.client, rest.POST, kt, req, "/url_rules/batch/create")
}

// BatchUpdateTCloudUrlRule 批量更新腾讯云Url规则
func (cli *LoadBalancerClient) BatchUpdateTCloudUrlRule(kt *kit.Kit, req *dataproto.TCloudUrlRuleBatchUpdateReq) error {

	return common.RequestNoResp[dataproto.TCloudUrlRuleBatchUpdateReq](
		cli.client, rest.PATCH, kt, req, "/url_rules/batch/update")
}

// BatchDeleteTCloudUrlRule 批量删除腾讯云Url规则
func (cli *LoadBalancerClient) BatchDeleteTCloudUrlRule(kt *kit.Kit, req *dataproto.LoadBalancerBatchDeleteReq) error {

	return common.RequestNoResp[dataproto.LoadBalancerBatchDeleteReq](
		cli.client, rest.DELETE, kt, req, "/url_rules/batch")
}

// BatchCreateTCloudListener 批量创建腾讯云监听器
func (cli *LoadBalancerClient) BatchCreateTCloudListener(kt *kit.Kit, req *dataproto.ListenerBatchCreateReq) (
	*core.BatchCreateResult, error) {

	return common.Request[dataproto.ListenerBatchCreateReq, core.BatchCreateResult](
		cli.client, rest.POST, kt, req, "/listeners/batch/create")
}

// BatchCreateTCloudListenerWithRule 批量创建腾讯云监听器+规则
func (cli *LoadBalancerClient) BatchCreateTCloudListenerWithRule(kt *kit.Kit,
	req *dataproto.ListenerWithRuleBatchCreateReq) (*core.BatchCreateResult, error) {

	return common.Request[dataproto.ListenerWithRuleBatchCreateReq, core.BatchCreateResult](
		cli.client, rest.POST, kt, req, "/listeners/rules/batch/create")
}

// BatchUpdateTCloudListener 批量更新腾讯云监听器
func (cli *LoadBalancerClient) BatchUpdateTCloudListener(kt *kit.Kit, req *dataproto.TCloudListenerUpdateReq) error {

	return common.RequestNoResp[dataproto.TCloudListenerUpdateReq](
		cli.client, rest.PATCH, kt, req, "/listeners/batch/update")
}

// ListListener list listener with tcloud extension.
func (cli *LoadBalancerClient) ListListener(kt *kit.Kit, req *core.ListReq) (
	*dataproto.TCloudListenerListResult, error) {

	return common.Request[core.ListReq, dataproto.TCloudListenerListResult](cli.client,
		rest.POST, kt, req, "/load_balancers/listeners/list")
}
