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

package huawei

import (
	"fmt"

	"hcm/cmd/hc-service/logics/res-sync/common"
	adcore "hcm/pkg/adaptor/types/core"
	adaptordisk "hcm/pkg/adaptor/types/disk"
	"hcm/pkg/api/core"
	coredisk "hcm/pkg/api/core/cloud/disk"
	"hcm/pkg/api/data-service/cloud/disk"
	"hcm/pkg/criteria/constant"
	"hcm/pkg/criteria/enumor"
	"hcm/pkg/criteria/errf"
	"hcm/pkg/criteria/validator"
	"hcm/pkg/dal/dao/tools"
	"hcm/pkg/kit"
	"hcm/pkg/logs"
	"hcm/pkg/runtime/filter"
	"hcm/pkg/tools/assert"
	"hcm/pkg/tools/converter"
	"hcm/pkg/tools/slice"
)

// SyncDiskOption ...
type SyncDiskOption struct {
	BootMap map[string]struct{}
}

// Validate ...
func (opt SyncDiskOption) Validate() error {
	return validator.Validate.Struct(opt)
}

// Disk 同步华为云磁盘资源
// 该方法负责将华为云上的磁盘资源与本地数据库中的磁盘数据进行同步
// 包括新增、更新和删除操作
func (cli *client) Disk(kt *kit.Kit, params *SyncBaseParams, opt *SyncDiskOption) (*SyncResult, error) {
	if err := validator.ValidateTool(params, opt); err != nil {
		return nil, errf.NewFromErr(errf.InvalidParameter, err)
	}

	diskFromCloud, err := cli.listDiskFromCloud(kt, params)
	if err != nil {
		return nil, err
	}

	diskFromDB, err := cli.listDiskFromDB(kt, params)
	if err != nil {
		return nil, err
	}

	if len(diskFromCloud) == 0 && len(diskFromDB) == 0 {
		return new(SyncResult), nil
	}

	if opt.BootMap != nil {
		for i, d := range diskFromCloud {
			_, exists := opt.BootMap[d.GetCloudID()]
			diskFromCloud[i].Boot = converter.ValToPtr(exists)
		}
	}

	addSlice, updateMap, delCloudIDs := common.Diff[adaptordisk.HuaWeiDisk, *coredisk.Disk[coredisk.HuaWeiExtension]](
		diskFromCloud, diskFromDB, isDiskChange)

	if len(delCloudIDs) > 0 {
		if err := cli.deleteDisk(kt, params.AccountID, params.Region, delCloudIDs); err != nil {
			return nil, err
		}
	}

	if len(addSlice) > 0 {
		if err = cli.createDisk(kt, params.AccountID, params.Region, addSlice); err != nil {
			return nil, err
		}
	}

	if len(updateMap) > 0 {
		if err = cli.updateDisk(kt, params.AccountID, updateMap); err != nil {
			return nil, err
		}
	}

	return new(SyncResult), nil
}

// deleteDisk 删除磁盘
// 该方法用于删除在云上已不存在但在数据库中仍有记录的磁盘
func (cli *client) deleteDisk(kt *kit.Kit, accountID string, region string, delCloudIDs []string) error {
	if len(delCloudIDs) <= 0 {
		return fmt.Errorf("delCloudIDs is <= 0, not delete")
	}

	checkParams := &SyncBaseParams{
		AccountID: accountID,
		Region:    region,
		CloudIDs:  delCloudIDs,
	}
	delDiskFromCloud, err := cli.listDiskFromCloud(kt, checkParams)
	if err != nil {
		return err
	}

	if len(delDiskFromCloud) > 0 {
		logs.Errorf("[%s] validate disk not exist failed, before delete, opt: %v, failed_count: %d, rid: %s",
			enumor.HuaWei, checkParams, len(delDiskFromCloud), kt.Rid)
		return fmt.Errorf("validate disk not exist failed, before delete")
	}

	deleteReq := &disk.DiskDeleteReq{
		Filter: tools.ContainersExpression("cloud_id", delCloudIDs),
	}
	if _, err = cli.dbCli.Global.DeleteDisk(kt.Ctx, kt.Header(), deleteReq); err != nil {
		logs.Errorf("[%s] request dataservice to batch delete disk failed, err: %v, rid: %s", enumor.HuaWei,
			err, kt.Rid)
		return err
	}

	logs.Infof("[%s] sync disk to delete disk success, accountID: %s, count: %d, rid: %s", enumor.HuaWei,
		accountID, len(delCloudIDs), kt.Rid)

	return nil
}

// updateDisk 更新磁盘信息
// 该方法用于更新数据库中磁盘的信息，使其与云上的最新状态保持一致
func (cli *client) updateDisk(kt *kit.Kit, accountID string, updateMap map[string]adaptordisk.HuaWeiDisk) error {

	if len(updateMap) <= 0 {
		return fmt.Errorf("updateMap is <= 0, not update")
	}

	disks := make([]*disk.DiskExtUpdateReq[coredisk.HuaWeiExtension], 0)

	for id, one := range updateMap {

		attachments := make([]*coredisk.HuaWeiDiskAttachment, 0)
		if len(one.Attachments) > 0 {
			for _, v := range one.Attachments {
				tmp := &coredisk.HuaWeiDiskAttachment{
					AttachedAt:   v.AttachedAt,
					AttachmentId: v.AttachmentId,
					DeviceName:   v.Device,
					HostName:     v.HostName,
					Id:           v.Id,
					InstanceId:   v.ServerId,
					DiskId:       v.VolumeId,
				}
				attachments = append(attachments, tmp)
			}
		}

		disk := &disk.DiskExtUpdateReq[coredisk.HuaWeiExtension]{
			ID:           id,
			Memo:         converter.ValToPtr(one.Description),
			Status:       one.Status,
			IsSystemDisk: one.Boot,
			Extension: &coredisk.HuaWeiExtension{
				ChargeType:  one.DiskChargeType,
				ExpireTime:  one.ExpireTime,
				ServiceType: one.ServiceType,
				Encrypted:   one.Encrypted,
				Attachment:  attachments,
				Bootable:    one.Bootable,
			},
		}

		disks = append(disks, disk)
	}

	var updateReq disk.DiskExtBatchUpdateReq[coredisk.HuaWeiExtension]
	for _, disk := range disks {
		updateReq = append(updateReq, disk)
	}
	if _, err := cli.dbCli.HuaWei.BatchUpdateDisk(kt.Ctx, kt.Header(), &updateReq); err != nil {
		logs.Errorf("[%s] request dataservice huawei BatchUpdateDisk failed, err: %v, rid: %s", enumor.HuaWei,
			err, kt.Rid)
		return err
	}

	logs.Infof("[%s] sync disk to update disk success, accountID: %s, count: %d, rid: %s", enumor.HuaWei,
		accountID, len(updateMap), kt.Rid)

	return nil
}

// createDisk 创建磁盘
// 该方法用于将云上新发现的磁盘信息同步到数据库中
func (cli *client) createDisk(kt *kit.Kit, accountID string, region string,
	addSlice []adaptordisk.HuaWeiDisk) error {

	if len(addSlice) <= 0 {
		return fmt.Errorf("addSlice is <= 0, not create")
	}

	var createReq disk.DiskExtBatchCreateReq[coredisk.HuaWeiExtension]

	for _, one := range addSlice {
		attachments := make([]*coredisk.HuaWeiDiskAttachment, 0)
		if len(one.Attachments) > 0 {
			for _, v := range one.Attachments {
				tmp := &coredisk.HuaWeiDiskAttachment{
					AttachedAt:   v.AttachedAt,
					AttachmentId: v.AttachmentId,
					DeviceName:   v.Device,
					HostName:     v.HostName,
					Id:           v.Id,
					InstanceId:   v.ServerId,
					DiskId:       v.VolumeId,
				}
				attachments = append(attachments, tmp)
			}
		}

		disk := &disk.DiskExtCreateReq[coredisk.HuaWeiExtension]{
			AccountID:    accountID,
			Name:         one.Name,
			CloudID:      one.Id,
			Region:       region,
			Zone:         one.AvailabilityZone,
			DiskSize:     uint64(one.Size),
			DiskType:     one.VolumeType,
			Status:       one.Status,
			Memo:         converter.ValToPtr(one.Description),
			IsSystemDisk: converter.PtrToVal(one.Boot),
			Extension: &coredisk.HuaWeiExtension{
				ChargeType:  one.DiskChargeType,
				ExpireTime:  one.ExpireTime,
				ServiceType: one.ServiceType,
				Encrypted:   one.Encrypted,
				Attachment:  attachments,
				Bootable:    one.Bootable,
			},
		}

		createReq = append(createReq, disk)
	}

	_, err := cli.dbCli.HuaWei.BatchCreateDisk(kt.Ctx, kt.Header(), &createReq)
	if err != nil {
		logs.Errorf("[%s] request dataservice to create huawei disk failed, err: %v, rid: %s", enumor.HuaWei,
			err, kt.Rid)
		return err
	}

	logs.Infof("[%s] sync disk to create disk success, accountID: %s, count: %d, rid: %s", enumor.HuaWei,
		accountID, len(addSlice), kt.Rid)

	return nil
}

// listDiskFromCloud 从华为云获取磁盘列表
// 该方法通过华为云API获取指定区域和云磁盘ID的磁盘信息
func (cli *client) listDiskFromCloud(kt *kit.Kit, params *SyncBaseParams) ([]adaptordisk.HuaWeiDisk, error) {
	if err := params.Validate(); err != nil {
		return nil, errf.NewFromErr(errf.InvalidParameter, err)
	}

	results := make([]adaptordisk.HuaWeiDisk, 0)
	elems := slice.Split(params.CloudIDs, 60)

	for _, partDisks := range elems {
		opt := &adaptordisk.HuaWeiDiskListOption{
			Region:   params.Region,
			CloudIDs: partDisks,
			Page: &adcore.HuaWeiPage{
				Limit: converter.ValToPtr(int32(adcore.HuaWeiQueryLimit)),
			},
		}
		result, err := cli.cloudCli.ListDisk(kt, opt)
		if err != nil {
			logs.Errorf("[%s] list disk from cloud failed, err: %v, account: %s, opt: %v, rid: %s", enumor.HuaWei,
				err, params.AccountID, opt, kt.Rid)
			return nil, err
		}

		results = append(results, result...)
	}

	return results, nil
}

// listDiskFromDB 从数据库获取磁盘列表
// 该方法根据账户ID、区域和云磁盘ID列表从数据库中查询磁盘信息
func (cli *client) listDiskFromDB(kt *kit.Kit, params *SyncBaseParams) (
	[]*coredisk.Disk[coredisk.HuaWeiExtension], error) {

	if err := params.Validate(); err != nil {
		return nil, errf.NewFromErr(errf.InvalidParameter, err)
	}

	req := &core.ListReq{
		Filter: &filter.Expression{
			Op: filter.And,
			Rules: []filter.RuleFactory{
				&filter.AtomRule{
					Field: "account_id",
					Op:    filter.Equal.Factory(),
					Value: params.AccountID,
				},
				&filter.AtomRule{
					Field: "cloud_id",
					Op:    filter.In.Factory(),
					Value: params.CloudIDs,
				},
				&filter.AtomRule{
					Field: "region",
					Op:    filter.Equal.Factory(),
					Value: params.Region,
				},
			},
		},
		Page: core.NewDefaultBasePage(),
	}
	result, err := cli.dbCli.HuaWei.ListDisk(kt.Ctx, kt.Header(), req)
	if err != nil {
		logs.Errorf("[%s] list disk from db failed, err: %v, account: %s, req: %v, rid: %s", enumor.HuaWei,
			err, params.AccountID, req, kt.Rid)
		return nil, err
	}

	return result.Details, nil
}

// RemoveDiskDeleteFromCloud 清理云上已删除的磁盘
// 该方法用于检查并删除在云上已不存在但在数据库中仍有记录的磁盘
func (cli *client) RemoveDiskDeleteFromCloud(kt *kit.Kit, accountID string, region string) error {
	req := &core.ListReq{
		Fields: []string{"id", "cloud_id"},
		Filter: &filter.Expression{
			Op: filter.And,
			Rules: []filter.RuleFactory{
				&filter.AtomRule{Field: "account_id", Op: filter.Equal.Factory(), Value: accountID},
				&filter.AtomRule{Field: "region", Op: filter.Equal.Factory(), Value: region},
			},
		},
		Page: &core.BasePage{
			Start: 0,
			Limit: constant.BatchOperationMaxLimit,
		},
	}
	for {
		resultFromDB, err := cli.dbCli.Global.ListDisk(kt, req)
		if err != nil {
			logs.Errorf("[%s] request dataservice to list disk failed, err: %v, req: %v, rid: %s", enumor.HuaWei,
				err, req, kt.Rid)
			return err
		}

		cloudIDs := make([]string, 0)
		for _, one := range resultFromDB.Details {
			cloudIDs = append(cloudIDs, one.CloudID)
		}

		if len(cloudIDs) == 0 {
			break
		}

		params := &SyncBaseParams{
			AccountID: accountID,
			Region:    region,
			CloudIDs:  cloudIDs,
		}
		resultFromCloud, err := cli.listDiskFromCloud(kt, params)
		if err != nil {
			return err
		}

		// 如果有资源没有查询出来，说明数据被从云上删除
		if len(resultFromCloud) != len(cloudIDs) {
			cloudIDMap := converter.StringSliceToMap(cloudIDs)
			for _, one := range resultFromCloud {
				delete(cloudIDMap, one.Id)
			}

			cloudIDs := converter.MapKeyToStringSlice(cloudIDMap)
			if len(cloudIDs) > 0 {
				if err := cli.deleteDisk(kt, accountID, region, cloudIDs); err != nil {
					return err
				}
			}
		}

		if len(resultFromDB.Details) < constant.BatchOperationMaxLimit {
			break
		}

		req.Page.Start += constant.BatchOperationMaxLimit
	}

	return nil
}

// isDiskChange 判断磁盘是否发生变化
// 该函数比较云上磁盘和数据库中磁盘的各个属性，判断是否需要更新
// 返回true表示磁盘信息有变化，需要更新；返回false表示无变化
func isDiskChange(cloud adaptordisk.HuaWeiDisk, db *coredisk.Disk[coredisk.HuaWeiExtension]) bool {

	if cloud.Status != db.Status {
		return true
	}

	if cloud.Description != converter.PtrToVal(db.Memo) {
		return true
	}

	if cloud.ServiceType != db.Extension.ServiceType {
		return true
	}

	if !assert.IsPtrBoolEqual(cloud.Encrypted, db.Extension.Encrypted) {
		return true
	}

	if cloud.Bootable != db.Extension.Bootable {
		return true
	}

	for _, dbValue := range db.Extension.Attachment {
		isEqual := false
		for _, cloudValue := range cloud.Attachments {
			if dbValue.AttachedAt == cloudValue.AttachedAt && dbValue.AttachmentId == cloudValue.AttachmentId &&
				dbValue.DeviceName == cloudValue.Device && dbValue.HostName == cloudValue.HostName &&
				dbValue.Id == cloudValue.Id && dbValue.InstanceId == cloudValue.ServerId &&
				dbValue.DiskId == cloudValue.VolumeId {
				isEqual = true
				break
			}
		}
		if !isEqual {
			return true
		}
	}

	if cloud.DiskChargeType != db.Extension.ChargeType {
		return true
	}

	if cloud.ExpireTime != db.Extension.ExpireTime {
		return true
	}

	if cloud.Boot != nil && *cloud.Boot != db.IsSystemDisk {
		return true
	}

	return false
}
