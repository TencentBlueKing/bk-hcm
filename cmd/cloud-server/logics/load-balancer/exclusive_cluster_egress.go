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
	dataservice "hcm/pkg/client/data-service"
	"hcm/pkg/criteria/enumor"
	"hcm/pkg/dal/dao/tools"
	"hcm/pkg/kit"
	"hcm/pkg/runtime/filter"
	"hcm/pkg/tools/slice"
)

// ComputeExclusiveClusterEgressSet 计算可能被分配到的独占集群允许的出口列表：只传四层集群ID时返回这些 TGW 集群的
// 出口去重列表；只传七层标签时返回当前业务下该标签的全部 STGW 集群的出口去重列表；两者都传时返回两个列表的交集。
func ComputeExclusiveClusterEgressSet(kt *kit.Kit, cli *dataservice.Client, bkBizID int64, clusterTag string,
	cloudClusterIDs []string) ([]string, error) {

	var tgwEgresses, stgwEgresses []string
	var err error

	if len(cloudClusterIDs) != 0 {
		tgwFilter := tools.ExpressionAnd(
			tools.RuleEqual("bk_biz_id", bkBizID),
			tools.RuleEqual("cluster_type", enumor.TGWClusterType),
			tools.RuleIn("cloud_id", cloudClusterIDs),
		)
		if tgwEgresses, err = queryExclusiveClusterEgresses(kt, cli, tgwFilter); err != nil {
			return nil, err
		}
	}

	if len(clusterTag) != 0 {
		stgwFilter := tools.ExpressionAnd(
			tools.RuleEqual("bk_biz_id", bkBizID),
			tools.RuleEqual("cluster_type", enumor.STGWClusterType),
			tools.RuleEqual("cluster_tag", clusterTag),
		)
		if stgwEgresses, err = queryExclusiveClusterEgresses(kt, cli, stgwFilter); err != nil {
			return nil, err
		}
	}

	switch {
	case len(cloudClusterIDs) != 0 && len(clusterTag) != 0:
		return slice.Intersection(tgwEgresses, stgwEgresses), nil
	case len(cloudClusterIDs) != 0:
		return tgwEgresses, nil
	default:
		return stgwEgresses, nil
	}
}

// queryExclusiveClusterEgresses 按过滤条件查询本地独占集群表，返回去重后的出口列表。
func queryExclusiveClusterEgresses(kt *kit.Kit, cli *dataservice.Client, expr *filter.Expression) ([]string, error) {
	clusters, err := cli.Global.ListAllExclusiveCluster(kt, expr, []string{"egress"})
	if err != nil {
		return nil, err
	}

	egresses := make([]string, 0, len(clusters))
	for _, one := range clusters {
		if len(one.Egress) != 0 {
			egresses = append(egresses, one.Egress)
		}
	}
	return slice.Unique(egresses), nil
}
