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

package dispatcher

import (
	"errors"
	"fmt"
	"slices"
	"strconv"
	"strings"

	"hcm/cmd/woa-server/logics/plan/fetcher"
	ptypes "hcm/cmd/woa-server/types/plan"
	"hcm/pkg/cc"
	"hcm/pkg/criteria/constant"
	"hcm/pkg/criteria/enumor"
	rpt "hcm/pkg/dal/table/resource-plan/res-plan-ticket"
	"hcm/pkg/kit"
	"hcm/pkg/logs"
	"hcm/pkg/thirdparty/cvmapi"
	cvt "hcm/pkg/tools/converter"
	"hcm/pkg/tools/math"
	"hcm/pkg/tools/retry"
	"hcm/pkg/tools/uuid"

	"github.com/shopspring/decimal"
)

// AdjustAbleRemainObj adjust able resource plan remained avail cpu core.
type AdjustAbleRemainObj struct {
	OriginDemand *cvmapi.CvmCbsPlanQueryItem
	AdjustType   enumor.CrpAdjustType
	// expectTime 当 adjustType 为 CrpAdjustTypeDelay 时，记录最新的期望交付时间
	ExpectTime string
	// WillConsume 记录CRP中预测在本次调整的量
	WillConsume int64
	// TransferTarget 用于转移场景，按照用户需求分组计算并存储转移量
	TransferTarget map[rpt.UpdatedRPDemandItem]int64
}

// CrpTicketCreator crp ticket creator
// CRP预测修改请求分为两部分：对CRP原有预测的调减，对用户更新内容的追加
type CrpTicketCreator struct {
	resFetcher fetcher.Fetcher
	crpCli     cvmapi.CVMClientInterface

	// adjCRPDemandsRst 记录对CRP中可修改的原有预测，将要产生调减的内容
	adjCRPDemandsRst map[string]*AdjustAbleRemainObj
	// appendUpdateDemand 记录将要追加的用户更新的预测需求内容
	appendUpdateDemand []*cvmapi.AdjustUpdatedData
	// adjustAbleDemands 记录每个本地预测需求对应的，CRP中可修改的原有预测
	adjustAbleDemands map[string][]*cvmapi.CvmCbsPlanQueryItem
	// transferAbleDemands 记录CRP中可转移的预测（来自中转产品），按预测所属年份分组
	transferAbleDemands map[int][]*cvmapi.CvmCbsPlanQueryItem
}

// NewCrpTicketCreator new CrpTicketCreator
func NewCrpTicketCreator(fetch fetcher.Fetcher, crpCli cvmapi.CVMClientInterface) *CrpTicketCreator {
	return &CrpTicketCreator{
		resFetcher:          fetch,
		crpCli:              crpCli,
		adjCRPDemandsRst:    make(map[string]*AdjustAbleRemainObj),
		appendUpdateDemand:  make([]*cvmapi.AdjustUpdatedData, 0),
		adjustAbleDemands:   make(map[string][]*cvmapi.CvmCbsPlanQueryItem),
		transferAbleDemands: make(map[int][]*cvmapi.CvmCbsPlanQueryItem),
	}
}

// CreateCRPTicket create crp ticket
func (c *CrpTicketCreator) CreateCRPTicket(kt *kit.Kit, subTicket *ptypes.SubTicketInfo) (string, error) {
	// 1. 新增单走新增流程
	if subTicket.Type == enumor.RPTicketTypeAdd {
		return c.createAddCrpTicket(kt, subTicket)
	}

	// 2. 生成调整请求
	srcData, updateData, err := c.constructCrpAdjustReqParams(kt, subTicket)
	if err != nil {
		logs.Errorf("failed to construct adjust crp plan order request params, ticketID: %s, err: %v, rid: %s",
			subTicket.ID, err, kt.Rid)
		return "", err
	}

	adjustReq := &cvmapi.CvmCbsPlanAdjustReq{
		ReqMeta: cvmapi.ReqMeta{
			Id:      cvmapi.CvmId,
			JsonRpc: cvmapi.CvmJsonRpc,
			Method:  cvmapi.CvmCbsPlanAdjustMethod,
		},
		Params: &cvmapi.CvmCbsPlanAdjustParam{
			BaseInfo: &cvmapi.AdjustBaseInfo{
				DeptId:          int(subTicket.VirtualDeptID),
				DeptName:        subTicket.VirtualDeptName,
				PlanProductName: subTicket.PlanProductName,
				Desc:            "",
			},
			SrcData:     srcData,
			UpdatedData: updateData,
			UserName:    subTicket.Applicant,
		},
	}

	switch subTicket.DemandClass {
	case enumor.DemandClassCVM:
		adjustReq.Params.BaseInfo.Desc = cvmapi.CvmCbsPlanDefaultCvmDesc
	case enumor.DemandClassCA:
		adjustReq.Params.BaseInfo.Desc = cvmapi.CvmCbsPlanDefaultCADesc
	default:
		logs.Warnf("failed to construct adjust desc, unsupported demand class: %s, rid: %s",
			subTicket.DemandClass, kt.Rid)
	}

	resp := new(cvmapi.CvmCbsPlanAdjustResp)
	rangeMS := [2]uint{CreateCrpTicketDefaultRetryDelayMinMS, CreateCrpTicketDefaultRetryDelayMaxMS}
	policy := retry.NewRetryPolicy(0, rangeMS)
	for {
		resp, err = c.crpCli.AdjustCvmCbsPlans(kt.Ctx, kt.Header(), adjustReq)
		if err != nil {
			logs.Errorf("failed to add cvm & cbs plan order, err: %v, sub_ticket_id: %s, rid: %s", err, subTicket.ID,
				kt.Rid)
			return "", err
		}
		// 仅在碰到限频错误时进行重试
		if resp.Error.Code == CreateCrpTicketRespRateLimitCode {
			if policy.RetryCount()+1 < CreateCrpTicketDefaultRetryTimes {
				// 	非最后一次重试，继续sleep
				logs.Warnf("call crp rate limit, will sleep for retry, retry count: %d, err: %v, crp_trace: %s, rid: %s",
					policy.RetryCount(), resp.Error, resp.TraceId, kt.Rid)
				policy.Sleep()
				continue
			}
		}
		// 其他情况都跳过
		break
	}

	if resp.Error.Code != 0 {
		logs.Errorf("failed to adjust cvm & cbs plan order, subTicketID: %s, code: %d, msg: %s, req: %v, "+
			"crp_trace: %s, rid: %s", subTicket.ID, resp.Error.Code, resp.Error.Message, cvt.PtrToVal(adjustReq),
			resp.TraceId, kt.Rid)
		if strings.Contains(resp.Error.Message, constant.CRPResPlanDemandIsOverLimit) {
			return "", fmt.Errorf(constant.CRPResPlanDemandIsOverLimitMessage,
				strings.Join(cc.WoaServer().ResPlan.CRPOverLimitContact, ","))
		}
		return "", fmt.Errorf("failed to create crp ticket, code: %d, msg: %s", resp.Error.Code, resp.Error.Message)
	}

	sn := resp.Result.OrderId
	if sn == "" {
		logs.Errorf("failed to adjust cvm & cbs plan order, for return empty order id, subTicketID: %s, rid: %s",
			subTicket.ID, kt.Rid)
		return "", errors.New("failed to create crp ticket, for return empty order id")
	}

	return sn, nil
}

// constructCrpAdjustReq construct cvm cbs plan adjust request.
func (c *CrpTicketCreator) constructCrpAdjustReqParams(kt *kit.Kit, subTicket *ptypes.SubTicketInfo) (
	[]*cvmapi.AdjustSrcData, []*cvmapi.AdjustUpdatedData, error) {

	// 根据子单的需求类型（新增、删除）决定转移的方向
	if subTicket.Type == enumor.RPTicketTypeTransfer {
		transferDirection := subTicket.Type
		if len(subTicket.Demands) == 0 {
			logs.Errorf("sub ticket has no demands, sub ticket id: %s, rid: %s", subTicket.ID, kt.Rid)
			return nil, nil, errors.New("sub ticket has no demands")
		}
		if subTicket.Demands[0].Original == nil {
			transferDirection = enumor.RPTicketTypeAdd
		} else {
			transferDirection = enumor.RPTicketTypeDelete
		}

		switch transferDirection {
		case enumor.RPTicketTypeAdd:
			return c.constructAddTransferAdjustReqParams(kt, subTicket)
		case enumor.RPTicketTypeDelete:
			return c.constructDelTransferAdjustReqParams(kt, subTicket)
		default:
			logs.Errorf("unsupported transfer direction: %s, id: %s, rid: %s", transferDirection, subTicket.ID,
				kt.Rid)
			return nil, nil, errors.New("unsupported transfer direction")
		}
	}

	return c.constructNormalCrpAdjustReqParams(kt, subTicket)
}

// constructAddTransferAdjustReqParams 构建CRP调整单请求参数，用于新增转移调整场景（中转产品 -> 本业务）
func (c *CrpTicketCreator) constructAddTransferAdjustReqParams(kt *kit.Kit, subTicket *ptypes.SubTicketInfo) (
	[]*cvmapi.AdjustSrcData, []*cvmapi.AdjustUpdatedData, error) {

	// 1. 整合所有需求的项目类型和技术大类
	obsProjectMap := make([]enumor.ObsProject, 0)
	technicalClassMap := make([]string, 0)
	for _, d := range subTicket.Demands {
		if d.Updated == nil {
			logs.Errorf("updated demand is nil, sub ticket id: %s, rid: %s", subTicket.ID, kt.Rid)
			return nil, nil, errors.New("updated demand is nil")
		}
		obsProjectMap = append(obsProjectMap, d.Updated.ObsProject)
		technicalClassMap = append(technicalClassMap, d.Updated.Cvm.TechnicalClass)
	}

	// 2. 查询CRP中可转移的预测（中转产品）
	err := c.queryTransferCRPDemands(kt, obsProjectMap, technicalClassMap)
	if err != nil {
		logs.Errorf("query crp demands failed, err: %v, sub ticket id: %s, rid: %s", err, subTicket.ID, kt.Rid)
		return nil, nil, err
	}

	// 3. 构造调整请求的详细内容。从可转移的中转产品预测中进行调减，以及从预测需求内容中进行追加
	for _, demand := range subTicket.Demands {
		if err := c.constructAdjustDemandDetails(kt, subTicket, demand); err != nil {
			logs.Errorf("failed to construct adjust demand details, err: %v, rid: %s", err, kt.Rid)
			return nil, nil, err
		}
	}

	// 3. 根据暂存的修改信息生成最终的变更请求内容
	srcData := make([]*cvmapi.AdjustSrcData, 0)
	updatedData := make([]*cvmapi.AdjustUpdatedData, 0)

	for _, adjustObj := range c.adjCRPDemandsRst {
		adjustAbleD := adjustObj.OriginDemand
		willConsume := adjustObj.WillConsume

		deviceCore := float64(adjustAbleD.CoreAmount) / adjustAbleD.CvmAmount
		// 和CRP确认保留4位小数可以，但是肯定会存在误差
		willChangeCvm, err := math.RoundToDecimalPlaces(float64(willConsume)/deviceCore, 4)
		if err != nil {
			logs.Errorf("failed to round change cvm to 4 decimal places, err: %v, crp_demand:%s, change cvm: %f, "+
				"rid: %s", err, adjustAbleD.DemandId, float64(willConsume)/deviceCore, kt.Rid)
			return nil, nil, err
		}

		// srcItem为中转产品的预测
		srcItem := &cvmapi.AdjustSrcData{
			// 转移一定是常规修改
			AdjustType:          string(enumor.CrpAdjustTypeUpdate),
			CvmCbsPlanQueryItem: adjustAbleD.Clone(),
		}
		// updatedItem为中转产品的预测
		updatedItem := &cvmapi.AdjustUpdatedData{
			AdjustType:          string(enumor.CrpAdjustTypeUpdate),
			CvmCbsPlanQueryItem: adjustAbleD.Clone(),
		}
		// 短租项目预测需要提供isAutoReturnPlan参数
		if adjustAbleD.ProjectName == enumor.ObsProjectShortLease {
			srcItem.IsAutoReturnPlan = true
			updatedItem.IsAutoReturnPlan = true
		}
		// transferItem为转移到本业务的预测
		transferItems, err := c.constructTransferAppendDataToBiz(kt, subTicket, deviceCore, adjustAbleD,
			adjustObj.TransferTarget)
		if err != nil {
			logs.Errorf("failed to construct transfer append data, err: %v, rid: %s", err, kt.Rid)
			return nil, nil, err
		}

		// TODO 硬盘跟机器大小毫无关系，且预测扣除时也不考虑硬盘大小，这里先不进行硬盘的扣除，避免出现负数；但是CBS类型的调减会没有效果
		updatedItem.CvmAmount = max(updatedItem.CvmAmount-willChangeCvm, 0)
		updatedItem.CoreAmount -= willConsume

		srcData = append(srcData, srcItem)
		updatedData = append(updatedData, updatedItem)
		updatedData = append(updatedData, transferItems...)
	}

	return srcData, updatedData, nil
}

// constructTransferCrpAdjustReqParams 构建CRP调整单请求参数，用于删除转移调整场景（本业务 -> 中转产品）
func (c *CrpTicketCreator) constructDelTransferAdjustReqParams(kt *kit.Kit, subTicket *ptypes.SubTicketInfo) (
	[]*cvmapi.AdjustSrcData, []*cvmapi.AdjustUpdatedData, error) {

	// 1. 通过通配的方式，查询可修改的CRP原有预测
	err := c.getAllCRPAdjustAbleDemands(kt, subTicket.Type, subTicket.Demands,
		subTicket.PlanProductName, subTicket.OpProductName)
	if err != nil {
		logs.Errorf("failed to get crp adjust able demands, err: %v, rid: %s", err, kt.Rid)
		return nil, nil, err
	}

	// 2. 构造调整请求的详细内容。从可修改的CRP预测中进行调减，以及从预测需求内容中进行追加
	for _, demand := range subTicket.Demands {
		if err := c.constructAdjustDemandDetails(kt, subTicket, demand); err != nil {
			logs.Errorf("failed to construct adjust demand details, err: %v, rid: %s", err, kt.Rid)
			return nil, nil, err
		}
	}

	// 3. 根据暂存的修改信息生成最终的变更请求内容
	srcData := make([]*cvmapi.AdjustSrcData, 0)
	updatedData := make([]*cvmapi.AdjustUpdatedData, 0)

	for _, adjustObj := range c.adjCRPDemandsRst {
		adjustAbleD := adjustObj.OriginDemand
		willConsume := adjustObj.WillConsume

		deviceCore := float64(adjustAbleD.CoreAmount) / adjustAbleD.CvmAmount
		// 和CRP确认保留2位小数可以，但是肯定会存在误差
		willChangeCvm, err := math.RoundToDecimalPlaces(float64(willConsume)/deviceCore, 2)
		if err != nil {
			logs.Errorf("failed to round change cvm to 2 decimal places, err: %v, crp_demand:%s, change cvm: %f, "+
				"rid: %s", err, adjustAbleD.DemandId, float64(willConsume)/deviceCore, kt.Rid)
			return nil, nil, err
		}

		srcItem := &cvmapi.AdjustSrcData{
			// 转移一定是常规修改
			AdjustType:          string(enumor.CrpAdjustTypeUpdate),
			CvmCbsPlanQueryItem: adjustAbleD.Clone(),
		}
		updatedItem := &cvmapi.AdjustUpdatedData{
			AdjustType:          string(enumor.CrpAdjustTypeUpdate),
			CvmCbsPlanQueryItem: adjustAbleD.Clone(),
		}
		// 短租项目预测需要提供isAutoReturnPlan参数
		if adjustAbleD.ProjectName == enumor.ObsProjectShortLease {
			srcItem.IsAutoReturnPlan = true
			updatedItem.IsAutoReturnPlan = true
		}
		// transferItem为转移到中转池的预测
		transferItems := c.constructTransferAppendDataToPool(adjustAbleD, willChangeCvm, willConsume)

		// if adjust type is update, updated item are normal.
		// if adjust type is delay, updated will fill parameter TimeAdjustCvmAmount with OriginOs.
		// if adjust type is cancel, error.
		// TODO 硬盘跟机器大小毫无关系，且预测扣除时也不考虑硬盘大小，这里先不进行硬盘的扣除，避免出现负数；但是CBS类型的调减会没有效果
		updatedItem.CvmAmount = max(updatedItem.CvmAmount-willChangeCvm, 0)
		updatedItem.CoreAmount -= willConsume

		srcData = append(srcData, srcItem)
		updatedData = append(updatedData, updatedItem)
		updatedData = append(updatedData, transferItems)
	}

	return srcData, updatedData, nil
}

// constructNormalCrpAdjustReqParams 构建CRP调整单请求参数，用于常规调整场景（修改、延期、删除等，非转移）
func (c *CrpTicketCreator) constructNormalCrpAdjustReqParams(kt *kit.Kit, subTicket *ptypes.SubTicketInfo) (
	[]*cvmapi.AdjustSrcData, []*cvmapi.AdjustUpdatedData, error) {

	// 1. 通过通配的方式，查询可修改的CRP原有预测
	err := c.getAllCRPAdjustAbleDemands(kt, subTicket.Type, subTicket.Demands,
		subTicket.PlanProductName, subTicket.OpProductName)
	if err != nil {
		logs.Errorf("failed to get crp adjust able demands, err: %v, rid: %s", err, kt.Rid)
		return nil, nil, err
	}

	// 2. 构造调整请求的详细内容。从可修改的CRP预测中进行调减，以及从预测需求内容中进行追加
	for _, demand := range subTicket.Demands {
		if err := c.constructAdjustDemandDetails(kt, subTicket, demand); err != nil {
			logs.Errorf("failed to construct adjust demand details, err: %v, rid: %s", err, kt.Rid)
			return nil, nil, err
		}
	}

	// 3. 根据暂存的修改信息生成最终的变更请求内容
	srcData := make([]*cvmapi.AdjustSrcData, 0)
	updatedData := make([]*cvmapi.AdjustUpdatedData, 0)
	// 3.1. 构造CRP原始预测调减的请求数据
	for _, adjustObj := range c.adjCRPDemandsRst {
		adjustAbleD := adjustObj.OriginDemand
		willConsume := adjustObj.WillConsume

		deviceCore := float64(adjustAbleD.CoreAmount) / adjustAbleD.CvmAmount
		// 和CRP确认保留2位小数可以，但是肯定会存在误差
		willChangeCvm, err := math.RoundToDecimalPlaces(float64(willConsume)/deviceCore, 2)
		if err != nil {
			logs.Errorf("failed to round change cvm to 2 decimal places, err: %v, crp_demand:%s, change cvm: %f, "+
				"rid: %s", err, adjustAbleD.DemandId, float64(willConsume)/deviceCore, kt.Rid)
			return nil, nil, err
		}

		srcItem := &cvmapi.AdjustSrcData{
			AdjustType:          string(adjustObj.AdjustType),
			CvmCbsPlanQueryItem: adjustAbleD.Clone(),
		}
		updatedItem := &cvmapi.AdjustUpdatedData{
			AdjustType:          string(adjustObj.AdjustType),
			CvmCbsPlanQueryItem: adjustAbleD.Clone(),
		}

		// 短租项目预测需要提供isAutoReturnPlan参数
		if adjustAbleD.ProjectName == enumor.ObsProjectShortLease {
			srcItem.IsAutoReturnPlan = true
			updatedItem.IsAutoReturnPlan = true
		}

		// if adjust type is update, updated item are normal.
		// if adjust type is delay, updated will fill parameter TimeAdjustCvmAmount with OriginOs.
		// if adjust type is cancel, error.
		switch adjustObj.AdjustType {
		case enumor.CrpAdjustTypeUpdate:
			// TODO 硬盘跟机器大小毫无关系，且预测扣除时也不考虑硬盘大小，这里先不进行硬盘的扣除，避免出现负数；但是CBS类型的调减会没有效果
			updatedItem.CvmAmount = max(updatedItem.CvmAmount-willChangeCvm, 0)
			updatedItem.CoreAmount -= willConsume
		case enumor.CrpAdjustTypeDelay:
			updatedItem.ReturnPlanTime = ""
			updatedItem.UseTime = adjustObj.ExpectTime
			updatedItem.TimeAdjustCvmAmount = willChangeCvm
		default:
			logs.Errorf("failed to construct adjust updated data. unsupported adjust type: %s, rid: %s",
				adjustObj.AdjustType, kt.Rid)
			return nil, nil, fmt.Errorf("unsupported adjust type: %s", adjustObj.AdjustType)
		}

		srcData = append(srcData, srcItem)
		updatedData = append(updatedData, updatedItem)
	}

	// 3.2. 构造追加的调整量
	for _, appendItem := range c.appendUpdateDemand {
		updatedData = append(updatedData, appendItem)
	}

	return srcData, updatedData, nil
}

// getCRPAdjustAbleDemandsForAdjustDemands get all adjustable demands from CRP based on the request's adjust demands.
func (c *CrpTicketCreator) getAllCRPAdjustAbleDemands(kt *kit.Kit, ticketType enumor.RPTicketType,
	demands rpt.ResPlanDemands, planProductName string, opProductName string) error {

	for _, demand := range demands {
		// 在转移拆单模式下，出现original为空是正常的，logs记录即可，不需要报错
		if demand.Original == nil {
			logs.Warnf("failed to construct adjust request, demand original is nil, rid: %s", kt.Rid)
			continue
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
		AbleDemandsRst, err := c.queryAdjustAbleDemands(kt, adjustAbleReq)
		if err != nil {
			logs.Errorf("failed to query adjust able demands, err: %v, req: %+v, rid: %s", err, adjustAbleReq,
				kt.Rid)
			return err
		}

		// 对AbleDemandsRst进行排序，returnPlanTime和本地相同的预测优先
		slices.SortFunc(AbleDemandsRst, func(a, b *cvmapi.CvmCbsPlanQueryItem) int {
			if a.ReturnPlanTime == demand.Original.ReturnPlanTime &&
				b.ReturnPlanTime != demand.Original.ReturnPlanTime {
				return -1
			}
			if a.ReturnPlanTime != demand.Original.ReturnPlanTime &&
				b.ReturnPlanTime == demand.Original.ReturnPlanTime {
				return 1
			}
			return 0
		})

		adjustAbleDemands := make([]*cvmapi.CvmCbsPlanQueryItem, 0)
		// 仅延期和转移单，可使用“已评审”的CRP预测进行修改
		// 转移单不可使用“未评审”的CRP预测
		for _, ad := range AbleDemandsRst {
			switch ad.ReviewStatus {
			case enumor.ResPlanReviewStatusPass:
				if ticketType != enumor.RPTicketTypeTransfer && ticketType != enumor.RPTicketTypeDelay {
					continue
				}
			case enumor.ResPlanReviewStatusPending:
				if ticketType == enumor.RPTicketTypeTransfer {
					continue
				}
			default:
				logs.Errorf("unsupported review status: %s, id: %s, rid: %s", ad.ReviewStatus, ad.SliceId, kt.Rid)
				return fmt.Errorf("unsupported review status: %s", ad.ReviewStatus)
			}
			adjustAbleDemands = append(adjustAbleDemands, ad)
		}

		c.adjustAbleDemands[demand.Original.DemandID] = adjustAbleDemands
	}

	return nil
}

// constructAdjustDemandDetails 构造调整预测请求参数
// 因crp不支持部分调整，这里通过将crp中的预测（可能有多条）调减调整的原始量，再在crp中追加一条调整的目标量，来间接实现部分调整。
// 在 prePrepareAdjustAbleData 中计算并记录crp中原始预测的调减量，将追加的目标量暂存在 appendUpdateDemand 中
func (c *CrpTicketCreator) constructAdjustDemandDetails(kt *kit.Kit, subTicket *ptypes.SubTicketInfo,
	demand rpt.ResPlanDemand) error {

	// isTransfer: 涉及中转产品转移操作的需求，尽量不使用未评审的预测
	// isAppend: 所有调整前后会造成预算追加的需求，尽量不使用已评审的预测
	var isTransfer, isAppend bool
	var adjustType enumor.CrpAdjustType
	switch subTicket.Type {
	case enumor.RPTicketTypeAdjust:
		isAppend = true
		adjustType = enumor.CrpAdjustTypeUpdate
	case enumor.RPTicketTypeDelay:
		adjustType = enumor.CrpAdjustTypeDelay
		// 这里会包含主分类不变的修改需求，并不属于延期，需将其修改回常规修改类型
		// 将这种修改需求标记为延期主要为了识别不会造成预算追加的需求
		if demand.Updated.ExpectTime == demand.Original.ExpectTime {
			adjustType = enumor.CrpAdjustTypeUpdate
		}
	case enumor.RPTicketTypeDelete:
		isAppend = true
		// 我们的删除对crp来说是部分调减
		adjustType = enumor.CrpAdjustTypeUpdate
	case enumor.RPTicketTypeTransfer:
		isTransfer = true
		adjustType = enumor.CrpAdjustTypeTransfer
	default:
		logs.Errorf("unsupported sub ticket type: %s， rid: %s", subTicket.Type, kt.Rid)
		return errors.New("unsupported sub ticket type")
	}

	// 预处理，计算crp中的预测调减结果
	if demand.Original == nil {
		// 追加需求，从中转产品转入
		if adjustType == enumor.CrpAdjustTypeTransfer {
			err := c.prePrepareTransferableData(kt, subTicket.ID, demand)
			if err != nil {
				logs.Errorf("failed to pre prepare transferable data, err: %v, rid: %s", err, kt.Rid)
				return err
			}
		}
	} else {
		// 修改，从可修改预测中调减
		err := c.prePrepareAdjustAbleData(kt, isTransfer, isAppend, adjustType, demand)
		if err != nil {
			logs.Errorf("failed to pre prepare adjust able data, err: %v, rid: %s", err, kt.Rid)
			return err
		}
	}

	// 加急延期、删除、转移时不追加预测
	if adjustType == enumor.CrpAdjustTypeDelay || demand.Updated == nil || subTicket.Type == enumor.RPTicketTypeTransfer {
		return nil
	}

	// 追加一条等于调整目标量的预测
	addItem, err := c.constructAdjustAppendData(kt, subTicket, demand)
	if err != nil {
		logs.Errorf("failed to construct adjust append data, err: %v, rid: %s", err, kt.Rid)
		return err
	}

	c.appendUpdateDemand = append(c.appendUpdateDemand, addItem)

	return nil
}

// prePrepareTransferableData 预处理可转移的数据.
// 适用于多个子订单会重复调整到同一条预测数据的场景，先将所有可能的影响汇总到 adjCRPDemandsRst 中
// TODO 逻辑和 splitter.matchTransferCRPDemands 重复，再抽一层
func (c *CrpTicketCreator) prePrepareTransferableData(kt *kit.Kit, id string, demand rpt.ResPlanDemand) error {

	if demand.Updated == nil {
		logs.Errorf("updated demand is nil, ticket id: %s, rid: %s", id, kt.Rid)
		return errors.New("updated demand is nil")
	}

	needDemand := demand.Updated
	expectYear, err := strconv.Atoi(needDemand.ExpectTime[:4])
	if err != nil {
		logs.Errorf("failed to parse expect year, err: %v, expect_time: %s, rid: %s", err, needDemand.ExpectTime,
			kt.Rid)
		return err
	}
	// 遍历可用于调减的crp预测，凑齐调减总量
	needCpuCores := demand.Updated.Cvm.CpuCore
	for _, transAbleD := range c.transferAbleDemands[expectYear] {
		if needCpuCores <= 0 {
			break
		}

		// 未评审需求跳过，不记录
		if transAbleD.ReviewStatus == enumor.ResPlanReviewStatusPending {
			continue
		}

		var canConsume int64
		// 项目类型和技术大类需一致
		if transAbleD.ProjectName != needDemand.ObsProject ||
			transAbleD.TechnicalClass != needDemand.Cvm.TechnicalClass {
			continue
		}

		remainedCpuCores := transAbleD.RealCoreAmount
		if _, ok := c.adjCRPDemandsRst[transAbleD.SliceId]; ok {
			remainedCpuCores -= c.adjCRPDemandsRst[transAbleD.SliceId].WillConsume
		}

		canConsume = min(needCpuCores, remainedCpuCores)
		// CvmAmount虽然理论上大于等于RealCoreAmount，但是为确保后续除法计算不出异常，判断下CvmAmount的大小
		if canConsume <= 0 || transAbleD.CvmAmount == 0 {
			continue
		}

		if _, ok := c.adjCRPDemandsRst[transAbleD.SliceId]; !ok {
			c.adjCRPDemandsRst[transAbleD.SliceId] = &AdjustAbleRemainObj{
				OriginDemand:   transAbleD.Clone(),
				TransferTarget: make(map[rpt.UpdatedRPDemandItem]int64),
			}
		}
		adjustAbleRemain := c.adjCRPDemandsRst[transAbleD.SliceId]
		adjustAbleRemain.WillConsume += canConsume
		adjustAbleRemain.TransferTarget[cvt.PtrToVal(needDemand.Clone())] = canConsume
		needCpuCores -= canConsume
	}

	if needCpuCores > 0 {
		logs.Errorf("crp demand remained is not enough to deduction, adjust: %+v, need cpu cores: %d, rid: %s",
			needDemand.Cvm, needCpuCores, kt.Rid)
		return fmt.Errorf("crp demand remained is not enough to deduction, adjust cores: %d, need cores: %d",
			needDemand.Cvm.CpuCore, needCpuCores)
	}

	return nil
}

// prePrepareAdjustAbleData 预处理可调减的数据.
// 适用于多个子订单会重复调整到同一条预测数据的场景，先将所有可能的影响汇总到 adjCRPDemandsRst 中
// TODO 逻辑和 splitter.matchReviewedCRPDemands 类似，再抽一层
func (c *CrpTicketCreator) prePrepareAdjustAbleData(kt *kit.Kit, isTransfer, isAppend bool,
	adjustType enumor.CrpAdjustType, demand rpt.ResPlanDemand) error {

	adjustAbleDemands, ok := c.adjustAbleDemands[demand.Original.DemandID]
	if !ok {
		logs.Errorf("failed to get adjust able demands, demand id: %s, rid: %s", demand.Original.DemandID, kt.Rid)
		return errors.New("no adjust able demands")
	}

	// 查询原始预测的预测内外情况
	demandDetail, err := c.resFetcher.GetResPlanDemandDetail(kt, demand.Original.DemandID, []int64{})
	if err != nil {
		logs.Errorf("failed to get res plan demand detail, err: %v, demand_id: %s, rid: %s", err,
			demand.Original.DemandID, kt.Rid)
		return err
	}

	// 遍历可用于调减的crp预测，凑齐调减总量
	needCpuCores := demand.Original.Cvm.CpuCore
	for _, adjustAbleD := range adjustAbleDemands {
		if needCpuCores <= 0 {
			break
		}

		// 转移时，不操作未评审的预测
		if isTransfer {
			if adjustAbleD.ReviewStatus == enumor.ResPlanReviewStatusPending {
				continue
			}
		}
		// 非延期、转移场景（删除、修改），不操作已评审的预测，避免影响预算
		if isAppend {
			if adjustAbleD.ReviewStatus == enumor.ResPlanReviewStatusPass {
				continue
			}
		}

		// 取消类型的调整单据，不需要给expectTime
		var updateExpectTime string
		if demand.Updated != nil {
			updateExpectTime = demand.Updated.ExpectTime
		}

		canConsume, err := c.calcAdjustAbleDCanConsumeCPU(kt, adjustAbleD, adjustType, needCpuCores, demandDetail,
			updateExpectTime)
		if err != nil {
			logs.Warnf("adjust type is delay, but updated is nil, need adjust: %+v, slice: %s, rid: %s",
				cvt.PtrToVal(demand.Original), adjustAbleD.SliceId, kt.Rid)
			continue
		}
		logs.Infof("pre prepare adjust able data, adjust_able_demand: %+v, rid: %s", adjustAbleD, kt.Rid)
		logs.Infof("pre prepare adjust able data, demand: %+v, needCpuCores: %d, canConsume: %d, rid: %s",
			demand.Original, needCpuCores, canConsume, kt.Rid)
		// 短租预测场景，如果即将修改的crp预测的预期退回时间与本地不一致，给出警告
		if demandDetail.ObsProject == enumor.ObsProjectShortLease &&
			adjustAbleD.ReturnPlanTime != cvt.PtrToVal(demandDetail.ReturnPlanTime) {
			logs.Warnf("pre prepare adjust able data, return plan time is not equal, local return time: %s, "+
				"crp return time: %s, rid: %s", demandDetail.ReturnPlanTime, adjustAbleD.ReturnPlanTime, kt.Rid)
		}

		needCpuCores -= canConsume
	}

	if needCpuCores > 0 {
		logs.Errorf("crp demand remained is not enough to deduction, adjust: %+v, need cpu cores: %d, rid: %s",
			demand.Original.Cvm, needCpuCores, kt.Rid)
		return fmt.Errorf("crp demand remained is not enough to deduction, adjust cores: %d, need cores: %d",
			demand.Original.Cvm.CpuCore, needCpuCores)
	}

	return nil
}

// calcAdjustAbleDCanConsumeCPU 计算单个可调整的crp预测，有多少CPU可以被用在demand的调整中
// 并将结果暂存在 adjCRPDemandsRst 中
func (c *CrpTicketCreator) calcAdjustAbleDCanConsumeCPU(kt *kit.Kit, adjustAbleD *cvmapi.CvmCbsPlanQueryItem,
	adjustType enumor.CrpAdjustType, needCpuCores int64, demandDetail *ptypes.GetPlanDemandDetailResp,
	updateExpectTime string) (int64, error) {

	var canConsume int64
	// 磁盘类型需一致，未知的磁盘类型（包括空值）除外
	if adjustAbleD.DiskType.Name() != demandDetail.DiskTypeName {
		// 加急延期场景CRP会根据原预测的磁盘类型创建新的预测，此时需保证磁盘类型完全一致，不能为空
		if adjustType == enumor.CrpAdjustTypeDelay || adjustAbleD.DiskType.Name() != "" {
			return canConsume, nil
		}
	}

	// 预测内外需一致
	if demandDetail.PlanType.GetCode() != enumor.PlanType(adjustAbleD.InPlan).GetCode() {
		return canConsume, nil
	}

	remainedCpuCores := adjustAbleD.RealCoreAmount
	if _, ok := c.adjCRPDemandsRst[adjustAbleD.SliceId]; ok {
		remainedCpuCores -= c.adjCRPDemandsRst[adjustAbleD.SliceId].WillConsume
	}

	canConsume = min(needCpuCores, remainedCpuCores)
	// CvmAmount虽然理论上大于等于RealCoreAmount，但是为确保后续除法计算不出异常，判断下CvmAmount的大小
	if canConsume <= 0 || adjustAbleD.CvmAmount == 0 {
		return 0, nil
	}

	if _, ok := c.adjCRPDemandsRst[adjustAbleD.SliceId]; !ok {
		c.adjCRPDemandsRst[adjustAbleD.SliceId] = &AdjustAbleRemainObj{
			OriginDemand: adjustAbleD.Clone(),
			AdjustType:   adjustType,
		}
		// 取消类型的调整单据，不需要expectTime
		if updateExpectTime != "" {
			c.adjCRPDemandsRst[adjustAbleD.SliceId].ExpectTime = updateExpectTime
		}
	}
	adjustAbleRemain := c.adjCRPDemandsRst[adjustAbleD.SliceId]

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

// constructAdjustAppendData construct adjust append data.
func (c *CrpTicketCreator) constructAdjustAppendData(kt *kit.Kit, subTicket *ptypes.SubTicketInfo,
	demand rpt.ResPlanDemand) (*cvmapi.AdjustUpdatedData, error) {

	// 根据HCM的diskType，生成CRP的diskType和diskName
	diskTypeName := demand.Updated.Cbs.DiskType.Name()
	diskType, err := enumor.GetCRPDiskTypeFromCRPName(diskTypeName)
	if err != nil {
		return nil, err
	}

	demandItem := &cvmapi.CvmCbsPlanQueryItem{
		SliceId:         uuid.UUID(),
		ProjectName:     demand.Updated.ObsProject,
		PlanProductId:   int(subTicket.PlanProductID),
		PlanProductName: subTicket.PlanProductName,
		ProductId:       int(subTicket.OpProductID),
		ProductName:     subTicket.OpProductName,
		InstanceType:    demand.Updated.Cvm.DeviceClass,
		CityId:          0,
		CityName:        demand.Updated.RegionName,
		ZoneId:          0,
		ZoneName:        demand.Updated.ZoneName,
		InstanceModel:   demand.Updated.Cvm.DeviceType,
		UseTime:         demand.Updated.ExpectTime,
		CvmAmount:       demand.Updated.Cvm.Os.InexactFloat64(),
		InstanceIO:      int(demand.Updated.Cbs.DiskIo),
		DiskType:        diskType,
		DiskTypeName:    diskTypeName,
		AllDiskAmount:   demand.Updated.Cbs.DiskSize,
	}
	// 修改场景追加调整后的预测时，需要提供预期退回时间参数
	if demand.Updated.ObsProject == enumor.ObsProjectShortLease {
		demandItem.IsAutoReturnPlan = true
		demandItem.ReturnPlanTime = demand.Updated.ReturnPlanTime
		if demand.Updated.ReturnPlanTime == "" {
			return nil, errors.New("short-term lease project must provide return plan time")
		}
	}

	updatedData := &cvmapi.AdjustUpdatedData{
		AdjustType:          string(enumor.CrpAdjustTypeUpdate),
		CvmCbsPlanQueryItem: demandItem,
	}

	return updatedData, nil
}

// constructTransferAppendData construct transfer append data. Use for transfer to biz.
func (c *CrpTicketCreator) constructTransferAppendDataToBiz(kt *kit.Kit, subTicket *ptypes.SubTicketInfo,
	deviceCore float64, source *cvmapi.CvmCbsPlanQueryItem, transferTarget map[rpt.UpdatedRPDemandItem]int64) (
	[]*cvmapi.AdjustUpdatedData, error) {

	allAppendData := make([]*cvmapi.AdjustUpdatedData, 0)
	for key, transferCore := range transferTarget {
		// 和CRP确认保留4位小数可以，但是肯定会存在误差
		transferCVM, err := math.RoundToDecimalPlaces(float64(transferCore)/deviceCore, 4)
		if err != nil {
			logs.Errorf("failed to round change cvm to 4 decimal places, err: %v, demand: %+v, change cvm: %f, "+
				"rid: %s", err, key, float64(transferCore)/deviceCore, kt.Rid)
			return nil, err
		}

		// 根据HCM的diskType，生成CRP的diskType和diskName
		diskTypeName := key.Cbs.DiskType.Name()
		diskType, err := enumor.GetCRPDiskTypeFromCRPName(diskTypeName)
		if err != nil {
			return nil, err
		}

		// TODO 目前转移后的机型以中转产品中的机型为准，实际应以业务提单为准；
		//  待基准由核数调整为技术大类资源基准（如高IO的云盘数、GPU的卡数）后再做调整
		demandItem := &cvmapi.CvmCbsPlanQueryItem{
			// 转移中追加的目标数据sliceID给默认值，同一个修改单的updated中不能有两个相同的sliceID
			SliceId:         uuid.UUID(),
			ProjectName:     key.ObsProject,
			PlanProductId:   int(subTicket.PlanProductID),
			PlanProductName: subTicket.PlanProductName,
			ProductId:       int(subTicket.OpProductID),
			ProductName:     subTicket.OpProductName,
			InstanceType:    source.InstanceType,
			CityId:          0,
			CityName:        key.RegionName,
			ZoneId:          0,
			ZoneName:        key.ZoneName,
			InstanceModel:   source.InstanceModel,
			UseTime:         key.ExpectTime,
			ReturnPlanTime:  key.ReturnPlanTime,
			CvmAmount:       transferCVM,
			InstanceIO:      int(key.Cbs.DiskIo),
			DiskType:        diskType,
			DiskTypeName:    key.Cbs.DiskTypeName,
			// TODO 用户的云盘需求会在这里被丢弃，避免出现一对多的情况下多次提交CBS需求
			AllDiskAmount: 0,
		}

		// 短租项目预测需要提供isAutoReturnPlan和returnPlanTime参数
		if key.ObsProject == enumor.ObsProjectShortLease {
			demandItem.IsAutoReturnPlan = true
		}

		// TODO 临时处理：标准型机器使用核心数作为技术大类，因此可以直接以业务提单为准
		if key.Cvm.DeviceFamily == string(enumor.DeviceFamilyStandard) {
			expectDeviceCore := decimal.NewFromInt(key.Cvm.CpuCore).Div(key.Cvm.Os.Decimal).InexactFloat64()
			transferCVM, err = math.RoundToDecimalPlaces(float64(transferCore)/expectDeviceCore, 4)
			if err != nil {
				logs.Errorf("failed to round change cvm to 4 decimal places, err: %v, demand: %+v, change cvm: %f, "+
					"rid: %s", err, key, float64(transferCore)/expectDeviceCore, kt.Rid)
				return nil, err
			}
			demandItem.InstanceType = key.Cvm.DeviceClass
			demandItem.InstanceModel = key.Cvm.DeviceType
			demandItem.CvmAmount = transferCVM
		}

		allAppendData = append(allAppendData, &cvmapi.AdjustUpdatedData{
			AdjustType:          string(enumor.CrpAdjustTypeUpdate),
			CvmCbsPlanQueryItem: demandItem,
		})
	}

	return allAppendData, nil
}

// constructTransferAppendData construct transfer append data. Use for transfer to pool.
func (c *CrpTicketCreator) constructTransferAppendDataToPool(source *cvmapi.CvmCbsPlanQueryItem,
	targetOS float64, targetCPU int64) *cvmapi.AdjustUpdatedData {

	demandItem := &cvmapi.CvmCbsPlanQueryItem{
		// 转移中追加的目标数据sliceID给默认值，同一个修改单的updated中不能有两个相同的sliceID
		SliceId:         uuid.UUID(),
		ProjectName:     source.ProjectName,
		PlanProductId:   cvmapi.TransferPlanProductID,
		PlanProductName: cvmapi.TransferPlanProductName,
		ProductId:       cvmapi.TransferOpProductID,
		ProductName:     cvmapi.TransferOpProductName,
		InstanceType:    source.InstanceType,
		CityId:          0,
		CityName:        source.CityName,
		ZoneId:          0,
		ZoneName:        source.ZoneName,
		InstanceModel:   source.InstanceModel,
		UseTime:         source.UseTime,
		ReturnPlanTime:  source.ReturnPlanTime,
		CvmAmount:       targetOS,
		CoreAmount:      targetCPU,
		InstanceIO:      source.InstanceIO,
		DiskType:        source.DiskType,
		DiskTypeName:    source.DiskTypeName,
		AllDiskAmount:   0,
	}

	// 短租项目预测需要提供isAutoReturnPlan和returnPlanTime参数
	if source.ProjectName == enumor.ObsProjectShortLease {
		demandItem.IsAutoReturnPlan = true
	}

	updatedData := &cvmapi.AdjustUpdatedData{
		AdjustType:          string(enumor.CrpAdjustTypeUpdate),
		CvmCbsPlanQueryItem: demandItem,
	}

	return updatedData
}
