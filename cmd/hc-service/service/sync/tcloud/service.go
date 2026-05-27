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

package tcloud

import (
	"net/http"

	cloudadaptor "hcm/cmd/hc-service/logics/cloud-adaptor"
	ressync "hcm/cmd/hc-service/logics/res-sync"
	"hcm/cmd/hc-service/service/capability"
	"hcm/pkg/client"
	dataservice "hcm/pkg/client/data-service"
	"hcm/pkg/rest"
)

// InitService initial tcloud sync service
func InitService(cap *capability.Capability) {
	v := &service{
		ad:      cap.CloudAdaptor,
		cs:      cap.ClientSet,
		dataCli: cap.ClientSet.DataService(),
		syncCli: cap.ResSyncCli,
	}

	h := rest.NewHandler()
	h.Path("/vendors/tcloud")

	h.Add("SyncVpc", http.MethodPost, "/vpcs/sync", v.SyncVpc)
	h.Add("SyncSubnet", http.MethodPost, "/subnets/sync", v.SyncSubnet)
	h.Add("SyncDisk", http.MethodPost, "/disks/sync", v.SyncDisk)
	h.Add("SyncCvmWithRelRes", http.MethodPost, "/cvms/with/relation_resources/sync", v.SyncCvmWithRelRes)
	h.Add("SyncSecurityGroup", http.MethodPost, "/security_groups/sync", v.SyncSecurityGroup)
	h.Add("SyncSecurityGroupUsageBiz", http.MethodPost,
		"/security_groups/usage_biz_rels/sync", v.SyncSecurityGroupUsageBiz)
	h.Add("SyncEip", http.MethodPost, "/eips/sync", v.SyncEip)
	h.Add("SyncRoute", http.MethodPost, "/route_tables/sync", v.SyncRouteTable)
	h.Add("SyncZone", http.MethodPost, "/zones/sync", v.SyncZone)
	h.Add("SyncRegion", http.MethodPost, "/regions/sync", v.SyncRegion)
	h.Add("SyncImage", http.MethodPost, "/images/sync", v.SyncImage)
	h.Add("SyncSubAccount", http.MethodPost, "/sub_accounts/sync", v.SyncSubAccount)
	h.Add("SyncArgsTpl", http.MethodPost, "/argument_templates/sync", v.SyncArgsTpl)
	h.Add("SyncCert", http.MethodPost, "/certs/sync", v.SyncCert)
	h.Add("SyncLoadBalancer", http.MethodPost, "/load_balancers/sync", v.SyncLoadBalancer)
	h.Add("SyncLoadBalancerListener", http.MethodPost, "/listeners/sync", v.SyncLoadBalancerListener)
	h.Add("SyncCvmCCInfo", http.MethodPost, "/cvms/cc_info/sync", v.SyncCvmCCInfo)
	h.Add("SyncCvmCCInfoByCond", http.MethodPost, "/cvms/cc_info/by_condition/sync", v.SyncCvmCCInfoByCond)
	h.Add("SyncPermissionTemplate", http.MethodPost, "/permission_templates/sync", v.SyncPermissionTemplate)

	h.Load(cap.WebService)
}

type service struct {
	ad      *cloudadaptor.CloudAdaptorClient
	cs      *client.ClientSet
	dataCli *dataservice.Client
	syncCli ressync.Interface
}
