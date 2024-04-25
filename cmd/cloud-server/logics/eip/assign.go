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

package eip

import (
	"fmt"

	logicaudit "hcm/cmd/cloud-server/logics/audit"
	"hcm/pkg/api/core"
	datarelproto "hcm/pkg/api/data-service/cloud"
	dataproto "hcm/pkg/api/data-service/cloud/eip"
	dataservice "hcm/pkg/client/data-service"
	"hcm/pkg/criteria/constant"
	"hcm/pkg/criteria/enumor"
	"hcm/pkg/dal/dao/tools"
	"hcm/pkg/kit"
	"hcm/pkg/logs"
	"hcm/pkg/tools/slice"
)

// Assign 分配EIP到业务下，isBind表示是分配绑定了的EIP，还是未绑定的EIP，校验有所不同
func Assign(kt *kit.Kit, cli *dataservice.Client, ids []string, bizID uint64, isBind bool) error {

	if len(ids) == 0 {
		return fmt.Errorf("ids is required")
	}

	if err := ValidateBeforeAssign(kt, cli, int64(bizID), ids, isBind); err != nil {
		return err
	}

	// create assign audit
	audit := logicaudit.NewAudit(cli)
	if err := audit.ResBizAssignAudit(kt, enumor.EipAuditResType, ids, int64(bizID)); err != nil {
		logs.Errorf("create assign eip audit failed, err: %v, rid: %s", err, kt.Rid)
		return err
	}

	req := &dataproto.EipBatchUpdateReq{
		IDs:     ids,
		BkBizID: bizID,
	}
	_, err := cli.Global.BatchUpdateEip(kt, req)
	if err != nil {
		logs.Errorf("batch update eip failed, err: %v, rid: %s", err, kt.Rid)
		return err
	}

	return nil
}

// ValidateBeforeAssign 分配前置校验
func ValidateBeforeAssign(kt *kit.Kit, cli *dataservice.Client, targetBizId int64, eipIds []string, isBind bool) error {
	// 判断是否已经分配
	// 允许已经在目标业务下
	listReq := &core.ListReq{
		Filter: tools.ExpressionAnd(
			tools.RuleIn("id", eipIds),
			tools.RuleNotIn("bk_biz_id", []int64{constant.UnassignedBiz, targetBizId}),
		),
		Page: core.NewDefaultBasePage(),
	}
	listResp, err := cli.Global.ListEip(kt, listReq)
	if err != nil {
		logs.Errorf("list eip failed, err: %v, req: %+v, rid: %s", err, listReq, kt.Rid)
		return err
	}

	if len(listResp.Details) != 0 {
		return fmt.Errorf("eip(ids=%v) already assigned", slice.Map(listResp.Details,
			func(eip *dataproto.EipResult) string { return eip.ID }))
	}

	// 判断是否关联资源
	listRelReq := &core.ListReq{
		Filter: tools.ContainersExpression("eip_id", eipIds),
		Page:   core.NewDefaultBasePage(),
	}
	listRelResp, err := cli.Global.ListEipCvmRel(kt, listRelReq)
	if err != nil {
		logs.Errorf("list eip cvm rel failed, err: %v, rid: %s", err, kt.Rid)
		return err
	}

	// 如果分配未绑定资源的EIP，但关系表有数据，需要报错
	if !isBind && len(listRelResp.Details) != 0 {
		return fmt.Errorf("eip(ids=%v) already bind cvm", slice.Map(listRelResp.Details,
			func(eip *datarelproto.EipCvmRelResult) string { return eip.EipID }))
	}

	// 如果分配绑定资源的EIP，但实际未绑定，也需要报错
	if isBind {
		eipBindMap := make(map[string]bool)
		for _, one := range listRelResp.Details {
			eipBindMap[one.EipID] = true
		}

		if len(eipIds) != len(eipBindMap) {
			unBindIDs := slice.Filter(eipIds, func(id string) bool {
				return !eipBindMap[id]
			})
			return fmt.Errorf("eip(ids=%v) not bind cvm", unBindIDs)
		}
	}

	return nil
}
