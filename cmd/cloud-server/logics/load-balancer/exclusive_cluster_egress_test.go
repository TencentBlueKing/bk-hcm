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

package lblogic

import (
	"net/http"
	"strings"
	"testing"

	"hcm/pkg/tools/slice"

	"github.com/stretchr/testify/require"
)

// newEgressMockHandler 按请求体中的 cluster_type 返回对应的四层/七层出口记录：TGW 请求返回 tgwEgresses，
// STGW 请求返回 stgwEgresses。
func newEgressMockHandler(t *testing.T, tgwEgresses, stgwEgresses []string) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		body := readBody(t, r)

		var egresses []string
		switch {
		case strings.Contains(body, `"TGW"`):
			egresses = tgwEgresses
		case strings.Contains(body, `"STGW"`):
			egresses = stgwEgresses
		default:
			t.Fatalf("unexpected filter without cluster_type: %s", body)
		}

		details := make([]map[string]any, 0, len(egresses))
		for _, e := range egresses {
			details = append(details, map[string]any{"egress": e})
		}
		writeOKResp(t, w, map[string]any{"count": uint64(len(details)), "details": details})
	}
}

// TestComputeExclusiveClusterEgressSet_OnlyL4NotMatch 只传四层，带宽包出口不在 TGW 出口集合内。
func TestComputeExclusiveClusterEgressSet_OnlyL4NotMatch(t *testing.T) {
	cli := newTestDataServiceClient(t, newEgressMockHandler(t, []string{"center_egress1"}, nil))

	allowed, err := ComputeExclusiveClusterEgressSet(testKit(), cli, 213, "", []string{"tgw-1"})
	require.NoError(t, err)
	require.False(t, slice.IsItemInSlice(allowed, "center_egress2"))
}

// TestComputeExclusiveClusterEgressSet_OnlyL4MultiEgressMatch 只传四层多个出口，带宽包出口命中其一。
func TestComputeExclusiveClusterEgressSet_OnlyL4MultiEgressMatch(t *testing.T) {
	cli := newTestDataServiceClient(t, newEgressMockHandler(t, []string{"center_egress1", "center_egress2"}, nil))

	allowed, err := ComputeExclusiveClusterEgressSet(testKit(), cli, 213, "", []string{"tgw-1", "tgw-2"})
	require.NoError(t, err)
	require.True(t, slice.IsItemInSlice(allowed, "center_egress1"))
}

// TestComputeExclusiveClusterEgressSet_OnlyL7NotMatch 只传七层，带宽包出口不在标签出口集合内。
func TestComputeExclusiveClusterEgressSet_OnlyL7NotMatch(t *testing.T) {
	cli := newTestDataServiceClient(t, newEgressMockHandler(t, nil, []string{"center_egress1"}))

	allowed, err := ComputeExclusiveClusterEgressSet(testKit(), cli, 213, "ziyan-serven", nil)
	require.NoError(t, err)
	require.False(t, slice.IsItemInSlice(allowed, "center_egress2"))
}

// TestComputeExclusiveClusterEgressSet_OnlyL7Match 只传七层，带宽包出口属于标签出口集合。
func TestComputeExclusiveClusterEgressSet_OnlyL7Match(t *testing.T) {
	cli := newTestDataServiceClient(t,
		newEgressMockHandler(t, nil, []string{"center_egress1", "center_egress2"}))

	allowed, err := ComputeExclusiveClusterEgressSet(testKit(), cli, 213, "ziyan-serven", nil)
	require.NoError(t, err)
	require.True(t, slice.IsItemInSlice(allowed, "center_egress1"))
}

// TestComputeExclusiveClusterEgressSet_BothWithinIntersection 四层七层都选，出口落在交集内。
func TestComputeExclusiveClusterEgressSet_BothWithinIntersection(t *testing.T) {
	cli := newTestDataServiceClient(t, newEgressMockHandler(t,
		[]string{"center_egress1", "center_egress2"},
		[]string{"center_egress1", "center_egress3"},
	))

	allowed, err := ComputeExclusiveClusterEgressSet(testKit(), cli, 213, "ziyan-serven", []string{"tgw-1"})
	require.NoError(t, err)
	require.True(t, slice.IsItemInSlice(allowed, "center_egress1"))
}

// TestComputeExclusiveClusterEgressSet_BothOnlyInL4 出口只属于四层出口集合，不在交集内。
func TestComputeExclusiveClusterEgressSet_BothOnlyInL4(t *testing.T) {
	cli := newTestDataServiceClient(t, newEgressMockHandler(t,
		[]string{"center_egress1", "center_egress2"},
		[]string{"center_egress3"},
	))

	allowed, err := ComputeExclusiveClusterEgressSet(testKit(), cli, 213, "ziyan-serven", []string{"tgw-1"})
	require.NoError(t, err)
	require.False(t, slice.IsItemInSlice(allowed, "center_egress1"))
}

// TestComputeExclusiveClusterEgressSet_BothOnlyInL7 出口只属于七层出口集合，不在交集内。
func TestComputeExclusiveClusterEgressSet_BothOnlyInL7(t *testing.T) {
	cli := newTestDataServiceClient(t, newEgressMockHandler(t,
		[]string{"center_egress2"},
		[]string{"center_egress1", "center_egress3"},
	))

	allowed, err := ComputeExclusiveClusterEgressSet(testKit(), cli, 213, "ziyan-serven", []string{"tgw-1"})
	require.NoError(t, err)
	require.False(t, slice.IsItemInSlice(allowed, "center_egress1"))
}

// TestComputeExclusiveClusterEgressSet_BothIntersectionEmpty 四层七层出口集合交集为空。
func TestComputeExclusiveClusterEgressSet_BothIntersectionEmpty(t *testing.T) {
	cli := newTestDataServiceClient(t, newEgressMockHandler(t,
		[]string{"center_egress1"},
		[]string{"center_egress2"},
	))

	allowed, err := ComputeExclusiveClusterEgressSet(testKit(), cli, 213, "ziyan-serven", []string{"tgw-1"})
	require.NoError(t, err)
	require.Empty(t, allowed)
	require.False(t, slice.IsItemInSlice(allowed, "center_egress1"))
	require.False(t, slice.IsItemInSlice(allowed, "center_egress2"))
}
