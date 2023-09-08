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

	csvpc "hcm/pkg/api/cloud-server/vpc"
	"hcm/pkg/api/core"
	"hcm/pkg/api/core/cloud"
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
	testTcloudAccountID := "00000003"
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
			listResult, err := cli.CloudServer().Vpc.ListInRes(kt.Ctx, kt.Header(), &listReq)
			So(err, ShouldBeNil)
			So(listResult, ShouldNotBeNil)
			So(len(listResult.Details), ShouldEqual, 0)
		})

		// 2. 创建vpc
		Convey("create tcloud vpc", func() {
			kt := cases.GenApiKit()

			created, err := cli.CloudServer().Vpc.CreateTCloudVpc(kt, &createReq)
			So(err, ShouldBeNil)
			So(created.ID, ShouldNotBeEmpty)
			createdVpc.ID = created.ID

		})

		// 3. 查询创建结果
		Convey("check created vpc and subnet", func() {
			kt := cases.GenApiKit()
			Convey("check created vpc", func() {
				listReq := core.ListReq{Page: &core.BasePage{Limit: 500}, Filter: tools.AllExpression()}
				listResult, err := cli.CloudServer().Vpc.ListInRes(kt.Ctx, kt.Header(), &listReq)
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
				routeTableResult, err := cli.CloudServer().RouteTable.ListInRes(kt.Ctx, kt.Header(), &routeTableListReq)
				So(err, ShouldBeNil)
				So(routeTableResult, ShouldNotBeNil)
				So(routeTableResult.Details, ShouldHaveLength, 1)
				So(routeTableResult.Details[0].Name, ShouldEqual, "default")
				So(routeTableResult.Details[0].CloudVpcID, ShouldEqual, createdVpc.CloudID)
				So(routeTableResult.Details[0].CloudID, ShouldNotBeEmpty)
			})

		})
	})

	// 4. 修改vpc属性并验证
	// 5. 删除vpc并验证
}
