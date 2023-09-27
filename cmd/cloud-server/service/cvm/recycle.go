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
	corerecyclerecord "hcm/pkg/api/core/recycle-record"
	protoaudit "hcm/pkg/api/data-service/audit"
	"hcm/pkg/api/data-service/cloud"
	recyclerecord "hcm/pkg/api/data-service/recycle-record"
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
	"hcm/pkg/tools/json"
	"hcm/pkg/tools/maps"
	"hcm/pkg/tools/slice"
)

// RecycleCvm recycle cvm.
func (svc *cvmSvc) RecycleCvm(cts *rest.Contexts) (interface{}, error) {
	return svc.recycleCvmSvc(cts, handler.ResValidWithAuth)
}

// RecycleBizCvm recycle biz cvm.
func (svc *cvmSvc) RecycleBizCvm(cts *rest.Contexts) (interface{}, error) {
	return svc.recycleCvmSvc(cts, handler.BizValidWithAuth)
}

// recycleCvmSvc cvm 标记回收 接口对接、前置校验和审计
func (svc *cvmSvc) recycleCvmSvc(cts *rest.Contexts, validHandler handler.ValidWithAuthHandler) (interface{}, error) {
	req := new(proto.CvmRecycleReq)
	if err := cts.DecodeInto(req); err != nil {
		return nil, err
	}
	if err := req.Validate(); err != nil {
		return nil, errf.NewFromErr(errf.InvalidParameter, err)
	}

	basicInfoReq := cloud.ListResourceBasicInfoReq{
		ResourceType: enumor.CvmCloudResType,
		IDs:          slice.Map(req.Infos, func(e proto.CvmRecycleInfo) string { return e.ID }),
		Fields:       append(types.CommonBasicInfoFields, "region", "recycle_status"),
	}
	basicInfoMap, err := svc.client.DataService().Global.Cloud.
		ListResourceBasicInfo(cts.Kit.Ctx, cts.Kit.Header(), basicInfoReq)
	if err != nil {
		return nil, err
	}

	// validate biz and authorize
	err = validHandler(cts, &handler.ValidWithAuthOption{Authorizer: svc.authorizer, ResType: meta.Cvm,
		Action: meta.Recycle, BasicInfos: basicInfoMap})
	if err != nil {
		return nil, err
	}

	// 1. 预检，有一个失败则全部失败，且不进审计
	if err := svc.cvmLgc.RecyclePreCheck(cts.Kit, basicInfoMap); err != nil {
		logs.Errorf("recycle precheck fail, err: %v, rid: %s", err, cts.Kit.Rid)
		return nil, err
	}

	cvmStatus := make(map[string]*recycle.CvmDetail, len(req.Infos))
	for _, cvmRecycleReq := range req.Infos {
		cvmStatus[cvmRecycleReq.ID] = &recycle.CvmDetail{
			Vendor:           basicInfoMap[cvmRecycleReq.ID].Vendor,
			AccountID:        basicInfoMap[cvmRecycleReq.ID].AccountID,
			CvmID:            cvmRecycleReq.ID,
			CvmRecycleDetail: corerecyclerecord.CvmRecycleDetail{CvmRecycleOptions: cvmRecycleReq.CvmRecycleOptions},
		}
	}

	auditInfos := slice.Map(req.Infos, func(info proto.CvmRecycleInfo) protoaudit.CloudResRecycleAuditInfo {
		return protoaudit.CloudResRecycleAuditInfo{ResID: info.ID, Data: info.CvmRecycleOptions}
	})
	// create recycle audit
	auditReq := &protoaudit.CloudResourceRecycleAuditReq{ResType: enumor.CvmAuditResType, Action: protoaudit.Recycle,
		Infos: auditInfos,
	}
	if err = svc.audit.ResRecycleAudit(cts.Kit, auditReq); err != nil {
		logs.Errorf("create recycle audit failed, err: %v, rid: %s", err, cts.Kit.Rid)
		return nil, err
	}

	defer func(svc *cvmSvc, kt *kit.Kit, cvmStatus map[string]*recycle.CvmDetail) {
		err := svc.recycleCleanUp(kt, cvmStatus)
		if err != nil {
			logs.Errorf("failed to cleanup recycle, err: %v, rid: %s", err, kt.Rid)
		}
	}(svc, cts.Kit, cvmStatus)

	taskID, err := svc.recycleCvm(cts.Kit, req, cvmStatus)
	if err != nil {
		return nil, err
	}
	return recycle.RecycleResult{TaskID: taskID}, nil
}

// recycleCvm  回收核心逻辑（创建recycle record）
// 1. 获取磁盘信息
// 2. 解绑不随主机回收磁盘
// 3. 获取eip信息
// 4. 解绑不随主机回收eip
// 5. 标记磁盘和eip为被动回收
// 6. 回收主机 (仅回收前置步骤成功的）
func (svc *cvmSvc) recycleCvm(kt *kit.Kit, req *proto.CvmRecycleReq,
	cvmStatus map[string]*recycle.CvmDetail) (taskID string, err error) {
	// 获取磁盘信息
	if err := svc.diskLgc.BatchGetDiskInfo(kt, cvmStatus); err != nil {
		logs.Errorf("failed to get disk info of cvm, err: %v, rid: %s", err, kt.Rid)
		return "", err
	}
	// 过滤出不随主机回收的磁盘，并解绑
	failed, err := svc.diskLgc.BatchDetach(kt,
		maps.FilterByValue(cvmStatus, func(c *recycle.CvmDetail) bool { return !c.WithDisk }))
	if err != nil {
		logs.Errorf("failed to detach some disks of cvm(%v), err: %v, rid: %s", failed, err, kt.Rid)
	}

	// 获取eip信息
	if err := svc.eipLgc.BatchGetEipInfo(kt, cvmStatus); err != nil {
		logs.Errorf("failed to get eip info of cvm, err: %v, rid: %s", err, kt.Rid)
	}

	// 过滤出不随主机回收的Eip，并解绑
	failed, err = svc.eipLgc.BatchUnbind(kt,
		maps.FilterByValue(cvmStatus, func(c *recycle.CvmDetail) bool { return !c.WithEip }))
	if err != nil {
		logs.Errorf("failed to unbind eip of cvm(%v), err: %v, rid: %s", failed, err, kt.Rid)
	}

	// 创建回收任务
	opt := &recyclerecord.BatchRecycleReq{
		ResType:            enumor.CvmCloudResType,
		DefaultRecycleTime: cc.CloudServer().Recycle.AutoDeleteTime,
	}

	for _, info := range req.Infos {
		// 过滤掉已经失败的id
		if recCvm := cvmStatus[info.ID]; recCvm != nil && recCvm.FailedAt == "" {
			opt.Infos = append(opt.Infos,
				recyclerecord.RecycleReq{ID: info.ID, Detail: cvmStatus[info.ID].CvmRecycleDetail})
		}
	}
	// 标记磁盘和eip为回收
	err = svc.markRelatedRecycleStatus(kt, cvmStatus)
	if err != nil {
		return "", err
	}

	// 批量加入回收任务
	taskID, err = svc.client.DataService().Global.RecycleRecord.BatchRecycleCloudRes(kt.Ctx, kt.Header(), opt)
	if err != nil {
		logs.Errorf("fail to recycle cvm, err: %v, rid: %s", err, kt.Rid)
		for _, info := range opt.Infos {
			cvmStatus[info.ID].FailedAt = enumor.CvmCloudResType
		}
		return "", err
	}

	return taskID, nil
}

// recycleCleanUp 处理回收失败需要尝试重新绑定的eip、disk
func (svc *cvmSvc) recycleCleanUp(kt *kit.Kit, cvmStatus map[string]*recycle.CvmDetail) error {

	eipRebind := make(map[string]*recycle.CvmDetail, len(cvmStatus))
	diskRebind := make(map[string]*recycle.CvmDetail, len(cvmStatus))

	for cvmId, detail := range cvmStatus {
		switch detail.FailedAt {
		case "":
			continue
		case enumor.DiskCloudResType:
			continue
		case enumor.EipCloudResType:
			diskRebind[cvmId] = detail
		case enumor.CvmCloudResType:
			// 	重新挂载磁盘和绑定eip
			eipRebind[cvmId] = detail
			diskRebind[cvmId] = detail
		default:
			return fmt.Errorf("unknown failed type: %v", detail.FailedAt)
		}
	}
	// 	尝试重新挂载磁盘
	err := svc.eipLgc.BatchRebind(kt, eipRebind)
	if err != nil {
		return err
	}
	err = svc.diskLgc.BatchReattachDisk(kt, diskRebind)
	if err != nil {
		return err
	}
	return nil
}

// markRelatedRecycleStatus 将关联资源标记为回收状态
func (svc *cvmSvc) markRelatedRecycleStatus(kt *kit.Kit, cvmStatus map[string]*recycle.CvmDetail) error {
	var diskIds, eipIds []string
	for _, recCvm := range cvmStatus {
		// 过滤掉已经失败的id
		if recCvm.FailedAt != "" {
			continue
		}
		if recCvm.WithDisk {
			for _, disk := range recCvm.DiskList {
				diskIds = append(diskIds, disk.DiskID)
			}
		}
		if recCvm.WithEip {
			for _, eip := range recCvm.EipList {
				eipIds = append(eipIds, eip.EipID)
			}
		}

	}
	// 标记磁盘和eip为回收
	if len(diskIds) > 0 {
		err := svc.client.DataService().Global.RecycleRecord.BatchUpdateRecycleStatus(kt,
			&recyclerecord.BatchUpdateRecycleStatusReq{
				ResType:       enumor.DiskCloudResType,
				IDs:           diskIds,
				RecycleStatus: enumor.RecycleStatus,
			})
		if err != nil {
			logs.Errorf("fail to mark disk recycling status, err: %v, disk ids: %v, rid: %s", err, diskIds, kt.Rid)
			return err
		}
	}
	if len(eipIds) > 0 {
		err := svc.client.DataService().Global.RecycleRecord.BatchUpdateRecycleStatus(kt,
			&recyclerecord.BatchUpdateRecycleStatusReq{
				ResType:       enumor.EipCloudResType,
				IDs:           eipIds,
				RecycleStatus: enumor.RecycleStatus,
			})
		if err != nil {
			logs.Errorf("fail to mark eip recycling status, err: %v, eip ids: %v, rid: %s", err, eipIds, kt.Rid)

			return err
		}
	}
	return nil

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

		err = svc.diskLgc.DetachDisk(kt, info.Vendor, id, diskID)
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

// RecoverBizCvm recover biz cvm.
func (svc *cvmSvc) RecoverBizCvm(cts *rest.Contexts) (interface{}, error) {
	return svc.recoverCvm(cts, handler.BizValidWithAuth)
}

// RecoverCvm recover cvm.
func (svc *cvmSvc) RecoverCvm(cts *rest.Contexts) (interface{}, error) {
	return svc.recoverCvm(cts, handler.ResValidWithAuth)
}

func (svc *cvmSvc) recoverCvm(cts *rest.Contexts, validHandler handler.ValidWithAuthHandler) (interface{}, error) {
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
	auditInfos := make([]protoaudit.CloudResRecycleAuditInfo, 0, len(records.Details))

	ids := make([]string, 0, len(records.Details))
	for _, record := range records.Details {
		ids = append(ids, record.ResID)
		auditInfos = append(auditInfos, protoaudit.CloudResRecycleAuditInfo{ResID: record.ResID, Data: record.Detail})
	}

	basicInfoReq := cloud.ListResourceBasicInfoReq{
		ResourceType: enumor.CvmCloudResType,
		IDs:          ids,
		Fields:       append(types.CommonBasicInfoFields, "recycle_status"),
	}
	basicInfoMap, err := svc.client.DataService().Global.Cloud.ListResourceBasicInfo(cts.Kit.Ctx, cts.Kit.Header(),
		basicInfoReq)
	if err != nil {
		return nil, err
	}

	// validate biz and authorize
	err = validHandler(cts, &handler.ValidWithAuthOption{Authorizer: svc.authorizer, ResType: meta.Cvm,
		Action: meta.Recover, BasicInfos: basicInfoMap})
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

	if err := svc.recoveryRelatedRes(cts, records.Details, err); err != nil {
		return nil, err
	}
	return nil, nil
}

func (svc *cvmSvc) recoveryRelatedRes(cts *rest.Contexts, records []corerecyclerecord.RecycleRecord, err error) error {
	cvmRecycleDetails := make([]corerecyclerecord.CvmRecycleDetail, len(records))
	var diskIds, eipIds []string
	for i, record := range records {
		if detailStr, ok := record.Detail.(string); ok {
			if err := json.UnmarshalFromString(detailStr, &cvmRecycleDetails[i]); err != nil {
				return err
			}
		} else {
			return fmt.Errorf("wrong type of recycle detail, got: %t, need string", record.Detail)
		}
		if cvmRecycleDetails[i].WithDisk && len(cvmRecycleDetails[i].DiskList) > 0 {
			for _, disk := range cvmRecycleDetails[i].DiskList {
				diskIds = append(diskIds, disk.DiskID)
			}
		}
		if cvmRecycleDetails[i].WithEip && len(cvmRecycleDetails[i].EipList) > 0 {
			for _, eip := range cvmRecycleDetails[i].EipList {
				eipIds = append(eipIds, eip.EipID)
			}
		}
	}
	if len(diskIds) > 0 {
		updateOpt := recyclerecord.BatchUpdateRecycleStatusReq{
			RecycleStatus: enumor.RecoverRecycleRecordStatus,
			IDs:           diskIds,
			ResType:       enumor.DiskCloudResType,
		}
		err = svc.client.DataService().Global.RecycleRecord.BatchUpdateRecycleStatus(cts.Kit, &updateOpt)
		if err != nil {
			logs.Errorf("fail to recover related res recycle status, err: %v, disk ids: %v, rid: %s",
				err, diskIds, cts.Kit.Rid)
			return err
		}
	}

	if len(eipIds) > 0 {
		updateOpt := recyclerecord.BatchUpdateRecycleStatusReq{
			RecycleStatus: enumor.RecoverRecycleRecordStatus,
			IDs:           eipIds,
			ResType:       enumor.EipCloudResType,
		}
		err = svc.client.DataService().Global.RecycleRecord.BatchUpdateRecycleStatus(cts.Kit, &updateOpt)
		if err != nil {
			logs.Errorf("fail to recover related res recycle status, err: %v, disk ids: %v, rid: %s",
				err, diskIds, cts.Kit.Rid)
			return err
		}
	}
	return nil
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
			return fmt.Errorf("record: %s not is wait_recycle status", one.ID)
		}

		if one.ResType != enumor.CvmCloudResType {
			return fmt.Errorf("record: %d not is cvm recycle record", one.ID)
		}
	}

	return nil
}

// BatchDeleteBizRecycledCvm batch delete biz recycled cvm.
func (svc *cvmSvc) BatchDeleteBizRecycledCvm(cts *rest.Contexts) (interface{}, error) {
	return svc.batchDeleteRecycledCvm(cts, handler.BizValidWithAuth)
}

// BatchDeleteRecycledCvm 立即销毁回收任务中的主机
func (svc *cvmSvc) BatchDeleteRecycledCvm(cts *rest.Contexts) (interface{}, error) {
	return svc.batchDeleteRecycledCvm(cts, handler.ResValidWithAuth)
}

// batchDeleteRecycledCvm 对接web端用户手动删除销毁接口
func (svc *cvmSvc) batchDeleteRecycledCvm(cts *rest.Contexts,
	validHandler handler.ValidWithAuthHandler) (interface{}, error) {

	req := new(proto.CvmDeleteRecycledReq)
	if err := cts.DecodeInto(req); err != nil {
		return nil, err
	}
	if err := req.Validate(); err != nil {
		return nil, errf.NewFromErr(errf.InvalidParameter, err)
	}

	// 1. 检查用户输入是否都在回收站中
	listReq := &core.ListReq{Filter: tools.ContainersExpression("id", req.RecordIDs)}
	listReq.Page = &core.BasePage{Limit: constant.BatchOperationMaxLimit}

	records, err := svc.client.DataService().Global.RecycleRecord.
		ListRecycleRecord(cts.Kit.Ctx, cts.Kit.Header(), listReq)
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
	cvmIDToBizID := make(map[string]int64, len(records.Details))
	for _, recordDetail := range records.Details {
		cvmIDs = append(cvmIDs, recordDetail.ResID)
		cvmIDToBizID[recordDetail.ResID] = recordDetail.BkBizID
	}

	basicInfoReq := cloud.ListResourceBasicInfoReq{ResourceType: enumor.CvmCloudResType, IDs: cvmIDs}
	basicInfoReq.Fields = append(types.CommonBasicInfoFields, "region", "recycle_status")
	basicInfoMap, err := svc.client.DataService().Global.Cloud.ListResourceBasicInfo(cts.Kit.Ctx, cts.Kit.Header(),
		basicInfoReq)
	if err != nil {
		return nil, err
	}

	// validate biz and authorize
	err = validHandler(cts, &handler.ValidWithAuthOption{Authorizer: svc.authorizer, ResType: meta.Cvm,
		Action: meta.DeleteRecycled, BasicInfos: basicInfoMap})
	if err != nil {
		return nil, err
	}

	// 创建回收任务的主机的业务id已经被清掉了，需要借助recycle record 来获取
	// TODO: 已经保留业务id，考虑将下面逻辑删除
	for cvmID, bizID := range cvmIDToBizID {
		info := basicInfoMap[cvmID]
		info.BkBizID = bizID
		basicInfoMap[cvmID] = info
	}

	// 调用实际删除逻辑
	delRes, err := svc.cvmLgc.DestroyRecycledCvm(cts.Kit, basicInfoMap, records.Details)
	if err != nil {
		return delRes, err
	}

	updateReq := &recyclerecord.BatchUpdateReq{
		Data: make([]recyclerecord.UpdateReq, 0, len(req.RecordIDs)),
	}
	for _, id := range req.RecordIDs {
		updateReq.Data = append(updateReq.Data,
			recyclerecord.UpdateReq{ID: id, Status: enumor.RecycledRecycleRecordStatus})
	}

	if err = svc.client.DataService().Global.RecycleRecord.BatchUpdateRecycleRecord(cts.Kit.Ctx, cts.Kit.Header(),
		updateReq); err != nil {
		logs.Errorf("update recycle record status to recycled failed, err: %v, ids: %v, rid: %s", err, req.RecordIDs,
			cts.Kit.Rid)
		return nil, err
	}
	return nil, nil
}
