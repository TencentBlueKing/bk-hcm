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
	typecore "hcm/pkg/adaptor/types/core"
	"hcm/pkg/adaptor/types/disk"
	dataproto "hcm/pkg/api/data-service/cloud/disk"
	protodisk "hcm/pkg/api/hc-service/disk"
	dataservice "hcm/pkg/client/data-service"
	"hcm/pkg/kit"
	"hcm/pkg/logs"
	"hcm/pkg/runtime/filter"
)

// SyncGcpDisk sync disk self
func SyncGcpDisk(kt *kit.Kit, req *protodisk.DiskSyncReq,
	ad *cloudclient.CloudAdaptorClient, dataCli *dataservice.Client) (interface{}, error) {

	cloudMap, err := getDatasFromGcpForDiskSync(kt, req, ad)
	if err != nil {
		logs.Errorf("request getDatasFromGcpForDiskSync failed, err: %v, rid: %s", err, kt.Rid)
		return nil, err
	}

	dsMap, err := GetDatasFromDSForDiskSync(kt, req, dataCli)
	if err != nil {
		logs.Errorf("request getDatasFromDSForDiskSync failed, err: %v, rid: %s", err, kt.Rid)
		return nil, err
	}

	err = diffGcpDiskSync(kt, cloudMap, dsMap, req, dataCli)
	if err != nil {
		logs.Errorf("request diffGcpDiskSync failed, err: %v, rid: %s", err, kt.Rid)
		return nil, err
	}

	return nil, nil
}

func getDatasFromGcpForDiskSync(kt *kit.Kit, req *protodisk.DiskSyncReq,
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

		datas, token, err := client.ListDisk(kt, listOpt)
		if err != nil {
			logs.Errorf("request adaptor to list gcp disk failed, err: %v, rid: %s", err, kt.Rid)
			return nil, err
		}

		for _, data := range datas {
			sg := new(GcpDiskSyncDiff)
			sg.Disk = data
			cloudMap[fmt.Sprint(data.Id)] = sg
		}

		if len(token) == 0 {
			break
		}
		nextToken = token
	}

	return cloudMap, nil
}

func diffGcpDiskSync(kt *kit.Kit, cloudMap map[string]*GcpDiskSyncDiff,
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
		err := diffGcpSyncUpdate(kt, cloudMap, dsMap, updateCloudIDs, dataCli)
		if err != nil {
			logs.Errorf("request diffGcpSyncUpdate failed, err: %v, rid: %s", err, kt.Rid)
			return err
		}
	}

	if len(addCloudIDs) > 0 {
		_, err := diffGcpDiskSyncAdd(kt, cloudMap, req, addCloudIDs, dataCli)
		if err != nil {
			logs.Errorf("request diffGcpDiskSyncAdd failed, err: %v, rid: %s", err, kt.Rid)
			return err
		}
	}

	return nil
}

func diffGcpDiskSyncAdd(kt *kit.Kit, cloudMap map[string]*GcpDiskSyncDiff,
	req *protodisk.DiskSyncReq, addCloudIDs []string, dataCli *dataservice.Client) ([]string, error) {

	var createReq dataproto.DiskExtBatchCreateReq[dataproto.GcpDiskExtensionCreateReq]

	for _, id := range addCloudIDs {
		disk := &dataproto.DiskExtCreateReq[dataproto.GcpDiskExtensionCreateReq]{
			AccountID:  req.AccountID,
			Name:       cloudMap[id].Disk.Name,
			CloudID:    id,
			Region:     req.Region,
			Zone:       cloudMap[id].Disk.Zone,
			DiskSize:   uint64(cloudMap[id].Disk.SizeGb),
			DiskType:   cloudMap[id].Disk.Type,
			DiskStatus: cloudMap[id].Disk.Status,
			Memo:       &cloudMap[id].Disk.Description,
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

func diffGcpSyncUpdate(kt *kit.Kit, cloudMap map[string]*GcpDiskSyncDiff,
	dsMap map[string]*DiskSyncDS, updateCloudIDs []string, dataCli *dataservice.Client) error {

	var updateReq dataproto.DiskExtBatchUpdateReq[dataproto.AwsDiskExtensionUpdateReq]

	for _, id := range updateCloudIDs {
		if cloudMap[id].Disk.Status == dsMap[id].HcDisk.DiskStatus &&
			cloudMap[id].Disk.Description == *dsMap[id].HcDisk.Memo {
			continue
		}
		disk := &dataproto.DiskExtUpdateReq[dataproto.AwsDiskExtensionUpdateReq]{
			ID:         dsMap[id].HcDisk.ID,
			DiskStatus: cloudMap[id].Disk.Status,
			Memo:       &cloudMap[id].Disk.Description,
		}
		updateReq = append(updateReq, disk)
	}

	if len(updateReq) > 0 {
		if _, err := dataCli.Aws.BatchUpdateDisk(kt.Ctx, kt.Header(), &updateReq); err != nil {
			logs.Errorf("request dataservice gcp BatchUpdateDisk failed, err: %v, rid: %s", err, kt.Rid)
			return err
		}
	}

	return nil
}
