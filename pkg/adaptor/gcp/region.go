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

package gcp

import (
	typesRegion "hcm/pkg/adaptor/types/region"
	"hcm/pkg/kit"
	"hcm/pkg/logs"
	"hcm/pkg/tools/converter"
)

// ListRegion list region.
// reference: https://cloud.google.com/compute/docs/reference/rest/v1/regions/list
func (g *Gcp) ListRegion(kt *kit.Kit) (*typesRegion.GcpRegionListResult, error) {
	client, err := g.clientSet.computeClient(kt)
	if err != nil {
		return nil, err
	}

	resp, err := client.Regions.List(g.clientSet.credential.CloudProjectID).Context(kt.Ctx).Do()
	if err != nil {
		logs.Errorf("list gcp region failed, err: %v, rid: %s", err, kt.Rid)
		return nil, err
	}

	details := make([]typesRegion.GcpRegion, 0, len(resp.Items))
	for _, data := range resp.Items {
		tmpData := typesRegion.GcpRegion{
			RegionID:    data.Description,
			RegionName:  data.Name,
			RegionState: data.Status, // UP、DOWN
		}
		details = append(details, tmpData)
	}

	return &typesRegion.GcpRegionListResult{Count: converter.ValToPtr(uint64(len(resp.Items))),
		Details: details}, nil
}
