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
	"hcm/pkg/api/core"
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
func CheckExclusiveClusterOwnership(kt *kit.Kit, cli *dataservice.Client, bkBizID int64, clusterTag string,
	cloudClusterIDs []string) error {

	if len(clusterTag) != 0 {
		if err := checkClusterTagOwnership(kt, cli, bkBizID, clusterTag); err != nil {
			return err
		}
	}

	if len(cloudClusterIDs) != 0 {
		if err := checkClusterIDsOwnership(kt, cli, bkBizID, cloudClusterIDs); err != nil {
			return err
		}
	}

	return nil
}

// checkClusterTagOwnership 校验七层独占集群标签是否归属当前业务。
func checkClusterTagOwnership(kt *kit.Kit, cli *dataservice.Client, bkBizID int64, clusterTag string) error {
	req := &core.ListReq{
		Fields: []string{"id"},
		Filter: tools.ExpressionAnd(
			tools.RuleEqual("bk_biz_id", bkBizID),
			tools.RuleEqual("cluster_type", enumor.STGWClusterType),
			tools.RuleEqual("cluster_tag", clusterTag),
		),
		Page: core.NewCountPage(),
	}
	result, err := cli.Global.ListExclusiveCluster(kt, req)
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
func checkClusterIDsOwnership(kt *kit.Kit, cli *dataservice.Client, bkBizID int64, cloudClusterIDs []string) error {
	remaining := cvt.StringSliceToMap(cloudClusterIDs)

	req := &core.ListReq{
		Fields: []string{"cloud_id"},
		Filter: tools.ExpressionAnd(
			tools.RuleEqual("bk_biz_id", bkBizID),
			tools.RuleEqual("cluster_type", enumor.TGWClusterType),
			tools.RuleIn("cloud_id", cloudClusterIDs),
		),
		Page: core.NewDefaultBasePage(),
	}
	result, err := cli.Global.ListExclusiveCluster(kt, req)
	if err != nil {
		logs.Errorf("check cloud_cluster_ids ownership failed, err: %v, bizID: %d, rid: %s", err, bkBizID, kt.Rid)
		return err
	}

	for _, one := range result.Details {
		delete(remaining, one.CloudID)
	}

	if len(remaining) != 0 {
		return errf.Newf(errf.PermissionDenied, "cloud_cluster_ids does not belong to biz(id=%d)", bkBizID)
	}

	return nil
}
