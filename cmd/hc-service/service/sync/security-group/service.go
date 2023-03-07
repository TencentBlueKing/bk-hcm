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
	cloudclient "hcm/cmd/hc-service/service/cloud-adaptor"
	dataservice "hcm/pkg/client/data-service"
	"hcm/pkg/rest"
)

// InitSyncSecurityGroupService initial the sync security group service
func InitSyncSecurityGroupService(cap *capability.Capability) {
	svc := &syncSecurityGroupSvc{
		adaptor: cap.CloudAdaptor,
		dataCli: cap.ClientSet.DataService(),
	}

	h := rest.NewHandler()

	h.Add("SyncTCloudSecurityGroup", "POST", "/vendors/tcloud/security_groups/sync", svc.SyncTCloudSecurityGroup)
	h.Add("SyncTCloudSGRule", "POST", "/vendors/tcloud/security_groups/{security_group_id}/rules/sync",
		svc.SyncTCloudSGRule)

	h.Add("SyncHuaWeiSecurityGroup", "POST", "/vendors/huawei/security_groups/sync", svc.SyncHuaWeiSecurityGroup)
	h.Add("SyncHuaWeiSGRule", "POST", "/vendors/huawei/security_groups/{security_group_id}/rules/sync",
		svc.SyncHuaWeiSGRule)

	h.Add("SyncAwsSecurityGroup", "POST", "/vendors/aws/security_groups/sync", svc.SyncAwsSecurityGroup)
	h.Add("SyncAwsSGRule", "POST", "/vendors/aws/security_groups/{security_group_id}/rules/sync",
		svc.SyncAwsSGRule)

	h.Add("SyncAzureSecurityGroup", "POST", "/vendors/azure/security_groups/sync", svc.SyncAzureSecurityGroup)
	h.Add("SyncAzureSGRule", "POST", "/vendors/azure/security_groups/{security_group_id}/rules/sync",
		svc.SyncAzureSGRule)

	h.Load(cap.WebService)
}

type syncSecurityGroupSvc struct {
	adaptor *cloudclient.CloudAdaptorClient
	dataCli *dataservice.Client
}
