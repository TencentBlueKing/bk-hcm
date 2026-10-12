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
	"fmt"

	"hcm/pkg/api/core"
	corelb "hcm/pkg/api/core/cloud/load-balancer"
	dataproto "hcm/pkg/api/data-service/cloud"
	"hcm/pkg/client/common"
	"hcm/pkg/kit"
	"hcm/pkg/rest"
	"hcm/pkg/runtime/filter"
)

// ListExclusiveCluster list load balancer exclusive cluster, extension is returned as raw json.
func (cli *restClient) ListExclusiveCluster(kt *kit.Kit, req *core.ListReq) (
	*dataproto.ExclusiveClusterListResult, error) {

	return common.Request[core.ListReq, dataproto.ExclusiveClusterListResult](
		cli.client, rest.POST, kt, req, "/load_balancer_exclusive_clusters/list")
}

// ListAllExclusiveCluster 翻页查询满足过滤条件的全部独占集群，extension 以原始 json 返回。
func (cli *restClient) ListAllExclusiveCluster(kt *kit.Kit, expr *filter.Expression, fields []string) (
	[]corelb.ExclusiveClusterRaw, error) {

	details := make([]corelb.ExclusiveClusterRaw, 0)
	page := core.NewDefaultBasePage()
	for {
		result, err := cli.ListExclusiveCluster(kt, &core.ListReq{Filter: expr, Fields: fields, Page: page})
		if err != nil {
			return nil, err
		}
		if result == nil {
			return nil, fmt.Errorf("list exclusive cluster got empty response")
		}

		details = append(details, result.Details...)
		if uint(len(result.Details)) < page.Limit {
			return details, nil
		}
		page.Start += uint32(page.Limit)
	}
}

// BatchDeleteExclusiveCluster batch delete load balancer exclusive cluster.
func (cli *restClient) BatchDeleteExclusiveCluster(kt *kit.Kit,
	req *dataproto.ExclusiveClusterBatchDeleteReq) error {

	return common.RequestNoResp[dataproto.ExclusiveClusterBatchDeleteReq](
		cli.client, rest.DELETE, kt, req, "/load_balancer_exclusive_clusters/batch")
}
