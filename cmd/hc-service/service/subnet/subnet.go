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

// Package subnet defines subnet service.
package subnet

import (
	"hcm/cmd/hc-service/service/capability"
	cloudadaptor "hcm/cmd/hc-service/service/cloud-adaptor"
	"hcm/pkg/client"
	"hcm/pkg/rest"
)

// InitSubnetService initial the subnet service
func InitSubnetService(cap *capability.Capability) {
	s := &subnet{
		ad: cap.CloudAdaptor,
		cs: cap.ClientSet,
	}

	h := rest.NewHandler()

	h.Add("TCloudSubnetUpdate", "PATCH", "/vendors/tcloud/subnets/{id}", s.TCloudSubnetUpdate)
	h.Add("AwsSubnetUpdate", "PATCH", "/vendors/aws/subnets/{id}", s.AwsSubnetUpdate)
	h.Add("HuaWeiSubnetUpdate", "PATCH", "/vendors/huawei/subnets/{id}", s.HuaWeiSubnetUpdate)
	h.Add("GcpSubnetUpdate", "PATCH", "/vendors/gcp/subnets/{id}", s.GcpSubnetUpdate)
	h.Add("AzureSubnetUpdate", "PATCH", "/vendors/azure/subnets/{id}", s.AzureSubnetUpdate)

	h.Add("TCloudSubnetDelete", "DELETE", "/vendors/tcloud/subnets/{id}", s.TCloudSubnetDelete)
	h.Add("AwsSubnetDelete", "DELETE", "/vendors/aws/subnets/{id}", s.AwsSubnetDelete)
	h.Add("HuaWeiSubnetDelete", "DELETE", "/vendors/huawei/subnets/{id}", s.HuaWeiSubnetDelete)
	h.Add("GcpSubnetDelete", "DELETE", "/vendors/gcp/subnets/{id}", s.GcpSubnetDelete)
	h.Add("AzureSubnetDelete", "DELETE", "/vendors/azure/subnets/{id}", s.AzureSubnetDelete)

	// count subnet available ips
	h.Add("TCloudSubnetCountIP", "POST", "/vendors/tcloud/subnets/{id}/ips/count", s.TCloudSubnetCountIP)
	h.Add("AwsSubnetCountIP", "POST", "/vendors/aws/subnets/{id}/ips/count", s.AwsSubnetCountIP)
	h.Add("AzureSubnetCountIP", "POST", "/vendors/azure/subnets/{id}/ips/count", s.AzureSubnetCountIP)
	h.Add("HuaWeiSubnetCountIP", "POST", "/vendors/huawei/subnets/{id}/ips/count", s.HuaWeiSubnetCountIP)

	h.Load(cap.WebService)
}

type subnet struct {
	ad *cloudadaptor.CloudAdaptorClient
	cs *client.ClientSet
}
