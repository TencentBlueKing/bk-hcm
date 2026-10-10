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
	"hcm/pkg/criteria/errf"
	cvt "hcm/pkg/tools/converter"

	"github.com/stretchr/testify/require"
)

// newVipCheckHandler 构造只带请求体的申请单处理器，用于不触达下游服务的 VIP 校验分支。
func newVipCheckHandler(vip *string, cloudClusterIDs []string) *ApplicationOfCreateTCloudLB {
	req := &hclb.TCloudLoadBalancerCreateReq{AccountID: "0000001", Region: "ap-guangzhou"}
	req.Vip = vip
	req.CloudClusterIDs = cloudClusterIDs
	return &ApplicationOfCreateTCloudLB{req: req}
}

// TestCheckVipIdle_NoVip 未指定VIP时直接通过，不查云。
func TestCheckVipIdle_NoVip(t *testing.T) {
	require.NoError(t, newVipCheckHandler(nil, []string{"tgw-1", "tgw-2"}).checkVipIdle())
	require.NoError(t, newVipCheckHandler(cvt.ValToPtr(""), nil).checkVipIdle())
}

// TestCheckVipIdle_ClusterIDCountNotOne 指定了VIP但四层集群ID不是恰好一个时返回 InvalidParameter，而不是越界。
func TestCheckVipIdle_ClusterIDCountNotOne(t *testing.T) {
	for _, ids := range [][]string{nil, {}, {"tgw-1", "tgw-2"}} {
		err := newVipCheckHandler(cvt.ValToPtr("1.1.1.1"), ids).checkVipIdle()
		require.Error(t, err)
		require.Equal(t, errf.InvalidParameter, err.(*errf.ErrorF).Code)
	}
}
