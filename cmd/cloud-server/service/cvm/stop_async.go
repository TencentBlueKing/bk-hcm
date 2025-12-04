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
	proto "hcm/pkg/api/cloud-server/cvm"
	"hcm/pkg/api/data-service/audit"
	"hcm/pkg/criteria/constant"
	"hcm/pkg/criteria/enumor"
	"hcm/pkg/criteria/errf"
	"hcm/pkg/iam/meta"
	"hcm/pkg/logs"
	"hcm/pkg/rest"
	"hcm/pkg/tools/hooks/handler"
)

// BatchAsyncStopCvm batch stop cvm.
func (svc *cvmSvc) BatchAsyncStopCvm(cts *rest.Contexts) (interface{}, error) {
	req := new(proto.BatchCvmPowerOperateReq)
	if err := cts.DecodeInto(req); err != nil {
		return nil, errf.NewFromErr(errf.DecodeRequestFailed, err)
	}
	if err := req.Validate(true); err != nil {
		return nil, errf.NewFromErr(errf.InvalidParameter, err)
	}

	return svc.batchAsyncStopCvmSvc(cts, constant.UnassignedBiz, handler.ResOperateAuth, req, true)
}

// BatchAsyncStopBizCvm batch stop biz cvm.
func (svc *cvmSvc) BatchAsyncStopBizCvm(cts *rest.Contexts) (interface{}, error) {
	bizID, err := cts.PathParameter("bk_biz_id").Int64()
	if err != nil {
		return nil, err
	}

	req := new(proto.BatchCvmPowerOperateReq)
	if err = cts.DecodeInto(req); err != nil {
		return nil, errf.NewFromErr(errf.DecodeRequestFailed, err)
	}
	if err = req.Validate(true); err != nil {
		return nil, errf.NewFromErr(errf.InvalidParameter, err)
	}

	return svc.batchAsyncStopCvmSvc(cts, bizID, handler.BizOperateAuth, req, true)
}

func (svc *cvmSvc) batchAsyncStopCvmSvc(cts *rest.Contexts, bkBizID int64, validHandler handler.ValidWithAuthHandler,
	req *proto.BatchCvmPowerOperateReq, verifyMoa bool) (interface{}, error) {

	if err := svc.validateAuthorize(cts, req.IDs, meta.Stop, validHandler); err != nil {
		logs.Errorf("validate authorize and create audit failed, err: %v, rid: %s", err, cts.Kit.Rid)
		return nil, err
	}

	if verifyMoa {
		if err := svc.validateMOAResult(cts.Kit, enumor.MoaSceneCVMStop, req.SessionID); err != nil {
			logs.Errorf("validate moa result failed, err: %v, sessionID: %s, rid: %s", err, req.SessionID, cts.Kit.Rid)
			return nil, err
		}
	}
	if err := svc.createAudit(cts, audit.Stop, req.IDs); err != nil {
		logs.Errorf("create audit for %s failed, err: %v, rid: %s", audit.Stop, err, cts.Kit.Rid)
		return nil, err
	}

	cvmList, err := svc.listCvmByIDs(cts.Kit, req.IDs)
	if err != nil {
		logs.Errorf("list cvm by ids failed, ids: %v, err: %v, rid: %s", req.IDs, err, cts.Kit.Rid)
		return nil, err
	}

	uniqueID, err := calCvmResetUniqueID(cts.Kit, bkBizID, req.IDs)
	if err != nil {
		logs.Errorf("cal cvm reset unique key failed, err: %v, cvmIDs: %v, rid: %s", err, req.IDs, cts.Kit.Rid)
		return "", err
	}

	// 请求来源
	source := enumor.TaskManagementSourceAPI
	if len(req.Source) > 0 {
		source = req.Source
	}

	taskManagementID, err := svc.cvmLgc.CvmPowerOperation(cts.Kit, bkBizID, uniqueID,
		source, enumor.TaskStopCvm, cvmList)
	if err != nil {
		logs.Errorf("build flow and task management failed, err: %v, rid: %s", err, cts.Kit.Rid)
		return nil, err
	}

	return proto.BatchCvmOperateResp{
		TaskManagementID: taskManagementID,
	}, nil
}

// BatchSopsAsyncStopBizCvm batch stop biz cvm for sops.
func (svc *cvmSvc) BatchSopsAsyncStopBizCvm(cts *rest.Contexts) (interface{}, error) {
	bizID, err := cts.PathParameter("bk_biz_id").Int64()
	if err != nil {
		return nil, err
	}

	req := new(proto.BatchCvmPowerOperateReq)
	if err = cts.DecodeInto(req); err != nil {
		return nil, errf.NewFromErr(errf.DecodeRequestFailed, err)
	}

	if err = req.Validate(false); err != nil {
		return nil, errf.NewFromErr(errf.InvalidParameter, err)
	}

	return svc.batchAsyncStopCvmSvc(cts, bizID, handler.BizOperateAuth, req, false)
}
