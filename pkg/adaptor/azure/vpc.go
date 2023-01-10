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

	"hcm/pkg/adaptor/types"
	"hcm/pkg/adaptor/types/core"
	"hcm/pkg/api/core/cloud"
	"hcm/pkg/criteria/enumor"
	"hcm/pkg/kit"
	"hcm/pkg/logs"
	"hcm/pkg/tools/cidr"
	"hcm/pkg/tools/converter"

	"github.com/Azure/azure-sdk-for-go/sdk/azcore/runtime"
	"github.com/Azure/azure-sdk-for-go/sdk/resourcemanager/network/armnetwork"
)

// UpdateVpc update vpc.
// TODO right now only memo is supported to update, add other update operations later.
func (a *Azure) UpdateVpc(kt *kit.Kit, opt *types.AzureVpcUpdateOption) error {
	return nil
}

// DeleteVpc delete vpc.
// reference: https://docs.microsoft.com/en-us/rest/api/virtualnetwork/virtual-networks/delete
func (a *Azure) DeleteVpc(kt *kit.Kit, opt *core.AzureDeleteOption) error {
	if err := opt.Validate(); err != nil {
		return err
	}

	vpcClient, err := a.clientSet.vpcClient()
	if err != nil {
		return fmt.Errorf("new vpc client failed, err: %v", err)
	}

	poller, err := vpcClient.BeginDelete(kt.Ctx, opt.ResourceGroupName, opt.ResourceID, nil)
	if err != nil {
		logs.Errorf("delete azure vpc failed, err: %v, rid: %s", err, kt.Rid)
		return err
	}

	_, err = poller.PollUntilDone(kt.Ctx, new(runtime.PollUntilDoneOptions))
	if err != nil {
		return err
	}

	return nil
}

// ListVpc list vpc.
// reference: https://learn.microsoft.com/en-us/rest/api/virtualnetwork/virtual-networks/list
func (a *Azure) ListVpc(kt *kit.Kit, opt *core.AzureListOption) (*types.AzureVpcListResult, error) {
	if err := opt.Validate(); err != nil {
		return nil, err
	}

	vpcClient, err := a.clientSet.vpcClient()
	if err != nil {
		return nil, fmt.Errorf("new vpc client failed, err: %v", err)
	}

	req := new(armnetwork.VirtualNetworksClientListOptions)

	pager := vpcClient.NewListPager(opt.ResourceGroupName, req)
	if err != nil {
		logs.Errorf("list azure vpc failed, err: %v, rid: %s", err, kt.Rid)
		return nil, fmt.Errorf("list azure vpc failed, err: %v", err)
	}

	details := make([]types.AzureVpc, 0)
	for pager.More() {
		page, err := pager.NextPage(kt.Ctx)
		if err != nil {
			return nil, fmt.Errorf("list azure vpc but get next page failed, err: %v", err)
		}

		for _, vpc := range page.Value {
			details = append(details, converter.PtrToVal(convertVpc(vpc, opt.ResourceGroupName)))
		}
	}

	return &types.AzureVpcListResult{Details: details}, nil
}

func convertVpc(data *armnetwork.VirtualNetwork, resourceGroup string) *types.AzureVpc {
	if data == nil {
		return nil
	}

	v := &types.AzureVpc{
		Spec: &cloud.VpcSpec{
			CloudID:  converter.PtrToVal(data.ID),
			Name:     converter.PtrToVal(data.Name),
			Category: enumor.BizVpcCategory,
		},
		Extension: &cloud.AzureVpcExtension{
			ResourceGroup: resourceGroup,
			Region:        converter.PtrToVal(data.Location),
			Cidr:          nil,
		},
	}

	if data.Properties == nil {
		return v
	}

	if data.Properties.DhcpOptions != nil {
		v.Extension.DNSServers = converter.PtrToSlice(data.Properties.DhcpOptions.DNSServers)
	}

	if data.Properties.AddressSpace != nil {
		for _, prefix := range data.Properties.AddressSpace.AddressPrefixes {
			if prefix == nil {
				continue
			}

			addressType, err := cidr.CidrIPAddressType(*prefix)
			if err != nil {
				logs.Errorf("get cidr ip address type failed, cidr: %v, err: %v", *prefix, err)
			}

			v.Extension.Cidr = append(v.Extension.Cidr, cloud.AzureCidr{
				Type: addressType,
				Cidr: *prefix,
			})
		}
	}

	return v
}
