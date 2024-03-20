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
	"errors"
	"fmt"

	"hcm/pkg/adaptor/poller"
	"hcm/pkg/adaptor/types"
	"hcm/pkg/adaptor/types/core"
	typelb "hcm/pkg/adaptor/types/load-balancer"
	"hcm/pkg/criteria/errf"
	"hcm/pkg/kit"
	"hcm/pkg/logs"
	"hcm/pkg/tools/converter"

	clb "github.com/tencentcloud/tencentcloud-sdk-go/tencentcloud/clb/v20180317"
	"github.com/tencentcloud/tencentcloud-sdk-go/tencentcloud/common"
)

// ListLoadBalancer 查询clb列表：如果指定的LoadBalancerIds不存在，该接口不会报错
// reference: https://cloud.tencent.com/document/api/214/30685
func (t *TCloudImpl) ListLoadBalancer(kt *kit.Kit, opt *typelb.TCloudListOption) ([]typelb.TCloudClb, error) {
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

	clbs := make([]typelb.TCloudClb, 0, len(resp.Response.LoadBalancerSet))
	for _, one := range resp.Response.LoadBalancerSet {
		clbs = append(clbs, typelb.TCloudClb{LoadBalancer: one})
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
func (t *TCloudImpl) ListListeners(kt *kit.Kit, opt *typelb.TCloudListListenersOption) (
	[]typelb.TCloudListener, int32, error) {

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
		req.Protocol = common.StringPtr(string(opt.Protocol))
	}

	if opt.Port > 0 {
		req.Port = common.Int64Ptr(opt.Port)
	}

	resp, err := client.DescribeListenersWithContext(kt.Ctx, req)
	if err != nil {
		logs.Errorf("list tcloud listeners failed, req: %+v, err: %v, rid: %s", req, err, kt.Rid)
		return nil, 0, err
	}

	listeners := make([]typelb.TCloudListener, 0, len(resp.Response.Listeners))
	for _, one := range resp.Response.Listeners {
		listeners = append(listeners, typelb.TCloudListener{Listener: one})
	}

	totalCount := int32(0)
	if resp != nil && resp.Response != nil && resp.Response.TotalCount != nil {
		totalCount = int32(*resp.Response.TotalCount)
	}

	return listeners, totalCount, nil
}

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

	if len(opt.CloudIDs) != 0 {
		req.ListenerIds = common.StringPtrs(opt.CloudIDs)
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

// CreateLoadBalancer reference: https://cloud.tencent.com/document/api/214/30692
// 如果创建成功返回对应clb id, 需要检查对应的`SuccessCloudIDs`参数。
func (t *TCloudImpl) CreateLoadBalancer(kt *kit.Kit, opt *typelb.TCloudCreateClbOption) (
	*poller.BaseDoneResult, error) {
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

	createResp, err := client.CreateLoadBalancerWithContext(kt.Ctx, req)
	if err != nil {
		logs.Errorf("create tencent cloud clb instance failed, req: %+v, err: %v, rid: %s", req, err, kt.Rid)
		return nil, err
	}
	/*
		NOTE：云上接口`CreateLoadBalancer`返回实例`ID`列表并不代表实例创建成功。`CreateLoadBalancer`接口文档声称可根据
		[DescribeLoadBalancers](https://cloud.tencent.com/document/api/214/30685)接口返回的`LoadBalancerSet`中
		对应实例的`ID`的状态来判断创建是否完成：如果实例状态由“0(创建中)”变为“1(正常运行)”，则为创建成功。
		但是实际上对于创建失败的任务使用`DescribeLoadBalancers`接口无法判断，该情况并不会返回错误，只会静默返回空值。
		因此，用`DescribeLoadBalancers`这个接口难以确定是创建时间过长还是创建失败。
		这里通过`DescribeTaskStatus`接口查询对应CLB创建任务状态，该接口可以明确创建失败。
		具体实现参考`createClbPollingHandler`中 `Poll`和`Done`方法的实现。
	*/

	respPoller := poller.Poller[*TCloudImpl, map[string]*clb.DescribeTaskStatusResponseParams, poller.BaseDoneResult]{
		Handler: &createClbPollingHandler{opt.Region},
	}

	reqID := createResp.Response.RequestId
	result, err := respPoller.PollUntilDone(t, kt, []*string{reqID}, types.NewBatchCreateClbPollerOption())
	if err != nil {
		return nil, err
	}
	if len(result.SuccessCloudIDs) == 0 {
		return nil, errf.Newf(errf.CloudVendorError,
			"no any lb being created, TencentCloudSDK RequestId: %s", converter.PtrToVal(reqID))
	}
	return result, nil
}

// DescribeResources 查询用户在当前地域支持可用区列表和资源列表
// https://cloud.tencent.com/document/api/214/70213
func (t *TCloudImpl) DescribeResources(kt *kit.Kit, opt *typelb.TCloudDescribeResourcesOption) (
	*clb.DescribeResourcesResponseParams, error) {

	if opt == nil {
		return nil, errf.New(errf.InvalidParameter, "describe resource option can not be nil")
	}

	if err := opt.Validate(); err != nil {
		return nil, errf.NewFromErr(errf.InvalidParameter, err)
	}

	client, err := t.clientSet.ClbClient(opt.Region)
	if err != nil {
		return nil, fmt.Errorf("init tencent cloud clb client failed, region: %s, err: %v", opt.Region, err)
	}
	req := clb.NewDescribeResourcesRequest()
	if len(opt.MasterZones) != 0 {
		req.Filters = append(req.Filters, &clb.Filter{
			Name:   common.StringPtr("master-zone"),
			Values: common.StringPtrs(opt.MasterZones),
		})
	}
	if len(opt.ISP) != 0 {
		req.Filters = append(req.Filters, &clb.Filter{
			Name:   common.StringPtr("isp"),
			Values: common.StringPtrs(opt.ISP),
		})
	}
	if len(opt.IPVersion) != 0 {
		req.Filters = append(req.Filters, &clb.Filter{
			Name:   common.StringPtr("ip-version"),
			Values: common.StringPtrs(opt.IPVersion),
		})
	}

	req.Limit = opt.Limit
	req.Offset = opt.Offset

	resp, err := client.DescribeResourcesWithContext(kt.Ctx, req)
	if err != nil {
		logs.Errorf("tencent cloud describe resources failed, req: %+v, err: %v, rid: %s", req, err, kt.Rid)
		return nil, err
	}
	return resp.Response, nil
}

func (t *TCloudImpl) formatCreateClbRequest(opt *typelb.TCloudCreateClbOption) *clb.CreateLoadBalancerRequest {
	req := clb.NewCreateLoadBalancerRequest()
	// 负载均衡实例的名称
	req.LoadBalancerName = opt.LoadBalancerName
	// 负载均衡实例的网络类型。OPEN：公网属性， INTERNAL：内网属性。
	req.LoadBalancerType = common.StringPtr(string(opt.LoadBalancerType))
	// 仅适用于公网负载均衡, IP版本
	req.AddressIPVersion = (*string)(opt.AddressIPVersion)
	// 负载均衡后端目标设备所属的网络
	req.VpcId = opt.VpcID
	// 负载均衡实例的类型。1：通用的负载均衡实例，目前只支持传入1。
	req.Forward = common.Int64Ptr(int64(typelb.DefaultLoadBalancerInstType))
	// 是否支持绑定跨地域/跨Vpc绑定IP的功能
	req.SnatPro = opt.SnatPro
	// Target是否放通来自CLB的流量。开启放通（true）：只验证CLB上的安全组；不开启放通（false）：需同时验证CLB和后端实例上的安全组
	req.LoadBalancerPassToTarget = opt.LoadBalancerPassToTarget
	// 是否创建域名化负载均衡
	req.DynamicVip = opt.DynamicVip
	req.SubnetId = opt.SubnetID
	req.Vip = opt.Vip
	req.Number = opt.Number
	req.ProjectId = opt.ProjectID
	req.SlaType = opt.SlaType
	req.ClusterIds = append(req.ClusterIds, opt.ClusterIds...)
	// 用于保证请求幂等性的字符串。该字符串由客户生成，需保证不同请求之间唯一，最大值不超过64个字符。若不指定该参数则无法保证请求的幂等性。
	req.ClientToken = opt.ClientToken
	req.ClusterTag = opt.ClusterTag
	req.EipAddressId = opt.EipAddressID
	req.SlaveZoneId = opt.SlaveZoneID
	req.Egress = opt.Egress
	req.ZoneId = opt.ZoneID
	req.MasterZoneId = opt.MasterZoneID

	req.BandwidthPackageId = opt.BandwidthPackageID
	req.Tags = opt.Tags
	req.SnatIps = opt.SnatIps

	// 使用默认ISP时传递空即可
	ispVal := converter.PtrToVal(opt.VipIsp)
	if ispVal != "" && ispVal != typelb.TCloudDefaultISP {
		req.VipIsp = opt.VipIsp
	}

	if opt.InternetChargeType != nil || opt.InternetMaxBandwidthOut != nil {
		req.InternetAccessible = &clb.InternetAccessible{
			InternetChargeType:      opt.InternetChargeType,
			InternetMaxBandwidthOut: opt.InternetMaxBandwidthOut,
			BandwidthpkgSubType:     opt.BandwidthpkgSubType,
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

var _ poller.PollingHandler[*TCloudImpl, map[string]*clb.DescribeTaskStatusResponseParams, poller.BaseDoneResult] = new(createClbPollingHandler)

type createClbPollingHandler struct {
	region string
}

// Done CLB 创建成功状态判断
func (h *createClbPollingHandler) Done(clbStatusMap map[string]*clb.DescribeTaskStatusResponseParams) (
	bool, *poller.BaseDoneResult) {

	result := &poller.BaseDoneResult{
		SuccessCloudIDs: make([]string, 0),
		FailedCloudIDs:  make([]string, 0),
		UnknownCloudIDs: make([]string, 0),
	}

	for _, status := range clbStatusMap {
		if status.Status == nil {
			return false, nil
		}
		switch converter.PtrToVal(status.Status) {
		case CLBTaskStatusRunning:
			// 还有任务在运行则是没有成功
			return false, nil
		case CLBTaskStatusFail:
			result.FailedCloudIDs = converter.PtrToSlice(status.LoadBalancerIds)
		case CLBTaskStatusSuccess:
			result.SuccessCloudIDs = converter.PtrToSlice(status.LoadBalancerIds)
		}
	}
	return true, result
}

// Poll 返回CLB创建任务结果
func (h *createClbPollingHandler) Poll(client *TCloudImpl, kt *kit.Kit, requestIDs []*string) (
	map[string]*clb.DescribeTaskStatusResponseParams, error) {

	taskOpt := &typelb.TCloudDescribeTaskStatusOption{Region: h.region}
	result := make(map[string]*clb.DescribeTaskStatusResponseParams)
	// 查询对应异步任务状态
	for _, reqID := range requestIDs {
		taskOpt.TaskId = converter.PtrToVal(reqID)
		if taskOpt.TaskId == "" {
			return nil, errors.New("empty request ID")
		}
		status, err := client.CLBDescribeTaskStatus(kt, taskOpt)
		if err != nil {
			return nil, err
		}

		result[taskOpt.TaskId] = status
	}
	return result, nil
}

// SetLoadBalancerSecurityGroups reference: https://cloud.tencent.com/document/api/214/34903
func (t *TCloudImpl) SetLoadBalancerSecurityGroups(kt *kit.Kit, opt *typelb.TCloudSetClbSecurityGroupOption) (
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

// DeleteLoadBalancer reference: https://cloud.tencent.com/document/api/214/30689
func (t *TCloudImpl) DeleteLoadBalancer(kt *kit.Kit, opt *typelb.TCloudDeleteOption) error {

	if opt == nil {
		return errf.New(errf.InvalidParameter, "delete clb option is required")
	}

	if err := opt.Validate(); err != nil {
		return errf.NewFromErr(errf.InvalidParameter, err)
	}

	client, err := t.clientSet.ClbClient(opt.Region)
	if err != nil {
		return fmt.Errorf("init tencent cloud clb client failed, region: %s, err: %v", opt.Region, err)
	}

	req := clb.NewDeleteLoadBalancerRequest()

	req.LoadBalancerIds = common.StringPtrs(opt.CloudIDs)

	resp, err := client.DeleteLoadBalancerWithContext(kt.Ctx, req)
	if err != nil {
		logs.Errorf("delete tcloud clb failed(RequestID:%s ), opt: %+v, err: %v, rid: %s",
			resp.Response.RequestId, opt, err, kt.Rid)
		return err
	}

	return nil
}

// UpdateLoadBalancer https://cloud.tencent.com/document/api/214/30680
func (t *TCloudImpl) UpdateLoadBalancer(kt *kit.Kit, opt *typelb.TCloudUpdateOption) (dealName *string, err error) {

	if opt == nil {
		return nil, errf.New(errf.InvalidParameter, "update clb option is required")
	}

	if err = opt.Validate(); err != nil {
		return nil, errf.NewFromErr(errf.InvalidParameter, err)
	}

	client, err := t.clientSet.ClbClient(opt.Region)
	if err != nil {
		return nil, fmt.Errorf("init tencent cloud clb client failed, region: %s, err: %v", opt.Region, err)
	}

	req := clb.NewModifyLoadBalancerAttributesRequest()

	req.LoadBalancerId = converter.ValToPtr(opt.LoadBalancerId)
	req.LoadBalancerPassToTarget = opt.LoadBalancerPassToTarget
	req.LoadBalancerName = opt.LoadBalancerName
	req.TargetRegionInfo = opt.TargetRegionInfo
	req.SnatPro = opt.SnatPro
	req.DeleteProtect = opt.DeleteProtect
	req.ModifyClassicDomain = opt.ModifyClassicDomain

	if opt.InternetChargeType != nil || opt.InternetMaxBandwidthOut != nil {
		req.InternetChargeInfo = &clb.InternetAccessible{
			InternetChargeType:      opt.InternetChargeType,
			InternetMaxBandwidthOut: opt.InternetMaxBandwidthOut,
			BandwidthpkgSubType:     opt.BandwidthpkgSubType,
		}
	}
	resp, err := client.ModifyLoadBalancerAttributesWithContext(kt.Ctx, req)
	if err != nil {
		logs.Errorf("delete tcloud lb failed,  err: %v, resp: %+v, opt: %+v,rid: %s", err, resp, opt, kt.Rid)
		return dealName, err
	}

	return resp.Response.DealName, nil
}

const (
	CLBTaskStatusSuccess = 0
	CLBTaskStatusFail    = 1
	CLBTaskStatusRunning = 2
)

// CLBDescribeTaskStatus 查询异步任务状态。
// 对于非查询类的接口（创建/删除负载均衡实例、监听器、规则以及绑定或解绑后端服务等），
// 在接口调用成功后，都需要使用本接口查询任务最终是否执行成功。
// https://cloud.tencent.com/document/api/214/30683
func (t *TCloudImpl) CLBDescribeTaskStatus(kt *kit.Kit, opt *typelb.TCloudDescribeTaskStatusOption) (
	*clb.DescribeTaskStatusResponseParams, error) {

	if opt == nil {
		return nil, errf.New(errf.InvalidParameter, "describe task status option can not be nil")
	}

	if err := opt.Validate(); err != nil {
		return nil, errf.NewFromErr(errf.InvalidParameter, err)
	}

	client, err := t.clientSet.ClbClient(opt.Region)
	if err != nil {
		return nil, fmt.Errorf("init tencent cloud clb client failed, region: %s, err: %v", opt.Region, err)
	}
	req := clb.NewDescribeTaskStatusRequest()
	if opt.TaskId != "" {
		req.TaskId = converter.ValToPtr(opt.TaskId)
	}
	if opt.DealName != "" {
		req.DealName = converter.ValToPtr(opt.DealName)
	}

	resp, err := client.DescribeTaskStatusWithContext(kt.Ctx, req)
	if err != nil {
		logs.Errorf("tencent cloud describe task status failed, req: %+v, err: %v, rid: %s", req, err, kt.Rid)
		return nil, err
	}
	return resp.Response, nil
}
