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
	"strings"
	"testing"

	"hcm/pkg/dal/dao/tools"
	"hcm/pkg/dal/table/types"
)

func TestIacPluginSchemaReqValidate(t *testing.T) {
	validHash := "sha256:" + strings.Repeat("b", 64)

	createReq := &IacPluginSchemaBatchCreateReq{
		Items: []IacPluginSchemaCreateField{
			{
				PluginName:    "tencentcloud",
				PluginVersion: "1.81.5",
				SchemaHash:    validHash,
				SchemaDetail:  types.JsonField(`{"capabilities":{"apply":true}}`),
			},
		},
	}
	if err := createReq.Validate(); err != nil {
		t.Fatalf("expected valid IacPluginSchemaBatchCreateReq to pass, got: %v", err)
	}

	emptyCreateReq := &IacPluginSchemaBatchCreateReq{}
	if err := emptyCreateReq.Validate(); err == nil {
		t.Fatalf("expected empty items to fail validation")
	}

	invalidHashReq := &IacPluginSchemaBatchCreateReq{
		Items: []IacPluginSchemaCreateField{
			{
				PluginName:    "tencentcloud",
				PluginVersion: "1.81.5",
				SchemaHash:    "bad",
				SchemaDetail:  types.JsonField(`{}`),
			},
		},
	}
	if err := invalidHashReq.Validate(); err == nil {
		t.Fatalf("expected invalid schema_hash to fail validation")
	}

	unsupportedAlgoReq := &IacPluginSchemaBatchCreateReq{
		Items: []IacPluginSchemaCreateField{
			{
				PluginName:    "tencentcloud",
				PluginVersion: "1.81.5",
				SchemaHash:    "md5:" + strings.Repeat("b", 32),
				SchemaDetail:  types.JsonField(`{}`),
			},
		},
	}
	if err := unsupportedAlgoReq.Validate(); err == nil || !strings.Contains(err.Error(), "unsupported") {
		t.Fatalf("expected unsupported algorithm to fail validation, got: %v", err)
	}

	updateReq := &IacPluginSchemaBatchUpdateReq{
		Items: []IacPluginSchemaUpdateField{
			{
				ID:           "schema-001",
				SchemaHash:   validHash,
				SchemaDetail: types.JsonField(`{"capabilities":{"apply":true}}`),
			},
		},
	}
	if err := updateReq.Validate(); err != nil {
		t.Fatalf("expected valid IacPluginSchemaBatchUpdateReq to pass, got: %v", err)
	}

	emptyUpdateReq := &IacPluginSchemaBatchUpdateReq{
		Items: []IacPluginSchemaUpdateField{
			{
				ID: "schema-001",
			},
		},
	}
	if err := emptyUpdateReq.Validate(); err == nil {
		t.Fatalf("expected empty update fields to fail validation")
	}

	deleteReq := &IacPluginSchemaBatchDeleteReq{
		Filter: tools.EqualExpression("id", "schema-001"),
	}
	if err := deleteReq.Validate(); err != nil {
		t.Fatalf("expected valid IacPluginSchemaBatchDeleteReq to pass, got: %v", err)
	}

	badDeleteReq := &IacPluginSchemaBatchDeleteReq{}
	if err := badDeleteReq.Validate(); err == nil {
		t.Fatalf("expected nil filter to fail validation")
	}
}
