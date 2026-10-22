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
	"testing"

	"hcm/pkg/dal/table"
	"hcm/pkg/dal/table/types"
	"hcm/pkg/tools/converter"
)

func TestIacPluginInstanceTableValidate(t *testing.T) {
	row := IacPluginInstanceTable{
		PluginName:  "tencentcloud",
		Name:        "tcloud-prod",
		Attrs:       types.JsonField(`{"secret_id":"AKID-xxx"}`),
		SecretAttrs: types.JsonField(`{"secret_key":"cipher-xxx"}`),
		Extension:   types.JsonField(`{"account_id":"acc-001"}`),
		Memo:        converter.ValToPtr("prod instance"),
		Creator:     "admin",
		Reviser:     "admin",
	}

	if err := row.InsertValidate(); err != nil {
		t.Fatalf("expected valid row to pass InsertValidate, got: %v", err)
	}
	if err := row.TableName().Validate(); err != nil {
		t.Fatalf("expected table name to be registered, got: %v", err)
	}
	if !table.TableMap[row.TableName()].EnableTenant {
		t.Fatalf("expected iac_plugin_instance EnableTenant to be true")
	}
	if row.Columns() == nil || len(row.ColumnDescriptors()) == 0 {
		t.Fatalf("expected non-empty column descriptors")
	}

	// ID set should fail InsertValidate
	withID := row
	withID.ID = "inst-001"
	if err := withID.InsertValidate(); err == nil {
		t.Fatalf("expected InsertValidate to fail when ID is set")
	}

	// Missing plugin_name should fail InsertValidate
	missingPlugin := row
	missingPlugin.PluginName = ""
	if err := missingPlugin.InsertValidate(); err == nil {
		t.Fatalf("expected InsertValidate to fail when plugin_name is empty")
	}

	// Missing name should fail InsertValidate
	missingName := row
	missingName.Name = ""
	if err := missingName.InsertValidate(); err == nil {
		t.Fatalf("expected InsertValidate to fail when name is empty")
	}

	// UpdateValidate valid
	updateRow := IacPluginInstanceTable{
		Name:    "tcloud-prod-v2",
		Reviser: "admin",
	}
	if err := updateRow.UpdateValidate(); err != nil {
		t.Fatalf("expected valid updateRow to pass UpdateValidate, got: %v", err)
	}

	// Updating id, plugin_name, or creator should fail UpdateValidate
	for _, tc := range []struct {
		name string
		row  IacPluginInstanceTable
	}{
		{
			name: "id set",
			row:  IacPluginInstanceTable{ID: "inst-001", Reviser: "admin"},
		},
		{
			name: "plugin_name set",
			row:  IacPluginInstanceTable{PluginName: "aws", Reviser: "admin"},
		},
		{
			name: "creator set",
			row:  IacPluginInstanceTable{Creator: "user1", Reviser: "admin"},
		},
		{
			name: "reviser missing",
			row:  IacPluginInstanceTable{Name: "new-name"},
		},
	} {
		if err := tc.row.UpdateValidate(); err == nil {
			t.Fatalf("expected UpdateValidate to fail for %s", tc.name)
		}
	}
}
