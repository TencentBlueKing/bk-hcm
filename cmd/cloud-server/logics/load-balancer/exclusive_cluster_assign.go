/*
 *
 * TencentBlueKing is pleased to support the open source community by making
 * 蓝鲸智云 - 混合云管理平台 (BlueKing - Hybrid Cloud Management System) available.
 * Copyright (C) 2024 THL A29 Limited,
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

	logicaudit "hcm/cmd/cloud-server/logics/audit"
	"hcm/pkg/api/core"
	dataproto "hcm/pkg/api/data-service/cloud"
	dataservice "hcm/pkg/client/data-service"
	"hcm/pkg/criteria/constant"
	"hcm/pkg/criteria/enumor"
	"hcm/pkg/dal/dao/tools"
	"hcm/pkg/kit"
	"hcm/pkg/logs"
)

// AssignExclusiveClusterToBiz 分配独占集群到业务下，独占集群没有关联子资源，不需要级联分配。
func AssignExclusiveClusterToBiz(kt *kit.Kit, cli *dataservice.Client, ids []string, bizID int64) error {
	if len(ids) == 0 {
		return fmt.Errorf("cluster ids is required")
	}

	// 校验独占集群分配前状态
	if err := ValidateExclusiveClusterBeforeAssign(kt, cli, ids, bizID); err != nil {
		return err
	}

	// create assign audit
	audit := logicaudit.NewAudit(cli)
	if err := audit.ResBizAssignAudit(kt, enumor.LoadBalancerExclusiveClusterAuditResType, ids, bizID); err != nil {
		logs.Errorf("create assign exclusive cluster audit failed, err: %v, rid: %s", err, kt.Rid)
		return err
	}

	// 分配独占集群
	update := &dataproto.ExclusiveClusterBatchUpdateBizIDReq{ClusterIDs: ids, BkBizID: bizID}
	if err := cli.Global.BatchUpdateExclusiveClusterBizID(kt, update); err != nil {
		logs.Errorf("BatchUpdateExclusiveClusterBizID failed, err: %v, req: %+v, rid: %s", err, update, kt.Rid)
		return err
	}

	return nil
}

// ValidateExclusiveClusterBeforeAssign 分配独占集群前校验，仅"已分配给其它业务"的集群才拒绝，已分配给目标业务
// 本身视为幂等放行，与 load-balancer 分配业务接口的判定条件一致。
func ValidateExclusiveClusterBeforeAssign(kt *kit.Kit, cli *dataservice.Client, ids []string, bizID int64) error {
	listReq := &core.ListReq{
		Fields: []string{"id", "bk_biz_id"},
		Filter: tools.ContainersExpression("id", ids),
		Page:   core.NewDefaultBasePage(),
	}
	result, err := cli.Global.ListExclusiveCluster(kt, listReq)
	if err != nil {
		logs.Errorf("list exclusive cluster failed, err: %v, rid: %s", err, kt.Rid)
		return err
	}
	if result == nil {
		return fmt.Errorf("list exclusive cluster got empty response, ids: %v, rid: %s", ids, kt.Rid)
	}

	// 判断是否已经分配到其它业务下
	assignedIDs := make([]string, 0)
	for _, one := range result.Details {
		if one.BkBizID != constant.UnassignedBiz && one.BkBizID != bizID {
			assignedIDs = append(assignedIDs, one.ID)
		}
	}

	// 存在已经分配到其它业务下的独占集群，整批拒绝
	if len(assignedIDs) != 0 {
		return fmt.Errorf("exclusive cluster(ids=%v) already assigned to other biz", assignedIDs)
	}

	return nil
}
