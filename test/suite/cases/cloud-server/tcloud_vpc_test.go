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

package cloudserver

import (
	"testing"

	cloudserver "hcm/pkg/api/cloud-server"
	csvpc "hcm/pkg/api/cloud-server/vpc"
	"hcm/pkg/api/core"
	"hcm/pkg/api/core/cloud"
	"hcm/pkg/criteria/constant"
	"hcm/pkg/dal/dao/tools"
	"hcm/pkg/runtime/filter"
	"hcm/pkg/tools/converter"
	"hcm/test/suite"
	"hcm/test/suite/cases"

	. "github.com/smartystreets/goconvey/convey"
)

// TestTCloudVPC 测试腾讯云VPC 相关api
func TestTCloudVPC(t *testing.T) {
	cli := suite.GetClientSet()
	testVpcName := "vpc1"
	testTcloudAccountID := accountID
	var createdVpc cloud.BaseVpc
	var createdSubnet cloud.BaseSubnet

	createReq := csvpc.TCloudVpcCreateReq{
		AccountID: testTcloudAccountID,
		Region:    "ap-mariana",
		Name:      testVpcName,
		IPv4Cidr:  "172.31.0.0/16",
		BkCloudID: 9911,
		Memo:      converter.ValToPtr("created by suit test"),
	}
	createReq.Subnet.Name = "subnet_of_vpc1"
	createReq.Subnet.Zone = "ap-mariana-6"
	createReq.Subnet.IPv4Cidr = "172.31.1.0/24"

	Convey("test vpc", t, func() {
		// 1. 列出空vpc
		Convey("list tcloud vpc", func() {
			kt := cases.GenApiKit()
			listReq := core.ListReq{Page: &core.BasePage{Limit: 500}, Filter: tools.AllExpression()}
			listResult, err := cli.CloudServer().Vpc.ListInRes(kt, &listReq)
			So(err, ShouldBeNil)
			So(listResult, ShouldNotBeNil)
			So(len(listResult.Details), ShouldEqual, 0)
		})

		// 2. 创建vpc，与其是要创建出对应的子网，子网创建会创建对应路由表
		Convey("create tcloud vpc", func() {
			kt := cases.GenApiKit()

			created, err := cli.CloudServer().Vpc.CreateTCloudVpc(kt, &createReq)
			So(err, ShouldBeNil)
			So(created.ID, ShouldNotBeEmpty)
			createdVpc.ID = created.ID

		})

		// 3. 尝试分配到业务下
		Convey("vpc assign to business", func() {
			kt := cases.GenApiKit()
			vpcAssign := &csvpc.AssignVpcToBizReq{
				VpcIDs:  []string{createdVpc.ID},
				BkBizID: constant.SuiteTestBizID,
			}
			err := cli.CloudServer().Vpc.Assign(kt, vpcAssign)
			So(err, ShouldBeNil)

		})

		// 4. 查询创建结果，包括VPC、对应的子网，以及子网对应的路由表
		Convey("check created vpc and subnet", func() {
			kt := cases.GenApiKit()
			Convey("check created vpc", func() {
				listReq := core.ListReq{Page: &core.BasePage{Limit: 500}, Filter: tools.AllExpression()}
				listResult, err := cli.CloudServer().Vpc.ListInBiz(kt, constant.SuiteTestBizID, &listReq)
				So(err, ShouldBeNil)
				So(listResult, ShouldNotBeNil)
				So(listResult.Details, ShouldHaveLength, 1)
				So(listResult.Details[0].Name, ShouldEqual, testVpcName)
				So(listResult.Details[0].ID, ShouldEqual, createdVpc.ID)
				So(listResult.Details[0].CloudID, ShouldNotBeEmpty)
				createdVpc = listResult.Details[0]
			})
			Convey("check generated subnet", func() {
				// 	查询对应的子网和路由表
				subnetListReq := core.ListReq{Page: &core.BasePage{Limit: 500}, Filter: tools.AllExpression()}
				subnetListReq.Filter = &filter.Expression{
					Op: filter.And,
					Rules: []filter.RuleFactory{
						&filter.AtomRule{Field: "account_id", Op: filter.Equal.Factory(), Value: testTcloudAccountID},
						&filter.AtomRule{Field: "cloud_vpc_id", Op: filter.Equal.Factory(), Value: createdVpc.CloudID},
					},
				}
				subnetListResult, err := cli.CloudServer().Subnet.ListInRes(kt.Ctx, kt.Header(), &subnetListReq)
				So(err, ShouldBeNil)
				So(subnetListResult, ShouldNotBeNil)
				So(subnetListResult.Details, ShouldHaveLength, 1)
				So(subnetListResult.Details[0].Name, ShouldEqual, createReq.Subnet.Name)
				So(subnetListResult.Details[0].CloudVpcID, ShouldEqual, createdVpc.CloudID)
				So(subnetListResult.Details[0].CloudID, ShouldNotBeEmpty)
				createdSubnet = subnetListResult.Details[0]

				Convey("assign subnet to biz", func() {
					subnetAssign := &cloudserver.AssignSubnetToBizReq{
						SubnetIDs: []string{createdSubnet.ID},
						BkBizID:   constant.SuiteTestBizID,
					}
					err = cli.CloudServer().Subnet.Assign(kt, subnetAssign)
					So(err, ShouldBeNil)

					subnetListResult, err := cli.CloudServer().Subnet.ListInBiz(kt, constant.SuiteTestBizID,
						&subnetListReq)
					So(err, ShouldBeNil)
					So(subnetListResult, ShouldNotBeNil)
					So(subnetListResult.Details, ShouldHaveLength, 1)
				})

			})

			Convey("check generated route table", func() {
				routeTableListReq := core.ListReq{Page: &core.BasePage{Limit: 500}, Filter: tools.AllExpression()}
				routeTableListReq.Filter = &filter.Expression{
					Op: filter.And,
					Rules: []filter.RuleFactory{
						&filter.AtomRule{Field: "account_id", Op: filter.Equal.Factory(), Value: testTcloudAccountID},
						&filter.AtomRule{Field: "cloud_id", Op: filter.Equal.Factory(),
							Value: createdSubnet.CloudRouteTableID},
					},
				}
				routeTableResult, err := cli.CloudServer().RouteTable.ListInRes(kt, &routeTableListReq)
				So(err, ShouldBeNil)
				So(routeTableResult, ShouldNotBeNil)
				So(routeTableResult.Details, ShouldHaveLength, 1)
				So(routeTableResult.Details[0].Name, ShouldEqual, "default")
				So(routeTableResult.Details[0].CloudVpcID, ShouldEqual, createdVpc.CloudID)
				So(routeTableResult.Details[0].CloudID, ShouldNotBeEmpty)

				Convey("assign route table to biz", func() {
					subnetAssign := &cloudserver.AssignRouteTableToBizReq{
						RouteTableIDs: []string{createdSubnet.ID},
						BkBizID:       constant.SuiteTestBizID,
					}

					err = cli.CloudServer().RouteTable.Assign(kt, subnetAssign)
					So(err, ShouldBeNil)
					routeTableResult, err := cli.CloudServer().RouteTable.ListInBiz(kt, constant.SuiteTestBizID,
						&routeTableListReq)
					So(err, ShouldBeNil)
					So(routeTableResult, ShouldNotBeNil)
					So(routeTableResult.Details, ShouldHaveLength, 1)
				})

			})

		})

		// 4. 修改vpc属性并验证
		Convey("update vpc", func() {
			kt := cases.GenApiKit()
			updateReq := &csvpc.VpcUpdateReq{
				Memo: converter.ValToPtr("vpc-name-updated"),
			}
			err := cli.CloudServer().Vpc.UpdateBiz(kt, constant.SuiteTestBizID, createdVpc.ID, updateReq)
			So(err, ShouldBeNil)

			vpcResult, err := cli.CloudServer().Vpc.GetInBiz(kt, constant.SuiteTestBizID, createdVpc.ID)
			So(err, ShouldBeNil)
			So(vpcResult.Memo, ShouldEqual, updateReq.Memo)
		})
		// 5. 删除vpc并验证
		Convey("delete vpc and verify", func() {
			kt := cases.GenApiKit()
			err := cli.CloudServer().Vpc.DeleteBiz(kt, constant.SuiteTestBizID, createdVpc.ID)
			So(err, ShouldBeNil)

			listReq := core.ListReq{Page: &core.BasePage{Limit: 2}, Filter: tools.AllExpression()}
			bizListResult, err := cli.CloudServer().Vpc.ListInBiz(kt, constant.SuiteTestBizID, &listReq)
			So(err, ShouldBeNil)
			So(bizListResult, ShouldNotBeNil)
			So(len(bizListResult.Details), ShouldEqual, 0)

			resListResult, err := cli.CloudServer().Vpc.ListInRes(kt, &listReq)
			So(err, ShouldBeNil)
			So(resListResult, ShouldNotBeNil)
			So(len(resListResult.Details), ShouldEqual, 0)

		})
	})

}
