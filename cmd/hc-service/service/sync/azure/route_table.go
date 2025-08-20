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
	routetable "hcm/pkg/adaptor/types/route-table"
	"hcm/pkg/api/hc-service/sync"
	"hcm/pkg/criteria/constant"
	"hcm/pkg/criteria/enumor"
	"hcm/pkg/kit"
	"hcm/pkg/logs"
	"hcm/pkg/rest"
	"hcm/pkg/tools/slice"

	"github.com/Azure/azure-sdk-for-go/sdk/resourcemanager/network/armnetwork/v2"
)

// SyncRouteTable ....
func (svc *service) SyncRouteTable(cts *rest.Contexts) (interface{}, error) {
	return nil, handler.ResourceSync(cts, &routeTableHandler{cli: svc.syncCli})
}

// routeTableHandler routeTable sync handler.
type routeTableHandler struct {
	cli ressync.Interface

	// Perpare 构建参数
	request *sync.AzureSyncReq
	syncCli azure.Interface
	pager   *adazure.Pager[armnetwork.RouteTablesClientListResponse, routetable.AzureRouteTable]
}

var _ handler.Handler = new(routeTableHandler)

// Prepare ...
func (hd *routeTableHandler) Prepare(cts *rest.Contexts) error {
	request, syncCli, err := defaultPrepare(cts, hd.cli)
	if err != nil {
		return err
	}

	hd.request = request
	hd.syncCli = syncCli

	listOpt := &typecore.AzureListOption{
		ResourceGroupName: hd.request.ResourceGroupName,
	}
	pager, err := hd.syncCli.CloudCli().ListRouteTableByPage(cts.Kit, listOpt)
	if err != nil {
		logs.Errorf("list routeTable by page failed, err: %v, rid: %s", err, cts.Kit.Rid)
		return err
	}

	hd.pager = pager

	return nil
}

// Next ...
func (hd *routeTableHandler) Next(kt *kit.Kit) ([]string, error) {
	if !hd.pager.More() {
		return nil, nil
	}

	total := make([]routetable.AzureRouteTable, 0)
	for hd.pager.More() && len(total) < constant.CloudResourceSyncMaxLimit {
		result, err := hd.pager.NextPage(kt)
		if err != nil {
			logs.Errorf("list routeTable next page failed, err: %v, rid: %s", err, kt.Rid)
			return nil, fmt.Errorf("list routeTable next page failed, err: %v", err)
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
func (hd *routeTableHandler) Sync(kt *kit.Kit, cloudIDs []string) error {
	cloudIDElems := slice.Split(cloudIDs, constant.CloudResourceSyncMaxLimit)

	for _, partCloudIDs := range cloudIDElems {
		params := &azure.SyncBaseParams{
			AccountID:         hd.request.AccountID,
			ResourceGroupName: hd.request.ResourceGroupName,
			CloudIDs:          partCloudIDs,
		}
		if _, err := hd.syncCli.RouteTable(kt, params, new(azure.SyncRouteTableOption)); err != nil {
			logs.Errorf("sync azure routeTable failed, err: %v, opt: %v, rid: %s", err, params, kt.Rid)
			return err
		}
	}

	return nil
}

// RemoveDeleteFromCloud ...
func (hd *routeTableHandler) RemoveDeleteFromCloud(kt *kit.Kit) error {
	err := hd.syncCli.RemoveRouteTableDeleteFromCloud(kt, hd.request.AccountID, hd.request.ResourceGroupName)
	if err != nil {
		logs.Errorf("remove routeTable delete from cloud failed, err: %v, accountID: %s, resGroupName: %s, rid: %s",
			err, hd.request.AccountID, hd.request.ResourceGroupName, kt.Rid)
		return err
	}

	return nil
}

// Name ...
func (hd *routeTableHandler) Name() enumor.CloudResourceType {
	return enumor.RouteTableCloudResType
}
