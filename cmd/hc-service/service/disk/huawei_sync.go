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
)

// HuaWeiSyncDisk sync huawei to hcm
func HuaWeiSyncDisk(da *diskAdaptor, cts *rest.Contexts) (interface{}, error) {
	req, err := da.decodeDiskSyncReq(cts)
	if err != nil {
		logs.Errorf("request decodeDiskSyncReq failed, err: %v, rid: %s", err, cts.Kit.Rid)
		return nil, err
	}

	cloudMap, err := da.getDatasFromHuaWeiForDiskSync(cts, req)
	if err != nil {
		logs.Errorf("request getDatasFromHuaWeiForDiskSync failed, err: %v, rid: %s", err, cts.Kit.Rid)
		return nil, err
	}

	dsMap, err := da.getDatasFromDSForDiskSync(cts, req)
	if err != nil {
		logs.Errorf("request getDatasFromDSForDiskSync failed, err: %v, rid: %s", err, cts.Kit.Rid)
		return nil, err
	}

	err = da.diffHuaWeiDiskSync(cts, cloudMap, dsMap, req)
	if err != nil {
		logs.Errorf("request diffHuaWeiDiskSync failed, err: %v, rid: %s", err, cts.Kit.Rid)
		return nil, err
	}

	return nil, nil
}

// getDatasFromHuaWeiForDiskSync get datas from cloud
func (da *diskAdaptor) getDatasFromHuaWeiForDiskSync(cts *rest.Contexts,
	req *protodisk.DiskSyncReq,
) (map[string]*proto.HuaWeiDiskSyncDiff, error) {
	client, err := da.adaptor.HuaWei(cts.Kit, req.AccountID)
	if err != nil {
		return nil, err
	}

	// TODO 分页逻辑
	limit := int32(core.HuaWeiQueryLimit)
	var marker *string = nil
	opt := &disk.HuaWeiDiskListOption{
		Region: req.Region,
		Page:   &core.HuaWeiPage{Limit: &limit, Marker: marker},
	}
	datas, err := client.ListDisk(cts.Kit, opt)
	if err != nil {
		logs.Errorf("request adaptor to list huawei disk failed, opt: %v, err: %v, rid: %s", opt, err,
			cts.Kit.Rid)
		return nil, err
	}

	cloudMap := make(map[string]*proto.HuaWeiDiskSyncDiff)
	for _, data := range datas {
		sg := new(proto.HuaWeiDiskSyncDiff)
		sg.Disk = data
		cloudMap[data.Id] = sg
	}

	return cloudMap, nil
}

// diffTCloudDiskSync diff cloud data-service
func (da *diskAdaptor) diffHuaWeiDiskSync(cts *rest.Contexts, cloudMap map[string]*proto.HuaWeiDiskSyncDiff,
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
		err := da.diffHuaWeiSyncUpdate(cts, cloudMap, dsMap, updateCloudIDs)
		if err != nil {
			logs.Errorf("request diffHuaWeiSyncUpdate failed, err: %v, rid: %s", err, cts.Kit.Rid)
			return err
		}
	}
	if len(addCloudIDs) > 0 {
		_, err := da.diffHuaWeiDiskSyncAdd(cts, cloudMap, req, addCloudIDs)
		if err != nil {
			logs.Errorf("request diffHuaWeiDiskSyncAdd failed, err: %v, rid: %s", err, cts.Kit.Rid)
			return err
		}
	}

	return nil
}

// diffHuaWeiDiskSyncAdd for add
func (da *diskAdaptor) diffHuaWeiDiskSyncAdd(cts *rest.Contexts, cloudMap map[string]*proto.HuaWeiDiskSyncDiff,
	req *proto.DiskSyncReq, addCloudIDs []string,
) ([]string, error) {
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

	results, err := da.dataCli.HuaWei.BatchCreateDisk(cts.Kit.Ctx, cts.Kit.Header(), &createReq)
	if err != nil {
		logs.Errorf("request dataservice to create huawei disk failed, err: %v, rid: %s", err, cts.Kit.Rid)
		return nil, err
	}

	return results.IDs, nil
}

// diffHuaWeiSyncUpdate for update
func (da *diskAdaptor) diffHuaWeiSyncUpdate(cts *rest.Contexts, cloudMap map[string]*proto.HuaWeiDiskSyncDiff,
	dsMap map[string]*protodisk.DiskSyncDS, updateCloudIDs []string,
) error {
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
		if _, err := da.dataCli.HuaWei.BatchUpdateDisk(cts.Kit.Ctx, cts.Kit.Header(), &updateReq); err != nil {
			logs.Errorf("request dataservice huawei BatchUpdateDisk failed, err: %v, rid: %s", err, cts.Kit.Rid)
			return err
		}
	}

	return nil
}
