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

package azure

import (
	"fmt"

	region "hcm/pkg/adaptor/types/region"
	"hcm/pkg/kit"

	"github.com/Azure/azure-sdk-for-go/sdk/resourcemanager/resources/armsubscriptions"
)

// ListRegion list region.
// reference: https://learn.microsoft.com/en-us/rest/api/resources/subscriptions/list-locations?tabs=HTTP#examples
func (az *Azure) ListRegion(kit *kit.Kit) ([]*region.AzureRegion, error) {

	client, err := az.clientSet.regionClient()
	if err != nil {
		return nil, fmt.Errorf("new region failed, err: %v", err)
	}

	regions := make([]*armsubscriptions.Location, 0)
	pager := client.NewListLocationsPager(az.clientSet.credential.CloudSubscriptionID, nil)
	for pager.More() {
		nextResult, err := pager.NextPage(kit.Ctx)
		if err != nil {
			return nil, fmt.Errorf("failed to advance page: %v", err)
		}
		regions = append(regions, nextResult.Value...)
	}

	typeRegions := make([]*region.AzureRegion, 0)
	for _, v := range regions {
		tmp := &region.AzureRegion{
			ID:                  SPtrToLowerSPtr(v.ID),
			Name:                SPtrToLowerSPtr(v.Name),
			Type:                string(*v.Type),
			DisplayName:         v.DisplayName,
			RegionalDisplayName: v.RegionalDisplayName,
		}
		typeRegions = append(typeRegions, tmp)
	}

	return typeRegions, nil
}
