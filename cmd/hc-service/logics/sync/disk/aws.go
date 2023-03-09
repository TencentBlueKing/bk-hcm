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
	dataproto "hcm/pkg/api/data-service/cloud/disk"
	protodisk "hcm/pkg/api/hc-service/disk"
	dataservice "hcm/pkg/client/data-service"
	"hcm/pkg/kit"
	"hcm/pkg/logs"
)

// SyncAwsDisk sync disk self
func SyncAwsDisk(kt *kit.Kit, req *protodisk.DiskSyncReq,
	ad *cloudclient.CloudAdaptorClient, dataCli *dataservice.Client) (interface{}, error) {

	cloudMap, err := getDatasFromAwsForDiskSync(kt, req, ad)
	if err != nil {
		logs.Errorf("request getDatasFromAwsForDiskSync failed, err: %v, rid: %s", err, kt.Rid)
		return nil, err
	}

	dsMap, err := GetDatasFromDSForDiskSync(kt, req, dataCli)
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

func getDatasFromAwsForDiskSync(kt *kit.Kit, req *protodisk.DiskSyncReq,
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

	datas, err := client.ListDisk(kt, listOpt)
	if err != nil {
		logs.Errorf("request adaptor to list aws disk failed, err: %v, rid: %s", err, kt.Rid)
		return nil, err
	}

	cloudMap := make(map[string]*AwsDiskSyncDiff)
	for _, data := range datas {
		sg := new(AwsDiskSyncDiff)
		sg.Disk = data
		cloudMap[*data.VolumeId] = sg
	}

	return cloudMap, nil
}

func diffAwsDiskSync(kt *kit.Kit, cloudMap map[string]*AwsDiskSyncDiff, dsMap map[string]*DiskSyncDS,
	req *protodisk.DiskSyncReq, dataCli *dataservice.Client) error {

	addCloudIDs := getAddCloudIDs(cloudMap, dsMap)
	deleteCloudIDs, updateCloudIDs := getDeleteAndUpdateCloudIDs(dsMap)

	if len(deleteCloudIDs) > 0 {
		err := diffDiskSyncDelete(kt, deleteCloudIDs, dataCli)
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
	req *protodisk.DiskSyncReq, addCloudIDs []string, dataCli *dataservice.Client) ([]string, error) {

	var createReq dataproto.DiskExtBatchCreateReq[dataproto.AwsDiskExtensionCreateReq]

	for _, id := range addCloudIDs {
		disk := &dataproto.DiskExtCreateReq[dataproto.AwsDiskExtensionCreateReq]{
			AccountID: req.AccountID,
			Name:      "todo",
			CloudID:   id,
			Region:    req.Region,
			Zone:      *cloudMap[id].Disk.AvailabilityZone,
			DiskSize:  uint64(*cloudMap[id].Disk.Size),
			DiskType:  *cloudMap[id].Disk.VolumeType,
			Status:    *cloudMap[id].Disk.State,
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

func diffAwsSyncUpdate(kt *kit.Kit, cloudMap map[string]*AwsDiskSyncDiff,
	dsMap map[string]*DiskSyncDS, updateCloudIDs []string, dataCli *dataservice.Client) error {

	var updateReq dataproto.DiskExtBatchUpdateReq[dataproto.AwsDiskExtensionUpdateReq]

	for _, id := range updateCloudIDs {
		if *cloudMap[id].Disk.State == dsMap[id].HcDisk.Status {
			continue
		}
		disk := &dataproto.DiskExtUpdateReq[dataproto.AwsDiskExtensionUpdateReq]{
			ID:     dsMap[id].HcDisk.ID,
			Status: *cloudMap[id].Disk.State,
		}
		updateReq = append(updateReq, disk)
	}

	if len(updateReq) > 0 {
		if _, err := dataCli.Aws.BatchUpdateDisk(kt.Ctx, kt.Header(), &updateReq); err != nil {
			logs.Errorf("request dataservice aws BatchUpdateDisk failed, err: %v, rid: %s", err, kt.Rid)
			return err
		}
	}

	return nil
}
