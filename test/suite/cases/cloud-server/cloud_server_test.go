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

	"hcm/test/suite"

	. "github.com/smartystreets/goconvey/convey"
)

var TCloudAccountID string

func TestCloudServer(t *testing.T) {
	SetDefaultFailureMode(FailureHalts)
	// 清库
	Convey("Prepare Job", t, func() {
		err := suite.ClearData()
		So(err, ShouldBeNil)
	})
	// 腾讯云相关测试
	// 准备测试数据
	TestCreateBaseRes(t)
	// 	Region 相关测试
	TestTCloudRegion(t)
	// 	VPC相关测试
	TestTCloudVPC(t)
	TestSubnet(t)
	TestRouteTable(t)
}

func TestCreateBaseRes(t *testing.T) {
	TestAddAccount(t)

	TestCreateSharedVpc(t)
}
