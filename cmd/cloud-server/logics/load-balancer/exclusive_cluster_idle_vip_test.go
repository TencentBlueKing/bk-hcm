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
	"net/http/httptest"
	"testing"

	hcservice "hcm/pkg/client/hc-service"
	"hcm/pkg/criteria/errf"
	"hcm/pkg/rest/client"

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

// newTestHCServiceClient 启动一个 httptest server 承载给定 handler，返回指向该 server 的 hc-service 客户端。
func newTestHCServiceClient(t *testing.T, handler http.Handler) *hcservice.Client {
	t.Helper()

	srv := httptest.NewServer(handler)
	t.Cleanup(srv.Close)

	cap := &client.Capability{Client: srv.Client(), Discover: staticServerDiscovery{addr: srv.URL}}
	return hcservice.NewClient(cap, "v1")
}

// TestCheckExclusiveClusterVipIdle_VipIdle 指定VIP闲置时云上返回该VIP，校验通过，且请求中带上了 vip 条件。
func TestCheckExclusiveClusterVipIdle_VipIdle(t *testing.T) {
	cli := newTestHCServiceClient(t, http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		require.Contains(t, readBody(t, r), `"vip":"1.1.1.2"`)
		writeOKResp(t, w, map[string]any{"count": 1, "details": []string{"1.1.1.2"}})
	}))

	err := CheckExclusiveClusterVipIdle(testKit(), cli, "acc-1", "ap-guangzhou", "tgw-1", "1.1.1.2")
	require.NoError(t, err)
}

// TestCheckExclusiveClusterVipIdle_VipNotIdle 指定VIP已被占用时云上不返回该VIP，返回 InvalidParameter。
func TestCheckExclusiveClusterVipIdle_VipNotIdle(t *testing.T) {
	cli := newTestHCServiceClient(t, http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		writeOKResp(t, w, map[string]any{"count": 0, "details": []string{}})
	}))

	err := CheckExclusiveClusterVipIdle(testKit(), cli, "acc-1", "ap-guangzhou", "tgw-1", "1.1.1.2")
	require.Error(t, err)
	require.Equal(t, errf.InvalidParameter, err.(*errf.ErrorF).Code)
}

// TestCheckExclusiveClusterVipIdle_DescribeFailed 查云失败时直接返回错误。
func TestCheckExclusiveClusterVipIdle_DescribeFailed(t *testing.T) {
	cli := newTestHCServiceClient(t, http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		_, err := w.Write([]byte(`{"code": 1, "message": "mock server error"}`))
		require.NoError(t, err)
	}))

	err := CheckExclusiveClusterVipIdle(testKit(), cli, "acc-1", "ap-guangzhou", "tgw-1", "1.1.1.2")
	require.Error(t, err)
}
