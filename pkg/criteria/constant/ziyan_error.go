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

package constant

const (
	// RollingServerSyncFailed rolling server sync failed.
	RollingServerSyncFailed WarnSign = "rolling_server_sync_failed"

	// ResPlanTicketWatchFailed res plan ticket watch failed.
	ResPlanTicketWatchFailed WarnSign = "res_plan_ticket_watch_failed"
	// DemandPenaltyBaseGenerateFailed res plan demand penalty base generate failed.
	DemandPenaltyBaseGenerateFailed WarnSign = "demand_penalty_base_generate_failed"
	// DemandPenaltyRatioReportFailed res plan demand penalty ratio report failed.
	DemandPenaltyRatioReportFailed WarnSign = "demand_penalty_ratio_report_failed"
	// ResPlanExpireNotificationPushFailed res plan expire notification push failed.
	ResPlanExpireNotificationPushFailed WarnSign = "res_plan_expire_notification_push_failed"
	// RollingServerReturnNotificationPushFailed rolling server return notification push failed.
	RollingServerReturnNotificationPushFailed WarnSign = "rolling_server_return_notification_push_failed"
	// ResPlanRefreshTransferQuotaFailed res plan refresh transfer quota failed.
	ResPlanRefreshTransferQuotaFailed WarnSign = "res_plan_refresh_transfer_quota_failed"
	// ResPlanNearExpiredDemandTransferFailed 临期预测转移失败.
	ResPlanNearExpiredDemandTransferFailed WarnSign = "res_plan_near_expired_demand_transfer_failed"
	// ResPlanConfirmNotificationFailed 资源预测确认通知推送失败.
	ResPlanConfirmNotificationFailed WarnSign = "res_plan_confirm_notification_failed"

	// CvmResetSystemUpdatePwdFailed cvm reset system update pwd failed.
	CvmResetSystemUpdatePwdFailed WarnSign = "cvm_reset_system_update_pwd_failed"
	// CvmApplyOrderCrpProductTimeout cvm apply order crp product timeout.
	CvmApplyOrderCrpProductTimeout WarnSign = "cvm_apply_order_crp_product_timeout"

	// CvmRecycleFailed 主机回收失败
	CvmRecycleFailed WarnSign = "cvm_recycle_failed"
	// CvmRecycleStuck 主机回收单据长时间未更新
	CvmRecycleStuck WarnSign = "cvm_recycle_stuck"

	// WaitAndCheckBPaasFailed 等待并检查BPaas失败
	WaitAndCheckBPaasFailed WarnSign = "WaitAndCheckBPaasFailed"
)
