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
	"hcm/pkg/adaptor/aws"
	typcore "hcm/pkg/adaptor/types/core"
	typesdisk "hcm/pkg/adaptor/types/disk"
	protodisk "hcm/pkg/api/hc-service/disk"
	"hcm/pkg/api/hc-service/sync"
	"hcm/pkg/criteria/errf"
	"hcm/pkg/logs"
	"hcm/pkg/rest"
	"hcm/pkg/runtime/filter"
	"hcm/pkg/tools/converter"
)

// SyncAwsDisk ...
func (svc *syncDiskSvc) SyncAwsDisk(cts *rest.Contexts) (interface{}, error) {
	syncReq := new(sync.SyncAwsDiskReq)
	if err := cts.DecodeInto(syncReq); err != nil {
		return nil, errf.NewFromErr(errf.DecodeRequestFailed, err)
	}

	if err := syncReq.Validate(); err != nil {
		return nil, errf.NewFromErr(errf.InvalidParameter, err)
	}

	req := &disk.SyncAwsDiskOption{
		AccountID: syncReq.AccountID,
		Region:    syncReq.Region,
	}
	client, err := svc.adaptor.Aws(cts.Kit, req.AccountID)
	if err != nil {
		return nil, err
	}

	allCloudIDs := make(map[string]struct{})
	nextToken := ""
	for {
		listOpt := &typesdisk.AwsDiskListOption{
			Region: req.Region,
			Page: &typcore.AwsPage{
				MaxResults: converter.ValToPtr(int64(filter.DefaultMaxInLimit)),
			},
		}
		if nextToken != "" {
			listOpt.Page.NextToken = converter.ValToPtr(nextToken)
		}

		datas, token, err := client.ListDisk(cts.Kit, listOpt)
		if err != nil {
			logs.Errorf("request adaptor to list aws disk failed, err: %v, rid: %s", err, cts.Kit.Rid)
			return nil, err
		}

		cloudIDs := make([]string, 0, len(datas))
		for _, one := range datas {
			cloudIDs = append(cloudIDs, *one.VolumeId)
			allCloudIDs[*one.VolumeId] = struct{}{}
		}

		if len(cloudIDs) > 0 {
			req.CloudIDs = cloudIDs
		}
		_, err = disk.SyncAwsDisk(cts.Kit, req, svc.adaptor, svc.dataCli)
		if err != nil {
			logs.Errorf("request to sync aws disk failed, err: %v, rid: %s", err, cts.Kit.Rid)
			return nil, err
		}

		if token == nil {
			break
		}
		nextToken = *token
	}

	commReq := &protodisk.DiskSyncReq{
		AccountID: req.AccountID,
		Region:    req.Region,
	}
	dsIDs, err := disk.GetDatasFromDSForDiskSync(cts.Kit, commReq, svc.dataCli)
	if err != nil {
		logs.Errorf("request getTCloudEipAllDS failed, err: %v, rid: %s", err, cts.Kit.Rid)
		return nil, err
	}

	deleteIDs := make([]string, 0)
	for id := range dsIDs {
		if _, ok := allCloudIDs[id]; !ok {
			deleteIDs = append(deleteIDs, id)
		}
	}

	err = svc.deleteAwsDisk(cts, client, req, deleteIDs)
	if err != nil {
		logs.Errorf("request deleteAwsDisk failed, err: %v, rid: %s", err, cts.Kit.Rid)
		return nil, err
	}

	return nil, nil
}

func (svc *syncDiskSvc) deleteAwsDisk(cts *rest.Contexts, client *aws.Aws,
	req *disk.SyncAwsDiskOption, deleteIDs []string) error {

	if len(deleteIDs) > 0 {
		realDeleteIDs := make([]string, 0)
		nextToken := ""
		for {
			listOpt := &typesdisk.AwsDiskListOption{
				Region: req.Region,
				Page: &typcore.AwsPage{
					MaxResults: converter.ValToPtr(int64(filter.DefaultMaxInLimit)),
				},
			}
			if nextToken != "" {
				listOpt.Page.NextToken = converter.ValToPtr(nextToken)
			}

			datas, token, err := client.ListDisk(cts.Kit, listOpt)
			if err != nil {
				logs.Errorf("request adaptor to list aws disk failed, err: %v, rid: %s", err, cts.Kit.Rid)
				return err
			}

			for _, id := range deleteIDs {
				realDeleteFlag := true
				for _, data := range datas {
					if *data.VolumeId == id {
						realDeleteFlag = false
						break
					}
				}

				if realDeleteFlag {
					realDeleteIDs = append(realDeleteIDs, id)
				}
			}

			if token == nil {
				break
			}
			nextToken = *token
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
