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

	cvt "hcm/pkg/tools/converter"

	tfjson "github.com/hashicorp/terraform-json"
)

const (
	// MinListLimit is the minimum allowed limit for POST /list.
	MinListLimit = 1
	// MaxListLimit is the maximum allowed limit for POST /list.
	MaxListLimit = 500
)

// TaskStatus represents the execution status of Apply, GetResult, and Stop.
type TaskStatus string

const (
	// TaskStatusRunning indicates the async apply task is still running.
	TaskStatusRunning TaskStatus = "running"
	// TaskStatusSucceeded indicates the apply task finished normally.
	TaskStatusSucceeded TaskStatus = "succeeded"
	// TaskStatusFailed indicates the apply task failed.
	TaskStatusFailed TaskStatus = "failed"
)

// PathStep represents a single step in a Diagnostic attribute path.
// Exactly one of Attribute, Key, or Index must be set.
type PathStep struct {
	// Attribute selects an object attribute or nested block by schema field name.
	Attribute *string `json:"attribute,omitempty"`
	// Key selects an element of a map by key string.
	Key *string `json:"key,omitempty"`
	// Index selects an element of a list by 0-based index.
	Index *int `json:"index,omitempty"`
}

// NewAttributeStep creates a PathStep for an attribute or nested block name.
func NewAttributeStep(attr string) PathStep {
	return PathStep{Attribute: cvt.ValToPtr(attr)}
}

// NewKeyStep creates a PathStep for a map key.
func NewKeyStep(key string) PathStep {
	return PathStep{Key: cvt.ValToPtr(key)}
}

// NewIndexStep creates a PathStep for a 0-based list index.
func NewIndexStep(index int) PathStep {
	return PathStep{Index: cvt.ValToPtr(index)}
}

// Validate checks that exactly one of Attribute, Key, or Index is set and valid.
func (s PathStep) Validate() error {
	count := 0
	if s.Attribute != nil {
		count++
	}
	if s.Key != nil {
		count++
	}
	if s.Index != nil {
		count++
	}
	if count != 1 {
		return errors.New("path step must have exactly one of attribute, key, or index")
	}
	if s.Attribute != nil && strings.TrimSpace(cvt.PtrToVal(s.Attribute)) == "" {
		return errors.New("path step attribute cannot be empty")
	}
	if s.Index != nil && cvt.PtrToVal(s.Index) < 0 {
		return errors.New("path step index must be >= 0")
	}
	return nil
}

// Diagnostic describes a single validation issue in user configuration.
// It embeds tfjson.Diagnostic to stay consistent with Terraform diagnostic JSON format,
// and supplements southbound protocol fields Path and Message.
type Diagnostic struct {
	tfjson.Diagnostic

	// Path is the ordered path from the root of config; empty array means not tied to a single field.
	Path []PathStep `json:"path"`
	// Message is the human-readable issue description without repeating the field name.
	Message string `json:"message"`
}

// EffectiveSeverity returns the diagnostic severity, defaulting to "error" when empty.
func (d Diagnostic) EffectiveSeverity() tfjson.DiagnosticSeverity {
	if d.Severity == "" {
		return tfjson.DiagnosticSeverityError
	}
	return d.Severity
}

// Validate validates the Diagnostic structure and all its PathSteps.
func (d Diagnostic) Validate() error {
	if strings.TrimSpace(d.Message) == "" {
		return errors.New("diagnostic message is required")
	}
	sev := d.EffectiveSeverity()
	if sev != tfjson.DiagnosticSeverityError && sev != tfjson.DiagnosticSeverityWarning {
		return fmt.Errorf("invalid diagnostic severity %q", d.Severity)
	}
	for i, step := range d.Path {
		if err := step.Validate(); err != nil {
			return fmt.Errorf("diagnostic path[%d] invalid: %v", i, err)
		}
	}
	return nil
}

// HealthData is the response data of GET /health.
type HealthData struct {
	// PluginVersion is the current version of the plugin.
	PluginVersion string `json:"plugin_version"`
}

// Progress represents the optional execution progress in Apply, GetResult, and Stop responses.
type Progress struct {
	// Total is the optional total item/stage count.
	Total *int `json:"total,omitempty"`
	// Succeeded is the optional succeeded item/stage count.
	Succeeded *int `json:"succeeded,omitempty"`
	// Failed is the optional failed item/stage count.
	Failed *int `json:"failed,omitempty"`
	// Description is the optional display message for progress.
	Description string `json:"description,omitempty"`
}

func validateCommonReq(typeName, launchConfig string) error {
	if strings.TrimSpace(typeName) == "" {
		return errors.New("type_name is required")
	}
	if strings.TrimSpace(launchConfig) == "" {
		return errors.New("launch_config is required")
	}
	return nil
}

// ApplyReq is the request body of POST /apply.
type ApplyReq struct {
	// TypeName is the resource type name in resource_schemas.
	TypeName string `json:"type_name"`
	// LaunchConfig is the encrypted plugin instance launch configuration.
	LaunchConfig string `json:"launch_config"`
	// PriorState is the current state before apply; nil when creating.
	PriorState map[string]any `json:"prior_state"`
	// ProposedState is the target state to apply; nil when deleting.
	ProposedState map[string]any `json:"proposed_state"`
}

// Validate validates ApplyReq fields.
func (r *ApplyReq) Validate() error {
	if r == nil {
		return errors.New("apply request is nil")
	}
	if err := validateCommonReq(r.TypeName, r.LaunchConfig); err != nil {
		return err
	}
	if r.PriorState == nil && r.ProposedState == nil {
		return errors.New("prior_state and proposed_state cannot both be nil")
	}
	return nil
}

// ApplyData is the response data of POST /apply, POST /result, and POST /stop.
type ApplyData struct {
	// Status is the task status: enumeration values such as: running/succeeded/failed.
	Status TaskStatus `json:"status"`
	// Message is the task failure reason (required when status=failed) or status note.
	Message string `json:"message,omitempty"`
	// ApplyID is the async execution identifier (required when status=running).
	ApplyID string `json:"apply_id,omitempty"`
	// State is the actual resource state after execution (or checkpoint/partial state).
	State map[string]any `json:"state"`
	// Progress is the optional execution progress summary.
	Progress *Progress `json:"progress,omitempty"`
}

// GetResultReq is the request body of POST /result.
type GetResultReq struct {
	// TypeName is the resource type name matching the original Apply call.
	TypeName string `json:"type_name"`
	// LaunchConfig is the encrypted plugin instance launch configuration.
	LaunchConfig string `json:"launch_config"`
	// ApplyID is the execution ID returned by Apply.
	ApplyID string `json:"apply_id"`
}

// Validate validates GetResultReq fields.
func (r *GetResultReq) Validate() error {
	if r == nil {
		return errors.New("get result request is nil")
	}
	if err := validateCommonReq(r.TypeName, r.LaunchConfig); err != nil {
		return err
	}
	if strings.TrimSpace(r.ApplyID) == "" {
		return errors.New("apply_id is required")
	}
	return nil
}

// StopReq is the request body of POST /stop.
type StopReq struct {
	// TypeName is the resource type name matching the original Apply call.
	TypeName string `json:"type_name"`
	// LaunchConfig is the encrypted plugin instance launch configuration.
	LaunchConfig string `json:"launch_config"`
	// ApplyID is the execution ID returned by Apply.
	ApplyID string `json:"apply_id"`
}

// Validate validates StopReq fields.
func (r *StopReq) Validate() error {
	if r == nil {
		return errors.New("stop request is nil")
	}
	if err := validateCommonReq(r.TypeName, r.LaunchConfig); err != nil {
		return err
	}
	if strings.TrimSpace(r.ApplyID) == "" {
		return errors.New("apply_id is required")
	}
	return nil
}

// PlanReq is the request body of POST /plan.
type PlanReq struct {
	// TypeName is the resource type name in resource_schemas.
	TypeName string `json:"type_name"`
	// LaunchConfig is the encrypted plugin instance launch configuration.
	LaunchConfig string `json:"launch_config"`
	// PriorState is the current state before plan; nil when creating.
	PriorState map[string]any `json:"prior_state"`
	// ProposedState is the target state to plan; nil when deleting.
	ProposedState map[string]any `json:"proposed_state"`
}

// Validate validates PlanReq fields.
func (r *PlanReq) Validate() error {
	if r == nil {
		return errors.New("plan request is nil")
	}
	if err := validateCommonReq(r.TypeName, r.LaunchConfig); err != nil {
		return err
	}
	if r.PriorState == nil && r.ProposedState == nil {
		return errors.New("prior_state and proposed_state cannot both be nil")
	}
	return nil
}

// PlanData is the response data of POST /plan.
type PlanData struct {
	// State is the planned resource state; nil when deleting.
	State map[string]any `json:"state"`
	// ReplacePaths lists field names that require resource replacement (delete then recreate).
	ReplacePaths []string `json:"replace_paths"`
}

// ValidateReq is the request body of POST /validate.
type ValidateReq struct {
	// TypeName is the resource or data source type name.
	TypeName string `json:"type_name"`
	// LaunchConfig is the encrypted plugin instance launch configuration.
	LaunchConfig string `json:"launch_config"`
	// Config is the user configuration object to validate.
	Config map[string]any `json:"config"`
}

// Validate validates ValidateReq fields.
func (r *ValidateReq) Validate() error {
	if r == nil {
		return errors.New("validate request is nil")
	}
	if err := validateCommonReq(r.TypeName, r.LaunchConfig); err != nil {
		return err
	}
	if r.Config == nil {
		return errors.New("config is required")
	}
	return nil
}

// ValidateData is the response data of POST /validate.
type ValidateData struct {
	// Diagnostics contains all validation issues found in config.
	Diagnostics []Diagnostic `json:"diagnostics"`
}

// IsPassed returns true if Diagnostics contains no item with severity "error".
func (v *ValidateData) IsPassed() bool {
	if v == nil {
		return true
	}
	for _, d := range v.Diagnostics {
		if d.EffectiveSeverity() == tfjson.DiagnosticSeverityError {
			return false
		}
	}
	return true
}

// ImportReq is the request body of POST /import.
type ImportReq struct {
	// TypeName is the resource type name in resource_schemas.
	TypeName string `json:"type_name"`
	// LaunchConfig is the encrypted plugin instance launch configuration.
	LaunchConfig string `json:"launch_config"`
	// ImportID is the external system object identifier.
	ImportID string `json:"import_id"`
}

// Validate validates ImportReq fields.
func (r *ImportReq) Validate() error {
	if r == nil {
		return errors.New("import request is nil")
	}
	if err := validateCommonReq(r.TypeName, r.LaunchConfig); err != nil {
		return err
	}
	if strings.TrimSpace(r.ImportID) == "" {
		return errors.New("import_id is required")
	}
	return nil
}

// RefreshReq is the request body of POST /refresh.
type RefreshReq struct {
	// TypeName is the resource type name in resource_schemas.
	TypeName string `json:"type_name"`
	// LaunchConfig is the encrypted plugin instance launch configuration.
	LaunchConfig string `json:"launch_config"`
	// State is the previously saved resource state used to locate and refresh the object.
	State map[string]any `json:"state"`
}

// Validate validates RefreshReq fields.
func (r *RefreshReq) Validate() error {
	if r == nil {
		return errors.New("refresh request is nil")
	}
	if err := validateCommonReq(r.TypeName, r.LaunchConfig); err != nil {
		return err
	}
	if len(r.State) == 0 {
		return errors.New("state is required")
	}
	return nil
}

// ReadDataReq is the request body of POST /read_data.
type ReadDataReq struct {
	// TypeName is the data source type name in data_source_schemas.
	TypeName string `json:"type_name"`
	// LaunchConfig is the encrypted plugin instance launch configuration.
	LaunchConfig string `json:"launch_config"`
	// Config is the query filter configuration for the data source.
	Config map[string]any `json:"config"`
}

// Validate validates ReadDataReq fields.
func (r *ReadDataReq) Validate() error {
	if r == nil {
		return errors.New("read_data request is nil")
	}
	if err := validateCommonReq(r.TypeName, r.LaunchConfig); err != nil {
		return err
	}
	if r.Config == nil {
		return errors.New("config is required")
	}
	return nil
}

// StateData is the response data of POST /import, POST /refresh, and POST /read_data.
type StateData struct {
	// State is the complete resource or data source state; nil when the refreshed object no longer exists.
	State map[string]any `json:"state"`
}

// ListReq is the request body of POST /list.
type ListReq struct {
	// TypeName is the resource type name in resource_schemas.
	TypeName string `json:"type_name"`
	// LaunchConfig is the encrypted plugin instance launch configuration.
	LaunchConfig string `json:"launch_config"`
	// Config is the optional filter condition defined by the plugin.
	Config map[string]any `json:"config,omitempty"`
	// Limit is the maximum number of items per page (1-500).
	Limit int `json:"limit"`
	// PageToken is the pagination cursor (empty string for the first page).
	PageToken string `json:"page_token"`
}

// Validate validates ListReq fields.
func (r *ListReq) Validate() error {
	if r == nil {
		return errors.New("list request is nil")
	}
	if err := validateCommonReq(r.TypeName, r.LaunchConfig); err != nil {
		return err
	}
	if r.Limit < MinListLimit || r.Limit > MaxListLimit {
		return fmt.Errorf("limit must be between %d and %d, got %d", MinListLimit, MaxListLimit, r.Limit)
	}
	return nil
}

// ListItem represents a single resource item returned by POST /list.
type ListItem struct {
	// ImportID is the external object ID that can be passed to Import.
	ImportID string `json:"import_id"`
	// State is the complete state of the resource object.
	State map[string]any `json:"state"`
}

// ListData is the response data of POST /list.
type ListData struct {
	// Items contains the resource objects on the current page.
	Items []ListItem `json:"items"`
	// NextPageToken is the cursor for the next page; empty string indicates the last page.
	NextPageToken string `json:"next_page_token"`
}
