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
	"testing"

	protolb "hcm/pkg/api/hc-service/load-balancer"
	"hcm/pkg/criteria/errf"
	"hcm/pkg/kit"
	cvt "hcm/pkg/tools/converter"

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

// requireBodyHasFields 断言请求体的过滤条件里包含所有给定字段。
func requireBodyHasFields(t *testing.T, body string, fields ...string) {
	t.Helper()
	for _, field := range fields {
		require.Contains(t, body, `"field":"`+field+`"`)
	}
}

// newOwnershipReq 构造归属校验用的创建请求，clusterTag 为空表示不指定标签。
func newOwnershipReq(clusterTag string, cloudClusterIDs []string) *protolb.TCloudLoadBalancerCreateReq {
	req := &protolb.TCloudLoadBalancerCreateReq{
		AccountID: "0000001",
		BkBizID:   213,
		Region:    "ap-guangzhou",
	}
	req.CloudClusterIDs = cloudClusterIDs
	if clusterTag != "" {
		req.ClusterTag = cvt.ValToPtr(clusterTag)
	}
	return req
}

// TestCheckExclusiveClusterOwnership_ClusterTagBelongsToBiz cluster_tag 归属当前业务，校验通过。
func TestCheckExclusiveClusterOwnership_ClusterTagBelongsToBiz(t *testing.T) {
	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		body := readBody(t, r)
		require.Contains(t, body, `"field":"cluster_tag"`)
		// 标签归属校验必须同时限定业务、账号和地域
		requireBodyHasFields(t, body, "bk_biz_id", "account_id", "region")
		writeOKResp(t, w, map[string]any{"count": 1, "details": []any{}})
	})
	cli := newTestDataServiceClient(t, handler)

	err := CheckExclusiveClusterOwnership(testKit(), cli, newOwnershipReq("ziyan-serven", nil))
	require.NoError(t, err)
}

// TestCheckExclusiveClusterOwnership_ClusterTagNotBelongToBiz cluster_tag 不属于当前业务，返回 PermissionDenied
// 且不泄露其它业务的集群清单。
func TestCheckExclusiveClusterOwnership_ClusterTagNotBelongToBiz(t *testing.T) {
	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		writeOKResp(t, w, map[string]any{"count": 0, "details": []any{}})
	})
	cli := newTestDataServiceClient(t, handler)

	err := CheckExclusiveClusterOwnership(testKit(), cli, newOwnershipReq("other-biz-tag", nil))
	require.Error(t, err)
	require.Equal(t, errf.PermissionDenied, err.(*errf.ErrorF).Code)
	require.NotContains(t, err.Error(), "other-biz-tag")
}

// TestCheckExclusiveClusterOwnership_AllClusterIDsBelongToBiz 全部四层集群 ID 都归属当前业务，校验通过。
func TestCheckExclusiveClusterOwnership_AllClusterIDsBelongToBiz(t *testing.T) {
	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		body := readBody(t, r)
		require.Contains(t, body, `"field":"cloud_id"`)
		// 四层集群归属校验必须同时限定业务、账号和地域
		requireBodyHasFields(t, body, "bk_biz_id", "account_id", "region")
		writeOKResp(t, w, map[string]any{
			"count": 2,
			"details": []map[string]any{
				{"cloud_id": "tgw-1"},
				{"cloud_id": "tgw-2"},
			},
		})
	})
	cli := newTestDataServiceClient(t, handler)

	err := CheckExclusiveClusterOwnership(testKit(), cli, newOwnershipReq("", []string{"tgw-1", "tgw-2"}))
	require.NoError(t, err)
}

// TestCheckExclusiveClusterOwnership_SomeClusterIDNotBelongToBiz 某个四层集群 ID 不属于当前业务时返回
// PermissionDenied，错误信息只列出不属于当前业务的那部分入参 ID，不含已归属的 ID，也不回显集群实际归属信息。
func TestCheckExclusiveClusterOwnership_SomeClusterIDNotBelongToBiz(t *testing.T) {
	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// 本地表只命中 tgw-1，tgw-2 属于业务 B，不应出现在响应中
		writeOKResp(t, w, map[string]any{
			"count":   1,
			"details": []map[string]any{{"cloud_id": "tgw-1"}},
		})
	})
	cli := newTestDataServiceClient(t, handler)

	err := CheckExclusiveClusterOwnership(testKit(), cli, newOwnershipReq("", []string{"tgw-1", "tgw-2"}))
	require.Error(t, err)
	require.Equal(t, errf.PermissionDenied, err.(*errf.ErrorF).Code)
	require.Contains(t, err.Error(), "tgw-2")
	require.NotContains(t, err.Error(), "tgw-1")
}

// TestCheckExclusiveClusterOwnership_BothEmpty 两个集群标识都为空时直接通过（结构校验已保证独占型下不会出现该组合）。
func TestCheckExclusiveClusterOwnership_BothEmpty(t *testing.T) {
	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		t.Fatal("should not query data service when both cluster_tag and cloud_cluster_ids are empty")
	})
	cli := newTestDataServiceClient(t, handler)

	err := CheckExclusiveClusterOwnership(testKit(), cli, newOwnershipReq("", nil))
	require.NoError(t, err)
}
