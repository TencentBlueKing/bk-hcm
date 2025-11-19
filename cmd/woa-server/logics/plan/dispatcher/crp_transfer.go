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
	"hcm/pkg/kit"
	"hcm/pkg/logs"
	"hcm/pkg/thirdparty/cvmapi"
)

// createTransferOutCrpTicket create transfer out crp ticket.
func (c *CrpTicketCreator) createTransferOutCrpTicket(kt *kit.Kit, subTicket *ptypes.SubTicketInfo,
	srcData []*cvmapi.AdjustSrcData) (string, error) {

	transReq := c.constructTransReq(subTicket, srcData)
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
func (c *CrpTicketCreator) constructTransReq(subTicket *ptypes.SubTicketInfo,
	srcData []*cvmapi.AdjustSrcData) *cvmapi.TransOrderReq {

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
				SkipTodo:             true,
			},
		},
	}

	for _, item := range srcData {
		transOrder.Params.TransferDetailList = append(transOrder.Params.TransferDetailList, cvmapi.TransOrderDetail{
			SliceId:             item.SliceId,
			CityId:              item.CityId,
			CityName:            item.CityName,
			ZoneId:              item.ZoneId,
			ZoneName:            item.ZoneName,
			InstanceType:        item.InstanceType,
			InstanceModel:       item.InstanceModel,
			CvmAmount:           int(item.CvmAmount),
			RamAmount:           int(item.RamAmount),
			CoreAmount:          int(item.CoreAmount),
			InstanceIO:          item.InstanceIO,
			DiskType:            item.DiskType,
			DiskTypeName:        item.DiskTypeName,
			AllDiskAmount:       int(item.AllDiskAmount),
			Desc:                "", // 需要从其他字段获取或留空
			ProjectName:         item.ProjectName,
			RequirementWeekType: item.RequirementWeekType,
			Year:                item.Year,
			Month:               item.Month,
			UseTime:             item.UseTime,
			BgId:                item.BgId,
			BgName:              item.BgName,
			DeptId:              item.DeptId,
			DeptName:            item.DeptName,
			PlanProductId:       item.PlanProductId,
			PlanProductName:     item.PlanProductName,
			ProductName:         item.ProductName,
			ReviewStatus:        item.ReviewStatus,
			CoreType:            item.CoreType,
			CoreTypeName:        item.CoreTypeName,
		})
	}

	return transOrder
}
