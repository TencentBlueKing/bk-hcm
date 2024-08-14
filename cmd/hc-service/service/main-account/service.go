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

// Package mainaccount Package service defines service.
package mainaccount

import (
	"net/http"

	cloudadaptor "hcm/cmd/hc-service/logics/cloud-adaptor"
	"hcm/cmd/hc-service/service/capability"
	"hcm/pkg/rest"
)

// InitService initial the service
func InitService(cap *capability.Capability) {
	svc := &service{
		ad: cap.CloudAdaptor,
	}

	h := rest.NewHandler()

	// 创建二级账号目前只支持AWS和GCP
	h.Add("AwsCreateMainAccount", http.MethodPost, "/vendors/aws/main_accounts/create", svc.AwsCreateMainAccount)
	h.Add("GcpCreateMainAccount", http.MethodPost, "/vendors/gcp/main_accounts/create", svc.GcpCreateMainAccount)

	h.Load(cap.WebService)

}

type service struct {
	ad *cloudadaptor.CloudAdaptorClient
}
