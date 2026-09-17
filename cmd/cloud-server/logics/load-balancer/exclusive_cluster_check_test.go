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
	"context"
	"io"
	"net/http"
	"strings"
	"testing"

	"hcm/pkg/criteria/errf"
	"hcm/pkg/kit"

	"github.com/stretchr/testify/require"
)

// readBody 读取请求体为字符串，供测试 handler 判断本次请求查询的是 cluster_tag 还是 cloud_id。
func readBody(t *testing.T, r *http.Request) string {
	t.Helper()
	b, err := io.ReadAll(r.Body)
	require.NoError(t, err)
	return string(b)
}

func testKit() *kit.Kit {
	return &kit.Kit{Ctx: context.Background()}
}

// TestCheckExclusiveClusterOwnership_ClusterTagBelongsToBiz cluster_tag 归属当前业务，校验通过（AC-010 反例）。
func TestCheckExclusiveClusterOwnership_ClusterTagBelongsToBiz(t *testing.T) {
	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		body := readBody(t, r)
		require.Contains(t, body, `"field":"cluster_tag"`)
		writeOKResp(t, w, map[string]any{"count": 1, "details": []any{}})
	})
	cli := newTestDataServiceClient(t, handler)

	err := CheckExclusiveClusterOwnership(testKit(), cli, 213, "ziyan-serven", nil)
	require.NoError(t, err)
}

// TestCheckExclusiveClusterOwnership_ClusterTagNotBelongToBiz cluster_tag 不属于当前业务，返回 PermissionDenied
// 且不泄露其它业务的集群清单（AC-010）。
func TestCheckExclusiveClusterOwnership_ClusterTagNotBelongToBiz(t *testing.T) {
	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		writeOKResp(t, w, map[string]any{"count": 0, "details": []any{}})
	})
	cli := newTestDataServiceClient(t, handler)

	err := CheckExclusiveClusterOwnership(testKit(), cli, 213, "other-biz-tag", nil)
	require.Error(t, err)
	require.Equal(t, errf.PermissionDenied, err.(*errf.ErrorF).Code)
	require.NotContains(t, err.Error(), "other-biz-tag")
}

// TestCheckExclusiveClusterOwnership_AllClusterIDsBelongToBiz 全部四层集群 ID 都归属当前业务，校验通过。
func TestCheckExclusiveClusterOwnership_AllClusterIDsBelongToBiz(t *testing.T) {
	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		body := readBody(t, r)
		require.Contains(t, body, `"field":"cloud_id"`)
		writeOKResp(t, w, map[string]any{
			"count": 2,
			"details": []map[string]any{
				{"cloud_id": "tgw-1"},
				{"cloud_id": "tgw-2"},
			},
		})
	})
	cli := newTestDataServiceClient(t, handler)

	err := CheckExclusiveClusterOwnership(testKit(), cli, 213, "", []string{"tgw-1", "tgw-2"})
	require.NoError(t, err)
}

// TestCheckExclusiveClusterOwnership_SomeClusterIDNotBelongToBiz 某个四层集群 ID 不属于当前业务时返回
// PermissionDenied，越权访问不泄露其它业务集群信息（AC-010、越权场景）。
func TestCheckExclusiveClusterOwnership_SomeClusterIDNotBelongToBiz(t *testing.T) {
	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// 本地表只命中 tgw-1，tgw-2 属于业务 B，不应出现在响应中
		writeOKResp(t, w, map[string]any{
			"count":   1,
			"details": []map[string]any{{"cloud_id": "tgw-1"}},
		})
	})
	cli := newTestDataServiceClient(t, handler)

	err := CheckExclusiveClusterOwnership(testKit(), cli, 213, "", []string{"tgw-1", "tgw-2"})
	require.Error(t, err)
	require.Equal(t, errf.PermissionDenied, err.(*errf.ErrorF).Code)
	require.False(t, strings.Contains(err.Error(), "tgw-2"),
		"error message must not leak other biz's cluster ids")
}

// TestCheckExclusiveClusterOwnership_BothEmpty 两个集群标识都为空时直接通过（结构校验已保证独占型下不会出现该组合）。
func TestCheckExclusiveClusterOwnership_BothEmpty(t *testing.T) {
	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		t.Fatal("should not query data service when both cluster_tag and cloud_cluster_ids are empty")
	})
	cli := newTestDataServiceClient(t, handler)

	err := CheckExclusiveClusterOwnership(testKit(), cli, 213, "", nil)
	require.NoError(t, err)
}
