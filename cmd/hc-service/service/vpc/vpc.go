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

// Package vpc defines vpc service.
package vpc

import (
	"hcm/cmd/hc-service/logics/subnet"
	"hcm/cmd/hc-service/service/capability"
	cloudadaptor "hcm/cmd/hc-service/service/cloud-adaptor"
	"hcm/pkg/client"
	"hcm/pkg/rest"
)

// InitVpcService initial the vpc service
func InitVpcService(cap *capability.Capability) {
	v := &vpc{
		ad:     cap.CloudAdaptor,
		cs:     cap.ClientSet,
		subnet: subnet.NewSubnet(cap.ClientSet, cap.CloudAdaptor),
	}

	h := rest.NewHandler()

	h.Add("TCloudVpcCreate", "POST", "/vendors/tcloud/vpcs/create", v.TCloudVpcCreate)
	h.Add("AwsVpcCreate", "POST", "/vendors/aws/vpcs/create", v.AwsVpcCreate)
	h.Add("HuaWeiVpcCreate", "POST", "/vendors/huawei/vpcs/create", v.HuaWeiVpcCreate)
	h.Add("GcpVpcCreate", "POST", "/vendors/gcp/vpcs/create", v.GcpVpcCreate)
	h.Add("AzureVpcCreate", "POST", "/vendors/azure/vpcs/create", v.AzureVpcCreate)

	h.Add("TCloudVpcUpdate", "PATCH", "/vendors/tcloud/vpcs/{id}", v.TCloudVpcUpdate)
	h.Add("AwsVpcUpdate", "PATCH", "/vendors/aws/vpcs/{id}", v.AwsVpcUpdate)
	h.Add("HuaWeiVpcUpdate", "PATCH", "/vendors/huawei/vpcs/{id}", v.HuaWeiVpcUpdate)
	h.Add("GcpVpcUpdate", "PATCH", "/vendors/gcp/vpcs/{id}", v.GcpVpcUpdate)
	h.Add("AzureVpcUpdate", "PATCH", "/vendors/azure/vpcs/{id}", v.AzureVpcUpdate)

	h.Add("TCloudVpcDelete", "DELETE", "/vendors/tcloud/vpcs/{id}", v.TCloudVpcDelete)
	h.Add("AwsVpcDelete", "DELETE", "/vendors/aws/vpcs/{id}", v.AwsVpcDelete)
	h.Add("HuaWeiVpcDelete", "DELETE", "/vendors/huawei/vpcs/{id}", v.HuaWeiVpcDelete)
	h.Add("GcpVpcDelete", "DELETE", "/vendors/gcp/vpcs/{id}", v.GcpVpcDelete)
	h.Add("AzureVpcDelete", "DELETE", "/vendors/azure/vpcs/{id}", v.AzureVpcDelete)

	h.Load(cap.WebService)
}

type vpc struct {
	ad     *cloudadaptor.CloudAdaptorClient
	cs     *client.ClientSet
	subnet *subnet.Subnet
}
