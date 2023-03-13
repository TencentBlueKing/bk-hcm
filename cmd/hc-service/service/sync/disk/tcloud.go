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
	disk "hcm/cmd/hc-service/logics/sync/disk"
	"hcm/pkg/adaptor/tcloud"
	"hcm/pkg/adaptor/types/core"
	typesdisk "hcm/pkg/adaptor/types/disk"
	protodisk "hcm/pkg/api/hc-service/disk"
	"hcm/pkg/api/hc-service/sync"
	"hcm/pkg/criteria/errf"
	"hcm/pkg/logs"
	"hcm/pkg/rest"
)

// SyncTCloudDisk ...
func (svc *syncDiskSvc) SyncTCloudDisk(cts *rest.Contexts) (interface{}, error) {
	req := new(sync.SyncTCloudDiskReq)
	if err := cts.DecodeInto(req); err != nil {
		return nil, errf.NewFromErr(errf.DecodeRequestFailed, err)
	}

	if err := req.Validate(); err != nil {
		return nil, errf.NewFromErr(errf.InvalidParameter, err)
	}

	client, err := svc.adaptor.TCloud(cts.Kit, req.AccountID)
	if err != nil {
		return nil, err
	}

	allCloudIDs := make(map[string]struct{})
	offset := 0
	for {
		opt := &typesdisk.TCloudDiskListOption{
			Region: req.Region,
			Page:   &core.TCloudPage{Offset: uint64(offset), Limit: uint64(core.TCloudQueryLimit)},
		}

		datas, err := client.ListDisk(cts.Kit, opt)
		if err != nil {
			logs.Errorf("request adaptor to list tcloud disk failed, err: %v, rid: %s", err, cts.Kit.Rid)
			return nil, err
		}

		cloudIDs := make([]string, 0, len(datas))
		for _, one := range datas {
			cloudIDs = append(cloudIDs, *one.DiskId)
			allCloudIDs[*one.DiskId] = struct{}{}
		}

		if len(cloudIDs) > 0 {
			req.CloudIDs = cloudIDs
		}
		_, err = disk.SyncTCloudDisk(cts.Kit, req, svc.adaptor, svc.dataCli)
		if err != nil {
			logs.Errorf("request to sync tcloud disk failed, err: %v, rid: %s", err, cts.Kit.Rid)
			return nil, err
		}

		offset += len(datas)
		if uint(len(datas)) < core.TCloudQueryLimit {
			break
		}
	}

	commReq := &protodisk.DiskSyncReq{
		AccountID: req.AccountID,
		Region:    req.Region,
	}
	dsIDs, err := disk.GetDatasFromDSForDiskSync(cts.Kit, commReq, svc.dataCli)
	if err != nil {
		logs.Errorf("request GetDatasFromDSForDiskSync failed, err: %v, rid: %s", err, cts.Kit.Rid)
		return nil, err
	}

	deleteIDs := make([]string, 0)
	for id := range dsIDs {
		if _, ok := allCloudIDs[id]; !ok {
			deleteIDs = append(deleteIDs, id)
		}
	}

	err = svc.deleteTCloudDisk(cts, client, req, deleteIDs)
	if err != nil {
		logs.Errorf("request deleteTCloudDisk failed, err: %v, rid: %s", err, cts.Kit.Rid)
		return nil, err
	}

	return nil, nil
}

func (svc *syncDiskSvc) deleteTCloudDisk(cts *rest.Contexts, client *tcloud.TCloud,
	req *sync.SyncTCloudDiskReq, deleteIDs []string) error {

	if len(deleteIDs) > 0 {
		realDeleteIDs := make([]string, 0)
		offset := 0
		for {
			opt := &typesdisk.TCloudDiskListOption{
				Region: req.Region,
				Page:   &core.TCloudPage{Offset: uint64(offset), Limit: uint64(core.TCloudQueryLimit)},
			}

			datas, err := client.ListDisk(cts.Kit, opt)
			if err != nil {
				logs.Errorf("request adaptor to list tcloud disk failed, err: %v, rid: %s", err, cts.Kit.Rid)
				return err
			}

			for _, id := range deleteIDs {
				realDeleteFlag := true
				for _, data := range datas {
					if *data.DiskId == id {
						realDeleteFlag = false
						break
					}
				}

				if realDeleteFlag {
					realDeleteIDs = append(realDeleteIDs, id)
				}
			}

			offset += len(datas)
			if uint(len(datas)) < core.TCloudQueryLimit {
				break
			}
		}

		if len(realDeleteIDs) > 0 {
			err := disk.DiffDiskSyncDelete(cts.Kit, realDeleteIDs, svc.dataCli)
			if err != nil {
				logs.Errorf("request diffDiskSyncDelete failed, err: %v, rid: %s", err, cts.Kit.Rid)
				return err
			}
		}
	}

	return nil
}
