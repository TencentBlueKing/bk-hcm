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

package cvm

import (
	"fmt"

	proto "hcm/pkg/api/cloud-server/cvm"
	"hcm/pkg/api/core"
	dataproto "hcm/pkg/api/data-service/cloud"
	"hcm/pkg/criteria/constant"
	"hcm/pkg/criteria/enumor"
	"hcm/pkg/criteria/errf"
	"hcm/pkg/iam/meta"
	"hcm/pkg/logs"
	"hcm/pkg/rest"
	"hcm/pkg/runtime/filter"
)

// AssignCvmToBiz assign cvm to biz.
func (svc *cvmSvc) AssignCvmToBiz(cts *rest.Contexts) (interface{}, error) {
	req := new(proto.AssignCvmToBizReq)
	if err := cts.DecodeInto(req); err != nil {
		return nil, err
	}
	if err := req.Validate(); err != nil {
		return nil, errf.NewFromErr(errf.InvalidParameter, err)
	}

	listReq := &dataproto.CvmListReq{
		Field: []string{"id", "bk_biz_id", "bk_cloud_id"},
		Filter: &filter.Expression{
			Op: filter.And,
			Rules: []filter.RuleFactory{
				&filter.AtomRule{Field: "id", Op: filter.In.Factory(), Value: req.CvmIDs},
			},
		},
		Page: core.NewDefaultBasePage(),
	}
	result, err := svc.client.DataService().Global.Cvm.ListCvm(cts.Kit.Ctx, cts.Kit.Header(), listReq)
	if err != nil {
		logs.Errorf("list cvm failed, err: %v, rid: %s", err, cts.Kit.Rid)
		return nil, err
	}

	accountIDMap := make(map[string]struct{}, 0)
	assignedIDs := make([]string, 0)
	unBindCloudIDs := make([]string, 0)
	for _, one := range result.Details {
		accountIDMap[one.AccountID] = struct{}{}

		if one.BkBizID != constant.UnassignedBiz {
			assignedIDs = append(assignedIDs, one.ID)
		}

		if one.BkCloudID == constant.UnbindBkCloudID {
			unBindCloudIDs = append(unBindCloudIDs, one.ID)
		}
	}

	// authorize
	authRes := make([]meta.ResourceAttribute, 0, len(accountIDMap))
	for accountID := range accountIDMap {
		authRes = append(authRes, meta.ResourceAttribute{Basic: &meta.Basic{Type: meta.Cvm,
			Action: meta.Assign, ResourceID: accountID}, BizID: req.BkBizID})
	}
	if err = svc.authorizer.AuthorizeWithPerm(cts.Kit, authRes...); err != nil {
		logs.Errorf("assign cvm failed, err: %v, req: %+v, rid: %s", err, req, cts.Kit.Rid)
		return nil, err
	}

	if len(assignedIDs) != 0 {
		return nil, fmt.Errorf("cvm(ids=%v) already assigned", assignedIDs)
	}

	if len(unBindCloudIDs) != 0 {
		return nil, fmt.Errorf("cvm(ids=%v) not bind cloud area", unBindCloudIDs)
	}

	// create assign audit.
	if err = svc.audit.ResBizAssignAudit(cts.Kit, enumor.CvmAuditResType, req.CvmIDs, req.BkBizID); err != nil {
		logs.Errorf("create assign audit failed, err: %v, rid: %s", err, cts.Kit.Rid)
		return nil, err
	}

	update := &dataproto.CvmCommonInfoBatchUpdateReq{IDs: req.CvmIDs, BkBizID: req.BkBizID}
	if err := svc.client.DataService().Global.Cvm.BatchUpdateCvmCommonInfo(cts.Kit.Ctx, cts.Kit.Header(),
		update); err != nil {

		logs.Errorf("batch update cvm common info failed, err: %v, req: %v, rid: %s", err, update,
			cts.Kit.Rid)
		return nil, err
	}

	return nil, nil
}
