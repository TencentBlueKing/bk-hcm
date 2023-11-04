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
	"strconv"

	"hcm/cmd/cloud-server/logics/async"
	actioncvm "hcm/cmd/task-server/logics/action/cvm"
	proto "hcm/pkg/api/cloud-server/cvm"
	protoaudit "hcm/pkg/api/data-service/audit"
	dataproto "hcm/pkg/api/data-service/cloud"
	ts "hcm/pkg/api/task-server"
	"hcm/pkg/async/action"
	"hcm/pkg/criteria/enumor"
	"hcm/pkg/criteria/errf"
	"hcm/pkg/dal/dao/types"
	"hcm/pkg/iam/meta"
	"hcm/pkg/logs"
	"hcm/pkg/rest"
	"hcm/pkg/tools/hooks/handler"
)

// BatchStartCvm batch start cvm.
func (svc *cvmSvc) BatchStartCvm(cts *rest.Contexts) (interface{}, error) {
	return svc.batchStartCvmSvc(cts, handler.ResOperateAuth)
}

// BatchStartBizCvm batch start biz cvm.
func (svc *cvmSvc) BatchStartBizCvm(cts *rest.Contexts) (interface{}, error) {
	return svc.batchStartCvmSvc(cts, handler.BizOperateAuth)
}

func (svc *cvmSvc) batchStartCvmSvc(cts *rest.Contexts, validHandler handler.ValidWithAuthHandler) (interface{},
	error) {

	req := new(proto.BatchStartCvmReq)
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
		Action: meta.Start, BasicInfos: basicInfoMap})
	if err != nil {
		return nil, err
	}

	if err = svc.audit.ResBaseOperationAudit(cts.Kit, enumor.CvmAuditResType, protoaudit.Start, req.IDs); err != nil {
		logs.Errorf("create operation audit failed, err: %v, rid: %s", err, cts.Kit.Rid)
		return nil, err
	}

	tasks, err := buildOperationTasks(enumor.ActionStartCvm, basicInfoMap)
	if err != nil {
		return nil, err
	}
	addReq := &ts.AddCustomFlowReq{
		Name:  enumor.FlowStartCvm,
		Tasks: tasks,
	}
	result, err := svc.client.TaskServer().CreateCustomFlow(cts.Kit, addReq)
	if err != nil {
		logs.Errorf("call taskserver to create custom flow failed, err: %v, rid: %s", err, cts.Kit.Rid)
		return nil, err
	}

	return result, async.WaitTaskToEnd(cts.Kit, svc.client.TaskServer(), result.ID)
}

func buildOperationTasks(actionName enumor.ActionName, basicInfoMap map[string]types.CloudResourceBasicInfo) (
	[]ts.CustomFlowTask, error) {

	paramMaps := make(map[string]*actioncvm.CvmOperationOption)
	for _, info := range basicInfoMap {
		switch info.Vendor {
		case enumor.TCloud, enumor.Aws, enumor.HuaWei:
			key := info.AccountID + "_" + info.Region
			_, exist := paramMaps[key]
			if !exist {
				paramMaps[key] = &actioncvm.CvmOperationOption{
					Vendor:    info.Vendor,
					AccountID: info.AccountID,
					Region:    info.Region,
					IDs:       make([]string, 0),
				}
			}
			paramMaps[key].IDs = append(paramMaps[key].IDs, info.ID)

		case enumor.Azure, enumor.Gcp:
			paramMaps[info.AccountID+"_"+info.ID] = &actioncvm.CvmOperationOption{
				Vendor:    info.Vendor,
				AccountID: info.AccountID,
				IDs:       []string{info.ID},
			}

		default:
			return nil, fmt.Errorf("vendor: %s not support", info.Vendor)
		}
	}

	tasks := make([]ts.CustomFlowTask, 0, len(paramMaps))
	count := 1
	for _, one := range paramMaps {
		tasks = append(tasks, ts.CustomFlowTask{
			ActionID:   action.ActIDType(strconv.Itoa(count)),
			ActionName: actionName,
			Params:     *one,
			DependOn:   nil,
		})
		count++
	}
	return tasks, nil
}
