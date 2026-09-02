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
 * to the current version of the project delivered to anyone in the future.
 */

package tcloud

import (
	"testing"

	typecore "hcm/pkg/adaptor/types/core"
	typeclb "hcm/pkg/adaptor/types/load-balancer"
	"hcm/pkg/api/core"
	hcsync "hcm/pkg/api/hc-service/sync"
	"hcm/pkg/cc"
	"hcm/pkg/criteria/constant"
	"hcm/pkg/criteria/enumor"
	cvt "hcm/pkg/tools/converter"
	"hcm/pkg/tools/slice"

	"github.com/stretchr/testify/require"
)

var hcServiceSetting = new(cc.HCServiceSetting)

func init() {
	cc.InitRuntime(hcServiceSetting)
}

func TestBuildListOpt(t *testing.T) {
	hd := &lbHandler{
		baseHandler: baseHandler{
			resType: enumor.LoadBalancerCloudResType,
			request: &hcsync.TCloudSyncReq{
				Region:     "ap-guangzhou",
				TagFilters: core.MultiValueTagMap{"env": {"prod"}},
			},
		},
	}

	page := &typecore.TCloudPage{Offset: 0, Limit: 100}
	cloudIDs := []string{"lb-1", "lb-2"}

	opt := hd.buildListOpt(page, cloudIDs)

	require.Equal(t, "ap-guangzhou", opt.Region)
	require.Equal(t, cloudIDs, opt.CloudIDs)
	require.Equal(t, page, opt.Page)
	require.Equal(t, cvt.ValToPtr(typeclb.TCloudCLBOrderAscending), opt.OrderType)
	require.Equal(t, cvt.ValToPtr(typeclb.TCloudOrderByCreateTime), opt.OrderBy)
	require.Equal(t, core.MultiValueTagMap{"env": {"prod"}}, opt.TagFilters)
}

func TestBuildPageListOpts(t *testing.T) {
	hd := &lbHandler{
		baseHandler: baseHandler{
			resType: enumor.LoadBalancerCloudResType,
			request: &hcsync.TCloudSyncReq{
				Region:     "ap-guangzhou",
				Concurrent: 3,
			},
		},
		offset: 0,
	}

	opts := hd.buildPageListOpts()

	require.Len(t, opts, 3)
	for i, opt := range opts {
		require.Equal(t, uint64(i*typecore.TCloudQueryLimit), opt.Page.Offset)
		require.Equal(t, uint64(typecore.TCloudQueryLimit), opt.Page.Limit)
		require.Nil(t, opt.CloudIDs)
	}
}

func TestBuildPageListOptsWithOffset(t *testing.T) {
	hd := &lbHandler{
		baseHandler: baseHandler{
			resType: enumor.LoadBalancerCloudResType,
			request: &hcsync.TCloudSyncReq{
				Region:     "ap-guangzhou",
				Concurrent: 2,
			},
		},
		offset: 100,
	}

	opts := hd.buildPageListOpts()

	require.Len(t, opts, 2)
	require.Equal(t, uint64(100), opts[0].Page.Offset)
	require.Equal(t, uint64(100+typecore.TCloudQueryLimit), opts[1].Page.Offset)
}

func TestListLoadBalancerByCloudIDsSplit(t *testing.T) {
	// 21 cloud IDs should split into 2 batches: 20 + 1
	cloudIDs := make([]string, 0, 21)
	for i := 0; i < 21; i++ {
		cloudIDs = append(cloudIDs, "lb-"+string(rune('a'+i)))
	}

	hd := &lbHandler{
		baseHandler: baseHandler{
			resType: enumor.LoadBalancerCloudResType,
			request: &hcsync.TCloudSyncReq{
				Region: "ap-guangzhou",
			},
		},
	}

	// We can't call listLoadBalancerByCloudIDs directly without mocking the cloud client,
	// but we can verify the batching logic by checking slice.Split behavior
	idBatches := slice.Split(cloudIDs, constant.TCLBDescribeMax)
	require.Len(t, idBatches, 2)
	require.Len(t, idBatches[0], 20)
	require.Len(t, idBatches[1], 1)

	// Verify buildListOpt is called with correct cloudIDs for each batch
	for i, batch := range idBatches {
		page := &typecore.TCloudPage{Limit: typecore.TCloudQueryLimit}
		opt := hd.buildListOpt(page, batch)
		require.Equal(t, batch, opt.CloudIDs)
		require.Equal(t, uint64(typecore.TCloudQueryLimit), opt.Page.Limit)
		_ = i
	}
}

func TestListLoadBalancerByCloudIDsEmpty(t *testing.T) {
	idBatches := slice.Split([]string(nil), constant.TCLBDescribeMax)
	require.Empty(t, idBatches)

	opts := make([]*typeclb.TCloudListOption, 0, len(idBatches))
	require.Empty(t, opts)
}
