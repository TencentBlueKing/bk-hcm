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

package networkinterface

import (
	"hcm/cmd/hc-service/service/capability"
	cloudclient "hcm/cmd/hc-service/service/cloud-adaptor"
	dataservice "hcm/pkg/client/data-service"
	"hcm/pkg/rest"
)

// InitNetworkInterfaceService initial the network interface service
func InitNetworkInterfaceService(cap *capability.Capability) {
	n := &networkInterfaceAdaptor{
		adaptor: cap.CloudAdaptor,
		dataCli: cap.ClientSet.DataService(),
	}

	h := rest.NewHandler()

	// network interface sync
	h.Add("AzureNetworkInterfaceSync", "POST", "/vendors/azure/network_interfaces/sync",
		n.AzureNetworkInterfaceSync)
	h.Add("GcpNetworkInterfaceSync", "POST", "/vendors/gcp/network_interfaces/sync",
		n.GcpNetworkInterfaceSync)
	h.Add("HuaWeiNetworkInterfaceSync", "POST", "/vendors/huawei/network_interfaces/sync",
		n.HuaWeiNetworkInterfaceSync)

	h.Load(cap.WebService)
}

type networkInterfaceAdaptor struct {
	adaptor *cloudclient.CloudAdaptorClient
	dataCli *dataservice.Client
}
