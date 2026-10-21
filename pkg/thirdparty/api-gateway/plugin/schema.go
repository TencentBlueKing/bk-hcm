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
	"errors"
	"fmt"
	"strings"

	tfjson "github.com/hashicorp/terraform-json"
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
// It embeds tfjson.ProviderSchema to stay consistent with Terraform provider schema JSON format,
// and supplements HCM southbound protocol metadata fields.
type PluginSchema struct {
	tfjson.ProviderSchema

	// ProtocolVersion is the southbound payload format version.
	ProtocolVersion string `json:"protocol_version"`
	// Plugin contains the plugin name and version.
	Plugin PluginInfo `json:"plugin"`
	// SchemaHash is the fingerprint of the entire schema.
	SchemaHash string `json:"schema_hash"`
	// Capabilities declares the implemented southbound endpoints.
	Capabilities Capabilities `json:"capabilities"`
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
	if err := validateSchema(s.ConfigSchema, "provider"); err != nil {
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
		if err := validateSchema(resSchema, "resource_schemas."+typeName); err != nil {
			return err
		}
	}
	for typeName, dsSchema := range s.DataSourceSchemas {
		if !strings.HasPrefix(typeName, prefix) {
			return fmt.Errorf("data_source type_name %q must start with %q", typeName, prefix)
		}
		if err := validateSchema(dsSchema, "data_source_schemas."+typeName); err != nil {
			return err
		}
	}
	return nil
}

func (s *PluginSchema) validateCapabilities() error {
	if err := s.validateResourceCapabilities(); err != nil {
		return err
	}
	return s.validateDataSourceCapabilities()
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
		if s.Capabilities.Stop && !s.Capabilities.AsyncApply {
			return errors.New("capabilities.async_apply must be true when capabilities.stop is true")
		}
		return nil
	}
	if s.Capabilities.Apply || s.Capabilities.Plan || s.Capabilities.Refresh ||
		s.Capabilities.AsyncApply || s.Capabilities.Stop || s.Capabilities.Import || s.Capabilities.List {
		return errors.New("resource capabilities must not be true when resource_schemas is empty")
	}
	return nil
}

func (s *PluginSchema) validateDataSourceCapabilities() error {
	if len(s.DataSourceSchemas) > 0 && !s.Capabilities.ReadData {
		return errors.New("capabilities.read_data must be true when data_source_schemas is not empty")
	}
	if len(s.DataSourceSchemas) == 0 && s.Capabilities.ReadData {
		return errors.New("capabilities.read_data must not be true when data_source_schemas is empty")
	}
	return nil
}

func validateSchema(s *tfjson.Schema, path string) error {
	if s == nil || s.Block == nil {
		return fmt.Errorf("%s.block is required", path)
	}
	return nil
}
