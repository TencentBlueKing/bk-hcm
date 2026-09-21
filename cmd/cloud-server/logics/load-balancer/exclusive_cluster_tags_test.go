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
	"testing"

	"hcm/pkg/dal/dao/tools"
	"hcm/pkg/runtime/filter"

	"github.com/stretchr/testify/require"
)

// TestAggregateExclusiveClusterTags_GroupsByTagAndType 正常场景：多条集群记录按 (cluster_tag, cluster_type)
// 分组聚合，同 tag 不同 type 拆分为不同分组，分组内保留出现顺序。
func TestAggregateExclusiveClusterTags_GroupsByTagAndType(t *testing.T) {
	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		writeOKResp(t, w, map[string]any{
			"count": 3,
			"details": []map[string]any{
				{
					"id": "id-1", "cloud_id": "stgw-1", "name": "cluster-1", "cluster_tag": "tagA",
					"cluster_type": "STGW", "egress": "1.1.1.1", "isp": "BGP",
					"extension": map[string]any{
						"clusters_zone": map[string]any{
							"master_zone": []string{"ap-guangzhou-1"},
							"slave_zone":  []string{"ap-guangzhou-2"},
						},
					},
				},
				{
					"id": "id-2", "cloud_id": "stgw-2", "name": "cluster-2", "cluster_tag": "tagA",
					"cluster_type": "STGW", "egress": "1.1.1.2", "isp": "BGP",
					"extension": map[string]any{
						"clusters_zone": map[string]any{"master_zone": []string{"ap-guangzhou-3"}},
					},
				},
				{
					"id": "id-3", "cloud_id": "tgw-1", "name": "cluster-3", "cluster_tag": "tagA",
					"cluster_type": "TGW", "egress": "1.1.1.3", "isp": "BGP",
					"extension": map[string]any{},
				},
			},
		})
	})
	cli := newTestDataServiceClient(t, handler)

	result, err := AggregateExclusiveClusterTags(testKit(), cli, tools.ExpressionAnd(
		tools.RuleEqual("account_id", "acc-1")))
	require.NoError(t, err)
	require.Len(t, result.Details, 2)

	require.Equal(t, "tagA", result.Details[0].ClusterTag)
	require.EqualValues(t, "STGW", result.Details[0].ClusterType)
	require.Len(t, result.Details[0].Clusters, 2)
	require.Equal(t, "stgw-1", result.Details[0].Clusters[0].CloudClusterID)
	require.Equal(t, []string{"ap-guangzhou-1"}, result.Details[0].Clusters[0].ClusterZone.MasterZone)
	require.Equal(t, []string{"ap-guangzhou-2"}, result.Details[0].Clusters[0].ClusterZone.SlaveZone)
	require.Equal(t, "stgw-2", result.Details[0].Clusters[1].CloudClusterID)
	require.Equal(t, []string{"ap-guangzhou-3"}, result.Details[0].Clusters[1].ClusterZone.MasterZone)

	require.Equal(t, "tagA", result.Details[1].ClusterTag)
	require.EqualValues(t, "TGW", result.Details[1].ClusterType)
	require.Len(t, result.Details[1].Clusters, 1)
	require.Equal(t, "tgw-1", result.Details[1].Clusters[0].CloudClusterID)
}

// TestAggregateExclusiveClusterTags_EmptyResult 空结果场景：data-service 未返回任何候选集群，聚合结果为空数组
// 而非 nil。
func TestAggregateExclusiveClusterTags_EmptyResult(t *testing.T) {
	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		writeOKResp(t, w, map[string]any{"count": 0, "details": []any{}})
	})
	cli := newTestDataServiceClient(t, handler)

	result, err := AggregateExclusiveClusterTags(testKit(), cli, tools.ExpressionAnd(
		tools.RuleEqual("account_id", "acc-1")))
	require.NoError(t, err)
	require.NotNil(t, result.Details)
	require.Len(t, result.Details, 0)
}

// TestAggregateExclusiveClusterTags_SkipsEmptyClusterTag 集群标签为空的记录不参与聚合（R-002）。
func TestAggregateExclusiveClusterTags_SkipsEmptyClusterTag(t *testing.T) {
	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		writeOKResp(t, w, map[string]any{
			"count": 2,
			"details": []map[string]any{
				{
					"id": "id-1", "cloud_id": "tgw-1", "name": "cluster-1", "cluster_tag": "",
					"cluster_type": "TGW", "extension": map[string]any{},
				},
				{
					"id": "id-2", "cloud_id": "stgw-1", "name": "cluster-2", "cluster_tag": "tagA",
					"cluster_type": "STGW", "extension": map[string]any{},
				},
			},
		})
	})
	cli := newTestDataServiceClient(t, handler)

	result, err := AggregateExclusiveClusterTags(testKit(), cli, tools.ExpressionAnd(
		tools.RuleEqual("account_id", "acc-1")))
	require.NoError(t, err)
	require.Len(t, result.Details, 1)
	require.Equal(t, "tagA", result.Details[0].ClusterTag)
}

// TestAggregateExclusiveClusterTags_ForwardsCallerFilter 聚合函数原样转发调用方传入的过滤条件。
// 其中已包含 handler.ListBizAuthRes 拼装的 bk_biz_id 归属条件与业务过滤条件。
// 本函数不重复拼装、也不放宽这些条件，从而保证跨业务数据不会被聚合进结果。
func TestAggregateExclusiveClusterTags_ForwardsCallerFilter(t *testing.T) {
	var gotBody string
	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		gotBody = readBody(t, r)
		writeOKResp(t, w, map[string]any{"count": 0, "details": []any{}})
	})
	cli := newTestDataServiceClient(t, handler)

	callerFilter := tools.ExpressionAnd(
		tools.RuleEqual("bk_biz_id", int64(213)),
		tools.RuleEqual("cluster_type", "TGW"),
	)
	_, err := AggregateExclusiveClusterTags(testKit(), cli, callerFilter)
	require.NoError(t, err)
	require.Contains(t, gotBody, `"field":"bk_biz_id"`)
	require.Contains(t, gotBody, `"field":"cluster_type"`)
}

// TestAggregateExclusiveClusterTags_DataServiceError 底层 data-service 调用失败时错误直接透传。
func TestAggregateExclusiveClusterTags_DataServiceError(t *testing.T) {
	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{"code":1,"message":"mock server error","data":null}`))
	})
	cli := newTestDataServiceClient(t, handler)

	_, err := AggregateExclusiveClusterTags(testKit(), cli, tools.ExpressionAnd(
		tools.RuleEqual("account_id", "acc-1")))
	require.Error(t, err)
	require.Contains(t, err.Error(), "mock server error")
}

// TestBuildExclusiveClusterZoneRules 覆盖单可用区走顶层 zone、
// 主备/多主可用区走 extension JSON 数组的 DB 过滤口径。
func TestBuildExclusiveClusterZoneRules(t *testing.T) {
	testCases := []struct {
		name      string
		zones     []string
		backZones []string
		want      []*filter.AtomRule
	}{
		{
			name: "empty zones and back_zones returns no rule",
		},
		{
			name:  "single zone without back_zones filters top-level zone",
			zones: []string{"ap-guangzhou-1"},
			want:  []*filter.AtomRule{tools.RuleEqual("zone", "ap-guangzhou-1")},
		},
		{
			name:  "two zones without back_zones filters extension master_zone",
			zones: []string{"ap-guangzhou-1", "ap-guangzhou-2"},
			want: []*filter.AtomRule{
				tools.RuleJSONContains(exclusiveClusterMasterZoneJSONField, "ap-guangzhou-1"),
				tools.RuleJSONContains(exclusiveClusterMasterZoneJSONField, "ap-guangzhou-2"),
			},
		},
		{
			name:      "single zone with back_zones filters extension master and slave",
			zones:     []string{"ap-guangzhou-1"},
			backZones: []string{"ap-guangzhou-2"},
			want: []*filter.AtomRule{
				tools.RuleJSONContains(exclusiveClusterMasterZoneJSONField, "ap-guangzhou-1"),
				tools.RuleJSONContains(exclusiveClusterSlaveZoneJSONField, "ap-guangzhou-2"),
			},
		},
		{
			name:      "only back_zones filters extension slave_zone",
			backZones: []string{"ap-guangzhou-2"},
			want: []*filter.AtomRule{
				tools.RuleJSONContains(exclusiveClusterSlaveZoneJSONField, "ap-guangzhou-2"),
			},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			require.Equal(t, tc.want, BuildExclusiveClusterZoneRules(tc.zones, tc.backZones))
		})
	}
}

// TestAggregateExclusiveClusterTags_InvalidExtension extension 无法反序列化时返回错误，
// 而不是静默丢弃数据。
func TestAggregateExclusiveClusterTags_InvalidExtension(t *testing.T) {
	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		writeOKResp(t, w, map[string]any{
			"count": 1,
			"details": []map[string]any{
				{
					"id": "id-1", "cloud_id": "stgw-1", "name": "cluster-1", "cluster_tag": "tagA",
					"cluster_type": "STGW", "extension": "not-a-json-object",
				},
			},
		})
	})
	cli := newTestDataServiceClient(t, handler)

	_, err := AggregateExclusiveClusterTags(testKit(), cli, tools.ExpressionAnd(
		tools.RuleEqual("account_id", "acc-1")))
	require.Error(t, err)
}
