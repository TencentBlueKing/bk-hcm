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
	corert "hcm/pkg/api/core/cloud/route-table"
	"hcm/pkg/api/hc-service/sync"
	"hcm/pkg/client"
	"hcm/pkg/criteria/constant"
	"hcm/pkg/dal/dao/tools"
	"hcm/pkg/kit"
	"hcm/pkg/runtime/filter"
	"hcm/pkg/tools/converter"
	"hcm/test/suite"
	"hcm/test/suite/cases"

	. "github.com/smartystreets/goconvey/convey"
)

var (
	// ResVpcID vpc id in res to share with other kind of resources
	ResVpcID string
	// ResVpcCloudID vpc cloud id in res to share with other kind of resources
	ResVpcCloudID string

	// BizVpcID vpc id in biz to share with other kind of resources
	BizVpcID string

	// BizVpcCloudID vpc cloud id in biz to share with other kind of resources
	BizVpcCloudID string
)

type vpcTester struct {
	kt                *kit.Kit
	cli               *client.ClientSet
	vpcName           string
	accountID         string
	createdVpc        cloud.BaseVpc
	createdSubnet     cloud.BaseSubnet
	createdRouteTable corert.BaseRouteTable
	vpcCreateReq      csvpc.TCloudVpcCreateReq
	alreadyInRes      int
	alreadyInBiz      int
}

func newVpcTester(accountID, vpcName string) *vpcTester {
	vpcCreateReq := csvpc.TCloudVpcCreateReq{
		Name:      vpcName,
		AccountID: accountID,
		Region:    constant.SuiteRegion,
		IPv4Cidr:  "172.31.0.0/16",
		Memo:      converter.ValToPtr("memo(create vpc)"),
	}
	vpcCreateReq.Subnet.Name = "subnet_of_vpc1"
	vpcCreateReq.Subnet.Zone = constant.SuiteZone
	vpcCreateReq.Subnet.IPv4Cidr = "172.31.1.0/24"
	return &vpcTester{
		kt:           cases.GenApiKit(),
		cli:          suite.GetClientSet(),
		vpcName:      vpcName,
		accountID:    accountID,
		vpcCreateReq: vpcCreateReq,
		alreadyInRes: 2,
		alreadyInBiz: 1,
	}
}

func (v *vpcTester) listInitVpc() {
	kt := v.kt.NewSubKit()
	listReq := core.ListReq{Page: &core.BasePage{Limit: 500}, Filter: tools.AllExpression()}
	listResult, err := v.cli.CloudServer().Vpc.ListInRes(kt, &listReq)
	So(err, ShouldBeNil)
	So(listResult, ShouldNotBeNil)
	So(listResult.Details, ShouldHaveLength, v.alreadyInRes)
}

func (v *vpcTester) createVpcInRes() {
	kt := v.kt.NewSubKit()

	created, err := v.cli.CloudServer().Vpc.CreateTCloudVpc(kt, &v.vpcCreateReq)
	So(err, ShouldBeNil)
	So(created.ID, ShouldNotBeEmpty)
	v.createdVpc.ID = created.ID

}
func (v *vpcTester) assignVpc() {
	kt := v.kt.NewSubKit()
	vpcAssign := &csvpc.AssignVpcToBizReq{
		VpcIDs:  []string{v.createdVpc.ID},
		BkBizID: constant.SuiteTestBizID,
	}
	err := v.cli.CloudServer().Vpc.Assign(kt, vpcAssign)
	So(err, ShouldBeNil)
}
func (v *vpcTester) checkCreatedVpcInBiz() {
	kt := v.kt.NewSubKit()

	listReq := core.ListReq{Page: &core.BasePage{Limit: 500}, Filter: tools.EqualExpression("id", v.createdVpc.ID)}
	listResult, err := v.cli.CloudServer().Vpc.ListInBiz(kt, constant.SuiteTestBizID, &listReq)
	So(err, ShouldBeNil)
	So(listResult, ShouldNotBeNil)
	So(listResult.Details, ShouldHaveLength, 1)
	v.createdVpc = listResult.Details[0]
	So(v.createdVpc.Name, ShouldEqual, v.vpcCreateReq.Name)
	So(v.createdVpc.CloudID, ShouldNotBeEmpty)
	So(v.createdVpc.Memo, ShouldEqual, v.vpcCreateReq.Memo)
}

func (v *vpcTester) checkGeneratedSubnet() {
	kt := v.kt.NewSubKit()

	// 	查询对应的子网和路由表
	subnetListReq := core.ListReq{Page: &core.BasePage{Limit: 10}}
	subnetListReq.Filter = &filter.Expression{
		Op: filter.And,
		Rules: []filter.RuleFactory{
			&filter.AtomRule{Field: "account_id", Op: filter.Equal.Factory(), Value: v.accountID},
			&filter.AtomRule{Field: "cloud_vpc_id", Op: filter.Equal.Factory(), Value: v.createdVpc.CloudID},
		},
	}
	subnetListResult, err := v.cli.CloudServer().Subnet.ListInRes(kt, &subnetListReq)
	So(err, ShouldBeNil)
	So(subnetListResult, ShouldNotBeNil)
	So(subnetListResult.Details, ShouldHaveLength, 1)
	v.createdSubnet = subnetListResult.Details[0]
	So(v.createdSubnet.Name, ShouldEqual, v.vpcCreateReq.Subnet.Name)
	So(v.createdSubnet.CloudVpcID, ShouldEqual, v.createdVpc.CloudID)
	So(v.createdSubnet.CloudID, ShouldNotBeEmpty)
}

func (v *vpcTester) checkGeneratedRouteTable() {

	kt := v.kt.NewSubKit()
	routeTableListReq := core.ListReq{Page: &core.BasePage{Limit: 10}}
	routeTableListReq.Filter = &filter.Expression{
		Op: filter.And,
		Rules: []filter.RuleFactory{
			&filter.AtomRule{Field: "account_id", Op: filter.Equal.Factory(), Value: v.accountID},
			&filter.AtomRule{Field: "cloud_id", Op: filter.Equal.Factory(),
				Value: v.createdSubnet.CloudRouteTableID},
		},
	}
	routeTableResult, err := v.cli.CloudServer().RouteTable.ListInRes(kt, &routeTableListReq)
	So(err, ShouldBeNil)
	So(routeTableResult, ShouldNotBeNil)
	So(routeTableResult.Details, ShouldHaveLength, 1)
	v.createdRouteTable = routeTableResult.Details[0]
	So(v.createdRouteTable.Name, ShouldEqual, "default")
	So(v.createdRouteTable.CloudVpcID, ShouldEqual, v.createdVpc.CloudID)
	So(v.createdRouteTable.CloudID, ShouldNotBeEmpty)

}
func (v *vpcTester) updateVpc() {
	kt := v.kt.NewSubKit()
	updateReq := &csvpc.VpcUpdateReq{
		Memo: converter.ValToPtr("memo(update vpc)"),
	}
	err := v.cli.CloudServer().Vpc.UpdateInBiz(kt, constant.SuiteTestBizID, v.createdVpc.ID, updateReq)
	So(err, ShouldBeNil)

	vpcResult, err := v.cli.CloudServer().Vpc.GetInBiz(kt, constant.SuiteTestBizID, v.createdVpc.ID)
	So(err, ShouldBeNil)
	So(vpcResult.Memo, ShouldEqual, updateReq.Memo)
}

func (v *vpcTester) delete() {
	kt := v.kt.NewSubKit()

	// 获取vpc 信息

	// 删除 vpc
	err := v.cli.CloudServer().Vpc.DeleteInBiz(kt, constant.SuiteTestBizID, v.createdVpc.ID)
	So(err, ShouldBeNil)

	// 确保vpc 被删除
	listReq := core.ListReq{Page: &core.BasePage{Limit: 500}, Filter: tools.EqualExpression("id", v.createdVpc.ID)}
	bizListResult, err := v.cli.CloudServer().Vpc.ListInRes(kt, &listReq)
	So(err, ShouldBeNil)
	So(bizListResult, ShouldNotBeNil)
	So(bizListResult.Details, ShouldHaveLength, 0)
	syncReq := &sync.TCloudSyncReq{
		AccountID: v.accountID,
		Region:    constant.SuiteRegion,
	}
	err = v.cli.HCService().TCloud.Subnet.SyncSubnet(kt.Ctx, kt.Header(), syncReq)
	So(err, ShouldBeNil)
	err = v.cli.HCService().TCloud.RouteTable.SyncRouteTable(kt.Ctx, kt.Header(), syncReq)
	So(err, ShouldBeNil)

	// 确保关联子网被删除
	listReq.Filter = tools.EqualExpression("cloud_vpc_id", v.createdVpc.CloudID)
	subnetResult, err := v.cli.CloudServer().Subnet.ListInRes(kt, &listReq)
	So(err, ShouldBeNil)
	So(subnetResult, ShouldNotBeNil)
	So(subnetResult.Details, ShouldHaveLength, 0)

	// 确保关联路由表被删除
	routeResult, err := v.cli.CloudServer().RouteTable.ListInRes(kt, &listReq)
	So(err, ShouldBeNil)
	So(routeResult, ShouldNotBeNil)
	So(routeResult.Details, ShouldHaveLength, 0)

}

// TestTCloudVPC 测试腾讯云VPC、子网、路由表 相关api, 这三个类资源基本是一起出现，所以一起测试
func TestTCloudVPC(t *testing.T) {

	tester := newVpcTester(TCloudAccountID, "vpc1")
	Convey("VPC test", t, func() {
		// 1. 列出空vpc
		Convey("list VPC in res", tester.listInitVpc)

		// 2. 创建vpc，同时创建出对应的子网，子网创建会创建对应路由表
		Convey("create tcloud VPC res", tester.createVpcInRes)

		// 3. 尝试分配到业务下
		Convey("assign VPC to business", tester.assignVpc)

		// 4. 查询创建结果，包括VPC、对应的子网，以及子网对应的路由表
		Convey("check created VPC in business", tester.checkCreatedVpcInBiz)
		Convey("check generated subnet", tester.checkGeneratedSubnet)
		Convey("check generated route table", tester.checkGeneratedRouteTable)

		// 5. 修改vpc属性并验证
		Convey("update VPC", tester.updateVpc)

		// 6. 删除vpc并验证
		Convey("delete VPC and verify", tester.delete)
	})

}

// TestCreateSharedVpc 创建给其他资源测试用的共享vpc，只会创建不会删除
func TestCreateSharedVpc(t *testing.T) {

	Convey("create shared vpc", t, func() {
		kt := cases.GenApiKit()
		cli := suite.GetClientSet()
		vpcCreateReq := newVpcCreateReq("test-vpc-res", TCloudAccountID, "192.168.1.0/24", "[res] share vpc for test")
		created, err := cli.CloudServer().Vpc.CreateTCloudVpc(kt, &vpcCreateReq)
		So(err, ShouldBeNil)
		So(created.ID, ShouldNotBeEmpty)
		ResVpcID = created.ID

		vpc, err := cli.CloudServer().Vpc.ListInRes(kt,
			&core.ListReq{Filter: tools.EqualExpression("id", ResVpcID), Page: core.NewDefaultBasePage()},
		)
		So(err, ShouldBeNil)
		So(vpc.Details, ShouldHaveLength, 1)
		ResVpcCloudID = vpc.Details[0].CloudID

		bizVpcCreateReq := newVpcCreateReq("test-vpc-biz", TCloudAccountID, "192.168.2.0/24",
			"[biz] share vpc for test")
		bizCreated, err := cli.CloudServer().Vpc.CreateTCloudVpc(kt, &bizVpcCreateReq)
		So(err, ShouldBeNil)
		So(bizCreated.ID, ShouldNotBeEmpty)
		BizVpcID = bizCreated.ID

		resVpc, err := cli.CloudServer().Vpc.ListInRes(kt,
			&core.ListReq{Filter: tools.EqualExpression("id", BizVpcID), Page: core.NewDefaultBasePage()},
		)
		So(err, ShouldBeNil)
		So(resVpc.Details, ShouldHaveLength, 1)
		BizVpcCloudID = resVpc.Details[0].CloudID

		// 分配到业务下
		vpcAssign := &csvpc.AssignVpcToBizReq{
			VpcIDs:  []string{BizVpcID},
			BkBizID: constant.SuiteTestBizID,
		}
		err = cli.CloudServer().Vpc.Assign(kt, vpcAssign)
		So(err, ShouldBeNil)

	})
}

func newVpcCreateReq(vpcName, accountID, cidr, memo string) csvpc.TCloudVpcCreateReq {
	vpcCreateReq := csvpc.TCloudVpcCreateReq{
		Name:      vpcName,
		AccountID: accountID,
		Region:    constant.SuiteRegion,
		IPv4Cidr:  cidr,
		Memo:      &memo,
	}
	vpcCreateReq.Subnet.Name = "subnet_of_" + vpcName
	vpcCreateReq.Subnet.Zone = constant.SuiteZone
	vpcCreateReq.Subnet.IPv4Cidr = cidr
	return vpcCreateReq
}
