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
	"hcm/pkg/rest"
)

// DeleteSecurityGroupRule delete security group rule.
func (svc securityGroupSvc) DeleteSecurityGroupRule(cts *rest.Contexts) (interface{}, error) {
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
