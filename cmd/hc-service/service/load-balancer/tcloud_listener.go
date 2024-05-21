/*
 *
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

package loadbalancer

import (
	loadbalancer "hcm/pkg/adaptor/types/load-balancer"
	protolb "hcm/pkg/api/hc-service/load-balancer"
	"hcm/pkg/criteria/errf"
	"hcm/pkg/rest"
)

// QueryListenerTargetsByCloudIDs 直接从云上查询监听器RS列表
func (svc *clbSvc) QueryListenerTargetsByCloudIDs(cts *rest.Contexts) (any, error) {

	req := new(protolb.QueryTCloudListenerTargets)
	if err := cts.DecodeInto(req); err != nil {
		return nil, errf.NewFromErr(errf.DecodeRequestFailed, err)
	}

	if err := req.Validate(); err != nil {
		return nil, errf.NewFromErr(errf.InvalidParameter, err)
	}
	tcloud, err := svc.ad.TCloud(cts.Kit, req.AccountID)
	if err != nil {
		return nil, err
	}
	listOpt := &loadbalancer.TCloudListTargetsOption{
		Region:         req.Region,
		LoadBalancerId: req.LoadBalancerCloudId,
		ListenerIds:    req.ListenerCloudIDs,
		Protocol:       req.Protocol,
		Port:           req.Port,
	}
	return tcloud.ListTargets(cts.Kit, listOpt)
}
