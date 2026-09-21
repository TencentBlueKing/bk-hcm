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
	"testing"

	hclb "hcm/pkg/api/hc-service/load-balancer"
	cvt "hcm/pkg/tools/converter"

	"github.com/stretchr/testify/require"
)

// TestRenderSlaType_Exclusive 独占型申请单规格展示为「独占型」，不受空 sla_type 影响。
func TestRenderSlaType_Exclusive(t *testing.T) {
	req := &hclb.TCloudLoadBalancerCreateReq{}
	req.Exclusive = cvt.ValToPtr(int64(1))
	req.SlaType = cvt.ValToPtr("")

	require.Equal(t, exclusiveSlaTypeName, renderSlaType(req))
}

// TestRenderSlaType_SharedWhenSlaTypeEmpty 非独占且 sla_type 为空时展示共享型。
func TestRenderSlaType_SharedWhenSlaTypeEmpty(t *testing.T) {
	req := &hclb.TCloudLoadBalancerCreateReq{}

	require.Equal(t, sharedSlaTypeName, renderSlaType(req))
}

// TestRenderSlaType_PerformanceCapacity 非独占且指定性能容量档位时透传档位编码。
func TestRenderSlaType_PerformanceCapacity(t *testing.T) {
	req := &hclb.TCloudLoadBalancerCreateReq{}
	req.SlaType = cvt.ValToPtr("clb.c2.medium")

	require.Equal(t, "clb.c2.medium", renderSlaType(req))
}

// TestFormatTgwClusterItsmLine_WithTag 本地表命中且带标签时同时展示名称、云上 ID、标签。
func TestFormatTgwClusterItsmLine_WithTag(t *testing.T) {
	got := formatTgwClusterItsmLine("tgw-cluster-1", "tgw-38feq8c6", "ziyan-chiji")

	require.Equal(t, "tgw-cluster-1（云上ID：tgw-38feq8c6，标签：ziyan-chiji）", got)
}

// TestFormatTgwClusterItsmLine_WithoutTag 本地表命中但未打标签时不拼接标签段。
func TestFormatTgwClusterItsmLine_WithoutTag(t *testing.T) {
	got := formatTgwClusterItsmLine("tgw-cluster-1", "tgw-38feq8c6", "")

	require.Equal(t, "tgw-cluster-1（云上ID：tgw-38feq8c6）", got)
}

// TestFormatTgwClusterItsmLine_LocalUnsynced 本地表未同步时名称使用占位文案。
func TestFormatTgwClusterItsmLine_LocalUnsynced(t *testing.T) {
	got := formatTgwClusterItsmLine("", "tgw-deleted", "")

	require.Equal(t, "本地未同步（云上ID：tgw-deleted）", got)
}

// TestRenderExclusiveClusterItems 独占集群 ITSM 表单项包含七层标签、指定 VIP、四层集群。
func TestRenderExclusiveClusterItems(t *testing.T) {
	items := renderExclusiveClusterItems("ziyan-serven", "1.1.1.1",
		[]string{"tgw-cluster-1（云上ID：tgw-38feq8c6，标签：ziyan-chiji）"})

	require.Equal(t, []formItem{
		{Label: "七层独占集群标签", Value: "ziyan-serven"},
		{Label: "指定VIP", Value: "1.1.1.1"},
		{Label: "四层独占集群", Value: "tgw-cluster-1（云上ID：tgw-38feq8c6，标签：ziyan-chiji）"},
	}, items)
}

// TestRenderExclusiveClusterItems_MultipleTgw 多个四层集群用分号拼成一行，避免 ITSM 换行丢失标签前缀。
func TestRenderExclusiveClusterItems_MultipleTgw(t *testing.T) {
	items := renderExclusiveClusterItems("", "", []string{
		"tgw-cluster-1（云上ID：tgw-1）",
		"tgw-cluster-2（云上ID：tgw-2，标签：ziyan-chiji）",
	})

	require.Equal(t, "tgw-cluster-1（云上ID：tgw-1）；tgw-cluster-2（云上ID：tgw-2，标签：ziyan-chiji）",
		items[2].Value)
}

// TestRenderExclusiveClusterItems_EmptyFields 七层标签、VIP、四层集群都为空时使用占位符。
func TestRenderExclusiveClusterItems_EmptyFields(t *testing.T) {
	items := renderExclusiveClusterItems("", "", nil)

	require.Equal(t, []formItem{
		{Label: "七层独占集群标签", Value: emptyItsmValue},
		{Label: "指定VIP", Value: emptyItsmValue},
		{Label: "四层独占集群", Value: emptyItsmValue},
	}, items)
}
