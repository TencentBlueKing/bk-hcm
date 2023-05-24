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

package aws

import (
	"strings"

	"hcm/pkg/adaptor/types/core"
	routetable "hcm/pkg/adaptor/types/route-table"
	"hcm/pkg/kit"
	"hcm/pkg/logs"
	"hcm/pkg/tools/converter"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/ec2"
)

// UpdateRouteTable update route table.
// TODO right now only memo is supported to update, add other update operations later.
func (a *Aws) UpdateRouteTable(kt *kit.Kit, opt *routetable.AwsRouteTableUpdateOption) error {
	return nil
}

// DeleteRouteTable delete route table.
// reference: https://docs.amazonaws.cn/AWSEC2/latest/APIReference/API_DeleteRouteTable.html
func (a *Aws) DeleteRouteTable(kt *kit.Kit, opt *core.BaseRegionalDeleteOption) error {
	if err := opt.Validate(); err != nil {
		return err
	}

	client, err := a.clientSet.ec2Client(opt.Region)
	if err != nil {
		return err
	}

	req := &ec2.DeleteRouteTableInput{
		RouteTableId: aws.String(opt.ResourceID),
	}
	_, err = client.DeleteRouteTableWithContext(kt.Ctx, req)
	if err != nil {
		logs.Errorf("delete aws route table failed, err: %v, rid: %s", err, kt.Rid)
		return err
	}

	return nil
}

// ListRouteTable list route table.
// reference: https://docs.amazonaws.cn/AWSEC2/latest/APIReference/API_DescribeRouteTables.html
func (a *Aws) ListRouteTable(kt *kit.Kit, opt *routetable.AwsRouteTableListOption) (
	*routetable.AwsRouteTableListResult, error) {

	if err := opt.Validate(); err != nil {
		return nil, err
	}

	client, err := a.clientSet.ec2Client(opt.Region)
	if err != nil {
		return nil, err
	}

	req := new(ec2.DescribeRouteTablesInput)

	if len(opt.CloudIDs) != 0 {
		req.RouteTableIds = aws.StringSlice(opt.CloudIDs)
	}

	if len(opt.SubnetIDs) != 0 {
		req.Filters = append(req.Filters, &ec2.Filter{
			Name:   aws.String("association.subnet-id"),
			Values: aws.StringSlice(opt.SubnetIDs),
		})
	}

	if opt.Page != nil {
		req.NextToken = opt.Page.NextToken
		req.MaxResults = opt.Page.MaxResults
	}

	resp, err := client.DescribeRouteTablesWithContext(kt.Ctx, req)
	if err != nil {
		if !strings.Contains(err.Error(), ErrRouteTableNotFound) {
			logs.Errorf("list aws route table failed, err: %v, rid: %s", err, kt.Rid)
		}

		return nil, err
	}

	details := make([]routetable.AwsRouteTable, 0, len(resp.RouteTables))
	for _, routeTable := range resp.RouteTables {
		details = append(details, converter.PtrToVal(convertRouteTable(routeTable, opt.Region)))
	}

	return &routetable.AwsRouteTableListResult{NextToken: resp.NextToken, Details: details}, nil
}

func convertRouteTable(data *ec2.RouteTable, region string) *routetable.AwsRouteTable {
	if data == nil {
		return nil
	}

	r := &routetable.AwsRouteTable{
		CloudID:    converter.PtrToVal(data.RouteTableId),
		CloudVpcID: converter.PtrToVal(data.VpcId),
		Region:     region,
		Extension:  new(routetable.AwsRouteTableExtension),
	}

	name, _ := parseTags(data.Tags)
	r.Name = name

	for _, route := range data.Routes {
		if route == nil {
			continue
		}

		r.Extension.Routes = append(r.Extension.Routes, *convertRoute(route, r.CloudID))
	}

	for _, asst := range data.Associations {
		if asst == nil {
			continue
		}

		asstRouteTableID := converter.PtrToVal(asst.RouteTableId)
		if asstRouteTableID != r.CloudID {
			logs.Errorf("tcloud route table %s association id %s not match", r.CloudID, asstRouteTableID)
		}

		if converter.PtrToVal(asst.Main) {
			r.Extension.Main = true
		}

		if asst.SubnetId == nil && asst.GatewayId == nil {
			continue
		}

		r.Extension.Associations = append(r.Extension.Associations, routetable.AwsRouteTableAsst{
			AssociationState: converter.PtrToVal(converter.PtrToVal(asst.AssociationState).State),
			CloudGatewayID:   asst.GatewayId,
			CloudSubnetID:    asst.SubnetId,
		})
	}

	return r
}
