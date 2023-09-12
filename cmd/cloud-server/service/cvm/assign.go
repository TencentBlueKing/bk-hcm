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
	"hcm/pkg/dal/dao/types"
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

	// authorize
	basicInfoReq := dataproto.ListResourceBasicInfoReq{
		ResourceType: enumor.CvmCloudResType,
		IDs:          req.CvmIDs,
		Fields:       append(types.CommonBasicInfoFields, "region", "recycle_status"),
	}
	basicInfoMap, err := svc.client.DataService().Global.Cloud.ListResourceBasicInfo(cts.Kit.Ctx, cts.Kit.Header(),
		basicInfoReq)
	if err != nil {
		return nil, err
	}

	authRes := make([]meta.ResourceAttribute, 0, len(basicInfoMap))
	for _, info := range basicInfoMap {
		authRes = append(authRes, meta.ResourceAttribute{Basic: &meta.Basic{Type: meta.Cvm,
			Action: meta.Assign, ResourceID: info.AccountID}, BizID: info.BkBizID})
	}
	err = svc.authorizer.AuthorizeWithPerm(cts.Kit, authRes...)
	if err != nil {
		return nil, err
	}

	listReq := &dataproto.CvmListReq{
		Field: []string{"id"},
		Filter: &filter.Expression{
			Op: filter.And,
			Rules: []filter.RuleFactory{
				&filter.AtomRule{Field: "id", Op: filter.In.Factory(), Value: req.CvmIDs},
				&filter.AtomRule{Field: "bk_biz_id", Op: filter.NotEqual.Factory(), Value: constant.UnassignedBiz},
			},
		},
		Page: core.NewDefaultBasePage(),
	}
	result, err := svc.client.DataService().Global.Cvm.ListCvm(cts.Kit.Ctx, cts.Kit.Header(), listReq)
	if err != nil {
		logs.Errorf("list cvm failed, err: %v, rid: %s", err, cts.Kit.Rid)
		return nil, err
	}
	if len(result.Details) != 0 {
		ids := make([]string, len(result.Details))
		for index, one := range result.Details {
			ids[index] = one.ID
		}
		return nil, fmt.Errorf("cvm(ids=%v) already assigned", ids)
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
