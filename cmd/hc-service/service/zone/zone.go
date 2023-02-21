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

package zone

import (
	"hcm/cmd/hc-service/service/capability"
	cloudadaptor "hcm/cmd/hc-service/service/cloud-adaptor"
	dataservice "hcm/pkg/client/data-service"
	"hcm/pkg/rest"
)

// InitZoneService initial the zone service
func InitZoneService(cap *capability.Capability) {
	z := &zoneHC{
		ad:      cap.CloudAdaptor,
		dataCli: cap.ClientSet.DataService(),
	}

	h := rest.NewHandler()

	h.Add("SyncHuaWeiZone", "POST", "/vendors/huawei/zones/sync", z.SyncHuaWeiZone)

	h.Add("SyncAwsZone", "POST", "/vendors/aws/zones/sync", z.SyncAwsZone)

	h.Add("SyncGcpZone", "POST", "/vendors/gcp/zones/sync", z.SyncGcpZone)

	h.Add("SyncTCloudZone", "POST", "/vendors/tcloud/zones/sync", z.SyncTCloudZone)

	h.Load(cap.WebService)
}

type zoneHC struct {
	ad      *cloudadaptor.CloudAdaptorClient
	dataCli *dataservice.Client
}
