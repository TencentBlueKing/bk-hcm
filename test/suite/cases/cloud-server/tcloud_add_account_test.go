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
	"strconv"
	"testing"

	protocloud "hcm/pkg/api/data-service/cloud"
	"hcm/pkg/criteria/constant"
	"hcm/pkg/criteria/enumor"
	cvt "hcm/pkg/tools/converter"
	"hcm/pkg/tools/rand"
	"hcm/test/suite"
	"hcm/test/suite/cases"

	. "github.com/smartystreets/goconvey/convey"
)

// TestAddAccount 添加测试账号到数据库
func TestAddAccount(t *testing.T) {
	cli := suite.GetClientSet()
	tcloudExt := &protocloud.TCloudAccountExtensionCreateReq{
		CloudMainAccountID: strconv.Itoa(rand.RandomRange([2]int{100000000000, 110000000000})),
		CloudSubAccountID:  strconv.Itoa(rand.RandomRange([2]int{100000000000, 110000000000})),
		CloudSecretID:      rand.Prefix("AKID", 32),
		CloudSecretKey:     rand.Prefix("SEC", 61),
	}
	Convey("prepare TCloud account test ", t, func() {
		kt := cases.GenApiKit()
		addAccountReq := protocloud.AccountCreateReq[protocloud.TCloudAccountExtensionCreateReq]{
			Name:      "tcloud-test-account",
			Managers:  []string{constant.SuiteTestUserKey},
			Type:      enumor.ResourceAccount,
			Site:      enumor.ChinaSite,
			Memo:      cvt.ValToPtr("suite test"),
			BkBizIDs:  []int64{constant.SuiteTestBizID},
			Extension: tcloudExt,
		}
		// TODO: test cloud server add account api
		ret, err := cli.DataService().TCloud.Account.Create(kt.Ctx, kt.Header(), &addAccountReq)
		So(err, ShouldBeNil)
		So(ret.ID, ShouldNotBeEmpty)
		TCloudAccountID = ret.ID

	})

}
