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
	"strconv"

	"hcm/pkg/adaptor/types/core"
	routetable "hcm/pkg/adaptor/types/route-table"
	"hcm/pkg/kit"
	"hcm/pkg/logs"
	"hcm/pkg/tools/converter"

	vpc "github.com/tencentcloud/tencentcloud-sdk-go/tencentcloud/vpc/v20170312"
)

// UpdateRouteTable update route table.
// TODO right now only memo is supported to update, add other update operations later.
func (t *TCloudImpl) UpdateRouteTable(_ *kit.Kit, _ *routetable.TCloudRouteTableUpdateOption) error {
	return nil
}

// DeleteRouteTable delete route table.
// reference: https://cloud.tencent.com/document/api/215/15771
func (t *TCloudImpl) DeleteRouteTable(kt *kit.Kit, opt *core.BaseRegionalDeleteOption) error {
	if err := opt.Validate(); err != nil {
		return err
	}

	VpcClient, err := t.clientSet.VpcClient(opt.Region)
	if err != nil {
		return fmt.Errorf("new route table client failed, err: %v", err)
	}

	req := vpc.NewDeleteRouteTableRequest()
	req.RouteTableId = converter.ValToPtr(opt.ResourceID)

	_, err = VpcClient.DeleteRouteTableWithContext(kt.Ctx, req)
	if err != nil {
		logs.Errorf("delete tencent cloud route table failed, err: %v, rid: %s", err, kt.Rid)
		return err
	}

	return nil
}

// ListRouteTable list route table.
// reference: https://cloud.tencent.com/document/api/215/15763
func (t *TCloudImpl) ListRouteTable(kt *kit.Kit, opt *core.TCloudListOption) (*routetable.TCloudRouteTableListResult,
	error) {

	if err := opt.Validate(); err != nil {
		return nil, err
	}

	VpcClient, err := t.clientSet.VpcClient(opt.Region)
	if err != nil {
		return nil, fmt.Errorf("new route table client failed, err: %v", err)
	}

	req := vpc.NewDescribeRouteTablesRequest()
	if len(opt.CloudIDs) != 0 {
		req.RouteTableIds = converter.SliceToPtr(opt.CloudIDs)
		req.Limit = converter.ValToPtr(strconv.FormatUint(core.TCloudQueryLimit, 10))
	}

	if opt.Page != nil {
		req.Offset = converter.ValToPtr(strconv.FormatUint(opt.Page.Offset, 10))
		req.Limit = converter.ValToPtr(strconv.FormatUint(opt.Page.Limit, 10))
	}

	// **NOTICE** this api will not return default route
	resp, err := VpcClient.DescribeRouteTablesWithContext(kt.Ctx, req)
	if err != nil {
		logs.Errorf("list tencent cloud route table failed, opt: %+v, err: %v, rid: %s", opt, err, kt.Rid)
		return nil, fmt.Errorf("list tencent cloud route table failed, err: %v", err)
	}

	details := make([]routetable.TCloudRouteTable, 0, len(resp.Response.RouteTableSet))

	for _, data := range resp.Response.RouteTableSet {
		details = append(details, converter.PtrToVal(convertRouteTable(data, opt.Region)))
	}

	return &routetable.TCloudRouteTableListResult{Count: resp.Response.TotalCount, Details: details}, nil
}

// CountRouteTable 基于 DescribeRouteTablesWithContext
// reference: https://cloud.tencent.com/document/api/215/15763
func (t *TCloudImpl) CountRouteTable(kt *kit.Kit, region string) (int32, error) {

	client, err := t.clientSet.VpcClient(region)
	if err != nil {
		return 0, fmt.Errorf("new tcloud vpc client failed, err: %v", err)
	}

	req := vpc.NewDescribeRouteTablesRequest()
	req.Limit = converter.ValToPtr("1")
	resp, err := client.DescribeRouteTablesWithContext(kt.Ctx, req)
	if err != nil {
		logs.Errorf("count tcloud route table failed, err: %v, region: %s, rid: %s", err, region, kt.Rid)
		return 0, err
	}
	return int32(*resp.Response.TotalCount), nil
}

func convertRouteTable(data *vpc.RouteTable, region string) *routetable.TCloudRouteTable {
	if data == nil {
		return nil
	}

	r := &routetable.TCloudRouteTable{
		CloudID:    converter.PtrToVal(data.RouteTableId),
		Name:       converter.PtrToVal(data.RouteTableName),
		CloudVpcID: converter.PtrToVal(data.VpcId),
		Region:     region,
		Extension: &routetable.TCloudRouteTableExtension{
			Main: converter.PtrToVal(data.Main),
		},
	}

	for _, asst := range data.AssociationSet {
		if asst == nil {
			continue
		}

		asstRouteTableID := converter.PtrToVal(asst.RouteTableId)
		if asstRouteTableID != r.CloudID {
			logs.Errorf("tcloud route table %s association id %s not match", r.CloudID, asstRouteTableID)
		}

		r.Extension.Associations = append(r.Extension.Associations, routetable.TCloudRouteTableAsst{
			CloudSubnetID: converter.PtrToVal(asst.SubnetId),
		})
	}

	for _, route := range data.RouteSet {
		if route == nil {
			continue
		}

		r.Extension.Routes = append(r.Extension.Routes, *convertRoute(route, r.CloudID))
	}

	return r
}
