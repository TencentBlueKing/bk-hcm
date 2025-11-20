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

	csmonitoring "hcm/pkg/api/cloud-server/monitoring"
	protomonitoring "hcm/pkg/api/hc-service/monitoring"
	"hcm/pkg/criteria/errf"
	"hcm/pkg/iam/meta"
	"hcm/pkg/logs"
	"hcm/pkg/rest"
)

// GcpListTimeSeries lists GCP Cloud Monitoring time series data.
func (svc *monitoringSvc) GcpListTimeSeries(cts *rest.Contexts) (any, error) {
	req := new(csmonitoring.GcpListTimeSeriesReq)
	if err := cts.DecodeInto(req); err != nil {
		logs.Errorf("decode gcp list time series request failed, err: %v, rid: %s", err, cts.Kit.Rid)
		return nil, err
	}

	if err := req.Validate(); err != nil {
		logs.Errorf("validate gcp list time series request failed, req: %+v, err: %v, rid: %s", req, err, cts.Kit.Rid)
		return nil, err
	}

	// Check permission for monitoring data access
	if err := svc.checkPermission(cts, meta.Monitoring, meta.Find); err != nil {
		logs.Errorf("check permission failed, user: %s, err: %v, rid: %s", cts.Kit.User, err, cts.Kit.Rid)
		return nil, err
	}

	// Convert to hc-service request
	hcReq := &protomonitoring.GcpListTimeSeriesReq{
		RootAccountID:           req.RootAccountID,
		MainAccountID:           req.MainAccountID,
		GcpListTimeSeriesOption: req.GcpListTimeSeriesOption,
	}

	// Call hc-service
	result, err := svc.client.HCService().Gcp.Monitoring.ListTimeSeries(cts.Kit.Ctx, cts.Kit.Header(), hcReq)
	if err != nil {
		logs.Errorf("list gcp time series failed, main account id: %s, err: %v, rid: %s",
			req.MainAccountID, err, cts.Kit.Rid)
		return nil, err
	}

	return result, nil
}

func (svc *monitoringSvc) checkPermission(cts *rest.Contexts, resType meta.ResourceType, action meta.Action) error {
	resources := make([]meta.ResourceAttribute, 0)
	resources = append(resources, meta.ResourceAttribute{
		Basic: &meta.Basic{
			Type:   resType,
			Action: action,
		},
	})

	_, authorized, err := svc.authorizer.Authorize(cts.Kit, resources...)
	if err != nil {
		return errf.NewFromErr(
			errf.PermissionDenied,
			fmt.Errorf("check %s permissions failed, err: %v", action, err),
		)
	}

	if !authorized {
		return errf.NewFromErr(errf.PermissionDenied, fmt.Errorf("you have not permission of %s", action))
	}

	return nil
}
