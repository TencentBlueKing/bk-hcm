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

package hclb

import (
	"testing"

	typelb "hcm/pkg/adaptor/types/load-balancer"
	"hcm/pkg/tools/converter"

	"github.com/stretchr/testify/require"
)

// baseValidSpec 返回一个满足基础结构校验（LoadBalancerType/CloudVpcID/LoadBalancerPassToTarget 必填）的公网
// 独占型请求基线，测试用例在此基础上覆盖独占相关字段。
func baseValidSpec() *TCloudLoadBalancerSpec {
	return &TCloudLoadBalancerSpec{
		LoadBalancerType:         typelb.OpenLoadBalancerType,
		CloudVpcID:               converter.ValToPtr("vpc-1"),
		LoadBalancerPassToTarget: converter.ValToPtr(true),
	}
}

// TestValidateSpec_AC001_OnlyCloudClusterIDs 只传四层集群的独占型请求通过结构校验。
func TestValidateSpec_AC001_OnlyCloudClusterIDs(t *testing.T) {
	spec := baseValidSpec()
	spec.Exclusive = converter.ValToPtr(int64(1))
	spec.CloudClusterIDs = []string{"tgw-38feq8c6"}

	require.NoError(t, spec.ValidateSpec())
}

// TestValidateSpec_AC002_BothClusterIdentifiersEmpty 独占型但 cluster_tag/cloud_cluster_ids 都为空。
func TestValidateSpec_AC002_BothClusterIdentifiersEmpty(t *testing.T) {
	spec := baseValidSpec()
	spec.Exclusive = converter.ValToPtr(int64(1))

	require.Error(t, spec.ValidateSpec())
}

// TestValidateSpec_AC003_OnlyClusterTag 只传七层标签的独占型请求通过结构校验。
func TestValidateSpec_AC003_OnlyClusterTag(t *testing.T) {
	spec := baseValidSpec()
	spec.Exclusive = converter.ValToPtr(int64(1))
	spec.ClusterTag = converter.ValToPtr("ziyan-serven")

	require.NoError(t, spec.ValidateSpec())
}

// TestValidateSpec_AC004_NonExclusiveWithClusterIdentifiers 非独占型却传了集群标识。
func TestValidateSpec_AC004_NonExclusiveWithClusterIdentifiers(t *testing.T) {
	t.Run("cluster_tag set", func(t *testing.T) {
		spec := baseValidSpec()
		spec.ClusterTag = converter.ValToPtr("ziyan-serven")
		require.Error(t, spec.ValidateSpec())
	})

	t.Run("cloud_cluster_ids set", func(t *testing.T) {
		spec := baseValidSpec()
		spec.CloudClusterIDs = []string{"tgw-38feq8c6"}
		require.Error(t, spec.ValidateSpec())
	})
}

// TestValidateSpec_AC005_ExclusiveInternal 独占型请求为内网负载均衡。
func TestValidateSpec_AC005_ExclusiveInternal(t *testing.T) {
	spec := baseValidSpec()
	spec.LoadBalancerType = typelb.InternalLoadBalancerType
	spec.CloudSubnetID = converter.ValToPtr("subnet-1")
	spec.Exclusive = converter.ValToPtr(int64(1))
	spec.CloudClusterIDs = []string{"tgw-38feq8c6"}

	require.Error(t, spec.ValidateSpec())
}

// TestValidateSpec_AC006_ExclusiveWithSlaType 独占型请求同时指定性能容量档位。
func TestValidateSpec_AC006_ExclusiveWithSlaType(t *testing.T) {
	spec := baseValidSpec()
	spec.Exclusive = converter.ValToPtr(int64(1))
	spec.CloudClusterIDs = []string{"tgw-38feq8c6"}
	spec.SlaType = converter.ValToPtr("clb.c2.medium")

	require.Error(t, spec.ValidateSpec())
}

// TestValidateSpec_AC007_VipWithoutCloudClusterIDs 仅七层场景指定 VIP。
func TestValidateSpec_AC007_VipWithoutCloudClusterIDs(t *testing.T) {
	spec := baseValidSpec()
	spec.Exclusive = converter.ValToPtr(int64(1))
	spec.ClusterTag = converter.ValToPtr("ziyan-serven")
	spec.Vip = converter.ValToPtr("1.1.1.1")

	require.Error(t, spec.ValidateSpec())
}

// TestValidateSpec_AC008_VipWithNonUniqueClusterIDs 指定 VIP 但四层集群 ID 不唯一（0 个或 2 个以上）。
func TestValidateSpec_AC008_VipWithNonUniqueClusterIDs(t *testing.T) {
	t.Run("two cloud_cluster_ids", func(t *testing.T) {
		spec := baseValidSpec()
		spec.Exclusive = converter.ValToPtr(int64(1))
		spec.CloudClusterIDs = []string{"tgw-1", "tgw-2"}
		spec.Vip = converter.ValToPtr("1.1.1.1")
		spec.RequireCount = converter.ValToPtr(uint64(1))

		require.Error(t, spec.ValidateSpec())
	})
}

// TestValidateSpec_AC009_VipWithRequireCountGreaterThanOne 指定 VIP 但购买数量大于 1。
func TestValidateSpec_AC009_VipWithRequireCountGreaterThanOne(t *testing.T) {
	spec := baseValidSpec()
	spec.Exclusive = converter.ValToPtr(int64(1))
	spec.CloudClusterIDs = []string{"tgw-38feq8c6"}
	spec.Vip = converter.ValToPtr("1.1.1.1")
	spec.RequireCount = converter.ValToPtr(uint64(2))

	require.Error(t, spec.ValidateSpec())
}

// TestValidateSpec_AC016_MultipleClusterIDsWithoutVip 多个四层集群 ID 表示随机分配，不传 VIP 时创建成功。
func TestValidateSpec_AC016_MultipleClusterIDsWithoutVip(t *testing.T) {
	spec := baseValidSpec()
	spec.Exclusive = converter.ValToPtr(int64(1))
	spec.CloudClusterIDs = []string{"tgw-1", "tgw-2"}

	require.NoError(t, spec.ValidateSpec())
}

// TestValidateSpec_AC017_VipHappyPath 指定 VIP 的合法组合可以通过结构校验。
func TestValidateSpec_AC017_VipHappyPath(t *testing.T) {
	spec := baseValidSpec()
	spec.Exclusive = converter.ValToPtr(int64(1))
	spec.CloudClusterIDs = []string{"tgw-38feq8c6"}
	spec.Vip = converter.ValToPtr("1.1.1.1")
	spec.RequireCount = converter.ValToPtr(uint64(1))

	require.NoError(t, spec.ValidateSpec())
}

// TestValidateSpec_AC018_SingleLineIspNotBandwidthPackage 单线运营商未使用共享带宽包计费。
func TestValidateSpec_AC018_SingleLineIspNotBandwidthPackage(t *testing.T) {
	spec := baseValidSpec()
	spec.Exclusive = converter.ValToPtr(int64(1))
	spec.ClusterTag = converter.ValToPtr("ziyan-serven")
	spec.VipIsp = converter.ValToPtr("CMCC")

	require.Error(t, spec.ValidateSpec())
}

// TestValidateSpec_AC019_BandwidthPackageChargeWithoutID 共享带宽包计费但未传带宽包 ID。
func TestValidateSpec_AC019_BandwidthPackageChargeWithoutID(t *testing.T) {
	spec := baseValidSpec()
	spec.Exclusive = converter.ValToPtr(int64(1))
	spec.ClusterTag = converter.ValToPtr("ziyan-serven")
	spec.InternetChargeType = converter.ValToPtr(typelb.BandwidthPackage)

	require.Error(t, spec.ValidateSpec())
}

// TestValidateSpec_AC015_LegacySharedRequestUnaffected 存量共享型请求（不传任何独占字段）不受影响。
func TestValidateSpec_AC015_LegacySharedRequestUnaffected(t *testing.T) {
	spec := baseValidSpec()

	require.NoError(t, spec.ValidateSpec())
}

// TestValidateSpec_CloudClusterIDsExceedLimit cloud_cluster_ids 数量超过上限时返回错误。
func TestValidateSpec_CloudClusterIDsExceedLimit(t *testing.T) {
	spec := baseValidSpec()
	spec.Exclusive = converter.ValToPtr(int64(1))

	ids := make([]string, cloudClusterIDsMaxLimit+1)
	for i := range ids {
		ids[i] = "tgw-" + string(rune('a'+i%26))
	}
	spec.CloudClusterIDs = ids

	require.Error(t, spec.ValidateSpec())
}
