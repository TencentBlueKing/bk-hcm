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
	protolb "hcm/pkg/api/hc-service/load-balancer"
	dataservice "hcm/pkg/client/data-service"
	"hcm/pkg/criteria/enumor"
	"hcm/pkg/dal/dao/tools"
	"hcm/pkg/kit"
	cvt "hcm/pkg/tools/converter"
	"hcm/pkg/tools/slice"
)

// ComputeExclusiveClusterEgressSet 计算可能被分配到的独占集群允许的出口列表：只传四层集群ID时返回这些 TGW 集群的
// 出口去重列表；只传七层标签时返回当前业务下该标签的全部 STGW 集群的出口去重列表；两者都传时返回两个列表的交集。
func ComputeExclusiveClusterEgressSet(kt *kit.Kit, cli *dataservice.Client,
	req *protolb.TCloudLoadBalancerCreateReq) ([]string, error) {

	var tgwEgresses, stgwEgresses []string
	clusterTag := cvt.PtrToVal(req.ClusterTag)

	if len(req.CloudClusterIDs) != 0 {
		tgwClusters, err := listExclusiveClustersByCloudIDs(kt, cli,
			req.CloudClusterIDs, []string{"egress"},
			tools.RuleEqual("bk_biz_id", req.BkBizID),
			tools.RuleEqual("cluster_type", enumor.TGWClusterType),
			tools.RuleEqual("account_id", req.AccountID),
			tools.RuleEqual("region", req.Region),
		)
		if err != nil {
			return nil, err
		}
		tgwEgresses = collectEgresses(tgwClusters)
	}

	if len(clusterTag) != 0 {
		stgwFilter := tools.ExpressionAnd(
			tools.RuleEqual("bk_biz_id", req.BkBizID),
			tools.RuleEqual("cluster_type", enumor.STGWClusterType),
			tools.RuleEqual("cluster_tag", clusterTag),
			tools.RuleEqual("account_id", req.AccountID),
			tools.RuleEqual("region", req.Region),
		)
		stgwClusters, err := cli.Global.ListAllExclusiveCluster(kt, stgwFilter, []string{"egress"})
		if err != nil {
			return nil, err
		}
		stgwEgresses = collectEgresses(stgwClusters)
	}

	switch {
	case len(req.CloudClusterIDs) != 0 && len(clusterTag) != 0:
		return slice.Intersection(tgwEgresses, stgwEgresses), nil
	case len(req.CloudClusterIDs) != 0:
		return tgwEgresses, nil
	default:
		return stgwEgresses, nil
	}
}

// collectEgresses 提取独占集群的出口并去重，忽略出口为空的集群。
func collectEgresses(clusters []corelb.ExclusiveClusterRaw) []string {
	egresses := make([]string, 0, len(clusters))
	for _, one := range clusters {
		if len(one.Egress) != 0 {
			egresses = append(egresses, one.Egress)
		}
	}
	return slice.Unique(egresses)
}
