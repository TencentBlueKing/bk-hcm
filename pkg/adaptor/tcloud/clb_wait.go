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

package tcloud

import (
	"slices"

	"hcm/pkg/adaptor/poller"
	"hcm/pkg/adaptor/types"
	typelb "hcm/pkg/adaptor/types/load-balancer"
	"hcm/pkg/criteria/constant"
	"hcm/pkg/kit"
	"hcm/pkg/logs"
	cvt "hcm/pkg/tools/converter"
	"hcm/pkg/tools/slice"
)

// clbVisiblePollResult 是一轮查询的结果：每个目标云 ID 是否可见，以及本轮查到的 CLB。
type clbVisiblePollResult struct {
	visible map[string]bool
	lbs     []typelb.TCloudClb
}

// clbVisibleDoneResult 是等待结束时的结果。lbs 仅在目标云 ID 全部可见时有值。
type clbVisibleDoneResult struct {
	base poller.BaseDoneResult
	lbs  []typelb.TCloudClb
}

// WaitLoadBalancerVisible 等待 CLB 在云上可查询，不返回错误。
// 返回仍查不到的云 ID；目标全部可见时，同时返回最后一轮查到的 CLB，供同步直接使用。
// 未齐套或查询失败时，查到的 CLB 为 nil。opt 为空时使用 types.NewCLBVisibleWaitPollerOption()。
func WaitLoadBalancerVisible(kt *kit.Kit, client TCloud, region string, cloudIDs []string,
	opt *poller.PollUntilDoneOption,
) ([]string, []typelb.TCloudClb) {

	if len(cloudIDs) == 0 {
		return nil, nil
	}
	if opt == nil {
		opt = types.NewCLBVisibleWaitPollerOption()
	}

	waitPoller := poller.Poller[TCloud, clbVisiblePollResult, clbVisibleDoneResult]{
		Handler: &clbVisibleWaitHandler{region: region},
	}
	result, err := waitPoller.PollUntilDone(client, kt, cvt.SliceToPtr(cloudIDs), opt)
	if err != nil || result == nil {
		logs.Warnf("wait lb visible on cloud failed, treat all as invisible, region: %s, cloud_ids: %v, err: %v, "+
			"rid: %s", region, cloudIDs, err, kt.Rid)
		return slices.Clone(cloudIDs), nil
	}

	if len(result.base.UnknownCloudIDs) > 0 {
		logs.Warnf("lb still invisible on cloud after waiting, region: %s, invisible: %v, cloud_ids: %v, rid: %s",
			region, result.base.UnknownCloudIDs, cloudIDs, kt.Rid)
		return result.base.UnknownCloudIDs, nil
	}
	return nil, result.lbs
}

var _ poller.PollingHandler[TCloud, clbVisiblePollResult, clbVisibleDoneResult] = new(clbVisibleWaitHandler)

type clbVisibleWaitHandler struct {
	region string
}

// Done 全部云 ID 可查询时返回 true；结果始终非空，仍查不到的云 ID 放入 UnknownCloudIDs。
// 全部可见时带上本轮查到的 CLB。
func (h *clbVisibleWaitHandler) Done(polled clbVisiblePollResult) (bool, *clbVisibleDoneResult) {
	result := &clbVisibleDoneResult{
		base: poller.BaseDoneResult{
			SuccessCloudIDs: make([]string, 0, len(polled.visible)),
			UnknownCloudIDs: make([]string, 0),
		},
	}
	for cloudID, visible := range polled.visible {
		if visible {
			result.base.SuccessCloudIDs = append(result.base.SuccessCloudIDs, cloudID)
			continue
		}
		result.base.UnknownCloudIDs = append(result.base.UnknownCloudIDs, cloudID)
	}
	slices.Sort(result.base.SuccessCloudIDs)
	slices.Sort(result.base.UnknownCloudIDs)

	allVisible := len(result.base.UnknownCloudIDs) == 0
	if allVisible {
		result.lbs = polled.lbs
	}
	return allVisible, result
}

// Poll 按批查询云上 CLB，返回每个目标云 ID 是否可查询，以及本轮查到的目标 CLB。
func (h *clbVisibleWaitHandler) Poll(client TCloud, kt *kit.Kit, cloudIDs []*string) (clbVisiblePollResult, error) {
	visibleMap := make(map[string]bool, len(cloudIDs))
	for _, cloudID := range cloudIDs {
		visibleMap[cvt.PtrToVal(cloudID)] = false
	}
	collected := make([]typelb.TCloudClb, 0, len(cloudIDs))

	for _, batch := range slice.Split(cvt.PtrToSlice(cloudIDs), constant.TCLBDescribeMax) {
		opt := &typelb.TCloudListOption{Region: h.region, CloudIDs: batch}
		lbs, err := client.ListLoadBalancer(kt, opt)
		if err != nil {
			logs.Warnf("list lb for visible wait failed, region: %s, cloud_ids: %v, err: %v, rid: %s",
				h.region, batch, err, kt.Rid)
			return clbVisiblePollResult{}, err
		}
		for _, lb := range lbs {
			cloudID := lb.GetCloudID()
			visible, ok := visibleMap[cloudID]
			if !ok || visible {
				continue
			}
			visibleMap[cloudID] = true
			collected = append(collected, lb)
		}
	}
	return clbVisiblePollResult{visible: visibleMap, lbs: collected}, nil
}
