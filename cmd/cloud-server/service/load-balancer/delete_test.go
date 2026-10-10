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
	"testing"

	actionlb "hcm/cmd/task-server/logics/action/load-balancer"
	"hcm/pkg/criteria/enumor"
	"hcm/pkg/dal/dao/types"

	"github.com/stretchr/testify/assert"
)

func Test_buildLBDeletionTasks(t *testing.T) {
	type testCase struct {
		name      string
		infoMap   map[string]types.CloudResourceBasicInfo
		wantCount int
	}

	tests := []testCase{
		{
			name:      "single group with 20 IDs",
			infoMap:   buildInfoMap("acc-1", enumor.TCloud, "ap-guangzhou", 20, 0),
			wantCount: 1,
		},
		{
			name:      "single group with 21 IDs",
			infoMap:   buildInfoMap("acc-1", enumor.TCloud, "ap-guangzhou", 21, 0),
			wantCount: 2,
		},
		{
			name:      "single group with 40 IDs",
			infoMap:   buildInfoMap("acc-1", enumor.TCloud, "ap-guangzhou", 40, 0),
			wantCount: 2,
		},
		{
			name:      "single group with 1 ID",
			infoMap:   buildInfoMap("acc-1", enumor.TCloud, "ap-guangzhou", 1, 0),
			wantCount: 1,
		},
		{
			name: "mixed groups A:25 + B:10",
			infoMap: mergeMaps(
				buildInfoMap("acc-a", enumor.TCloud, "ap-guangzhou", 25, 0),
				buildInfoMap("acc-b", enumor.TCloud, "ap-shanghai", 10, 100),
			),
			wantCount: 3,
		},
		{
			name: "all single groups 3+5+7",
			infoMap: mergeMaps(
				buildInfoMap("acc-a", enumor.TCloud, "ap-guangzhou", 3, 0),
				buildInfoMap("acc-b", enumor.TCloud, "ap-shanghai", 5, 10),
				buildInfoMap("acc-c", enumor.TCloud, "ap-beijing", 7, 20),
			),
			wantCount: 3,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tasks := buildLBDeletionTasks(tt.infoMap)

			// 验证 task 总数正确
			assert.Equal(t, tt.wantCount, len(tasks))

			// 收集所有 task 信息用于后续验证
			actionIDs := make(map[string]struct{}, len(tasks))
			totalIDs := make([]string, 0, len(tt.infoMap))

			// 解析每个 task 的 Params 并校验
			for _, task := range tasks {
				// 验证 ActionID 唯一
				_, exists := actionIDs[string(task.ActionID)]
				assert.False(t, exists, "ActionID %s should be unique", task.ActionID)
				actionIDs[string(task.ActionID)] = struct{}{}

				// 验证 ActionName
				assert.Equal(t, string(enumor.ActionDeleteLoadBalancer), string(task.ActionName))

				// 验证 Retry 不为空
				assert.NotNil(t, task.Retry)

				// Params 是 DeleteLoadBalancerOption 值（非指针）
				opt, ok := task.Params.(actionlb.DeleteLoadBalancerOption)
				assert.True(t, ok, "Params should be DeleteLoadBalancerOption")

				// 验证 IDs 长度 ≤ 20
				assert.LessOrEqual(t, len(opt.IDs), 20,
					"batch IDs length %d should <= 20", len(opt.IDs))

				// 验证 IDs 非空
				assert.Greater(t, len(opt.IDs), 0,
					"batch IDs should not be empty")

				// 验证 AccountID/Region/Vendor 非空
				assert.NotEmpty(t, opt.AccountID)
				assert.NotEmpty(t, opt.Region)
				assert.NotEmpty(t, opt.Vendor)

				totalIDs = append(totalIDs, opt.IDs...)
			}

			// 验证所有 task 的 IDs 合并后等于输入全集（无丢失无重复）
			expectedIDs := make([]string, 0, len(tt.infoMap))
			for id := range tt.infoMap {
				expectedIDs = append(expectedIDs, id)
			}

			assert.ElementsMatch(t, expectedIDs, totalIDs,
				"merged IDs should match input IDs exactly")
		})
	}
}

// buildInfoMap 构建指定 account+vendor+region 的测试数据，idOffset 用于避免不同组的 ID 冲突
func buildInfoMap(accountID string, vendor enumor.Vendor, region string, count int,
	idOffset int) map[string]types.CloudResourceBasicInfo {

	infoMap := make(map[string]types.CloudResourceBasicInfo, count)
	for i := 0; i < count; i++ {
		id := fmt.Sprintf("lb-%d", idOffset+i)
		infoMap[id] = types.CloudResourceBasicInfo{
			ID:        id,
			Vendor:    vendor,
			AccountID: accountID,
			Region:    region,
		}
	}
	return infoMap
}

// mergeMaps 合并多个 map
func mergeMaps(maps ...map[string]types.CloudResourceBasicInfo) map[string]types.CloudResourceBasicInfo {
	result := make(map[string]types.CloudResourceBasicInfo)
	for _, m := range maps {
		for k, v := range m {
			result[k] = v
		}
	}
	return result
}
