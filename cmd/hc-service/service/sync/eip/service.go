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

package eip

import (
	"hcm/cmd/hc-service/service/capability"
	cloudclient "hcm/cmd/hc-service/service/cloud-adaptor"
	dataservice "hcm/pkg/client/data-service"
	"hcm/pkg/rest"
)

// InitSyncEipService initial the sync eip service
func InitSyncEipService(cap *capability.Capability) {
	svc := &syncEipSvc{
		adaptor: cap.CloudAdaptor,
		dataCli: cap.ClientSet.DataService(),
	}

	h := rest.NewHandler()

	h.Add("SyncTCloudEip", "POST", "/vendors/tcloud/eips/sync", svc.SyncTCloudEip)
	h.Add("SyncHuaWeiEip", "POST", "/vendors/huawei/eips/sync", svc.SyncHuaWeiEip)
	h.Add("SyncAwsEip", "POST", "/vendors/aws/eips/sync", svc.SyncAwsEip)
	h.Add("SyncAzureEip", "POST", "/vendors/azure/eips/sync", svc.SyncAzureEip)
	h.Add("SyncGcpEip", "POST", "/vendors/gcp/eips/sync", svc.SyncGcpEip)

	h.Load(cap.WebService)
}

type syncEipSvc struct {
	adaptor *cloudclient.CloudAdaptorClient
	dataCli *dataservice.Client
}
