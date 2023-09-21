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
	"hcm/pkg/tools/converter"
	"hcm/pkg/tools/hooks/handler"
	"hcm/pkg/tools/maps"
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
		IDs:          converter.Map(req.Infos, func(e proto.CvmRecycleInfo) string { return e.ID }),
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

	auditInfos := converter.Map(req.Infos, func(info proto.CvmRecycleInfo) protoaudit.CloudResRecycleAuditInfo {
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

	return svc.recycleCvm(cts.Kit, req, basicInfoMap)
}

// recycleCvm 回收核心逻辑（创建recycle record）
func (svc *cvmSvc) recycleCvm(kt *kit.Kit, req *proto.CvmRecycleReq,
	cvmInfoMap map[string]types.CloudResourceBasicInfo) (interface{}, error) {

	// TODO: 简化核心逻辑
	recycleResult := new(core.BatchOperateAllResult)
	leftCvmInfo := maps.Clone(cvmInfoMap)
	markRecycleFail := func(cvmId string, reason error) {
		delete(leftCvmInfo, cvmId)
		recycleResult.Failed = append(recycleResult.Failed, core.FailedInfo{ID: cvmId, Error: reason})
	}
	// 1. 预检，有一个失败则全部失败
	checkResult := svc.cvmLgc.RecyclePreCheck(kt, leftCvmInfo)
	if len(checkResult.Failed) > 0 {
		return nil, checkResult.Failed[0].Error
	}

	eipFailedCvmIds := make([]string, 0)
	// 2. 解绑不随主机回收的disk
	detachDiskCvm := make(map[string]types.CloudResourceBasicInfo)
	for _, info := range req.Infos {
		if info.CvmRecycleOptions != nil && info.CvmRecycleOptions.WithDisk == false {
			detachDiskCvm[info.ID] = cvmInfoMap[info.ID]
		}
	}
	diskResult, diskRollBack, err := svc.diskLgc.BatchDetachWithRollback(kt, detachDiskCvm)
	if err != nil {
		for cvmId, err := range diskResult.FailedCvm {
			markRecycleFail(cvmId, err)
		}
	}
	if len(leftCvmInfo) == 0 {
		return recycleResult, recycleResult.Failed[0].Error
	}
	// 3. 解绑不随主机回收的eip
	detachEipCvmIDs := make([]string, 0)
	for _, info := range req.Infos {
		if info.CvmRecycleOptions != nil && info.CvmRecycleOptions.WithEip == false && leftCvmInfo[info.ID].ID == info.ID {
			detachEipCvmIDs = append(detachEipCvmIDs, info.ID)
		}
	}
	eipResult, eipRollback, err := svc.eipLgc.BatchDisassociateWithRollback(kt, detachEipCvmIDs)
	if err != nil {
		for cvmId, err := range eipResult.FailedCvm {
			// 逐个标记失败
			markRecycleFail(cvmId, err)
			eipFailedCvmIds = append(eipFailedCvmIds, cvmId)
		}
	}
	recycleFailedCvmIds := make([]string, 0, len(leftCvmInfo))
	defer func() {
		eipRollback(kt, recycleFailedCvmIds)
		diskRollBack(kt, append(eipFailedCvmIds, recycleFailedCvmIds...))
	}()
	if len(leftCvmInfo) == 0 {
		return recycleResult, recycleResult.Failed[0].Error
	}
	// 创建回收任务
	opt := &recyclerecord.BatchRecycleReq{
		ResType:            enumor.CvmCloudResType,
		DefaultRecycleTime: cc.CloudServer().Recycle.AutoDeleteTime,
	}

	for _, info := range req.Infos {
		// 过滤掉已经失败的id
		if _, exists := leftCvmInfo[info.ID]; exists {
			opt.Infos = append(opt.Infos, recyclerecord.RecycleReq{ID: info.ID, Detail: info.CvmRecycleOptions})
		}
	}
	taskID, err := svc.client.DataService().Global.RecycleRecord.BatchRecycleCloudRes(kt.Ctx, kt.Header(), opt)
	if err != nil {
		for _, info := range opt.Infos {
			markRecycleFail(info.ID, err)
			// 加入待回滚id
			recycleFailedCvmIds = append(recycleFailedCvmIds, info.ID)
		}
	}
	if len(recycleResult.Failed) > 0 {
		return recycleResult, recycleResult.Failed[0].Error
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
	for cvmID, bizID := range cvmIDToBizID {
		info := basicInfoMap[cvmID]
		info.BkBizID = bizID
		basicInfoMap[cvmID] = info
	}

	// 调用实际删除逻辑
	delRes, err := svc.cvmLgc.DestroyRecycledCvm(cts.Kit, basicInfoMap)
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
