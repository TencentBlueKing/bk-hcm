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

package plugin

import (
	"testing"

	"hcm/pkg/api/core"
	"hcm/pkg/tools/converter"
	"hcm/pkg/tools/json"
)

func TestPluginInstanceSerialization(t *testing.T) {
	inst := &PluginInstance{
		ID:         "inst-001",
		PluginName: "tencentcloud",
		Name:       "tcloud-prod",
		Properties: map[string]any{
			"secret_id":  "AKID-xxx",
			"secret_key": "SK-xxx",
		},
		Extension: map[string]any{
			"account_id": "acc-001",
		},
		Memo:     converter.ValToPtr("prod instance"),
		TenantID: "default",
		Revision: &core.Revision{
			Creator:   "admin",
			Reviser:   "admin",
			CreatedAt: "2026-10-20T00:00:00Z",
			UpdatedAt: "2026-10-20T00:00:00Z",
		},
	}

	raw, err := json.Marshal(inst)
	if err != nil {
		t.Fatalf("marshal PluginInstance failed: %v", err)
	}

	var topMap map[string]any
	if err := json.Unmarshal(raw, &topMap); err != nil {
		t.Fatalf("unmarshal into map failed: %v", err)
	}

	for _, key := range []string{"id", "plugin_name", "name", "properties", "extension", "memo", "creator", "reviser"} {
		if _, ok := topMap[key]; !ok {
			t.Errorf("expected key %q in serialized PluginInstance, got: %v", key, topMap)
		}
	}
	for _, forbidden := range []string{"attrs", "secret_attrs"} {
		if _, exists := topMap[forbidden]; exists {
			t.Errorf("unexpected internal split key %q in PluginInstance JSON", forbidden)
		}
	}

	var decoded PluginInstance
	if err := json.Unmarshal(raw, &decoded); err != nil {
		t.Fatalf("unmarshal into PluginInstance failed: %v", err)
	}
	if decoded.ID != inst.ID || decoded.PluginName != inst.PluginName || decoded.Name != inst.Name {
		t.Fatalf("round-trip mismatch: got %+v", decoded)
	}
}
