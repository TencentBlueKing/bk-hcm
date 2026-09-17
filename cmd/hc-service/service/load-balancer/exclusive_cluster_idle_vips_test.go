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
	"errors"
	"sort"
	"strconv"
	"testing"

	typelb "hcm/pkg/adaptor/types/load-balancer"
	"hcm/pkg/kit"

	"github.com/stretchr/testify/require"
)

// fakeClusterResourceDescriber clusterResourceDescriber 的 fake 实现，按调用次数依次返回预置的分页结果。
type fakeClusterResourceDescriber struct {
	pages   []*typelb.TCloudDescribeClusterResourcesResult
	err     error
	offsets []uint64
}

func (f *fakeClusterResourceDescriber) DescribeClusterResources(_ *kit.Kit,
	opt *typelb.TCloudDescribeClusterResourcesOption) (*typelb.TCloudDescribeClusterResourcesResult, error) {

	f.offsets = append(f.offsets, *opt.Offset)
	if f.err != nil {
		return nil, f.err
	}

	idx := len(f.offsets) - 1
	if idx >= len(f.pages) {
		return nil, errors.New("fakeClusterResourceDescriber: no more pages configured")
	}
	return f.pages[idx], nil
}

// TestListAllClusterIdleVips_SinglePage 单页即可取全：total_count 小于等于页大小时只请求一页。
func TestListAllClusterIdleVips_SinglePage(t *testing.T) {
	fake := &fakeClusterResourceDescriber{
		pages: []*typelb.TCloudDescribeClusterResourcesResult{
			{
				TotalCount: 2,
				Resources: []typelb.TCloudClusterResource{
					{Vip: "1.1.1.1", Idle: true},
					{Vip: "1.1.1.2", Idle: true},
				},
			},
		},
	}

	vips, err := listAllClusterIdleVips(testKit(), fake, "ap-guangzhou", "tgw-1")
	require.NoError(t, err)
	sort.Strings(vips)
	require.Equal(t, []string{"1.1.1.1", "1.1.1.2"}, vips)
	require.Len(t, fake.offsets, 1)
	require.EqualValues(t, 0, fake.offsets[0])
}

// TestListAllClusterIdleVips_MultiPage 多页翻页取全：total_count 超过单页大小时自动翻页直到取全，且对重复
// VIP 去重。
func TestListAllClusterIdleVips_MultiPage(t *testing.T) {
	total := clusterIdleVipPageLimit + 1
	page1 := make([]typelb.TCloudClusterResource, 0, clusterIdleVipPageLimit)
	for i := uint64(0); i < clusterIdleVipPageLimit; i++ {
		page1 = append(page1, typelb.TCloudClusterResource{Vip: "vip-" + strconv.FormatUint(i, 10), Idle: true})
	}
	// 第二页与第一页有一个重复 VIP，验证去重逻辑。
	page2 := []typelb.TCloudClusterResource{
		{Vip: "vip-0", Idle: true},
		{Vip: "vip-last", Idle: true},
	}

	fake := &fakeClusterResourceDescriber{
		pages: []*typelb.TCloudDescribeClusterResourcesResult{
			{TotalCount: total, Resources: page1},
			{TotalCount: total, Resources: page2},
		},
	}

	vips, err := listAllClusterIdleVips(testKit(), fake, "ap-guangzhou", "tgw-1")
	require.NoError(t, err)
	require.Len(t, vips, int(clusterIdleVipPageLimit)+1)
	require.Len(t, fake.offsets, 2)
	require.EqualValues(t, 0, fake.offsets[0])
	require.EqualValues(t, clusterIdleVipPageLimit, fake.offsets[1])
}

// TestListAllClusterIdleVips_TotalCountZero total_count 为 0 时，只请求一页且返回空列表。
func TestListAllClusterIdleVips_TotalCountZero(t *testing.T) {
	fake := &fakeClusterResourceDescriber{
		pages: []*typelb.TCloudDescribeClusterResourcesResult{
			{TotalCount: 0, Resources: []typelb.TCloudClusterResource{}},
		},
	}

	vips, err := listAllClusterIdleVips(testKit(), fake, "ap-guangzhou", "tgw-1")
	require.NoError(t, err)
	require.Len(t, vips, 0)
	require.Len(t, fake.offsets, 1)
}

// TestListAllClusterIdleVips_AdaptorError adaptor 调用失败时错误直接透传，不吞掉错误继续翻页。
func TestListAllClusterIdleVips_AdaptorError(t *testing.T) {
	fake := &fakeClusterResourceDescriber{err: errors.New("mock adaptor error")}

	_, err := listAllClusterIdleVips(testKit(), fake, "ap-guangzhou", "tgw-1")
	require.Error(t, err)
	require.Contains(t, err.Error(), "mock adaptor error")
	require.Len(t, fake.offsets, 1)
}
