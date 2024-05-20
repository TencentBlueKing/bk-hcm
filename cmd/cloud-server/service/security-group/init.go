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

package securitygroup

import (
	"net/http"

	"hcm/cmd/cloud-server/logics/audit"
	"hcm/cmd/cloud-server/service/capability"
	"hcm/pkg/client"
	"hcm/pkg/iam/auth"
	"hcm/pkg/rest"
)

// InitSecurityGroupService initial the security group service
func InitSecurityGroupService(c *capability.Capability) {
	svc := &securityGroupSvc{
		client:     c.ApiClient,
		authorizer: c.Authorizer,
		audit:      c.Audit,
	}

	h := rest.NewHandler()

	// 资源下安全组相关接口
	h.Add("CreateSecurityGroup", http.MethodPost, "/security_groups/create", svc.CreateSecurityGroup)
	h.Add("GetSecurityGroup", http.MethodGet, "/security_groups/{id}", svc.GetSecurityGroup)
	h.Add("BatchUpdateSecurityGroup", http.MethodPatch, "/security_groups/{id}", svc.UpdateSecurityGroup)
	h.Add("BatchDeleteSecurityGroup", http.MethodDelete, "/security_groups/batch", svc.BatchDeleteSecurityGroup)
	h.Add("ListSecurityGroup", http.MethodPost, "/security_groups/list", svc.ListSecurityGroup)
	h.Add("ListSecurityGroupsByCvmID", http.MethodGet, "/security_groups/cvms/{cvm_id}", svc.ListSecurityGroupsByCvmID)
	h.Add("AssignSecurityGroupToBiz", http.MethodPost, "/security_groups/assign/bizs", svc.AssignSecurityGroupToBiz)
	h.Add("AssociateCvm", http.MethodPost, "/security_groups/associate/cvms", svc.AssociateCvm)
	h.Add("DisassociateCvm", http.MethodPost, "/security_groups/disassociate/cvms", svc.DisassociateCvm)
	h.Add("AssociateSubnet", http.MethodPost, "/security_groups/associate/subnets", svc.AssociateSubnet)
	h.Add("DisAssociateSubnet", http.MethodPost, "/security_groups/disassociate/subnets", svc.DisAssociateSubnet)
	h.Add("AssociateNetworkInterface", http.MethodPost, "/security_groups/associate/network_interfaces",
		svc.AssociateNetworkInterface)
	h.Add("DisAssociateNetworkInterface", http.MethodPost, "/security_groups/disassociate/network_interfaces",
		svc.DisAssociateNetworkInterface)

	h.Add("CreateSecurityGroupRule", http.MethodPost,
		"/vendors/{vendor}/security_groups/{security_group_id}/rules/create", svc.CreateSecurityGroupRule)
	h.Add("ListSecurityGroupRule", http.MethodPost,
		"/vendors/{vendor}/security_groups/{security_group_id}/rules/list", svc.ListSecurityGroupRule)
	h.Add("UpdateSecurityGroupRule", http.MethodPut,
		"/vendors/{vendor}/security_groups/{security_group_id}/rules/{id}", svc.UpdateSecurityGroupRule)
	h.Add("DeleteSecurityGroupRule", http.MethodDelete,
		"/vendors/{vendor}/security_groups/{security_group_id}/rules/{id}", svc.DeleteSecurityGroupRule)
	h.Add("GetAzureDefaultSGRule", http.MethodGet, "/vendors/azure/default/security_groups/rules/{type}",
		svc.GetAzureDefaultSGRule)

	// 业务下安全组相关接口
	h.Add("CreateBizSecurityGroup", http.MethodPost, "/bizs/{bk_biz_id}/security_groups/create",
		svc.CreateBizSecurityGroup)
	h.Add("GetBizSecurityGroup", http.MethodGet, "/bizs/{bk_biz_id}/security_groups/{id}", svc.GetBizSecurityGroup)
	h.Add("UpdateBizSecurityGroup", http.MethodPatch, "/bizs/{bk_biz_id}/security_groups/{id}",
		svc.UpdateBizSecurityGroup)
	h.Add("BatchDeleteBizSecurityGroup", http.MethodDelete, "/bizs/{bk_biz_id}/security_groups/batch",
		svc.BatchDeleteBizSecurityGroup)
	h.Add("ListBizSecurityGroup", http.MethodPost, "/bizs/{bk_biz_id}/security_groups/list", svc.ListBizSecurityGroup)
	h.Add("ListBizSecurityGroupsByCvmID", http.MethodGet, "/bizs/{bk_biz_id}/security_groups/cvms/{cvm_id}",
		svc.ListBizSecurityGroupsByCvmID)
	h.Add("AssociateBizCvm", http.MethodPost, "/bizs/{bk_biz_id}/security_groups/associate/cvms", svc.AssociateBizCvm)
	h.Add("DisassociateCvm", http.MethodPost, "/bizs/{bk_biz_id}/security_groups/disassociate/cvms",
		svc.DisassociateBizCvm)
	h.Add("AssociateBizSubnet", http.MethodPost, "/bizs/{bk_biz_id}/security_groups/associate/subnets",
		svc.AssociateBizSubnet)
	h.Add("DisAssociateBizSubnet", http.MethodPost, "/bizs/{bk_biz_id}/security_groups/disassociate/subnets",
		svc.DisAssociateBizSubnet)
	h.Add("AssociateBizNIC", http.MethodPost, "/bizs/{bk_biz_id}/security_groups/associate/network_interfaces",
		svc.AssociateBizNIC)
	h.Add("DisAssociateBizNIC", http.MethodPost, "/bizs/{bk_biz_id}/security_groups/disassociate/network_interfaces",
		svc.DisAssociateBizNIC)

	h.Add("CreateBizSGRule", http.MethodPost,
		"/bizs/{bk_biz_id}/vendors/{vendor}/security_groups/{security_group_id}/rules/create", svc.CreateBizSGRule)
	h.Add("ListBizSGRule", http.MethodPost,
		"/bizs/{bk_biz_id}/vendors/{vendor}/security_groups/{security_group_id}/rules/list", svc.ListBizSGRule)
	h.Add("UpdateBizSGRule", http.MethodPut,
		"/bizs/{bk_biz_id}/vendors/{vendor}/security_groups/{security_group_id}/rules/{id}", svc.UpdateBizSGRule)
	h.Add("DeleteBizSGRule", http.MethodDelete,
		"/bizs/{bk_biz_id}/vendors/{vendor}/security_groups/{security_group_id}/rules/{id}", svc.DeleteBizSGRule)

	initSecurityGroupServiceHooks(svc, h)

	h.Load(c.WebService)
}

type securityGroupSvc struct {
	client     *client.ClientSet
	authorizer auth.Authorizer
	audit      audit.Interface
}
