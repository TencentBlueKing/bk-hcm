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

// LoadBalancerLister 只包含查询 CLB 列表能力的适配器
type LoadBalancerLister interface {
	ListLoadBalancer(kt *kit.Kit, opt *typelb.TCloudListOption) ([]typelb.TCloudClb, error)
}

var _ LoadBalancerLister = TCloud(nil)

// WaitLoadBalancerVisible 等待 CLB 在云上可查询，返回到达等待上限时仍查不到的云 ID，不返回错误。
// opt 为空时使用 types.NewCLBVisibleWaitPollerOption()。
func WaitLoadBalancerVisible(kt *kit.Kit, lister LoadBalancerLister, region string, cloudIDs []string,
	opt *poller.PollUntilDoneOption) []string {

	if len(cloudIDs) == 0 {
		return nil
	}
	if opt == nil {
		opt = types.NewCLBVisibleWaitPollerOption()
	}

	waitPoller := poller.Poller[LoadBalancerLister, map[string]bool, poller.BaseDoneResult]{
		Handler: &clbVisibleWaitHandler{region: region},
	}
	result, err := waitPoller.PollUntilDone(lister, kt, cvt.SliceToPtr(cloudIDs), opt)
	if err != nil || result == nil {
		logs.Warnf("wait lb visible on cloud failed, treat all as invisible, region: %s, cloud_ids: %v, err: %v, "+
			"rid: %s", region, cloudIDs, err, kt.Rid)
		return slices.Clone(cloudIDs)
	}

	if len(result.UnknownCloudIDs) > 0 {
		logs.Warnf("lb still invisible on cloud after waiting, region: %s, invisible: %v, cloud_ids: %v, rid: %s",
			region, result.UnknownCloudIDs, cloudIDs, kt.Rid)
	}
	return result.UnknownCloudIDs
}

var _ poller.PollingHandler[LoadBalancerLister, map[string]bool, poller.BaseDoneResult] = new(clbVisibleWaitHandler)

type clbVisibleWaitHandler struct {
	region string
}

// Done 全部云 ID 可查询时返回 true；结果始终非空，仍查不到的云 ID 放入 UnknownCloudIDs
func (h *clbVisibleWaitHandler) Done(visibleMap map[string]bool) (bool, *poller.BaseDoneResult) {
	result := &poller.BaseDoneResult{
		SuccessCloudIDs: make([]string, 0, len(visibleMap)),
		UnknownCloudIDs: make([]string, 0),
	}
	for cloudID, visible := range visibleMap {
		if visible {
			result.SuccessCloudIDs = append(result.SuccessCloudIDs, cloudID)
			continue
		}
		result.UnknownCloudIDs = append(result.UnknownCloudIDs, cloudID)
	}
	slices.Sort(result.SuccessCloudIDs)
	slices.Sort(result.UnknownCloudIDs)

	return len(result.UnknownCloudIDs) == 0, result
}

// Poll 按批查询云上 CLB，返回每个目标云 ID 是否可查询
func (h *clbVisibleWaitHandler) Poll(lister LoadBalancerLister, kt *kit.Kit, cloudIDs []*string) (
	map[string]bool, error) {

	visibleMap := make(map[string]bool, len(cloudIDs))
	for _, cloudID := range cloudIDs {
		visibleMap[cvt.PtrToVal(cloudID)] = false
	}

	for _, batch := range slice.Split(cvt.PtrToSlice(cloudIDs), constant.TCLBDescribeMax) {
		opt := &typelb.TCloudListOption{Region: h.region, CloudIDs: batch}
		lbs, err := lister.ListLoadBalancer(kt, opt)
		if err != nil {
			logs.Warnf("list lb for visible wait failed, region: %s, cloud_ids: %v, err: %v, rid: %s",
				h.region, batch, err, kt.Rid)
			return nil, err
		}
		for _, lb := range lbs {
			if _, ok := visibleMap[lb.GetCloudID()]; ok {
				visibleMap[lb.GetCloudID()] = true
			}
		}
	}
	return visibleMap, nil
}
