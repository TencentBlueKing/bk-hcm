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

package cvm

import (
	"net/http"

	"hcm/cmd/hc-service/service/capability"
	cloudclient "hcm/cmd/hc-service/service/cloud-adaptor"
	dataservice "hcm/pkg/client/data-service"
	"hcm/pkg/rest"
)

// InitSyncCvmService initial the cvm service
func InitSyncCvmService(cap *capability.Capability) {
	svc := &syncCvmSvc{
		adaptor: cap.CloudAdaptor,
		dataCli: cap.ClientSet.DataService(),
	}

	h := rest.NewHandler()

	h.Add("SyncTCloudCvm", http.MethodPost, "/vendors/tcloud/cvms/sync", svc.SyncTCloudCvm)
	h.Add("SyncTCloudCvmWithRelResource", http.MethodPost, "/vendors/tcloud/cvms/with/relation_resource/sync",
		svc.SyncTCloudCvmWithRelResource)

	h.Add("SyncHuaWeiCvm", http.MethodPost, "/vendors/huawei/cvms/sync", svc.SyncHuaWeiCvm)
	h.Add("SyncHuaWeiCvmWithRelResource", http.MethodPost, "/vendors/huawei/cvms/with/relation_resource/sync",
		svc.SyncHuaWeiCvmWithRelResource)

	h.Add("SyncAwsCvm", http.MethodPost, "/vendors/aws/cvms/sync", svc.SyncAwsCvm)
	h.Add("SyncAwsCvmWithRelResource", http.MethodPost, "/vendors/aws/cvms/with/relation_resource/sync",
		svc.SyncAwsCvmWithRelResource)

	h.Add("SyncAzureCvm", http.MethodPost, "/vendors/azure/cvms/sync", svc.SyncAzureCvm)
	h.Add("SyncAzureCvmWithRelResource", http.MethodPost, "/vendors/azure/cvms/with/relation_resource/sync",
		svc.SyncAzureCvmWithRelResource)

	h.Add("SyncGcpCvm", http.MethodPost, "/vendors/gcp/cvms/sync", svc.SyncGcpCvm)
	h.Add("SyncGcpCvmWithRelResource", http.MethodPost, "/vendors/gcp/cvms/with/relation_resource/sync",
		svc.SyncGcpCvmWithRelResource)

	h.Load(cap.WebService)
}

type syncCvmSvc struct {
	adaptor *cloudclient.CloudAdaptorClient
	dataCli *dataservice.Client
}
