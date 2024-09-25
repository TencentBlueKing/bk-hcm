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
	"fmt"

	ressync "hcm/cmd/hc-service/logics/res-sync"
	"hcm/cmd/hc-service/logics/res-sync/tcloud"
	"hcm/cmd/hc-service/service/sync/handler"
	typecore "hcm/pkg/adaptor/types/core"
	typeclb "hcm/pkg/adaptor/types/load-balancer"
	"hcm/pkg/api/core"
	corelb "hcm/pkg/api/core/cloud/load-balancer"
	"hcm/pkg/api/hc-service/sync"
	dataservice "hcm/pkg/client/data-service"
	"hcm/pkg/criteria/enumor"
	"hcm/pkg/criteria/errf"
	"hcm/pkg/dal/dao/tools"
	"hcm/pkg/kit"
	"hcm/pkg/logs"
	"hcm/pkg/rest"
	cvt "hcm/pkg/tools/converter"
)

// SyncLoadBalancerListener 同步负载均衡监听器接口
func (svc *service) SyncLoadBalancerListener(cts *rest.Contexts) (any, error) {
	return nil, handler.ResourceSync(cts, &lblHandler{cli: svc.syncCli, dataCli: svc.dataCli})
}

// lblHandler lb listener sync handler.
type lblHandler struct {
	cli ressync.Interface

	request *sync.TCloudListenerSyncReq
	syncCli tcloud.Interface
	offset  uint64
	dataCli *dataservice.Client
	lbInfo  *corelb.TCloudLoadBalancer
}

var _ handler.Handler = new(lblHandler)

// Prepare ...
func (hd *lblHandler) Prepare(cts *rest.Contexts) error {
	req := new(sync.TCloudListenerSyncReq)
	if err := cts.DecodeInto(req); err != nil {
		return errf.NewFromErr(errf.DecodeRequestFailed, err)
	}

	if err := req.Validate(); err != nil {
		return errf.NewFromErr(errf.InvalidParameter, err)
	}

	// 查询负载均衡本地数据
	listReq := &core.ListReq{
		Filter: tools.ExpressionAnd(
			tools.RuleEqual("account_id", req.AccountID),
			tools.RuleEqual("cloud_id", req.LoadBalancerCloudID)),
		Page: core.NewDefaultBasePage(),
	}
	lbResp, err := hd.dataCli.TCloud.LoadBalancer.ListLoadBalancer(cts.Kit, listReq)
	if err != nil {
		logs.Errorf("fail to query load balancer for sync listener, err: %v, listReq:%#v, rid: %s",
			err, listReq, cts.Kit.Rid)
		return err
	}
	// 本地没有的数据不支持同步
	if len(lbResp.Details) == 0 {
		return fmt.Errorf("load balancer(%s) cannot be found in database", hd.request.LoadBalancerCloudID)
	}
	hd.lbInfo = cvt.ValToPtr(lbResp.Details[0])

	syncCli, err := hd.cli.TCloud(cts.Kit, req.AccountID)
	if err != nil {
		return err
	}

	hd.request = req
	hd.syncCli = syncCli

	return nil
}

// Next ...
func (hd *lblHandler) Next(kt *kit.Kit) ([]string, error) {
	listOpt := &typeclb.TCloudListListenersOption{
		Region:         hd.request.Region,
		LoadBalancerId: hd.request.LoadBalancerCloudID,
		CloudIDs:       hd.request.CloudIDs,
	}

	lbResult, err := hd.syncCli.CloudCli().ListListener(kt, listOpt)
	if err != nil {
		logs.Errorf("request adaptor list tcloud load balancer failed, err: %v, opt: %v, rid: %s", err, listOpt, kt.Rid)
		return nil, err
	}

	if len(lbResult) == 0 {
		return nil, nil
	}

	cloudIDs := make([]string, 0, len(lbResult))
	for _, one := range lbResult {
		cloudIDs = append(cloudIDs, one.GetCloudID())
	}

	hd.offset += typecore.TCloudQueryLimit
	return cloudIDs, nil
}

// Sync ...
func (hd *lblHandler) Sync(kt *kit.Kit, cloudIDs []string) error {
	params := &tcloud.SyncBaseParams{
		AccountID: hd.request.AccountID,
		Region:    hd.request.Region,
		CloudIDs:  cloudIDs,
	}
	opt := &tcloud.SyncListenerOption{
		BizID:              hd.lbInfo.BkBizID,
		LBID:               hd.lbInfo.ID,
		CloudLBID:          hd.request.LoadBalancerCloudID,
		CachedLoadBalancer: hd.lbInfo,
	}
	if _, err := hd.syncCli.Listener(kt, params, opt); err != nil {
		logs.Errorf("sync tcloud load balancer with rel failed, err: %v, params: %v, opt: %v, rid: %s",
			err, params, opt, kt.Rid)
		return err
	}

	return nil
}

// RemoveDeleteFromCloud ...
func (hd *lblHandler) RemoveDeleteFromCloud(kt *kit.Kit) error {
	params := &tcloud.ListenerSyncRemovedParams{
		AccountID:          hd.request.AccountID,
		Region:             hd.request.Region,
		CloudIDs:           hd.request.CloudIDs,
		BizID:              hd.lbInfo.BkBizID,
		LBID:               hd.lbInfo.ID,
		CloudLBID:          hd.request.LoadBalancerCloudID,
		CachedLoadBalancer: hd.lbInfo,
	}
	err := hd.syncCli.RemoveListenerDeleteFromCloud(kt, params)
	if err != nil {
		logs.Errorf("fail to remove deleted tcloud listener, err: %v, params: %v, rid: %s",
			err, params, kt.Rid)
		return err
	}
	return nil
}

// Name load_balancer
func (hd *lblHandler) Name() enumor.CloudResourceType {
	return enumor.LoadBalancerCloudResType
}
