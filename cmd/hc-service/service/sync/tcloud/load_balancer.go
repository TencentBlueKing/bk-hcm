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
	typeclb "hcm/pkg/adaptor/types/load-balancer"
	"hcm/pkg/api/hc-service/sync"
	"hcm/pkg/criteria/enumor"
	"hcm/pkg/kit"
	"hcm/pkg/logs"
	"hcm/pkg/rest"
	"hcm/pkg/tools/converter"
)

// SyncLoadBalancer 同步负载均衡接口
func (svc *service) SyncLoadBalancer(cts *rest.Contexts) (interface{}, error) {
	return nil, handler.ResourceSyncV2(cts, &lbHandler{cli: svc.syncCli})
}

// lbHandler lb sync handler.
type lbHandler struct {
	cli ressync.Interface

	request *sync.TCloudSyncReq
	syncCli tcloud.Interface
	// 缓存负载均衡数据，避免反复调用云上接口拉数据
	lbCache []typeclb.TCloudClb
	offset  uint64
}

var _ handler.HandlerV2 = new(lbHandler)

// Prepare ...
func (hd *lbHandler) Prepare(cts *rest.Contexts) error {
	request, syncCli, err := defaultPrepare(cts, hd.cli)
	if err != nil {
		return err
	}

	hd.request = request
	hd.syncCli = syncCli

	return nil
}

// Next ...
func (hd *lbHandler) Next(kt *kit.Kit) ([]string, error) {

	if len(hd.request.CloudIDs) > 0 {
		// 指定资源同步的情况，直接返回指定资源id即可
		return hd.request.CloudIDs, nil
	}

	cloudIDs := make([]string, 0, 100)
	hd.lbCache = make([]typeclb.TCloudClb, 0, 100)

	listOpt := &typeclb.TCloudListOption{
		Region: hd.request.Region,
		Page: &typecore.TCloudPage{
			Offset: hd.offset,
			Limit:  typecore.TCloudQueryLimit,
		},
	}

	lbResult, err := hd.syncCli.CloudCli().ListLoadBalancer(kt, listOpt)
	if err != nil {
		logs.Errorf("request adaptor list tcloud load balancer failed, err: %v, opt: %+v, rid: %s",
			err, listOpt, kt.Rid)
		return nil, err
	}

	if len(lbResult) == 0 {
		return nil, nil
	}

	for _, one := range lbResult {
		cloudIDs = append(cloudIDs, converter.PtrToVal(one.LoadBalancerId))
	}
	hd.lbCache = append(hd.lbCache, lbResult...)
	hd.offset += uint64(len(lbResult))
	return cloudIDs, nil
}

// Sync ...
func (hd *lbHandler) Sync(kt *kit.Kit, cloudIDs []string) error {
	params := &tcloud.SyncBaseParams{
		AccountID: hd.request.AccountID,
		Region:    hd.request.Region,
		CloudIDs:  cloudIDs,
	}
	if _, err := hd.syncCli.LoadBalancerWithListener(kt, params, new(tcloud.SyncLBOption)); err != nil {
		logs.Errorf("sync tcloud load balancer with rel failed, err: %v, opt: %v, rid: %s", err, params, kt.Rid)
		return err
	}

	return nil
}

// RemoveDeleteFromCloud ...
func (hd *lbHandler) RemoveDeleteFromCloud(kt *kit.Kit) error {

	params := &tcloud.SyncRemovedParams{
		AccountID: hd.request.AccountID,
		Region:    hd.request.Region,
		CloudIDs:  hd.request.CloudIDs,
	}
	if err := hd.syncCli.RemoveLoadBalancerDeleteFromCloud(kt, params); err != nil {
		logs.Errorf("remove load balancer delete from cloud failed, err: %v, accountID: %s, region: %s, rid: %s", err,
			hd.request.AccountID, hd.request.Region, kt.Rid)
		return err
	}

	return nil
}

// RemoveDeleteFromCloudV2 清理云上已删除资源
func (hd *lbHandler) RemoveDeleteFromCloudV2(kt *kit.Kit, allCloudIDMap map[string]struct{}) error {

	params := &tcloud.SyncRemovedParams{
		AccountID: hd.request.AccountID,
		Region:    hd.request.Region,
		CloudIDs:  hd.request.CloudIDs,
	}
	err := hd.syncCli.RemoveLoadBalancerDeleteFromCloudV2(kt, params, allCloudIDMap)
	if err != nil {
		logs.Errorf("remove sg delete from cloud failed, err: %v, accountID: %s, region: %s, rid: %s", err,
			hd.request.AccountID, hd.request.Region, kt.Rid)
		return err
	}
	return nil
}

// Name load_balancer
func (hd *lbHandler) Name() enumor.CloudResourceType {
	return enumor.LoadBalancerCloudResType
}
