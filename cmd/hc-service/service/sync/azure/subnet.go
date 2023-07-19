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

package azure

import (
	"fmt"

	ressync "hcm/cmd/hc-service/logics/res-sync"
	"hcm/cmd/hc-service/logics/res-sync/azure"
	"hcm/cmd/hc-service/service/sync/handler"
	adazure "hcm/pkg/adaptor/azure"
	typecore "hcm/pkg/adaptor/types/core"
	"hcm/pkg/adaptor/types/subnet"
	"hcm/pkg/api/hc-service/sync"
	"hcm/pkg/criteria/constant"
	"hcm/pkg/criteria/enumor"
	"hcm/pkg/criteria/errf"
	"hcm/pkg/kit"
	"hcm/pkg/logs"
	"hcm/pkg/rest"
	"hcm/pkg/tools/slice"

	"github.com/Azure/azure-sdk-for-go/sdk/resourcemanager/network/armnetwork/v2"
)

// SyncSubnet ....
func (svc *service) SyncSubnet(cts *rest.Contexts) (interface{}, error) {
	return nil, handler.ResourceSync(cts, &subnetHandler{cli: svc.syncCli})
}

// subnetHandler subnet sync handler.
type subnetHandler struct {
	cli ressync.Interface

	// Prepare 构建参数
	request *sync.AzureSubnetSyncReq
	syncCli azure.Interface
	pager   *adazure.Pager[armnetwork.SubnetsClientListResponse, adtysubnet.AzureSubnet]
}

var _ handler.Handler = new(subnetHandler)

// Prepare ...
func (hd *subnetHandler) Prepare(cts *rest.Contexts) error {
	request := new(sync.AzureSubnetSyncReq)
	if err := cts.DecodeInto(request); err != nil {
		return errf.NewFromErr(errf.DecodeRequestFailed, err)
	}

	if err := request.Validate(); err != nil {
		return errf.NewFromErr(errf.InvalidParameter, err)
	}

	syncCli, err := hd.cli.Azure(cts.Kit, request.AccountID)
	if err != nil {
		return err
	}

	hd.request = request
	hd.syncCli = syncCli

	listOpt := &adtysubnet.AzureSubnetListOption{
		AzureListOption: typecore.AzureListOption{
			ResourceGroupName: hd.request.ResourceGroupName,
		},
		CloudVpcID: hd.request.CloudVpcID,
	}
	pager, err := hd.syncCli.CloudCli().ListSubnetByPage(cts.Kit, listOpt)
	if err != nil {
		logs.Errorf("list subnet by page failed, err: %v, rid: %s", err, cts.Kit.Rid)
		return err
	}

	hd.pager = pager

	return nil
}

// Next ...
func (hd *subnetHandler) Next(kt *kit.Kit) ([]string, error) {
	if !hd.pager.More() {
		return nil, nil
	}

	total := make([]adtysubnet.AzureSubnet, 0)
	for hd.pager.More() && len(total) < constant.CloudResourceSyncMaxLimit {
		result, err := hd.pager.NextPage(kt)
		if err != nil {
			logs.Errorf("list subnet next page failed, err: %v, rid: %s", err, kt.Rid)
			return nil, fmt.Errorf("list subnet next page failed, err: %v", err)
		}

		total = append(total, result...)
	}

	cloudIDs := make([]string, 0, len(total))
	for _, one := range total {
		cloudIDs = append(cloudIDs, one.CloudID)
	}

	return cloudIDs, nil
}

// Sync ...
func (hd *subnetHandler) Sync(kt *kit.Kit, cloudIDs []string) error {
	cloudIDElems := slice.Split(cloudIDs, constant.CloudResourceSyncMaxLimit)

	for _, partCloudIDs := range cloudIDElems {
		params := &azure.SyncBaseParams{
			AccountID:         hd.request.AccountID,
			ResourceGroupName: hd.request.ResourceGroupName,
			CloudIDs:          partCloudIDs,
		}
		opt := &azure.SyncSubnetOption{
			CloudVpcID: hd.request.CloudVpcID,
		}
		if _, err := hd.syncCli.Subnet(kt, params, opt); err != nil {
			logs.Errorf("sync azure subnet failed, err: %v, opt: %v, rid: %s", err, params, kt.Rid)
			return err
		}
	}

	return nil
}

// RemoveDeleteFromCloud ...
func (hd *subnetHandler) RemoveDeleteFromCloud(kt *kit.Kit) error {
	if err := hd.syncCli.RemoveSubnetDeleteFromCloud(kt, hd.request.AccountID, hd.request.ResourceGroupName,
		hd.request.CloudVpcID); err != nil {

		logs.Errorf("remove subnet delete from cloud failed, err: %v, accountID: %s, resGroupName: %s, rid: %s", err,
			hd.request.AccountID, hd.request.ResourceGroupName, kt.Rid)
		return err
	}

	return nil
}

// Name ...
func (hd *subnetHandler) Name() enumor.CloudResourceType {
	return enumor.SubnetCloudResType
}
