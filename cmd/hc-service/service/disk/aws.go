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
	"hcm/pkg/adaptor/types/disk"
	dataproto "hcm/pkg/api/data-service/cloud/disk"
	proto "hcm/pkg/api/hc-service/disk"
	protodisk "hcm/pkg/api/hc-service/disk"
	"hcm/pkg/criteria/constant"
	"hcm/pkg/criteria/errf"
	"hcm/pkg/logs"
	"hcm/pkg/rest"
)

// AwsCreateDisk ...
func AwsCreateDisk(da *diskAdaptor, cts *rest.Contexts) (interface{}, error) {
	req := new(proto.AwsDiskCreateReq)
	if err := cts.DecodeInto(req); err != nil {
		return nil, errf.NewFromErr(errf.DecodeRequestFailed, err)
	}
	if err := req.Validate(); err != nil {
		return nil, errf.NewFromErr(errf.InvalidParameter, err)
	}

	client, err := da.adaptor.Aws(cts.Kit, req.Base.AccountID)
	if err != nil {
		return nil, err
	}

	diskSize := int64(req.Base.DiskSize)
	opt := &disk.AwsDiskCreateOption{
		Region:   req.Base.Region,
		Zone:     &req.Base.Zone,
		DiskType: &req.Base.DiskType,
		DiskSize: &diskSize,
	}
	client.CreateDisk(cts.Kit, opt)

	// TODO save to data-service

	return nil, nil
}

// AwsSyncDisk...
func AwsSyncDisk(da *diskAdaptor, cts *rest.Contexts) (interface{}, error) {

	req, err := da.decodeDiskSyncReq(cts)
	if err != nil {
		logs.Errorf("request decodeDiskSyncReq failed, err: %v, rid: %s", err, cts.Kit.Rid)
		return nil, err
	}

	cloudMap, err := da.getDatasFromAwsForDiskSync(cts, req)
	if err != nil {
		logs.Errorf("request getDatasFromAwsForDiskSync failed, err: %v, rid: %s", err, cts.Kit.Rid)
		return nil, err
	}

	dsMap, err := da.getDatasFromDSForDiskSync(cts, req)
	if err != nil {
		logs.Errorf("request getDatasFromDSForDiskSync failed, err: %v, rid: %s", err, cts.Kit.Rid)
		return nil, err
	}

	err = da.diffAwsDiskSync(cts, cloudMap, dsMap, req)
	if err != nil {
		logs.Errorf("request diffAwsDiskSync failed, err: %v, rid: %s", err, cts.Kit.Rid)
		return nil, err
	}

	return nil, nil
}

// getDatasFromAwsForDiskSync get datas from cloud
func (da *diskAdaptor) getDatasFromAwsForDiskSync(cts *rest.Contexts,
	req *protodisk.DiskSyncReq) (map[string]*proto.AwsDiskSyncDiff, error) {

	client, err := da.adaptor.Aws(cts.Kit, req.AccountID)
	if err != nil {
		return nil, err
	}

	listOpt := &disk.AwsDiskListOption{
		Region: req.Region,
	}
	datas, err := client.ListDisk(cts.Kit, listOpt)
	if err != nil {
		logs.Errorf("request adaptor to list aws disk failed, err: %v, rid: %s", err, cts.Kit.Rid)
		return nil, err
	}

	cloudMap := make(map[string]*proto.AwsDiskSyncDiff)
	for _, data := range datas {
		sg := new(proto.AwsDiskSyncDiff)
		sg.Disk = data
		cloudMap[*data.VolumeId] = sg
	}

	return cloudMap, nil
}

// diffAwsDiskSync diff cloud data-service
func (da *diskAdaptor) diffAwsDiskSync(cts *rest.Contexts, cloudMap map[string]*proto.AwsDiskSyncDiff,
	dsMap map[string]*protodisk.DiskSyncDS, req *proto.DiskSyncReq) error {

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
		err := da.diffAwsSyncUpdate(cts, cloudMap, dsMap, updateCloudIDs)
		if err != nil {
			logs.Errorf("request diffAwsSyncUpdate failed, err: %v, rid: %s", err, cts.Kit.Rid)
			return err
		}
	}

	if len(addCloudIDs) > 0 {
		_, err := da.diffAwsDiskSyncAdd(cts, cloudMap, req, addCloudIDs)
		if err != nil {
			logs.Errorf("request diffAwsDiskSyncAdd failed, err: %v, rid: %s", err, cts.Kit.Rid)
			return err
		}
	}

	return nil
}

// diffAwsDiskSyncAdd for add
func (da *diskAdaptor) diffAwsDiskSyncAdd(cts *rest.Contexts, cloudMap map[string]*proto.AwsDiskSyncDiff,
	req *proto.DiskSyncReq, addCloudIDs []string) ([]string, error) {

	var createReq dataproto.DiskExtBatchCreateReq[dataproto.AwsDiskExtensionCreateReq]

	for _, id := range addCloudIDs {
		disk := &dataproto.DiskExtCreateReq[dataproto.AwsDiskExtensionCreateReq]{
			AccountID:  req.AccountID,
			Name:       constant.DiskDefaultName,
			CloudID:    id,
			Region:     req.Region,
			Zone:       *cloudMap[id].Disk.AvailabilityZone,
			DiskSize:   uint64(*cloudMap[id].Disk.Size),
			DiskType:   *cloudMap[id].Disk.VolumeType,
			DiskStatus: *cloudMap[id].Disk.State,
		}
		createReq = append(createReq, disk)
	}

	if len(createReq) <= 0 {
		return make([]string, 0), nil
	}

	results, err := da.dataCli.Aws.BatchCreateDisk(cts.Kit.Ctx, cts.Kit.Header(), &createReq)
	if err != nil {
		logs.Errorf("request dataservice to create aws disk failed, err: %v, rid: %s", err, cts.Kit.Rid)
		return nil, err
	}

	return results.IDs, nil
}

// diffHuaWeiSyncUpdate for update
func (da *diskAdaptor) diffAwsSyncUpdate(cts *rest.Contexts, cloudMap map[string]*proto.AwsDiskSyncDiff,
	dsMap map[string]*protodisk.DiskSyncDS, updateCloudIDs []string) error {

	var updateReq dataproto.DiskExtBatchUpadteReq[dataproto.AwsDiskExtensionUpdateReq]

	for _, id := range updateCloudIDs {
		if *cloudMap[id].Disk.State == dsMap[id].HcDisk.DiskStatus {
			continue
		}
		disk := &dataproto.DiskExtUpdateReq[dataproto.AwsDiskExtensionUpdateReq]{
			ID:         dsMap[id].HcDisk.ID,
			DiskStatus: *cloudMap[id].Disk.State,
		}
		updateReq = append(updateReq, disk)
	}

	if len(updateReq) > 0 {
		if _, err := da.dataCli.Aws.BatchUpdateDisk(cts.Kit.Ctx, cts.Kit.Header(), &updateReq); err != nil {
			logs.Errorf("request dataservice aws BatchUpdateDisk failed, err: %v, rid: %s", err, cts.Kit.Rid)
			return err
		}
	}

	return nil
}
