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
	"errors"
	"fmt"
	"slices"
	"strings"
	"time"

	ptypes "hcm/cmd/woa-server/types/plan"
	"hcm/pkg/criteria/constant"
	"hcm/pkg/criteria/enumor"
	rpt "hcm/pkg/dal/table/resource-plan/res-plan-ticket"
	"hcm/pkg/kit"
	"hcm/pkg/logs"
	"hcm/pkg/thirdparty/cvmapi"
	cvt "hcm/pkg/tools/converter"
	"hcm/pkg/tools/math"
	"hcm/pkg/tools/uuid"
)

const (
	// pollSplitOrderInterval is the interval between consecutive QueryPlanOrder polls.
	pollSplitOrderInterval = 2 * time.Second
	// pollSplitOrderTimeout is the maximum time to wait for a split order to be approved.
	pollSplitOrderTimeout = 30 * time.Second
)

// errAdjustInProcessing is the sentinel error returned by splitAdjustOrder when CRP rejects
// the request with AdjustDemandIsInProcessingException. The outer retry loop uses this to
// trigger a re-entry iteration instead of failing the sub-ticket immediately.
var errAdjustInProcessing = errors.New(constant.CRPResPlanDemandIsInProcessing)

// splitTargetEntry records a split target candidate and the required core gap for one demand.
type splitTargetEntry struct {
	Source *cvmapi.CvmCbsPlanQueryItem
	Gap    int64
}

// mergedSplitTarget records a split target candidate and the aggregated gaps from multiple demands
// that selected the same SliceId.
type mergedSplitTarget struct {
	Source *cvmapi.CvmCbsPlanQueryItem
	Gaps   []int64
}

// candidateWithRemain pairs a CvmCbsPlanQueryItem with its effective remaining core count
// (after deducting already-consumed amounts from matchResult).
type candidateWithRemain struct {
	item   *cvmapi.CvmCbsPlanQueryItem
	remain int64
}

// greedyMatch selects transfer pool demands to satisfy the given demand's core requirement.
// It is a package-level function (not a CrpTicketCreator method) to prevent accidental access
// to c.adjCRPDemandsRst; matching state is written only to the caller-provided matchResult.
//
// Algorithm:
//
//	Step 1 – Sort candidates with remain <= needCores by remain descending and greedily consume them.
//	Step 2 – If needCores > 0, pick the smallest candidate with remain > needCores as splitTarget.
//	         Tolerance: if the gap is within TransferCoreToleranceThreshold, treat as exact match.
//
// Returns:
//
//	gap        > 0 and splitTarget != nil : need to split splitTarget by gap cores before transfer.
//	gap        > 0 and splitTarget == nil : transfer pool is insufficient; caller should fail.
//	gap        == 0                       : all needCores satisfied; no split needed.
func greedyMatch(kt *kit.Kit, demand rpt.ResPlanDemand, candidates []*cvmapi.CvmCbsPlanQueryItem,
	matchResult map[string]*AdjustAbleRemainObj) (gap int64, splitTarget *cvmapi.CvmCbsPlanQueryItem, err error) {

	if demand.Updated == nil {
		return 0, nil, errors.New("updated demand is nil")
	}

	needDemand := demand.Updated
	needCores := needDemand.Cvm.CpuCore
	if needCores == 0 {
		// CBS-only demand: no cores needed, nothing to match.
		return 0, nil, nil
	}

	filtered := filterMatchCandidates(needDemand, candidates, matchResult)
	needCores, consumed := consumeFittingCandidates(needDemand, filtered, matchResult, needCores)
	if needCores <= 0 {
		return 0, nil, nil
	}

	gap, splitTarget = selectSplitCandidate(needDemand, filtered, consumed, needCores, matchResult)
	return gap, splitTarget, nil
}

// filterMatchCandidates filters the raw candidates list and computes effective remaining cores
// for each entry (deducting already-consumed amounts from matchResult).
func filterMatchCandidates(needDemand *rpt.UpdatedRPDemandItem, candidates []*cvmapi.CvmCbsPlanQueryItem,
	matchResult map[string]*AdjustAbleRemainObj) []candidateWithRemain {

	filtered := make([]candidateWithRemain, 0, len(candidates))
	for _, c := range candidates {
		if c.IsInProcessing == 1 {
			continue
		}
		if c.ReviewStatus == enumor.ResPlanReviewStatusPending {
			continue
		}
		if c.ProjectName != needDemand.ObsProject || c.TechnicalClass != needDemand.Cvm.TechnicalClass {
			continue
		}
		// Guard against division-by-zero in downstream CvmAmount calculations.
		if c.CvmAmount == 0 {
			continue
		}
		remain := c.RealCoreAmount
		if obj, ok := matchResult[c.SliceId]; ok {
			remain -= obj.WillConsume
		}
		if remain <= 0 {
			continue
		}
		filtered = append(filtered, candidateWithRemain{item: c, remain: remain})
	}
	return filtered
}

// consumeFittingCandidates implements Step 1 of greedyMatch: consumes candidates whose effective
// remain <= needCores (sorted by remain descending to minimise CRP lock contention).
// Returns the updated needCores and the set of consumed SliceIds.
func consumeFittingCandidates(needDemand *rpt.UpdatedRPDemandItem, filtered []candidateWithRemain,
	matchResult map[string]*AdjustAbleRemainObj, needCores int64) (int64, map[string]bool) {

	slices.SortFunc(filtered, func(a, b candidateWithRemain) int {
		if a.remain > b.remain {
			return -1
		}
		if a.remain < b.remain {
			return 1
		}
		return 0
	})

	consumed := make(map[string]bool)
	for _, cand := range filtered {
		if needCores <= 0 {
			break
		}
		if cand.remain > needCores {
			continue
		}
		consumed[cand.item.SliceId] = true
		initMatchEntry(matchResult, cand.item)
		obj := matchResult[cand.item.SliceId]
		obj.WillConsume += cand.remain
		obj.TransferTarget[cvt.PtrToVal(needDemand.Clone())] += cand.remain
		needCores -= cand.remain
	}
	return needCores, consumed
}

// selectSplitCandidate implements Step 2 of greedyMatch: picks the smallest candidate with
// remain > needCores. Applies the tolerance threshold to avoid unnecessary splits.
// Pre-deducts the gap into matchResult regardless of whether a split is triggered.
func selectSplitCandidate(needDemand *rpt.UpdatedRPDemandItem, filtered []candidateWithRemain,
	consumed map[string]bool, needCores int64,
	matchResult map[string]*AdjustAbleRemainObj) (int64, *cvmapi.CvmCbsPlanQueryItem) {

	overCandidates := make([]candidateWithRemain, 0)
	for _, cand := range filtered {
		if !consumed[cand.item.SliceId] && cand.remain > needCores {
			overCandidates = append(overCandidates, cand)
		}
	}
	if len(overCandidates) == 0 {
		return needCores, nil
	}

	slices.SortFunc(overCandidates, func(a, b candidateWithRemain) int {
		if a.remain < b.remain {
			return -1
		}
		if a.remain > b.remain {
			return 1
		}
		return 0
	})

	best := overCandidates[0]
	initMatchEntry(matchResult, best.item)
	obj := matchResult[best.item.SliceId]
	obj.WillConsume += needCores
	obj.TransferTarget[cvt.PtrToVal(needDemand.Clone())] += needCores

	// Tolerance check: treat as exact match to avoid an unnecessary split round-trip.
	if best.remain-needCores <= constant.TransferCoreToleranceThreshold {
		return 0, nil
	}
	return needCores, best.item
}

// initMatchEntry ensures matchResult has an entry for the given item's SliceId.
func initMatchEntry(matchResult map[string]*AdjustAbleRemainObj, item *cvmapi.CvmCbsPlanQueryItem) {
	if _, ok := matchResult[item.SliceId]; !ok {
		matchResult[item.SliceId] = &AdjustAbleRemainObj{
			OriginDemand:   item.Clone(),
			TransferTarget: make(map[rpt.UpdatedRPDemandItem]int64),
		}
	}
}

// mergeSplitTargets merges multiple per-demand split targets by SliceId.
// When the same SliceId is selected by multiple demands, all their gaps are aggregated
// so that a single multi-way adjustOrder can split the source into N+1 segments.
func mergeSplitTargets(targets []splitTargetEntry) map[string]*mergedSplitTarget {
	result := make(map[string]*mergedSplitTarget, len(targets))
	for _, t := range targets {
		sliceID := t.Source.SliceId
		if _, ok := result[sliceID]; !ok {
			result[sliceID] = &mergedSplitTarget{Source: t.Source, Gaps: make([]int64, 0)}
		}
		result[sliceID].Gaps = append(result[sliceID].Gaps, t.Gap)
	}
	return result
}

// splitAdjustOrder constructs and submits an adjustOrder request to split transfer pool demands.
// It supports multi-way splits: when one SliceId is targeted by multiple demands it is split
// into N+1 segments (one per demand gap + one remainder segment).
//
// Returns (orderSN, nil) on success.
// Returns (_, errAdjustInProcessing) when CRP rejects with AdjustDemandIsInProcessingException.
// Returns (_, otherError) for all other CRP errors.
func (c *CrpTicketCreator) splitAdjustOrder(kt *kit.Kit, subTicket *ptypes.SubTicketInfo,
	mergedTargets map[string]*mergedSplitTarget) (string, error) {

	srcData, updatedData, err := buildSplitAdjustReqData(mergedTargets)
	if err != nil {
		return "", err
	}

	adjustReq := &cvmapi.CvmCbsPlanAdjustReq{
		ReqMeta: cvmapi.ReqMeta{Id: cvmapi.CvmId, JsonRpc: cvmapi.CvmJsonRpc, Method: cvmapi.CvmCbsPlanAdjustMethod},
		Params: &cvmapi.CvmCbsPlanAdjustParam{
			BaseInfo: &cvmapi.AdjustBaseInfo{
				DeptId:          int(subTicket.VirtualDeptID),
				DeptName:        subTicket.VirtualDeptName,
				PlanProductName: cvmapi.TransferPlanProductName,
				Desc:            cvmapi.CvmCbsPlanDefaultCvmDesc,
			},
			SrcData:     srcData,
			UpdatedData: updatedData,
			UserName:    subTicket.Applicant,
		},
	}

	resp, err := c.crpCli.AdjustCvmCbsPlans(kt.Ctx, kt.Header(), adjustReq)
	if err != nil {
		logs.Errorf("failed to call adjustOrder for transfer pool split, err: %v, "+
			"sub_ticket_id: %s, rid: %s", err, subTicket.ID, kt.Rid)
		return "", err
	}

	if resp.Error.Code != 0 {
		logs.Errorf("failed to split crp demand via adjustOrder, code: %d, msg: %s, "+
			"sub_ticket_id: %s, crp_trace: %s, rid: %s",
			resp.Error.Code, resp.Error.Message, subTicket.ID, resp.TraceId, kt.Rid)
		if strings.Contains(resp.Error.Message, constant.CRPResPlanDemandIsInProcessing) {
			return "", errAdjustInProcessing
		}
		return "", fmt.Errorf("failed to split crp demand, code: %d, msg: %s",
			resp.Error.Code, resp.Error.Message)
	}

	if resp.Result == nil || resp.Result.OrderId == "" {
		return "", errors.New("split adjustOrder returned empty orderSN")
	}
	return resp.Result.OrderId, nil
}

// buildSplitAdjustReqData constructs the srcData and updatedData slices for a split adjustOrder.
// Each mergedTarget produces one srcItem and N+1 updatedItems (N gap segments + one remainder).
func buildSplitAdjustReqData(mergedTargets map[string]*mergedSplitTarget) (
	[]*cvmapi.AdjustSrcData, []*cvmapi.AdjustUpdatedData, error) {

	srcData := make([]*cvmapi.AdjustSrcData, 0, len(mergedTargets))
	updatedData := make([]*cvmapi.AdjustUpdatedData, 0)

	for _, mt := range mergedTargets {
		source := mt.Source
		totalGap := int64(0)
		for _, g := range mt.Gaps {
			totalGap += g
		}
		if totalGap > source.CoreAmount {
			return nil, nil, fmt.Errorf("split gap sum %d exceeds source CoreAmount %d for SliceId %s",
				totalGap, source.CoreAmount, source.SliceId)
		}

		srcItem := &cvmapi.AdjustSrcData{
			AdjustType:          string(enumor.CrpAdjustTypeUpdate),
			CvmCbsPlanQueryItem: source.Clone(),
		}
		if source.ProjectName == enumor.ObsProjectShortLease {
			srcItem.IsAutoReturnPlan = true
		}
		srcData = append(srcData, srcItem)

		deviceCore := float64(source.CoreAmount) / source.CvmAmount
		gaps, remain, err := buildSplitUpdatedItems(source, mt.Gaps, deviceCore)
		if err != nil {
			return nil, nil, err
		}
		updatedData = append(updatedData, gaps...)
		updatedData = append(updatedData, remain...)
	}
	return srcData, updatedData, nil
}

// buildSplitUpdatedItems constructs the updatedData entries for a single source demand:
// one new item per gap (with a fresh UUID) and one remainder item (keeping the original SliceId).
func buildSplitUpdatedItems(source *cvmapi.CvmCbsPlanQueryItem, gaps []int64, deviceCore float64) (
	gapItems []*cvmapi.AdjustUpdatedData, remainItems []*cvmapi.AdjustUpdatedData, err error) {

	for _, gap := range gaps {
		splitCvmAmount, e := math.RoundToDecimalPlaces(float64(gap)/deviceCore, 4)
		if e != nil {
			return nil, nil, fmt.Errorf("failed to compute split CvmAmount for gap %d: %v", gap, e)
		}
		splitItem := source.Clone()
		splitItem.SliceId = uuid.UUID()
		splitItem.CoreAmount = gap
		splitItem.CvmAmount = splitCvmAmount
		gapItems = append(gapItems, &cvmapi.AdjustUpdatedData{
			AdjustType:          string(enumor.CrpAdjustTypeUpdate),
			CvmCbsPlanQueryItem: splitItem,
		})
	}

	totalGap := int64(0)
	for _, g := range gaps {
		totalGap += g
	}
	remainder := source.CoreAmount - totalGap
	if remainder > 0 {
		remainCvmAmount, e := math.RoundToDecimalPlaces(float64(remainder)/deviceCore, 4)
		if e != nil {
			return nil, nil, fmt.Errorf("failed to compute remainder CvmAmount for %d: %v", remainder, e)
		}
		remainItem := source.Clone()
		remainItem.CoreAmount = remainder
		remainItem.CvmAmount = remainCvmAmount
		remainItems = append(remainItems, &cvmapi.AdjustUpdatedData{
			AdjustType:          string(enumor.CrpAdjustTypeUpdate),
			CvmCbsPlanQueryItem: remainItem,
		})
	}
	return gapItems, remainItems, nil
}

// pollSplitOrderUntilApproved polls the CRP plan order status until it reaches a terminal state.
// It reuses the QueryPlanOrder call pattern from checkCrpTicket.
//
// Returns nil when PlanOrderStatusApproved.
// Returns error when PlanOrderStatusRejected or the poll timeout (30s) is exceeded.
func (c *CrpTicketCreator) pollSplitOrderUntilApproved(kt *kit.Kit, orderSN string) error {
	req := &cvmapi.QueryPlanOrderReq{
		ReqMeta: cvmapi.ReqMeta{Id: cvmapi.CvmId, JsonRpc: cvmapi.CvmJsonRpc, Method: cvmapi.CvmCbsPlanOrderQueryMethod},
		Params:  &cvmapi.QueryPlanOrderParam{OrderIds: []string{orderSN}},
	}

	deadline := time.Now().Add(pollSplitOrderTimeout)
	for {
		resp, err := c.crpCli.QueryPlanOrder(kt.Ctx, kt.Header(), req)
		if err != nil {
			logs.Errorf("failed to query plan order for split, sn: %s, err: %v, rid: %s", orderSN, err, kt.Rid)
			return err
		}
		if resp.Error.Code != 0 {
			logs.Errorf("query plan order for split returned error, sn: %s, code: %d, msg: %s, rid: %s",
				orderSN, resp.Error.Code, resp.Error.Message, kt.Rid)
			return fmt.Errorf("query plan order failed, code: %d, msg: %s", resp.Error.Code, resp.Error.Message)
		}
		if resp.Result == nil {
			logs.Errorf("query plan order for split returned nil result, sn: %s, rid: %s", orderSN, kt.Rid)
			return fmt.Errorf("query plan order returned nil result, sn: %s", orderSN)
		}
		planItem, ok := resp.Result[orderSN]
		if !ok {
			logs.Errorf("split order not found in query result, sn: %s, rid: %s", orderSN, kt.Rid)
			return fmt.Errorf("split order not found in query result: %s", orderSN)
		}

		switch planItem.Data.BaseInfo.Status {
		case cvmapi.PlanOrderStatusApproved:
			return nil
		case cvmapi.PlanOrderStatusRejected:
			logs.Errorf("split order was rejected, sn: %s, rid: %s", orderSN, kt.Rid)
			return fmt.Errorf("split order was rejected: %s", orderSN)
		}

		if time.Now().After(deadline) {
			logs.Errorf("poll split order timed out, sn: %s, rid: %s", orderSN, kt.Rid)
			return fmt.Errorf("poll split order timed out: %s", orderSN)
		}
		time.Sleep(pollSplitOrderInterval)
	}
}
