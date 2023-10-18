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

	cloudproto "hcm/pkg/api/cloud-server"
	"hcm/pkg/api/core"
	"hcm/pkg/api/hc-service/sync"
	"hcm/pkg/criteria/constant"
	"hcm/pkg/criteria/enumor"
	"hcm/pkg/dal/dao/tools"
	"hcm/pkg/tools/converter"
	"hcm/test/suite"
	"hcm/test/suite/cases"

	. "github.com/smartystreets/goconvey/convey"
)

func TestSubnet(t *testing.T) {

	Convey("test subnet", t, func() {

		cli := suite.GetClientSet()
		kt := cases.GenApiKit()
		vpcID := BizVpcCloudID

		So(vpcID, ShouldNotBeEmpty)

		var subnetID string
		createReq := cloudproto.TCloudSubnetCreateReq{
			BaseSubnetCreateReq: &cloudproto.BaseSubnetCreateReq{
				Vendor:     enumor.TCloud,
				AccountID:  TCloudAccountID,
				CloudVpcID: vpcID,
				Name:       "tcloud-subnet-abc",
				Memo:       converter.ValToPtr("memo(create subnet)"),
			},
			Region:   constant.SuiteRegion,
			Zone:     constant.SuiteZone,
			IPv4Cidr: "192.168.1.0/24",
		}
		createResult, err := cli.CloudServer().Subnet.Create(kt, &createReq)
		So(err, ShouldBeNil)
		So(createResult.ID, ShouldNotBeEmpty)
		subnetID = createResult.ID
		subnetListReq := core.ListReq{Page: &core.BasePage{Limit: 10},
			Filter: tools.EqualExpression("id", subnetID)}

		Convey("assign subnet to business", func() {
			assignReq := &cloudproto.AssignSubnetToBizReq{
				SubnetIDs: []string{subnetID},
				BkBizID:   constant.SuiteTestBizID,
			}
			err = cli.CloudServer().Subnet.Assign(kt, assignReq)
			So(err, ShouldBeNil)

			// 业务下查找
			subnetListResult, err := cli.CloudServer().Subnet.ListInBiz(kt, constant.SuiteTestBizID, &subnetListReq)
			So(err, ShouldBeNil)
			So(subnetListResult, ShouldNotBeNil)
			So(subnetListResult.Details, ShouldHaveLength, 1)
			createdSubnet1 := subnetListResult.Details[0]
			// 业务下
			So(createdSubnet1.BkBizID, ShouldEqual, constant.SuiteTestBizID)
			So(createdSubnet1.Name, ShouldEqual, createReq.Name)
			So(createdSubnet1.Ipv4Cidr, ShouldEqual, []string{createReq.IPv4Cidr})
			So(createdSubnet1.CloudVpcID, ShouldEqual, vpcID)
			So(createdSubnet1.CloudID, ShouldNotBeEmpty)
			So(createdSubnet1.Memo, ShouldEqual, createReq.Memo)

			Convey("update subnet", func() {
				updateReq := &cloudproto.SubnetUpdateReq{
					Memo: converter.ValToPtr("memo(update subnet)"),
				}
				err = cli.CloudServer().Subnet.UpdateInBiz(kt, constant.SuiteTestBizID, subnetID, updateReq)
				So(err, ShouldBeNil)

				subnetListResult, err := cli.CloudServer().Subnet.ListInBiz(kt, constant.SuiteTestBizID, &subnetListReq)
				So(err, ShouldBeNil)
				So(subnetListResult, ShouldNotBeNil)
				So(subnetListResult.Details, ShouldHaveLength, 1)
				subnet := subnetListResult.Details[0]
				So(subnet.Memo, ShouldEqual, updateReq.Memo)

				Convey("delete subnet", func() {

					err = cli.CloudServer().Subnet.Delete(kt, &cloudproto.BatchDeleteReq{IDs: []string{subnetID}})
					So(err, ShouldBeNil)
					err = cli.HCService().TCloud.RouteTable.SyncRouteTable(kt.Ctx, kt.Header(), &sync.TCloudSyncReq{
						AccountID: TCloudAccountID,
						Region:    constant.SuiteRegion,
					})
					So(err, ShouldBeNil)

					resListResult, err := cli.CloudServer().Subnet.ListInRes(kt, &subnetListReq)
					So(err, ShouldBeNil)
					So(resListResult, ShouldNotBeNil)
					So(resListResult.Details, ShouldHaveLength, 0)
				})
			})
		})

	})
}
