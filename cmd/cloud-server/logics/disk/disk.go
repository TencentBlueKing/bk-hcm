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
	"hcm/pkg/api/cloud-server/recycle"
	"hcm/pkg/api/core"
	recyclerecord "hcm/pkg/api/core/recycle-record"
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
	"hcm/pkg/tools/slice"
)

// Interface define disk interface.
type Interface interface {
	DetachDisk(kt *kit.Kit, vendor enumor.Vendor, cvmID, diskID string) error
	DeleteDisk(kt *kit.Kit, vendor enumor.Vendor, diskID string) error
	DeleteRecycledDisk(kt *kit.Kit, infoMap map[string]types.CloudResourceBasicInfo) (*core.BatchOperateResult, error)

	BatchGetDiskInfo(kt *kit.Kit, cvmDetail map[string]*recycle.CvmDetail) (err error)
	BatchDetach(kt *kit.Kit, cvmRecycleMap map[string]*recycle.CvmDetail) (failed []string, err error)
	BatchReattachDisk(kt *kit.Kit, cvmRecycleMap map[string]*recycle.CvmDetail) (err error)
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
	relReq := &core.ListReq{
		Filter: tools.ContainersExpression("disk_id", ids),
		Page:   &core.BasePage{Count: true},
	}
	relRes, err := d.client.DataService().Global.ListDiskCvmRel(kt, relReq)
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

func (d *disk) fillAwsDisks(kt *kit.Kit, cvmDetails []*recycle.CvmDetail) error {

	cvmDiskMap := map[string][]recyclerecord.DiskAttachInfo{}
	cvmIds := slice.Map(cvmDetails, func(info *recycle.CvmDetail) string { return info.CvmID })
	relWithCvm, err := d.client.DataService().Aws.ListDiskCvmRelWithDisk(
		kt.Ctx, kt.Header(), &cloud.DiskCvmRelWithDiskListReq{CvmIDs: cvmIds})

	if err != nil {
		logs.Errorf("[aws] failed to ListDiskCvmRelWithDisk, err: %v, cvmIds:%v, rid:%s", err, cvmIds, kt.Rid)
		return err
	}
	for _, rel := range relWithCvm {
		if rel.Extension == nil || len(rel.Extension.Attachment) < 1 {
			return errf.Newf(errf.Unknown, "[Aws] no attachment found in cvm related disk, err: %v, cvmId: %v, rid:%s",
				err, rel.CvmID, kt.Rid)
		}
		if rel.IsSystemDisk {
			continue
		}
		cvmDiskMap[rel.CvmID] = append(cvmDiskMap[rel.CvmID], recyclerecord.DiskAttachInfo{
			DiskID:     rel.Disk.ID,
			DeviceName: converter.PtrToVal(rel.Extension.Attachment[0].DeviceName),
		})
	}
	for _, cvmDetail := range cvmDetails {
		if diskList, exists := cvmDiskMap[cvmDetail.CvmID]; exists {
			cvmDetail.DiskList = diskList
		}
	}
	return nil
}

func (d *disk) fillAzureDisk(kt *kit.Kit, cvmDetails []*recycle.CvmDetail) error {

	cvmDiskMap := map[string][]recyclerecord.DiskAttachInfo{}
	cvmIds := slice.Map(cvmDetails, func(info *recycle.CvmDetail) string { return info.CvmID })

	relWithCvm, err := d.client.DataService().Azure.ListDiskCvmRelWithDisk(kt.Ctx, kt.Header(),
		&cloud.DiskCvmRelWithDiskListReq{CvmIDs: cvmIds})

	if err != nil {
		logs.Errorf("[azure] failed to ListDiskCvmRelWithDisk, err: %v, cvmIds:%v, rid:%s", err, cvmIds, kt.Rid)
		return err
	}
	for _, rel := range relWithCvm {
		if rel.IsSystemDisk {
			continue
		}
		cvmDiskMap[rel.CvmID] = append(cvmDiskMap[rel.CvmID], recyclerecord.DiskAttachInfo{
			DiskID: rel.Disk.ID,
			// TODO:!!! 没有保存caching type，难以重新attach，暂时先按None恢复,-> 在vm 属性的storageProfile里面
			CachingType: "None",
		})
	}
	for _, cvmDetail := range cvmDetails {
		if diskList, exists := cvmDiskMap[cvmDetail.CvmID]; exists {
			cvmDetail.DiskList = diskList
		}
	}
	return nil
}

func (d *disk) fillDisk(kt *kit.Kit, vendor enumor.Vendor, cvmDetails []*recycle.CvmDetail) error {

	cvmDiskMap := map[string][]recyclerecord.DiskAttachInfo{}
	cvmIds := slice.Map(cvmDetails, func(info *recycle.CvmDetail) string { return info.CvmID })

	relWithCvm, err := d.client.DataService().Global.ListDiskCvmRelWithDisk(kt.Ctx, kt.Header(),
		&cloud.DiskCvmRelWithDiskListReq{CvmIDs: cvmIds})

	if err != nil {
		logs.Errorf("[%s] fail to ListDiskCvmRelWithDisk, err: %v, cvmIDs: %v, rid: %s",
			vendor, err, cvmIds, kt.Rid)
		return err
	}

	for _, rel := range relWithCvm {
		if rel.IsSystemDisk {
			continue
		}
		cvmDiskMap[rel.CvmID] = append(cvmDiskMap[rel.CvmID], recyclerecord.DiskAttachInfo{DiskID: rel.BaseDisk.ID})
	}

	for _, cvmDetail := range cvmDetails {
		if diskList, exists := cvmDiskMap[cvmDetail.CvmID]; exists {
			cvmDetail.DiskList = diskList
		}
	}
	return nil
}

// BatchGetDiskInfo 获取并填充磁盘信息
func (d *disk) BatchGetDiskInfo(kt *kit.Kit, cvmDetail map[string]*recycle.CvmDetail) (err error) {

	if len(cvmDetail) == 0 {
		return nil
	}
	if len(cvmDetail) > constant.BatchOperationMaxLimit {
		return errf.Newf(errf.InvalidParameter, "cvmIDs should <= %d", constant.BatchOperationMaxLimit)
	}

	infoByVendor := classifier.ClassifyMap(cvmDetail, func(v *recycle.CvmDetail) enumor.Vendor { return v.Vendor })
	// Aws 和Azure 参数不一样，需要通过with ext 获取特定参数
	for vendor, infos := range infoByVendor {

		switch vendor {
		case enumor.Aws:
			if err := d.fillAwsDisks(kt, infos); err != nil {
				return err
			}
		case enumor.Azure:
			if err := d.fillAzureDisk(kt, infos); err != nil {
				return err
			}
		case enumor.Gcp, enumor.HuaWei, enumor.TCloud:
			if err := d.fillDisk(kt, vendor, infos); err != nil {
				return err
			}
		default:
			return errf.Newf(errf.InvalidParameter, "unknown vendor: %v", vendor)
		}
	}
	return nil
}

// BatchDetach  批量解绑，返回的失败cvm, 用户自行决定是否回滚
func (d *disk) BatchDetach(kt *kit.Kit, cvmRecycleMap map[string]*recycle.CvmDetail) (failed []string,
	lastErr error) {
	if len(cvmRecycleMap) == 0 {
		return nil, nil
	}
	kt = kt.NewSubKit()

	for _, detail := range cvmRecycleMap {
		if detail.FailedAt != "" {
			continue
		}
		// 同一个cvm的磁盘解绑只失败一次
		for i, disk := range detail.DiskList {
			err := d.DetachDisk(kt, detail.Vendor, detail.CvmID, disk.DiskID)
			if err != nil {
				lastErr = err
				// 标记失败
				detail.DiskList[i].Err = err
				detail.FailedAt = enumor.DiskCloudResType
				failed = append(failed, detail.CvmID)
				logs.Errorf("failed to detach disk, err: %v cvmId: %s, diskId: %s, rid:%s",
					err, detail.CvmID, disk.DiskID, kt.Rid)
				// 继续处理下一个主机
				break
			}
		}
	}

	return failed, lastErr
}

// BatchReattachDisk 批量重新挂载磁盘, 仅处理磁盘卸载没有失败的磁盘
func (d *disk) BatchReattachDisk(kt *kit.Kit, cvmRecycleMap map[string]*recycle.CvmDetail) (err error) {
	for cvmId, detail := range cvmRecycleMap {

		for _, disk := range detail.DiskList {
			if disk.Err != nil {
				break
			}
			operationInfo := protoaudit.CloudResourceOperationInfo{
				ResType:           enumor.DiskAuditResType,
				ResID:             disk.DiskID,
				Action:            protoaudit.Associate,
				AssociatedResType: enumor.CvmAuditResType,
				AssociatedResID:   cvmId,
			}

			err := d.audit.ResOperationAudit(kt, operationInfo)
			if err != nil {
				logs.Errorf("create attach disk audit failed, err: %v, rid: %s", err, kt.Rid)
				return err
			}
			switch detail.Vendor {
			case enumor.Azure:
				err = d.client.HCService().Azure.Disk.AttachDisk(kt.Ctx, kt.Header(), &hcproto.AzureDiskAttachReq{
					CvmID: cvmId, DiskID: disk.DiskID, CachingType: disk.CachingType})
			case enumor.Aws:
				err = d.client.HCService().Aws.Disk.AttachDisk(kt.Ctx, kt.Header(), &hcproto.AwsDiskAttachReq{
					CvmID: cvmId, DiskID: disk.DiskID, DeviceName: disk.DeviceName})
			case enumor.TCloud:
				err = d.client.HCService().TCloud.Disk.AttachDisk(kt.Ctx, kt.Header(),
					&hcproto.TCloudDiskAttachReq{CvmID: cvmId, DiskID: disk.DiskID})
			case enumor.HuaWei:
				err = d.client.HCService().HuaWei.Disk.AttachDisk(kt.Ctx, kt.Header(),
					&hcproto.HuaWeiDiskAttachReq{CvmID: cvmId, DiskID: disk.DiskID})
			case enumor.Gcp:
				err = d.client.HCService().Gcp.Disk.AttachDisk(kt.Ctx, kt.Header(),
					&hcproto.GcpDiskAttachReq{CvmID: cvmId, DiskID: disk.DiskID})
			default:
				err = errf.NewFromErr(errf.InvalidParameter, fmt.Errorf("unknown vendor: %s", detail.Vendor))
			}
			if err != nil {
				logs.Errorf("fail to reattach, err: %v, cvmId: %s, disk: %v, rid: %s", err, cvmId, disk, kt.Rid)
			}
		}
	}
	return nil
}
