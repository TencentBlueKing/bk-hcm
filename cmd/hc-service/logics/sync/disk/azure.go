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

package disk

import (
	"fmt"

	cloudclient "hcm/cmd/hc-service/service/cloud-adaptor"
	typescore "hcm/pkg/adaptor/types/core"
	"hcm/pkg/api/core"
	datadisk "hcm/pkg/api/data-service/cloud/disk"
	dataproto "hcm/pkg/api/data-service/cloud/disk"
	dataservice "hcm/pkg/client/data-service"
	"hcm/pkg/criteria/constant"
	"hcm/pkg/criteria/errf"
	"hcm/pkg/criteria/validator"
	"hcm/pkg/kit"
	"hcm/pkg/logs"
	"hcm/pkg/runtime/filter"
	"hcm/pkg/tools/assert"
	"hcm/pkg/tools/converter"
	"hcm/pkg/tools/slice"
)

// SyncAzureDiskOption define sync azure disk option.
type SyncAzureDiskOption struct {
	AccountID         string   `json:"account_id" validate:"required"`
	ResourceGroupName string   `json:"resource_group_name" validate:"required"`
	CloudIDs          []string `json:"cloud_ids" validate:"omitempty"`
}

// Validate SyncAzureDiskOption
func (opt SyncAzureDiskOption) Validate() error {
	if err := validator.Validate.Struct(opt); err != nil {
		return err
	}

	if len(opt.CloudIDs) > constant.RelResourceOperationMaxLimit {
		return fmt.Errorf("cloudIDs should <= %d", constant.RelResourceOperationMaxLimit)
	}

	return nil
}

// SyncAzureDisk sync disk self
func SyncAzureDisk(kt *kit.Kit, req *SyncAzureDiskOption, ad *cloudclient.CloudAdaptorClient,
	dataCli *dataservice.Client) (interface{}, error) {

	return SyncAzureDiskWithOs(kt, req, nil, ad, dataCli)
}

// SyncAzureDiskWithOs sync disk with cvm os device info
func SyncAzureDiskWithOs(kt *kit.Kit, req *SyncAzureDiskOption, osMap map[string]struct{},
	ad *cloudclient.CloudAdaptorClient, dataCli *dataservice.Client) (interface{}, error) {

	if err := req.Validate(); err != nil {
		return nil, errf.NewFromErr(errf.InvalidParameter, err)
	}

	cloudMap, err := getDatasFromAzureForDiskSync(kt, req, ad)
	if err != nil {
		logs.Errorf("request getDatasFromAzureForDiskSync failed, err: %v, rid: %s", err, kt.Rid)
		return nil, err
	}

	dsMap, err := getDatasFromAzureDSForDiskSync(kt, req, dataCli)
	if err != nil {
		logs.Errorf("request getDatasFromAzureDSForDiskSync failed, err: %v, rid: %s", err, kt.Rid)
		return nil, err
	}

	err = diffAzureDiskSync(kt, cloudMap, dsMap, req, osMap, dataCli)
	if err != nil {
		logs.Errorf("request diffAzureDiskSync failed, err: %v, rid: %s", err, kt.Rid)
		return nil, err
	}

	return nil, nil
}

func getDatasFromAzureDSForDiskSync(kt *kit.Kit, req *SyncAzureDiskOption,
	dataCli *dataservice.Client) (map[string]*AzureDiskSyncDS, error) {

	start := 0
	resultsHcm := make([]*datadisk.DiskExtResult[dataproto.AzureDiskExtensionResult], 0)
	for {
		dataReq := &datadisk.DiskListReq{
			Filter: &filter.Expression{
				Op: filter.And,
				Rules: []filter.RuleFactory{
					&filter.AtomRule{
						Field: "account_id",
						Op:    filter.Equal.Factory(),
						Value: req.AccountID,
					},
					&filter.AtomRule{
						Field: "extension.resource_group_name",
						Op:    filter.JSONEqual.Factory(),
						Value: req.ResourceGroupName,
					},
				},
			},
			Page: &core.BasePage{Start: uint32(start), Limit: filter.DefaultMaxInLimit},
		}

		if len(req.CloudIDs) > 0 {
			filter := filter.AtomRule{Field: "cloud_id", Op: filter.In.Factory(), Value: req.CloudIDs}
			dataReq.Filter.Rules = append(dataReq.Filter.Rules, filter)
		}

		results, err := dataCli.Azure.ListDisk(kt.Ctx, kt.Header(), dataReq)
		if err != nil {
			logs.Errorf("from data-service list disk failed, err: %v, rid: %s", err, kt.Rid)
		}

		if results == nil {
			break
		}

		resultsHcm = append(resultsHcm, results.Details...)
		start += len(results.Details)
		if uint(len(results.Details)) < dataReq.Page.Limit {
			break
		}
	}

	dsMap := make(map[string]*AzureDiskSyncDS)
	for _, result := range resultsHcm {
		disk := new(AzureDiskSyncDS)
		disk.IsUpdated = false
		disk.HcDisk = result
		dsMap[result.CloudID] = disk
	}

	return dsMap, nil
}

func getDatasFromAzureForDiskSync(kt *kit.Kit, req *SyncAzureDiskOption,
	ad *cloudclient.CloudAdaptorClient) (map[string]*AzureDiskSyncDiff, error) {

	client, err := ad.Azure(kt, req.AccountID)
	if err != nil {
		return nil, err
	}

	listOpt := &typescore.AzureListByIDOption{
		ResourceGroupName: req.ResourceGroupName,
	}
	if len(req.CloudIDs) > 0 {
		listOpt.CloudIDs = req.CloudIDs
	}

	datas, err := client.ListDiskByID(kt, listOpt)
	if err != nil {
		logs.Errorf("request adaptor to list azure disk failed, err: %v, rid: %s", err, kt.Rid)
		return nil, err
	}

	cloudMap := make(map[string]*AzureDiskSyncDiff)
	for _, data := range datas {
		disk := new(AzureDiskSyncDiff)
		disk.Disk = data
		cloudMap[*data.ID] = disk
	}

	return cloudMap, nil
}

func diffAzureDiskSync(kt *kit.Kit, cloudMap map[string]*AzureDiskSyncDiff, dsMap map[string]*AzureDiskSyncDS,
	req *SyncAzureDiskOption, osMap map[string]struct{}, dataCli *dataservice.Client) error {

	addCloudIDs := make([]string, 0)
	for id := range cloudMap {
		if _, ok := dsMap[id]; !ok {
			addCloudIDs = append(addCloudIDs, id)
		} else {
			dsMap[id].IsUpdated = true
		}
	}

	deleteCloudIDs := make([]string, 0)
	updateCloudIDs := make([]string, 0)
	for id, one := range dsMap {
		if !one.IsUpdated {
			deleteCloudIDs = append(deleteCloudIDs, id)
		} else {
			updateCloudIDs = append(updateCloudIDs, id)
		}
	}

	if len(deleteCloudIDs) > 0 {
		err := DiffDiskSyncDelete(kt, deleteCloudIDs, dataCli)
		if err != nil {
			logs.Errorf("request diffDiskSyncDeletek failed, err: %v, rid: %s", err, kt.Rid)
			return err
		}
	}

	if len(updateCloudIDs) > 0 {
		err := diffAzureSyncUpdate(kt, cloudMap, dsMap, updateCloudIDs, dataCli, req, osMap)
		if err != nil {
			logs.Errorf("request diffAzureSyncUpdate failed, err: %v, rid: %s", err, kt.Rid)
			return err
		}
	}

	if len(addCloudIDs) > 0 {
		_, err := diffAzureDiskSyncAdd(kt, cloudMap, req, addCloudIDs, dataCli, osMap)
		if err != nil {
			logs.Errorf("request diffAzureDiskSyncAdd failed, err: %v, rid: %s", err, kt.Rid)
			return err
		}
	}

	return nil
}

func diffAzureDiskSyncAdd(kt *kit.Kit, cloudMap map[string]*AzureDiskSyncDiff, req *SyncAzureDiskOption,
	addCloudIDs []string, dataCli *dataservice.Client, osMap map[string]struct{}) ([]string, error) {

	var createReq dataproto.DiskExtBatchCreateReq[dataproto.AzureDiskExtensionCreateReq]

	for _, id := range addCloudIDs {
		disk := &dataproto.DiskExtCreateReq[dataproto.AzureDiskExtensionCreateReq]{
			AccountID: req.AccountID,
			Name:      converter.PtrToVal(cloudMap[id].Disk.Name),
			CloudID:   id,
			Region:    converter.PtrToVal(cloudMap[id].Disk.Location),
			DiskSize:  uint64(*cloudMap[id].Disk.DiskSize) / 1024 / 1024 / 1024,
			DiskType:  converter.PtrToVal(cloudMap[id].Disk.Type),
			Status:    string(*cloudMap[id].Disk.Status),
			Zone:      "",
			// 该云没有此字段
			Memo: nil,
			Extension: &dataproto.AzureDiskExtensionCreateReq{
				ResourceGroupName: req.ResourceGroupName,
				OSType:            converter.PtrToVal(cloudMap[id].Disk.OSType),
				SKUName:           cloudMap[id].Disk.SKUName,
				SKUTier:           cloudMap[id].Disk.SKUTier,
			},
		}
		if len(cloudMap[id].Disk.Zones) > 0 {
			disk.Zone = converter.PtrToVal(cloudMap[id].Disk.Zones[0])
		}

		if _, exists := osMap[id]; exists {
			disk.IsSystemDisk = true
		}

		createReq = append(createReq, disk)
	}

	if len(createReq) <= 0 {
		return make([]string, 0), nil
	}

	results, err := dataCli.Azure.BatchCreateDisk(kt.Ctx, kt.Header(), &createReq)
	if err != nil {
		logs.Errorf("request dataservice to create azure disk failed, err: %v, rid: %s", err, kt.Rid)
		return nil, err
	}

	return results.IDs, nil
}

func isAzureDiskChange(db *AzureDiskSyncDS, cloud *AzureDiskSyncDiff, isSystemDisk bool) bool {

	if converter.PtrToVal(cloud.Disk.Status) != db.HcDisk.Status {
		return true
	}

	if !assert.IsPtrStringEqual(cloud.Disk.SKUName, db.HcDisk.Extension.SKUName) {
		return true
	}

	if !assert.IsPtrStringEqual(cloud.Disk.SKUTier, db.HcDisk.Extension.SKUTier) {
		return true
	}

	if converter.PtrToVal(cloud.Disk.OSType) != db.HcDisk.Extension.OSType {
		return true
	}

	if isSystemDisk != db.HcDisk.IsSystemDisk {
		return true
	}

	return false
}

func diffAzureSyncUpdate(kt *kit.Kit, cloudMap map[string]*AzureDiskSyncDiff, dsMap map[string]*AzureDiskSyncDS,
	updateCloudIDs []string, dataCli *dataservice.Client, req *SyncAzureDiskOption, osMap map[string]struct{}) error {

	disks := make([]*dataproto.DiskExtUpdateReq[dataproto.AzureDiskExtensionUpdateReq], 0)

	for _, id := range updateCloudIDs {
		_, isSystemDisk := osMap[id]

		if !isAzureDiskChange(dsMap[id], cloudMap[id], isSystemDisk) {
			continue
		}

		disk := &dataproto.DiskExtUpdateReq[dataproto.AzureDiskExtensionUpdateReq]{
			ID:           dsMap[id].HcDisk.ID,
			Status:       *cloudMap[id].Disk.Status,
			IsSystemDisk: &isSystemDisk,
			Extension: &dataproto.AzureDiskExtensionUpdateReq{
				ResourceGroupName: req.ResourceGroupName,
				OSType:            converter.PtrToVal(cloudMap[id].Disk.OSType),
				SKUName:           cloudMap[id].Disk.SKUName,
				SKUTier:           cloudMap[id].Disk.SKUTier,
			},
		}
		disks = append(disks, disk)
	}

	if len(disks) > 0 {
		elems := slice.Split(disks, typescore.TCloudQueryLimit)
		for _, partDisks := range elems {
			var updateReq dataproto.DiskExtBatchUpdateReq[dataproto.AzureDiskExtensionUpdateReq]
			for _, disk := range partDisks {
				updateReq = append(updateReq, disk)
			}
			if _, err := dataCli.Azure.BatchUpdateDisk(kt.Ctx, kt.Header(), &updateReq); err != nil {
				logs.Errorf("request dataservice azure BatchUpdateDisk failed, err: %v, rid: %s", err, kt.Rid)
				return err
			}
		}
	}

	return nil
}
