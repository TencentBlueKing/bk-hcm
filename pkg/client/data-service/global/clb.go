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

// ClbClient is data service clb api client.
type ClbClient struct {
	client rest.ClientInterface
}

// NewClbClient create a new clb api client.
func NewClbClient(client rest.ClientInterface) *ClbClient {
	return &ClbClient{
		client: client,
	}
}

// ListLoadBalancer list clb.
func (cli *ClbClient) ListLoadBalancer(kt *kit.Kit, req *core.ListReq) (*dataproto.ClbListResult, error) {
	return common.Request[core.ListReq, dataproto.ClbListResult](cli.client, rest.POST, kt, req, "/load_balancers/list")
}

// BatchUpdateClbBizInfo update biz
func (cli *ClbClient) BatchUpdateClbBizInfo(kt *kit.Kit, req *dataproto.ClbBizBatchUpdateReq) error {
	return common.RequestNoResp[dataproto.ClbBizBatchUpdateReq](cli.client, rest.PATCH,
		kt, req, "/load_balancers/biz/batch/update")

}

// ListListener list listener.
func (cli *ClbClient) ListListener(kt *kit.Kit, req *dataproto.ListListenerReq) (
	*dataproto.ListenerListResult, error) {

	return common.Request[dataproto.ListListenerReq, dataproto.ListenerListResult](
		cli.client, rest.POST, kt, req, "/load_balancers/listeners/list")
}

// ListUrlRule list url rule.
func (cli *ClbClient) ListUrlRule(kt *kit.Kit, req *dataproto.ListTCloudURLRuleReq) (
	*dataproto.TCloudURLRuleListResult, error) {

	return common.Request[dataproto.ListTCloudURLRuleReq, dataproto.TCloudURLRuleListResult](
		cli.client, rest.POST, kt, req, "/load_balancers/url_rules/list")
}

// ListTarget list target.
func (cli *ClbClient) ListTarget(kt *kit.Kit, req *core.ListReq) (*dataproto.ClbTargetListResult, error) {
	return common.Request[core.ListReq, dataproto.ClbTargetListResult](
		cli.client, rest.POST, kt, req, "/load_balancers/targets/list")
}

// ListTargetGroup list target group.
func (cli *ClbClient) ListTargetGroup(kt *kit.Kit, req *core.ListReq) (*dataproto.ClbTargetGroupListResult, error) {
	return common.Request[core.ListReq, dataproto.ClbTargetGroupListResult](
		cli.client, rest.POST, kt, req, "/load_balancers/target_groups/list")
}
