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
	adtypes "hcm/pkg/adaptor/types"
	adcore "hcm/pkg/adaptor/types/core"
	typelb "hcm/pkg/adaptor/types/load-balancer"
	"hcm/pkg/api/core"
	protolb "hcm/pkg/api/hc-service/load-balancer"
	"hcm/pkg/criteria/enumor"
	"hcm/pkg/criteria/errf"
	"hcm/pkg/dal/dao/tools"
	"hcm/pkg/kit"
	"hcm/pkg/logs"
	"hcm/pkg/runtime/filter"
	cvt "hcm/pkg/tools/converter"
)

// exclusiveClusterAdaptor 独占集群下云前复核所需的最小 adaptor 能力集合，从完整的 tcloud.TCloud 接口中按需
// 截取，便于单元测试注入 fake 实现，无需实现整个 tcloud.TCloud 接口。
type exclusiveClusterAdaptor interface {
	// ListBandwidthPackage 查询带宽包（用于获取出口）
	ListBandwidthPackage(kt *kit.Kit, opt *adtypes.TCloudListBwPkgOption) (*adtypes.TCloudListBwPkgResult, error)
	// DescribeClusterResources 查询独占集群内资源（含 VIP 闲置状态）
	DescribeClusterResources(kt *kit.Kit, opt *typelb.TCloudDescribeClusterResourcesOption) (
		*typelb.TCloudDescribeClusterResourcesResult, error)
}

// recheckExclusiveBeforeDeliver 下云前（真正调用云创建之前）完整重跑一遍提单时的独占集群校验（归属+出口一致性），
// 并额外复核指定 vip 是否仍闲置。判定逻辑与 cloud-server CheckReq 阶段一致，因服务边界不能跨服务调用 cloud-server
// 的实现，这里基于 svc.dataCli.Global.ListExclusiveCluster 重新实现一份等价逻辑。
func (svc *clbSvc) recheckExclusiveBeforeDeliver(kt *kit.Kit, tcloudAdpt exclusiveClusterAdaptor,
	req *protolb.TCloudLoadBalancerCreateReq) error {

	clusterTag := cvt.PtrToVal(req.ClusterTag)
	if err := svc.checkExclusiveClusterOwnership(kt, req.BkBizID, clusterTag, req.CloudClusterIDs); err != nil {
		return err
	}

	if req.InternetChargeType != nil && *req.InternetChargeType == typelb.BandwidthPackage {
		if err := svc.checkBandwidthPackageEgress(kt, tcloudAdpt, req); err != nil {
			return err
		}
	}

	if cvt.PtrToVal(req.Vip) != "" {
		return svc.checkVipIdle(kt, tcloudAdpt, req.Region, req.CloudClusterIDs[0], cvt.PtrToVal(req.Vip))
	}

	return nil
}

// checkExclusiveClusterOwnership 重新校验 cluster_tag/cloud_cluster_ids 是否仍归属当前业务。
func (svc *clbSvc) checkExclusiveClusterOwnership(kt *kit.Kit, bkBizID int64, clusterTag string,
	cloudClusterIDs []string) error {

	if len(clusterTag) != 0 {
		req := &core.ListReq{
			Fields: []string{"id"},
			Filter: tools.ExpressionAnd(
				tools.RuleEqual("bk_biz_id", bkBizID),
				tools.RuleEqual("cluster_type", enumor.STGWClusterType),
				tools.RuleEqual("cluster_tag", clusterTag),
			),
			Page: core.NewCountPage(),
		}
		result, err := svc.dataCli.Global.ListExclusiveCluster(kt, req)
		if err != nil {
			return err
		}
		if result.Count == 0 {
			return errf.Newf(errf.PermissionDenied, "cluster_tag does not belong to biz(id=%d)", bkBizID)
		}
	}

	if len(cloudClusterIDs) == 0 {
		return nil
	}

	remaining := cvt.StringSliceToMap(cloudClusterIDs)
	req := &core.ListReq{
		Fields: []string{"cloud_id"},
		Filter: tools.ExpressionAnd(
			tools.RuleEqual("bk_biz_id", bkBizID),
			tools.RuleEqual("cluster_type", enumor.TGWClusterType),
			tools.RuleIn("cloud_id", cloudClusterIDs),
		),
		Page: core.NewDefaultBasePage(),
	}
	result, err := svc.dataCli.Global.ListExclusiveCluster(kt, req)
	if err != nil {
		return err
	}
	for _, one := range result.Details {
		delete(remaining, one.CloudID)
	}
	if len(remaining) != 0 {
		return errf.Newf(errf.PermissionDenied, "cloud_cluster_ids does not belong to biz(id=%d)", bkBizID)
	}

	return nil
}

// checkBandwidthPackageEgress 重新校验共享带宽包出口与本次可能分配集群出口的一致性（R-008）。
func (svc *clbSvc) checkBandwidthPackageEgress(kt *kit.Kit, tcloudAdpt exclusiveClusterAdaptor,
	req *protolb.TCloudLoadBalancerCreateReq) error {

	bwPkgID := cvt.PtrToVal(req.BandwidthPackageID)
	bwResult, err := tcloudAdpt.ListBandwidthPackage(kt, &adtypes.TCloudListBwPkgOption{
		Region:      req.Region,
		Page:        &adcore.TCloudPage{Offset: 0, Limit: 1},
		PkgCloudIds: []string{bwPkgID},
	})
	if err != nil {
		return err
	}
	if len(bwResult.Packages) == 0 {
		return errf.Newf(errf.InvalidParameter, "bandwidth package(%s) not found", bwPkgID)
	}
	egress := bwResult.Packages[0].Egress

	allowedSet, err := svc.computeExclusiveClusterEgressSet(kt, req.BkBizID, cvt.PtrToVal(req.ClusterTag),
		req.CloudClusterIDs)
	if err != nil {
		return err
	}
	if _, ok := allowedSet[egress]; !ok {
		return errf.Newf(errf.InvalidParameter,
			"bandwidth package egress(%s) is not allowed by the selected exclusive cluster(s)", egress)
	}

	return nil
}

// computeExclusiveClusterEgressSet 计算「本次可能分配到的集群」允许的出口集合，规则同「共享带宽包出口一致性校验」。
func (svc *clbSvc) computeExclusiveClusterEgressSet(kt *kit.Kit, bkBizID int64, clusterTag string,
	cloudClusterIDs []string) (map[string]struct{}, error) {

	var e4Set, e7Set map[string]struct{}
	var err error

	if len(cloudClusterIDs) != 0 {
		e4Set, err = svc.queryExclusiveClusterEgress(kt, tools.ExpressionAnd(
			tools.RuleEqual("bk_biz_id", bkBizID),
			tools.RuleEqual("cluster_type", enumor.TGWClusterType),
			tools.RuleIn("cloud_id", cloudClusterIDs),
		))
		if err != nil {
			return nil, err
		}
	}

	if len(clusterTag) != 0 {
		e7Set, err = svc.queryExclusiveClusterEgress(kt, tools.ExpressionAnd(
			tools.RuleEqual("bk_biz_id", bkBizID),
			tools.RuleEqual("cluster_type", enumor.STGWClusterType),
			tools.RuleEqual("cluster_tag", clusterTag),
		))
		if err != nil {
			return nil, err
		}
	}

	switch {
	case len(cloudClusterIDs) != 0 && len(clusterTag) != 0:
		return intersectEgressSet(e4Set, e7Set), nil
	case len(cloudClusterIDs) != 0:
		return e4Set, nil
	default:
		return e7Set, nil
	}
}

// queryExclusiveClusterEgress 按过滤条件查询本地独占集群表，返回去重后的出口集合。
func (svc *clbSvc) queryExclusiveClusterEgress(kt *kit.Kit, expr *filter.Expression) (map[string]struct{}, error) {
	req := &core.ListReq{
		Fields: []string{"egress"},
		Filter: expr,
		Page:   core.NewDefaultBasePage(),
	}
	result, err := svc.dataCli.Global.ListExclusiveCluster(kt, req)
	if err != nil {
		return nil, err
	}

	set := make(map[string]struct{}, len(result.Details))
	for _, one := range result.Details {
		if len(one.Egress) != 0 {
			set[one.Egress] = struct{}{}
		}
	}
	return set, nil
}

// intersectEgressSet 返回两个出口集合的交集。
func intersectEgressSet(a, b map[string]struct{}) map[string]struct{} {
	result := make(map[string]struct{})
	for egress := range a {
		if _, ok := b[egress]; ok {
			result[egress] = struct{}{}
		}
	}
	return result
}

// checkVipIdle 复核指定 vip 在对应四层集群下是否仍然闲置。按 cluster-id + vip + idle 在云上过滤，
// 避免集群资源较多时因分页导致目标 vip 未返回而误判。
func (svc *clbSvc) checkVipIdle(kt *kit.Kit, tcloudAdpt exclusiveClusterAdaptor, region, clusterID,
	vip string) error {
	result, err := tcloudAdpt.DescribeClusterResources(kt, &typelb.TCloudDescribeClusterResourcesOption{
		Region:    region,
		ClusterID: clusterID,
		Vip:       vip,
		Idle:      cvt.ValToPtr(true),
	})
	if err != nil {
		logs.Errorf("describe cluster resources failed, cluster_id: %s, vip: %s, err: %v, rid: %s",
			clusterID, vip, err, kt.Rid)
		return err
	}

	for _, one := range result.Resources {
		if one.Vip == vip && one.Idle {
			return nil
		}
	}

	return errf.Newf(errf.InvalidParameter, "vip(%s) is not idle or not found in cluster(%s)", vip, clusterID)
}
