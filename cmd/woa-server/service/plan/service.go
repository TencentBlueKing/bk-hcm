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
	"net/http"

	"hcm/cmd/woa-server/logics/biz"
	"hcm/cmd/woa-server/logics/plan"
	"hcm/cmd/woa-server/service/capability"
	"hcm/pkg/client"
	"hcm/pkg/dal/dao"
	"hcm/pkg/iam/auth"
	"hcm/pkg/rest"
)

// InitService initial the plan service.
func InitService(c *capability.Capability) {
	s := &service{
		dao:            c.Dao,
		planController: c.PlanController,
		authorizer:     c.Authorizer,
		bizLogics:      c.BizLogic,
		client:         c.Client,
	}
	h := rest.NewHandler()
	s.initPlanService(h)

	bizH := rest.NewHandler()
	bizH.Path("/bizs/{bk_biz_id}")
	s.initBizPlanService(bizH)

	h.Load(c.WebService)
	bizH.Load(c.WebService)
}

type service struct {
	dao            dao.Set
	planController plan.Logics
	authorizer     auth.Authorizer
	bizLogics      biz.Logics
	client         *client.ClientSet
}

func (s *service) initPlanService(h *rest.Handler) {
	// meta
	// TODO: 这里的url跟meta包里的url边界划分不清晰
	h.Add("ListDemandClass", http.MethodGet, "/plan/demand_class/list", s.ListDemandClass)
	h.Add("ListResMode", http.MethodGet, "/plan/res_mode/list", s.ListResMode)
	h.Add("ListDemandSource", http.MethodGet, "/plan/demand_source/list", s.ListDemandSource)
	h.Add("ListResPlanTicketStatus", http.MethodGet, "/plan/res_plan_ticket_status/list", s.ListRPTicketStatus)
	h.Add("GetDemandAvailableTime", http.MethodPost, "/plans/demands/available_times/get", s.GetDemandAvailableTime)

	// ticket
	h.Add("ListResPlanTicket", http.MethodPost, "/plans/resources/tickets/list", s.ListResPlanTicket)
	h.Add("GetResPlanTicket", http.MethodGet, "/plans/resources/tickets/{id}", s.GetResPlanTicket)
	h.Add("GetResPlanTicketAudit", http.MethodGet,
		"/plans/resources/tickets/{ticket_id}/audit", s.GetResPlanTicketAudit)
	h.Add("ApproveResPlanTicketITSMNode", http.MethodPost,
		"/plans/resources/tickets/{ticket_id}/approve_itsm_node", s.ApproveResPlanTicketITSMNode)
	h.Add("RetryResPlanTicket", http.MethodPost,
		"/plans/resources/tickets/{ticket_id}/retry", s.RetryResPlanTicket)
	h.Add("TerminateResPlanTicket", http.MethodPost,
		"/plans/resources/tickets/{ticket_id}/terminate", s.TerminateResPlanTicket)
	// ticket audit
	h.Add("ListResPlanItsmTicket", http.MethodPost,
		"/plans/resources/itsm/ticket/list", s.ListResPlanItsmTicket)
	h.Add("ListResPlanCrpTicket", http.MethodPost,
		"/plans/resources/crp/ticket/list", s.ListResPlanCrpTicket)

	// sub_ticket
	h.Add("ListResPlanSubTicket", http.MethodPost, "/plans/resources/sub_tickets/list",
		s.ListResPlanSubTicket)
	h.Add("GetResPlanSubTicketDetail", http.MethodGet, "/plans/resources/sub_tickets/{sub_ticket_id}",
		s.GetResPlanSubTicketDetail)
	h.Add("GetResPlanSubTicketAudit", http.MethodGet, "/plans/resources/sub_tickets/{sub_ticket_id}/audit",
		s.GetResPlanSubTicketAudit)
	h.Add("ApproveResPlanTicketAdminNode", http.MethodPost,
		"/plans/resources/sub_tickets/{sub_ticket_id}/approve_admin_node", s.ApproveResPlanSubTicketAdminNode)
	h.Add("BatchApproveResPlanSubTicketAdminNodes", http.MethodPost,
		"/plans/resources/sub_tickets/approve_admin_node/batch", s.BatchApproveResPlanSubTicketAdminNodes)

	// demand
	h.Add("ListResPlanDemand", http.MethodPost, "/plans/resources/demands/list", s.ListResPlanDemand)
	h.Add("GetPlanDemandDetail", http.MethodGet, "/plans/demands/{id}", s.GetPlanDemandDetail)
	h.Add("ListPlanDemandChangelog", http.MethodPost, "/plans/demands/change_logs/list", s.ListPlanDemandChangeLog)
	h.Add("BatchUpdateResPlanDemand", http.MethodPatch, "/plans/resources/demands/batch", s.BatchUpdateResPlanDemand)
	h.Add("ConfirmResPlanDemands", http.MethodPost, "/plans/resources/demands/confirm", s.ConfirmResPlanDemands)
	// gpu demand
	h.Add("ListResPlanDemandGpuSubOrder", http.MethodPost,
		"/plans/resources/gpu/demands/suborders/list", s.ListResPlanDemandGpuSubOrder)
	h.Add("BatchUpdateResPlanDemandGpuSubOrder", http.MethodPost,
		"/plans/resources/gpu/demands/suborders/batch", s.BatchUpdateResPlanDemandGpuSubOrder)

	// verify
	h.Add("VerifyResPlanDemandV2", http.MethodPost, "/plans/resources/demands/verify", s.VerifyResPlanDemandV2)
	h.Add("GetCvmChargeTypeDeviceTypeV2", http.MethodPost, "/config/findmany/config/cvm/charge_type/device_type",
		s.GetCvmChargeTypeDeviceTypeV2)

	// repair history data
	h.Add("RepairResPlanDemand", http.MethodPost, "/plans/resources/demands/repair", s.RepairResPlanDemand)
	h.Add("SyncDemandFromCRPOrder", http.MethodPost, "/plans/demands/sync", s.SyncDemandFromCRPOrder)
	// penalty
	h.Add("CalcPenaltyBase", http.MethodPost, "/plans/penalty/base/calc", s.CalcPenaltyBase)
	h.Add("CalcAndPushPenaltyRatio", http.MethodPost, "/plans/penalty/ratio/push", s.CalcAndPushPenaltyRatio)
	h.Add("PushExpireNotification", http.MethodPost, "/plans/demands/expire_notifications/push",
		s.PushExpireNotification)

	// res plan confirm notice
	h.Add("PushResPlanConfirmNotice", http.MethodPost,
		"/plans/resources/demands/confirm_notifications/push", s.PushResPlanConfirmNotice)

	// demand week
	h.Add("ImportDemandWeek", http.MethodPost, "/plans/demand_week/import", s.ImportDemandWeek)

	// woa device type physical rel
	h.Add("CreateDeviceTypePhysicalRel", http.MethodPost, "/plans/device_type_physical_rels/batch/create",
		s.CreateDeviceTypePhysicalRel)

	// resource plan transfer quota
	h.Add("GetTransferQuotaConfigs", http.MethodGet,
		"/plans/resources/transfer_quotas/configs", s.GetTransferQuotaConfigs)
	h.Add("ListResPlanTransferQuotaSummary", http.MethodPost,
		"/plans/resources/transfer_quotas/summary", s.ListResPlanTransferQuotaSummary)
	// resource plan transfer applied record
	h.Add("ListResPlanTransferAppliedRecord", http.MethodPost,
		"/plans/resources/transfer_applied_records/list", s.ListResPlanTransferAppliedRecord)

	// gpu demand order
	h.Add("ListResPlanDemandGpuOrder", http.MethodPost,
		"/plans/resources/gpu/demands/orders/list", s.ListResPlanDemandGpuOrder)
	h.Add("BatchSetResPlanDemandGpuOrderPending", http.MethodPost,
		"/plans/resources/gpu/demands/orders/batch/pending", s.BatchSetResPlanDemandGpuOrderPending)
	h.Add("BatchRejectResPlanDemandGpuOrder", http.MethodPost,
		"/plans/resources/gpu/demands/orders/batch/reject", s.BatchRejectResPlanDemandGpuOrder)
	h.Add("BatchTerminateResPlanDemandGpuOrder", http.MethodPost,
		"/plans/resources/gpu/demands/orders/batch/terminate", s.BatchTerminateResPlanDemandGpuOrder)
}

// initBizService 初始化业务下接口
func (s *service) initBizPlanService(h *rest.Handler) {
	// biz
	h.Add("GetBizOrgRel", http.MethodGet, "/org/relation", s.GetBizOrgRel)

	// ticket
	h.Add("ListBizResPlanTicket", http.MethodPost, "/plans/resources/tickets/list",
		s.ListBizResPlanTicket)
	h.Add("CreateBizResPlanTicket", http.MethodPost, "/plans/resources/tickets/create",
		s.CreateBizResPlanTicket)
	h.Add("GetBizResPlanTicket", http.MethodGet, "/plans/resources/tickets/{id}", s.GetBizResPlanTicket)
	h.Add("GetBizResPlanTicketAudit", http.MethodGet, "/plans/resources/tickets/{ticket_id}/audit",
		s.GetBizResPlanTicketAudit)
	h.Add("ApproveBizResPlanTicketITSMNode", http.MethodPost,
		"/plans/resources/tickets/{ticket_id}/approve_itsm_node", s.ApproveBizResPlanTicketITSMNode)
	h.Add("RetryBizResPlanTicket", http.MethodPost, "/plans/resources/tickets/{ticket_id}/retry",
		s.RetryBizResPlanTicket)
	h.Add("TerminateBizResPlanTicket", http.MethodPost,
		"/plans/resources/tickets/{ticket_id}/terminate", s.TerminateBizResPlanTicket)

	// sub_ticket
	h.Add("ListBizResPlanSubTicket", http.MethodPost, "/plans/resources/sub_tickets/list",
		s.ListBizResPlanSubTicket)
	h.Add("GetBizResPlanSubTicketDetail", http.MethodGet, "/plans/resources/sub_tickets/{sub_ticket_id}",
		s.GetBizResPlanSubTicketDetail)
	h.Add("GetBizResPlanSubTicketAudit", http.MethodGet,
		"/plans/resources/sub_tickets/{sub_ticket_id}/audit", s.GetBizResPlanSubTicketAudit)
	h.Add("ApproveBizResPlanTicketAdminNode", http.MethodPost,
		"/plans/resources/sub_tickets/{sub_ticket_id}/approve_admin_node", s.ApproveBizResPlanSubTicketAdminNode)
	h.Add("BatchApproveBizResPlanSubTicketAdminNodes", http.MethodPost,
		"/plans/resources/sub_tickets/approve_admin_node/batch", s.BatchApproveBizResPlanSubTicketAdminNodes)

	// demand
	h.Add("ListBizResPlanDemand", http.MethodPost, "/plans/resources/demands/list", s.ListBizResPlanDemand)
	h.Add("GetBizPlanDemandDetail", http.MethodGet, "/plans/demands/{id}", s.GetBizPlanDemandDetail)
	h.Add("ListBizPlanDemandChangeLog", http.MethodPost, "/plans/demands/change_logs/list",
		s.ListBizPlanDemandChangeLog)
	h.Add("AdjustBizResPlanDemand", http.MethodPost, "/plans/resources/demands/adjust",
		s.AdjustBizResPlanDemand)
	h.Add("CancelBizResPlanDemand", http.MethodPost, "/plans/resources/demands/cancel",
		s.CancelBizResPlanDemand)
	h.Add("AutoTransferBizResPlanDemand", http.MethodPost, "/plans/resources/demands/auto_transfer",
		s.AutoTransferBizResPlanDemand)
	h.Add("ConfirmBizResPlanDemands", http.MethodPost, "/plans/resources/demands/confirm",
		s.ConfirmBizResPlanDemands)
	// gpu demand
	h.Add("ListBizResPlanDemandGpuSubOrder", http.MethodPost,
		"/plans/resources/gpu/demands/suborders/list", s.ListBizResPlanDemandGpuSubOrder)
	h.Add("BatchUpdateBizResPlanDemandGpuSubOrder", http.MethodPost,
		"/plans/resources/gpu/demands/suborders/batch", s.BatchUpdateBizResPlanDemandGpuSubOrder)
	h.Add("BatchTerminateBizResPlanDemandGpuSubOrder", http.MethodPost,
		"/plans/resources/gpu/demands/suborders/batch/terminate", s.BatchTerminateBizResPlanDemandGpuSubOrder)

	// resource plan transfer quota
	h.Add("ListBizResPlanTransferQuotaSummary", http.MethodPost, "/plans/resources/transfer_quotas/summary",
		s.ListBizResPlanTransferQuotaSummary)
	// resource plan transfer applied record
	h.Add("ListBizResPlanTransferAppliedRecord", http.MethodPost,
		"/plans/resources/transfer_applied_records/list", s.ListBizResPlanTransferAppliedRecord)

	// gpu demand order
	h.Add("ListBizResPlanDemandGpuOrder", http.MethodPost,
		"/plans/resources/gpu/demands/orders/list", s.ListBizResPlanDemandGpuOrder)
	h.Add("BatchTerminateBizResPlanDemandGpuOrder", http.MethodPost,
		"/plans/resources/gpu/demands/orders/batch/terminate", s.BatchTerminateBizResPlanDemandGpuOrder)
}
