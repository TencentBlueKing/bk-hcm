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
	"hcm/pkg/adaptor/huawei"
	"hcm/pkg/adaptor/types/core"
	typcore "hcm/pkg/adaptor/types/core"
	typesdisk "hcm/pkg/adaptor/types/disk"
	protodisk "hcm/pkg/api/hc-service/disk"
	"hcm/pkg/criteria/errf"
	"hcm/pkg/logs"
	"hcm/pkg/rest"
)

// SyncHuaWeiDisk ...
func (svc *syncDiskSvc) SyncHuaWeiDisk(cts *rest.Contexts) (interface{}, error) {
	req := new(disk.SyncHuaWeiDiskOption)
	if err := cts.DecodeInto(req); err != nil {
		return nil, errf.NewFromErr(errf.DecodeRequestFailed, err)
	}

	if err := req.Validate(); err != nil {
		return nil, errf.NewFromErr(errf.InvalidParameter, err)
	}

	client, err := svc.adaptor.HuaWei(cts.Kit, req.AccountID)
	if err != nil {
		return nil, err
	}

	allCloudIDs := make(map[string]struct{})
	limit := int32(typcore.HuaWeiQueryLimit)
	var marker *string = nil
	for {
		opt := &typesdisk.HuaWeiDiskListOption{
			Region: req.Region,
			Page:   &core.HuaWeiPage{Limit: &limit, Marker: marker},
		}

		datas, err := client.ListDisk(cts.Kit, opt)
		if err != nil {
			logs.Errorf("request adaptor to list huawei disk failed, err: %v, rid: %s", err, cts.Kit.Rid)
			return nil, err
		}

		cloudIDs := make([]string, 0, len(datas))
		for index, one := range datas {
			cloudIDs = append(cloudIDs, one.Id)
			allCloudIDs[one.Id] = struct{}{}
			if index == len(datas)-1 {
				marker = &one.Id
			}
		}

		if len(cloudIDs) > 0 {
			req.CloudIDs = cloudIDs
		}
		_, err = disk.SyncHuaWeiDisk(cts.Kit, req, svc.adaptor, svc.dataCli)
		if err != nil {
			logs.Errorf("request to sync huawei disk failed, err: %v, rid: %s", err, cts.Kit.Rid)
			return nil, err
		}

		if int32(len(datas)) < limit {
			break
		}
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

	err = svc.deleteHuaWeiDisk(cts, client, req, deleteIDs)
	if err != nil {
		logs.Errorf("request deleteHuaWeiDisk failed, err: %v, rid: %s", err, cts.Kit.Rid)
		return nil, err
	}

	return nil, nil
}

func (svc *syncDiskSvc) deleteHuaWeiDisk(cts *rest.Contexts, client *huawei.HuaWei,
	req *disk.SyncHuaWeiDiskOption, deleteIDs []string) error {

	if len(deleteIDs) > 0 {
		realDeleteIDs := make([]string, 0)

		limit := int32(typcore.HuaWeiQueryLimit)
		var marker *string = nil
		for {
			opt := &typesdisk.HuaWeiDiskListOption{
				Region: req.Region,
				Page:   &core.HuaWeiPage{Limit: &limit, Marker: marker},
			}

			datas, err := client.ListDisk(cts.Kit, opt)
			if err != nil {
				logs.Errorf("request adaptor to list huawei disk failed, err: %v, rid: %s", err, cts.Kit.Rid)
				return err
			}

			if len(datas) > 0 {
				marker = &datas[len(datas)-1].Id
			}

			for _, id := range deleteIDs {
				realDeleteFlag := true
				for _, data := range datas {
					if data.Id == id {
						realDeleteFlag = false
						break
					}
				}

				if realDeleteFlag {
					realDeleteIDs = append(realDeleteIDs, id)
				}
			}

			if int32(len(datas)) < limit {
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
