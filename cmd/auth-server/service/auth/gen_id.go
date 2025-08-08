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

package auth

import (
	"strconv"

	"hcm/pkg/criteria/errf"
	"hcm/pkg/iam/client"
	"hcm/pkg/iam/meta"
	"hcm/pkg/iam/sys"
)

// genSkipResource generate iam resource for resource, using skip action.
func genSkipResource(_ *meta.ResourceAttribute) (client.ActionID, []client.Resource, error) {
	return sys.Skip, make([]client.Resource, 0), nil
}

// genAccountResource generate account related iam resource.
func genAccountResource(a *meta.ResourceAttribute) (client.ActionID, []client.Resource, error) {
	res := client.Resource{
		System: sys.SystemIDHCM,
		Type:   sys.Account,
	}

	// compatible for authorize any
	if len(a.ResourceID) > 0 {
		res.ID = a.ResourceID
	}

	switch a.Basic.Action {
	case meta.Find:
		// find account is related to hcm account resource
		return sys.AccountFind, []client.Resource{res}, nil
	case meta.KeyAccess:
		// access account secret keys is related to hcm account resource
		return sys.AccountKeyAccess, []client.Resource{res}, nil
	case meta.Import:
		return sys.AccountImport, make([]client.Resource, 0), nil
	case meta.Update:
		// update account is related to hcm account resource
		return sys.AccountEdit, []client.Resource{res}, nil
	case meta.UpdateRRT:
		// update account RecycleReserveTime is related to hcm account resource
		return sys.RecycleBinConfig, []client.Resource{res}, nil
	case meta.Delete:
		// update account RecycleReserveTime is related to hcm account resource
		return sys.AccountDelete, []client.Resource{res}, nil
	default:
		return "", nil, errf.Newf(errf.InvalidParameter, "unsupported hcm action: %s", a.Basic.Action)
	}
}

// genIaaSResourceResource generate iaas resource related iam resource.
func genIaaSResourceResource(a *meta.ResourceAttribute) (client.ActionID, []client.Resource, error) {
	if a.Basic.Action != meta.Assign && a.BizID > 0 {
		return genBizIaaSResResource(a)
	}

	res := client.Resource{
		System: sys.SystemIDHCM,
		Type:   sys.Account,
	}

	// compatible for authorize any
	if len(a.ResourceID) > 0 {
		res.ID = a.ResourceID
	}

	switch a.Basic.Action {
	case meta.Find, meta.Assign:
		// find & assign action use generic cloud resource auth.
		return genCloudResResource(a)
	case meta.Create, meta.Apply:
		// create resource is related to hcm account resource
		return sys.IaaSResCreate, []client.Resource{res}, nil
	case meta.Update:
		// update resource is related to hcm account resource
		return sys.IaaSResOperate, []client.Resource{res}, nil
	case meta.Delete, meta.Recycle:
		// delete resource is related to hcm account resource
		return sys.IaaSResDelete, []client.Resource{res}, nil
	case meta.Destroy, meta.Recover:
		return sys.RecycleBinOperate, []client.Resource{res}, nil
	default:
		return "", nil, errf.Newf(errf.InvalidParameter, "unsupported hcm action: %s", a.Basic.Action)
	}
}

// genBizIaaSResResource generate biz iaas resource related iam resource.
func genBizIaaSResResource(a *meta.ResourceAttribute) (client.ActionID, []client.Resource, error) {
	res := client.Resource{
		System: sys.SystemIDCMDB,
		Type:   sys.Biz,
	}

	// compatible for authorize any
	if a.BizID > 0 {
		res.ID = strconv.FormatInt(a.BizID, 10)
	}

	switch a.Basic.Action {
	case meta.Find:
		return sys.BizAccess, []client.Resource{res}, nil
	case meta.Create, meta.Apply:
		return sys.BizIaaSResCreate, []client.Resource{res}, nil
	case meta.Update:
		return sys.BizIaaSResOperate, []client.Resource{res}, nil
	case meta.Delete, meta.Recycle:
		return sys.BizIaaSResDelete, []client.Resource{res}, nil
	case meta.Destroy, meta.Recover:
		return sys.BizRecycleBinOperate, []client.Resource{res}, nil
	default:
		return "", nil, errf.Newf(errf.InvalidParameter, "unsupported hcm action: %s", a.Basic.Action)
	}
}

func genIaaSResourceSecurityGroupRule(a *meta.ResourceAttribute) (client.ActionID, []client.Resource, error) {
	if a.Basic.Action != meta.Assign && a.BizID > 0 {
		return genBizIaaSResSecurityGroupRule(a)
	}

	res := client.Resource{
		System: sys.SystemIDHCM,
		Type:   sys.Account,
	}

	// compatible for authorize any
	if len(a.ResourceID) > 0 {
		res.ID = a.ResourceID
	}

	switch a.Basic.Action {
	case meta.Find, meta.Assign:
		// find & assign action use generic cloud resource auth.
		return genCloudResResource(a)
	case meta.Create, meta.Update, meta.Delete:
		// create update delete resource is related to hcm operate
		return sys.IaaSResOperate, []client.Resource{res}, nil
	default:
		return "", nil, errf.Newf(errf.InvalidParameter, "unsupported hcm action: %s", a.Basic.Action)
	}
}

func genBizIaaSResSecurityGroupRule(a *meta.ResourceAttribute) (client.ActionID, []client.Resource, error) {
	res := client.Resource{
		System: sys.SystemIDCMDB,
		Type:   sys.Biz,
	}

	// compatible for authorize any
	if a.BizID > 0 {
		res.ID = strconv.FormatInt(a.BizID, 10)
	}

	switch a.Basic.Action {
	case meta.Find:
		return sys.BizAccess, []client.Resource{res}, nil
	case meta.Create, meta.Update, meta.Delete:
		return sys.BizIaaSResOperate, []client.Resource{res}, nil
	default:
		return "", nil, errf.Newf(errf.InvalidParameter, "unsupported hcm action: %s", a.Basic.Action)
	}
}

// genVpcResource generate vpc related iam resource.
func genVpcResource(a *meta.ResourceAttribute) (client.ActionID, []client.Resource, error) {
	return genIaaSResourceResource(a)
}

// genSubnetResource generate subnet related iam resource.
func genSubnetResource(a *meta.ResourceAttribute) (client.ActionID, []client.Resource, error) {
	return genIaaSResourceResource(a)
}

// genDiskResource generate disk related iam resource.
func genDiskResource(a *meta.ResourceAttribute) (client.ActionID, []client.Resource, error) {
	res := client.Resource{
		System: sys.SystemIDHCM,
		Type:   sys.Account,
	}

	// compatible for authorize any
	if len(a.ResourceID) > 0 {
		res.ID = a.ResourceID
	}

	bizRes := client.Resource{
		System: sys.SystemIDCMDB,
		Type:   sys.Biz,
		ID:     strconv.FormatInt(a.BizID, 10),
	}

	switch a.Basic.Action {
	case meta.Associate, meta.Disassociate:
		if a.BizID > 0 {
			return sys.BizIaaSResOperate, []client.Resource{bizRes}, nil
		}
		return sys.IaaSResOperate, []client.Resource{res}, nil
	default:
		return genIaaSResourceResource(a)
	}
}

// genSecurityGroupResource generate security group related iam resource.
func genSecurityGroupResource(a *meta.ResourceAttribute) (client.ActionID, []client.Resource, error) {
	res := client.Resource{
		System: sys.SystemIDHCM,
		Type:   sys.Account,
	}

	// compatible for authorize any
	if len(a.ResourceID) > 0 {
		res.ID = a.ResourceID
	}

	bizRes := client.Resource{
		System: sys.SystemIDCMDB,
		Type:   sys.Biz,
		ID:     strconv.FormatInt(a.BizID, 10),
	}

	switch a.Basic.Action {
	case meta.Associate, meta.Disassociate:
		if a.BizID > 0 {
			return sys.BizIaaSResOperate, []client.Resource{bizRes}, nil
		}
		return sys.IaaSResOperate, []client.Resource{res}, nil
	default:
		return genIaaSResourceResource(a)
	}
}

// genSecurityGroupRuleResource generate security group rule related iam resource.
func genSecurityGroupRuleResource(a *meta.ResourceAttribute) (client.ActionID, []client.Resource, error) {
	return genIaaSResourceSecurityGroupRule(a)
}

// genGcpFirewallRuleResource generate gcp firewall rule related iam resource.
func genGcpFirewallRuleResource(a *meta.ResourceAttribute) (client.ActionID, []client.Resource, error) {
	return genIaaSResourceResource(a)
}

// genRecycleBinResource generate recycle bin related iam resource.
func genRecycleBinResource(a *meta.ResourceAttribute) (client.ActionID, []client.Resource, error) {
	res := client.Resource{
		System: sys.SystemIDHCM,
		Type:   sys.Account,
		ID:     a.ResourceID,
	}

	bizRes := client.Resource{
		System: sys.SystemIDCMDB,
		Type:   sys.Biz,
		ID:     strconv.FormatInt(a.BizID, 10),
	}

	switch a.Basic.Action {
	case meta.Find:
		if a.BizID > 0 {
			return sys.BizAccess, []client.Resource{bizRes}, nil
		}
		return sys.RecycleBinAccess, []client.Resource{res}, nil
	case meta.Recycle, meta.Recover:
		if a.BizID > 0 {
			return sys.BizRecycleBinOperate, []client.Resource{bizRes}, nil
		}
		return sys.RecycleBinOperate, []client.Resource{res}, nil
	default:
		return "", nil, errf.Newf(errf.InvalidParameter, "unsupported hcm action: %s", a.Basic.Action)
	}
}

// genAuditResource generate audit log related iam resource.
func genAuditResource(a *meta.ResourceAttribute) (client.ActionID, []client.Resource, error) {
	if a.BizID > 0 {
		return genBizAuditResource(a)
	}
	return genResourceAuditResource(a)
}

// genBizAuditResource generate biz audit log related iam resource.
func genBizAuditResource(a *meta.ResourceAttribute) (client.ActionID, []client.Resource, error) {
	res := client.Resource{
		System: sys.SystemIDCMDB,
		Type:   sys.Biz,
		ID:     strconv.FormatInt(a.BizID, 10),
	}

	switch a.Basic.Action {
	case meta.Find:
		return sys.BizOperationRecordFind, []client.Resource{res}, nil
	default:
		return "", nil, errf.Newf(errf.InvalidParameter, "unsupported hcm action: %s", a.Basic.Action)
	}
}

// genResourceAuditResource generate resource audit log related iam resource.
func genResourceAuditResource(a *meta.ResourceAttribute) (client.ActionID, []client.Resource, error) {
	res := client.Resource{
		System: sys.SystemIDHCM,
		Type:   sys.Account,
		ID:     a.ResourceID,
	}

	switch a.Basic.Action {
	case meta.Find:
		return sys.OperationRecordFind, []client.Resource{res}, nil
	default:
		return "", nil, errf.Newf(errf.InvalidParameter, "unsupported hcm action: %s", a.Basic.Action)
	}
}

func genCvmResource(a *meta.ResourceAttribute) (client.ActionID, []client.Resource, error) {
	res := client.Resource{
		System: sys.SystemIDHCM,
		Type:   sys.Account,
	}

	// compatible for authorize any
	if len(a.ResourceID) > 0 {
		res.ID = a.ResourceID
	}

	bizRes := client.Resource{
		System: sys.SystemIDCMDB,
		Type:   sys.Biz,
		ID:     strconv.FormatInt(a.BizID, 10),
	}

	switch a.Basic.Action {
	case meta.Stop, meta.Reboot, meta.Start, meta.ResetPwd:
		if a.BizID > 0 {
			return sys.BizIaaSResOperate, []client.Resource{bizRes}, nil
		}
		return sys.IaaSResOperate, []client.Resource{res}, nil
	default:
		return genIaaSResourceResource(a)
	}
}

func genSubAccountResource(a *meta.ResourceAttribute) (client.ActionID, []client.Resource, error) {
	res := client.Resource{
		System: sys.SystemIDHCM,
		Type:   sys.Account,
		ID:     a.ResourceID,
	}

	switch a.Basic.Action {
	case meta.Find:
		return sys.AccountFind, []client.Resource{res}, nil
	case meta.Update:
		return sys.SubAccountEdit, []client.Resource{res}, nil
	default:
		return "", nil, errf.Newf(errf.InvalidParameter, "unsupported hcm action: %s", a.Basic.Action)
	}
}

// genRouteTableResource generate route table's related iam resource.
func genRouteTableResource(a *meta.ResourceAttribute) (client.ActionID, []client.Resource, error) {
	return genIaaSResourceResource(a)
}

// genRouteResource generate route's related iam resource.
func genRouteResource(a *meta.ResourceAttribute) (client.ActionID, []client.Resource, error) {
	return genIaaSResourceResource(a)
}

// genBizResource generate biz's related iam resource.
func genBizResource(a *meta.ResourceAttribute) (client.ActionID, []client.Resource, error) {
	res := client.Resource{
		System: sys.SystemIDCMDB,
		Type:   sys.Biz,
	}

	// compatible for authorize any
	if a.BizID > 0 {
		res.ID = strconv.FormatInt(a.BizID, 10)
	}

	switch a.Basic.Action {
	case meta.Access:
		return sys.BizAccess, []client.Resource{res}, nil
	default:
		return genIaaSResourceResource(a)
	}
}

// genNetworkInterfaceResource generate network interface related iam resource.
func genNetworkInterfaceResource(a *meta.ResourceAttribute) (client.ActionID, []client.Resource, error) {
	return genIaaSResourceResource(a)
}

// genEipResource ...
func genEipResource(a *meta.ResourceAttribute) (client.ActionID, []client.Resource, error) {
	res := client.Resource{
		System: sys.SystemIDHCM,
		Type:   sys.Account,
	}

	// compatible for authorize any
	if len(a.ResourceID) > 0 {
		res.ID = a.ResourceID
	}

	bizRes := client.Resource{
		System: sys.SystemIDCMDB,
		Type:   sys.Biz,
		ID:     strconv.FormatInt(a.BizID, 10),
	}

	switch a.Basic.Action {
	case meta.Associate, meta.Disassociate:
		if a.BizID > 0 {
			return sys.BizIaaSResOperate, []client.Resource{bizRes}, nil
		}
		return sys.IaaSResOperate, []client.Resource{res}, nil
	default:
		return genIaaSResourceResource(a)
	}
}

// genCloudResResource generate all cloud resource related iam resource.
func genCloudResResource(a *meta.ResourceAttribute) (client.ActionID, []client.Resource, error) {
	res := client.Resource{
		System: sys.SystemIDHCM,
		Type:   sys.Account,
		ID:     a.ResourceID,
	}

	switch a.Basic.Action {
	case meta.Find:
		// find resource is related to hcm account resource
		return sys.ResourceFind, []client.Resource{res}, nil
	case meta.Assign:
		// assign resource to biz is related to hcm account & cmdb biz resource
		bizRes := client.Resource{
			System: sys.SystemIDCMDB,
			Type:   sys.Biz,
			ID:     strconv.FormatInt(a.BizID, 10),
		}
		return sys.ResourceAssign, []client.Resource{res, bizRes}, nil
	default:
		return "", nil, errf.Newf(errf.InvalidParameter, "unsupported hcm action: %s", a.Basic.Action)
	}
}

func genCloudSelectionSchemeResource(a *meta.ResourceAttribute) (client.ActionID, []client.Resource, error) {
	res := client.Resource{
		System: sys.SystemIDHCM,
		Type:   sys.CloudSelectionScheme,
		ID:     a.ResourceID,
	}

	switch a.Basic.Action {
	case meta.Find:
		return sys.CloudSelectionSchemeFind, []client.Resource{res}, nil
	case meta.Update:
		return sys.CloudSelectionSchemeEdit, []client.Resource{res}, nil
	case meta.Delete:
		return sys.CloudSelectionSchemeDelete, []client.Resource{res}, nil
	case meta.Create:
		return sys.CloudSelectionRecommend, make([]client.Resource, 0), nil

	default:
		return "", nil, errf.Newf(errf.InvalidParameter, "unsupported hcm action: %s", a.Basic.Action)
	}
}

// genBizCollectionResource 业务收藏
func genBizCollectionResource(a *meta.ResourceAttribute) (client.ActionID, []client.Resource, error) {
	bizRes := client.Resource{
		System: sys.SystemIDCMDB,
		Type:   sys.Biz,
		ID:     strconv.FormatInt(a.BizID, 10),
	}

	return sys.BizAccess, []client.Resource{bizRes}, nil
}

// genProxyResourceFind 代理资源访问权限.
func genProxyResourceFind(a *meta.ResourceAttribute) (client.ActionID, []client.Resource, error) {
	if a.BizID != 0 {
		bizRes := client.Resource{
			System: sys.SystemIDCMDB,
			Type:   sys.Biz,
			ID:     strconv.FormatInt(a.BizID, 10),
		}

		return sys.BizAccess, []client.Resource{bizRes}, nil
	}

	res := client.Resource{
		System: sys.SystemIDHCM,
		Type:   sys.Account,
		ID:     a.ResourceID,
	}
	return sys.ResourceFind, []client.Resource{res}, nil
}

// genCostManageResource generate cost manage related iam resource.
func genCostManageResource(a *meta.ResourceAttribute) (client.ActionID, []client.Resource, error) {
	switch a.Basic.Action {
	case meta.Find:
		return sys.CostManage, make([]client.Resource, 0), nil
	default:
		return "", nil, errf.Newf(errf.InvalidParameter, "unsupported hcm action: %s", a.Basic.Action)
	}
}

// genArgumentTemplateResource generate argument template related iam resource.
func genArgumentTemplateResource(a *meta.ResourceAttribute) (client.ActionID, []client.Resource, error) {
	return genIaaSResourceResource(a)
}

// genCertResource generate cert related iam resource.
func genCertResource(a *meta.ResourceAttribute) (client.ActionID, []client.Resource, error) {
	res := client.Resource{
		System: sys.SystemIDHCM,
		Type:   sys.Account,
	}

	// compatible for authorize any
	if len(a.ResourceID) > 0 {
		res.ID = a.ResourceID
	}

	bizRes := client.Resource{
		System: sys.SystemIDCMDB,
		Type:   sys.Biz,
		ID:     strconv.FormatInt(a.BizID, 10),
	}

	switch a.Basic.Action {
	case meta.Find, meta.Assign:
		return genIaaSResourceResource(a)
	case meta.Create:
		if a.BizID > 0 {
			return sys.BizCertResCreate, []client.Resource{bizRes}, nil
		}
		return sys.CertResCreate, []client.Resource{res}, nil
	case meta.Update:
		// update resource is related to hcm account resource
		if a.BizID > 0 {
			return sys.BizIaaSResOperate, []client.Resource{bizRes}, nil
		}
		return sys.IaaSResOperate, []client.Resource{res}, nil
	case meta.Delete:
		if a.BizID > 0 {
			return sys.BizCertResDelete, []client.Resource{bizRes}, nil
		}
		return sys.CertResDelete, []client.Resource{res}, nil
	default:
		return "", nil, errf.Newf(errf.InvalidParameter, "unsupported hcm action: %s", a.Basic.Action)
	}
}

// genLoadBalancerResource generate load balancer related iam resource.
func genLoadBalancerResource(a *meta.ResourceAttribute) (client.ActionID, []client.Resource, error) {
	res := client.Resource{
		System: sys.SystemIDHCM,
		Type:   sys.Account,
	}

	// compatible for authorize any
	if len(a.ResourceID) > 0 {
		res.ID = a.ResourceID
	}

	if a.Basic.Action != meta.Assign && a.BizID > 0 {
		return genBizLoadBalancerResource(a)
	}
	switch a.Basic.Action {
	case meta.Associate, meta.Disassociate:
		if a.BizID > 0 {
			bizRes := client.Resource{
				System: sys.SystemIDCMDB,
				Type:   sys.Biz,
				ID:     strconv.FormatInt(a.BizID, 10),
			}
			return sys.BizCLBResOperate, []client.Resource{bizRes}, nil
		}
		return sys.IaaSResOperate, []client.Resource{res}, nil
	case meta.Create, meta.Apply:
		return sys.CLBResCreate, []client.Resource{res}, nil
	case meta.Find, meta.Assign:
		// find & assign action use generic cloud resource auth.
		return genCloudResResource(a)
	case meta.Update:
		// update resource is related to hcm account resource
		return sys.CLBResOperate, []client.Resource{res}, nil
	case meta.Delete:
		// delete resource is related to hcm account resource
		return sys.CLBResDelete, []client.Resource{res}, nil
	default:
		return "", nil, errf.Newf(errf.InvalidParameter, "unsupported hcm action: %s", a.Basic.Action)
	}
}

func genBizLoadBalancerResource(a *meta.ResourceAttribute) (client.ActionID, []client.Resource, error) {
	res := client.Resource{
		System: sys.SystemIDCMDB,
		Type:   sys.Biz,
	}

	// compatible for authorize any
	if a.BizID > 0 {
		res.ID = strconv.FormatInt(a.BizID, 10)
	}

	switch a.Basic.Action {
	case meta.Find:
		return sys.BizAccess, []client.Resource{res}, nil
	case meta.Create, meta.Apply:
		return sys.BizCLBResCreate, []client.Resource{res}, nil
	case meta.Update:
		return sys.BizCLBResOperate, []client.Resource{res}, nil
	case meta.Delete:
		return sys.BizCLBResDelete, []client.Resource{res}, nil
	default:
		return "", nil, errf.Newf(errf.InvalidParameter, "unsupported hcm action: %s", a.Basic.Action)
	}
}

// genListenerResource generate clb listener related iam resource.
func genListenerResource(a *meta.ResourceAttribute) (client.ActionID, []client.Resource, error) {
	return genLoadBalancerRelatedResources(a)
}

// genLoadBalancerRelatedResources 生成负载均衡下属资源的权限点
func genLoadBalancerRelatedResources(a *meta.ResourceAttribute) (client.ActionID, []client.Resource, error) {
	res := client.Resource{
		System: sys.SystemIDHCM,
		Type:   sys.Account,
	}

	// compatible for authorize any
	if len(a.ResourceID) > 0 {
		res.ID = a.ResourceID
	}

	if a.BizID > 0 {
		return genBizLoadBalancerRelatedResources(a)
	}
	switch a.Basic.Action {
	case meta.Find:
		return genCloudResResource(a)
	case meta.Create, meta.Update, meta.Delete:
		// update resource is related to hcm account resource
		return sys.CLBResOperate, []client.Resource{res}, nil
	default:
		return "", nil, errf.Newf(errf.InvalidParameter, "unsupported hcm action: %s", a.Basic.Action)
	}
}

// genBizLoadBalancerRelatedResources 生成业务下负载均衡下属资源的权限点
func genBizLoadBalancerRelatedResources(a *meta.ResourceAttribute) (client.ActionID, []client.Resource, error) {
	res := client.Resource{
		System: sys.SystemIDCMDB,
		Type:   sys.Biz,
	}

	// compatible for authorize any
	if a.BizID > 0 {
		res.ID = strconv.FormatInt(a.BizID, 10)
	}

	switch a.Basic.Action {
	case meta.Find:
		return sys.BizAccess, []client.Resource{res}, nil
	case meta.Create, meta.Update, meta.Delete:
		return sys.BizCLBResOperate, []client.Resource{res}, nil
	default:
		return "", nil, errf.Newf(errf.InvalidParameter, "unsupported hcm action: %s", a.Basic.Action)
	}
}

// genTargetGroupResource generate target group related iam resource.
func genTargetGroupResource(a *meta.ResourceAttribute) (client.ActionID, []client.Resource, error) {
	res := client.Resource{
		System: sys.SystemIDHCM,
		Type:   sys.Account,
	}

	// compatible for authorize any
	if len(a.ResourceID) > 0 {
		res.ID = a.ResourceID
	}

	bizRes := client.Resource{
		System: sys.SystemIDCMDB,
		Type:   sys.Biz,
		ID:     strconv.FormatInt(a.BizID, 10),
	}

	switch a.Basic.Action {
	case meta.Associate, meta.Disassociate:
		if a.BizID > 0 {
			return sys.BizCLBResOperate, []client.Resource{bizRes}, nil
		}
		return sys.IaaSResOperate, []client.Resource{res}, nil
	default:
		return genLoadBalancerRelatedResources(a)
	}
}

// genUrlRuleResource generate clb listener related iam resource.
func genUrlRuleResource(a *meta.ResourceAttribute) (client.ActionID, []client.Resource, error) {
	return genLoadBalancerRelatedResources(a)
}

func genMainAccountRuleResource(a *meta.ResourceAttribute) (client.ActionID, []client.Resource, error) {
	res := client.Resource{
		System: sys.SystemIDHCM,
		Type:   sys.MainAccount,
	}
	if len(a.ResourceID) > 0 {
		res.ID = a.ResourceID
	}

	switch a.Basic.Action {
	case meta.Find:
		return sys.MainAccountFind, []client.Resource{res}, nil
	case meta.Update:
		return sys.MainAccountEdit, []client.Resource{res}, nil
	case meta.Create:
		return sys.MainAccountCreate, []client.Resource{res}, nil
	default:
		return "", nil, errf.Newf(errf.InvalidParameter, "unsupported hcm action: %s", a.Basic.Action)
	}

}

func genRootAccountRuleResource(a *meta.ResourceAttribute) (client.ActionID, []client.Resource, error) {
	switch a.Basic.Action {
	case meta.Find, meta.Create, meta.Update:
		return sys.RootAccountManage, make([]client.Resource, 0), nil
	default:
		return "", nil, errf.Newf(errf.InvalidParameter, "unsupported hcm action: %s", a.Basic.Action)
	}
}

// 生成账单账号权限映射
func genAccountBillResource(a *meta.ResourceAttribute) (client.ActionID, []client.Resource, error) {
	switch a.Basic.Action {
	case meta.Find, meta.Delete, meta.Import, meta.Create, meta.Update, meta.Access:
		return sys.AccountBillManage, make([]client.Resource, 0), nil
	default:
		return "", nil, errf.Newf(errf.InvalidParameter, "unsupported hcm action: %s", a.Basic.Action)
	}
}

func genAccountBillThirdPartyResource(a *meta.ResourceAttribute) (client.ActionID, []client.Resource, error) {
	// TODO 改为 属性鉴权
	res := client.Resource{
		System: sys.SystemIDHCM,
		Type:   sys.BillCloudVendor,
	}
	switch a.Basic.Action {
	case meta.Find:
		if a.Type != meta.AccountBillThirdParty {
			return "", nil, errf.Newf(errf.InvalidParameter, "unsupported hcm res type: %s", a.Basic.Type)
		}
		res.ID = a.ResourceID
		return sys.AccountBillPull, []client.Resource{res}, nil
	default:
		return "", nil, errf.Newf(errf.InvalidParameter, "unsupported hcm action: %s", a.Basic.Action)
	}
}

func genImageResource(a *meta.ResourceAttribute) (client.ActionID, []client.Resource, error) {
	res := client.Resource{
		System: sys.SystemIDHCM,
		Type:   sys.Account,
	}

	// compatible for authorize any
	if len(a.ResourceID) > 0 {
		res.ID = a.ResourceID
	}

	switch a.Basic.Action {
	case meta.Find:
		if a.BizID > 0 {
			return genBizIaaSResResource(a)
		}
		return sys.AccountFind, []client.Resource{res}, nil
	default:
		return "", nil, errf.Newf(errf.InvalidParameter, "unsupported hcm action of image resource: %s", a.Basic.Action)
	}
}

func genTaskManagementResource(a *meta.ResourceAttribute) (client.ActionID, []client.Resource, error) {
	res := client.Resource{
		System: sys.SystemIDHCM,
		Type:   sys.Biz,
		ID:     strconv.FormatInt(a.BizID, 10),
	}

	switch a.Basic.Action {
	case meta.Find:
		return sys.BizAccess, []client.Resource{res}, nil
	case meta.Create, meta.Update, meta.Delete:
		return sys.BizTaskManagementOperate, []client.Resource{res}, nil
	default:
		return "", nil, errf.Newf(errf.InvalidParameter, "unsupported hcm action: %s", a.Basic.Action)
	}
}

func genCosBucket(a *meta.ResourceAttribute) (client.ActionID, []client.Resource, error) {
	res := client.Resource{
		System: sys.SystemIDHCM,
		Type:   sys.Account,
		ID:     a.ResourceID,
	}

	switch a.Basic.Action {
	case meta.Create:
		return sys.CosBucketCreate, []client.Resource{res}, nil
	case meta.Find:
		return sys.CosBucketFind, []client.Resource{res}, nil
	case meta.Delete:
		return sys.CosBucketDelete, []client.Resource{res}, nil
	default:
		return "", nil, errf.Newf(errf.InvalidParameter, "unsupported hcm action: %s", a.Basic.Action)
	}
}

func genCloudSelectionResource(*meta.ResourceAttribute) (client.ActionID, []client.Resource, error) {
	return sys.CloudSelectionRecommend, make([]client.Resource, 0), nil
}
