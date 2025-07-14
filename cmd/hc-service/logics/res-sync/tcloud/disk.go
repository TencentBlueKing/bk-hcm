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

package tcloud

import (
	"fmt"

	"hcm/cmd/hc-service/logics/res-sync/common"
	adcore "hcm/pkg/adaptor/types/core"
	typesdisk "hcm/pkg/adaptor/types/disk"
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
)

// SyncDiskOption ...
type SyncDiskOption struct {
}

// Validate ...
func (opt SyncDiskOption) Validate() error {
	return validator.Validate.Struct(opt)
}

// Disk ...
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

	addSlice, updateMap, delCloudIDs := common.Diff[typesdisk.TCloudDisk, *coredisk.Disk[coredisk.TCloudExtension]](
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
			enumor.TCloud, checkParams, len(delDiskFromCloud), kt.Rid)
		return fmt.Errorf("validate disk not exist failed, before delete")
	}

	deleteReq := &disk.DiskDeleteReq{
		Filter: tools.ContainersExpression("cloud_id", delCloudIDs),
	}
	if _, err = cli.dbCli.Global.DeleteDisk(kt.Ctx, kt.Header(), deleteReq); err != nil {
		logs.Errorf("[%s] request dataservice to batch delete disk failed, err: %v, rid: %s", enumor.TCloud,
			err, kt.Rid)
		return err
	}

	logs.Infof("[%s] sync disk to delete disk success, accountID: %s, count: %d, rid: %s", enumor.TCloud,
		accountID, len(delCloudIDs), kt.Rid)

	return nil
}

func (cli *client) updateDisk(kt *kit.Kit, accountID string, updateMap map[string]typesdisk.TCloudDisk) error {
	if len(updateMap) <= 0 {
		return fmt.Errorf("updateMap is <= 0, not update")
	}

	disks := make([]*disk.DiskExtUpdateReq[coredisk.TCloudExtension], 0)

	for id, one := range updateMap {
		disk := &disk.DiskExtUpdateReq[coredisk.TCloudExtension]{
			ID:           id,
			Status:       converter.PtrToVal(one.DiskState),
			IsSystemDisk: converter.ValToPtr(one.Boot),
			Extension: &coredisk.TCloudExtension{
				DiskChargeType: converter.PtrToVal(one.DiskChargeType),
				DiskChargePrepaid: &coredisk.TCloudDiskChargePrepaid{
					RenewFlag: one.RenewFlag,
					Period:    one.DifferDaysOfDeadline,
				},
				Encrypted:          one.Encrypt,
				Attached:           one.Attached,
				DiskUsage:          one.DiskUsage,
				InstanceId:         one.InstanceId,
				InstanceType:       one.InstanceType,
				DeleteWithInstance: one.DeleteWithInstance,
				DeadlineTime:       one.DeadlineTime,
				BackupDisk:         one.BackupDisk,
			},
		}

		disks = append(disks, disk)
	}

	var updateReq disk.DiskExtBatchUpdateReq[coredisk.TCloudExtension]
	for _, disk := range disks {
		updateReq = append(updateReq, disk)
	}
	if _, err := cli.dbCli.TCloud.BatchUpdateDisk(kt.Ctx, kt.Header(), &updateReq); err != nil {
		logs.Errorf("[%s] request dataservice BatchUpdateDisk failed, err: %v, rid: %s", enumor.TCloud,
			err, kt.Rid)
		return err
	}

	logs.Infof("[%s] sync disk to update disk success, accountID: %s, count: %d, rid: %s", enumor.TCloud,
		accountID, len(updateMap), kt.Rid)

	return nil
}

func (cli *client) createDisk(kt *kit.Kit, accountID string, region string,
	addSlice []typesdisk.TCloudDisk) error {

	if len(addSlice) <= 0 {
		return fmt.Errorf("addSlice is <= 0, not create")
	}

	var createReq disk.DiskExtBatchCreateReq[coredisk.TCloudExtension]

	for _, one := range addSlice {
		disk := &disk.DiskExtCreateReq[coredisk.TCloudExtension]{
			AccountID: accountID,
			Name:      converter.PtrToVal(one.DiskName),
			CloudID:   converter.PtrToVal(one.DiskId),
			Region:    region,
			Zone:      converter.PtrToVal(one.Placement.Zone),
			DiskSize:  converter.PtrToVal(one.DiskSize),
			DiskType:  converter.PtrToVal(one.DiskType),
			Status:    converter.PtrToVal(one.DiskState),
			// tcloud no memo
			Memo: nil,
			Extension: &coredisk.TCloudExtension{
				DiskChargeType: converter.PtrToVal(one.DiskChargeType),
				DiskChargePrepaid: &coredisk.TCloudDiskChargePrepaid{
					RenewFlag: one.RenewFlag,
					Period:    one.DifferDaysOfDeadline,
				},
				Encrypted:          one.Encrypt,
				Attached:           one.Attached,
				DiskUsage:          one.DiskUsage,
				InstanceId:         one.InstanceId,
				InstanceType:       one.InstanceType,
				DeleteWithInstance: one.DeleteWithInstance,
				DeadlineTime:       one.DeadlineTime,
				BackupDisk:         one.BackupDisk,
			},
		}

		if one.DiskUsage != nil && converter.PtrToVal(one.DiskUsage) == "SYSTEM_DISK" {
			disk.IsSystemDisk = true
		}

		createReq = append(createReq, disk)
	}

	_, err := cli.dbCli.TCloud.BatchCreateDisk(kt.Ctx, kt.Header(), &createReq)
	if err != nil {
		logs.Errorf("[%s] request dataservice to create tcloud disk failed, err: %v, rid: %s", enumor.TCloud,
			err, kt.Rid)
		return err
	}

	logs.Infof("[%s] sync disk to create disk success, accountID: %s, count: %d, rid: %s", enumor.TCloud,
		accountID, len(addSlice), kt.Rid)

	return nil
}

func (cli *client) listDiskFromCloud(kt *kit.Kit, params *SyncBaseParams) ([]typesdisk.TCloudDisk, error) {
	if err := params.Validate(); err != nil {
		return nil, errf.NewFromErr(errf.InvalidParameter, err)
	}

	opt := &adcore.TCloudListOption{
		Region:   params.Region,
		CloudIDs: params.CloudIDs,
		Page: &adcore.TCloudPage{
			Offset: 0,
			Limit:  adcore.TCloudQueryLimit,
		},
	}
	result, err := cli.cloudCli.ListDisk(kt, opt)
	if err != nil {
		logs.Errorf("[%s] list disk from cloud failed, err: %v, account: %s, opt: %v, rid: %s", enumor.TCloud,
			err, params.AccountID, opt, kt.Rid)
		return nil, err
	}

	return result, nil
}

func (cli *client) listDiskFromDB(kt *kit.Kit, params *SyncBaseParams) (
	[]*coredisk.Disk[coredisk.TCloudExtension], error) {

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
	result, err := cli.dbCli.TCloud.ListDisk(kt.Ctx, kt.Header(), req)
	if err != nil {
		logs.Errorf("[%s] list disk from db failed, err: %v, account: %s, req: %v, rid: %s", enumor.TCloud,
			err, params.AccountID, req, kt.Rid)
		return nil, err
	}

	return result.Details, nil
}

func isDiskChange(cloud typesdisk.TCloudDisk, db *coredisk.Disk[coredisk.TCloudExtension]) bool {

	if converter.PtrToVal(cloud.DiskState) != db.Status {
		return true
	}

	if converter.PtrToVal(cloud.DiskChargeType) != db.Extension.DiskChargeType {
		return true
	}

	if !assert.IsPtrBoolEqual(cloud.Encrypt, db.Extension.Encrypted) {
		return true
	}

	if !assert.IsPtrBoolEqual(cloud.Attached, db.Extension.Attached) {
		return true
	}

	if !assert.IsPtrStringEqual(cloud.DiskUsage, db.Extension.DiskUsage) {
		return true
	}

	if cloud.Boot != db.IsSystemDisk {
		return true
	}

	if !assert.IsPtrStringEqual(cloud.InstanceId, db.Extension.InstanceId) {
		return true
	}

	if !assert.IsPtrStringEqual(cloud.InstanceType, db.Extension.InstanceType) {
		return true
	}

	if !assert.IsPtrStringEqual(cloud.RenewFlag, db.Extension.DiskChargePrepaid.RenewFlag) {
		return true
	}

	if !assert.IsPtrInt64Equal(cloud.DifferDaysOfDeadline, db.Extension.DiskChargePrepaid.Period) {
		return true
	}

	if !assert.IsPtrStringEqual(cloud.DeadlineTime, db.Extension.DeadlineTime) {
		return true
	}

	if !assert.IsPtrBoolEqual(cloud.DeleteWithInstance, db.Extension.DeleteWithInstance) {
		return true
	}

	if !assert.IsPtrBoolEqual(cloud.BackupDisk, db.Extension.BackupDisk) {
		return true
	}

	return false
}

// RemoveDiskDeleteFromCloud ...
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
			logs.Errorf("[%s] request dataservice to list disk failed, err: %v, req: %v, rid: %s", enumor.TCloud,
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
				delete(cloudIDMap, converter.PtrToVal(one.DiskId))
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
