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
	"io"
	"net/http"
	"strconv"
	"sync"
	"testing"

	"hcm/pkg/dal/dao/tools"

	"github.com/stretchr/testify/require"
)

// batchQueryRecorder 记录数据服务收到的每一次列表请求，并按请求里的 cloud_id 条件回放命中记录。
type batchQueryRecorder struct {
	mu sync.Mutex
	// batchSizes 每个批次请求的 cloud_id 数量，翻页的后续页不重复记录
	batchSizes []int
	// pageStarts 所有请求的翻页起点
	pageStarts []int
	// bodies 所有请求的原始请求体
	bodies []string
}

// handler 返回一个只对第一页回放全部 cloud_id 命中记录的数据服务 handler。
func (r *batchQueryRecorder) handler(t *testing.T) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
		var body struct {
			Filter struct {
				Rules []struct {
					Field string `json:"field"`
					Op    string `json:"op"`
					Value any    `json:"value"`
				} `json:"rules"`
			} `json:"filter"`
			Page struct {
				Start int `json:"start"`
			} `json:"page"`
		}
		raw, err := io.ReadAll(req.Body)
		require.NoError(t, err)
		require.NoError(t, json.Unmarshal(raw, &body))

		var ids []any
		for _, rule := range body.Filter.Rules {
			if rule.Field == "cloud_id" && rule.Op == "in" {
				ids, _ = rule.Value.([]any)
			}
		}

		r.mu.Lock()
		r.pageStarts = append(r.pageStarts, body.Page.Start)
		r.bodies = append(r.bodies, string(raw))
		if body.Page.Start == 0 {
			r.batchSizes = append(r.batchSizes, len(ids))
		}
		r.mu.Unlock()

		details := make([]map[string]any, 0, len(ids))
		if body.Page.Start == 0 {
			for _, id := range ids {
				details = append(details, map[string]any{"cloud_id": id, "egress": "egress-" + id.(string)})
			}
		}
		writeOKResp(t, w, map[string]any{"details": details})
	})
}

func newCloudIDs(n int) []string {
	ids := make([]string, n)
	for i := range ids {
		ids[i] = "tgw-" + strconv.Itoa(i)
	}
	return ids
}

// TestListExclusiveClustersByCloudIDs_SplitIntoBatches 云上ID数量超过 in 条件上限时按上限分批查询，
// 每一批满页后继续翻页，最终返回全部记录。
func TestListExclusiveClustersByCloudIDs_SplitIntoBatches(t *testing.T) {
	rec := &batchQueryRecorder{}
	cli := newTestDataServiceClient(t, rec.handler(t))

	ids := newCloudIDs(1200)
	clusters, err := listExclusiveClustersByCloudIDs(testKit(), cli, ids, []string{"cloud_id"})
	require.NoError(t, err)
	require.Len(t, clusters, 1200)
	require.Equal(t, []int{500, 500, 200}, rec.batchSizes)
	// 前两批各满一页（500 条），因此各自多请求了一次下一页；最后一批不足一页不再翻页
	require.Equal(t, []int{0, 500, 0, 500, 0}, rec.pageStarts)
}

// TestListExclusiveClustersByCloudIDs_BaseRulesApplyToEveryBatch 调用方传入的固定过滤条件在每一个批次、每一页
// 的请求里都生效。
func TestListExclusiveClustersByCloudIDs_BaseRulesApplyToEveryBatch(t *testing.T) {
	rec := &batchQueryRecorder{}
	cli := newTestDataServiceClient(t, rec.handler(t))

	_, err := listExclusiveClustersByCloudIDs(testKit(), cli, newCloudIDs(600), []string{"cloud_id"},
		tools.RuleEqual("bk_biz_id", 213), tools.RuleEqual("account_id", "0000001"))
	require.NoError(t, err)
	require.NotEmpty(t, rec.bodies)
	for _, body := range rec.bodies {
		requireBodyHasFields(t, body, "bk_biz_id", "account_id", "cloud_id")
	}
}

// TestListExclusiveClustersByCloudIDs_EmptyIDs 云上ID为空时不发起任何查询。
func TestListExclusiveClustersByCloudIDs_EmptyIDs(t *testing.T) {
	rec := &batchQueryRecorder{}
	cli := newTestDataServiceClient(t, rec.handler(t))

	clusters, err := listExclusiveClustersByCloudIDs(testKit(), cli, nil, []string{"cloud_id"})
	require.NoError(t, err)
	require.Empty(t, clusters)
	require.Empty(t, rec.bodies)
}

// TestListExclusiveClustersByCloudIDs_DuplicateIDs 重复的云上ID只查询一次。
func TestListExclusiveClustersByCloudIDs_DuplicateIDs(t *testing.T) {
	rec := &batchQueryRecorder{}
	cli := newTestDataServiceClient(t, rec.handler(t))

	clusters, err := listExclusiveClustersByCloudIDs(testKit(), cli, []string{"tgw-1", "tgw-1", "tgw-2"},
		[]string{"cloud_id"})
	require.NoError(t, err)
	require.Len(t, clusters, 2)
	require.Equal(t, []int{2}, rec.batchSizes)
}

// TestCheckExclusiveClusterOwnership_ManyClusterIDs 四层集群ID超过单批上限时，归属校验分批查询后仍能全部命中。
func TestCheckExclusiveClusterOwnership_ManyClusterIDs(t *testing.T) {
	rec := &batchQueryRecorder{}
	cli := newTestDataServiceClient(t, rec.handler(t))

	err := CheckExclusiveClusterOwnership(testKit(), cli, newOwnershipReq("", newCloudIDs(1200)))
	require.NoError(t, err)
	require.Equal(t, []int{500, 500, 200}, rec.batchSizes)
}

// TestComputeExclusiveClusterEgressSet_ManyClusterIDs 四层集群ID超过单批上限时，出口集合覆盖所有批次的集群。
func TestComputeExclusiveClusterEgressSet_ManyClusterIDs(t *testing.T) {
	rec := &batchQueryRecorder{}
	cli := newTestDataServiceClient(t, rec.handler(t))

	egresses, err := ComputeExclusiveClusterEgressSet(testKit(), cli, newOwnershipReq("", newCloudIDs(1200)))
	require.NoError(t, err)
	require.Len(t, egresses, 1200)
}
