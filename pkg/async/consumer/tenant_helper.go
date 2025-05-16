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

package consumer

import (
	"fmt"

	actcli "hcm/cmd/task-server/logics/action/cli"
	"hcm/pkg/api/core"
	"hcm/pkg/dal/dao/tools"
	"hcm/pkg/logs"
)

// listTenantIDs 获取所有租户ID
func listTenantIDs() ([]string, error) {
	kt := NewKit()
	tenantIDs := make([]string, 0)
	page := core.NewDefaultBasePage()
	for {
		result, err := actcli.GetDataService().Global.Tenant.List(kt, &core.ListReq{
			Page:   page,
			Fields: []string{"tenant_id"},
			Filter: tools.EqualExpression("status", "enable"),
		})
		if err != nil {
			logs.Errorf("list tenant failed, err: %v, rid: %s", err, kt.Rid)
			return nil, fmt.Errorf("list tenant failed, err: %v, rid: %s", err, kt.Rid)
		}

		for _, t := range result.Details {
			tenantIDs = append(tenantIDs, t.TenantID)
		}

		// 如果当前页数据不足一页，说明后面没有更多数据了
		if uint(len(result.Details)) < page.Limit {
			break
		}
		page.Start += uint32(page.Limit)
	}
	return tenantIDs, nil
}

// distributeTenantTasks 获取租户列表并将任务分发到协程池
func distributeTenantTasks(pool *workerPool) error {
	tenantIDs, err := listTenantIDs()
	if err != nil {
		logs.Errorf("distributeTenantTasks failed to list tenants, err: %v", err)
		return err
	}

	for _, tenantID := range tenantIDs {
		pool.submit(tenantID)
	}
	return nil
}
