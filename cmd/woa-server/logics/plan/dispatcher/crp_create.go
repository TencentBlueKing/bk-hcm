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
	"strings"

	ptypes "hcm/cmd/woa-server/types/plan"
	rpproto "hcm/pkg/api/data-service/resource-plan"
	"hcm/pkg/cc"
	"hcm/pkg/criteria/constant"
	"hcm/pkg/criteria/enumor"
	"hcm/pkg/kit"
	"hcm/pkg/logs"
	"hcm/pkg/thirdparty/cvmapi"
	"hcm/pkg/tools/retry"
)

const (
	// CreateCrpTicketRespRateLimitCode 创建crp单据响应限流错误码
	CreateCrpTicketRespRateLimitCode = -254

	// CreateCrpTicketDefaultRetryTimes 创建crp单据默认重试次数
	CreateCrpTicketDefaultRetryTimes = 3
	// CreateCrpTicketDefaultRetryDelayMinMS 创建crp单据默认重试最小延迟时间
	CreateCrpTicketDefaultRetryDelayMinMS = 5000 // 5s
	// CreateCrpTicketDefaultRetryDelayMaxMS 创建crp单据默认重试最大延迟时间
	CreateCrpTicketDefaultRetryDelayMaxMS = 10000 // 10s
)

// createCrpTicket create crp ticket.
func (d *Dispatcher) createCrpTicket(kt *kit.Kit, subTicket *ptypes.SubTicketInfo) error {
	if subTicket == nil {
		logs.Errorf("failed to create crp ticket, sub ticket is nil, rid: %s", kt.Rid)
		return errors.New("sub ticket is nil")
	}
	if len(subTicket.Demands) == 0 {
		logs.Errorf("failed to create crp ticket, sub ticket has no demand, sub_ticket_id: %s, rid: %s",
			subTicket.ID, kt.Rid)
		return errors.New("sub ticket has no demand")
	}

	// call crp api to create crp ticket.
	crpCreator := NewCrpTicketCreator(d.resFetcher, d.crpCli)
	sn, err := crpCreator.CreateCRPTicket(kt, subTicket)
	if err != nil {
		// 因CRP单据修改冲突导致的提单失败，不返回报错，记录日志后返回队列继续等待
		if strings.Contains(err.Error(), constant.CRPResPlanDemandIsInProcessing) {
			logs.Warnf("failed to create crp ticket, as crp res plan demand is in processing, err: %v, "+
				"sub_ticket_id: %s, rid: %s", err, subTicket.ID, kt.Rid)
			return nil
		}

		// 这里主要返回的error是crp ticket创建失败，且ticket状态更新失败的日志在函数内已打印，这里可以忽略该错误
		_ = d.updateSubTicketStatusFailed(kt, subTicket, err.Error())
		logs.Errorf("failed to create crp ticket with different ticket type, err: %v, sub_ticket_id: %s, rid: %s",
			err, subTicket.ID, kt.Rid)
		return err
	}

	// save crp sn and crp url to resource plan sub ticket status table.
	update := &rpproto.ResPlanSubTicketUpdateReq{
		ID:     subTicket.ID,
		Stage:  enumor.RPSubTicketStageCRPAudit,
		CrpSN:  sn,
		CrpURL: cvmapi.CvmPlanLinkPrefix + sn,
	}

	if err = d.updateSubTicket(kt, subTicket, update); err != nil {
		logs.Errorf("failed to update resource plan sub ticket, err: %v, sub_ticket_id: %s, rid: %s", err,
			subTicket.ID, kt.Rid)
		return err
	}
	return nil
}

// createAddCrpTicket create add crp ticket.
func (c *CrpTicketCreator) createAddCrpTicket(kt *kit.Kit, subTicket *ptypes.SubTicketInfo) (string, error) {
	addReq, err := c.constructAddReq(kt, subTicket)
	if err != nil {
		logs.Errorf("failed to construct add cvm & cbs plan order request, err: %v, sub_ticket_id: %s, rid: %s",
			err, subTicket.ID, kt.Rid)
		return "", err
	}

	resp := new(cvmapi.AddCvmCbsPlanResp)
	rangeMS := [2]uint{CreateCrpTicketDefaultRetryDelayMinMS, CreateCrpTicketDefaultRetryDelayMaxMS}
	policy := retry.NewRetryPolicy(0, rangeMS)
	for {
		resp, err = c.crpCli.AddCvmCbsPlan(kt.Ctx, kt.Header(), addReq)
		if err != nil {
			logs.Errorf("failed to add cvm & cbs plan order, err: %v, sub_ticket_id: %s, rid: %s", err, subTicket.ID,
				kt.Rid)
			return "", err
		}
		// 仅在碰到限频错误时进行重试
		if resp.Error.Code == CreateCrpTicketRespRateLimitCode {
			if policy.RetryCount()+1 < CreateCrpTicketDefaultRetryTimes {
				// 	非最后一次重试，继续sleep
				logs.Warnf(
					"call crp rate limit, will sleep for retry, retry count: %d, err: %v, crp_trace: %s, rid: %s",
					policy.RetryCount(), resp.Error, resp.TraceId, kt.Rid)
				policy.Sleep()
				continue
			}
		}
		// 其他情况都跳过
		break
	}

	if resp.Error.Code != 0 {
		logs.Errorf("failed to create add plan order, code: %d, msg: %s, crp_trace: %s, "+
			"sub_ticket_id: %s, rid: %s", resp.Error.Code, resp.Error.Message, resp.TraceId, subTicket.ID, kt.Rid)
		if strings.Contains(resp.Error.Message, constant.CRPResPlanDemandIsOverLimit) {
			return "", fmt.Errorf(constant.CRPResPlanDemandIsOverLimitMessage,
				strings.Join(cc.WoaServer().ResPlan.CRPOverLimitContact, ","))
		}
		return "", fmt.Errorf("failed to create add plan order, code: %d, msg: %s", resp.Error.Code,
			resp.Error.Message)
	}

	sn := resp.Result.OrderId
	if sn == "" {
		logs.Errorf("failed to create add plan order, for return empty order id, crp_trace: %s, "+
			"sub_ticket_id: %s, rid: %s", resp.TraceId, subTicket.ID, kt.Rid)
		return "", errors.New("failed to create add plan order, for return empty order id")
	}

	return sn, nil
}

// constructAddReq construct cvm cbs plan add request.
func (c *CrpTicketCreator) constructAddReq(kt *kit.Kit, subTicket *ptypes.SubTicketInfo) (
	*cvmapi.AddCvmCbsPlanReq, error) {

	addReq := &cvmapi.AddCvmCbsPlanReq{
		ReqMeta: cvmapi.ReqMeta{
			Id:      cvmapi.CvmId,
			JsonRpc: cvmapi.CvmJsonRpc,
			Method:  cvmapi.CvmCbsPlanAddMethod,
		},
		Params: &cvmapi.AddCvmCbsPlanParam{
			Operator: subTicket.Applicant,
			DeptName: subTicket.VirtualDeptName,
			Desc:     "",
			Items:    make([]*cvmapi.AddPlanItem, 0),
		},
	}

	switch subTicket.DemandClass {
	case enumor.DemandClassCVM:
		addReq.Params.Desc = cvmapi.CvmCbsPlanDefaultCvmDesc
	case enumor.DemandClassCA:
		addReq.Params.Desc = cvmapi.CvmCbsPlanDefaultCADesc
	default:
		logs.Warnf("failed to construct add desc, unsupported demand class: %s, rid: %s", subTicket.DemandClass,
			kt.Rid)
	}

	for _, demand := range subTicket.Demands {
		if demand.Updated == nil {
			logs.Errorf("failed to create add crp ticket, demand updated is nil, rid: %s", kt.Rid)
			return nil, errors.New("demand updated is nil")
		}

		planItem := &cvmapi.AddPlanItem{
			UseTime:         demand.Updated.ExpectTime,
			ProjectName:     string(demand.Updated.ObsProject),
			PlanProductName: subTicket.PlanProductName,
			ProductName:     subTicket.OpProductName,
			CityName:        demand.Updated.RegionName,
			ZoneName:        demand.Updated.ZoneName,
			Desc:            demand.Updated.Remark,
			InstanceIO:      int(demand.Updated.Cbs.DiskIo),
			DiskTypeName:    demand.Updated.Cbs.DiskType.Name(),
			DiskAmount:      int(demand.Updated.Cbs.DiskSize),
		}

		// 仅在有 Cvm 需求的情况下填充 Cvm 相关字段
		if !demand.Updated.Cvm.IsEmpty() {
			planItem.CoreTypeName = demand.Updated.Cvm.CoreType
			planItem.InstanceModel = demand.Updated.Cvm.DeviceType
			planItem.CvmAmount = demand.Updated.Cvm.Os.InexactFloat64()
			planItem.CoreAmount = int(demand.Updated.Cvm.CpuCore)
		}

		if demand.Updated.ObsProject == enumor.ObsProjectShortLease {
			planItem.IsAutoReturnPlan = true
			planItem.ReturnPlanTime = demand.Updated.ReturnPlanTime
		}

		addReq.Params.Items = append(addReq.Params.Items, planItem)
	}

	return addReq, nil
}
