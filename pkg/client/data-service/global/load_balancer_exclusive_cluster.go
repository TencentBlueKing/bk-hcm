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

// ListExclusiveCluster list load balancer exclusive cluster, extension is returned as raw json.
func (cli *restClient) ListExclusiveCluster(kt *kit.Kit, req *core.ListReq) (
	*dataproto.ExclusiveClusterListResult, error) {

	return common.Request[core.ListReq, dataproto.ExclusiveClusterListResult](
		cli.client, rest.POST, kt, req, "/load_balancer_exclusive_clusters/list")
}

// BatchUpdateExclusiveClusterBizID batch assign load balancer exclusive cluster to a business.
func (cli *restClient) BatchUpdateExclusiveClusterBizID(kt *kit.Kit,
	req *dataproto.ExclusiveClusterBatchUpdateBizIDReq) error {

	return common.RequestNoResp[dataproto.ExclusiveClusterBatchUpdateBizIDReq](
		cli.client, rest.PATCH, kt, req, "/load_balancer_exclusive_clusters/biz")
}

// BatchDeleteExclusiveCluster batch delete load balancer exclusive cluster.
func (cli *restClient) BatchDeleteExclusiveCluster(kt *kit.Kit,
	req *dataproto.ExclusiveClusterBatchDeleteReq) error {

	return common.RequestNoResp[dataproto.ExclusiveClusterBatchDeleteReq](
		cli.client, rest.DELETE, kt, req, "/load_balancer_exclusive_clusters/batch")
}
