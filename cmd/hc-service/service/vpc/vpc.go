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
	"hcm/cmd/hc-service/service/capability"
	cloudadaptor "hcm/cmd/hc-service/service/cloud-adaptor"
	"hcm/pkg/client"
	"hcm/pkg/rest"
)

// InitVpcService initial the vpc service
func InitVpcService(cap *capability.Capability) {
	a := &vpc{
		ad: cap.CloudAdaptor,
		cs: cap.ClientSet,
	}

	h := rest.NewHandler()

	h.Add("TCloudVpcUpdate", "PATCH", "/vendors/tcloud/vpcs/{id}", a.TCloudVpcUpdate)
	h.Add("AwsVpcUpdate", "PATCH", "/vendors/aws/vpcs/{id}", a.AwsVpcUpdate)
	h.Add("HuaWeiVpcUpdate", "PATCH", "/vendors/huawei/vpcs/{id}", a.HuaWeiVpcUpdate)
	h.Add("GcpVpcUpdate", "PATCH", "/vendors/gcp/vpcs/{id}", a.GcpVpcUpdate)
	h.Add("AzureVpcUpdate", "PATCH", "/vendors/azure/vpcs/{id}", a.AzureVpcUpdate)

	h.Add("TCloudVpcDelete", "DELETE", "/vendors/tcloud/vpcs/{id}", a.TCloudVpcDelete)
	h.Add("AwsVpcDelete", "DELETE", "/vendors/aws/vpcs/{id}", a.AwsVpcDelete)
	h.Add("HuaWeiVpcDelete", "DELETE", "/vendors/huawei/vpcs/{id}", a.HuaWeiVpcDelete)
	h.Add("GcpVpcDelete", "DELETE", "/vendors/gcp/vpcs/{id}", a.GcpVpcDelete)
	h.Add("AzureVpcDelete", "DELETE", "/vendors/azure/vpcs/{id}", a.AzureVpcDelete)

	h.Load(cap.WebService)
}

type vpc struct {
	ad *cloudadaptor.CloudAdaptorClient
	cs *client.ClientSet
}
