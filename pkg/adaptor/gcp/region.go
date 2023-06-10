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
	"hcm/pkg/adaptor/types/core"
	typesRegion "hcm/pkg/adaptor/types/region"
	"hcm/pkg/kit"
	"hcm/pkg/logs"
	"hcm/pkg/tools/converter"
)

// ListRegion list region.
// reference: https://cloud.google.com/compute/docs/reference/rest/v1/regions/list
func (g *Gcp) ListRegion(kt *kit.Kit, opt *core.GcpListOption) (*typesRegion.GcpRegionListResult, error) {
	if err := opt.Validate(); err != nil {
		return nil, err
	}

	client, err := g.clientSet.computeClient(kt)
	if err != nil {
		return nil, err
	}

	listCall := client.Regions.List(g.clientSet.credential.CloudProjectID).Context(kt.Ctx)

	if len(opt.CloudIDs) > 0 {
		listCall.Filter(generateResourceFilter("description", opt.CloudIDs))
	}

	if opt.Page != nil {
		listCall.MaxResults(opt.Page.PageSize).PageToken(opt.Page.PageToken)
	}

	resp, err := listCall.Do()
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
			SelfLink:    data.SelfLink,
		}
		details = append(details, tmpData)
	}

	return &typesRegion.GcpRegionListResult{
		NextPageToken: resp.NextPageToken,
		Count:         converter.ValToPtr(uint64(len(resp.Items))),
		Details:       details,
	}, nil
}
