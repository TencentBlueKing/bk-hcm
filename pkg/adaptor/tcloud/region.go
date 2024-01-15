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

package tcloud

import (
	"fmt"

	typesRegion "hcm/pkg/adaptor/types/region"
	"hcm/pkg/kit"
	"hcm/pkg/logs"
	"hcm/pkg/tools/converter"
)

// ListRegion list region.
// reference: https://cloud.tencent.com/document/product/1278/55255
func (t *TCloudImpl) ListRegion(kt *kit.Kit) (*typesRegion.TCloudRegionListResult, error) {
	CvmClient, err := t.clientSet.CvmClient("")
	if err != nil {
		return nil, fmt.Errorf("new cvm client failed, err: %v", err)
	}

	resp, err := CvmClient.DescribeRegionsWithContext(kt.Ctx, nil)
	if err != nil {
		logs.Errorf("list tcloud region failed, err: %v, rid: %s", err, kt.Rid)
		return nil, fmt.Errorf("list tcloud region failed, err: %v", err)
	}

	details := make([]typesRegion.TCloudRegion, 0, len(resp.Response.RegionSet))
	for _, data := range resp.Response.RegionSet {
		tmpData := typesRegion.TCloudRegion{
			RegionID:    converter.PtrToVal(data.Region),
			RegionName:  converter.PtrToVal(data.RegionName),
			RegionState: converter.PtrToVal(data.RegionState),
		}
		details = append(details, tmpData)
	}

	return &typesRegion.TCloudRegionListResult{Count: resp.Response.TotalCount, Details: details}, nil
}
