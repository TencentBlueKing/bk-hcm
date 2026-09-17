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

package lblogic

import (
	"encoding/json"
	"fmt"

	cslb "hcm/pkg/api/cloud-server/load-balancer"
	"hcm/pkg/api/core"
	corelb "hcm/pkg/api/core/cloud/load-balancer"
	dataservice "hcm/pkg/client/data-service"
	"hcm/pkg/criteria/enumor"
	"hcm/pkg/kit"
	"hcm/pkg/logs"
	"hcm/pkg/runtime/filter"
)

// exclusiveClusterTagGroupKey 独占集群标签聚合分组的键，由集群标签+集群类型两个维度组成。
type exclusiveClusterTagGroupKey struct {
	tag         string
	clusterType enumor.ClusterType
}

// AggregateExclusiveClusterTags 按 (cluster_tag, cluster_type) 对独占集群做分组聚合，供业务视角标签聚合查询接口
// （N-04）使用。bizFilterExpr 由调用方通过 handler.ListBizAuthRes 生成（已包含 bk_biz_id 归属过滤及 account_id/
// region/isp/cluster_type 等业务过滤条件），本函数只负责取数后的内存聚合与 zone 过滤，不重复拼装归属过滤条件。
func AggregateExclusiveClusterTags(kt *kit.Kit, cli *dataservice.Client, bizFilterExpr *filter.Expression,
	req *cslb.ListExclusiveClusterTagsReq) (*cslb.ListExclusiveClusterTagsResult, error) {

	listReq := &core.ListReq{
		Filter: bizFilterExpr,
		Page:   core.NewDefaultBasePage(),
	}
	result, err := cli.Global.ListExclusiveCluster(kt, listReq)
	if err != nil {
		logs.Errorf("list exclusive cluster for tags aggregation failed, err: %v, rid: %s", err, kt.Rid)
		return nil, err
	}

	groupOrder := make([]exclusiveClusterTagGroupKey, 0)
	groups := make(map[exclusiveClusterTagGroupKey][]cslb.ExclusiveClusterTagItem)

	for _, one := range result.Details {
		// R-002: 集群标签为空的记录不参与聚合。
		if len(one.ClusterTag) == 0 {
			continue
		}

		if len(req.Zone) != 0 {
			matched, err := exclusiveClusterMatchesZone(one.Extension, req.Zone)
			if err != nil {
				logs.Errorf("unmarshal exclusive cluster(id=%s) extension failed, err: %v, rid: %s",
					one.ID, err, kt.Rid)
				return nil, fmt.Errorf("unmarshal exclusive cluster(id=%s) extension failed, err: %v", one.ID, err)
			}
			// R-005: zone 不命中该集群的 master_zone 时跳过该集群。
			if !matched {
				continue
			}
		}

		key := exclusiveClusterTagGroupKey{tag: one.ClusterTag, clusterType: one.ClusterType}
		if _, exist := groups[key]; !exist {
			groupOrder = append(groupOrder, key)
		}
		groups[key] = append(groups[key], cslb.ExclusiveClusterTagItem{
			CloudClusterID: one.CloudID,
			ClusterID:      one.ID,
			ClusterName:    one.Name,
			Egress:         one.Egress,
			Isp:            one.Isp,
			Zone:           one.Zone,
		})
	}

	details := make([]cslb.ExclusiveClusterTagGroup, 0, len(groupOrder))
	for _, key := range groupOrder {
		details = append(details, cslb.ExclusiveClusterTagGroup{
			ClusterTag:  key.tag,
			ClusterType: key.clusterType,
			Clusters:    groups[key],
		})
	}

	return &cslb.ListExclusiveClusterTagsResult{Details: details}, nil
}

// exclusiveClusterMatchesZone 判断独占集群的 extension.clusters_zone.master_zone 是否包含指定可用区。
func exclusiveClusterMatchesZone(rawExt json.RawMessage, zone string) (bool, error) {
	ext := new(corelb.TCloudExclusiveClusterExtension)
	if err := json.Unmarshal(rawExt, ext); err != nil {
		return false, err
	}

	for _, one := range ext.ClustersZone.MasterZone {
		if one == zone {
			return true, nil
		}
	}

	return false, nil
}
