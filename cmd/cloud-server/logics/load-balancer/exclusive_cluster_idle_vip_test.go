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

	"hcm/pkg/criteria/errf"

	"github.com/stretchr/testify/require"
)

// TestCheckExclusiveClusterIdleVipQueryable_NotFound 云上集群 ID 在本地表中查不到，返回 InvalidParameter。
func TestCheckExclusiveClusterIdleVipQueryable_NotFound(t *testing.T) {
	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		body := readBody(t, r)
		require.Contains(t, body, `"field":"cloud_id"`)
		writeOKResp(t, w, map[string]any{"count": 0, "details": []any{}})
	})
	cli := newTestDataServiceClient(t, handler)

	err := CheckExclusiveClusterIdleVipQueryable(testKit(), cli, 213, "tgw-not-exist")
	require.Error(t, err)
	require.Equal(t, errf.InvalidParameter, err.(*errf.ErrorF).Code)
}

// TestCheckExclusiveClusterIdleVipQueryable_NotTGWType 集群存在但类型非 TGW（如 STGW），返回 InvalidParameter。
func TestCheckExclusiveClusterIdleVipQueryable_NotTGWType(t *testing.T) {
	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		writeOKResp(t, w, map[string]any{
			"count":   1,
			"details": []map[string]any{{"bk_biz_id": 213, "cluster_type": "STGW"}},
		})
	})
	cli := newTestDataServiceClient(t, handler)

	err := CheckExclusiveClusterIdleVipQueryable(testKit(), cli, 213, "stgw-1")
	require.Error(t, err)
	require.Equal(t, errf.InvalidParameter, err.(*errf.ErrorF).Code)
}

// TestCheckExclusiveClusterIdleVipQueryable_NotBelongToBiz 集群存在、类型为TGW，但归属其它业务，返回
// PermissionDenied 且不在错误信息中回显该集群实际归属的业务 ID。
func TestCheckExclusiveClusterIdleVipQueryable_NotBelongToBiz(t *testing.T) {
	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		writeOKResp(t, w, map[string]any{
			"count":   1,
			"details": []map[string]any{{"bk_biz_id": 999, "cluster_type": "TGW"}},
		})
	})
	cli := newTestDataServiceClient(t, handler)

	err := CheckExclusiveClusterIdleVipQueryable(testKit(), cli, 213, "tgw-1")
	require.Error(t, err)
	require.Equal(t, errf.PermissionDenied, err.(*errf.ErrorF).Code)
	require.NotContains(t, err.Error(), "999")
}

// TestCheckExclusiveClusterIdleVipQueryable_Pass 集群存在、类型为TGW、且归属当前业务，校验通过。
func TestCheckExclusiveClusterIdleVipQueryable_Pass(t *testing.T) {
	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		writeOKResp(t, w, map[string]any{
			"count":   1,
			"details": []map[string]any{{"bk_biz_id": 213, "cluster_type": "TGW"}},
		})
	})
	cli := newTestDataServiceClient(t, handler)

	err := CheckExclusiveClusterIdleVipQueryable(testKit(), cli, 213, "tgw-1")
	require.NoError(t, err)
}

// TestCheckExclusiveClusterIdleVipQueryable_DataServiceError 底层 data-service 调用失败时错误直接透传。
func TestCheckExclusiveClusterIdleVipQueryable_DataServiceError(t *testing.T) {
	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{"code":1,"message":"mock server error","data":null}`))
	})
	cli := newTestDataServiceClient(t, handler)

	err := CheckExclusiveClusterIdleVipQueryable(testKit(), cli, 213, "tgw-1")
	require.Error(t, err)
	require.Contains(t, err.Error(), "mock server error")
}
