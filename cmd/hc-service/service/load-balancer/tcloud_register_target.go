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
	"errors"
	"fmt"

	typeslb "hcm/pkg/adaptor/types/load-balancer"
	"hcm/pkg/api/core"
	protolb "hcm/pkg/api/hc-service/load-balancer"
	"hcm/pkg/criteria/enumor"
	"hcm/pkg/criteria/errf"
	"hcm/pkg/dal/dao/tools"
	"hcm/pkg/logs"
	"hcm/pkg/rest"
	cvt "hcm/pkg/tools/converter"
)

// RegisterTargetToListenerRule 批量到云上注册RS, 对接adaptor, 无db操作
func (svc *clbSvc) RegisterTargetToListenerRule(cts *rest.Contexts) (any, error) {
	lbID := cts.PathParameter("lb_id").String()
	if len(lbID) == 0 {
		return nil, errf.New(errf.InvalidParameter, "lb_id is required")
	}

	req := new(protolb.BatchRegisterTCloudTargetReq)
	if err := cts.DecodeInto(req); err != nil {
		return nil, errf.NewFromErr(errf.DecodeRequestFailed, err)
	}

	if err := req.Validate(); err != nil {
		return nil, errf.NewFromErr(errf.InvalidParameter, err)
	}
	// 获取负载均衡信息
	lbResp, err := svc.dataCli.Global.LoadBalancer.ListLoadBalancer(cts.Kit, &core.ListReq{
		Filter: tools.EqualExpression("id", lbID),
		Page:   core.NewDefaultBasePage(),
	})
	if err != nil {
		logs.Errorf("fail to list find load balancer, err: %v, rid: %s", err, cts.Kit.Rid)
		return nil, err
	}
	if len(lbResp.Details) < 1 {
		return nil, errf.New(errf.RecordNotFound, "lb not found")
	}
	lb := lbResp.Details[0]

	adpt, err := svc.ad.TCloud(cts.Kit, lb.AccountID)
	if err != nil {
		return nil, err
	}

	opt := &typeslb.TCloudRegisterTargetsOption{
		Region:         lb.Region,
		LoadBalancerId: lb.CloudID,
		Targets:        make([]*typeslb.BatchTarget, 0, len(req.Targets)),
	}
	for _, target := range req.Targets {
		tmpRs := &typeslb.BatchTarget{
			ListenerId: cvt.ValToPtr(req.CloudListenerID),
			Port:       cvt.ValToPtr(target.Port),
			Type:       cvt.ValToPtr(string(target.TargetType)),
			Weight:     cvt.ValToPtr(target.Weight),
		}
		switch target.TargetType {
		case enumor.CvmInstType:
			tmpRs.InstanceId = cvt.ValToPtr(target.CloudInstID)
		case enumor.EniInstType:
			// 跨域rs 通过指定为 ip
			tmpRs.EniIp = cvt.ValToPtr(target.EniIp)
		default:
			return nil, errors.New(string("invalid target type: " + target.TargetType))
		}
		// 只有七层规则才需要传该参数
		if req.RuleType == enumor.Layer7RuleType {
			tmpRs.LocationId = cvt.ValToPtr(req.CloudRuleID)
		}
		opt.Targets = append(opt.Targets, tmpRs)
	}
	failLblIds, err := adpt.RegisterTargets(cts.Kit, opt)
	if err != nil {
		logs.Errorf("fail to register rs, err: %v, rid: %s", err, cts.Kit.Rid)
		return nil, err
	}
	if len(failLblIds) > 0 {
		return nil, fmt.Errorf("some listener fail to bind: %v", failLblIds)
	}
	return nil, nil
}
