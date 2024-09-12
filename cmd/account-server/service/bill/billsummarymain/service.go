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

package billsummarymain

import (
	"net/http"

	"hcm/cmd/account-server/logics/audit"
	"hcm/cmd/account-server/service/capability"
	"hcm/pkg/client"
	"hcm/pkg/iam/auth"
	"hcm/pkg/rest"
	"hcm/pkg/thirdparty/esb"
)

// InitService initial the main account service
func InitService(c *capability.Capability) {
	svc := &service{
		client:     c.ApiClient,
		authorizer: c.Authorizer,
		audit:      c.Audit,
		esbClient:  c.EsbClient,
	}

	h := rest.NewHandler()

	// register handler
	h.Add("ListMainAccountSummary", http.MethodPost, "/bills/main_account_summarys/list", svc.ListMainAccountSummary)
	h.Add("SumMainAccountSummary", http.MethodPost, "/bills/main_account_summarys/sum", svc.SumMainAccountSummary)
	h.Add("ExportMainAccountSummary", http.MethodPost,
		"/bills/main_account_summarys/export", svc.ExportMainAccountSummary)

	h.Load(c.WebService)
}

type service struct {
	client     *client.ClientSet
	authorizer auth.Authorizer
	audit      audit.Interface
	esbClient  esb.Client
}
