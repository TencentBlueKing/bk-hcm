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

// Package rollingserver 滚服相关接口
package rollingserver

import (
	"net/http"

	rslogic "hcm/cmd/woa-server/logics/rolling-server"
	"hcm/cmd/woa-server/service/capability"
	"hcm/pkg/client"
	"hcm/pkg/criteria/enumor"
	"hcm/pkg/cron/core"
	"hcm/pkg/iam/auth"
	"hcm/pkg/rest"
)

// InitService initial the service
func InitService(c *capability.Capability) {
	s := &service{
		client:             c.Client,
		authorizer:         c.Authorizer,
		rollingServerLogic: c.RsLogic,
		tasks:              c.Tasks,
	}
	h := rest.NewHandler()
	h.Path("/rolling_servers")

	s.initService(h)
	h.Load(c.WebService)

	// 业务下的接口
	bizH := rest.NewHandler()
	bizH.Path("/bizs/{bk_biz_id}/rolling_servers")
	s.bizService(bizH)
	bizH.Load(c.WebService)
}

type service struct {
	authorizer         auth.Authorizer
	client             *client.ClientSet
	rollingServerLogic rslogic.Logics
	tasks              map[enumor.CronTask]core.Task
}

// initService 资源下的接口
func (s *service) initService(h *rest.Handler) {
	// 全局配额
	h.Add("CreateGlobalQuotaConfigs", http.MethodPost, "/global_configs/batch/create", s.CreateGlobalQuotaConfigs)
	h.Add("GetGlobalQuotaConfigs", http.MethodGet, "/global_config", s.GetGlobalQuotaConfigs)
	h.Add("DeleteGlobalQuotaConfig", http.MethodDelete, "/global_config/{id}", s.DeleteGlobalQuotaConfig)

	// 业务滚服配额
	h.Add("AdjustQuotaOffsets", http.MethodPatch, "/quota_offsets/batch", s.AdjustQuotaOffsets)
	h.Add("CreateBizQuotaConfigs", http.MethodPost, "/biz_quotas/batch/create", s.CreateBizQuotaConfigs)
	h.Add("ListBizsWithExistQuota", http.MethodPost, "/exist_quota_bizs/list", s.ListBizsWithExistQuota)
	h.Add("ListBizQuotaConfigs", http.MethodPost, "/biz_quotas/list", s.ListBizQuotaConfigs)
	h.Add("ListQuotaOffsetsAdjustRecords", http.MethodPost, "/quota_offsets/adjust_records/list",
		s.ListQuotaOffsetsAdjustRecords)

	h.Add("ListAppliedRecords", http.MethodPost, "/applied_records/list", s.ListAppliedRecords)
	h.Add("UpdateAppliedRecordsNoticeState", http.MethodPost, "/applied_records/notice/{state}/update",
		s.UpdateAppliedRecordsNoticeState)
	h.Add("UpdateAppliedRecordExemptedCore", http.MethodPatch, "/applied_records/exempted_returned_core",
		s.UpdateAppliedRecordExemptedCore)
	h.Add("ListReturnedRecords", http.MethodPost, "/returned_records/list", s.ListReturnedRecords)
	h.Add("GetCpuCoreSummary", http.MethodPost, "/cpu_core/summary", s.GetCpuCoreSummary)

	h.Add("ListFineDetails", http.MethodPost, "/fine_details/list", s.ListFineDetails)

	h.Add("ListBills", http.MethodPost, "/bills/list", s.ListBills)
	h.Add("SyncBills", http.MethodPost, "/bills/sync", s.SyncBills)
	h.Add("PushRollingServerReturnNotification", http.MethodPost, "/return_notifications/push",
		s.PushReturnNotification)

	h.Add("ManualTriggerCrossMonthTermination", http.MethodPost,
		s.tasks[enumor.CronTaskRollingMonthlyTerminateNotice].GetURL(), s.TerminateLastMonthOrders)
}

// bizService 业务下的接口
func (s *service) bizService(h *rest.Handler) {
	h.Add("ListBizAppliedRecords", http.MethodPost, "/applied_records/list", s.ListBizAppliedRecords)
	h.Add("UpdateBizAppliedRecordsNoticeState", http.MethodPost, "/applied_records/notice/{state}/update",
		s.UpdateBizAppliedRecordsNoticeState)
	h.Add("ListBizReturnedRecords", http.MethodPost, "/returned_records/list", s.ListBizReturnedRecords)
	h.Add("GetBizCpuCoreSummary", http.MethodPost, "/cpu_core/summary", s.GetBizCpuCoreSummary)

	// 业务配额
	h.Add("ListBizBizQuotaConfigs", http.MethodPost, "/biz_quotas/list", s.ListBizBizQuotaConfigs)
}
