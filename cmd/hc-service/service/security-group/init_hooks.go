/*
 * TencentBlueKing is pleased to support the open source community by making
 * 蓝鲸智云 - 混合云管理平台 (BlueKing - Hybrid Cloud Management System) available.
 * Copyright (C) 2024 THL A29 Limited,
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

import "hcm/pkg/rest"

func initSecurityGroupServiceHooks(svc *securityGroup, h *rest.Handler) {

	h.Add("CreateZiyanSecurityGroup", "POST",
		"/vendors/tcloud-ziyan/security_groups/create", svc.CreateZiyanSecurityGroup)
	h.Add("DeleteZiyanSecurityGroup", "DELETE",
		"/vendors/tcloud-ziyan/security_groups/{id}", svc.DeleteZiyanSecurityGroup)
	h.Add("UpdateZiyanSecurityGroup", "PATCH",
		"/vendors/tcloud-ziyan/security_groups/{id}", svc.UpdateZiyanSecurityGroup)
	h.Add("BatchCreateZiyanSGRule", "POST",
		"/vendors/tcloud-ziyan/security_groups/{security_group_id}/rules/batch/create", svc.BatchCreateZiyanSGRule)
	h.Add("UpdateZiyanSGRule", "PUT",
		"/vendors/tcloud-ziyan/security_groups/{security_group_id}/rules/{id}", svc.UpdateZiyanSGRule)
	h.Add("DeleteZiyanSGRule", "DELETE",
		"/vendors/tcloud-ziyan/security_groups/{security_group_id}/rules/{id}", svc.DeleteZiyanSGRule)

	h.Add("BatchUpdateZiyanSGRule", "PUT",
		"/vendors/tcloud-ziyan/security_groups/{security_group_id}/rules/batch/update", svc.BatchUpdateZiyanSGRule)

	h.Add("ZiyanSGBatchAssociateCvm", "POST",
		"/vendors/tcloud-ziyan/security_groups/associate/cvms/batch", svc.ZiyanSGBatchAssociateCvm)
	h.Add("ZiyanSGBatchDisassociateCvm", "POST",
		"/vendors/tcloud-ziyan/security_groups/disassociate/cvms/batch", svc.ZiyanSGBatchDisassociateCvm)
	h.Add("ZiyanSGBatchAssociateLoadBalancers", "POST",
		"/vendors/tcloud-ziyan/security_groups/associate/load_balancers/batch", svc.ZiyanSGBatchAssociateLoadBalancers)
	h.Add("ZiyanSGBatchDisassociateLoadBalancers", "POST",
		"/vendors/tcloud-ziyan/security_groups/disassociate/load_balancers/batch",
		svc.ZiyanSGBatchDisassociateLoadBalancers)

	h.Add("ZiyanSecurityGroupAssociateLoadBalancer", "POST",
		"/vendors/tcloud-ziyan/security_groups/associate/load_balancers",
		svc.ZiyanSecurityGroupAssociateLoadBalancer)
	h.Add("ZiyanSecurityGroupDisassociateLoadBalancer", "POST",
		"/vendors/tcloud-ziyan/security_groups/disassociate/load_balancers",
		svc.ZiyanSecurityGroupDisassociateLoadBalancer)
	h.Add("ZiyanListSecurityGroupStatistic", "POST",
		"/vendors/tcloud-ziyan/security_groups/statistic", svc.ZiyanListSecurityGroupStatistic)
	h.Add("ZiyanCloneSecurityGroup", "POST",
		"/vendors/tcloud-ziyan/security_groups/clone", svc.ZiyanCloneSecurityGroup)

	h.Add("ZiyanSecurityGroupAssociateCvm", "POST",
		"/vendors/tcloud-ziyan/security_groups/associate/cvms", svc.ZiyanSecurityGroupAssociateCvm)
	h.Add("TCloudSecurityGroupDisassociateCvm", "POST",
		"/vendors/tcloud-ziyan/security_groups/disassociate/cvms", svc.ZiyanSecurityGroupDisassociateCvm)
}
