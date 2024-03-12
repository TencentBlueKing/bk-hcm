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

package clb

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

// Assign 分配负载均衡及关联资源到业务下
func Assign(kt *kit.Kit, cli *dataservice.Client, ids []string, bizID int64) error {

	if len(ids) == 0 {
		return fmt.Errorf("ids is required")
	}

	// 校验负载均衡信息
	if err := ValidateBeforeAssign(kt, cli, ids); err != nil {
		return err
	}

	// 获取负载均衡关联资源
	lblIds, ruleIds, err := GetClbRelResIDs(kt, cli, ids)
	if err != nil {
		return err
	}

	// 校验负载均衡关联资源信息
	if err := ValidateClbRelResBeforeAssign(kt, cli, lblIds, ruleIds); err != nil {
		return err
	}

	// create assign audit
	audit := logicaudit.NewAudit(cli)
	if err := audit.ResBizAssignAudit(kt, enumor.LoadBalancerAuditResType, ids, bizID); err != nil {
		logs.Errorf("create assign clb audit failed, err: %v, rid: %s", err, kt.Rid)
		return err
	}

	// 分配负载均衡关联资源
	if err := AssignClbRelRes(kt, cli, lblIds, ruleIds, bizID); err != nil {
		return err
	}

	// 分配负载均衡
	update := &dataproto.ClbBizBatchUpdateReq{IDs: ids, BkBizID: bizID}
	if err := cli.Global.LoadBalancer.BatchUpdateClbBizInfo(kt, update); err != nil {
		logs.Errorf("BatchUpdateClbBizInfo failed, err: %v, req: %v, rid: %s", err, update, kt.Rid)
		return err
	}

	return nil
}

// AssignClbRelRes 分配负载均衡关联的监听器和规则
func AssignClbRelRes(kt *kit.Kit, cli *dataservice.Client, lblIds []string, ruleIds []string, bizID int64) error {

	// TODO 分配关联监听器和规则

	return nil
}

// GetClbRelResIDs 获取clb关联资源列表，包括监听器和7层规则
func GetClbRelResIDs(kt *kit.Kit, cli *dataservice.Client, ids []string) (
	lblIds []string, ruleIds []string, err error) {
	//TODO 补充关联监听器和规则的获取
	return nil, nil, nil
}

// ValidateClbRelResBeforeAssign 校验clb关联资源在分配前
func ValidateClbRelResBeforeAssign(kt *kit.Kit, cli *dataservice.Client, lblIds []string, ruleIds []string) error {
	//TODO 补充监听器和规则的分配校验

	return nil
}

// ValidateBeforeAssign 分配负载均衡前校验
func ValidateBeforeAssign(kt *kit.Kit, cli *dataservice.Client, ids []string) error {
	listReq := &core.ListReq{
		Fields: []string{"id", "bk_biz_id"},
		Filter: tools.ContainersExpression("id", ids),
		Page:   core.NewDefaultBasePage(),
	}
	result, err := cli.Global.LoadBalancer.ListClb(kt, listReq)
	if err != nil {
		logs.Errorf("list clb failed, err: %v, rid: %s", err, kt.Rid)
		return err
	}

	// 判断是否已经分配到业务下
	assignedIDs := make([]string, 0)
	for _, one := range result.Details {
		if one.BkBizID != constant.UnassignedBiz {
			assignedIDs = append(assignedIDs, one.ID)
		}
	}

	// 存在已经分配到业务下的clb实例，报错
	if len(assignedIDs) != 0 {
		return fmt.Errorf("clb(ids=%v) already assigned", assignedIDs)
	}

	return nil
}
