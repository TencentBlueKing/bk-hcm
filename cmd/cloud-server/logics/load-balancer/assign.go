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
	"hcm/pkg/tools/slice"
)

// AssignTCloud 分配负载均衡及关联资源到业务下
// 目前在负载均衡和监听器下有业务字段，监听器不能独立分配业务，分配时需要将监听器和对应的目标组分配到业务下
func AssignTCloud(kt *kit.Kit, cli *dataservice.Client, ids []string, bizID int64) error {

	if len(ids) == 0 {
		return fmt.Errorf("ids is required")
	}

	// 校验负载均衡信息
	if err := ValidateBeforeAssign(kt, cli, ids, bizID); err != nil {
		return err
	}

	// 获取负载均衡关联资源
	lblIds, tgIDs, err := GetLoadBalancerRelateResIDs(kt, cli, ids)
	if err != nil {
		return err
	}

	// 校验负载均衡关联资源信息
	if err := ValidateLBRelatedBeforeAssign(kt, cli, lblIds, tgIDs); err != nil {
		return err
	}

	// create assign audit
	audit := logicaudit.NewAudit(cli)
	if err := audit.ResBizAssignAudit(kt, enumor.LoadBalancerAuditResType, ids, bizID); err != nil {
		logs.Errorf("create assign clb audit failed, err: %v, rid: %s", err, kt.Rid)
		return err
	}

	// 分配负载均衡关联资源
	if err := AssignLoadBalancerRelated(kt, cli, lblIds, tgIDs, bizID); err != nil {
		return err
	}

	// 分配负载均衡
	update := &dataproto.BizBatchUpdateReq{IDs: ids, BkBizID: bizID}
	if err := cli.Global.LoadBalancer.BatchUpdateLbBizInfo(kt, update); err != nil {
		logs.Errorf("BatchUpdateLbBizInfo failed, err: %v, req: %+v, rid: %s", err, update, kt.Rid)
		return err
	}

	return nil
}

// ValidateBeforeAssign 分配负载均衡前校验
func ValidateBeforeAssign(kt *kit.Kit, cli *dataservice.Client, ids []string, bizID int64) error {
	listReq := &core.ListReq{
		Fields: []string{"id", "bk_biz_id"},
		Filter: tools.ContainersExpression("id", ids),
		Page:   core.NewDefaultBasePage(),
	}
	result, err := cli.Global.LoadBalancer.ListLoadBalancer(kt, listReq)
	if err != nil {
		logs.Errorf("list clb failed, err: %v, rid: %s", err, kt.Rid)
		return err
	}

	// 判断是否已经分配到业务下
	assignedIDs := make([]string, 0)
	for _, one := range result.Details {
		if one.BkBizID != constant.UnassignedBiz && one.BkBizID != bizID {
			assignedIDs = append(assignedIDs, one.ID)
		}
	}

	// 存在已经分配到业务下的clb实例，报错
	if len(assignedIDs) != 0 {
		return fmt.Errorf("load balancer(ids=%v) already assigned", assignedIDs)
	}

	return nil
}

// GetLoadBalancerRelateResIDs 获取clb关联资源列表，包括监听器和目标组
func GetLoadBalancerRelateResIDs(kt *kit.Kit, cli *dataservice.Client, lbIds []string) (
	lblIds []string, tgIDs []string, err error) {
	lbReq := &core.ListReq{
		Filter: tools.ContainersExpression("lb_id", lbIds),
		Page:   core.NewDefaultBasePage(),
	}
	lblResp, err := cli.Global.LoadBalancer.ListListener(kt, lbReq)
	if err != nil {
		logs.Errorf("fail to list listener for lb(ids=%v), err: %v, rid: %s", lbIds, err, kt.Rid)
		return nil, nil, err
	}
	for _, lbl := range lblResp.Details {
		lblIds = append(lblIds, lbl.ID)
	}

	tgRelResp, err := cli.Global.LoadBalancer.ListTargetGroupListenerRel(kt, lbReq)
	if err != nil {
		logs.Errorf("fail to list load balancer(ids=%v) related target group relation, err: %v, rid: %s",
			lbIds, err, kt.Rid)
		return nil, nil, err
	}
	for _, rel := range tgRelResp.Details {
		tgIDs = append(tgIDs, rel.TargetGroupID)
	}

	return lblIds, tgIDs, nil
}

// ValidateLBRelatedBeforeAssign 在分配前校验lb关联资源信息
func ValidateLBRelatedBeforeAssign(kt *kit.Kit, cli *dataservice.Client, lblIds []string,
	tgIds []string) error {

	// 目前都是以负载均衡粒度分配到业务，因此暂不做关联资源分配校验
	return nil
}

// AssignLoadBalancerRelated 分配负载均衡关联的监听器和规则
func AssignLoadBalancerRelated(kt *kit.Kit, cli *dataservice.Client, lblIds []string, tgIds []string,
	bizID int64) error {

	if len(lblIds) != 0 {
		// 分配关联规则、关联目标组
		for _, lblIdBatch := range slice.Split(lblIds, constant.BatchOperationMaxLimit) {
			updateLbl := &dataproto.BizBatchUpdateReq{IDs: lblIdBatch, BkBizID: bizID}
			if err := cli.Global.LoadBalancer.BatchUpdateListenerBizInfo(kt, updateLbl); err != nil {
				logs.Errorf("batch update listener biz info failed, err: %v, req: %+v, rid: %s", err, updateLbl, kt.Rid)
				return err
			}
		}

	}
	if len(tgIds) != 0 {
		for _, tgIdBatch := range slice.Split(tgIds, constant.BatchOperationMaxLimit) {
			updateTg := &dataproto.BizBatchUpdateReq{IDs: tgIdBatch, BkBizID: bizID}
			if err := cli.Global.LoadBalancer.BatchUpdateTargetGroupBizInfo(kt, updateTg); err != nil {
				logs.Errorf("batch update target group biz info failed, err: %v, req: %+v, rid: %s", err, updateTg,
					kt.Rid)
				return err
			}
		}
	}

	return nil
}
