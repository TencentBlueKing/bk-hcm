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
	"bytes"
	"context"
	"errors"
	"fmt"
	"maps"

	"hcm/cmd/hc-service/types"
	"hcm/pkg/plugin/tfplugin5"
)

// ValidateResourceTypeConfig checks one resource config.
// An error fails this call. It does not block a later plan.
func (e *PluginExecutor) ValidateResourceTypeConfig(ctx context.Context, typeName string, config map[string]any) (
	*types.ValidateResult, error) {

	return callRPC(e, func() (*types.ValidateResult, error) {
		schema, err := e.readyResource(typeName)
		if err != nil {
			return nil, err
		}
		encoded, err := e.encode(schema, config)
		if err != nil {
			return nil, err
		}
		resp, err := e.rpc.ValidateResourceTypeConfig(ctx, &tfplugin5.ValidateResourceTypeConfig_Request{
			TypeName: typeName,
			Config:   encoded,
		})
		if err != nil {
			return nil, err
		}
		diags := resp.GetDiagnostics()
		e.logWarnings(diags)
		return &types.ValidateResult{Diagnostics: diags}, diagnosticFailure(diags)
	})
}

// UpgradeResourceState sends RawState without treating it as a DynamicValue.
// JSON and flatmap are mutually exclusive. The upgraded state is returned to the caller.
func (e *PluginExecutor) UpgradeResourceState(ctx context.Context, typeName string, stateVersion int64,
	raw *types.RawState) (*types.UpgradeResult, error) {

	return callRPC(e, func() (*types.UpgradeResult, error) {
		schema, err := e.readyResource(typeName)
		if err != nil {
			return nil, err
		}
		rawState, err := protocolRawState(raw)
		if err != nil {
			return nil, err
		}
		resp, err := e.rpc.UpgradeResourceState(ctx, &tfplugin5.UpgradeResourceState_Request{
			TypeName: typeName,
			Version:  stateVersion,
			RawState: rawState,
		})
		if err != nil {
			return nil, err
		}
		diags := resp.GetDiagnostics()
		upgraded, decErr := e.decodeObject(schema, resp.GetUpgradedState())
		e.logWarnings(diags)
		result := &types.UpgradeResult{UpgradedState: upgraded, Diagnostics: diags}
		if decErr != nil {
			return result, decErr
		}
		return result, diagnosticFailure(diags)
	})
}

// ReadResource reads one resource. A nil current state stays nil.
func (e *PluginExecutor) ReadResource(ctx context.Context, typeName string, current map[string]any, private []byte) (
	*types.ReadResourceResult, error) {

	return callRPC(e, func() (*types.ReadResourceResult, error) {
		schema, err := e.readyResource(typeName)
		if err != nil {
			return nil, err
		}
		encoded, err := e.encode(schema, current)
		if err != nil {
			return nil, err
		}
		resp, err := e.rpc.ReadResource(ctx, &tfplugin5.ReadResource_Request{
			TypeName:     typeName,
			CurrentState: encoded,
			Private:      bytes.Clone(private),
		})
		if err != nil {
			return nil, err
		}
		diags := resp.GetDiagnostics()
		state, decErr := e.decodeObject(schema, resp.GetNewState())
		e.logWarnings(diags)
		result := &types.ReadResourceResult{NewState: state, Private: bytes.Clone(resp.GetPrivate()),
			Diagnostics: diags}
		if decErr != nil {
			return result, decErr
		}
		return result, diagnosticFailure(diags)
	})
}

// PlanResourceChange sends only the plan RPC.
// Empty prior and proposed states send nothing. A delete plan returns an empty planned state.
func (e *PluginExecutor) PlanResourceChange(ctx context.Context, typeName string,
	prior, proposed, config map[string]any, priorPrivate []byte) (*types.PlanResult, error) {

	return callRPC(e, func() (*types.PlanResult, error) {
		schema, err := e.readyResource(typeName)
		if err != nil {
			return nil, err
		}
		if prior == nil && proposed == nil {
			return nil, errors.New("terraform plugin plan has no prior state and no proposed state")
		}
		encodedPrior, err := e.encode(schema, prior)
		if err != nil {
			return nil, err
		}
		encodedProposed, err := e.encode(schema, proposed)
		if err != nil {
			return nil, err
		}
		encodedConfig, err := e.encode(schema, config)
		if err != nil {
			return nil, err
		}
		resp, err := e.rpc.PlanResourceChange(ctx, &tfplugin5.PlanResourceChange_Request{
			TypeName:         typeName,
			PriorState:       encodedPrior,
			ProposedNewState: encodedProposed,
			Config:           encodedConfig,
			PriorPrivate:     bytes.Clone(priorPrivate),
		})
		if err != nil {
			return nil, err
		}
		diags := resp.GetDiagnostics()
		e.logWarnings(diags)
		result := &types.PlanResult{
			RequiresReplace: resp.GetRequiresReplace(),
			PlannedPrivate:  bytes.Clone(resp.GetPlannedPrivate()),
			Diagnostics:     diags,
		}
		if prior != nil && proposed == nil {
			return result, diagnosticFailure(diags)
		}
		planned, decErr := e.decodeObject(schema, resp.GetPlannedState())
		result.PlannedState = planned
		if decErr != nil {
			return result, decErr
		}
		return result, diagnosticFailure(diags)
	})
}

// ApplyResourceChange applies one plan.
// A nil planned state is a delete and the returned state is empty.
// An apply error still returns the plugin's new state.
func (e *PluginExecutor) ApplyResourceChange(ctx context.Context, typeName string,
	prior, planned, config map[string]any, plannedPrivate []byte) (*types.ApplyResult, error) {

	return callRPC(e, func() (*types.ApplyResult, error) {
		schema, err := e.readyResource(typeName)
		if err != nil {
			return nil, err
		}
		encodedPrior, err := e.encode(schema, prior)
		if err != nil {
			return nil, err
		}
		encodedPlanned, err := e.encode(schema, planned)
		if err != nil {
			return nil, err
		}
		encodedConfig, err := e.encode(schema, config)
		if err != nil {
			return nil, err
		}
		resp, err := e.rpc.ApplyResourceChange(ctx, &tfplugin5.ApplyResourceChange_Request{
			TypeName:       typeName,
			PriorState:     encodedPrior,
			PlannedState:   encodedPlanned,
			Config:         encodedConfig,
			PlannedPrivate: bytes.Clone(plannedPrivate),
		})
		if err != nil {
			return nil, err
		}
		diags := resp.GetDiagnostics()
		e.logWarnings(diags)
		result := &types.ApplyResult{Private: bytes.Clone(resp.GetPrivate()), Diagnostics: diags}
		if planned == nil {
			return result, diagnosticFailure(diags)
		}
		state, decErr := e.decodeObject(schema, resp.GetNewState())
		result.NewState = state
		if decErr != nil {
			return result, decErr
		}
		return result, diagnosticFailure(diags)
	})
}

// ImportResourceState imports an existing resource.
// Each returned item is decoded with its own type schema. This method does not read or apply the resource.
func (e *PluginExecutor) ImportResourceState(ctx context.Context, typeName, id string) (*types.ImportResult, error) {
	return callRPC(e, func() (*types.ImportResult, error) {
		if _, err := e.readyResource(typeName); err != nil {
			return nil, err
		}
		resp, err := e.rpc.ImportResourceState(ctx, &tfplugin5.ImportResourceState_Request{
			TypeName: typeName,
			Id:       id,
		})
		if err != nil {
			return nil, err
		}
		diags := resp.GetDiagnostics()
		imported := make([]types.ImportedResource, 0, len(resp.GetImportedResources()))
		var decErr error
		for _, item := range resp.GetImportedResources() {
			if item == nil {
				continue
			}
			itemSchema, err := e.readyResource(item.GetTypeName())
			if err != nil {
				return &types.ImportResult{Diagnostics: diags}, err
			}
			state, err := e.decodeObject(itemSchema, item.GetState())
			if err != nil && decErr == nil {
				decErr = err
			}
			imported = append(imported, types.ImportedResource{
				TypeName: item.GetTypeName(),
				State:    state,
				Private:  bytes.Clone(item.GetPrivate()),
			})
		}
		e.logWarnings(diags)
		result := &types.ImportResult{Imported: imported, Diagnostics: diags}
		if decErr != nil {
			return result, decErr
		}
		return result, diagnosticFailure(diags)
	})
}

// ValidateDataSourceConfig checks one data source config.
// An error fails this call. It does not block a later read.
func (e *PluginExecutor) ValidateDataSourceConfig(ctx context.Context, typeName string, config map[string]any) (
	*types.ValidateResult, error) {

	return callRPC(e, func() (*types.ValidateResult, error) {
		schema, err := e.readyDataSource(typeName)
		if err != nil {
			return nil, err
		}
		encoded, err := e.encode(schema, config)
		if err != nil {
			return nil, err
		}
		resp, err := e.rpc.ValidateDataSourceConfig(ctx, &tfplugin5.ValidateDataSourceConfig_Request{
			TypeName: typeName,
			Config:   encoded,
		})
		if err != nil {
			return nil, err
		}
		diags := resp.GetDiagnostics()
		e.logWarnings(diags)
		return &types.ValidateResult{Diagnostics: diags}, diagnosticFailure(diags)
	})
}

// ReadDataSource reads one data source. It does not plan or apply.
func (e *PluginExecutor) ReadDataSource(ctx context.Context, typeName string, config map[string]any) (
	*types.ReadDataSourceResult, error) {

	return callRPC(e, func() (*types.ReadDataSourceResult, error) {
		schema, err := e.readyDataSource(typeName)
		if err != nil {
			return nil, err
		}
		encoded, err := e.encode(schema, config)
		if err != nil {
			return nil, err
		}
		resp, err := e.rpc.ReadDataSource(ctx, &tfplugin5.ReadDataSource_Request{
			TypeName: typeName,
			Config:   encoded,
		})
		if err != nil {
			return nil, err
		}
		diags := resp.GetDiagnostics()
		state, decErr := e.decodeObject(schema, resp.GetState())
		e.logWarnings(diags)
		result := &types.ReadDataSourceResult{State: state, Diagnostics: diags}
		if decErr != nil {
			return result, decErr
		}
		return result, diagnosticFailure(diags)
	})
}

func protocolRawState(raw *types.RawState) (*tfplugin5.RawState, error) {
	if raw == nil {
		return nil, errors.New("terraform plugin raw state must set json or flatmap, not both")
	}
	hasJSON := raw.JSON != nil
	hasFlatmap := raw.Flatmap != nil
	if hasJSON == hasFlatmap {
		return nil, errors.New("terraform plugin raw state must set json or flatmap, not both")
	}
	return &tfplugin5.RawState{Json: bytes.Clone(raw.JSON), Flatmap: maps.Clone(raw.Flatmap)}, nil
}

func (e *PluginExecutor) readyResource(typeName string) (*tfplugin5.Schema, error) {
	e.mu.Lock()
	defer e.mu.Unlock()
	if !e.configured {
		return nil, errors.New("terraform plugin is not configured")
	}
	schema, ok := e.resourceSchemas[typeName]
	if !ok {
		return nil, fmt.Errorf("resource type %q is not in the schema", typeName)
	}
	return schema, nil
}

func (e *PluginExecutor) readyDataSource(typeName string) (*tfplugin5.Schema, error) {
	e.mu.Lock()
	defer e.mu.Unlock()
	if !e.configured {
		return nil, errors.New("terraform plugin is not configured")
	}
	schema, ok := e.dataSourceSchemas[typeName]
	if !ok {
		return nil, fmt.Errorf("data source type %q is not in the schema", typeName)
	}
	return schema, nil
}
