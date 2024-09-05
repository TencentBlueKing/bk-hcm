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

package billsummaryroot

import (
	"net/http"

	"hcm/cmd/account-server/logics/audit"
	"hcm/cmd/account-server/service/capability"
	"hcm/pkg/client"
	"hcm/pkg/iam/auth"
	"hcm/pkg/rest"
)

// InitService initial the main account service
func InitService(c *capability.Capability) {
	svc := &service{
		client:     c.ApiClient,
		authorizer: c.Authorizer,
		audit:      c.Audit,
	}

	h := rest.NewHandler()

	// register handler
	h.Add("ListRootAccountSummary", http.MethodPost, "/bills/root_account_summarys/list", svc.ListRootAccountSummary)
	h.Add("ReaccountRootAccountSummary",
		http.MethodPost, "/bills/root_account_summarys/reaccount", svc.ReaccountRootAccountSummary)
	h.Add("SumRootAccountSummary", http.MethodPost, "/bills/root_account_summarys/sum", svc.SumRootAccountSummary)
	h.Add("ConfirmRootAccountSummary",
		http.MethodPost, "bills/root_account_summarys/confirm", svc.ConfirmRootAccountSummary)
	h.Add("ExportRootAccountSummary", http.MethodPost, "/bills/root_account_summarys/export",
		svc.ExportRootAccountSummary)

	h.Load(c.WebService)
}

type service struct {
	client     *client.ClientSet
	authorizer auth.Authorizer
	audit      audit.Interface
}
