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

// Package scheduler 调度器
package scheduler

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"reflect"
	"regexp"
	"sort"
	"strconv"
	"strings"
	"time"

	"hcm/cmd/woa-server/dal/task/dao"
	"hcm/cmd/woa-server/dal/task/table"
	"hcm/cmd/woa-server/logics/biz"
	"hcm/cmd/woa-server/logics/config"
	"hcm/cmd/woa-server/logics/dissolve"
	greenchannel "hcm/cmd/woa-server/logics/green-channel"
	"hcm/cmd/woa-server/logics/plan"
	rollingserver "hcm/cmd/woa-server/logics/rolling-server"
	shortrental "hcm/cmd/woa-server/logics/short-rental"
	"hcm/cmd/woa-server/logics/task/informer"
	"hcm/cmd/woa-server/logics/task/scheduler/dispatcher"
	"hcm/cmd/woa-server/logics/task/scheduler/generator"
	"hcm/cmd/woa-server/logics/task/scheduler/matcher"
	"hcm/cmd/woa-server/logics/task/scheduler/recommender"
	"hcm/cmd/woa-server/logics/task/scheduler/record"
	model "hcm/cmd/woa-server/model/task"
	cfgtype "hcm/cmd/woa-server/types/config"
	configtypes "hcm/cmd/woa-server/types/config"
	rstypes "hcm/cmd/woa-server/types/rolling-server"
	types "hcm/cmd/woa-server/types/task"
	"hcm/pkg"
	"hcm/pkg/adaptor/types/cvm"
	"hcm/pkg/api/core"
	"hcm/pkg/cc"
	"hcm/pkg/client"
	"hcm/pkg/criteria/constant"
	"hcm/pkg/criteria/enumor"
	"hcm/pkg/criteria/errf"
	"hcm/pkg/criteria/mapstr"
	"hcm/pkg/dal"
	"hcm/pkg/dal/dao/tools"
	"hcm/pkg/kit"
	"hcm/pkg/logs"
	"hcm/pkg/thirdparty"
	"hcm/pkg/thirdparty/api-gateway/bkbotapproval"
	"hcm/pkg/thirdparty/api-gateway/cmdb"
	"hcm/pkg/thirdparty/api-gateway/cmsi"
	"hcm/pkg/thirdparty/api-gateway/itsm"
	"hcm/pkg/thirdparty/cvmapi"
	cvt "hcm/pkg/tools/converter"
	"hcm/pkg/tools/language"
	"hcm/pkg/tools/maps"
	"hcm/pkg/tools/metadata"
	"hcm/pkg/tools/querybuilder"
	"hcm/pkg/tools/slice"
	"hcm/pkg/tools/util"

	"go.mongodb.org/mongo-driver/mongo"
)

// Interface scheduler interface
type Interface interface {
	// UpdateApplyTicket creates or updates resource apply ticket
	UpdateApplyTicket(kt *kit.Kit, param *types.ApplyReq) (*types.CreateApplyOrderResult, error)
	// GetApplyTicket gets resource apply ticket
	GetApplyTicket(kit *kit.Kit, param *types.GetApplyTicketReq) (*types.GetApplyTicketRst, error)
	// GetApplyAuditItsm gets resource apply ticket itsm audit info
	GetApplyAuditItsm(kit *kit.Kit, param *types.GetApplyAuditItsmReq) (*types.GetApplyAuditItsmRst, error)
	// GetApplyAuditCrp gets resource apply ticket crp audit info
	GetApplyAuditCrp(kit *kit.Kit, param *types.GetApplyAuditCrpReq, resType types.ResourceType) (
		*types.GetApplyAuditCrpRst, error)
	// AuditTicket audit resource apply ticket
	AuditTicket(kit *kit.Kit, param *types.ApplyAuditReq) error
	// AutoAuditTicket system automatic audit resource apply ticket callback
	AutoAuditTicket(kit *kit.Kit, param *types.ApplyAutoAuditReq) (*types.ApplyAutoAuditRst, error)
	// ApproveTicket approve apply ticket callback
	ApproveTicket(kit *kit.Kit, param *types.ApproveApplyReq) error

	// CreateApplyOrder creates resource apply order
	CreateApplyOrder(kit *kit.Kit, param *types.ApplyReq) (*types.CreateApplyOrderResult, error)
	// GetApplyOrder gets resource apply order info
	GetApplyOrder(kit *kit.Kit, param *types.GetApplyParam) (*types.GetApplyOrderRst, error)
	// GetApplyDetail gets resource apply order detail info
	GetApplyDetail(kit *kit.Kit, param *types.GetApplyDetailReq) (*types.GetApplyDetailRst, error)
	// GetApplyGenerate gets resource apply order generate records
	GetApplyGenerate(kit *kit.Kit, param *types.GetApplyGenerateReq) (*types.GetApplyGenerateRst, error)
	// GetApplyInit gets resource apply order init records
	GetApplyInit(kit *kit.Kit, param *types.GetApplyInitReq) (*types.GetApplyInitRst, error)
	// GetApplyDiskCheck gets resource apply order disk check records
	GetApplyDiskCheck(kit *kit.Kit, param *types.GetApplyInitReq) (*types.GetApplyDiskCheckRst, error)
	// GetApplyDeliver gets resource apply order deliver records
	GetApplyDeliver(kit *kit.Kit, param *types.GetApplyDeliverReq) (*types.GetApplyDeliverRst, error)
	// GetApplyDevice get resource apply delivered devices
	GetApplyDevice(kit *kit.Kit, param *types.GetApplyDeviceReq) (*types.GetApplyDeviceRst, error)
	// ExportDeliverDevice export resource apply delivered devices
	ExportDeliverDevice(kit *kit.Kit, param *types.ExportDeliverDeviceReq) (*types.GetApplyDeviceRst, error)
	// GetMatchDevice get resource apply match devices
	GetMatchDevice(kit *kit.Kit, param *types.GetMatchDeviceReq) (*types.GetMatchDeviceRst, error)
	// MatchDevice execute resource apply match devices
	MatchDevice(kit *kit.Kit, param *types.MatchDeviceReq) error
	// MatchPoolDevice execute resource apply match devices from resource pool
	MatchPoolDevice(kit *kit.Kit, param *types.MatchPoolDeviceReq) error
	// GetAffinityMatchDetail get affinity match detail
	GetAffinityMatchDetail(kit *kit.Kit, param *types.AffinityMatchReq) (*types.AffinityMatchResp, error)
	// PauseApplyOrder pauses resource apply order
	PauseApplyOrder(kit *kit.Kit, param mapstr.MapStr) error
	// ResumeApplyOrder resumes resource apply order
	ResumeApplyOrder(kit *kit.Kit, param mapstr.MapStr) error
	// StartApplyOrder starts resource apply order
	StartApplyOrder(kit *kit.Kit, param *types.StartApplyOrderReq) error
	// TerminateApplyOrder terminates resource apply order
	TerminateApplyOrder(kit *kit.Kit, param *types.TerminateApplyOrderReq) error
	// ModifyApplyOrder modify resource apply order
	ModifyApplyOrder(kit *kit.Kit, param *types.ModifyApplyReq) error
	// ConfirmApplyModify confirm resource apply modify
	ConfirmApplyModify(kit *kit.Kit, param *types.ConfirmApplyModifyReq) (*types.ConfirmApplyModifyResp, error)
	// RecommendApplyOrder get resource apply order modification recommendation
	RecommendApplyOrder(kit *kit.Kit, param *types.RecommendApplyReq) (*types.RecommendApplyRst, error)
	// GetApplyModify gets resource apply order modify records
	GetApplyModify(kit *kit.Kit, param *types.GetApplyModifyReq) (*types.GetApplyModifyRst, error)
	// DeliverDevice deliver one device to business
	DeliverDevice(info *types.DeviceInfo, order *types.ApplyOrder) error
	// SetDeviceDelivered set device info delivered
	SetDeviceDelivered(info *types.DeviceInfo) error
	// GetGenerateRecords check and update cvm device
	GetGenerateRecords(kt *kit.Kit, orderId string) ([]*types.GenerateRecord, error)
	// AddCvmDevices check and update cvm device
	AddCvmDevices(kit *kit.Kit, taskId string, generateId uint64, order *types.ApplyOrder) error
	// UpdateOrderStatus check generate record by order id
	UpdateOrderStatus(resType types.ResourceType, suborderID string) error
	// UpdateHostOperator update operator of host
	UpdateHostOperator(info *types.DeviceInfo, hostId int64, operator string) error
	// ProcessInitStep process init step
	ProcessInitStep(device *types.DeviceInfo) error
	// CheckSopsUpdate check if the sops task is completed and update the initialization status
	CheckSopsUpdate(bkBizID int64, info *types.DeviceInfo, jobUrl string, jobIDStr string) error
	// RunDiskCheck run disk check
	RunDiskCheck(order *types.ApplyOrder, devices []*types.DeviceInfo) ([]*types.DeviceInfo, error)
	// DeliverDevices deliver devices to business
	DeliverDevices(kt *kit.Kit, order *types.ApplyOrder, observeDevices []*types.DeviceInfo) error
	// FinalApplyStep after deliver device, update generate record status and order status
	FinalApplyStep(kt *kit.Kit, genRecord *types.GenerateRecord, order *types.ApplyOrder) error
	// GetMatcher get matcher
	GetMatcher() *matcher.Matcher
	// GetGenerator get generator
	GetGenerator() *generator.Generator

	// CheckRollingServerHost check rolling server host
	CheckRollingServerHost(kt *kit.Kit, param *types.CheckRollingServerHostReq) (
		*types.CheckRollingServerHostResp, error)

	// CancelApplyTicketItsm cancel apply ticket which in itsm
	CancelApplyTicketItsm(kt *kit.Kit, req *types.CancelApplyTicketItsmReq) error
	// CancelApplyTicketCrp cancel apply ticket which in crp
	CancelApplyTicketCrp(kt *kit.Kit, req *types.CancelApplyTicketCrpReq) error
	// VerifyCvmGPUChargeMonth verify cvm gpu charge month
	VerifyCvmGPUChargeMonth(kt *kit.Kit, subOrders []*types.Suborder) error

	// CreateUpgradeTicketANDOrder create upgrade ticket and order
	CreateUpgradeTicketANDOrder(kt *kit.Kit, param *types.ApplyReq) (*types.CreateUpgradeCrpOrderResult, error)
	// UpdateApplyTicketDemand update apply ticket demand
	UpdateApplyTicketDemand(kt *kit.Kit, param *types.ApplyTicket) error
}

// scheduler provides resource apply service
type scheduler struct {
	lang           language.CCLanguageIf
	itsm           itsm.Client
	cc             cmdb.Client
	dispatcher     *dispatcher.Dispatcher
	generator      *generator.Generator
	matcher        *matcher.Matcher
	recommend      *recommender.Recommender
	configLogics   config.Logics
	rsLogics       rollingserver.Logics
	srLogics       shortrental.Logics
	gcLogics       greenchannel.Logics
	crpCli         cvmapi.CVMClientInterface
	bizLogic       biz.Logics
	bkBotApproval  bkbotapproval.Client
	dissolveLogics dissolve.Logics
	cmsiClient     cmsi.Client
	apiClientSet   *client.ClientSet
}

// New creates a scheduler
func New(ctx context.Context, rsLogics rollingserver.Logics, srLogics shortrental.Logics, gcLogics greenchannel.Logics,
	thirdCli *thirdparty.Client, cmdbCli cmdb.Client, informerIf informer.Interface, clientConf cc.ClientConfig,
	planLogics plan.Logics, bizLogic biz.Logics, configLogics config.Logics, dissolveLogics dissolve.Logics,
	cmsiCli cmsi.Client, apiClientSet *client.ClientSet) (*scheduler, error) {

	// new recommend module
	recommend, err := recommender.New(ctx, thirdCli)
	if err != nil {
		return nil, err
	}

	// new matcher
	match, err := matcher.New(ctx, rsLogics, thirdCli, cmdbCli, clientConf, informerIf, planLogics, configLogics,
		cmsiCli)
	if err != nil {
		return nil, err
	}

	// new generator
	generate, err := generator.New(ctx, rsLogics, thirdCli, cmdbCli, clientConf, configLogics)
	if err != nil {
		return nil, err
	}

	// new dispatcher
	dispatch, err := dispatcher.New(ctx, informerIf)
	if err != nil {
		return nil, err
	}
	dispatch.SetGenerator(generate)

	scheduler := &scheduler{
		lang:           language.NewFromCtx(language.EmptyLanguageSetting),
		itsm:           thirdCli.ITSM,
		crpCli:         thirdCli.CVM,
		cc:             cmdbCli,
		dispatcher:     dispatch,
		generator:      generate,
		matcher:        match,
		recommend:      recommend,
		configLogics:   configLogics,
		rsLogics:       rsLogics,
		srLogics:       srLogics,
		gcLogics:       gcLogics,
		bizLogic:       bizLogic,
		bkBotApproval:  thirdCli.BkBotApproval,
		dissolveLogics: dissolveLogics,
		apiClientSet:   apiClientSet,
	}

	return scheduler, nil
}

// GetDispatcher get dispatcher
func (s *scheduler) GetDispatcher() *dispatcher.Dispatcher {
	return s.dispatcher
}

// GetGenerator get generator
func (s *scheduler) GetGenerator() *generator.Generator { return s.generator }

// UpdateApplyTicket creates or updates resource apply ticket
func (s *scheduler) UpdateApplyTicket(kt *kit.Kit, param *types.ApplyReq) (*types.CreateApplyOrderResult, error) {
	if param.OrderId <= 0 {
		return s.createApplyTicket(kt, param, types.TicketStageUncommit)
	}

	return s.updateApplyTicket(kt, param, types.TicketStageUncommit)
}

func (s *scheduler) createApplyTicket(kt *kit.Kit, param *types.ApplyReq,
	stage types.TicketStage) (*types.CreateApplyOrderResult, error) {

	orderId, err := model.Operation().ApplyOrder().NextSequence(kt.Ctx)
	if err != nil {
		return nil, errf.Newf(pkg.CCErrObjectDBOpErrno, err.Error())
	}

	for _, suborder := range param.Suborders {
		if suborder.Source == "" {
			suborder.Source = enumor.ApplyTicketSrcBusiness
		}
	}
	now := time.Now()
	ticket := &types.ApplyTicket{
		OrderId:      orderId,
		Stage:        stage,
		BkBizId:      param.BkBizId,
		User:         param.User,
		Follower:     param.Follower,
		EnableNotice: param.EnableNotice,
		RequireType:  param.RequireType,
		ExpectTime:   param.ExpectTime,
		Remark:       param.Remark,
		Suborders:    param.Suborders,
		CreateAt:     now,
		UpdateAt:     now,
	}

	logs.V(9).Infof("ticket data: %+v", ticket)

	if err := model.Operation().ApplyTicket().CreateApplyTicket(kt.Ctx, ticket); err != nil {
		logs.Errorf("failed to create apply ticket, err: %v, rid: %s", err, kt.Rid)
		return nil, err
	}

	rst := &types.CreateApplyOrderResult{
		OrderId: orderId,
	}

	return rst, nil
}

func (s *scheduler) updateApplyTicket(kt *kit.Kit, param *types.ApplyReq,
	stage types.TicketStage) (*types.CreateApplyOrderResult, error) {

	filter := mapstr.MapStr{
		"order_id": param.OrderId,
	}

	origin, err := model.Operation().ApplyTicket().GetApplyTicket(kt.Ctx, &filter)
	if err != nil {
		logs.Errorf("failed to update apply ticket, err: %v, rid: %s", err, kt.Rid)
		return nil, err
	}

	if origin.Stage != types.TicketStageUncommit {
		logs.Errorf("failed to update apply ticket, for invalid stage: %s != %s, rid: %s", origin.Stage,
			types.TicketStageUncommit, kt.Rid)
		return nil, fmt.Errorf("invalid ticket stage:%s != %s", origin.Stage, types.TicketStageUncommit)
	}

	update := mapstr.MapStr{
		"order_id":      param.OrderId,
		"stage":         stage,
		"bk_biz_id":     param.BkBizId,
		"bk_username":   param.User,
		"follower":      param.Follower,
		"enable_notice": param.EnableNotice,
		"require_type":  param.RequireType,
		"expect_time":   param.ExpectTime,
		"remark":        param.Remark,
		"suborders":     param.Suborders,
		"update_at":     time.Now(),
	}

	if err := model.Operation().ApplyTicket().UpdateApplyTicket(kt.Ctx, &filter, update); err != nil {
		logs.Errorf("failed to update apply ticket, err: %v, rid: %s", err, kt.Rid)
		return nil, err
	}

	rst := &types.CreateApplyOrderResult{
		OrderId: param.OrderId,
	}

	return rst, nil
}

// GetApplyTicket gets resource apply ticket
func (s *scheduler) GetApplyTicket(kit *kit.Kit, param *types.GetApplyTicketReq) (
	*types.GetApplyTicketRst, error) {

	filter := mapstr.MapStr{
		"order_id": param.OrderId,
	}
	// 业务下查询时，只查询传入业务对应的单据
	if param.BkBizID > 0 && param.BkBizID != constant.UnassignedBiz {
		filter["bk_biz_id"] = param.BkBizID
	}

	inst, err := model.Operation().ApplyTicket().GetApplyTicket(kit.Ctx, &filter)
	if err != nil {
		logs.Errorf("failed to get apply ticket, err: %v, rid: %s", err, kit.Rid)
		return nil, err
	}

	// 转换旧版可用区到新版可用区，方便前端统一展示
	for _, ticket := range inst.Suborders {
		if ticket.Spec != nil && len(ticket.Spec.Zone) > 0 && len(ticket.Spec.Zones) == 0 {
			ticket.Spec.Zones = []string{ticket.Spec.Zone}
			// 分Campus
			if ticket.Spec.Zone == cvmapi.CvmSeparateCampus {
				ticket.Spec.Zones = []string{cvmapi.CvmZoneAll}
				ticket.Spec.ResAssign = enumor.CampusResAssign
			}
		}
	}

	rst := &types.GetApplyTicketRst{
		ApplyTicket: inst,
	}

	return rst, nil
}

// GetApplyAuditItsm gets resource apply ticket audit info
func (s *scheduler) GetApplyAuditItsm(kt *kit.Kit, param *types.GetApplyAuditItsmReq) (
	*types.GetApplyAuditItsmRst, error) {

	filter := mapstr.MapStr{
		"order_id": param.OrderId,
	}
	// 业务下查询时，只查询传入业务对应的单据
	if param.BkBizID > 0 && param.BkBizID != constant.UnassignedBiz {
		filter["bk_biz_id"] = param.BkBizID
	}

	inst, err := model.Operation().ApplyTicket().GetApplyTicket(kt.Ctx, &filter)
	if err != nil {
		logs.Errorf("failed to get apply ticket audit info, err: %v, rid: %s", err, kt.Rid)
		return nil, err
	}

	if inst.ItsmTicketId == "" {
		logs.Errorf("failed to get apply ticket audit info, for itsm ticket sn is empty, rid: %s", kt.Rid)
		return nil, fmt.Errorf("failed to get apply ticket audit info, for itsm ticket sn is empty")
	}

	statusResp, err := s.itsm.GetTicketStatus(kt, inst.ItsmTicketId)
	if err != nil {
		logs.Errorf("failed to get apply ticket audit info, err: %v, rid: %s", err, kt.Rid)
		return nil, err
	}

	if statusResp.Code != 0 {
		logs.Errorf("failed to get apply ticket audit info, code: %d, msg: %s, rid: %s", statusResp.Code,
			statusResp.ErrMsg, kt.Rid)
		return nil, fmt.Errorf("failed to get apply ticket audit info, code: %d, msg: %s", statusResp.Code,
			statusResp.ErrMsg)
	}

	rst, err := s.getItsmApplyAuditRst(kt, param, statusResp, inst)
	if err != nil {
		return nil, err
	}

	return rst, nil
}

func (s *scheduler) getItsmApplyAuditRst(kt *kit.Kit, param *types.GetApplyAuditItsmReq,
	statusResp *itsm.GetTicketStatusResp, inst *types.ApplyTicket) (*types.GetApplyAuditItsmRst, error) {

	status := statusResp.Data.CurrentStatus
	link := statusResp.Data.TicketUrl

	logResp, err := s.itsm.GetTicketLog(kt, inst.ItsmTicketId)
	if err != nil {
		logs.Errorf("failed to get apply ticket audit info, err: %v, rid: %s", err, kt.Rid)
		return nil, err
	}

	if logResp.Code != 0 {
		logs.Errorf("failed to get apply ticket audit info, code: %d, msg: %s, rid: %s", logResp.Code, logResp.ErrMsg,
			kt.Rid)
		return nil, fmt.Errorf("failed to get apply ticket audit info, code: %d, msg: %s", logResp.Code, logResp.ErrMsg)
	}

	rst := &types.GetApplyAuditItsmRst{
		ApplyAuditItsm: &types.ApplyAuditItsm{
			OrderId:        param.OrderId,
			ItsmTicketId:   inst.ItsmTicketId,
			ItsmTicketLink: link,
			Status:         status,
			CurrentSteps:   make([]*types.ApplyAuditItsmStep, 0),
			Logs:           make([]*types.ApplyAuditItsmLog, 0),
		},
	}

	for _, step := range statusResp.Data.CurrentSteps {
		// 校验审批人是否有该业务的访问权限
		processorUsers := strings.Split(step.Processors, ",")
		processorAuth, err := s.bizLogic.BatchCheckUserBizAccessAuth(kt, param.BkBizID, processorUsers)
		if err != nil {
			return nil, err
		}

		rst.CurrentSteps = append(rst.CurrentSteps, &types.ApplyAuditItsmStep{
			Name:           step.Name,
			Processors:     processorUsers,
			StateId:        step.StateId,
			ProcessorsAuth: processorAuth,
		})
	}
	for _, log := range logResp.Data.Logs {
		rst.Logs = append(rst.Logs, &types.ApplyAuditItsmLog{
			Operator:  log.Operator,
			OperateAt: log.OperateAt,
			Message:   log.Message,
			Source:    log.Source,
		})
	}
	return rst, nil
}

// GetApplyAuditCrp gets resource apply ticket audit info
func (s *scheduler) GetApplyAuditCrp(kit *kit.Kit, param *types.GetApplyAuditCrpReq, resType types.ResourceType) (
	*types.GetApplyAuditCrpRst, error) {

	logsReq := cvmapi.NewCvmQueryApproveLogReq(&cvmapi.GetCvmApproveLogParams{OrderId: param.CrpTicketId})
	logsResp, err := s.crpCli.GetCvmApproveLogs(kit.Ctx, kit.Header(), logsReq)
	if err != nil {
		logs.Errorf("failed to get cvm approve logs, err: %v, rid: %s", err, kit.Rid)
		return nil, err
	}

	if logsResp.Error.Code != 0 {
		logs.Errorf("failed to get cvm approve logs, err: %v, trace id: %s, rid: %s",
			logsResp.Error.Message, logsResp.TraceId, kit.Rid)
		return nil, errors.New(logsResp.Error.Message)
	}

	if logsResp.Result == nil || len(logsResp.Result.Data) == 0 {
		logs.Errorf("cvm approve logs is empty, trace id: %s, rid: %s", logsResp.TraceId, kit.Rid)
		return nil, fmt.Errorf("cvm approve logs is empty")
	}

	currentStepStatus := new(cvmapi.OrderItem)
	switch resType {
	case types.ResourceTypeUpgradeCvm:
		currentStepStatus, err = s.getUpgradeCurrentStepStatus(kit, param.CrpTicketId)
		if err != nil {
			logs.Errorf("failed to get upgrade current step status, err: %v, ticket_id: %s, rid: %s", err,
				param.CrpTicketId, kit.Rid)
			return nil, err
		}
	default:
		currentStepStatus, err = s.getCVMCurrentStepStatus(kit, param.CrpTicketId)
		if err != nil {
			logs.Errorf("failed to get cvm current step status, err: %v, ticket_id: %s, rid: %s", err,
				param.CrpTicketId, kit.Rid)
			return nil, err
		}
	}

	resp := &types.GetApplyAuditCrpRst{
		ApplyAuditCrp: &types.ApplyAuditCrp{
			CrpTicketId:   param.CrpTicketId,
			CrpTicketLink: fmt.Sprintf("%s%s", cvmapi.CvmOrderLinkPrefix, param.CrpTicketId),
			CurrentStep: types.ApplyAuditCrpStep{
				CurrentTaskNo:   logsResp.Result.CurrentTaskNo,
				CurrentTaskName: logsResp.Result.CurrentTaskName,
			},
		},
	}

	for _, log := range logsResp.Result.Data {
		resp.Logs = append(resp.Logs, types.ApplyAuditCrpLog{
			TaskNo:        log.TaskNo,
			TaskName:      log.TaskName,
			OperateResult: log.OperateResult,
			Operator:      log.Operator,
			OperateInfo:   log.OperateInfo,
			OperateTime:   log.OperateTime,
		})
	}

	resp = generateApplyAuditCurStep(resp, currentStepStatus)

	return resp, nil
}

func (s *scheduler) getCVMCurrentStepStatus(kt *kit.Kit, crpTicketID string) (*cvmapi.OrderItem, error) {
	orderReq := cvmapi.NewOrderQueryReq(&cvmapi.OrderQueryParam{OrderId: []string{crpTicketID}})
	orderResp, err := s.crpCli.QueryCvmOrders(kt.Ctx, kt.Header(), orderReq)
	if err != nil {
		logs.Errorf("failed to query cvm order, err: %v, rid: %s", err, kt.Rid)
		return nil, err
	}

	if orderResp.Error.Code != 0 {
		logs.Errorf("failed to query cvm order, err: %v, trace id: %s, rid: %s",
			orderResp.Error.Message, orderResp.TraceId, kt.Rid)
		return nil, errors.New(orderResp.Error.Message)
	}

	if orderResp.Result == nil || len(orderResp.Result.Data) == 0 {
		logs.Errorf("crp order is empty, trace id: %s, rid: %s", orderResp.TraceId, kt.Rid)
		return nil, fmt.Errorf("crp order is empty")
	}

	return orderResp.Result.Data[0], nil
}

func (s *scheduler) getUpgradeCurrentStepStatus(kt *kit.Kit, crpTicketID string) (*cvmapi.OrderItem, error) {
	orderReq := cvmapi.NewCvmUpgradeDetailReq(&cvmapi.UpgradeDetailParam{OrderID: crpTicketID})
	orderResp, err := s.crpCli.QueryCvmUpgradeDetail(kt, orderReq)
	if err != nil {
		logs.Errorf("failed to query upgrade order, err: %v, rid: %s", err, kt.Rid)
		return nil, err
	}

	if orderResp.Error.Code != 0 {
		logs.Errorf("failed to query upgrade order, err: %v, trace id: %s, rid: %s",
			orderResp.Error.Message, orderResp.TraceId, kt.Rid)
		return nil, errors.New(orderResp.Error.Message)
	}

	if orderResp.Result == nil {
		logs.Errorf("crp upgrade order is empty, trace id: %s, rid: %s", orderResp.TraceId, kt.Rid)
		return nil, fmt.Errorf("crp upgrade order is empty")
	}

	orderInfo := &cvmapi.OrderItem{
		OrderId:           crpTicketID,
		Status:            int(orderResp.Result.Status),
		StatusDesc:        orderResp.Result.StatusMsg,
		FailInstanceInfos: nil,
		CreateTime:        orderResp.Result.CreateTime,
	}

	if orderResp.Result.Status == enumor.CrpUpgradeOrderFailed {
		orderInfo.FailInstanceInfos = append(orderInfo.FailInstanceInfos, cvmapi.FailInstanceInfo{
			ErrorMsg: orderResp.Result.StatusMsg,
		})
	}
	return orderInfo, nil
}

// generateApplyAuditCurStep generate apply audit current step
func generateApplyAuditCurStep(auditRst *types.GetApplyAuditCrpRst,
	crpOrderRst *cvmapi.OrderItem) *types.GetApplyAuditCrpRst {

	auditRst.CurrentStep.Status = crpOrderRst.Status
	auditRst.CurrentStep.StatusDesc = crpOrderRst.StatusDesc

	for _, failedInfo := range crpOrderRst.FailInstanceInfos {
		auditRst.CurrentStep.FailInstanceInfo = append(auditRst.CurrentStep.FailInstanceInfo, types.FailInstanceInfo{
			ErrorMsgTypeEn: failedInfo.ErrorMsgTypeEn,
			ErrorType:      failedInfo.ErrorType,
			ErrorMsgTypeCn: failedInfo.ErrorMsgTypeCn,
			RequestId:      failedInfo.RequestId,
			ErrorMsg:       failedInfo.ErrorMsg,
			Operator:       failedInfo.Operator,
			ErrorCount:     failedInfo.ErrorCount,
		})
	}

	return auditRst
}

// AuditTicket audit resource apply ticket
func (s *scheduler) AuditTicket(kit *kit.Kit, param *types.ApplyAuditReq) error {
	req := &itsm.OperateNodeReq{
		Sn:         param.ItsmTicketId,
		StateId:    param.StateId,
		Operator:   param.Operator,
		ActionType: itsm.ActionTypeTransition,
		Fields:     make([]*itsm.TicketField, 0),
	}

	keys, ok := itsm.MapStateKey[param.StateId]
	if !ok {
		logs.Errorf("failed to audit apply ticket, invalid state id: %d, rid: %s", param.StateId, kit.Rid)
		return fmt.Errorf("failed to audit apply ticket, invalid state id: %d", param.StateId)
	}

	req.Fields = append(req.Fields, &itsm.TicketField{
		Key:   keys[0],
		Value: strconv.FormatBool(param.Approval),
	})
	req.Fields = append(req.Fields, &itsm.TicketField{
		Key:   keys[1],
		Value: param.Remark,
	})

	resp, err := s.itsm.OperateNode(kit, req)
	if err != nil {
		logs.Errorf("failed to audit apply ticket, err: %v, rid: %s", err, kit.Rid)
		return err
	}

	if resp.Code != 0 {
		logs.Errorf("failed to audit apply ticket, code: %d, msg: %s, rid: %s", resp.Code, resp.ErrMsg, kit.Rid)
		return fmt.Errorf("failed to audit apply ticket, order id: %d, sn: %s, code: %d, msg: %s", param.OrderId,
			param.ItsmTicketId, resp.Code, resp.ErrMsg)
	}

	return nil
}

type checker func(s *scheduler, kit *kit.Kit, order *types.ApplyTicket) (string, bool, error)

// AutoAuditTicket system automatic audit resource apply ticket callback
func (s *scheduler) AutoAuditTicket(kit *kit.Kit, param *types.ApplyAutoAuditReq) (*types.ApplyAutoAuditRst, error) {
	filter := mapstr.MapStr{
		"order_id": param.OrderId,
	}

	order, err := model.Operation().ApplyTicket().GetApplyTicket(kit.Ctx, &filter)
	if err != nil {
		logs.Errorf("failed to auto audit order %d, err: %v, rid: %s", param.OrderId, err, kit.Rid)
		return nil, fmt.Errorf("failed to auto audit order %d, err: %v", param.OrderId, err)
	}

	if order.Stage != types.TicketStageAudit {
		logs.Errorf("failed to auto audit order %d, for invalid stage %s != AUDIT, rid: %s", param.OrderId, order.Stage,
			kit.Rid)
		return nil, fmt.Errorf("order %d is not at AUDIT stage", param.OrderId)
	}

	rst := &types.ApplyAutoAuditRst{
		Operator: "icr",
		Approval: 1,
		Remark:   "approved",
	}

	checkerRules := []checker{
		checkResourceType,
		checkTotalDevice,
		checkRequireType,
	}
	for _, checkerRule := range checkerRules {
		reason, needAudit, err := checkerRule(s, kit, order)
		if err != nil {
			logs.Errorf("failed to check %s, err: %v, rid: %s", reflect.TypeOf(checkerRule).Name(), err, kit.Rid)
			return nil, err
		}

		if needAudit {
			rst.Approval = 0
			rst.Remark = reason
			return rst, nil
		}
	}

	return rst, nil
}

// checkTotalDevice auto audit threshold device number
const auditThresholdDevice = uint(50)

// checkTotalDevice check total device number
func checkTotalDevice(s *scheduler, kt *kit.Kit, order *types.ApplyTicket) (string, bool, error) {
	totalDevice := uint(0)
	for _, suborder := range order.Suborders {
		totalDevice += suborder.Replicas
	}

	threshold, err := getAutoAuditDeviceThreshold(s, kt)
	if err != nil {
		logs.Warnf("failed to get auto audit device threshold from global_config, err: %v, rid: %s", err, kt.Rid)
		// 如果查询失败，使用默认值
		threshold = auditThresholdDevice
	}
	logs.Infof("auto audit device threshold: %d, rid: %s", threshold, kt.Rid)

	if totalDevice > threshold {
		reason := fmt.Sprintf("order %d apply device number %d exceed auto audit threshold %d", order.OrderId,
			totalDevice, threshold)
		return reason, true, nil
	}

	return "", false, nil
}

// getAutoAuditDeviceThreshold 从 global_config 表获取自动审批设备阈值
func getAutoAuditDeviceThreshold(s *scheduler, kt *kit.Kit) (uint, error) {
	req := core.ListReq{
		Filter: tools.ExpressionAnd(
			tools.RuleEqual("config_type", enumor.GlobalConfigTypeAutoAudit),
			tools.RuleEqual("config_key", enumor.GlobalConfigAutoAuditDeviceThreshold),
		),
		Page: core.NewDefaultBasePage(),
	}

	cfgResp, err := s.apiClientSet.DataService().Global.GlobalConfig.List(kt, &req)
	if err != nil {
		logs.Errorf("failed to list global config, err: %v, req: %+v, rid: %s", err, req, kt.Rid)
		return 0, err
	}

	if len(cfgResp.Details) == 0 {
		return 0, errors.New("auto audit device threshold config not found")
	}

	// 解析 config_value，期望是int64
	threshold, err := util.GetInt64ByInterface(cfgResp.Details[0].ConfigValue)
	if err != nil {
		logs.Errorf("failed to convert auto audit device threshold, err: %v, value: %v, rid: %s", err,
			cfgResp.Details[0].ConfigValue, kt.Rid)
		return 0, err
	}

	if threshold < 0 {
		return 0, errors.New("auto audit device must be greater than or equal to 0")
	}

	return uint(threshold), nil
}

// checkRequireType check require type
func checkRequireType(s *scheduler, kit *kit.Kit, order *types.ApplyTicket) (string, bool, error) {
	switch order.RequireType {
	case enumor.RequireTypeGreenChannel:
		greenChannelConfig, err := s.gcLogics.GetConfigs(kit)
		if err != nil {
			return "", false, err
		}
		totalAppliedCore := uint(0)
		for _, suborder := range order.Suborders {
			totalAppliedCore += suborder.AppliedCore
		}
		if int64(totalAppliedCore) > greenChannelConfig.AuditThreshold {
			return fmt.Sprintf("order %d apply core %d exceed green channel auto approval audit threshold %d",
				order.OrderId, totalAppliedCore, greenChannelConfig.AuditThreshold), true, nil
		}
		return "", false, nil

	case enumor.RequireTypeDissolve:
		approvalLimit, err := s.dissolveLogics.Config().GetApprovalLimit(kit)
		if err != nil {
			logs.Errorf("failed to get approval limit, err: %v, rid: %s", err, kit.Rid)
			return "", false, err
		}
		if approvalLimit == nil {
			logs.Errorf("approval limit not exist, rid: %s", kit.Rid)
			return "", false, errors.New("approval limit not exist")
		}
		bizSummaryMap, err := s.dissolveLogics.Table().ListBizCpuCoreSummary(kit, []int64{order.BkBizId})
		if err != nil {
			logs.Errorf("list biz dissolve cpu core summary failed, err: %v, bizID: %d, rid: %s", err, order.BkBizId,
				kit.Rid)
			return "", false, err
		}
		summary, ok := bizSummaryMap[order.BkBizId]
		if !ok {
			logs.Errorf("can not find biz dissolve cpu core summary, bizID: %d, rid: %s", order.BkBizId, kit.Rid)
			return "", false, fmt.Errorf("can not find biz dissolve cpu core summary, bizID: %d", order.BkBizId)
		}
		if summary.TotalCore == 0 {
			logs.Errorf("total core is zero, bizID: %d, rid: %s", order.BkBizId, kit.Rid)
			return "", false, fmt.Errorf("total core is zero, bizID: %d", order.BkBizId)
		}
		cur := float64(summary.DeliveredCore) / float64(summary.TotalCore) * 100
		if cur >= cvt.PtrToVal(approvalLimit) {
			return fmt.Sprintf("delivery percentage greater than limit, cur: %f, limit: %f", cur,
				cvt.PtrToVal(approvalLimit)), true, nil
		}
		return "", false, nil
	default:
		return "", false, nil
	}
}

// checkResourceType ...
func checkResourceType(_ *scheduler, _ *kit.Kit, order *types.ApplyTicket) (string, bool, error) {
	// 所有物理机资源申请，都需要人工审核
	for _, suborder := range order.Suborders {
		if suborder.ResourceType == types.ResourceTypePm {
			reason := fmt.Sprintf("order %d apply resource type %s, but require type is %s",
				order.OrderId, suborder.ResourceType, order.RequireType)
			return reason, true, nil
		}
	}

	return "", false, nil
}

// ApproveTicket approve or reject resource apply ticket
func (s *scheduler) ApproveTicket(kt *kit.Kit, param *types.ApproveApplyReq) error {
	filter := mapstr.MapStr{
		"order_id": param.OrderId,
	}

	stage := types.TicketStageTerminate
	if param.Approval {
		stage = types.TicketStageRunning
	}
	update := mapstr.MapStr{
		"stage":     stage,
		"update_at": time.Now(),
	}

	err := dal.RunTransaction(kt, func(sc mongo.SessionContext) error {
		if err := model.Operation().ApplyTicket().UpdateApplyTicket(sc, &filter, update); err != nil {
			logs.Errorf("failed to update apply ticket, orderId: %d, err: %v, rid: %s", param.OrderId, err, kt.Rid)
			return err
		}

		sessionKit := &kit.Kit{Ctx: sc, User: kt.User, Rid: kt.Rid, AppCode: kt.AppCode, TenantID: kt.TenantID,
			RequestSource: kt.RequestSource}
		if param.Approval {
			if err := s.createSubOrders(sessionKit, param.OrderId); err != nil {
				logs.Errorf("failed to create subOrders, orderId: %d, err: %v, rid: %s", param.OrderId, err, kt.Rid)
				return err
			}
		}

		return nil
	})

	if err != nil {
		update["stage"] = types.TicketStageTerminate
		if updateErr := model.Operation().ApplyTicket().UpdateApplyTicket(kt.Ctx, &filter, update); updateErr != nil {
			logs.Errorf("failed to update apply ticket, orderId: %d, err: %v, rid: %s", param.OrderId, updateErr,
				kt.Rid)
			return updateErr
		}
		logs.Errorf("failed to approve apply ticket %d, err: %v, rid: %s", param.OrderId, err, kt.Rid)
		return err
	}
	return nil
}

func (s *scheduler) createSubOrders(kt *kit.Kit, orderId uint64) error {
	filter := mapstr.MapStr{
		"order_id": orderId,
	}

	ticket, err := model.Operation().ApplyTicket().GetApplyTicket(kt.Ctx, &filter)
	if err != nil {
		logs.Errorf("failed to get apply ticket by filter: %+v, err: %v, rid: %s", filter, err, kt.Rid)
		return err
	}

	// 将资源池采购的子单放到最后创建，防止业务看到的子单号不连续
	subOrders := make([]*types.Suborder, 0)
	purchaseToResPoolSuborders := make([]*types.Suborder, 0)
	for _, suborder := range ticket.Suborders {
		if suborder.Source == enumor.ApplyTicketSrcPurchaseToResPool {
			purchaseToResPoolSuborders = append(purchaseToResPoolSuborders, suborder)
			continue
		}
		if suborder.Source == "" {
			suborder.Source = enumor.ApplyTicketSrcBusiness
		}
		subOrders = append(subOrders, suborder)
	}
	subOrders = append(subOrders, purchaseToResPoolSuborders...)

	now := time.Now()
	suborders := make([]*types.ApplyOrder, len(subOrders))
	purchaseToResPoolUser := cc.WoaServer().ApplyTicketConfig.PurchaseToResourcePool.User
	for index, suborder := range subOrders {
		// TODO: delete debug log
		logs.V(5).Infof("suborder data: %+v", suborder)

		subOrder := &types.ApplyOrder{
			OrderId:           orderId,
			SubOrderId:        fmt.Sprintf("%d-%d", orderId, index+1),
			BkBizId:           ticket.BkBizId,
			User:              ticket.User,
			Follower:          ticket.Follower,
			Auditor:           "",
			RequireType:       ticket.RequireType,
			ExpectTime:        ticket.ExpectTime,
			ResourceType:      suborder.ResourceType,
			Source:            suborder.Source,
			Spec:              suborder.Spec,
			AntiAffinityLevel: suborder.AntiAffinityLevel,
			EnableDiskCheck:   suborder.EnableDiskCheck,
			Description:       ticket.Remark,
			Remark:            suborder.Remark,
			Stage:             types.TicketStageRunning,
			Status:            types.ApplyStatusWaitForMatch,
			OriginNum:         suborder.Replicas,
			TotalNum:          suborder.Replicas,
			PendingNum:        suborder.Replicas,
			SuccessNum:        0,
			AppliedCore:       suborder.AppliedCore,
			ObsProject:        ticket.RequireType.ToObsProject(),
			RetryTime:         0,
			ModifyTime:        0,
			CreateAt:          now,
			UpdateAt:          now,
		}
		if suborder.Source == enumor.ApplyTicketSrcPurchaseToResPool {
			subOrder.Stage = types.TicketStageUncommit
			subOrder.User = purchaseToResPoolUser
		}
		logs.V(4).Infof("suborder data: %+v", subOrder)

		if err := model.Operation().ApplyOrder().CreateApplyOrder(kt.Ctx, subOrder); err != nil {
			logs.Errorf("failed to create apply order, err: %v, rid: %s", err, kt.Rid)
			return err
		}

		// init all step record
		if err := s.initAllSteps(kt, subOrder.SubOrderId, subOrder.TotalNum, subOrder.EnableDiskCheck); err != nil {
			logs.Errorf("failed to init apply step record, err: %v, rid: %s", err, kt.Rid)
			return err
		}

		suborders[index] = subOrder
	}

	if err = s.doCreateOrderPostOp(kt, ticket, suborders); err != nil {
		logs.Errorf("do create order post op failed, err: %v, ticket: %+v, suborders: %v, rid: %s", err,
			cvt.PtrToVal(ticket), suborders, kt.Rid)
		return err
	}

	return nil
}

func (s *scheduler) doCreateOrderPostOp(kt *kit.Kit, ticket *types.ApplyTicket, suborders []*types.ApplyOrder) error {
	switch ticket.RequireType {
	case enumor.RequireTypeRollServer, enumor.RequireTypeSpringResPool:
		if err := s.createRollingAppliedRecord(kt, ticket, suborders); err != nil {
			logs.Errorf("create rolling applied record failed, err: %v, ticket: %+v, rid: %s", err, *ticket, kt.Rid)
			return err
		}

	case enumor.RequireTypeGreenChannel:
		if err := s.canApplyGreenChannelHost(kt, ticket); err != nil {
			logs.Errorf("apply green channel host failed, err: %v, ticket: %+v, rid: %s", err, *ticket, kt.Rid)
			return err
		}
	}

	return nil
}

func (s *scheduler) createRollingAppliedRecord(kt *kit.Kit, ticket *types.ApplyTicket,
	suborders []*types.ApplyOrder) error {

	if len(suborders) == 0 || !ticket.RequireType.IsNeedQuotaManage() {
		return nil
	}

	if len(suborders) != len(ticket.Suborders) {
		logs.Errorf("suborder length(%d) not equal ticket suborders length(%d), orderID: %d, rid: %s", len(suborders),
			len(ticket.Suborders), ticket.OrderId, kt.Rid)
		return fmt.Errorf("suborder length(%d) not equal ticket suborders length(%d), orderID: %d", len(suborders),
			len(ticket.Suborders), ticket.OrderId)
	}

	isResPoolBiz, err := s.rsLogics.IsResPoolBiz(kt, ticket.BkBizId)
	if err != nil {
		logs.Errorf("unable to confirm whether biz is resource pool, err: %v, bizID: %d, rid: %s", err,
			suborders[0].BkBizId, kt.Rid)
		return err
	}

	appliedType := enumor.NormalAppliedType
	if isResPoolBiz {
		appliedType = enumor.ResourcePoolAppliedType
	}

	deviceTypeCountMap := make(map[string]int)
	for _, suborder := range ticket.Suborders {
		if _, ok := deviceTypeCountMap[suborder.Spec.DeviceType]; !ok {
			deviceTypeCountMap[suborder.Spec.DeviceType] = 0
		}
		deviceTypeCountMap[suborder.Spec.DeviceType]++
	}
	count, err := s.rsLogics.GetCpuCoreSum(kt, deviceTypeCountMap)
	if err != nil {
		logs.Errorf("get cpu core sum failed, err: %v, deviceTypeCountMap: %v, rid: %s", err, deviceTypeCountMap,
			kt.Rid)
		return err
	}
	canApply, reason, err := s.rsLogics.CanApplyHost(kt, ticket.BkBizId, uint(count), appliedType)
	if err != nil {
		logs.Errorf("determine can apply host failed, err: %v, ticket: %+v, rid: %s", err, *ticket, kt.Rid)
		return err
	}

	if !canApply {
		logs.Errorf("can not apply host, ticket: %+v, reason: %s, rid: %s", *ticket, reason, kt.Rid)
		return fmt.Errorf("%s", reason)
	}

	records := make([]rstypes.CreateAppliedRecordData, len(suborders))
	for i, suborder := range suborders {
		appliedRecord := rstypes.CreateAppliedRecordData{
			BizID:       suborder.BkBizId,
			OrderID:     suborder.OrderId,
			SubOrderID:  suborder.SubOrderId,
			DeviceType:  suborder.Spec.DeviceType,
			Count:       int(ticket.Suborders[i].Replicas),
			AppliedType: appliedType,
			RequireType: suborder.RequireType,
		}
		records[i] = appliedRecord
	}

	if err = s.rsLogics.CreateAppliedRecord(kt, records); err != nil {
		logs.Errorf("create rolling server applied record failed, err: %v, req: %+v, rid: %s", err, records, kt.Rid)
		return err
	}

	return nil
}

func (s *scheduler) canApplyGreenChannelHost(kt *kit.Kit, ticket *types.ApplyTicket) error {
	if ticket.RequireType != enumor.RequireTypeGreenChannel {
		return nil
	}

	var appliedCount uint = 0
	for _, suborder := range ticket.Suborders {
		appliedCount += suborder.AppliedCore
	}

	canApply, reason, err := s.gcLogics.CanApplyHost(kt, ticket.BkBizId, appliedCount)
	if err != nil {
		logs.Errorf("determine can apply green channel host failed, err: %v, bizID: %d, total: %d, rid: %s", err,
			ticket.BkBizId, appliedCount, kt.Rid)
		return err
	}
	if !canApply {
		logs.Errorf("can not apply green channel host, bizID: %d, reason: %s, rid: %s", ticket.BkBizId, reason, kt.Rid)
		return fmt.Errorf("%s", reason)
	}

	return nil
}

// initAllSteps init apply order all steps
func (s *scheduler) initAllSteps(kt *kit.Kit, suborderId string, total uint,
	enableDiskCheck bool) error {
	// init commit step
	stepID := 1
	if err := record.CreateCommitStep(kt.Ctx, suborderId, total, stepID); err != nil {
		logs.Errorf("order %s failed to create commit step, err: %v, rid: %s", suborderId, err, kt.Rid)
		return err
	}

	// init generate step
	stepID++
	if err := record.CreateGenerateStep(kt.Ctx, suborderId, total, stepID); err != nil {
		logs.Errorf("order %s failed to create generate step, err: %v, rid: %s", suborderId, err, kt.Rid)
		return err
	}

	// init init step
	stepID++
	if err := record.CreateInitStep(kt.Ctx, suborderId, total, stepID); err != nil {
		logs.Errorf("order %s failed to create init step, err: %v, rid: %s", suborderId, err, kt.Rid)
		return err
	}

	if enableDiskCheck {
		// init disk check step
		stepID++
		if err := record.CreateDiskCheckStep(kt.Ctx, suborderId, total, stepID); err != nil {
			logs.Errorf("order %s failed to create disk check step, err: %v, rid: %s", suborderId, err, kt.Rid)
			return err
		}
	}

	// init deliver step
	stepID++
	if err := record.CreateDeliverStep(kt.Ctx, suborderId, total, stepID); err != nil {
		logs.Errorf("order %s failed to create deliver step, err: %v, rid: %s", suborderId, err, kt.Rid)
		return err
	}

	return nil
}

// CreateApplyOrder creates resource apply order
func (s *scheduler) CreateApplyOrder(kt *kit.Kit, param *types.ApplyReq) (*types.CreateApplyOrderResult, error) {
	rst := new(types.CreateApplyOrderResult)
	var err error = nil

	if param.RequireType == enumor.RequireTypeRollServer {
		if err = s.checkRollingServer(kt, param); err != nil {
			logs.Errorf("failed to check rolling server, err: %v, rid: %s", err, kt.Rid)
			return nil, err
		}
	}

	// GPU特殊机型的计费时长校验
	if err = s.VerifyCvmGPUChargeMonth(kt, param.Suborders); err != nil {
		logs.Errorf("failed to verify cvm gpu charge month, err: %v, rid: %s", err, kt.Rid)
		return nil, err
	}

	param, err = s.fillCVMAppliedCore(kt, param)
	if err != nil {
		logs.Errorf("failed to fill applied core, err: %v, rid: %s", err, kt.Rid)
		return nil, err
	}

	txnErr := dal.RunTransaction(kt, func(sc mongo.SessionContext) error {
		sessionKit := &kit.Kit{Ctx: sc, Rid: kt.Rid}
		if param.OrderId <= 0 {
			rst, err = s.createApplyTicket(sessionKit, param, types.TicketStageAudit)
		} else {
			rst, err = s.updateApplyTicket(sessionKit, param, types.TicketStageAudit)
		}
		if err != nil {
			logs.Errorf("failed to create apply order, orderId: %d, err: %v, rid: %s", param.OrderId, err, kt.Rid)
			return err
		}

		resType := types.ResourceTypeCvm
		if len(param.Suborders) > 0 && param.Suborders[0] != nil {
			resType = param.Suborders[0].ResourceType
		}
		resp, err := s.itsm.CreateApplyTicket(sessionKit, param.User, rst.OrderId, param.BkBizId, param.Remark,
			string(resType))
		if err != nil {
			logs.Errorf("failed to create apply order, for create itsm ticket err: %v, rid: %s, orderId: %d, BkBIzId: %d",
				err, kt.Rid, rst.OrderId, param.BkBizId)
			return err
		}

		if resp.Code != 0 {
			logs.Errorf("failed to create apply order, for create itsm ticket err, code: %d, msg: %s, rid: %s, orderId: %d,"+
				" BkBIzId: %d", resp.Code, resp.ErrMsg, kt.Rid, rst.OrderId, param.BkBizId)
			return err
		}

		if err = s.setTicketId(sessionKit, rst.OrderId, resp.Data.Sn); err != nil {
			logs.Errorf("failed to create apply order, for set ticket id err: %v, rid: %s, orderId: %d, sn: %s",
				err, kt.Rid, rst.OrderId, resp.Data.Sn)
			return err
		}
		return nil
	})

	return rst, txnErr
}

// VerifyCvmGPUChargeMonth GPU特殊机型的计费时长校验
func (s *scheduler) VerifyCvmGPUChargeMonth(kt *kit.Kit, subOrders []*types.Suborder) error {
	if len(subOrders) == 0 {
		return nil
	}

	// 计费模式为包年包月，需要根据机型校验年限
	deviceTypes := make([]string, 0)
	for _, suborder := range subOrders {
		if suborder.ResourceType == types.ResourceTypeCvm && suborder.Spec != nil &&
			suborder.Spec.ChargeType == cvmapi.ChargeTypePrePaid {
			deviceTypes = append(deviceTypes, suborder.Spec.DeviceType)
		}
	}

	// 没有符合条件的机型，不需要校验
	if len(deviceTypes) == 0 {
		return nil
	}

	deviceTypeInfoMap, err := s.configLogics.Device().ListDeviceTypeInfoFromCrp(kt, deviceTypes)
	if err != nil {
		logs.Errorf("get crp cvm instance info by device type failed, err: %v, deviceTypes: %v, rid: %s",
			err, deviceTypes, kt.Rid)
		return err
	}

	// 计费模式为包年包月，特殊机型+GPU的机型，计费时长必须为5年
	for _, suborder := range subOrders {
		if suborder.Spec == nil || suborder.Spec.ChargeType != cvmapi.ChargeTypePrePaid {
			continue
		}

		deviceInfo, ok := deviceTypeInfoMap[suborder.Spec.DeviceType]
		if !ok {
			continue
		}

		// 特殊机型+GPU的机型，计费时长必须为5年
		if deviceInfo.InstanceTypeClass == cvmapi.SpecialType &&
			strings.Contains(deviceInfo.InstanceGroup, constant.GpuInstanceClass) &&
			suborder.Spec.ChargeMonths != constant.GPUDeviceTypeChargeMonth {

			logs.Warnf("special gpu instance charge month must be %d months, deviceTypes: %v, deviceType: %s, "+
				"chargeMonth: %d, deviceInfo: %+v, rid: %s", constant.GPUDeviceTypeChargeMonth, deviceTypes,
				suborder.Spec.DeviceType, suborder.Spec.ChargeMonths, deviceInfo, kt.Rid)
			return errf.New(errf.CvmApplyVerifyFailed, fmt.Sprintf("special gpu instance charge month "+
				"must be %d months", constant.GPUDeviceTypeChargeMonth))
		}
	}

	return nil
}

func (s *scheduler) checkRollingServer(kt *kit.Kit, param *types.ApplyReq) error {
	for _, suborder := range param.Suborders {
		// 根据继承的主机信息，获取云主机信息
		ccReq := &getHostFromCCReq{
			CloudInstID: suborder.Spec.InheritInstanceId,
		}
		inheritHostInfo, err := s.getInheritedHostFromCC(kt, ccReq)
		if err != nil {
			err = fmt.Errorf("failed to get inherited host from cc, err: %v", err)
			return err
		}

		// 获取当前选择机型和继承主机的机型信息
		hostInfoMap, err := s.configLogics.Device().ListCvmInstanceInfoByDeviceTypes(kt,
			[]string{suborder.Spec.DeviceType, inheritHostInfo.SvrDeviceClassName})
		if err != nil {
			err = fmt.Errorf("failed to get host info from crp, err: %v", err)
			return err
		}

		selectDeviceType, ok := hostInfoMap[suborder.Spec.DeviceType]
		if !ok {
			err = fmt.Errorf("failed to get select device type info, device type: %s",
				suborder.Spec.DeviceType)
			return err
		}

		inheritDeviceType, ok := hostInfoMap[inheritHostInfo.SvrDeviceClassName]
		if !ok {
			err = fmt.Errorf("failed to get inherit host device type info, device type: %s",
				inheritHostInfo.SvrDeviceClassName)
			return err
		}

		// 判断选择的机型和继承主机的机型是否属于同一个机型族
		if selectDeviceType.DeviceGroup != inheritDeviceType.DeviceGroup {
			err = fmt.Errorf("select device type and inherit device type is not in the same device group,"+
				" select device type: %s, select device type group: %s,"+
				" inherit device type: %s, inherit device type group: %s",
				selectDeviceType.DeviceType, selectDeviceType.DeviceGroup,
				inheritDeviceType.DeviceType, inheritDeviceType.DeviceGroup)
			return err
		}
	}

	return nil
}

func (s *scheduler) fillCVMAppliedCore(kt *kit.Kit, param *types.ApplyReq) (*types.ApplyReq, error) {
	if param == nil {
		logs.Errorf("failed to fill applied core, param is nil, rid: %s", kt.Rid)
		return nil, errf.New(errf.InvalidParameter, "param is nil")
	}

	deviceTypes := make([]string, 0)
	for _, suborder := range param.Suborders {
		if suborder.ResourceType == types.ResourceTypeCvm {
			deviceTypes = append(deviceTypes, suborder.Spec.DeviceType)
		}

	}
	deviceTypeInfoMap, err := s.configLogics.Device().ListCvmInstanceInfoByDeviceTypes(kt, deviceTypes)
	if err != nil {
		logs.Errorf("get cvm instance info by device type failed, err: %v, device_types: %v, rid: %s",
			err, deviceTypes, kt.Rid)
		return nil, err
	}

	for i, suborder := range param.Suborders {
		if suborder.ResourceType != types.ResourceTypeCvm {
			continue
		}

		deviceType := suborder.Spec.DeviceType
		deviceTypeInfo, ok := deviceTypeInfoMap[deviceType]
		if !ok {
			logs.Errorf("can not find device_type, type: %s, rid: %s", deviceType, kt.Rid)
			return nil, fmt.Errorf("can not find device_type, type: %s", deviceType)
		}

		param.Suborders[i].AppliedCore = uint(deviceTypeInfo.CPUAmount) * suborder.Replicas
		// 补充机型族、大小核心数据
		if param.Suborders[i].Spec != nil {
			param.Suborders[i].Spec.DeviceGroup = deviceTypeInfo.DeviceGroup
			param.Suborders[i].Spec.DeviceSize = deviceTypeInfo.CoreType
		}
	}

	return param, nil
}

func (s *scheduler) setTicketId(kt *kit.Kit, orderId uint64, itsmTicketId string) error {
	filter := mapstr.MapStr{
		"order_id": orderId,
	}

	doc := mapstr.MapStr{
		"itsm_ticket_id": itsmTicketId,
		"update_at":      time.Now(),
	}

	if err := model.Operation().ApplyTicket().UpdateApplyTicket(kt.Ctx, &filter, doc); err != nil {
		logs.Errorf("failed to update apply ticket, err: %v, rid: %s", err, kt.Rid)
		return err
	}

	return nil
}

// GetApplyOrder gets resource apply order info
func (s *scheduler) GetApplyOrder(kt *kit.Kit, param *types.GetApplyParam) (*types.GetApplyOrderRst, error) {
	orderFilter := param.GetFilter(false)
	ticketFilter := param.GetFilter(true)

	page := metadata.BasePage{
		Sort:  "-create_at",
		Limit: pkg.BKNoLimit,
		Start: 0,
	}

	tickets, err := model.Operation().ApplyTicket().FindManyApplyTicket(kt.Ctx, page, ticketFilter)
	if err != nil {
		logs.Errorf("get apply ticket failed, err: %v, rid: %s", err, kt.Rid)
		return nil, err
	}

	orders, err := model.Operation().ApplyOrder().FindManyApplyOrder(kt.Ctx, page, orderFilter)
	if err != nil {
		logs.Errorf("get apply order failed, err: %v, rid: %s", err, kt.Rid)
		return nil, err
	}
	mergedOrders := s.mergeApplyTicketOrder(kt, tickets, orders, param.GetProduct)
	total := len(mergedOrders)

	// 翻页超过当前总数，直接返回空列表
	if param.Page.Start > total {
		logs.Warnf("start out of range, cnt: %d, param page: %+v, rid: %s", total, param.Page, kt.Rid)
		return &types.GetApplyOrderRst{
			Count: int64(total),
			Info:  []*types.UnifyOrder{},
		}, nil
	}

	begin := max(0, param.Page.Start)
	end := total
	if param.Page.Limit > 0 {
		end = min(begin+param.Page.Limit, total)
	}

	rst := &types.GetApplyOrderRst{
		Count: int64(total),
		Info:  mergedOrders[begin:end],
	}

	return rst, nil
}

func (s *scheduler) mergeApplyTicketOrder(kt *kit.Kit, tickets []*types.ApplyTicket,
	orders []*types.ApplyOrder, getProduct bool) []*types.UnifyOrder {

	mergeOrders := types.UnifyOrderList{}

	unifyTickets := s.ticketToUnifyOrder(tickets)
	unifyOrders := s.orderToUnifyOrder(kt, orders, getProduct)

	mergeOrders = append(mergeOrders, unifyTickets...)
	mergeOrders = append(mergeOrders, unifyOrders...)

	sort.Sort(sort.Reverse(mergeOrders))
	return mergeOrders
}

func (s *scheduler) ticketToUnifyOrder(tickets []*types.ApplyTicket) []*types.UnifyOrder {
	unifyOrders := make([]*types.UnifyOrder, 0)

	for _, ticket := range tickets {
		total := uint(0)
		for _, suborder := range ticket.Suborders {
			total += suborder.Replicas
		}
		order := &types.UnifyOrder{
			OrderId:     ticket.OrderId,
			BkBizId:     ticket.BkBizId,
			User:        ticket.User,
			RequireType: ticket.RequireType,
			ExpectTime:  ticket.ExpectTime,
			Description: ticket.Remark,
			Stage:       ticket.Stage,
			TotalNum:    total,
			CreateAt:    ticket.CreateAt,
			UpdateAt:    ticket.UpdateAt,
		}

		unifyOrders = append(unifyOrders, order)
	}

	return unifyOrders
}

func (s *scheduler) orderToUnifyOrder(kt *kit.Kit, orders []*types.ApplyOrder, getProduct bool) []*types.UnifyOrder {
	unifyOrders := make([]*types.UnifyOrder, 0)

	for _, order := range orders {
		// 获取实际生产成功的总数量
		productNum := uint(0)
		if getProduct {
			deviceInfos, err := s.matcher.GetUnreleasedDevice(order.SubOrderId)
			if err != nil {
				// 记录日志不影响获取订单信息
				logs.Warnf("order to unify get has product device list failed, subOrderID: %s, err: %v, rid: %s",
					order.SubOrderId, err, kt.Rid)
			}
			productNum = uint(len(deviceInfos))
		}

		// 转换旧版可用区到新版可用区，方便前端统一展示
		if order.Spec != nil && len(order.Spec.Zone) > 0 && len(order.Spec.Zones) == 0 {
			order.Spec.Zones = []string{order.Spec.Zone}
			// 分Campus
			if order.Spec.Zone == cvmapi.CvmSeparateCampus {
				order.Spec.Zones = []string{cvmapi.CvmZoneAll}
				order.Spec.ResAssign = enumor.CampusResAssign
			}
		}

		unifyOrder := &types.UnifyOrder{
			OrderId:           order.OrderId,
			SubOrderId:        order.SubOrderId,
			BkBizId:           order.BkBizId,
			User:              order.User,
			RequireType:       order.RequireType,
			ResourceType:      order.ResourceType,
			ExpectTime:        order.ExpectTime,
			Description:       order.Description,
			Remark:            order.Remark,
			Spec:              order.Spec,
			AntiAffinityLevel: order.AntiAffinityLevel,
			EnableDiskCheck:   order.EnableDiskCheck,
			Stage:             order.Stage,
			Status:            order.Status,
			OriginNum:         order.OriginNum,
			TotalNum:          order.TotalNum,
			SuccessNum:        order.SuccessNum,
			PendingNum:        order.PendingNum,
			ProductNum:        productNum,
			ModifyTime:        order.ModifyTime,
			Source:            order.Source,
			CreateAt:          order.CreateAt,
			UpdateAt:          order.UpdateAt,
		}
		unifyOrders = append(unifyOrders, unifyOrder)
	}

	return unifyOrders
}

// GetApplyDetail gets resource apply order detail info
func (s *scheduler) GetApplyDetail(kit *kit.Kit, param *types.GetApplyDetailReq) (*types.GetApplyDetailRst, error) {
	filter := &mapstr.MapStr{
		"suborder_id": param.SuborderId,
	}

	insts, err := model.Operation().ApplyStep().FindManyApplyStep(kit.Ctx, filter)
	if err != nil {
		logs.Errorf("get apply order detail info failed, err: %v, rid: %s", err, kit.Rid)
		return nil, err
	}

	rst := &types.GetApplyDetailRst{
		Count: int64(len(insts)),
		Info:  insts,
	}

	return rst, nil
}

// GetApplyGenerate gets resource apply order generate records
func (s *scheduler) GetApplyGenerate(kit *kit.Kit, param *types.GetApplyGenerateReq) (*types.GetApplyGenerateRst,
	error) {

	filter, err := param.GetFilter()
	if err != nil {
		logs.Errorf("get apply order generate record failed, err: %v, rid: %s", err, kit.Rid)
		return nil, err
	}
	filter["suborder_id"] = param.SuborderId

	count, err := model.Operation().GenerateRecord().CountGenerateRecord(kit.Ctx, filter)
	if err != nil {
		return nil, err
	}

	insts, err := model.Operation().GenerateRecord().FindManyGenerateRecord(kit.Ctx, param.Page, filter)
	if err != nil {
		return nil, err
	}

	rst := &types.GetApplyGenerateRst{
		Count: int64(count),
		Info:  insts,
	}

	return rst, nil
}

// GetApplyInit gets resource apply order init records
func (s *scheduler) GetApplyInit(kit *kit.Kit, param *types.GetApplyInitReq) (*types.GetApplyInitRst, error) {
	filter, err := param.GetFilter()
	if err != nil {
		logs.Errorf("get apply order init record failed, err: %v, rid: %s", err, kit.Rid)
		return nil, err
	}
	filter["suborder_id"] = param.SuborderId

	count, err := model.Operation().InitRecord().CountInitRecord(kit.Ctx, filter)
	if err != nil {
		return nil, err
	}

	insts, err := model.Operation().InitRecord().FindManyInitRecord(kit.Ctx, param.Page, filter)
	if err != nil {
		return nil, err
	}

	rst := &types.GetApplyInitRst{
		Count: int64(count),
		Info:  insts,
	}

	return rst, nil
}

// GetApplyDiskCheck gets resource apply order disk check records
func (s *scheduler) GetApplyDiskCheck(kit *kit.Kit, param *types.GetApplyInitReq) (*types.GetApplyDiskCheckRst,
	error) {

	filter, err := param.GetFilter()
	if err != nil {
		logs.Errorf("get apply order disk check record failed, err: %v, rid: %s", err, kit.Rid)
		return nil, err
	}
	filter["suborder_id"] = param.SuborderId

	count, err := model.Operation().DiskCheckRecord().CountDiskCheckRecord(kit.Ctx, filter)
	if err != nil {
		return nil, err
	}

	insts, err := model.Operation().DiskCheckRecord().FindManyDiskCheckRecord(kit.Ctx, param.Page, filter)
	if err != nil {
		return nil, err
	}

	rst := &types.GetApplyDiskCheckRst{
		Count: int64(count),
		Info:  insts,
	}

	return rst, nil
}

// GetApplyDeliver gets resource apply order deliver records
func (s *scheduler) GetApplyDeliver(kit *kit.Kit, param *types.GetApplyDeliverReq) (*types.GetApplyDeliverRst, error) {
	filter, err := param.GetFilter()
	if err != nil {
		logs.Errorf("get apply order deliver record failed, err: %v, rid: %s", err, kit.Rid)
		return nil, err
	}
	filter["suborder_id"] = param.SuborderId

	count, err := model.Operation().DeliverRecord().CountDeliverRecord(kit.Ctx, filter)
	if err != nil {
		return nil, err
	}

	insts, err := model.Operation().DeliverRecord().FindManyDeliverRecord(kit.Ctx, param.Page, filter)
	if err != nil {
		return nil, err
	}

	rst := &types.GetApplyDeliverRst{
		Count: int64(count),
		Info:  insts,
	}

	return rst, nil
}

// GetApplyDevice get resource apply delivered devices
func (s *scheduler) GetApplyDevice(kit *kit.Kit, param *types.GetApplyDeviceReq) (*types.GetApplyDeviceRst, error) {
	filter, err := param.GetFilter()
	if err != nil {
		logs.Errorf("failed to get apply order device info, err: %v, rid: %s", err, kit.Rid)
		return nil, err
	}
	// return delivered device only
	filter["is_delivered"] = true

	if param.Page.EnableCount {
		count, err := model.Operation().DeviceInfo().CountDeviceInfo(kit.Ctx, filter)
		if err != nil {
			logs.Errorf("failed to count apply device, filter: %+v, err: %v, rid: %s", filter, err, kit.Rid)
			return nil, err
		}
		return &types.GetApplyDeviceRst{Count: int64(count)}, nil
	}

	insts, err := model.Operation().DeviceInfo().FindManyDeviceInfo(kit.Ctx, param.Page, filter)
	if err != nil {
		logs.Errorf("failed to find apply device, filter: %+v, page: %+v, err: %v, rid: %s",
			filter, param.Page, err, kit.Rid)
		return nil, err
	}

	return &types.GetApplyDeviceRst{Info: insts}, nil
}

// ExportDeliverDevice export resource apply delivered devices
func (s *scheduler) ExportDeliverDevice(kit *kit.Kit, param *types.ExportDeliverDeviceReq) (*types.GetApplyDeviceRst,
	error) {

	filter, err := param.GetFilter()
	if err != nil {
		logs.Errorf("failed to export apply delivered device info, err: %v, rid: %s", err, kit.Rid)
		return nil, err
	}
	// return delivered device only
	filter["is_delivered"] = true

	count, err := model.Operation().DeviceInfo().CountDeviceInfo(kit.Ctx, filter)
	if err != nil {
		return nil, err
	}

	page := metadata.BasePage{
		Start: 0,
		Limit: pkg.BKNoLimit,
	}

	insts, err := model.Operation().DeviceInfo().FindManyDeviceInfo(kit.Ctx, page, filter)
	if err != nil {
		return nil, err
	}

	rst := &types.GetApplyDeviceRst{
		Count: int64(count),
		Info:  insts,
	}

	return rst, nil
}

// GetMatchDevice get resource apply match devices
func (s *scheduler) GetMatchDevice(kit *kit.Kit, param *types.GetMatchDeviceReq) (*types.GetMatchDeviceRst, error) {
	rule := s.buildMatchDeviceFilter(kit, param)
	req := s.buildCMDBQueryRequest(rule)

	resp, err := s.cc.ListBizHost(kit, req)
	if err != nil {
		logs.Errorf("failed to get cc host info, err: %v, rid: %s", err, kit.Rid)
		return nil, err
	}

	return s.convertHostsToMatchDevices(resp.Info, param.PendingNum), nil
}

// buildMatchDeviceFilter builds the filter rule for matching devices
func (s *scheduler) buildMatchDeviceFilter(kit *kit.Kit, param *types.GetMatchDeviceReq) *querybuilder.CombinedRule {
	rule := &querybuilder.CombinedRule{
		Condition: querybuilder.ConditionAnd,
		Rules:     make([]querybuilder.Rule, 0),
	}
	if len(param.Ips) != 0 {
		rule.Rules = append(rule.Rules, querybuilder.AtomRule{
			Field:    "bk_host_innerip",
			Operator: querybuilder.OperatorIn,
			Value:    param.Ips,
		})
		rule.Rules = append(rule.Rules, querybuilder.AtomRule{
			Field:    "bk_cloud_id",
			Operator: querybuilder.OperatorEqual,
			Value:    0,
		})
	}
	if param.Spec != nil {
		s.buildRegionZoneFilter(kit, rule, param)
		s.buildSpecFilter(rule, param.Spec)
	}

	return rule
}

// buildRegionZoneFilter builds region and zone filter rules based on resource type
func (s *scheduler) buildRegionZoneFilter(kt *kit.Kit, rule *querybuilder.CombinedRule,
	param *types.GetMatchDeviceReq) {

	if param.ResourceType != types.ResourceTypeCvm {
		s.buildNonCvmRegionZoneFilter(rule, param.Spec)
	}
	s.buildCvmRegionZoneFilter(kt, rule, param.Spec)
}

// buildNonCvmRegionZoneFilter builds region/zone filter for non-CVM resources
func (s *scheduler) buildNonCvmRegionZoneFilter(rule *querybuilder.CombinedRule, spec *types.MatchSpec) {
	if len(spec.Region) > 0 {
		rule.Rules = append(rule.Rules, querybuilder.AtomRule{
			Field:    "bk_zone_name",
			Operator: querybuilder.OperatorIn,
			Value:    spec.Region,
		})
	}
	if len(spec.Zone) > 0 {
		rule.Rules = append(rule.Rules, querybuilder.AtomRule{
			Field:    "sub_zone",
			Operator: querybuilder.OperatorIn,
			Value:    spec.Zone,
		})
	}
}

// buildCvmRegionZoneFilter builds region/zone filter for CVM resources
func (s *scheduler) buildCvmRegionZoneFilter(kt *kit.Kit, rule *querybuilder.CombinedRule, spec *types.MatchSpec) {
	if len(spec.Zone) > 0 {
		zoneIDs := s.getCvmZoneIDs(kt, spec.Region, spec.Zone)

		rule.Rules = append(rule.Rules, querybuilder.AtomRule{
			Field:    "bk_cloud_zone",
			Operator: querybuilder.OperatorIn,
			Value:    zoneIDs,
		})

	} else if len(spec.Region) > 0 {
		regionIDs := s.getCvmRegionIDs(kt, spec.Region)

		rule.Rules = append(rule.Rules, querybuilder.AtomRule{
			Field:    "bk_cloud_region",
			Operator: querybuilder.OperatorIn,
			Value:    regionIDs,
		})
	}
}

// getCvmZoneIDs gets zone IDs for CVM by querying zone config
func (s *scheduler) getCvmZoneIDs(kt *kit.Kit, regions []string, zones []string) []string {
	req := &cfgtype.GetZoneParam{Zone: zones}
	if len(regions) > 0 {
		req.Region = regions
	}

	zoneResult, err := s.configLogics.Zone().GetZone(kt, req)
	if err != nil {
		logs.Errorf("failed to get zone config, err: %v, rid: %s", err, kt.Rid)
		return nil
	}

	zoneIDs := make([]string, 0, len(zoneResult.Info))
	for _, zone := range zoneResult.Info {
		zoneIDs = append(zoneIDs, zone.Zone)
	}
	return zoneIDs
}

// getCvmRegionIDs gets region IDs for CVM by querying zone config
func (s *scheduler) getCvmRegionIDs(kt *kit.Kit, regions []string) []string {
	req := &cfgtype.GetZoneParam{Region: regions}
	zoneResult, err := s.configLogics.Zone().GetZone(kt, req)
	if err != nil {
		logs.Errorf("failed to get zone config, err: %v, rid: %s", err, kt.Rid)
		return nil
	}

	regionIDs := make([]string, 0)
	for _, zone := range zoneResult.Info {
		regionIDs = append(regionIDs, zone.Region)
	}
	return slice.Unique(regionIDs)
}

// buildSpecFilter builds filter rules for device specifications
func (s *scheduler) buildSpecFilter(rule *querybuilder.CombinedRule, spec *types.MatchSpec) {
	if len(spec.DeviceType) > 0 {
		rule.Rules = append(rule.Rules, querybuilder.AtomRule{
			Field:    "svr_device_class",
			Operator: querybuilder.OperatorIn,
			Value:    spec.DeviceType,
		})
	}
	if len(spec.OsType) > 0 {
		re := regexp.MustCompile(`([.*+?^${}()|[\]\\])`)
		osType := re.ReplaceAllString(spec.OsType, `\$1`)
		rule.Rules = append(rule.Rules, querybuilder.AtomRule{
			Field:    "bk_os_name",
			Operator: querybuilder.OperatorContains,
			Value:    osType,
		})
	}
	if len(spec.RaidType) > 0 {
		rule.Rules = append(rule.Rules, querybuilder.AtomRule{
			Field:    "raid_name",
			Operator: querybuilder.OperatorIn,
			Value:    spec.RaidType,
		})
	}
	if len(spec.Isp) > 0 {
		rule.Rules = append(rule.Rules, querybuilder.AtomRule{
			Field:    "bk_ip_oper_name",
			Operator: querybuilder.OperatorIn,
			Value:    spec.Isp,
		})
	}
	if len(spec.InstanceChargeType) > 0 {
		rule.Rules = append(rule.Rules, querybuilder.AtomRule{
			Field:    "instance_charge_type",
			Operator: querybuilder.OperatorEqual,
			Value:    spec.InstanceChargeType,
		})
	}
}

// buildCMDBQueryRequest builds CMDB query request
func (s *scheduler) buildCMDBQueryRequest(rule *querybuilder.CombinedRule) *cmdb.ListBizHostParams {
	req := &cmdb.ListBizHostParams{
		BizID:       931,
		BkModuleIDs: []int64{239149},
		Fields: []string{
			"bk_host_id",
			"bk_asset_id",
			"bk_host_innerip",
			"bk_host_outerip",
			// 外网运营商
			"bk_ip_oper_name",
			// 机型
			"svr_device_class",
			"bk_os_name",
			// 地域
			"bk_zone_name",
			// 可用区
			"sub_zone",
			"module_name",
			// 机架号，string类型
			"rack_id",
			"idc_unit_name",
			// 逻辑区域
			"logic_domain",
			"raid_name",
			"svr_input_time",
			"instance_charge_type",
			"billing_start_time",
			"billing_expire_time",
		},
		Page: &cmdb.BasePage{
			Start: 0,
			Limit: pkg.BKMaxInstanceLimit,
		},
	}
	if len(rule.Rules) > 0 {
		req.HostPropertyFilter = &cmdb.QueryFilter{
			Rule: rule,
		}
	}

	return req
}

// convertHostsToMatchDevices converts CMDB hosts to match devices
func (s *scheduler) convertHostsToMatchDevices(hosts []cmdb.Host, pendingNum int64) *types.GetMatchDeviceRst {
	// TODO: filter and sort devices
	rst := &types.GetMatchDeviceRst{
		Count: 0,
		Info:  make([]*types.MatchDevice, 0),
	}
	tagNum := int64(0)
	for _, host := range hosts {
		rackId, err := strconv.Atoi(host.RackId)
		if err != nil {
			logs.Warnf("failed to convert host %d rack_id %s to int", host.BkHostID, host.RackId)
			rackId = 0
		}
		tag := false
		if tagNum < pendingNum {
			tag = true
			tagNum++
		}
		device := &types.MatchDevice{
			BkHostId:           host.BkHostID,
			AssetId:            host.BkAssetID,
			Ip:                 host.GetUniqIp(),
			OuterIp:            host.BkHostOuterIP,
			Isp:                host.BkIpOerName,
			DeviceType:         host.SvrDeviceClass,
			OsType:             host.BkOSName,
			Region:             host.BkZoneName,
			Zone:               host.SubZone,
			Module:             host.ModuleName,
			Equipment:          int64(rackId),
			IdcUnit:            host.IdcUnitName,
			IdcLogicArea:       host.LogicDomain,
			RaidType:           host.RaidName,
			InputTime:          host.SvrInputTime,
			MatchScore:         1.0,
			MatchTag:           tag,
			InstanceChargeType: host.InstanceChargeType,
			BillingStartTime:   host.BillingStartTime,
			BillingExpireTime:  host.BillingExpireTime,
		}

		rst.Info = append(rst.Info, device)
	}
	rst.Count = int64(len(rst.Info))
	return rst
}

// MatchDevice execute resource apply match devices
func (s *scheduler) MatchDevice(kt *kit.Kit, param *types.MatchDeviceReq) error {
	// get order by suborder id
	order, err := s.generator.GetApplyOrder(param.SuborderId)
	if err != nil {
		logs.Errorf("failed to match cvm when get apply order, err: %v, order id: %s, rid: %s", err, param.SuborderId,
			kt.Rid)
		return fmt.Errorf("failed to match cvm, err: %v, order id: %s", err, param.SuborderId)
	}

	// 获取实际生产成功的数量
	deviceInfos, err := s.matcher.GetUnreleasedDevice(param.SuborderId)
	if err != nil {
		logs.Errorf("failed to get product device info, subOrderID: %s, err: %v, rid: %s",
			param.SuborderId, err, kt.Rid)
		return err
	}

	productSuccCount := len(deviceInfos)
	// 手工匹配的数量 + 已生产的数量 不能大于 需要交付的总数量
	if len(param.Device)+productSuccCount > int(order.TotalNum) {
		logs.Errorf("match device && successfully generator amount exceeds origin requirement, subOrderID: %s, "+
			"productNum: %d, TotalNum: %d, rid: %s", order.SubOrderId, productSuccCount, order.TotalNum, kt.Rid)
		return fmt.Errorf("match device && successfully generator amount %d exceeds origin value %d",
			len(param.Device)+productSuccCount, order.TotalNum)
	}

	if err = s.generator.MatchCVM(kt, param, order); err != nil {
		logs.Errorf("failed to match devices, err: %v, rid: %s", err, kt.Rid)
		return err
	}

	return nil
}

// MatchPoolDevice execute resource apply match devices from resource pool
func (s *scheduler) MatchPoolDevice(_ *kit.Kit, param *types.MatchPoolDeviceReq) error {
	go s.generator.MatchPoolDevice(param)

	return nil
}

// PauseApplyOrder pauses resource apply order
func (s *scheduler) PauseApplyOrder(_ *kit.Kit, _ mapstr.MapStr) error {
	// TODO
	return nil
}

// ResumeApplyOrder resumes resource apply order
func (s *scheduler) ResumeApplyOrder(_ *kit.Kit, _ mapstr.MapStr) error {
	// TODO
	return nil
}

// StartApplyOrder starts resource apply order
func (s *scheduler) StartApplyOrder(kt *kit.Kit, param *types.StartApplyOrderReq) error {
	filter := map[string]interface{}{
		"suborder_id": mapstr.MapStr{
			pkg.BKDBIN: param.SuborderID,
		},
	}

	page := metadata.BasePage{}

	insts, err := model.Operation().ApplyOrder().FindManyApplyOrder(kt.Ctx, page, filter)
	if err != nil {
		logs.Errorf("failed to get apply order, err: %v, rid: %s", err, kt.Rid)
		return err
	}

	cnt := len(insts)
	if cnt == 0 {
		logs.Errorf("found no apply order to start, orderNum: %d, rid: %s", cnt, kt.Rid)
		return fmt.Errorf("found no apply order to start")
	}

	// check status
	for _, order := range insts {
		// cannot start apply order if its stage is not SUSPEND、CONFIRMING
		if order.Stage != types.TicketStageSuspend && order.Stage != types.TicketStageConfirming {
			logs.Errorf("cannot start order %s, for its stage %s != %s and %s, rid: %s", order.SubOrderId, order.Stage,
				types.TicketStageSuspend, types.TicketStageConfirming, kt.Rid)
			return fmt.Errorf("cannot start order %s, for its stage %s != %s and %s", order.SubOrderId, order.Stage,
				types.TicketStageSuspend, types.TicketStageConfirming)
		}

		// TODO 暂不支持重试升降配类型单据
		if order.ResourceType == types.ResourceTypeUpgradeCvm {
			logs.Errorf("cannot start order %s, for its resource type is %s, rid: %s", order.SubOrderId,
				order.ResourceType, kt.Rid)
			return fmt.Errorf("CVM升降配单据暂不支持重试")
		}
	}

	// set order status wait
	if err = s.startOrder(kt, insts); err != nil {
		logs.Errorf("failed to start apply order, err: %v, rid: %s", err, kt.Rid)
		return fmt.Errorf("failed to start apply order, err: %v", err)
	}

	return nil
}

func (s *scheduler) startOrder(kt *kit.Kit, orders []*types.ApplyOrder) error {
	now := time.Now()
	for _, order := range orders {
		// cannot start apply order if its stage is not SUSPEND、CONFIRMING
		if order.Stage != types.TicketStageSuspend && order.Stage != types.TicketStageConfirming {
			logs.Errorf("cannot start order %s, for its stage %s != %s and %s, rid: %s", order.SubOrderId, order.Stage,
				types.TicketStageSuspend, types.TicketStageConfirming, kt.Rid)
			return fmt.Errorf("cannot start order %s, for its stage %s != %s and %s", order.SubOrderId, order.Stage,
				types.TicketStageSuspend, types.TicketStageConfirming)
		}

		if err := s.startSubOrderFailedStep(kt, order.SubOrderId); err != nil {
			logs.Errorf("failed to start order failed step, err: %v, sub orderID: %s, rid: %s", err, order.SubOrderId,
				kt.Rid)
			return err
		}

		filter := &mapstr.MapStr{
			"suborder_id": order.SubOrderId,
		}

		update := &mapstr.MapStr{
			"stage":      types.TicketStageRunning,
			"status":     types.ApplyStatusWaitForMatch,
			"retry_time": 0,
			"update_at":  now,
		}

		if err := model.Operation().ApplyOrder().UpdateApplyOrder(context.Background(), filter, update); err != nil {
			logs.Errorf("failed to set order %s running, err: %v, rid: %s", order.SubOrderId, err, kt.Rid)
			return fmt.Errorf("failed to set order %s running, err: %v", order.SubOrderId, err)
		}

		go func(suborderID string) {
			if err := s.retryFailedDevices(kt, suborderID); err != nil {
				logs.Errorf("failed to retry failed devices, err: %v, sub orderID: %s, rid: %s", err, suborderID,
					kt.Rid)
			}
		}(order.SubOrderId)
	}

	return nil
}

func (s *scheduler) startSubOrderFailedStep(kt *kit.Kit, subOrderID string) error {
	filter := mapstr.MapStr{
		"suborder_id": subOrderID,
		"status":      types.StepStatusFailed,
	}

	now := time.Now()
	doc := mapstr.MapStr{
		"status":    types.StepStatusHandling,
		"message":   types.StepMsgHandling,
		"start_at":  now,
		"update_at": now,
	}

	if err := model.Operation().ApplyStep().UpdateApplyStep(kt.Ctx, &filter, &doc); err != nil {
		logs.Errorf("failed to start order failed step, err: %v, sub orderID: %s, rid: %s", err, subOrderID, kt.Rid)
		return err
	}

	return nil
}

// retryTimeoutMin 重试失败主机对应超时分钟数
const retryTimeoutMin = 10

func (s *scheduler) retryFailedDevices(oldKt *kit.Kit, subOrderID string) error {
	kt := core.NewBackendKit()
	kt.Rid = oldKt.Rid

	timeout := retryTimeoutMin * time.Minute
	startTime := time.Now()

	for {
		if time.Since(startTime) > timeout {
			logs.Errorf("retry failed devices timeout, sub order id: %s, rid: %s", subOrderID, kt.Rid)
			return fmt.Errorf("retry failed devices timeout, sub order id: %s", subOrderID)
		}

		filter := &mapstr.MapStr{"suborder_id": subOrderID}
		order, err := model.Operation().ApplyOrder().GetApplyOrder(kt.Ctx, filter)
		if err != nil {
			logs.Errorf("failed to get apply order, err: %v, sub order id: %s, rid: %s", err, subOrderID, kt.Rid)
			return err
		}

		// 子单的状态是types.ApplyStatusWaitForMatch时，表示子单在主流程等待调度，此时也需要在这里进行等待
		if order.Status == types.ApplyStatusWaitForMatch {
			logs.Warnf("sleep one second, order status is %s, sub order id: %s, rid: %s", order.Status, subOrderID,
				kt.Rid)
			time.Sleep(time.Second)
			continue
		}

		// 如果子单的状态是types.ApplyStatusMatching时，表示子单正在生产主机，此时可以进行失败主机重试，否则返回错误
		if order.Status != types.ApplyStatusMatching {
			logs.Errorf("order status is not matching, id: %s, status: %s, rid: %s", subOrderID, order.Status, kt.Rid)
			return fmt.Errorf("order status is not matching, id: %s, status: %s", subOrderID, order.Status)
		}

		devices, err := model.Operation().DeviceInfo().GetDeviceInfo(kt.Ctx, filter)
		if err != nil {
			logs.Errorf("failed to get binding devices to sub order id: %s, err: %v, rid: %s", subOrderID, err, kt.Rid)
			return err
		}

		// 在主机生产的主流程中，如果生产的数量已经大于子单所需要的数量时，也会触发失败主机的重试，所以这里不需要重复操作了
		if len(devices) >= int(order.TotalNum) {
			logs.Infof("devices len(%d) greater than order total(%d), sub order id: %s, rid: %s", len(devices),
				order.TotalNum, subOrderID, kt.Rid)
			return nil
		}

		genIDs := make([]int64, 0)
		for _, device := range devices {
			if !device.IsDelivered {
				genIDs = append(genIDs, int64(device.GenerateId))
			}
		}
		if len(genIDs) == 0 {
			return nil
		}

		genIDs = slice.Unique(genIDs)
		filter = &mapstr.MapStr{"generate_id": &mapstr.MapStr{pkg.BKDBIN: genIDs}}
		update := mapstr.MapStr{"is_matched": false, "update_at": time.Now()}
		if err = model.Operation().GenerateRecord().UpdateGenerateRecord(kt.Ctx, filter, &update); err != nil {
			logs.Errorf("failed to update generate record, err: %v, generate ids: %v, sub order id: %s, update: %+v, "+
				"rid: %s", err, genIDs, subOrderID, update, kt.Rid)
			return err
		}
		logs.Infof("update generate record, generate ids: %v, sub order id: %s, rid: %s", genIDs, subOrderID, kt.Rid)

		return nil
	}
}

// TerminateApplyOrder terminates resource apply order
func (s *scheduler) TerminateApplyOrder(kit *kit.Kit, param *types.TerminateApplyOrderReq) error {
	filter := map[string]interface{}{
		"suborder_id": mapstr.MapStr{
			pkg.BKDBIN: param.SuborderID,
		},
	}

	page := metadata.BasePage{}

	insts, err := model.Operation().ApplyOrder().FindManyApplyOrder(kit.Ctx, page, filter)
	if err != nil {
		logs.Errorf("failed to get apply order, err: %v, rid: %s", err, kit.Rid)
		return err
	}

	cnt := len(insts)
	if cnt == 0 {
		logs.Errorf("found no apply order to terminate, rid: %s", kit.Rid)
		return fmt.Errorf("found no apply order to terminate")
	}

	// check status
	for _, order := range insts {
		// cannot terminate apply order if its stage is not SUSPEND
		if order.Stage != types.TicketStageSuspend {
			logs.Errorf("cannot terminate order %s, for its stage %s != %s", order.SubOrderId, order.Stage,
				types.TicketStageSuspend)
			return fmt.Errorf("cannot terminate order %s, for its stage %s != %s", order.SubOrderId, order.Stage,
				types.TicketStageSuspend)
		}
	}

	// set order status terminate
	if err := s.terminateOrder(insts); err != nil {
		logs.Errorf("failed to terminate apply order, err: %v", err)
		return fmt.Errorf("failed to terminate apply order, err: %v", err)
	}

	return nil
}

func (s *scheduler) terminateOrder(orders []*types.ApplyOrder) error {
	now := time.Now()
	for _, order := range orders {
		// cannot terminate apply order if its stage is not SUSPEND
		if order.Stage != types.TicketStageSuspend {
			logs.Errorf("cannot terminate order %s, for its stage %s != %s", order.SubOrderId, order.Status,
				types.TicketStageSuspend)
			return fmt.Errorf("cannot terminate order %s, for its stage %s != %s", order.SubOrderId, order.Stage,
				types.TicketStageSuspend)
		}

		filter := &mapstr.MapStr{
			"suborder_id": order.SubOrderId,
		}

		update := &mapstr.MapStr{
			"stage":     types.TicketStageTerminate,
			"update_at": now,
		}

		if err := model.Operation().ApplyOrder().UpdateApplyOrder(context.Background(), filter, update); err != nil {
			logs.Warnf("failed to set order %s terminate, err: %v", order.SubOrderId, err)
			return fmt.Errorf("failed to set order %s terminate, err: %v", order.SubOrderId, err)
		}
	}

	return nil
}

// ModifyApplyOrder 修改需求重试
func (s *scheduler) ModifyApplyOrder(kt *kit.Kit, param *types.ModifyApplyReq) error {
	filter := &mapstr.MapStr{
		"suborder_id": mapstr.MapStr{
			pkg.BKDBEQ: param.SuborderID,
		},
	}

	order, err := model.Operation().ApplyOrder().GetApplyOrder(kt.Ctx, filter)
	if err != nil {
		logs.Errorf("failed to get apply order, err: %v, rid: %s", err, kt.Rid)
		return err
	}

	if order.Source == enumor.ApplyTicketSrcPurchaseToResPool {
		logs.Errorf("cannot modify order %s, for its source is %s, rid: %s", order.SubOrderId, order.Source, kt.Rid)
		return fmt.Errorf("cannot modify order %s, for its source is %s", order.SubOrderId, order.Source)
	}

	// cannot modify apply order if its stage is not SUSPEND
	if order.Stage != types.TicketStageSuspend {
		logs.Errorf("cannot modify order %s, for its stage %s != %s, rid: %s", order.SubOrderId, order.Stage,
			types.TicketStageSuspend, kt.Rid)
		return fmt.Errorf("cannot modify order %s, for its stage %s != %s", order.SubOrderId, order.Stage,
			types.TicketStageSuspend)
	}

	// TODO 暂不支持重试升降配类型单据
	if order.ResourceType == types.ResourceTypeUpgradeCvm {
		logs.Errorf("cannot modify order %s, for its resource type is %s, rid: %s", order.SubOrderId,
			order.ResourceType, kt.Rid)
		return fmt.Errorf("CVM升降配暂不支持修改单据重试")
	}

	// validate modification
	if err = s.validateModification(kt, order, param); err != nil {
		logs.Errorf("modification is invalid, subOrderID: %s, err: %v, rid: %s", order.SubOrderId, err, kt.Rid)
		return err
	}

	// 订单申请人本人修改，可以直接修改，不需要发送消息确认
	if kt.User == order.User {
		// modify apply order
		if err = s.modifyOrder(kt, order, param); err != nil {
			logs.Errorf("failed to modify apply order, subOrderID: %s, err: %v, rid: %s", order.SubOrderId, err, kt.Rid)
			return fmt.Errorf("failed to modify apply order, err: %v", err)
		}

		// create apply order modify record
		if _, err = s.createModifyRecord(kt, order, param, enumor.ApprovedCvmModifyStatus); err != nil {
			logs.Errorf("failed to create apply order modify record, subOrderID: %s, err: %v, rid: %s",
				order.SubOrderId, err, kt.Rid)
			return err
		}
		return nil
	}

	// create apply order modify record
	modifyID, err := s.createModifyRecord(kt, order, param, enumor.ApprovalPendingCvmModifyStatus)
	if err != nil {
		logs.Errorf("failed to create apply order modify record, subOrderID: %s, err: %v, rid: %s",
			order.SubOrderId, err, kt.Rid)
		return err
	}

	// 给用户发送确认消息
	if err = s.sendConfirmMessage(kt, order, modifyID, param); err != nil {
		logs.Errorf("failed to send apply order modify message, subOrderID: %s, modifyID: %s, err: %v, rid: %s",
			order.SubOrderId, modifyID, err, kt.Rid)
		return err
	}

	// modify apply order confirming
	if err = s.modifyOrderStatusConfirming(kt, order); err != nil {
		logs.Errorf("failed to modify apply order confirming, subOrderID: %s, err: %v, rid: %s",
			order.SubOrderId, err, kt.Rid)
		return fmt.Errorf("failed to modify apply order confirming, err: %v", err)
	}

	return nil
}

// sendConfirmMessage 发送[蓝鲸审批助手]消息
func (s *scheduler) sendConfirmMessage(kt *kit.Kit, order *types.ApplyOrder, modifyID uint64,
	param *types.ModifyApplyReq) error {

	// 主机申请单的详情链接
	applyOrderURL := fmt.Sprintf(s.itsm.GetItsmConfig().ApplyLinkFormat, order.OrderId, order.BkBizId, order.BkBizId,
		order.ResourceType)
	// 同意的Action
	approveActionJSON, err := json.Marshal(bkbotapproval.CvmApplyModifyConfirmCallbackData{
		BkBizID:    order.BkBizId,
		Action:     bkbotapproval.ApproveActionType,
		SuborderID: order.SubOrderId,
		ModifyID:   modifyID,
	})
	if err != nil {
		logs.Errorf("failed to marshal approve action, subOrderID: %s, modifyID: %d, err: %v, rid: %s",
			order.SubOrderId, modifyID, err, kt.Rid)
		return err
	}

	// 拒绝的Action
	rejectActionJSON, err := json.Marshal(bkbotapproval.CvmApplyModifyConfirmCallbackData{
		BkBizID:    order.BkBizId,
		Action:     bkbotapproval.RejectActionType,
		SuborderID: order.SubOrderId,
		ModifyID:   modifyID,
	})
	if err != nil {
		logs.Errorf("failed to marshal reject action, subOrderID: %s, modifyID: %d, err: %v, rid: %s",
			order.SubOrderId, modifyID, err, kt.Rid)
		return err
	}

	modifyCompare := getModifyApplyCompare(order, param)
	loc, err := time.LoadLocation(cc.WoaServer().LocalTimezone)
	if err != nil {
		logs.Warnf("get location time zone: %s failed, err: %v, rid: %s", cc.WoaServer().LocalTimezone, err, kt.Rid)
		loc = time.UTC
	}
	createTime := order.CreateAt.In(loc).Format(constant.DateTimeLayout)
	modifyTime := time.Now().In(loc).Format(constant.DateTimeLayout)
	callbackURL := fmt.Sprintf("%s/api/v1/woa/bizs/%d/task/confirm/apply/record/modify",
		cc.WoaServer().BkApigwHCMURL, order.BkBizId)
	req := &bkbotapproval.SendMessageTplReq{
		Title:     "海垒主机申请单需求调整授权",
		Approvers: []string{order.User},
		Receiver:  []string{order.User},
		Summary: fmt.Sprintf(`单号：[%s](%s)\n提单时间：%s\n\n\n%s%s%s%s%s\n\n\n%s%s%s%s%s\n\n\n修改人：%s\n修改时间：%s\n
注意：如6小时内未确认操作，将自动释放本次修改`, order.SubOrderId, applyOrderURL, createTime, modifyCompare.PreDeviceType,
			modifyCompare.PreZone, modifyCompare.PreNum, modifyCompare.PreVpc, modifyCompare.PreSubnet,
			modifyCompare.CurDeviceType, modifyCompare.CurZone, modifyCompare.CurNum, modifyCompare.CurVpc,
			modifyCompare.CurSubnet, kt.User, modifyTime),
		Actions: []bkbotapproval.MessageAction{
			{
				Name:         "确认修改",
				Color:        bkbotapproval.GreenButtonColor,
				CallbackURL:  callbackURL,
				CallbackData: approveActionJSON,
			},
			{
				Name:         "拒绝修改",
				Color:        bkbotapproval.RedButtonColor,
				CallbackURL:  callbackURL,
				CallbackData: rejectActionJSON,
			},
		},
	}
	err = s.bkBotApproval.SendMessage(kt, req)
	if err != nil {
		logs.Errorf("failed to send bkbotapproval modify message, subOrderID: %s, modifyID: %d, err: %v, rid: %s",
			order.SubOrderId, modifyID, err, kt.Rid)
		return err
	}

	return nil
}

// getModifyApplyCompare 获取变更前后的对比信息
func getModifyApplyCompare(order *types.ApplyOrder, param *types.ModifyApplyReq) types.ConfirmApplyModifyCompare {
	modifyCompare := types.ConfirmApplyModifyCompare{}
	// 机型是否有变化
	modifyCompare.PreDeviceType = fmt.Sprintf("修改前机型：%s\n", order.Spec.DeviceType)
	modifyCompare.CurDeviceType = fmt.Sprintf("修改后机型：%s", param.Spec.DeviceType)
	if order.Spec.DeviceType != param.Spec.DeviceType {
		modifyCompare.CurDeviceType += "<font color=red>（有调整）</font>"
	}
	modifyCompare.CurDeviceType += "\n"

	// 园区是否有变化
	oldZones := getApplyOrderZones(order.Spec)
	newZones := getApplyOrderZones(param.Spec)
	modifyCompare.PreZone = fmt.Sprintf("修改前园区：%s\n", strings.Join(oldZones, "、"))
	modifyCompare.CurZone = fmt.Sprintf("修改后园区：%s", strings.Join(newZones, "、"))
	if order.Spec.DeviceType != param.Spec.DeviceType {
		modifyCompare.CurZone += "<font color=red>（有调整）</font>"
	}
	modifyCompare.CurZone += "\n"

	// 交付数量是否有变化
	modifyCompare.PreNum = fmt.Sprintf("修改前待交付数量：%d\n", order.TotalNum-param.ProductNum)
	modifyCompare.CurNum = fmt.Sprintf("修改后待交付数量：%d\n", param.TotalNum-param.ProductNum)

	// VPC是否有变化
	if order.Spec.Vpc != param.Spec.Vpc {
		modifyCompare.PreVpc = fmt.Sprintf("修改前VPC：%s\n", order.Spec.Vpc)
		modifyCompare.CurVpc = fmt.Sprintf("修改后VPC：%s\n", param.Spec.Vpc)
	}

	// 子网是否有变化
	if order.Spec.Subnet != param.Spec.Subnet {
		modifyCompare.PreSubnet = fmt.Sprintf("修改前子网：%s\n", order.Spec.Subnet)
		modifyCompare.CurSubnet = fmt.Sprintf("修改后子网：%s\n", param.Spec.Subnet)
	}
	return modifyCompare
}

// getApplyOrderZones 获取申请单的可用区
func getApplyOrderZones(spec *types.ResourceSpec) []string {
	if spec == nil {
		return []string{}
	}
	if len(spec.Zones) > 0 {
		return spec.Zones
	}
	if len(spec.Zone) > 0 {
		return []string{spec.Zone}
	}
	return []string{}
}

func (s *scheduler) validateModification(kt *kit.Kit, order *types.ApplyOrder, param *types.ModifyApplyReq) error {
	// validate replicas and modify param
	param, err := s.validateReplicasAndModifyParam(kt, order, param)
	if err != nil {
		logs.Errorf("failed to validate modify replicas num, subOrderID: %s, err: %v, rid: %s",
			order.SubOrderId, err, kt.Rid)
		return err
	}

	// validate device type
	if err = s.validateModifyDeviceType(kt, order, param); err != nil {
		logs.Errorf("failed to validate modify device type, subOrderID: %s, err: %v, rid: %s",
			order.SubOrderId, err, kt.Rid)
		return err
	}

	// validate zone
	if err = s.validateModifyZone(kt, order, param); err != nil {
		logs.Errorf("failed to validate modify zone, subOrderID: %s, err: %v, rid: %s", order.SubOrderId, err, kt.Rid)
		return err
	}

	return nil
}

func (s *scheduler) validateReplicasAndModifyParam(kt *kit.Kit, order *types.ApplyOrder, param *types.ModifyApplyReq) (
	*types.ModifyApplyReq, error) {

	if param.Replicas <= 0 {
		logs.Errorf("modified replicas should be positive integer, subOrderID: %s, rid: %s", order.SubOrderId, kt.Rid)
		return nil, errors.New("modified replicas should be positive integer")
	}

	// 获取实际生产成功的数量
	deviceInfos, err := s.matcher.GetUnreleasedDevice(order.SubOrderId)
	if err != nil {
		logs.Errorf("failed to get generate records, subOrderID: %s, err: %v, rid: %s", order.SubOrderId, err, kt.Rid)
		return nil, err
	}

	productSuccCount := uint(len(deviceInfos))
	// 剩余生产数量 + 已生产的数量 不能大于 原始需求总数量
	if param.Replicas+productSuccCount > order.OriginNum {
		logs.Errorf("modified replicas && successfully generator amount exceeds origin value, subOrderID: %s, "+
			"productNum: %d, originNum: %d, rid: %s", order.SubOrderId, productSuccCount, order.OriginNum, kt.Rid)
		return nil, fmt.Errorf("modified replicas && successfully generator amount %d exceeds origin value %d",
			productSuccCount, order.OriginNum)
	}

	// 核心数的校验
	deviceTypeInfoMap, err := s.configLogics.Device().ListCvmInstanceInfoByDeviceTypes(
		kt, []string{param.Spec.DeviceType})
	if err != nil {
		logs.Errorf("modified replicas get cvm instance info by device type failed, err: %v, device_type: %s, rid: %s",
			err, param.Spec.DeviceType, kt.Rid)
		return nil, err
	}

	deviceTypeInfo, ok := deviceTypeInfoMap[param.Spec.DeviceType]
	if !ok {
		logs.Errorf("modified replicas can not find device_type, subOrderID: %s, deviceType: %s, rid: %s",
			order.SubOrderId, param.Spec.DeviceType, kt.Rid)
		return nil, fmt.Errorf("can not find device_type, deviceType: %s", param.Spec.DeviceType)
	}

	// 获取“已生产”的机型及数量并计算“已生产”的总核数
	_, deliverGroupCntMap := calProductDeviceTypeCountMap(deviceInfos, false)
	productSuccCPUCore, _, err := s.matcher.GetCpuCoreSum(kt, deliverGroupCntMap)
	if err != nil {
		logs.Errorf("get product cpu core failed, subOrderID: %s, err: %v, deviceTypeCountMap: %+v, rid: %s",
			order.SubOrderId, err, deliverGroupCntMap, kt.Rid)
		return nil, err
	}

	modifySumCPUCore := uint(deviceTypeInfo.CPUAmount) * param.Replicas
	// 修改的设备类型总核数 + “已生产”的总核数 不能大于 申请总核数
	if modifySumCPUCore+uint(productSuccCPUCore) > order.AppliedCore {
		logs.Errorf("modified replicas cpuCore && delivered cpuCore exceeds applied cpuCore, subOrderID: %s, "+
			"modifySumCPUCore: %d, productSuccCPUCore: %d, deliveredCore: %d, appliedCore: %d, rid: %s",
			order.SubOrderId, modifySumCPUCore, productSuccCPUCore, order.DeliveredCore, order.AppliedCore, kt.Rid)
		return nil, fmt.Errorf("modified replicas cpuCore %d && product cpuCore %d exceeds applied cpuCore %d",
			modifySumCPUCore, productSuccCPUCore, order.AppliedCore)
	}

	// 需要交付的总数量
	param.TotalNum = param.Replicas + productSuccCount
	// 已生产成功的总数量
	param.ProductNum = productSuccCount

	logs.Infof("validate order replicas success, subOrderID: %s, productSuccCount: %d, deviceCore: %d, "+
		"modifySumCPUCore: %d, param: %+v, rid: %s", order.SubOrderId, productSuccCount, deviceTypeInfo.CPUAmount,
		modifySumCPUCore, cvt.PtrToVal(param), kt.Rid)

	return param, nil
}

func (s *scheduler) validateModifyDeviceType(kt *kit.Kit, order *types.ApplyOrder, param *types.ModifyApplyReq) error {
	// 小额绿通需要按常规项目去校验 --story=125266150
	requireType := order.RequireType.ToRequireTypeWhenGetDevice()

	originDeviceGroup, originDeviceSize, err := s.getDeviceGroup(kt, requireType, order.Spec.DeviceType,
		order.Spec.Region, order.Spec.Zone)
	if err != nil {
		logs.Errorf("failed to get device group, err: %v", err)
		return err
	}

	modifiedDeviceGroup, modifiedDeviceSize, err := s.getDeviceGroup(kt, requireType, param.Spec.DeviceType,
		param.Spec.Region, param.Spec.Zone)
	if err != nil {
		logs.Errorf("failed to get device group, err: %v", err)
		return err
	}

	// modification is valid if found no device config
	if originDeviceGroup == "" {
		return nil
	}

	// 机型族不一致
	if originDeviceGroup != modifiedDeviceGroup {
		logs.Errorf("modify device type is invalid, for its device group changed, subOrderID: %s, "+
			"originDeviceGroup: %s, modifiedDeviceGroup: %s, rid: %s",
			order.SubOrderId, originDeviceGroup, modifiedDeviceGroup, kt.Rid)
		return errf.Newf(errf.InvalidParameter, "modify device type is invalid, for its device group changed, "+
			"originDeviceGroup: %s, modifiedDeviceGroup: %s", originDeviceGroup, modifiedDeviceGroup)
	}

	// 大小核心不一致
	if originDeviceSize != modifiedDeviceSize {
		logs.Errorf("modify device size is invalid, for its device group changed, subOrderID: %s, "+
			"originDeviceSize: %s, modifiedDeviceSize: %s, rid: %s",
			order.SubOrderId, originDeviceSize, modifiedDeviceSize, kt.Rid)
		return errf.Newf(errf.InvalidParameter, "modify device type is invalid, for its device size changed, "+
			"originDeviceSize: %s, modifiedDeviceSize: %s", originDeviceSize, modifiedDeviceSize)
	}

	return nil
}

func (s *scheduler) getDeviceGroup(kt *kit.Kit, requireType enumor.RequireType, deviceType, region, zone string) (
	string, string, error) {

	rules := []querybuilder.Rule{
		querybuilder.AtomRule{
			Field:    "device_type",
			Operator: querybuilder.OperatorEqual,
			Value:    deviceType,
		},
		querybuilder.AtomRule{
			Field:    "require_type",
			Operator: querybuilder.OperatorEqual,
			Value:    requireType,
		},
		querybuilder.AtomRule{
			Field:    "region",
			Operator: querybuilder.OperatorEqual,
			Value:    region,
		},
	}

	if zone != "" && zone != cvmapi.CvmSeparateCampus {
		rules = append(rules, querybuilder.AtomRule{
			Field:    "zone",
			Operator: querybuilder.OperatorEqual,
			Value:    zone,
		})
	}

	req := &configtypes.GetDeviceParam{
		Filter: &querybuilder.QueryFilter{
			Rule: querybuilder.CombinedRule{
				Condition: querybuilder.ConditionAnd,
				Rules:     rules,
			},
		},
		Page: metadata.BasePage{
			Limit: 1,
			Start: 0,
		},
	}

	deviceInfo, err := s.configLogics.Device().GetDevice(kt, req)
	if err != nil {
		logs.Errorf("failed to get device info, err: %v", err)
		return "", "", err
	}

	num := len(deviceInfo.Info)
	if num == 0 {
		// return empty when found no device config
		return "", "", nil
	} else if num != 1 {
		logs.Errorf("failed to get device info, for len %d != 1", num)
		return "", "", fmt.Errorf("failed to get device info, for len %d != 1", num)
	}

	// 机型族
	deviceGroup, ok := deviceInfo.Info[0].Label["device_group"]
	if !ok {
		return "", "", errors.New("get invalid empty device group")
	}
	deviceGroupStr, ok := deviceGroup.(string)
	if !ok {
		return "", "", errors.New("get invalid non-string device group")
	}

	// 机型核心类型
	deviceSize, ok := deviceInfo.Info[0].Label["device_size"]
	if !ok {
		return "", "", errors.New("get invalid empty device size")
	}
	deviceSizeStr, ok := deviceSize.(string)
	if !ok {
		return "", "", errors.New("get invalid non-string device size")
	}

	return deviceGroupStr, deviceSizeStr, nil
}

func (s *scheduler) validateModifyZone(kt *kit.Kit, order *types.ApplyOrder, param *types.ModifyApplyReq) error {
	if param.Spec.Region != order.Spec.Region {
		logs.Errorf("validate modify region cannot be modified, subOrderID: %s, rid: %s", order.SubOrderId, kt.Rid)
		return errors.New("region cannot be modified")
	}

	return nil
}

func (s *scheduler) modifyOrder(kt *kit.Kit, order *types.ApplyOrder, param *types.ModifyApplyReq) error {
	now := time.Now()
	// cannot modify apply order if its stage is not SUSPEND、CONFIRMING
	if order.Stage != types.TicketStageSuspend && order.Stage != types.TicketStageConfirming {
		logs.Errorf("cannot modify order %s, for its stage %s != %s and %s, rid: %s", order.SubOrderId, order.Status,
			types.TicketStageSuspend, types.TicketStageConfirming, kt.Rid)
		return fmt.Errorf("cannot modify order %s, for its stage %s != %s and %s", order.SubOrderId, order.Status,
			types.TicketStageSuspend, types.TicketStageConfirming)
	}

	if err := s.startSubOrderFailedStep(kt, order.SubOrderId); err != nil {
		logs.Errorf("failed to start order failed step, err: %v, sub orderID: %s, rid: %s", err, order.SubOrderId,
			kt.Rid)
		return err
	}

	filter := &mapstr.MapStr{
		"suborder_id": order.SubOrderId,
	}

	update := &mapstr.MapStr{
		"spec.region":          param.Spec.Region,
		"spec.zone":            param.Spec.Zone,
		"spec.device_type":     param.Spec.DeviceType,
		"spec.image_id":        param.Spec.ImageId,
		"spec.disk_size":       param.Spec.DiskSize,
		"spec.disk_type":       param.Spec.DiskType,
		"spec.network_type":    param.Spec.NetworkType,
		"spec.vpc":             param.Spec.Vpc,
		"spec.subnet":          param.Spec.Subnet,
		"spec.failed_zone_ids": []string{}, // 修改需求重试时需要清空已失败的可用区，也就是全可用区重试
		"spec.zones":           param.Spec.Zones,
		"spec.res_assign":      param.Spec.ResAssign,
		"stage":                types.TicketStageRunning,
		"status":               types.ApplyStatusWaitForMatch,
		"total_num":            param.TotalNum,
		"pending_num":          param.TotalNum - param.ProductNum,
		"retry_time":           0,
		"modify_time":          order.ModifyTime + 1,
		"update_at":            now,
	}

	if err := model.Operation().ApplyOrder().UpdateApplyOrder(context.Background(), filter, update); err != nil {
		logs.Errorf("failed to set order %s terminate, err: %v, rid: %s", order.SubOrderId, err, kt.Rid)
		return fmt.Errorf("failed to set order %s terminate, err: %v", order.SubOrderId, err)
	}

	go func(suborderID string) {
		if err := s.retryFailedDevices(kt, suborderID); err != nil {
			logs.Errorf("failed to retry failed devices, err: %v, sub orderID: %s, rid: %s", err, suborderID,
				kt.Rid)
		}
	}(order.SubOrderId)

	return nil
}

// modifyOrderStatusConfirming 修改主机申请单状态为“待用户确认”
func (s *scheduler) modifyOrderStatusConfirming(kt *kit.Kit, order *types.ApplyOrder) error {
	// cannot modify apply order if its stage is not SUSPEND
	if order.Stage != types.TicketStageSuspend {
		logs.Errorf("cannot modify order %s, for its stage %s != %s, rid: %s", order.SubOrderId, order.Status,
			types.TicketStageSuspend, kt.Rid)
		return fmt.Errorf("cannot modify order %s, for its stage %s != %s", order.SubOrderId, order.Status,
			types.TicketStageSuspend)
	}

	filter := &mapstr.MapStr{
		"suborder_id": order.SubOrderId,
	}

	update := &mapstr.MapStr{
		"stage":  types.TicketStageConfirming,
		"status": types.ApplyStatusConfirming,
	}

	if err := model.Operation().ApplyOrder().UpdateApplyOrder(context.Background(), filter, update); err != nil {
		logs.Errorf("failed to update order %s status, err: %v, rid: %s", order.SubOrderId, err, kt.Rid)
		return fmt.Errorf("failed to update order %s status, err: %v", order.SubOrderId, err)
	}

	return nil
}

func (s *scheduler) createModifyRecord(kt *kit.Kit, order *types.ApplyOrder, param *types.ModifyApplyReq,
	status enumor.CvmModifyRecordStatus) (uint64, error) {

	id, err := dao.Set().ModifyRecord().NextSequence(kt.Ctx)
	if err != nil {
		logs.Errorf("failed to get modify record next sequence id, subOrderID: %s, err: %v, rid: %s",
			order.SubOrderId, err, kt.Rid)
		return 0, errf.Newf(pkg.CCErrObjectDBOpErrno, err.Error())
	}

	modifyRecord := &table.ModifyRecord{
		ID:         id,
		SuborderID: order.SubOrderId,
		User:       kt.User,
		Details: &table.ModifyDetail{
			PreData: &table.ModifyData{
				TotalNum:    order.TotalNum,
				Region:      order.Spec.Region,
				Zone:        order.Spec.Zone,
				DeviceType:  order.Spec.DeviceType,
				ImageId:     order.Spec.ImageId,
				DiskSize:    order.Spec.DiskSize,
				DiskType:    order.Spec.DiskType,
				NetworkType: order.Spec.NetworkType,
				Vpc:         order.Spec.Vpc,
				Subnet:      order.Spec.Subnet,
				SystemDisk:  order.Spec.SystemDisk,
				DataDisk:    order.Spec.DataDisk,
				Zones:       order.Spec.Zones,
				ResAssign:   order.Spec.ResAssign,
			},
			CurData: &table.ModifyData{
				TotalNum:    param.TotalNum,
				Replicas:    param.Replicas,
				Region:      param.Spec.Region,
				Zone:        param.Spec.Zone,
				DeviceType:  param.Spec.DeviceType,
				ImageId:     param.Spec.ImageId,
				DiskSize:    param.Spec.DiskSize,
				DiskType:    param.Spec.DiskType,
				NetworkType: param.Spec.NetworkType,
				Vpc:         param.Spec.Vpc,
				Subnet:      param.Spec.Subnet,
				SystemDisk:  param.Spec.SystemDisk,
				DataDisk:    param.Spec.DataDisk,
				Zones:       param.Spec.Zones,
				ResAssign:   param.Spec.ResAssign,
			},
		},
		CreateAt: time.Now(),
		Status:   status,
	}

	if err = dao.Set().ModifyRecord().CreateModifyRecord(kt.Ctx, modifyRecord); err != nil {
		logs.Errorf("failed to create modify record, subOrderID: %s, err: %v, rid: %s", order.SubOrderId, err, kt.Rid)
		return 0, err
	}

	return id, nil
}

// updateModifyRecordStatus 更新变更记录状态
func (s *scheduler) updateModifyRecordData(kt *kit.Kit, subOrderID string, modifyID uint64,
	status enumor.CvmModifyRecordStatus) error {

	filter := mapstr.MapStr{
		"id":          modifyID,
		"suborder_id": subOrderID,
	}

	update := &mapstr.MapStr{
		"status":    status,
		"approver":  kt.User,
		"update_at": time.Now(),
	}

	if err := dao.Set().ModifyRecord().UpdateModifyRecord(kt.Ctx, &filter, update); err != nil {
		logs.Errorf("failed to update modify record, err: %v, subOrderID: %s, modifyID: %d, rid: %s",
			err, subOrderID, modifyID, kt.Rid)
		return err
	}

	return nil
}

// RecommendApplyOrder get resource apply order modification recommendation
func (s *scheduler) RecommendApplyOrder(kt *kit.Kit, param *types.RecommendApplyReq) (*types.RecommendApplyRst,
	error) {

	rst, err := s.recommend.GetApplyRecommendation(param.SuborderID)
	if err != nil {
		logs.Errorf("failed to get apply recommendation, err: %v, rid: %s", err, kt.Rid)
		return nil, fmt.Errorf("failed to get apply recommendation, err: %v", err)
	}

	return rst, nil
}

// GetApplyModify gets resource apply order modify records
func (s *scheduler) GetApplyModify(kt *kit.Kit, param *types.GetApplyModifyReq) (*types.GetApplyModifyRst, error) {
	filter, err := param.GetFilter()
	if err != nil {
		logs.Errorf("failed to get apply order modify record, for get filter err: %v, rid: %s", err, kt.Rid)
		return nil, err
	}

	rst := &types.GetApplyModifyRst{}
	if param.Page.EnableCount {
		cnt, err := dao.Set().ModifyRecord().CountModifyRecord(kt.Ctx, filter)
		if err != nil {
			logs.Errorf("failed to get apply order modify record count, err: %v, rid: %s", err, kt.Rid)
			return nil, err
		}
		rst.Count = int64(cnt)
		rst.Info = make([]*table.ModifyRecord, 0)
		return rst, nil
	}

	insts, err := dao.Set().ModifyRecord().FindManyModifyRecord(kt.Ctx, param.Page, filter)
	if err != nil {
		logs.Errorf("failed to get apply order modify record, err: %v, rid: %s", err, kt.Rid)
		return nil, err
	}
	rst.Count = 0
	rst.Info = insts

	return rst, nil
}

// ProcessInitStep processes orders with init step
func (s *scheduler) ProcessInitStep(device *types.DeviceInfo) error {
	_, errMap := s.matcher.ProcessInitStep([]*types.DeviceInfo{device})
	if len(errMap) != 0 {
		return errMap[0]
	}
	return nil
}

// CheckSopsUpdate checks sops task and update status, return err if sops task failed or update failed
func (s *scheduler) CheckSopsUpdate(bkBizID int64, info *types.DeviceInfo, jobUrl string, jobIDStr string) error {
	return s.matcher.CheckSopsUpdate(bkBizID, info, jobUrl, jobIDStr)
}

// RunDiskCheck runs disk check
func (s *scheduler) RunDiskCheck(order *types.ApplyOrder, devices []*types.DeviceInfo) ([]*types.DeviceInfo, error) {
	return s.matcher.RunDiskCheck(order, devices)
}

// DeliverDevices delivers devices to order biz
func (s *scheduler) DeliverDevices(kt *kit.Kit, order *types.ApplyOrder, observeDevices []*types.DeviceInfo) error {
	return s.matcher.DeliverDevices(kt, order, observeDevices)
}

// FinalApplyStep checks whether the record is updated
func (s *scheduler) FinalApplyStep(kt *kit.Kit, genRecord *types.GenerateRecord, order *types.ApplyOrder) error {
	return s.matcher.FinalApplyStep(kt, genRecord, order)
}

// GetGenerateRecords get generate record by order id
func (s *scheduler) GetGenerateRecords(kt *kit.Kit, subOrderId string) ([]*types.GenerateRecord, error) {
	recordInfo, err := s.matcher.GetOrderGenRecords(subOrderId)
	if err != nil {
		logs.Errorf("failed to get generate generateRecord by subOrderId, subOrderId: %s, err: %v, rid: %s", subOrderId,
			err, kt.Rid)
		return nil, err
	}
	return recordInfo, nil
}

// AddCvmDevices check and add cvm device
func (s *scheduler) AddCvmDevices(kt *kit.Kit, taskId string, generateId uint64, order *types.ApplyOrder) error {
	var err error
	switch order.ResourceType {
	// 升降配使用不同的CRP接口轮询单据状态
	case types.ResourceTypeUpgradeCvm:
		err = s.generator.AddUpgradeCvmDevices(kt, taskId, generateId, order)
	default:
		err = s.generator.AddCvmDevices(kt, taskId, generateId, order, order.Spec.Zone)
	}

	if err != nil {
		logs.Errorf("failed to check and update cvm device, orderId: %s, err: %v, rid: %s", err, order.SubOrderId,
			kt.Rid)
		return err
	}
	return nil
}

// UpdateOrderStatus check generate record by order id
func (s *scheduler) UpdateOrderStatus(resType types.ResourceType, suborderID string) error {
	return s.generator.UpdateOrderStatus(resType, suborderID)
}

// GetMatcher get matcher
func (s *scheduler) GetMatcher() *matcher.Matcher {
	return s.matcher
}

// DeliverDevice deliver device
func (s *scheduler) DeliverDevice(info *types.DeviceInfo, order *types.ApplyOrder) error {
	return s.matcher.DeliverDevice(info, order)
}

// UpdateHostOperator update host operator
func (s *scheduler) UpdateHostOperator(info *types.DeviceInfo, hostId int64, operator string) error {
	return s.matcher.UpdateHostOperator(info, hostId, operator)
}

// SetDeviceDelivered set device delivered
func (s *scheduler) SetDeviceDelivered(info *types.DeviceInfo) error {
	return s.matcher.SetDeviceDelivered(info)
}

// CheckRollingServerHost check rolling server host
func (s *scheduler) CheckRollingServerHost(kt *kit.Kit, param *types.CheckRollingServerHostReq) (
	*types.CheckRollingServerHostResp, error) {

	ccReq := &getHostFromCCReq{
		AssetID: param.AssetID,
		BizID:   param.BizID,
	}
	host, err := s.getInheritedHostFromCC(kt, ccReq)
	if err != nil {
		logs.Errorf("get rolling server host from bkcc failed, err: %v, rid: %s", err, kt.Rid)
		return nil, err
	}
	if err = s.checkInheritedHost(kt, param, host); err != nil {
		logs.Errorf("check inherited host failed, err: %v, param: %+v, host: %+v, rid: %s", err, param, host, kt.Rid)
		return nil, err
	}

	chargeMonths := calculateMonths(time.Now(), host.BillingExpireTime)

	// 兜底逻辑，如果当前时间加申请的月份数时间还是小于原来的套餐时间，那么就加上一个月
	if time.Now().AddDate(0, chargeMonths, 0).Before(host.BillingExpireTime) {
		chargeMonths++
	}

	// 校验机型是否匹配
	cvmInfoMap, err := s.configLogics.Device().ListCvmInstanceInfoByDeviceTypes(kt, []string{host.SvrDeviceClassName})
	if err != nil {
		logs.Errorf("get cvm instance info failed, err: %v, rid: %s", err, kt.Rid)
		return nil, err
	}

	if _, ok := cvmInfoMap[host.SvrDeviceClassName]; !ok {
		err = fmt.Errorf("device type not match cvm instance info, host: %+v", host)
		logs.Errorf("check inherited host failed, err: %v, rid: %s", err, kt.Rid)
		return nil, err
	}

	return &types.CheckRollingServerHostResp{
		DeviceType:           host.SvrDeviceClassName,
		DeviceGroup:          cvmInfoMap[host.SvrDeviceClassName].DeviceGroup,
		InstanceChargeType:   host.InstanceChargeType,
		BillingStartTime:     host.BillingStartTime,
		OldBillingExpireTime: host.BillingExpireTime,
		NewBillingExpireTime: time.Now().AddDate(0, chargeMonths, 0),
		ChargeMonths:         chargeMonths,
		CloudInstID:          host.BkCloudInstID,
	}, nil
}

// getHostFromCCReq get host from cc request
type getHostFromCCReq struct {
	AssetID     string
	CloudInstID string
	BizID       int64
}

func (s *scheduler) getInheritedHostFromCC(kt *kit.Kit, param *getHostFromCCReq) (*cmdb.Host, error) {
	rule := querybuilder.CombinedRule{
		Condition: querybuilder.ConditionOr,
		Rules:     []querybuilder.Rule{},
	}

	if len(param.AssetID) > 0 {
		rule.Rules = append(rule.Rules,
			querybuilder.AtomRule{
				Field:    "bk_asset_id",
				Operator: querybuilder.OperatorEqual,
				Value:    param.AssetID,
			})
	}

	if len(param.CloudInstID) > 0 {
		rule.Rules = append(rule.Rules,
			querybuilder.AtomRule{
				Field:    "bk_cloud_inst_id",
				Operator: querybuilder.OperatorEqual,
				Value:    param.CloudInstID,
			})
	}

	if len(rule.Rules) == 0 {
		return nil, errf.New(errf.InvalidParameter, "asset_id and cloud_inst_id can not be empty at the same time")
	}

	fields := []string{"bk_svr_device_cls_name", "instance_charge_type", "billing_start_time", "billing_expire_time",
		"bk_cloud_inst_id", "dept_name"}

	if param.BizID != 0 {
		req := &cmdb.ListBizHostParams{
			BizID:              param.BizID,
			HostPropertyFilter: &cmdb.QueryFilter{Rule: rule},
			Fields:             fields,
			Page:               &cmdb.BasePage{Start: 0, Limit: 1},
		}
		resp, err := s.cc.ListBizHost(kt, req)
		if err != nil {
			logs.Errorf("list host from cc failed, err: %v, req: %+v, rid: %s", err, req, kt.Rid)
			return nil, err
		}

		if len(resp.Info) != 1 {
			logs.Errorf("host is invalid, count: %d, param: %+v, rid: %s", len(resp.Info), param, kt.Rid)
			return nil, errors.New("该主机不在当前业务")
		}

		return &resp.Info[0], nil
	}

	req := &cmdb.ListHostReq{
		HostPropertyFilter: &cmdb.QueryFilter{Rule: rule},
		Fields:             fields,
		Page:               cmdb.BasePage{Start: 0, Limit: 1},
	}
	resp, err := s.cc.ListHost(kt, req)
	if err != nil {
		logs.Errorf("list host from cc failed, err: %v, req: %+v, rid: %s", err, req, kt.Rid)
		return nil, err
	}

	if len(resp.Info) != 1 {
		logs.Errorf("host is invalid, count: %d, param: %+v, rid: %s", len(resp.Info), param, kt.Rid)
		return nil, errors.New("该主机在配置平台(bkcc)不存在")
	}

	return cvt.ValToPtr(resp.Info[0]), nil
}

func (s *scheduler) checkInheritedHost(kt *kit.Kit, param *types.CheckRollingServerHostReq, host *cmdb.Host) error {
	if host == nil {
		return errf.New(errf.InvalidParameter, "host not found")
	}

	if host.DeptName != constant.IEGDeptName {
		return errors.New("主机的所属的运维部门不是IEG")
	}

	if host.InstanceChargeType == "" {
		return errors.New("该主机计费模式为空")
	}

	if host.BillingStartTime.IsZero() {
		return errors.New("该主机无计费开始时间")
	}

	if host.InstanceChargeType == string(cvm.Prepaid) && host.BillingExpireTime.IsZero() {
		return errors.New("该主机无计费结束时间")
	}

	deviceTypeInfoMap, err := s.configLogics.Device().ListDeviceTypeInfoFromCrp(kt, []string{host.SvrDeviceClassName})
	if err != nil {
		logs.Errorf("list device type info from crp failed, err: %v, val: %s, rid: %s", err, host.SvrDeviceClassName,
			kt.Rid)
		return err
	}
	deviceTypeInfo, ok := deviceTypeInfoMap[host.SvrDeviceClassName]
	if !ok {
		return errors.New("该主机未找到对应的机型信息")
	}
	if deviceTypeInfo.InstanceTypeClass != cvmapi.CommonType {
		return errors.New("该主机机型不是通用机型")
	}

	req := &cvmapi.InstanceQueryReq{
		ReqMeta: cvmapi.ReqMeta{
			Id:      cvmapi.CvmId,
			JsonRpc: cvmapi.CvmJsonRpc,
			Method:  cvmapi.CvmInstanceStatusMethod,
		},
		Params: &cvmapi.InstanceQueryParam{
			AssetId: []string{param.AssetID},
		},
	}
	resp, err := s.crpCli.QueryCvmInstances(kt.Ctx, kt.Header(), req)
	if err != nil {
		logs.Errorf("query cvm instance failed, err: %v, req: %+v, rid: %s", err, *req, kt.Rid)
		return err
	}
	if resp.Error.Code != 0 {
		logs.Errorf("failed to query cvm instance, code: %d, msg: %s, asset id: %s, rid: %s", resp.Error.Code,
			resp.Error.Message, param.AssetID, kt.Rid)
		return fmt.Errorf("failed to query cvm instance, code: %d, msg: %s", resp.Error.Code, resp.Error.Message)
	}
	if resp.Result == nil {
		logs.Errorf("failed to query cvm instance, for result is nil, asset id: %s, rid: %s", param.AssetID, kt.Rid)
		return errors.New("failed to query cvm instance, for result is nil")
	}
	if len(resp.Result.Data) != 1 {
		logs.Errorf("failed to query cvm instance, for data num %d != 1, asset id: %s, rid: %s", len(resp.Result.Data),
			param.AssetID, kt.Rid)
		return fmt.Errorf("failed to query cvm instance, for data num %d != 1", len(resp.Result.Data))
	}

	if resp.Result.Data[0].CloudRegion != param.Region {
		return errors.New("继承主机的地域与当前所选不匹配")
	}

	return nil
}

func calculateMonths(startTime, endTime time.Time) int {
	// 计算年份差和月份差
	yearDiff := endTime.Year() - startTime.Year()
	monthDiff := endTime.Month() - startTime.Month()

	// 总月数 = 年份差 * 12 + 月份差
	totalMonths := yearDiff*12 + int(monthDiff)

	// 如果结束时间的日大于开始时间的日，则添加一个月
	if endTime.Day() > startTime.Day() {
		totalMonths++
	}

	return totalMonths
}

// CancelApplyTicketItsm ...
func (s *scheduler) CancelApplyTicketItsm(kt *kit.Kit, req *types.CancelApplyTicketItsmReq) error {
	filter := mapstr.MapStr{
		"order_id": req.OrderID,
	}

	applyTicket, err := model.Operation().ApplyTicket().GetApplyTicket(kt.Ctx, &filter)
	if err != nil {
		logs.Errorf("failed to get apply ticket, err: %v, rid: %s", err, kt.Rid)
		return err
	}

	ticketStatusResp, err := s.itsm.GetTicketStatus(kt, applyTicket.ItsmTicketId)
	if err != nil {
		logs.Errorf("failed to get ticket status, err: %v, rid: %s", err, kt.Rid)
		return err
	}

	// 0. 判断单据状态
	err = checkTicketCanCancel(kt, applyTicket, ticketStatusResp.Data)
	if err != nil {
		return err
	}

	// 1. 关闭 hcm 单据
	applyReq := &types.ApproveApplyReq{
		OrderId:  applyTicket.OrderId,
		Operator: kt.User,
		Approval: false,
		Remark:   fmt.Sprintf("%s手动取消单据", kt.User),
	}
	err = s.ApproveTicket(kt, applyReq)
	if err != nil {
		logs.Errorf("failed to approve ticket, err: %v, rid: %s", err, kt.Rid)
		return err
	}

	// 2. 关闭 itsm 单据
	actionMsg := fmt.Sprintf("%s手动取消单据", kt.User)
	err = s.itsm.TerminateTicket(kt, applyTicket.ItsmTicketId, enumor.ItsmOperatorHcm, actionMsg)
	if err != nil {
		logs.Errorf("failed to cancel itsm ticket, err: %v, rid: %s", err, kt.Rid)
		return err
	}

	return nil
}

// checkTicketCanCancel 检查单据是否可以撤单
func checkTicketCanCancel(kt *kit.Kit, applyTicket *types.ApplyTicket, ticketStatusRst *itsm.GetTicketStatusRst) error {
	// 1. 只有提单人可以撤单
	if applyTicket.User != kt.User {
		return errors.New("only ticket creator can cancel ticket")
	}

	// 2. 单据需要处于运行中或暂停状态
	ticketStatus := itsm.Status(ticketStatusRst.CurrentStatus)
	if ticketStatus != (itsm.StatusRunning) && ticketStatus != (itsm.StatusSuspended) {
		return errors.New("ticket status is not running or suspended")
	}

	// 3. 单据只有处于指定节点时才可以取消
	canCancel := false
	for _, step := range ticketStatusRst.CurrentSteps {
		if checkStepCanCancel(step.Name, applyCancelNodes()) {
			canCancel = true
			break
		}
	}
	if !canCancel {
		return errors.New("ticket steps cannot be cancelled")
	}

	return nil
}

// applyCancelNodes 资源申请服务可以撤单的节点
func applyCancelNodes() []string {
	nodes := make([]string, 0)
	for _, flow := range cc.WoaServer().CancelItsmFlows {
		if flow.ServiceName == enumor.ItsmServiceNameApply {
			for _, node := range flow.StateNodes {
				nodes = append(nodes, node.NodeName)
			}
			break
		}
	}

	return nodes
}

// checkStepCanCancel 检查单据步骤是否可以撤单
func checkStepCanCancel(nodeName string, cancelNodes []string) bool {
	for _, node := range cancelNodes {
		if node == nodeName {
			return true
		}
	}

	return false
}

// CancelApplyTicketCrp ...
func (s *scheduler) CancelApplyTicketCrp(kt *kit.Kit, req *types.CancelApplyTicketCrpReq) error {
	// common filter and page
	filter := map[string]interface{}{
		"suborder_id": req.SubOrderID,
	}
	page := metadata.BasePage{
		Limit: 1,
		Start: 0,
	}

	orders, err := model.Operation().ApplyOrder().FindManyApplyOrder(kt.Ctx, page, filter)
	if err != nil {
		logs.Errorf("failed to get apply order, err: %v, rid: %s", err, kt.Rid)
		return err
	}

	if len(orders) == 0 {
		return errf.New(errf.InvalidParameter, "order is not exist")
	}

	order := orders[0]

	switch order.Status {
	case types.ApplyStatusMatching, types.ApplyStatusMatchedSome, types.ApplyStatusGracefulTerminate:
		break
	default:
		return errf.New(errf.InvalidParameter, fmt.Sprintf("order status is %s cannot cancel", order.Status))
	}

	generateRecords, err := model.Operation().GenerateRecord().FindManyGenerateRecord(kt.Ctx, page, filter)
	if err != nil {
		return err
	}

	// 检查是否有单据尚未发起 crp 请求
	for _, generateRecord := range generateRecords {
		if generateRecord.TaskId == "" {
			return fmt.Errorf("has task still in init,can't revoke suborder, generate id: %d, order id: %s",
				generateRecord.GenerateId, order.SubOrderId)
		}
	}

	// 获取所有未完成的task
	unFinishedTasks := make([]string, 0)
	for _, generateRecord := range generateRecords {
		if taskIsUnFinish(generateRecord) {
			unFinishedTasks = append(unFinishedTasks, generateRecord.TaskId)
		}
	}

	// 筛选可以撤单的crp task
	canRevokeTasks := s.filterCanRevokeCrpTask(kt, unFinishedTasks)
	if len(canRevokeTasks) == 0 {
		return errors.New("no task can revoke")
	}

	// 开始执行撤单程序
	if err = s.revokeApplyOrder(kt, order.SubOrderId, canRevokeTasks); err != nil {
		return err
	}

	return nil
}

// taskIsUnFinish check crp task is finish
func taskIsUnFinish(generateRecord *types.GenerateRecord) bool {
	if generateRecord.Status == types.GenerateStatusHandling {
		return true
	}

	return false
}

// filterCanRevokeCrpTask 筛选可以撤单的 crp 单据
func (s *scheduler) filterCanRevokeCrpTask(kt *kit.Kit, taskIDs []string) []string {
	params := &cvmapi.OrderQueryParam{
		OrderId: taskIDs,
		Status:  make([]int, 0, len(enumor.CrpOrderStatusCanRevoke)),
	}
	req := cvmapi.NewOrderQueryReq(params)
	for _, status := range enumor.CrpOrderStatusCanRevoke {
		req.Params.Status = append(req.Params.Status, int(status))
	}

	ordersResp, err := s.crpCli.QueryCvmOrders(kt.Ctx, kt.Header(), req)
	if err != nil {
		logs.Errorf("failed to query cvm orders, err: %v, rid: %s", err, kt.Rid)
		return nil
	}

	if len(ordersResp.Result.Data) == 0 {
		return nil
	}

	taskIDs = make([]string, 0, len(ordersResp.Result.Data))
	for _, order := range ordersResp.Result.Data {
		taskIDs = append(taskIDs, order.OrderId)
	}

	return taskIDs
}

// revokeApplyOrder revoke apply order
func (s *scheduler) revokeApplyOrder(kt *kit.Kit, subOrderId string, taskIDs []string) error {
	// 1. 修改 apply order 状态为 GracefulTerminate
	filter := &mapstr.MapStr{
		"suborder_id": subOrderId,
	}
	doc := &mapstr.MapStr{
		"status": types.ApplyStatusGracefulTerminate,
	}
	err := model.Operation().ApplyOrder().UpdateApplyOrder(kt.Ctx, filter, doc)
	if err != nil {
		return err
	}

	// 2. 发起 CRP 撤单
	// CRP 单据撤销失败，只记录日志
	errs := make([]string, 0)
	for _, taskID := range taskIDs {
		params := &cvmapi.RevokeCvmOrderParams{
			OrderId: taskID,
		}
		req := cvmapi.NewRevokeCvmOrderReq(params)
		resp, err := s.crpCli.RevokeCvmOrder(kt.Ctx, kt.Header(), req)
		if err != nil {
			errs = append(errs, fmt.Sprintf("taskID: %s, err: %v", taskID, err))
			logs.Warnf("failed to revoke cvm order, taskID: %s, err: %v, rid: %s", taskID, err, kt.Rid)
			continue
		}

		if resp.RespMeta.Error.Code != 0 {
			err = fmt.Errorf("failed to revoke cvm order, trace id: %s, code: %d, msg: %s",
				resp.TraceId, resp.RespMeta.Error.Code, resp.RespMeta.Error.Message)
			logs.Warnf("failed to revoke cvm order, err: %v, rid: %s", err, kt.Rid)
			continue
		}
	}

	if len(errs) > 0 {
		return errors.New(strings.Join(errs, "\n"))
	}

	return nil
}

// calProductDeviceTypeCountMap calculate device type count map
func calProductDeviceTypeCountMap(devices []*types.DeviceInfo, checkDelivered bool) (
	map[string]int, map[types.DeliveredCVMKey]int) {

	deviceTypeCountMap := make(map[string]int)
	deliverGroupCntMap := make(map[types.DeliveredCVMKey]int)

	for _, device := range devices {
		if checkDelivered && !device.IsDelivered {
			continue
		}

		if _, ok := deviceTypeCountMap[device.DeviceType]; !ok {
			deviceTypeCountMap[device.DeviceType] = 0
		}
		deviceTypeCountMap[device.DeviceType]++

		deliveredKey := types.DeliveredCVMKey{
			DeviceType: device.DeviceType,
			Region:     device.CloudRegion,
			Zone:       device.CloudZone,
		}

		if _, ok := deliverGroupCntMap[deliveredKey]; !ok {
			deliverGroupCntMap[deliveredKey] = 0
		}
		deliverGroupCntMap[deliveredKey]++
	}
	return deviceTypeCountMap, deliverGroupCntMap
}

// ConfirmApplyModify modify resource apply modify 用户确认变更记录点接口
func (s *scheduler) ConfirmApplyModify(kt *kit.Kit, param *types.ConfirmApplyModifyReq) (
	*types.ConfirmApplyModifyResp, error) {

	// 查询主机申请单信息
	filter := &mapstr.MapStr{
		"suborder_id": param.SuborderID,
	}
	order, err := model.Operation().ApplyOrder().GetApplyOrder(kt.Ctx, filter)
	if err != nil {
		logs.Errorf("failed to get apply order, err: %v, param: %+v, rid: %s", err, cvt.PtrToVal(param), kt.Rid)
		return nil, err
	}

	// 验证确认信息
	if err = validateConfirm(kt, param, order); err != nil {
		return nil, err
	}

	// 查询主机申请单变更记录
	modifyRecord, isExpire, err := s.checkAndGetModifyRecord(kt, param, order)
	if err != nil {
		// 该变更单已过期，需要把主机申请单，由“待确认”更新为“备货异常”
		if isExpire && order.Stage == types.TicketStageConfirming {
			if expireErr := s.updateApplyOrderStatus(kt, order,
				types.TicketStageSuspend, types.ApplyStatusTerminate); expireErr != nil {
				logs.Errorf("failed to update apply order status SUSPEND, subOrderID: %s, err: %v, param: %+v, rid: %s",
					order.SubOrderId, expireErr, cvt.PtrToVal(param), kt.Rid)
				return nil, expireErr
			}
		}
		return nil, err
	}

	// 获取实际生产成功的数量
	deviceInfos, err := s.matcher.GetUnreleasedDevice(order.SubOrderId)
	if err != nil {
		logs.Errorf("failed to get generate records, subOrderID: %s, err: %v, rid: %s", order.SubOrderId, err, kt.Rid)
		return nil, err
	}
	productSuccCount := uint(len(deviceInfos))

	if productSuccCount > modifyRecord.Details.CurData.TotalNum {
		return nil, fmt.Errorf("cannot confirm apply modify, subOrderID: %s, product success count %d > "+
			"modify record count %d", order.SubOrderId, productSuccCount, modifyRecord.Details.CurData.TotalNum)
	}

	return s.auditApplyModifyCallback(kt, param, order, modifyRecord, productSuccCount)
}

func (s *scheduler) auditApplyModifyCallback(kt *kit.Kit, param *types.ConfirmApplyModifyReq, order *types.ApplyOrder,
	modifyRecord *table.ModifyRecord, productSuccCount uint) (*types.ConfirmApplyModifyResp, error) {

	// 审批状态
	auditStatus := enumor.ApprovalRejectedCvmModifyStatus
	if param.Action == bkbotapproval.ApproveActionType {
		auditStatus = enumor.ApprovedCvmModifyStatus
	}

	// update apply order modify record
	err := s.updateModifyRecordData(kt, order.SubOrderId, modifyRecord.ID, auditStatus)
	if err != nil {
		logs.Errorf("failed to update apply order modify record, subOrderID: %s, modifyID: %d, err: %v, param: %+v, "+
			"rid: %s", order.SubOrderId, modifyRecord.ID, err, cvt.PtrToVal(param), kt.Rid)
		return nil, err
	}

	// 审批不通过的话，直接返回
	if auditStatus != enumor.ApprovedCvmModifyStatus {
		// 需要把主机申请单，由“待确认”更新为“备货异常”
		if order.Stage == types.TicketStageConfirming {
			if err = s.updateApplyOrderStatus(kt, order,
				types.TicketStageSuspend, types.ApplyStatusTerminate); err != nil {
				logs.Errorf("failed to update apply order status SUSPEND, subOrderID: %s, err: %v, param: %+v, rid: %s",
					order.SubOrderId, err, cvt.PtrToVal(param), kt.Rid)
				return nil, err
			}
		}
		return &types.ConfirmApplyModifyResp{
			ResponseMsg:   "单据拒绝成功",
			ResponseColor: bkbotapproval.RedButtonColor,
			RequestID:     kt.Rid,
		}, nil
	}

	// modify apply order
	maReq := &types.ModifyApplyReq{
		TotalNum:   modifyRecord.Details.CurData.TotalNum,
		ProductNum: productSuccCount,
		Spec: &types.ResourceSpec{
			Region:      modifyRecord.Details.CurData.Region,
			Zone:        modifyRecord.Details.CurData.Zone,
			DeviceType:  modifyRecord.Details.CurData.DeviceType,
			ImageId:     modifyRecord.Details.CurData.ImageId,
			DiskSize:    modifyRecord.Details.CurData.DiskSize,
			DiskType:    modifyRecord.Details.CurData.DiskType,
			NetworkType: modifyRecord.Details.CurData.NetworkType,
			Vpc:         modifyRecord.Details.CurData.Vpc,
			Subnet:      modifyRecord.Details.CurData.Subnet,
			Zones:       modifyRecord.Details.CurData.Zones,
			ResAssign:   modifyRecord.Details.CurData.ResAssign,
		},
	}
	if err = s.modifyOrder(kt, order, maReq); err != nil {
		logs.Errorf("failed to modify apply order confirmed, subOrderID: %s, err: %v, rid: %s",
			order.SubOrderId, err, kt.Rid)
		return nil, fmt.Errorf("failed to modify apply order confirming, err: %v", err)
	}

	return &types.ConfirmApplyModifyResp{
		ResponseMsg:   "单据确认成功",
		ResponseColor: bkbotapproval.GreenButtonColor,
		RequestID:     kt.Rid,
	}, nil
}

func (s *scheduler) checkAndGetModifyRecord(kt *kit.Kit, param *types.ConfirmApplyModifyReq, order *types.ApplyOrder) (
	*table.ModifyRecord, bool, error) {

	mrReq := &types.GetApplyModifyReq{ID: []uint64{param.ModifyID}}
	listModify, err := s.GetApplyModify(kt, mrReq)
	if err != nil {
		logs.Errorf("failed to get apply modify record, err: %v, param: %+v, rid: %s", err, cvt.PtrToVal(param), kt.Rid)
		return nil, false, err
	}

	if len(listModify.Info) != 1 {
		return nil, false, fmt.Errorf("cannot confirm apply modify, subOrderID: %s, modifyID: %d, "+
			"modify record count %d != 1", order.SubOrderId, param.ModifyID, len(listModify.Info))
	}

	modifyRecord := listModify.Info[0]
	// 变更状态不是“待确认”，则报错提示
	if modifyRecord.Status != enumor.ApprovalPendingCvmModifyStatus {
		auditStatusName := enumor.CvmModifyRecordStatusMap[modifyRecord.Status]
		logs.Errorf("cannot confirm apply modify, subOrderID: %s, modifyID: %d, action: %s, 变更单不能重复操作，"+
			"当前状态是: %s, rid: %s", order.SubOrderId, param.ModifyID, param.Action, auditStatusName, kt.Rid)
		return nil, false, fmt.Errorf("cannot confirm apply modify, subOrderID: %s, modifyID: %d, action: %s, "+
			"变更单不能重复操作，当前状态是: %s", order.SubOrderId, param.ModifyID, param.Action, auditStatusName)
	}

	// 计算6小时之前的时间
	sixHoursAgo := time.Now().Add(-6 * time.Hour)
	// 判断modifyRecord.CreateAt是否早于6小时之前的时间
	if modifyRecord.CreateAt.Before(sixHoursAgo) {
		// 如果审批超过6小时，需要自动释放
		logs.Errorf("cannot confirm apply modify, subOrderID: %s, modifyID: %d, action: %s, 该变更单已超过6小时，"+
			"已自动释放本次修改, rid: %s", order.SubOrderId, param.ModifyID, param.Action, kt.Rid)
		return nil, true, fmt.Errorf("cannot confirm apply modify, subOrderID: %s, modifyID: %d, action: %s, "+
			"该变更单已超过6小时，已自动释放本次修改", order.SubOrderId, param.ModifyID, param.Action)
	}

	if modifyRecord.Details == nil || modifyRecord.Details.CurData == nil {
		logs.Errorf("cannot confirm apply modify, subOrderID: %s, modifyID: %d, action: %s, modify record details "+
			"is nil, rid: %s", order.SubOrderId, param.ModifyID, param.Action, kt.Rid)
		return nil, false, fmt.Errorf("cannot confirm apply modify, subOrderID: %s, modifyID: %d, action: %s, modify record "+
			"details is nil", order.SubOrderId, param.ModifyID, param.Action)
	}

	return modifyRecord, false, nil
}

func validateConfirm(kt *kit.Kit, param *types.ConfirmApplyModifyReq, order *types.ApplyOrder) error {
	// 校验主机申请单是否属于该业务
	if order.BkBizId != param.BkBizID {
		logs.Errorf("cannot confirm apply modify %s, for its bizid %d != %d, rid: %s", order.SubOrderId,
			order.BkBizId, param.BkBizID, kt.Rid)
		return fmt.Errorf("cannot confirm apply modify %s, for its bizid %d != %d", order.SubOrderId,
			order.BkBizId, param.BkBizID)
	}

	// 校验当前用户是否是主机申请人
	if kt.User != order.User {
		logs.Errorf("cannot confirm apply modify %s, apply order not belonging to the current user: %s, rid: %s",
			order.SubOrderId, kt.User, kt.Rid)
		return fmt.Errorf("cannot confirm apply modify %s, apply order not belonging to the current user: %s",
			order.SubOrderId, kt.User)
	}

	// cannot confirm apply modify if its stage is not SUSPEND、CONFIRMING
	if order.Stage != types.TicketStageSuspend && order.Stage != types.TicketStageConfirming {
		logs.Errorf("cannot confirm apply modify, subOrderID: %s, modifyID: %d, action: %s, stage: %s, param: %+v, "+
			"单据处于备货中或已完成，确认失败, rid: %s", order.SubOrderId, param.ModifyID, param.Action, order.Stage,
			cvt.PtrToVal(param), kt.Rid)
		return fmt.Errorf("cannot confirm apply modify, subOrderID: %s, modifyID: %d, action: %s, stage: %s, "+
			"单据处于备货中或已完成，确认失败", order.SubOrderId, param.ModifyID, param.Action, order.Stage)
	}

	// 暂不支持重试升降配类型单据
	if order.ResourceType == types.ResourceTypeUpgradeCvm {
		logs.Errorf("cannot confirm apply modify %s, for its resource type is %s, rid: %s", order.SubOrderId,
			order.ResourceType, kt.Rid)
		return fmt.Errorf("CVM升降配暂不支持修改单据重试")
	}
	return nil
}

// GetAffinityMatchDetail 亲和性检查
func (s *scheduler) GetAffinityMatchDetail(kit *kit.Kit, param *types.AffinityMatchReq) (*types.AffinityMatchResp,
	error) {
	affinityService, err := NewAffinityService(s.configLogics, s.crpCli)
	if err != nil {
		logs.Errorf("failed to create affinity service, err: %v, rid: %s", err, kit.Rid)
		return nil, err
	}
	return affinityService.GetAffinityMatchDetail(kit, param)
}

// UpdateApplyTicketDemand update apply ticket demand
func (s *scheduler) UpdateApplyTicketDemand(kt *kit.Kit, param *types.ApplyTicket) error {
	if err := s.validateUpdateApplyTicketDemand(kt, param); err != nil {
		logs.Errorf("validate update apply ticket demand failed, err: %v, ticket id: %d, rid: %s", err, param.OrderId,
			kt.Rid)
		return err
	}

	filter := mapstr.MapStr{
		"order_id": param.OrderId,
	}
	update := mapstr.MapStr{
		"suborders":     param.Suborders,
		"old_suborders": param.OldSuborders,
		"update_at":     time.Now(),
	}
	if err := model.Operation().ApplyTicket().UpdateApplyTicket(kt.Ctx, &filter, &update); err != nil {
		logs.Errorf("failed to update apply ticket, err: %v, ticket id: %d, rid: %s", err, param.OrderId, kt.Rid)
		return err
	}

	return nil
}

func (s *scheduler) validateUpdateApplyTicketDemand(kt *kit.Kit, param *types.ApplyTicket) error {
	if param.Stage != types.TicketStageAudit {
		logs.Errorf("failed to update apply ticket demand, for invalid stage: %s != %s, ticket id: %d, rid: %s",
			param.Stage, types.TicketStageAudit, param.OrderId, kt.Rid)
		return fmt.Errorf("invalid ticket stage:%s != %s", param.Stage, types.TicketStageAudit)
	}
	if len(param.OldSuborders) == 0 {
		logs.Errorf("old suborders length is 0, ticket id: %d, rid: %s", param.OrderId, kt.Rid)
		return fmt.Errorf("old suborders length is 0")
	}
	if len(param.Suborders) == 0 {
		logs.Errorf("suborders length is 0, ticket id: %d, rid: %s", param.OrderId, kt.Rid)
		return fmt.Errorf("suborders length is 0")
	}

	deviceTypeInfoMap, err := s.getDeviceTypeInfo(kt, param)
	if err != nil {
		logs.Errorf("get device type info failed, err: %v, ticket id: %d, rid: %s", err, param.OrderId, kt.Rid)
		return err
	}

	if err = s.validateTicketBaseInfo(kt, param, deviceTypeInfoMap); err != nil {
		logs.Errorf("validate ticket base info failed, err: %v, ticket id: %d, rid: %s", err, param.OrderId, kt.Rid)
		return err
	}

	if err = s.validateFuzzyMatchDemand(kt, param, deviceTypeInfoMap); err != nil {
		logs.Errorf("validate fuzzy match demand failed, err: %v, ticket id: %d, rid: %s", err, param.OrderId, kt.Rid)
		return err
	}

	return nil
}

func (s *scheduler) validateTicketBaseInfo(kt *kit.Kit, param *types.ApplyTicket,
	deviceTypeInfoMap map[string]configtypes.DeviceTypeCpuItem) error {

	for index, subOrder := range param.Suborders {
		if subOrder.Spec == nil {
			logs.Errorf("suborder spec is nil, suborder index: %d, rid: %s", index, kt.Rid)
			return fmt.Errorf("suborder spec is nil, suborder index: %d", index)
		}

		deviceTypeInfo, ok := deviceTypeInfoMap[subOrder.Spec.DeviceType]
		if !ok {
			logs.Errorf("device type info not found, device type: %s, rid: %s", subOrder.Spec.DeviceType, kt.Rid)
			return fmt.Errorf("device type info not found, device type: %s", subOrder.Spec.DeviceType)
		}

		subOrderCpuTotal := subOrder.Replicas * uint(deviceTypeInfo.CPUAmount)
		if subOrderCpuTotal != subOrder.AppliedCore {
			logs.Errorf("suborder applied core is not equal to suborder cpu total, suborder index: %d, "+
				"applied core: %d, cpu total: %d, rid: %s", index, subOrder.AppliedCore, subOrderCpuTotal, kt.Rid)
			return fmt.Errorf("suborder applied core is not equal to suborder cpu total, suborder index: %d, "+
				"applied core: %d, cpu total: %d", index, subOrder.AppliedCore, subOrderCpuTotal)
		}
	}

	return nil
}

// validateFuzzyMatchDemand 模糊匹配需求校验
func (s *scheduler) validateFuzzyMatchDemand(kt *kit.Kit, param *types.ApplyTicket,
	deviceTypeInfoMap map[string]configtypes.DeviceTypeCpuItem) error {

	curDeviceDemandMap, err := s.buildSubOrderDemandMap(kt, deviceTypeInfoMap, param.Suborders)
	if err != nil {
		logs.Errorf("build suborder demand map failed, err: %v, ticket id: %d, rid: %s", err, param.OrderId, kt.Rid)
		return err
	}

	oldDeviceDemandMap, err := s.buildSubOrderDemandMap(kt, deviceTypeInfoMap, param.OldSuborders)
	if err != nil {
		logs.Errorf("build old suborder demand map failed, err: %v, ticket id: %d, rid: %s", err, param.OrderId, kt.Rid)
		return err
	}

	for deviceInfo, replicas := range curDeviceDemandMap {
		oldReplicas, ok := oldDeviceDemandMap[deviceInfo]
		if !ok {
			logs.Errorf("old device demand map not found, device info: %+v, rid: %s", deviceInfo, kt.Rid)
			return fmt.Errorf("old device demand map not found, device info: %+v", deviceInfo)
		}
		if oldReplicas < replicas {
			logs.Errorf("old device demand is less than cur device demand, device info: %+v, old replicas: %d, "+
				"cur replicas: %d, ticket id: %d, rid: %s", deviceInfo, oldReplicas, replicas, param.OrderId, kt.Rid)
			return fmt.Errorf("old device demand is less than cur device demand, device info: %+v, old replicas: %d"+
				", cur replicas: %d", deviceInfo, oldReplicas, replicas)
		}
	}

	return nil
}

func (s *scheduler) getDeviceTypeInfo(kt *kit.Kit, param *types.ApplyTicket) (map[string]configtypes.DeviceTypeCpuItem,
	error) {

	deviceTypeMap := make(map[string]struct{})
	for index, suborder := range param.Suborders {
		if suborder.Spec == nil {
			logs.Errorf("suborder spec is nil, ticket id: %d, suborder index: %d, rid: %s", param.OrderId, index,
				kt.Rid)
			return nil, fmt.Errorf("suborder spec is nil, ticket id: %d, suborder index: %d", param.OrderId, index)
		}
		deviceTypeMap[suborder.Spec.DeviceType] = struct{}{}
	}
	for index, suborder := range param.OldSuborders {
		if suborder.Spec == nil {
			logs.Errorf("old suborder spec is nil, ticket id: %d, suborder index: %d, rid: %s", param.OrderId, index,
				kt.Rid)
			return nil, fmt.Errorf("old suborder spec is nil, ticket id: %d, suborder index: %d", param.OrderId, index)
		}
		deviceTypeMap[suborder.Spec.DeviceType] = struct{}{}
	}
	deviceTypeInfoMap, err := s.configLogics.Device().ListCvmInstanceInfoByDeviceTypes(kt, maps.Keys(deviceTypeMap))
	if err != nil {
		logs.Errorf("list cvm instance info by device types failed, err: %v, device types: %v, rid: %s", err,
			maps.Keys(deviceTypeMap), kt.Rid)
		return nil, err
	}

	return deviceTypeInfoMap, nil
}

func (s *scheduler) buildSubOrderDemandMap(kt *kit.Kit, deviceTypeInfoMap map[string]configtypes.DeviceTypeCpuItem,
	suborders []*types.Suborder) (map[configtypes.DeviceFuzzyMatchInfo]uint, error) {

	deviceDemandMap := make(map[configtypes.DeviceFuzzyMatchInfo]uint)
	for index, suborder := range suborders {
		if suborder.Spec == nil {
			logs.Errorf("suborder spec is nil, suborder index: %d, rid: %s", index, kt.Rid)
			return nil, fmt.Errorf("suborder spec is nil, suborder index: %d", index)
		}

		deviceTypeInfo, ok := deviceTypeInfoMap[suborder.Spec.DeviceType]
		if !ok {
			logs.Errorf("device type info not found, device type: %s, rid: %s", suborder.Spec.DeviceType, kt.Rid)
			return nil, fmt.Errorf("device type info not found, device type: %s", suborder.Spec.DeviceType)
		}
		deviceFuzzyMatchInfo := configtypes.DeviceFuzzyMatchInfo{
			TechnicalClass: deviceTypeInfo.TechnicalClass,
			CoreType:       deviceTypeInfo.CoreType,
		}
		if _, ok = deviceDemandMap[deviceFuzzyMatchInfo]; !ok {
			deviceDemandMap[deviceFuzzyMatchInfo] = 0
		}

		subOrderCpuTotal := suborder.Replicas * uint(deviceTypeInfo.CPUAmount)
		deviceDemandMap[deviceFuzzyMatchInfo] += subOrderCpuTotal
	}

	return deviceDemandMap, nil
}
