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
	"time"

	cloudserver "hcm/pkg/api/cloud-server"
	csvpc "hcm/pkg/api/cloud-server/vpc"
	"hcm/pkg/api/core"
	"hcm/pkg/api/core/cloud"
	corert "hcm/pkg/api/core/cloud/route-table"
	"hcm/pkg/api/hc-service/sync"
	"hcm/pkg/criteria/constant"
	"hcm/pkg/criteria/enumor"
	"hcm/pkg/dal/dao/tools"
	"hcm/pkg/runtime/filter"
	"hcm/pkg/tools/converter"
	"hcm/test/suite"
	"hcm/test/suite/cases"

	. "github.com/smartystreets/goconvey/convey"
)

// TestTCloudVPC 测试腾讯云VPC、子网、路由表 相关api, 这三个类资源基本是一起出现，所以一起测试
func TestTCloudVPC(t *testing.T) {
	cli := suite.GetClientSet()
	testVpcName := "vpc1"
	testTcloudAccountID := accountID
	var createdVpc cloud.BaseVpc
	var createdSubnet cloud.BaseSubnet
	var createdRouteTable corert.BaseRouteTable

	createReq := csvpc.TCloudVpcCreateReq{
		AccountID: testTcloudAccountID,
		Region:    constant.SuiteRegion,
		Name:      testVpcName,
		IPv4Cidr:  "172.31.0.0/16",
		BkCloudID: 9911,
		Memo:      converter.ValToPtr("created by suite test"),
	}
	createReq.Subnet.Name = "subnet_of_vpc1"
	createReq.Subnet.Zone = constant.SuiteZone
	createReq.Subnet.IPv4Cidr = "172.31.1.0/24"

	Convey("VPC test", t, func() {
		// 1. 列出空vpc
		Convey("list VPC in res", func() {
			kt := cases.GenApiKit()
			listReq := core.ListReq{Page: &core.BasePage{Limit: 500}, Filter: tools.AllExpression()}
			listResult, err := cli.CloudServer().Vpc.ListInRes(kt, &listReq)
			So(err, ShouldBeNil)
			So(listResult, ShouldNotBeNil)
			So(len(listResult.Details), ShouldEqual, 0)
		})

		// 2. 创建vpc，与其是要创建出对应的子网，子网创建会创建对应路由表
		Convey("create tcloud VPC res", func() {
			kt := cases.GenApiKit()

			created, err := cli.CloudServer().Vpc.CreateTCloudVpc(kt, &createReq)
			So(err, ShouldBeNil)
			So(created.ID, ShouldNotBeEmpty)
			createdVpc.ID = created.ID

		})

		// 3. 尝试分配到业务下
		Convey("assign VPC to business", func() {
			kt := cases.GenApiKit()
			vpcAssign := &csvpc.AssignVpcToBizReq{
				VpcIDs:  []string{createdVpc.ID},
				BkBizID: constant.SuiteTestBizID,
			}
			err := cli.CloudServer().Vpc.Assign(kt, vpcAssign)
			So(err, ShouldBeNil)

		})

		// 4. 查询创建结果，包括VPC、对应的子网，以及子网对应的路由表
		Convey("check created VPC and subnet", func() {
			kt := cases.GenApiKit()
			Convey("check created VPC in business", func() {
				listReq := core.ListReq{Page: &core.BasePage{Limit: 500}, Filter: tools.AllExpression()}
				listResult, err := cli.CloudServer().Vpc.ListInBiz(kt, constant.SuiteTestBizID, &listReq)
				So(err, ShouldBeNil)
				So(listResult, ShouldNotBeNil)
				So(listResult.Details, ShouldHaveLength, 1)
				createdVpc = listResult.Details[0]
				So(createdVpc.Name, ShouldEqual, testVpcName)
				So(createdVpc.ID, ShouldEqual, createdVpc.ID)
				So(createdVpc.CloudID, ShouldNotBeEmpty)

			})
			Convey("check generated subnet", func() {
				// 	查询对应的子网和路由表
				subnetListReq := core.ListReq{Page: &core.BasePage{Limit: 10}, Filter: tools.AllExpression()}
				subnetListReq.Filter = &filter.Expression{
					Op: filter.And,
					Rules: []filter.RuleFactory{
						&filter.AtomRule{Field: "account_id", Op: filter.Equal.Factory(), Value: testTcloudAccountID},
						&filter.AtomRule{Field: "cloud_vpc_id", Op: filter.Equal.Factory(), Value: createdVpc.CloudID},
					},
				}
				subnetListResult, err := cli.CloudServer().Subnet.ListInRes(kt, &subnetListReq)
				So(err, ShouldBeNil)
				So(subnetListResult, ShouldNotBeNil)
				So(subnetListResult.Details, ShouldHaveLength, 1)
				createdSubnet = subnetListResult.Details[0]
				So(createdSubnet.Name, ShouldEqual, createReq.Subnet.Name)
				So(createdSubnet.CloudVpcID, ShouldEqual, createdVpc.CloudID)
				So(createdSubnet.CloudID, ShouldNotBeEmpty)

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
				createdRouteTable = routeTableResult.Details[0]
				So(createdRouteTable.Name, ShouldEqual, "default")
				So(createdRouteTable.CloudVpcID, ShouldEqual, createdVpc.CloudID)
				So(createdRouteTable.CloudID, ShouldNotBeEmpty)

				Convey("assign route table to business", func() {
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
		Convey("update VPC", func() {
			kt := cases.GenApiKit()
			updateReq := &csvpc.VpcUpdateReq{
				Memo: converter.ValToPtr("vpc-memo-updated"),
			}
			err := cli.CloudServer().Vpc.UpdateBiz(kt, constant.SuiteTestBizID, createdVpc.ID, updateReq)
			So(err, ShouldBeNil)

			vpcResult, err := cli.CloudServer().Vpc.GetInBiz(kt, constant.SuiteTestBizID, createdVpc.ID)
			So(err, ShouldBeNil)
			So(vpcResult.Memo, ShouldEqual, updateReq.Memo)
		})

		// 测试创建子网
		Convey("test subnet", func() {
			kt := cases.GenApiKit()
			subnetCreate := cloudserver.TCloudSubnetCreateReq{
				BaseSubnetCreateReq: &cloudserver.BaseSubnetCreateReq{
					Vendor:     enumor.TCloud,
					AccountID:  accountID,
					CloudVpcID: createdVpc.CloudID,
					Name:       "tcloud-subnet-abc",
					Memo:       converter.ValToPtr("memo"),
				},
				Region:   constant.SuiteRegion,
				Zone:     constant.SuiteZone,
				IPv4Cidr: "192.168.1.0/24",
			}
			createResult, err := cli.CloudServer().Subnet.Create(kt, &subnetCreate)
			So(err, ShouldBeNil)
			So(createResult.ID, ShouldNotBeEmpty)

			Convey("assign subnet to business", func() {
				subnetAssign := &cloudserver.AssignSubnetToBizReq{
					SubnetIDs: []string{createResult.ID},
					BkBizID:   constant.SuiteTestBizID,
				}
				err = cli.CloudServer().Subnet.Assign(kt, subnetAssign)
				So(err, ShouldBeNil)

				subnetListReq := core.ListReq{Page: &core.BasePage{Limit: 10},
					Filter: tools.EqualExpression("id", createResult.ID)}
				subnetListResult, err := cli.CloudServer().Subnet.ListInRes(kt, &subnetListReq)
				So(err, ShouldBeNil)
				So(subnetListResult, ShouldNotBeNil)
				So(subnetListResult.Details, ShouldHaveLength, 1)
				createdSubnet1 := subnetListResult.Details[0]
				So(createdSubnet1.Name, ShouldEqual, subnetCreate.Name)
				So(createdSubnet1.Ipv4Cidr, ShouldEqual, []string{subnetCreate.IPv4Cidr})
				So(createdSubnet1.CloudVpcID, ShouldEqual, createdVpc.CloudID)
				So(createdSubnet1.CloudID, ShouldNotBeEmpty)
			})
		})

		// 5. 删除vpc并验证
		Convey("delete VPC and verify", func() {
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

			err = cli.HCService().TCloud.Subnet.SyncSubnet(kt.Ctx, kt.Header(), &sync.TCloudSyncReq{
				AccountID: accountID,
				Region:    constant.SuiteRegion,
			})
			So(err, ShouldBeNil)
			time.Sleep(time.Second)
			// 查询子网
			subnetResult, err := cli.CloudServer().Subnet.ListInRes(kt, &listReq)
			So(err, ShouldBeNil)
			So(subnetResult, ShouldNotBeNil)
			So(len(subnetResult.Details), ShouldEqual, 0)

		})
	})

}
