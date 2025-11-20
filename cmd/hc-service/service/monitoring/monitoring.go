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

// Package monitoring provides monitoring service for querying cloud metrics.
package monitoring

import (
	"net/http"

	"hcm/cmd/hc-service/logics/cloud-adaptor"
	"hcm/cmd/hc-service/service/capability"
	"hcm/pkg/client"
	"hcm/pkg/rest"
)

// InitMonitoringService initialize the monitoring service.
func InitMonitoringService(cap *capability.Capability) {
	svc := &monitoringSvc{
		ad: cap.CloudAdaptor,
		cs: cap.ClientSet,
	}

	h := rest.NewHandler()

	h.Add("GcpListTimeSeries", http.MethodPost, "/vendors/gcp/monitoring/time_series/list", svc.GcpListTimeSeries)

	h.Load(cap.WebService)
}

type monitoringSvc struct {
	ad *cloudadaptor.CloudAdaptorClient
	cs *client.ClientSet
}
