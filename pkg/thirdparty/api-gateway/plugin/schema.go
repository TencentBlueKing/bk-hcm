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
	"encoding/json"
	"errors"
	"fmt"
	"strings"
)

// NestingMode defines the nesting mode of a nested block in Schema.
type NestingMode string

const (
	// NestingModeSingle represents a single object nested block.
	NestingModeSingle NestingMode = "single"
	// NestingModeList represents an ordered array nested block.
	NestingModeList NestingMode = "list"
	// NestingModeSet represents an unordered array nested block.
	NestingModeSet NestingMode = "set"
	// NestingModeMap represents a map object nested block.
	NestingModeMap NestingMode = "map"
)

// PluginInfo represents the plugin identity and version in Schema.
type PluginInfo struct {
	// Name is the globally unique plugin name and type_name prefix.
	Name string `json:"name"`
	// Version is the plugin version, matching /health plugin_version.
	Version string `json:"version"`
}

// Capabilities represents the declared capabilities of a southbound plugin.
type Capabilities struct {
	// Apply indicates whether POST /apply is implemented.
	Apply bool `json:"apply,omitempty"`
	// Plan indicates whether POST /plan is implemented.
	Plan bool `json:"plan,omitempty"`
	// Refresh indicates whether POST /refresh is implemented.
	Refresh bool `json:"refresh,omitempty"`
	// ReadData indicates whether POST /read_data is implemented.
	ReadData bool `json:"read_data,omitempty"`
	// AsyncApply indicates whether async apply (status=running + POST /result) is supported.
	AsyncApply bool `json:"async_apply,omitempty"`
	// Stop indicates whether POST /stop is implemented.
	Stop bool `json:"stop,omitempty"`
	// Validate indicates whether POST /validate is implemented.
	Validate bool `json:"validate,omitempty"`
	// Import indicates whether POST /import is implemented.
	Import bool `json:"import,omitempty"`
	// List indicates whether POST /list is implemented.
	List bool `json:"list,omitempty"`
}

// PluginSchema represents the full southbound schema returned by GET /schema.
type PluginSchema struct {
	// ProtocolVersion is the southbound payload format version.
	ProtocolVersion string `json:"protocol_version"`
	// Plugin contains the plugin name and version.
	Plugin PluginInfo `json:"plugin"`
	// SchemaHash is the fingerprint of the entire schema.
	SchemaHash string `json:"schema_hash"`
	// Capabilities declares the implemented southbound endpoints.
	Capabilities Capabilities `json:"capabilities"`
	// Provider defines the schema of launch_config plaintext parameters.
	Provider Schema `json:"provider"`
	// ResourceSchemas defines all managed resource types keyed by type_name.
	ResourceSchemas map[string]Schema `json:"resource_schemas"`
	// DataSourceSchemas defines all managed data source types keyed by type_name.
	DataSourceSchemas map[string]Schema `json:"data_source_schemas"`
}

// Schema represents a provider, resource, or data source schema definition.
type Schema struct {
	// Version is the state migration version number (default 0).
	Version int `json:"version"`
	// Block is the root block definition of the schema.
	Block *Block `json:"block"`
}

// Block represents a Terraform-compatible schema block containing attributes and nested blocks.
type Block struct {
	// Description is the optional description of the block.
	Description string `json:"description,omitempty"`
	// Attributes maps attribute names to their Attribute definitions.
	Attributes map[string]Attribute `json:"attributes"`
	// BlockTypes maps nested block names to their NestedBlock definitions.
	BlockTypes map[string]NestedBlock `json:"block_types"`
}

// Attribute represents a single field definition in a schema block.
type Attribute struct {
	// Type is the Terraform JSON type representation (e.g., "string", ["list", "string"]).
	Type json.RawMessage `json:"type"`
	// Required indicates the attribute must be provided by the user.
	Required bool `json:"required,omitempty"`
	// Optional indicates the attribute may be omitted by the user.
	Optional bool `json:"optional,omitempty"`
	// Computed indicates the attribute value can be computed by the plugin.
	Computed bool `json:"computed,omitempty"`
	// Sensitive indicates the attribute contains sensitive data and should be masked.
	Sensitive bool `json:"sensitive,omitempty"`
	// Description is the optional description of the attribute.
	Description string `json:"description,omitempty"`
}

// NestedBlock represents a nested block definition inside a Block.
type NestedBlock struct {
	// NestingMode specifies how nested instances are encoded: single/list/set/map.
	NestingMode NestingMode `json:"nesting_mode"`
	// MinItems is the optional minimum number of block items.
	MinItems int `json:"min_items,omitempty"`
	// MaxItems is the optional maximum number of block items.
	MaxItems int `json:"max_items,omitempty"`
	// Block is the inner block definition.
	Block *Block `json:"block"`
}

// Validate validates the entire PluginSchema according to the southbound protocol rules.
func (s *PluginSchema) Validate() error {
	if s == nil {
		return errors.New("plugin schema is nil")
	}
	if strings.TrimSpace(s.ProtocolVersion) == "" {
		return errors.New("protocol_version is required")
	}
	if strings.TrimSpace(s.Plugin.Name) == "" {
		return errors.New("plugin.name is required")
	}
	if strings.TrimSpace(s.Plugin.Version) == "" {
		return errors.New("plugin.version is required")
	}
	if strings.TrimSpace(s.SchemaHash) == "" {
		return errors.New("schema_hash is required")
	}
	if err := s.Provider.Validate("provider"); err != nil {
		return err
	}
	if err := s.validateTypeSchemas(); err != nil {
		return err
	}
	return s.validateCapabilities()
}

func (s *PluginSchema) validateTypeSchemas() error {
	prefix := s.Plugin.Name + "_"
	for typeName, resSchema := range s.ResourceSchemas {
		if !strings.HasPrefix(typeName, prefix) {
			return fmt.Errorf("resource type_name %q must start with %q", typeName, prefix)
		}
		if err := resSchema.Validate("resource_schemas." + typeName); err != nil {
			return err
		}
	}
	for typeName, dsSchema := range s.DataSourceSchemas {
		if !strings.HasPrefix(typeName, prefix) {
			return fmt.Errorf("data_source type_name %q must start with %q", typeName, prefix)
		}
		if err := dsSchema.Validate("data_source_schemas." + typeName); err != nil {
			return err
		}
	}
	return nil
}

func (s *PluginSchema) validateCapabilities() error {
	if err := s.validateResourceCapabilities(); err != nil {
		return err
	}
	if len(s.DataSourceSchemas) > 0 && !s.Capabilities.ReadData {
		return errors.New("capabilities.read_data must be true when data_source_schemas is not empty")
	}
	if len(s.DataSourceSchemas) == 0 && s.Capabilities.ReadData {
		return errors.New("capabilities.read_data must not be true when data_source_schemas is empty")
	}
	if s.Capabilities.Stop && !s.Capabilities.AsyncApply {
		return errors.New("capabilities.async_apply must be true when capabilities.stop is true")
	}
	return nil
}

func (s *PluginSchema) validateResourceCapabilities() error {
	if len(s.ResourceSchemas) > 0 {
		if !s.Capabilities.Apply {
			return errors.New("capabilities.apply must be true when resource_schemas is not empty")
		}
		if !s.Capabilities.Plan {
			return errors.New("capabilities.plan must be true when resource_schemas is not empty")
		}
		if !s.Capabilities.Refresh {
			return errors.New("capabilities.refresh must be true when resource_schemas is not empty")
		}
		return nil
	}
	if s.Capabilities.Apply || s.Capabilities.Plan || s.Capabilities.Refresh ||
		s.Capabilities.AsyncApply || s.Capabilities.Stop || s.Capabilities.Import || s.Capabilities.List {
		return errors.New("resource capabilities must not be true when resource_schemas is empty")
	}
	return nil
}

// Validate validates a Schema and its root Block.
func (s Schema) Validate(path string) error {
	if s.Block == nil {
		return fmt.Errorf("%s.block is required", path)
	}
	return s.Block.Validate(path + ".block")
}

// Validate validates a Block, its attributes, and its nested block_types recursively.
func (b *Block) Validate(path string) error {
	if b == nil {
		return fmt.Errorf("%s is nil", path)
	}
	for attrName, attr := range b.Attributes {
		if err := attr.Validate(fmt.Sprintf("%s.attributes.%s", path, attrName)); err != nil {
			return err
		}
	}
	for blockName, nested := range b.BlockTypes {
		if err := nested.Validate(fmt.Sprintf("%s.block_types.%s", path, blockName)); err != nil {
			return err
		}
	}
	return nil
}

// Validate validates an Attribute's type and required/optional/computed modifier combinations.
func (a Attribute) Validate(path string) error {
	if len(a.Type) == 0 {
		return fmt.Errorf("%s.type is required", path)
	}
	if a.Required && (a.Optional || a.Computed) {
		return fmt.Errorf("%s: required cannot be combined with optional or computed", path)
	}
	if !a.Required && !a.Optional && !a.Computed {
		return fmt.Errorf("%s: at least one of required, optional, or computed must be true", path)
	}
	return nil
}

// Validate validates a NestedBlock's nesting_mode, item bounds, and inner Block.
func (nb NestedBlock) Validate(path string) error {
	switch nb.NestingMode {
	case NestingModeSingle, NestingModeList, NestingModeSet, NestingModeMap:
	default:
		return fmt.Errorf("%s: invalid nesting_mode %q", path, nb.NestingMode)
	}
	if nb.MinItems < 0 || nb.MaxItems < 0 {
		return fmt.Errorf("%s: min_items and max_items must be >= 0", path)
	}
	if nb.MaxItems > 0 && nb.MinItems > nb.MaxItems {
		return fmt.Errorf("%s: min_items (%d) cannot be greater than max_items (%d)", path, nb.MinItems, nb.MaxItems)
	}
	if nb.Block == nil {
		return fmt.Errorf("%s.block is required", path)
	}
	return nb.Block.Validate(path + ".block")
}
