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

package huawei

import (
	ressync "hcm/cmd/hc-service/logics/res-sync"
	"hcm/cmd/hc-service/logics/res-sync/huawei"
	"hcm/cmd/hc-service/service/sync/handler"
	"hcm/pkg/adaptor/types/core"
	"hcm/pkg/adaptor/types/disk"
	"hcm/pkg/api/hc-service/sync"
	"hcm/pkg/criteria/constant"
	"hcm/pkg/criteria/enumor"
	"hcm/pkg/kit"
	"hcm/pkg/logs"
	"hcm/pkg/rest"
	"hcm/pkg/tools/converter"
)

// SyncDisk ....
func (svc *service) SyncDisk(cts *rest.Contexts) (interface{}, error) {
	return nil, handler.ResourceSync(cts, &diskHandler{cli: svc.syncCli})
}

// diskHandler disk sync handler.
type diskHandler struct {
	cli ressync.Interface

	// Prepare 构建参数
	request *sync.HuaWeiSyncReq
	syncCli huawei.Interface
	// marker 分页查询起始的资源ID，为空时查询第一页
	marker *string
}

var _ handler.Handler = new(diskHandler)

// Prepare ...
func (hd *diskHandler) Prepare(cts *rest.Contexts) error {
	request, syncCli, err := defaultPrepare(cts, hd.cli)
	if err != nil {
		return err
	}

	hd.request = request
	hd.syncCli = syncCli

	return nil
}

// Next ...
func (hd *diskHandler) Next(kt *kit.Kit) ([]string, error) {
	listOpt := &disk.HuaWeiDiskListOption{
		Region: hd.request.Region,
		Page: &core.HuaWeiPage{
			Limit:  converter.ValToPtr(int32(constant.CloudResourceSyncMaxLimit)),
			Marker: hd.marker,
		},
	}

	diskResult, err := hd.syncCli.CloudCli().ListDisk(kt, listOpt)
	if err != nil {
		logs.Errorf("request adaptor list huawei disk failed, err: %v, opt: %v, rid: %s", err, listOpt, kt.Rid)
		return nil, err
	}

	if len(diskResult) == 0 {
		return nil, nil
	}

	cloudIDs := make([]string, 0, len(diskResult))
	for _, one := range diskResult {
		cloudIDs = append(cloudIDs, one.Id)
	}

	hd.marker = converter.ValToPtr(diskResult[len(diskResult)-1].Id)
	return cloudIDs, nil
}

// Sync ...
func (hd *diskHandler) Sync(kt *kit.Kit, cloudIDs []string) error {
	params := &huawei.SyncBaseParams{
		AccountID: hd.request.AccountID,
		Region:    hd.request.Region,
		CloudIDs:  cloudIDs,
	}
	if _, err := hd.syncCli.Disk(kt, params, new(huawei.SyncDiskOption)); err != nil {
		logs.Errorf("sync huawei disk failed, err: %v, opt: %v, rid: %s", err, params, kt.Rid)
		return err
	}

	return nil
}

// RemoveDeleteFromCloud ...
func (hd *diskHandler) RemoveDeleteFromCloud(kt *kit.Kit) error {
	if err := hd.syncCli.RemoveDiskDeleteFromCloud(kt, hd.request.AccountID, hd.request.Region); err != nil {
		logs.Errorf("remove disk delete from cloud failed, err: %v, accountID: %s, region: %s, rid: %s", err,
			hd.request.AccountID, hd.request.Region, kt.Rid)
		return err
	}

	return nil
}

// Name ...
func (hd *diskHandler) Name() enumor.CloudResourceType {
	return enumor.DiskCloudResType
}
