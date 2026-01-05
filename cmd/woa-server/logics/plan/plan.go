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

package plan

import (
	"context"
	"errors"
	"fmt"
	"runtime/debug"
	"time"

	"hcm/cmd/woa-server/logics/biz"
	demandtime "hcm/cmd/woa-server/logics/plan/demand-time"
	"hcm/cmd/woa-server/logics/plan/dispatcher"
	"hcm/cmd/woa-server/logics/plan/fetcher"
	"hcm/cmd/woa-server/storage/driver/mongodb"
	"hcm/cmd/woa-server/types/device"
	mtypes "hcm/cmd/woa-server/types/meta"
	ptypes "hcm/cmd/woa-server/types/plan"
	ttypes "hcm/cmd/woa-server/types/task"
	"hcm/pkg/api/core"
	rpproto "hcm/pkg/api/data-service/resource-plan"
	"hcm/pkg/cc"
	"hcm/pkg/client"
	"hcm/pkg/criteria/constant"
	"hcm/pkg/criteria/enumor"
	"hcm/pkg/dal/dao"
	"hcm/pkg/dal/dao/tools"
	rpdaotypes "hcm/pkg/dal/dao/types/resource-plan"
	rpts "hcm/pkg/dal/table/resource-plan/res-plan-ticket-status"
	wdt "hcm/pkg/dal/table/resource-plan/woa-device-type"
	"hcm/pkg/kit"
	"hcm/pkg/logs"
	"hcm/pkg/runtime/filter"
	"hcm/pkg/serviced"
	"hcm/pkg/thirdparty/api-gateway/cmsi"
	"hcm/pkg/thirdparty/api-gateway/finops"
	"hcm/pkg/thirdparty/api-gateway/itsm"
	"hcm/pkg/thirdparty/cvmapi"
	"hcm/pkg/tools/times"
)

// Logics provides management interface for resource plan.
type Logics interface {
	// ListResPlanDemandAndOverview list res plan demand and overview.
	ListResPlanDemandAndOverview(kt *kit.Kit, req *ptypes.ListResPlanDemandReq) (*ptypes.ListResPlanDemandResp, error)
	// GetResPlanDemandDetail get res plan demand detail.
	GetResPlanDemandDetail(kt *kit.Kit, demandID string, bkBizIDs []int64) (*ptypes.GetPlanDemandDetailResp, error)
	// QueryIEGDemands query IEG crp demands.
	QueryIEGDemands(kt *kit.Kit, req *QueryIEGDemandsReq) ([]*cvmapi.CvmCbsPlanQueryItem, error)
	// AreAllDemandBelongToBiz return whether all demands belong to biz.
	AreAllDemandBelongToBiz(kt *kit.Kit, demandIDs []string, bkBizID int64) (bool, error)
	// ListCrpDemandChangeLog list crp demand change log.
	ListCrpDemandChangeLog(kt *kit.Kit, req *ptypes.ListDemandChangeLogReq) (*ptypes.ListDemandChangeLogResp, error)

	// AdjustBizResPlanDemand adjust biz res plan demand.
	AdjustBizResPlanDemand(kt *kit.Kit, req *ptypes.AdjustRPDemandReq, bkBizID int64, bizOrgRel *mtypes.BizOrgRel) (
		ticketID string, retErr error)
	// CancelBizResPlanDemand cancel biz res plan demand.
	CancelBizResPlanDemand(kt *kit.Kit, req *ptypes.CancelRPDemandReq, bkBizID int64, bizOrgRel *mtypes.BizOrgRel) (
		string, error)

	// GetDemandAvailableTime get demand available time.
	GetDemandAvailableTime(kt *kit.Kit, expectTime time.Time) (*ptypes.DemandAvailTimeResp, error)
	// ExamineDemandClass examine whether all demands are the same demand class, and return the demand class.
	ExamineDemandClass(kt *kit.Kit, demandIDs []string) (enumor.DemandClass, error)
	// IsDeviceMatched return whether each device type in deviceTypeSlice can use deviceType's resource plan.
	IsDeviceMatched(kt *kit.Kit, deviceTypeSlice []string, deviceType string) ([]bool, error)
	// GetProdResPlanPool get op product resource plan pool.
	GetProdResPlanPool(kt *kit.Kit, prodID int64) (ResPlanPool, error)
	// GetProdResConsumePool get op product resource consume pool.
	GetProdResConsumePool(kt *kit.Kit, prodID, planProdID int64) (ResPlanPool, error)
	// GetProdResRemainPool get op product resource remain pool.
	// @param prodID is the op product id.
	// @param planProdID is the corresponding plan product id of the op product id.
	// @return prodRemainedPool is the op product in plan and out plan remained resource plan pool.
	// @return prodMaxAvailablePool is the op product in plan and out plan remained max available resource plan pool.
	// NOTE: maxAvailableInPlanPool = totalInPlan * 120% - consumeInPlan, because the special rules of the crp system.
	GetProdResRemainPool(kt *kit.Kit, prodID, planProdID int64) (ResPlanPool, ResPlanPool, error)
	// VerifyProdDemands verify whether the needs of op product can be satisfied.
	VerifyProdDemands(kt *kit.Kit, prodID, planProdID int64, needs []VerifyResPlanElem) ([]VerifyResPlanResElem, error)
	// GetProdResConsumePoolV2 get biz resource consume pool.
	GetProdResConsumePoolV2(kt *kit.Kit, bkBizIDs []int64, startDay, endDay time.Time) (ResPlanConsumePool, error)
	// VerifyResPlanDemandV2 verify resource plan demand for subOrders.
	VerifyResPlanDemandV2(kt *kit.Kit, bkBizID int64, requireType enumor.RequireType, subOrders []ttypes.Suborder) (
		[]ptypes.VerifyResPlanDemandElem, error)
	// VerifyProdDemandsV2 verify whether the needs of biz can be satisfied.
	VerifyProdDemandsV2(kt *kit.Kit, bkBizID int64, requireType enumor.RequireType, needs []VerifyResPlanElemV2) (
		[]VerifyResPlanResElem, error)
	// GetPlanTypeAvlDeviceTypesV2 get plan type avl device types.
	GetPlanTypeAvlDeviceTypesV2(kt *kit.Kit, planType enumor.PlanTypeCode, req *ptypes.GetCvmChargeTypeDeviceTypeReq,
		prodRemainMap map[ResPlanPoolKeyV2]map[string]int64) ([]ptypes.DeviceTypeAvailable, error)
	// GetProdResRemainPoolMatch get prod res remain pool match.
	GetProdResRemainPoolMatch(kt *kit.Kit, bkBizID int64, requireType enumor.RequireType) (
		ResPlanPoolMatch, ResPlanPoolMatch, error)
	// AddMatchedPlanDemandExpendLogs add matched plan demand expend logs.
	AddMatchedPlanDemandExpendLogs(kt *kit.Kit, bkBizID int64, subOrder *ttypes.ApplyOrder,
		verifyGroups []VerifyResPlanElemV2) error
	// GetAllDeviceTypeMap get all device type map.
	GetAllDeviceTypeMap(kt *kit.Kit) (map[string]wdt.WoaDeviceTypeTable, error)
	// SyncDeviceTypesFromCRP sync device types from crp.
	SyncDeviceTypesFromCRP(kt *kit.Kit, deviceTypes []string) error

	// CreateAuditFlow creates an audit flow for resource plan ticket.
	CreateAuditFlow(kt *kit.Kit, ticketID string) error
	// CreateResPlanTicket create resource plan ticket.
	CreateResPlanTicket(kt *kit.Kit, req *CreateResPlanTicketReq) (string, error)
	// GetResPlanTicketStatusInfo get res plan ticket status info.
	GetResPlanTicketStatusInfo(kt *kit.Kit, ticketID string) (*ptypes.GetRPTicketStatusInfo, error)
	// GetResPlanTicketAudit get res plan ticket audit.
	GetResPlanTicketAudit(kt *kit.Kit, ticketID string, bkBizID int64) (*ptypes.GetResPlanTicketAuditResp, error)
	// GetCrpCurrentApprove get crp current approve.
	GetCrpCurrentApprove(kt *kit.Kit, bkBizID int64, orderID string) ([]*ptypes.CrpAuditStep, error)
	// ListAllResPlanTicket list all res plan ticket with status.
	ListAllResPlanTicket(kt *kit.Kit, listFilter *filter.Expression) ([]rpdaotypes.RPTicketWithStatus, error)
	// ListResPlanTicketWithRes list res plan ticket with res.
	ListResPlanTicketWithRes(kt *kit.Kit, req *core.ListReq) (*ptypes.RPTicketWithStatusAndResListRst, error)
	// GetResPlanTicketStatusByBiz get res plan ticket status by biz.
	GetResPlanTicketStatusByBiz(kt *kit.Kit, ticketID string, bizID int64) (*ptypes.GetRPTicketStatusInfo, error)
	// ApproveTicketITSMByBiz approve ticket itsm by biz.
	ApproveTicketITSMByBiz(kt *kit.Kit, ticketID string, param *itsm.ApproveNodeOpt) error

	// QueryCrpDemandsQuota 查询crp预测额度
	QueryCrpDemandsQuota(kt *kit.Kit, obsProject []enumor.ObsProject, technicalClasses []string) (
		[]*cvmapi.CvmCbsPlanQueryItem, error)
	// ListRemainTransferQuota 查询剩余可转移额度
	ListRemainTransferQuota(kt *kit.Kit, req *ptypes.ListResPlanTransferQuotaSummaryReq) (
		*ptypes.ResPlanTransferQuotaSummaryResp, error)
	// GetPlanTransferQuotaConfigs 获取预测转移额度配置
	GetPlanTransferQuotaConfigs(kt *kit.Kit) (ptypes.TransferQuotaConfig, error)
	// UpdatePlanTransferQuotaConfigs 更新预测转移额度配置
	UpdatePlanTransferQuotaConfigs(kt *kit.Kit, req *ptypes.UpdatePlanTransferQuotaConfigsReq) error

	// ListResPlanSubTicket list res plan sub ticket.
	ListResPlanSubTicket(kt *kit.Kit, req *ptypes.ListResPlanSubTicketReq) (*ptypes.ListResPlanSubTicketResp, error)
	// GetResPlanSubTicketDetail get res plan sub ticket detail.
	GetResPlanSubTicketDetail(kt *kit.Kit, subTicketID string) (*ptypes.GetSubTicketDetailResp, string, error)
	// GetResPlanSubTicketAudit get res plan sub ticket audit.
	GetResPlanSubTicketAudit(kt *kit.Kit, bizID int64, subTicketID string) (*ptypes.GetSubTicketAuditResp, string,
		error)
	// ApproveResPlanSubTicketAdmin approve res plan ticket admin.
	ApproveResPlanSubTicketAdmin(kt *kit.Kit, subTicketID string, bizID int64,
		req *ptypes.AuditResPlanTicketAdminReq) error
	// RetryResPlanFailedSubTickets retry res plan failed sub tickets.
	RetryResPlanFailedSubTickets(kt *kit.Kit, ticketID string) error
	// TerminateResPlanFailedTicket terminate res plan failed ticket.
	TerminateResPlanFailedTicket(kt *kit.Kit, ticketID string) error

	// CreateDemandWeek create demand week.
	CreateDemandWeek(kt *kit.Kit, createReqs []rpproto.ResPlanWeekCreateReq) (*core.BatchCreateResult, error)

	// CalcPenaltyBase calc penalty base.
	CalcPenaltyBase(kt *kit.Kit, baseDay time.Time, bkBizIDs []int64) error
	// CalcPenaltyRatioAndPush calc penalty ratio and push.
	CalcPenaltyRatioAndPush(kt *kit.Kit, baseTime time.Time) error
	// PushExpireNotifications push expire notifications.
	PushExpireNotifications(kt *kit.Kit, bkBizIDs []int64, extraReceivers []string) error
	// PushResPlanConfirmNotice push res plan confirm notice.
	PushResPlanConfirmNotice(kt *kit.Kit, bkBizIDs []int64) ([]int64, []int64, error)
	// ConfirmResPlanDemands confirm res plan demands.
	ConfirmResPlanDemands(kt *kit.Kit, bkBizID int64, demandIDs []string) ([]string, []string, error)

	// RepairResPlanDemandFromTicket repair res plan demand from ticket.
	RepairResPlanDemandFromTicket(kt *kit.Kit, bkBizIDs []int64, ticketTimeRange times.DateRange) error
	// SyncDemandFromCRPOrder sync demand from crp order.
	SyncDemandFromCRPOrder(kt *kit.Kit, crpSN string, priorBizIDs []int64, opProdToBizID map[string]int64) error

	// ApplyDestroyOrderToResPlanDemand 将销毁单返还的预测预算保存到本地
	ApplyDestroyOrderToResPlanDemand(kt *kit.Kit, destroyOrderID string) error
	// AutoTransferBizResPlanDemandByID 根据业务ID和需求ID自动转移预测
	AutoTransferBizResPlanDemandByID(kt *kit.Kit, bkBizID int64, demandIDs []string) ([]string, error)
}

// Controller motivates the resource plan ticket status flow.
type Controller struct {
	resPlanCfg     cc.ResPlan
	dao            dao.Set
	sd             serviced.State
	client         *client.ClientSet
	bkHcmURL       string
	CmsiClient     cmsi.Client
	itsmCli        itsm.Client
	finOpsCli      finops.Client
	itsmFlow       cc.ItsmFlow
	crpCli         cvmapi.CVMClientInterface
	bizLogics      biz.Logics
	deviceTypesMap *device.DeviceTypesMap
	demandTime     demandtime.DemandTime
	ctx            context.Context

	resFetcher fetcher.Fetcher
	dispatcher *dispatcher.Dispatcher
}

// New creates a resource plan ticket controller instance.
func New(sd serviced.State, client *client.ClientSet, dao dao.Set, cmsiCli cmsi.Client, itsmCli itsm.Client,
	finOpsCli finops.Client, crpCli cvmapi.CVMClientInterface, bizLogic biz.Logics) (Logics, error) {

	var itsmFlowCfg cc.ItsmFlow
	for _, itsmFlow := range cc.WoaServer().ItsmFlows {
		if itsmFlow.ServiceName == enumor.TicketSvcNameResPlan {
			itsmFlowCfg = itsmFlow
			break
		}
	}

	demandTimeCli := demandtime.NewDemandTimeFromTable(client)
	deviceTypesMap := device.NewDeviceTypesMap(dao)
	fetch := fetcher.New(dao, demandTimeCli, client, crpCli, itsmCli, bizLogic, deviceTypesMap)

	ctx := context.Background()
	// new dispatcher
	dispatch, err := dispatcher.New(ctx, sd, client, dao, itsmCli, crpCli, bizLogic, deviceTypesMap, fetch)
	if err != nil {
		logs.Errorf("new res plan dispatcher failed: %v", err)
		return nil, err
	}

	ctrl := &Controller{
		resPlanCfg:     cc.WoaServer().ResPlan,
		dao:            dao,
		sd:             sd,
		client:         client,
		bkHcmURL:       cc.WoaServer().BkHcmURL,
		CmsiClient:     cmsiCli,
		itsmCli:        itsmCli,
		finOpsCli:      finOpsCli,
		itsmFlow:       itsmFlowCfg,
		crpCli:         crpCli,
		bizLogics:      bizLogic,
		deviceTypesMap: deviceTypesMap,
		demandTime:     demandTimeCli,
		ctx:            ctx,
		resFetcher:     fetch,
		dispatcher:     dispatch,
	}

	go ctrl.Run()

	return ctrl, nil
}

func (c *Controller) recoverLog(keywords constant.WarnSign) {
	if r := recover(); r != nil {
		logs.Errorf("%s: panic: %v\n%s", keywords, r, debug.Stack())
	}
}

// Run starts dispatcher
func (c *Controller) Run() {
	// controller启动后需等待一段时间，mongo等服务初始化完成后才能开始定时任务 TODO 用统一的任务调度模块来执行定时任务，确保在初始化之后
	for {
		if mongodb.Client() == nil {
			logs.Warnf("mongodb client is not ready, wait seconds to retry")
			time.Sleep(constant.IntervalWaitTaskStart)
			continue
		}
		break
	}

	loc, err := time.LoadLocation(cc.WoaServer().LocalTimezone)
	if err != nil {
		logs.Warnf("%s: load location: %s failed: %v", constant.ResPlanExpireNotificationPushFailed,
			cc.WoaServer().LocalTimezone, err)
		loc = time.UTC
	}

	// 每周一生成12周后的核算基准数据
	go func() {
		defer c.recoverLog(constant.DemandPenaltyBaseGenerateFailed)

		c.generatePenaltyBase(c.ctx)
	}()

	// 每月最后7天，每天下午18:00计算当月罚金分摊比例并推送到CRP
	go func() {
		if !c.resPlanCfg.ReportPenaltyRatio {
			return
		}

		defer c.recoverLog(constant.DemandPenaltyRatioReportFailed)

		c.calcAndReportPenaltyRatioToCRP(c.ctx, loc)
	}()

	// 预测到期提醒通知
	go func() {
		if !c.resPlanCfg.ExpireNotification.Enable {
			return
		}

		defer c.recoverLog(constant.ResPlanExpireNotificationPushFailed)

		c.pushExpireNotificationsRegular(c.ctx, loc)
	}()

	// 临期预测转移
	go func() {
		if !c.resPlanCfg.NearExpiredTransfer.Enable {
			return
		}
		defer c.recoverLog(constant.ResPlanNearExpiredDemandTransferFailed)
		c.autoTransferNearExpireDemand(c.ctx, loc)
	}()

	// 预测确认通知
	go func() {
		if !c.resPlanCfg.ConfirmNotice.Enable {
			return
		}
		defer c.recoverLog(constant.ResPlanConfirmNotificationFailed)
		c.runConfirmNotice(c.ctx, loc)
	}()

	select {
	case <-c.ctx.Done():
		logs.Infof("resource plan ticket controller exits")
	}
}

// CreateAuditFlow creates an audit flow for resource plan ticket.
// TODO ITSM单的创建也应在 dispatcher 中执行
func (c *Controller) CreateAuditFlow(kt *kit.Kit, ticketID string) error {
	ticket, err := c.resFetcher.GetTicketInfo(kt, ticketID)
	if err != nil {
		logs.Errorf("failed to get ticket info, err: %v", err)
		return err
	}

	sn, err := c.createItsmTicket(kt, ticket)
	if err != nil {
		logs.Errorf("failed to create itsm ticket, err: %v, rid: %s", err, kt.Rid)
		return err
	}

	itsmStatus, err := c.itsmCli.GetTicketStatus(kt, sn)
	if err != nil {
		logs.Errorf("failed to get itsm ticket status, err: %v, id: %s, rid: %s", err, ticket.ID, kt.Rid)
		return err
	}

	update := &rpts.ResPlanTicketStatusTable{
		TicketID: ticketID,
		Status:   enumor.RPTicketStatusAuditing,
		ItsmSN:   sn,
		ItsmURL:  itsmStatus.Data.TicketUrl,
	}

	if err = c.updateTicketStatus(kt, update); err != nil {
		logs.Errorf("failed to update resource plan ticket status, err: %v, rid: %s", err, kt.Rid)
		return err
	}

	return nil
}

// updateTicketStatus update ticket status.
func (c *Controller) updateTicketStatus(kt *kit.Kit, ticket *rpts.ResPlanTicketStatusTable) error {
	expr := tools.EqualExpression("ticket_id", ticket.TicketID)
	if err := c.dao.ResPlanTicketStatus().Update(kt, expr, ticket); err != nil {
		logs.Errorf("failed to update resource plan ticket status, err: %v, ticket_id: %s, rid: %s", err,
			ticket.TicketID, kt.Rid)
		return err
	}
	return nil
}

func (c *Controller) createItsmTicket(kt *kit.Kit, ticket *ptypes.TicketInfo) (string, error) {
	if ticket == nil {
		return "", errors.New("ticket is nil")
	}

	// TODO：待修改
	contentTemplate := `业务：%s(%d)
预测类型：%s
CPU变更核数：%d
内存变更量(GB)：%d
云盘变更量(GB)：%d
`
	content := fmt.Sprintf(contentTemplate, ticket.BkBizName, ticket.BkBizID, ticket.DemandClass,
		ticket.UpdatedCpuCore-ticket.OriginalCpuCore, ticket.UpdatedMemory-ticket.OriginalMemory,
		ticket.UpdatedDiskSize-ticket.OriginalDiskSize)
	createTicketReq := &itsm.CreateTicketParams{
		ServiceID:      c.itsmFlow.ServiceID,
		Creator:        ticket.Applicant,
		Title:          fmt.Sprintf("%s(业务ID: %d)资源预测申请", ticket.BkBizName, ticket.BkBizID),
		ContentDisplay: content,
		ExtraFields: map[string]interface{}{"res_plan_url": fmt.Sprintf(c.itsmFlow.RedirectUrlTemplate,
			ticket.ID, ticket.BkBizID)},
	}

	sn, err := c.itsmCli.CreateTicket(kt, createTicketReq)
	if err != nil {
		logs.Errorf("create itsm ticket failed, err: %v, rid: %s", err, kt.Rid)
		return "", err
	}

	return sn, nil
}

// ApproveTicketITSMByBiz 审批 预测单itsm节点
func (c *Controller) ApproveTicketITSMByBiz(kt *kit.Kit, ticketID string, param *itsm.ApproveNodeOpt) error {

	if err := c.itsmCli.ApproveNode(kt, param); err != nil {
		logs.Errorf("failed to approve itsm node of plan ticket %s, err: %v, rid: %s", ticketID, err, kt.Rid)
		return err
	}
	return nil
}
