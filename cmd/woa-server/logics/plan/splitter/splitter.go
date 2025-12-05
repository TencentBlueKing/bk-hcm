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

// Package splitter ...
package splitter

import (
	"errors"
	"fmt"

	"hcm/cmd/woa-server/logics/plan/fetcher"
	"hcm/cmd/woa-server/types/device"
	ptypes "hcm/cmd/woa-server/types/plan"
	"hcm/pkg/client"
	"hcm/pkg/criteria/enumor"
	"hcm/pkg/dal/dao"
	rpt "hcm/pkg/dal/table/resource-plan/res-plan-ticket"
	"hcm/pkg/kit"
	"hcm/pkg/logs"
	"hcm/pkg/thirdparty/cvmapi"
)

// Splitter ...
type Splitter interface {
	SplitDeleteTicket(kt *kit.Kit, demands rpt.ResPlanDemands, planProductName, opProductName string) error
}

// SubTicketSplitter split res plan ticket to sub ticket
type SubTicketSplitter struct {
	dao         dao.Set
	client      *client.ClientSet
	crpCli      cvmapi.CVMClientInterface
	resFetcher  fetcher.Fetcher
	deviceTypes *device.DeviceTypesMap

	// adjustAbleDemands 记录每个本地预测需求对应的，CRP中可修改的原有预测
	adjustAbleDemands map[string][]*cvmapi.CvmCbsPlanQueryItem
	// adjCRPDemandsRst 记录对CRP中可修改的原有预测，已被调整使用的部分，用于解决多条需求共用一个可修改预测的场景
	adjCRPDemandsRst map[string]*AdjustAbleRemainObj
	// adjSplitGroupDemands 记录预测需求的拆分结果，按照子单的类型分组
	adjSplitGroupDemands map[enumor.RPTicketType][]*rpt.ResPlanDemand

	// transferAbleDemands 记录CRP中可转移的预测（来自中转产品），按预测所属年份分组
	transferAbleDemands map[int][]*cvmapi.CvmCbsPlanQueryItem
	// transferCRPDemandRst 记录对CRP中可转移的预测，已被转移使用的部分，用于解决多条需求共用一个可转移预测的场景
	transferCRPDemandRst map[string]*AdjustAbleRemainObj
}

// AdjustAbleRemainObj adjust able resource plan remained avail cpu core.
type AdjustAbleRemainObj struct {
	OriginDemand *cvmapi.CvmCbsPlanQueryItem
	AdjustType   enumor.CrpAdjustType
	// expectTime 当 adjustType 为 CrpAdjustTypeDelay 时，记录最新的期望交付时间
	ExpectTime  string
	WillConsume int64
}

// New create a SubTicketSplitter
func New(dao dao.Set, cli *client.ClientSet, crpCli cvmapi.CVMClientInterface, resFetcher fetcher.Fetcher,
	deviceMap *device.DeviceTypesMap) *SubTicketSplitter {

	return &SubTicketSplitter{
		dao:         dao,
		client:      cli,
		crpCli:      crpCli,
		resFetcher:  resFetcher,
		deviceTypes: deviceMap,

		adjustAbleDemands:    make(map[string][]*cvmapi.CvmCbsPlanQueryItem),
		adjCRPDemandsRst:     make(map[string]*AdjustAbleRemainObj),
		adjSplitGroupDemands: make(map[enumor.RPTicketType][]*rpt.ResPlanDemand),
		transferAbleDemands:  make(map[int][]*cvmapi.CvmCbsPlanQueryItem),
		transferCRPDemandRst: make(map[string]*AdjustAbleRemainObj),
	}
}

// getCRPAdjustAbleDemandsForAdjustDemands get all adjustable demands from CRP based on the request's adjust demands.
func (s *SubTicketSplitter) getAllCRPAdjustAbleDemands(kt *kit.Kit, demands rpt.ResPlanDemands,
	planProductName string, opProductName string) error {

	for _, demand := range demands {
		if demand.Original == nil {
			logs.Errorf("failed to construct adjust request, demand original is nil, rid: %s", kt.Rid)
			return errors.New("demand original is nil")
		}

		// query crp for a set of res plan demands that can be adjusted.
		adjustAbleReq := &ptypes.AdjustAbleDemandsReq{
			RegionName:      demand.Original.RegionName,
			DeviceFamily:    demand.Original.Cvm.DeviceFamily,
			DeviceType:      demand.Original.Cvm.DeviceType,
			ExpectTime:      demand.Original.ExpectTime,
			PlanProductName: planProductName,
			OpProductName:   opProductName,
			ObsProject:      demand.Original.ObsProject,
			DiskType:        demand.Original.Cbs.DiskType,
			ResMode:         demand.Original.Cvm.ResMode,
		}
		adjustAbleDemands, err := s.queryAdjustAbleDemands(kt, adjustAbleReq)
		if err != nil {
			logs.Errorf("failed to query adjust able demands, err: %v, req: %+v, rid: %s", err, adjustAbleReq,
				kt.Rid)
			return err
		}

		s.adjustAbleDemands[demand.Original.DemandID] = adjustAbleDemands
	}

	return nil
}

// queryAdjustAbleDemands query demands that can be adjusted.
// TODO 在splitter 和 dispatcher 中都调用，可以抽象为一个公共函数
func (s *SubTicketSplitter) queryAdjustAbleDemands(kt *kit.Kit, req *ptypes.AdjustAbleDemandsReq) (
	[]*cvmapi.CvmCbsPlanQueryItem, error) {

	if err := req.Validate(); err != nil {
		return nil, err
	}

	// init request parameter.
	queryReq := &cvmapi.CvmCbsAdjustAblePlanQueryReq{
		ReqMeta: cvmapi.ReqMeta{
			Id:      cvmapi.CvmId,
			JsonRpc: cvmapi.CvmJsonRpc,
			Method:  cvmapi.CvmCbsAdjustAblePlanQueryMethod,
		},
		Params: convAdjustAbleQueryParam(req),
	}

	rst, err := s.crpCli.QueryAdjustAbleDemand(kt.Ctx, kt.Header(), queryReq)
	if err != nil {
		logs.Errorf("failed to query adjust able demands, err: %v, req: %+v, rid: %s", err, *queryReq.Params,
			kt.Rid)
		return nil, err
	}

	if rst.Error.Code != 0 {
		logs.Errorf("failed to query adjust able demands, err: %v, crp_trace: %s, rid: %s", rst.Error.Message,
			rst.TraceId, kt.Rid)
		return nil, errors.New(rst.Error.Message)
	}

	if rst.Result == nil || len(rst.Result.Data) == 0 {
		logs.Errorf("failed to query adjust able demands, return is empty, crp_trace: %s, rid: %s",
			rst.TraceId, kt.Rid)
		return nil, errors.New("no demands can be adjusted in CRP")
	}

	return rst.Result.Data, nil
}

// convAdjustAbleQueryParam 构造 cvmapi.CvmCbsAdjustAblePlanQueryMethod 接口的查询参数
func convAdjustAbleQueryParam(req *ptypes.AdjustAbleDemandsReq) *cvmapi.CvmCbsAdjustAblePlanQueryParam {
	reqParams := new(cvmapi.CvmCbsAdjustAblePlanQueryParam)

	if len(req.RegionName) > 0 {
		reqParams.CityName = req.RegionName
	}

	if len(req.DeviceFamily) > 0 {
		reqParams.InstanceFamily = req.DeviceFamily
	}

	if len(req.DeviceType) > 0 {
		reqParams.InstanceModel = req.DeviceType
	}

	if len(req.ExpectTime) > 0 {
		reqParams.UseTime = req.ExpectTime
	}

	if len(req.PlanProductName) > 0 {
		reqParams.PlanProductName = req.PlanProductName
	}

	if len(req.OpProductName) > 0 {
		reqParams.ProductName = req.OpProductName
	}

	if len(req.ObsProject) > 0 {
		reqParams.ProjectName = string(req.ObsProject)
	}

	if len(req.DiskType) > 0 {
		// TODO 因CRP会在拆单时改变云盘的类型，且云盘不会参与模糊匹配，因此查询时暂时不考虑diskType参数.
		// reqParams.DiskTypeName = req.DiskType.Name()
	}

	if len(req.ResMode) > 0 {
		reqParams.ResourceMode = string(req.ResMode)
	}

	return reqParams
}

// matchReviewedCRPDemands 从所有可用于调整的CRP预测中，匹配已评审的预测，并计算返回已评审和未评审的核数
func (s *SubTicketSplitter) matchReviewedCRPDemands(kt *kit.Kit, demand rpt.ResPlanDemand,
	adjustType enumor.CrpAdjustType) (reviewedCore int64, unReviewedCore int64, err error) {

	demandID := demand.Original.DemandID
	// 1. 查询原始预测的预测内外情况
	demandDetail, err := s.resFetcher.GetResPlanDemandDetail(kt, demandID, []int64{})
	if err != nil {
		logs.Errorf("failed to get res plan demand detail, err: %v, demand_id: %s, rid: %s", err, demandID,
			kt.Rid)
		return reviewedCore, unReviewedCore, err
	}

	adjustAbleDemands, ok := s.adjustAbleDemands[demandID]
	if !ok {
		err = fmt.Errorf("not found adjust able demands, demand id: %s", demandID)
		logs.Errorf("failed to get adjust able demands, err: %v, rid: %s", err, kt.Rid)
		return reviewedCore, unReviewedCore, err
	}

	// 2. 计算可转移的预测量，将转移的消耗记录到 adjCRPDemandsRst
	needCpuCores := demand.Original.Cvm.CpuCore
	for _, adjustAbleD := range adjustAbleDemands {
		if needCpuCores <= 0 {
			break
		}

		// 未评审需求跳过，不记录
		if adjustAbleD.ReviewStatus == enumor.ResPlanReviewStatusPending {
			continue
		}

		// 取消类型的调整单据，不需要给expectTime
		var updateExpectTime string
		if demand.Updated != nil {
			updateExpectTime = demand.Updated.ExpectTime
		}

		canConsume, err := s.calcAdjustAbleDCanConsumeCPU(kt, adjustAbleD, adjustType, needCpuCores, demandDetail,
			updateExpectTime)
		if err != nil {
			logs.Warnf("crp demand %s has been used by other demand, cannot share with demand %s, skip, rid: %s",
				adjustAbleD.SliceId, demandID, kt.Rid)
			continue
		}

		needCpuCores -= canConsume
	}

	unReviewedCore = needCpuCores
	reviewedCore = demand.Original.Cvm.CpuCore - needCpuCores
	return reviewedCore, unReviewedCore, nil
}

// calcAdjustAbleDCanConsumeCPU 计算单个可调整的crp预测，有多少CPU可以被用在demand的调整中
// 并将结果暂存在 adjCRPDemandsRst 中
func (s *SubTicketSplitter) calcAdjustAbleDCanConsumeCPU(kt *kit.Kit, adjustAbleD *cvmapi.CvmCbsPlanQueryItem,
	adjustType enumor.CrpAdjustType, needCpuCores int64, demandDetail *ptypes.GetPlanDemandDetailResp,
	updateExpectTime string) (int64, error) {

	var canConsume int64
	// 磁盘类型需一致，未知的磁盘类型（包括空值）除外
	adjustDiskTypeName := adjustAbleD.DiskType.Name()
	if adjustDiskTypeName != "" && adjustDiskTypeName != demandDetail.DiskTypeName {
		return canConsume, nil
	}

	// 预测内外需一致
	if demandDetail.PlanType.GetCode() != enumor.PlanType(adjustAbleD.InPlan).GetCode() {
		return canConsume, nil
	}

	remainedCpuCores := adjustAbleD.RealCoreAmount
	if _, ok := s.adjCRPDemandsRst[adjustAbleD.SliceId]; ok {
		remainedCpuCores -= s.adjCRPDemandsRst[adjustAbleD.SliceId].WillConsume
	}

	canConsume = min(needCpuCores, remainedCpuCores)
	// CvmAmount虽然理论上大于等于RealCoreAmount，但是为确保后续除法计算不出异常，判断下CvmAmount的大小
	if canConsume <= 0 || adjustAbleD.CvmAmount == 0 {
		return 0, nil
	}

	if _, ok := s.adjCRPDemandsRst[adjustAbleD.SliceId]; !ok {
		s.adjCRPDemandsRst[adjustAbleD.SliceId] = &AdjustAbleRemainObj{
			OriginDemand: adjustAbleD.Clone(),
			AdjustType:   adjustType,
		}
		// 取消类型的调整单据，不需要expectTime
		if updateExpectTime != "" {
			s.adjCRPDemandsRst[adjustAbleD.SliceId].ExpectTime = updateExpectTime
		}
	}
	adjustAbleRemain := s.adjCRPDemandsRst[adjustAbleD.SliceId]

	if adjustType != adjustAbleRemain.AdjustType {
		logs.Warnf("adjust type is not eq, adjust type: %s, slice: %s, adjust demand: %+v, rid: %s", adjustType,
			adjustAbleD.SliceId, demandDetail, kt.Rid)
		return canConsume, errors.New("adjust type is not eq")
	}

	if adjustType == enumor.CrpAdjustTypeDelay {
		if updateExpectTime == "" {
			logs.Warnf("adjust type is delay, but updated expect time is nil, slice: %s, adjust demand: %+v, rid: %s",
				adjustAbleD.SliceId, demandDetail, kt.Rid)
			return canConsume, errors.New("adjust type is delay, but updated expect time is nil")
		}
		// 如果多条通配的预测一起延期，延期的期望时间必须一致。否则不能在同一个CRP demand中操作调整
		if updateExpectTime != adjustAbleRemain.ExpectTime {
			logs.Warnf("adjust type is delay, but expect time is not eq, slice: %s, adjust demand: %+v, rid: %s",
				adjustAbleD.SliceId, demandDetail, kt.Rid)
			return canConsume, errors.New("adjust type is delay, but expect time is not eq")
		}
	}

	adjustAbleRemain.WillConsume += canConsume

	return canConsume, nil
}
