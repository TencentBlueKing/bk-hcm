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

	dt "hcm/pkg/api/core/cloud/device-type"
	"hcm/pkg/criteria/enumor"
	rpt "hcm/pkg/dal/table/resource-plan/res-plan-ticket"
	ttypes "hcm/pkg/dal/table/types"
	"hcm/pkg/kit"
	"hcm/pkg/logs"
	"hcm/pkg/thirdparty/cvmapi"

	"github.com/shopspring/decimal"
)

// SplitDeleteTicket split res plan delete ticket to sub ticket.
// The virtualDeptID parameter is used to gate transfer pool logic: only dept 1041 (CvmCbsPlanDeptId) may
// transfer demands into the transfer pool. Other depts skip the transfer branch entirely.
func (s *SubTicketSplitter) SplitDeleteTicket(kt *kit.Kit, ticketID string, virtualDeptID int64,
	ticketType enumor.RPTicketType, demands rpt.ResPlanDemands, planProductName, opProductName string) error {

	// 1. 准备拆分后的子单，存储在 adjSplitGroupDemands 中备用
	err := s.prepareDeleteSubTickets(kt, ticketID, virtualDeptID, ticketType, demands, planProductName, opProductName)
	if err != nil {
		logs.Errorf("failed to prepare delete sub tickets, err: %v, rid: %s", err, kt.Rid)
		return err
	}

	// 2. 对所有拆分后的变更需求，整理并创建出1～2个子单
	err = s.createSubTicket(kt, ticketID, demands, enumor.RPTicketTypeDelete)
	if err != nil {
		logs.Errorf("failed to create sub ticket, err: %v, rid: %s", err, kt.Rid)
		return err
	}

	return nil
}

// prepareDeleteSubTickets 准备删除场景下拆分后的子单，存储在 adjSplitGroupDemands 中备用
// 该方法可在调整场景复用
// The virtualDeptID parameter gates transfer pool logic: only dept CvmCbsPlanDeptId (1041) may transfer
// demands into the transfer pool. Non-1041 depts bypass the transfer matching and proceed as regular deletes.
func (s *SubTicketSplitter) prepareDeleteSubTickets(kt *kit.Kit, ticketID string, virtualDeptID int64,
	ticketType enumor.RPTicketType, demands rpt.ResPlanDemands, planProductName, opProductName string) error {

	// Only dept 1041 (CvmCbsPlanDeptId) may transfer demands into the transfer pool.
	// For other depts, skip the transfer matching and put all demands directly into the regular delete group.
	if virtualDeptID != cvmapi.CvmCbsPlanDeptId {
		for idx := range demands {
			s.adjSplitGroupDemands[enumor.RPTicketTypeDelete] = append(
				s.adjSplitGroupDemands[enumor.RPTicketTypeDelete], &demands[idx])
		}
		return nil
	}

	// 1.查询CRP中可修改的预测
	err := s.getAllCRPAdjustAbleDemands(kt, demands, planProductName, opProductName)
	if err != nil {
		logs.Errorf("failed to query adjust able demands, err: %v, demands: %+v, rid: %s", err, demands, kt.Rid)
		return err
	}

	// 1.1. 获取机型信息列表，用于后续操作
	deviceTypeMap, err := s.deviceTypes.GetDeviceTypes(kt)
	if err != nil {
		logs.Errorf("get device type map failed, err: %v, rid: %s", err, kt.Rid)
		return err
	}

	// 2. 对每个变更需求，匹配已评审的CRP预测，并按照已评审和未评审拆分为2个子单
	for _, demand := range demands {
		transferableCore, nonTransferableCore, err := s.matchReviewedCRPDemands(kt, demand, enumor.CrpAdjustTypeCancel)
		if err != nil {
			logs.Errorf("failed to match reviewed crp demands, err: %v, demand: %+v, rid: %s", err, demand, kt.Rid)
			return err
		}

		err = s.splitDemandInDeleteScenarios(kt, ticketID, ticketType, demand, transferableCore, nonTransferableCore,
			deviceTypeMap)
		if err != nil {
			logs.Errorf("failed to split demand in delete scenarios, err: %v, ticket id: %s, update: %+v, rid: %s",
				err, ticketID, demand.Updated, kt.Rid)
			return err
		}
	}

	return nil
}

// splitDemandInDeleteScenarios 在删除场景下，根据本业务中已评审的CPU核数将需求拆分到 adjSplitGroupDemands 备用
// 删除场景只会从本业务转移到中转产品，因此拆分时仅需关注demand的original部分，即纯删除的部分
func (s *SubTicketSplitter) splitDemandInDeleteScenarios(kt *kit.Kit, ticketID string, ticketType enumor.RPTicketType,
	demand rpt.ResPlanDemand, transferableCore, nonTransferableCore int64,
	deviceTypeMap map[string]dt.DistinctDeviceType) error {

	if demand.Original == nil {
		logs.Errorf("original is nil, ticket id: %s, updated: %+v, rid: %s", ticketID, demand.Updated, kt.Rid)
		return errors.New("demand original is nil")
	}

	deviceType := demand.Original.Cvm.DeviceType
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
		transferDemand.Original.Cvm.CpuCore = transferableCore
		transferDemand.Original.Cvm.Os = ttypes.Decimal{Decimal: transferableOS}
		transferDemand.Original.Cvm.Memory = transferableMem
		transferType := enumor.RPTicketTypeTransferOUT
		// 当主单类型为 automatic_transfer 时，转出单免审
		if ticketType == enumor.RPTicketTypeAutomaticTransfer {
			transferType = enumor.RPTicketTypeTransferOUTExempt
		}
		s.adjSplitGroupDemands[transferType] = append(
			s.adjSplitGroupDemands[transferType], transferDemand)
	}

	// 不可转移的预测
	if nonTransferableCore > 0 {
		nonTransferableOS := decimal.NewFromInt(nonTransferableCore).Div(decimal.NewFromInt(deviceCPUCore))
		nonTransferableMem := nonTransferableOS.Mul(decimal.NewFromInt(deviceMemory)).IntPart()

		nonTransferDemand := demand.Clone()
		nonTransferDemand.Original.Cvm.CpuCore = nonTransferableCore
		nonTransferDemand.Original.Cvm.Os = ttypes.Decimal{Decimal: nonTransferableOS}
		nonTransferDemand.Original.Cvm.Memory = nonTransferableMem
		// 当存在可转移预测时，避免云盘二次申请
		nonTransferDemand.Original.Cbs.DiskSize = 0
		s.adjSplitGroupDemands[enumor.RPTicketTypeDelete] = append(s.adjSplitGroupDemands[enumor.RPTicketTypeDelete],
			nonTransferDemand)
	}
	return nil
}
