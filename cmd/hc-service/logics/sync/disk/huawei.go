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
	"hcm/pkg/adaptor/types/core"
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

// SyncHuaWeiDiskOption define sync huawei disk option.
type SyncHuaWeiDiskOption struct {
	AccountID string   `json:"account_id" validate:"required"`
	Region    string   `json:"region" validate:"required"`
	CloudIDs  []string `json:"cloud_ids" validate:"omitempty"`
}

// Validate SyncHuaWeiDiskOption
func (opt SyncHuaWeiDiskOption) Validate() error {
	if err := validator.Validate.Struct(opt); err != nil {
		return err
	}

	if len(opt.CloudIDs) > constant.RelResourceOperationMaxLimit {
		return fmt.Errorf("cloudIDs should <= %d", constant.RelResourceOperationMaxLimit)
	}

	return nil
}

// SyncHuaWeiDisk sync disk self
func SyncHuaWeiDisk(kt *kit.Kit, req *SyncHuaWeiDiskOption,
	ad *cloudclient.CloudAdaptorClient, dataCli *dataservice.Client) (interface{}, error) {

	if err := req.Validate(); err != nil {
		return nil, errf.NewFromErr(errf.InvalidParameter, err)
	}

	cloudMap, err := getDatasFromHuaWeiForDiskSync(kt, req, ad)
	if err != nil {
		logs.Errorf("request getDatasFromHuaWeiForDiskSync failed, err: %v, rid: %s", err, kt.Rid)
		return nil, err
	}

	dsMap, err := getDatasFromDSForHuaWeiDiskSync(kt, req, dataCli)
	if err != nil {
		logs.Errorf("request getDatasFromDSForDiskSync failed, err: %v, rid: %s", err, kt.Rid)
		return nil, err
	}

	err = diffHuaWeiDiskSync(kt, cloudMap, dsMap, req, dataCli)
	if err != nil {
		logs.Errorf("request diffHuaWeiDiskSync failed, err: %v, rid: %s", err, kt.Rid)
		return nil, err
	}

	return nil, nil
}

func getDatasFromDSForHuaWeiDiskSync(kt *kit.Kit, req *SyncHuaWeiDiskOption,
	dataCli *dataservice.Client) (map[string]*HuaWeiDiskSyncDS, error) {
	start := 0
	resultsHcm := make([]*datadisk.DiskExtResult[dataproto.HuaWeiDiskExtensionResult], 0)
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

		results, err := dataCli.HuaWei.ListDisk(kt.Ctx, kt.Header(), dataReq)
		if err != nil {
			logs.Errorf("from data-service list disk failed, err: %v, rid: %s", err, kt.Rid)
			return nil, err
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

	dsMap := make(map[string]*HuaWeiDiskSyncDS)
	for _, result := range resultsHcm {
		disk := new(HuaWeiDiskSyncDS)
		disk.IsUpdated = false
		disk.HcDisk = result
		dsMap[result.CloudID] = disk
	}

	return dsMap, nil
}

func getDatasFromHuaWeiForDiskSync(kt *kit.Kit, req *SyncHuaWeiDiskOption,
	ad *cloudclient.CloudAdaptorClient) (map[string]*HuaWeiDiskSyncDiff, error) {

	client, err := ad.HuaWei(kt, req.AccountID)
	if err != nil {
		return nil, err
	}

	var marker *string = nil
	limit := int32(core.HuaWeiQueryLimit)

	opt := &disk.HuaWeiDiskListOption{
		Region: req.Region,
		Page:   &core.HuaWeiPage{Limit: &limit, Marker: marker},
	}
	if len(req.CloudIDs) > 0 {
		opt.CloudIDs = req.CloudIDs
	}

	datas, err := client.ListDisk(kt, opt)
	if err != nil {
		logs.Errorf("request adaptor to list huawei disk failed, err: %v, rid: %s", err, kt.Rid)
		return nil, err
	}

	cloudMap := make(map[string]*HuaWeiDiskSyncDiff)
	for _, data := range datas {
		disk := new(HuaWeiDiskSyncDiff)
		disk.Disk = data
		cloudMap[data.Id] = disk
	}

	return cloudMap, nil
}

func diffHuaWeiDiskSync(kt *kit.Kit, cloudMap map[string]*HuaWeiDiskSyncDiff, dsMap map[string]*HuaWeiDiskSyncDS,
	req *SyncHuaWeiDiskOption, dataCli *dataservice.Client) error {

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
		err := diffHuaWeiSyncUpdate(kt, cloudMap, dsMap, updateCloudIDs, dataCli)
		if err != nil {
			logs.Errorf("request diffHuaWeiSyncUpdate failed, err: %v, rid: %s", err, kt.Rid)
			return err
		}
	}

	if len(addCloudIDs) > 0 {
		_, err := diffHuaWeiDiskSyncAdd(kt, cloudMap, req, addCloudIDs, dataCli)
		if err != nil {
			logs.Errorf("request diffHuaWeiDiskSyncAdd failed, err: %v, rid: %s", err, kt.Rid)
			return err
		}
	}

	return nil
}

func diffHuaWeiDiskSyncAdd(kt *kit.Kit, cloudMap map[string]*HuaWeiDiskSyncDiff, req *SyncHuaWeiDiskOption,
	addCloudIDs []string, dataCli *dataservice.Client) ([]string, error) {

	var createReq dataproto.DiskExtBatchCreateReq[dataproto.HuaWeiDiskExtensionCreateReq]

	for _, id := range addCloudIDs {

		attachments := make([]*dataproto.HuaWeiDiskAttachment, 0)
		if len(cloudMap[id].Disk.Attachments) > 0 {
			for _, v := range cloudMap[id].Disk.Attachments {
				tmp := &dataproto.HuaWeiDiskAttachment{
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

		disk := &dataproto.DiskExtCreateReq[dataproto.HuaWeiDiskExtensionCreateReq]{
			AccountID: req.AccountID,
			Name:      cloudMap[id].Disk.Name,
			CloudID:   id,
			Region:    req.Region,
			Zone:      cloudMap[id].Disk.AvailabilityZone,
			DiskSize:  uint64(cloudMap[id].Disk.Size),
			DiskType:  cloudMap[id].Disk.VolumeType,
			Status:    cloudMap[id].Disk.Status,
			Memo:      &cloudMap[id].Disk.Description,
			Extension: &dataproto.HuaWeiDiskExtensionCreateReq{
				ServiceType: cloudMap[id].Disk.ServiceType,
				Encrypted:   cloudMap[id].Disk.Encrypted,
				Attachment:  attachments,
				Bootable:    cloudMap[id].Disk.Bootable,
			},
		}
		createReq = append(createReq, disk)
	}

	if len(createReq) <= 0 {
		return make([]string, 0), nil
	}

	results, err := dataCli.HuaWei.BatchCreateDisk(kt.Ctx, kt.Header(), &createReq)
	if err != nil {
		logs.Errorf("request dataservice to create huawei disk failed, err: %v, rid: %s", err, kt.Rid)
		return nil, err
	}

	return results.IDs, nil
}

func isHuaWeiDiskChange(db *HuaWeiDiskSyncDS, cloud *HuaWeiDiskSyncDiff) bool {

	if cloud.Disk.Status != db.HcDisk.Status {
		return true
	}

	if cloud.Disk.Description != converter.PtrToVal(db.HcDisk.Memo) {
		return true
	}

	if cloud.Disk.ServiceType != db.HcDisk.Extension.ServiceType {
		return true
	}

	if !assert.IsPtrBoolEqual(cloud.Disk.Encrypted, db.HcDisk.Extension.Encrypted) {
		return true
	}

	if cloud.Disk.Bootable != db.HcDisk.Extension.Bootable {
		return true
	}

	for _, dbValue := range db.HcDisk.Extension.Attachment {
		isEqual := false
		for _, cloudValue := range cloud.Disk.Attachments {
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

	return false
}

func diffHuaWeiSyncUpdate(kt *kit.Kit, cloudMap map[string]*HuaWeiDiskSyncDiff, dsMap map[string]*HuaWeiDiskSyncDS,
	updateCloudIDs []string, dataCli *dataservice.Client) error {

	disks := make([]*dataproto.DiskExtUpdateReq[dataproto.HuaWeiDiskExtensionUpdateReq], 0)

	for _, id := range updateCloudIDs {

		if !isHuaWeiDiskChange(dsMap[id], cloudMap[id]) {
			continue
		}

		attachments := make([]*dataproto.HuaWeiDiskAttachment, 0)
		if len(cloudMap[id].Disk.Attachments) > 0 {
			for _, v := range cloudMap[id].Disk.Attachments {
				tmp := &dataproto.HuaWeiDiskAttachment{
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

		disk := &dataproto.DiskExtUpdateReq[dataproto.HuaWeiDiskExtensionUpdateReq]{
			ID:     dsMap[id].HcDisk.ID,
			Memo:   &cloudMap[id].Disk.Description,
			Status: cloudMap[id].Disk.Status,
			Extension: &dataproto.HuaWeiDiskExtensionUpdateReq{
				ServiceType: cloudMap[id].Disk.ServiceType,
				Encrypted:   cloudMap[id].Disk.Encrypted,
				Attachment:  attachments,
				Bootable:    cloudMap[id].Disk.Bootable,
			},
		}

		disks = append(disks, disk)
	}

	if len(disks) > 0 {
		elems := slice.Split(disks, typescore.TCloudQueryLimit)
		for _, partDisks := range elems {
			var updateReq dataproto.DiskExtBatchUpdateReq[dataproto.HuaWeiDiskExtensionUpdateReq]
			for _, disk := range partDisks {
				updateReq = append(updateReq, disk)
			}
			if _, err := dataCli.HuaWei.BatchUpdateDisk(kt.Ctx, kt.Header(), &updateReq); err != nil {
				logs.Errorf("request dataservice huawei BatchUpdateDisk failed, err: %v, rid: %s", err, kt.Rid)
				return err
			}
		}
	}

	return nil
}
