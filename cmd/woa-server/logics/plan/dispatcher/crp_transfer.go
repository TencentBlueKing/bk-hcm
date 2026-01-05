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

	ptypes "hcm/cmd/woa-server/types/plan"
	"hcm/pkg/criteria/enumor"
	"hcm/pkg/kit"
	"hcm/pkg/logs"
	"hcm/pkg/thirdparty/cvmapi"
)

// createTransferOutCrpTicket create transfer out crp ticket.
func (c *CrpTicketCreator) createTransferOutCrpTicket(kt *kit.Kit, subTicket *ptypes.SubTicketInfo,
	srcData []*cvmapi.AdjustSrcData, updateData []*cvmapi.AdjustUpdatedData) (string, error) {

	transReq := c.constructTransReq(kt, subTicket, srcData, updateData)
	resp, err := c.crpCli.CreateTransOrder(kt.Ctx, kt.Header(), transReq)
	if err != nil {
		logs.Errorf("failed to transfer plan order, err: %v, sub_ticket_id: %s, rid: %s", err, subTicket.ID,
			kt.Rid)
		return "", err
	}

	if resp.Error.Code != 0 {
		logs.Errorf("failed to create transfer plan order, code: %d, msg: %s, crp_trace: %s, "+
			"sub_ticket_id: %s, rid: %s", resp.Error.Code, resp.Error.Message, resp.TraceId, subTicket.ID, kt.Rid)
		return "", fmt.Errorf("failed to create transfer plan order, code: %d, msg: %s", resp.Error.Code,
			resp.Error.Message)
	}

	sn := resp.Result.OrderId
	if sn == "" {
		logs.Errorf("failed to transOrder, for return empty order id, crp_trace: %s, "+
			"sub_ticket_id: %s, rid: %s", resp.TraceId, subTicket.ID, kt.Rid)
		return "", errors.New("failed to create crp ticket, for return empty order id")
	}

	return sn, nil
}

// constructTransReq construct cvm cbs plan trans request.
func (c *CrpTicketCreator) constructTransReq(kt *kit.Kit, subTicket *ptypes.SubTicketInfo,
	srcData []*cvmapi.AdjustSrcData, updateData []*cvmapi.AdjustUpdatedData) *cvmapi.TransOrderReq {

	transOrder := &cvmapi.TransOrderReq{
		ReqMeta: cvmapi.ReqMeta{
			Id:      cvmapi.CvmId,
			JsonRpc: cvmapi.CvmJsonRpc,
			Method:  cvmapi.CvmCbsPlanTransOrderMethod,
		},
		Params: &cvmapi.TransOrderParams{
			BaseInfo: cvmapi.TransOrderBaseInfo{
				DeptId:               int(subTicket.VirtualDeptID),
				DeptName:             subTicket.VirtualDeptName,
				PlanProductId:        subTicket.PlanProductID,
				PlanProductName:      subTicket.PlanProductName,
				ProductID:            subTicket.OpProductID,
				ProductName:          subTicket.OpProductName,
				BgName:               cvmapi.CvmCbsPlanQueryBgName,
				AfterDeptId:          cvmapi.CvmCbsPlanDeptId,
				AfterDeptName:        cvmapi.CvmCbsPlanQueryBgName,
				AfterPlanProductId:   cvmapi.TransferPlanProductID,
				AfterPlanProductName: cvmapi.TransferPlanProductName,
				AfterProductID:       cvmapi.TransferOpProductID,
				AfterProductName:     cvmapi.TransferOpProductName,
				AfterBgName:          cvmapi.CvmCbsPlanQueryBgName,
				// transfer_exempt 类型免审，transfer 类型不免审
				SkipTodo: subTicket.Type == enumor.RPTicketTypeTransferExempt,
			},
		},
	}

	sliceIDToUpdateData := make(map[string]*cvmapi.AdjustUpdatedData)
	for _, item := range updateData {
		sliceIDToUpdateData[item.SliceId] = item
	}

	for _, item := range srcData {
		updateTo, ok := sliceIDToUpdateData[item.SliceId]
		if !ok {
			// 理论上src一定有对应的update，没有时说明没有发生调减，直接跳过
			continue
		}

		logs.Infof("create transfer out crp ticket, ticketID: %s, srcData: %+v, newCore: %d, rid: %s",
			subTicket.ID, item.CvmCbsPlanQueryItem, updateTo.CoreAmount, kt.Rid)

		transOrderDetail := item.ToTransOrderDetail()
		// update表示原预测剩余的核心数，因此这里用 src - update得出需要转移的核心数、实例数、内存数
		cvmAmountNew := item.CvmAmount - updateTo.CvmAmount
		coreAmountNew := item.CoreAmount - updateTo.CoreAmount
		transOrderDetail.CvmAmount = cvmAmountNew
		transOrderDetail.CoreAmount = coreAmountNew
		transOrderDetail.RamAmount = int64(item.RamAmount - updateTo.RamAmount)
		// 对于标准型、小核心，CRP会以 coreAmountNew 为准，其他机型以 cvmAmountNew 为准，这里我们都传保持兼容
		transOrderDetail.CvmAmountNew = cvmAmountNew
		transOrderDetail.CoreAmountNew = coreAmountNew

		transOrder.Params.TransferDetailList = append(transOrder.Params.TransferDetailList, transOrderDetail)
	}

	return transOrder
}
