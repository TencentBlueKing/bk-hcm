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

package splitter

import (
	"errors"
	"fmt"
	"strconv"
	"time"

	ptypes "hcm/cmd/woa-server/types/plan"
	"hcm/pkg/criteria/enumor"
	rpt "hcm/pkg/dal/table/resource-plan/res-plan-ticket"
	wdt "hcm/pkg/dal/table/resource-plan/woa-device-type"
	ttypes "hcm/pkg/dal/table/types"
	"hcm/pkg/kit"
	"hcm/pkg/logs"

	"github.com/shopspring/decimal"
)

// SplitAddTicket split res plan append ticket to sub ticket
func (s *SubTicketSplitter) SplitAddTicket(kt *kit.Kit, ticketID string, demands rpt.ResPlanDemands) error {

	// 1. 准备拆分后的子单，存储在 adjSplitGroupDemands 中备用
	_, _, err := s.prepareAddSubTickets(kt, ticketID, demands)
	if err != nil {
		logs.Errorf("failed to prepare add sub tickets, err: %v, rid: %s", err, kt.Rid)
		return err
	}

	// 2. 对每个变更需求，匹配可转移的CRP预测，并按照可转移和不可转移拆分为2个子单
	err = s.createSubTicket(kt, ticketID, demands, enumor.RPTicketTypeAdd)
	if err != nil {
		logs.Errorf("failed to create sub ticket, err: %v, rid: %s", err, kt.Rid)
		return err
	}

	return nil
}

// prepareAddSubTickets 准备追加场景下拆分后的子单，存储在 adjSplitGroupDemands 中备用
// 该方法可在调整场景复用
func (s *SubTicketSplitter) prepareAddSubTickets(kt *kit.Kit, ticketID string, demands rpt.ResPlanDemands) (
	bool, rpt.ResPlanDemands, error) {

	// 1.1. 获取机型信息列表，用于后续操作
	deviceTypeMap, err := s.deviceTypes.GetDeviceTypes(kt)
	if err != nil {
		logs.Errorf("get device type map failed, err: %v, rid: %s", err, kt.Rid)
		return false, nil, err
	}

	// 1.2. 区分并处理纯 CBS 和包含 CVM 的需求
	cvmDemands, obsProjects, technicalClasses, err := s.separateAndProcessDemands(kt, ticketID, demands)
	if err != nil {
		return false, nil, err
	}

	// 如果没有 CVM 需求，直接返回
	if len(cvmDemands) == 0 {
		return false, nil, nil
	}

	// 1.4. 额度池余额判断
	// TODO 额度需请求CRP接口（QueryCvmCbsPlans），可以和下方的 queryTransferCRPDemands 合并
	canTransfer, err := s.canTransferByQuota(kt, obsProjects)
	if err != nil {
		logs.Errorf("failed to check quota, err: %v, ticket id: %s, rid: %s", err, ticketID, kt.Rid)
		return false, nil, err
	}

	// 2. 查询CRP中可转移的预测（中转产品）
	if canTransfer {
		err := s.queryTransferCRPDemands(kt, obsProjects, technicalClasses)
		if err != nil {
			logs.Errorf("query crp demands failed, err: %v, ticket id: %s, rid: %s", err, ticketID, kt.Rid)
			return canTransfer, nil, err
		}
		// 2.1. 对每个变更需求，匹配可转移的CRP预测，匹配不完的拆分为另一个子单
		for _, demand := range cvmDemands {
			transferableCore, nonTransferableCore, err := s.matchTransferCRPDemands(kt, ticketID, demand)
			if err != nil {
				logs.Errorf("failed to match transfer crp demands, err: %v, ticket id: %s, update: %+v, rid: %s",
					err, ticketID, demand.Updated, kt.Rid)
				return canTransfer, nil, err
			}

			err = s.splitDemandInAddScenarios(kt, ticketID, demand, transferableCore, nonTransferableCore,
				deviceTypeMap)
			if err != nil {
				logs.Errorf("failed to split demand in add scenarios, err: %v, ticket id: %s, update: %+v, rid: %s",
					err, ticketID, demand.Updated, kt.Rid)
				return canTransfer, nil, err
			}
		}
	}

	return canTransfer, cvmDemands, nil
}

// separateAndProcessDemands 分离并处理仅包含 CBS 的和包含 CVM 的需求
func (s *SubTicketSplitter) separateAndProcessDemands(kt *kit.Kit, ticketID string, demands rpt.ResPlanDemands) (
	cvmDemands rpt.ResPlanDemands, obsProjects []enumor.ObsProject, technicalClasses []string, err error) {

	obsProjects = make([]enumor.ObsProject, 0)
	technicalClasses = make([]string, 0)
	cbsOnlyDemands := make([]*rpt.ResPlanDemand, 0)
	cvmDemands = make(rpt.ResPlanDemands, 0)

	for idx, d := range demands {
		if d.Updated == nil {
			logs.Errorf("updated demand is nil, ticket id: %s, rid: %s", ticketID, kt.Rid)
			return cvmDemands, obsProjects, technicalClasses, errors.New("updated demand is nil")
		}

		// 仅包含 CBS 需求，单独处理
		if d.Updated.Cvm.IsEmpty() {
			cbsOnlyDemands = append(cbsOnlyDemands, &demands[idx])
			continue
		}

		// 对 CVM 需求，拆分出 CBS 部分单独处理
		if d.Updated.Cbs.DiskSize > 0 {
			cbsDemand := d.ExtractCbsDemand()
			cbsOnlyDemands = append(cbsOnlyDemands, cbsDemand)
			// 将原需求的 CBS 置为 0
			demands[idx].ClearCbsDemand()
		}

		cvmDemands = append(cvmDemands, demands[idx])
		obsProjects = append(obsProjects, d.Updated.ObsProject)
		technicalClasses = append(technicalClasses, d.Updated.Cvm.TechnicalClass)
	}

	// 处理仅包含 CBS 的需求，直接放入 adjSplitGroupDemands，不需要考虑转移逻辑
	if len(cbsOnlyDemands) > 0 {
		s.adjSplitGroupDemands[enumor.RPTicketTypeAdd] = append(s.adjSplitGroupDemands[enumor.RPTicketTypeAdd],
			cbsOnlyDemands...)
	}

	return cvmDemands, obsProjects, technicalClasses, nil
}

// canTransferByQuota 根据额度判断是否可通过转移进行预测追加
func (s *SubTicketSplitter) canTransferByQuota(kt *kit.Kit, obsProjects []enumor.ObsProject) (bool, error) {
	transferQuota, err := s.resFetcher.GetPlanTransferQuotaConfigs(kt)
	if err != nil {
		logs.Errorf("get plan transfer quota configs failed, err: %v, rid: %s", err, kt.Rid)
		return false, err
	}

	listQuotaReq := &ptypes.ListResPlanTransferQuotaSummaryReq{
		Year:       int64(time.Now().Year()),
		ObsProject: obsProjects,
	}
	remainQuotaObj, err := s.resFetcher.ListRemainTransferQuota(kt, listQuotaReq)
	if err != nil {
		logs.Errorf("list remain transfer quota failed, err: %v, req: %+v, rid: %s", err, listQuotaReq, kt.Rid)
		return false, err
	}

	canTransfer := true
	// 剩余额度小于可转移额度
	if remainQuotaObj.RemainQuota < transferQuota.Quota {
		canTransfer = false
	}
	return canTransfer, nil
}

// queryCRPDemands 查询CRP中的预测，按照项目类型和技术大类
func (s *SubTicketSplitter) queryTransferCRPDemands(kt *kit.Kit, obsProjects []enumor.ObsProject,
	technicalClasses []string) error {

	crpDemands, err := s.resFetcher.QueryCRPTransferPoolDemands(kt, obsProjects, technicalClasses)
	if err != nil {
		logs.Errorf("failed to query transfer pool demands, err: %v, obs_project: %v, technical_classes: %v, rid: %s",
			err, obsProjects, technicalClasses, kt.Rid)
		return err
	}

	for _, demand := range crpDemands {
		s.transferAbleDemands[demand.Year] = append(s.transferAbleDemands[demand.Year], demand)
	}
	return nil
}

// matchTransferCRPDemands 从所有可用于转移的CRP预测中，匹配相同项目类型和技术大类的，并计算返回可转移和不可转移的核数
func (s *SubTicketSplitter) matchTransferCRPDemands(kt *kit.Kit, ticketID string, demand rpt.ResPlanDemand) (
	transferableCore int64, nonTransferableCore int64, err error) {

	if demand.Updated == nil {
		logs.Errorf("updated demand is nil, ticket id: %s, rid: %s", ticketID, kt.Rid)
		return transferableCore, nonTransferableCore, errors.New("updated demand is nil")
	}

	needDemand := demand.Updated
	expectYear, err := strconv.Atoi(needDemand.ExpectTime[:4])
	if err != nil {
		logs.Errorf("failed to parse expect year, err: %v, expect_time: %s, rid: %s", err, needDemand.ExpectTime,
			kt.Rid)
		return transferableCore, nonTransferableCore, err
	}
	// 1. 计算可转移的预测量，将转移的消耗记录到 transferCRPDemandRst
	needCpuCores := needDemand.Cvm.CpuCore
	for _, transAbleD := range s.transferAbleDemands[expectYear] {
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
		if _, ok := s.transferCRPDemandRst[transAbleD.SliceId]; ok {
			remainedCpuCores -= s.transferCRPDemandRst[transAbleD.SliceId].WillConsume
		}

		canConsume = min(needCpuCores, remainedCpuCores)
		// CvmAmount虽然理论上大于等于RealCoreAmount，但是为确保后续除法计算不出异常，判断下CvmAmount的大小
		if canConsume <= 0 || transAbleD.CvmAmount == 0 {
			continue
		}

		if _, ok := s.transferCRPDemandRst[transAbleD.SliceId]; !ok {
			s.transferCRPDemandRst[transAbleD.SliceId] = &AdjustAbleRemainObj{
				OriginDemand: transAbleD.Clone(),
			}
		}
		adjustAbleRemain := s.transferCRPDemandRst[transAbleD.SliceId]
		adjustAbleRemain.WillConsume += canConsume
		needCpuCores -= canConsume
	}

	nonTransferableCore = needCpuCores
	transferableCore = needDemand.Cvm.CpuCore - needCpuCores
	return transferableCore, nonTransferableCore, nil
}

// splitDemandInAddScenarios 在新增场景下，根据中转产品中可转移的CPU核数将需求拆分到 adjSplitGroupDemands 备用
// 新增场景只会从中转产品转移到本业务，因此拆分时仅需关注demand的updated部分，即纯新增的部分
func (s *SubTicketSplitter) splitDemandInAddScenarios(kt *kit.Kit, ticketID string, demand rpt.ResPlanDemand,
	transferableCore, nonTransferableCore int64, deviceTypeMap map[string]wdt.WoaDeviceTypeTable) error {

	if demand.Updated == nil {
		logs.Errorf("updated is nil, ticket id: %s, original: %+v, rid: %s", ticketID, demand.Original, kt.Rid)
		return errors.New("demand updated is nil")
	}

	deviceType := demand.Updated.Cvm.DeviceType
	deviceCPUCore := deviceTypeMap[deviceType].CpuCore
	deviceMemory := deviceTypeMap[deviceType].Memory
	if deviceCPUCore == 0 {
		logs.Errorf("cannot found device type %s, cpu core is 0, rid: %s", deviceType, kt.Rid)
		return fmt.Errorf("cannot found device type: %s", deviceType)
	}

	// 2.3. 按照可转移和不可转移将该变更需求拆分为2部分，记录到 adjSplitGroupDemands
	// 可转移的预测
	if transferableCore > 0 {
		transferableOS := decimal.NewFromInt(transferableCore).Div(decimal.NewFromInt(deviceCPUCore))
		transferableMem := transferableOS.Mul(decimal.NewFromInt(deviceMemory)).IntPart()

		transferDemand := demand.Clone()
		transferDemand.Updated.Cvm.CpuCore = transferableCore
		transferDemand.Updated.Cvm.Os = ttypes.Decimal{Decimal: transferableOS}
		transferDemand.Updated.Cvm.Memory = transferableMem
		s.adjSplitGroupDemands[enumor.RPTicketTypeTransferIN] = append(
			s.adjSplitGroupDemands[enumor.RPTicketTypeTransferIN], transferDemand)
	}

	// 不可转移的预测
	if nonTransferableCore > 0 {
		nonTransferableOS := decimal.NewFromInt(nonTransferableCore).Div(decimal.NewFromInt(deviceCPUCore))
		nonTransferableMem := nonTransferableOS.Mul(decimal.NewFromInt(deviceMemory)).IntPart()

		nonTransferDemand := demand.Clone()
		nonTransferDemand.Updated.Cvm.CpuCore = nonTransferableCore
		nonTransferDemand.Updated.Cvm.Os = ttypes.Decimal{Decimal: nonTransferableOS}
		nonTransferDemand.Updated.Cvm.Memory = nonTransferableMem
		// 当存在可转移预测时，避免云盘二次申请
		nonTransferDemand.Updated.Cbs.DiskSize = 0
		s.adjSplitGroupDemands[enumor.RPTicketTypeAdd] = append(s.adjSplitGroupDemands[enumor.RPTicketTypeAdd],
			nonTransferDemand)
	}
	return nil
}
