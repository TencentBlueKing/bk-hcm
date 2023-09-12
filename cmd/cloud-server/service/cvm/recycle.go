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
 */

package cvm

import (
	"fmt"

	proto "hcm/pkg/api/cloud-server/cvm"
	"hcm/pkg/api/cloud-server/recycle"
	"hcm/pkg/api/core"
	protoaudit "hcm/pkg/api/data-service/audit"
	"hcm/pkg/api/data-service/cloud"
	recyclerecord "hcm/pkg/api/data-service/recycle-record"
	"hcm/pkg/criteria/constant"
	"hcm/pkg/criteria/enumor"
	"hcm/pkg/criteria/errf"
	"hcm/pkg/dal/dao/tools"
	"hcm/pkg/dal/dao/types"
	"hcm/pkg/iam/meta"
	"hcm/pkg/kit"
	"hcm/pkg/logs"
	"hcm/pkg/rest"
	"hcm/pkg/runtime/filter"
	"hcm/pkg/tools/hooks/handler"
)

// RecycleCvm recycle cvm.
func (svc *cvmSvc) RecycleCvm(cts *rest.Contexts) (interface{}, error) {
	return svc.recycleCvmSvc(cts, handler.ResValidWithAuth)
}

// RecycleBizCvm recycle biz cvm.
func (svc *cvmSvc) RecycleBizCvm(cts *rest.Contexts) (interface{}, error) {
	return svc.recycleCvmSvc(cts, handler.BizValidWithAuth)
}

func (svc *cvmSvc) recycleCvmSvc(cts *rest.Contexts, validHandler handler.ValidWithAuthHandler) (interface{}, error) {
	req := new(proto.CvmRecycleReq)
	if err := cts.DecodeInto(req); err != nil {
		return nil, err
	}

	if err := req.Validate(); err != nil {
		return nil, errf.NewFromErr(errf.InvalidParameter, err)
	}

	ids := make([]string, 0, len(req.Infos))
	auditInfos := make([]protoaudit.CloudResRecycleAuditInfo, 0, len(req.Infos))
	for _, info := range req.Infos {
		ids = append(ids, info.ID)
		auditInfos = append(auditInfos, protoaudit.CloudResRecycleAuditInfo{ResID: info.ID,
			Data: info.CvmRecycleOptions})
	}

	basicInfoReq := cloud.ListResourceBasicInfoReq{
		ResourceType: enumor.CvmCloudResType,
		IDs:          ids,
		Fields:       append(types.CommonBasicInfoFields, "region", "recycle_status"),
	}
	basicInfoMap, err := svc.client.DataService().Global.Cloud.ListResourceBasicInfo(cts.Kit.Ctx, cts.Kit.Header(),
		basicInfoReq)
	if err != nil {
		return nil, err
	}

	// validate biz and authorize
	err = validHandler(cts, &handler.ValidWithAuthOption{Authorizer: svc.authorizer, ResType: meta.Cvm,
		Action: meta.Recycle, BasicInfos: basicInfoMap})
	if err != nil {
		return nil, err
	}

	// create recycle audit
	auditReq := &protoaudit.CloudResourceRecycleAuditReq{
		ResType: enumor.CvmAuditResType,
		Action:  protoaudit.Recycle,
		Infos:   auditInfos,
	}
	if err = svc.audit.ResRecycleAudit(cts.Kit, auditReq); err != nil {
		logs.Errorf("create recycle audit failed, err: %v, rid: %s", err, cts.Kit.Rid)
		return nil, err
	}

	return svc.recycleCvm(cts.Kit, req, basicInfoMap)
}

func (svc *cvmSvc) recycleCvm(kt *kit.Kit, req *proto.CvmRecycleReq, infoMap map[string]types.CloudResourceBasicInfo) (
	interface{}, error) {

	// filter out stopped cvm
	ids := make([]string, 0, len(infoMap))
	for id := range infoMap {
		ids = append(ids, id)
	}

	notStoppedRule := filter.AtomRule{Field: "status", Op: filter.NotIn.Factory(), Value: []string{"STOPPING",
		"STOPPED", "stopping", "stopped", "SUSPENDING", "SUSPENDED", "PowerState/stopped", "SHUTOFF"}}
	notStoppedFilter, err := tools.And(tools.ContainersExpression("id", ids), notStoppedRule)
	if err != nil {
		return nil, err
	}

	notStoppedReq := &cloud.CvmListReq{Field: []string{"id"}, Filter: notStoppedFilter, Page: core.NewDefaultBasePage()}
	startCvmRes, err := svc.client.DataService().Global.Cvm.ListCvm(kt.Ctx, kt.Header(), notStoppedReq)
	if err != nil {
		return nil, err
	}

	notStoppedMap := make(map[string]types.CloudResourceBasicInfo)
	for _, cvm := range startCvmRes.Details {
		notStoppedMap[cvm.ID] = infoMap[cvm.ID]
	}

	// stop cvm
	stopRes, err := svc.cvmLgc.BatchStopCvm(kt, notStoppedMap)
	if err != nil {
		logs.Errorf("stop cvm failed, err: %v, resp: %+v, infos: %+v, rid: %s", err, stopRes, notStoppedMap, kt.Rid)
		return nil, err
	}

	// detach disk if needed
	detachDiskCvmIDs := make([]string, 0)
	for _, info := range req.Infos {
		if info.CvmRecycleOptions != nil && !info.CvmRecycleOptions.WithDisk {
			detachDiskCvmIDs = append(detachDiskCvmIDs, info.ID)
		}
	}
	detachRes, err := svc.detachDiskByCvmIDs(kt, detachDiskCvmIDs, infoMap)
	if err != nil {
		return nil, err
	}

	res := new(core.BatchOperateAllResult)
	failedIDMap := make(map[string]struct{})
	if detachRes != nil {
		if len(detachRes.Failed) == len(detachDiskCvmIDs) {
			return res, res.Failed[0].Error
		}

		res.Failed = detachRes.Failed
		for _, info := range detachRes.Failed {
			failedIDMap[info.ID] = struct{}{}
		}
	}

	// create recycle record
	opt := &recyclerecord.BatchRecycleReq{ResType: enumor.CvmCloudResType, Infos: make([]recyclerecord.RecycleReq, 0)}
	for _, info := range req.Infos {
		if _, exists := failedIDMap[info.ID]; exists {
			continue
		}
		opt.Infos = append(opt.Infos, recyclerecord.RecycleReq{ID: info.ID, Detail: info.CvmRecycleOptions})
	}
	taskID, err := svc.client.DataService().Global.RecycleRecord.BatchRecycleCloudRes(kt.Ctx, kt.Header(), opt)
	if err != nil {
		return nil, err
	}

	if len(res.Failed) > 0 {
		return res, res.Failed[0].Error
	}
	return &recycle.RecycleResult{TaskID: taskID}, nil
}

func (svc *cvmSvc) detachDiskByCvmIDs(kt *kit.Kit, ids []string, basicInfoMap map[string]types.CloudResourceBasicInfo) (
	*core.BatchOperateAllResult, error) {

	if len(ids) == 0 {
		return nil, nil
	}

	if len(ids) > constant.BatchOperationMaxLimit {
		return nil, errf.Newf(errf.InvalidParameter, "ids should <= %d", constant.BatchOperationMaxLimit)
	}

	listReq := &cloud.DiskCvmRelListReq{
		Filter: tools.ContainersExpression("cvm_id", ids),
		Page:   core.NewDefaultBasePage(),
	}
	relRes, err := svc.client.DataService().Global.ListDiskCvmRel(kt.Ctx, kt.Header(), listReq)
	if err != nil {
		return nil, err
	}

	if len(relRes.Details) == 0 {
		return nil, nil
	}

	res := &core.BatchOperateAllResult{
		Succeeded: make([]string, 0),
		Failed:    make([]core.FailedInfo, 0),
	}

	cvmDiskMap := make(map[string]string)
	for _, detail := range relRes.Details {
		cvmDiskMap[detail.CvmID] = detail.DiskID
	}

	for _, id := range ids {
		diskID, exists := cvmDiskMap[id]
		if !exists {
			res.Succeeded = append(res.Succeeded, id)
			continue
		}

		info, exists := basicInfoMap[id]
		if !exists {
			res.Succeeded = append(res.Succeeded, id)
			continue
		}

		err = svc.diskLgc.DetachDisk(kt, info.Vendor, id, diskID, info.AccountID)
		if err != nil {
			res.Failed = append(res.Failed, core.FailedInfo{ID: id, Error: err})
			continue
		}
		res.Succeeded = append(res.Succeeded, id)
	}

	if len(res.Failed) > 0 {
		return res, res.Failed[0].Error
	}
	return res, nil
}

// RecoverCvm recover cvm.
func (svc *cvmSvc) RecoverCvm(cts *rest.Contexts) (interface{}, error) {
	req := new(proto.CvmRecoverReq)
	if err := cts.DecodeInto(req); err != nil {
		return nil, err
	}

	if err := req.Validate(); err != nil {
		return nil, errf.NewFromErr(errf.InvalidParameter, err)
	}

	listReq := &core.ListReq{
		Filter: tools.ContainersExpression("id", req.RecordIDs),
		Page:   &core.BasePage{Limit: constant.BatchOperationMaxLimit},
	}
	records, err := svc.client.DataService().Global.RecycleRecord.ListRecycleRecord(cts.Kit.Ctx, cts.Kit.Header(),
		listReq)
	if err != nil {
		return nil, err
	}

	if len(records.Details) != len(req.RecordIDs) {
		return nil, errf.New(errf.InvalidParameter, "some record_ids are not in recycle bin")
	}

	if err = svc.validateRecycleRecord(records); err != nil {
		return nil, err
	}

	// authorize
	authRes := make([]meta.ResourceAttribute, 0, len(records.Details))
	auditInfos := make([]protoaudit.CloudResRecycleAuditInfo, 0, len(records.Details))
	for _, record := range records.Details {
		authRes = append(authRes, meta.ResourceAttribute{Basic: &meta.Basic{Type: meta.RecycleBin, Action: meta.Recover,
			ResourceID: record.AccountID}, BizID: record.BkBizID})
		auditInfos = append(auditInfos, protoaudit.CloudResRecycleAuditInfo{ResID: record.ResID, Data: record.Detail})
	}
	err = svc.authorizer.AuthorizeWithPerm(cts.Kit, authRes...)
	if err != nil {
		return nil, err
	}

	// create recover audit
	auditReq := &protoaudit.CloudResourceRecycleAuditReq{
		ResType: enumor.CvmAuditResType,
		Action:  protoaudit.Recover,
		Infos:   auditInfos,
	}
	if err = svc.audit.ResRecycleAudit(cts.Kit, auditReq); err != nil {
		logs.Errorf("create recycle audit failed, err: %v, rid: %s", err, cts.Kit.Rid)
		return nil, err
	}

	opt := &recyclerecord.BatchRecoverReq{
		ResType:   enumor.CvmCloudResType,
		RecordIDs: req.RecordIDs,
	}
	err = svc.client.DataService().Global.RecycleRecord.BatchRecoverCloudResource(cts.Kit.Ctx, cts.Kit.Header(), opt)
	if err != nil {
		return nil, err
	}

	return nil, nil
}

// validateRecycleRecord 只能批量处理处于同一个回收任务的且是等待回收的记录。
func (svc *cvmSvc) validateRecycleRecord(records *recyclerecord.ListResult) error {
	taskID := ""
	for _, one := range records.Details {
		if len(taskID) == 0 {
			taskID = one.TaskID
		} else if taskID != one.TaskID {
			return fmt.Errorf("only cvms in one task can be reclaimed at the same time")
		}

		if one.Status != enumor.WaitingRecycleRecordStatus {
			return fmt.Errorf("record: %d not is wait_recycle status", one.ID)
		}

		if one.ResType != enumor.CvmCloudResType {
			return fmt.Errorf("record: %d not is cvm recycle record", one.ID)
		}
	}

	return nil
}

// BatchDeleteRecycledCvm batch delete recycled cvm.
func (svc *cvmSvc) BatchDeleteRecycledCvm(cts *rest.Contexts) (interface{}, error) {
	req := new(proto.CvmDeleteRecycledReq)
	if err := cts.DecodeInto(req); err != nil {
		return nil, err
	}

	if err := req.Validate(); err != nil {
		return nil, errf.NewFromErr(errf.InvalidParameter, err)
	}

	listReq := &core.ListReq{
		Filter: tools.ContainersExpression("id", req.RecordIDs),
		Page:   &core.BasePage{Limit: constant.BatchOperationMaxLimit},
	}
	records, err := svc.client.DataService().Global.RecycleRecord.ListRecycleRecord(cts.Kit.Ctx, cts.Kit.Header(),
		listReq)
	if err != nil {
		return nil, err
	}

	if len(records.Details) != len(req.RecordIDs) {
		return nil, errf.New(errf.InvalidParameter, "some record_ids are not in recycle bin")
	}

	if err = svc.validateRecycleRecord(records); err != nil {
		return nil, err
	}

	cvmIDs := make([]string, 0, len(records.Details))
	for _, one := range records.Details {
		cvmIDs = append(cvmIDs, one.ResID)
	}

	basicInfoReq := cloud.ListResourceBasicInfoReq{
		ResourceType: enumor.CvmCloudResType,
		IDs:          cvmIDs,
		Fields:       append(types.CommonBasicInfoFields, "region", "recycle_status"),
	}
	basicInfoMap, err := svc.client.DataService().Global.Cloud.ListResourceBasicInfo(cts.Kit.Ctx, cts.Kit.Header(),
		basicInfoReq)
	if err != nil {
		return nil, err
	}

	// validate biz and authorize
	err = handler.RecycleValidWithAuth(cts, &handler.ValidWithAuthOption{Authorizer: svc.authorizer, ResType: meta.Cvm,
		Action: meta.Recycle, BasicInfos: basicInfoMap})
	if err != nil {
		return nil, err
	}

	delRes, err := svc.cvmLgc.DeleteRecycledCvm(cts.Kit, basicInfoMap)
	if err != nil {
		return delRes, err
	}

	updateReq := &recyclerecord.BatchUpdateReq{
		Data: make([]recyclerecord.UpdateReq, 0, len(req.RecordIDs)),
	}
	for _, id := range req.RecordIDs {
		updateReq.Data = append(updateReq.Data, recyclerecord.UpdateReq{
			ID:     id,
			Status: enumor.RecycledRecycleRecordStatus,
		})
	}
	err = svc.client.DataService().Global.RecycleRecord.BatchUpdateRecycleRecord(cts.Kit.Ctx, cts.Kit.Header(),
		updateReq)
	if err != nil {
		logs.Errorf("update recycle record status to recycled failed, err: %v, ids: %v, rid: %s", err, req.RecordIDs,
			cts.Kit.Rid)
		return nil, err
	}

	return nil, nil
}
