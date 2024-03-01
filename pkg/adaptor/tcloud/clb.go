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

	"hcm/pkg/adaptor/poller"
	"hcm/pkg/adaptor/types"
	typeclb "hcm/pkg/adaptor/types/clb"
	"hcm/pkg/adaptor/types/core"
	"hcm/pkg/criteria/errf"
	"hcm/pkg/kit"
	"hcm/pkg/logs"
	"hcm/pkg/tools/converter"
	"hcm/pkg/tools/slice"

	clb "github.com/tencentcloud/tencentcloud-sdk-go/tencentcloud/clb/v20180317"
	"github.com/tencentcloud/tencentcloud-sdk-go/tencentcloud/common"
)

// ListClb list clb.
// reference: https://cloud.tencent.com/document/api/214/30685
func (t *TCloudImpl) ListClb(kt *kit.Kit, opt *typeclb.TCloudListOption) ([]typeclb.TCloudClb, error) {
	if opt == nil {
		return nil, errf.New(errf.InvalidParameter, "list option is required")
	}

	if err := opt.Validate(); err != nil {
		return nil, errf.NewFromErr(errf.InvalidParameter, err)
	}

	client, err := t.clientSet.ClbClient(opt.Region)
	if err != nil {
		return nil, fmt.Errorf("new tcloud clb client failed, region: %s, err: %v", opt.Region, err)
	}

	req := clb.NewDescribeLoadBalancersRequest()
	// 负载均衡实例ID。实例ID数量上限为20个
	if len(opt.CloudIDs) != 0 {
		req.LoadBalancerIds = common.StringPtrs(opt.CloudIDs)
		req.Limit = common.Int64Ptr(int64(core.TCloudQueryLimit))
	}

	if opt.Page != nil {
		req.Offset = common.Int64Ptr(int64(opt.Page.Offset))
		req.Limit = common.Int64Ptr(int64(opt.Page.Limit))
	}

	resp, err := client.DescribeLoadBalancersWithContext(kt.Ctx, req)
	if err != nil {
		logs.Errorf("list tcloud clb failed, req: %+v, err: %v, rid: %s", req, err, kt.Rid)
		return nil, err
	}

	clbs := make([]typeclb.TCloudClb, 0, len(resp.Response.LoadBalancerSet))
	for _, one := range resp.Response.LoadBalancerSet {
		clbs = append(clbs, typeclb.TCloudClb{LoadBalancer: one})
	}

	return clbs, nil
}

// CountClb count clb of region
// reference: https://cloud.tencent.com/document/api/214/30685
func (t *TCloudImpl) CountClb(kt *kit.Kit, region string) (int32, error) {
	client, err := t.clientSet.ClbClient(region)
	if err != nil {
		return 0, fmt.Errorf("new tcloud clb client failed, region: %s, err: %v", region, err)
	}

	req := clb.NewDescribeLoadBalancersRequest()
	req.Limit = converter.ValToPtr(int64(1))
	resp, err := client.DescribeLoadBalancersWithContext(kt.Ctx, req)
	if err != nil {
		logs.Errorf("count tcloud clb failed, region:%s, req: %+v, err: %v, rid: %s", region, req, err, kt.Rid)
		return 0, err
	}

	return int32(*resp.Response.TotalCount), nil
}

// ListListeners list listeners.
// reference: https://cloud.tencent.com/document/api/214/30686
func (t *TCloudImpl) ListListeners(kt *kit.Kit, opt *typeclb.TCloudListListenersOption) (
	[]typeclb.TCloudListeners, int32, error) {

	if opt == nil {
		return nil, 0, errf.New(errf.InvalidParameter, "list option is required")
	}

	if err := opt.Validate(); err != nil {
		return nil, 0, errf.NewFromErr(errf.InvalidParameter, err)
	}

	client, err := t.clientSet.ClbClient(opt.Region)
	if err != nil {
		return nil, 0, fmt.Errorf("new tcloud clb client failed, region: %s, err: %v", opt.Region, err)
	}

	req := clb.NewDescribeListenersRequest()
	req.LoadBalancerId = common.StringPtr(opt.LoadBalancerId)

	if len(opt.CloudIDs) != 0 {
		req.ListenerIds = common.StringPtrs(opt.CloudIDs)
	}

	if len(opt.Protocol) != 0 {
		req.Protocol = common.StringPtr(opt.Protocol)
	}

	if opt.Port > 0 {
		req.Port = common.Int64Ptr(opt.Port)
	}

	resp, err := client.DescribeListenersWithContext(kt.Ctx, req)
	if err != nil {
		logs.Errorf("list tcloud listeners failed, req: %+v, err: %v, rid: %s", req, err, kt.Rid)
		return nil, 0, err
	}

	listeners := make([]typeclb.TCloudListeners, 0, len(resp.Response.Listeners))
	for _, one := range resp.Response.Listeners {
		listeners = append(listeners, typeclb.TCloudListeners{Listener: one})
	}

	totalCount := int32(0)
	if resp != nil && resp.Response != nil && resp.Response.TotalCount != nil {
		totalCount = int32(*resp.Response.TotalCount)
	}

	return listeners, totalCount, nil
}

// ListTargets 获取监听器后端绑定的机器列表信息.
// reference: https://cloud.tencent.com/document/api/214/30686
func (t *TCloudImpl) ListTargets(kt *kit.Kit, opt *typeclb.TCloudListTargetsOption) (
	[]typeclb.TCloudListenerTargets, error) {

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

	if len(opt.CloudIDs) != 0 {
		req.ListenerIds = common.StringPtrs(opt.CloudIDs)
	}

	if len(opt.Protocol) != 0 {
		req.Protocol = common.StringPtr(opt.Protocol)
	}

	if opt.Port > 0 {
		req.Port = common.Int64Ptr(opt.Port)
	}

	resp, err := client.DescribeTargetsWithContext(kt.Ctx, req)
	if err != nil {
		logs.Errorf("list tcloud listener targets failed, req: %+v, err: %v, rid: %s", req, err, kt.Rid)
		return nil, err
	}

	listeners := make([]typeclb.TCloudListenerTargets, 0, len(resp.Response.Listeners))
	for _, one := range resp.Response.Listeners {
		listeners = append(listeners, typeclb.TCloudListenerTargets{ListenerBackend: one})
	}

	return listeners, nil
}

// CreateClb reference: https://cloud.tencent.com/document/api/214/30692
// NOTE：返回实例`ID`列表并不代表实例创建成功，可根据 [DescribeLoadBalancers](https://cloud.tencent.com/document/api/214/30685)
// 接口查询返回的LoadBalancerSet中对应实例的`ID`的状态来判断创建是否完成；如果实例状态由“0(创建中)”变为“1(正常运行)”，则为创建成功。
func (t *TCloudImpl) CreateClb(kt *kit.Kit, opt *typeclb.TCloudCreateClbOption) (*poller.BaseDoneResult, error) {
	if opt == nil {
		return nil, errf.New(errf.InvalidParameter, "create option is required")
	}

	if err := opt.Validate(); err != nil {
		return nil, errf.NewFromErr(errf.InvalidParameter, err)
	}

	client, err := t.clientSet.ClbClient(opt.Region)
	if err != nil {
		return nil, fmt.Errorf("init tencent cloud clb client failed, region: %s, err: %v", opt.Region, err)
	}

	req := t.formatCreateClbRequest(opt)

	resp, err := client.CreateLoadBalancer(req)
	if err != nil {
		logs.Errorf("run tencent cloud clb instance failed, req: %+v, err: %v, rid: %s", req, err, kt.Rid)
		return nil, err
	}

	handler := &createClbPollingHandler{
		opt.Region,
	}
	respPoller := poller.Poller[*TCloudImpl, []typeclb.TCloudClb, poller.BaseDoneResult]{Handler: handler}
	result, err := respPoller.PollUntilDone(t, kt, resp.Response.LoadBalancerIds, types.NewBatchCreateClbPollerOption())
	if err != nil {
		return nil, err
	}

	return result, nil
}

func (t *TCloudImpl) formatCreateClbRequest(opt *typeclb.TCloudCreateClbOption) *clb.CreateLoadBalancerRequest {
	req := clb.NewCreateLoadBalancerRequest()
	// 负载均衡实例的名称
	req.LoadBalancerName = common.StringPtr(opt.LoadBalancerName)
	// 负载均衡实例的网络类型。OPEN：公网属性， INTERNAL：内网属性。
	req.LoadBalancerType = common.StringPtr(string(opt.LoadBalancerType))
	// 仅适用于公网负载均衡, IP版本
	req.AddressIPVersion = common.StringPtr(string(opt.AddressIPVersion))
	// 负载均衡后端目标设备所属的网络
	req.VpcId = common.StringPtr(opt.VpcID)
	// 负载均衡实例的类型。1：通用的负载均衡实例，目前只支持传入1。
	req.Forward = common.Int64Ptr(int64(typeclb.DefaultLoadBalancerInstType))
	// 是否支持绑定跨地域/跨Vpc绑定IP的功能
	req.SnatPro = common.BoolPtr(opt.SnatPro)
	// Target是否放通来自CLB的流量。开启放通（true）：只验证CLB上的安全组；不开启放通（false）：需同时验证CLB和后端实例上的安全组
	req.LoadBalancerPassToTarget = common.BoolPtr(opt.LoadBalancerPassToTarget)
	// 是否创建域名化负载均衡
	req.DynamicVip = common.BoolPtr(opt.DynamicVip)
	req.SubnetId = common.StringPtr(opt.SubnetID)
	req.Vip = common.StringPtr(opt.Vip)
	req.Number = common.Uint64Ptr(opt.Number)
	req.ProjectId = common.Int64Ptr(opt.ProjectID)
	req.SlaType = common.StringPtr(opt.SlaType)
	req.ClusterIds = append(req.ClusterIds, opt.ClusterIds...)
	// 用于保证请求幂等性的字符串。该字符串由客户生成，需保证不同请求之间唯一，最大值不超过64个字符。若不指定该参数则无法保证请求的幂等性。
	req.ClientToken = common.StringPtr(opt.ClientToken)
	req.ClusterTag = common.StringPtr(opt.ClusterTag)
	req.EipAddressId = common.StringPtr(opt.EipAddressID)
	req.SlaveZoneId = common.StringPtr(opt.SlaveZoneID)
	req.Egress = common.StringPtr(opt.Egress)
	req.MasterZoneId = common.StringPtr(opt.MasterZoneID)
	req.ZoneId = common.StringPtr(opt.ZoneID)
	req.VipIsp = common.StringPtr(opt.VipIsp)
	req.BandwidthPackageId = common.StringPtr(opt.BandwidthPackageID)
	req.Tags = opt.Tags
	req.SnatIps = opt.SnatIps

	if opt.InternetAccessible != nil {
		req.InternetAccessible = &clb.InternetAccessible{
			InternetChargeType:      opt.InternetAccessible.InternetChargeType,
			InternetMaxBandwidthOut: opt.InternetAccessible.InternetMaxBandwidthOut,
			BandwidthpkgSubType:     opt.InternetAccessible.BandwidthpkgSubType,
		}
	}

	if opt.ExclusiveCluster != nil {
		req.ExclusiveCluster = &clb.ExclusiveCluster{
			L4Clusters:       opt.ExclusiveCluster.L4Clusters,
			L7Clusters:       opt.ExclusiveCluster.L7Clusters,
			ClassicalCluster: opt.ExclusiveCluster.ClassicalCluster,
		}
	}

	return req
}

var _ poller.PollingHandler[*TCloudImpl, []typeclb.TCloudClb, poller.BaseDoneResult] = new(createClbPollingHandler)

type createClbPollingHandler struct {
	region string
}

// Done ...
func (h *createClbPollingHandler) Done(clbs []typeclb.TCloudClb) (bool, *poller.BaseDoneResult) {
	result := &poller.BaseDoneResult{
		SuccessCloudIDs: make([]string, 0),
		FailedCloudIDs:  make([]string, 0),
		UnknownCloudIDs: make([]string, 0),
	}
	flag := true
	for _, item := range clbs {
		// 不是[正常运行]的状态
		if converter.PtrToVal(item.Status) != uint64(typeclb.SuccessStatus) {
			flag = false
			result.FailedCloudIDs = append(result.FailedCloudIDs, *item.LoadBalancerId)
			continue
		}

		result.SuccessCloudIDs = append(result.SuccessCloudIDs, *item.LoadBalancerId)
	}

	return flag, result
}

// Poll ...
func (h *createClbPollingHandler) Poll(client *TCloudImpl, kt *kit.Kit, cloudIDs []*string) (
	[]typeclb.TCloudClb, error) {

	// 负载均衡实例ID。实例ID数量上限为20个
	cloudIDSplit := slice.Split(cloudIDs, 20)

	clbs := make([]typeclb.TCloudClb, 0, len(cloudIDs))
	for idx, partIDs := range cloudIDSplit {
		opt := &typeclb.TCloudListOption{
			Region:   h.region,
			CloudIDs: converter.PtrToSlice(partIDs),
			Page: &core.TCloudPage{
				Offset: uint64(idx),
				Limit:  uint64(core.TCloudQueryLimit),
			},
		}
		resp, err := client.ListClb(kt, opt)
		if err != nil {
			return nil, err
		}

		clbs = append(clbs, resp...)
	}

	if len(clbs) != len(cloudIDs) {
		return nil, fmt.Errorf("batch query clb count: %d not equal return count: %d", len(cloudIDs), len(clbs))
	}

	return clbs, nil
}

// SetClbSecurityGroups reference: https://cloud.tencent.com/document/api/214/34903
func (t *TCloudImpl) SetClbSecurityGroups(kt *kit.Kit, opt *typeclb.TCloudSetClbSecurityGroupOption) (
	*clb.SetLoadBalancerSecurityGroupsResponseParams, error) {

	if opt == nil {
		return nil, errf.New(errf.InvalidParameter, "set clb security group option is required")
	}

	if err := opt.Validate(); err != nil {
		return nil, errf.NewFromErr(errf.InvalidParameter, err)
	}

	client, err := t.clientSet.ClbClient(opt.Region)
	if err != nil {
		return nil, fmt.Errorf("init tencent cloud clb client failed, region: %s, err: %v", opt.Region, err)
	}

	req := clb.NewSetLoadBalancerSecurityGroupsRequest()
	req.LoadBalancerId = common.StringPtr(opt.LoadBalancerID)
	if len(opt.SecurityGroups) > 0 {
		req.SecurityGroups = common.StringPtrs(opt.SecurityGroups)
	}

	resp, err := client.SetLoadBalancerSecurityGroupsWithContext(kt.Ctx, req)
	if err != nil {
		logs.Errorf("run tencent cloud clb set security group failed, opt: %+v, err: %v, rid: %s", opt, err, kt.Rid)
		return nil, err
	}

	return resp.Response, nil
}
