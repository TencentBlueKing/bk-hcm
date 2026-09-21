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

package application

import (
	"context"
	"testing"

	"hcm/pkg/api/core"
	corelb "hcm/pkg/api/core/cloud/load-balancer"
	dataproto "hcm/pkg/api/data-service/cloud"
	"hcm/pkg/criteria/enumor"
	"hcm/pkg/kit"

	"github.com/stretchr/testify/require"
	"github.com/tidwall/gjson"
)

func testKit() *kit.Kit {
	return &kit.Kit{Ctx: context.Background()}
}

// fakeExclusiveClusterLister 测试用 exclusiveClusterLister fake 实现。
type fakeExclusiveClusterLister struct {
	result *dataproto.ExclusiveClusterListResult
	err    error
	called bool
}

func (f *fakeExclusiveClusterLister) ListExclusiveCluster(_ *kit.Kit, _ *core.ListReq) (
	*dataproto.ExclusiveClusterListResult, error) {
	f.called = true
	return f.result, f.err
}

// TestEnrichExclusiveClusterContent_NonExclusive 非独占型申请单详情的 clusters 为空数组。
func TestEnrichExclusiveClusterContent_NonExclusive(t *testing.T) {
	content := `{"bk_biz_id":213,"exclusive":0}`

	result := enrichExclusiveClusterContent(testKit(), &fakeExclusiveClusterLister{}, enumor.CreateLoadBalancer,
		content)

	require.True(t, gjson.Get(result, "clusters").IsArray())
	require.Len(t, gjson.Get(result, "clusters").Array(), 0)
}

// TestEnrichExclusiveClusterContent_NotCreateLoadBalancer 非 create_load_balancer 类型的申请单不做任何富化。
func TestEnrichExclusiveClusterContent_NotCreateLoadBalancer(t *testing.T) {
	content := `{"bk_biz_id":213,"exclusive":1,"cluster_tag":"ziyan-serven"}`

	result := enrichExclusiveClusterContent(testKit(), &fakeExclusiveClusterLister{}, enumor.CreateCvm, content)

	require.Equal(t, content, result)
}

// TestEnrichExclusiveClusterContent_OnlyClusterTag 七层集群在提单时通常没有具体集群 ID，只有 cluster_tag/
// cluster_type 有值，且不需要查询本地表。
func TestEnrichExclusiveClusterContent_OnlyClusterTag(t *testing.T) {
	lister := &fakeExclusiveClusterLister{}
	content := `{"bk_biz_id":213,"exclusive":1,"cluster_tag":"ziyan-serven"}`

	result := enrichExclusiveClusterContent(testKit(), lister, enumor.CreateLoadBalancer, content)

	require.False(t, lister.called, "should not query data service when only cluster_tag is set")
	clusters := gjson.Get(result, "clusters").Array()
	require.Len(t, clusters, 1)
	require.Equal(t, "ziyan-serven", clusters[0].Get("cluster_tag").String())
	require.Equal(t, "STGW", clusters[0].Get("cluster_type").String())
	require.Equal(t, "", clusters[0].Get("cloud_cluster_id").String())
	require.Equal(t, "", clusters[0].Get("cluster_id").String())
}

// TestEnrichExclusiveClusterContent_ClusterIDsFoundLocally 四层集群 ID 在本地表命中时补齐 cluster_id/
// cluster_name/cluster_tag。
func TestEnrichExclusiveClusterContent_ClusterIDsFoundLocally(t *testing.T) {
	lister := &fakeExclusiveClusterLister{
		result: &dataproto.ExclusiveClusterListResult{
			Count: 1,
			Details: []corelb.ExclusiveClusterRaw{
				{BaseExclusiveCluster: corelb.BaseExclusiveCluster{
					ID: "00000001", CloudID: "tgw-38feq8c6", Name: "tgw-cluster-1",
					ClusterTag: "ziyan-chiji",
				}},
			},
		},
	}
	content := `{"bk_biz_id":213,"exclusive":1,"cloud_cluster_ids":["tgw-38feq8c6"]}`

	result := enrichExclusiveClusterContent(testKit(), lister, enumor.CreateLoadBalancer, content)

	clusters := gjson.Get(result, "clusters").Array()
	require.Len(t, clusters, 1)
	require.Equal(t, "tgw-38feq8c6", clusters[0].Get("cloud_cluster_id").String())
	require.Equal(t, "00000001", clusters[0].Get("cluster_id").String())
	require.Equal(t, "tgw-cluster-1", clusters[0].Get("cluster_name").String())
	require.Equal(t, "ziyan-chiji", clusters[0].Get("cluster_tag").String())
	require.Equal(t, "TGW", clusters[0].Get("cluster_type").String())
}

// TestEnrichExclusiveClusterContent_ClusterIDNotSyncedLocally 本地表未同步/已删除的集群，cluster_id/cluster_name
// 为空字符串，cloud_cluster_id 仍返回原值。
func TestEnrichExclusiveClusterContent_ClusterIDNotSyncedLocally(t *testing.T) {
	lister := &fakeExclusiveClusterLister{result: &dataproto.ExclusiveClusterListResult{}}
	content := `{"bk_biz_id":213,"exclusive":1,"cloud_cluster_ids":["tgw-deleted"]}`

	result := enrichExclusiveClusterContent(testKit(), lister, enumor.CreateLoadBalancer, content)

	clusters := gjson.Get(result, "clusters").Array()
	require.Len(t, clusters, 1)
	require.Equal(t, "tgw-deleted", clusters[0].Get("cloud_cluster_id").String())
	require.Equal(t, "", clusters[0].Get("cluster_id").String())
	require.Equal(t, "", clusters[0].Get("cluster_name").String())
	require.Equal(t, "", clusters[0].Get("cluster_tag").String())
}

// TestEnrichExclusiveClusterContent_PreservesOtherFields 富化不影响 content 中原有字段。
func TestEnrichExclusiveClusterContent_PreservesOtherFields(t *testing.T) {
	content := `{"bk_biz_id":213,"exclusive":0,"vendor":"tcloud","name":"test-lb"}`

	result := enrichExclusiveClusterContent(testKit(), &fakeExclusiveClusterLister{}, enumor.CreateLoadBalancer,
		content)

	require.Equal(t, "tcloud", gjson.Get(result, "vendor").String())
	require.Equal(t, "test-lb", gjson.Get(result, "name").String())
	require.EqualValues(t, 213, gjson.Get(result, "bk_biz_id").Int())
}
