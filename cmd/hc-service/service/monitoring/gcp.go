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

package monitoring

import (
	"fmt"

	protomonitoring "hcm/pkg/api/hc-service/monitoring"
	"hcm/pkg/logs"
	"hcm/pkg/rest"
)

// GcpListTimeSeries lists GCP Cloud Monitoring time series data.
func (svc *monitoringSvc) GcpListTimeSeries(cts *rest.Contexts) (any, error) {
	req := new(protomonitoring.GcpListTimeSeriesReq)
	if err := cts.DecodeInto(req); err != nil {
		return nil, err
	}

	if err := req.Validate(); err != nil {
		return nil, err
	}

	// Get project id from main account
	mainAccountInfo, err := svc.cs.DataService().Gcp.MainAccount.Get(cts.Kit, req.MainAccountID)
	if err != nil {
		logs.Errorf("get gcp main account failed, main account id: %s, err: %+v, rid: %s",
			req.MainAccountID, err, cts.Kit.Rid)
		return nil, err
	}
	if mainAccountInfo.Extension == nil || mainAccountInfo.Extension.CloudProjectID == "" {
		return nil, fmt.Errorf("main account: %s cloud project id is empty", req.MainAccountID)
	}

	// Get GCP adaptor client using root account id
	gcpCli, err := svc.ad.GcpRoot(cts.Kit, req.RootAccountID)
	if err != nil {
		logs.Errorf("get gcp client failed, root account id: %s, err: %v, rid: %s",
			req.RootAccountID, err, cts.Kit.Rid)
		return nil, err
	}

	// Call GCP monitoring API with project id as parameter
	result, err := gcpCli.ListMonitorTimeSeries(cts.Kit, mainAccountInfo.Extension.CloudProjectID,
		&req.GcpListTimeSeriesOption)
	if err != nil {
		logs.Errorf("list gcp time series failed, main account id: %s, err: %v, rid: %s",
			req.MainAccountID, err, cts.Kit.Rid)
		return nil, err
	}

	return result, nil
}
