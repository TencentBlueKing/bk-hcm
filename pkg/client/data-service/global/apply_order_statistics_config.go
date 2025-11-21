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
 *
 * to the current version of the project delivered to anyone in the future.
 */

package global

import (
	"hcm/pkg/api/core"
	"hcm/pkg/client/common"
	tableapplystat "hcm/pkg/dal/table/cvm-apply-order-statistics-config"
	"hcm/pkg/kit"
	"hcm/pkg/rest"
)

// ApplyOrderStatisticsConfigClient is data service client for apply order statistics config.
type ApplyOrderStatisticsConfigClient struct {
	client rest.ClientInterface
}

// NewApplyOrderStatisticsConfigClient creates a new client.
func NewApplyOrderStatisticsConfigClient(client rest.ClientInterface) *ApplyOrderStatisticsConfigClient {
	return &ApplyOrderStatisticsConfigClient{client: client}
}

// List apply order statistics configs.
func (c *ApplyOrderStatisticsConfigClient) List(kt *kit.Kit, req *core.ListReq,
) (*core.ListResultT[tableapplystat.CvmApplyOrderStatisticsConfigTable], error) {
	return common.Request[core.ListReq, core.ListResultT[tableapplystat.CvmApplyOrderStatisticsConfigTable]](
		c.client, rest.POST, kt, req, "/apply_order_statistics_configs/list")
}
