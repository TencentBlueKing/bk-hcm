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

package disk

import (
	"fmt"

	csdisk "hcm/pkg/api/cloud-server/disk"
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
	"hcm/pkg/tools/hooks/handler"
)

// RecycleDisk recycle disk.
func (svc *diskSvc) RecycleDisk(cts *rest.Contexts) (interface{}, error) {
	return svc.recycleDiskSvc(cts, handler.ResValidWithAuth)
}

// RecycleBizDisk recycle biz disk.
func (svc *diskSvc) RecycleBizDisk(cts *rest.Contexts) (interface{}, error) {
	return svc.recycleDiskSvc(cts, handler.BizValidWithAuth)
}

func (svc *diskSvc) recycleDiskSvc(cts *rest.Contexts, validHandler handler.ValidWithAuthHandler) (interface{}, error) {
	req := new(csdisk.DiskRecycleReq)
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
		auditInfos = append(auditInfos, protoaudit.CloudResRecycleAuditInfo{ResID: info.ID, Data: info.DiskRecycleOptions})
	}

	basicInfoReq := cloud.ListResourceBasicInfoReq{
		ResourceType: enumor.DiskCloudResType,
		IDs:          ids,
		Fields:       append(types.CommonBasicInfoFields, "recycle_status"),
	}
	basicInfoMap, err := svc.client.DataService().Global.Cloud.ListResourceBasicInfo(cts.Kit.Ctx, cts.Kit.Header(),
		basicInfoReq)
	if err != nil {
		return nil, err
	}

	// validate biz and authorize
	err = validHandler(cts, &handler.ValidWithAuthOption{Authorizer: svc.authorizer, ResType: meta.Disk,
		Action: meta.Recycle, BasicInfos: basicInfoMap})
	if err != nil {
		return nil, err
	}

	// create recycle audit
	auditReq := &protoaudit.CloudResourceRecycleAuditReq{
		ResType: enumor.DiskAuditResType,
		Action:  protoaudit.Recycle,
		Infos:   auditInfos,
	}
	if err = svc.audit.ResRecycleAudit(cts.Kit, auditReq); err != nil {
		logs.Errorf("create recycle audit failed, err: %v, rid: %s", err, cts.Kit.Rid)
		return nil, err
	}

	return svc.recycleDisk(cts.Kit, req, ids, basicInfoMap)
}

func (svc *diskSvc) recycleDisk(kt *kit.Kit, req *csdisk.DiskRecycleReq, ids []string,
	infoMap map[string]types.CloudResourceBasicInfo) (interface{}, error) {

	// detach disk from cvm
	detachRes, err := svc.detachDiskByIDs(kt, ids, infoMap)
	if err != nil {
		logs.Errorf("detach disks failed, err: %v, ids: +v, result: %+v, rid: %s", err, ids, detachRes, kt.Rid)
		return nil, err
	}

	res := new(core.BatchOperateAllResult)

	failedIDMap := make(map[string]struct{})
	if detachRes != nil {
		if len(detachRes.Failed) == len(ids) {
			return res, res.Failed[0].Error
		}

		res.Failed = detachRes.Failed
		for _, info := range detachRes.Failed {
			failedIDMap[info.ID] = struct{}{}
		}
	}

	// create recycle record
	opt := &recyclerecord.BatchRecycleReq{
		ResType: enumor.DiskCloudResType,
		Infos:   make([]recyclerecord.RecycleReq, 0),
	}
	for _, info := range req.Infos {
		if _, exists := failedIDMap[info.ID]; exists {
			continue
		}
		opt.Infos = append(opt.Infos, recyclerecord.RecycleReq{
			ID:     info.ID,
			Detail: info.DiskRecycleOptions,
		})
	}

	taskID, err := svc.client.DataService().Global.RecycleRecord.BatchRecycleCloudRes(kt.Ctx, kt.Header(),
		opt)
	if err != nil {
		for _, id := range detachRes.Succeeded {
			res.Failed = append(res.Failed, core.FailedInfo{ID: id, Error: err})
		}
		return res, err
	}

	if len(res.Failed) > 0 {
		return res, res.Failed[0].Error
	}
	return &recycle.RecycleResult{TaskID: taskID}, nil
}

func (svc *diskSvc) detachDiskByIDs(kt *kit.Kit, ids []string, basicInfoMap map[string]types.CloudResourceBasicInfo) (
	*core.BatchOperateAllResult, error) {

	if len(ids) == 0 {
		return nil, nil
	}

	if len(ids) > constant.BatchOperationMaxLimit {
		return nil, errf.Newf(errf.InvalidParameter, "ids should <= %d", constant.BatchOperationMaxLimit)
	}

	listReq := &cloud.DiskCvmRelListReq{
		Filter: tools.ContainersExpression("disk_id", ids),
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

	diskCvmMap := make(map[string]string)
	for _, detail := range relRes.Details {
		diskCvmMap[detail.DiskID] = detail.CvmID
	}

	for _, id := range ids {
		cvmID, exists := diskCvmMap[id]
		if !exists {
			res.Succeeded = append(res.Succeeded, id)
			continue
		}

		info, exists := basicInfoMap[id]
		if !exists {
			res.Succeeded = append(res.Succeeded, id)
			continue
		}

		err = svc.diskLgc.DetachDisk(kt, info.Vendor, cvmID, id, info.AccountID)
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

// validateRecycleRecord 只能批量处理处于同一个回收任务的且是等待回收的记录。
func (svc *diskSvc) validateRecycleRecord(records *recyclerecord.ListResult) error {
	taskID := ""
	for _, one := range records.Details {
		if len(taskID) == 0 {
			taskID = one.TaskID
		} else if taskID != one.TaskID {
			return fmt.Errorf("only disks in one task can be reclaimed at the same time")
		}

		if one.Status != enumor.WaitingRecycleRecordStatus {
			return fmt.Errorf("record: %d not is wait_recycle status", one.ID)
		}

		if one.ResType != enumor.DiskCloudResType {
			return fmt.Errorf("record: %d not is disk recycle record", one.ID)
		}
	}

	return nil
}

// RecoverDisk recover disk.
func (svc *diskSvc) RecoverDisk(cts *rest.Contexts) (interface{}, error) {
	req := new(csdisk.DiskRecoverReq)
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
		ResType: enumor.DiskAuditResType,
		Action:  protoaudit.Recover,
		Infos:   auditInfos,
	}
	if err = svc.audit.ResRecycleAudit(cts.Kit, auditReq); err != nil {
		logs.Errorf("create recycle audit failed, err: %v, rid: %s", err, cts.Kit.Rid)
		return nil, err
	}

	opt := &recyclerecord.BatchRecoverReq{
		ResType:   enumor.DiskCloudResType,
		RecordIDs: req.RecordIDs,
	}
	err = svc.client.DataService().Global.RecycleRecord.BatchRecoverCloudResource(cts.Kit.Ctx, cts.Kit.Header(), opt)
	if err != nil {
		return nil, err
	}

	return nil, nil
}

// BatchDeleteRecycledDisk batch delete recycled disks.
func (svc *diskSvc) BatchDeleteRecycledDisk(cts *rest.Contexts) (interface{}, error) {
	req := new(csdisk.DiskDeleteRecycleReq)
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

	diskIDs := make([]string, 0, len(records.Details))
	for _, one := range records.Details {
		diskIDs = append(diskIDs, one.ResID)
	}

	basicInfoReq := cloud.ListResourceBasicInfoReq{
		ResourceType: enumor.DiskCloudResType,
		IDs:          diskIDs,
		Fields:       append(types.CommonBasicInfoFields, "recycle_status"),
	}
	basicInfoMap, err := svc.client.DataService().Global.Cloud.ListResourceBasicInfo(cts.Kit.Ctx, cts.Kit.Header(),
		basicInfoReq)
	if err != nil {
		return nil, err
	}

	// validate biz and authorize
	err = handler.RecycleValidWithAuth(cts, &handler.ValidWithAuthOption{Authorizer: svc.authorizer, ResType: meta.Vpc,
		Action: meta.Recycle, BasicInfos: basicInfoMap})
	if err != nil {
		return nil, err
	}

	delRes, err := svc.diskLgc.DeleteRecycledDisk(cts.Kit, basicInfoMap)
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
