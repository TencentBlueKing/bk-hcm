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
	"fmt"

	"hcm/pkg/api/core"
	dataservice "hcm/pkg/client/data-service"
	"hcm/pkg/criteria/enumor"
	"hcm/pkg/dal/dao/tools"
	"hcm/pkg/kit"
	"hcm/pkg/runtime/filter"
)

// ComputeExclusiveClusterEgressSet 计算「本次可能分配到的集群」允许的出口集合（R-008）：只选四层返回
// cloud_cluster_ids 对应 TGW 的出口去重集合（E4_set）；只选七层返回当前业务已分配该标签下全部 STGW 的出口去重集合
// （E7_set）；四层七层都选返回 E4_set 与 E7_set 的交集。
func ComputeExclusiveClusterEgressSet(kt *kit.Kit, cli *dataservice.Client, bkBizID int64, clusterTag string,
	cloudClusterIDs []string) (map[string]struct{}, error) {

	var e4Set, e7Set map[string]struct{}
	var err error

	if len(cloudClusterIDs) != 0 {
		e4Filter := tools.ExpressionAnd(
			tools.RuleEqual("bk_biz_id", bkBizID),
			tools.RuleEqual("cluster_type", enumor.TGWClusterType),
			tools.RuleIn("cloud_id", cloudClusterIDs),
		)
		if e4Set, err = queryExclusiveClusterEgressSet(kt, cli, e4Filter); err != nil {
			return nil, err
		}
	}

	if len(clusterTag) != 0 {
		e7Filter := tools.ExpressionAnd(
			tools.RuleEqual("bk_biz_id", bkBizID),
			tools.RuleEqual("cluster_type", enumor.STGWClusterType),
			tools.RuleEqual("cluster_tag", clusterTag),
		)
		if e7Set, err = queryExclusiveClusterEgressSet(kt, cli, e7Filter); err != nil {
			return nil, err
		}
	}

	switch {
	case len(cloudClusterIDs) != 0 && len(clusterTag) != 0:
		return intersectEgressSet(e4Set, e7Set), nil
	case len(cloudClusterIDs) != 0:
		return e4Set, nil
	default:
		return e7Set, nil
	}
}

// queryExclusiveClusterEgressSet 按过滤条件查询本地独占集群表，返回去重后的出口集合。
func queryExclusiveClusterEgressSet(kt *kit.Kit, cli *dataservice.Client, expr *filter.Expression) (
	map[string]struct{}, error) {

	req := &core.ListReq{
		Fields: []string{"egress"},
		Filter: expr,
		Page:   core.NewDefaultBasePage(),
	}
	result, err := cli.Global.ListExclusiveCluster(kt, req)
	if err != nil {
		return nil, err
	}

	set := make(map[string]struct{}, len(result.Details))
	for _, one := range result.Details {
		if len(one.Egress) != 0 {
			set[one.Egress] = struct{}{}
		}
	}
	return set, nil
}

// intersectEgressSet 返回两个出口集合的交集。
func intersectEgressSet(a, b map[string]struct{}) map[string]struct{} {
	result := make(map[string]struct{})
	for egress := range a {
		if _, ok := b[egress]; ok {
			result[egress] = struct{}{}
		}
	}
	return result
}

// CheckBandwidthPackageEgress 校验带宽包出口是否落在允许的出口集合内，不满足返回错误信息。
func CheckBandwidthPackageEgress(allowedSet map[string]struct{}, egress string) error {
	if _, ok := allowedSet[egress]; !ok {
		return fmt.Errorf("bandwidth package egress(%s) is not allowed by the selected exclusive cluster(s)", egress)
	}
	return nil
}
