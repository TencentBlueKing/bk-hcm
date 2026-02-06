/*
 * Tencent is pleased to support the open source community by making 蓝鲸 available.
 * Copyright (C) 2017-2018 THL A29 Limited, a Tencent company. All rights reserved.
 * Licensed under the MIT License (the "License"); you may not use this file except
 * in compliance with the License. You may obtain a copy of the License at
 * http://opensource.org/licenses/MIT
 * Unless required by applicable law or agreed to in writing, software distributed under
 * the License is distributed on an "AS IS" BASIS, WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND,
 * either express or implied. See the License for the specific language governing permissions and
 * limitations under the License.
 */

package cvmapi

import (
	"context"
	"fmt"
	"net/http"

	"hcm/pkg/kit"
	"hcm/pkg/logs"
	"hcm/pkg/rest"
	"hcm/pkg/rest/client"

	"github.com/prometheus/client_golang/prometheus"
)

// CVMClientInterface cvm api interface
type CVMClientInterface interface {
	// CreateCvmOrder creates cvm order
	CreateCvmOrder(ctx context.Context, header http.Header, req *OrderCreateReq) (*OrderCreateResp, error)
	// QueryCvmOrders query cvm orders
	QueryCvmOrders(ctx context.Context, header http.Header, req *OrderQueryReq) (*OrderQueryResp, error)
	// QueryCvmInstances query cvm instances
	QueryCvmInstances(ctx context.Context, header http.Header, req *InstanceQueryReq) (*InstanceQueryResp, error)
	// QueryCvmCapacity query cvm capacity
	QueryCvmCapacity(ctx context.Context, header http.Header, req *CapacityReq) (*CapacityResp, error)
	// QueryCvmVpc query cvm subnet info
	QueryCvmVpc(ctx context.Context, header http.Header, req *VpcReq) (*VpcResp, error)
	// QueryRealCvmSubnet query real cvm subnet info
	QueryRealCvmSubnet(kt *kit.Kit, subnetReq SubnetRealParam) (*SubnetResp, error)
	// GetApproveLog get approve log
	GetApproveLog(ctx context.Context, header http.Header, req *GetApproveLogReq) (*GetApproveLogResp, error)
	// CreateCvmReturnOrder creates cvm return order
	CreateCvmReturnOrder(ctx context.Context, header http.Header, req *ReturnReq) (*OrderCreateResp, error)
	// QueryCvmReturnOrders query cvm return order status
	QueryCvmReturnOrders(ctx context.Context, header http.Header, req *OrderQueryReq) (*ReturnQueryResp, error)
	// QueryCvmReturnDetail query cvm return order detail
	QueryCvmReturnDetail(ctx context.Context, header http.Header, req *ReturnDetailReq) (*ReturnDetailResp, error)
	// CreateUpgradeOrder creates cvm upgrade order
	CreateUpgradeOrder(kt *kit.Kit, req *UpgradeReq) (*OrderCreateResp, error)
	// QueryCvmUpgradeDetail query cvm upgrade detail
	QueryCvmUpgradeDetail(kt *kit.Kit, req *UpgradeDetailReq) (*UpgradeDetailResp, error)
	// GetCvmProcess check if cvm is in any process like "退回"
	GetCvmProcess(ctx context.Context, header http.Header, req *GetCvmProcessReq) (*GetCvmProcessResp, error)
	// GetErpProcess check if physical machine is in any process like "退回"
	GetErpProcess(ctx context.Context, header http.Header, req *GetErpProcessReq) (*GetErpProcessResp, error)
	// QueryCvmInstanceType query cvm instance type
	QueryCvmInstanceType(kt *kit.Kit, params *QueryCvmInstanceTypeParams) (*QueryCvmInstanceTypeResp, error)
	// GetInstanceTypeInfo get instance type info
	GetInstanceTypeInfo(kt *kit.Kit, params *GetInstanceTypeInfoParams) (*GetInstanceTypeInfoResp, error)
	// GetCvmApproveLogs get cvm approve logs
	GetCvmApproveLogs(ctx context.Context, header http.Header, req *GetCvmApproveLogReq) (*GetCvmApproveLogsResp, error)
	// RevokeCvmOrder revoke cvm order
	RevokeCvmOrder(ctx context.Context, header http.Header, req *RevokeCvmOrderReq) (*RevokeCvmOrderResp, error)

	// QueryCvmCbsPlans query cvm and cbs plan info
	QueryCvmCbsPlans(ctx context.Context, header http.Header, req *CvmCbsPlanQueryReq) (*CvmCbsPlanQueryResp, error)
	// QueryAdjustAbleDemand query cvm and cbs plan info which can be adjusted
	QueryAdjustAbleDemand(ctx context.Context, header http.Header, req *CvmCbsAdjustAblePlanQueryReq) (
		*CvmCbsPlanQueryResp, error)
	// AdjustCvmCbsPlans adjust cvm and cbs plan info
	AdjustCvmCbsPlans(ctx context.Context, header http.Header, req *CvmCbsPlanAdjustReq) (*CvmCbsPlanAdjustResp, error)
	// AddCvmCbsPlan add cvm and cbs plan order
	AddCvmCbsPlan(ctx context.Context, header http.Header, req *AddCvmCbsPlanReq) (*AddCvmCbsPlanResp, error)
	// QueryPlanOrder query cvm and cbs plan order
	QueryPlanOrder(ctx context.Context, header http.Header, req *QueryPlanOrderReq) (*QueryPlanOrderResp, error)
	// QueryPlanOrderChange query cvm and cbs plan order change
	QueryPlanOrderChange(ctx context.Context, header http.Header, req *PlanOrderChangeReq) (*PlanOrderChangeResp, error)
	// QueryDemandChangeLog query demand change log
	QueryDemandChangeLog(ctx context.Context, header http.Header, req *DemandChangeLogQueryReq) (
		*DemandChangeLogQueryResp, error)
	// ReportPenaltyRatio report penalty ratio
	ReportPenaltyRatio(ctx context.Context, header http.Header, req *CvmCbsPlanPenaltyRatioReportReq) (
		*CvmCbsPlanPenaltyRatioReportResp, error)

	// QueryReturnPlan query return plan
	QueryReturnPlan(ctx context.Context, header http.Header, req *QueryReturnPlanReq) (*QueryReturnPlanResp, error)
	// QueryOrderList 根据销毁单据查询预测返还信息
	QueryOrderList(ctx context.Context, header http.Header, req *QueryOrderListReq) (
		*QueryOrderListResp, error)
	// MatchSwapGroup CRP亲合度可申领量匹配
	MatchSwapGroup(ctx context.Context, header http.Header, req *MatchSwapGroupReq) (*MatchSwapGroupResp, error)
	// QueryMatchTask CRP查询匹配单状态
	QueryMatchTask(ctx context.Context, header http.Header, req *QueryMatchTaskReq) (*QueryMatchTaskResp, error)
	// CreateTransOrder 预测转移
	CreateTransOrder(ctx context.Context, header http.Header, req *TransOrderReq) (
		*TransOrderResp, error)
}

// NewCVMClientInterface creates a cvm api instance
func NewCVMClientInterface(opts CVMCli, reg prometheus.Registerer) (CVMClientInterface, error) {
	cli, err := client.NewClient(nil)
	if err != nil {
		return nil, err
	}

	c := &client.Capability{
		Client: cli,
		Discover: &ServerDiscovery{
			name:    "cvm api",
			servers: []string{opts.CvmApiAddr},
		},
		MetricOpts: client.MetricOption{Register: reg},
	}

	cvm := &cvmApi{
		client: rest.NewClient(c, "/"),
	}

	return cvm, nil
}

// cvmApi cvm api interface implementation
type cvmApi struct {
	client rest.ClientInterface
}

// CreateCvmOrder creates cvm order
func (c *cvmApi) CreateCvmOrder(ctx context.Context, header http.Header, req *OrderCreateReq) (*OrderCreateResp,
	error) {

	subPath := "/apply/api/cvm"
	resp := new(OrderCreateResp)
	err := c.client.Post().
		WithContext(ctx).
		Body(req).
		SubResourcef(subPath).
		WithParam(CvmApiKey, CvmApiKeyVal).
		WithHeaders(header).
		Do().
		Into(resp)

	if err != nil {
		logs.Errorf("scheduler:cvm:create:order:failed, err: %v, subPath: %s, req: %+v", err, subPath, req)
		return nil, err
	}

	return resp, nil
}

// QueryCvmOrders query cvm orders
func (c *cvmApi) QueryCvmOrders(ctx context.Context, header http.Header, req *OrderQueryReq) (*OrderQueryResp, error) {
	subPath := "/apply/api/cvm"
	resp := new(OrderQueryResp)
	err := c.client.Post().
		WithContext(ctx).
		Body(req).
		SubResourcef(subPath).
		WithParam(CvmApiKey, CvmApiKeyVal).
		WithHeaders(header).
		Do().
		Into(resp)

	if err != nil {
		logs.Errorf("scheduler:cvm:query:order:failed, err: %v, subPath: %s, req: %+v", err, subPath, req)
		return nil, err
	}

	return resp, nil
}

// QueryCvmInstances query cvm instances
func (c *cvmApi) QueryCvmInstances(ctx context.Context, header http.Header, req *InstanceQueryReq) (*InstanceQueryResp,
	error) {

	subPath := "/apply/api/cvm"
	resp := new(InstanceQueryResp)
	err := c.client.Post().
		WithContext(ctx).
		Body(req).
		SubResourcef(subPath).
		WithParam(CvmApiKey, CvmApiKeyVal).
		WithHeaders(header).
		Do().
		Into(resp)

	return resp, err
}

// QueryCvmCapacity query cvm inventory
func (c *cvmApi) QueryCvmCapacity(ctx context.Context, header http.Header, req *CapacityReq) (*CapacityResp, error) {
	subPath := "/capacity/api/queryApplyCapacity"
	resp := new(CapacityResp)
	err := c.client.Post().
		WithContext(ctx).
		Body(req).
		SubResourcef(subPath).
		WithParam(CvmApiKey, CvmApiKeyVal).
		WithHeaders(header).
		Do().
		Into(resp)

	return resp, err
}

// QueryCvmVpc query cvm subnet info
func (c *cvmApi) QueryCvmVpc(ctx context.Context, header http.Header, req *VpcReq) (*VpcResp, error) {
	subPath := "/apply/api/cvm"
	resp := new(VpcResp)
	err := c.client.Post().
		WithContext(ctx).
		Body(req).
		SubResourcef(subPath).
		WithParam(CvmApiKey, CvmApiKeyVal).
		WithHeaders(header).
		Do().
		Into(resp)

	return resp, err
}

// QueryRealCvmSubnet query real cvm subnet info
func (c *cvmApi) QueryRealCvmSubnet(kt *kit.Kit, subnetReq SubnetRealParam) (*SubnetResp, error) {
	req := &SubnetRealReq{
		ReqMeta: ReqMeta{
			Id:      CvmId,
			JsonRpc: CvmJsonRpc,
			Method:  CvmRealSubnetMethod,
		},
		Params: &SubnetRealParam{
			DeptId:      CvmDeptId,
			Region:      subnetReq.Region,
			CloudCampus: subnetReq.CloudCampus,
			VpcId:       subnetReq.VpcId,
		},
	}

	subPath := "/capacity/api"
	resp := new(SubnetResp)
	err := c.client.Post().
		WithContext(kt.Ctx).
		Body(req).
		SubResourcef(subPath).
		WithParam(CvmApiKey, CvmApiKeyVal).
		WithHeaders(kt.Header()).
		Do().
		Into(resp)

	if err != nil {
		logs.Errorf("query real cvm subnet from crp failed, subnetReq: %+v, err: %+v, rid: %s", subnetReq, err, kt.Rid)
		return nil, err
	}

	if resp.Error.Code != 0 {
		logs.Errorf("query real cvm subnet from crp failed, subnetReq: %+v, errCode: %d, errMsg: %s, crpTraceID: %s, "+
			"rid: %s", subnetReq, resp.Error.Code, resp.Error.Message, resp.TraceId, kt.Rid)
		return nil, fmt.Errorf("query real cvm subnet from crp failed, errCode: %d, errMsg: %s, crpTraceID: %s",
			resp.Error.Code, resp.Error.Message, resp.TraceId)
	}

	return resp, nil
}

// GetApproveLog get approve log
func (c *cvmApi) GetApproveLog(ctx context.Context, header http.Header, req *GetApproveLogReq) (*GetApproveLogResp,
	error) {

	subPath := "/apply/api/cvm/getApproveLog"
	resp := new(GetApproveLogResp)
	err := c.client.Post().
		WithContext(ctx).
		Body(req).
		SubResourcef(subPath).
		WithParam(CvmApiKey, CvmApiKeyVal).
		WithHeaders(header).
		Do().
		Into(resp)

	return resp, err
}

// QueryCvmCbsPlans query cvm and cbs plans
func (c *cvmApi) QueryCvmCbsPlans(ctx context.Context, header http.Header, req *CvmCbsPlanQueryReq) (
	*CvmCbsPlanQueryResp, error) {

	subPath := "/yunti-demand/external"
	resp := new(CvmCbsPlanQueryResp)
	err := c.client.Post().
		WithContext(ctx).
		Body(req).
		SubResourcef(subPath).
		WithParam(CvmApiKey, CvmApiKeyVal).
		WithHeaders(header).
		Do().
		Into(resp)

	return resp, err
}

// QueryAdjustAbleDemand query adjust able demand
func (c *cvmApi) QueryAdjustAbleDemand(ctx context.Context, header http.Header, req *CvmCbsAdjustAblePlanQueryReq) (
	*CvmCbsPlanQueryResp, error) {

	subPath := "/yunti-demand/external"
	resp := new(CvmCbsPlanQueryResp)
	err := c.client.Post().
		WithContext(ctx).
		Body(req).
		SubResourcef(subPath).
		WithParam(CvmApiKey, CvmApiKeyVal).
		WithHeaders(header).
		Do().
		Into(resp)

	return resp, err
}

// AdjustCvmCbsPlans adjust cvm and cbs plans
func (c *cvmApi) AdjustCvmCbsPlans(ctx context.Context, header http.Header, req *CvmCbsPlanAdjustReq) (
	*CvmCbsPlanAdjustResp, error) {

	subPath := "/yunti-demand/external"
	resp := new(CvmCbsPlanAdjustResp)
	err := c.client.Post().
		WithContext(ctx).
		Body(req).
		SubResourcef(subPath).
		WithParam(CvmApiKey, CvmApiKeyVal).
		WithHeaders(header).
		Do().
		Into(resp)

	return resp, err
}

// AddCvmCbsPlan add cvm and cbs plan order
func (c *cvmApi) AddCvmCbsPlan(ctx context.Context, header http.Header, req *AddCvmCbsPlanReq) (*AddCvmCbsPlanResp,
	error) {

	subPath := "/yunti-demand/external"
	resp := new(AddCvmCbsPlanResp)
	err := c.client.Post().
		WithContext(ctx).
		Body(req).
		SubResourcef(subPath).
		WithParam(CvmApiKey, CvmApiKeyVal).
		WithHeaders(header).
		Do().
		Into(resp)

	return resp, err
}

// QueryPlanOrder query cvm and cbs plan order
func (c *cvmApi) QueryPlanOrder(ctx context.Context, header http.Header, req *QueryPlanOrderReq) (*QueryPlanOrderResp,
	error) {

	subPath := "/yunti-demand/external"
	resp := new(QueryPlanOrderResp)
	err := c.client.Post().
		WithContext(ctx).
		Body(req).
		SubResourcef(subPath).
		WithParam(CvmApiKey, CvmApiKeyVal).
		WithHeaders(header).
		Do().
		Into(resp)

	return resp, err
}

// QueryPlanOrderChange query cvm and cbs plan order change
func (c *cvmApi) QueryPlanOrderChange(ctx context.Context, header http.Header, req *PlanOrderChangeReq) (
	*PlanOrderChangeResp, error) {

	subPath := "/yunti-demand/external"
	resp := new(PlanOrderChangeResp)
	err := c.client.Post().
		WithContext(ctx).
		Body(req).
		SubResourcef(subPath).
		WithParam(CvmApiKey, CvmApiKeyVal).
		WithHeaders(header).
		Do().
		Into(resp)

	return resp, err
}

// QueryDemandChangeLog query cvm and cbs demand change log
func (c *cvmApi) QueryDemandChangeLog(ctx context.Context, header http.Header, req *DemandChangeLogQueryReq) (
	*DemandChangeLogQueryResp, error) {

	subPath := "/yunti-demand/external"
	resp := new(DemandChangeLogQueryResp)
	err := c.client.Post().
		WithContext(ctx).
		Body(req).
		SubResourcef(subPath).
		WithParam(CvmApiKey, CvmApiKeyVal).
		WithHeaders(header).
		Do().
		Into(resp)

	return resp, err
}

// ReportPenaltyRatio report penalty ratio
func (c *cvmApi) ReportPenaltyRatio(ctx context.Context, header http.Header, req *CvmCbsPlanPenaltyRatioReportReq) (
	*CvmCbsPlanPenaltyRatioReportResp, error) {

	subPath := "/tocservice/obs/"
	resp := new(CvmCbsPlanPenaltyRatioReportResp)
	err := c.client.Post().
		WithContext(ctx).
		Body(req).
		SubResourcef(subPath).
		WithParam(CvmApiKey, CvmApiKeyVal).
		WithHeaders(header).
		Do().
		Into(resp)

	return resp, err
}

// QueryReturnPlan query cvm return plan
func (c *cvmApi) QueryReturnPlan(ctx context.Context, header http.Header, req *QueryReturnPlanReq) (
	*QueryReturnPlanResp, error) {

	subPath := "/yunti-return/webapi/cvm"
	resp := new(QueryReturnPlanResp)
	err := c.client.Post().
		WithContext(ctx).
		Body(req).
		SubResourcef(subPath).
		WithParam(CvmApiKey, CvmApiKeyVal).
		WithHeaders(header).
		Do().
		Into(resp)

	return resp, err
}

// CreateCvmReturnOrder creates cvm return order
func (c *cvmApi) CreateCvmReturnOrder(ctx context.Context, header http.Header, req *ReturnReq) (*OrderCreateResp,
	error) {

	subPath := "/apply/api"
	resp := new(OrderCreateResp)
	err := c.client.Post().
		WithContext(ctx).
		Body(req).
		SubResourcef(subPath).
		WithParam(CvmApiKey, CvmApiKeyVal).
		WithHeaders(header).
		Do().
		Into(resp)

	return resp, err
}

// QueryCvmReturnOrders query cvm return order status
func (c *cvmApi) QueryCvmReturnOrders(ctx context.Context, header http.Header, req *OrderQueryReq) (*ReturnQueryResp,
	error) {

	subPath := "/apply/api"
	resp := new(ReturnQueryResp)
	err := c.client.Post().
		WithContext(ctx).
		Body(req).
		SubResourcef(subPath).
		WithParam(CvmApiKey, CvmApiKeyVal).
		WithHeaders(header).
		Do().
		Into(resp)

	return resp, err
}

// QueryCvmReturnDetail query cvm return order detail
func (c *cvmApi) QueryCvmReturnDetail(ctx context.Context, header http.Header, req *ReturnDetailReq) (*ReturnDetailResp,
	error) {

	subPath := "/apply/api"
	resp := new(ReturnDetailResp)
	err := c.client.Post().
		WithContext(ctx).
		Body(req).
		SubResourcef(subPath).
		WithParam(CvmApiKey, CvmApiKeyVal).
		WithHeaders(header).
		Do().
		Into(resp)

	return resp, err
}

// CreateUpgradeOrder creates cvm upgrade order
func (c *cvmApi) CreateUpgradeOrder(kt *kit.Kit, req *UpgradeReq) (*OrderCreateResp,
	error) {

	subPath := "/upgrade/api"
	resp := new(OrderCreateResp)
	err := c.client.Post().
		WithContext(kt.Ctx).
		Body(req).
		SubResourcef(subPath).
		WithParam(CvmApiKey, CvmApiKeyVal).
		WithHeaders(kt.Header()).
		Do().
		Into(resp)

	return resp, err
}

// QueryCvmUpgradeDetail query cvm upgrade order detail
func (c *cvmApi) QueryCvmUpgradeDetail(kt *kit.Kit, req *UpgradeDetailReq) (
	*UpgradeDetailResp, error) {

	subPath := "/upgrade/api"
	resp := new(UpgradeDetailResp)
	err := c.client.Post().
		WithContext(kt.Ctx).
		Body(req).
		SubResourcef(subPath).
		WithParam(CvmApiKey, CvmApiKeyVal).
		WithHeaders(kt.Header()).
		Do().
		Into(resp)

	return resp, err
}

// GetCvmProcess check if cvm is in any process like "退回"
func (c *cvmApi) GetCvmProcess(ctx context.Context, header http.Header, req *GetCvmProcessReq) (*GetCvmProcessResp,
	error) {

	subPath := "/operation/api/"
	resp := new(GetCvmProcessResp)
	err := c.client.Post().
		WithContext(ctx).
		Body(req).
		SubResourcef(subPath).
		WithParam(CvmApiKey, CvmApiKeyVal).
		WithHeaders(header).
		Do().
		Into(resp)

	if err != nil {
		logs.Errorf("recycle:cvm:get:GetCvmProcess:failed, err: %v, subPath: %s, req: %+v", err, subPath, req)
		return nil, err
	}

	return resp, err
}

// GetErpProcess check if physical machine is in any process like "退回"
func (c *cvmApi) GetErpProcess(ctx context.Context, header http.Header, req *GetErpProcessReq) (*GetErpProcessResp,
	error) {

	subPath := "/operation/api/"
	resp := new(GetErpProcessResp)
	err := c.client.Post().
		WithContext(ctx).
		Body(req).
		SubResourcef(subPath).
		WithParam(CvmApiKey, CvmApiKeyVal).
		WithHeaders(header).
		Do().
		Into(resp)

	if err != nil {
		logs.Errorf("recycle:cvm:get:GetErpProcess:failed, err: %v, subPath: %s, req: %+v", err, subPath, req)
		return nil, err
	}

	return resp, err
}

// QueryCvmInstanceType query cvm instance type
func (c *cvmApi) QueryCvmInstanceType(kt *kit.Kit, params *QueryCvmInstanceTypeParams) (
	*QueryCvmInstanceTypeResp, error) {

	req := &QueryCvmInstanceTypeReq{
		ReqMeta: ReqMeta{
			Id:      CvmId,
			JsonRpc: CvmJsonRpc,
			Method:  QueryCvmInstanceType,
		},
		Params: params,
	}
	subPath := "/apply/api/"
	resp := new(QueryCvmInstanceTypeResp)
	err := c.client.Post().
		WithContext(kt.Ctx).
		Body(req).
		SubResourcef(subPath).
		WithParam(CvmApiKey, CvmApiKeyVal).
		WithHeaders(kt.Header()).
		Do().
		Into(resp)

	if err != nil {
		logs.Errorf("query cvm instance type failed, err: %v, subPath: %s, req: %+v", err, subPath, req)
		return nil, err
	}

	if resp.Error.Code != 0 {
		logs.Errorf("query cvm instance type code error, subPath: %s, code: %d, msg: %s, crpTraceID: %s, req: %+v",
			subPath, resp.Error.Code, resp.Error.Message, resp.TraceId, req)
		return nil, fmt.Errorf("query cvm instance type code error, code: %d, msg: %s, crpTraceID: %s",
			resp.Error.Code, resp.Error.Message, resp.TraceId)
	}

	return resp, nil
}

// GetInstanceTypeInfo get instance type info
func (c *cvmApi) GetInstanceTypeInfo(kt *kit.Kit, params *GetInstanceTypeInfoParams) (
	*GetInstanceTypeInfoResp, error) {

	req := &GetInstanceTypeInfoReq{
		ReqMeta: ReqMeta{
			Id:      CvmId,
			JsonRpc: CvmJsonRpc,
			Method:  GetInstanceTypeInfoMethod,
		},
		Params: params,
	}
	subPath := "/apply/api/"
	resp := new(GetInstanceTypeInfoResp)
	err := c.client.Post().
		WithContext(kt.Ctx).
		Body(req).
		SubResourcef(subPath).
		WithParam(CvmApiKey, CvmApiKeyVal).
		WithHeaders(kt.Header()).
		Do().
		Into(resp)

	if err != nil {
		logs.Errorf("get instance type info failed, err: %v, subPath: %s, req: %+v", err, subPath, req)
		return nil, err
	}

	if resp.Error.Code != 0 {
		logs.Errorf("get instance type info code error, subPath: %s, code: %d, msg: %s, crpTraceID: %s, req: %+v",
			subPath, resp.Error.Code, resp.Error.Message, resp.TraceId, req)
		return nil, fmt.Errorf("get instance type info code error, code: %d, msg: %s, crpTraceID: %s",
			resp.Error.Code, resp.Error.Message, resp.TraceId)
	}

	return resp, nil
}

// GetCvmApproveLogs get cvm approve logs
func (c *cvmApi) GetCvmApproveLogs(ctx context.Context, header http.Header,
	req *GetCvmApproveLogReq) (*GetCvmApproveLogsResp, error) {

	subPath := "/api/approve"
	resp := new(GetCvmApproveLogsResp)
	err := c.client.Post().
		WithContext(ctx).
		Body(req).
		SubResourcef(subPath).
		WithParam(CvmApiKey, CvmApiKeyVal).
		WithHeaders(header).
		Do().
		Into(resp)

	return resp, err
}

// RevokeCvmOrder revoke cvm order
func (c *cvmApi) RevokeCvmOrder(ctx context.Context, header http.Header, req *RevokeCvmOrderReq) (
	*RevokeCvmOrderResp, error) {

	subPath := "/apply/api/"
	resp := new(RevokeCvmOrderResp)
	err := c.client.Post().
		WithContext(ctx).
		Body(req).
		SubResourcef(subPath).
		WithParam(CvmApiKey, CvmApiKeyVal).
		WithHeaders(header).
		Do().
		Into(resp)

	return resp, err
}

// QueryOrderList ...
func (c *cvmApi) QueryOrderList(ctx context.Context, header http.Header, req *QueryOrderListReq) (
	*QueryOrderListResp, error) {

	subPath := "/yunti-demand/external"
	resp := new(QueryOrderListResp)
	err := c.client.Post().
		WithContext(ctx).
		Body(req).
		SubResourcef(subPath).
		WithParam(CvmApiKey, CvmApiKeyVal).
		WithHeaders(header).
		Do().
		Into(resp)

	return resp, err
}

// CreateTransOrder 需求转移
func (c *cvmApi) CreateTransOrder(ctx context.Context, header http.Header, req *TransOrderReq) (
	*TransOrderResp, error) {

	subPath := "/yunti-demand/external"
	resp := new(TransOrderResp)
	err := c.client.Post().
		WithContext(ctx).
		Body(req).
		SubResourcef(subPath).
		WithParam(CvmApiKey, CvmApiKeyVal).
		WithHeaders(header).
		Do().
		Into(resp)

	return resp, err
}

// MatchSwapGroup CRP亲合度可申领量匹配
func (c *cvmApi) MatchSwapGroup(ctx context.Context, header http.Header, req *MatchSwapGroupReq) (*MatchSwapGroupResp,
	error) {
	subPath := "/packer/api/"
	resp := new(MatchSwapGroupResp)
	err := c.client.Post().
		WithContext(ctx).
		Body(req).
		SubResourcef(subPath).
		WithParam(CvmApiKey, CvmApiKeyVal).
		WithHeaders(header).
		Do().
		Into(resp)

	if err != nil {
		logs.Errorf("cvm:match:swap:group:failed, err: %v, subPath: %s, req: %+v", err, subPath, req)
		return nil, err
	}

	return resp, nil
}

// QueryMatchTask CRP查询匹配单状态
func (c *cvmApi) QueryMatchTask(ctx context.Context, header http.Header, req *QueryMatchTaskReq) (*QueryMatchTaskResp,
	error) {
	subPath := "/packer/api/"
	resp := new(QueryMatchTaskResp)
	err := c.client.Post().
		WithContext(ctx).
		Body(req).
		SubResourcef(subPath).
		WithParam(CvmApiKey, CvmApiKeyVal).
		WithHeaders(header).
		Do().
		Into(resp)

	if err != nil {
		logs.Errorf("cvm:query:match:task:failed, err: %v, subPath: %s, req: %+v", err, subPath, req)
		return nil, err
	}

	return resp, nil
}
