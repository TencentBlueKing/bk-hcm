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

// Package routetable defines route table service.
package routetable

import (
	"hcm/cmd/hc-service/service/capability"
	cloudadaptor "hcm/cmd/hc-service/service/cloud-adaptor"
	"hcm/pkg/client"
	dataservice "hcm/pkg/client/data-service"
	"hcm/pkg/rest"
)

// InitRouteTableService initial the route table service
func InitRouteTableService(cap *capability.Capability) {
	v := &routeTable{
		ad:      cap.CloudAdaptor,
		cs:      cap.ClientSet,
		dataCli: cap.ClientSet.DataService(),
	}

	h := rest.NewHandler()

	h.Add("TCloudRouteTableSync", "POST", "/vendors/tcloud/route_tables/sync", v.SyncTCloudRouteTable)
	h.Add("HuaWeiRouteTableSync", "POST", "/vendors/huawei/route_tables/sync", v.SyncHuaWeiRouteTable)
	h.Add("AwsRouteTableSync", "POST", "/vendors/aws/route_tables/sync", v.SyncAwsRouteTable)
	h.Add("AzureRouteTableSync", "POST", "/vendors/azure/route_tables/sync", v.SyncAzureRouteTable)
	h.Add("GcpRouteTableSync", "POST", "/vendors/gcp/route_tables/sync", v.SyncGcpRouteTable)

	h.Load(cap.WebService)
}

type routeTable struct {
	ad      *cloudadaptor.CloudAdaptorClient
	cs      *client.ClientSet
	dataCli *dataservice.Client
}
