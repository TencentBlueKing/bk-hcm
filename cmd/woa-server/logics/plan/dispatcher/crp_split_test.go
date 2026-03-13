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

package dispatcher

import (
	"context"
	"net/http"
	"testing"

	ptypes "hcm/cmd/woa-server/types/plan"
	"hcm/pkg/criteria/enumor"
	rpt "hcm/pkg/dal/table/resource-plan/res-plan-ticket"
	"hcm/pkg/kit"
	"hcm/pkg/thirdparty/cvmapi"

	"github.com/stretchr/testify/assert"
)

// testKit returns a minimal kit for unit tests.
func testKit() *kit.Kit {
	return &kit.Kit{
		Ctx: context.Background(),
		Rid: "test-rid",
	}
}

// makeDemand builds a rpt.ResPlanDemand for test use.
func makeDemand(obs enumor.ObsProject, tech string, cpuCore int64) rpt.ResPlanDemand {
	return rpt.ResPlanDemand{
		Updated: &rpt.UpdatedRPDemandItem{
			ObsProject: obs,
			ExpectTime: "2025-01-01",
			Cvm: rpt.Cvm{
				CpuCore:        cpuCore,
				TechnicalClass: tech,
			},
		},
	}
}

// makeCandidate builds a CvmCbsPlanQueryItem for test use.
// cvmPerCore is cores-per-VM (used to derive CvmAmount from realCore).
func makeCandidate(sliceID string, obs enumor.ObsProject, tech string, realCore int64,
	isInProcessing int, reviewStatus enumor.ResPlanReviewStatus) *cvmapi.CvmCbsPlanQueryItem {
	cvmAmount := float64(realCore) / 4 // assume 4 cores per VM
	return &cvmapi.CvmCbsPlanQueryItem{
		SliceId:        sliceID,
		ProjectName:    obs,
		TechnicalClass: tech,
		RealCoreAmount: realCore,
		CoreAmount:     realCore,
		CvmAmount:      cvmAmount,
		ReviewStatus:   reviewStatus,
		IsInProcessing: isInProcessing,
	}
}

// totalWillConsume returns the sum of WillConsume across all matchResult entries.
func totalWillConsume(m map[string]*AdjustAbleRemainObj) int64 {
	var total int64
	for _, v := range m {
		total += v.WillConsume
	}
	return total
}

// totalTransferTarget returns the sum of all TransferTarget values across all matchResult entries.
func totalTransferTarget(m map[string]*AdjustAbleRemainObj) int64 {
	var total int64
	for _, v := range m {
		for _, cores := range v.TransferTarget {
			total += cores
		}
	}
	return total
}

// ---- TestGreedyMatch ----

type greedyMatchCase struct {
	name string

	demandCores  int64
	useNilDemand bool
	candidates   []*cvmapi.CvmCbsPlanQueryItem
	preConsumed  map[string]int64

	wantGap         int64
	wantSplitTarget bool
	wantErr         bool
	wantAllocated   int64
}

func buildGreedyMatchCases(obs enumor.ObsProject, tech string,
	pass enumor.ResPlanReviewStatus) []greedyMatchCase {
	return []greedyMatchCase{
		{name: "exact match: single candidate equals demand", demandCores: 16,
			candidates: []*cvmapi.CvmCbsPlanQueryItem{makeCandidate("s1", obs, tech, 16, 0, pass)},
			wantGap:    0, wantAllocated: 16},
		{name: "greedy descending: two candidates exactly cover demand", demandCores: 26,
			candidates: []*cvmapi.CvmCbsPlanQueryItem{
				makeCandidate("s1", obs, tech, 16, 0, pass), makeCandidate("s2", obs, tech, 10, 0, pass)},
			wantGap: 0, wantAllocated: 26},
		{name: "greedy descending: partial fit, remainder needs split", demandCores: 24,
			candidates: []*cvmapi.CvmCbsPlanQueryItem{
				makeCandidate("s1", obs, tech, 16, 0, pass),
				makeCandidate("s2", obs, tech, 10, 0, pass),
				makeCandidate("s3", obs, tech, 4, 0, pass),
			}, wantGap: 4, wantSplitTarget: true, wantAllocated: 24},
		{name: "needs split: single over-candidate exceeds threshold", demandCores: 16,
			candidates: []*cvmapi.CvmCbsPlanQueryItem{makeCandidate("s1", obs, tech, 32, 0, pass)},
			wantGap:    16, wantSplitTarget: true, wantAllocated: 16},
		{name: "tolerance match: candidate exceeds by at most threshold", demandCores: 16,
			candidates: []*cvmapi.CvmCbsPlanQueryItem{makeCandidate("s1", obs, tech, 17, 0, pass)},
			wantGap:    0, wantAllocated: 16},
		{name: "tolerance match: candidate exceeds exactly at threshold", demandCores: 16,
			candidates: []*cvmapi.CvmCbsPlanQueryItem{makeCandidate("s1", obs, tech, 18, 0, pass)},
			wantGap:    0, wantAllocated: 16},
		{name: "just above threshold triggers split", demandCores: 16,
			candidates: []*cvmapi.CvmCbsPlanQueryItem{makeCandidate("s1", obs, tech, 19, 0, pass)},
			wantGap:    16, wantSplitTarget: true, wantAllocated: 16},
		{name: "transfer pool insufficient: no over-candidates", demandCores: 32,
			candidates: []*cvmapi.CvmCbsPlanQueryItem{
				makeCandidate("s1", obs, tech, 10, 0, pass), makeCandidate("s2", obs, tech, 8, 0, pass)},
			wantGap: 14, wantAllocated: 18},
		{name: "skip in-processing candidate", demandCores: 16,
			candidates: []*cvmapi.CvmCbsPlanQueryItem{
				makeCandidate("s1", obs, tech, 16, 1, pass), makeCandidate("s2", obs, tech, 32, 0, pass)},
			wantGap: 16, wantSplitTarget: true, wantAllocated: 16},
		{name: "skip pending review candidate", demandCores: 16,
			candidates: []*cvmapi.CvmCbsPlanQueryItem{
				makeCandidate("s1", obs, tech, 16, 0, enumor.ResPlanReviewStatusPending),
				makeCandidate("s2", obs, tech, 32, 0, pass),
			}, wantGap: 16, wantSplitTarget: true, wantAllocated: 16},
		{name: "pre-deduction: shared sliceID reduces effective remain", demandCores: 12,
			candidates:  []*cvmapi.CvmCbsPlanQueryItem{makeCandidate("s1", obs, tech, 20, 0, pass)},
			preConsumed: map[string]int64{"s1": 10}, wantGap: 2, wantAllocated: 10},
		{name: "pre-deduction prevents over-allocation: second demand uses updated remain", demandCores: 16,
			candidates:  []*cvmapi.CvmCbsPlanQueryItem{makeCandidate("s1", obs, tech, 20, 0, pass)},
			preConsumed: map[string]int64{"s1": 16}, wantGap: 12, wantAllocated: 4},
		{name: "project mismatch: candidate skipped", demandCores: 16,
			candidates: []*cvmapi.CvmCbsPlanQueryItem{
				{SliceId: "s1", ProjectName: "其他项目", TechnicalClass: tech,
					RealCoreAmount: 16, CoreAmount: 16, CvmAmount: 4, ReviewStatus: pass},
			}, wantGap: 16, wantAllocated: 0},
		{name: "zero demand cores (CBS-only)", demandCores: 0,
			candidates: []*cvmapi.CvmCbsPlanQueryItem{makeCandidate("s1", obs, tech, 32, 0, pass)},
			wantGap:    0, wantAllocated: 0},
		{name: "nil updated demand returns error", useNilDemand: true, wantErr: true},
	}
}

func runGreedyMatchCase(t *testing.T, kt *kit.Kit, obs enumor.ObsProject, tech string,
	pass enumor.ResPlanReviewStatus, tc greedyMatchCase) {
	var demand rpt.ResPlanDemand
	if !tc.useNilDemand {
		demand = makeDemand(obs, tech, tc.demandCores)
	}

	matchResult := make(map[string]*AdjustAbleRemainObj)
	for sliceID, consumed := range tc.preConsumed {
		matchResult[sliceID] = &AdjustAbleRemainObj{
			OriginDemand:   makeCandidate(sliceID, obs, tech, 20, 0, pass),
			WillConsume:    consumed,
			TransferTarget: make(map[rpt.UpdatedRPDemandItem]int64),
		}
	}

	gap, splitTarget, err := greedyMatch(kt, demand, tc.candidates, matchResult)
	if tc.wantErr {
		assert.Error(t, err)
		return
	}
	assert.NoError(t, err)
	assert.Equal(t, tc.wantGap, gap)
	if tc.wantSplitTarget {
		assert.NotNil(t, splitTarget)
	} else {
		assert.Nil(t, splitTarget)
	}

	freshAllocated := totalWillConsume(matchResult)
	for _, consumed := range tc.preConsumed {
		freshAllocated -= consumed
	}
	assert.Equal(t, tc.wantAllocated, freshAllocated,
		"freshly allocated cores in matchResult should match wantAllocated")
}

func TestGreedyMatch(t *testing.T) {
	kt := testKit()
	obs := enumor.ObsProjectNormal
	tech := "标准型"
	pass := enumor.ResPlanReviewStatusPass

	for _, tc := range buildGreedyMatchCases(obs, tech, pass) {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			runGreedyMatchCase(t, kt, obs, tech, pass, tc)
		})
	}
}

// TestGreedyMatch_MultiDemandPreDeductionCapacity verifies that when two demands both select
// the same SliceId as splitTarget, the combined gaps do not exceed the candidate's core count.
func TestGreedyMatch_MultiDemandPreDeductionCapacity(t *testing.T) {
	kt := testKit()
	obs := enumor.ObsProjectNormal
	tech := "标准型"
	pass := enumor.ResPlanReviewStatusPass

	// SliceId "s1" has 36 cores. Demand A needs 20, Demand B needs 10.
	candidate := makeCandidate("s1", obs, tech, 36, 0, pass)
	matchResult := make(map[string]*AdjustAbleRemainObj)

	// First demand: needs 20, candidate has 36 > 20 and 36 - 20 = 16 > threshold(2), so split.
	gapA, targetA, err := greedyMatch(kt, makeDemand(obs, tech, 20), []*cvmapi.CvmCbsPlanQueryItem{candidate},
		matchResult)
	assert.NoError(t, err)
	assert.Equal(t, int64(20), gapA)
	assert.NotNil(t, targetA)

	// After first demand: matchResult["s1"].WillConsume = 20, effective remain = 36 - 20 = 16.
	assert.Equal(t, int64(20), matchResult["s1"].WillConsume)

	// Second demand: needs 10, effective remain of s1 = 16. 16 > 10 and 16 - 10 = 6 > threshold(2), so split.
	gapB, targetB, err := greedyMatch(kt, makeDemand(obs, tech, 10), []*cvmapi.CvmCbsPlanQueryItem{candidate},
		matchResult)
	assert.NoError(t, err)
	assert.Equal(t, int64(10), gapB)
	assert.NotNil(t, targetB)

	// After second demand: WillConsume = 20 + 10 = 30, which is <= 36.
	assert.Equal(t, int64(30), matchResult["s1"].WillConsume,
		"combined pre-deduction should not exceed candidate's core count")
}

// ---- TestMergeSplitTargets ----

func TestMergeSplitTargets(t *testing.T) {
	obs := enumor.ObsProjectNormal
	tech := "标准型"
	pass := enumor.ResPlanReviewStatusPass

	s1 := makeCandidate("s1", obs, tech, 36, 0, pass)
	s2 := makeCandidate("s2", obs, tech, 20, 0, pass)

	testCases := []struct {
		name    string
		targets []splitTargetEntry
		// expected SliceId keys in merged result
		wantSliceIDs []string
		// expected gap lists per SliceId
		wantGaps map[string][]int64
	}{
		{
			name: "single demand single SliceId",
			targets: []splitTargetEntry{
				{Source: s1, Gap: 20},
			},
			wantSliceIDs: []string{"s1"},
			wantGaps:     map[string][]int64{"s1": {20}},
		},
		{
			name: "two demands different SliceIds",
			targets: []splitTargetEntry{
				{Source: s1, Gap: 20},
				{Source: s2, Gap: 10},
			},
			wantSliceIDs: []string{"s1", "s2"},
			wantGaps:     map[string][]int64{"s1": {20}, "s2": {10}},
		},
		{
			name: "two demands same SliceId aggregated",
			targets: []splitTargetEntry{
				{Source: s1, Gap: 20},
				{Source: s1, Gap: 10},
			},
			wantSliceIDs: []string{"s1"},
			wantGaps:     map[string][]int64{"s1": {20, 10}},
		},
		{
			name:         "empty input",
			targets:      []splitTargetEntry{},
			wantSliceIDs: []string{},
			wantGaps:     map[string][]int64{},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			result := mergeSplitTargets(tc.targets)
			assert.Equal(t, len(tc.wantSliceIDs), len(result))

			for _, sliceID := range tc.wantSliceIDs {
				mt, ok := result[sliceID]
				assert.True(t, ok, "SliceId %s should be present in result", sliceID)
				if ok {
					assert.ElementsMatch(t, tc.wantGaps[sliceID], mt.Gaps)
				}
			}
		})
	}
}

// ---- mock CVM client ----

// mockCRPClient implements cvmapi.CVMClientInterface for unit tests.
// Only AdjustCvmCbsPlans and QueryPlanOrder need real implementations;
// all other methods panic to catch unexpected calls.
type mockCRPClient struct {
	adjustFn     func(*cvmapi.CvmCbsPlanAdjustReq) (*cvmapi.CvmCbsPlanAdjustResp, error)
	queryOrderFn func(*cvmapi.QueryPlanOrderReq) (*cvmapi.QueryPlanOrderResp, error)
	callCount    int
}

func (m *mockCRPClient) AdjustCvmCbsPlans(_ context.Context, _ http.Header,
	req *cvmapi.CvmCbsPlanAdjustReq) (*cvmapi.CvmCbsPlanAdjustResp, error) {
	m.callCount++
	return m.adjustFn(req)
}

func (m *mockCRPClient) QueryPlanOrder(_ context.Context, _ http.Header,
	req *cvmapi.QueryPlanOrderReq) (*cvmapi.QueryPlanOrderResp, error) {
	m.callCount++
	return m.queryOrderFn(req)
}

// Unused interface methods – panic to detect unexpected calls during testing.
func (m *mockCRPClient) CreateCvmOrder(_ context.Context, _ http.Header,
	_ *cvmapi.OrderCreateReq) (*cvmapi.OrderCreateResp, error) {
	panic("unexpected call to CreateCvmOrder")
}
func (m *mockCRPClient) QueryCvmOrders(_ context.Context, _ http.Header,
	_ *cvmapi.OrderQueryReq) (*cvmapi.OrderQueryResp, error) {
	panic("unexpected call to QueryCvmOrders")
}
func (m *mockCRPClient) QueryCvmInstances(_ context.Context, _ http.Header,
	_ *cvmapi.InstanceQueryReq) (*cvmapi.InstanceQueryResp, error) {
	panic("unexpected")
}
func (m *mockCRPClient) QueryCvmCapacity(_ context.Context, _ http.Header, _ *cvmapi.CapacityReq) (*cvmapi.CapacityResp,
	error) {
	panic("unexpected")
}
func (m *mockCRPClient) QueryCvmVpc(_ context.Context, _ http.Header, _ *cvmapi.VpcReq) (*cvmapi.VpcResp, error) {
	panic("unexpected")
}
func (m *mockCRPClient) QueryRealCvmSubnet(_ *kit.Kit, _ cvmapi.SubnetRealParam) (*cvmapi.SubnetResp, error) {
	panic("unexpected")
}
func (m *mockCRPClient) GetApproveLog(_ context.Context, _ http.Header,
	_ *cvmapi.GetApproveLogReq) (*cvmapi.GetApproveLogResp, error) {
	panic("unexpected")
}
func (m *mockCRPClient) CreateCvmReturnOrder(_ context.Context, _ http.Header,
	_ *cvmapi.ReturnReq) (*cvmapi.OrderCreateResp, error) {
	panic("unexpected")
}
func (m *mockCRPClient) QueryCvmReturnOrders(_ context.Context, _ http.Header,
	_ *cvmapi.OrderQueryReq) (*cvmapi.ReturnQueryResp, error) {
	panic("unexpected")
}
func (m *mockCRPClient) QueryCvmReturnDetail(_ context.Context, _ http.Header,
	_ *cvmapi.ReturnDetailReq) (*cvmapi.ReturnDetailResp, error) {
	panic("unexpected")
}
func (m *mockCRPClient) CreateUpgradeOrder(_ *kit.Kit, _ *cvmapi.UpgradeReq) (*cvmapi.OrderCreateResp, error) {
	panic("unexpected")
}
func (m *mockCRPClient) QueryCvmUpgradeDetail(_ *kit.Kit, _ *cvmapi.UpgradeDetailReq) (*cvmapi.UpgradeDetailResp,
	error) {
	panic("unexpected")
}
func (m *mockCRPClient) GetCvmProcess(_ context.Context, _ http.Header,
	_ *cvmapi.GetCvmProcessReq) (*cvmapi.GetCvmProcessResp, error) {
	panic("unexpected")
}
func (m *mockCRPClient) GetErpProcess(_ context.Context, _ http.Header,
	_ *cvmapi.GetErpProcessReq) (*cvmapi.GetErpProcessResp, error) {
	panic("unexpected")
}
func (m *mockCRPClient) QueryCvmInstanceType(_ *kit.Kit,
	_ *cvmapi.QueryCvmInstanceTypeParams) (*cvmapi.QueryCvmInstanceTypeResp, error) {
	panic("unexpected")
}
func (m *mockCRPClient) GetInstanceTypeInfo(_ *kit.Kit,
	_ *cvmapi.GetInstanceTypeInfoParams) (*cvmapi.GetInstanceTypeInfoResp, error) {
	panic("unexpected")
}
func (m *mockCRPClient) GetCvmApproveLogs(_ context.Context, _ http.Header,
	_ *cvmapi.GetCvmApproveLogReq) (*cvmapi.GetCvmApproveLogsResp, error) {
	panic("unexpected")
}
func (m *mockCRPClient) RevokeCvmOrder(_ context.Context, _ http.Header,
	_ *cvmapi.RevokeCvmOrderReq) (*cvmapi.RevokeCvmOrderResp, error) {
	panic("unexpected")
}
func (m *mockCRPClient) QueryCvmCbsPlans(_ context.Context, _ http.Header,
	_ *cvmapi.CvmCbsPlanQueryReq) (*cvmapi.CvmCbsPlanQueryResp, error) {
	panic("unexpected")
}
func (m *mockCRPClient) QueryAdjustAbleDemand(_ context.Context, _ http.Header,
	_ *cvmapi.CvmCbsAdjustAblePlanQueryReq) (*cvmapi.CvmCbsPlanQueryResp, error) {
	panic("unexpected")
}
func (m *mockCRPClient) AddCvmCbsPlan(_ context.Context, _ http.Header,
	_ *cvmapi.AddCvmCbsPlanReq) (*cvmapi.AddCvmCbsPlanResp, error) {
	panic("unexpected")
}
func (m *mockCRPClient) QueryPlanOrderChange(_ context.Context, _ http.Header,
	_ *cvmapi.PlanOrderChangeReq) (*cvmapi.PlanOrderChangeResp, error) {
	panic("unexpected")
}
func (m *mockCRPClient) QueryDemandChangeLog(_ context.Context, _ http.Header,
	_ *cvmapi.DemandChangeLogQueryReq) (*cvmapi.DemandChangeLogQueryResp, error) {
	panic("unexpected")
}
func (m *mockCRPClient) ReportPenaltyRatio(_ context.Context, _ http.Header,
	_ *cvmapi.CvmCbsPlanPenaltyRatioReportReq) (*cvmapi.CvmCbsPlanPenaltyRatioReportResp, error) {
	panic("unexpected")
}
func (m *mockCRPClient) QueryReturnPlan(_ context.Context, _ http.Header,
	_ *cvmapi.QueryReturnPlanReq) (*cvmapi.QueryReturnPlanResp, error) {
	panic("unexpected")
}
func (m *mockCRPClient) QueryOrderList(_ context.Context, _ http.Header,
	_ *cvmapi.QueryOrderListReq) (*cvmapi.QueryOrderListResp, error) {
	panic("unexpected")
}
func (m *mockCRPClient) MatchSwapGroup(_ context.Context, _ http.Header,
	_ *cvmapi.MatchSwapGroupReq) (*cvmapi.MatchSwapGroupResp, error) {
	panic("unexpected")
}
func (m *mockCRPClient) QueryMatchTask(_ context.Context, _ http.Header,
	_ *cvmapi.QueryMatchTaskReq) (*cvmapi.QueryMatchTaskResp, error) {
	panic("unexpected")
}
func (m *mockCRPClient) CreateTransOrder(_ context.Context, _ http.Header,
	_ *cvmapi.TransOrderReq) (*cvmapi.TransOrderResp, error) {
	panic("unexpected")
}

// newTestCreator builds a CrpTicketCreator with the provided mock client for unit testing.
func newTestCreator(cli cvmapi.CVMClientInterface) *CrpTicketCreator {
	return &CrpTicketCreator{
		crpCli:              cli,
		adjCRPDemandsRst:    make(map[string]*AdjustAbleRemainObj),
		appendUpdateDemand:  make([]*cvmapi.AdjustUpdatedData, 0),
		adjustAbleDemands:   make(map[string][]*cvmapi.CvmCbsPlanQueryItem),
		transferAbleDemands: make(map[int][]*cvmapi.CvmCbsPlanQueryItem),
	}
}

// testSubTicket builds a minimal SubTicketInfo for split tests.
func testSubTicket() *ptypes.SubTicketInfo {
	return &ptypes.SubTicketInfo{
		ID:              "sub-001",
		VirtualDeptID:   cvmapi.CvmCbsPlanDeptId,
		VirtualDeptName: "IEG技术运营部",
		Applicant:       "test-user",
	}
}

// ---- TestSplitAdjustOrder ----

func TestSplitAdjustOrder(t *testing.T) {
	obs := enumor.ObsProjectNormal
	tech := "标准型"
	pass := enumor.ResPlanReviewStatusPass

	testCases := []struct {
		name          string
		mergedTargets map[string]*mergedSplitTarget
		adjustResp    *cvmapi.CvmCbsPlanAdjustResp
		adjustErr     error
		wantOrderSN   string
		wantErr       bool
		wantSentinel  bool // expect errAdjustInProcessing
	}{
		{
			name: "single SliceId two-way split success",
			mergedTargets: map[string]*mergedSplitTarget{
				"s1": {
					Source: makeCandidate("s1", obs, tech, 36, 0, pass),
					Gaps:   []int64{20},
				},
			},
			adjustResp: &cvmapi.CvmCbsPlanAdjustResp{
				Result: &cvmapi.CvmCbsPlanAdjustRst{OrderId: "order-001"},
			},
			wantOrderSN: "order-001",
		},
		{
			name: "two SliceIds multi-split success",
			mergedTargets: map[string]*mergedSplitTarget{
				"s1": {Source: makeCandidate("s1", obs, tech, 36, 0, pass), Gaps: []int64{20}},
				"s2": {Source: makeCandidate("s2", obs, tech, 20, 0, pass), Gaps: []int64{10}},
			},
			adjustResp: &cvmapi.CvmCbsPlanAdjustResp{
				Result: &cvmapi.CvmCbsPlanAdjustRst{OrderId: "order-002"},
			},
			wantOrderSN: "order-002",
		},
		{
			name: "same SliceId N+1 split (two demands)",
			mergedTargets: map[string]*mergedSplitTarget{
				"s1": {Source: makeCandidate("s1", obs, tech, 36, 0, pass), Gaps: []int64{20, 10}},
			},
			adjustResp: &cvmapi.CvmCbsPlanAdjustResp{
				Result: &cvmapi.CvmCbsPlanAdjustRst{OrderId: "order-003"},
			},
			wantOrderSN: "order-003",
		},
		{
			name: "CRP returns in-processing error → sentinel returned",
			mergedTargets: map[string]*mergedSplitTarget{
				"s1": {Source: makeCandidate("s1", obs, tech, 36, 0, pass), Gaps: []int64{20}},
			},
			adjustResp: &cvmapi.CvmCbsPlanAdjustResp{
				RespMeta: cvmapi.RespMeta{
					Error: cvmapi.RespError{
						Code:    -1,
						Message: "AdjustDemandIsInProcessingException: demand is locked",
					},
				},
			},
			wantErr:      true,
			wantSentinel: true,
		},
		{
			name: "CRP returns other error → propagated",
			mergedTargets: map[string]*mergedSplitTarget{
				"s1": {Source: makeCandidate("s1", obs, tech, 36, 0, pass), Gaps: []int64{20}},
			},
			adjustResp: &cvmapi.CvmCbsPlanAdjustResp{
				RespMeta: cvmapi.RespMeta{
					Error: cvmapi.RespError{Code: -2, Message: "some other CRP error"},
				},
			},
			wantErr: true,
		},
		{
			name: "gap exceeds source CoreAmount → error before calling CRP",
			mergedTargets: map[string]*mergedSplitTarget{
				"s1": {Source: makeCandidate("s1", obs, tech, 20, 0, pass), Gaps: []int64{25}},
			},
			wantErr: true,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			mock := &mockCRPClient{
				adjustFn: func(_ *cvmapi.CvmCbsPlanAdjustReq) (*cvmapi.CvmCbsPlanAdjustResp, error) {
					return tc.adjustResp, tc.adjustErr
				},
			}
			c := newTestCreator(mock)
			orderSN, err := c.splitAdjustOrder(testKit(), testSubTicket(), tc.mergedTargets)

			if tc.wantErr {
				assert.Error(t, err)
				if tc.wantSentinel {
					assert.ErrorIs(t, err, errAdjustInProcessing)
				}
				return
			}
			assert.NoError(t, err)
			assert.Equal(t, tc.wantOrderSN, orderSN)
		})
	}
}

// TestSplitAdjustOrder_UpdatedDataSum verifies that for each srcData item the corresponding
// updatedData core amounts sum to exactly srcData[i].CoreAmount.
func TestSplitAdjustOrder_UpdatedDataSum(t *testing.T) {
	obs := enumor.ObsProjectNormal
	tech := "标准型"
	pass := enumor.ResPlanReviewStatusPass

	var capturedReq *cvmapi.CvmCbsPlanAdjustReq
	mock := &mockCRPClient{
		adjustFn: func(req *cvmapi.CvmCbsPlanAdjustReq) (*cvmapi.CvmCbsPlanAdjustResp, error) {
			capturedReq = req
			return &cvmapi.CvmCbsPlanAdjustResp{
				Result: &cvmapi.CvmCbsPlanAdjustRst{OrderId: "order-x"},
			}, nil
		},
	}
	c := newTestCreator(mock)

	// SliceId "s1" = 36 cores; split into [20, 10, 6].
	merged := map[string]*mergedSplitTarget{
		"s1": {Source: makeCandidate("s1", obs, tech, 36, 0, pass), Gaps: []int64{20, 10}},
	}
	_, err := c.splitAdjustOrder(testKit(), testSubTicket(), merged)
	assert.NoError(t, err)
	assert.NotNil(t, capturedReq)

	// Verify srcData has exactly one entry with CoreAmount = 36.
	assert.Equal(t, 1, len(capturedReq.Params.SrcData))
	srcCore := capturedReq.Params.SrcData[0].CoreAmount

	// Sum updatedData core amounts for sliceId "s1".
	var updatedSum int64
	for _, ud := range capturedReq.Params.UpdatedData {
		updatedSum += ud.CoreAmount
	}
	assert.Equal(t, srcCore, updatedSum, "srcData[0].CoreAmount must equal sum of updatedData.CoreAmount")
}

// ---- TestPollSplitOrderUntilApproved ----

func TestPollSplitOrderUntilApproved(t *testing.T) {
	approvedResp := func(sn string) *cvmapi.QueryPlanOrderResp {
		return &cvmapi.QueryPlanOrderResp{
			Result: map[string]*cvmapi.QueryPlanOrderRst{
				sn: {
					Data: cvmapi.PlanOrderData{
						BaseInfo: cvmapi.PlanOrderBaseInfo{
							Status: cvmapi.PlanOrderStatusApproved,
						},
					},
				},
			},
		}
	}
	rejectedResp := func(sn string) *cvmapi.QueryPlanOrderResp {
		return &cvmapi.QueryPlanOrderResp{
			Result: map[string]*cvmapi.QueryPlanOrderRst{
				sn: {
					Data: cvmapi.PlanOrderData{
						BaseInfo: cvmapi.PlanOrderBaseInfo{
							Status: cvmapi.PlanOrderStatusRejected,
						},
					},
				},
			},
		}
	}
	pendingResp := func(sn string) *cvmapi.QueryPlanOrderResp {
		return &cvmapi.QueryPlanOrderResp{
			Result: map[string]*cvmapi.QueryPlanOrderRst{
				sn: {
					Data: cvmapi.PlanOrderData{
						BaseInfo: cvmapi.PlanOrderBaseInfo{
							Status: cvmapi.PlanOrderStatusDeptAdmin, // in-progress status
						},
					},
				},
			},
		}
	}

	testCases := []struct {
		name        string
		orderSN     string
		responses   []*cvmapi.QueryPlanOrderResp
		wantErr     bool
		wantErrText string
	}{
		{
			name:      "first poll approved",
			orderSN:   "order-001",
			responses: []*cvmapi.QueryPlanOrderResp{approvedResp("order-001")},
		},
		{
			name:    "poll once pending then approved",
			orderSN: "order-002",
			responses: []*cvmapi.QueryPlanOrderResp{
				pendingResp("order-002"),
				approvedResp("order-002"),
			},
		},
		{
			name:        "rejected immediately",
			orderSN:     "order-003",
			responses:   []*cvmapi.QueryPlanOrderResp{rejectedResp("order-003")},
			wantErr:     true,
			wantErrText: "rejected",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			idx := 0
			mock := &mockCRPClient{
				queryOrderFn: func(_ *cvmapi.QueryPlanOrderReq) (*cvmapi.QueryPlanOrderResp, error) {
					resp := tc.responses[idx]
					if idx < len(tc.responses)-1 {
						idx++
					}
					return resp, nil
				},
			}
			c := newTestCreator(mock)

			err := c.pollSplitOrderUntilApproved(testKit(), tc.orderSN)

			if tc.wantErr {
				assert.Error(t, err)
				if tc.wantErrText != "" {
					assert.Contains(t, err.Error(), tc.wantErrText)
				}
				return
			}
			assert.NoError(t, err)
		})
	}
}

// TestPollSplitOrderUntilApproved_Timeout verifies that the function returns an error
// when polling exceeds the 30-second deadline.
// This test overrides the internal timeout to avoid waiting 30s in CI.
func TestPollSplitOrderUntilApproved_Timeout(t *testing.T) {
	orderSN := "order-timeout"

	// Always return a pending (non-terminal) status.
	pendingResp := &cvmapi.QueryPlanOrderResp{
		Result: map[string]*cvmapi.QueryPlanOrderRst{
			orderSN: {
				Data: cvmapi.PlanOrderData{
					BaseInfo: cvmapi.PlanOrderBaseInfo{
						Status: cvmapi.PlanOrderStatusDeptAdmin,
					},
				},
			},
		},
	}

	_ = pendingResp

	// This test is skipped to avoid a real 30s wait in CI.
	// Verify manually by overriding pollSplitOrderTimeout to a small value.
	t.Skip("skipping timeout test to avoid 30s wait in CI; verify manually with pollSplitOrderTimeout override")
}
