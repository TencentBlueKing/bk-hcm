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

package argstpl

import (
	"errors"
	"fmt"

	logicaudit "hcm/cmd/cloud-server/logics/audit"
	"hcm/pkg/api/core"
	coreargstpl "hcm/pkg/api/core/cloud/argument-template"
	protocloud "hcm/pkg/api/data-service/cloud"
	dataservice "hcm/pkg/client/data-service"
	"hcm/pkg/criteria/constant"
	"hcm/pkg/criteria/enumor"
	"hcm/pkg/kit"
	"hcm/pkg/logs"
	"hcm/pkg/runtime/filter"
	"hcm/pkg/tools/slice"
)

// Assign 分配参数模版到业务下
func Assign(kt *kit.Kit, cli *dataservice.Client, ids []string, bizID int64) error {
	if len(ids) == 0 {
		return errors.New("ids is required")
	}

	if err := ValidateBeforeAssign(kt, cli, ids); err != nil {
		return err
	}

	// create assign audit
	audit := logicaudit.NewAudit(cli)
	if err := audit.ResBizAssignAudit(kt, enumor.ArgumentTemplateAuditResType, ids, bizID); err != nil {
		logs.Errorf("create assign argstpl audit failed, ids: %v, bizID: %d, err: %v, rid: %s", ids, bizID, err, kt.Rid)
		return err
	}

	// assign
	req := &protocloud.ArgsTplBatchUpdateExprReq{
		IDs:     ids,
		BkBizID: bizID,
	}
	_, err := cli.Global.ArgsTpl.BatchUpdateArgsTpl(kt, req)
	if err != nil {
		logs.Errorf("batch update argstpl db failed, ids: %v, bizID: %d, err: %v, rid: %s", ids, bizID, err, kt.Rid)
		return err
	}

	return nil
}

// ValidateBeforeAssign 分配前置校验
func ValidateBeforeAssign(kt *kit.Kit, cli *dataservice.Client, ids []string) error {
	// 判断是否已经分配
	listReq := &core.ListReq{
		Filter: &filter.Expression{
			Op: filter.And,
			Rules: []filter.RuleFactory{
				&filter.AtomRule{Field: "id", Op: filter.In.Factory(), Value: ids},
				&filter.AtomRule{Field: "bk_biz_id", Op: filter.NotEqual.Factory(), Value: constant.UnassignedBiz},
			},
		},
		Page: core.NewDefaultBasePage(),
	}
	listResp, err := cli.Global.ArgsTpl.ListArgsTpl(kt, listReq)
	if err != nil {
		logs.Errorf("list argument template failed, req: %+v, err: %v, rid: %s", listReq, err, kt.Rid)
		return err
	}

	if len(listResp.Details) != 0 {
		return fmt.Errorf("argument template(ids=%v) already assigned", slice.Map(listResp.Details,
			func(at coreargstpl.BaseArgsTpl) string { return at.ID }))
	}

	return nil
}
