/*
 * TencentBlueKing is pleased to support the open source community by making
 * 蓝鲸智云 - 混合云管理平台 (BlueKing - Hybrid Cloud Management System) available.
 * Copyright (C) 2024 THL A29 Limited,
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
	"testing"

	typeslb "hcm/pkg/adaptor/types/load-balancer"
	corelb "hcm/pkg/api/core/cloud/load-balancer"
	"hcm/pkg/criteria/enumor"
	cvt "hcm/pkg/tools/converter"

	"github.com/stretchr/testify/require"
	tclb "github.com/tencentcloud/tencentcloud-sdk-go/tencentcloud/clb/v20180317"
)

func newTestClb(clusterTag *string, clusterIDs []string) typeslb.TCloudClb {
	return typeslb.TCloudClb{LoadBalancer: &tclb.LoadBalancer{
		ClusterTag: clusterTag,
		ClusterIds: cvt.SliceToPtr(clusterIDs),
	}}
}

// TestBuildExclusiveClusterExtension_NonExclusive 非独占型实例 exclusive 为 false，clusters 为空数组。
func TestBuildExclusiveClusterExtension_NonExclusive(t *testing.T) {
	cloud := newTestClb(nil, nil)

	exclusive, clusters := buildExclusiveClusterExtension(cloud, map[string]corelb.BaseExclusiveCluster{})

	require.False(t, cvt.PtrToVal(exclusive))
	require.NotNil(t, clusters)
	require.Len(t, clusters, 0)
}

// TestBuildExclusiveClusterExtension_OnlyL4 只使用四层独占集群时数组只含 TGW 类型元素。
func TestBuildExclusiveClusterExtension_OnlyL4(t *testing.T) {
	cloud := newTestClb(nil, []string{"tgw-1"})
	clusterMap := map[string]corelb.BaseExclusiveCluster{
		"tgw-1": {ID: "000001", Name: "tgw-cluster-1", ClusterType: enumor.TGWClusterType},
	}

	exclusive, clusters := buildExclusiveClusterExtension(cloud, clusterMap)

	require.True(t, cvt.PtrToVal(exclusive))
	require.Len(t, clusters, 1)
	require.Equal(t, "tgw-1", clusters[0].CloudClusterID)
	require.Equal(t, "000001", clusters[0].ClusterID)
	require.Equal(t, "tgw-cluster-1", clusters[0].ClusterName)
	require.Equal(t, string(enumor.TGWClusterType), clusters[0].ClusterType)
}

// TestBuildExclusiveClusterExtension_MixedL4AndL7ClusterIds 云上 ClusterIds 混合了四层与七层落地 ID，按本地表
// 记录的 cluster_type 分别正确标注。
func TestBuildExclusiveClusterExtension_MixedL4AndL7ClusterIds(t *testing.T) {
	cloud := newTestClb(nil, []string{"tgw-1", "stgw-1"})
	clusterMap := map[string]corelb.BaseExclusiveCluster{
		"tgw-1":  {ID: "000001", Name: "tgw-cluster-1", ClusterType: enumor.TGWClusterType},
		"stgw-1": {ID: "000002", Name: "stgw-cluster-1", ClusterType: enumor.STGWClusterType, ClusterTag: "tag-1"},
	}

	exclusive, clusters := buildExclusiveClusterExtension(cloud, clusterMap)

	require.True(t, cvt.PtrToVal(exclusive))
	require.Len(t, clusters, 2)

	byID := make(map[string]corelb.TCloudExtensionCluster, len(clusters))
	for _, c := range clusters {
		byID[c.CloudClusterID] = c
	}
	require.Equal(t, string(enumor.TGWClusterType), byID["tgw-1"].ClusterType)
	require.Equal(t, string(enumor.STGWClusterType), byID["stgw-1"].ClusterType)
	require.Equal(t, "tag-1", byID["stgw-1"].ClusterTag)
}

// TestBuildExclusiveClusterExtension_ClusterTagWithoutLandingSTGW 云侧只回传七层标签、未回传具体落地 ID 时，
// 补一条只有 cluster_tag/cluster_type 有值的元素。
func TestBuildExclusiveClusterExtension_ClusterTagWithoutLandingSTGW(t *testing.T) {
	cloud := newTestClb(cvt.ValToPtr("tag-1"), []string{"tgw-1"})
	clusterMap := map[string]corelb.BaseExclusiveCluster{
		"tgw-1": {ID: "000001", Name: "tgw-cluster-1", ClusterType: enumor.TGWClusterType},
	}

	exclusive, clusters := buildExclusiveClusterExtension(cloud, clusterMap)

	require.True(t, cvt.PtrToVal(exclusive))
	require.Len(t, clusters, 2)

	var stgwElem *corelb.TCloudExtensionCluster
	for i := range clusters {
		if clusters[i].ClusterType == string(enumor.STGWClusterType) {
			stgwElem = &clusters[i]
		}
	}
	require.NotNil(t, stgwElem)
	require.Equal(t, "tag-1", stgwElem.ClusterTag)
	require.Equal(t, "", stgwElem.CloudClusterID)
	require.Equal(t, "", stgwElem.ClusterID)
}

// TestBuildExclusiveClusterExtension_ClusterIDNotSyncedLocally 本地表未同步或已删除时仍返回云上原值，层级按
// TGW 兜底。
func TestBuildExclusiveClusterExtension_ClusterIDNotSyncedLocally(t *testing.T) {
	cloud := newTestClb(nil, []string{"tgw-deleted"})

	exclusive, clusters := buildExclusiveClusterExtension(cloud, map[string]corelb.BaseExclusiveCluster{})

	require.True(t, cvt.PtrToVal(exclusive))
	require.Len(t, clusters, 1)
	require.Equal(t, "tgw-deleted", clusters[0].CloudClusterID)
	require.Equal(t, "", clusters[0].ClusterID)
	require.Equal(t, "", clusters[0].ClusterName)
}

// TestBuildExclusiveClusterExtension_ClusterTagWithLandingSTGW 云侧已回传具体落地 STGW ID 时，不再额外补占位
// 元素。
func TestBuildExclusiveClusterExtension_ClusterTagWithLandingSTGW(t *testing.T) {
	cloud := newTestClb(cvt.ValToPtr("tag-1"), []string{"stgw-1"})
	clusterMap := map[string]corelb.BaseExclusiveCluster{
		"stgw-1": {ID: "000002", Name: "stgw-cluster-1", ClusterType: enumor.STGWClusterType, ClusterTag: "tag-1"},
	}

	exclusive, clusters := buildExclusiveClusterExtension(cloud, clusterMap)

	require.True(t, cvt.PtrToVal(exclusive))
	require.Len(t, clusters, 1)
	require.Equal(t, "stgw-1", clusters[0].CloudClusterID)
}
