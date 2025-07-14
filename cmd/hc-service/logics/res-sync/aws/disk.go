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

package aws

import (
	"fmt"
	"strings"

	"hcm/cmd/hc-service/logics/res-sync/common"
	"hcm/pkg/adaptor/aws"
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
	"hcm/pkg/tools/times"
)

// SyncDiskOption ...
type SyncDiskOption struct {
	BootMap map[string]struct{}
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
	if opt.BootMap != nil {
		// 标记启动盘
		for i, d := range diskFromCloud {
			_, exists := opt.BootMap[d.GetCloudID()]
			diskFromCloud[i].Boot = converter.ValToPtr(exists)
		}
	}
	addSlice, updateMap, delCloudIDs := common.Diff[adaptordisk.AwsDisk, *coredisk.Disk[coredisk.AwsExtension]](
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

func (cli *client) updateDisk(kt *kit.Kit, accountID string, updateMap map[string]adaptordisk.AwsDisk) error {

	if len(updateMap) <= 0 {
		return fmt.Errorf("updateMap is <= 0, not update")
	}

	disks := make([]*disk.DiskExtUpdateReq[coredisk.AwsExtension], 0)

	for id, one := range updateMap {
		name := ""
		for _, tag := range one.Tags {
			if tag != nil {
				if converter.PtrToVal(tag.Key) == "ImportMode" {
					name = converter.PtrToVal(tag.Value)
				}
			}
		}

		attachments := make([]*coredisk.AwsDiskAttachment, 0)
		if len(one.Attachments) > 0 {
			for _, v := range one.Attachments {
				if v != nil {
					tmp := &coredisk.AwsDiskAttachment{
						AttachTime:          times.ConvStdTimeFormat(*v.AttachTime),
						DeleteOnTermination: v.DeleteOnTermination,
						DeviceName:          v.Device,
						InstanceId:          v.InstanceId,
						Status:              v.State,
						DiskId:              v.VolumeId,
					}
					attachments = append(attachments, tmp)
				}
			}
		}

		disk := &disk.DiskExtUpdateReq[coredisk.AwsExtension]{
			ID:           id,
			Status:       converter.PtrToVal(one.State),
			Name:         name,
			IsSystemDisk: one.Boot,
			Extension: &coredisk.AwsExtension{
				Attachment: attachments,
				Encrypted:  one.Encrypted,
			},
		}

		disks = append(disks, disk)
	}

	var updateReq disk.DiskExtBatchUpdateReq[coredisk.AwsExtension]
	for _, disk := range disks {
		updateReq = append(updateReq, disk)
	}
	if _, err := cli.dbCli.Aws.BatchUpdateDisk(kt.Ctx, kt.Header(), &updateReq); err != nil {
		logs.Errorf("[%s] request dataservice aws BatchUpdateDisk failed, err: %v, rid: %s", enumor.Aws,
			err, kt.Rid)
		return err
	}

	logs.Infof("[%s] sync disk to update disk success, accountID: %s, count: %d, rid: %s", enumor.Aws,
		accountID, len(updateMap), kt.Rid)

	return nil
}

func (cli *client) createDisk(kt *kit.Kit, accountID string, region string, addSlice []adaptordisk.AwsDisk) error {

	if len(addSlice) <= 0 {
		return fmt.Errorf("addSlice is <= 0, not create")
	}

	var createReq disk.DiskExtBatchCreateReq[coredisk.AwsExtension]

	for _, one := range addSlice {
		attachments := make([]*coredisk.AwsDiskAttachment, 0)
		if len(one.Attachments) > 0 {
			for _, v := range one.Attachments {
				if v != nil {
					tmp := &coredisk.AwsDiskAttachment{
						AttachTime:          times.ConvStdTimeFormat(*v.AttachTime),
						DeleteOnTermination: v.DeleteOnTermination,
						DeviceName:          v.Device,
						InstanceId:          v.InstanceId,
						Status:              v.State,
						DiskId:              v.VolumeId,
					}
					attachments = append(attachments, tmp)
				}
			}
		}

		name := ""
		for _, tag := range one.Tags {
			if tag != nil {
				if converter.PtrToVal(tag.Key) == "ImportMode" {
					name = converter.PtrToVal(tag.Value)
				}
			}
		}

		disk := &disk.DiskExtCreateReq[coredisk.AwsExtension]{
			AccountID:    accountID,
			Name:         name,
			CloudID:      converter.PtrToVal(one.VolumeId),
			Region:       region,
			Zone:         converter.PtrToVal(one.AvailabilityZone),
			DiskSize:     uint64(converter.PtrToVal(one.Size)),
			DiskType:     converter.PtrToVal(one.VolumeType),
			Status:       converter.PtrToVal(one.State),
			IsSystemDisk: converter.PtrToVal(one.Boot),
			// 该云没有此字段
			Memo: nil,
			Extension: &coredisk.AwsExtension{
				Attachment: attachments,
				Encrypted:  one.Encrypted,
			},
		}

		createReq = append(createReq, disk)
	}

	_, err := cli.dbCli.Aws.BatchCreateDisk(kt.Ctx, kt.Header(), &createReq)
	if err != nil {
		logs.Errorf("[%s] request dataservice to create aws disk failed, err: %v, rid: %s", enumor.Aws,
			err, kt.Rid)
		return err
	}

	logs.Infof("[%s] sync disk to create disk success, accountID: %s, count: %d, rid: %s", enumor.Aws,
		accountID, len(addSlice), kt.Rid)

	return nil
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
			enumor.Aws, checkParams, len(delDiskFromCloud), kt.Rid)
		return fmt.Errorf("validate disk not exist failed, before delete")
	}

	deleteReq := &disk.DiskDeleteReq{
		Filter: tools.ContainersExpression("cloud_id", delCloudIDs),
	}
	if _, err = cli.dbCli.Global.DeleteDisk(kt.Ctx, kt.Header(), deleteReq); err != nil {
		logs.Errorf("[%s] request dataservice to batch delete disk failed, err: %v, rid: %s", enumor.Aws,
			err, kt.Rid)
		return err
	}

	logs.Infof("[%s] sync disk to delete disk success, accountID: %s, count: %d, rid: %s", enumor.Aws,
		accountID, len(delCloudIDs), kt.Rid)

	return nil
}

func (cli *client) listDiskFromCloud(kt *kit.Kit, params *SyncBaseParams) ([]adaptordisk.AwsDisk, error) {
	if err := params.Validate(); err != nil {
		return nil, errf.NewFromErr(errf.InvalidParameter, err)
	}

	opt := &adaptordisk.AwsDiskListOption{
		Region:   params.Region,
		CloudIDs: params.CloudIDs,
	}

	result, _, err := cli.cloudCli.ListDisk(kt, opt)
	if err != nil {
		if strings.Contains(err.Error(), aws.ErrDiskNotFound) {
			return make([]adaptordisk.AwsDisk, 0), nil
		}

		logs.Errorf("[%s] list disk from cloud failed, err: %v, account: %s, opt: %v, rid: %s", enumor.Aws, err,
			params.AccountID, opt, kt.Rid)
		return nil, err
	}

	return result, nil
}

func (cli *client) listDiskFromDB(kt *kit.Kit, params *SyncBaseParams) (
	[]*coredisk.Disk[coredisk.AwsExtension], error) {

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
	result, err := cli.dbCli.Aws.ListDisk(kt.Ctx, kt.Header(), req)
	if err != nil {
		logs.Errorf("[%s] list disk from db failed, err: %v, account: %s, req: %v, rid: %s", enumor.Aws,
			err, params.AccountID, req, kt.Rid)
		return nil, err
	}

	return result.Details, nil
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
			logs.Errorf("[%s] request dataservice to list disk failed, err: %v, req: %v, rid: %s", enumor.Aws,
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

		var delCloudIDs []string
		if len(cloudIDs) != 0 {
			params := &SyncBaseParams{
				AccountID: accountID,
				Region:    region,
				CloudIDs:  cloudIDs,
			}
			delCloudIDs, err = cli.listRemoveDiskID(kt, params)
			if err != nil {
				return err
			}
		}

		if len(delCloudIDs) != 0 {
			if err = cli.deleteDisk(kt, accountID, region, delCloudIDs); err != nil {
				return err
			}
		}

		if len(resultFromDB.Details) < constant.BatchOperationMaxLimit {
			break
		}

		req.Page.Start += constant.BatchOperationMaxLimit
	}

	return nil
}

func (cli *client) listRemoveDiskID(kt *kit.Kit, params *SyncBaseParams) ([]string, error) {
	if err := params.Validate(); err != nil {
		return nil, errf.NewFromErr(errf.InvalidParameter, err)
	}

	delCloudIDs := make([]string, 0)
	cloudIDs := params.CloudIDs
	for {
		opt := &adaptordisk.AwsDiskListOption{
			Region:   params.Region,
			CloudIDs: cloudIDs,
		}
		_, _, err := cli.cloudCli.ListDisk(kt, opt)
		if err != nil {
			if strings.Contains(err.Error(), aws.ErrDiskNotFound) {
				var delCloudID string
				cloudIDs, delCloudID = removeNotFoundCloudID(cloudIDs, err)
				delCloudIDs = append(delCloudIDs, delCloudID)

				if len(cloudIDs) <= 0 {
					break
				}

				continue
			}

			logs.Errorf("[%s] list disk from cloud failed, err: %v, account: %s, opt: %v, rid: %s", enumor.Aws, err,
				params.AccountID, opt, kt.Rid)
			return nil, err
		}

		break
	}

	return delCloudIDs, nil
}

func isDiskChange(cloud adaptordisk.AwsDisk, db *coredisk.Disk[coredisk.AwsExtension]) bool {

	if converter.PtrToVal(cloud.State) != db.Status {
		return true
	}

	if !assert.IsPtrBoolEqual(cloud.Encrypted, db.Extension.Encrypted) {
		return true
	}

	if cloud.Boot != nil && *cloud.Boot != db.IsSystemDisk {
		return true
	}
	for _, dbValue := range db.Extension.Attachment {
		isEqual := false
		for _, cloudValue := range cloud.Attachments {
			if dbValue.AttachTime == times.ConvStdTimeFormat(*cloudValue.AttachTime) &&
				assert.IsPtrBoolEqual(dbValue.DeleteOnTermination, cloudValue.DeleteOnTermination) &&
				assert.IsPtrStringEqual(dbValue.DeviceName, cloudValue.Device) &&
				assert.IsPtrStringEqual(dbValue.InstanceId, cloudValue.InstanceId) &&
				assert.IsPtrStringEqual(dbValue.Status, cloudValue.State) &&
				assert.IsPtrStringEqual(dbValue.DiskId, cloudValue.VolumeId) {
				isEqual = true
				break
			}
		}
		if !isEqual {
			return true
		}
	}

	return false
}
