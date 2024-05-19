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
	"fmt"

	typelb "hcm/pkg/adaptor/types/load-balancer"
	"hcm/pkg/criteria/errf"
	"hcm/pkg/kit"
	"hcm/pkg/logs"
	cvt "hcm/pkg/tools/converter"

	clb "github.com/tencentcloud/tencentcloud-sdk-go/tencentcloud/clb/v20180317"
	"github.com/tencentcloud/tencentcloud-sdk-go/tencentcloud/common"
)

// ListTargets 获取监听器后端绑定的机器列表信息.
// reference: https://cloud.tencent.com/document/api/214/30686
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

	resp, err := client.DescribeTargetsWithContext(kt.Ctx, req)
	if err != nil {
		logs.Errorf("list tcloud listener targets failed, req: %+v, err: %v, rid: %s", req, err, kt.Rid)
		return nil, err
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

// ListTargetGroup DescribeTargetGroupList 获取目标组列表 不返回关联关系
// https://cloud.tencent.com/document/api/214/40555
func (t *TCloudImpl) ListTargetGroup(kt *kit.Kit, opt *typelb.ListTargetGroupOption) ([]typelb.TargetGroup, error) {
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

	req := clb.NewDescribeTargetGroupListRequest()
	if len(opt.TargetGroupIds) != 0 {
		req.TargetGroupIds = common.StringPtrs(opt.TargetGroupIds)
	}

	if opt.Page != nil {
		req.Offset = cvt.ValToPtr(opt.Page.Offset)
		req.Limit = cvt.ValToPtr(opt.Page.Limit)
	}

	resp, err := client.DescribeTargetGroupListWithContext(kt.Ctx, req)
	if err != nil {
		logs.Errorf("list tcloud target group list failed, req: %+v, err: %v, rid: %s", req, err, kt.Rid)
		return nil, err
	}

	groups := make([]typelb.TargetGroup, 0, len(resp.Response.TargetGroupSet))
	for _, one := range resp.Response.TargetGroupSet {
		groups = append(groups, typelb.TargetGroup{TargetGroupInfo: one})
	}

	return groups, nil
}

// ListTargetGroupInstance DescribeTargetGroupInstances 获取目标组绑定的服务器
// https://cloud.tencent.com/document/api/214/40556
func (t *TCloudImpl) ListTargetGroupInstance(kt *kit.Kit, opt *typelb.ListTargetGroupInstanceOption) (
	[]typelb.TargetGroupBackend, error) {

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

	req := clb.NewDescribeTargetGroupInstancesRequest()
	if len(opt.TargetGroupIds) != 0 {
		req.Filters = append(req.Filters, &clb.Filter{
			Name:   cvt.ValToPtr("TargetGroupId"),
			Values: cvt.SliceToPtr(opt.TargetGroupIds),
		})
	}
	if len(opt.BindIPs) != 0 {
		req.Filters = append(req.Filters, &clb.Filter{
			Name:   cvt.ValToPtr("BindIP"),
			Values: cvt.SliceToPtr(opt.BindIPs),
		})
	}
	if len(opt.InstanceIds) != 0 {
		req.Filters = append(req.Filters, &clb.Filter{
			Name:   cvt.ValToPtr("InstanceId"),
			Values: cvt.SliceToPtr(opt.InstanceIds),
		})
	}

	if opt.Page != nil {
		req.Offset = cvt.ValToPtr(opt.Page.Offset)
		req.Limit = cvt.ValToPtr(opt.Page.Limit)
	}

	resp, err := client.DescribeTargetGroupInstancesWithContext(kt.Ctx, req)
	if err != nil {
		logs.Errorf("list tcloud target group list failed, req: %+v, err: %v, rid: %s", req, err, kt.Rid)
		return nil, err
	}

	groups := make([]typelb.TargetGroupBackend, 0, len(resp.Response.TargetGroupInstanceSet))
	for _, one := range resp.Response.TargetGroupInstanceSet {
		groups = append(groups, typelb.TargetGroupBackend{TargetGroupBackend: one})
	}

	return groups, nil
}
