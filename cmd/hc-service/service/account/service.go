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

// Package account Package service defines service.
package account

import (
	"net/http"

	"hcm/cmd/hc-service/logics/cloud-adaptor"
	"hcm/cmd/hc-service/service/capability"
	"hcm/pkg/rest"
)

// InitAccountService initial the service
func InitAccountService(cap *capability.Capability) {
	svc := &service{
		ad: cap.CloudAdaptor,
	}

	h := rest.NewHandler()
	// 联通性和云上字段匹配校验
	h.Add("TCloudAccountCheck", http.MethodPost, "/vendors/tcloud/accounts/check", svc.TCloudAccountCheck)
	h.Add("AwsAccountCheck", http.MethodPost, "/vendors/aws/accounts/check", svc.AwsAccountCheck)
	h.Add("HuaWeiAccountCheck", http.MethodPost, "/vendors/huawei/accounts/check", svc.HuaWeiAccountCheck)
	h.Add("GcpAccountCheck", http.MethodPost, "/vendors/gcp/accounts/check", svc.GcpAccountCheck)
	h.Add("AzureAccountCheck", http.MethodPost, "/vendors/azure/accounts/check", svc.AzureAccountCheck)

	// 获取账号配额
	h.Add("GetTCloudAccountZoneQuota", http.MethodPost, "/vendors/tcloud/accounts/zones/quotas",
		svc.GetTCloudAccountZoneQuota)
	h.Add("GetHuaWeiAccountRegionQuota", http.MethodPost, "/vendors/huawei/accounts/regions/quotas",
		svc.GetHuaWeiAccountRegionQuota)
	h.Add("GetGcpAccountRegionQuota", http.MethodPost, "/vendors/gcp/accounts/regions/quotas",
		svc.GetGcpAccountRegionQuota)

	// 通过秘钥获取账号信息
	h.Add("TCloudGetInfoBySecret", http.MethodPost, "/vendors/tcloud/accounts/secret", svc.TCloudGetInfoBySecret)
	h.Add("AwsGetInfoBySecret", http.MethodPost, "/vendors/aws/accounts/secret", svc.AwsGetInfoBySecret)
	h.Add("HuaWeiGetInfoBySecret", http.MethodPost, "/vendors/huawei/accounts/secret", svc.HuaWeiGetInfoBySecret)
	h.Add("GcpGetInfoBySecret", http.MethodPost, "/vendors/gcp/accounts/secret", svc.GcpGetInfoBySecret)
	h.Add("AzureGetInfoBySecret", http.MethodPost, "/vendors/azure/accounts/secret", svc.AzureGetInfoBySecret)

	// 通过秘钥获取资源数量
	h.Add("HuaWeiGetResCountBySecret", http.MethodPost, "/vendors/huawei/accounts/res_counts/by_secrets",
		svc.HuaWeiGetResCountBySecret)
	h.Add("GetGcpResCountBySecret", http.MethodPost, "/vendors/gcp/accounts/res_counts/by_secrets",
		svc.GetGcpResCountBySecret)
	h.Add("GetAzureResCountBySecret", http.MethodPost, "/vendors/azure/accounts/res_counts/by_secrets",
		svc.GetAzureResCountBySecret)
	h.Add("TCloudGetResCountBySecret", http.MethodPost, "/vendors/tcloud/accounts/res_counts/by_secrets",
		svc.TCloudGetResCountBySecret)
	h.Add("AwsGetResCountBySecret", http.MethodPost, "/vendors/aws/accounts/res_counts/by_secrets",
		svc.AwsGetResCountBySecret)

	// 通过密钥获取账号权限策略
	h.Add("ListTCloudAuthPolicies", http.MethodPost, "/vendors/tcloud/accounts/auth_policies/list",
		svc.ListTCloudAuthPolicies)

	// 获取腾讯云账号用户网络类型
	h.Add("GetTCloudNetworkAccountType", http.MethodGet, "/vendors/tcloud/accounts/{account_id}/network_type",
		svc.GetTCloudNetworkAccountType)

	initAccountServiceHooks(svc, h)

	h.Load(cap.WebService)
}

type service struct {
	ad *cloudadaptor.CloudAdaptorClient
}
