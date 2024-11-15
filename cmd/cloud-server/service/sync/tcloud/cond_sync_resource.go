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
	"hcm/pkg/api/hc-service/sync"
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
	enumor.LoadBalancerCloudResType: CondSyncLoadBalancer,
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
