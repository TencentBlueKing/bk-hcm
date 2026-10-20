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

package lblogic

import (
	corelb "hcm/pkg/api/core/cloud/load-balancer"
	dataservice "hcm/pkg/client/data-service"
	"hcm/pkg/dal/dao/tools"
	"hcm/pkg/kit"
	"hcm/pkg/runtime/filter"
	"hcm/pkg/tools/slice"
)

// listExclusiveClustersByCloudIDs 按云上集群ID查询本地独占集群表，baseRules 为附加的固定过滤条件。
// 云上集群ID数量超过数据库 in 条件的元素上限时按上限分批查询，每一批再翻页取全，不依赖调用方限制ID数量。
func listExclusiveClustersByCloudIDs(kt *kit.Kit, cli *dataservice.Client, cloudIDs []string, fields []string,
	baseRules ...*filter.AtomRule) ([]corelb.ExclusiveClusterRaw, error) {

	clusters := make([]corelb.ExclusiveClusterRaw, 0, len(cloudIDs))
	for _, batch := range slice.Split(slice.Unique(cloudIDs), int(filter.DefaultMaxInLimit)) {
		rules := append(append(make([]*filter.AtomRule, 0, len(baseRules)+1), baseRules...),
			tools.RuleIn("cloud_id", batch))

		part, err := cli.Global.ListAllExclusiveCluster(kt, tools.ExpressionAnd(rules...), fields)
		if err != nil {
			return nil, err
		}
		clusters = append(clusters, part...)
	}

	return clusters, nil
}
