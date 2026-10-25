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

// Package types holds the plugin executor request and response values.
package types

import "hcm/pkg/plugin/tfplugin5"

// ProviderSchema is the adapted provider schema returned to the caller.
type ProviderSchema struct {
	ProtocolVersion   string                     `json:"protocol_version"`
	Plugin            PluginIdentity             `json:"plugin"`
	SchemaHash        string                     `json:"schema_hash"`
	Capabilities      map[string]bool            `json:"capabilities"`
	Provider          *SchemaDocument            `json:"provider"`
	ResourceSchemas   map[string]*SchemaDocument `json:"resource_schemas"`
	DataSourceSchemas map[string]*SchemaDocument `json:"data_source_schemas"`
	Diagnostics       []*tfplugin5.Diagnostic    `json:"diagnostics"`
}

// PluginIdentity is the name and version recorded for the process.
type PluginIdentity struct {
	Name    string `json:"name"`
	Version string `json:"version"`
}

// SchemaDocument is an adapted schema. Version is the protocol schema version.
// An absent block is an empty block.
type SchemaDocument struct {
	Version int64  `json:"version"`
	Block   *Block `json:"block"`
}

// Block is the adapted configuration block.
type Block struct {
	Description string                  `json:"description,omitempty"`
	Attributes  map[string]*Attribute   `json:"attributes"`
	BlockTypes  map[string]*NestedBlock `json:"block_types"`
}

// Attribute is an adapted schema attribute.
// Type is the TF JSON type. A primitive decodes to a string. A compound type decodes to an array.
type Attribute struct {
	Type        any    `json:"type"`
	Required    bool   `json:"required"`
	Optional    bool   `json:"optional"`
	Computed    bool   `json:"computed"`
	Sensitive   bool   `json:"sensitive,omitempty"`
	Description string `json:"description,omitempty"`
}

// NestedBlock is a nested block inside a block.
type NestedBlock struct {
	Nesting  string `json:"nesting"`
	MinItems int64  `json:"min_items"`
	MaxItems int64  `json:"max_items"`
	Block    *Block `json:"block"`
}

// ConfigureResult is the outcome of ConfigureProvider.
type ConfigureResult struct {
	Diagnostics []*tfplugin5.Diagnostic
}

// ValidateResult is the outcome of a resource or data source validation.
type ValidateResult struct {
	Diagnostics []*tfplugin5.Diagnostic
}

// UpgradeResult is the outcome of UpgradeResourceState.
type UpgradeResult struct {
	UpgradedState map[string]any
	Diagnostics   []*tfplugin5.Diagnostic
}

// ReadResourceResult is the outcome of ReadResource.
type ReadResourceResult struct {
	NewState    map[string]any
	Private     []byte
	Diagnostics []*tfplugin5.Diagnostic
}

// PlanResult is the outcome of PlanResourceChange.
type PlanResult struct {
	PlannedState    map[string]any
	RequiresReplace []*tfplugin5.AttributePath
	PlannedPrivate  []byte
	Diagnostics     []*tfplugin5.Diagnostic
}

// ApplyResult is the outcome of ApplyResourceChange.
type ApplyResult struct {
	NewState    map[string]any
	Private     []byte
	Diagnostics []*tfplugin5.Diagnostic
}

// ImportResult is the outcome of ImportResourceState.
type ImportResult struct {
	Imported    []ImportedResource
	Diagnostics []*tfplugin5.Diagnostic
}

// ImportedResource is a resource returned by import.
type ImportedResource struct {
	TypeName string
	State    map[string]any
	Private  []byte
}

// ReadDataSourceResult is the outcome of ReadDataSource.
type ReadDataSourceResult struct {
	State       map[string]any
	Diagnostics []*tfplugin5.Diagnostic
}

// RawState is the not-yet-decoded state passed to UpgradeResourceState.
// JSON and Flatmap are mutually exclusive.
type RawState struct {
	JSON    []byte
	Flatmap map[string]string
}
