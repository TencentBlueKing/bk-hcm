/*
 * TencentBlueKing is pleased to support the open source community by making
 * 蓝鲸智云 - 混合云管理平台 (BlueKing - Hybrid Cloud Management System) available.
 * Copyright (C) 2024 THL A29 Limited,
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

package sync_test

import (
	"testing"

	"hcm/cmd/cloud-server/service/sync/aws"
	"hcm/cmd/cloud-server/service/sync/azure"
	"hcm/cmd/cloud-server/service/sync/huawei"
	"hcm/cmd/cloud-server/service/sync/tcloud"
	"hcm/pkg/criteria/enumor"

	"github.com/stretchr/testify/assert"
)

func TestGetCondSyncFunc_Subnet(t *testing.T) {
	vendors := []struct {
		name string
		fn   func(enumor.CloudResourceType) (any, bool)
	}{
		{name: string(enumor.TCloud), fn: func(res enumor.CloudResourceType) (any, bool) {
			f, ok := tcloud.GetCondSyncFunc(res)
			return f, ok
		}},
		{name: string(enumor.HuaWei), fn: func(res enumor.CloudResourceType) (any, bool) {
			f, ok := huawei.GetCondSyncFunc(res)
			return f, ok
		}},
		{name: string(enumor.Aws), fn: func(res enumor.CloudResourceType) (any, bool) {
			f, ok := aws.GetCondSyncFunc(res)
			return f, ok
		}},
		{name: string(enumor.Azure), fn: func(res enumor.CloudResourceType) (any, bool) {
			f, ok := azure.GetCondSyncFunc(res)
			return f, ok
		}},
	}

	for _, vendor := range vendors {
		t.Run(vendor.name, func(t *testing.T) {
			syncFunc, ok := vendor.fn(enumor.SubnetCloudResType)
			assert.True(t, ok)
			assert.NotNil(t, syncFunc)
		})
	}
}
