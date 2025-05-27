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

package other

import (
	"fmt"

	ressync "hcm/cmd/hc-service/logics/res-sync"
	"hcm/cmd/hc-service/logics/res-sync/other"
	"hcm/cmd/hc-service/service/sync/handler"
	"hcm/pkg/api/core"
	"hcm/pkg/api/hc-service/sync"
	"hcm/pkg/cc"
	"hcm/pkg/criteria/constant"
	"hcm/pkg/criteria/enumor"
	"hcm/pkg/criteria/errf"
	"hcm/pkg/hooks"
	"hcm/pkg/kit"
	"hcm/pkg/logs"
	"hcm/pkg/rest"
	"hcm/pkg/thirdparty/api-gateway/cmdb"
	"hcm/pkg/tools/converter"
)

// SyncHostWithRelRes ....
func (svc *service) SyncHostWithRelRes(cts *rest.Contexts) (interface{}, error) {
	return nil, handler.ResourceSyncV2[cmdb.HostWithCloudID](cts, &hostHandler{cli: svc.syncCli})
}

// SyncHostWithRelResByCond ....
func (svc *service) SyncHostWithRelResByCond(cts *rest.Contexts) (interface{}, error) {
	req := new(sync.OtherSyncHostByCondReq)
	if err := cts.DecodeInto(req); err != nil {
		return nil, errf.NewFromErr(errf.DecodeRequestFailed, err)
	}

	if err := req.Validate(); err != nil {
		return nil, errf.NewFromErr(errf.InvalidParameter, err)
	}

	syncCli, err := svc.syncCli.Other(cts.Kit, req.AccountID)
	if err != nil {
		return nil, err
	}
	op := &hostHandler{syncCli: syncCli}
	params := &other.SyncHostParams{AccountID: req.AccountID, BizID: req.BizID, HostIDs: req.HostIDs}
	if err = op.SyncByCond(cts.Kit, params); err != nil {
		logs.Errorf("sync host by condition failed, err: %v, req: %+v, rid: %s", err, req, cts.Kit.Rid)
		return nil, err
	}

	return nil, nil
}

// DeleteHostByCond ....
func (svc *service) DeleteHostByCond(cts *rest.Contexts) (interface{}, error) {
	req := new(sync.OtherDelHostByCondReq)
	if err := cts.DecodeInto(req); err != nil {
		return nil, errf.NewFromErr(errf.DecodeRequestFailed, err)
	}

	if err := req.Validate(); err != nil {
		return nil, errf.NewFromErr(errf.InvalidParameter, err)
	}

	syncCli, err := svc.syncCli.Other(cts.Kit, req.AccountID)
	if err != nil {
		return nil, err
	}
	op := &hostHandler{syncCli: syncCli}
	if err := op.DeleteHost(cts.Kit, &other.DelHostParams{DelHostIDs: req.HostIDs}); err != nil {
		logs.Errorf("sync host by condition failed, err: %v, req: %+v, rid: %s", err, req, cts.Kit.Rid)
		return nil, err
	}

	return nil, nil
}

// hostHandler cvm sync handler.
type hostHandler struct {
	cli ressync.Interface

	// Prepare 构建参数
	request *sync.OtherSyncHostReq
	syncCli other.Interface
	offset  uint64
}

var _ handler.HandlerV2[cmdb.HostWithCloudID] = new(hostHandler)

// Prepare ...
func (hd *hostHandler) Prepare(cts *rest.Contexts) error {
	req := new(sync.OtherSyncHostReq)
	if err := cts.DecodeInto(req); err != nil {
		return errf.NewFromErr(errf.DecodeRequestFailed, err)
	}

	if err := req.Validate(); err != nil {
		return errf.NewFromErr(errf.InvalidParameter, err)
	}

	syncCli, err := hd.cli.Other(cts.Kit, req.AccountID)
	if err != nil {
		return err
	}

	hd.request = req
	hd.syncCli = syncCli

	return nil
}

// Next ...
func (hd *hostHandler) Next(kt *kit.Kit) ([]cmdb.HostWithCloudID, error) {
	ccHosts, err := hd.listCCHost(kt, core.DefaultMaxPageLimit)
	if err != nil {
		logs.Errorf("failed to list cc host, err: %v, req: %+v, rid: %s", err, hd.request, kt.Rid)
		return nil, err
	}

	if len(ccHosts) == 0 {
		return nil, nil
	}

	hosts := make([]cmdb.HostWithCloudID, 0)
	for _, host := range ccHosts {
		hosts = append(hosts, cmdb.HostWithCloudID{
			Host:    host,
			BizID:   hd.request.BizID,
			CloudID: other.BuildCloudIDFromHostID(host.BkHostID),
		})
	}

	hd.offset += uint64(core.DefaultMaxPageLimit)

	return hosts, nil
}

func (hd *hostHandler) listCCHost(kt *kit.Kit, limit uint) ([]cmdb.Host, error) {
	if hd.request.BizID == constant.HostPoolBiz {
		params := &cmdb.ListResourcePoolHostsParams{
			Fields: cmdb.HostFields,
			Page: &cmdb.BasePage{
				Start: int64(hd.offset),
				Limit: int64(limit),
				Sort:  "bk_host_id",
			},
		}
		result, err := cmdb.CmdbClient().ListResourcePoolHosts(kt, params)
		if err != nil {
			logs.Errorf("failed to list resource pool host, err: %v, req: %+v, rid: %s", err, params, kt.Rid)
			return nil, err
		}
		return result.Info, nil
	}

	params := &cmdb.ListBizHostParams{
		BizID:  hd.request.BizID,
		Fields: cmdb.HostFields,
		Page: &cmdb.BasePage{
			Start: int64(hd.offset),
			Limit: int64(limit),
			Sort:  "bk_host_id",
		},
	}
	var err error
	params, err = hooks.AdjustOtherSyncerListCCHostParams(kt, params)
	if err != nil {
		logs.Errorf("failed to adjust params, err: %v, req: %+v, rid: %s", err, converter.PtrToVal(params), kt.Rid)
		return nil, err
	}

	result, err := cmdb.CmdbClient().ListBizHost(kt, params)
	if err != nil {
		logs.Errorf("call cmdb to list biz host failed, err: %v, req: %+v, rid: %s", err, params, kt.Rid)
		return nil, err
	}

	return result.Info, nil
}

// Sync ...
func (hd *hostHandler) Sync(kt *kit.Kit, hosts []cmdb.HostWithCloudID) error {
	hostIDs := make([]int64, 0, len(hosts))
	hostCache := make(map[int64]cmdb.HostWithCloudID, len(hosts))
	for _, host := range hosts {
		hostIDs = append(hostIDs, host.BkHostID)
		hostCache[host.BkHostID] = host
	}
	params := &other.SyncHostParams{
		AccountID: hd.request.AccountID,
		BizID:     hd.request.BizID,
		HostIDs:   hostIDs,
		HostCache: hostCache,
	}

	return hd.SyncByCond(kt, params)
}

// SyncByCond ...
func (hd *hostHandler) SyncByCond(kt *kit.Kit, params *other.SyncHostParams) error {
	if err := hd.syncCli.Host(kt, params); err != nil {
		logs.Errorf("sync other vendor host failed, err: %v, opt: %v, rid: %s", err, params, kt.Rid)
		return err
	}

	return nil
}

// RemoveDeletedFromCloud ...
func (hd *hostHandler) RemoveDeletedFromCloud(kt *kit.Kit, allCloudIDMap map[string]struct{}) error {
	ccBizExistHostIDs := make(map[int64]struct{}, len(allCloudIDMap))
	for hostIDStr := range allCloudIDMap {
		hostID, err := other.GetHostIDFromCloudID(hostIDStr)
		if err != nil {
			logs.Errorf("failed to get host id from cloud id, err: %v, cloud id: %s, rid: %s", err, hostIDStr, kt.Rid)
			return err
		}
		ccBizExistHostIDs[hostID] = struct{}{}
	}

	params := &other.DelHostParams{BizID: hd.request.BizID, CCBizExistHostIDs: ccBizExistHostIDs}
	return hd.DeleteHost(kt, params)
}

// DeleteHost ...
func (hd *hostHandler) DeleteHost(kt *kit.Kit, params *other.DelHostParams) error {
	err := hd.syncCli.RemoveHostByCCInfo(kt, params)
	if err != nil {
		logs.Errorf("remove host by cc host ids failed, err: %v, param: %+v, rid: %s", err, converter.PtrToVal(params),
			kt.Rid)
		return err
	}

	return nil
}

// SyncConcurrent ...
func (hd *hostHandler) SyncConcurrent() uint {
	if hd.request != nil && hd.request.Concurrent != 0 {
		return hd.request.Concurrent
	}
	// read from config file
	_, syncing := cc.HCService().SyncConfig.GetSyncConcurrent(enumor.Other, enumor.CvmCloudResType,
		cc.ConcurrentWildcard)
	return max(syncing, 1)
}

// Describe ...
func (hd *hostHandler) Describe() string {
	if hd.request == nil {
		return fmt.Sprintf("other %s(-)", hd.Resource())
	}
	return fmt.Sprintf("other %s(bizID=%d,account=%s)", hd.Resource(), hd.request.BizID, hd.request.AccountID)
}

// Resource ...
func (hd *hostHandler) Resource() enumor.CloudResourceType {
	return enumor.CvmCloudResType
}
