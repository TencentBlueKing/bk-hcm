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
	"hcm/pkg/criteria/enumor"
	"hcm/pkg/criteria/errf"
	"hcm/pkg/iam/meta"
	"hcm/pkg/logs"
	"hcm/pkg/rest"
	"hcm/pkg/tools/hooks/handler"
)

// DeleteSecurityGroupRule delete security group rule.
func (svc *securityGroupSvc) DeleteSecurityGroupRule(cts *rest.Contexts) (interface{}, error) {
	return svc.deleteSGRule(cts, handler.ResValidWithAuth)
}

// DeleteBizSGRule delete biz security group rule.
func (svc *securityGroupSvc) DeleteBizSGRule(cts *rest.Contexts) (interface{}, error) {
	return svc.deleteSGRule(cts, handler.BizValidWithAuth)
}

func (svc *securityGroupSvc) deleteSGRule(cts *rest.Contexts, validHandler handler.ValidWithAuthHandler) (interface{},
	error) {

	vendor := enumor.Vendor(cts.PathParameter("vendor").String())
	if len(vendor) == 0 {
		return nil, errf.New(errf.InvalidParameter, "vendor is required")
	}

	sgID := cts.PathParameter("security_group_id").String()
	if len(sgID) == 0 {
		return nil, errf.New(errf.InvalidParameter, "security group id is required")
	}

	id := cts.PathParameter("id").String()
	if len(id) == 0 {
		return nil, errf.New(errf.InvalidParameter, "id is required")
	}

	basicInfo, err := svc.client.DataService().Global.Cloud.GetResBasicInfo(cts.Kit,
		enumor.SecurityGroupCloudResType, sgID)
	if err != nil {
		return nil, err
	}

	// validate biz and authorize
	err = validHandler(cts, &handler.ValidWithAuthOption{Authorizer: svc.authorizer, ResType: meta.SecurityGroupRule,
		Action: meta.Delete, BasicInfo: basicInfo})
	if err != nil {
		return nil, err
	}

	// create delete audit.
	err = svc.audit.ChildResDeleteAudit(cts.Kit, enumor.SecurityGroupRuleAuditResType, sgID, []string{id})
	if err != nil {
		logs.Errorf("create delete audit failed, err: %v, rid: %s", err, cts.Kit.Rid)
		return nil, err
	}

	switch vendor {
	case enumor.TCloud:
		return nil, svc.client.HCService().TCloud.SecurityGroup.DeleteSecurityGroupRule(cts.Kit.Ctx,
			cts.Kit.Header(), sgID, id)

	case enumor.Aws:
		return nil, svc.client.HCService().Aws.SecurityGroup.DeleteSecurityGroupRule(cts.Kit.Ctx,
			cts.Kit.Header(), sgID, id)

	case enumor.HuaWei:
		return nil, svc.client.HCService().HuaWei.SecurityGroup.DeleteSecurityGroupRule(cts.Kit.Ctx,
			cts.Kit.Header(), sgID, id)

	case enumor.Azure:
		return nil, svc.client.HCService().Azure.SecurityGroup.DeleteSecurityGroupRule(cts.Kit.Ctx,
			cts.Kit.Header(), sgID, id)

	default:
		return nil, errf.Newf(errf.Unknown, "vendor: %s not support", vendor)
	}
}
