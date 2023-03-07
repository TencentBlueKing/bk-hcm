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
	cloudclient "hcm/cmd/hc-service/service/cloud-adaptor"
	"hcm/pkg/adaptor/types/disk"
	"hcm/pkg/api/core"
	datadisk "hcm/pkg/api/data-service/cloud/disk"
	dataproto "hcm/pkg/api/data-service/cloud/disk"
	protodisk "hcm/pkg/api/hc-service/disk"
	dataservice "hcm/pkg/client/data-service"
	"hcm/pkg/kit"
	"hcm/pkg/logs"
	"hcm/pkg/runtime/filter"
)

// SyncAzureDisk sync disk self
func SyncAzureDisk(kt *kit.Kit, req *protodisk.DiskSyncReq,
	ad *cloudclient.CloudAdaptorClient, dataCli *dataservice.Client) (interface{}, error) {

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

	err = diffAzureDiskSync(kt, cloudMap, dsMap, req, dataCli)
	if err != nil {
		logs.Errorf("request diffAzureDiskSync failed, err: %v, rid: %s", err, kt.Rid)
		return nil, err
	}

	return nil, nil
}

func getDatasFromAzureForDiskSync(kt *kit.Kit, req *protodisk.DiskSyncReq,
	ad *cloudclient.CloudAdaptorClient) (map[string]*AzureDiskSyncDiff, error) {

	client, err := ad.Azure(kt, req.AccountID)
	if err != nil {
		return nil, err
	}

	listOpt := &disk.AzureDiskListOption{
		ResourceGroupName: req.ResourceGroupName,
	}
	if len(req.CloudIDs) > 0 {
		listOpt.CloudIDs = req.CloudIDs
	}
	datas, err := client.ListDisk(kt, listOpt)
	if err != nil {
		logs.Errorf("request adaptor to list azure disk failed, err: %v, rid: %s", err, kt.Rid)
		return nil, err
	}

	cloudMap := make(map[string]*AzureDiskSyncDiff)
	for _, data := range datas {
		sg := new(AzureDiskSyncDiff)
		sg.Disk = data
		cloudMap[*data.ID] = sg
	}

	return cloudMap, nil
}

func getDatasFromAzureDSForDiskSync(kt *kit.Kit, req *protodisk.DiskSyncReq,
	dataCli *dataservice.Client) (map[string]*DiskSyncDS, error) {

	start := 0
	resultsHcm := make([]*datadisk.DiskResult, 0)
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
					filter.AtomRule{
						Field: "region",
						Op:    filter.Equal.Factory(),
						Value: req.Region,
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

		results, err := dataCli.Global.ListDisk(kt.Ctx, kt.Header(), dataReq)
		if err != nil {
			logs.Errorf("from data-service list disk failed, err: %v, rid: %s", err, kt.Rid)
			return nil, err
		}

		resultsHcm = append(resultsHcm, results.Details...)
		start += len(results.Details)
		if uint(len(results.Details)) < core.DefaultMaxPageLimit {
			break
		}
	}

	dsMap := make(map[string]*DiskSyncDS)
	for _, result := range resultsHcm {
		sg := new(DiskSyncDS)
		sg.IsUpdated = false
		sg.HcDisk = result
		dsMap[result.CloudID] = sg
	}

	return dsMap, nil
}

func diffAzureDiskSync(kt *kit.Kit, cloudMap map[string]*AzureDiskSyncDiff,
	dsMap map[string]*DiskSyncDS, req *protodisk.DiskSyncReq, dataCli *dataservice.Client) error {

	addCloudIDs := getAddCloudIDs(cloudMap, dsMap)
	deleteCloudIDs, updateCloudIDs := getDeleteAndUpdateCloudIDs(dsMap)

	if len(deleteCloudIDs) > 0 {
		err := diffDiskSyncDelete(kt, deleteCloudIDs, dataCli)
		if err != nil {
			logs.Errorf("request diffDiskSyncDeletek failed, err: %v, rid: %s", err, kt.Rid)
			return err
		}
	}

	if len(updateCloudIDs) > 0 {
		err := diffAzureSyncUpdate(kt, cloudMap, dsMap, updateCloudIDs, dataCli)
		if err != nil {
			logs.Errorf("request diffAzureSyncUpdate failed, err: %v, rid: %s", err, kt.Rid)
			return err
		}
	}

	if len(addCloudIDs) > 0 {
		_, err := diffAzureDiskSyncAdd(kt, cloudMap, req, addCloudIDs, dataCli)
		if err != nil {
			logs.Errorf("request diffAzureDiskSyncAdd failed, err: %v, rid: %s", err, kt.Rid)
			return err
		}
	}

	return nil
}

func diffAzureDiskSyncAdd(kt *kit.Kit, cloudMap map[string]*AzureDiskSyncDiff, req *protodisk.DiskSyncReq,
	addCloudIDs []string, dataCli *dataservice.Client) ([]string, error) {

	var createReq dataproto.DiskExtBatchCreateReq[dataproto.AzureDiskExtensionCreateReq]

	for _, id := range addCloudIDs {
		disk := &dataproto.DiskExtCreateReq[dataproto.AzureDiskExtensionCreateReq]{
			AccountID: req.AccountID,
			Name:      *cloudMap[id].Disk.Name,
			CloudID:   id,
			Region:    req.Region,
			Zone:      *cloudMap[id].Disk.Location,
			DiskSize:  uint64(*cloudMap[id].Disk.Properties.DiskSizeBytes),
			DiskType:  *cloudMap[id].Disk.Type,
			Status:    string(*cloudMap[id].Disk.Properties.DiskState),
			Extension: &dataproto.AzureDiskExtensionCreateReq{
				ResourceGroupName: req.ResourceGroupName,
			},
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

func diffAzureSyncUpdate(kt *kit.Kit, cloudMap map[string]*AzureDiskSyncDiff,
	dsMap map[string]*DiskSyncDS, updateCloudIDs []string, dataCli *dataservice.Client) error {

	var updateReq dataproto.DiskExtBatchUpdateReq[dataproto.AzureDiskExtensionUpdateReq]

	for _, id := range updateCloudIDs {
		if string(*cloudMap[id].Disk.Properties.DiskState) == dsMap[id].HcDisk.Status {
			continue
		}

		disk := &dataproto.DiskExtUpdateReq[dataproto.AzureDiskExtensionUpdateReq]{
			ID:     dsMap[id].HcDisk.ID,
			Status: string(*cloudMap[id].Disk.Properties.DiskState),
		}
		updateReq = append(updateReq, disk)
	}

	if len(updateReq) > 0 {
		if _, err := dataCli.Azure.BatchUpdateDisk(kt.Ctx, kt.Header(), &updateReq); err != nil {
			logs.Errorf("request dataservice azure BatchUpdateDisk failed, err: %v, rid: %s", err, kt.Rid)
			return err
		}
	}

	return nil
}
