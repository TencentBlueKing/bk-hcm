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
	"crypto/sha256"
	"encoding/hex"
	"regexp"

	"hcm/pkg/criteria/enumor"
	"hcm/pkg/criteria/errf"
	"hcm/pkg/tools/json"
)

const (
	// SchemaHashPrefix 合法 schema_hash 前缀。
	SchemaHashPrefix = "sha256:"
	// SchemaHashLength 合法 schema_hash 总长度（"sha256:" + 64 位 hex）。
	SchemaHashLength = 71
)

var schemaHashRegexp = regexp.MustCompile(`^sha256:[0-9a-f]{64}$`)

var validCapabilities = map[string]struct{}{
	CapabilityApply:      {},
	CapabilityPlan:       {},
	CapabilityRefresh:    {},
	CapabilityReadData:   {},
	CapabilityAsyncApply: {},
	CapabilityStop:       {},
	CapabilityValidate:   {},
	CapabilityImport:     {},
	CapabilityList:       {},
}

// ComputeRawSchemaHash 计算 RawSchema 的规范 SHA-256（"sha256:" + 64 位小写 hex）。
func ComputeRawSchemaHash(raw *RawSchema) (string, error) {
	if raw == nil {
		return "", errf.New(errf.InvalidParameter, "raw schema is nil")
	}

	rawBytes, err := json.Marshal(raw)
	if err != nil {
		return "", errf.Newf(errf.InvalidParameter, "marshal raw schema failed, err: %v", err)
	}

	sum := sha256.Sum256(rawBytes)
	return SchemaHashPrefix + hex.EncodeToString(sum[:]), nil
}

// ValidateSchemaHashFormat 校验 schemaHash 是否符合 "sha256:" + 64 位小写 hex。
func ValidateSchemaHashFormat(schemaHash string) error {
	if len(schemaHash) != SchemaHashLength || !schemaHashRegexp.MatchString(schemaHash) {
		return errf.Newf(errf.InvalidParameter,
			"invalid schema_hash format: %q, expected sha256:<64 lowercase hex chars>", schemaHash)
	}
	return nil
}

// ValidateCapabilities 按南向协议能力位规则校验 RawSchema.Capabilities。
func ValidateCapabilities(raw *RawSchema) error {
	if raw == nil {
		return errf.New(errf.InvalidParameter, "raw schema is nil")
	}

	if len(raw.Capabilities) == 0 {
		return errf.New(errf.InvalidParameter, "capabilities cannot be empty")
	}

	for key := range raw.Capabilities {
		if _, ok := validCapabilities[key]; !ok {
			return errf.Newf(errf.InvalidParameter, "unsupported capability key: %s", key)
		}
	}

	if len(raw.ResourceSchemas) > 0 {
		requiredCaps := []string{CapabilityApply, CapabilityPlan, CapabilityRefresh}
		for _, capKey := range requiredCaps {
			if !raw.Capabilities[capKey] {
				return errf.Newf(errf.InvalidParameter,
					"capability %q must be true when resource_schemas is non-empty", capKey)
			}
		}
	}

	if len(raw.DataSourceSchemas) > 0 {
		if !raw.Capabilities[CapabilityReadData] {
			return errf.New(errf.InvalidParameter,
				"capability \"read_data\" must be true when data_source_schemas is non-empty")
		}
	} else {
		if _, declared := raw.Capabilities[CapabilityReadData]; declared {
			return errf.New(errf.InvalidParameter,
				"capability \"read_data\" must not be declared when data_source_schemas is empty")
		}
	}

	if raw.Capabilities[CapabilityStop] && !raw.Capabilities[CapabilityAsyncApply] {
		return errf.New(errf.InvalidParameter,
			"capability \"async_apply\" must be true when capability \"stop\" is true")
	}

	return nil
}

// ValidatePluginSchema 按期望插件身份与协议规则校验 PluginSchema 全部字段。
func ValidatePluginSchema(schema *PluginSchema, expectedPluginName, expectedPluginVersion string,
	protocol enumor.IaCPluginProtocol) error {

	if schema == nil {
		return errf.New(errf.InvalidParameter, "plugin schema is nil")
	}

	if len(schema.PluginName) == 0 {
		return errf.New(errf.InvalidParameter, "plugin_name is required")
	}
	if len(schema.PluginVersion) == 0 {
		return errf.New(errf.InvalidParameter, "plugin_version is required")
	}

	if schema.PluginName != expectedPluginName {
		return errf.Newf(errf.InvalidParameter,
			"plugin_name mismatch: schema has %q, expected %q", schema.PluginName, expectedPluginName)
	}
	if schema.PluginVersion != expectedPluginVersion {
		return errf.Newf(errf.InvalidParameter,
			"plugin_version mismatch: schema has %q, expected %q", schema.PluginVersion, expectedPluginVersion)
	}

	switch protocol {
	case enumor.IaCPluginProtocolHTTP:
		if len(schema.ProtocolVersion) == 0 {
			return errf.New(errf.InvalidParameter, "protocol_version is required for http protocol plugin")
		}
	case enumor.IaCPluginProtocolTFProvider:
		// protocol_version is optional for tf_provider protocol
	default:
		return errf.Newf(errf.InvalidParameter, "unsupported plugin protocol: %s", protocol)
	}

	if err := ValidateSchemaHashFormat(schema.SchemaHash); err != nil {
		return err
	}

	computedHash, err := ComputeRawSchemaHash(&schema.RawSchema)
	if err != nil {
		return err
	}
	if schema.SchemaHash != computedHash {
		return errf.Newf(errf.InvalidParameter,
			"schema_hash mismatch: declared %s, computed %s", schema.SchemaHash, computedHash)
	}

	if err := ValidateCapabilities(&schema.RawSchema); err != nil {
		return err
	}

	if schema.Provider == nil || schema.Provider.Block == nil {
		return errf.New(errf.InvalidParameter, "provider and provider.block are required")
	}

	return nil
}
