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
	"hcm/pkg/rest"
)

// InitRouteTableService initial the route table service
func InitRouteTableService(cap *capability.Capability) {
	r := &routeTable{
		ad: cap.CloudAdaptor,
		cs: cap.ClientSet,
	}

	h := rest.NewHandler()

	h.Add("TCloudRouteTableUpdate", "PATCH", "/vendors/tcloud/route_tables/{id}", r.TCloudRouteTableUpdate)
	h.Add("AwsRouteTableUpdate", "PATCH", "/vendors/aws/route_tables/{id}", r.AwsRouteTableUpdate)
	h.Add("HuaWeiRouteTableUpdate", "PATCH", "/vendors/huawei/route_tables/{id}", r.HuaWeiRouteTableUpdate)
	h.Add("AzureRouteTableUpdate", "PATCH", "/vendors/azure/route_tables/{id}", r.AzureRouteTableUpdate)

	h.Add("TCloudRouteTableDelete", "DELETE", "/vendors/tcloud/route_tables/{id}", r.TCloudRouteTableDelete)
	h.Add("AwsRouteTableDelete", "DELETE", "/vendors/aws/route_tables/{id}", r.AwsRouteTableDelete)
	h.Add("HuaWeiRouteTableDelete", "DELETE", "/vendors/huawei/route_tables/{id}", r.HuaWeiRouteTableDelete)
	h.Add("AzureRouteTableDelete", "DELETE", "/vendors/azure/route_tables/{id}", r.AzureRouteTableDelete)

	h.Load(cap.WebService)
}

type routeTable struct {
	ad *cloudadaptor.CloudAdaptorClient
	cs *client.ClientSet
}
