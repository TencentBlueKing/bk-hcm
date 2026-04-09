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

package hccvm

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestAzureMonitorDataReqValidate_Success(t *testing.T) {
	top := int32(10)
	req := &AzureMonitorDataReq{
		AccountID:    "acc-1",
		Region:       "eastus",
		MetricName:   "LanOuttraffic",
		Period:       300,
		StartTime:    "2026-04-09T00:00:00Z",
		EndTime:      "2026-04-09T01:00:00Z",
		Top:          &top,
		OrderBy:      "total desc",
		ResultType:   "Data",
		InstanceIDs:  []string{"/subscriptions/s1/resourcegroups/rg1/providers/microsoft.compute/virtualmachines/vm1"},
	}

	err := req.Validate()
	require.NoError(t, err)
}

func TestAzureMonitorDataReqValidate_OrderByRequiresTop(t *testing.T) {
	req := &AzureMonitorDataReq{
		AccountID:    "acc-1",
		Region:       "eastus",
		MetricName:   "LanOuttraffic",
		Period:       300,
		StartTime:    "2026-04-09T00:00:00Z",
		EndTime:      "2026-04-09T01:00:00Z",
		OrderBy:      "total desc",
		InstanceIDs:  []string{"/subscriptions/s1/resourcegroups/rg1/providers/microsoft.compute/virtualmachines/vm1"},
	}

	err := req.Validate()
	require.Error(t, err)
	require.Contains(t, err.Error(), "orderby requires top")
}
