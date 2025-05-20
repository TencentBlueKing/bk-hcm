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
	"encoding/json"
	"fmt"

	"hcm/cmd/cloud-server/logics/async"
	"hcm/cmd/cloud-server/service/common"
	actioncvm "hcm/cmd/task-server/logics/action/cvm"
	cloudserver "hcm/pkg/api/cloud-server"
	cscvm "hcm/pkg/api/cloud-server/cvm"
	ts "hcm/pkg/api/task-server"
	"hcm/pkg/async/action"
	"hcm/pkg/criteria/constant"
	"hcm/pkg/criteria/enumor"
	"hcm/pkg/criteria/errf"
	"hcm/pkg/iam/meta"
	"hcm/pkg/logs"
	"hcm/pkg/rest"
	"hcm/pkg/tools/converter"
)

// CreateCvm create cvm.
func (svc *cvmSvc) CreateCvm(cts *rest.Contexts) (interface{}, error) {

	req := new(cloudserver.ResourceCreateReq)
	if err := cts.DecodeInto(req); err != nil {
		logs.Errorf("create cvm request decode failed, err: %v, rid: %s", err, cts.Kit.Rid)
		return nil, errf.NewFromErr(errf.DecodeRequestFailed, err)
	}

	authRes := meta.ResourceAttribute{Basic: &meta.Basic{Type: meta.Cvm, Action: meta.Create,
		ResourceID: req.AccountID}}
	if err := svc.authorizer.AuthorizeWithPerm(cts.Kit, authRes); err != nil {
		logs.Errorf("create cvm auth failed, err: %v, rid: %s", err, cts.Kit.Rid)
		return nil, err
	}

	info, err := svc.client.DataService().Global.Cloud.GetResBasicInfo(cts.Kit,
		enumor.AccountCloudResType, req.AccountID)
	if err != nil {
		logs.Errorf("get account basic info failed, err: %v, rid: %s", err, cts.Kit.Rid)
		return nil, err
	}

	tasks := make([]ts.CustomFlowTask, 0)
	switch info.Vendor {
	case enumor.TCloud:
		tasks, err = svc.buildCreateTCloudCvmTasks(req.Data)
	case enumor.Aws:
		tasks, err = svc.buildCreateAwsCvmTasks(req.Data)
	case enumor.HuaWei:
		tasks, err = svc.buildCreateHuaWeiCvmTasks(req.Data)
	case enumor.Gcp:
		tasks, err = svc.buildCreateGcpCvmTasks(req.Data)
	case enumor.Azure:
		tasks, err = svc.buildCreateAzureCvmTasks(req.Data)
	default:
		return nil, fmt.Errorf("vendor: %s not support", info.Vendor)
	}
	if err != nil {
		logs.Errorf("build create %s cvm tasks failed, err: %v, rid: %s", info.Vendor, err, cts.Kit.Rid)
		return nil, err
	}

	addReq := &ts.AddCustomFlowReq{
		Name:  enumor.FlowCreateCvm,
		Tasks: tasks,
	}
	result, err := svc.client.TaskServer().CreateCustomFlow(cts.Kit, addReq)
	if err != nil {
		logs.Errorf("call taskserver to create custom flow failed, err: %v, rid: %s", err, cts.Kit.Rid)
		return nil, err
	}

	return result, async.WaitTaskToEnd(cts.Kit, svc.client.TaskServer(), result.ID)
}

func (svc *cvmSvc) buildCreateAzureCvmTasks(body json.RawMessage) ([]ts.CustomFlowTask, error) {

	req := new(cscvm.AzureCvmCreateReq)
	if err := json.Unmarshal(body, req); err != nil {
		return nil, errf.NewFromErr(errf.DecodeRequestFailed, err)
	}

	if err := req.Validate(false); err != nil {
		return nil, errf.NewFromErr(errf.InvalidParameter, err)
	}

	opt := &actioncvm.AssignCvmOption{
		BizID:     constant.UnassignedBiz,
		BkCloudID: converter.ValToPtr(constant.UnassignedBkCloudID),
	}
	tasks := actioncvm.BuildCreateCvmTasks(req.RequiredCount, 1, opt,
		func(actionID action.ActIDType, count int64) ts.CustomFlowTask {
			req.RequiredCount = count
			return ts.CustomFlowTask{
				ActionID:   actionID,
				ActionName: enumor.ActionCreateCvm,
				Params: &actioncvm.CreateOption{
					Vendor:         enumor.Azure,
					AzureCreateReq: *common.ConvAzureCvmCreateReq(req),
				},
			}
		})

	return tasks, nil
}

func (svc *cvmSvc) buildCreateHuaWeiCvmTasks(body json.RawMessage) ([]ts.CustomFlowTask, error) {

	req := new(cscvm.HuaWeiCvmCreateReq)
	if err := json.Unmarshal(body, req); err != nil {
		return nil, errf.NewFromErr(errf.DecodeRequestFailed, err)
	}

	if err := req.Validate(false); err != nil {
		return nil, errf.NewFromErr(errf.InvalidParameter, err)
	}

	opt := &actioncvm.AssignCvmOption{
		BizID:     constant.UnassignedBiz,
		BkCloudID: converter.ValToPtr(constant.UnassignedBkCloudID),
	}
	tasks := actioncvm.BuildCreateCvmTasks(req.RequiredCount, constant.BatchCreateCvmFromCloudMaxLimit, opt,
		func(actionID action.ActIDType, count int64) ts.CustomFlowTask {
			req.RequiredCount = count
			return ts.CustomFlowTask{
				ActionID:   actionID,
				ActionName: enumor.ActionCreateCvm,
				Params: &actioncvm.CreateOption{
					Vendor:               enumor.HuaWei,
					HuaWeiBatchCreateReq: *common.ConvHuaWeiCvmCreateReq(req),
				},
			}
		})

	return tasks, nil
}

func (svc *cvmSvc) buildCreateGcpCvmTasks(body json.RawMessage) ([]ts.CustomFlowTask, error) {

	req := new(cscvm.GcpCvmCreateReq)
	if err := json.Unmarshal(body, req); err != nil {
		return nil, errf.NewFromErr(errf.DecodeRequestFailed, err)
	}

	if err := req.Validate(false); err != nil {
		return nil, errf.NewFromErr(errf.InvalidParameter, err)
	}

	opt := &actioncvm.AssignCvmOption{
		BizID:     constant.UnassignedBiz,
		BkCloudID: converter.ValToPtr(constant.UnassignedBkCloudID),
	}
	tasks := actioncvm.BuildCreateCvmTasks(req.RequiredCount, constant.BatchCreateCvmFromCloudMaxLimit, opt,
		func(actionID action.ActIDType, count int64) ts.CustomFlowTask {
			req.RequiredCount = count
			return ts.CustomFlowTask{
				ActionID:   actionID,
				ActionName: enumor.ActionCreateCvm,
				Params: &actioncvm.CreateOption{
					Vendor:            enumor.Gcp,
					GcpBatchCreateReq: *common.ConvGcpCvmCreateReq(req),
				},
			}
		})

	return tasks, nil
}

func (svc *cvmSvc) buildCreateAwsCvmTasks(body json.RawMessage) ([]ts.CustomFlowTask, error) {

	req := new(cscvm.AwsCvmCreateReq)
	if err := json.Unmarshal(body, req); err != nil {
		return nil, errf.NewFromErr(errf.DecodeRequestFailed, err)
	}

	if err := req.Validate(false); err != nil {
		return nil, errf.NewFromErr(errf.InvalidParameter, err)
	}

	opt := &actioncvm.AssignCvmOption{
		BizID:     constant.UnassignedBiz,
		BkCloudID: converter.ValToPtr(constant.UnassignedBkCloudID),
	}
	tasks := actioncvm.BuildCreateCvmTasks(req.RequiredCount, constant.BatchCreateCvmFromCloudMaxLimit, opt,
		func(actionID action.ActIDType, count int64) ts.CustomFlowTask {
			req.RequiredCount = count
			return ts.CustomFlowTask{
				ActionID:   actionID,
				ActionName: enumor.ActionCreateCvm,
				Params: &actioncvm.CreateOption{
					Vendor:            enumor.Aws,
					AwsBatchCreateReq: *common.ConvAwsCvmCreateReq(req),
				},
			}
		})

	return tasks, nil
}

func (svc *cvmSvc) buildCreateTCloudCvmTasks(body json.RawMessage) ([]ts.CustomFlowTask, error) {

	req := new(cscvm.TCloudCvmCreateReq)
	if err := json.Unmarshal(body, req); err != nil {
		return nil, errf.NewFromErr(errf.DecodeRequestFailed, err)
	}

	if err := req.Validate(false); err != nil {
		return nil, errf.NewFromErr(errf.InvalidParameter, err)
	}

	opt := &actioncvm.AssignCvmOption{
		BizID:     constant.UnassignedBiz,
		BkCloudID: converter.ValToPtr(constant.UnassignedBkCloudID),
	}
	tasks := actioncvm.BuildCreateCvmTasks(req.RequiredCount, constant.BatchCreateCvmFromCloudMaxLimit, opt,
		func(actionID action.ActIDType, count int64) ts.CustomFlowTask {
			req.RequiredCount = count
			return ts.CustomFlowTask{
				ActionID:   actionID,
				ActionName: enumor.ActionCreateCvm,
				Params: &actioncvm.CreateOption{
					Vendor:               enumor.TCloud,
					TCloudBatchCreateReq: *common.ConvTCloudCvmCreateReq(req),
				},
			}
		})

	return tasks, nil
}
