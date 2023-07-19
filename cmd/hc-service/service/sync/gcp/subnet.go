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

package gcp

import (
	ressync "hcm/cmd/hc-service/logics/res-sync"
	"hcm/cmd/hc-service/logics/res-sync/gcp"
	"hcm/cmd/hc-service/service/sync/handler"
	typecore "hcm/pkg/adaptor/types/core"
	"hcm/pkg/adaptor/types/subnet"
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
	request   *sync.GcpSyncReq
	syncCli   gcp.Interface
	pageToken string
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
	listOpt := &adtysubnet.GcpSubnetListOption{
		Region: hd.request.Region,
		GcpListOption: typecore.GcpListOption{
			Page: &typecore.GcpPage{
				PageSize:  constant.CloudResourceSyncMaxLimit,
				PageToken: hd.pageToken,
			},
		},
	}

	subnetResult, err := hd.syncCli.CloudCli().ListSubnet(kt, listOpt)
	if err != nil {
		logs.Errorf("request adaptor list gcp subnet failed, err: %v, opt: %v, rid: %s", err, listOpt, kt.Rid)
		return nil, err
	}

	if len(subnetResult.Details) == 0 {
		return nil, nil
	}

	cloudIDs := make([]string, 0, len(subnetResult.Details))
	for _, one := range subnetResult.Details {
		cloudIDs = append(cloudIDs, one.CloudID)
	}

	hd.pageToken = subnetResult.NextPageToken
	return cloudIDs, nil
}

// Sync ...
func (hd *subnetHandler) Sync(kt *kit.Kit, cloudIDs []string) error {
	params := &gcp.SyncBaseParams{
		AccountID: hd.request.AccountID,
		CloudIDs:  cloudIDs,
	}
	opt := &gcp.SyncSubnetOption{
		Region: hd.request.Region,
	}
	if _, err := hd.syncCli.Subnet(kt, params, opt); err != nil {
		logs.Errorf("sync gcp subnet failed, err: %v, opt: %v, rid: %s", err, params, kt.Rid)
		return err
	}

	return nil
}

// RemoveDeleteFromCloud ...
func (hd *subnetHandler) RemoveDeleteFromCloud(kt *kit.Kit) error {
	if err := hd.syncCli.RemoveSubnetDeleteFromCloud(kt, hd.request.AccountID, hd.request.Region); err != nil {
		logs.Errorf("remove subnet delete from cloud failed, err: %v, accountID: %s, rid: %s", err,
			hd.request.AccountID, kt.Rid)
		return err
	}

	return nil
}

// Name ...
func (hd *subnetHandler) Name() enumor.CloudResourceType {
	return enumor.SubnetCloudResType
}
