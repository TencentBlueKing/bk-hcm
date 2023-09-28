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
	routetable "hcm/pkg/adaptor/types/route-table"
	"hcm/pkg/api/hc-service/sync"
	"hcm/pkg/criteria/constant"
	"hcm/pkg/criteria/enumor"
	"hcm/pkg/criteria/errf"
	"hcm/pkg/kit"
	"hcm/pkg/logs"
	"hcm/pkg/rest"
)

// SyncRoute ....
func (svc *service) SyncRoute(cts *rest.Contexts) (interface{}, error) {
	return nil, handler.ResourceSync(cts, &routeHandler{cli: svc.syncCli})
}

// routeHandler route sync handler.
type routeHandler struct {
	cli ressync.Interface

	// Perpare 构建参数
	request   *sync.GcpRouteSyncReq
	syncCli   gcp.Interface
	pageToken string
}

var _ handler.Handler = new(routeHandler)

// Prepare ...
func (hd *routeHandler) Prepare(cts *rest.Contexts) error {
	req := new(sync.GcpRouteSyncReq)
	if err := cts.DecodeInto(req); err != nil {
		return errf.NewFromErr(errf.DecodeRequestFailed, err)
	}

	if err := req.Validate(); err != nil {
		return errf.NewFromErr(errf.InvalidParameter, err)
	}

	syncCli, err := hd.cli.Gcp(cts.Kit, req.AccountID)
	if err != nil {
		return err
	}

	hd.request = req
	hd.syncCli = syncCli

	return nil
}

// Next ...
func (hd *routeHandler) Next(kt *kit.Kit) ([]string, error) {
	listOpt := &routetable.GcpListOption{
		Page: &typecore.GcpPage{
			PageToken: hd.pageToken,
			PageSize:  constant.CloudResourceSyncMaxLimit,
		},
	}
	routeResult, err := hd.syncCli.CloudCli().ListRoute(kt, listOpt)
	if err != nil {
		logs.Errorf("request adaptor list gcp route failed, err: %v, opt: %v, rid: %s", err, listOpt, kt.Rid)
		return nil, err
	}

	if len(routeResult.Details) == 0 {
		return nil, nil
	}

	cloudIDs := make([]string, 0, len(routeResult.Details))
	for _, one := range routeResult.Details {
		cloudIDs = append(cloudIDs, one.CloudID)
	}

	hd.pageToken = routeResult.NextPageToken
	return cloudIDs, nil
}

// Sync ...
func (hd *routeHandler) Sync(kt *kit.Kit, cloudIDs []string) error {
	params := &gcp.SyncBaseParams{
		AccountID: hd.request.AccountID,
		CloudIDs:  cloudIDs,
	}
	if _, err := hd.syncCli.Route(kt, params, &gcp.SyncRouteOption{}); err != nil {
		logs.Errorf("sync gcp route failed, err: %v, opt: %v, rid: %s", err, params, kt.Rid)
		return err
	}

	return nil
}

// RemoveDeleteFromCloud ...
func (hd *routeHandler) RemoveDeleteFromCloud(kt *kit.Kit) error {
	if err := hd.syncCli.RemoveRouteDeleteFromCloud(kt, hd.request.AccountID); err != nil {
		logs.Errorf("remove route delete from cloud failed, err: %v, accountID: %s, rid: %s", err,
			hd.request.AccountID, kt.Rid)
		return err
	}

	return nil
}

// Name ...
func (hd *routeHandler) Name() enumor.CloudResourceType {
	return enumor.RouteCloudResType
}
