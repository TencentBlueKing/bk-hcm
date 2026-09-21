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
	"bytes"
	"encoding/json"
	"fmt"

	"hcm/pkg/api/core"
	dataproto "hcm/pkg/api/data-service/cloud"
	"hcm/pkg/criteria/enumor"
	"hcm/pkg/dal/dao/tools"
	"hcm/pkg/kit"
	"hcm/pkg/logs"

	"github.com/tidwall/gjson"
)

// exclusiveClusterContentInfo 申请单详情 content 中富化的单个独占集群信息，字段含义详见
// openspec/changes/load-balancer-exclusive-cluster-purchase 的 load-balancer-exclusive-cluster-detail 规格。
type exclusiveClusterContentInfo struct {
	CloudClusterID string `json:"cloud_cluster_id"`
	ClusterID      string `json:"cluster_id"`
	ClusterName    string `json:"cluster_name"`
	ClusterTag     string `json:"cluster_tag"`
	ClusterType    string `json:"cluster_type"`
}

// exclusiveClusterLister 富化 content 所需的最小数据面能力，从 `*global.Client` 中按需截取，便于单元测试注入
// fake 实现。
type exclusiveClusterLister interface {
	ListExclusiveCluster(kt *kit.Kit, req *core.ListReq) (*dataproto.ExclusiveClusterListResult, error)
}

// enrichExclusiveClusterContent 申请单详情 content 独占集群信息富化（读时计算，不落库）：仅
// type=create_load_balancer 时生效，非独占型 clusters 为空数组，独占型按 cluster_tag/cloud_cluster_ids
// 拼出 clusters 数组注入 content。
func (a *applicationSvc) enrichExclusiveClusterContent(kt *kit.Kit, appType enumor.ApplicationType,
	content string) string {

	return enrichExclusiveClusterContent(kt, a.client.DataService().Global, appType, content)
}

// enrichExclusiveClusterContent 独立于 applicationSvc 的富化实现，lister 由调用方注入，便于单元测试。
func enrichExclusiveClusterContent(kt *kit.Kit, lister exclusiveClusterLister, appType enumor.ApplicationType,
	content string) string {

	if appType != enumor.CreateLoadBalancer {
		return content
	}

	parsed := gjson.Parse(content)
	clusters := make([]exclusiveClusterContentInfo, 0)
	if parsed.Get("exclusive").Int() == 1 {
		bkBizID := parsed.Get("bk_biz_id").Int()
		clusterTag := parsed.Get("cluster_tag").String()

		var cloudClusterIDs []string
		for _, one := range parsed.Get("cloud_cluster_ids").Array() {
			if id := one.String(); id != "" {
				cloudClusterIDs = append(cloudClusterIDs, id)
			}
		}

		var err error
		clusters, err = buildExclusiveClusterContentInfos(kt, lister, bkBizID, clusterTag, cloudClusterIDs)
		if err != nil {
			logs.Errorf("build exclusive cluster content info failed, err: %v, bizID: %d, rid: %s",
				err, bkBizID, kt.Rid)
			// 富化失败不影响原有 content 的正常展示，直接跳过 clusters 富化
			return content
		}
	}

	clustersRaw, err := json.Marshal(clusters)
	if err != nil {
		logs.Errorf("marshal exclusive clusters failed, err: %v, rid: %s", err, kt.Rid)
		return content
	}

	return injectJSONField(content, "clusters", string(clustersRaw))
}

// localTgwCluster 四层独占集群本地表反查结果，用于补齐申请单详情 clusters 中 TGW 元素的展示字段。
type localTgwCluster struct {
	ID         string
	Name       string
	ClusterTag string
}

// buildExclusiveClusterContentInfos 按 cluster_tag/cloud_cluster_ids 拼出 clusters 数组：cluster_tag 对应七层
// 独占集群，提单时通常没有具体落地集群 ID，只填 cluster_tag/cluster_type；cloud_cluster_ids 里每个 ID 对应一个四层
// 独占集群，按 ID 反查本地表补齐 cluster_id/cluster_name/cluster_tag，本地表未同步或已删除时对应字段留空，
// cloud_cluster_id 仍返回原值。
func buildExclusiveClusterContentInfos(kt *kit.Kit, lister exclusiveClusterLister, bkBizID int64, clusterTag string,
	cloudClusterIDs []string) ([]exclusiveClusterContentInfo, error) {

	clusters := make([]exclusiveClusterContentInfo, 0, len(cloudClusterIDs)+1)

	if len(clusterTag) != 0 {
		clusters = append(clusters, exclusiveClusterContentInfo{
			ClusterTag:  clusterTag,
			ClusterType: string(enumor.STGWClusterType),
		})
	}

	if len(cloudClusterIDs) == 0 {
		return clusters, nil
	}

	req := &core.ListReq{
		Fields: []string{"id", "cloud_id", "name", "cluster_tag"},
		Filter: tools.ExpressionAnd(
			tools.RuleEqual("bk_biz_id", bkBizID),
			tools.RuleEqual("cluster_type", enumor.TGWClusterType),
			tools.RuleIn("cloud_id", cloudClusterIDs),
		),
		Page: core.NewDefaultBasePage(),
	}
	result, err := lister.ListExclusiveCluster(kt, req)
	if err != nil {
		return nil, err
	}

	localByCloudID := make(map[string]localTgwCluster, len(result.Details))
	for _, one := range result.Details {
		localByCloudID[one.CloudID] = localTgwCluster{ID: one.ID, Name: one.Name, ClusterTag: one.ClusterTag}
	}

	for _, cloudID := range cloudClusterIDs {
		info := exclusiveClusterContentInfo{CloudClusterID: cloudID, ClusterType: string(enumor.TGWClusterType)}
		if local, ok := localByCloudID[cloudID]; ok {
			info.ClusterID = local.ID
			info.ClusterName = local.Name
			info.ClusterTag = local.ClusterTag
		}
		clusters = append(clusters, info)
	}

	return clusters, nil
}

// injectJSONField 将一个已序列化的 JSON 值以指定 key 注入到 content 顶层对象中，写法与 RemoveSenseField 保持
// 一致，避免引入额外的 JSON 写库依赖。
func injectJSONField(content, key, rawValue string) string {
	buffer := bytes.Buffer{}

	m := gjson.Parse(content).Map()
	for k, v := range m {
		if k == key {
			continue
		}
		buffer.WriteString(fmt.Sprintf(`"%s":%s,`, k, v.Raw))
	}
	buffer.WriteString(fmt.Sprintf(`"%s":%s,`, key, rawValue))

	ext := buffer.String()
	return fmt.Sprintf("{%s}", ext[:len(ext)-1])
}
