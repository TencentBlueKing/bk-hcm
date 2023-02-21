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
	"hcm/pkg/api/core"
	datadisk "hcm/pkg/api/data-service/cloud/disk"
	dataproto "hcm/pkg/api/data-service/cloud/disk"
	proto "hcm/pkg/api/hc-service/disk"
	protodisk "hcm/pkg/api/hc-service/disk"
	"hcm/pkg/criteria/errf"
	"hcm/pkg/logs"
	"hcm/pkg/rest"
	"hcm/pkg/runtime/filter"
)

// AzureCreateDisk ...
func AzureCreateDisk(da *diskAdaptor, cts *rest.Contexts) (interface{}, error) {
	req := new(proto.AzureDiskCreateReq)
	if err := cts.DecodeInto(req); err != nil {
		return nil, errf.NewFromErr(errf.DecodeRequestFailed, err)
	}
	if err := req.Validate(); err != nil {
		return nil, errf.NewFromErr(errf.InvalidParameter, err)
	}

	client, err := da.adaptor.Azure(cts.Kit, req.Base.AccountID)
	if err != nil {
		return nil, err
	}

	diskSize := int32(req.Base.DiskSize)
	opt := &disk.AzureDiskCreateOption{
		Name:              req.Base.Name,
		ResourceGroupName: req.Extension.ResourceGroupName,
		Region:            &req.Base.Region,
		Zone:              &req.Base.Zone,
		DiskType:          req.Base.DiskType,
		DiskSize:          &diskSize,
	}
	client.CreateDisk(cts.Kit, opt)

	// TODO save to data-service

	return nil, nil
}

// AzureSyncDisk...
func AzureSyncDisk(da *diskAdaptor, cts *rest.Contexts) (interface{}, error) {

	req, err := da.decodeDiskSyncReq(cts)
	if err != nil {
		logs.Errorf("request decodeDiskSyncReq failed, err: %v, rid: %s", err, cts.Kit.Rid)
		return nil, err
	}

	cloudMap, err := da.getDatasFromAzureForDiskSync(cts, req)
	if err != nil {
		logs.Errorf("request getDatasFromAzureForDiskSync failed, err: %v, rid: %s", err, cts.Kit.Rid)
		return nil, err
	}

	dsMap, err := da.getDatasFromAzureDSForDiskSync(cts, req)
	if err != nil {
		logs.Errorf("request getDatasFromAzureDSForDiskSync failed, err: %v, rid: %s", err, cts.Kit.Rid)
		return nil, err
	}

	err = da.diffAzureDiskSync(cts, cloudMap, dsMap, req)
	if err != nil {
		logs.Errorf("request diffAzureDiskSync failed, err: %v, rid: %s", err, cts.Kit.Rid)
		return nil, err
	}

	return nil, nil
}

// getDatasFromAzureDSForDiskSync get azure datas from data-service
func (da *diskAdaptor) getDatasFromAzureDSForDiskSync(cts *rest.Contexts,
	req *protodisk.DiskSyncReq) (map[string]*protodisk.DiskSyncDS, error) {

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

		results, err := da.dataCli.Global.ListDisk(cts.Kit.Ctx, cts.Kit.Header(), dataReq)
		if err != nil {
			logs.Errorf("from data-service list disk failed, err: %v, rid: %s", err, cts.Kit.Rid)
			return nil, err
		}

		resultsHcm = append(resultsHcm, results.Details...)
		start += len(results.Details)
		if uint(len(results.Details)) < core.DefaultMaxPageLimit {
			break
		}
	}

	dsMap := make(map[string]*protodisk.DiskSyncDS)
	for _, result := range resultsHcm {
		sg := new(protodisk.DiskSyncDS)
		sg.IsUpdated = false
		sg.HcDisk = result
		dsMap[result.CloudID] = sg
	}

	return dsMap, nil
}

// getDatasFromAwsForDiskSync get datas from cloud
func (da *diskAdaptor) getDatasFromAzureForDiskSync(cts *rest.Contexts,
	req *protodisk.DiskSyncReq) (map[string]*proto.AzureDiskSyncDiff, error) {

	client, err := da.adaptor.Azure(cts.Kit, req.AccountID)
	if err != nil {
		return nil, err
	}

	listOpt := &disk.AzureDiskListOption{
		ResourceGroupName: req.ResourceGroupName,
	}
	datas, err := client.ListDisk(cts.Kit, listOpt)
	if err != nil {
		logs.Errorf("request adaptor to list azure disk failed, err: %v, rid: %s", err, cts.Kit.Rid)
		return nil, err
	}

	cloudMap := make(map[string]*proto.AzureDiskSyncDiff)
	for _, data := range datas {
		sg := new(proto.AzureDiskSyncDiff)
		sg.Disk = data
		cloudMap[*data.ID] = sg
	}

	return cloudMap, nil
}

// diffAzureDiskSync diff cloud data-service
func (da *diskAdaptor) diffAzureDiskSync(cts *rest.Contexts, cloudMap map[string]*proto.AzureDiskSyncDiff,
	dsMap map[string]*protodisk.DiskSyncDS, req *proto.DiskSyncReq) error {

	addCloudIDs := getAddCloudIDs(cloudMap, dsMap)
	deleteCloudIDs, updateCloudIDs := getDeleteAndUpdateCloudIDs(dsMap)

	if len(deleteCloudIDs) > 0 {
		err := da.diffDiskSyncDelete(cts, deleteCloudIDs)
		if err != nil {
			logs.Errorf("request diffDiskSyncDeletek failed, err: %v, rid: %s", err, cts.Kit.Rid)
			return err
		}
	}

	if len(updateCloudIDs) > 0 {
		err := da.diffAzureSyncUpdate(cts, cloudMap, dsMap, updateCloudIDs)
		if err != nil {
			logs.Errorf("request diffAzureSyncUpdate failed, err: %v, rid: %s", err, cts.Kit.Rid)
			return err
		}
	}

	if len(addCloudIDs) > 0 {
		_, err := da.diffAzureDiskSyncAdd(cts, cloudMap, req, addCloudIDs)
		if err != nil {
			logs.Errorf("request diffAzureDiskSyncAdd failed, err: %v, rid: %s", err, cts.Kit.Rid)
			return err
		}
	}

	return nil
}

// diffAzureDiskSyncAdd for add
func (da *diskAdaptor) diffAzureDiskSyncAdd(cts *rest.Contexts, cloudMap map[string]*proto.AzureDiskSyncDiff,
	req *proto.DiskSyncReq, addCloudIDs []string) ([]string, error) {

	var createReq dataproto.DiskExtBatchCreateReq[dataproto.AzureDiskExtensionCreateReq]

	for _, id := range addCloudIDs {
		disk := &dataproto.DiskExtCreateReq[dataproto.AzureDiskExtensionCreateReq]{
			AccountID:  req.AccountID,
			Name:       *cloudMap[id].Disk.Name,
			CloudID:    id,
			Region:     req.Region,
			Zone:       *cloudMap[id].Disk.Location,
			DiskSize:   uint64(*cloudMap[id].Disk.Properties.DiskSizeBytes),
			DiskType:   *cloudMap[id].Disk.Type,
			DiskStatus: string(*cloudMap[id].Disk.Properties.DiskState),
			Extension: &dataproto.AzureDiskExtensionCreateReq{
				ResourceGroupName: req.ResourceGroupName,
			},
		}
		createReq = append(createReq, disk)
	}

	if len(createReq) <= 0 {
		return make([]string, 0), nil
	}

	results, err := da.dataCli.Azure.BatchCreateDisk(cts.Kit.Ctx, cts.Kit.Header(), &createReq)
	if err != nil {
		logs.Errorf("request dataservice to create azure disk failed, err: %v, rid: %s", err, cts.Kit.Rid)
		return nil, err
	}

	return results.IDs, nil
}

// diffAzureSyncUpdate for update
func (da *diskAdaptor) diffAzureSyncUpdate(cts *rest.Contexts, cloudMap map[string]*proto.AzureDiskSyncDiff,
	dsMap map[string]*protodisk.DiskSyncDS, updateCloudIDs []string) error {

	var updateReq dataproto.DiskExtBatchUpadteReq[dataproto.AzureDiskExtensionUpdateReq]

	for _, id := range updateCloudIDs {
		if string(*cloudMap[id].Disk.Properties.DiskState) == dsMap[id].HcDisk.DiskStatus {
			continue
		}

		disk := &dataproto.DiskExtUpdateReq[dataproto.AzureDiskExtensionUpdateReq]{
			ID:         dsMap[id].HcDisk.ID,
			DiskStatus: string(*cloudMap[id].Disk.Properties.DiskState),
		}
		updateReq = append(updateReq, disk)
	}

	if len(updateReq) > 0 {
		if _, err := da.dataCli.Azure.BatchUpdateDisk(cts.Kit.Ctx, cts.Kit.Header(), &updateReq); err != nil {
			logs.Errorf("request dataservice azure BatchUpdateDisk failed, err: %v, rid: %s", err, cts.Kit.Rid)
			return err
		}
	}

	return nil
}
