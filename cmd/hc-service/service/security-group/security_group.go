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
	"hcm/cmd/hc-service/logics/cloud-adaptor"
	ressync "hcm/cmd/hc-service/logics/res-sync"
	"hcm/cmd/hc-service/service/capability"
	dataservice "hcm/pkg/client/data-service"
	"hcm/pkg/rest"
)

// InitSecurityGroupService initial the security group service
func InitSecurityGroupService(cap *capability.Capability) {
	sg := &securityGroup{
		ad:      cap.CloudAdaptor,
		dataCli: cap.ClientSet.DataService(),
		syncCli: cap.ResSyncCli,
	}

	h := rest.NewHandler()
	tcloudService(h, sg)

	h.Add("AwsSecurityGroupAssociateCvm", "POST", "/vendors/aws/security_groups/associate/cvms",
		sg.AwsSecurityGroupAssociateCvm)
	h.Add("AwsSecurityGroupDisassociateCvm", "POST", "/vendors/aws/security_groups/disassociate/cvms",
		sg.AwsSecurityGroupDisassociateCvm)
	h.Add("CreateAwsSecurityGroup", "POST", "/vendors/aws/security_groups/create", sg.CreateAwsSecurityGroup)
	h.Add("DeleteAwsSecurityGroup", "DELETE", "/vendors/aws/security_groups/{id}", sg.DeleteAwsSecurityGroup)
	h.Add("BatchCreateAwsSGRule", "POST", "/vendors/aws/security_groups/{security_group_id}/rules/batch/create",
		sg.BatchCreateAwsSGRule)
	h.Add("UpdateAwsSGRule", "PUT", "/vendors/aws/security_groups/{security_group_id}/rules/{id}",
		sg.UpdateAwsSGRule)
	h.Add("DeleteAwsSGRule", "DELETE", "/vendors/aws/security_groups/{security_group_id}/rules/{id}",
		sg.DeleteAwsSGRule)
	h.Add("AwsListSecurityGroupStatistic", "POST", "/vendors/aws/security_groups/statistic",
		sg.AwsListSecurityGroupStatistic)

	h.Add("HuaWeiSecurityGroupAssociateCvm", "POST", "/vendors/huawei/security_groups/associate/cvms",
		sg.HuaWeiSecurityGroupAssociateCvm)
	h.Add("HuaWeiSecurityGroupDisassociateCvm", "POST", "/vendors/huawei/security_groups/disassociate/cvms",
		sg.HuaWeiSecurityGroupDisassociateCvm)
	h.Add("CreateHuaWeiSecurityGroup", "POST", "/vendors/huawei/security_groups/create", sg.CreateHuaWeiSecurityGroup)
	h.Add("DeleteHuaWeiSecurityGroup", "DELETE", "/vendors/huawei/security_groups/{id}", sg.DeleteHuaWeiSecurityGroup)
	h.Add("UpdateHuaWeiSecurityGroup", "PATCH", "/vendors/huawei/security_groups/{id}", sg.UpdateHuaWeiSecurityGroup)
	h.Add("CreateHuaWeiSGRule", "POST", "/vendors/huawei/security_groups/{security_group_id}/rules/create",
		sg.CreateHuaWeiSGRule)
	h.Add("DeleteHuaWeiSGRule", "DELETE", "/vendors/huawei/security_groups/{security_group_id}/rules/{id}",
		sg.DeleteHuaWeiSGRule)
	h.Add("HuaweiListSecurityGroupStatistic", "POST", "/vendors/huawei/security_groups/statistic",
		sg.HuaweiListSecurityGroupStatistic)

	h.Add("AzureSecurityGroupAssociateSubnet", "POST", "/vendors/azure/security_groups/associate/subnets",
		sg.AzureSecurityGroupAssociateSubnet)
	h.Add("AzureSecurityGroupAssociateNI", "POST", "/vendors/azure/security_groups/associate/network_interfaces",
		sg.AzureSecurityGroupAssociateNI)
	h.Add("AzureSecurityGroupDisassociateSubnet", "POST", "/vendors/azure/security_groups/disassociate/subnets",
		sg.AzureSGDisassociateSubnet)
	h.Add("AzureSecurityGroupDisassociateNI", "POST", "/vendors/azure/security_groups/disassociate/network_interfaces",
		sg.AzureSecurityGroupDisassociateNI)
	h.Add("CreateAzureSecurityGroup", "POST", "/vendors/azure/security_groups/create", sg.CreateAzureSecurityGroup)
	h.Add("DeleteAzureSecurityGroup", "DELETE", "/vendors/azure/security_groups/{id}", sg.DeleteAzureSecurityGroup)
	h.Add("UpdateAzureSecurityGroup", "PATCH", "/vendors/azure/security_groups/{id}", sg.UpdateAzureSecurityGroup)
	h.Add("BatchCreateAzureSGRule", "POST", "/vendors/azure/security_groups/{security_group_id}/rules/batch/create",
		sg.BatchCreateAzureSGRule)
	h.Add("UpdateAzureSGRule", "PUT", "/vendors/azure/security_groups/{security_group_id}/rules/{id}",
		sg.UpdateAzureSGRule)
	h.Add("DeleteAzureSGRule", "DELETE", "/vendors/azure/security_groups/{security_group_id}/rules/{id}",
		sg.DeleteAzureSGRule)
	h.Add("AzureListSecurityGroupStatistic", "POST", "/vendors/azure/security_groups/statistic",
		sg.AzureListSecurityGroupStatistic)

	// CLB负载均衡
	h.Add("TCloudSGAssociateLoadBalancer", "POST",
		"/vendors/tcloud/security_groups/associate/load_balancers", sg.TCloudSGAssociateLoadBalancer)
	h.Add("TCloudSGDisassociateLoadBalancer", "POST",
		"/vendors/tcloud/security_groups/disassociate/load_balancers", sg.TCloudSGDisassociateLoadBalancer)

	initSecurityGroupServiceHooks(sg, h)

	h.Load(cap.WebService)
}

func tcloudService(h *rest.Handler, sg *securityGroup) {
	h.Add("TCloudSecurityGroupAssociateCvm", "POST", "/vendors/tcloud/security_groups/associate/cvms",
		sg.TCloudSecurityGroupAssociateCvm)
	h.Add("TCloudSecurityGroupDisassociateCvm", "POST", "/vendors/tcloud/security_groups/disassociate/cvms",
		sg.TCloudSecurityGroupDisassociateCvm)
	h.Add("CreateTCloudSecurityGroup", "POST", "/vendors/tcloud/security_groups/create", sg.CreateTCloudSecurityGroup)
	h.Add("DeleteTCloudSecurityGroup", "DELETE", "/vendors/tcloud/security_groups/{id}", sg.DeleteTCloudSecurityGroup)
	h.Add("UpdateTCloudSecurityGroup", "PATCH", "/vendors/tcloud/security_groups/{id}", sg.UpdateTCloudSecurityGroup)
	h.Add("BatchCreateTCloudSGRule", "POST", "/vendors/tcloud/security_groups/{security_group_id}/rules/batch/create",
		sg.BatchCreateTCloudSGRule)
	h.Add("UpdateTCloudSGRule", "PUT", "/vendors/tcloud/security_groups/{security_group_id}/rules/{id}",
		sg.UpdateTCloudSGRule)
	h.Add("BatchUpdateTCloudSGRule", "PUT", "/vendors/tcloud/security_groups/{security_group_id}/rules/batch/update",
		sg.BatchUpdateTCloudSGRule)
	h.Add("DeleteTCloudSGRule", "DELETE", "/vendors/tcloud/security_groups/{security_group_id}/rules/{id}",
		sg.DeleteTCloudSGRule)
	h.Add("TCloudSGBatchAssociateCloudCvm", "POST",
		"/vendors/tcloud/security_groups/associate/cvms/batch", sg.TCloudSGBatchAssociateCvm)
	h.Add("TCloudSGBatchDisassociateCloudCvm", "POST",
		"/vendors/tcloud/security_groups/disassociate/cvms/batch", sg.TCloudSGBatchDisassociateCvm)
	h.Add("TCloudListSecurityGroupStatistic", "POST", "/vendors/tcloud/security_groups/statistic",
		sg.TCloudListSecurityGroupStatistic)
	h.Add("TCloudCloneSecurityGroup", "POST", "/vendors/tcloud/security_groups/clone",
		sg.TCloudCloneSecurityGroup)
}

type securityGroup struct {
	ad      *cloudadaptor.CloudAdaptorClient
	dataCli *dataservice.Client
	syncCli ressync.Interface
}
