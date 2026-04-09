/*
 * TencentBlueKing is pleased to support the open source community by making
 * 蓝鲸智云 - 混合云管理平台 (BlueKing - Hybrid Cloud Management System) available.
 * Copyright (C) 2026 THL A29 Limited,
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

package cscvm

import (
	"testing"

	"hcm/pkg/criteria/enumor"

	"github.com/stretchr/testify/require"
)

func TestGetMonitorDataReqValidate_AwsRequiresUTCFields(t *testing.T) {
	req := &GetMonitorDataReq{
		MetricName: "LanOuttraffic",
		Period:     300,
		IDs:        []string{"id-1"},
	}

	err := req.Validate(enumor.Aws)
	require.Error(t, err)
	require.Contains(t, err.Error(), "start_time and end_time are required")
}

func TestGetMonitorDataReqValidate_AwsRejectsOtherVendorFields(t *testing.T) {
	req := &GetMonitorDataReq{
		MetricName: "LanOuttraffic",
		Period:     300,
		IDs:        []string{"id-1"},
		StartTime:  "2026-04-09T00:00:00Z",
		EndTime:    "2026-04-09T01:00:00Z",
		Namespace:  "SYS.ECS",
	}

	err := req.Validate(enumor.Aws)
	require.Error(t, err)
	require.Contains(t, err.Error(), "only supported for vendor huawei")
}

func TestGetMonitorDataReqValidate_TCloudRejectsAwsFields(t *testing.T) {
	req := &GetMonitorDataReq{
		MetricName: "CPUUsage",
		Period:     60,
		IDs:        []string{"id-1"},
		StartTime:  "2026-04-09 08:00:00",
		EndTime:    "2026-04-09 09:00:00",
	}

	err := req.Validate(enumor.TCloud)
	require.NoError(t, err)
}
