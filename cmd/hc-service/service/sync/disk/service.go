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

package disk

import (
	"hcm/cmd/hc-service/service/capability"
	cloudclient "hcm/cmd/hc-service/service/cloud-adaptor"
	dataservice "hcm/pkg/client/data-service"
	"hcm/pkg/rest"
)

// InitSyncDiskService initial the sync disk service
func InitSyncDiskService(cap *capability.Capability) {
	svc := &syncDiskSvc{
		adaptor: cap.CloudAdaptor,
		dataCli: cap.ClientSet.DataService(),
	}

	h := rest.NewHandler()

	h.Add("SyncTCloudDisk", "POST", "/vendors/tcloud/disks/sync", svc.SyncTCloudDisk)
	h.Add("SyncHuaWeiDisk", "POST", "/vendors/huawei/disks/sync", svc.SyncHuaWeiDisk)
	h.Add("SyncAwsDisk", "POST", "/vendors/aws/disks/sync", svc.SyncAwsDisk)
	h.Add("SyncAzureDisk", "POST", "/vendors/azure/disks/sync", svc.SyncAzureDisk)
	h.Add("SyncGcpDisk", "POST", "/vendors/gcp/disks/sync", svc.SyncGcpDisk)

	h.Load(cap.WebService)
}

type syncDiskSvc struct {
	adaptor *cloudclient.CloudAdaptorClient
	dataCli *dataservice.Client
}
