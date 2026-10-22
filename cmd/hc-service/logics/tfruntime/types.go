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

package tfruntime

import (
	"strings"

	gwplugin "hcm/pkg/thirdparty/api-gateway/plugin"
)

// ProviderSchema 是 GetProviderSchema 返回的统一 Schema，字段与对接方案 6.2 一致。
// 当前指向南向协议已有的 PluginSchema，与其他分支合并时只需调整这一处别名。
type ProviderSchema = gwplugin.PluginSchema

// Severity Diagnostic 的严重程度。
type Severity string

const (
	// SeverityError 该调用失败，执行器转为错误并终止后续调用。
	SeverityError Severity = "Error"
	// SeverityWarning 只记录日志，不影响调用结果。
	SeverityWarning Severity = "Warning"
)

// AttributePath 属性路径，每一段是一个属性名或索引的字符串形式。空表示不指向单个字段。
type AttributePath []string

// String 以点号连接路径各段。
func (p AttributePath) String() string {
	return strings.Join(p, ".")
}

// Diagnostic 响应中对异常字段与原因的描述。
type Diagnostic struct {
	Severity  Severity      `json:"severity"`
	Summary   string        `json:"summary"`
	Detail    string        `json:"detail,omitempty"`
	Attribute AttributePath `json:"attribute,omitempty"`
}

// Object 按 schema 组织的 config / state 对象。nil 表示 null；值为 Unknown{} 表示未知值。
// DynamicValue 的编解码与按 schema 校验由执行器实现负责，本契约不暴露 tftypes。
type Object map[string]any

// Unknown 表示一个当前未知的值（例如计划阶段尚不可知的 Computed 字段）。
type Unknown struct{}

// RawState 尚未按当前 schema 解码的状态，JSON 与 Flatmap 只填其中之一。
type RawState struct {
	JSON    []byte            `json:"json,omitempty"`
	Flatmap map[string]string `json:"flatmap,omitempty"`
}

// ConfigureProviderRequest 配置 Provider。凭证在 Config 中，由上层注入，不通过进程环境变量传递。
type ConfigureProviderRequest struct {
	// Config Provider 配置，未设置的字段为 null。
	Config Object
}

// ConfigureProviderResponse ...
type ConfigureProviderResponse struct {
	Diagnostics []Diagnostic
}

// ValidateResourceTypeConfigRequest 校验资源声明。
type ValidateResourceTypeConfigRequest struct {
	// TypeName 资源类型名，必须是 ResourceSchemas 的 key。
	TypeName string
	// Config 资源声明，未设置的字段为 null，可含未知值。
	Config Object
}

// ValidateResourceTypeConfigResponse ...
type ValidateResourceTypeConfigResponse struct {
	Diagnostics []Diagnostic
}

// ValidateDataSourceConfigRequest 校验数据源的查询条件。
type ValidateDataSourceConfigRequest struct {
	// TypeName 数据源类型名，必须是 DataSourceSchemas 的 key。
	TypeName string
	// Config 查询条件。
	Config Object
}

// ValidateDataSourceConfigResponse ...
type ValidateDataSourceConfigResponse struct {
	Diagnostics []Diagnostic
}

// UpgradeResourceStateRequest 调用方持有的状态 schema 版本落后时，在计划之前把状态升到当前版本。
type UpgradeResourceStateRequest struct {
	TypeName string
	// Version 调用方状态当前的 schema 版本。
	Version int64
	// RawState 尚未按当前 schema 解码的状态。
	RawState RawState
}

// UpgradeResourceStateResponse ...
type UpgradeResourceStateResponse struct {
	// UpgradedState 当前 schema 下的状态。
	UpgradedState Object
	Diagnostics   []Diagnostic
}

// ReadResourceRequest 用已有状态读取资源现状。
type ReadResourceRequest struct {
	TypeName string
	// CurrentState 调用方保存的当前状态。
	CurrentState Object
	// Private 上次该资源返回的私有字节，按字节传入。
	Private []byte
}

// ReadResourceResponse ...
type ReadResourceResponse struct {
	// NewState 插件读到的当前状态，资源已不存在时为 nil。
	NewState Object
	// Private 插件要求随该资源继续保存的私有字节，执行器不解码。
	Private     []byte
	Diagnostics []Diagnostic
}

// PlanResourceChangeRequest 只计算计划，不改外部系统。
//
// PriorState 与 ProposedNewState 是否为 nil 决定操作：
// PriorState 为 nil 是创建，两者都有值是更新，ProposedNewState 为 nil 是删除。
type PlanResourceChangeRequest struct {
	TypeName string
	// Config 调用方给出的资源声明。
	Config Object
	// PriorState 变更前状态，创建时为 nil。
	PriorState Object
	// ProposedNewState 目标状态，删除时为 nil。由调用方传入，执行器不在本方法内合成。
	ProposedNewState Object
	// PriorPrivate 上次 ApplyResourceChange 或 ReadResource 返回的 Private，没有则为空。
	PriorPrivate []byte
}

// PlanResourceChangeResponse ...
type PlanResourceChangeResponse struct {
	// PlannedState 应用之后的预期状态，删除计划为 nil。当时未知的属性保持 Unknown。
	PlannedState Object
	// RequiresReplace 必须先删除再创建的字段，执行器把该列表交给调用方，不在本次调用里删除再创建。
	RequiresReplace []AttributePath
	// PlannedPrivate 交给后续 ApplyResourceChange 的私有字节，按字节传递。
	PlannedPrivate []byte
	Diagnostics    []Diagnostic
}

// ApplyResourceChangeRequest 按已得到的计划修改外部系统。
type ApplyResourceChangeRequest struct {
	TypeName string
	// Config 与对应 PlanResourceChange 的声明相同。
	Config Object
	// PriorState 与计划时相同。
	PriorState Object
	// PlannedState 使用计划响应的 PlannedState。
	PlannedState Object
	// PlannedPrivate 使用计划响应的 PlannedPrivate。
	PlannedPrivate []byte
}

// ApplyResourceChangeResponse ...
type ApplyResourceChangeResponse struct {
	// NewState 应用后的实际状态，删除成功为 nil，不得再含 Unknown。
	NewState Object
	// Private 随资源保存，下次作为 PriorPrivate 或 ReadResource 的 Private 传入。执行器不持久化。
	Private []byte
	// Diagnostics Severity 为 Error 时本次应用失败，已返回的 NewState 仍交给调用方。
	Diagnostics []Diagnostic
}

// ImportResourceStateRequest 按外部标识读入已有资源，不创建资源，也不再调用 ReadResource。
type ImportResourceStateRequest struct {
	TypeName string
	// ID 调用方给出的外部标识，格式由该资源类型的 schema 约定。
	ID string
}

// ImportedResource 读入的一个资源。
type ImportedResource struct {
	TypeName string
	State    Object
	Private  []byte
}

// ImportResourceStateResponse ...
type ImportResourceStateResponse struct {
	ImportedResources []ImportedResource
	Diagnostics       []Diagnostic
}

// ReadDataSourceRequest 按查询条件读取只读数据，不经过 Plan 与 Apply。
type ReadDataSourceRequest struct {
	TypeName string
	// Config 查询条件。
	Config Object
}

// ReadDataSourceResponse ...
type ReadDataSourceResponse struct {
	// State 查询结果，字段服从该数据源的 schema。
	State       Object
	Diagnostics []Diagnostic
}
