/*
 * TencentBlueKing is pleased to support the open source community by making
 * 蓝鲸智云 - 混合云管理平台 (BlueKing - Hybrid Cloud Management System) available.
 * Copyright (C) 2024 THL A29 Limited,
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

package tziyan

import (
	gosync "sync"

	"hcm/pkg/api/core"
	protoimage "hcm/pkg/api/hc-service/image"
	"hcm/pkg/api/hc-service/region"
	"hcm/pkg/api/hc-service/sync"
	"hcm/pkg/api/hc-service/zone"
	"hcm/pkg/client"
	"hcm/pkg/criteria/enumor"
	"hcm/pkg/kit"
	"hcm/pkg/logs"
)

// CondSyncParams 条件同步选项
type CondSyncParams struct {
	AccountID string   `json:"account_id" validate:"required"`
	Regions   []string `json:"regions,omitempty" validate:"max=20"`
	CloudIDs  []string `json:"cloud_ids,omitempty" validate:"max=20"`

	TagFilters core.MultiValueTagMap `json:"tag_filters,omitempty" validate:"max=10"`
}

// CondSyncFunc sync resource by given condition
type CondSyncFunc func(kt *kit.Kit, cliSet *client.ClientSet, params *CondSyncParams) error

var condSyncFuncMap = map[enumor.CloudResourceType]CondSyncFunc{
	enumor.RegionCloudResType:        CondSyncRegion,
	enumor.ZoneCloudResType:          CondSyncZone,
	enumor.ImageCloudResType:         CondSyncImage,
	enumor.LoadBalancerCloudResType:  CondSyncLoadBalancer,
	enumor.SecurityGroupCloudResType: CondSyncSecurityGroup,
	enumor.DeviceType:                CondSyncDeviceType,
}

// GetCondSyncFunc ...
func GetCondSyncFunc(res enumor.CloudResourceType) (syncFunc CondSyncFunc, ok bool) {
	syncFunc, ok = condSyncFuncMap[res]
	return syncFunc, ok
}

// CondSyncLoadBalancer ...
func CondSyncLoadBalancer(kt *kit.Kit, cliSet *client.ClientSet, params *CondSyncParams) error {
	syncReq := sync.TCloudSyncReq{
		AccountID:  params.AccountID,
		CloudIDs:   params.CloudIDs,
		TagFilters: params.TagFilters,
	}
	for i := range params.Regions {
		syncReq.Region = params.Regions[i]
		err := cliSet.HCService().TCloudZiyan.Clb.SyncLoadBalancer(kt, &syncReq)
		if err != nil {
			logs.Errorf("[%s] conditional sync load balancer failed, err: %v, req: %+v, rid: %s",
				enumor.TCloudZiyan, err, syncReq, kt.Rid)
			return err
		}
		logs.Infof("[%s] conditional sync load balancer end, req: %+v, rid: %s", enumor.TCloudZiyan, syncReq, kt.Rid)
	}
	return nil
}

// CondSyncSecurityGroup ...
func CondSyncSecurityGroup(kt *kit.Kit, cliSet *client.ClientSet, params *CondSyncParams) error {
	syncReq := sync.TCloudSyncReq{
		AccountID: params.AccountID,
		CloudIDs:  params.CloudIDs,
	}
	// TagFilters 和 CloudIDs 不能同时传入, 上层调用的时候又默认注入了tagFilter, 因此当CloudIDs为空时, 才TagFilters传入
	if len(params.CloudIDs) == 0 {
		syncReq.TagFilters = params.TagFilters
	}
	for i := range params.Regions {
		syncReq.Region = params.Regions[i]
		err := cliSet.HCService().TCloudZiyan.SecurityGroup.SyncSecurityGroup(kt, &syncReq)
		if err != nil {
			logs.Errorf("[%s] conditional sync security group failed, err: %v, req: %+v, rid: %s",
				enumor.TCloudZiyan, err, syncReq, kt.Rid)
			return err
		}
		logs.Infof("[%s] conditional sync security group end, req: %+v, rid: %s",
			enumor.TCloudZiyan, syncReq, kt.Rid)
	}
	return nil
}

// CondSyncRegion sync region
func CondSyncRegion(kt *kit.Kit, cliSet *client.ClientSet, params *CondSyncParams) error {
	syncReq := &region.TCloudRegionSyncReq{
		AccountID: params.AccountID,
	}
	err := cliSet.HCService().TCloudZiyan.Region.Sync(kt, syncReq)
	if err != nil {
		logs.Errorf("[%s] conditional sync region failed, err: %v, req: %+v, rid: %s",
			enumor.TCloudZiyan, err, syncReq, kt.Rid)
		return err
	}
	logs.Infof("[%s] conditional sync region end, req: %+v, rid: %s", enumor.TCloudZiyan, syncReq, kt.Rid)
	return nil
}

// CondSyncZone sync zone
func CondSyncZone(kt *kit.Kit, cliSet *client.ClientSet, params *CondSyncParams) error {
	// zone不能根据cloudID或tagFilter进行部分同步
	syncReq := &zone.TCloudZoneSyncReq{
		AccountID: params.AccountID,
	}
	for i := range params.Regions {
		syncReq.Region = params.Regions[i]
		err := cliSet.HCService().TCloudZiyan.Zone.SyncZone(kt, syncReq)
		if err != nil {
			logs.Errorf("[%s] conditional sync zone failed, err: %v, req: %+v, rid: %s",
				enumor.TCloudZiyan, err, syncReq, kt.Rid)
			return err
		}
		logs.Infof("[%s] conditional sync zone end, req: %+v, rid: %s", enumor.TCloudZiyan, syncReq, kt.Rid)
	}
	return nil
}

// CondSyncImage sync image
func CondSyncImage(kt *kit.Kit, cliSet *client.ClientSet, params *CondSyncParams) error {
	pipeline := make(chan bool, syncConcurrencyCount)
	var firstErr error
	var wg gosync.WaitGroup
	for _, oneRegion := range params.Regions {
		pipeline <- true
		wg.Add(1)

		go func(region string) {
			defer func() {
				wg.Done()
				<-pipeline
			}()

			// cloud不能根据cloudID或tagFilter进行部分同步
			req := &protoimage.TCloudImageSyncReq{
				AccountID: params.AccountID,
				Region:    region,
			}
			err := cliSet.HCService().TCloudZiyan.Image.SyncImage(kt.Ctx, kt.Header(), req)
			if firstErr == nil && err != nil {
				logs.Errorf("sync tcloud ziyan image failed, err: %v, req: %v, rid: %s", err, req, kt.Rid)
				firstErr = err
				return
			}
		}(oneRegion)
	}

	wg.Wait()

	if firstErr != nil {
		return firstErr
	}

	return nil
}

// CondSyncDeviceType sync device type
func CondSyncDeviceType(kt *kit.Kit, cliSet *client.ClientSet, params *CondSyncParams) error {
	syncReq := sync.TCloudSyncReq{AccountID: params.AccountID}
	for i := range params.Regions {
		syncReq.Region = params.Regions[i]
		err := cliSet.HCService().TCloudZiyan.DeviceType.SyncDeviceType(kt, &syncReq)
		if err != nil {
			logs.Errorf("[%s] conditional sync device type failed, err: %v, req: %+v, rid: %s",
				enumor.TCloudZiyan, err, syncReq, kt.Rid)
			return err
		}
		logs.Infof("[%s] conditional sync device type end, req: %+v, rid: %s", enumor.TCloudZiyan, syncReq, kt.Rid)
	}
	return nil
}
