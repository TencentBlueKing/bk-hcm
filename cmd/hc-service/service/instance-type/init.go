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

package instancetype

import (
	"hcm/cmd/hc-service/service/capability"
	cloudclient "hcm/cmd/hc-service/service/cloud-adaptor"
	dataservice "hcm/pkg/client/data-service"
	"hcm/pkg/rest"
)

type instanceTypeAdaptor struct {
	adaptor *cloudclient.CloudAdaptorClient
	dataCli *dataservice.Client
}

// InitInstanceTypeService ...
func InitInstanceTypeService(cap *capability.Capability) {
	i := &instanceTypeAdaptor{
		adaptor: cap.CloudAdaptor,
		dataCli: cap.ClientSet.DataService(),
	}

	h := rest.NewHandler()

	h.Add("ListForTCloud", "POST", "/vendors/tcloud/instance_types/list", i.ListForTCloud)
	h.Add("ListForAws", "POST", "/vendors/aws/instance_types/list", i.ListForAws)
	h.Add("ListForHuaWei", "POST", "/vendors/huawei/instance_types/list", i.ListForHuaWei)
	h.Add("ListForAzure", "POST", "/vendors/azure/instance_types/list", i.ListForAzure)
	h.Add("ListForGcp", "POST", "/vendors/gcp/instance_types/list", i.ListForGcp)

	h.Load(cap.WebService)
}
