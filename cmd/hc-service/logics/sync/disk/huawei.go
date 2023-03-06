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
	"hcm/pkg/adaptor/types/core"
	"hcm/pkg/adaptor/types/disk"
	dataproto "hcm/pkg/api/data-service/cloud/disk"
	protodisk "hcm/pkg/api/hc-service/disk"
	dataservice "hcm/pkg/client/data-service"
	"hcm/pkg/kit"
	"hcm/pkg/logs"
)

// SyncHuaWeiDisk sync disk self
func SyncHuaWeiDisk(kt *kit.Kit, req *protodisk.DiskSyncReq,
	ad *cloudclient.CloudAdaptorClient, dataCli *dataservice.Client) (interface{}, error) {

	cloudMap, err := getDatasFromHuaWeiForDiskSync(kt, req, ad)
	if err != nil {
		logs.Errorf("request getDatasFromHuaWeiForDiskSync failed, err: %v, rid: %s", err, kt.Rid)
		return nil, err
	}

	dsMap, err := GetDatasFromDSForDiskSync(kt, req, dataCli)
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

func getDatasFromHuaWeiForDiskSync(kt *kit.Kit, req *protodisk.DiskSyncReq,
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
		sg := new(HuaWeiDiskSyncDiff)
		sg.Disk = data
		cloudMap[data.Id] = sg
	}

	return cloudMap, nil
}

func diffHuaWeiDiskSync(kt *kit.Kit, cloudMap map[string]*HuaWeiDiskSyncDiff, dsMap map[string]*DiskSyncDS,
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

func diffHuaWeiDiskSyncAdd(kt *kit.Kit, cloudMap map[string]*HuaWeiDiskSyncDiff, req *protodisk.DiskSyncReq,
	addCloudIDs []string, dataCli *dataservice.Client) ([]string, error) {

	var createReq dataproto.DiskExtBatchCreateReq[dataproto.HuaWeiDiskExtensionCreateReq]

	for _, id := range addCloudIDs {
		disk := &dataproto.DiskExtCreateReq[dataproto.HuaWeiDiskExtensionCreateReq]{
			AccountID:  req.AccountID,
			Name:       cloudMap[id].Disk.Name,
			CloudID:    id,
			Region:     req.Region,
			Zone:       cloudMap[id].Disk.AvailabilityZone,
			DiskSize:   uint64(cloudMap[id].Disk.Size),
			DiskType:   cloudMap[id].Disk.VolumeType,
			DiskStatus: cloudMap[id].Disk.Status,
			Memo:       &cloudMap[id].Disk.Description,
			Extension: &dataproto.HuaWeiDiskExtensionCreateReq{
				DiskChargeType: cloudMap[id].Disk.ServiceType,
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

func diffHuaWeiSyncUpdate(kt *kit.Kit, cloudMap map[string]*HuaWeiDiskSyncDiff, dsMap map[string]*DiskSyncDS,
	updateCloudIDs []string, dataCli *dataservice.Client) error {

	var updateReq dataproto.DiskExtBatchUpdateReq[dataproto.HuaWeiDiskExtensionUpdateReq]

	for _, id := range updateCloudIDs {
		if cloudMap[id].Disk.Description == *dsMap[id].HcDisk.Memo &&
			cloudMap[id].Disk.Status == dsMap[id].HcDisk.DiskStatus {
			continue
		}
		disk := &dataproto.DiskExtUpdateReq[dataproto.HuaWeiDiskExtensionUpdateReq]{
			ID:         dsMap[id].HcDisk.ID,
			Memo:       &cloudMap[id].Disk.Description,
			DiskStatus: cloudMap[id].Disk.Status,
		}
		updateReq = append(updateReq, disk)
	}

	if len(updateReq) > 0 {
		if _, err := dataCli.HuaWei.BatchUpdateDisk(kt.Ctx, kt.Header(), &updateReq); err != nil {
			logs.Errorf("request dataservice huawei BatchUpdateDisk failed, err: %v, rid: %s", err, kt.Rid)
			return err
		}
	}

	return nil
}
