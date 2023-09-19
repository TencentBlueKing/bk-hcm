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

// Package disk ...
package disk

import (
	"fmt"

	"hcm/cmd/cloud-server/logics/audit"
	"hcm/pkg/api/core"
	protoaudit "hcm/pkg/api/data-service/audit"
	"hcm/pkg/api/data-service/cloud"
	hcproto "hcm/pkg/api/hc-service/disk"
	"hcm/pkg/client"
	"hcm/pkg/criteria/constant"
	"hcm/pkg/criteria/enumor"
	"hcm/pkg/criteria/errf"
	"hcm/pkg/dal/dao/tools"
	"hcm/pkg/dal/dao/types"
	"hcm/pkg/kit"
	"hcm/pkg/logs"
	"hcm/pkg/tools/classifier"
	"hcm/pkg/tools/converter"
)

// Interface define disk interface.
type Interface interface {
	DetachDisk(kt *kit.Kit, vendor enumor.Vendor, cvmID, diskID string) error
	DeleteDisk(kt *kit.Kit, vendor enumor.Vendor, diskID string) error
	DeleteRecycledDisk(kt *kit.Kit, infoMap map[string]types.CloudResourceBasicInfo) (*core.BatchOperateResult, error)

	DetachDataDiskByCvmIDs(kt *kit.Kit, cvmIds []string,
		cvmInfoMap map[string]types.CloudResourceBasicInfo) (*core.BatchOperateAllResult, error)

	BatchDetachWithRollback(kt *kit.Kit, cvmInfoMap map[string]types.CloudResourceBasicInfo) (
		batchResult, BatchRollBackFunc, error)
}
type disk struct {
	client *client.ClientSet
	audit  audit.Interface
}

// NewDisk new disk.
func NewDisk(client *client.ClientSet, audit audit.Interface) Interface {
	return &disk{
		client: client,
		audit:  audit,
	}
}

// BatchRollBackFunc 批量操作回滚操作
type BatchRollBackFunc func(kt *kit.Kit, rollbackIds []string) (*core.BatchOperateAllResult, error)

type diskReattachInfo struct {
	Vendor      enumor.Vendor
	CvmID       string
	DiskID      string
	CachingType string
	DeviceName  string
}

type batchResult struct {
	SucceedResCvm map[string]string
	FailedCvm     map[string]error
}

// DetachDisk detach disk from cvm.
func (d *disk) DetachDisk(kt *kit.Kit, vendor enumor.Vendor, cvmID, diskID string) error {
	// create audit
	operationInfo := protoaudit.CloudResourceOperationInfo{
		ResType:           enumor.DiskAuditResType,
		ResID:             diskID,
		Action:            protoaudit.Disassociate,
		AssociatedResType: enumor.CvmAuditResType,
		AssociatedResID:   cvmID,
	}

	err := d.audit.ResOperationAudit(kt, operationInfo)
	if err != nil {
		logs.Errorf("create detach disk audit failed, err: %v, rid: %s", err, kt.Rid)
		return err
	}

	detachReq := &hcproto.DiskDetachReq{
		CvmID:  cvmID,
		DiskID: diskID,
	}

	switch vendor {
	case enumor.TCloud:
		return d.client.HCService().TCloud.Disk.DetachDisk(kt.Ctx, kt.Header(), detachReq)
	case enumor.Aws:
		return d.client.HCService().Aws.Disk.DetachDisk(kt.Ctx, kt.Header(), detachReq)
	case enumor.HuaWei:
		return d.client.HCService().HuaWei.Disk.DetachDisk(kt.Ctx, kt.Header(), detachReq)
	case enumor.Gcp:
		return d.client.HCService().Gcp.Disk.DetachDisk(kt.Ctx, kt.Header(), detachReq)
	case enumor.Azure:
		return d.client.HCService().Azure.Disk.DetachDisk(kt.Ctx, kt.Header(), detachReq)
	default:
		return errf.NewFromErr(errf.InvalidParameter, fmt.Errorf("no support vendor: %s", vendor))
	}
}

// DeleteDisk delete disk.
func (d *disk) DeleteDisk(kt *kit.Kit, vendor enumor.Vendor, diskID string) error {
	// create delete audit.
	err := d.audit.ResDeleteAudit(kt, enumor.DiskAuditResType, []string{diskID})
	if err != nil {
		logs.Errorf("create delete audit failed, err: %v, rid: %s", err, kt.Rid)
		return err
	}

	deleteReq := &hcproto.DiskDeleteReq{DiskID: diskID}

	switch vendor {
	case enumor.TCloud:
		return d.client.HCService().TCloud.Disk.DeleteDisk(kt.Ctx, kt.Header(), deleteReq)
	case enumor.Aws:
		return d.client.HCService().Aws.Disk.DeleteDisk(kt.Ctx, kt.Header(), deleteReq)
	case enumor.HuaWei:
		return d.client.HCService().HuaWei.Disk.DeleteDisk(kt.Ctx, kt.Header(), deleteReq)
	case enumor.Gcp:
		return d.client.HCService().Gcp.Disk.DeleteDisk(kt.Ctx, kt.Header(), deleteReq)
	case enumor.Azure:
		return d.client.HCService().Azure.Disk.DeleteDisk(kt.Ctx, kt.Header(), deleteReq)
	default:
		return errf.NewFromErr(errf.InvalidParameter, fmt.Errorf("no support vendor: %s", vendor))
	}
}

// DeleteRecycledDisk batch delete recycled disk.
func (d *disk) DeleteRecycledDisk(kt *kit.Kit, basicInfoMap map[string]types.CloudResourceBasicInfo) (
	*core.BatchOperateResult, error) {

	if len(basicInfoMap) == 0 {
		return nil, nil
	}

	if len(basicInfoMap) > constant.BatchOperationMaxLimit {
		return nil, errf.Newf(errf.InvalidParameter, "disk length should <= %d", constant.BatchOperationMaxLimit)
	}

	ids := make([]string, 0, len(basicInfoMap))
	for id := range basicInfoMap {
		ids = append(ids, id)
	}

	// check if disks are all detached
	relReq := &cloud.DiskCvmRelListReq{
		Filter: tools.ContainersExpression("disk_id", ids),
		Page:   &core.BasePage{Count: true},
	}
	relRes, err := d.client.DataService().Global.ListDiskCvmRel(kt.Ctx, kt.Header(), relReq)
	if err != nil {
		return nil, err
	}

	if converter.PtrToVal(relRes.Count) > 0 {
		logs.Errorf("some recycled disks(ids: %+v) are attached, cannot be deleted, rid: %s", ids, kt.Rid)
		return nil, errf.New(errf.InvalidParameter, "recycled disk is attached, cannot be deleted")
	}

	res := new(core.BatchOperateResult)

	// delete disk
	for _, id := range ids {
		info := basicInfoMap[id]
		err = d.DeleteDisk(kt, info.Vendor, id)
		if err != nil {
			res.Failed = &core.FailedInfo{ID: id, Error: err}
			return res, err
		}
		res.Succeeded = append(res.Succeeded, id)
	}

	return nil, nil
}

// DetachDataDiskByCvmIDs 解绑cvm下的数据盘
func (d *disk) DetachDataDiskByCvmIDs(kt *kit.Kit, cvmIds []string,
	cvmInfoMap map[string]types.CloudResourceBasicInfo) (*core.BatchOperateAllResult, error) {

	if len(cvmIds) == 0 {
		return nil, nil
	}

	if len(cvmIds) > constant.BatchOperationMaxLimit {
		return nil, errf.Newf(errf.InvalidParameter, "cvmIds should <= %d", constant.BatchOperationMaxLimit)
	}

	listReq := &cloud.DiskCvmRelListReq{
		Filter: tools.ContainersExpression("cvm_id", cvmIds),
		Page:   core.NewDefaultBasePage(),
	}
	relRes, err := d.client.DataService().Global.ListDiskCvmRel(kt.Ctx, kt.Header(), listReq)
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

	for _, cvmId := range cvmIds {
		diskID, exists := cvmDiskMap[cvmId]
		if !exists {
			res.Succeeded = append(res.Succeeded, cvmId)
			continue
		}

		cvmInfo, exists := cvmInfoMap[cvmId]
		if !exists {
			res.Succeeded = append(res.Succeeded, cvmId)
			continue
		}

		err = d.DetachDisk(kt, cvmInfo.Vendor, cvmId, diskID)
		if err != nil {
			res.Failed = append(res.Failed, core.FailedInfo{ID: cvmId, Error: err})
			continue
		}
		res.Succeeded = append(res.Succeeded, cvmId)
	}

	if len(res.Failed) > 0 {
		return res, res.Failed[0].Error
	}
	return res, nil
}

// DetachDataDiskByCvmIDs 解绑cvm下的数据盘
func (d *disk) getDiskByCvm(kt *kit.Kit, cvmInfos map[string]types.CloudResourceBasicInfo) (
	diskCvmMap map[string]string, reattachMap map[string]diskReattachInfo, err error) {

	if len(cvmInfos) == 0 {
		return nil, nil, nil
	}
	if len(cvmInfos) > constant.BatchOperationMaxLimit {
		return nil, nil, errf.Newf(errf.InvalidParameter, "cvmIDs should <= %d", constant.BatchOperationMaxLimit)
	}

	reattachMap = make(map[string]diskReattachInfo)
	diskCvmMap = make(map[string]string)

	infoByVendor := classifier.ClassifyBasicInfoByVendor(cvmInfos)
	// Aws 和Azure 参数不一样，需要通过with ext 获取特定参数
	for vendor, infos := range infoByVendor {

		cvmIds := converter.Map(infos, func(info types.CloudResourceBasicInfo) string { return info.ID })
		switch vendor {
		case enumor.Aws:
			err := d.getAwsDisk(kt, cvmIds, diskCvmMap, reattachMap)
			if err != nil {
				return nil, nil, err
			}
		case enumor.Azure:
			err := d.getAzureDisk(kt, cvmIds, diskCvmMap, reattachMap)
			if err != nil {
				return nil, nil, err
			}
		case enumor.Gcp, enumor.HuaWei, enumor.TCloud:
			err := d.getDiskInfo(kt, vendor, cvmIds, diskCvmMap, reattachMap)
			if err != nil {
				return nil, nil, err
			}
		default:
			return nil, nil, errf.Newf(errf.InvalidParameter, "unknown vendor: %v", vendor)
		}
	}
	return diskCvmMap, reattachMap, nil
}

func (d *disk) getDiskInfo(kt *kit.Kit, vendor enumor.Vendor, cvmIds []string, diskCvmMap map[string]string,
	reattachMap map[string]diskReattachInfo) error {

	relWithCvm, err := d.client.DataService().Global.ListDiskCvmRelWithDisk(kt.Ctx, kt.Header(),
		&cloud.DiskCvmRelWithDiskListReq{
			CvmIDs: cvmIds,
		})

	if err != nil {
		logs.Errorf("[%s] fail to ListDiskCvmRelWithDisk, err: %v, cvmIDs: %v, rid: %s",
			vendor, err, cvmIds, kt.Rid)
		return err
	}

	for _, rel := range relWithCvm {
		diskCvmMap[rel.ID] = rel.CvmID
		reattachMap[rel.ID] = diskReattachInfo{
			Vendor: vendor,
			CvmID:  rel.CvmID,
			DiskID: rel.DiskResult.ID,
		}
	}
	return nil
}

func (d *disk) getAzureDisk(kt *kit.Kit, cvmIds []string, diskCvmMap map[string]string,
	reattachMap map[string]diskReattachInfo) error {

	relWithCvm, err := d.client.DataService().Azure.ListDiskCvmRelWithDisk(kt.Ctx, kt.Header(),
		&cloud.DiskCvmRelWithDiskListReq{
			CvmIDs: cvmIds,
		})

	if err != nil {
		logs.Errorf("[Azure] failed to ListDiskCvmRelWithDisk, err: %v, cvmIds:%v, rid:%s", err, cvmIds, kt.Rid)
		return err
	}
	for _, rel := range relWithCvm {
		diskCvmMap[rel.ID] = rel.CvmID
		reattachMap[rel.ID] = diskReattachInfo{
			Vendor: enumor.Azure,
			CvmID:  rel.CvmID,
			DiskID: rel.DiskExtResult.ID,
			// TODO:!!! 没有保存caching type，难以重新attach，暂时先按None恢复,-> 在vm 属性的storageProfile里面
			CachingType: "None",
		}
	}
	return nil
}

func (d *disk) getAwsDisk(kt *kit.Kit, cvmIds []string, diskCvmMap map[string]string,
	reattachMap map[string]diskReattachInfo) error {

	relWithCvm, err := d.client.DataService().Aws.ListDiskCvmRelWithDisk(kt.Ctx, kt.Header(),
		&cloud.DiskCvmRelWithDiskListReq{
			CvmIDs: cvmIds,
		})

	if err != nil {
		logs.Errorf("[Aws] failed to ListDiskCvmRelWithDisk, err: %v, cvmIds:%v, rid:%s", err, cvmIds, kt.Rid)
		return err
	}
	for _, rel := range relWithCvm {
		diskCvmMap[rel.ID] = rel.CvmID
		if rel.Extension == nil || len(rel.Extension.Attachment) < 1 {
			return errf.Newf(errf.Unknown, "[Aws] no disk attachment in disk, err: %v, cvmId: %v, rid:%s",
				err, rel.CvmID, kt.Rid)
		}

		reattachMap[rel.ID] = diskReattachInfo{
			Vendor:     enumor.Aws,
			CvmID:      rel.CvmID,
			DiskID:     rel.DiskExtResult.ID,
			DeviceName: converter.PtrToVal(rel.Extension.Attachment[0].DeviceName),
		}
	}
	return nil
}

// BatchDetachWithRollback  批量解绑，返回回滚函数，返回的失败cvm, 用户自行决定是否回滚
func (d *disk) BatchDetachWithRollback(kt *kit.Kit, cvmInfoMap map[string]types.CloudResourceBasicInfo) (
	batchResult, BatchRollBackFunc, error) {

	// 1. 获取disk和cvm关联信息以及d磁盘重新挂载信息
	detachResult := batchResult{map[string]string{}, map[string]error{}}

	diskCvmMap, diskMap, err := d.getDiskByCvm(kt, cvmInfoMap)
	if err != nil {
		for cvmId := range cvmInfoMap {
			detachResult.FailedCvm[cvmId] = err
		}
		return detachResult, nil, err
	}

	rollback := func(kt *kit.Kit, rollbackIds []string) (*core.BatchOperateAllResult, error) {
		if len(rollbackIds) == 0 {
			return nil, nil
		}
		logs.V(3).Infof("rollback for BatchDisassociateEip, rollback cvm ids: %v, rid:%s", rollbackIds, kt.Rid)
		reattachResult := &core.BatchOperateAllResult{}
		rbCvmIds := converter.StringSliceToMap(rollbackIds)
		var err error
		for diskId, cvmId := range diskCvmMap {
			if _, ok := rbCvmIds[cvmId]; !ok {
				continue
			}
			if err = d.reattach(kt, diskMap[diskId]); err != nil {
				reattachResult.Failed = append(reattachResult.Failed, core.FailedInfo{ID: cvmId, Error: err})
			} else {
				reattachResult.Succeeded = append(reattachResult.Succeeded, cvmId)
			}
		}
		return reattachResult, err
	}

	// 2. 尝试卸载磁盘
	for diskId, cvmId := range diskCvmMap {
		// 如果cvm存在多个磁盘，只失败一次就行了，不要重复失败
		if detachResult.FailedCvm[cvmId] != nil {
			continue
		}
		err = d.DetachDisk(kt, diskMap[diskId].Vendor, cvmId, diskId)
		if err != nil {
			detachResult.FailedCvm[cvmId] = err
		} else {
			detachResult.SucceedResCvm[cvmId] = diskId
		}
	}
	return detachResult, rollback, err
}

// reattach 重新重新挂载磁盘
func (d *disk) reattach(kt *kit.Kit, attachInfo diskReattachInfo) error {

	operationInfo := protoaudit.CloudResourceOperationInfo{
		ResType:           enumor.DiskAuditResType,
		ResID:             attachInfo.DiskID,
		Action:            protoaudit.Associate,
		AssociatedResType: enumor.CvmAuditResType,
		AssociatedResID:   attachInfo.CvmID,
	}

	err := d.audit.ResOperationAudit(kt, operationInfo)
	if err != nil {
		logs.Errorf("create attach disk audit failed, err: %v, rid: %s", err, kt.Rid)
		return err
	}
	switch attachInfo.Vendor {
	case enumor.Azure:
		err = d.client.HCService().Azure.Disk.AttachDisk(kt.Ctx, kt.Header(), &hcproto.AzureDiskAttachReq{
			CvmID: attachInfo.CvmID, DiskID: attachInfo.DiskID, CachingType: attachInfo.CachingType})
	case enumor.Aws:
		err = d.client.HCService().Aws.Disk.AttachDisk(kt.Ctx, kt.Header(), &hcproto.AwsDiskAttachReq{
			CvmID: attachInfo.CvmID, DiskID: attachInfo.DiskID, DeviceName: attachInfo.DeviceName})
	case enumor.TCloud:
		err = d.client.HCService().TCloud.Disk.AttachDisk(kt.Ctx, kt.Header(),
			&hcproto.TCloudDiskAttachReq{CvmID: attachInfo.CvmID, DiskID: attachInfo.DiskID})
	case enumor.HuaWei:
		err = d.client.HCService().HuaWei.Disk.AttachDisk(kt.Ctx, kt.Header(),
			&hcproto.HuaWeiDiskAttachReq{CvmID: attachInfo.CvmID, DiskID: attachInfo.DiskID})
	case enumor.Gcp:
		err = d.client.HCService().Gcp.Disk.AttachDisk(kt.Ctx, kt.Header(),
			&hcproto.GcpDiskAttachReq{CvmID: attachInfo.CvmID, DiskID: attachInfo.DiskID})
	default:
		err = errf.NewFromErr(errf.InvalidParameter, fmt.Errorf("unknown vendor: %s", attachInfo.Vendor))
	}
	if err != nil {
		logs.Errorf("fail to reattach, err: %v, params: %v, rid: %s", err, attachInfo, kt.Rid)
	}
	return err
}
