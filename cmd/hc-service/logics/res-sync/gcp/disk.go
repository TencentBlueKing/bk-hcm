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

package gcp

import (
	"fmt"

	"hcm/cmd/hc-service/logics/res-sync/common"
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
	"hcm/pkg/tools/converter"
)

// SyncDiskOption ...
type SyncDiskOption struct {
	BootMap map[string]struct{} `json:"boot_map" validate:"omitempty"`
	Zone    string              `json:"zone" validate:"required"`
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

	diskFromCloud, err := cli.listDiskFromCloud(kt, params, opt)
	if err != nil {
		return nil, err
	}

	diskFromDB, err := cli.listDiskFromDB(kt, params, opt)
	if err != nil {
		return nil, err
	}

	if len(diskFromCloud) == 0 && len(diskFromDB) == 0 {
		return new(SyncResult), nil
	}

	if opt.BootMap != nil {
		// 标记启动盘
		for i, d := range diskFromCloud {
			_, exists := opt.BootMap[d.SelfLink]
			diskFromCloud[i].Boot = converter.ValToPtr(exists)

		}
	}

	addSlice, updateMap, delCloudIDs := common.Diff[adaptordisk.GcpDisk, *coredisk.Disk[coredisk.GcpExtension]](
		diskFromCloud, diskFromDB, isDiskChange)

	if len(delCloudIDs) > 0 {
		if err := cli.deleteDisk(kt, params.AccountID, opt.Zone, delCloudIDs); err != nil {
			return nil, err
		}
	}

	if len(addSlice) > 0 {
		if err = cli.createDisk(kt, params.AccountID, opt.Zone, addSlice); err != nil {
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

func (cli *client) listDiskFromCloud(kt *kit.Kit, params *SyncBaseParams,
	syncOpt *SyncDiskOption) ([]adaptordisk.GcpDisk, error) {

	if err := params.Validate(); err != nil {
		return nil, errf.NewFromErr(errf.InvalidParameter, err)
	}

	listOpt := &adaptordisk.GcpDiskListOption{
		Zone:     syncOpt.Zone,
		CloudIDs: params.CloudIDs,
	}
	result, _, err := cli.cloudCli.ListDisk(kt, listOpt)
	if err != nil {
		logs.Errorf("[%s] list disk from cloud failed, err: %v, account: %s, listOpt: %v, rid: %s", enumor.Gcp,
			err, params.AccountID, listOpt, kt.Rid)
		return nil, err
	}

	return result, nil
}

func (cli *client) listDiskFromDBBySelfLink(kt *kit.Kit, params *ListDiskBySelfLinkOption) (
	[]*coredisk.Disk[coredisk.GcpExtension], error) {

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
					Field: "extension.self_link",
					Op:    filter.JSONIn.Factory(),
					Value: params.SelfLink,
				},
				&filter.AtomRule{
					Field: "zone",
					Op:    filter.Equal.Factory(),
					Value: params.Zone,
				},
			},
		},
		Page: core.NewDefaultBasePage(),
	}
	result, err := cli.dbCli.Gcp.ListDisk(kt.Ctx, kt.Header(), req)
	if err != nil {
		logs.Errorf("[%s] list disk from db failed, err: %v, account: %s, req: %v, rid: %s", enumor.Gcp,
			err, params.AccountID, req, kt.Rid)
		return nil, err
	}

	return result.Details, nil
}

func (cli *client) listDiskFromDB(kt *kit.Kit, params *SyncBaseParams, option *SyncDiskOption) (
	[]*coredisk.Disk[coredisk.GcpExtension], error) {

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
					Field: "zone",
					Op:    filter.Equal.Factory(),
					Value: option.Zone,
				},
			},
		},
		Page: core.NewDefaultBasePage(),
	}
	result, err := cli.dbCli.Gcp.ListDisk(kt.Ctx, kt.Header(), req)
	if err != nil {
		logs.Errorf("[%s] list disk from db failed, err: %v, account: %s, req: %v, rid: %s", enumor.Gcp,
			err, params.AccountID, req, kt.Rid)
		return nil, err
	}

	return result.Details, nil
}

func (cli *client) updateDisk(kt *kit.Kit, accountID string, updateMap map[string]adaptordisk.GcpDisk) error {

	if len(updateMap) <= 0 {
		return fmt.Errorf("updateMap is <= 0, not update")
	}

	disks := make([]*disk.DiskExtUpdateReq[coredisk.GcpExtension], 0)

	for id, one := range updateMap {
		disk := &disk.DiskExtUpdateReq[coredisk.GcpExtension]{
			ID:           id,
			Region:       one.Region,
			Status:       one.Status,
			IsSystemDisk: one.Boot,
			Memo:         converter.ValToPtr(one.Description),
			Extension: &coredisk.GcpExtension{
				SelfLink:    one.SelfLink,
				SourceImage: one.SourceImage,
				Description: one.Description,
				// TODO: not find
				Encrypted: nil,
			},
		}

		disks = append(disks, disk)
	}

	var updateReq disk.DiskExtBatchUpdateReq[coredisk.GcpExtension]
	for _, disk := range disks {
		updateReq = append(updateReq, disk)
	}
	if _, err := cli.dbCli.Gcp.BatchUpdateDisk(kt.Ctx, kt.Header(), &updateReq); err != nil {
		logs.Errorf("[%s] request dataservice gcp BatchUpdateDisk failed, err: %v, rid: %s", enumor.Gcp,
			err, kt.Rid)
		return err
	}

	logs.Infof("[%s] sync disk to update disk success, accountID: %s, count: %d, rid: %s", enumor.Gcp,
		accountID, len(updateMap), kt.Rid)

	return nil
}

func (cli *client) createDisk(kt *kit.Kit, accountID string, zone string, addSlice []adaptordisk.GcpDisk) error {

	if len(addSlice) <= 0 {
		return fmt.Errorf("addSlice is <= 0, not create")
	}

	var createReq disk.DiskExtBatchCreateReq[coredisk.GcpExtension]

	for _, one := range addSlice {
		disk := &disk.DiskExtCreateReq[coredisk.GcpExtension]{
			AccountID:    accountID,
			Name:         one.Name,
			CloudID:      fmt.Sprint(one.Id),
			Region:       one.Region,
			Zone:         zone,
			DiskSize:     uint64(one.SizeGb),
			DiskType:     one.Type,
			Status:       one.Status,
			IsSystemDisk: converter.PtrToVal(one.Boot),
			Memo:         converter.ValToPtr(one.Description),
			Extension: &coredisk.GcpExtension{
				SelfLink:    one.SelfLink,
				SourceImage: one.SourceImage,
				Description: one.Description,
				// TODO: not find
				Encrypted: nil,
			},
		}

		createReq = append(createReq, disk)
	}

	_, err := cli.dbCli.Gcp.BatchCreateDisk(kt.Ctx, kt.Header(), &createReq)
	if err != nil {
		logs.Errorf("[%s] request dataservice to create gcp disk failed, err: %v, rid: %s", enumor.Gcp,
			err, kt.Rid)
		return err
	}

	logs.Infof("[%s] sync disk to create disk success, accountID: %s, count: %d, rid: %s", enumor.Gcp,
		accountID, len(addSlice), kt.Rid)

	return nil
}

func (cli *client) deleteDisk(kt *kit.Kit, accountID string, zone string, delCloudIDs []string) error {
	if len(delCloudIDs) <= 0 {
		return fmt.Errorf("delCloudIDs is <= 0, not delete")
	}

	checkParams := &SyncBaseParams{
		AccountID: accountID,
		CloudIDs:  delCloudIDs,
	}
	delDiskFromCloud, err := cli.listDiskFromCloud(kt, checkParams, &SyncDiskOption{Zone: zone})
	if err != nil {
		return err
	}

	if len(delDiskFromCloud) > 0 {
		logs.Errorf("[%s] validate disk not exist failed, before delete, opt: %v, failed_count: %d, rid: %s",
			enumor.Gcp, checkParams, len(delDiskFromCloud), kt.Rid)
		return fmt.Errorf("validate disk not exist failed, before delete")
	}

	deleteReq := &disk.DiskDeleteReq{
		Filter: tools.ContainersExpression("cloud_id", delCloudIDs),
	}
	if _, err = cli.dbCli.Global.DeleteDisk(kt.Ctx, kt.Header(), deleteReq); err != nil {
		logs.Errorf("[%s] request dataservice to batch delete disk failed, err: %v, rid: %s", enumor.Gcp,
			err, kt.Rid)
		return err
	}

	logs.Infof("[%s] sync disk to delete disk success, accountID: %s, count: %d, rid: %s", enumor.Gcp,
		accountID, len(delCloudIDs), kt.Rid)

	return nil
}

// RemoveDiskDeleteFromCloud ...
func (cli *client) RemoveDiskDeleteFromCloud(kt *kit.Kit, accountID string, zone string) error {
	req := &core.ListReq{
		Fields: []string{"id", "cloud_id"},
		Filter: &filter.Expression{
			Op: filter.And,
			Rules: []filter.RuleFactory{
				&filter.AtomRule{Field: "account_id", Op: filter.Equal.Factory(), Value: accountID},
				&filter.AtomRule{Field: "zone", Op: filter.Equal.Factory(), Value: zone},
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
			logs.Errorf("[%s] request dataservice to list disk failed, err: %v, req: %v, rid: %s", enumor.Gcp,
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
			CloudIDs:  cloudIDs,
		}
		resultFromCloud, err := cli.listDiskFromCloud(kt, params, &SyncDiskOption{Zone: zone})
		if err != nil {
			return err
		}

		// 如果有资源没有查询出来，说明数据被从云上删除
		if len(resultFromCloud) != len(cloudIDs) {
			cloudIDMap := converter.StringSliceToMap(cloudIDs)
			for _, one := range resultFromCloud {
				delete(cloudIDMap, fmt.Sprint(one.Id))
			}

			cloudIDs := converter.MapKeyToStringSlice(cloudIDMap)
			if len(cloudIDs) > 0 {
				if err := cli.deleteDisk(kt, accountID, zone, cloudIDs); err != nil {
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

func isDiskChange(cloud adaptordisk.GcpDisk, db *coredisk.Disk[coredisk.GcpExtension]) bool {

	if cloud.Status != db.Status {
		return true
	}

	if cloud.Region != db.Region {
		return true
	}

	if cloud.Description != converter.PtrToVal(db.Memo) {
		return true
	}

	if cloud.SelfLink != db.Extension.SelfLink {
		return true
	}

	if cloud.SourceImage != db.Extension.SourceImage {
		return true
	}

	if cloud.Description != db.Extension.Description {
		return true
	}

	if cloud.Boot != nil && *cloud.Boot != db.IsSystemDisk {
		return true
	}

	return false
}
