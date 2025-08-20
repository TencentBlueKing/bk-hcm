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

package gcp

import (
	"hcm/cmd/hc-service/logics/cloud-adaptor"
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
	h.Path("/vendors/gcp")

	h.Add("SyncVpc", "POST", "/vpcs/sync", v.SyncVpc)
	h.Add("SyncSubnet", "POST", "/subnets/sync", v.SyncSubnet)
	h.Add("SyncDisk", "POST", "/disks/sync", v.SyncDisk)
	h.Add("SyncFirewallRule", "POST", "/firewalls/rules/sync", v.SyncFirewallRule)
	h.Add("SyncCvmWithRelRes", "POST", "/cvms/with/relation_resources/sync", v.SyncCvmWithRelRes)
	h.Add("SyncEip", "POST", "/eips/sync", v.SyncEip)
	h.Add("SyncRoute", "POST", "/routes/sync", v.SyncRoute)
	h.Add("SyncZone", "POST", "/zones/sync", v.SyncZone)
	h.Add("SyncRegion", "POST", "/regions/sync", v.SyncRegion)
	h.Add("SyncImage", "POST", "/images/sync", v.SyncImage)
	h.Add("SyncSubAccount", "POST", "/sub_accounts/sync", v.SyncSubAccount)
	h.Add("SyncCvmCCInfo", "POST", "/cvms/cc_info/sync", v.SyncCvmCCInfo)
	h.Add("SyncCvmCCInfoByCond", "POST", "/cvms/cc_info/by_condition/sync", v.SyncCvmCCInfoByCond)

	h.Load(cap.WebService)
}

type service struct {
	ad      *cloudadaptor.CloudAdaptorClient
	cs      *client.ClientSet
	dataCli *dataservice.Client
	syncCli ressync.Interface
}
