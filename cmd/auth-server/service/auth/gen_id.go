/*
 * TencentBlueKing is pleased to support the open source community by making
 * 蓝鲸智云 - 混合云管理平台 (BlueKing - Hybrid Cloud Management System) available.
 * Copyright (C) 2022 THL A29 Limited,
 * a Tencent company. All rights reserved.
 * Licensed under the MIT License (the "License");
 * you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at http://opensource.org/licenses/MIT
 * Unless required by accountlicable law or agreed to in writing,
 * software distributed under the License is distributed on
 * an "AS IS" BASIS, WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND,
 * either express or implied. See the License for the
 * specific language governing permissions and limitations under the License.
 *
 * We undertake not to change the open source license (MIT license) accountlicable
 *
 * to the current version of the project delivered to anyone in the future.
 */

package auth

import (
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
	case meta.Delete:
		// delete account is related to hcm account resource
		return sys.AccountDelete, []client.Resource{res}, nil
	default:
		return "", nil, errf.Newf(errf.InvalidParameter, "unsupported hcm action: %s", a.Basic.Action)
	}
}

// genResourceResource generate resource related iam resource.
func genResourceResource(a *meta.ResourceAttribute) (client.ActionID, []client.Resource, error) {
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
		// find resource is related to hcm account resource
		return sys.ResourceFind, []client.Resource{res}, nil
	case meta.Assign:
		// access resource secret keys is related to hcm account resource
		return sys.ResourceAssign, []client.Resource{res}, nil
	case meta.Update, meta.Delete, meta.Create:
		// update resource is related to hcm account resource
		return sys.ResourceManage, []client.Resource{res}, nil
	default:
		return "", nil, errf.Newf(errf.InvalidParameter, "unsupported hcm action: %s", a.Basic.Action)
	}
}

// genVpcResource generate vpc related iam resource.
func genVpcResource(a *meta.ResourceAttribute) (client.ActionID, []client.Resource, error) {
	return genResourceResource(a)
}

// genSubnetResource generate subnet related iam resource.
func genSubnetResource(a *meta.ResourceAttribute) (client.ActionID, []client.Resource, error) {
	return genResourceResource(a)
}

// genDiskResource generate disk related iam resource.
func genDiskResource(a *meta.ResourceAttribute) (client.ActionID, []client.Resource, error) {
	return genResourceResource(a)
}

// genSecurityGroupResource generate security group related iam resource.
func genSecurityGroupResource(a *meta.ResourceAttribute) (client.ActionID, []client.Resource, error) {
	return genResourceResource(a)
}

// genSecurityGroupRuleResource generate security group rule related iam resource.
func genSecurityGroupRuleResource(a *meta.ResourceAttribute) (client.ActionID, []client.Resource, error) {
	return genResourceResource(a)
}

// genGcpFirewallRuleResource generate gcp firewall rule related iam resource.
func genGcpFirewallRuleResource(a *meta.ResourceAttribute) (client.ActionID, []client.Resource, error) {
	return genResourceResource(a)
}

// genRecycleBinResource generate recycle bin related iam resource.
func genRecycleBinResource(a *meta.ResourceAttribute) (client.ActionID, []client.Resource, error) {
	switch a.Basic.Action {
	case meta.Find:
		return sys.ResourceFind, make([]client.Resource, 0), nil
	case meta.Recycle:
		return sys.RecycleBinManage, make([]client.Resource, 0), nil
	case meta.Recover:
		return sys.RecycleBinManage, make([]client.Resource, 0), nil
	default:
		return "", nil, errf.Newf(errf.InvalidParameter, "unsupported hcm action: %s", a.Basic.Action)
	}
}

// genAuditResource generate audit log related iam resource.
func genAuditResource(a *meta.ResourceAttribute) (client.ActionID, []client.Resource, error) {
	switch a.Basic.Action {
	case meta.Find:
		return sys.AuditFind, make([]client.Resource, 0), nil
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

	switch a.Basic.Action {
	case meta.Find:
		// find resource is related to hcm account resource
		return sys.ResourceFind, []client.Resource{res}, nil
	case meta.Assign:
		// access resource secret keys is related to hcm account resource
		return sys.ResourceAssign, []client.Resource{res}, nil
	case meta.Update, meta.Delete, meta.Create, meta.Stop, meta.Reboot, meta.Start, meta.ResetPwd:
		// update resource is related to hcm account resource
		return sys.ResourceManage, []client.Resource{res}, nil
	default:
		return "", nil, errf.Newf(errf.InvalidParameter, "unsupported hcm action: %s", a.Basic.Action)
	}
}

// genRouteTableResource generate route table's related iam resource.
func genRouteTableResource(a *meta.ResourceAttribute) (client.ActionID, []client.Resource, error) {
	return genResourceResource(a)
}

// genRouteResource generate route's related iam resource.
func genRouteResource(a *meta.ResourceAttribute) (client.ActionID, []client.Resource, error) {
	return genResourceResource(a)
}
