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

	"hcm/pkg/adaptor/types/eip"
	"hcm/pkg/kit"
	"hcm/pkg/logs"
)

// ListEip ...
// reference: https://learn.microsoft.com/zh-cn/rest/api/virtualnetwork/public-ip-addresses/list-all?tabs=HTTP
func (a *Azure) ListEip(kt *kit.Kit, opt *eip.AzureEipListOption) (*eip.AzureEipListResult, error) {
	client, err := a.clientSet.publicIPAddressesClient()
	if err != nil {
		return nil, err
	}

	eips := make([]*eip.AzureEip, 0)
	pager := client.NewListAllPager(nil)
	for pager.More() {
		nextResult, err := pager.NextPage(kt.Ctx)
		if err != nil {
			logs.Errorf("list azure eip failed, err: %v, rid: %s", err, kt.Rid)
			return nil, fmt.Errorf("list azure eip failed, err: %v", err)
		}
		for _, v := range nextResult.Value {
			state := string(*v.Properties.ProvisioningState)
			sku := string(*v.SKU.Name)
			eIp := &eip.AzureEip{
				CloudID:  *v.ID,
				Name:     v.Name,
				Region:   *v.Location,
				Status:   &state,
				PublicIp: v.Properties.IPAddress,
				SKU:      &sku,
			}
			if v.Properties.IPConfiguration != nil {
				eIp.IpConfigurationID = v.Properties.IPConfiguration.ID
			}

			eips = append(eips, eIp)
		}
	}

	return &eip.AzureEipListResult{Details: eips}, nil
}
