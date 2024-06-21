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

package loadbalancer

import (
	"fmt"

	typeslb "hcm/pkg/adaptor/types/load-balancer"
	protolb "hcm/pkg/api/hc-service/load-balancer"
	"hcm/pkg/criteria/errf"
	"hcm/pkg/logs"
	"hcm/pkg/rest"
)

// TCloudCreateSnatIps 创建跨域2.0 snat ip， 如果没有开启跨域2.0会自动开启
func (svc *clbSvc) TCloudCreateSnatIps(cts *rest.Contexts) (any, error) {

	req := new(protolb.TCloudCreateSnatIpReq)
	if err := cts.DecodeInto(req); err != nil {
		return nil, errf.NewFromErr(errf.DecodeRequestFailed, err)
	}

	if err := req.Validate(); err != nil {
		return nil, errf.NewFromErr(errf.InvalidParameter, err)
	}

	adpt, err := svc.ad.TCloud(cts.Kit, req.AccountID)
	if err != nil {
		return nil, err
	}
	opt := &typeslb.TCloudCreateSnatIpOpt{
		Region:         req.Region,
		LoadBalancerId: req.LoadBalancerCloudId,
		SnatIps:        req.SnatIPs,
	}
	createErr := adpt.CreateLoadBalancerSnatIps(cts.Kit, opt)
	if createErr != nil {
		logs.Errorf("fail to call tcloud to create snat ip, err: %v, req: %+v, rid: %s", createErr, req, cts.Kit.Rid)
		// 添加失败也同步
	}

	if err := svc.lbSync(cts.Kit, adpt, req.AccountID, req.Region, []string{req.LoadBalancerCloudId}); err != nil {
		logs.Errorf("fail to sync load balancer for create snat ip, rid: %s", cts.Kit.Rid)
		// 添加成功情况下才返回同步的错误
		if createErr == nil {
			return nil, fmt.Errorf("fail to sync load balancer for create snat ip, err: %w", err)
		}
	}
	return nil, createErr
}

// TCloudDeleteSnatIps 删除snat ip
func (svc *clbSvc) TCloudDeleteSnatIps(cts *rest.Contexts) (any, error) {
	req := new(protolb.TCloudDeleteSnatIpReq)
	if err := cts.DecodeInto(req); err != nil {
		return nil, errf.NewFromErr(errf.DecodeRequestFailed, err)
	}

	if err := req.Validate(); err != nil {
		return nil, errf.NewFromErr(errf.InvalidParameter, err)
	}

	adpt, err := svc.ad.TCloud(cts.Kit, req.AccountID)
	if err != nil {
		return nil, err
	}
	opt := &typeslb.TCloudDeleteSnatIpOpt{
		Region:         req.Region,
		LoadBalancerId: req.LoadBalancerCloudId,
		Ips:            req.Ips,
	}
	delErr := adpt.DeleteLoadBalancerSnatIps(cts.Kit, opt)
	if delErr != nil {
		logs.Errorf("fail to call tcloud to delete snat ip, err: %v, req: %+v, rid: %s", delErr, req, cts.Kit.Rid)
		// 删除失败也同步
	}
	if err := svc.lbSync(cts.Kit, adpt, req.AccountID, req.Region, []string{req.LoadBalancerCloudId}); err != nil {
		logs.Errorf("fail to sync load balancer for delete snat ip, rid: %s", cts.Kit.Rid)
		// 删除成功情况下才返回同步的错误
		if delErr == nil {
			return nil, fmt.Errorf("fail to sync load balancer for delete snat ip, err: %w", err)
		}
	}
	return nil, delErr
}
