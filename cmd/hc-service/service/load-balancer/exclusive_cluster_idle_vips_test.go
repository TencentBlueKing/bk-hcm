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
	"errors"
	"sort"
	"strconv"
	"testing"

	typelb "hcm/pkg/adaptor/types/load-balancer"
	"hcm/pkg/criteria/constant"
	"hcm/pkg/kit"
	cvt "hcm/pkg/tools/converter"

	"github.com/stretchr/testify/require"
	tclb "github.com/tencentcloud/tencentcloud-sdk-go/tencentcloud/clb/v20180317"
)

func testKit() *kit.Kit {
	return &kit.Kit{Ctx: context.Background()}
}

// fakeClusterResourceDescriber clusterResourceDescriber 的 fake 实现，按调用次数依次返回预置的分页结果。
type fakeClusterResourceDescriber struct {
	pages      []*tclb.DescribeClusterResourcesResponseParams
	err        error
	offsets    []uint64
	vips       [][]string
	clusterIDs [][]string
}

func (f *fakeClusterResourceDescriber) DescribeClusterResources(_ *kit.Kit,
	opt *typelb.TCloudDescribeClusterResourcesOption) (*tclb.DescribeClusterResourcesResponseParams, error) {

	f.offsets = append(f.offsets, *opt.Offset)
	f.vips = append(f.vips, opt.Vip)
	f.clusterIDs = append(f.clusterIDs, opt.ClusterID)
	if f.err != nil {
		return nil, f.err
	}

	idx := len(f.offsets) - 1
	if idx >= len(f.pages) {
		return nil, errors.New("fakeClusterResourceDescriber: no more pages configured")
	}
	return f.pages[idx], nil
}

// idleResource 构造一条闲置 VIP 资源。
func idleResource(vip string) *tclb.ClusterResource {
	return &tclb.ClusterResource{Vip: cvt.ValToPtr(vip), Idle: cvt.ValToPtr("True")}
}

// resourcePage 构造一页查询结果。
func resourcePage(total uint64, items ...*tclb.ClusterResource) *tclb.DescribeClusterResourcesResponseParams {
	return &tclb.DescribeClusterResourcesResponseParams{
		TotalCount:         cvt.ValToPtr(total),
		ClusterResourceSet: items,
	}
}

// TestListAllClusterIdleVips_SinglePage 单页即可取全：本页未取满时只请求一页，查询条件带上集群ID且不带 vip。
func TestListAllClusterIdleVips_SinglePage(t *testing.T) {
	fake := &fakeClusterResourceDescriber{
		pages: []*tclb.DescribeClusterResourcesResponseParams{
			resourcePage(2, idleResource("1.1.1.1"), idleResource("1.1.1.2")),
		},
	}

	vips, err := listAllClusterIdleVips(testKit(), fake, "ap-guangzhou", "tgw-1", "")
	require.NoError(t, err)
	sort.Strings(vips)
	require.Equal(t, []string{"1.1.1.1", "1.1.1.2"}, vips)
	require.Len(t, fake.offsets, 1)
	require.EqualValues(t, 0, fake.offsets[0])
	require.Equal(t, []string{"tgw-1"}, fake.clusterIDs[0])
	require.Empty(t, fake.vips[0])
}

// TestListAllClusterIdleVips_MultiPage 多页翻页取全：上一页取满时自动翻页直到取全，且对重复 VIP 去重。
func TestListAllClusterIdleVips_MultiPage(t *testing.T) {
	pageLimit := uint64(constant.BatchOperationMaxLimit)
	total := pageLimit + 1
	page1 := make([]*tclb.ClusterResource, 0, pageLimit)
	for i := uint64(0); i < pageLimit; i++ {
		page1 = append(page1, idleResource("vip-"+strconv.FormatUint(i, 10)))
	}

	fake := &fakeClusterResourceDescriber{
		pages: []*tclb.DescribeClusterResourcesResponseParams{
			resourcePage(total, page1...),
			// 第二页与第一页有一个重复 VIP，验证去重逻辑。
			resourcePage(total, idleResource("vip-0"), idleResource("vip-last")),
		},
	}

	vips, err := listAllClusterIdleVips(testKit(), fake, "ap-guangzhou", "tgw-1", "")
	require.NoError(t, err)
	require.Len(t, vips, int(pageLimit)+1)
	require.Len(t, fake.offsets, 2)
	require.EqualValues(t, 0, fake.offsets[0])
	require.EqualValues(t, pageLimit, fake.offsets[1])
}

// TestListAllClusterIdleVips_TotalCountZero total_count 为 0 时，只请求一页且返回空列表。
func TestListAllClusterIdleVips_TotalCountZero(t *testing.T) {
	fake := &fakeClusterResourceDescriber{
		pages: []*tclb.DescribeClusterResourcesResponseParams{resourcePage(0)},
	}

	vips, err := listAllClusterIdleVips(testKit(), fake, "ap-guangzhou", "tgw-1", "")
	require.NoError(t, err)
	require.Len(t, vips, 0)
	require.Len(t, fake.offsets, 1)
}

// TestListAllClusterIdleVips_AdaptorError adaptor 调用失败时错误直接透传，不吞掉错误继续翻页。
func TestListAllClusterIdleVips_AdaptorError(t *testing.T) {
	fake := &fakeClusterResourceDescriber{err: errors.New("mock adaptor error")}

	_, err := listAllClusterIdleVips(testKit(), fake, "ap-guangzhou", "tgw-1", "")
	require.Error(t, err)
	require.Contains(t, err.Error(), "mock adaptor error")
	require.Len(t, fake.offsets, 1)
}

// TestListAllClusterIdleVips_WithVip 指定 vip 时只查询该 VIP：查询条件带上 vip，且一次请求即可结束。
func TestListAllClusterIdleVips_WithVip(t *testing.T) {
	fake := &fakeClusterResourceDescriber{
		pages: []*tclb.DescribeClusterResourcesResponseParams{resourcePage(1, idleResource("1.1.1.2"))},
	}

	vips, err := listAllClusterIdleVips(testKit(), fake, "ap-guangzhou", "tgw-1", "1.1.1.2")
	require.NoError(t, err)
	require.Equal(t, []string{"1.1.1.2"}, vips)
	require.Equal(t, [][]string{{"1.1.1.2"}}, fake.vips)
	require.Len(t, fake.offsets, 1)
}

// fullIdleVipPage 构造一页取满的闲置 VIP 结果，totalCount 由调用方指定，用于模拟云侧返回的总数不可信。
func fullIdleVipPage(prefix string, totalCount uint64) *tclb.DescribeClusterResourcesResponseParams {
	resources := make([]*tclb.ClusterResource, 0, constant.BatchOperationMaxLimit)
	for i := 0; i < constant.BatchOperationMaxLimit; i++ {
		resources = append(resources, idleResource(prefix+strconv.Itoa(i)))
	}
	return resourcePage(totalCount, resources...)
}

// TestListAllClusterIdleVips_StopsOnPartialPage 以本页是否取满作为终止依据：即使云侧返回的 total_count 偏小，
// 只要上一页取满就继续翻页，直到出现未取满的页才结束。
func TestListAllClusterIdleVips_StopsOnPartialPage(t *testing.T) {
	fake := &fakeClusterResourceDescriber{
		pages: []*tclb.DescribeClusterResourcesResponseParams{
			fullIdleVipPage("a-", 0),
			resourcePage(0, idleResource("last")),
		},
	}

	vips, err := listAllClusterIdleVips(testKit(), fake, "ap-guangzhou", "tgw-1", "")
	require.NoError(t, err)
	require.Len(t, vips, constant.BatchOperationMaxLimit+1)
	require.Len(t, fake.offsets, 2)
}

// TestListAllClusterIdleVips_ExceedMaxPages 云侧一直返回满页时，翻页次数达到上限后返回错误，不无限翻页。
func TestListAllClusterIdleVips_ExceedMaxPages(t *testing.T) {
	pages := make([]*tclb.DescribeClusterResourcesResponseParams, 0, constant.ExclusiveClusterIdleVipMaxPages+1)
	for i := 0; i <= constant.ExclusiveClusterIdleVipMaxPages; i++ {
		pages = append(pages, fullIdleVipPage(strconv.Itoa(i)+"-", 0))
	}
	fake := &fakeClusterResourceDescriber{pages: pages}

	_, err := listAllClusterIdleVips(testKit(), fake, "ap-guangzhou", "tgw-1", "")
	require.Error(t, err)
	require.Contains(t, err.Error(), "exceeds max pages")
	require.Len(t, fake.offsets, constant.ExclusiveClusterIdleVipMaxPages)
}

// TestListAllClusterIdleVips_ExactlyMaxPages 恰好在最后一个允许的页取完（该页未取满）时不报错。
func TestListAllClusterIdleVips_ExactlyMaxPages(t *testing.T) {
	pages := make([]*tclb.DescribeClusterResourcesResponseParams, 0, constant.ExclusiveClusterIdleVipMaxPages)
	for i := 0; i < constant.ExclusiveClusterIdleVipMaxPages-1; i++ {
		pages = append(pages, fullIdleVipPage(strconv.Itoa(i)+"-", 0))
	}
	pages = append(pages, &tclb.DescribeClusterResourcesResponseParams{
		ClusterResourceSet: []*tclb.ClusterResource{idleResource("last")},
	})
	fake := &fakeClusterResourceDescriber{pages: pages}

	vips, err := listAllClusterIdleVips(testKit(), fake, "ap-guangzhou", "tgw-1", "")
	require.NoError(t, err)
	require.Len(t, vips, (constant.ExclusiveClusterIdleVipMaxPages-1)*constant.BatchOperationMaxLimit+1)
	require.Len(t, fake.offsets, constant.ExclusiveClusterIdleVipMaxPages)
}
