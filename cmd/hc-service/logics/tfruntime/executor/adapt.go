/*
 * TencentBlueKing is pleased to support the open source community by making
 * 蓝鲸智云 - 混合云管理平台 (BlueKing - Hybrid Cloud Management System) available.
 * Copyright (C) 2022 THL A29 Limited,
 * a Tencent company. All rights reserved.
 * Licensed under the MIT License (the "License");
 * you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at http://opensource.org/licenses/MIT
 * Unless required by applicable law or agreed to in writing,
 * software distributed under the License is distributed on an "AS IS" BASIS,
 * WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
 * See the License for the specific language governing permissions and
 * limitations under the License.
 *
 * We undertake not to change the open source license (MIT license) applicable
 *
 * to the current version of the project delivered to anyone in the future.
 */

package executor

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"

	"hcm/cmd/hc-service/types"
	"hcm/pkg/logs"
	"hcm/pkg/plugin/tfplugin5"
)

func adaptSchema(in *tfplugin5.Schema) (*types.SchemaDocument, error) {
	if in == nil || in.GetBlock() == nil {
		return &types.SchemaDocument{Block: &types.Block{
			Attributes: map[string]*types.Attribute{},
			BlockTypes: map[string]*types.NestedBlock{},
		}}, nil
	}
	block, err := adaptBlock(in.GetBlock())
	if err != nil {
		return nil, err
	}
	return &types.SchemaDocument{Version: in.GetVersion(), Block: block}, nil
}

func adaptBlock(in *tfplugin5.Schema_Block) (*types.Block, error) {
	if in == nil {
		return &types.Block{
			Attributes: map[string]*types.Attribute{},
			BlockTypes: map[string]*types.NestedBlock{},
		}, nil
	}
	out := &types.Block{
		Attributes:  map[string]*types.Attribute{},
		BlockTypes:  map[string]*types.NestedBlock{},
		Description: in.GetDescription(),
	}
	for _, attr := range in.GetAttributes() {
		if attr == nil {
			continue
		}
		converted, err := adaptAttribute(attr)
		if err != nil {
			return nil, err
		}
		out.Attributes[attr.GetName()] = converted
	}
	for _, nested := range in.GetBlockTypes() {
		if nested == nil {
			continue
		}
		converted, err := adaptNestedBlock(nested)
		if err != nil {
			return nil, err
		}
		out.BlockTypes[nested.GetTypeName()] = converted
	}
	return out, nil
}

func adaptAttribute(in *tfplugin5.Schema_Attribute) (*types.Attribute, error) {
	var typ any
	if raw := in.GetType(); len(raw) > 0 {
		if err := json.Unmarshal(raw, &typ); err != nil {
			return nil, fmt.Errorf("attribute %s: %w", in.GetName(), err)
		}
	}
	return &types.Attribute{
		Type:        typ,
		Required:    in.GetRequired(),
		Optional:    in.GetOptional(),
		Computed:    in.GetComputed(),
		Sensitive:   in.GetSensitive(),
		Description: in.GetDescription(),
	}, nil
}

func adaptNestedBlock(in *tfplugin5.Schema_NestedBlock) (*types.NestedBlock, error) {
	block, err := adaptBlock(in.GetBlock())
	if err != nil {
		return nil, err
	}
	return &types.NestedBlock{
		Nesting:  nestingName(in.GetNesting()),
		MinItems: in.GetMinItems(),
		MaxItems: in.GetMaxItems(),
		Block:    block,
	}, nil
}

func nestingName(mode tfplugin5.Schema_NestedBlock_NestingMode) string {
	switch mode {
	case tfplugin5.Schema_NestedBlock_SINGLE:
		return "single"
	case tfplugin5.Schema_NestedBlock_LIST:
		return "list"
	case tfplugin5.Schema_NestedBlock_SET:
		return "set"
	case tfplugin5.Schema_NestedBlock_MAP:
		return "map"
	case tfplugin5.Schema_NestedBlock_GROUP:
		return "group"
	default:
		return "invalid"
	}
}

func adaptSchemaMap(in map[string]*tfplugin5.Schema) (map[string]*types.SchemaDocument, error) {
	if len(in) == 0 {
		return map[string]*types.SchemaDocument{}, nil
	}
	out := make(map[string]*types.SchemaDocument, len(in))
	for name, schema := range in {
		converted, err := adaptSchema(schema)
		if err != nil {
			return nil, fmt.Errorf("schema %s: %w", name, err)
		}
		out[name] = converted
	}
	return out, nil
}

func capabilitiesFor(resourceCount, dataSourceCount int) map[string]bool {
	out := map[string]bool{}
	if resourceCount > 0 {
		out["apply"] = true
		out["plan"] = true
		out["validate"] = true
		out["import"] = true
		out["refresh"] = true
	}
	if dataSourceCount > 0 {
		out["read_data"] = true
	}
	return out
}

// schemaHash is the hex SHA-256 of the adapted provider, resource, and data source schemas.
// Map keys are sorted by encoding/json. The hash, diagnostics, plugin identity, and capabilities are excluded.
func schemaHash(provider *types.SchemaDocument, resources, dataSources map[string]*types.SchemaDocument) (string, error) {
	if resources == nil {
		resources = map[string]*types.SchemaDocument{}
	}
	if dataSources == nil {
		dataSources = map[string]*types.SchemaDocument{}
	}
	raw, err := json.Marshal(map[string]any{
		"data_source_schemas": dataSources,
		"provider":            provider,
		"resource_schemas":    resources,
	})
	if err != nil {
		return "", err
	}
	sum := sha256.Sum256(raw)
	return hex.EncodeToString(sum[:]), nil
}

func diagnosticError(items []*tfplugin5.Diagnostic) bool {
	for _, item := range items {
		if item.GetSeverity() == tfplugin5.Diagnostic_ERROR {
			return true
		}
	}
	return false
}

func (e *PluginExecutor) logWarnings(items []*tfplugin5.Diagnostic) {
	for _, item := range items {
		if item.GetSeverity() != tfplugin5.Diagnostic_WARNING {
			continue
		}
		logs.Warnf("plugin diagnostic: plugin=%s summary=%s", e.name, item.GetSummary())
	}
}
