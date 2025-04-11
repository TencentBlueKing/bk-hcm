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
	securitygroup "hcm/cmd/cloud-server/logics/security-group"
	"hcm/cmd/cloud-server/service/capability"
	"hcm/pkg/client"
	"hcm/pkg/iam/auth"
	"hcm/pkg/rest"
	"hcm/pkg/thirdparty/esb"
)

// InitSecurityGroupService initial the security group service
func InitSecurityGroupService(c *capability.Capability) {
	svc := &securityGroupSvc{
		client:     c.ApiClient,
		authorizer: c.Authorizer,
		audit:      c.Audit,
		sgLogic:    c.Logics.SecurityGroup,
		esb:        c.EsbClient,
	}

	h := rest.NewHandler()

	// 资源下安全组相关接口
	h.Add("CreateSecurityGroup", http.MethodPost, "/security_groups/create", svc.CreateSecurityGroup)
	h.Add("GetSecurityGroup", http.MethodGet, "/security_groups/{id}", svc.GetSecurityGroup)
	h.Add("BatchUpdateSecurityGroup", http.MethodPatch, "/security_groups/{id}", svc.UpdateSecurityGroup)
	h.Add("UpdateSecurityGroupMgmtAttr", http.MethodPatch, "/security_groups/{id}/mgmt_attrs",
		svc.UpdateSGMgmtAttr)
	h.Add("BatchUpdateSGMgmtAttr", http.MethodPatch, "/security_groups/mgmt_attrs/batch",
		svc.BatchUpdateSGMgmtAttr)
	h.Add("BatchDeleteSecurityGroup", http.MethodDelete, "/security_groups/batch", svc.BatchDeleteSecurityGroup)
	h.Add("ListSecurityGroup", http.MethodPost, "/security_groups/list", svc.ListSecurityGroup)

	h.Add("AssociateCvm", http.MethodPost, "/security_groups/associate/cvms", svc.AssociateCvm)
	h.Add("DisassociateCvm", http.MethodPost, "/security_groups/disassociate/cvms", svc.DisassociateCvm)
	h.Add("AssociateSubnet", http.MethodPost, "/security_groups/associate/subnets", svc.AssociateSubnet)
	h.Add("DisAssociateSubnet", http.MethodPost, "/security_groups/disassociate/subnets", svc.DisAssociateSubnet)
	h.Add("AssociateNetworkInterface", http.MethodPost, "/security_groups/associate/network_interfaces",
		svc.AssociateNetworkInterface)
	h.Add("DisAssociateNetworkInterface", http.MethodPost, "/security_groups/disassociate/network_interfaces",
		svc.DisAssociateNetworkInterface)
	h.Add("AssignBizPreview", http.MethodPost, "/security_groups/assign/bizs/preview", svc.AssignBizPreview)
	h.Add("BatchAssignBiz", http.MethodPost, "/security_groups/assign/bizs/batch", svc.BatchAssignBiz)

	h.Add("CreateSecurityGroupRule", http.MethodPost,
		"/vendors/{vendor}/security_groups/{security_group_id}/rules/create", svc.CreateSecurityGroupRule)
	h.Add("ListSecurityGroupRule", http.MethodPost,
		"/vendors/{vendor}/security_groups/{security_group_id}/rules/list", svc.ListSecurityGroupRule)
	h.Add("UpdateSecurityGroupRule", http.MethodPut,
		"/vendors/{vendor}/security_groups/{security_group_id}/rules/{id}", svc.UpdateSecurityGroupRule)
	h.Add("BatchUpdateSecurityGroupRule", http.MethodPut,
		"/vendors/{vendor}/security_groups/{security_group_id}/rules/batch/update", svc.BatchUpdateSecurityGroupRule)
	h.Add("DeleteSecurityGroupRule", http.MethodDelete,
		"/vendors/{vendor}/security_groups/{security_group_id}/rules/{id}", svc.DeleteSecurityGroupRule)
	h.Add("GetAzureDefaultSGRule", http.MethodGet, "/vendors/azure/default/security_groups/rules/{type}",
		svc.GetAzureDefaultSGRule)

	h.Add("ListBizSecurityGroupsByResID", http.MethodGet,
		"/security_groups/res/{res_type}/{res_id}", svc.ListSecurityGroupsByResID)
	h.Add("ListResourceIdBySecurityGroup", http.MethodPost,
		"/security_group/{id}/common/list", svc.ListResourceIdBySecurityGroup)

	h.Add("QueryRelatedResourceCount", http.MethodPost,
		"/security_groups/related_resources/query_count", svc.QueryRelatedResourceCount)
	h.Add("ListSecurityGroupRelBusiness", http.MethodPost,
		"/security_groups/{security_group_id}/related_resources/bizs/list", svc.ListSecurityGroupRelBusiness)
	h.Add("ListSGRelCVMByBizID", http.MethodPost,
		"/security_groups/{sg_id}/related_resources/biz_resources/{res_biz_id}/cvms/list", svc.ListSGRelCVMByBizID)
	h.Add("ListSGRelLBByBizID", http.MethodPost,
		"/security_groups/{sg_id}/related_resources/biz_resources/{res_biz_id}/load_balancers/list",
		svc.ListSGRelLBByBizID)
	h.Add("CountSecurityGroupRules", http.MethodPost, "/security_groups/rules/count",
		svc.CountSecurityGroupRules)

	h.Add("BatchAssociateCvm", http.MethodPost,
		"/security_groups/associate/cvms/batch", svc.BatchAssociateCvm)
	h.Add("BatchDisassociateCvm", http.MethodPost,
		"/security_groups/disassociate/cvms/batch", svc.BatchDisassociateCvm)

	h.Add("BatchListResSecurityGroups", http.MethodPost, "/security_groups/res/{res_type}/batch",
		svc.BatchListResSecurityGroups)

	bizService(h, svc)
	initSecurityGroupServiceHooks(svc, h)

	h.Load(c.WebService)
}

func bizService(h *rest.Handler, svc *securityGroupSvc) {
	// 业务下安全组相关接口
	h.Add("CreateBizSecurityGroup", http.MethodPost, "/bizs/{bk_biz_id}/security_groups/create",
		svc.CreateBizSecurityGroup)
	h.Add("GetBizSecurityGroup", http.MethodGet, "/bizs/{bk_biz_id}/security_groups/{id}", svc.GetBizSecurityGroup)
	h.Add("UpdateBizSecurityGroup", http.MethodPatch, "/bizs/{bk_biz_id}/security_groups/{id}",
		svc.UpdateBizSecurityGroup)
	h.Add("UpdateSecurityGroupMgmtAttr", http.MethodPatch, "/bizs/{bk_biz_id}/security_groups/{id}/mgmt_attrs",
		svc.UpdateBizSGMgmtAttr)
	h.Add("BatchDeleteBizSecurityGroup", http.MethodDelete, "/bizs/{bk_biz_id}/security_groups/batch",
		svc.BatchDeleteBizSecurityGroup)
	h.Add("ListBizSecurityGroup", http.MethodPost, "/bizs/{bk_biz_id}/security_groups/list", svc.ListBizSecurityGroup)

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
	h.Add("ListBizSecurityGroupsByResID", http.MethodGet,
		"/bizs/{bk_biz_id}/security_groups/res/{res_type}/{res_id}", svc.ListBizSecurityGroupsByResID)
	h.Add("AssociateBizLb", http.MethodPost,
		"/bizs/{bk_biz_id}/security_groups/associate/load_balancers", svc.AssociateBizLb)
	h.Add("DisassociateBizLb", http.MethodPost, "/bizs/{bk_biz_id}/security_groups/disassociate/load_balancers",
		svc.DisassociateBizLb)

	h.Add("CreateBizSGRule", http.MethodPost,
		"/bizs/{bk_biz_id}/vendors/{vendor}/security_groups/{security_group_id}/rules/create", svc.CreateBizSGRule)
	h.Add("ListBizSGRule", http.MethodPost,
		"/bizs/{bk_biz_id}/vendors/{vendor}/security_groups/{security_group_id}/rules/list", svc.ListBizSGRule)
	h.Add("UpdateBizSGRule", http.MethodPut,
		"/bizs/{bk_biz_id}/vendors/{vendor}/security_groups/{security_group_id}/rules/{id}", svc.UpdateBizSGRule)
	h.Add("BatchUpdateBizSGRule", http.MethodPut,
		"/bizs/{bk_biz_id}/vendors/{vendor}/security_groups/{security_group_id}/rules/batch/update",
		svc.BatchUpdateBizSGRule)
	h.Add("DeleteBizSGRule", http.MethodDelete,
		"/bizs/{bk_biz_id}/vendors/{vendor}/security_groups/{security_group_id}/rules/{id}", svc.DeleteBizSGRule)

	h.Add("ListBizResourceIDBySecurityGroup", http.MethodPost,
		"/bizs/{bk_biz_id}/security_group/{id}/common/list", svc.ListBizResourceIDBySecurityGroup)

	h.Add("QueryBizRelatedResourceCount", http.MethodPost,
		"/bizs/{bk_biz_id}/security_groups/related_resources/query_count", svc.QueryBizRelatedResourceCount)
	h.Add("ListBizSecurityGroupRelBusiness", http.MethodPost,
		"/bizs/{bk_biz_id}/security_groups/{security_group_id}/related_resources/bizs/list",
		svc.ListBizSecurityGroupRelBusiness)
	h.Add("ListBizSGRelCVMByBizID", http.MethodPost,
		"/bizs/{bk_biz_id}/security_groups/{sg_id}/related_resources/biz_resources/{res_biz_id}/cvms/list",
		svc.ListBizSGRelCVMByBizID)
	h.Add("ListBizSGRelLBByBizID", http.MethodPost,
		"/bizs/{bk_biz_id}/security_groups/{sg_id}/related_resources/biz_resources/{res_biz_id}/load_balancers/list",
		svc.ListBizSGRelLBByBizID)
	h.Add("CountBizSecurityGroupRules", http.MethodPost, "/bizs/{bk_biz_id}/security_groups/rules/count",
		svc.CountBizSecurityGroupRules)
	h.Add("BizListSGMaintainerInfos", http.MethodPost,
		"/bizs/{bk_biz_id}/security_groups/maintainers_info/list", svc.BizListSGMaintainerInfos)

	h.Add("CloneBizSecurityGroup", http.MethodPost,
		"/bizs/{bk_biz_id}/security_groups/{id}/clone", svc.CloneBizSecurityGroup)

	h.Add("BatchAssociateBizCvm", http.MethodPost,
		"/bizs/{bk_biz_id}/security_groups/associate/cvms/batch", svc.BatchAssociateBizCvm)
	h.Add("BatchDisassociateBizCvm", http.MethodPost,
		"/bizs/{bk_biz_id}/security_groups/disassociate/cvms/batch", svc.BatchDisassociateBizCvm)

	h.Add("BizBatchListResSecurityGroups", http.MethodPost, "/bizs/{bk_biz_id}/security_groups/res/{res_type}/batch",
		svc.BizBatchListResSecurityGroups)
}

type securityGroupSvc struct {
	client     *client.ClientSet
	authorizer auth.Authorizer
	audit      audit.Interface
	esb        esb.Client
	sgLogic    securitygroup.Interface
}
