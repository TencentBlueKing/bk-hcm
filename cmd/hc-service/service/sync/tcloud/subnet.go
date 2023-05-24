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

package tcloud

import (
	ressync "hcm/cmd/hc-service/logics/res-sync"
	"hcm/cmd/hc-service/logics/res-sync/tcloud"
	"hcm/cmd/hc-service/service/sync/handler"
	typecore "hcm/pkg/adaptor/types/core"
	"hcm/pkg/api/hc-service/sync"
	"hcm/pkg/criteria/constant"
	"hcm/pkg/criteria/enumor"
	"hcm/pkg/kit"
	"hcm/pkg/logs"
	"hcm/pkg/rest"
)

// SyncSubnet ....
func (svc *service) SyncSubnet(cts *rest.Contexts) (interface{}, error) {
	return nil, handler.ResourceSync(cts, &subnetHandler{cli: svc.syncCli})
}

// subnetHandler subnet sync handler.
type subnetHandler struct {
	cli ressync.Interface

	// Prepare 构建参数
	request *sync.TCloudSyncReq
	syncCli tcloud.Interface
	offset  uint64
}

var _ handler.Handler = new(subnetHandler)

// Prepare ...
func (hd *subnetHandler) Prepare(cts *rest.Contexts) error {
	request, syncCli, err := defaultPrepare(cts, hd.cli)
	if err != nil {
		return err
	}

	hd.request = request
	hd.syncCli = syncCli

	return nil
}

// Next ...
func (hd *subnetHandler) Next(kt *kit.Kit) ([]string, error) {
	listOpt := &typecore.TCloudListOption{
		Region: hd.request.Region,
		Page: &typecore.TCloudPage{
			Offset: hd.offset,
			Limit:  constant.CloudResourceSyncMaxLimit,
		},
	}

	subnetResult, err := hd.syncCli.CloudCli().ListSubnet(kt, listOpt)
	if err != nil {
		logs.Errorf("request adaptor list tcloud subnet failed, err: %v, opt: %v, rid: %s", err, listOpt, kt.Rid)
		return nil, err
	}

	if len(subnetResult.Details) == 0 {
		return nil, nil
	}

	cloudIDs := make([]string, 0, len(subnetResult.Details))
	for _, one := range subnetResult.Details {
		cloudIDs = append(cloudIDs, one.CloudID)
	}

	hd.offset += constant.CloudResourceSyncMaxLimit
	return cloudIDs, nil
}

// Sync ...
func (hd *subnetHandler) Sync(kt *kit.Kit, cloudIDs []string) error {
	params := &tcloud.SyncBaseParams{
		AccountID: hd.request.AccountID,
		Region:    hd.request.Region,
		CloudIDs:  cloudIDs,
	}
	if _, err := hd.syncCli.Subnet(kt, params, new(tcloud.SyncSubnetOption)); err != nil {
		logs.Errorf("sync tcloud subnet failed, err: %v, opt: %v, rid: %s", err, params, kt.Rid)
		return err
	}

	return nil
}

// RemoveDeleteFromCloud ...
func (hd *subnetHandler) RemoveDeleteFromCloud(kt *kit.Kit) error {
	if err := hd.syncCli.RemoveSubnetDeleteFromCloud(kt, hd.request.AccountID, hd.request.Region); err != nil {
		logs.Errorf("remove subnet delete from cloud failed, err: %v, accountID: %s, region: %s, rid: %s", err,
			hd.request.AccountID, hd.request.Region, kt.Rid)
		return err
	}

	return nil
}

// Name ...
func (hd *subnetHandler) Name() enumor.CloudResourceType {
	return enumor.SubnetCloudResType
}
