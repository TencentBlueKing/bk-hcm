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

package approvalprocess

import (
	"net/http"

	"hcm/cmd/cloud-server/service/capability"
	"hcm/pkg/api/core"
	dataproto "hcm/pkg/api/data-service"
	"hcm/pkg/client"
	"hcm/pkg/iam/auth"
	"hcm/pkg/logs"
	"hcm/pkg/rest"
	"hcm/pkg/runtime/filter"
)

// InitService initialize the service.
func InitService(c *capability.Capability) {
	svc := &service{
		client:     c.ApiClient,
		authorizer: c.Authorizer,
	}

	h := rest.NewHandler()

	h.Add("GetApprovalProcessServiceID", http.MethodGet, "/approval_processes/service_id",
		svc.GetApprovalProcessServiceID)

	h.Load(c.WebService)
}

type service struct {
	client     *client.ClientSet
	authorizer auth.Authorizer
}

// GetApprovalProcessServiceID 获取hcm itsm单据流程所在的服务目录ID列表
func (svc *service) GetApprovalProcessServiceID(cts *rest.Contexts) (interface{}, error) {

	req := &dataproto.ApprovalProcessListReq{
		Filter: &filter.Expression{
			Op:    filter.And,
			Rules: []filter.RuleFactory{},
		},
		// itsm 单据流程不可能超过500个，所以就不分页去查了。
		Page: core.NewDefaultBasePage(),
	}
	result, err := svc.client.DataService().Global.ApprovalProcess.List(cts.Kit.Ctx, cts.Kit.Header(), req)
	if err != nil {
		logs.Errorf("call data-service to list approval process failed, err: %v, rid: %s", err, cts.Kit.Rid)
		return nil, err
	}

	serviceIdMap := make(map[int64]struct{})
	serviceIds := make([]int64, 0)
	for _, one := range result.Details {
		if _, exists := serviceIdMap[one.ServiceID]; !exists {
			serviceIdMap[one.ServiceID] = struct{}{}
			serviceIds = append(serviceIds, one.ServiceID)
		}
	}

	return serviceIds, nil
}
