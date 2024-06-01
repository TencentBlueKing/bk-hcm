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

// Package cert ...
package cert

import (
	"hcm/cmd/hc-service/logics/cloud-adaptor"
	"hcm/cmd/hc-service/service/capability"
	"hcm/pkg/client"
	dataservice "hcm/pkg/client/data-service"
)

// InitCertService initial cert service.
func InitCertService(cap *capability.Capability) {
	svc := &certSvc{
		ad:      cap.CloudAdaptor,
		dataCli: cap.ClientSet.DataService(),
	}

	svc.initTCloudCertService(cap)
}

type certSvc struct {
	ad      *cloudadaptor.CloudAdaptorClient
	dataCli *dataservice.Client
	client  *client.ClientSet
}
