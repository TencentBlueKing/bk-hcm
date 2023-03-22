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
	"hcm/pkg/adaptor/types/disk"
	apicore "hcm/pkg/api/core"
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

// SyncAwsDiskOption define sync aws disk option.
type SyncAwsDiskOption struct {
	AccountID string   `json:"account_id" validate:"required"`
	Region    string   `json:"region" validate:"required"`
	CloudIDs  []string `json:"cloud_ids" validate:"omitempty"`
}

// Validate SyncAwsDiskOption
func (opt SyncAwsDiskOption) Validate() error {
	if err := validator.Validate.Struct(opt); err != nil {
		return err
	}

	if len(opt.CloudIDs) > constant.RelResourceOperationMaxLimit {
		return fmt.Errorf("cloudIDs should <= %d", constant.RelResourceOperationMaxLimit)
	}

	return nil
}

// SyncAwsDisk sync disk self
func SyncAwsDisk(kt *kit.Kit, req *SyncAwsDiskOption,
	ad *cloudclient.CloudAdaptorClient, dataCli *dataservice.Client) (interface{}, error) {

	if err := req.Validate(); err != nil {
		return nil, errf.NewFromErr(errf.InvalidParameter, err)
	}

	cloudMap, err := getDatasFromAwsForDiskSync(kt, req, ad)
	if err != nil {
		logs.Errorf("request getDatasFromAwsForDiskSync failed, err: %v, rid: %s", err, kt.Rid)
		return nil, err
	}

	dsMap, err := getDatasFromDSForAwsDiskSync(kt, req, dataCli)
	if err != nil {
		logs.Errorf("request getDatasFromDSForDiskSync failed, err: %v, rid: %s", err, kt.Rid)
		return nil, err
	}

	err = diffAwsDiskSync(kt, cloudMap, dsMap, req, dataCli)
	if err != nil {
		logs.Errorf("request diffAwsDiskSync failed, err: %v, rid: %s", err, kt.Rid)
		return nil, err
	}

	return nil, nil
}

func getDatasFromDSForAwsDiskSync(kt *kit.Kit, req *SyncAwsDiskOption,
	dataCli *dataservice.Client) (map[string]*AwsDiskSyncDS, error) {
	start := 0
	resultsHcm := make([]*datadisk.DiskExtResult[dataproto.AwsDiskExtensionResult], 0)
	for {
		dataReq := &datadisk.DiskListReq{
			Filter: &filter.Expression{
				Op: filter.And,
				Rules: []filter.RuleFactory{
					filter.AtomRule{Field: "account_id", Op: filter.Equal.Factory(), Value: req.AccountID},
					filter.AtomRule{Field: "region", Op: filter.Equal.Factory(), Value: req.Region},
				},
			},
			Page: &apicore.BasePage{Start: uint32(start), Limit: filter.DefaultMaxInLimit},
		}

		if len(req.CloudIDs) > 0 {
			filter := filter.AtomRule{Field: "cloud_id", Op: filter.In.Factory(), Value: req.CloudIDs}
			dataReq.Filter.Rules = append(dataReq.Filter.Rules, filter)
		}

		results, err := dataCli.Aws.ListDisk(kt.Ctx, kt.Header(), dataReq)
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

	dsMap := make(map[string]*AwsDiskSyncDS)
	for _, result := range resultsHcm {
		disk := new(AwsDiskSyncDS)
		disk.IsUpdated = false
		disk.HcDisk = result
		dsMap[result.CloudID] = disk
	}

	return dsMap, nil
}

func getDatasFromAwsForDiskSync(kt *kit.Kit, req *SyncAwsDiskOption,
	ad *cloudclient.CloudAdaptorClient) (map[string]*AwsDiskSyncDiff, error) {

	client, err := ad.Aws(kt, req.AccountID)
	if err != nil {
		return nil, err
	}

	listOpt := &disk.AwsDiskListOption{
		Region: req.Region,
	}
	if len(req.CloudIDs) > 0 {
		listOpt.CloudIDs = req.CloudIDs
	}

	datas, _, err := client.ListDisk(kt, listOpt)
	if err != nil {
		logs.Errorf("request adaptor to list aws disk failed, err: %v, rid: %s", err, kt.Rid)
		return nil, err
	}

	cloudMap := make(map[string]*AwsDiskSyncDiff)
	for _, data := range datas {
		disk := new(AwsDiskSyncDiff)
		disk.Disk = data
		cloudMap[*data.VolumeId] = disk
	}

	return cloudMap, nil
}

func diffAwsDiskSync(kt *kit.Kit, cloudMap map[string]*AwsDiskSyncDiff, dsMap map[string]*AwsDiskSyncDS,
	req *SyncAwsDiskOption, dataCli *dataservice.Client) error {

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
			logs.Errorf("request diffDiskSyncDelete failed, err: %v, rid: %s", err, kt.Rid)
			return err
		}
	}

	if len(updateCloudIDs) > 0 {
		err := diffAwsSyncUpdate(kt, cloudMap, dsMap, updateCloudIDs, dataCli)
		if err != nil {
			logs.Errorf("request diffAwsSyncUpdate failed, err: %v, rid: %s", err, kt.Rid)
			return err
		}
	}

	if len(addCloudIDs) > 0 {
		_, err := diffAwsDiskSyncAdd(kt, cloudMap, req, addCloudIDs, dataCli)
		if err != nil {
			logs.Errorf("request diffAwsDiskSyncAdd failed, err: %v, rid: %s", err, kt.Rid)
			return err
		}
	}

	return nil
}

func diffAwsDiskSyncAdd(kt *kit.Kit, cloudMap map[string]*AwsDiskSyncDiff,
	req *SyncAwsDiskOption, addCloudIDs []string, dataCli *dataservice.Client) ([]string, error) {

	var createReq dataproto.DiskExtBatchCreateReq[dataproto.AwsDiskExtensionCreateReq]

	for _, id := range addCloudIDs {
		attachments := make([]*dataproto.AwsDiskAttachment, 0)
		if len(cloudMap[id].Disk.Attachments) > 0 {
			for _, v := range cloudMap[id].Disk.Attachments {
				if v != nil {
					tmp := &dataproto.AwsDiskAttachment{
						AttachTime:          v.AttachTime,
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
		for _, tag := range cloudMap[id].Disk.Tags {
			if tag != nil {
				if converter.PtrToVal(tag.Key) == "Name" {
					name = converter.PtrToVal(tag.Value)
				}
			}
		}

		disk := &dataproto.DiskExtCreateReq[dataproto.AwsDiskExtensionCreateReq]{
			AccountID: req.AccountID,
			Name:      name,
			CloudID:   id,
			Region:    req.Region,
			Zone:      converter.PtrToVal(cloudMap[id].Disk.AvailabilityZone),
			DiskSize:  uint64(converter.PtrToVal(cloudMap[id].Disk.Size)),
			DiskType:  converter.PtrToVal(cloudMap[id].Disk.VolumeType),
			Status:    converter.PtrToVal(cloudMap[id].Disk.State),
			// 该云没有此字段
			Memo: nil,
			Extension: &dataproto.AwsDiskExtensionCreateReq{
				Attachment: attachments,
				Encrypted:  cloudMap[id].Disk.Encrypted,
			},
		}
		createReq = append(createReq, disk)
	}

	if len(createReq) <= 0 {
		return make([]string, 0), nil
	}

	results, err := dataCli.Aws.BatchCreateDisk(kt.Ctx, kt.Header(), &createReq)
	if err != nil {
		logs.Errorf("request dataservice to create aws disk failed, err: %v, rid: %s", err, kt.Rid)
		return nil, err
	}

	return results.IDs, nil
}

func isAwsDiskChange(db *AwsDiskSyncDS, cloud *AwsDiskSyncDiff, cloudName string) bool {

	if converter.PtrToVal(cloud.Disk.State) != db.HcDisk.Status {
		return true
	}

	if cloudName != db.HcDisk.Name {
		return true
	}

	if !assert.IsPtrBoolEqual(cloud.Disk.Encrypted, db.HcDisk.Extension.Encrypted) {
		return true
	}

	for _, dbValue := range db.HcDisk.Extension.Attachment {
		isEqual := false
		for _, cloudValue := range cloud.Disk.Attachments {
			if dbValue.AttachTime.String() == cloudValue.AttachTime.String() &&
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

func diffAwsSyncUpdate(kt *kit.Kit, cloudMap map[string]*AwsDiskSyncDiff,
	dsMap map[string]*AwsDiskSyncDS, updateCloudIDs []string, dataCli *dataservice.Client) error {

	disks := make([]*dataproto.DiskExtUpdateReq[dataproto.AwsDiskExtensionUpdateReq], 0)

	for _, id := range updateCloudIDs {
		name := ""
		for _, tag := range cloudMap[id].Disk.Tags {
			if tag != nil {
				if converter.PtrToVal(tag.Key) == "Name" {
					name = converter.PtrToVal(tag.Value)
				}
			}
		}

		if !isAwsDiskChange(dsMap[id], cloudMap[id], name) {
			continue
		}

		attachments := make([]*dataproto.AwsDiskAttachment, 0)
		if len(cloudMap[id].Disk.Attachments) > 0 {
			for _, v := range cloudMap[id].Disk.Attachments {
				if v != nil {
					tmp := &dataproto.AwsDiskAttachment{
						AttachTime:          v.AttachTime,
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

		disk := &dataproto.DiskExtUpdateReq[dataproto.AwsDiskExtensionUpdateReq]{
			ID:     dsMap[id].HcDisk.ID,
			Status: *cloudMap[id].Disk.State,
			Name:   name,
			Extension: &dataproto.AwsDiskExtensionUpdateReq{
				Attachment: attachments,
				Encrypted:  cloudMap[id].Disk.Encrypted,
			},
		}

		disks = append(disks, disk)
	}

	if len(disks) > 0 {
		elems := slice.Split(disks, typescore.TCloudQueryLimit)
		for _, partDisks := range elems {
			var updateReq dataproto.DiskExtBatchUpdateReq[dataproto.AwsDiskExtensionUpdateReq]
			for _, disk := range partDisks {
				updateReq = append(updateReq, disk)
			}
			if _, err := dataCli.Aws.BatchUpdateDisk(kt.Ctx, kt.Header(), &updateReq); err != nil {
				logs.Errorf("request dataservice aws BatchUpdateDisk failed, err: %v, rid: %s", err, kt.Rid)
				return err
			}
		}
	}

	return nil
}
