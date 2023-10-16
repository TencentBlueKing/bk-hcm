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
	"hcm/pkg/api/core"
	"hcm/pkg/criteria/constant"
	"hcm/pkg/dal/dao/tools"
	"hcm/test/suite"
	"hcm/test/suite/cases"

	. "github.com/smartystreets/goconvey/convey"
)

func TestRouteTable(t *testing.T) {
	Convey("test route table", t, func() {

		// 目前无法单独创建路由表， 使用前面创建好的共享路由表
		cli := suite.GetClientSet()
		kt := cases.GenApiKit()

		So(ResVpcCloudID, ShouldNotBeEmpty)

		listReq := &core.ListReq{
			Filter: tools.EqualExpression("cloud_vpc_id", ResVpcCloudID),
			Page:   &core.BasePage{Limit: 2},
		}

		rt, err := cli.CloudServer().RouteTable.ListInRes(kt, listReq)
		So(err, ShouldBeNil)
		So(rt.Details, ShouldHaveLength, 1)
		routeTable := rt.Details[0]

		Convey("assign route table to business", func() {

			rtAssign := &cloudserver.AssignRouteTableToBizReq{
				RouteTableIDs: []string{routeTable.ID},
				BkBizID:       constant.SuiteTestBizID,
			}

			err = cli.CloudServer().RouteTable.Assign(kt, rtAssign)
			So(err, ShouldBeNil)
			routeTableInBiz, err := cli.CloudServer().RouteTable.ListInBiz(kt, constant.SuiteTestBizID, listReq)
			So(err, ShouldBeNil)
			So(routeTableInBiz, ShouldNotBeNil)
			So(routeTableInBiz.Details, ShouldHaveLength, 1)
		})

	})

}
