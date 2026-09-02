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

package tcloud

import (
	"hcm/pkg/api/core"
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

	BkBizID    int64                 `json:"bk_biz_id,omitempty"`
	TagFilters core.MultiValueTagMap `json:"tag_filters,omitempty" validate:"max=10"`
}

// CondSyncFunc sync resource by given condition
type CondSyncFunc func(kt *kit.Kit, cliSet *client.ClientSet, params *CondSyncParams) error

// CondAsyncFunc asynchronously syncs a resource by the given condition.
type CondAsyncFunc func(kt *kit.Kit, cliSet *client.ClientSet, params *CondSyncParams) (any, error)

// CondSyncRoute defines the selected conditional sync route.
type CondSyncRoute struct {
	// SyncFunc is the synchronous conditional sync function.
	SyncFunc CondSyncFunc
	// AsyncFunc is the asynchronous conditional sync function.
	AsyncFunc CondAsyncFunc
}

var condSyncFuncMap = map[enumor.CloudResourceType]CondSyncFunc{
	enumor.RegionCloudResType:             CondSyncRegion,
	enumor.ZoneCloudResType:               CondSyncZone,
	enumor.LoadBalancerCloudResType:       CondSyncLoadBalancer,
	enumor.SecurityGroupCloudResType:      CondSyncSecurityGroup,
	enumor.SubAccountCloudResType:         CondSyncSubAccount,
	enumor.PermissionTemplateCloudResType: CondSyncPermissionTemplate,
}

var condAsyncFuncMap = map[enumor.CloudResourceType]CondAsyncFunc{
	enumor.LoadBalancerCloudResType: AsyncCondSyncLoadBalancer,
}

// GetCondSyncRoute gets the conditional sync route for a resource.
func GetCondSyncRoute(res enumor.CloudResourceType) (*CondSyncRoute, bool) {
	syncFunc, syncOK := condSyncFuncMap[res]
	asyncFunc, asyncOK := condAsyncFuncMap[res]
	if !syncOK && !asyncOK {
		return nil, false
	}

	return &CondSyncRoute{
		SyncFunc:  syncFunc,
		AsyncFunc: asyncFunc,
	}, true
}

// HasAsync checks whether the selected route has an asynchronous override.
func (r *CondSyncRoute) HasAsync() bool {
	return r != nil && r.AsyncFunc != nil
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
		err := cliSet.HCService().TCloud.Clb.SyncLoadBalancer(kt, &syncReq)
		if err != nil {
			logs.Errorf("[%s] conditional sync load balancer failed, err: %v, req: %+v, rid: %s",
				enumor.TCloud, err, syncReq, kt.Rid)
			return err
		}
		logs.Infof("[%s] conditional sync load balancer end, req: %+v, rid: %s", enumor.TCloud, syncReq, kt.Rid)
	}
	return nil
}

// CondSyncSecurityGroup ...
func CondSyncSecurityGroup(kt *kit.Kit, cliSet *client.ClientSet, params *CondSyncParams) error {
	syncReq := sync.TCloudSyncReq{
		AccountID:  params.AccountID,
		CloudIDs:   params.CloudIDs,
		TagFilters: params.TagFilters,
	}
	for i := range params.Regions {
		syncReq.Region = params.Regions[i]
		err := cliSet.HCService().TCloud.SecurityGroup.SyncSecurityGroup(kt.Ctx, kt.Header(), &syncReq)
		if err != nil {
			logs.Errorf("[%s] conditional sync security group failed, err: %v, req: %+v, rid: %s",
				enumor.TCloud, err, syncReq, kt.Rid)
			return err
		}
		logs.Infof("[%s] conditional sync security group end, req: %+v, rid: %s", enumor.TCloud, syncReq, kt.Rid)
	}
	return nil
}

// CondSyncRegion sync region
func CondSyncRegion(kt *kit.Kit, cliSet *client.ClientSet, params *CondSyncParams) error {
	syncReq := &region.TCloudRegionSyncReq{
		AccountID: params.AccountID,
	}
	err := cliSet.HCService().TCloud.Region.Sync(kt.Ctx, kt.Header(), syncReq)
	if err != nil {
		logs.Errorf("[%s] conditional sync region failed, err: %v, req: %+v, rid: %s",
			enumor.TCloud, err, syncReq, kt.Rid)
		return err
	}
	logs.Infof("[%s] conditional sync region end, req: %+v, rid: %s", enumor.TCloud, syncReq, kt.Rid)
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
		err := cliSet.HCService().TCloud.Zone.SyncZone(kt.Ctx, kt.Header(), syncReq)
		if err != nil {
			logs.Errorf("[%s] conditional sync zone failed, err: %v, req: %+v, rid: %s",
				enumor.TCloud, err, syncReq, kt.Rid)
			return err
		}
		logs.Infof("[%s] conditional sync zone end, req: %+v, rid: %s", enumor.TCloud, syncReq, kt.Rid)
	}
	return nil
}

// CondSyncSubAccount ...
func CondSyncSubAccount(kt *kit.Kit, cliSet *client.ClientSet, params *CondSyncParams) error {
	syncReq := &sync.TCloudGlobalSyncReq{
		AccountID: params.AccountID,
	}
	err := cliSet.HCService().TCloud.Account.SyncSubAccount(kt, syncReq)
	if err != nil {
		logs.Errorf("[%s] conditional sync sub account failed, err: %v, req: %+v, rid: %s",
			enumor.TCloud, err, syncReq, kt.Rid)
		return err
	}

	logs.Infof("[%s] conditional sync sub account end, req: %+v, rid: %s", enumor.TCloud, syncReq, kt.Rid)

	return nil
}

// CondSyncPermissionTemplate ...
func CondSyncPermissionTemplate(kt *kit.Kit, cliSet *client.ClientSet, params *CondSyncParams) error {
	syncReq := &sync.TCloudGlobalSyncReq{
		AccountID: params.AccountID,
	}
	err := cliSet.HCService().TCloud.Account.SyncPermissionTemplate(kt, syncReq)
	if err != nil {
		logs.Errorf("[%s] conditional sync permission template failed, err: %v, req: %+v, rid: %s",
			enumor.TCloud, err, syncReq, kt.Rid)
		return err
	}

	logs.Infof("[%s] conditional sync permission template end, req: %+v, rid: %s",
		enumor.TCloud, syncReq, kt.Rid)

	return nil
}
