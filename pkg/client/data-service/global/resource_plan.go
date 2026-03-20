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

package global

import (
	"hcm/pkg/api/core"
	dataproto "hcm/pkg/api/data-service"
	rpproto "hcm/pkg/api/data-service/resource-plan"
	"hcm/pkg/client/common"
	resplan "hcm/pkg/dal/dao/types/resource-plan"
	"hcm/pkg/kit"
	"hcm/pkg/rest"
)

// ResourcePlanClient is data service resource plan api client.
type ResourcePlanClient struct {
	client rest.ClientInterface
}

// NewResourcePlanClient create a new resource plan api client.
func NewResourcePlanClient(client rest.ClientInterface) *ResourcePlanClient {
	return &ResourcePlanClient{
		client: client,
	}
}

// --- resource plan demand ---

// ListResPlanDemand list resource plan demand
func (b *ResourcePlanClient) ListResPlanDemand(kt *kit.Kit, req *rpproto.ResPlanDemandListReq) (
	*rpproto.ResPlanDemandListResult, error) {

	return common.Request[rpproto.ResPlanDemandListReq, rpproto.ResPlanDemandListResult](
		b.client, rest.POST, kt, req, "/res_plans/res_plan_demands/list")
}

// BatchCreateResPlanDemand batch create resource plan demand
func (b *ResourcePlanClient) BatchCreateResPlanDemand(kt *kit.Kit, req *rpproto.ResPlanDemandBatchCreateReq) (
	*core.BatchCreateResult, error) {

	return common.Request[rpproto.ResPlanDemandBatchCreateReq, core.BatchCreateResult](
		b.client, rest.POST, kt, req, "/res_plans/res_plan_demands/batch/create")
}

// BatchUpdateResPlanDemand update resource plan demand
func (b *ResourcePlanClient) BatchUpdateResPlanDemand(kt *kit.Kit, req *rpproto.ResPlanDemandBatchUpdateReq) error {
	return common.RequestNoResp[rpproto.ResPlanDemandBatchUpdateReq](
		b.client, rest.PATCH, kt, req, "/res_plans/res_plan_demands/batch")
}

// DeleteResPlanDemand delete resource plan demand
func (b *ResourcePlanClient) DeleteResPlanDemand(kt *kit.Kit, req *dataproto.BatchDeleteReq) error {
	return common.RequestNoResp[dataproto.BatchDeleteReq](
		b.client, rest.DELETE, kt, req, "/res_plans/res_plan_demands/batch")
}

// LockResPlanDemand lock resource plan demand
func (b *ResourcePlanClient) LockResPlanDemand(kt *kit.Kit, req *rpproto.ResPlanDemandLockOpReq) error {
	return common.RequestNoResp[rpproto.ResPlanDemandLockOpReq](
		b.client, rest.PATCH, kt, req, "/res_plans/res_plan_demands/lock")
}

// UnlockResPlanDemand unlock resource plan demand
func (b *ResourcePlanClient) UnlockResPlanDemand(kt *kit.Kit, req *rpproto.ResPlanDemandLockOpReq) error {
	return common.RequestNoResp[rpproto.ResPlanDemandLockOpReq](
		b.client, rest.PATCH, kt, req, "/res_plans/res_plan_demands/unlock")
}

// BatchUpsertResPlanDemand upsert resource plan demand
func (b *ResourcePlanClient) BatchUpsertResPlanDemand(kt *kit.Kit, req *rpproto.ResPlanDemandBatchUpsertReq) (
	*core.BatchCreateResult, error) {

	return common.Request[rpproto.ResPlanDemandBatchUpsertReq, core.BatchCreateResult](
		b.client, rest.POST, kt, req, "/res_plans/res_plan_demands/batch/upsert")
}

// --- demand penalty base ---

// ListDemandPenaltyBase list resource plan demand
func (b *ResourcePlanClient) ListDemandPenaltyBase(kt *kit.Kit, req *rpproto.DemandPenaltyBaseListReq) (
	*rpproto.DemandPenaltyBaseListResult, error) {

	return common.Request[rpproto.DemandPenaltyBaseListReq, rpproto.DemandPenaltyBaseListResult](
		b.client, rest.POST, kt, req, "/res_plans/demand_penalty_bases/list")
}

// BatchCreateDemandPenaltyBase batch create resource plan demand
func (b *ResourcePlanClient) BatchCreateDemandPenaltyBase(kt *kit.Kit, req *rpproto.DemandPenaltyBaseCreateReq) (
	*core.BatchCreateResult, error) {

	return common.Request[rpproto.DemandPenaltyBaseCreateReq, core.BatchCreateResult](
		b.client, rest.POST, kt, req, "/res_plans/demand_penalty_bases/batch/create")
}

// BatchUpdateDemandPenaltyBase update resource plan demand
func (b *ResourcePlanClient) BatchUpdateDemandPenaltyBase(kt *kit.Kit,
	req *rpproto.DemandPenaltyBaseBatchUpdateReq) error {
	return common.RequestNoResp[rpproto.DemandPenaltyBaseBatchUpdateReq](
		b.client, rest.PATCH, kt, req, "/res_plans/demand_penalty_bases/batch")
}

// DeleteDemandPenaltyBase delete resource plan demand
func (b *ResourcePlanClient) DeleteDemandPenaltyBase(kt *kit.Kit, req *dataproto.BatchDeleteReq) error {
	return common.RequestNoResp[dataproto.BatchDeleteReq](
		b.client, rest.DELETE, kt, req, "/res_plans/demand_penalty_bases/batch")
}

// --- demand changelog ---

// ListDemandChangelog list resource plan demand
func (b *ResourcePlanClient) ListDemandChangelog(kt *kit.Kit, req *rpproto.DemandChangelogListReq) (
	*rpproto.DemandChangelogListResult, error) {

	return common.Request[rpproto.DemandChangelogListReq, rpproto.DemandChangelogListResult](
		b.client, rest.POST, kt, req, "/res_plans/demand_changelogs/list")
}

// BatchCreateDemandChangelog batch create resource plan demand
func (b *ResourcePlanClient) BatchCreateDemandChangelog(kt *kit.Kit, req *rpproto.DemandChangelogCreateReq) (
	*core.BatchCreateResult, error) {

	return common.Request[rpproto.DemandChangelogCreateReq, core.BatchCreateResult](
		b.client, rest.POST, kt, req, "/res_plans/demand_changelogs/batch/create")
}

// BatchUpdateDemandChangelog update resource plan demand
func (b *ResourcePlanClient) BatchUpdateDemandChangelog(kt *kit.Kit,
	req *rpproto.DemandChangelogBatchUpdateReq) error {
	return common.RequestNoResp[rpproto.DemandChangelogBatchUpdateReq](
		b.client, rest.PATCH, kt, req, "/res_plans/demand_changelogs/batch")
}

// DeleteDemandChangelog delete resource plan demand
func (b *ResourcePlanClient) DeleteDemandChangelog(kt *kit.Kit, req *dataproto.BatchDeleteReq) error {
	return common.RequestNoResp[dataproto.BatchDeleteReq](
		b.client, rest.DELETE, kt, req, "/res_plans/demand_changelogs/batch")
}

// --- demand gpu template ---

// ListDemandGpuTemplate list demand gpu template
func (b *ResourcePlanClient) ListDemandGpuTemplate(kt *kit.Kit, req *rpproto.DemandGpuTemplateListReq) (
	*rpproto.DemandGpuTemplateListResult, error) {

	return common.Request[rpproto.DemandGpuTemplateListReq, rpproto.DemandGpuTemplateListResult](
		b.client, rest.POST, kt, req, "/res_plans/demand_gpu_templates/list")
}

// BatchCreateDemandGpuTemplate batch create demand gpu template
func (b *ResourcePlanClient) BatchCreateDemandGpuTemplate(kt *kit.Kit, req *rpproto.DemandGpuTemplateBatchCreateReq) (
	*core.BatchCreateResult, error) {

	return common.Request[rpproto.DemandGpuTemplateBatchCreateReq, core.BatchCreateResult](
		b.client, rest.POST, kt, req, "/res_plans/demand_gpu_templates/batch/create")
}

// BatchUpdateDemandGpuTemplate batch update demand gpu template
func (b *ResourcePlanClient) BatchUpdateDemandGpuTemplate(kt *kit.Kit,
	req *rpproto.DemandGpuTemplateBatchUpdateReq) error {
	return common.RequestNoResp[rpproto.DemandGpuTemplateBatchUpdateReq](
		b.client, rest.PATCH, kt, req, "/res_plans/demand_gpu_templates/batch")
}

// DeleteDemandGpuTemplate delete demand gpu template
func (b *ResourcePlanClient) DeleteDemandGpuTemplate(kt *kit.Kit, req *dataproto.BatchDeleteReq) error {
	return common.RequestNoResp[dataproto.BatchDeleteReq](
		b.client, rest.DELETE, kt, req, "/res_plans/demand_gpu_templates/batch")
}

// --- res plan week ---

// ListResPlanWeek list resource plan week
func (b *ResourcePlanClient) ListResPlanWeek(kt *kit.Kit, req *rpproto.ResPlanWeekListReq) (
	*rpproto.ResPlanWeekListResult, error) {

	return common.Request[rpproto.ResPlanWeekListReq, rpproto.ResPlanWeekListResult](
		b.client, rest.POST, kt, req, "/res_plans/res_plan_weeks/list")
}

// BatchCreateResPlanWeek batch create resource plan week
func (b *ResourcePlanClient) BatchCreateResPlanWeek(kt *kit.Kit, req *rpproto.ResPlanWeekBatchCreateReq) (
	*core.BatchCreateResult, error) {

	return common.Request[rpproto.ResPlanWeekBatchCreateReq, core.BatchCreateResult](
		b.client, rest.POST, kt, req, "/res_plans/res_plan_weeks/batch/create")
}

// BatchUpdateResPlanWeek update resource plan week
func (b *ResourcePlanClient) BatchUpdateResPlanWeek(kt *kit.Kit, req *rpproto.ResPlanWeekBatchUpdateReq) error {
	return common.RequestNoResp[rpproto.ResPlanWeekBatchUpdateReq](
		b.client, rest.PATCH, kt, req, "/res_plans/res_plan_weeks/batch")
}

// DeleteResPlanWeek delete resource plan week
func (b *ResourcePlanClient) DeleteResPlanWeek(kt *kit.Kit, req *dataproto.BatchDeleteReq) error {
	return common.RequestNoResp[dataproto.BatchDeleteReq](
		b.client, rest.DELETE, kt, req, "/res_plans/res_plan_weeks/batch")
}

// --- res plan sub ticket ---

// ListResPlanSubTicket list resource plan sub ticket
func (b *ResourcePlanClient) ListResPlanSubTicket(kt *kit.Kit, req *rpproto.ResPlanSubTicketListReq) (
	*rpproto.ResPlanSubTicketListResult, error) {

	return common.Request[rpproto.ResPlanSubTicketListReq, rpproto.ResPlanSubTicketListResult](
		b.client, rest.POST, kt, req, "/res_plans/res_plan_sub_tickets/list")
}

// BatchCreateResPlanSubTicket batch create resource plan sub ticket
func (b *ResourcePlanClient) BatchCreateResPlanSubTicket(kt *kit.Kit, req *rpproto.ResPlanSubTicketBatchCreateReq) (
	*core.BatchCreateResult, error) {

	return common.Request[rpproto.ResPlanSubTicketBatchCreateReq, core.BatchCreateResult](
		b.client, rest.POST, kt, req, "/res_plans/res_plan_sub_tickets/batch/create")
}

// BatchUpdateResPlanSubTicket update resource plan sub ticket
func (b *ResourcePlanClient) BatchUpdateResPlanSubTicket(kt *kit.Kit,
	req *rpproto.ResPlanSubTicketBatchUpdateReq) error {
	return common.RequestNoResp[rpproto.ResPlanSubTicketBatchUpdateReq](
		b.client, rest.PATCH, kt, req, "/res_plans/res_plan_sub_tickets/batch")
}

// UpdateResPlanSubTicketStatusCAS update resource plan sub ticket status with CAS
func (b *ResourcePlanClient) UpdateResPlanSubTicketStatusCAS(kt *kit.Kit,
	req *rpproto.ResPlanSubTicketStatusUpdateReq) error {
	return common.RequestNoResp[rpproto.ResPlanSubTicketStatusUpdateReq](
		b.client, rest.PATCH, kt, req, "/res_plans/res_plan_sub_tickets/status/cas")
}

// DeleteResPlanSubTicket delete resource plan sub ticket
func (b *ResourcePlanClient) DeleteResPlanSubTicket(kt *kit.Kit, req *dataproto.BatchDeleteReq) error {
	return common.RequestNoResp[dataproto.BatchDeleteReq](
		b.client, rest.DELETE, kt, req, "/res_plans/res_plan_sub_tickets/batch")
}

// --- res plan transfer applied record ---

// ListResPlanTransferAppliedRecord list resource plan transfer applied record
func (b *ResourcePlanClient) ListResPlanTransferAppliedRecord(kt *kit.Kit, req *rpproto.TransferAppliedRecordListReq) (
	*rpproto.ResPlanTransferAppliedRecordListResult, error) {

	return common.Request[rpproto.TransferAppliedRecordListReq, rpproto.ResPlanTransferAppliedRecordListResult](
		b.client, rest.POST, kt, req, "/res_plans/transfer_applied_records/list")
}

// BatchCreateResPlanTransferAppliedRecord batch create resource plan transfer applied record
func (b *ResourcePlanClient) BatchCreateResPlanTransferAppliedRecord(kt *kit.Kit,
	req *rpproto.TransferAppliedRecordBatchCreateReq) (*core.BatchCreateResult, error) {

	return common.Request[rpproto.TransferAppliedRecordBatchCreateReq, core.BatchCreateResult](
		b.client, rest.POST, kt, req, "/res_plans/transfer_applied_records/batch/create")
}

// BatchUpdateResPlanTransferAppliedRecord update resource plan transfer applied record
func (b *ResourcePlanClient) BatchUpdateResPlanTransferAppliedRecord(kt *kit.Kit,
	req *rpproto.TransferAppliedRecordBatchUpdateReq) error {

	return common.RequestNoResp[rpproto.TransferAppliedRecordBatchUpdateReq](
		b.client, rest.PATCH, kt, req, "/res_plans/transfer_applied_records/batch")
}

// DeleteResPlanTransferAppliedRecord delete resource plan transfer applied record
func (b *ResourcePlanClient) DeleteResPlanTransferAppliedRecord(kt *kit.Kit, req *dataproto.BatchDeleteReq) error {
	return common.RequestNoResp[dataproto.BatchDeleteReq](
		b.client, rest.DELETE, kt, req, "/res_plans/transfer_applied_records/batch")
}

// SumResPlanTransferAppliedRecord sum resource plan transfer applied record
func (b *ResourcePlanClient) SumResPlanTransferAppliedRecord(kt *kit.Kit, req *rpproto.TransferAppliedRecordListReq) (
	*resplan.SumTransferAppliedRecord, error) {

	return common.Request[rpproto.TransferAppliedRecordListReq, resplan.SumTransferAppliedRecord](
		b.client, rest.POST, kt, req, "/res_plans/transfer_applied_records/sum")
}

// --- res plan demand gpu order ---

// BatchCreateResPlanDemandGpuOrder batch create res plan demand gpu order
func (b *ResourcePlanClient) BatchCreateResPlanDemandGpuOrder(kt *kit.Kit,
	req *rpproto.ResPlanDemandGpuOrderBatchCreateReq) (*core.BatchCreateResult, error) {

	return common.Request[rpproto.ResPlanDemandGpuOrderBatchCreateReq, core.BatchCreateResult](
		b.client, rest.POST, kt, req, "/res_plans/res_plan_demand_gpu_orders/batch/create")
}

// BatchUpdateResPlanDemandGpuOrder batch update res plan demand gpu order
func (b *ResourcePlanClient) BatchUpdateResPlanDemandGpuOrder(kt *kit.Kit,
	req *rpproto.ResPlanDemandGpuOrderBatchUpdateReq) error {

	return common.RequestNoResp[rpproto.ResPlanDemandGpuOrderBatchUpdateReq](
		b.client, rest.PATCH, kt, req, "/res_plans/res_plan_demand_gpu_orders/batch")
}

// ListResPlanDemandGpuOrder list res plan demand gpu order
func (b *ResourcePlanClient) ListResPlanDemandGpuOrder(kt *kit.Kit,
	req *rpproto.ResPlanDemandGpuOrderListReq) (*rpproto.ResPlanDemandGpuOrderListResult, error) {

	return common.Request[rpproto.ResPlanDemandGpuOrderListReq, rpproto.ResPlanDemandGpuOrderListResult](
		b.client, rest.POST, kt, req, "/res_plans/res_plan_demand_gpu_orders/list")
}

// DeleteResPlanDemandGpuOrder delete res plan demand gpu order
func (b *ResourcePlanClient) DeleteResPlanDemandGpuOrder(kt *kit.Kit, req *dataproto.BatchDeleteReq) error {
	return common.RequestNoResp[dataproto.BatchDeleteReq](
		b.client, rest.DELETE, kt, req, "/res_plans/res_plan_demand_gpu_orders/batch")
}

// --- resource plan demand gpu sub order ---

// ListResPlanDemandGpuSubOrder list resource plan demand gpu sub order
func (b *ResourcePlanClient) ListResPlanDemandGpuSubOrder(kt *kit.Kit,
	req *rpproto.ResPlanDemandGpuSubOrderListReq) (*rpproto.ResPlanDemandGpuSubOrderListResult, error) {

	return common.Request[rpproto.ResPlanDemandGpuSubOrderListReq, rpproto.ResPlanDemandGpuSubOrderListResult](
		b.client, rest.POST, kt, req, "/res_plans/res_plan_demand_gpu_suborders/list")
}

// BatchCreateResPlanDemandGpuSubOrder batch create resource plan demand gpu sub order
func (b *ResourcePlanClient) BatchCreateResPlanDemandGpuSubOrder(kt *kit.Kit,
	req *rpproto.ResPlanDemandGpuSubOrderBatchCreateReq) (*core.BatchCreateResult, error) {

	return common.Request[rpproto.ResPlanDemandGpuSubOrderBatchCreateReq, core.BatchCreateResult](
		b.client, rest.POST, kt, req, "/res_plans/res_plan_demand_gpu_suborders/batch/create")
}

// BatchUpdateResPlanDemandGpuSubOrder update resource plan demand gpu sub order
func (b *ResourcePlanClient) BatchUpdateResPlanDemandGpuSubOrder(kt *kit.Kit,
	req *rpproto.ResPlanDemandGpuSubOrderBatchUpdateReq) error {

	return common.RequestNoResp[rpproto.ResPlanDemandGpuSubOrderBatchUpdateReq](
		b.client, rest.PATCH, kt, req, "/res_plans/res_plan_demand_gpu_suborders/batch")
}

// BatchUpdateStatusResPlanDemandGpuSubOrder batch update sub orders to the same status by id list (max 100).
func (b *ResourcePlanClient) BatchUpdateStatusResPlanDemandGpuSubOrder(kt *kit.Kit,
	req *rpproto.ResPlanDemandGpuSubOrderBatchUpdateStatusReq) error {

	return common.RequestNoResp[rpproto.ResPlanDemandGpuSubOrderBatchUpdateStatusReq](
		b.client, rest.PATCH, kt, req, "/res_plans/res_plan_demand_gpu_suborders/batch/status")
}

// DeleteResPlanDemandGpuSubOrder delete resource plan demand gpu sub order
func (b *ResourcePlanClient) DeleteResPlanDemandGpuSubOrder(kt *kit.Kit, req *dataproto.BatchDeleteReq) error {
	return common.RequestNoResp[dataproto.BatchDeleteReq](
		b.client, rest.DELETE, kt, req, "/res_plans/res_plan_demand_gpu_suborders/batch")
}

// OverwriteResPlanDemandGpuSubOrders atomically replaces all sub orders of an order with new ones.
func (b *ResourcePlanClient) OverwriteResPlanDemandGpuSubOrders(kt *kit.Kit,
	req *rpproto.ResPlanDemandGpuSubOrderOverwriteReq) (*core.BatchCreateResult, error) {

	return common.Request[rpproto.ResPlanDemandGpuSubOrderOverwriteReq, core.BatchCreateResult](
		b.client, rest.POST, kt, req, "/res_plans/res_plan_demand_gpu_suborders/overwrite")
}
