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

package common

import (
	"fmt"

	"hcm/pkg/api/core"
	protocloud "hcm/pkg/api/data-service/cloud"
	dataproto "hcm/pkg/api/data-service/cloud/route-table"
	dataclient "hcm/pkg/client/data-service"
	"hcm/pkg/criteria/enumor"
	"hcm/pkg/kit"
	"hcm/pkg/logs"
	"hcm/pkg/runtime/filter"
	"hcm/pkg/tools/converter"
)

// CancelRouteTableSubnetRel cancel route table and subnet rel.
func CancelRouteTableSubnetRel(kt *kit.Kit, dataCli *dataclient.Client, vendor enumor.Vendor, delCloudIDs []string) error {
	if len(delCloudIDs) == 0 {
		return nil
	}

	expr := &filter.Expression{
		Op: filter.And,
		Rules: []filter.RuleFactory{
			&filter.AtomRule{
				Field: "vendor",
				Op:    filter.Equal.Factory(),
				Value: vendor,
			},
			&filter.AtomRule{
				Field: "cloud_route_table_id",
				Op:    filter.In.Factory(),
				Value: delCloudIDs,
			},
		},
	}
	dbQueryReq := &core.ListReq{
		Filter: expr,
		Page:   &core.BasePage{Count: false, Start: 0, Limit: core.DefaultMaxPageLimit},
	}
	dbList, err := dataCli.Global.Subnet.List(kt.Ctx, kt.Header(), dbQueryReq)
	if err != nil {
		logs.Errorf("[%s] routetable-route batch cancel route table and subnet rel failed. delIDs: %v, err: %v, "+
			"rid: %s", vendor, delCloudIDs, err, kt.Rid)
		return err
	}

	if len(dbList.Details) == 0 {
		return nil
	}

	var subnetUpdateReq = &protocloud.SubnetBaseInfoBatchUpdateReq{}
	for _, item := range dbList.Details {
		tmpSubnet := protocloud.SubnetBaseInfoUpdateReq{
			IDs: []string{item.ID},
			Data: &protocloud.SubnetUpdateBaseInfo{
				Name:              converter.ValToPtr(item.Name),
				Ipv4Cidr:          item.Ipv4Cidr,
				Ipv6Cidr:          item.Ipv6Cidr,
				Memo:              item.Memo,
				BkBizID:           item.BkBizID,
				CloudRouteTableID: converter.ValToPtr(""),
				RouteTableID:      converter.ValToPtr(""),
			},
		}
		subnetUpdateReq.Subnets = append(subnetUpdateReq.Subnets, tmpSubnet)
	}
	return dataCli.Global.Subnet.BatchUpdateBaseInfo(kt.Ctx, kt.Header(), subnetUpdateReq)
}

// UpdateSubnetRouteTableByIDs update subnet's route_table
func UpdateSubnetRouteTableByIDs(kt *kit.Kit, vendor enumor.Vendor, subnetMap map[string]dataproto.RouteTableSubnetReq,
	dataCli *dataclient.Client) error {

	tmpCloudIDs := make([]string, 0)
	tmpCloudSubnetIDs := make([]string, 0)
	for tmpSubnetID, tmpRouteItem := range subnetMap {
		tmpCloudIDs = append(tmpCloudIDs, tmpRouteItem.CloudRouteTableID)
		tmpCloudSubnetIDs = append(tmpCloudSubnetIDs, tmpSubnetID)
	}
	subnetList, err := listSubnet(kt, vendor, dataCli, tmpCloudSubnetIDs)
	if err != nil {
		return err
	}

	rtListReq := &core.ListReq{
		Fields: []string{"id", "cloud_id"},
		Filter: &filter.Expression{
			Op: filter.And,
			Rules: []filter.RuleFactory{
				&filter.AtomRule{
					Field: "cloud_id",
					Op:    filter.In.Factory(),
					Value: tmpCloudIDs,
				},
				&filter.AtomRule{
					Field: "vendor",
					Op:    filter.Equal.Factory(),
					Value: vendor,
				},
			},
		},
		Page: core.NewDefaultBasePage(),
	}
	routeTableList, err := dataCli.Global.RouteTable.List(kt.Ctx, kt.Header(), rtListReq)
	if err != nil {
		return err
	}
	routeTableInfoMap := make(map[string]dataproto.RouteTableSubnetReq, 0)
	for _, rtItem := range routeTableList.Details {
		routeTableInfoMap[rtItem.CloudID] = dataproto.RouteTableSubnetReq{
			RouteTableID:      rtItem.ID,
			CloudRouteTableID: rtItem.CloudID,
		}
	}

	tmpSubnetArr := make([]protocloud.SubnetBaseInfoUpdateReq, 0)
	for _, tmpItem := range subnetList.Details {
		rtSubnetInfo, ok := subnetMap[tmpItem.CloudID]
		if !ok {
			continue
		}
		tmpSubnetReq := protocloud.SubnetBaseInfoUpdateReq{
			IDs: []string{tmpItem.ID},
			Data: &protocloud.SubnetUpdateBaseInfo{
				CloudRouteTableID: converter.ValToPtr(rtSubnetInfo.CloudRouteTableID),
			},
		}
		// 检查routeTable表的cloud_id是否存在
		if rtInfo, ok := routeTableInfoMap[rtSubnetInfo.CloudRouteTableID]; ok {
			tmpSubnetReq.Data.RouteTableID = converter.ValToPtr(rtInfo.RouteTableID)
		}
		tmpSubnetArr = append(tmpSubnetArr, tmpSubnetReq)
	}

	if len(tmpSubnetArr) > 0 {
		subnetReq := &protocloud.SubnetBaseInfoBatchUpdateReq{
			Subnets: tmpSubnetArr,
		}
		err = dataCli.Global.Subnet.BatchUpdateBaseInfo(kt.Ctx, kt.Header(), subnetReq)
	}

	return nil
}

func listSubnet(kt *kit.Kit, vendor enumor.Vendor, dataCli *dataclient.Client, tmpCloudSubnetIDs []string) (
	*protocloud.SubnetListResult, error) {

	subnetListReq := &core.ListReq{
		Fields: []string{"id", "cloud_id", "cloud_route_table_id", "route_table_id"},
		Filter: &filter.Expression{
			Op: filter.And,
			Rules: []filter.RuleFactory{
				&filter.AtomRule{
					Field: "cloud_id",
					Op:    filter.In.Factory(),
					Value: tmpCloudSubnetIDs,
				},
				&filter.AtomRule{
					Field: "vendor",
					Op:    filter.Equal.Factory(),
					Value: vendor,
				},
			},
		},
		Page: core.NewDefaultBasePage(),
	}
	subnetList, err := dataCli.Global.Subnet.List(kt.Ctx, kt.Header(), subnetListReq)
	if err != nil {
		logs.Errorf("%s-routetable update subnet route_table_id failed. cloud_ids: %v, err: %v",
			vendor, tmpCloudSubnetIDs, err)
		return nil, err
	}

	if len(subnetList.Details) == 0 {
		return nil, fmt.Errorf("subnets not find")
	}

	return subnetList, nil
}
