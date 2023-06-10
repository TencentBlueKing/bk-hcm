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

package resourcegroup

import (
	"net/http"

	"hcm/cmd/cloud-server/service/capability"
	cloudproto "hcm/pkg/api/cloud-server/resource-group"
	dataproto "hcm/pkg/api/data-service/cloud/resource-group"
	"hcm/pkg/client"
	"hcm/pkg/criteria/errf"
	"hcm/pkg/iam/auth"
	"hcm/pkg/rest"
)

// InitResourceGroupService initialize the resource group service.
func InitResourceGroupService(c *capability.Capability) {
	svc := &ResourceGroupSvc{
		client:     c.ApiClient,
		authorizer: c.Authorizer,
	}

	h := rest.NewHandler()

	h.Add("ListAzureResourceGroup", http.MethodPost, "/vendors/azure/resource_groups/list", svc.ListAzureResourceGroup)

	h.Load(c.WebService)
}

// ResourceGroupSvc resource group svc
type ResourceGroupSvc struct {
	client     *client.ClientSet
	authorizer auth.Authorizer
}

// ListAzureResourceGroup ...
func (dSvc *ResourceGroupSvc) ListAzureResourceGroup(cts *rest.Contexts) (interface{}, error) {
	req := new(cloudproto.ResourceGroupListReq)
	if err := cts.DecodeInto(req); err != nil {
		return nil, err
	}

	if err := req.Validate(); err != nil {
		return nil, errf.NewFromErr(errf.InvalidParameter, err)
	}

	return dSvc.client.DataService().Azure.ResourceGroup.ListResourceGroup(
		cts.Kit.Ctx,
		cts.Kit.Header(),
		&dataproto.AzureRGListReq{
			Filter: req.Filter,
			Page:   req.Page,
		},
	)
}
