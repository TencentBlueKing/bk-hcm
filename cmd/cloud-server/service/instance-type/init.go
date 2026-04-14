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

package instancetype

import (
	"net/http"

	"hcm/cmd/cloud-server/logics/audit"
	"hcm/cmd/cloud-server/service/capability"
	"hcm/pkg/client"
	"hcm/pkg/iam/auth"
	"hcm/pkg/rest"
)

type instanceTypeSvc struct {
	client     *client.ClientSet
	authorizer auth.Authorizer
	audit      audit.Interface
}

// InitInstanceTypeService ...
func InitInstanceTypeService(c *capability.Capability) {
	svc := &instanceTypeSvc{
		client:     c.ApiClient,
		authorizer: c.Authorizer,
		audit:      c.Audit,
	}

	h := rest.NewHandler()

	// 业务下。
	h.Add("ListInBiz", http.MethodPost, "/bizs/{bk_biz_id}/instance_types/list", svc.ListInBiz)

	// 资源下。
	h.Add("ListInRes", http.MethodPost, "/instance_types/list", svc.ListInRes)

	// AWS AssumeRole cross-account data pass-through (resource scope).
	h.Add("ListAssumeRoleInstanceTypeInRes", http.MethodPost,
		"/vendors/aws/assume_role/instance_types/list", svc.ListAssumeRoleInstanceTypeInRes)
	h.Add("ListAssumeRoleInstanceInRes", http.MethodPost,
		"/vendors/aws/assume_role/instances/list", svc.ListAssumeRoleInstanceInRes)
	h.Add("ListAssumeRoleMetricDataInRes", http.MethodPost,
		"/vendors/aws/assume_role/cloudwatch/metric_data/list", svc.ListAssumeRoleMetricDataInRes)
	h.Add("ListAssumeRoleMetricsInRes", http.MethodPost,
		"/vendors/aws/assume_role/cloudwatch/metrics/list", svc.ListAssumeRoleMetricsInRes)

	h.Load(c.WebService)
}
