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

	"hcm/pkg/adaptor/types/core"
	routetable "hcm/pkg/adaptor/types/route-table"
	"hcm/pkg/kit"
	"hcm/pkg/logs"
	"hcm/pkg/tools/converter"

	"github.com/Azure/azure-sdk-for-go/sdk/azcore/runtime"
	"github.com/Azure/azure-sdk-for-go/sdk/resourcemanager/network/armnetwork/v2"
)

// UpdateRouteTable update route table.
// TODO right now only memo is supported to update, add other update operations later.
func (az *Azure) UpdateRouteTable(_ *kit.Kit, _ *routetable.AzureRouteTableUpdateOption) error {
	return nil
}

// DeleteRouteTable delete route table.
// reference: https://learn.microsoft.com/en-us/rest/api/virtualnetwork/route-tables/delete?tabs=HTTP
func (az *Azure) DeleteRouteTable(kt *kit.Kit, opt *core.AzureDeleteOption) error {
	if err := opt.Validate(); err != nil {
		return err
	}

	routeTableClient, err := az.clientSet.routeTableClient()
	if err != nil {
		return fmt.Errorf("new route table client failed, err: %v", err)
	}

	poller, err := routeTableClient.BeginDelete(kt.Ctx, opt.ResourceGroupName, opt.ResourceID, nil)
	if err != nil {
		logs.Errorf("delete azure route table failed, err: %v, rid: %s", err, kt.Rid)
		return err
	}

	_, err = poller.PollUntilDone(kt.Ctx, new(runtime.PollUntilDoneOptions))
	if err != nil {
		return err
	}

	return nil
}

// ListRouteTable list route table.
// reference: https://learn.microsoft.com/en-us/rest/api/virtualnetwork/route-tables/list?tabs=HTTP
func (az *Azure) ListRouteTable(kt *kit.Kit, opt *core.AzureListOption) (*routetable.AzureRouteTableListResult, error) {
	if err := opt.Validate(); err != nil {
		return nil, err
	}

	routeTableClient, err := az.clientSet.routeTableClient()
	if err != nil {
		return nil, fmt.Errorf("new route table client failed, err: %v", err)
	}

	req := new(armnetwork.RouteTablesClientListOptions)

	pager := routeTableClient.NewListPager(opt.ResourceGroupName, req)
	if err != nil {
		logs.Errorf("list azure route table failed, err: %v, rid: %s", err, kt.Rid)
		return nil, fmt.Errorf("list azure route table failed, err: %v", err)
	}

	details := make([]routetable.AzureRouteTable, 0)
	for pager.More() {
		page, err := pager.NextPage(kt.Ctx)
		if err != nil {
			return nil, fmt.Errorf("list azure route table but get next page failed, err: %v", err)
		}

		for _, routeTable := range page.Value {
			details = append(details, converter.PtrToVal(az.ConvertRouteTable(routeTable, opt.ResourceGroupName,
				az.clientSet.credential.CloudSubscriptionID)))
		}
	}

	return &routetable.AzureRouteTableListResult{Details: details}, nil
}

type routeTableResultHandler struct {
	resGroupName string
	a            *Azure
}

func (handler *routeTableResultHandler) BuildResult(resp armnetwork.RouteTablesClientListResponse) []routetable.AzureRouteTable {
	details := make([]routetable.AzureRouteTable, 0, len(resp.Value))
	for _, routeTable := range resp.Value {
		details = append(details, converter.PtrToVal(handler.a.ConvertRouteTable(routeTable, handler.resGroupName,
			handler.a.clientSet.credential.CloudSubscriptionID)))
	}

	return details
}

// ListRouteTableByPage ...
// reference: https://learn.microsoft.com/en-us/rest/api/virtualnetwork/route-tables/list?tabs=HTTP
func (az *Azure) ListRouteTableByPage(kt *kit.Kit, opt *core.AzureListOption) (
	*Pager[armnetwork.RouteTablesClientListResponse, routetable.AzureRouteTable], error) {

	client, err := az.clientSet.routeTableClient()
	if err != nil {
		return nil, err
	}

	azurePager := client.NewListPager(opt.ResourceGroupName, nil)

	pager := &Pager[armnetwork.RouteTablesClientListResponse, routetable.AzureRouteTable]{
		pager: azurePager,
		resultHandler: &routeTableResultHandler{
			resGroupName: opt.ResourceGroupName,
			a:            az,
		},
	}

	return pager, nil
}

// ListRouteTablePage list route table page.
// reference: https://learn.microsoft.com/en-us/rest/api/virtualnetwork/route-tables/list?tabs=HTTP
func (az *Azure) ListRouteTablePage(opt *core.AzureListByIDOption) (
	*runtime.Pager[armnetwork.RouteTablesClientListResponse], string, error) {

	if err := opt.Validate(); err != nil {
		return nil, "", err
	}

	routeTableClient, err := az.clientSet.routeTableClient()
	if err != nil {
		return nil, "", fmt.Errorf("new route table client failed, err: %v", err)
	}

	req := new(armnetwork.RouteTablesClientListOptions)

	pager := routeTableClient.NewListPager(opt.ResourceGroupName, req)
	return pager, az.clientSet.credential.CloudSubscriptionID, nil
}

// ListRouteTableByID list route table.
// reference: https://learn.microsoft.com/en-us/rest/api/virtualnetwork/route-tables/list?tabs=HTTP
func (az *Azure) ListRouteTableByID(kt *kit.Kit, opt *core.AzureListByIDOption) (
	*routetable.AzureRouteTableListResult, error) {

	if err := opt.Validate(); err != nil {
		return nil, err
	}

	routeTableClient, err := az.clientSet.routeTableClient()
	if err != nil {
		return nil, fmt.Errorf("new route table client failed, err: %v", err)
	}

	req := new(armnetwork.RouteTablesClientListOptions)

	pager := routeTableClient.NewListPager(opt.ResourceGroupName, req)

	idMap := converter.StringSliceToMap(opt.CloudIDs)
	details := make([]routetable.AzureRouteTable, 0, len(idMap))
	for pager.More() {
		nextResult, err := pager.NextPage(kt.Ctx)
		if err != nil {
			return nil, fmt.Errorf("list azure route table but get next page failed, err: %v", err)
		}

		for _, one := range nextResult.Value {
			id := SPtrToLowerSPtr(one.ID)
			if _, exist := idMap[*id]; exist {
				details = append(details, converter.PtrToVal(az.ConvertRouteTable(one, opt.ResourceGroupName,
					az.clientSet.credential.CloudSubscriptionID)))
				delete(idMap, *id)

				if len(idMap) == 0 {
					return &routetable.AzureRouteTableListResult{Details: details}, nil
				}
			}
		}
	}

	return &routetable.AzureRouteTableListResult{Details: details}, nil
}

// GetRouteTable get route table.
// reference: https://learn.microsoft.com/en-us/rest/api/virtualnetwork/route-tables/get?tabs=HTTP
func (az *Azure) GetRouteTable(kt *kit.Kit, opt *routetable.AzureRouteTableGetOption) (*routetable.AzureRouteTable,
	error) {

	if err := opt.Validate(); err != nil {
		return nil, err
	}

	routeTableClient, err := az.clientSet.routeTableClient()
	if err != nil {
		return nil, fmt.Errorf("new route table client failed, err: %v", err)
	}

	req := new(armnetwork.RouteTablesClientGetOptions)

	routeTableRes, err := routeTableClient.Get(kt.Ctx, opt.ResourceGroupName, opt.Name, req)
	if err != nil {
		logs.Errorf("list azure route table failed, err: %v, rid: %s", err, kt.Rid)
		return nil, fmt.Errorf("list azure route table failed, err: %v", err)
	}

	return az.ConvertRouteTable(&routeTableRes.RouteTable, opt.ResourceGroupName,
		az.clientSet.credential.CloudSubscriptionID), nil
}

// ConvertRouteTable ...
func (az *Azure) ConvertRouteTable(data *armnetwork.RouteTable, resourceGroup,
	subscription string) *routetable.AzureRouteTable {

	if data == nil {
		return nil
	}

	r := &routetable.AzureRouteTable{
		CloudID: SPtrToLowerStr(data.ID),
		Name:    SPtrToLowerStr(data.Name),
		Region:  SPtrToLowerNoSpaceStr(data.Location),
		Extension: &routetable.AzureRouteTableExtension{
			ResourceGroupName:   strings.ToLower(resourceGroup),
			CloudSubscriptionID: strings.ToLower(subscription),
			Routes:              nil,
			CloudSubnetIDs:      nil,
		},
	}

	if data.Properties == nil {
		return r
	}

	for _, route := range data.Properties.Routes {
		if route == nil {
			continue
		}

		r.Extension.Routes = append(r.Extension.Routes, *convertRoute(route, r.CloudID))
	}

	for _, subnet := range data.Properties.Subnets {
		if subnet == nil {
			continue
		}

		r.Extension.CloudSubnetIDs = append(r.Extension.CloudSubnetIDs, SPtrToLowerStr(subnet.ID))
	}

	return r
}
