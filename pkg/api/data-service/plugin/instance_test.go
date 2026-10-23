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

package dataplugin

import (
	"testing"

	"hcm/pkg/dal/dao/tools"
	"hcm/pkg/tools/converter"
)

func TestIacPluginInstanceReqValidate(t *testing.T) {
	createReq := &IacPluginInstanceBatchCreateReq{
		Items: []IacPluginInstanceCreateField{
			{
				PluginName:  "tencentcloud",
				Name:        "tcloud-prod",
				Attrs:       map[string]any{"secret_id": "AKID-xxx"},
				SecretAttrs: map[string]any{"secret_key": "SK-xxx"},
				Extension:   map[string]any{"account_id": "acc-001"},
				Memo:        converter.ValToPtr("memo"),
			},
		},
	}
	if err := createReq.Validate(); err != nil {
		t.Fatalf("expected valid IacPluginInstanceBatchCreateReq to pass, got: %v", err)
	}

	emptyCreateReq := &IacPluginInstanceBatchCreateReq{}
	if err := emptyCreateReq.Validate(); err == nil {
		t.Fatalf("expected empty items to fail validation")
	}

	badCreateReq := &IacPluginInstanceBatchCreateReq{
		Items: []IacPluginInstanceCreateField{
			{
				Name: "tcloud-prod",
			},
		},
	}
	if err := badCreateReq.Validate(); err == nil {
		t.Fatalf("expected missing plugin_name to fail validation")
	}

	updateReq := &IacPluginInstanceBatchUpdateReq{
		Items: []IacPluginInstanceUpdateField{
			{
				ID:   "inst-001",
				Name: "tcloud-prod-v2",
			},
		},
	}
	if err := updateReq.Validate(); err != nil {
		t.Fatalf("expected valid IacPluginInstanceBatchUpdateReq to pass, got: %v", err)
	}

	emptyUpdateReq := &IacPluginInstanceBatchUpdateReq{
		Items: []IacPluginInstanceUpdateField{
			{
				ID: "inst-001",
			},
		},
	}
	if err := emptyUpdateReq.Validate(); err == nil {
		t.Fatalf("expected empty update fields to fail validation")
	}

	deleteReq := &IacPluginInstanceBatchDeleteReq{
		Filter: tools.EqualExpression("id", "inst-001"),
	}
	if err := deleteReq.Validate(); err != nil {
		t.Fatalf("expected valid IacPluginInstanceBatchDeleteReq to pass, got: %v", err)
	}

	badDeleteReq := &IacPluginInstanceBatchDeleteReq{}
	if err := badDeleteReq.Validate(); err == nil {
		t.Fatalf("expected nil filter to fail validation")
	}
}
