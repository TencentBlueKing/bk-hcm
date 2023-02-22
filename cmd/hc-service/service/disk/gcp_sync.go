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

	"hcm/pkg/adaptor/types/disk"
	dataproto "hcm/pkg/api/data-service/cloud/disk"
	proto "hcm/pkg/api/hc-service/disk"
	protodisk "hcm/pkg/api/hc-service/disk"
	"hcm/pkg/logs"
	"hcm/pkg/rest"
)

// GcpSyncDisk...
func GcpSyncDisk(da *diskAdaptor, cts *rest.Contexts) (interface{}, error) {
	req, err := da.decodeDiskSyncReq(cts)
	if err != nil {
		logs.Errorf("request decodeDiskSyncReq failed, err: %v, rid: %s", err, cts.Kit.Rid)
		return nil, err
	}

	cloudMap, err := da.getDatasFromGcpForDiskSync(cts, req)
	if err != nil {
		logs.Errorf("request getDatasFromGcpForDiskSync failed, err: %v, rid: %s", err, cts.Kit.Rid)
		return nil, err
	}

	dsMap, err := da.getDatasFromDSForDiskSync(cts, req)
	if err != nil {
		logs.Errorf("request getDatasFromDSForDiskSync failed, err: %v, rid: %s", err, cts.Kit.Rid)
		return nil, err
	}

	err = da.diffGcpDiskSync(cts, cloudMap, dsMap, req)
	if err != nil {
		logs.Errorf("request diffGcpDiskSync failed, err: %v, rid: %s", err, cts.Kit.Rid)
		return nil, err
	}

	return nil, nil
}

// getDatasFromGcpForDiskSync get datas from cloud
func (da *diskAdaptor) getDatasFromGcpForDiskSync(cts *rest.Contexts,
	req *protodisk.DiskSyncReq,
) (map[string]*proto.GcpDiskSyncDiff, error) {
	client, err := da.adaptor.Gcp(cts.Kit, req.AccountID)
	if err != nil {
		return nil, err
	}

	listOpt := &disk.GcpDiskListOption{
		Zone: req.Zone,
	}
	datas, err := client.ListDisk(cts.Kit, listOpt)
	if err != nil {
		logs.Errorf("request adaptor to list gcp disk failed, err: %v, rid: %s", err, cts.Kit.Rid)
		return nil, err
	}

	cloudMap := make(map[string]*proto.GcpDiskSyncDiff)
	for _, data := range datas {
		sg := new(proto.GcpDiskSyncDiff)
		sg.Disk = data
		cloudMap[fmt.Sprint(data.Id)] = sg
	}

	return cloudMap, nil
}

// diffGcpDiskSync diff gcp data-service
func (da *diskAdaptor) diffGcpDiskSync(cts *rest.Contexts, cloudMap map[string]*proto.GcpDiskSyncDiff,
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
		err := da.diffGcpSyncUpdate(cts, cloudMap, dsMap, updateCloudIDs)
		if err != nil {
			logs.Errorf("request diffGcpSyncUpdate failed, err: %v, rid: %s", err, cts.Kit.Rid)
			return err
		}
	}

	if len(addCloudIDs) > 0 {
		_, err := da.diffGcpDiskSyncAdd(cts, cloudMap, req, addCloudIDs)
		if err != nil {
			logs.Errorf("request diffGcpDiskSyncAdd failed, err: %v, rid: %s", err, cts.Kit.Rid)
			return err
		}
	}

	return nil
}

// diffGcpDiskSyncAdd for add
func (da *diskAdaptor) diffGcpDiskSyncAdd(cts *rest.Contexts, cloudMap map[string]*proto.GcpDiskSyncDiff,
	req *proto.DiskSyncReq, addCloudIDs []string,
) ([]string, error) {
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

	results, err := da.dataCli.Gcp.BatchCreateDisk(cts.Kit.Ctx, cts.Kit.Header(), &createReq)
	if err != nil {
		logs.Errorf("request dataservice to create gcp disk failed, err: %v, rid: %s", err, cts.Kit.Rid)
		return nil, err
	}

	return results.IDs, nil
}

// diffGcpSyncUpdate for update
func (da *diskAdaptor) diffGcpSyncUpdate(cts *rest.Contexts, cloudMap map[string]*proto.GcpDiskSyncDiff,
	dsMap map[string]*protodisk.DiskSyncDS, updateCloudIDs []string,
) error {
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
		if _, err := da.dataCli.Aws.BatchUpdateDisk(cts.Kit.Ctx, cts.Kit.Header(), &updateReq); err != nil {
			logs.Errorf("request dataservice gcp BatchUpdateDisk failed, err: %v, rid: %s", err, cts.Kit.Rid)
			return err
		}
	}

	return nil
}
