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

	typelb "hcm/pkg/adaptor/types/load-balancer"
	"hcm/pkg/api/core"
	hcproto "hcm/pkg/api/hc-service/load-balancer"
	"hcm/pkg/api/hc-service/sync"
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

// SyncLoadBalancer 同步负载均衡
func (c *ClbClient) SyncLoadBalancer(kt *kit.Kit, req *sync.TCloudSyncReq) error {

	return common.RequestNoResp[sync.TCloudSyncReq](c.client, http.MethodPost, kt, req, "/load_balancers/sync")
}

// DescribeResources ...
func (c *ClbClient) DescribeResources(kt *kit.Kit, req *hcproto.TCloudDescribeResourcesOption) (
	*tclb.DescribeResourcesResponseParams, error) {

	return common.Request[hcproto.TCloudDescribeResourcesOption, tclb.DescribeResourcesResponseParams](
		c.client, http.MethodPost, kt, req, "/load_balancers/resources/describe")
}

// BatchCreate ...
func (c *ClbClient) BatchCreate(kt *kit.Kit, req *hcproto.TCloudLoadBalancerCreateReq) (*hcproto.BatchCreateResult,
	error) {
	return common.Request[hcproto.TCloudLoadBalancerCreateReq, hcproto.BatchCreateResult](
		c.client, http.MethodPost, kt, req, "/load_balancers/batch/create")
}

// Update ...
func (c *ClbClient) Update(kt *kit.Kit, id string, req *hcproto.TCloudLBUpdateReq) error {
	return common.RequestNoResp[hcproto.TCloudLBUpdateReq](c.client, http.MethodPatch,
		kt, req, "/load_balancers/%s", id)
}

// CreateListener 创建监听器
func (c *ClbClient) CreateListener(kt *kit.Kit, req *hcproto.ListenerWithRuleCreateReq) (
	*hcproto.ListenerWithRuleCreateResult, error) {

	return common.Request[hcproto.ListenerWithRuleCreateReq, hcproto.ListenerWithRuleCreateResult](
		c.client, http.MethodPost, kt, req, "/listeners/create")
}

// UpdateListener 更新监听器
func (c *ClbClient) UpdateListener(kt *kit.Kit, id string, req *hcproto.ListenerWithRuleUpdateReq) (
	*hcproto.BatchCreateResult, error) {

	return common.Request[hcproto.ListenerWithRuleUpdateReq, hcproto.BatchCreateResult](
		c.client, http.MethodPatch, kt, req, "/listeners/%s", id)
}

// DeleteListener 删除监听器
func (c *ClbClient) DeleteListener(kt *kit.Kit, req *core.BatchDeleteReq) error {
	return common.RequestNoResp[core.BatchDeleteReq](c.client, http.MethodDelete, kt, req, "/listeners/batch")
}

// BatchCreateUrlRule 批量创建规则
func (c *ClbClient) BatchCreateUrlRule(kt *kit.Kit, lblID string, req *hcproto.TCloudRuleBatchCreateReq) (
	*hcproto.BatchCreateResult, error) {

	return common.Request[hcproto.TCloudRuleBatchCreateReq, hcproto.BatchCreateResult](
		c.client, http.MethodPost, kt, req, "/listeners/%s/rules/batch/create", lblID)
}

// UpdateUrlRule 更新规则
func (c *ClbClient) UpdateUrlRule(kt *kit.Kit, lblID, ruleID string, req *hcproto.TCloudRuleUpdateReq) error {
	return common.RequestNoResp[hcproto.TCloudRuleUpdateReq](c.client, http.MethodPatch, kt, req,
		"/listeners/%s/rules/%s", lblID, ruleID)
}

// BatchDeleteUrlRule 批量删除规则
func (c *ClbClient) BatchDeleteUrlRule(kt *kit.Kit, lblID string, req *hcproto.TCloudRuleDeleteByIDReq) error {

	return common.RequestNoResp[hcproto.TCloudRuleDeleteByIDReq](c.client,
		http.MethodDelete, kt, req, "/listeners/%s/rules/batch", lblID)
}

// BatchDeleteUrlRuleByDomain 批量删除规则
func (c *ClbClient) BatchDeleteUrlRuleByDomain(kt *kit.Kit, lblID string,
	req *hcproto.TCloudRuleDeleteByDomainReq) error {

	return common.RequestNoResp[hcproto.TCloudRuleDeleteByDomainReq](c.client,
		http.MethodDelete, kt, req, "/listeners/%s/rules/by/domain/batch", lblID)
}

// UpdateDomainAttr 更新域名属性
func (c *ClbClient) UpdateDomainAttr(kt *kit.Kit, id string, req *hcproto.DomainAttrUpdateReq) error {
	return common.RequestNoResp[hcproto.DomainAttrUpdateReq](
		c.client, http.MethodPatch, kt, req, "/listeners/%s/domains", id)
}

// BatchAddRs 批量添加RS
func (c *ClbClient) BatchAddRs(kt *kit.Kit, targetGroupID string, req *hcproto.TCloudBatchOperateTargetReq) (
	*hcproto.BatchCreateResult, error) {

	return common.Request[hcproto.TCloudBatchOperateTargetReq, hcproto.BatchCreateResult](
		c.client, http.MethodPost, kt, req, "/target_groups/%s/targets/create", targetGroupID)
}

// BatchRemoveTarget 批量移除RS
func (c *ClbClient) BatchRemoveTarget(kt *kit.Kit, targetGroupID string, req *hcproto.TCloudBatchOperateTargetReq) (
	*hcproto.BatchCreateResult, error) {

	return common.Request[hcproto.TCloudBatchOperateTargetReq, hcproto.BatchCreateResult](
		c.client, http.MethodDelete, kt, req, "/target_groups/%s/targets/batch", targetGroupID)
}

// BatchModifyTargetPort 批量修改RS端口
func (c *ClbClient) BatchModifyTargetPort(kt *kit.Kit, targetGroupID string,
	req *hcproto.TCloudBatchOperateTargetReq) error {

	return common.RequestNoResp[hcproto.TCloudBatchOperateTargetReq](
		c.client, http.MethodPatch, kt, req, "/target_groups/%s/targets/port", targetGroupID)
}

// BatchModifyTargetWeight 批量修改RS权重
func (c *ClbClient) BatchModifyTargetWeight(kt *kit.Kit, targetGroupID string,
	req *hcproto.TCloudBatchOperateTargetReq) error {

	return common.RequestNoResp[hcproto.TCloudBatchOperateTargetReq](
		c.client, http.MethodPatch, kt, req, "/target_groups/%s/targets/weight", targetGroupID)
}

// BatchRegisterTargetToListenerRule 注册rs到监听器、规则上
func (c *ClbClient) BatchRegisterTargetToListenerRule(kt *kit.Kit, lbID string,
	req *hcproto.BatchRegisterTCloudTargetReq) error {

	return common.RequestNoResp[hcproto.BatchRegisterTCloudTargetReq](
		c.client, http.MethodPost, kt, req, "/load_balancers/%s/targets/create", lbID)
}

// BatchDeleteLoadBalancer 批量删除云负载均衡
func (c *ClbClient) BatchDeleteLoadBalancer(kt *kit.Kit, req *hcproto.TCloudBatchDeleteLoadbalancerReq) error {

	return common.RequestNoResp[hcproto.TCloudBatchDeleteLoadbalancerReq](
		c.client, http.MethodDelete, kt, req, "/load_balancers/batch")
}

// UpdateListenerHealthCheck 更新健康检查
func (c *ClbClient) UpdateListenerHealthCheck(kt *kit.Kit, lblID string,
	req *hcproto.HealthCheckUpdateReq) error {

	return common.RequestNoResp[hcproto.HealthCheckUpdateReq](c.client, http.MethodPatch, kt, req,
		"/listeners/%s/health_check", lblID)
}

// ListTargetHealth 查询目标组所在负载均衡的端口健康数据
func (c *ClbClient) ListTargetHealth(kt *kit.Kit, req *hcproto.TCloudTargetHealthReq) (
	*hcproto.TCloudTargetHealthResp, error) {

	return common.Request[hcproto.TCloudTargetHealthReq, hcproto.TCloudTargetHealthResp](
		c.client, http.MethodPost, kt, req, "/load_balancers/targets/health")
}

// QueryListenerTargetsByCloudIDs 查询监听器下的rs
func (c *ClbClient) QueryListenerTargetsByCloudIDs(kt *kit.Kit, req *hcproto.QueryTCloudListenerTargets) (
	*[]typelb.TCloudListenerTarget, error) {

	return common.Request[hcproto.QueryTCloudListenerTargets, []typelb.TCloudListenerTarget](
		c.client, http.MethodPost, kt, req, "/targets/query_by_cloud_ids")
}

// InquiryPrice 负载均衡购买询价
func (c *ClbClient) InquiryPrice(kt *kit.Kit, req *hcproto.TCloudLoadBalancerCreateReq) (*typelb.TCloudLBPrice, error) {
	return common.Request[hcproto.TCloudLoadBalancerCreateReq, typelb.TCloudLBPrice](
		c.client, http.MethodPost, kt, req, "/load_balancers/prices/inquiry")
}

// ListQuota 负载均衡配额列表
func (c *ClbClient) ListQuota(kt *kit.Kit, req *hcproto.TCloudListLoadBalancerQuotaReq) (
	*[]typelb.TCloudLoadBalancerQuota, error) {

	return common.Request[hcproto.TCloudListLoadBalancerQuotaReq, []typelb.TCloudLoadBalancerQuota](
		c.client, http.MethodPost, kt, req, "/load_balancers/quota")
}

// CreateSnatIp ...
func (c *ClbClient) CreateSnatIp(kt *kit.Kit, req *hcproto.TCloudCreateSnatIpReq) error {
	return common.RequestNoResp[hcproto.TCloudCreateSnatIpReq](c.client, http.MethodPost, kt, req,
		"/load_balancers/snat_ips/create")
}

// DeleteSnatIp ...
func (c *ClbClient) DeleteSnatIp(kt *kit.Kit, req *hcproto.TCloudDeleteSnatIpReq) error {
	return common.RequestNoResp[hcproto.TCloudDeleteSnatIpReq](c.client, http.MethodDelete, kt, req,
		"/load_balancers/snat_ips")
}
