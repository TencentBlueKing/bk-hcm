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
	"sort"

	"hcm/pkg/api/core"
	protolb "hcm/pkg/api/hc-service/load-balancer"
	dataservice "hcm/pkg/client/data-service"
	"hcm/pkg/criteria/enumor"
	"hcm/pkg/criteria/errf"
	"hcm/pkg/dal/dao/tools"
	"hcm/pkg/kit"
	"hcm/pkg/logs"
	cvt "hcm/pkg/tools/converter"
)

// CheckExclusiveClusterOwnership 校验独占集群 cluster_tag/cloud_cluster_ids 是否归属当前业务：cluster_tag 非空时要求
// 当前业务下存在使用该标签的 STGW（七层）集群；cloud_cluster_ids 非空时要求每一个 ID 都是当前业务已分配的 TGW（四层）
// 集群云上 ID。任一不满足均返回 PermissionDenied，且不在错误信息中回显查询到的其它业务集群数据。
func CheckExclusiveClusterOwnership(kt *kit.Kit, cli *dataservice.Client,
	req *protolb.TCloudLoadBalancerCreateReq) error {

	if len(cvt.PtrToVal(req.ClusterTag)) != 0 {
		if err := checkClusterTagOwnership(kt, cli, req); err != nil {
			return err
		}
	}

	if len(req.CloudClusterIDs) != 0 {
		if err := checkClusterIDsOwnership(kt, cli, req); err != nil {
			return err
		}
	}

	return nil
}

// checkClusterTagOwnership 校验七层独占集群标签是否归属当前业务。
func checkClusterTagOwnership(kt *kit.Kit, cli *dataservice.Client, req *protolb.TCloudLoadBalancerCreateReq) error {
	bkBizID := req.BkBizID

	listReq := &core.ListReq{
		Fields: []string{"id"},
		Filter: tools.ExpressionAnd(
			tools.RuleEqual("bk_biz_id", bkBizID),
			tools.RuleEqual("cluster_type", enumor.STGWClusterType),
			tools.RuleEqual("cluster_tag", cvt.PtrToVal(req.ClusterTag)),
			tools.RuleEqual("account_id", req.AccountID),
			tools.RuleEqual("region", req.Region),
		),
		Page: core.NewCountPage(),
	}
	result, err := cli.Global.ListExclusiveCluster(kt, listReq)
	if err != nil {
		logs.Errorf("check cluster_tag ownership failed, err: %v, bizID: %d, rid: %s", err, bkBizID, kt.Rid)
		return err
	}

	if result.Count == 0 {
		return errf.Newf(errf.PermissionDenied, "cluster_tag does not belong to biz(id=%d)", bkBizID)
	}

	return nil
}

// checkClusterIDsOwnership 校验四层独占集群云上 ID 列表是否都归属当前业务。
func checkClusterIDsOwnership(kt *kit.Kit, cli *dataservice.Client, req *protolb.TCloudLoadBalancerCreateReq) error {
	bkBizID := req.BkBizID
	remaining := cvt.StringSliceToMap(req.CloudClusterIDs)

	clusters, err := listExclusiveClustersByCloudIDs(kt, cli, req.CloudClusterIDs, []string{"cloud_id"},
		tools.RuleEqual("bk_biz_id", bkBizID),
		tools.RuleEqual("cluster_type", enumor.TGWClusterType),
		tools.RuleEqual("account_id", req.AccountID),
		tools.RuleEqual("region", req.Region),
	)
	if err != nil {
		logs.Errorf("check cloud_cluster_ids ownership failed, err: %v, bizID: %d, rid: %s", err, bkBizID, kt.Rid)
		return err
	}

	for _, one := range clusters {
		delete(remaining, one.CloudID)
	}

	if len(remaining) != 0 {
		missing := cvt.MapKeyToStringSlice(remaining)
		sort.Strings(missing)
		logs.Errorf("cloud_cluster_ids(%v) does not belong to biz(id=%d), rid: %s", missing, bkBizID, kt.Rid)
		return errf.Newf(errf.PermissionDenied, "cloud_cluster_ids(%v) does not belong to biz(id=%d)", missing,
			bkBizID)
	}

	return nil
}
