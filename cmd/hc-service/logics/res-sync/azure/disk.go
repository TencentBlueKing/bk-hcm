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

package azure

import (
	"fmt"

	"hcm/cmd/hc-service/logics/res-sync/common"
	typescore "hcm/pkg/adaptor/types/core"
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
	addSlice, updateMap, delCloudIDs := common.Diff[typesdisk.AzureDisk, *coredisk.Disk[coredisk.AzureExtension]](
		diskFromCloud, diskFromDB, isDiskChange)

	if len(delCloudIDs) > 0 {
		if err := cli.deleteDisk(kt, params.AccountID, params.ResourceGroupName, delCloudIDs); err != nil {
			return nil, err
		}
	}

	if len(addSlice) > 0 {
		if err = cli.createDisk(kt, params.AccountID, params.ResourceGroupName, addSlice); err != nil {
			return nil, err
		}
	}

	if len(updateMap) > 0 {
		if err = cli.updateDisk(kt, params.AccountID, params.ResourceGroupName, updateMap); err != nil {
			return nil, err
		}
	}

	return new(SyncResult), nil
}

func (cli *client) updateDisk(kt *kit.Kit, accountID string, resGroupName string,
	updateMap map[string]typesdisk.AzureDisk) error {

	if len(updateMap) <= 0 {
		return fmt.Errorf("updateMap is <= 0, not update")
	}

	disks := make([]*disk.DiskExtUpdateReq[coredisk.AzureExtension], 0)

	for id, one := range updateMap {

		oneDisk := &disk.DiskExtUpdateReq[coredisk.AzureExtension]{
			ID:           id,
			Status:       converter.PtrToVal(one.Status),
			IsSystemDisk: one.Boot,
			Extension: &coredisk.AzureExtension{
				ResourceGroupName: resGroupName,
				OSType:            converter.PtrToVal(one.OSType),
				SKUName:           one.SKUName,
				SKUTier:           one.SKUTier,
				Zones:             one.Zones,
			},
		}
		disks = append(disks, oneDisk)
	}

	var updateReq disk.DiskExtBatchUpdateReq[coredisk.AzureExtension]
	for _, disk := range disks {
		updateReq = append(updateReq, disk)
	}
	if _, err := cli.dbCli.Azure.BatchUpdateDisk(kt.Ctx, kt.Header(), &updateReq); err != nil {
		logs.Errorf("[%s] request dataservice azure BatchUpdateDisk failed, err: %v, rid: %s", enumor.Azure,
			err, kt.Rid)
		return err
	}

	logs.Infof("[%s] sync disk to update disk success, accountID: %s, count: %d, rid: %s", enumor.Azure,
		accountID, len(updateMap), kt.Rid)

	return nil
}

func (cli *client) createDisk(kt *kit.Kit, accountID string, resGroupName string,
	addSlice []typesdisk.AzureDisk) error {

	if len(addSlice) <= 0 {
		return fmt.Errorf("addSlice is <= 0, not create")
	}

	var createReq disk.DiskExtBatchCreateReq[coredisk.AzureExtension]

	for _, one := range addSlice {
		disk := &disk.DiskExtCreateReq[coredisk.AzureExtension]{
			AccountID:    accountID,
			Name:         converter.PtrToVal(one.Name),
			CloudID:      converter.PtrToVal(one.ID),
			Region:       converter.PtrToVal(one.Location),
			DiskSize:     uint64(converter.PtrToVal(one.DiskSize)) / 1024 / 1024 / 1024,
			DiskType:     converter.PtrToVal(one.Type),
			Status:       string(converter.PtrToVal(one.Status)),
			Zone:         "",
			IsSystemDisk: converter.PtrToVal(one.Boot),
			// 该云没有此字段
			Memo: nil,
			Extension: &coredisk.AzureExtension{
				ResourceGroupName: resGroupName,
				OSType:            converter.PtrToVal(one.OSType),
				SKUName:           one.SKUName,
				SKUTier:           one.SKUTier,
				Zones:             one.Zones,
			},
		}

		createReq = append(createReq, disk)
	}

	_, err := cli.dbCli.Azure.BatchCreateDisk(kt.Ctx, kt.Header(), &createReq)
	if err != nil {
		logs.Errorf("[%s] request dataservice to create azure disk failed, err: %v, rid: %s", enumor.Azure,
			err, kt.Rid)
		return err
	}

	logs.Infof("[%s] sync disk to create disk success, accountID: %s, count: %d, rid: %s", enumor.Azure,
		accountID, len(addSlice), kt.Rid)

	return nil
}

func (cli *client) deleteDisk(kt *kit.Kit, accountID string, resGroupName string,
	delCloudIDs []string) error {

	if len(delCloudIDs) <= 0 {
		return fmt.Errorf("delCloudIDs is <= 0, not delete")
	}

	checkParams := &SyncBaseParams{
		AccountID:         accountID,
		ResourceGroupName: resGroupName,
		CloudIDs:          delCloudIDs,
	}
	delDiskFromCloud, err := cli.listDiskFromCloud(kt, checkParams)
	if err != nil {
		return err
	}

	if len(delDiskFromCloud) > 0 {
		logs.Errorf("[%s] validate disk not exist failed, before delete, opt: %v, failed_count: %d, rid: %s",
			enumor.Azure, checkParams, len(delDiskFromCloud), kt.Rid)
		return fmt.Errorf("validate disk not exist failed, before delete")
	}

	deleteReq := &disk.DiskDeleteReq{
		Filter: tools.ContainersExpression("cloud_id", delCloudIDs),
	}
	if _, err = cli.dbCli.Global.DeleteDisk(kt.Ctx, kt.Header(), deleteReq); err != nil {
		logs.Errorf("[%s] request dataservice to batch delete disk failed, err: %v, rid: %s", enumor.Azure,
			err, kt.Rid)
		return err
	}

	logs.Infof("[%s] sync disk to delete disk success, accountID: %s, count: %d, rid: %s", enumor.Azure,
		accountID, len(delCloudIDs), kt.Rid)

	return nil
}

func (cli *client) listDiskFromCloud(kt *kit.Kit, params *SyncBaseParams) ([]typesdisk.AzureDisk, error) {
	if err := params.Validate(); err != nil {
		return nil, errf.NewFromErr(errf.InvalidParameter, err)
	}

	opt := &typescore.AzureListByIDOption{
		ResourceGroupName: params.ResourceGroupName,
		CloudIDs:          params.CloudIDs,
	}
	result, err := cli.cloudCli.ListDiskByID(kt, opt)
	if err != nil {
		logs.Errorf("[%s] list disk from cloud failed, err: %v, account: %s, opt: %v, rid: %s", enumor.Azure,
			err, params.AccountID, opt, kt.Rid)
		return nil, err
	}

	disks := make([]typesdisk.AzureDisk, 0, len(result))
	for _, one := range result {
		disks = append(disks, converter.PtrToVal(one))
	}

	return disks, nil
}

func (cli *client) listDiskFromDB(kt *kit.Kit, params *SyncBaseParams) (
	[]*coredisk.Disk[coredisk.AzureExtension], error) {

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
					Field: "extension.resource_group_name",
					Op:    filter.JSONEqual.Factory(),
					Value: params.ResourceGroupName,
				},
			},
		},
		Page: core.NewDefaultBasePage(),
	}
	result, err := cli.dbCli.Azure.ListDisk(kt.Ctx, kt.Header(), req)
	if err != nil {
		logs.Errorf("[%s] list disk from db failed, err: %v, account: %s, req: %v, rid: %s", enumor.Azure,
			err, params.AccountID, req, kt.Rid)
		return nil, err
	}

	return result.Details, nil
}

// RemoveDiskDeleteFromCloud ...
func (cli *client) RemoveDiskDeleteFromCloud(kt *kit.Kit, accountID string, resGroupName string) error {
	req := &core.ListReq{
		Fields: []string{"id", "cloud_id"},
		Filter: &filter.Expression{
			Op: filter.And,
			Rules: []filter.RuleFactory{
				&filter.AtomRule{Field: "account_id", Op: filter.Equal.Factory(), Value: accountID},
				&filter.AtomRule{Field: "extension.resource_group_name", Op: filter.JSONEqual.Factory(),
					Value: resGroupName},
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
			logs.Errorf("[%s] request dataservice to list disk failed, err: %v, req: %v, rid: %s", enumor.Azure,
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
			AccountID:         accountID,
			ResourceGroupName: resGroupName,
			CloudIDs:          cloudIDs,
		}
		resultFromCloud, err := cli.listDiskFromCloud(kt, params)
		if err != nil {
			return err
		}

		// 如果有资源没有查询出来，说明数据被从云上删除
		if len(resultFromCloud) != len(cloudIDs) {
			cloudIDMap := converter.StringSliceToMap(cloudIDs)
			for _, one := range resultFromCloud {
				delete(cloudIDMap, converter.PtrToVal(one.ID))
			}

			cloudIDs := converter.MapKeyToStringSlice(cloudIDMap)
			if len(cloudIDs) > 0 {
				if err := cli.deleteDisk(kt, accountID, resGroupName, cloudIDs); err != nil {
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

func isDiskChange(cloud typesdisk.AzureDisk, db *coredisk.Disk[coredisk.AzureExtension]) bool {

	if converter.PtrToVal(cloud.Status) != db.Status {
		return true
	}

	if !assert.IsPtrStringEqual(cloud.SKUName, db.Extension.SKUName) {
		return true
	}

	if !assert.IsPtrStringEqual(cloud.SKUTier, db.Extension.SKUTier) {
		return true
	}

	if converter.PtrToVal(cloud.OSType) != db.Extension.OSType {
		return true
	}

	if !assert.IsPtrStringSliceEqual(cloud.Zones, db.Extension.Zones) {
		return true
	}
	if cloud.Boot != nil && *cloud.Boot != db.IsSystemDisk {
		return true
	}

	return false
}
