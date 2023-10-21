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

package securitygroup

import (
	"hcm/cmd/cloud-server/logics/async"
	actionsg "hcm/cmd/task-server/logics/action/security-group"
	proto "hcm/pkg/api/cloud-server"
	dataproto "hcm/pkg/api/data-service/cloud"
	ts "hcm/pkg/api/task-server"
	"hcm/pkg/async/action"
	"hcm/pkg/criteria/enumor"
	"hcm/pkg/criteria/errf"
	"hcm/pkg/iam/meta"
	"hcm/pkg/logs"
	"hcm/pkg/rest"
	"hcm/pkg/tools/counter"
	"hcm/pkg/tools/hooks/handler"
)

// BatchDeleteSecurityGroup batch delete security group.
func (svc *securityGroupSvc) BatchDeleteSecurityGroup(cts *rest.Contexts) (interface{}, error) {
	return svc.batchDeleteSecurityGroup(cts, handler.ResValidWithAuth)
}

// BatchDeleteBizSecurityGroup batch delete biz security group.
func (svc *securityGroupSvc) BatchDeleteBizSecurityGroup(cts *rest.Contexts) (interface{}, error) {
	return svc.batchDeleteSecurityGroup(cts, handler.BizValidWithAuth)
}

func (svc *securityGroupSvc) batchDeleteSecurityGroup(cts *rest.Contexts, validHandler handler.ValidWithAuthHandler) (
	interface{}, error) {

	req := new(proto.SecurityGroupBatchDeleteReq)
	if err := cts.DecodeInto(req); err != nil {
		return nil, err
	}

	if err := req.Validate(); err != nil {
		return nil, errf.NewFromErr(errf.InvalidParameter, err)
	}

	basicInfoReq := dataproto.ListResourceBasicInfoReq{
		ResourceType: enumor.SecurityGroupCloudResType,
		IDs:          req.IDs,
	}
	basicInfoMap, err := svc.client.DataService().Global.Cloud.ListResourceBasicInfo(cts.Kit.Ctx, cts.Kit.Header(),
		basicInfoReq)
	if err != nil {
		return nil, err
	}

	// validate biz and authorize
	err = validHandler(cts, &handler.ValidWithAuthOption{Authorizer: svc.authorizer, ResType: meta.SecurityGroup,
		Action: meta.Delete, BasicInfos: basicInfoMap})
	if err != nil {
		return nil, err
	}

	// create delete audit.
	if err := svc.audit.ResDeleteAudit(cts.Kit, enumor.SecurityGroupAuditResType, req.IDs); err != nil {
		logs.Errorf("create delete audit failed, err: %v, rid: %s", err, cts.Kit.Rid)
		return nil, err
	}

	tasks := make([]ts.CustomFlowTask, 0, len(req.IDs))

	nextID := counter.NewNumStringCounter(1, 10)
	for _, info := range basicInfoMap {
		tasks = append(tasks, ts.CustomFlowTask{
			ActionID:   action.ActIDType(nextID()),
			ActionName: enumor.ActionDeleteSecurityGroup,
			Params: actionsg.DeleteSGOption{
				Vendor: info.Vendor,
				ID:     info.ID,
			},
			DependOn: nil,
		})
	}
	flowReq := &ts.AddCustomFlowReq{
		Name:  enumor.FlowDeleteSecurityGroup,
		Tasks: tasks,
	}

	result, err := svc.client.TaskServer().CreateCustomFlow(cts.Kit, flowReq)
	if err != nil {
		logs.Errorf("call taskserver to create custom flow failed, err: %v, rid: %s", err, cts.Kit.Rid)
		return nil, err
	}

	if err := async.WaitTaskToEnd(cts.Kit, svc.client.TaskServer(), result.ID); err != nil {
		return nil, err
	}
	return result, nil
}
