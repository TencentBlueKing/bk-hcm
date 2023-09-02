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
	"strings"

	"github.com/Azure/azure-sdk-for-go/sdk/resourcemanager/network/armnetwork/v2"

	"hcm/pkg/adaptor/types/core"
	"hcm/pkg/adaptor/types/eip"
	"hcm/pkg/criteria/enumor"
	"hcm/pkg/criteria/errf"
	"hcm/pkg/kit"
	"hcm/pkg/logs"
	"hcm/pkg/tools/converter"
)

// ListEipByID ...
// reference: https://learn.microsoft.com/zh-cn/rest/api/virtualnetwork/public-ip-addresses/list-all?tabs=HTTP
func (az *Azure) ListEipByID(kt *kit.Kit, opt *core.AzureListByIDOption) (*eip.AzureEipListResult, error) {
	client, err := az.clientSet.publicIPAddressesClient()
	if err != nil {
		return nil, err
	}

	idMap := converter.StringSliceToMap(opt.CloudIDs)

	eips := make([]*eip.AzureEip, 0, len(idMap))
	pager := client.NewListPager(opt.ResourceGroupName, nil)
	for pager.More() {
		nextResult, err := pager.NextPage(kt.Ctx)
		if err != nil {
			logs.Errorf("list azure eip failed, err: %v, rid: %s", err, kt.Rid)
			return nil, fmt.Errorf("list azure eip failed, err: %v", err)
		}

		for _, one := range nextResult.Value {
			eipTmp := convertEip(one, opt.ResourceGroupName)

			if len(opt.CloudIDs) > 0 {
				id := SPtrToLowerSPtr(one.ID)
				if _, exist := idMap[*id]; exist {
					eips = append(eips, eipTmp)
					delete(idMap, *id)
					if len(idMap) == 0 {
						return &eip.AzureEipListResult{Details: eips}, nil
					}
				}
			} else {
				eips = append(eips, eipTmp)
			}
		}
	}

	return &eip.AzureEipListResult{Details: eips}, nil
}

// CountEip count eip.
// reference: https://learn.microsoft.com/zh-cn/rest/api/virtualnetwork/public-ip-addresses/list-all?tabs=HTTP
func (az *Azure) CountEip(kt *kit.Kit) (int32, error) {

	client, err := az.clientSet.publicIPAddressesClient()
	if err != nil {
		return 0, fmt.Errorf("new eip client failed, err: %v", err)
	}

	var count int32
	pager := client.NewListAllPager(nil)
	for pager.More() {
		nextResult, err := pager.NextPage(kt.Ctx)
		if err != nil {
			logs.Errorf("list eip next page failed, err: %v, rid: %s", err, kt.Rid)
			return 0, fmt.Errorf("failed to advance page: %v", err)
		}

		count += int32(len(nextResult.Value))
	}

	return count, nil
}

// ListEipByPage ...
// reference: https://learn.microsoft.com/zh-cn/rest/api/virtualnetwork/public-ip-addresses/list-all?tabs=HTTP
func (az *Azure) ListEipByPage(kt *kit.Kit, opt *core.AzureListOption) (
	*Pager[armnetwork.PublicIPAddressesClientListResponse, eip.AzureEip], error) {

	client, err := az.clientSet.publicIPAddressesClient()
	if err != nil {
		return nil, err
	}

	azurePager := client.NewListPager(opt.ResourceGroupName, nil)

	pager := &Pager[armnetwork.PublicIPAddressesClientListResponse, eip.AzureEip]{
		pager: azurePager,
		resultHandler: &eipResultHandler{
			resGroupName: opt.ResourceGroupName,
		},
	}

	return pager, nil
}

func convertEip(one *armnetwork.PublicIPAddress, resGroupName string) *eip.AzureEip {
	eipTmp := &eip.AzureEip{
		CloudID:                strings.ToLower(*one.ID),
		Name:                   SPtrToLowerSPtr(one.Name),
		Region:                 StrToLowerNoSpaceStr(*one.Location),
		Status:                 converter.ValToPtr(string(enumor.EipUnBind)),
		PublicIp:               one.Properties.IPAddress,
		Zones:                  one.Zones,
		ResourceGroupName:      strings.ToLower(resGroupName),
		Location:               one.Location,
		PublicIPAddressVersion: (*string)(one.Properties.PublicIPAddressVersion),
	}

	if one.Properties.DNSSettings != nil {
		eipTmp.Fqdn = one.Properties.DNSSettings.Fqdn
	}

	if one.Properties.IPConfiguration != nil {
		if one.Properties.IPConfiguration.ID != nil {
			eipTmp.IpConfigurationID = SPtrToLowerSPtr(one.Properties.IPConfiguration.ID)
			eipTmp.Status = converter.ValToPtr(string(enumor.EipBind))
		}
	}

	if one.SKU != nil {
		if one.SKU.Name != nil {
			eipTmp.SKU = converter.ValToPtr(string(*one.SKU.Name))
		}
		if one.SKU.Tier != nil {
			eipTmp.SKUTier = converter.ValToPtr(string(*one.SKU.Tier))
		}
	}
	return eipTmp
}

type eipResultHandler struct {
	resGroupName string
}

func (handler *eipResultHandler) BuildResult(resp armnetwork.PublicIPAddressesClientListResponse) []eip.AzureEip {
	details := make([]eip.AzureEip, 0, len(resp.Value))
	for _, eip := range resp.Value {
		details = append(details, converter.PtrToVal(convertEip(eip, handler.resGroupName)))
	}

	return details
}

// DeleteEip ...
// reference: https://learn.microsoft.com/zh-cn/rest/api/virtualnetwork/public-ip-addresses/delete?tabs=HTTP
func (az *Azure) DeleteEip(kt *kit.Kit, opt *eip.AzureEipDeleteOption) error {
	if opt == nil {
		return errf.New(errf.InvalidParameter, "azure eip delete option is required")
	}

	if err := opt.Validate(); err != nil {
		return err
	}

	client, err := az.clientSet.publicIPAddressesClient()
	if err != nil {
		return err
	}
	pollerResp, err := client.BeginDelete(kt.Ctx, opt.ResourceGroupName, opt.EipName, nil)
	if err != nil {
		return fmt.Errorf("failed to finish the request:  %v", err)
	}
	_, err = pollerResp.PollUntilDone(kt.Ctx, nil)

	return err
}

// AssociateEip ...
// reference: https://learn.microsoft.com/zh-cn/rest/api/virtualnetwork/network-interfaces/create-or-update?tabs=Go
func (az *Azure) AssociateEip(kt *kit.Kit, opt *eip.AzureEipAssociateOption) error {
	if opt == nil {
		return errf.New(errf.InvalidParameter, "azure eip associate option is required")
	}

	params, err := opt.ToInterfaceParams()
	if err != nil {
		return err
	}

	client, err := az.clientSet.networkInterfaceClient()
	if err != nil {
		return err
	}

	pollerResp, err := client.BeginCreateOrUpdate(
		kt.Ctx,
		opt.ResourceGroupName,
		*opt.NetworkInterface.Name,
		*params,
		nil,
	)
	if err != nil {
		return fmt.Errorf("azure associate eip failed to finish the request, eipID: %s, err: %v",
			opt.CloudEipID, err)
	}
	_, err = pollerResp.PollUntilDone(kt.Ctx, nil)

	return err
}

// DisassociateEip ...
// reference: https://learn.microsoft.com/zh-cn/rest/api/virtualnetwork/network-interfaces/create-or-update?tabs=Go
func (az *Azure) DisassociateEip(kt *kit.Kit, opt *eip.AzureEipDisassociateOption) error {
	if opt == nil {
		return errf.New(errf.InvalidParameter, "azure eip associate option is required")
	}

	params, err := opt.ToInterfaceParams()
	if err != nil {
		return err
	}
	if params == nil {
		return nil
	}

	client, err := az.clientSet.networkInterfaceClient()
	if err != nil {
		return err
	}

	pollerResp, err := client.BeginCreateOrUpdate(
		kt.Ctx,
		opt.ResourceGroupName,
		*opt.NetworkInterface.Name,
		*params,
		nil,
	)
	if err != nil {
		return fmt.Errorf("azure disassociate eip failed to finish the request, eipID: %s, err: %v",
			opt.CloudEipID, err)
	}
	_, err = pollerResp.PollUntilDone(kt.Ctx, nil)

	return err
}

// CreateEip ...
// reference: https://learn.microsoft.com/zh-cn/rest/api/virtualnetwork/public-ip-addresses/create-or-update?tabs=HTTP
func (az *Azure) CreateEip(kt *kit.Kit, opt *eip.AzureEipCreateOption) (*string, error) {
	if opt == nil {
		return nil, errf.New(errf.InvalidParameter, "azure eip create option is required")
	}

	params, err := opt.ToPublicIPAddress()
	if err != nil {
		return nil, err
	}

	client, err := az.clientSet.publicIPAddressesClient()
	if err != nil {
		return nil, err
	}

	pollerResp, err := client.BeginCreateOrUpdate(kt.Ctx, opt.ResourceGroupName, opt.EipName, *params, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to finish the request:  %v", err)
	}
	resp, err := pollerResp.PollUntilDone(kt.Ctx, nil)

	return SPtrToLowerSPtr(resp.ID), err
}
