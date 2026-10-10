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
	"errors"
	"fmt"
	"testing"
	"time"

	"hcm/pkg/adaptor/poller"
	typelb "hcm/pkg/adaptor/types/load-balancer"
	"hcm/pkg/criteria/constant"
	"hcm/pkg/kit"
	cvt "hcm/pkg/tools/converter"
	"hcm/pkg/tools/retry"

	"github.com/stretchr/testify/assert"
	clb "github.com/tencentcloud/tencentcloud-sdk-go/tencentcloud/clb/v20180317"
)

const testWaitTimeoutSec = 1

type mockLBLister struct {
	calls [][]string
	// list 入参为第几次调用（从 1 开始）和本次查询的云 ID，返回云上可查询到的云 ID
	list func(call int, cloudIDs []string) ([]string, error)
}

func (m *mockLBLister) ListLoadBalancer(_ *kit.Kit, opt *typelb.TCloudListOption) ([]typelb.TCloudClb, error) {
	m.calls = append(m.calls, opt.CloudIDs)
	visibleIDs, err := m.list(len(m.calls), opt.CloudIDs)
	if err != nil {
		return nil, err
	}
	lbs := make([]typelb.TCloudClb, 0, len(visibleIDs))
	for _, id := range visibleIDs {
		lbs = append(lbs, typelb.TCloudClb{LoadBalancer: &clb.LoadBalancer{LoadBalancerId: cvt.ValToPtr(id)}})
	}
	return lbs, nil
}

func testWaitOption() *poller.PollUntilDoneOption {
	return &poller.PollUntilDoneOption{
		TimeoutTimeSecond: testWaitTimeoutSec,
		Retry:             retry.NewRetryPolicy(constant.TCLBVisibleWaitImmuneCount, [2]uint{50, 100}),
	}
}

func genCloudIDs(n int) []string {
	ids := make([]string, n)
	for i := range ids {
		ids[i] = fmt.Sprintf("lb-%03d", i)
	}
	return ids
}

func TestWaitLoadBalancerVisible_FirstRoundVisible(t *testing.T) {
	cloudIDs := genCloudIDs(3)
	lister := &mockLBLister{list: func(_ int, ids []string) ([]string, error) { return ids, nil }}

	missing := WaitLoadBalancerVisible(kit.New(), lister, "ap-guangzhou", cloudIDs, testWaitOption())

	assert.Empty(t, missing)
	assert.Len(t, lister.calls, 1)
}

func TestWaitLoadBalancerVisible_VisibleAtSecondRound(t *testing.T) {
	cloudIDs := genCloudIDs(3)
	lister := &mockLBLister{list: func(call int, ids []string) ([]string, error) {
		if call == 1 {
			return nil, nil
		}
		return ids, nil
	}}

	missing := WaitLoadBalancerVisible(kit.New(), lister, "ap-guangzhou", cloudIDs, testWaitOption())

	assert.Empty(t, missing)
	assert.Len(t, lister.calls, 2)
}

func TestWaitLoadBalancerVisible_TimeoutWithInvisible(t *testing.T) {
	cloudIDs := genCloudIDs(3)
	invisibleID := cloudIDs[1]
	lister := &mockLBLister{list: func(_ int, ids []string) ([]string, error) {
		visible := make([]string, 0, len(ids))
		for _, id := range ids {
			if id != invisibleID {
				visible = append(visible, id)
			}
		}
		return visible, nil
	}}

	start := time.Now()
	missing := WaitLoadBalancerVisible(kit.New(), lister, "ap-guangzhou", cloudIDs, testWaitOption())
	elapsed := time.Since(start)

	assert.Equal(t, []string{invisibleID}, missing)
	assert.Greater(t, len(lister.calls), 1)
	assert.LessOrEqual(t, elapsed, (testWaitTimeoutSec+3)*time.Second)
}

func TestWaitLoadBalancerVisible_SplitBatch(t *testing.T) {
	cloudIDs := genCloudIDs(25)
	lister := &mockLBLister{list: func(_ int, ids []string) ([]string, error) { return ids, nil }}

	missing := WaitLoadBalancerVisible(kit.New(), lister, "ap-guangzhou", cloudIDs, testWaitOption())

	assert.Empty(t, missing)
	assert.Len(t, lister.calls, 2)
	for _, ids := range lister.calls {
		assert.LessOrEqual(t, len(ids), constant.TCLBDescribeMax)
	}
	assert.Len(t, lister.calls[0], constant.TCLBDescribeMax)
	assert.Len(t, lister.calls[1], 5)
}

func TestWaitLoadBalancerVisible_AlwaysError(t *testing.T) {
	cloudIDs := genCloudIDs(3)
	lister := &mockLBLister{list: func(_ int, _ []string) ([]string, error) {
		return nil, errors.New("RequestLimitExceeded")
	}}

	missing := WaitLoadBalancerVisible(kit.New(), lister, "ap-guangzhou", cloudIDs, testWaitOption())

	assert.ElementsMatch(t, cloudIDs, missing)
	assert.Greater(t, len(lister.calls), 1)
}

func TestWaitLoadBalancerVisible_EmptyCloudIDs(t *testing.T) {
	lister := &mockLBLister{list: func(_ int, ids []string) ([]string, error) { return ids, nil }}

	missing := WaitLoadBalancerVisible(kit.New(), lister, "ap-guangzhou", nil, testWaitOption())

	assert.Empty(t, missing)
	assert.Empty(t, lister.calls)
}
