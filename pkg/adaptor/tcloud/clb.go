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

	typeclb "hcm/pkg/adaptor/types/clb"
	"hcm/pkg/adaptor/types/core"
	"hcm/pkg/criteria/errf"
	"hcm/pkg/kit"
	"hcm/pkg/logs"
	cvt "hcm/pkg/tools/converter"
	"hcm/pkg/tools/slice"

	clb "github.com/tencentcloud/tencentcloud-sdk-go/tencentcloud/clb/v20180317"
)

// CreateLoadBalancer 购买负载均衡实例 https://cloud.tencent.com/document/api/214/30692
func (t *TCloudImpl) CreateLoadBalancer(kt *kit.Kit, opt *typeclb.TCloudCLBCreateOpt) (
	clbIDList []string, err error) {

	if opt == nil {
		return nil, errf.New(errf.InvalidParameter, "create clb option is required")
	}

	if err := opt.Validate(); err != nil {
		return nil, errf.NewFromErr(errf.InvalidParameter, err)
	}
	createReq := clb.NewCreateLoadBalancerRequest()
	createReq.LoadBalancerName = cvt.ValToPtr(opt.Name)
	createReq.LoadBalancerType = cvt.ValToPtr(string(opt.LBType))
	createReq.VpcId = cvt.ValToPtr(opt.CloudVpcID)
	createReq.SubnetId = cvt.ValToPtr(opt.SubnetID)
	createReq.AddressIPVersion = (*string)(opt.AddressIPVersion)
	if opt.InternetMaxBandwidthOut != nil || opt.InternetChargeType != nil {
		createReq.InternetAccessible = &clb.InternetAccessible{
			InternetChargeType:      opt.InternetChargeType,
			InternetMaxBandwidthOut: opt.InternetMaxBandwidthOut,
		}
	}

	createReq.BandwidthPackageId = opt.BandwidthPackageId
	createReq.SlaType = opt.SlaType
	createReq.Vip = opt.Vip
	createReq.VipIsp = opt.VipIsp
	createReq.LoadBalancerPassToTarget = opt.LoadBalancerPassToTarget

	client, err := t.clientSet.clbClient(opt.Region)
	if err != nil {
		return nil, fmt.Errorf("new tcloud vpc client failed, err: %v", err)
	}

	created, err := client.CreateLoadBalancerWithContext(kt.Ctx, createReq)
	if err != nil {
		logs.Errorf("create tcloud clb failed, err: %v, rid: %s", err, kt.Rid)
		return nil, err
	}

	return slice.Map(created.Response.LoadBalancerIds, cvt.PtrToVal[string]), nil
}

// ListLoadBalancer 列出clb 列表
// https://cloud.tencent.com/document/api/214/30685
func (t *TCloudImpl) ListLoadBalancer(kt *kit.Kit, opt *typeclb.TCloudCLBListOpt) (clbList []typeclb.TCloudCLB,
	err error) {

	if opt == nil {
		return nil, errf.New(errf.InvalidParameter, "list clb option is required")
	}

	if err := opt.Validate(); err != nil {
		return nil, errf.NewFromErr(errf.InvalidParameter, err)
	}

	client, err := t.clientSet.clbClient(opt.Region)
	if err != nil {
		return nil, fmt.Errorf("new tcloud vpc client failed, err: %v", err)
	}

	req := clb.NewDescribeLoadBalancersRequest()
	if len(opt.CloudIDs) != 0 {
		req.LoadBalancerIds = cvt.SliceToPtr(opt.CloudIDs)
		req.Limit = cvt.ValToPtr(int64(core.TCloudQueryLimit))
	}

	if opt.Page != nil {
		req.Offset = cvt.ValToPtr(int64(opt.Page.Offset))
		req.Limit = cvt.ValToPtr(int64(opt.Page.Limit))
	}

	resp, err := client.DescribeLoadBalancersWithContext(kt.Ctx, req)
	if err != nil {
		logs.Errorf("list tcloud clb failed, err: %v, rid: %s", err, kt.Rid)
		return nil, err
	}

	clbs := make([]typeclb.TCloudCLB, 0, len(resp.Response.LoadBalancerSet))
	for _, one := range resp.Response.LoadBalancerSet {
		clbs = append(clbs, typeclb.TCloudCLB{LoadBalancer: one})
	}

	return clbs, nil
}

// DeleteLoadBalancer 删除CLB
// https://cloud.tencent.com/document/api/214/30689
func (t *TCloudImpl) DeleteLoadBalancer(kt *kit.Kit, opt *typeclb.TCloudCLBListOpt) (err error) {

	if opt == nil {
		return errf.New(errf.InvalidParameter, "list clb option is required")
	}

	if err := opt.Validate(); err != nil {
		return errf.NewFromErr(errf.InvalidParameter, err)
	}

	client, err := t.clientSet.clbClient(opt.Region)
	if err != nil {
		return fmt.Errorf("new tcloud vpc client failed, err: %v", err)
	}

	req := clb.NewDeleteLoadBalancerRequest()
	req.LoadBalancerIds = cvt.SliceToPtr(opt.CloudIDs)

	if _, err = client.DeleteLoadBalancerWithContext(kt.Ctx, req); err != nil {
		logs.Errorf("delete tcloud clb failed, err: %v, rid: %s", err, kt.Rid)
		return err
	}
	return nil
}

// ListListeners 列出clb下的监听器列表
// https://cloud.tencent.com/document/api/214/30686
func (t *TCloudImpl) ListListeners(kt *kit.Kit, opt *typeclb.TCloudListenerListOpt) (
	listeners []typeclb.TCloudListener, err error) {

	if opt == nil {
		return nil, errf.New(errf.InvalidParameter, "list listener option is required")
	}

	if err := opt.Validate(); err != nil {
		return nil, errf.NewFromErr(errf.InvalidParameter, err)
	}

	client, err := t.clientSet.clbClient(opt.Region)
	if err != nil {
		return nil, fmt.Errorf("new tcloud vpc client failed, err: %v", err)
	}

	req := clb.NewDescribeListenersRequest()
	if len(opt.ListenerIds) > 0 {
		req.ListenerIds = cvt.SliceToPtr(opt.ListenerIds)
	}
	req.LoadBalancerId = cvt.ValToPtr(opt.ClbID)
	resp, err := client.DescribeListenersWithContext(kt.Ctx, req)
	if err != nil {
		logs.Errorf("list tcloud clb failed, err: %v, rid: %s", err, kt.Rid)
		return nil, err
	}

	listeners = slice.Map(resp.Response.Listeners, func(one *clb.Listener) typeclb.TCloudListener {
		return typeclb.TCloudListener{Listener: one}
	})

	return listeners, nil
}

// ListTargets 查询负载均衡绑定的后端服务列表
// https://cloud.tencent.com/document/api/214/30684
func (t *TCloudImpl) ListTargets(kt *kit.Kit, opt *typeclb.TCloudTargetListOpt) (
	targets []typeclb.TCloudListenerBackend, err error) {

	if opt == nil {
		return nil, errf.New(errf.InvalidParameter, "list clb option is required")
	}

	if err := opt.Validate(); err != nil {
		return nil, errf.NewFromErr(errf.InvalidParameter, err)
	}

	client, err := t.clientSet.clbClient(opt.Region)
	if err != nil {
		return nil, fmt.Errorf("new tcloud vpc client failed, err: %v", err)
	}

	req := clb.NewDescribeTargetsRequest()
	if len(opt.ListenerIds) > 0 {
		req.ListenerIds = cvt.SliceToPtr(opt.ListenerIds)
	}
	req.LoadBalancerId = cvt.ValToPtr(opt.ClbID)
	resp, err := client.DescribeTargetsWithContext(kt.Ctx, req)
	if err != nil {
		logs.Errorf("list tcloud clb failed, err: %v, rid: %s", err, kt.Rid)
		return nil, err
	}

	targets = slice.Map(resp.Response.Listeners, func(one *clb.ListenerBackend) typeclb.TCloudListenerBackend {
		return typeclb.TCloudListenerBackend{ListenerBackend: one}
	})

	return targets, nil
}
