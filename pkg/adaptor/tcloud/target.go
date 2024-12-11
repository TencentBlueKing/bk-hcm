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

package tcloud

import (
	"errors"
	"fmt"

	typelb "hcm/pkg/adaptor/types/load-balancer"
	"hcm/pkg/criteria/errf"
	"hcm/pkg/kit"
	"hcm/pkg/logs"

	clb "github.com/tencentcloud/tencentcloud-sdk-go/tencentcloud/clb/v20180317"
	"github.com/tencentcloud/tencentcloud-sdk-go/tencentcloud/common"
)

// ListTargets 获取监听器后端绑定的机器列表信息.
// reference: https://cloud.tencent.com/document/api/214/30684
func (t *TCloudImpl) ListTargets(kt *kit.Kit, opt *typelb.TCloudListTargetsOption) (
	[]typelb.TCloudListenerTarget, error) {

	if opt == nil {
		return nil, errf.New(errf.InvalidParameter, "list option is required")
	}

	if err := opt.Validate(); err != nil {
		return nil, errf.NewFromErr(errf.InvalidParameter, err)
	}

	client, err := t.clientSet.ClbClient(opt.Region)
	if err != nil {
		return nil, fmt.Errorf("new tcloud targets client failed, region: %s, err: %v", opt.Region, err)
	}

	req := clb.NewDescribeTargetsRequest()
	req.LoadBalancerId = common.StringPtr(opt.LoadBalancerId)

	if len(opt.ListenerIds) != 0 {
		req.ListenerIds = common.StringPtrs(opt.ListenerIds)
	}

	if len(opt.Protocol) != 0 {
		req.Protocol = common.StringPtr(string(opt.Protocol))
	}

	if opt.Port > 0 {
		req.Port = common.Int64Ptr(opt.Port)
	}

	resp, err := NetworkErrRetry(client.DescribeTargetsWithContext, kt, req)
	if err != nil {
		logs.Errorf("fail to describe clb targets from tcloud, err: %v, req: %+v, rid: %s", err, req, kt.Rid)
		return nil, err
	}

	if resp == nil || resp.Response == nil {
		return nil, errors.New("empty target response from tcloud")
	}
	listeners := make([]typelb.TCloudListenerTarget, 0, len(resp.Response.Listeners))
	for _, one := range resp.Response.Listeners {
		listeners = append(listeners, typelb.TCloudListenerTarget{ListenerBackend: one})
	}

	return listeners, nil
}

// ListTargetHealth 获取负载均衡后端服务的健康检查状态
// reference: https://cloud.tencent.com/document/api/214/34898
func (t *TCloudImpl) ListTargetHealth(kt *kit.Kit, opt *typelb.TCloudListTargetHealthOption) (
	[]typelb.TCloudTargetHealth, error) {

	if opt == nil {
		return nil, errf.New(errf.InvalidParameter, "list option is required")
	}

	if err := opt.Validate(); err != nil {
		return nil, errf.NewFromErr(errf.InvalidParameter, err)
	}

	client, err := t.clientSet.ClbClient(opt.Region)
	if err != nil {
		return nil, fmt.Errorf("new tcloud targets client failed, region: %s, err: %v", opt.Region, err)
	}

	req := clb.NewDescribeTargetHealthRequest()
	req.LoadBalancerIds = common.StringPtrs(opt.LoadBalancerIDs)

	resp, err := client.DescribeTargetHealthWithContext(kt.Ctx, req)
	if err != nil {
		logs.Errorf("list tcloud listener targets health failed, req: %+v, err: %v, rid: %s", req, err, kt.Rid)
		return nil, err
	}

	healths := make([]typelb.TCloudTargetHealth, 0, len(resp.Response.LoadBalancers))
	for _, one := range resp.Response.LoadBalancers {
		healths = append(healths, typelb.TCloudTargetHealth{LoadBalancerHealth: one})
	}

	return healths, nil
}
