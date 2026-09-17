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

package loadbalancer

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	cslb "hcm/pkg/api/cloud-server/load-balancer"
	"hcm/pkg/cc"
	"hcm/pkg/client"
	"hcm/pkg/iam/auth"
	"hcm/pkg/iam/meta"
	"hcm/pkg/kit"
	"hcm/pkg/rest"
	"hcm/pkg/runtime/filter"
	"hcm/pkg/serviced"

	"github.com/emicklei/go-restful/v3"
	"github.com/stretchr/testify/require"
)

// fakeAuthorizer 测试用鉴权器：Authorize 的结果由 authorized/err 两个字段固定控制。
type fakeAuthorizer struct {
	authorized bool
	err        error
}

func (f *fakeAuthorizer) Authorize(_ *kit.Kit, resources ...meta.ResourceAttribute) (
	[]meta.Decision, bool, error) {
	if f.err != nil {
		return nil, false, f.err
	}
	decisions := make([]meta.Decision, len(resources))
	for i := range decisions {
		decisions[i] = meta.Decision{Authorized: f.authorized}
	}
	return decisions, f.authorized, nil
}

func (f *fakeAuthorizer) AuthorizeAny(_ *kit.Kit, _ ...meta.ResourceAttribute) ([]meta.Decision, error) {
	return nil, nil
}

func (f *fakeAuthorizer) AuthorizeWithPerm(_ *kit.Kit, _ ...meta.ResourceAttribute) error { return nil }

func (f *fakeAuthorizer) ListAuthorizedInstances(_ *kit.Kit, _ *meta.ListAuthResInput) (
	*meta.AuthorizedInstances, error) {
	return nil, nil
}

func (f *fakeAuthorizer) ListAuthInstWithFilter(_ *kit.Kit, _ *meta.ListAuthResInput, _ *filter.Expression,
	_ string) (*filter.Expression, bool, error) {
	return nil, false, nil
}

func (f *fakeAuthorizer) RegisterResourceCreatorAction(_ *kit.Kit, _ *meta.RegisterResCreatorActionInst) error {
	return nil
}

func (f *fakeAuthorizer) GetPermissionToApply(_ *kit.Kit, _ ...meta.ResourceAttribute) (
	*meta.IamPermission, error) {
	return nil, nil
}

func (f *fakeAuthorizer) GetApplyPermUrl(_ *kit.Kit, _ *meta.IamPermission) (string, error) {
	return "", nil
}

var _ auth.Authorizer = new(fakeAuthorizer)

// newTestLbSvc 构造一个 client 指向 srv 的 lbSvc，用于测试需要真实发起下游 HTTP 调用的 handler。
func newTestLbSvc(t *testing.T, srv *httptest.Server, authorizer auth.Authorizer) *lbSvc {
	t.Helper()
	cs := client.NewClientSet(srv.Client(), fakeSimpleDiscover{addr: srv.URL})
	return &lbSvc{client: cs, authorizer: authorizer}
}

// fakeSimpleDiscover 测试用服务发现：对任意服务名都返回同一个 httptest 地址，用于在单个 mock server 上同时
// 承接 data-service 与 hc-service 两类下游调用。
type fakeSimpleDiscover struct{ addr string }

func (d fakeSimpleDiscover) Discover(cc.Name) ([]string, error) { return []string{d.addr}, nil }
func (d fakeSimpleDiscover) Services() []cc.Name {
	return []cc.Name{cc.DataServiceName, cc.HCServiceName}
}
func (d fakeSimpleDiscover) ByLabels([]string) serviced.Discover             { return d }
func (d fakeSimpleDiscover) GetServiceAllNodeKeys(cc.Name) ([]string, error) { return nil, nil }

// newBizContext 构造一个携带 body 与 bk_biz_id 路径参数的 rest.Contexts，用于调用 biz 视角 handler。
func newBizContext(t *testing.T, bkBizID string, body any) *rest.Contexts {
	t.Helper()
	raw, err := json.Marshal(body)
	require.NoError(t, err)

	httpReq, err := http.NewRequest(http.MethodPost, "http://test/bizs/"+bkBizID+"/load_balancers/x",
		bytes.NewReader(raw))
	require.NoError(t, err)

	req := restful.NewRequest(httpReq)
	req.PathParameters()["bk_biz_id"] = bkBizID

	return &rest.Contexts{Kit: kit.New(), Request: req}
}

func writeOKRespCS(t *testing.T, w http.ResponseWriter, data any) {
	t.Helper()
	w.Header().Set("Content-Type", "application/json")
	require.NoError(t, json.NewEncoder(w).Encode(map[string]any{"code": 0, "message": "", "data": data}))
}

// TestListBizExclusiveClusterTags_InvalidParameter 必填参数缺失（isp 未传）时直接返回 InvalidParameter，
// 不发起任何下游调用。
func TestListBizExclusiveClusterTags_InvalidParameter(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		t.Fatalf("should not call downstream, got request: %s", r.URL.Path)
	}))
	t.Cleanup(srv.Close)

	svc := newTestLbSvc(t, srv, &fakeAuthorizer{authorized: true})
	cts := newBizContext(t, "213", map[string]any{"account_id": "acc-1", "region": "ap-guangzhou"})

	_, err := svc.ListBizExclusiveClusterTags(cts)
	require.Error(t, err)
}

// TestListBizExclusiveClusterTags_NoPermission 鉴权失败（无权限）时返回空结果，不报错、不泄露信息。
func TestListBizExclusiveClusterTags_NoPermission(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		t.Fatalf("should not query data service when unauthorized, got request: %s", r.URL.Path)
	}))
	t.Cleanup(srv.Close)

	svc := newTestLbSvc(t, srv, &fakeAuthorizer{authorized: false})
	cts := newBizContext(t, "213", map[string]any{"account_id": "acc-1", "region": "ap-guangzhou", "isp": "BGP"})

	result, err := svc.ListBizExclusiveClusterTags(cts)
	require.NoError(t, err)
	tagsResult, ok := result.(*cslb.ListExclusiveClusterTagsResult)
	require.True(t, ok)
	require.Len(t, tagsResult.Details, 0)
}

// TestListBizExclusiveClusterTags_RequestBodyIgnoresBkBizID 请求体结构体不包含 bk_biz_id 字段，bk_biz_id 只能
// 来自路径参数，避免越权指定业务。
func TestListBizExclusiveClusterTags_RequestBodyIgnoresBkBizID(t *testing.T) {
	raw, err := json.Marshal(cslb.ListExclusiveClusterTagsReq{AccountID: "acc-1", Region: "ap-guangzhou", Isp: "BGP"})
	require.NoError(t, err)
	require.False(t, strings.Contains(string(raw), "bk_biz_id"),
		"ListExclusiveClusterTagsReq must not contain bk_biz_id field")
}

// TestListBizExclusiveClusterIdleVips_InvalidParameter 必填参数缺失（cloud_cluster_id 未传）时直接返回
// InvalidParameter，不发起任何下游调用。
func TestListBizExclusiveClusterIdleVips_InvalidParameter(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		t.Fatalf("should not call downstream, got request: %s", r.URL.Path)
	}))
	t.Cleanup(srv.Close)

	svc := newTestLbSvc(t, srv, &fakeAuthorizer{authorized: true})
	cts := newBizContext(t, "213", map[string]any{"account_id": "acc-1", "region": "ap-guangzhou"})

	_, err := svc.ListBizExclusiveClusterIdleVips(cts)
	require.Error(t, err)
}

// TestListBizExclusiveClusterIdleVips_NoPermission 鉴权失败时直接拒绝，不查本地表、不发起云调用。
func TestListBizExclusiveClusterIdleVips_NoPermission(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		t.Fatalf("should not call downstream when unauthorized, got request: %s", r.URL.Path)
	}))
	t.Cleanup(srv.Close)

	svc := newTestLbSvc(t, srv, &fakeAuthorizer{authorized: false})
	cts := newBizContext(t, "213", map[string]any{
		"account_id": "acc-1", "region": "ap-guangzhou", "cloud_cluster_id": "tgw-1",
	})

	_, err := svc.ListBizExclusiveClusterIdleVips(cts)
	require.Error(t, err)
}

// TestListBizExclusiveClusterIdleVips_OwnershipRejected_NoCloudCall 集群不归属当前业务时拒绝，且不发起
// hc-service/云调用（越权拒绝且不发起云调用）。
func TestListBizExclusiveClusterIdleVips_OwnershipRejected_NoCloudCall(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if strings.Contains(r.URL.Path, "idle_vips/describe") {
			t.Fatalf("should not call hc-service when ownership check fails")
		}
		// ownership check: cluster belongs to a different biz.
		writeOKRespCS(t, w, map[string]any{
			"count":   1,
			"details": []map[string]any{{"bk_biz_id": 999, "cluster_type": "TGW"}},
		})
	}))
	t.Cleanup(srv.Close)

	svc := newTestLbSvc(t, srv, &fakeAuthorizer{authorized: true})
	cts := newBizContext(t, "213", map[string]any{
		"account_id": "acc-1", "region": "ap-guangzhou", "cloud_cluster_id": "tgw-1",
	})

	_, err := svc.ListBizExclusiveClusterIdleVips(cts)
	require.Error(t, err)
}

// TestListBizExclusiveClusterIdleVips_Success 正常查询：归属校验通过后调用 hc-service 拿到闲置 VIP 列表。
func TestListBizExclusiveClusterIdleVips_Success(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if strings.Contains(r.URL.Path, "idle_vips/describe") {
			writeOKRespCS(t, w, map[string]any{"count": 2, "details": []string{"1.1.1.1", "1.1.1.2"}})
			return
		}
		writeOKRespCS(t, w, map[string]any{
			"count":   1,
			"details": []map[string]any{{"bk_biz_id": 213, "cluster_type": "TGW"}},
		})
	}))
	t.Cleanup(srv.Close)

	svc := newTestLbSvc(t, srv, &fakeAuthorizer{authorized: true})
	cts := newBizContext(t, "213", map[string]any{
		"account_id": "acc-1", "region": "ap-guangzhou", "cloud_cluster_id": "tgw-1",
	})

	result, err := svc.ListBizExclusiveClusterIdleVips(cts)
	require.NoError(t, err)
	vipsResult, ok := result.(*cslb.ListExclusiveClusterIdleVipsResult)
	require.True(t, ok)
	require.EqualValues(t, 2, vipsResult.Count)
	require.ElementsMatch(t, []string{"1.1.1.1", "1.1.1.2"}, vipsResult.Details)
}
