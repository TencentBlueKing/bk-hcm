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
	"hcm/pkg/adaptor/types/core"
	"hcm/pkg/adaptor/types/disk"
	dataproto "hcm/pkg/api/data-service/cloud/disk"
	proto "hcm/pkg/api/hc-service/disk"
	protodisk "hcm/pkg/api/hc-service/disk"
	"hcm/pkg/logs"
	"hcm/pkg/rest"

	cbs "github.com/tencentcloud/tencentcloud-sdk-go/tencentcloud/cbs/v20170312"
)

// TCloudSyncDisk sync tcloud to hcm
func TCloudSyncDisk(da *diskAdaptor, cts *rest.Contexts) (interface{}, error) {
	req, err := da.decodeDiskSyncReq(cts)
	if err != nil {
		logs.Errorf("request decodeDiskSyncReq failed, err: %v, rid: %s", err, cts.Kit.Rid)
		return nil, err
	}

	cloudMap, err := da.getDatasFromTCloudForDiskSync(cts, req)
	if err != nil {
		logs.Errorf("request getDatasFromTCloudForDiskSync failed, err: %v, rid: %s", err, cts.Kit.Rid)
		return nil, err
	}

	dsMap, err := da.getDatasFromDSForDiskSync(cts, req)
	if err != nil {
		logs.Errorf("request getDatasFromDSForDiskSync failed, err: %v, rid: %s", err, cts.Kit.Rid)
		return nil, err
	}

	err = da.diffTCloudDiskSync(cts, cloudMap, dsMap, req)
	if err != nil {
		logs.Errorf("request diffTCloudDiskSync failed, err: %v, rid: %s", err, cts.Kit.Rid)
		return nil, err
	}
	return nil, nil
}

// getDatasFromTCloudForDiskSync get datas from cloud
func (da *diskAdaptor) getDatasFromTCloudForDiskSync(cts *rest.Contexts,
	req *protodisk.DiskSyncReq,
) (map[string]*proto.TCloudDiskSyncDiff, error) {
	client, err := da.adaptor.TCloud(cts.Kit, req.AccountID)
	if err != nil {
		return nil, err
	}

	offset := 0
	datasCloud := make([]*cbs.Disk, 0)
	for {
		opt := &disk.TCloudDiskListOption{
			Region: req.Region,
			Page:   &core.TCloudPage{Offset: uint64(offset), Limit: uint64(core.TCloudQueryLimit)},
		}

		datas, err := client.ListDisk(cts.Kit, opt)
		if err != nil {
			logs.Errorf("request adaptor to list tcloud disk failed, err: %v, opt: %v, rid: %s", err, opt, cts.Kit.Rid)
			return nil, err
		}

		offset += len(datas)
		datasCloud = append(datasCloud, datas...)
		if uint(len(datas)) < core.TCloudQueryLimit {
			break
		}
	}

	cloudMap := make(map[string]*proto.TCloudDiskSyncDiff)
	for _, data := range datasCloud {
		sg := new(proto.TCloudDiskSyncDiff)
		sg.Disk = data
		cloudMap[*data.DiskId] = sg
	}

	return cloudMap, nil
}

// diffTCloudDiskSync diff cloud data-service
func (da *diskAdaptor) diffTCloudDiskSync(cts *rest.Contexts, cloudMap map[string]*proto.TCloudDiskSyncDiff,
	dsMap map[string]*protodisk.DiskSyncDS, req *proto.DiskSyncReq,
) error {
	addCloudIDs := getAddCloudIDs(cloudMap, dsMap)
	deleteCloudIDs, updateCloudIDs := getDeleteAndUpdateCloudIDs(dsMap)

	if len(deleteCloudIDs) > 0 {
		err := da.diffDiskSyncDelete(cts, deleteCloudIDs)
		if err != nil {
			logs.Errorf("request diffDiskSyncDelete failed, err: %v, rid: %s", err, cts.Kit.Rid)
			return err
		}
	}

	if len(updateCloudIDs) > 0 {
		err := da.diffTCloudDiskSyncUpdate(cts, cloudMap, dsMap, updateCloudIDs)
		if err != nil {
			logs.Errorf("request diffTCloudDiskSyncUpdate failed, err: %v, rid: %s", err, cts.Kit.Rid)
			return err
		}
	}

	if len(addCloudIDs) > 0 {
		_, err := da.diffTCloudDiskSyncAdd(cts, cloudMap, req, addCloudIDs)
		if err != nil {
			logs.Errorf("request diffTCloudDiskSyncAdd failed, err: %v, rid: %s", err, cts.Kit.Rid)
			return err
		}
	}

	return nil
}

// diffTCloudDiskSyncAdd for add
func (da *diskAdaptor) diffTCloudDiskSyncAdd(cts *rest.Contexts, cloudMap map[string]*proto.TCloudDiskSyncDiff,
	req *proto.DiskSyncReq, addCloudIDs []string,
) ([]string, error) {
	var createReq dataproto.DiskExtBatchCreateReq[dataproto.TCloudDiskExtensionCreateReq]

	for _, id := range addCloudIDs {
		disk := &dataproto.DiskExtCreateReq[dataproto.TCloudDiskExtensionCreateReq]{
			AccountID:  req.AccountID,
			Name:       *cloudMap[id].Disk.DiskName,
			CloudID:    id,
			Region:     req.Region,
			Zone:       *cloudMap[id].Disk.Placement.Zone,
			DiskSize:   *cloudMap[id].Disk.DiskSize,
			DiskType:   *cloudMap[id].Disk.DiskType,
			DiskStatus: *cloudMap[id].Disk.DiskState,
			Extension: &dataproto.TCloudDiskExtensionCreateReq{
				DiskChargeType: *cloudMap[id].Disk.DiskChargeType,
			},
		}
		createReq = append(createReq, disk)
	}

	if len(createReq) <= 0 {
		return make([]string, 0), nil
	}

	results, err := da.dataCli.TCloud.BatchCreateDisk(cts.Kit.Ctx, cts.Kit.Header(), &createReq)
	if err != nil {
		logs.Errorf("request dataservice to create tcloud disk failed, err: %v, rid: %s", err, cts.Kit.Rid)
		return nil, err
	}

	return results.IDs, nil
}

// diffSecurityGroupSyncUpdate for update
func (da *diskAdaptor) diffTCloudDiskSyncUpdate(cts *rest.Contexts, cloudMap map[string]*proto.TCloudDiskSyncDiff,
	dsMap map[string]*protodisk.DiskSyncDS, updateCloudIDs []string,
) error {
	var updateReq dataproto.DiskExtBatchUpdateReq[dataproto.TCloudDiskExtensionUpdateReq]

	for _, id := range updateCloudIDs {
		if *cloudMap[id].Disk.DiskState == dsMap[id].HcDisk.DiskStatus {
			continue
		}
		disk := &dataproto.DiskExtUpdateReq[dataproto.TCloudDiskExtensionUpdateReq]{
			ID:         dsMap[id].HcDisk.ID,
			DiskStatus: *cloudMap[id].Disk.DiskState,
		}
		updateReq = append(updateReq, disk)
	}

	if len(updateReq) > 0 {
		if _, err := da.dataCli.TCloud.BatchUpdateDisk(cts.Kit.Ctx, cts.Kit.Header(), &updateReq); err != nil {
			logs.Errorf("request dataservice BatchUpdateDisk failed, err: %v, rid: %s", err, cts.Kit.Rid)
			return err
		}
	}

	return nil
}
