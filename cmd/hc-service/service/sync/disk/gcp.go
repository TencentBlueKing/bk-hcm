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
	disk "hcm/cmd/hc-service/logics/sync/disk"
	"hcm/pkg/adaptor/gcp"
	typcore "hcm/pkg/adaptor/types/core"
	typesdisk "hcm/pkg/adaptor/types/disk"
	protodisk "hcm/pkg/api/hc-service/disk"
	"hcm/pkg/criteria/errf"
	"hcm/pkg/logs"
	"hcm/pkg/rest"
	"hcm/pkg/runtime/filter"
)

// SyncGcpDisk ...
func (svc *syncDiskSvc) SyncGcpDisk(cts *rest.Contexts) (interface{}, error) {
	req := new(disk.SyncGcpDiskOption)
	if err := cts.DecodeInto(req); err != nil {
		return nil, errf.NewFromErr(errf.DecodeRequestFailed, err)
	}

	if err := req.Validate(); err != nil {
		return nil, errf.NewFromErr(errf.InvalidParameter, err)
	}

	client, err := svc.adaptor.Gcp(cts.Kit, req.AccountID)
	if err != nil {
		return nil, err
	}

	allCloudIDs := make(map[string]struct{})
	nextToken := ""
	for {
		listOpt := &typesdisk.GcpDiskListOption{
			Zone: req.Zone,
			Page: &typcore.GcpPage{
				PageToken: nextToken,
				PageSize:  int64(filter.DefaultMaxInLimit),
			},
		}

		datas, token, err := client.ListDisk(cts.Kit, listOpt)
		if err != nil {
			logs.Errorf("request adaptor to list gcp disk failed, err: %v, rid: %s", err, cts.Kit.Rid)
			return nil, err
		}

		cloudIDs := make([]string, 0, len(datas))
		for _, one := range datas {
			cloudIDs = append(cloudIDs, fmt.Sprint(one.Id))
			allCloudIDs[fmt.Sprint(one.Id)] = struct{}{}
		}

		if len(cloudIDs) > 0 {
			req.CloudIDs = cloudIDs
		}
		_, err = disk.SyncGcpDisk(cts.Kit, req, svc.adaptor, svc.dataCli)
		if err != nil {
			logs.Errorf("request to sync gcp disk failed, err: %v, rid: %s", err, cts.Kit.Rid)
			return nil, err
		}

		if len(token) == 0 {
			break
		}
		nextToken = token
	}

	commReq := &protodisk.DiskSyncReq{
		AccountID: req.AccountID,
		Zone:      req.Zone,
	}
	dsIDs, err := disk.GetDatasFromDSForGcpDiskSync(cts.Kit, commReq, svc.dataCli)
	if err != nil {
		logs.Errorf("request GetDatasFromDSForGcpDiskSync failed, err: %v, rid: %s", err, cts.Kit.Rid)
		return nil, err
	}

	deleteIDs := make([]string, 0)
	for id := range dsIDs {
		if _, ok := allCloudIDs[id]; !ok {
			deleteIDs = append(deleteIDs, id)
		}
	}

	err = svc.deleteGcpDisk(cts, client, req, deleteIDs)
	if err != nil {
		logs.Errorf("request deleteGcpDisk failed, err: %v, rid: %s", err, cts.Kit.Rid)
		return nil, err
	}

	return nil, nil
}

func (svc *syncDiskSvc) deleteGcpDisk(cts *rest.Contexts, client *gcp.Gcp,
	req *disk.SyncGcpDiskOption, deleteIDs []string) error {

	if len(deleteIDs) > 0 {
		realDeleteIDs := make([]string, 0)
		nextToken := ""
		for {
			listOpt := &typesdisk.GcpDiskListOption{
				Zone: req.Zone,
				Page: &typcore.GcpPage{
					PageToken: nextToken,
					PageSize:  int64(filter.DefaultMaxInLimit),
				},
			}

			datas, token, err := client.ListDisk(cts.Kit, listOpt)
			if err != nil {
				logs.Errorf("request adaptor to list gcp disk failed, err: %v, rid: %s", err, cts.Kit.Rid)
				return err
			}

			for _, id := range deleteIDs {
				realDeleteFlag := true
				for _, data := range datas {
					if fmt.Sprint(data.Id) == id {
						realDeleteFlag = false
						break
					}
				}

				if realDeleteFlag {
					realDeleteIDs = append(realDeleteIDs, id)
				}
			}

			if len(token) == 0 {
				break
			}
			nextToken = token
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
