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
	"fmt"

	typelb "hcm/pkg/adaptor/types/load-balancer"
	protolb "hcm/pkg/api/hc-service/load-balancer"
	"hcm/pkg/criteria/constant"
	"hcm/pkg/criteria/errf"
	"hcm/pkg/kit"
	"hcm/pkg/logs"
	"hcm/pkg/rest"
	cvt "hcm/pkg/tools/converter"

	tclb "github.com/tencentcloud/tencentcloud-sdk-go/tencentcloud/clb/v20180317"
)

// clusterResourceDescriber 查询独占集群闲置VIP翻页所需的最小 adaptor 能力集合，从完整的 tcloud.TCloud 接口中
// 按需截取，便于单元测试注入 fake 实现，无需实现整个 tcloud.TCloud 接口。
type clusterResourceDescriber interface {
	// DescribeClusterResources 查询独占集群内资源（含 VIP 闲置状态）
	DescribeClusterResources(kt *kit.Kit, opt *typelb.TCloudDescribeClusterResourcesOption) (
		*tclb.DescribeClusterResourcesResponseParams, error)
}

// TCloudDescribeClusterIdleVips 查询独占集群当前闲置的VIP列表，内部自动翻页取全并对结果去重，供cloud-server
// 业务视角闲置VIP查询接口调用，实时查云、不落库。
func (svc *clbSvc) TCloudDescribeClusterIdleVips(cts *rest.Contexts) (any, error) {
	req := new(protolb.TCloudDescribeClusterIdleVipsReq)
	if err := cts.DecodeInto(req); err != nil {
		return nil, errf.NewFromErr(errf.DecodeRequestFailed, err)
	}

	if err := req.Validate(); err != nil {
		return nil, errf.NewFromErr(errf.InvalidParameter, err)
	}

	client, err := svc.ad.TCloud(cts.Kit, req.AccountID)
	if err != nil {
		return nil, err
	}

	vips, err := listAllClusterIdleVips(cts.Kit, client, req.Region, req.ClusterID, req.Vip)
	if err != nil {
		return nil, err
	}

	return &protolb.TCloudDescribeClusterIdleVipsResult{Count: uint64(len(vips)), Details: vips}, nil
}

// listAllClusterIdleVips 按固定页大小翻页查询独占集群内全部闲置VIP，对结果按VIP去重后返回；vip 非空时只查询该VIP。
func listAllClusterIdleVips(kt *kit.Kit, client clusterResourceDescriber, region, clusterID, vip string) (
	[]string, error) {

	idle := true
	limit := uint64(constant.BatchOperationMaxLimit)
	offset := uint64(0)
	vipSet := make(map[string]struct{})

	var vipFilter []string
	if vip != "" {
		vipFilter = []string{vip}
	}

	for page := 0; ; page++ {
		if page >= constant.ExclusiveClusterIdleVipMaxPages {
			return nil, fmt.Errorf("describe cluster(%s) idle vips exceeds max pages(%d)", clusterID,
				constant.ExclusiveClusterIdleVipMaxPages)
		}

		result, err := client.DescribeClusterResources(kt, &typelb.TCloudDescribeClusterResourcesOption{
			Region:    region,
			ClusterID: []string{clusterID},
			Vip:       vipFilter,
			Idle:      &idle,
			Limit:     &limit,
			Offset:    &offset,
		})
		if err != nil {
			logs.Errorf("describe cluster(%s) idle vips failed, offset: %d, err: %v, rid: %s", clusterID, offset,
				err, kt.Rid)
			return nil, err
		}
		if result == nil {
			break
		}

		for _, one := range result.ClusterResourceSet {
			if one == nil || cvt.PtrToVal(one.Vip) == "" {
				continue
			}
			vipSet[cvt.PtrToVal(one.Vip)] = struct{}{}
		}

		// 翻页过程中闲置状态可能变化导致云侧 TotalCount 不稳定，以本页是否取满作为终止依据
		if uint64(len(result.ClusterResourceSet)) < limit {
			break
		}
		offset += limit
	}

	return cvt.MapKeyToStringSlice(vipSet), nil
}
