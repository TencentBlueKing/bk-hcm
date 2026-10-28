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

package adaptor

import (
	"errors"
	"sync"

	iacplugin "hcm/cmd/hc-service/logics/iac-plugin"
	"hcm/cmd/hc-service/logics/iac-plugin/tf-plugin/pool"
	"hcm/cmd/hc-service/logics/tfruntime/executor"
	"hcm/cmd/hc-service/types"
	coreplugin "hcm/pkg/api/core/plugin"
	proto "hcm/pkg/api/hc-service/iac-plugin"
	"hcm/pkg/criteria/constant"
	"hcm/pkg/criteria/enumor"
	"hcm/pkg/criteria/errf"
	"hcm/pkg/kit"
	"hcm/pkg/logs"
	"hcm/pkg/plugin/tfplugin5"
	cvt "hcm/pkg/tools/converter"
)

// TFPluginAdaptor 封装单次北向请求内本地 tf_provider 插件的进程池租约与多步 gRPC 串行编排。
// 结构体仅持有已完成 ConfigureProvider 的进程租约 lease，不保存实例明文 properties，杜绝后续日志泄露。
type TFPluginAdaptor struct {
	pluginName string
	instanceID string
	schema     *coreplugin.PluginSchema
	lease      pool.ProviderIF
	closeOnce  sync.Once
}

// NewTFPluginAdaptor 从目标插件进程池借出已完成 ConfigureProvider 的 Lease 并构造适配器。
func NewTFPluginAdaptor(kt *kit.Kit, p *pool.PluginPool, schema *coreplugin.PluginSchema,
	instanceID string, properties map[string]any) (*TFPluginAdaptor, error) {

	if kt == nil {
		return nil, errf.New(errf.InvalidParameter, "kit is nil")
	}
	if p == nil {
		return nil, errf.New(errf.InvalidParameter, "plugin pool is nil")
	}
	if schema == nil {
		return nil, errf.New(errf.InvalidParameter, "plugin schema is nil")
	}

	lease, err := p.Acquire(kt, properties)
	if err != nil {
		// 日志与错误消息仅记录插件名、实例 ID 与底层错误，绝不打印明文 properties。
		logs.Errorf("acquire plugin %s process lease for instance %s failed, err: %v, rid: %s",
			p.Name(), instanceID, err, kt.Rid)
		return nil, errf.Newf(errf.Aborted,
			"acquire plugin %s process lease for instance %s failed: %v", p.Name(), instanceID, err)
	}

	return &TFPluginAdaptor{
		pluginName: p.Name(),
		instanceID: instanceID,
		schema:     schema,
		lease:      lease,
	}, nil
}

// Close 幂等归还进程池租约并发槽位。
func (a *TFPluginAdaptor) Close() {
	if a == nil {
		return
	}
	a.closeOnce.Do(func() {
		if a.lease != nil {
			a.lease.Release()
		}
	})
}

// classifyRPCError 区分底层 gRPC/进程通信故障与插件业务返回：
//   - 当 err == nil 时，RPC 成功返回（无诊断，或仅包含 Diagnostic_WARNING / Diagnostic_INVALID）；
//   - 当 errors.Is(err, executor.ErrDiagnostic) 时，RPC 正常返回但包含 Diagnostic_ERROR 级别业务诊断，由调用方读取 Diagnostics 处理；
//   - 其他 err != nil 说明发生进程崩溃、连接断开或编解码异常，转换为外层 errf.Aborted。
func (a *TFPluginAdaptor) classifyRPCError(kt *kit.Kit, typeName string, err error) error {
	if err == nil || errors.Is(err, executor.ErrDiagnostic) {
		return nil
	}
	logs.Errorf("plugin %s instance %s rpc on %s failed, err: %v, rid: %s",
		a.pluginName, a.instanceID, typeName, err, kt.Rid)
	return errf.Newf(errf.Aborted, "plugin %s rpc on %s failed: %v", a.pluginName, typeName, err)
}

// changeParams 封装 Execute 与 Plan 共享的资源变更参数。
type changeParams struct {
	typeName     string
	priorState   map[string]any
	config       map[string]any
	recompute    []string
	priorPrivate *string
}

// changeContext 保存 prepareChangeContext 阶段产出的数据上下文。
type changeContext struct {
	// validateFailed 表示前置 ValidateResourceTypeConfig 或 ReadResource 出现 Error 级诊断，需立即阻断后续 RPC。
	validateFailed bool
	// shortCircuitDelete 表示删除操作在 ReadResource 阶段发现远端对象已不存在，可直接短路成功返回。
	shortCircuitDelete bool
	// drifted 记录前置 ReadResource 刷新到的状态是否与传入的 priorState 存在漂移（含远端已删除转重建）。
	drifted bool
	// currentState 为刷新后的最新基准状态（创建或转重建时为 nil）。
	currentState map[string]any
	// currentPrivate 为刷新后的最新基准 Private 字节切片。
	currentPrivate []byte
	// proposedNewState 为合成后的目标状态（删除时为 nil）。
	proposedNewState map[string]any
	// executorDiags 累积前置阶段产生的告警或错误诊断。
	executorDiags []*tfplugin5.Diagnostic
}

// resolveDriftOutcome 为无 I/O 的漂移分支决策纯函数，语义对齐三方 HTTP 插件的刷新后状态分支：
//   - 删除操作（config == nil）且远端已不存在（refreshedState == nil）：标记 shortCircuitDelete = true；
//   - 更新操作（config != nil）且远端已不存在（refreshedState == nil）：清空基准状态与 Private，转为创建并标记 drifted = true；
//   - 远端仍存在（refreshedState != nil）：以刷新后的状态与 Private 覆盖基准，并对比 priorState 记录 drifted。
func resolveDriftOutcome(priorState, config, refreshedState map[string]any, refreshedPrivate []byte) (
	shortCircuitDelete, drifted bool, baselineState map[string]any, baselinePrivate []byte) {

	if refreshedState == nil {
		if config == nil {
			return true, true, nil, nil
		}
		return false, true, nil, nil
	}
	return false, !stateEqual(priorState, refreshedState), refreshedState, refreshedPrivate
}

// prepareChangeContext 在同一进程 Lease 上串行执行 Execute 与 Plan 共用的前置步骤：
// 1. 静态校验 config 与 recompute（直接复用 iacplugin.ValidateConfigAndRecompute）；
// 2. 当 config != nil 时调用 ValidateResourceTypeConfig 预校验，含 Error 诊断则立即终止；
// 3. 当 priorState != nil 时调用 ReadResource 执行漂移刷新，按 resolveDriftOutcome 处理删除短路、更新转创建或覆盖基准态；
// 4. 调用 iacplugin.ComposeProposedState 合成 proposedNewState。
func (a *TFPluginAdaptor) prepareChangeContext(kt *kit.Kit, in changeParams) (*changeContext, error) {
	if in.priorState == nil && in.config == nil {
		return nil, errf.New(errf.InvalidParameter, "prior_state and config cannot both be nil")
	}
	if a.schema == nil || a.schema.ResourceSchemas == nil {
		return nil, errf.Newf(errf.InvalidParameter, "resource schema not found for %s", in.typeName)
	}
	resSchema := a.schema.ResourceSchemas[in.typeName]
	if resSchema == nil || resSchema.Block == nil {
		return nil, errf.Newf(errf.InvalidParameter, "resource schema block is nil for %s", in.typeName)
	}

	priorPrivateBytes, err := decodePrivate(in.priorPrivate)
	if err != nil {
		return nil, err
	}
	if err := iacplugin.ValidateConfigAndRecompute(resSchema.Block, in.config, in.recompute); err != nil {
		return nil, err
	}

	out := &changeContext{
		currentState:   in.priorState,
		currentPrivate: priorPrivateBytes,
	}

	// 创建或更新操作：先执行插件侧配置预校验，遇到 Error 诊断立即阻断，不触发 ReadResource 与 PlanResourceChange。
	if in.config != nil {
		valResp, valErr := a.lease.ValidateResourceTypeConfig(kt.Ctx, in.typeName, in.config)
		if err := a.classifyRPCError(kt, in.typeName, valErr); err != nil {
			return nil, err
		}
		if valResp != nil {
			out.executorDiags = append(out.executorDiags, valResp.Diagnostics...)
		}
		if valErr != nil {
			out.validateFailed = true
			return out, nil
		}
	}

	// 更新或删除操作：调用 ReadResource 拉取云端最新状态执行漂移检测。
	if in.priorState != nil {
		out, err = a.refreshPriorState(kt, in, priorPrivateBytes, out)
		if err != nil {
			return nil, err
		}
		if out.validateFailed || out.shortCircuitDelete {
			return out, nil
		}
	}

	// 删除操作的 proposedNewState 固定为 nil；创建与更新操作直接复用 iacplugin.ComposeProposedState 合成目标状态。
	if in.config != nil {
		proposed, err := iacplugin.ComposeProposedState(resSchema.Block, out.currentState, in.config, in.recompute)
		if err != nil {
			return nil, err
		}
		out.proposedNewState = proposed
	}
	return out, nil
}

// refreshPriorState 调用 ReadResource 刷新存量资源状态并应用漂移决策，返回更新后的 changeContext。
func (a *TFPluginAdaptor) refreshPriorState(kt *kit.Kit, in changeParams, priorPrivate []byte,
	ctx *changeContext) (*changeContext, error) {

	if ctx == nil {
		return nil, errf.New(errf.InvalidParameter, "change context is nil")
	}

	readResp, readErr := a.lease.ReadResource(kt.Ctx, in.typeName, in.priorState, priorPrivate)
	if err := a.classifyRPCError(kt, in.typeName, readErr); err != nil {
		return nil, err
	}
	next := cvt.PtrToVal(ctx)
	var newState map[string]any
	var newPrivate []byte
	if readResp != nil {
		next.executorDiags = append(next.executorDiags, readResp.Diagnostics...)
		newState = readResp.NewState
		newPrivate = readResp.Private
	}
	if readErr != nil {
		next.validateFailed = true
		return cvt.ValToPtr(next), nil
	}

	next.shortCircuitDelete, next.drifted, next.currentState, next.currentPrivate = resolveDriftOutcome(
		in.priorState, in.config, newState, newPrivate)
	return cvt.ValToPtr(next), nil
}

// Execute 在同一进程 Lease 上串行编排前置校验/漂移刷新、PlanResourceChange 与 ApplyResourceChange，
// 并在失败时按三级优先级挽回最新有效状态。
func (a *TFPluginAdaptor) Execute(kt *kit.Kit, req *proto.ExecuteReq) (*proto.ExecuteResult, error) {
	if req == nil {
		return nil, errf.New(errf.InvalidParameter, "execute request is nil")
	}

	prep, err := a.prepareChangeContext(kt, changeParams{
		typeName:     req.TypeName,
		priorState:   req.PriorState,
		config:       req.Config,
		recompute:    req.Recompute,
		priorPrivate: req.PriorPrivate,
	})
	if err != nil {
		return nil, err
	}
	if prep.validateFailed {
		return a.buildFailedExecuteResult(prep, nil, nil, prep.executorDiags), nil
	}
	if prep.shortCircuitDelete {
		return buildSucceededExecuteResult(nil, nil, prep.executorDiags), nil
	}

	planResp, executorDiags, ok, err := a.planChange(kt, req.TypeName, req.Config, prep)
	if err != nil {
		return nil, err
	}
	if !ok {
		return a.buildFailedExecuteResult(prep, nil, nil, executorDiags), nil
	}

	return a.applyPlannedChange(kt, req, prep, planResp, executorDiags)
}

// planChange 调用 PlanResourceChange 并合并返回累积的诊断列表，不修改入参 prep。
// 返回 ok == false 表示插件产生 Error 级诊断或空响应，调用方应终止后续步骤。
func (a *TFPluginAdaptor) planChange(kt *kit.Kit, typeName string, config map[string]any,
	prep *changeContext) (*types.PlanResult, []*tfplugin5.Diagnostic, bool, error) {

	planResp, planErr := a.lease.PlanResourceChange(
		kt.Ctx, typeName, prep.currentState, prep.proposedNewState, config, prep.currentPrivate)
	if err := a.classifyRPCError(kt, typeName, planErr); err != nil {
		return nil, nil, false, err
	}
	executorDiags := prep.executorDiags
	if planResp != nil {
		executorDiags = append(executorDiags, planResp.Diagnostics...)
	}
	if planResp == nil || planErr != nil {
		return nil, executorDiags, false, nil
	}
	return planResp, executorDiags, true, nil
}

// applyPlannedChange 调用 ApplyResourceChange 应用变更计划，成功时返回最新状态，失败时按挽回策略返回中间状态。
func (a *TFPluginAdaptor) applyPlannedChange(kt *kit.Kit, req *proto.ExecuteReq, prep *changeContext,
	planResp *types.PlanResult, priorDiags []*tfplugin5.Diagnostic) (*proto.ExecuteResult, error) {

	applyResp, applyErr := a.lease.ApplyResourceChange(
		kt.Ctx, req.TypeName, prep.currentState, planResp.PlannedState, req.Config, planResp.PlannedPrivate)
	if err := a.classifyRPCError(kt, req.TypeName, applyErr); err != nil {
		return nil, err
	}
	executorDiags := priorDiags
	var applyState map[string]any
	var applyPrivate []byte
	if applyResp != nil {
		applyState = applyResp.NewState
		applyPrivate = applyResp.Private
		executorDiags = append(executorDiags, applyResp.Diagnostics...)
	}
	if applyErr != nil {
		return a.buildFailedExecuteResult(prep, applyState, applyPrivate, executorDiags), nil
	}

	cleanedState, _ := nullifyUnknowns(applyState)
	return buildSucceededExecuteResult(cleanedState, encodePrivate(applyPrivate), executorDiags), nil
}

func buildSucceededExecuteResult(state map[string]any, private *string,
	executorDiags []*tfplugin5.Diagnostic) *proto.ExecuteResult {

	return &proto.ExecuteResult{
		Status:      enumor.IaCPluginTaskStatusSucceeded,
		State:       state,
		Private:     private,
		Identity:    nil,
		Diagnostics: convertDiagnostics(executorDiags),
	}
}

// buildFailedExecuteResult 按三级状态挽回策略构造失败的 ExecuteResult：
// 1. 优先返回 ApplyResourceChange 阶段已部分落地的非空 applyState；
// 2. 其次当 drifted == true 且 currentState != nil 时，返回前置 ReadResource 刷新到的最新云端状态；
// 3. 否则返回 nil，由上游保留调用前状态。
func (a *TFPluginAdaptor) buildFailedExecuteResult(prep *changeContext, applyState map[string]any,
	applyPrivate []byte, executorDiags []*tfplugin5.Diagnostic) *proto.ExecuteResult {

	var latestState map[string]any
	var latestPrivate []byte
	switch {
	case applyState != nil:
		latestState, _ = nullifyUnknowns(applyState)
		latestPrivate = applyPrivate
	case prep.drifted && prep.currentState != nil:
		latestState, _ = nullifyUnknowns(prep.currentState)
		latestPrivate = prep.currentPrivate
	default:
		latestState = nil
		latestPrivate = nil
	}

	diags := convertDiagnostics(executorDiags)
	return &proto.ExecuteResult{
		Status:      enumor.IaCPluginTaskStatusFailed,
		Message:     extractErrorSummary(diags),
		State:       latestState,
		Private:     encodePrivate(latestPrivate),
		Identity:    nil,
		Diagnostics: diags,
	}
}

// Plan 在同一进程 Lease 上串行执行前置校验/漂移刷新与 PlanResourceChange，
// 并将 PlannedState 中的 codec.Unknown 递归替换为 nil、将 RequiresReplace 转换为路径字符串切片。
func (a *TFPluginAdaptor) Plan(kt *kit.Kit, req *proto.PlanReq) (*proto.PlanResult, error) {
	if req == nil {
		return nil, errf.New(errf.InvalidParameter, "plan request is nil")
	}

	prep, err := a.prepareChangeContext(kt, changeParams{
		typeName:     req.TypeName,
		priorState:   req.PriorState,
		config:       req.Config,
		recompute:    req.Recompute,
		priorPrivate: req.PriorPrivate,
	})
	if err != nil {
		return nil, err
	}
	if prep.validateFailed || prep.shortCircuitDelete {
		return buildEmptyPlanResult(prep.executorDiags), nil
	}

	planResp, executorDiags, ok, err := a.planChange(kt, req.TypeName, req.Config, prep)
	if err != nil {
		return nil, err
	}
	if !ok {
		return buildEmptyPlanResult(executorDiags), nil
	}

	var plannedState map[string]any
	if req.Config != nil {
		plannedState, _ = nullifyUnknowns(planResp.PlannedState)
	}
	return &proto.PlanResult{
		State:        plannedState,
		ReplacePaths: convertReplacePaths(planResp.RequiresReplace),
		Diagnostics:  convertDiagnostics(executorDiags),
	}, nil
}

func buildEmptyPlanResult(executorDiags []*tfplugin5.Diagnostic) *proto.PlanResult {
	return &proto.PlanResult{
		State:        nil,
		ReplacePaths: []string{},
		Diagnostics:  convertDiagnostics(executorDiags),
	}
}

// Import 在同一进程 Lease 上串行调用 ImportResourceState 与 ReadResource 补全完整资源属性。
// 当 Imported 为空或 ReadResource 返回 NewState == nil（远端对象不存在）且插件未给 Error 诊断时，主动封装 Error Diagnostic。
func (a *TFPluginAdaptor) Import(kt *kit.Kit, req *proto.ImportReq) (*proto.StateResult, error) {
	if req == nil {
		return nil, errf.New(errf.InvalidParameter, "import request is nil")
	}

	imported, executorDiags, ok, err := a.importResource(kt, req.TypeName, req.ImportID)
	if err != nil {
		return nil, err
	}
	if !ok {
		return &proto.StateResult{
			State:       nil,
			Diagnostics: convertDiagnostics(executorDiags),
		}, nil
	}
	if imported == nil {
		return buildStateDiagnosticError(executorDiags, constant.TFPluginDiagNoImportedResource), nil
	}

	return a.readImportedResource(kt, req.TypeName, imported, executorDiags)
}

func (a *TFPluginAdaptor) importResource(kt *kit.Kit, typeName, importID string) (
	*types.ImportedResource, []*tfplugin5.Diagnostic, bool, error) {

	impResp, impErr := a.lease.ImportResourceState(kt.Ctx, typeName, importID)
	if err := a.classifyRPCError(kt, typeName, impErr); err != nil {
		return nil, nil, false, err
	}
	var executorDiags []*tfplugin5.Diagnostic
	if impResp != nil {
		executorDiags = impResp.Diagnostics
	}
	if impErr != nil {
		return nil, executorDiags, false, nil
	}
	if impResp == nil || len(impResp.Imported) == 0 {
		return nil, executorDiags, true, nil
	}
	return &impResp.Imported[0], executorDiags, true, nil
}

func (a *TFPluginAdaptor) readImportedResource(kt *kit.Kit, fallbackType string, imported *types.ImportedResource,
	priorDiags []*tfplugin5.Diagnostic) (*proto.StateResult, error) {

	readType := imported.TypeName
	if readType == "" {
		readType = fallbackType
	}
	readResp, readErr := a.lease.ReadResource(kt.Ctx, readType, imported.State, imported.Private)
	if err := a.classifyRPCError(kt, readType, readErr); err != nil {
		return nil, err
	}
	executorDiags := priorDiags
	if readResp != nil {
		executorDiags = append(executorDiags, readResp.Diagnostics...)
	}
	if readErr != nil {
		return &proto.StateResult{
			State:       nil,
			Diagnostics: convertDiagnostics(executorDiags),
		}, nil
	}
	if readResp == nil || readResp.NewState == nil {
		return buildStateDiagnosticError(executorDiags, constant.TFPluginDiagImportNonExistent), nil
	}

	return buildStateResult(readResp.NewState, readResp.Private, executorDiags), nil
}

func buildStateDiagnosticError(executorDiags []*tfplugin5.Diagnostic, summary string) *proto.StateResult {
	diags := append(convertDiagnostics(executorDiags), newErrorDiagnostic(summary, ""))
	return &proto.StateResult{
		State:       nil,
		Diagnostics: diags,
	}
}

func buildStateResult(rawState map[string]any, rawPrivate []byte,
	executorDiags []*tfplugin5.Diagnostic) *proto.StateResult {

	if rawState == nil {
		return &proto.StateResult{
			State:       nil,
			Private:     nil,
			Identity:    nil,
			Diagnostics: convertDiagnostics(executorDiags),
		}
	}
	cleanedState, _ := nullifyUnknowns(rawState)
	return &proto.StateResult{
		State:       cleanedState,
		Private:     encodePrivate(rawPrivate),
		Identity:    nil,
		Diagnostics: convertDiagnostics(executorDiags),
	}
}

// ReadData 在同一进程 Lease 上串行执行 ValidateDataSourceConfig 与 ReadDataSource，
// 并对空状态或残留 codec.Unknown 主动封装 Error Diagnostic。
func (a *TFPluginAdaptor) ReadData(kt *kit.Kit, req *proto.ReadDataReq) (*proto.ReadDataResult, error) {
	if req == nil {
		return nil, errf.New(errf.InvalidParameter, "read_data request is nil")
	}

	executorDiags, ok, err := a.validateDataSource(kt, req.TypeName, req.Config)
	if err != nil {
		return nil, err
	}
	if !ok {
		return &proto.ReadDataResult{
			State:       nil,
			Diagnostics: convertDiagnostics(executorDiags),
		}, nil
	}

	return a.readValidatedDataSource(kt, req, executorDiags)
}

func (a *TFPluginAdaptor) validateDataSource(kt *kit.Kit, typeName string,
	config map[string]any) ([]*tfplugin5.Diagnostic, bool, error) {

	valResp, valErr := a.lease.ValidateDataSourceConfig(kt.Ctx, typeName, config)
	if err := a.classifyRPCError(kt, typeName, valErr); err != nil {
		return nil, false, err
	}
	var executorDiags []*tfplugin5.Diagnostic
	if valResp != nil {
		executorDiags = valResp.Diagnostics
	}
	return executorDiags, valErr == nil, nil
}

func (a *TFPluginAdaptor) readValidatedDataSource(kt *kit.Kit, req *proto.ReadDataReq,
	priorDiags []*tfplugin5.Diagnostic) (*proto.ReadDataResult, error) {

	readResp, readErr := a.lease.ReadDataSource(kt.Ctx, req.TypeName, req.Config)
	if err := a.classifyRPCError(kt, req.TypeName, readErr); err != nil {
		return nil, err
	}
	executorDiags := priorDiags
	if readResp != nil {
		executorDiags = append(executorDiags, readResp.Diagnostics...)
	}
	if readErr != nil {
		return &proto.ReadDataResult{
			State:       nil,
			Diagnostics: convertDiagnostics(executorDiags),
		}, nil
	}
	if readResp == nil || readResp.State == nil {
		diags := append(convertDiagnostics(executorDiags),
			newErrorDiagnostic(constant.TFPluginDiagNullDataSourceState, ""))
		return &proto.ReadDataResult{
			State:       nil,
			Diagnostics: diags,
		}, nil
	}

	cleanedState, hadUnknown := nullifyUnknowns(readResp.State)
	diags := convertDiagnostics(executorDiags)
	if hadUnknown {
		diags = append(diags, newErrorDiagnostic(constant.TFPluginDiagStateNotWhollyKnown, ""))
	}
	return &proto.ReadDataResult{
		State:       cleanedState,
		Diagnostics: diags,
	}, nil
}

// Validate 按 req.ConfigType（"resource" 或 "data"）分流调用 ValidateResourceTypeConfig 或 ValidateDataSourceConfig。
func (a *TFPluginAdaptor) Validate(kt *kit.Kit, req *proto.ValidateReq) (*proto.ValidateResult, error) {
	if req == nil {
		return nil, errf.New(errf.InvalidParameter, "validate request is nil")
	}

	var (
		resp *types.ValidateResult
		err  error
	)
	switch req.ConfigType {
	case proto.ConfigTypeResource:
		resp, err = a.lease.ValidateResourceTypeConfig(kt.Ctx, req.TypeName, req.Config)
	case proto.ConfigTypeDataSource:
		resp, err = a.lease.ValidateDataSourceConfig(kt.Ctx, req.TypeName, req.Config)
	default:
		return nil, errf.Newf(errf.InvalidParameter, "unsupported config_type %q", req.ConfigType)
	}

	if classifyErr := a.classifyRPCError(kt, req.TypeName, err); classifyErr != nil {
		return nil, classifyErr
	}
	var executorDiags []*tfplugin5.Diagnostic
	if resp != nil {
		executorDiags = resp.Diagnostics
	}
	diags := convertDiagnostics(executorDiags)
	if diags == nil {
		diags = []coreplugin.Diagnostic{}
	}
	return &proto.ValidateResult{Diagnostics: diags}, nil
}

// Refresh 解码 Private 并单步调用 ReadResource 刷新资源状态；若远端已删除（NewState == nil 且无 Error 诊断），正常返回 State = nil。
func (a *TFPluginAdaptor) Refresh(kt *kit.Kit, req *proto.RefreshReq) (*proto.StateResult, error) {
	if req == nil {
		return nil, errf.New(errf.InvalidParameter, "refresh request is nil")
	}

	privateBytes, err := decodePrivate(req.Private)
	if err != nil {
		return nil, err
	}

	readResp, readErr := a.lease.ReadResource(kt.Ctx, req.TypeName, req.State, privateBytes)
	if err := a.classifyRPCError(kt, req.TypeName, readErr); err != nil {
		return nil, err
	}
	var executorDiags []*tfplugin5.Diagnostic
	if readResp != nil {
		executorDiags = readResp.Diagnostics
	}
	if readErr != nil || readResp == nil {
		return buildStateResult(nil, nil, executorDiags), nil
	}

	return buildStateResult(readResp.NewState, readResp.Private, executorDiags), nil
}
