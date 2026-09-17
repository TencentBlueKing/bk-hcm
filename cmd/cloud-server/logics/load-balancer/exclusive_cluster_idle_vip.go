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
)

// CheckExclusiveClusterIdleVipQueryable 校验独占集群闲置VIP查询目标集群是否可查询：集群必须存在于本地独占
// 集群表、类型必须为TGW（四层）、且必须已分配给当前业务，三项均满足才允许发起实时查云。区分"集群不存在"/
// "类型非TGW"（均返回InvalidParameter）与"存在但不归属当前业务"（返回PermissionDenied，且不在错误信息中
// 回显该集群实际归属的业务）两类错误语义。
func CheckExclusiveClusterIdleVipQueryable(kt *kit.Kit, cli *dataservice.Client, bkBizID int64,
	cloudClusterID string) error {

	listReq := &core.ListReq{
		Fields: []string{"bk_biz_id", "cluster_type"},
		Filter: tools.ExpressionAnd(
			tools.RuleEqual("cloud_id", cloudClusterID),
		),
		Page: core.NewDefaultBasePage(),
	}
	result, err := cli.Global.ListExclusiveCluster(kt, listReq)
	if err != nil {
		logs.Errorf("check exclusive cluster(cloud_cluster_id=%s) idle vip queryable failed, err: %v, rid: %s",
			cloudClusterID, err, kt.Rid)
		return err
	}

	if len(result.Details) == 0 {
		return errf.Newf(errf.InvalidParameter, "exclusive cluster(cloud_cluster_id=%s) not found", cloudClusterID)
	}

	one := result.Details[0]
	if one.ClusterType != enumor.TGWClusterType {
		return errf.Newf(errf.InvalidParameter,
			"exclusive cluster(cloud_cluster_id=%s) type(%s) does not support idle vip query", cloudClusterID,
			one.ClusterType)
	}

	if one.BkBizID != bkBizID {
		return errf.Newf(errf.PermissionDenied,
			"exclusive cluster(cloud_cluster_id=%s) does not belong to biz(id=%d)", cloudClusterID, bkBizID)
	}

	return nil
}
