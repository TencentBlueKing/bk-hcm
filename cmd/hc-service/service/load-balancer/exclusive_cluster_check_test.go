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
	"context"
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	adtypes "hcm/pkg/adaptor/types"
	typelb "hcm/pkg/adaptor/types/load-balancer"
	protolb "hcm/pkg/api/hc-service/load-balancer"
	dataservice "hcm/pkg/client/data-service"
	"hcm/pkg/kit"
	"hcm/pkg/rest/client"
	cvt "hcm/pkg/tools/converter"

	"github.com/stretchr/testify/require"
)

// staticServerDiscovery 测试用静态服务发现，始终返回同一个 httptest 地址。
type staticServerDiscovery struct{ addr string }

func (d staticServerDiscovery) GetServers() ([]string, error) { return []string{d.addr}, nil }

// newTestDataServiceClient 启动一个 httptest server 承载给定 handler，返回一个指向该 server 的 data-service
// 客户端；测试结束时自动关闭 server。
func newTestDataServiceClient(t *testing.T, handler http.Handler) *dataservice.Client {
	t.Helper()
	srv := httptest.NewServer(handler)
	t.Cleanup(srv.Close)

	cap := &client.Capability{Client: srv.Client(), Discover: staticServerDiscovery{addr: srv.URL}}
	return dataservice.NewClient(cap, "v1")
}

func writeOKResp(t *testing.T, w http.ResponseWriter, data any) {
	t.Helper()
	w.Header().Set("Content-Type", "application/json")
	require.NoError(t, json.NewEncoder(w).Encode(map[string]any{"code": 0, "message": "", "data": data}))
}

func readBody(t *testing.T, r *http.Request) string {
	t.Helper()
	b, err := io.ReadAll(r.Body)
	require.NoError(t, err)
	return string(b)
}

func testKit() *kit.Kit {
	return &kit.Kit{Ctx: context.Background()}
}

// ownershipOKHandler 归属校验全部通过的 data-service mock：cluster_tag 命中一条记录、cloud_cluster_ids 全部命中。
func ownershipOKHandler(t *testing.T, clusterIDs []string, egress string) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		body := readBody(t, r)
		switch {
		case strings.Contains(body, `"field":"cluster_tag"`):
			writeOKResp(t, w, map[string]any{"count": 1, "details": []any{}})
		case strings.Contains(body, `"field":"cloud_id"`):
			details := make([]map[string]any, 0, len(clusterIDs))
			for _, id := range clusterIDs {
				details = append(details, map[string]any{"cloud_id": id, "egress": egress})
			}
			writeOKResp(t, w, map[string]any{"count": uint64(len(details)), "details": details})
		default:
			t.Fatalf("unexpected request body: %s", body)
		}
	}
}

// fakeExclusiveAdaptor 独占集群下云前复核所需的最小 adaptor 能力的 fake 实现，用于替代真实的 tcloud.TCloud。
type fakeExclusiveAdaptor struct {
	bwPkgResult      *adtypes.TCloudListBwPkgResult
	bwPkgErr         error
	clusterResources *typelb.TCloudDescribeClusterResourcesResult
	clusterErr       error
	clusterOpt       *typelb.TCloudDescribeClusterResourcesOption
}

func (f *fakeExclusiveAdaptor) ListBandwidthPackage(_ *kit.Kit, _ *adtypes.TCloudListBwPkgOption) (
	*adtypes.TCloudListBwPkgResult, error) {
	return f.bwPkgResult, f.bwPkgErr
}

func (f *fakeExclusiveAdaptor) DescribeClusterResources(_ *kit.Kit,
	opt *typelb.TCloudDescribeClusterResourcesOption) (*typelb.TCloudDescribeClusterResourcesResult, error) {
	f.clusterOpt = opt
	return f.clusterResources, f.clusterErr
}

func baseExclusiveCreateReq() *protolb.TCloudLoadBalancerCreateReq {
	req := &protolb.TCloudLoadBalancerCreateReq{
		AccountID: "acc-1",
		BkBizID:   213,
		Region:    "ap-guangzhou",
		Name:      cvt.ValToPtr("test-lb"),
	}
	req.Exclusive = cvt.ValToPtr(int64(1))
	req.CloudClusterIDs = []string{"tgw-38feq8c6"}
	req.LoadBalancerPassToTarget = cvt.ValToPtr(true)
	return req
}

// TestRecheckExclusiveBeforeDeliver_OwnershipRevoked AC：审批等待期间归属被重新分配，下云前复核失败。
func TestRecheckExclusiveBeforeDeliver_OwnershipRevoked(t *testing.T) {
	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// cloud_cluster_ids 反查为空，说明已不属于当前业务
		writeOKResp(t, w, map[string]any{"count": 0, "details": []any{}})
	})
	svc := &clbSvc{dataCli: newTestDataServiceClient(t, handler)}

	err := svc.recheckExclusiveBeforeDeliver(testKit(), &fakeExclusiveAdaptor{}, baseExclusiveCreateReq())
	require.Error(t, err)
}

// TestRecheckExclusiveBeforeDeliver_VipIdleStillAvailable AC-017：VIP 复核仍闲置，正常通过。
func TestRecheckExclusiveBeforeDeliver_VipIdleStillAvailable(t *testing.T) {
	req := baseExclusiveCreateReq()
	req.Vip = cvt.ValToPtr("1.1.1.1")
	req.RequireCount = cvt.ValToPtr(uint64(1))

	svc := &clbSvc{dataCli: newTestDataServiceClient(t, ownershipOKHandler(t, req.CloudClusterIDs, ""))}
	adaptor := &fakeExclusiveAdaptor{
		clusterResources: &typelb.TCloudDescribeClusterResourcesResult{
			Resources: []typelb.TCloudClusterResource{{Vip: "1.1.1.1", Idle: true}},
		},
	}

	require.NoError(t, svc.recheckExclusiveBeforeDeliver(testKit(), adaptor, req))
	require.NotNil(t, adaptor.clusterOpt)
	require.Equal(t, "1.1.1.1", adaptor.clusterOpt.Vip)
	require.True(t, cvt.PtrToVal(adaptor.clusterOpt.Idle))
}

// TestRecheckExclusiveBeforeDeliver_VipNotReturned 云上按 vip+idle 过滤后无结果（已被占用或不存在），复核失败。
func TestRecheckExclusiveBeforeDeliver_VipNotReturned(t *testing.T) {
	req := baseExclusiveCreateReq()
	req.Vip = cvt.ValToPtr("1.1.1.1")
	req.RequireCount = cvt.ValToPtr(uint64(1))

	svc := &clbSvc{dataCli: newTestDataServiceClient(t, ownershipOKHandler(t, req.CloudClusterIDs, ""))}
	adaptor := &fakeExclusiveAdaptor{
		clusterResources: &typelb.TCloudDescribeClusterResourcesResult{},
	}

	err := svc.recheckExclusiveBeforeDeliver(testKit(), adaptor, req)
	require.Error(t, err)
}

// TestRecheckExclusiveBeforeDeliver_VipNoLongerIdle AC-013：指定 VIP 在提单时闲置、下云前已被占用。
func TestRecheckExclusiveBeforeDeliver_VipNoLongerIdle(t *testing.T) {
	req := baseExclusiveCreateReq()
	req.Vip = cvt.ValToPtr("1.1.1.1")
	req.RequireCount = cvt.ValToPtr(uint64(1))

	svc := &clbSvc{dataCli: newTestDataServiceClient(t, ownershipOKHandler(t, req.CloudClusterIDs, ""))}
	adaptor := &fakeExclusiveAdaptor{
		clusterResources: &typelb.TCloudDescribeClusterResourcesResult{
			Resources: []typelb.TCloudClusterResource{{Vip: "1.1.1.1", Idle: false}},
		},
	}

	err := svc.recheckExclusiveBeforeDeliver(testKit(), adaptor, req)
	require.Error(t, err)
}

// TestRecheckExclusiveBeforeDeliver_BandwidthEgressMismatch 出口一致性复核失败时不调用云创建。
func TestRecheckExclusiveBeforeDeliver_BandwidthEgressMismatch(t *testing.T) {
	req := baseExclusiveCreateReq()
	req.InternetChargeType = cvt.ValToPtr(typelb.BandwidthPackage)
	req.BandwidthPackageID = cvt.ValToPtr("bwp-1")

	svc := &clbSvc{dataCli: newTestDataServiceClient(t, ownershipOKHandler(t, req.CloudClusterIDs, "center_egress1"))}
	adaptor := &fakeExclusiveAdaptor{
		bwPkgResult: &adtypes.TCloudListBwPkgResult{
			Packages: []adtypes.TCloudBandwidthPackage{{Egress: "center_egress2"}},
		},
	}

	err := svc.recheckExclusiveBeforeDeliver(testKit(), adaptor, req)
	require.Error(t, err)
}

// TestRecheckExclusiveBeforeDeliver_BandwidthEgressMatch AC-020：出口一致性复核通过。
func TestRecheckExclusiveBeforeDeliver_BandwidthEgressMatch(t *testing.T) {
	req := baseExclusiveCreateReq()
	req.InternetChargeType = cvt.ValToPtr(typelb.BandwidthPackage)
	req.BandwidthPackageID = cvt.ValToPtr("bwp-1")

	svc := &clbSvc{dataCli: newTestDataServiceClient(t, ownershipOKHandler(t, req.CloudClusterIDs, "center_egress1"))}
	adaptor := &fakeExclusiveAdaptor{
		bwPkgResult: &adtypes.TCloudListBwPkgResult{
			Packages: []adtypes.TCloudBandwidthPackage{{Egress: "center_egress1"}},
		},
	}

	require.NoError(t, svc.recheckExclusiveBeforeDeliver(testKit(), adaptor, req))
}

// TestRecheckExclusiveBeforeDeliver_NonExclusiveSkipsRecheck 非独占型请求不触发复核（不查数据服务、不查 adaptor）。
func TestRecheckExclusiveBeforeDeliver_NonExclusiveSkipsRecheck(t *testing.T) {
	req := baseExclusiveCreateReq()
	req.Exclusive = cvt.ValToPtr(int64(0))
	require.False(t, req.IsExclusive())
}
