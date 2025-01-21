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
	apidata "hcm/pkg/api/data-service/cloud"
	"hcm/pkg/client/common"
	"hcm/pkg/criteria/enumor"
	"hcm/pkg/kit"
	"hcm/pkg/rest"
)

// NewResUsageBizRelRelClient create a new resource usage biz rel api client.
func NewResUsageBizRelRelClient(client rest.ClientInterface) *ResUsageBizRelClient {
	return &ResUsageBizRelClient{
		client: client,
	}
}

// ResUsageBizRelClient is data service resource usage biz rel api client.
type ResUsageBizRelClient struct {
	client rest.ClientInterface
}

// SetBizRels resource usage biz rels.
func (cli *ResUsageBizRelClient) SetBizRels(kt *kit.Kit, resType enumor.CloudResourceType, resID string,
	req *apidata.ResUsageBizRelUpdateReq) error {

	return common.RequestNoResp[apidata.ResUsageBizRelUpdateReq](cli.client, rest.PUT, kt, req,
		"/res_usage_biz_rels/res_types/%s/%s", resType, resID)
}

// ListResUsageBizRel resource usage biz rels.
func (cli *ResUsageBizRelClient) ListResUsageBizRel(kt *kit.Kit, req *core.ListReq) (
	*apidata.ListResUsageBizRelResult, error) {

	return common.Request[core.ListReq, apidata.ListResUsageBizRelResult](cli.client, rest.POST, kt, req,
		"/res_usage_biz_rels/list")
}
