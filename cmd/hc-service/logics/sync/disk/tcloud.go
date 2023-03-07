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
	"hcm/pkg/adaptor/tcloud"
	"hcm/pkg/adaptor/types/core"
	"hcm/pkg/adaptor/types/disk"
	dataproto "hcm/pkg/api/data-service/cloud/disk"
	protodisk "hcm/pkg/api/hc-service/disk"
	dataservice "hcm/pkg/client/data-service"
	"hcm/pkg/kit"
	"hcm/pkg/logs"

	cbs "github.com/tencentcloud/tencentcloud-sdk-go/tencentcloud/cbs/v20170312"
)

// SyncTCloudDisk sync disk self
func SyncTCloudDisk(kt *kit.Kit, req *protodisk.DiskSyncReq,
	ad *cloudclient.CloudAdaptorClient, dataCli *dataservice.Client) (interface{}, error) {

	cloudMap, err := getDatasFromTCloudForDiskSync(kt, req, ad)
	if err != nil {
		logs.Errorf("request getDatasFromTCloudForDiskSync failed, err: %v, rid: %s", err, kt.Rid)
		return nil, err
	}

	dsMap, err := GetDatasFromDSForDiskSync(kt, req, dataCli)
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

func getDatasFromTCloudForDiskSync(kt *kit.Kit, req *protodisk.DiskSyncReq,
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
	req *protodisk.DiskSyncReq) (map[string]*TCloudDiskSyncDiff, error) {

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
	req *protodisk.DiskSyncReq) (map[string]*TCloudDiskSyncDiff, error) {

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
	dsMap map[string]*DiskSyncDS, req *protodisk.DiskSyncReq, dataCli *dataservice.Client) error {
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
	req *protodisk.DiskSyncReq, addCloudIDs []string, dataCli *dataservice.Client) ([]string, error) {
	var createReq dataproto.DiskExtBatchCreateReq[dataproto.TCloudDiskExtensionCreateReq]

	for _, id := range addCloudIDs {
		disk := &dataproto.DiskExtCreateReq[dataproto.TCloudDiskExtensionCreateReq]{
			AccountID: req.AccountID,
			Name:      *cloudMap[id].Disk.DiskName,
			CloudID:   id,
			Region:    req.Region,
			Zone:      *cloudMap[id].Disk.Placement.Zone,
			DiskSize:  *cloudMap[id].Disk.DiskSize,
			DiskType:  *cloudMap[id].Disk.DiskType,
			Status:    *cloudMap[id].Disk.DiskState,
			Extension: &dataproto.TCloudDiskExtensionCreateReq{
				DiskChargeType: *cloudMap[id].Disk.DiskChargeType,
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

func diffTCloudDiskSyncUpdate(kt *kit.Kit, cloudMap map[string]*TCloudDiskSyncDiff,
	dsMap map[string]*DiskSyncDS, updateCloudIDs []string, dataCli *dataservice.Client) error {
	var updateReq dataproto.DiskExtBatchUpdateReq[dataproto.TCloudDiskExtensionUpdateReq]

	for _, id := range updateCloudIDs {
		if *cloudMap[id].Disk.DiskState == dsMap[id].HcDisk.Status {
			continue
		}
		disk := &dataproto.DiskExtUpdateReq[dataproto.TCloudDiskExtensionUpdateReq]{
			ID:     dsMap[id].HcDisk.ID,
			Status: *cloudMap[id].Disk.DiskState,
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
