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
	"hcm/cmd/cloud-server/logics/async"
	proto "hcm/pkg/api/cloud-server"
	dataproto "hcm/pkg/api/data-service/cloud"
	ts "hcm/pkg/api/task-server"
	"hcm/pkg/criteria/enumor"
	"hcm/pkg/criteria/errf"
	"hcm/pkg/dal/dao/types"
	"hcm/pkg/iam/meta"
	"hcm/pkg/logs"
	"hcm/pkg/rest"
	"hcm/pkg/tools/hooks/handler"
)

// BatchDeleteCvm batch delete cvm.
func (svc *cvmSvc) BatchDeleteCvm(cts *rest.Contexts) (interface{}, error) {
	return svc.batchDeleteCvmSvc(cts, handler.ResOperateAuth)
}

// BatchDeleteBizCvm batch delete biz cvm.
func (svc *cvmSvc) BatchDeleteBizCvm(cts *rest.Contexts) (interface{}, error) {
	return svc.batchDeleteCvmSvc(cts, handler.BizOperateAuth)
}

func (svc *cvmSvc) batchDeleteCvmSvc(cts *rest.Contexts, validHandler handler.ValidWithAuthHandler) (interface{},
	error) {

	req := new(proto.BatchDeleteReq)
	if err := cts.DecodeInto(req); err != nil {
		return nil, errf.NewFromErr(errf.DecodeRequestFailed, err)
	}

	if err := req.Validate(); err != nil {
		return nil, errf.NewFromErr(errf.InvalidParameter, err)
	}

	basicInfoReq := dataproto.ListResourceBasicInfoReq{
		ResourceType: enumor.CvmCloudResType,
		IDs:          req.IDs,
		Fields:       append(types.CommonBasicInfoFields, "region", "recycle_status"),
	}
	basicInfoMap, err := svc.client.DataService().Global.Cloud.ListResBasicInfo(cts.Kit, basicInfoReq)
	if err != nil {
		return nil, err
	}

	// validate biz and authorize
	err = validHandler(cts, &handler.ValidWithAuthOption{Authorizer: svc.authorizer, ResType: meta.Cvm,
		Action: meta.Delete, BasicInfos: basicInfoMap})
	if err != nil {
		return nil, err
	}

	if err = svc.audit.ResDeleteAudit(cts.Kit, enumor.CvmAuditResType, req.IDs); err != nil {
		logs.Errorf("create operation audit failed, err: %v, rid: %s", err, cts.Kit.Rid)
		return nil, err
	}

	tasks, err := buildOperationTasks(enumor.ActionDeleteCvm, basicInfoMap)
	if err != nil {
		return nil, err
	}
	addReq := &ts.AddCustomFlowReq{
		Name:  enumor.FlowDeleteCvm,
		Tasks: tasks,
	}
	result, err := svc.client.TaskServer().CreateCustomFlow(cts.Kit, addReq)
	if err != nil {
		logs.Errorf("call taskserver to create custom flow failed, err: %v, rid: %s", err, cts.Kit.Rid)
		return nil, err
	}

	return result, async.WaitTaskToEnd(cts.Kit, svc.client.TaskServer(), result.ID)
}
