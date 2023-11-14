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

package logicsrt

import (
	ressync "hcm/cmd/hc-service/logics/res-sync"
	"hcm/cmd/hc-service/logics/res-sync/huawei"
	"hcm/cmd/hc-service/service/sync/handler"
	"hcm/pkg/adaptor/types/core"
	routetable "hcm/pkg/adaptor/types/route-table"
	"hcm/pkg/api/hc-service/sync"
	"hcm/pkg/criteria/constant"
	"hcm/pkg/criteria/enumor"
	"hcm/pkg/criteria/errf"
	"hcm/pkg/kit"
	"hcm/pkg/logs"
	"hcm/pkg/rest"
	"hcm/pkg/tools/converter"
)

// HuaWeiRouteTableHandler routeTable sync handler.
type HuaWeiRouteTableHandler struct {
	Cli            ressync.Interface
	DisablePrepare bool

	// Perpare 构建参数
	Request *sync.HuaWeiSyncReq
	SyncCli huawei.Interface
	// Marker 取值为上一页数据的最后一条记录的id，为空时为查询第一页
	Marker *string
}

var _ handler.Handler = new(HuaWeiRouteTableHandler)

// Prepare ...
func (hd *HuaWeiRouteTableHandler) Prepare(cts *rest.Contexts) error {
	if hd.DisablePrepare {
		return nil
	}

	req := new(sync.HuaWeiSyncReq)
	if err := cts.DecodeInto(req); err != nil {
		return errf.NewFromErr(errf.DecodeRequestFailed, err)
	}

	if err := req.Validate(); err != nil {
		return errf.NewFromErr(errf.InvalidParameter, err)
	}

	syncCli, err := hd.Cli.HuaWei(cts.Kit, req.AccountID)
	if err != nil {
		return err
	}

	hd.Request = req
	hd.SyncCli = syncCli

	return nil
}

// Next ...
func (hd *HuaWeiRouteTableHandler) Next(kt *kit.Kit) ([]string, error) {
	listOpt := &routetable.HuaWeiRouteTableListOption{
		Region: hd.Request.Region,
		Page: &core.HuaWeiPage{
			Limit:  converter.ValToPtr(int32(constant.CloudResourceSyncMaxLimit)),
			Marker: hd.Marker,
		},
	}
	routeTableResult, err := hd.SyncCli.CloudCli().ListRouteTables(kt, listOpt)
	if err != nil {
		logs.Errorf("Request adaptor list huawei routeTable failed, err: %v, opt: %v, rid: %s", err, listOpt, kt.Rid)
		return nil, err
	}

	if len(routeTableResult) == 0 {
		return nil, nil
	}

	cloudIDs := make([]string, 0, len(routeTableResult))
	for _, one := range routeTableResult {
		cloudIDs = append(cloudIDs, one.CloudID)
	}

	hd.Marker = converter.ValToPtr(routeTableResult[len(routeTableResult)-1].CloudID)
	return cloudIDs, nil
}

// Sync ...
func (hd *HuaWeiRouteTableHandler) Sync(kt *kit.Kit, cloudIDs []string) error {
	params := &huawei.SyncBaseParams{
		AccountID: hd.Request.AccountID,
		Region:    hd.Request.Region,
		CloudIDs:  cloudIDs,
	}
	if _, err := hd.SyncCli.RouteTable(kt, params, new(huawei.SyncRouteTableOption)); err != nil {
		logs.Errorf("sync huawei routeTable failed, err: %v, opt: %v, rid: %s", err, params, kt.Rid)
		return err
	}

	return nil
}

// RemoveDeleteFromCloud ...
func (hd *HuaWeiRouteTableHandler) RemoveDeleteFromCloud(kt *kit.Kit) error {
	if err := hd.SyncCli.RemoveRouteTableDeleteFromCloud(kt, hd.Request.AccountID, hd.Request.Region); err != nil {
		logs.Errorf("remove routeTable delete from cloud failed, err: %v, accountID: %s, region: %s, rid: %s", err,
			hd.Request.AccountID, hd.Request.Region, kt.Rid)
		return err
	}

	return nil
}

// Name ...
func (hd *HuaWeiRouteTableHandler) Name() enumor.CloudResourceType {
	return enumor.RouteTableCloudResType
}
