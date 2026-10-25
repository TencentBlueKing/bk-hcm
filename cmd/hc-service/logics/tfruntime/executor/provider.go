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
	"context"
	"errors"

	"hcm/cmd/hc-service/types"
	"hcm/pkg/criteria/constant"
	"hcm/pkg/plugin/tfplugin5"
	"hcm/pkg/version"
)

// errDiagnostic is returned when a plugin response contains an Error diagnostic.
// The result value is still returned so the caller can read the diagnostics.
var errDiagnostic = errors.New("terraform plugin diagnostics contain an error")

// GetProviderSchema reads the provider schema and adapts it for the caller.
// The request has no business fields. The protocol schemas stay on the executor for later encoding.
func (e *PluginExecutor) GetProviderSchema(ctx context.Context) (*types.ProviderSchema, error) {
	return callRPC(e, func() (*types.ProviderSchema, error) {
		return e.loadSchemas(ctx)
	})
}

// loadSchemas calls GetSchema and stores the generated schemas. It does not enter the close latch.
func (e *PluginExecutor) loadSchemas(ctx context.Context) (*types.ProviderSchema, error) {
	resp, err := e.rpc.GetSchema(ctx, &tfplugin5.GetProviderSchema_Request{})
	if err != nil {
		return nil, err
	}
	providerSchema := resp.GetProvider()
	resourceSchemas := resp.GetResourceSchemas()
	dataSourceSchemas := resp.GetDataSourceSchemas()
	diags := resp.GetDiagnostics()
	provider, err := adaptSchema(providerSchema)
	if err != nil {
		return nil, err
	}
	resources, err := adaptSchemaMap(resourceSchemas)
	if err != nil {
		return nil, err
	}
	dataSources, err := adaptSchemaMap(dataSourceSchemas)
	if err != nil {
		return nil, err
	}
	hash, err := schemaHash(provider, resources, dataSources)
	if err != nil {
		return nil, err
	}
	e.storeSchemas(providerSchema, resourceSchemas, dataSourceSchemas)
	e.logWarnings(diags)
	return &types.ProviderSchema{
		ProtocolVersion:   constant.TFPluginSchemaProtocolVersion,
		Plugin:            types.PluginIdentity{Name: e.name, Version: e.version},
		SchemaHash:        hash,
		Capabilities:      capabilitiesFor(len(resources), len(dataSources)),
		Provider:          provider,
		ResourceSchemas:   resources,
		DataSourceSchemas: dataSources,
		Diagnostics:       diags,
	}, diagnosticFailure(diags)
}

// ConfigureProvider encodes Config, prepares it, then configures the plugin.
// A missing schema cache is loaded here, so this method can be the first call after start.
// A prepare error does not call configure. Credentials stay inside Config.
func (e *PluginExecutor) ConfigureProvider(ctx context.Context, config any) (*types.ConfigureResult, error) {
	return callRPC(e, func() (*types.ConfigureResult, error) {
		if !e.schemasLoaded() {
			if _, err := e.loadSchemas(ctx); err != nil {
				return nil, err
			}
		}
		schema, _ := e.storedProviderSchema()
		encoded, err := e.encode(schema, config)
		if err != nil {
			return nil, err
		}
		prepared, err := e.rpc.PrepareProviderConfig(ctx, &tfplugin5.PrepareProviderConfig_Request{
			Config: encoded,
		})
		if err != nil {
			return nil, err
		}
		diags := prepared.GetDiagnostics()
		e.logWarnings(diags)
		if diagnosticError(diags) {
			return &types.ConfigureResult{Diagnostics: diags}, errDiagnostic
		}
		configured, err := e.rpc.Configure(ctx, &tfplugin5.Configure_Request{
			TerraformVersion: version.VERSION,
			Config:           prepared.GetPreparedConfig(),
		})
		if err != nil {
			return nil, err
		}
		diags = configured.GetDiagnostics()
		e.logWarnings(diags)
		if err = diagnosticFailure(diags); err != nil {
			return &types.ConfigureResult{Diagnostics: diags}, err
		}
		e.markConfigured()
		return &types.ConfigureResult{Diagnostics: diags}, nil
	})
}

func diagnosticFailure(items []*tfplugin5.Diagnostic) error {
	if diagnosticError(items) {
		return errDiagnostic
	}
	return nil
}

func (e *PluginExecutor) storeSchemas(provider *tfplugin5.Schema, resources, dataSources map[string]*tfplugin5.Schema) {
	e.mu.Lock()
	defer e.mu.Unlock()
	e.providerSchema = provider
	if resources == nil {
		resources = map[string]*tfplugin5.Schema{}
	}
	if dataSources == nil {
		dataSources = map[string]*tfplugin5.Schema{}
	}
	e.resourceSchemas = resources
	e.dataSourceSchemas = dataSources
}

func (e *PluginExecutor) schemasLoaded() bool {
	e.mu.Lock()
	defer e.mu.Unlock()
	return e.resourceSchemas != nil
}

func (e *PluginExecutor) storedProviderSchema() (*tfplugin5.Schema, bool) {
	e.mu.Lock()
	defer e.mu.Unlock()
	if e.providerSchema == nil {
		return nil, false
	}
	return e.providerSchema, true
}

func (e *PluginExecutor) markConfigured() {
	e.mu.Lock()
	defer e.mu.Unlock()
	e.configured = true
}
