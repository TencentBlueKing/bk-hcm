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

package huawei

import (
	"fmt"
	"strings"

	"hcm/pkg/adaptor/types/core"
	routetable "hcm/pkg/adaptor/types/route-table"
	"hcm/pkg/kit"
	"hcm/pkg/logs"
	"hcm/pkg/tools/converter"

	"github.com/huaweicloud/huaweicloud-sdk-go-v3/services/vpc/v2/model"
)

// UpdateRouteTable update route table.
// reference: https://support.huaweicloud.com/api-vpc/vpc_apiroutetab_0004.html
func (h *HuaWei) UpdateRouteTable(kt *kit.Kit, opt *routetable.HuaWeiRouteTableUpdateOption) error {
	if err := opt.Validate(); err != nil {
		return err
	}

	vpcClient, err := h.clientSet.vpcClientV2(opt.Region)
	if err != nil {
		return fmt.Errorf("new route table client failed, err: %v", err)
	}

	req := &model.UpdateRouteTableRequest{
		RoutetableId: opt.ResourceID,
		Body: &model.UpdateRoutetableReqBody{
			Routetable: &model.UpdateRouteTableReq{
				Description: opt.Data.Memo,
			},
		},
	}

	_, err = vpcClient.UpdateRouteTable(req)
	if err != nil {
		logs.Errorf("update huawei route table failed, err: %v, rid: %s", err, kt.Rid)
		return err
	}

	return nil
}

// DeleteRouteTable delete route table.
// reference: https://support.huaweicloud.com/api-vpc/vpc_apiroutetab_0006.html
func (h *HuaWei) DeleteRouteTable(kt *kit.Kit, opt *core.BaseRegionalDeleteOption) error {
	if err := opt.Validate(); err != nil {
		return err
	}

	vpcClient, err := h.clientSet.vpcClientV2(opt.Region)
	if err != nil {
		return fmt.Errorf("new route table client failed, err: %v", err)
	}

	req := &model.DeleteRouteTableRequest{
		RoutetableId: opt.ResourceID,
	}

	_, err = vpcClient.DeleteRouteTable(req)
	if err != nil {
		logs.Errorf("delete huawei route table failed, err: %v, rid: %s", err, kt.Rid)
		return err
	}

	return nil
}

// ListRouteTables list route table ids.
// reference: https://support.huaweicloud.com/api-vpc/vpc_apiroutetab_0001.html
func (h *HuaWei) ListRouteTables(kt *kit.Kit, opt *routetable.HuaWeiRouteTableListOption) ([]routetable.HuaWeiRouteTable, error) {
	if err := opt.Validate(); err != nil {
		return nil, err
	}

	vpcClient, err := h.clientSet.vpcClientV2(opt.Region)
	if err != nil {
		return nil, fmt.Errorf("new route table client failed, err: %v", err)
	}

	req := new(model.ListRouteTablesRequest)

	if len(opt.ID) != 0 {
		req.Id = &opt.ID
	}

	if len(opt.VpcID) != 0 {
		req.VpcId = &opt.VpcID
	}

	if len(opt.SubnetID) != 0 {
		req.SubnetId = &opt.SubnetID
	}

	if opt.Page != nil {
		req.Marker = opt.Page.Marker
		req.Limit = opt.Page.Limit
	}

	resp, err := vpcClient.ListRouteTables(req)
	if err != nil {
		if strings.Contains(err.Error(), ErrDataNotFound) {
			return nil, nil
		}
		logs.Errorf("list huawei route table failed, err: %v, rid: %s", err, kt.Rid)
		return nil, fmt.Errorf("list huawei route table failed, err: %v", err)
	}

	routeTables := make([]routetable.HuaWeiRouteTable, 0)

	for _, one := range converter.PtrToVal(resp.Routetables) {
		tmp := routetable.HuaWeiRouteTable{
			CloudID:    one.Id,
			Name:       one.Name,
			CloudVpcID: one.VpcId,
			Region:     opt.Region,
			Memo:       converter.ValToPtr(one.Description),
			Extension: &routetable.HuaWeiRouteTableExtension{
				Default:  one.Default,
				TenantID: one.TenantId,
			},
		}

		routeTables = append(routeTables, tmp)
	}

	return routeTables, nil
}

// ListRouteTableIDs list route table ids.
// reference: https://support.huaweicloud.com/api-vpc/vpc_apiroutetab_0001.html
func (h *HuaWei) ListRouteTableIDs(kt *kit.Kit, opt *routetable.HuaWeiRouteTableListOption) ([]string, error) {
	if err := opt.Validate(); err != nil {
		return nil, err
	}

	vpcClient, err := h.clientSet.vpcClientV2(opt.Region)
	if err != nil {
		return nil, fmt.Errorf("new route table client failed, err: %v", err)
	}

	req := new(model.ListRouteTablesRequest)

	if len(opt.ID) != 0 {
		req.Id = &opt.ID
	}

	if len(opt.VpcID) != 0 {
		req.VpcId = &opt.VpcID
	}

	if len(opt.SubnetID) != 0 {
		req.SubnetId = &opt.SubnetID
	}

	if opt.Page != nil {
		req.Marker = opt.Page.Marker
		req.Limit = opt.Page.Limit
	}

	resp, err := vpcClient.ListRouteTables(req)
	if err != nil {
		if strings.Contains(err.Error(), ErrDataNotFound) {
			return make([]string, 0), nil
		}
		logs.Errorf("list huawei route table failed, err: %v, rid: %s", err, kt.Rid)
		return nil, fmt.Errorf("list huawei route table failed, err: %v", err)
	}

	routeTables := converter.PtrToVal(resp.Routetables)
	ids := make([]string, 0, len(routeTables))

	for _, data := range routeTables {
		ids = append(ids, data.Id)
	}

	return ids, nil
}

// GetRouteTable get route table.
// reference: https://support.huaweicloud.com/api-vpc/vpc_apiroutetab_0002.html
func (h *HuaWei) GetRouteTable(kt *kit.Kit, opt *routetable.HuaWeiRouteTableGetOption) (*routetable.HuaWeiRouteTable,
	error) {

	if err := opt.Validate(); err != nil {
		return nil, err
	}

	vpcClient, err := h.clientSet.vpcClientV2(opt.Region)
	if err != nil {
		return nil, fmt.Errorf("new route table client failed, err: %v", err)
	}

	req := &model.ShowRouteTableRequest{
		RoutetableId: opt.ID,
	}

	resp, err := vpcClient.ShowRouteTable(req)
	if err != nil {
		logs.Errorf("get huawei route table failed, err: %v, rid: %s", err, kt.Rid)
		return nil, fmt.Errorf("get huawei route table failed, err: %v", err)
	}

	return convertRouteTable(resp.Routetable, opt.Region), nil
}

func convertRouteTable(data *model.RouteTableResp, region string) *routetable.HuaWeiRouteTable {
	if data == nil {
		return nil
	}

	r := &routetable.HuaWeiRouteTable{
		CloudID:    data.Id,
		Name:       data.Name,
		Memo:       &data.Description,
		CloudVpcID: data.VpcId,
		Region:     region,
		Extension: &routetable.HuaWeiRouteTableExtension{
			Default:        data.Default,
			CloudSubnetIDs: make([]string, 0, len(data.Subnets)),
			TenantID:       data.TenantId,
		},
	}

	for _, subnet := range data.Subnets {
		r.Extension.CloudSubnetIDs = append(r.Extension.CloudSubnetIDs, subnet.Id)
	}

	for _, route := range data.Routes {
		r.Extension.Routes = append(r.Extension.Routes, convertRoute(route, r.CloudID))
	}

	return r
}
