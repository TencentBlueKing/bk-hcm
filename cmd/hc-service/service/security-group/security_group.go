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
	"hcm/cmd/hc-service/service/capability"
	cloudadaptor "hcm/cmd/hc-service/service/cloud-adaptor"
	dataservice "hcm/pkg/client/data-service"
	"hcm/pkg/rest"
)

// InitSecurityGroupService initial the security group service
func InitSecurityGroupService(cap *capability.Capability) {
	sg := &securityGroup{
		ad:      cap.CloudAdaptor,
		dataCli: cap.ClientSet.DataService(),
	}

	h := rest.NewHandler()
	h.Add("SyncTCloudSecurityGroup", "POST", "/vendors/tcloud/security_groups/sync", sg.SyncTCloudSecurityGroup)
	h.Add("SyncTCloudSGRule", "POST", "/vendors/tcloud/security_groups/{security_group_id}/rules/sync", sg.SyncTCloudSGRule)
	h.Add("CreateTCloudSecurityGroup", "POST", "/vendors/tcloud/security_groups/create", sg.CreateTCloudSecurityGroup)
	h.Add("DeleteTCloudSecurityGroup", "DELETE", "/vendors/tcloud/security_groups/{id}", sg.DeleteTCloudSecurityGroup)
	h.Add("UpdateTCloudSecurityGroup", "PATCH", "/vendors/tcloud/security_groups/{id}", sg.UpdateTCloudSecurityGroup)
	h.Add("BatchCreateTCloudSGRule", "POST", "/vendors/tcloud/security_groups/{security_group_id}/rules/batch/create",
		sg.BatchCreateTCloudSGRule)
	h.Add("UpdateTCloudSGRule", "PUT", "/vendors/tcloud/security_groups/{security_group_id}/rules/{id}",
		sg.UpdateTCloudSGRule)
	h.Add("DeleteTCloudSGRule", "DELETE", "/vendors/tcloud/security_groups/{security_group_id}/rules/{id}",
		sg.DeleteTCloudSGRule)

	h.Add("SyncAwsSecurityGroup", "POST", "/vendors/aws/security_groups/sync", sg.SyncAwsSecurityGroup)
	h.Add("SyncAwsSGRule", "POST", "/vendors/aws/security_groups/{security_group_id}/rules/sync", sg.SyncAwsSGRule)
	h.Add("UpdateAwsSecurityGroup", "PATCH", "/vendors/aws/security_groups/{id}", sg.CreateAwsSecurityGroup)
	h.Add("DeleteAwsSecurityGroup", "DELETE", "/vendors/aws/security_groups/{id}", sg.DeleteAwsSecurityGroup)
	h.Add("BatchCreateAwsSGRule", "POST", "/vendors/aws/security_groups/{security_group_id}/rules/batch/create",
		sg.BatchCreateAwsSGRule)
	h.Add("UpdateAwsSGRule", "PUT", "/vendors/aws/security_groups/{security_group_id}/rules/{id}",
		sg.UpdateAwsSGRule)
	h.Add("DeleteAwsSGRule", "DELETE", "/vendors/aws/security_groups/{security_group_id}/rules/{id}",
		sg.DeleteAwsSGRule)

	h.Add("SyncHuaWeiSecurityGroup", "POST", "/vendors/huawei/security_groups/sync", sg.SyncHuaWeiSecurityGroup)
	h.Add("SyncHuaWeiSGRule", "POST", "/vendors/huawei/security_groups/{security_group_id}/rules/sync", sg.SyncHuaWeiSGRule)
	h.Add("CreateHuaWeiSecurityGroup", "POST", "/vendors/huawei/security_groups/create", sg.CreateHuaWeiSecurityGroup)
	h.Add("DeleteHuaWeiSecurityGroup", "DELETE", "/vendors/huawei/security_groups/{id}", sg.DeleteHuaWeiSecurityGroup)
	h.Add("UpdateHuaWeiSecurityGroup", "PATCH", "/vendors/huawei/security_groups/{id}", sg.UpdateHuaWeiSecurityGroup)
	h.Add("CreateHuaWeiSGRule", "POST", "/vendors/huawei/security_groups/{security_group_id}/rules/create",
		sg.CreateHuaWeiSGRule)
	h.Add("DeleteAwsSGRule", "DELETE", "/vendors/huawei/security_groups/{security_group_id}/rules/{id}",
		sg.DeleteHuaWeiSGRule)

	h.Add("SyncAzureSecurityGroup", "POST", "/vendors/azure/security_groups/sync", sg.SyncAzureSecurityGroup)
	h.Add("SyncAzureSGRule", "POST", "/vendors/azure/security_groups/{security_group_id}/rules/sync", sg.SyncAzureSGRule)
	h.Add("CreateAzureSecurityGroup", "POST", "/vendors/azure/security_groups/create", sg.CreateAzureSecurityGroup)
	h.Add("DeleteAzureSecurityGroup", "DELETE", "/vendors/azure/security_groups/{id}", sg.DeleteAzureSecurityGroup)
	h.Add("UpdateAzureSecurityGroup", "PATCH", "/vendors/azure/security_groups/{id}", sg.UpdateAzureSecurityGroup)
	h.Add("BatchCreateAzureSGRule", "POST", "/vendors/azure/security_groups/{security_group_id}/rules/batch/create",
		sg.BatchCreateAzureSGRule)
	h.Add("UpdateAzureSGRule", "PUT", "/vendors/azure/security_groups/{security_group_id}/rules/{id}",
		sg.UpdateAzureSGRule)
	h.Add("DeleteAzureSGRule", "DELETE", "/vendors/azure/security_groups/{security_group_id}/rules/{id}",
		sg.DeleteAzureSGRule)

	h.Load(cap.WebService)
}

type securityGroup struct {
	ad      *cloudadaptor.CloudAdaptorClient
	dataCli *dataservice.Client
}
