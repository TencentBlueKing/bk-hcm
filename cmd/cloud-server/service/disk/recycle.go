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

	"hcm/cmd/cloud-server/logics/recycle"
	csdisk "hcm/pkg/api/cloud-server/disk"
	csrecycle "hcm/pkg/api/cloud-server/recycle"
	"hcm/pkg/api/core"
	corerr "hcm/pkg/api/core/recycle-record"
	protoaudit "hcm/pkg/api/data-service/audit"
	"hcm/pkg/api/data-service/cloud"
	dsrr "hcm/pkg/api/data-service/recycle-record"
	"hcm/pkg/cc"
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
	return svc.recycleDiskSvc(cts, handler.ResOperateAuth)
}

// RecycleBizDisk recycle biz disk.
func (svc *diskSvc) RecycleBizDisk(cts *rest.Contexts) (interface{}, error) {
	return svc.recycleDiskSvc(cts, handler.BizOperateAuth)
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
		auditInfos = append(auditInfos,
			protoaudit.CloudResRecycleAuditInfo{ResID: info.ID, Data: info.DiskRecycleOptions})
	}

	basicInfoReq := cloud.ListResourceBasicInfoReq{
		ResourceType: enumor.DiskCloudResType,
		IDs:          ids,
		Fields:       append(types.CommonBasicInfoFields, "recycle_status"),
	}
	basicInfoMap, err := svc.client.DataService().Global.Cloud.ListResBasicInfo(cts.Kit, basicInfoReq)
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
	opt := &dsrr.BatchRecycleReq{
		ResType:            enumor.DiskCloudResType,
		DefaultRecycleTime: cc.CloudServer().Recycle.AutoDeleteTime,
		Infos:              make([]dsrr.RecycleReq, 0),
	}
	for _, info := range req.Infos {
		if _, exists := failedIDMap[info.ID]; exists {
			continue
		}
		opt.Infos = append(opt.Infos, dsrr.RecycleReq{
			ID:     info.ID,
			Detail: info.DiskRecycleOptions,
		})
	}

	taskID, err := svc.client.DataService().Global.RecycleRecord.BatchRecycleCloudRes(kt, opt)
	if err != nil {
		for _, id := range detachRes.Succeeded {
			res.Failed = append(res.Failed, core.FailedInfo{ID: id, Error: err})
		}
		return res, err
	}

	if len(res.Failed) > 0 {
		return res, res.Failed[0].Error
	}
	return &csrecycle.RecycleResult{TaskID: taskID}, nil
}

func (svc *diskSvc) detachDiskByIDs(kt *kit.Kit, ids []string, basicInfoMap map[string]types.CloudResourceBasicInfo) (
	*core.BatchOperateAllResult, error) {

	if len(ids) == 0 {
		return nil, nil
	}

	if len(ids) > constant.BatchOperationMaxLimit {
		return nil, errf.Newf(errf.InvalidParameter, "ids should <= %d", constant.BatchOperationMaxLimit)
	}

	listReq := &core.ListReq{
		Filter: tools.ContainersExpression("disk_id", ids),
		Page:   core.NewDefaultBasePage(),
	}
	relRes, err := svc.client.DataService().Global.ListDiskCvmRel(kt, listReq)
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

		err = svc.diskLgc.DetachDisk(kt, info.Vendor, cvmID, id)
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
func (svc *diskSvc) validateRecycleRecord(records *dsrr.ListResult) error {
	taskID := ""
	for _, one := range records.Details {
		if len(taskID) == 0 {
			taskID = one.TaskID
		} else if taskID != one.TaskID {
			return fmt.Errorf("only disks in one task can be reclaimed at the same time")
		}

		if one.Status != enumor.WaitingRecycleRecordStatus {
			return fmt.Errorf("record: %s not is wait_recycle status", one.ID)
		}

		if one.ResType != enumor.DiskCloudResType {
			return fmt.Errorf("record: %s not is disk recycle record", one.ID)
		}
		if one.RecycleType == enumor.RecycleTypeRelated {
			return fmt.Errorf("related recycled disk(%s) can not be operated", one.ResID)
		}
	}

	return nil
}

// RecoverDisk recover disk.
func (svc *diskSvc) RecoverDisk(cts *rest.Contexts) (interface{}, error) {
	return svc.recoverDisk(cts, handler.ResOperateAuth)
}

// RecoverBizDisk recover biz disk.
func (svc *diskSvc) RecoverBizDisk(cts *rest.Contexts) (interface{}, error) {
	return svc.recoverDisk(cts, handler.BizOperateAuth)
}

func (svc *diskSvc) recoverDisk(cts *rest.Contexts, validHandler handler.ValidWithAuthHandler) (interface{}, error) {
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
	records, err := svc.client.DataService().Global.RecycleRecord.ListRecycleRecord(cts.Kit, listReq)
	if err != nil {
		return nil, err
	}

	if len(records.Details) != len(req.RecordIDs) {
		return nil, errf.New(errf.InvalidParameter, "some record_ids are not in recycle bin")
	}
	if err = svc.validateRecycleRecord(records); err != nil {
		return nil, err
	}

	diskIds := make([]string, 0, len(records.Details))
	auditInfos := make([]protoaudit.CloudResRecycleAuditInfo, 0, len(records.Details))
	for _, record := range records.Details {
		diskIds = append(diskIds, record.ResID)
		auditInfos = append(auditInfos, protoaudit.CloudResRecycleAuditInfo{ResID: record.ResID, Data: record.Detail})
	}

	basicInfoReq := cloud.ListResourceBasicInfoReq{
		ResourceType: enumor.DiskCloudResType,
		IDs:          diskIds,
		Fields:       append(types.CommonBasicInfoFields, "recycle_status"),
	}
	basicInfoMap, err := svc.client.DataService().Global.Cloud.ListResBasicInfo(cts.Kit, basicInfoReq)
	if err != nil {
		return nil, err
	}

	// validate biz and authorize
	err = validHandler(cts, &handler.ValidWithAuthOption{Authorizer: svc.authorizer, ResType: meta.Disk,
		Action: meta.Recover, BasicInfos: basicInfoMap})
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

	opt := &dsrr.BatchRecoverReq{
		ResType:   enumor.DiskCloudResType,
		RecordIDs: req.RecordIDs,
	}
	err = svc.client.DataService().Global.RecycleRecord.BatchRecoverCloudResource(cts.Kit, opt)
	if err != nil {
		return nil, err
	}

	return nil, nil
}

// BatchDeleteBizRecycledDisk batch delete biz recycled disks.
func (svc *diskSvc) BatchDeleteBizRecycledDisk(cts *rest.Contexts) (interface{}, error) {
	return svc.batchDeleteRecycledDisk(cts, handler.BizOperateAuth)
}

// BatchDeleteRecycledDisk batch delete recycled disks.
func (svc *diskSvc) BatchDeleteRecycledDisk(cts *rest.Contexts) (interface{}, error) {
	return svc.batchDeleteRecycledDisk(cts, handler.ResOperateAuth)
}

func (svc *diskSvc) batchDeleteRecycledDisk(cts *rest.Contexts,
	validHandler handler.ValidWithAuthHandler) (interface{}, error) {

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
	records, err := svc.client.DataService().Global.RecycleRecord.ListRecycleRecord(cts.Kit, listReq)
	if err != nil {
		return nil, err
	}

	if len(records.Details) != len(req.RecordIDs) {
		return nil, errf.New(errf.InvalidParameter, "some record_ids are not in recycle bin")
	}

	if err = svc.validateRecycleRecord(records); err != nil {
		return nil, err
	}

	opRet := new(core.BatchOperateResult)
	var recycleErr error
	for _, record := range records.Details {
		recycleErr = svc.destroyOneRecord(cts, validHandler, record)
		if recycleErr != nil {
			logs.Errorf("fail to destroy disk recycle record(%s), err: %v, rid:%s", record.ID, recycleErr, cts.Kit.Rid)

			opRet.Failed = &core.FailedInfo{ID: record.ID, Error: recycleErr}
			// 目前只处理找不到记录的错误
			if ef := errf.Error(recycleErr); ef != nil && ef.Code == errf.RecordNotFound {
				logicsrecycle.MarkRecordFailed(cts.Kit, svc.client.DataService(), recycleErr, []string{record.ID})
			}
			// TODO: 目前遇到错误就停止处理后面任务，转成异步任务后优化成多个错误互相不影响
			break
		}
		opRet.Succeeded = append(opRet.Succeeded, record.ID)
	}
	// 标记成功
	if len(opRet.Succeeded) > 0 {
		logicsrecycle.MarkRecordSuccess(cts.Kit, svc.client.DataService(), opRet.Succeeded)
	}
	return opRet, recycleErr
}

func (svc *diskSvc) destroyOneRecord(cts *rest.Contexts, validHandler handler.ValidWithAuthHandler,
	record corerr.RecycleRecord) error {

	basicInfoReq := cloud.ListResourceBasicInfoReq{
		ResourceType: enumor.DiskCloudResType,
		IDs:          []string{record.ResID},
		Fields:       append(types.CommonBasicInfoFields, "recycle_status"),
	}

	basicInfoMap, err := svc.client.DataService().Global.Cloud.ListResBasicInfo(cts.Kit, basicInfoReq)
	if err != nil {
		return err
	}

	// validate biz and authorize
	err = validHandler(cts, &handler.ValidWithAuthOption{Authorizer: svc.authorizer, ResType: meta.Disk,
		Action: meta.Destroy, BasicInfos: basicInfoMap})
	if err != nil {
		return err
	}

	if _, err := svc.diskLgc.DeleteRecycledDisk(cts.Kit, basicInfoMap); err != nil {
		return err
	}

	return nil
}
