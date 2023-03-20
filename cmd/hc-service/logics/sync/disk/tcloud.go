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
	"hcm/pkg/adaptor/tcloud"
	"hcm/pkg/adaptor/types/core"
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

	cbs "github.com/tencentcloud/tencentcloud-sdk-go/tencentcloud/cbs/v20170312"
)

// SyncTCloudDiskOption define sync tcloud disk option.
type SyncTCloudDiskOption struct {
	AccountID string   `json:"account_id" validate:"required"`
	Region    string   `json:"region" validate:"required"`
	CloudIDs  []string `json:"cloud_ids" validate:"omitempty"`
}

// Validate SyncTCloudDiskOption
func (opt SyncTCloudDiskOption) Validate() error {
	if err := validator.Validate.Struct(opt); err != nil {
		return err
	}

	if len(opt.CloudIDs) > constant.RelResourceOperationMaxLimit {
		return fmt.Errorf("cloudIDs should <= %d", constant.RelResourceOperationMaxLimit)
	}

	return nil
}

// SyncTCloudDisk sync disk self
func SyncTCloudDisk(kt *kit.Kit, req *SyncTCloudDiskOption,
	ad *cloudclient.CloudAdaptorClient, dataCli *dataservice.Client) (interface{}, error) {

	if err := req.Validate(); err != nil {
		return nil, errf.NewFromErr(errf.InvalidParameter, err)
	}

	cloudMap, err := getDatasFromTCloudForDiskSync(kt, req, ad)
	if err != nil {
		logs.Errorf("request getDatasFromTCloudForDiskSync failed, err: %v, rid: %s", err, kt.Rid)
		return nil, err
	}

	dsMap, err := getDatasFromDSForTCloudDiskSync(kt, req, dataCli)
	if err != nil {
		logs.Errorf("request getDatasFromDSForDiskSync failed, err: %v, rid: %s", err, kt.Rid)
		return nil, err
	}

	err = diffTCloudDiskSync(kt, cloudMap, dsMap, req, dataCli)
	if err != nil {
		logs.Errorf("request diffTCloudDiskSync failed, err: %v, rid: %s", err, kt.Rid)
		return nil, err
	}

	return nil, nil
}

func getDatasFromDSForTCloudDiskSync(kt *kit.Kit, req *SyncTCloudDiskOption,
	dataCli *dataservice.Client) (map[string]*TCloudDiskSyncDS, error) {
	start := 0
	resultsHcm := make([]*datadisk.DiskExtResult[dataproto.TCloudDiskExtensionResult], 0)
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

		results, err := dataCli.TCloud.ListDisk(kt.Ctx, kt.Header(), dataReq)
		if err != nil {
			logs.Errorf("from data-service list disk failed, err: %v, rid: %s", err, kt.Rid)
		}

		resultsHcm = append(resultsHcm, results.Details...)
		start += len(results.Details)
		if uint(len(results.Details)) < dataReq.Page.Limit {
			break
		}
	}

	dsMap := make(map[string]*TCloudDiskSyncDS)
	for _, result := range resultsHcm {
		sg := new(TCloudDiskSyncDS)
		sg.IsUpdated = false
		sg.HcDisk = result
		dsMap[result.CloudID] = sg
	}

	return dsMap, nil
}

func getDatasFromTCloudForDiskSync(kt *kit.Kit, req *SyncTCloudDiskOption,
	ad *cloudclient.CloudAdaptorClient) (map[string]*TCloudDiskSyncDiff, error) {

	client, err := ad.TCloud(kt, req.AccountID)
	if err != nil {
		return nil, err
	}

	cloudMap := make(map[string]*TCloudDiskSyncDiff)
	if len(req.CloudIDs) > 0 {
		cloudMap, err = getTCloudDiskByCloudIDsSync(kt, client, req)
		if err != nil {
			logs.Errorf("request to list tcloud disk by cloud_ids failed, err: %v, rid: %s", err, kt.Rid)
			return nil, err
		}
	} else {
		cloudMap, err = getTCloudDiskAllSync(kt, client, req)
		if err != nil {
			logs.Errorf("request to list all tcloud disk failed, err: %v, rid: %s", err, kt.Rid)
			return nil, err
		}
	}

	return cloudMap, nil
}

func getTCloudDiskAllSync(kt *kit.Kit, client *tcloud.TCloud,
	req *SyncTCloudDiskOption) (map[string]*TCloudDiskSyncDiff, error) {

	offset := 0
	datasCloud := make([]*cbs.Disk, 0)

	for {
		opt := &disk.TCloudDiskListOption{
			Region: req.Region,
			Page:   &core.TCloudPage{Offset: uint64(offset), Limit: uint64(core.TCloudQueryLimit)},
		}

		datas, err := client.ListDisk(kt, opt)
		if err != nil {
			logs.Errorf("request adaptor to list tcloud disk failed, err: %v, opt: %v, rid: %s", err, opt, kt.Rid)
			return nil, err
		}

		offset += len(datas)
		datasCloud = append(datasCloud, datas...)
		if uint(len(datas)) < core.TCloudQueryLimit {
			break
		}
	}

	cloudMap := make(map[string]*TCloudDiskSyncDiff)
	for _, data := range datasCloud {
		sg := new(TCloudDiskSyncDiff)
		sg.Disk = data
		cloudMap[*data.DiskId] = sg
	}

	return cloudMap, nil
}

func getTCloudDiskByCloudIDsSync(kt *kit.Kit, client *tcloud.TCloud,
	req *SyncTCloudDiskOption) (map[string]*TCloudDiskSyncDiff, error) {

	opt := &disk.TCloudDiskListOption{
		Region:   req.Region,
		CloudIDs: req.CloudIDs,
	}

	datas, err := client.ListDisk(kt, opt)
	if err != nil {
		logs.Errorf("request adaptor to list tcloud disk failed, err: %v, opt: %v, rid: %s", err, opt, kt.Rid)
		return nil, err
	}

	cloudMap := make(map[string]*TCloudDiskSyncDiff)
	for _, data := range datas {
		sg := new(TCloudDiskSyncDiff)
		sg.Disk = data
		cloudMap[*data.DiskId] = sg
	}

	return cloudMap, nil
}

func diffTCloudDiskSync(kt *kit.Kit, cloudMap map[string]*TCloudDiskSyncDiff,
	dsMap map[string]*TCloudDiskSyncDS, req *SyncTCloudDiskOption, dataCli *dataservice.Client) error {

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
		err := diffTCloudDiskSyncUpdate(kt, cloudMap, dsMap, updateCloudIDs, dataCli)
		if err != nil {
			logs.Errorf("request diffTCloudDiskSyncUpdate failed, err: %v, rid: %s", err, kt.Rid)
			return err
		}
	}

	if len(addCloudIDs) > 0 {
		_, err := diffTCloudDiskSyncAdd(kt, cloudMap, req, addCloudIDs, dataCli)
		if err != nil {
			logs.Errorf("request diffTCloudDiskSyncAdd failed, err: %v, rid: %s", err, kt.Rid)
			return err
		}
	}

	return nil
}

func diffTCloudDiskSyncAdd(kt *kit.Kit, cloudMap map[string]*TCloudDiskSyncDiff,
	req *SyncTCloudDiskOption, addCloudIDs []string, dataCli *dataservice.Client) ([]string, error) {
	var createReq dataproto.DiskExtBatchCreateReq[dataproto.TCloudDiskExtensionCreateReq]

	for _, id := range addCloudIDs {
		disk := &dataproto.DiskExtCreateReq[dataproto.TCloudDiskExtensionCreateReq]{
			AccountID: req.AccountID,
			Name:      converter.PtrToVal(cloudMap[id].Disk.DiskName),
			CloudID:   id,
			Region:    req.Region,
			Zone:      converter.PtrToVal(cloudMap[id].Disk.Placement.Zone),
			DiskSize:  converter.PtrToVal(cloudMap[id].Disk.DiskSize),
			DiskType:  converter.PtrToVal(cloudMap[id].Disk.DiskType),
			Status:    converter.PtrToVal(cloudMap[id].Disk.DiskState),
			// tcloud no memo
			Memo: nil,
			Extension: &dataproto.TCloudDiskExtensionCreateReq{
				DiskChargeType: converter.PtrToVal(cloudMap[id].Disk.DiskChargeType),
				DiskChargePrepaid: &dataproto.TCloudDiskChargePrepaid{
					RenewFlag: cloudMap[id].Disk.RenewFlag,
					Period:    cloudMap[id].Disk.DifferDaysOfDeadline,
				},
				Encrypted:    cloudMap[id].Disk.Encrypt,
				Attached:     cloudMap[id].Disk.Attached,
				DiskUsage:    cloudMap[id].Disk.DiskUsage,
				InstanceId:   cloudMap[id].Disk.InstanceId,
				InstanceType: cloudMap[id].Disk.InstanceType,
			},
		}
		createReq = append(createReq, disk)
	}

	if len(createReq) <= 0 {
		return make([]string, 0), nil
	}

	results, err := dataCli.TCloud.BatchCreateDisk(kt.Ctx, kt.Header(), &createReq)
	if err != nil {
		logs.Errorf("request dataservice to create tcloud disk failed, err: %v, rid: %s", err, kt.Rid)
		return nil, err
	}

	return results.IDs, nil
}

func isTCloudDiskChange(db *TCloudDiskSyncDS, cloud *TCloudDiskSyncDiff) bool {

	if converter.PtrToVal(cloud.Disk.DiskState) != db.HcDisk.Status {
		return true
	}

	if converter.PtrToVal(cloud.Disk.DiskChargeType) != db.HcDisk.Extension.DiskChargeType {
		return true
	}

	if !assert.IsPtrBoolEqual(cloud.Disk.Encrypt, db.HcDisk.Extension.Encrypted) {
		return true
	}

	if !assert.IsPtrBoolEqual(cloud.Disk.Attached, db.HcDisk.Extension.Attached) {
		return true
	}

	if !assert.IsPtrStringEqual(cloud.Disk.DiskUsage, db.HcDisk.Extension.DiskUsage) {
		return true
	}

	if !assert.IsPtrStringEqual(cloud.Disk.InstanceId, db.HcDisk.Extension.InstanceId) {
		return true
	}

	if !assert.IsPtrStringEqual(cloud.Disk.InstanceType, db.HcDisk.Extension.InstanceType) {
		return true
	}

	if !assert.IsPtrStringEqual(cloud.Disk.RenewFlag, db.HcDisk.Extension.DiskChargePrepaid.RenewFlag) {
		return true
	}

	if !assert.IsPtrInt64Equal(cloud.Disk.DifferDaysOfDeadline, db.HcDisk.Extension.DiskChargePrepaid.Period) {
		return true
	}

	return false
}

func diffTCloudDiskSyncUpdate(kt *kit.Kit, cloudMap map[string]*TCloudDiskSyncDiff,
	dsMap map[string]*TCloudDiskSyncDS, updateCloudIDs []string, dataCli *dataservice.Client) error {
	var updateReq dataproto.DiskExtBatchUpdateReq[dataproto.TCloudDiskExtensionUpdateReq]

	for _, id := range updateCloudIDs {

		if !isTCloudDiskChange(dsMap[id], cloudMap[id]) {
			continue
		}

		disk := &dataproto.DiskExtUpdateReq[dataproto.TCloudDiskExtensionUpdateReq]{
			ID:     dsMap[id].HcDisk.ID,
			Status: *cloudMap[id].Disk.DiskState,
			Extension: &dataproto.TCloudDiskExtensionUpdateReq{
				DiskChargeType: converter.PtrToVal(cloudMap[id].Disk.DiskChargeType),
				DiskChargePrepaid: &dataproto.TCloudDiskChargePrepaid{
					RenewFlag: cloudMap[id].Disk.RenewFlag,
					Period:    cloudMap[id].Disk.DifferDaysOfDeadline,
				},
				Encrypted:    cloudMap[id].Disk.Encrypt,
				Attached:     cloudMap[id].Disk.Attached,
				DiskUsage:    cloudMap[id].Disk.DiskUsage,
				InstanceId:   cloudMap[id].Disk.InstanceId,
				InstanceType: cloudMap[id].Disk.InstanceType,
			},
		}
		updateReq = append(updateReq, disk)
	}

	if len(updateReq) > 0 {
		if _, err := dataCli.TCloud.BatchUpdateDisk(kt.Ctx, kt.Header(), &updateReq); err != nil {
			logs.Errorf("request dataservice BatchUpdateDisk failed, err: %v, rid: %s", err, kt.Rid)
			return err
		}
	}

	return nil
}
