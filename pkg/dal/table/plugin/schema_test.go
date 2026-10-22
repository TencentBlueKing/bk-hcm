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

package tableplugin

import (
	"strings"
	"testing"

	"hcm/pkg/dal/table"
	"hcm/pkg/dal/table/types"
)

func TestIacPluginSchemaTableValidate(t *testing.T) {
	validHash := "sha256:" + strings.Repeat("a", 64)
	row := IacPluginSchemaTable{
		PluginName:    "tencentcloud",
		PluginVersion: "1.81.5",
		SchemaHash:    validHash,
		SchemaDetail:  types.JsonField(`{"capabilities":{"apply":true}}`),
		Creator:       "admin",
		Reviser:       "admin",
	}

	if err := row.InsertValidate(); err != nil {
		t.Fatalf("expected valid row to pass InsertValidate, got: %v", err)
	}
	if err := row.TableName().Validate(); err != nil {
		t.Fatalf("expected table name to be registered, got: %v", err)
	}
	if table.TableMap[row.TableName()].EnableTenant {
		t.Fatalf("expected iac_plugin_schema EnableTenant to be false")
	}
	if row.Columns() == nil || len(row.ColumnDescriptors()) == 0 {
		t.Fatalf("expected non-empty column descriptors")
	}
	if !strings.Contains(IacPluginSchemaTableColumns.ColumnExpr(), "schema_detail") {
		t.Fatalf("expected ColumnExpr to contain schema_detail, got: %s",
			IacPluginSchemaTableColumns.ColumnExpr())
	}
	if !strings.Contains(IacPluginSchemaTableColumns.ColonNameExpr(), ":schema_detail") {
		t.Fatalf("expected ColonNameExpr to contain :schema_detail, got: %s",
			IacPluginSchemaTableColumns.ColonNameExpr())
	}
	if IacPluginSchemaTableColumns.FieldsNamedExpr([]string{"id", "schema_detail"}) != "id,schema_detail" {
		t.Fatalf("expected FieldsNamedExpr([id, schema_detail]) to be id,schema_detail, got: %s",
			IacPluginSchemaTableColumns.FieldsNamedExpr([]string{"id", "schema_detail"}))
	}

	// ID set should fail InsertValidate
	withID := row
	withID.ID = "1"
	if err := withID.InsertValidate(); err == nil {
		t.Fatalf("expected InsertValidate to fail when ID is set")
	}

	// Invalid hash should fail InsertValidate
	badHash := row
	badHash.SchemaHash = "invalid-hash"
	if err := badHash.InsertValidate(); err == nil {
		t.Fatalf("expected InsertValidate to fail with invalid schema_hash")
	}

	// UpdateValidate tests
	updateRow := IacPluginSchemaTable{
		SchemaHash: validHash,
		Reviser:    "admin",
	}
	if err := updateRow.UpdateValidate(); err != nil {
		t.Fatalf("expected valid updateRow to pass UpdateValidate, got: %v", err)
	}

	badUpdateHash := IacPluginSchemaTable{
		SchemaHash: "invalid",
		Reviser:    "admin",
	}
	if err := badUpdateHash.UpdateValidate(); err == nil {
		t.Fatalf("expected UpdateValidate to fail when schema_hash is invalid")
	}

	badUpdate := IacPluginSchemaTable{
		PluginName: "tencentcloud",
		Reviser:    "admin",
	}
	if err := badUpdate.UpdateValidate(); err == nil {
		t.Fatalf("expected UpdateValidate to fail when plugin_name is set")
	}
}
