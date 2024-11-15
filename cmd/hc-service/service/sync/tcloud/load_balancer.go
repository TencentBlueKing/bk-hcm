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
	"sync"

	"hcm/cmd/hc-service/logics/res-sync/tcloud"
	"hcm/cmd/hc-service/service/sync/handler"
	typecore "hcm/pkg/adaptor/types/core"
	typeclb "hcm/pkg/adaptor/types/load-balancer"
	"hcm/pkg/criteria/enumor"
	"hcm/pkg/kit"
	"hcm/pkg/logs"
	"hcm/pkg/rest"
	cvt "hcm/pkg/tools/converter"
	"hcm/pkg/tools/slice"

	"golang.org/x/sync/errgroup"
)

// SyncLoadBalancer 同步负载均衡接口
func (svc *service) SyncLoadBalancer(cts *rest.Contexts) (interface{}, error) {
	hd := &lbHandler{
		baseHandler: baseHandler{
			resType: enumor.LoadBalancerCloudResType,
			cli:     svc.syncCli,
		},
	}
	return nil, handler.ResourceSyncV2(cts, hd)
}

// lbHandler lb sync handler.
type lbHandler struct {
	baseHandler
	offset uint64
}

var _ handler.HandlerV2[typeclb.TCloudClb] = new(lbHandler)

// Next ...
func (hd *lbHandler) Next(kt *kit.Kit) ([]typeclb.TCloudClb, error) {

	if len(hd.request.CloudIDs) > 0 {
		// 指定id只处理一次
		listOpt := &typeclb.TCloudListOption{
			Region:   hd.request.Region,
			CloudIDs: hd.request.CloudIDs,
			Page: &typecore.TCloudPage{
				Limit: typecore.TCloudQueryLimit,
			},
			OrderType:  cvt.ValToPtr(typeclb.TCloudCLBOrderAscending),
			OrderBy:    cvt.ValToPtr(typeclb.TCloudOrderByCreateTime),
			TagFilters: hd.request.TagFilters,
		}
		lbResult, err := hd.syncCli.CloudCli().ListLoadBalancer(kt, listOpt)
		if err != nil {
			logs.Errorf("request adaptor list tcloud load balancer failed, err: %v, opt: %+v, rid: %s",
				err, listOpt, kt.Rid)
			return nil, err
		}
		return lbResult, nil
	}

	var eg, _ = errgroup.WithContext(kt.Ctx)
	mu := &sync.Mutex{}
	results := make([]typeclb.TCloudClb, 0, typecore.TCloudQueryLimit*hd.SyncConcurrent())
	for i := uint(0); i < hd.SyncConcurrent(); i++ {
		offset := hd.offset + uint64(typecore.TCloudQueryLimit*i)
		eg.Go(func() error {
			listOpt := &typeclb.TCloudListOption{
				Region: hd.request.Region,
				Page: &typecore.TCloudPage{
					Offset: offset,
					Limit:  typecore.TCloudQueryLimit,
				},
				OrderType:  cvt.ValToPtr(typeclb.TCloudCLBOrderAscending),
				OrderBy:    cvt.ValToPtr(typeclb.TCloudOrderByCreateTime),
				TagFilters: hd.request.TagFilters,
			}
			lbResult, err := hd.syncCli.CloudCli().ListLoadBalancer(kt, listOpt)
			if err != nil {
				logs.Errorf("request adaptor list tcloud load balancer failed, err: %v, opt: %+v, rid: %s",
					err, listOpt, kt.Rid)
				return err
			}
			mu.Lock()
			results = append(results, lbResult...)
			mu.Unlock()
			return nil
		})

	}
	err := eg.Wait()
	if err != nil {
		return nil, err
	}

	if len(results) == 0 {
		return nil, nil
	}
	hd.offset += uint64(len(results))
	return results, nil
}

// Sync ...
func (hd *lbHandler) Sync(kt *kit.Kit, instances []typeclb.TCloudClb) error {

	params := &tcloud.SyncBaseParams{
		AccountID:  hd.request.AccountID,
		Region:     hd.request.Region,
		CloudIDs:   slice.Map(instances, typeclb.TCloudClb.GetCloudID),
		TagFilters: hd.request.TagFilters,
	}
	opt := &tcloud.SyncLBOption{
		PrefetchedLB: instances,
	}
	if _, err := hd.syncCli.LoadBalancerWithListener(kt, params, opt); err != nil {
		logs.Errorf("sync tcloud load balancer with rel failed, err: %v, opt: %v, rid: %s", err, params, kt.Rid)
		return err
	}

	return nil
}

// RemoveDeleteFromCloud ...
func (hd *lbHandler) RemoveDeleteFromCloud(kt *kit.Kit) error {

	params := &tcloud.SyncRemovedParams{
		AccountID:  hd.request.AccountID,
		Region:     hd.request.Region,
		CloudIDs:   hd.request.CloudIDs,
		TagFilters: hd.request.TagFilters,
	}
	if err := hd.syncCli.RemoveLoadBalancerDeleteFromCloud(kt, params); err != nil {
		logs.Errorf("remove load balancer delete from cloud failed, err: %v, accountID: %s, region: %s, rid: %s", err,
			hd.request.AccountID, hd.request.Region, kt.Rid)
		return err
	}

	return nil
}

// RemoveDeletedFromCloud 清理云上已删除资源
func (hd *lbHandler) RemoveDeletedFromCloud(kt *kit.Kit, allCloudIDMap map[string]struct{}) error {

	params := &tcloud.SyncRemovedParams{
		AccountID:  hd.request.AccountID,
		Region:     hd.request.Region,
		CloudIDs:   hd.request.CloudIDs,
		TagFilters: hd.request.TagFilters,
	}
	err := hd.syncCli.RemoveLoadBalancerDeleteFromCloudV2(kt, params, allCloudIDMap)
	if err != nil {
		logs.Errorf("remove clb delete from cloud failed, err: %v, cloud id: %v,account: %s, region: %s, rid: %s",
			err, hd.request.CloudIDs, hd.request.AccountID, hd.request.Region, kt.Rid)
		return err
	}
	return nil
}
