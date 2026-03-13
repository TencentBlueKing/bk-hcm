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

package splitter

import (
	"context"
	"testing"

	"hcm/pkg/criteria/enumor"
	rpt "hcm/pkg/dal/table/resource-plan/res-plan-ticket"
	"hcm/pkg/kit"
	"hcm/pkg/thirdparty/cvmapi"

	"github.com/stretchr/testify/assert"
)

// splitterTestKit returns a minimal kit for unit tests.
func splitterTestKit() *kit.Kit {
	return &kit.Kit{
		Ctx: context.Background(),
		Rid: "test-rid-splitter",
	}
}

// newTestSplitter returns a SubTicketSplitter with only the transfer-pool-related
// maps initialised (suitable for unit-testing matchTransferCRPDemands).
func newTestSplitter() *SubTicketSplitter {
	return &SubTicketSplitter{
		transferAbleDemands:  make(map[int][]*cvmapi.CvmCbsPlanQueryItem),
		transferCRPDemandRst: make(map[string]*AdjustAbleRemainObj),
		adjSplitGroupDemands: make(map[enumor.RPTicketType][]*rpt.ResPlanDemand),
		adjustAbleDemands:    make(map[string][]*cvmapi.CvmCbsPlanQueryItem),
		adjCRPDemandsRst:     make(map[string]*AdjustAbleRemainObj),
	}
}

// makeTransferCandidate builds a CvmCbsPlanQueryItem for transfer pool tests.
func makeTransferCandidate(sliceID string, obs enumor.ObsProject, tech string,
	realCore int64, isInProcessing int, reviewStatus enumor.ResPlanReviewStatus) *cvmapi.CvmCbsPlanQueryItem {
	return &cvmapi.CvmCbsPlanQueryItem{
		SliceId:        sliceID,
		ProjectName:    obs,
		TechnicalClass: tech,
		RealCoreAmount: realCore,
		CoreAmount:     realCore,
		CvmAmount:      float64(realCore) / 4,
		ReviewStatus:   reviewStatus,
		IsInProcessing: isInProcessing,
	}
}

// makeMatchDemand builds a rpt.ResPlanDemand for matchTransferCRPDemands tests.
func makeMatchDemand(obs enumor.ObsProject, tech, expectTime string, cpuCore int64) rpt.ResPlanDemand {
	return rpt.ResPlanDemand{
		Updated: &rpt.UpdatedRPDemandItem{
			ObsProject: obs,
			ExpectTime: expectTime,
			Cvm: rpt.Cvm{
				CpuCore:        cpuCore,
				TechnicalClass: tech,
			},
		},
	}
}

// TestMatchTransferCRPDemands_IsInProcessingFilter verifies that candidates with
// IsInProcessing == 1 are skipped and not counted toward the transferable core total.
func TestMatchTransferCRPDemands_IsInProcessingFilter(t *testing.T) {
	obs := enumor.ObsProjectNormal
	tech := "标准型"
	pass := enumor.ResPlanReviewStatusPass
	expectTime := "2025-01-01"
	year := 2025

	testCases := []struct {
		name                 string
		candidates           []*cvmapi.CvmCbsPlanQueryItem
		demandCores          int64
		wantTransferable     int64
		wantNonTransferable  int64
	}{
		{
			name: "all candidates available: fully transferable",
			candidates: []*cvmapi.CvmCbsPlanQueryItem{
				makeTransferCandidate("s1", obs, tech, 16, 0, pass),
			},
			demandCores:         16,
			wantTransferable:    16,
			wantNonTransferable: 0,
		},
		{
			name: "single in-processing candidate: fully non-transferable",
			candidates: []*cvmapi.CvmCbsPlanQueryItem{
				makeTransferCandidate("s1", obs, tech, 16, 1, pass), // in-processing → skip
			},
			demandCores:         16,
			wantTransferable:    0,
			wantNonTransferable: 16,
		},
		{
			name: "mixed: in-processing skipped, available consumed",
			candidates: []*cvmapi.CvmCbsPlanQueryItem{
				makeTransferCandidate("s1", obs, tech, 16, 1, pass), // skipped
				makeTransferCandidate("s2", obs, tech, 8, 0, pass),  // consumed
			},
			demandCores:         16,
			wantTransferable:    8,
			wantNonTransferable: 8,
		},
		{
			name: "in-processing candidate next to available: only available counted",
			candidates: []*cvmapi.CvmCbsPlanQueryItem{
				makeTransferCandidate("s1", obs, tech, 32, 0, pass),  // consumed (partial)
				makeTransferCandidate("s2", obs, tech, 16, 1, pass),  // skipped
			},
			demandCores:         20,
			wantTransferable:    20,
			wantNonTransferable: 0,
		},
		{
			name: "pending review candidate also skipped",
			candidates: []*cvmapi.CvmCbsPlanQueryItem{
				makeTransferCandidate("s1", obs, tech, 16, 0, enumor.ResPlanReviewStatusPending), // skipped
				makeTransferCandidate("s2", obs, tech, 16, 0, pass),                              // consumed
			},
			demandCores:         16,
			wantTransferable:    16,
			wantNonTransferable: 0,
		},
		{
			name: "project mismatch: candidate skipped",
			candidates: []*cvmapi.CvmCbsPlanQueryItem{
				{SliceId: "s1", ProjectName: "其他项目", TechnicalClass: tech,
					RealCoreAmount: 16, CoreAmount: 16, CvmAmount: 4, ReviewStatus: pass},
			},
			demandCores:         16,
			wantTransferable:    0,
			wantNonTransferable: 16,
		},
		{
			name: "technical class mismatch: candidate skipped",
			candidates: []*cvmapi.CvmCbsPlanQueryItem{
				makeTransferCandidate("s1", obs, "其他类型", 16, 0, pass),
			},
			demandCores:         16,
			wantTransferable:    0,
			wantNonTransferable: 16,
		},
		{
			name: "zero CvmAmount guard: candidate skipped",
			candidates: []*cvmapi.CvmCbsPlanQueryItem{
				{SliceId: "s1", ProjectName: obs, TechnicalClass: tech,
					RealCoreAmount: 16, CoreAmount: 16, CvmAmount: 0, ReviewStatus: pass},
			},
			demandCores:         16,
			wantTransferable:    0,
			wantNonTransferable: 16,
		},
		{
			name:                "no candidates: fully non-transferable",
			candidates:          []*cvmapi.CvmCbsPlanQueryItem{},
			demandCores:         16,
			wantTransferable:    0,
			wantNonTransferable: 16,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			s := newTestSplitter()
			s.transferAbleDemands[year] = tc.candidates

			demand := makeMatchDemand(obs, tech, expectTime, tc.demandCores)
			transferable, nonTransferable, err := s.matchTransferCRPDemands(splitterTestKit(), "ticket-001", demand)

			assert.NoError(t, err)
			assert.Equal(t, tc.wantTransferable, transferable, "transferable cores mismatch")
			assert.Equal(t, tc.wantNonTransferable, nonTransferable, "non-transferable cores mismatch")
		})
	}
}

// TestMatchTransferCRPDemands_PreDeduction verifies that WillConsume in transferCRPDemandRst
// reduces the effective capacity seen by subsequent demands on the same SliceId.
func TestMatchTransferCRPDemands_PreDeduction(t *testing.T) {
	obs := enumor.ObsProjectNormal
	tech := "标准型"
	pass := enumor.ResPlanReviewStatusPass
	expectTime := "2025-06-01"
	year := 2025

	s := newTestSplitter()
	s.transferAbleDemands[year] = []*cvmapi.CvmCbsPlanQueryItem{
		makeTransferCandidate("s1", obs, tech, 20, 0, pass),
	}

	// First demand consumes 10 cores from s1.
	d1 := makeMatchDemand(obs, tech, expectTime, 10)
	t1, nt1, err := s.matchTransferCRPDemands(splitterTestKit(), "ticket-001", d1)
	assert.NoError(t, err)
	assert.Equal(t, int64(10), t1)
	assert.Equal(t, int64(0), nt1)

	// After first demand: s1 has 10 remaining.
	assert.Equal(t, int64(10), s.transferCRPDemandRst["s1"].WillConsume)

	// Second demand of 15 cores: only 10 left in s1.
	d2 := makeMatchDemand(obs, tech, expectTime, 15)
	t2, nt2, err := s.matchTransferCRPDemands(splitterTestKit(), "ticket-001", d2)
	assert.NoError(t, err)
	assert.Equal(t, int64(10), t2)
	assert.Equal(t, int64(5), nt2)

	// Total WillConsume across both demands must not exceed RealCoreAmount.
	assert.LessOrEqual(t, s.transferCRPDemandRst["s1"].WillConsume, int64(20),
		"WillConsume must not exceed RealCoreAmount of the candidate")
}

// TestMatchTransferCRPDemands_NilUpdated verifies that passing a demand with nil Updated
// returns an error without panicking.
func TestMatchTransferCRPDemands_NilUpdated(t *testing.T) {
	s := newTestSplitter()
	demand := rpt.ResPlanDemand{} // Updated is nil
	_, _, err := s.matchTransferCRPDemands(splitterTestKit(), "ticket-001", demand)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "nil")
}
