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

package fetcher

import (
	"errors"
	"fmt"
	"strings"

	ptypes "hcm/cmd/woa-server/types/plan"
	"hcm/pkg/criteria/enumor"
	"hcm/pkg/kit"
	"hcm/pkg/logs"
	"hcm/pkg/thirdparty/cvmapi"
)

// GetCrpCurrentApprove 查询当前审批节点
func (f *ResPlanFetcher) GetCrpCurrentApprove(kt *kit.Kit, bkBizID int64, orderID string) ([]*ptypes.CrpAuditStep,
	error) {

	req := &cvmapi.QueryPlanOrderReq{
		ReqMeta: cvmapi.ReqMeta{
			Id:      cvmapi.CvmId,
			JsonRpc: cvmapi.CvmJsonRpc,
			Method:  cvmapi.CvmCbsPlanOrderQueryMethod,
		},
		Params: &cvmapi.QueryPlanOrderParam{
			OrderIds: []string{orderID},
		},
	}
	resp, err := f.crpCli.QueryPlanOrder(kt.Ctx, kt.Header(), req)
	if err != nil {
		logs.Errorf("failed to query crp plan order, err: %v, rid: %s", err, kt.Rid)
		return nil, err
	}

	if resp.Error.Code != 0 {
		logs.Errorf("failed to query crp plan order, code: %d, msg: %s, order id: %s, rid: %s", resp.Error.Code,
			resp.Error.Message, orderID, kt.Rid)
		return nil, fmt.Errorf("failed to query crp plan order, code: %d, msg: %s", resp.Error.Code, resp.Error.Message)
	}

	if resp.Result == nil {
		logs.Errorf("failed to query crp plan order, for result is empty, order id: %s, rid: %s", orderID, kt.Rid)
		return nil, errors.New("failed to query crp plan order, for result is empty")
	}

	orderItem, ok := resp.Result[orderID]
	if !ok {
		logs.Errorf("query crp plan order return no result by order id: %s, rid: %s", orderID, kt.Rid)
		return nil, fmt.Errorf("query crp plan order return no result by order id: %s", orderID)
	}

	// 如果processors为空，说明审批已经结束
	processors := orderItem.Data.BaseInfo.CurrentProcessor
	if processors == "" {
		return []*ptypes.CrpAuditStep{}, nil
	}

	// 校验审批人是否有该业务的访问权限
	processorUsers := strings.Split(processors, ";")
	processorAuth, err := f.bizLogics.BatchCheckUserBizAccessAuth(kt, bkBizID, processorUsers)
	if err != nil {
		return nil, err
	}

	var crpStatus enumor.CrpOrderStatus
	switch orderItem.Data.BaseInfo.Status {
	case cvmapi.PlanOrderStatusDeptAdmin:
		crpStatus = enumor.CrpOrderStatusDeptApprove
	case cvmapi.PlanOrderStatusPlanManager:
		crpStatus = enumor.CrpOrderStatusPlanApprove
	case cvmapi.PlanOrderStatusResManager:
		crpStatus = enumor.CrpOrderStatusResourceApprove
	default:
		crpStatus = enumor.CrpOrderStatus(orderItem.Data.BaseInfo.Status)
	}

	currentStep := &ptypes.CrpAuditStep{
		StateID:        "", // CRP接口暂时没有节点的ID，后续实现审批操作功能时，必须补全这个ID
		Name:           orderItem.Data.BaseInfo.StatusDesc,
		Status:         crpStatus,
		Processors:     processorUsers,
		ProcessorsAuth: processorAuth,
	}

	return []*ptypes.CrpAuditStep{currentStep}, nil
}

// GetCrpApproveLogs 查询Crp审批记录
func (f *ResPlanFetcher) GetCrpApproveLogs(kt *kit.Kit, orderID string) ([]*ptypes.CrpAuditLog, error) {
	req := &cvmapi.GetApproveLogReq{
		ReqMeta: cvmapi.ReqMeta{
			Id:      cvmapi.CvmId,
			JsonRpc: cvmapi.CvmJsonRpc,
			Method:  cvmapi.GetApproveLogMethod,
		},
		Params: &cvmapi.GetApproveLogParams{
			OrderId: []string{orderID},
		},
	}

	resp, err := f.crpCli.GetApproveLog(kt.Ctx, kt.Header(), req)
	if err != nil {
		logs.Errorf("failed to get crp approve log, err: %v, rid: %s", err, kt.Rid)
		return nil, err
	}

	if resp.Error.Code != 0 {
		logs.Errorf("failed to get crp approve log, code: %d, msg: %s, order id: %s, rid: %s", resp.Error.Code,
			resp.Error.Message, req.Params.OrderId, kt.Rid)
		return nil, fmt.Errorf("failed to get crp approve log, code: %d, msg: %s", resp.Error.Code, resp.Error.Message)
	}

	if resp.Result == nil {
		logs.Errorf("failed to get crp approve log, for result is empty, order id: %s, rid: %s", req.Params.OrderId,
			kt.Rid)
		return nil, errors.New("failed to get crp approve log, for result is empty")
	}

	orderLogs, ok := resp.Result[orderID]
	if !ok {
		return []*ptypes.CrpAuditLog{}, nil
	}

	// crp返回的审批记录是倒序的，需要反转
	auditLogs := make([]*ptypes.CrpAuditLog, len(orderLogs))
	for i := len(orderLogs) - 1; i >= 0; i-- {
		auditLogs[len(orderLogs)-1-i] = &ptypes.CrpAuditLog{
			Operator:  orderLogs[i].Operator,
			OperateAt: orderLogs[i].OperateTime,
			Message:   orderLogs[i].OperateResult,
			Name:      orderLogs[i].Activity,
		}
	}

	return auditLogs, nil
}

// GetOrderList 根据销毁单据查询预测返还信息
func (f *ResPlanFetcher) GetOrderList(kt *kit.Kit, orderID string) ([]*cvmapi.QueryOrderInfo, error) {
	req := &cvmapi.QueryOrderListReq{
		ReqMeta: cvmapi.ReqMeta{
			Id:      cvmapi.CvmId,
			JsonRpc: cvmapi.CvmJsonRpc,
			Method:  cvmapi.CvmQueryOrderList,
		},
		Params: &cvmapi.QueryOrderListParam{
			DestroyReturnPlanOrderId: []string{orderID},
			UserName:                 cvmapi.CvmApiKeyVal,
		},
	}

	resp, err := f.crpCli.QueryOrderList(kt.Ctx, kt.Header(), req)
	if err != nil {
		logs.Errorf("failed to get crp destroy order, err: %v, rid: %s", err, kt.Rid)
		return nil, err
	}

	if resp.Error.Code != 0 {
		logs.Errorf("failed to get crp destroy order, code: %d, crpTraceID: %s, msg: %s, order id: %v, rid: %s",
			resp.Error.Code, resp.TraceId, resp.Error.Message, req.Params.DestroyReturnPlanOrderId, kt.Rid)
		return nil, fmt.Errorf("failed to get crp destroy order, code: %d, msg: %s", resp.Error.Code, resp.Error.Message)
	}

	if resp.Result == nil || resp.Result.Data == nil {
		logs.Errorf("failed to get destroy order, for result is empty, crpTraceID: %s, order id: %v, rid: %s",
			resp.TraceId, req.Params.DestroyReturnPlanOrderId, kt.Rid)
		return nil, errors.New("failed to get destroy order, for result is empty")
	}

	return resp.Result.Data.Data, nil
}
