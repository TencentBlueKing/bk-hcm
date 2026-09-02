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
	"time"

	"hcm/cmd/hc-service/logics/res-sync/tcloud"
	"hcm/cmd/hc-service/service/sync/handler"
	typecore "hcm/pkg/adaptor/types/core"
	typeclb "hcm/pkg/adaptor/types/load-balancer"
	hcsync "hcm/pkg/api/hc-service/sync"
	"hcm/pkg/criteria/constant"
	"hcm/pkg/criteria/enumor"
	"hcm/pkg/criteria/errf"
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

// DeleteLoadBalancerByCond 按条件删除负载均衡接口
func (svc *service) DeleteLoadBalancerByCond(cts *rest.Contexts) (interface{}, error) {
	hd := &lbHandler{
		baseHandler: baseHandler{
			resType: enumor.LoadBalancerCloudResType,
			cli:     svc.syncCli,
		},
	}
	return hd.DeleteLoadBalancerByCond(cts)
}

// SyncLoadBalancerByCond 按条件同步负载均衡接口
func (svc *service) SyncLoadBalancerByCond(cts *rest.Contexts) (interface{}, error) {
	hd := &lbHandler{
		baseHandler: baseHandler{
			resType: enumor.LoadBalancerCloudResType,
			cli:     svc.syncCli,
		},
	}
	return hd.SyncLoadBalancerByCond(cts)
}

// lbHandler lb sync handler.
type lbHandler struct {
	baseHandler
	offset uint64
}

// DeleteLoadBalancerByCond 按条件删除负载均衡
func (hd *lbHandler) DeleteLoadBalancerByCond(cts *rest.Contexts) (interface{}, error) {
	req := new(hcsync.TCloudDelLoadBalancerByCondReq)
	if err := cts.DecodeInto(req); err != nil {
		return nil, errf.NewFromErr(errf.DecodeRequestFailed, err)
	}
	if err := req.Validate(); err != nil {
		return nil, errf.NewFromErr(errf.InvalidParameter, err)
	}

	syncCli, err := hd.cli.TCloud(cts.Kit, req.AccountID)
	if err != nil {
		logs.Errorf("init tcloud sync client failed, err: %v, account: %s, rid: %s", err, req.AccountID, cts.Kit.Rid)
		return nil, err
	}

	startedAt := time.Now()
	if err = syncCli.BatchDeleteLoadBalancer(cts.Kit, req.AccountID, req.Region, req.CloudIDs); err != nil {
		logs.Errorf("delete load balancer by condition failed, err: %v, account: %s, region: %s, "+
			"count: %d, rid: %s", err, req.AccountID, req.Region, len(req.CloudIDs), cts.Kit.Rid)
		return nil, err
	}
	logs.Infof("delete load balancer by condition success, account: %s, region: %s, count: %d, cost: %s, rid: %s",
		req.AccountID, req.Region, len(req.CloudIDs), time.Since(startedAt), cts.Kit.Rid)

	return nil, nil
}

// SyncLoadBalancerByCond 按条件同步负载均衡，只同步指定云上ID的实例，不做DB全量清理
func (hd *lbHandler) SyncLoadBalancerByCond(cts *rest.Contexts) (interface{}, error) {
	req := new(hcsync.TCloudSyncLoadBalancerByCondReq)
	if err := cts.DecodeInto(req); err != nil {
		return nil, errf.NewFromErr(errf.DecodeRequestFailed, err)
	}
	if err := req.Validate(); err != nil {
		return nil, errf.NewFromErr(errf.InvalidParameter, err)
	}

	syncCli, err := hd.cli.TCloud(cts.Kit, req.AccountID)
	if err != nil {
		logs.Errorf("init tcloud sync client failed, err: %v, account: %s, rid: %s", err, req.AccountID, cts.Kit.Rid)
		return nil, err
	}
	hd.syncCli = syncCli
	hd.request = &hcsync.TCloudSyncReq{
		AccountID:  req.AccountID,
		Region:     req.Region,
		TagFilters: req.TagFilters,
	}

	listStartedAt := time.Now()
	instances, err := hd.listLoadBalancerByCloudIDs(cts.Kit, req.CloudIDs)
	if err != nil {
		logs.Errorf("list load balancer by cloud ids failed, err: %v, account: %s, region: %s, "+
			"count: %d, rid: %s", err, req.AccountID, req.Region, len(req.CloudIDs), cts.Kit.Rid)
		return nil, err
	}
	logs.Infof("list load balancer by cloud ids done, account: %s, region: %s, count: %d, found: %d, "+
		"cost: %s, rid: %s", req.AccountID, req.Region, len(req.CloudIDs), len(instances),
		time.Since(listStartedAt), cts.Kit.Rid)
	if len(instances) == 0 {
		// 云上已不存在的实例由删除动作负责清理，此处直接跳过
		logs.Infof("no load balancer found on cloud, account: %s, region: %s, count: %d, rid: %s",
			req.AccountID, req.Region, len(req.CloudIDs), cts.Kit.Rid)
		return nil, nil
	}

	// 按单次同步上限切分，保持与全量同步一致的下发粒度，避免每批都被当作最后一批再按并发拆细
	allInstances := slice.Split(instances, constant.CloudResourceSyncMaxLimit)
	syncStartedAt := time.Now()
	if _, _, err = handler.SyncResourcesDetail(cts.Kit, hd, len(instances), allInstances); err != nil {
		logs.Errorf("sync load balancer detail failed, err: %v, account: %s, region: %s, count: %d, rid: %s",
			err, req.AccountID, req.Region, len(instances), cts.Kit.Rid)
		return nil, err
	}
	logs.Infof("sync load balancer by condition success, account: %s, region: %s, count: %d, cost: %s, rid: %s",
		req.AccountID, req.Region, len(instances), time.Since(syncStartedAt), cts.Kit.Rid)

	return nil, nil
}

var _ handler.HandlerV2[typeclb.TCloudClb] = new(lbHandler)

// Next ...
func (hd *lbHandler) Next(kt *kit.Kit) ([]typeclb.TCloudClb, error) {

	if len(hd.request.CloudIDs) > 0 {
		// 指定id只处理一次
		return hd.listLoadBalancerByCloudIDs(kt, hd.request.CloudIDs)
	}

	results, err := hd.concurrentListLoadBalancer(kt, hd.buildPageListOpts())
	if err != nil {
		return nil, err
	}

	if len(results) == 0 {
		return nil, nil
	}
	hd.offset += uint64(len(results))
	return results, nil
}

// buildListOpt 构造云上负载均衡查询参数，分页查询与指定ID查询共用
func (hd *lbHandler) buildListOpt(page *typecore.TCloudPage, cloudIDs []string) *typeclb.TCloudListOption {

	return &typeclb.TCloudListOption{
		Region:     hd.request.Region,
		CloudIDs:   cloudIDs,
		Page:       page,
		OrderType:  cvt.ValToPtr(typeclb.TCloudCLBOrderAscending),
		OrderBy:    cvt.ValToPtr(typeclb.TCloudOrderByCreateTime),
		TagFilters: hd.request.TagFilters,
	}
}

// buildPageListOpts 构造本轮分页查询参数，从当前游标开始，每个并发取一页
func (hd *lbHandler) buildPageListOpts() []*typeclb.TCloudListOption {

	concurrent := hd.SyncConcurrent()
	opts := make([]*typeclb.TCloudListOption, 0, concurrent)
	for i := uint(0); i < concurrent; i++ {
		page := &typecore.TCloudPage{
			Offset: hd.offset + uint64(typecore.TCloudQueryLimit*i),
			Limit:  typecore.TCloudQueryLimit,
		}
		opts = append(opts, hd.buildListOpt(page, nil))
	}

	return opts
}

// listLoadBalancerByCloudIDs 按云上ID拉取负载均衡，云上指定ID查询单次上限 constant.TCLBDescribeMax 个，分批并发查询
func (hd *lbHandler) listLoadBalancerByCloudIDs(kt *kit.Kit, cloudIDs []string) ([]typeclb.TCloudClb, error) {

	idBatches := slice.Split(cloudIDs, constant.TCLBDescribeMax)
	opts := make([]*typeclb.TCloudListOption, 0, len(idBatches))
	for _, idBatch := range idBatches {
		page := &typecore.TCloudPage{Limit: typecore.TCloudQueryLimit}
		opts = append(opts, hd.buildListOpt(page, idBatch))
	}

	return hd.concurrentListLoadBalancer(kt, opts)
}

// concurrentListLoadBalancer 并发执行给定的云上负载均衡查询，合并返回结果
func (hd *lbHandler) concurrentListLoadBalancer(kt *kit.Kit, opts []*typeclb.TCloudListOption) (
	[]typeclb.TCloudClb, error) {

	eg, _ := errgroup.WithContext(kt.Ctx)
	eg.SetLimit(int(hd.SyncConcurrent()))
	mu := &sync.Mutex{}
	results := make([]typeclb.TCloudClb, 0, len(opts)*typecore.TCloudQueryLimit)
	for _, listOpt := range opts {
		eg.Go(func() error {
			lbResult, err := hd.syncCli.CloudCli().ListLoadBalancer(kt, listOpt)
			if err != nil {
				logs.Errorf("request adaptor list tcloud load balancer failed, err: %v, account: %s, "+
					"region: %s, offset: %d, cloud_id_count: %d, rid: %s", err, hd.request.AccountID,
					hd.request.Region, listOpt.Page.Offset, len(listOpt.CloudIDs), kt.Rid)
				return err
			}
			mu.Lock()
			results = append(results, lbResult...)
			mu.Unlock()
			return nil
		})
	}
	if err := eg.Wait(); err != nil {
		return nil, err
	}

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
		logs.Errorf("sync tcloud load balancer with rel failed, err: %v, account: %s, region: %s, "+
			"count: %d, rid: %s", err, params.AccountID, params.Region, len(params.CloudIDs), kt.Rid)
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
