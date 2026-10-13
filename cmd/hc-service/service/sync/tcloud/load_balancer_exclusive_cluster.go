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
	"hcm/cmd/hc-service/logics/res-sync/tcloud"
	"hcm/cmd/hc-service/service/sync/handler"
	typecore "hcm/pkg/adaptor/types/core"
	typeslb "hcm/pkg/adaptor/types/load-balancer"
	"hcm/pkg/criteria/enumor"
	"hcm/pkg/kit"
	"hcm/pkg/logs"
	"hcm/pkg/rest"
	"hcm/pkg/tools/slice"
)

// exclusiveClusterNetwork、exclusiveClusterTypes 全量同步产品范围固定为公网TGW/STGW，非云API限制
var exclusiveClusterTypes = []enumor.ClusterType{enumor.TGWClusterType, enumor.STGWClusterType}

const exclusiveClusterNetwork = enumor.PublicClusterNetwork

// SyncExclusiveCluster 同步独占集群接口，请求指定cloud_ids时只同步对应实例
func (svc *service) SyncExclusiveCluster(cts *rest.Contexts) (interface{}, error) {
	hd := &exclusiveClusterHandler{
		baseHandler: baseHandler{
			resType: enumor.LoadBalancerExclusiveClusterCloudResType,
			cli:     svc.syncCli,
		},
	}
	return nil, handler.ResourceSyncV2(cts, hd)
}

// exclusiveClusterHandler 独占集群同步handler，范围固定为公网TGW/STGW
type exclusiveClusterHandler struct {
	baseHandler
	offset uint64
}

var _ handler.HandlerV2[typeslb.TCloudExclusiveCluster] = new(exclusiveClusterHandler)

// Next 串行分页拉取云上公网TGW/STGW独占集群，集群对象没有第二套详情接口，ClusterSet即全量
func (hd *exclusiveClusterHandler) Next(kt *kit.Kit) ([]typeslb.TCloudExclusiveCluster, error) {
	if len(hd.request.CloudIDs) > 0 {
		// 指定id只处理一次
		opt := &typeslb.TCloudExclusiveClusterListOption{
			Region:       hd.request.Region,
			CloudIDs:     hd.request.CloudIDs,
			Network:      exclusiveClusterNetwork,
			ClusterTypes: exclusiveClusterTypes,
			Page:         &typecore.TCloudPage{Limit: typecore.TCloudQueryLimit},
		}
		results, err := hd.syncCli.CloudCli().ListExclusiveClusters(kt, opt)
		if err != nil {
			logs.Errorf("request adaptor list tcloud exclusive cluster failed, err: %v, opt: %+v, rid: %s",
				err, opt, kt.Rid)
			return nil, err
		}
		return results, nil
	}

	opt := &typeslb.TCloudExclusiveClusterListOption{
		Region:       hd.request.Region,
		Network:      exclusiveClusterNetwork,
		ClusterTypes: exclusiveClusterTypes,
		Page:         &typecore.TCloudPage{Offset: hd.offset, Limit: typecore.TCloudQueryLimit},
	}
	results, err := hd.syncCli.CloudCli().ListExclusiveClusters(kt, opt)
	if err != nil {
		logs.Errorf("request adaptor list tcloud exclusive cluster failed, err: %v, opt: %+v, rid: %s",
			err, opt, kt.Rid)
		return nil, err
	}

	if len(results) == 0 {
		return nil, nil
	}
	hd.offset += typecore.TCloudQueryLimit
	return results, nil
}

// Sync ...
func (hd *exclusiveClusterHandler) Sync(kt *kit.Kit, instances []typeslb.TCloudExclusiveCluster) error {

	params := &tcloud.SyncBaseParams{
		AccountID: hd.request.AccountID,
		Region:    hd.request.Region,
		CloudIDs:  slice.Map(instances, typeslb.TCloudExclusiveCluster.GetCloudID),
	}
	if _, err := hd.syncCli.ExclusiveCluster(kt, params, new(tcloud.SyncExclusiveClusterOption)); err != nil {
		logs.Errorf("sync tcloud exclusive cluster failed, err: %v, account: %s, region: %s, count: %d, rid: %s",
			err, params.AccountID, params.Region, len(params.CloudIDs), kt.Rid)
		return err
	}

	return nil
}

// RemoveDeletedFromCloud 清理云上已删除的独占集群
func (hd *exclusiveClusterHandler) RemoveDeletedFromCloud(kt *kit.Kit, allCloudIDMap map[string]struct{}) error {

	params := &tcloud.SyncRemovedParams{
		AccountID: hd.request.AccountID,
		Region:    hd.request.Region,
		CloudIDs:  hd.request.CloudIDs,
	}
	if err := hd.syncCli.RemoveExclusiveClusterDeleteFromCloud(kt, params, allCloudIDMap); err != nil {
		logs.Errorf("remove exclusive cluster delete from cloud failed, err: %v, account: %s, region: %s, "+
			"rid: %s", err, hd.request.AccountID, hd.request.Region, kt.Rid)
		return err
	}

	return nil
}
