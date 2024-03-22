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

	"hcm/pkg/api/core"
	protolb "hcm/pkg/api/hc-service/load-balancer"
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
func (c *ClbClient) DescribeResources(kt *kit.Kit, req *protolb.TCloudDescribeResourcesOption) (
	*tclb.DescribeResourcesResponseParams, error) {

	return common.Request[protolb.TCloudDescribeResourcesOption, tclb.DescribeResourcesResponseParams](
		c.client, http.MethodPost, kt, req, "/load_balancers/resources/describe")
}

// BatchCreate ...
func (c *ClbClient) BatchCreate(kt *kit.Kit, req *protolb.TCloudBatchCreateReq) (*protolb.BatchCreateResult, error) {
	return common.Request[protolb.TCloudBatchCreateReq, protolb.BatchCreateResult](
		c.client, http.MethodPost, kt, req, "/load_balancers/batch/create")
}

// Update ...
func (c *ClbClient) Update(kt *kit.Kit, id string, req *protolb.TCloudLBUpdateReq) error {
	return common.RequestNoResp[protolb.TCloudLBUpdateReq](c.client, http.MethodPatch,
		kt, req, "/load_balancers/%s", id)
}

// CreateListener 创建监听器
func (c *ClbClient) CreateListener(kt *kit.Kit, req *protolb.ListenerWithRuleCreateReq) (
	*core.BatchCreateResult, error) {

	return common.Request[protolb.ListenerWithRuleCreateReq, core.BatchCreateResult](
		c.client, http.MethodPost, kt, req, "/listeners/create")
}

// UpdateListener 更新监听器
func (c *ClbClient) UpdateListener(kt *kit.Kit, id string, req *protolb.ListenerWithRuleUpdateReq) (
	*protolb.BatchCreateResult, error) {

	return common.Request[protolb.ListenerWithRuleUpdateReq, protolb.BatchCreateResult](
		c.client, http.MethodPatch, kt, req, "/listeners/%s", id)
}

// DeleteListener 删除监听器
func (c *ClbClient) DeleteListener(kt *kit.Kit, req *core.BatchDeleteReq) error {
	return common.RequestNoResp[core.BatchDeleteReq](c.client, http.MethodDelete, kt, req, "/listeners/batch")
}
