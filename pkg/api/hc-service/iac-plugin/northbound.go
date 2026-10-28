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

package iacplugin

import (
	"encoding/base64"
	"errors"
	"fmt"
	"strings"

	coreplugin "hcm/pkg/api/core/plugin"
	"hcm/pkg/criteria/enumor"
	"hcm/pkg/iacplugin"
	cvt "hcm/pkg/tools/converter"
)

func validateOptionalBase64(fieldName string, val *string) error {
	if val == nil {
		return nil
	}
	raw := cvt.PtrToVal(val)
	if len(raw) == 0 {
		return nil
	}
	if strings.TrimSpace(raw) == "" {
		return fmt.Errorf("%s cannot be whitespace only", fieldName)
	}
	if _, err := base64.StdEncoding.DecodeString(raw); err != nil {
		return fmt.Errorf("%s must be valid standard base64, err: %v", fieldName, err)
	}
	return nil
}

func validateRecompute(recompute []string) error {
	for i, item := range recompute {
		if strings.TrimSpace(item) == "" {
			return fmt.Errorf("recompute[%d] cannot be empty", i)
		}
	}
	return nil
}

// HealthReq defines the query parameters of GET /iac_plugins/health.
type HealthReq struct {
	// Plugin is the globally unique plugin name.
	Plugin string `json:"plugin"`
}

// Validate validates HealthReq.
func (req *HealthReq) Validate() error {
	if req == nil {
		return errors.New("health request is nil")
	}
	return validatePluginName(req.Plugin)
}

// HealthResult is the response data of GET /iac_plugins/health.
type HealthResult = coreplugin.HealthData

// GetSchemaReq defines the query parameters of GET /iac_plugins/schema.
type GetSchemaReq struct {
	// Plugin is the globally unique plugin name.
	Plugin string `json:"plugin"`
	// PluginVersion is the optional semantic version of the plugin.
	PluginVersion string `json:"plugin_version,omitempty"`
}

// Validate validates GetSchemaReq.
func (req *GetSchemaReq) Validate() error {
	if req == nil {
		return errors.New("get schema request is nil")
	}
	if err := validatePluginName(req.Plugin); err != nil {
		return err
	}
	if req.PluginVersion != "" {
		if _, err := iacplugin.ParseVersion(req.PluginVersion); err != nil {
			return err
		}
	}
	return nil
}

// CommonInstanceReq defines the common addressing fields for instance-scoped northbound requests.
// Northbound requests must never contain launch_config or plaintext credentials.
type CommonInstanceReq struct {
	// Plugin is the globally unique plugin name.
	Plugin string `json:"plugin"`
	// PluginInstanceID is the plugin instance ID in iac_plugin_instance.
	PluginInstanceID string `json:"plugin_instance_id"`
	// TypeName is the resource or data source type name.
	TypeName string `json:"type_name"`
}

// Validate validates CommonInstanceReq fields.
func (req *CommonInstanceReq) Validate() error {
	if req == nil {
		return errors.New("common instance request is nil")
	}
	if err := validatePluginName(req.Plugin); err != nil {
		return err
	}
	if strings.TrimSpace(req.PluginInstanceID) == "" {
		return errors.New("plugin_instance_id is required")
	}
	if strings.TrimSpace(req.TypeName) == "" {
		return errors.New("type_name is required")
	}
	return nil
}

// ExecuteReq is the request body of POST /iac_plugins/execute.
type ExecuteReq struct {
	CommonInstanceReq `json:",inline"`
	// PriorState is the saved state before execution; nil when creating.
	PriorState map[string]any `json:"prior_state"`
	// Config is the user desired configuration; nil when deleting.
	Config map[string]any `json:"config"`
	// Recompute lists optional+computed top-level attributes to explicitly set to nil on update.
	Recompute []string `json:"recompute,omitempty"`
	// PriorPrivate is the optional base64-encoded provider private state.
	PriorPrivate *string `json:"prior_private,omitempty"`
	// PriorIdentity is the optional resource identity object.
	PriorIdentity map[string]any `json:"prior_identity,omitempty"`
}

// Validate validates ExecuteReq fields.
func (req *ExecuteReq) Validate() error {
	if req == nil {
		return errors.New("execute request is nil")
	}
	if err := req.CommonInstanceReq.Validate(); err != nil {
		return err
	}
	if req.PriorState == nil && req.Config == nil {
		return errors.New("prior_state and config cannot both be nil")
	}
	if err := validateRecompute(req.Recompute); err != nil {
		return err
	}
	return validateOptionalBase64("prior_private", req.PriorPrivate)
}

// ExecuteResult is the response data of POST /iac_plugins/{execute,result,stop}.
type ExecuteResult struct {
	// Status is the execution status: enumeration values such as: running/succeeded/failed.
	Status enumor.IaCPluginTaskStatus `json:"status"`
	// Message is the task failure reason or status description.
	Message string `json:"message,omitempty"`
	// ApplyID is the async execution identifier when status is running.
	ApplyID string `json:"apply_id,omitempty"`
	// State is the latest resource state.
	State map[string]any `json:"state"`
	// Private is the provider private state.
	Private *string `json:"private"`
	// Identity is the resource identity object.
	Identity map[string]any `json:"identity"`
	// Progress is the optional execution progress summary.
	Progress *coreplugin.Progress `json:"progress,omitempty"`
	// Diagnostics contains optional diagnostic items.
	Diagnostics []coreplugin.Diagnostic `json:"diagnostics,omitempty"`
}

// GetResultReq is the request body of POST /iac_plugins/result.
type GetResultReq struct {
	CommonInstanceReq `json:",inline"`
	// ApplyID is the async execution ID returned by execute.
	ApplyID string `json:"apply_id"`
}

// Validate validates GetResultReq fields.
func (req *GetResultReq) Validate() error {
	if req == nil {
		return errors.New("get result request is nil")
	}
	if err := req.CommonInstanceReq.Validate(); err != nil {
		return err
	}
	if strings.TrimSpace(req.ApplyID) == "" {
		return errors.New("apply_id is required")
	}
	return nil
}

// StopReq is the request body of POST /iac_plugins/stop.
type StopReq struct {
	CommonInstanceReq `json:",inline"`
	// ApplyID is the async execution ID returned by execute.
	ApplyID string `json:"apply_id"`
}

// Validate validates StopReq fields.
func (req *StopReq) Validate() error {
	if req == nil {
		return errors.New("stop request is nil")
	}
	if err := req.CommonInstanceReq.Validate(); err != nil {
		return err
	}
	if strings.TrimSpace(req.ApplyID) == "" {
		return errors.New("apply_id is required")
	}
	return nil
}

// PlanReq is the request body of POST /iac_plugins/plan.
type PlanReq struct {
	CommonInstanceReq `json:",inline"`
	// PriorState is the saved state before plan; nil when creating.
	PriorState map[string]any `json:"prior_state"`
	// Config is the user desired configuration; nil when deleting.
	Config map[string]any `json:"config"`
	// Recompute lists optional+computed top-level attributes to explicitly set to nil on update.
	Recompute []string `json:"recompute,omitempty"`
	// PriorPrivate is the optional base64-encoded provider private state.
	PriorPrivate *string `json:"prior_private,omitempty"`
	// PriorIdentity is the optional resource identity object.
	PriorIdentity map[string]any `json:"prior_identity,omitempty"`
}

// Validate validates PlanReq fields.
func (req *PlanReq) Validate() error {
	if req == nil {
		return errors.New("plan request is nil")
	}
	if err := req.CommonInstanceReq.Validate(); err != nil {
		return err
	}
	if req.PriorState == nil && req.Config == nil {
		return errors.New("prior_state and config cannot both be nil")
	}
	if err := validateRecompute(req.Recompute); err != nil {
		return err
	}
	return validateOptionalBase64("prior_private", req.PriorPrivate)
}

// PlanResult is the response data of POST /iac_plugins/plan.
type PlanResult struct {
	// State is the planned resource state; nil when deleting.
	State map[string]any `json:"state"`
	// ReplacePaths lists field names that require replacement.
	ReplacePaths []string `json:"replace_paths"`
	// Diagnostics contains optional diagnostic items.
	Diagnostics []coreplugin.Diagnostic `json:"diagnostics,omitempty"`
}

const (
	// ConfigTypeResource 表示校验目标为受管资源（Resource）配置。
	ConfigTypeResource = "resource"
	// ConfigTypeDataSource 表示校验目标为数据源（Data Source）查询配置。
	ConfigTypeDataSource = "data"
)

// ValidateReq is the request body of POST /iac_plugins/validate.
type ValidateReq struct {
	CommonInstanceReq `json:",inline"`
	// ConfigType 标识待校验配置的目标类型，枚举值为 "resource" 或 "data"。
	ConfigType string `json:"config_type"`
	// Config is the user configuration object to validate.
	Config map[string]any `json:"config"`
}

// Validate validates ValidateReq fields.
func (req *ValidateReq) Validate() error {
	if req == nil {
		return errors.New("validate request is nil")
	}
	if err := req.CommonInstanceReq.Validate(); err != nil {
		return err
	}
	if req.ConfigType != ConfigTypeResource && req.ConfigType != ConfigTypeDataSource {
		return fmt.Errorf("config_type must be %q or %q", ConfigTypeResource, ConfigTypeDataSource)
	}
	if req.Config == nil {
		return errors.New("config is required")
	}
	return nil
}

// ValidateResult is the response data of POST /iac_plugins/validate.
type ValidateResult struct {
	// Diagnostics contains validation issues found in config.
	Diagnostics []coreplugin.Diagnostic `json:"diagnostics"`
}

// ImportReq is the request body of POST /iac_plugins/import.
type ImportReq struct {
	CommonInstanceReq `json:",inline"`
	// ImportID is the external resource object identifier.
	ImportID string `json:"import_id"`
}

// Validate validates ImportReq fields.
func (req *ImportReq) Validate() error {
	if req == nil {
		return errors.New("import request is nil")
	}
	if err := req.CommonInstanceReq.Validate(); err != nil {
		return err
	}
	if strings.TrimSpace(req.ImportID) == "" {
		return errors.New("import_id is required")
	}
	return nil
}

// RefreshReq is the request body of POST /iac_plugins/refresh.
type RefreshReq struct {
	CommonInstanceReq `json:",inline"`
	// State is the saved resource state used to refresh external state.
	State map[string]any `json:"state"`
	// Private is the optional base64-encoded provider private state.
	Private *string `json:"private,omitempty"`
	// Identity is the optional resource identity object.
	Identity map[string]any `json:"identity,omitempty"`
}

// Validate validates RefreshReq fields.
func (req *RefreshReq) Validate() error {
	if req == nil {
		return errors.New("refresh request is nil")
	}
	if err := req.CommonInstanceReq.Validate(); err != nil {
		return err
	}
	if len(req.State) == 0 {
		return errors.New("state is required")
	}
	return validateOptionalBase64("private", req.Private)
}

// StateResult is the response data of POST /iac_plugins/{import,refresh}.
type StateResult struct {
	// State is the latest resource state; nil when the refreshed object no longer exists.
	State map[string]any `json:"state"`
	// Private is the provider private state.
	Private *string `json:"private"`
	// Identity is the resource identity object.
	Identity map[string]any `json:"identity"`
	// Diagnostics contains optional diagnostic items.
	Diagnostics []coreplugin.Diagnostic `json:"diagnostics,omitempty"`
}

// ReadDataReq is the request body of POST /iac_plugins/read_data.
type ReadDataReq struct {
	CommonInstanceReq `json:",inline"`
	// Config is the query filter configuration for the data source.
	Config map[string]any `json:"config"`
}

// Validate validates ReadDataReq fields.
func (req *ReadDataReq) Validate() error {
	if req == nil {
		return errors.New("read_data request is nil")
	}
	if err := req.CommonInstanceReq.Validate(); err != nil {
		return err
	}
	if req.Config == nil {
		return errors.New("config is required")
	}
	return nil
}

// ReadDataResult is the response data of POST /iac_plugins/read_data.
type ReadDataResult struct {
	// State is the data source query result state.
	State map[string]any `json:"state"`
	// Diagnostics contains optional diagnostic items.
	Diagnostics []coreplugin.Diagnostic `json:"diagnostics,omitempty"`
}

// ListReq is the request body of POST /iac_plugins/list.
type ListReq struct {
	CommonInstanceReq `json:",inline"`
	// Config is the optional filter condition for listing resources.
	Config map[string]any `json:"config,omitempty"`
	// Limit is the maximum number of items per page (1-500).
	Limit int `json:"limit"`
	// PageToken is the pagination cursor (empty string for first page).
	PageToken string `json:"page_token,omitempty"`
}

// Validate validates ListReq fields.
func (req *ListReq) Validate() error {
	if req == nil {
		return errors.New("list request is nil")
	}
	if err := req.CommonInstanceReq.Validate(); err != nil {
		return err
	}
	if req.Limit < coreplugin.MinListLimit || req.Limit > coreplugin.MaxListLimit {
		return fmt.Errorf("limit must be between %d and %d, got %d",
			coreplugin.MinListLimit, coreplugin.MaxListLimit, req.Limit)
	}
	return nil
}

// ListResult is the response data of POST /iac_plugins/list.
type ListResult = coreplugin.ListData
