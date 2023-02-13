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
	routetable "hcm/pkg/adaptor/types/route-table"
	"hcm/pkg/tools/converter"

	"github.com/Azure/azure-sdk-for-go/sdk/resourcemanager/network/armnetwork/v2"
)

func convertRoute(data *armnetwork.Route, cloudRouteTableID string) *routetable.AzureRoute {
	if data == nil {
		return nil
	}

	r := &routetable.AzureRoute{
		CloudID:           converter.PtrToVal(data.ID),
		CloudRouteTableID: cloudRouteTableID,
		Name:              converter.PtrToVal(data.Name),
		AddressPrefix:     converter.PtrToVal(data.Properties.AddressPrefix),
		NextHopType:       string(converter.PtrToVal(data.Properties.NextHopType)),
		NextHopIPAddress:  data.Properties.NextHopIPAddress,
		ProvisioningState: string(converter.PtrToVal(data.Properties.ProvisioningState)),
	}

	if data.Properties == nil {
		return r
	}

	return r
}
