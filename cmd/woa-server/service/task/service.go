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

package task

import (
	"net/http"

	"hcm/cmd/woa-server/logics/config"
	"hcm/cmd/woa-server/logics/dissolve"
	gclogics "hcm/cmd/woa-server/logics/green-channel"
	planLogics "hcm/cmd/woa-server/logics/plan"
	taskLogics "hcm/cmd/woa-server/logics/task"
	"hcm/cmd/woa-server/service/capability"
	"hcm/pkg/client"
	"hcm/pkg/criteria/enumor"
	"hcm/pkg/cron/core"
	"hcm/pkg/iam/auth"
	"hcm/pkg/rest"
	"hcm/pkg/thirdparty/api-gateway/itsm"
)

// InitService initial the service
func InitService(c *capability.Capability) {
	logics := taskLogics.New(c.SchedulerIf, c.RecyclerIf, c.InformerIf, c.OperationIf, c.TaskStatistics)
	s := &service{
		client:         c.Client,
		logics:         logics,
		configLogics:   c.ConfigLogics,
		planLogics:     c.PlanController,
		authorizer:     c.Authorizer,
		itsmClient:     c.ThirdCli.ITSM,
		gcLogics:       c.GcLogic,
		dissolveLogics: c.DissolveLogic,
	}
	h := rest.NewHandler()
	h.Path("/task")

	s.initOperationService(h)
	s.initDeliverAnalysisService(h)
	s.initRecyclerService(h)
	s.initSchedulerService(h)

	h.Load(c.WebService)

	// 业务下的接口
	bizH := rest.NewHandler()
	bizH.Path("/bizs/{bk_biz_id}/task")
	bizService(bizH, s)

	bizH.Load(c.WebService)
}

type service struct {
	client         *client.ClientSet
	logics         taskLogics.Logics
	configLogics   config.Logics
	planLogics     planLogics.Logics
	authorizer     auth.Authorizer
	itsmClient     itsm.Client
	gcLogics       gclogics.Logics
	dissolveLogics dissolve.Logics
	Tasks          map[enumor.CronTask]core.Task
}

func (s *service) initOperationService(h *rest.Handler) {
	h.Add("GetApplyStatistics", http.MethodPost, "/find/operation/apply/statistics", s.GetApplyStatistics)
	h.Add("CreateApplyOrderStatisticsConfig", http.MethodPost, "/config/create/apply/order/statistics",
		s.CreateApplyOrderStatisticsConfig)
	h.Add("UpdateApplyOrderStatisticsConfig", http.MethodPut, "/config/update/apply/order/statistics",
		s.UpdateApplyOrderStatisticsConfig)
	h.Add("ListApplyOrderStatisticsConfig", http.MethodPost, "/config/findmany/apply/order/statistics",
		s.ListApplyOrderStatisticsConfig)
	h.Add("ListApplyOrderStatisticsYearMonths", http.MethodPost, "/config/findmany/apply/order/statistics/year_months",
		s.ListApplyOrderStatisticsYearMonths)
	h.Add("GetCompletionRateStatistics", http.MethodPost,
		"/apply/completion-rate/statistics", s.GetCompletionRateStatistics)
	h.Add("GetCompletionRateDetail", http.MethodPost, "/apply/completion-rate/detail", s.GetCompletionRateDetail)
	h.Add("GetApplyBizHostsStatistics", http.MethodPost,
		"/apply/findmany/bizs_hosts/statistics", s.GetApplyBizHostsStatistics)
	h.Add("GetApplyBizCpuCoresStatistics", http.MethodPost,
		"/apply/findmany/bizs_cpucores/statistics", s.GetApplyBizCpuCoresStatistics)
}

func (s *service) initDeliverAnalysisService(h *rest.Handler) {
	h.Add("GetAverageTimeConsumptionOverview", http.MethodPost, "/apply/analysis/average_time_consumption/overview",
		s.GetAverageTimeConsumptionOverview)
	h.Add("GetAverageTimeConsumptionCompare", http.MethodPost, "/apply/analysis/average_time_consumption/compare",
		s.GetAverageTimeConsumptionCompare)
	h.Add("GetOrderTimeCostOverview", http.MethodPost, "/apply/analysis/order_time_cost/overview",
		s.GetOrderTimeCostOverview)
	h.Add("GetOrderTimeCostCompare", http.MethodPost, "/apply/analysis/order_time_cost/compare",
		s.GetOrderTimeCostCompare)
	h.Add("GetProductionStageTimeCostOverview", http.MethodPost, "/apply/analysis/production_stage_time_cost/overview",
		s.GetProductionStageTimeCostOverview)
	h.Add("GetProductionStageTimeCostCompare", http.MethodPost, "/apply/analysis/production_stage_time_cost/compare",
		s.GetProductionStageTimeCostCompare)
	h.Add("GetPercentileTimeConsumptionOverview", http.MethodPost,
		"/apply/analysis/percentile_time_consumption/overview",
		s.GetPercentileTimeConsumptionOverview)
	h.Add("GetPercentileTimeConsumptionCompare", http.MethodPost, "/apply/analysis/percentile_time_consumption/compare",
		s.GetPercentileTimeConsumptionCompare)
	h.Add("GetDeliveryRateStatistics", http.MethodPost, "/apply/delivery-rate/statistics",
		s.GetDeliveryRateStatistics)
	h.Add("GetDeliveryRateDetail", http.MethodPost, "/apply/delivery-rate/detail",
		s.GetDeliveryRateDetail)
}

func (s *service) initRecyclerService(h *rest.Handler) {
	h.Add("GetRecyclability", http.MethodPost, "/findmany/recycle/recyclability", s.GetRecyclability)
	h.Add("PreviewRecycleOrder", http.MethodPost, "/preview/recycle/order", s.PreviewRecycleOrder)
	h.Add("AuditRecycleOrder", http.MethodPost, "/audit/recycle/order", s.AuditRecycleOrder)
	h.Add("CreateRecycleOrder", http.MethodPost, "/create/recycle/order", s.CreateRecycleOrder)
	h.Add("GetRecycleOrder", http.MethodPost, "/findmany/recycle/order", s.GetRecycleOrder)
	h.Add("GetBizRecycleOrder", http.MethodPost, "/findmany/biz/recycle/order", s.GetBizRecycleOrder)
	h.Add("GetRecycleDetect", http.MethodPost, "/findmany/recycle/detect", s.GetRecycleDetect)
	h.Add("ListDetectHost", http.MethodPost, "/list/recycle/detect/host", s.ListDetectHost)
	h.Add("ListDetectTask", http.MethodPost, "/list/detect/task", s.ListDetectTask)
	h.Add("GetRecycleDetectStep", http.MethodPost, "/findmany/recycle/detect/step", s.GetRecycleDetectStep)
	h.Add("StartRecycleOrder", http.MethodPost, "/start/recycle/order", s.StartRecycleOrder)
	h.Add("StartRecycleOrderByRecycleType", http.MethodPost,
		"/start/recycle/order/by/recycle_type", s.StartRecycleOrderByRecycleType)
	h.Add("StartRecycleDetect", http.MethodPost, "/start/recycle/detect", s.StartRecycleDetect)
	h.Add("ReviseRecycleOrder", http.MethodPost, "/revise/recycle/order", s.ReviseRecycleOrder)
	h.Add("PauseRecycleOrder", http.MethodPost, "/pause/recycle", s.PauseRecycleOrder)
	h.Add("ResumeRecycleOrder", http.MethodPost, "/resume/recycle/order", s.ResumeRecycleOrder)
	h.Add("TerminateRecycleOrder", http.MethodPost, "/terminate/recycle/order", s.TerminateRecycleOrder)
	h.Add("GetRecycleOrderHost", http.MethodPost, "/findmany/recycle/host", s.GetRecycleOrderHost)
	h.Add("GetRecycleRecordDeviceType", http.MethodGet, "/find/recycle/record/devicetype", s.GetRecycleRecordDeviceType)
	h.Add("GetRecycleRecordRegion", http.MethodGet, "/find/recycle/record/region", s.GetRecycleRecordRegion)
	h.Add("GetRecycleRecordZone", http.MethodGet, "/find/recycle/record/zone", s.GetRecycleRecordZone)
	h.Add("GetBizHostToRecycle", http.MethodPost, "/find/recycle/biz/host", s.GetBizHostToRecycle)
	h.Add("StartIdleCheck", http.MethodPost, "/start/cvms/idle_check", s.StartIdleCheck)

	// configs related api
	h.Add("GetRecycleStageCfg", http.MethodGet, "/find/config/recycle/stage", s.GetRecycleStageCfg)
	h.Add("GetRecycleStatusCfg", http.MethodGet, "/find/config/recycle/status", s.GetRecycleStatusCfg)
	h.Add("GetDetectStatusCfg", http.MethodGet, "/find/config/recycle/detect/status", s.GetDetectStatusCfg)
	h.Add("GetDetectStepCfg", http.MethodGet, "/find/config/recycle/detect/step", s.GetDetectStepCfg)
}

func (s *service) initSchedulerService(h *rest.Handler) {
	h.Add("UpdateApplyTicket", http.MethodPost, "/update/apply/ticket", s.UpdateApplyTicket)
	h.Add("GetApplyTicket", http.MethodPost, "/get/apply/ticket", s.GetApplyTicket)
	h.Add("GetApplyAuditItsm", http.MethodPost, "/get/apply/ticket/audit", s.GetApplyAuditItsm)
	h.Add("CancelApplyTicketItsm", http.MethodPost, "/apply/ticket/itsm_audit/cancel", s.CancelApplyTicketItsm)
	h.Add("CancelApplyTicketCrp", http.MethodPost, "/apply/ticket/crp_audit/cancel", s.CancelApplyTicketCrp)
	h.Add("AuditApplyTicket", http.MethodPost, "/audit/apply/ticket", s.AuditApplyTicket)
	h.Add("AutoAuditApplyTicket", http.MethodPost, "/autoaudit/apply/ticket", s.AutoAuditApplyTicket)
	h.Add("ApproveApplyTicket", http.MethodPost, "/approve/apply/ticket", s.ApproveApplyTicket)
	h.Add("CreateApplyOrder", http.MethodPost, "/create/apply", s.CreateApplyOrder)
	h.Add("GetApplyOrder", http.MethodPost, "/findmany/apply", s.GetApplyOrder)
	h.Add("GetBizApplyOrder", http.MethodPost, "/findmany/biz/apply", s.GetBizApplyOrder)
	h.Add("GetApplyStatus", http.MethodGet, "/find/apply/status/{order_id}", s.GetApplyStatus)
	h.Add("GetApplyDetail", http.MethodPost, "/find/apply/detail", s.GetApplyDetail)
	h.Add("GetApplyGenerate", http.MethodPost, "/find/apply/record/generate", s.GetApplyGenerate)
	h.Add("GetApplyInit", http.MethodPost, "/find/apply/record/init", s.GetApplyInit)
	h.Add("GetApplyDiskCheck", http.MethodPost, "/find/apply/record/disk_check", s.GetApplyDiskCheck)
	h.Add("GetApplyDeliver", http.MethodPost, "/find/apply/record/deliver", s.GetApplyDeliver)
	h.Add("GetApplyDevice", http.MethodPost, "/findmany/apply/device", s.GetApplyDevice)
	h.Add("GetDeliverDeviceByOrder", http.MethodPost, "/findmany/apply/deliver/device", s.GetDeliverDeviceByOrder)
	h.Add("ExportDeliverDevice", http.MethodPost, "/export/apply/deliver/device", s.ExportDeliverDevice)
	h.Add("MatchDevice", http.MethodPost, "/findmany/apply/match/device", s.GetMatchDevice)
	h.Add("MatchDevice", http.MethodPost, "/commit/apply/match", s.MatchDevice)
	h.Add("MatchPoolDevice", http.MethodPost, "/commit/apply/pool/match", s.MatchPoolDevice)
	h.Add("PauseApplyOrder", http.MethodPost, "/pause/apply", s.PauseApplyOrder)
	h.Add("ResumeApplyOrder", http.MethodPost, "/resume/apply", s.ResumeApplyOrder)
	h.Add("StartApplyOrder", http.MethodPost, "/start/apply", s.StartApplyOrder)
	h.Add("TerminateApplyOrder", http.MethodPost, "/terminate/apply", s.TerminateApplyOrder)
	h.Add("ModifyApplyOrder", http.MethodPost, "/modify/apply", s.ModifyApplyOrder)
	h.Add("RecommendApplyOrder", http.MethodPost, "/recommend/apply", s.RecommendApplyOrder)
	h.Add("GetApplyModify", http.MethodPost, "/find/apply/record/modify", s.GetApplyModify)

	h.Add("CheckRollingServerHost", http.MethodPost, "/check/rolling_server/host", s.CheckRollingServerHost)
	h.Add("GetApplyAuditCrp", http.MethodPost, "/apply/crp_ticket/audit/get", s.GetApplyAuditCrp)

	h.Add("ListApplyAuditInfo", http.MethodPost, "/apply/ticket/audit/info/list", s.ListApplyAuditInfo)
	h.Add("ApproveApplyTicketNode", http.MethodPost, "/approve/apply/ticket/node", s.ApproveApplyTicketNode)
	h.Add("FindApproveNodeResult", http.MethodPost, "/find/approve_node/result", s.FindApproveNodeResult)

	h.Add("ListHostApplyItsmTicket", http.MethodPost, "/apply/itsm/ticket/list", s.ListHostApplyItsmTicket)
	h.Add("ListHostApplyCrpTicket", http.MethodPost, "/apply/crp/ticket/list", s.ListHostApplyCrpTicket)

	h.Add("UpdateApplyTicketDemand", http.MethodPost, "/apply/ticket/demand/update", s.UpdateApplyTicketDemand)
}

// bizService 业务下的接口
func bizService(h *rest.Handler, s *service) {
	h.Add("CreateBizApplyOrder", http.MethodPost, "/create/apply", s.CreateBizApplyOrder)
	h.Add("UpdateBizApplyTicket", http.MethodPost, "/update/apply/ticket", s.UpdateBizApplyTicket)
	h.Add("StartBizApplyOrder", http.MethodPost, "/start/apply", s.StartBizApplyOrder)
	h.Add("TerminateBizApplyOrder", http.MethodPost, "/terminate/apply", s.TerminateBizApplyOrder)
	h.Add("ModifyBizApplyOrder", http.MethodPost, "/modify/apply", s.ModifyBizApplyOrder)
	h.Add("GetBizApplyOrder", http.MethodPost, "/findmany/apply", s.GetApplyBizOrder)
	h.Add("GetBizApplyTicket", http.MethodPost, "/get/apply/ticket", s.GetBizApplyTicket)
	h.Add("GetBizApplyAuditItsm", http.MethodPost, "/get/apply/ticket/audit", s.GetBizApplyAuditItsm)
	h.Add("GetBizApplyAuditCrp", http.MethodPost, "/apply/crp_ticket/audit/get", s.GetBizApplyAuditCrp)
	h.Add("GetBizApplyDetail", http.MethodPost, "/find/apply/detail", s.GetBizApplyDetail)
	h.Add("GetBizApplyGenerate", http.MethodPost, "/find/apply/record/generate", s.GetBizApplyGenerate)
	h.Add("GetBizApplyDevice", http.MethodPost, "/findmany/apply/device", s.GetBizApplyDevice)
	h.Add("GetBizApplyInit", http.MethodPost, "/find/apply/record/init", s.GetBizApplyInit)
	h.Add("GetBizApplyDeliver", http.MethodPost, "/find/apply/record/deliver", s.GetBizApplyDeliver)
	h.Add("GetBizMatchDevice", http.MethodPost, "/findmany/apply/match/device", s.GetBizMatchDevice)
	h.Add("MatchBizDevice", http.MethodPost, "/commit/apply/match", s.MatchBizDevice)
	h.Add("MatchBizPoolDevice", http.MethodPost, "/commit/apply/pool/match", s.MatchBizPoolDevice)
	h.Add("GetBizApplyModify", http.MethodPost, "/find/apply/record/modify", s.GetBizApplyModify)
	h.Add("ConfirmBizApplyModify", http.MethodPost, "/confirm/apply/record/modify", s.ConfirmBizApplyModify)
	h.Add("CreateBizRecycleOrder", http.MethodPost, "/create/recycle/order", s.CreateBizRecycleOrder)
	h.Add("PreviewBizRecycleOrder", http.MethodPost, "/preview/recycle/order", s.PreviewBizRecycleOrder)
	h.Add("TerminateBizRecycleOrder", http.MethodPost, "/terminate/recycle/order", s.TerminateBizRecycleOrder)
	h.Add("GetBizRecycleOrderHost", http.MethodPost, "/findmany/recycle/host", s.GetBizRecycleOrderHost)
	h.Add("GetRecycleBizOrder", http.MethodPost, "/findmany/recycle/order", s.GetRecycleBizOrder)
	h.Add("GetBizRecyclability", http.MethodPost, "/findmany/recycle/recyclability", s.GetBizRecyclability)
	h.Add("ReviseBizRecycleOrder", http.MethodPost, "/revise/recycle/order", s.ReviseBizRecycleOrder)
	h.Add("StartBizRecycleOrder", http.MethodPost, "/start/recycle/order", s.StartBizRecycleOrder)
	h.Add("GetBizRecycleDetect", http.MethodPost, "/findmany/recycle/detect", s.GetBizRecycleDetect)
	h.Add("ListBizDetectHost", http.MethodPost, "/list/recycle/detect/host", s.ListBizDetectHost)
	h.Add("GetBizRecycleDetectStep", http.MethodPost, "/findmany/recycle/detect/step", s.GetBizRecycleDetectStep)
	h.Add("CancelBizApplyTicketItsm", http.MethodPost, "/apply/ticket/itsm_audit/cancel", s.CancelBizApplyTicketItsm)
	h.Add("CancelBizApplyTicketCrp", http.MethodPost, "/apply/ticket/crp_audit/cancel", s.CancelBizApplyTicketCrp)
	h.Add("AuditBizApplyTicket", http.MethodPost, "/audit/apply/ticket", s.AuditBizApplyTicket)

	h.Add("CheckHostUworkTicketStatus", http.MethodPost, "/hosts/uwork_tickets/status/check",
		s.CheckHostUworkTicketStatus)

	// 升降配接口
	h.Add("CreateBizUpgradeCRPOrder", http.MethodPost, "/create/upgrade/crp_order", s.CreateBizUpgradeCRPOrder)
	// 亲和性检查接口
	h.Add("GetAffinityMatchDetail", http.MethodPost, "/apply/match/check", s.GetAffinityMatchDetail)
}
