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
	"errors"
	"fmt"

	cloudclient "hcm/cmd/hc-service/service/cloud-adaptor"
	typecore "hcm/pkg/adaptor/types/core"
	typescore "hcm/pkg/adaptor/types/core"
	"hcm/pkg/adaptor/types/disk"
	"hcm/pkg/api/core"
	apicore "hcm/pkg/api/core"
	datadisk "hcm/pkg/api/data-service/cloud/disk"
	dataproto "hcm/pkg/api/data-service/cloud/disk"
	"hcm/pkg/api/data-service/cloud/zone"
	dataservice "hcm/pkg/client/data-service"
	"hcm/pkg/criteria/constant"
	"hcm/pkg/criteria/enumor"
	"hcm/pkg/criteria/errf"
	"hcm/pkg/criteria/validator"
	"hcm/pkg/kit"
	"hcm/pkg/logs"
	"hcm/pkg/runtime/filter"
	"hcm/pkg/tools/converter"
	"hcm/pkg/tools/slice"
)

// SyncGcpDiskOption define sync gcp disk option.
type SyncGcpDiskOption struct {
	AccountID string   `json:"account_id" validate:"required"`
	Zone      string   `json:"zone" validate:"required"`
	CloudIDs  []string `json:"cloud_ids" validate:"omitempty"`
	SelfLinks []string `json:"self_links" validate:"omitempty"`
}

// Validate SyncGcpDiskOption
func (opt SyncGcpDiskOption) Validate() error {
	if err := validator.Validate.Struct(opt); err != nil {
		return err
	}

	if len(opt.CloudIDs) > constant.RelResourceOperationMaxLimit {
		return fmt.Errorf("cloudIDs should <= %d", constant.RelResourceOperationMaxLimit)
	}

	return nil
}

// SyncGcpDisk sync disk self
func SyncGcpDisk(kt *kit.Kit, req *SyncGcpDiskOption, ad *cloudclient.CloudAdaptorClient,
	dataCli *dataservice.Client) (interface{}, error) {

	return SyncGcpDiskWithBoot(kt, req, nil, ad, dataCli)
}

// SyncGcpDiskWithBoot sync disk with cvm boot device info
func SyncGcpDiskWithBoot(kt *kit.Kit, req *SyncGcpDiskOption, bootMap map[string]struct{},
	ad *cloudclient.CloudAdaptorClient, dataCli *dataservice.Client) (interface{}, error) {

	if err := req.Validate(); err != nil {
		return nil, errf.NewFromErr(errf.InvalidParameter, err)
	}

	cloudMap, err := getDatasFromGcpForDiskSync(kt, req, ad)
	if err != nil {
		logs.Errorf("request getDatasFromGcpForDiskSync failed, err: %v, rid: %s", err, kt.Rid)
		return nil, err
	}

	dsMap, err := getDatasFromDSForGcpDiskSync(kt, req, dataCli)
	if err != nil {
		logs.Errorf("request getDatasFromDSForDiskSync failed, err: %v, rid: %s", err, kt.Rid)
		return nil, err
	}

	err = diffGcpDiskSync(kt, cloudMap, dsMap, req, bootMap, dataCli)
	if err != nil {
		logs.Errorf("request diffGcpDiskSync failed, err: %v, rid: %s", err, kt.Rid)
		return nil, err
	}

	return nil, nil
}

func getDatasFromDSForGcpDiskSync(kt *kit.Kit, req *SyncGcpDiskOption,
	dataCli *dataservice.Client) (map[string]*GcpDiskSyncDS, error) {
	start := 0
	resultsHcm := make([]*datadisk.DiskExtResult[dataproto.GcpDiskExtensionResult], 0)
	for {
		dataReq := &datadisk.DiskListReq{
			Filter: &filter.Expression{
				Op: filter.And,
				Rules: []filter.RuleFactory{
					filter.AtomRule{Field: "account_id", Op: filter.Equal.Factory(), Value: req.AccountID},
					filter.AtomRule{Field: "zone", Op: filter.Equal.Factory(), Value: req.Zone},
				},
			},
			Page: &apicore.BasePage{Start: uint32(start), Limit: filter.DefaultMaxInLimit},
		}

		if len(req.CloudIDs) > 0 {
			filter := filter.AtomRule{Field: "cloud_id", Op: filter.In.Factory(), Value: req.CloudIDs}
			dataReq.Filter.Rules = append(dataReq.Filter.Rules, filter)
		}

		if len(req.SelfLinks) > 0 {
			filter := filter.AtomRule{Field: "extension.self_link", Op: filter.JSONIn.Factory(),
				Value: req.SelfLinks}
			dataReq.Filter.Rules = append(dataReq.Filter.Rules, filter)
		}

		results, err := dataCli.Gcp.ListDisk(kt.Ctx, kt.Header(), dataReq)
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

	dsMap := make(map[string]*GcpDiskSyncDS)
	for _, result := range resultsHcm {
		disk := new(GcpDiskSyncDS)
		disk.IsUpdated = false
		disk.HcDisk = result
		dsMap[result.CloudID] = disk
	}

	return dsMap, nil
}

func getDatasFromGcpForDiskSync(kt *kit.Kit, req *SyncGcpDiskOption,
	ad *cloudclient.CloudAdaptorClient) (map[string]*GcpDiskSyncDiff, error) {

	client, err := ad.Gcp(kt, req.AccountID)
	if err != nil {
		return nil, err
	}

	cloudMap := make(map[string]*GcpDiskSyncDiff)
	nextToken := ""
	for {
		listOpt := &disk.GcpDiskListOption{
			Zone: req.Zone,
			Page: &typecore.GcpPage{
				PageToken: nextToken,
				PageSize:  int64(filter.DefaultMaxInLimit),
			},
		}

		if nextToken != "" {
			listOpt.Page.PageToken = nextToken
		}

		if len(req.CloudIDs) > 0 {
			listOpt.CloudIDs = req.CloudIDs
			listOpt.Page = nil
		}

		if len(req.SelfLinks) > 0 {
			listOpt.SelfLinks = req.SelfLinks
			listOpt.Page = nil
		}

		datas, token, err := client.ListDisk(kt, listOpt)
		if err != nil {
			logs.Errorf("request adaptor to list gcp disk failed, err: %v, rid: %s", err, kt.Rid)
			return nil, err
		}

		for _, data := range datas {
			disk := new(GcpDiskSyncDiff)
			disk.Disk = data
			cloudMap[fmt.Sprint(data.Id)] = disk
		}

		if len(token) == 0 {
			break
		}
		nextToken = token
	}

	return cloudMap, nil
}

func diffGcpDiskSync(kt *kit.Kit, cloudMap map[string]*GcpDiskSyncDiff, dsMap map[string]*GcpDiskSyncDS,
	req *SyncGcpDiskOption, bootMap map[string]struct{}, dataCli *dataservice.Client) error {

	if bootMap == nil {
		bootMap = make(map[string]struct{})
	}

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
		err := diffGcpSyncUpdate(kt, cloudMap, req, dsMap, updateCloudIDs, bootMap, dataCli)
		if err != nil {
			logs.Errorf("request diffGcpSyncUpdate failed, err: %v, rid: %s", err, kt.Rid)
			return err
		}
	}

	if len(addCloudIDs) > 0 {
		_, err := diffGcpDiskSyncAdd(kt, cloudMap, req, addCloudIDs, bootMap, dataCli)
		if err != nil {
			logs.Errorf("request diffGcpDiskSyncAdd failed, err: %v, rid: %s", err, kt.Rid)
			return err
		}
	}

	return nil
}

func diffGcpDiskSyncAdd(kt *kit.Kit, cloudMap map[string]*GcpDiskSyncDiff, req *SyncGcpDiskOption,
	addCloudIDs []string, bootMap map[string]struct{}, dataCli *dataservice.Client) ([]string, error) {

	var createReq dataproto.DiskExtBatchCreateReq[dataproto.GcpDiskExtensionCreateReq]

	for _, id := range addCloudIDs {
		disk := &dataproto.DiskExtCreateReq[dataproto.GcpDiskExtensionCreateReq]{
			AccountID: req.AccountID,
			Name:      cloudMap[id].Disk.Name,
			CloudID:   id,
			Region:    cloudMap[id].Disk.Region,
			Zone:      req.Zone,
			DiskSize:  uint64(cloudMap[id].Disk.SizeGb),
			DiskType:  cloudMap[id].Disk.Type,
			Status:    cloudMap[id].Disk.Status,
			Memo:      &cloudMap[id].Disk.Description,
			Extension: &dataproto.GcpDiskExtensionCreateReq{
				SelfLink:    cloudMap[id].Disk.SelfLink,
				SourceImage: cloudMap[id].Disk.SourceImage,
				Description: cloudMap[id].Disk.Description,
				// TODO: not find
				Encrypted: nil,
			},
		}
		if disk.Region == "" {
			region, err := getRegion(kt, dataCli, req.Zone)
			if err != nil {
				logs.Errorf("request gcp disk to get region failed, err: %v, rid: %s", err, kt.Rid)
				return nil, err
			}
			disk.Region = region
		}
		if _, exists := bootMap[id]; exists {
			disk.IsSystemDisk = true
		}
		createReq = append(createReq, disk)
	}

	if len(createReq) <= 0 {
		return make([]string, 0), nil
	}

	results, err := dataCli.Gcp.BatchCreateDisk(kt.Ctx, kt.Header(), &createReq)
	if err != nil {
		logs.Errorf("request dataservice to create gcp disk failed, err: %v, rid: %s", err, kt.Rid)
		return nil, err
	}

	return results.IDs, nil
}

func getRegion(kt *kit.Kit, dataCli *dataservice.Client, gcpZone string) (string, error) {
	listReq := &zone.ZoneListReq{
		Filter: &filter.Expression{
			Op: filter.And,
			Rules: []filter.RuleFactory{
				&filter.AtomRule{
					Field: "vendor",
					Op:    filter.Equal.Factory(),
					Value: enumor.Gcp,
				},
				&filter.AtomRule{
					Field: "name",
					Op:    filter.Equal.Factory(),
					Value: gcpZone,
				},
			},
		},
		Page: core.DefaultBasePage,
	}
	result, err := dataCli.Global.Zone.ListZone(kt.Ctx, kt.Header(), listReq)
	if err != nil {
		logs.Errorf("list gcp zone failed, err: %v, rid: %s", err, kt.Rid)
		return "", err
	}

	if len(result.Details) == 0 {
		return "", errors.New("gcp zone is empty")
	}

	return result.Details[0].Region, nil
}

func isGcpDiskChange(db *GcpDiskSyncDS, cloud *GcpDiskSyncDiff, isSystemDisk bool) bool {

	if cloud.Disk.Status != db.HcDisk.Status {
		return true
	}

	if cloud.Disk.Region != db.HcDisk.Region {
		return true
	}

	if cloud.Disk.Description != converter.PtrToVal(db.HcDisk.Memo) {
		return true
	}

	if cloud.Disk.SelfLink != db.HcDisk.Extension.SelfLink {
		return true
	}

	if cloud.Disk.SourceImage != db.HcDisk.Extension.SourceImage {
		return true
	}

	if cloud.Disk.Description != db.HcDisk.Extension.Description {
		return true
	}

	if isSystemDisk != db.HcDisk.IsSystemDisk {
		return true
	}

	return false
}

func diffGcpSyncUpdate(kt *kit.Kit, cloudMap map[string]*GcpDiskSyncDiff, req *SyncGcpDiskOption,
	dsMap map[string]*GcpDiskSyncDS, updateCloudIDs []string, bootMap map[string]struct{},
	dataCli *dataservice.Client) error {

	disks := make([]*dataproto.DiskExtUpdateReq[dataproto.GcpDiskExtensionUpdateReq], 0)

	for _, id := range updateCloudIDs {
		if cloudMap[id].Disk.Region == "" {
			region, err := getRegion(kt, dataCli, req.Zone)
			if err != nil {
				logs.Errorf("request gcp disk to get region failed, err: %v, rid: %s", err, kt.Rid)
				return err
			}
			cloudMap[id].Disk.Region = region
		}

		_, isSystemDisk := bootMap[cloudMap[id].Disk.SelfLink]

		if !isGcpDiskChange(dsMap[id], cloudMap[id], isSystemDisk) {
			continue
		}

		disk := &dataproto.DiskExtUpdateReq[dataproto.GcpDiskExtensionUpdateReq]{
			ID:           dsMap[id].HcDisk.ID,
			Region:       cloudMap[id].Disk.Region,
			Status:       cloudMap[id].Disk.Status,
			IsSystemDisk: isSystemDisk,
			Memo:         &cloudMap[id].Disk.Description,
			Extension: &dataproto.GcpDiskExtensionUpdateReq{
				SelfLink:    cloudMap[id].Disk.SelfLink,
				SourceImage: cloudMap[id].Disk.SourceImage,
				Description: cloudMap[id].Disk.Description,
				// TODO: not find
				Encrypted: nil,
			},
		}

		disks = append(disks, disk)
	}

	if len(disks) > 0 {
		elems := slice.Split(disks, typescore.TCloudQueryLimit)
		for _, partDisks := range elems {
			var updateReq dataproto.DiskExtBatchUpdateReq[dataproto.GcpDiskExtensionUpdateReq]
			for _, disk := range partDisks {
				updateReq = append(updateReq, disk)
			}
			if _, err := dataCli.Gcp.BatchUpdateDisk(kt.Ctx, kt.Header(), &updateReq); err != nil {
				logs.Errorf("request dataservice gcp BatchUpdateDisk failed, err: %v, rid: %s", err, kt.Rid)
				return err
			}
		}
	}

	return nil
}
