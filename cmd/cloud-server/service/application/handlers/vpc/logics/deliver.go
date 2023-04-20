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

package logics

import (
	"hcm/cmd/cloud-server/logics/audit"
	protocloud "hcm/pkg/api/data-service/cloud"
	routetable "hcm/pkg/api/data-service/cloud/route-table"
	dataservice "hcm/pkg/client/data-service"
	"hcm/pkg/criteria/enumor"
	"hcm/pkg/kit"
)

// DeliverVpc 交付vpc
func DeliverVpc(kt *kit.Kit, bkBizID int64, datacli *dataservice.Client,
	audit audit.Interface, vpcID string) (map[string]interface{}, error) {
	// update vpc bk_biz_id
	vpcs := []protocloud.VpcBaseInfoUpdateReq{
		{
			IDs: []string{vpcID},
			Data: &protocloud.VpcUpdateBaseInfo{
				BkBizID: bkBizID,
			},
		},
	}

	err := datacli.Global.Vpc.BatchUpdateBaseInfo(
		kt.Ctx,
		kt.Header(),
		&protocloud.VpcBaseInfoBatchUpdateReq{Vpcs: vpcs},
	)
	if err != nil {
		return map[string]interface{}{"error": err.Error()}, err
	}

	// create deliver audit
	err = audit.ResDeliverAudit(kt, enumor.VpcCloudAuditResType, []string{vpcID}, bkBizID)
	if err != nil {
		return map[string]interface{}{"error": err.Error()}, err
	}

	return map[string]interface{}{}, nil
}

// DeliverSubnet 交付subnet
func DeliverSubnet(kt *kit.Kit, bkBizID int64, datacli *dataservice.Client,
	audit audit.Interface, subnetIDs []string) (map[string]interface{}, error) {
	// update subnet bk_biz_id
	subnets := []protocloud.SubnetBaseInfoUpdateReq{
		{
			IDs: subnetIDs,
			Data: &protocloud.SubnetUpdateBaseInfo{
				BkBizID: bkBizID,
			},
		},
	}

	err := datacli.Global.Subnet.BatchUpdateBaseInfo(
		kt.Ctx,
		kt.Header(),
		&protocloud.SubnetBaseInfoBatchUpdateReq{Subnets: subnets},
	)
	if err != nil {
		return map[string]interface{}{"error": err.Error()}, err
	}

	// create deliver audit
	err = audit.ResDeliverAudit(kt, enumor.SubnetAuditResType, subnetIDs, int64(bkBizID))
	if err != nil {
		return map[string]interface{}{"error": err.Error()}, err
	}

	return map[string]interface{}{}, nil
}

// DeliverRouteTable 交付routetable
func DeliverRouteTable(kt *kit.Kit, bkBizID int64, datacli *dataservice.Client,
	audit audit.Interface, routeTableIDs []string) (map[string]interface{}, error) {
	// update routetable bk_biz_id
	routeTables := []routetable.RouteTableBaseInfoUpdateReq{
		{
			IDs: routeTableIDs,
			Data: &routetable.RouteTableUpdateBaseInfo{
				BkBizID: bkBizID,
			},
		},
	}

	err := datacli.Global.RouteTable.BatchUpdateBaseInfo(
		kt.Ctx,
		kt.Header(),
		&routetable.RouteTableBaseInfoBatchUpdateReq{RouteTables: routeTables},
	)
	if err != nil {
		return map[string]interface{}{"error": err.Error()}, err
	}

	// create deliver audit
	err = audit.ResDeliverAudit(kt, enumor.RouteTableAuditResType, routeTableIDs, int64(bkBizID))
	if err != nil {
		return map[string]interface{}{"error": err.Error()}, err
	}

	return map[string]interface{}{}, nil
}
