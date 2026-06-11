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

package cvm

import (
	"testing"

	"hcm/pkg/criteria/constant"
	"hcm/pkg/criteria/enumor"
	"hcm/pkg/dal/table/cloud/cvm"
	"hcm/pkg/kit"
)

// TestBuildCmdbOperatorsWithoutAccountLookup covers the operator derivation branches that do not
// require an account lookup.
func TestBuildCmdbOperatorsWithoutAccountLookup(t *testing.T) {
	svc := &cvmSvc{}
	models := []*cvm.Table{
		// purchased host newly pushed to cmdb -> operator is the creator
		{ID: "cvm-1", CloudID: "cloud-1", BkBizID: 100, BkHostID: 0, Vendor: enumor.TCloud, Creator: "zhangsan"},
		// host already in cmdb -> operator not set
		{ID: "cvm-2", CloudID: "cloud-2", BkBizID: 100, BkHostID: 88, Vendor: enumor.TCloud, Creator: "lisi"},
		// cloud sync info update without creator -> operator not set
		{ID: "cvm-3", CloudID: "cloud-3", BkBizID: 100, BkHostID: 0, Vendor: enumor.TCloud, Creator: ""},
		// resource pool host -> operator not set
		{ID: "cvm-4", CloudID: "cloud-4", BkBizID: constant.UnassignedBiz, BkHostID: 0, Vendor: enumor.TCloud,
			Creator: "wangwu"},
		// other vendor -> operator not set
		{ID: "cvm-5", CloudID: "cloud-5", BkBizID: 100, BkHostID: 0, Vendor: enumor.Other, Creator: "zhaoliu"},
	}

	operators, err := buildCvmIDOperatorMap(svc, kit.New(), models)
	if err != nil {
		t.Fatalf("build cmdb operators failed, err: %v", err)
	}

	if len(operators) != 1 {
		t.Fatalf("expect 1 operators, got: %d, operators: %+v", len(operators), operators)
	}
	if operators["cvm-1"] == nil || *operators["cvm-1"] != "zhangsan" {
		t.Errorf("expect cvm-1 operator zhangsan, got: %+v", operators["cvm-1"])
	}
	for _, cvmID := range []string{"cvm-2", "cvm-3", "cvm-4", "cvm-5"} {
		if _, ok := operators[cvmID]; ok {
			t.Errorf("%s should not have operator", cvmID)
		}
	}
}
