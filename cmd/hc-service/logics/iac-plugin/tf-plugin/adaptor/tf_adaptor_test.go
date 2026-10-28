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
	"context"
	"encoding/base64"
	"encoding/json"
	"errors"
	"math/big"
	"os"
	"path/filepath"
	"sync"
	"sync/atomic"
	"testing"
	"time"

	"hcm/cmd/hc-service/logics/iac-plugin/tf-plugin/pool"
	"hcm/cmd/hc-service/logics/tfruntime/codec"
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

	tfjson "github.com/hashicorp/terraform-json"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/zclconf/go-cty/cty"
)

// mockProvider 模拟 pool.ProviderIF，记录每次 RPC 调用次数、入参并支持按场景定制返回值。
type mockProvider struct {
	mu sync.Mutex

	validateResCalls  atomic.Int32
	validateDataCalls atomic.Int32
	readResCalls      atomic.Int32
	planCalls         atomic.Int32
	applyCalls        atomic.Int32
	importCalls       atomic.Int32
	readDataCalls     atomic.Int32
	releaseCalls      atomic.Int32

	lastPlanPrior    map[string]any
	lastPlanProposed map[string]any
	lastPlanConfig   map[string]any
	lastPlanPrivate  []byte
	lastApplyPrior   map[string]any
	lastApplyPlanned map[string]any
	lastApplyPrivate []byte
	lastReadState    map[string]any
	lastReadPrivate  []byte

	validateResFn  func(typeName string, cfg map[string]any) (*types.ValidateResult, error)
	validateDataFn func(typeName string, cfg map[string]any) (*types.ValidateResult, error)
	readResFn      func(typeName string, cur map[string]any, priv []byte) (*types.ReadResourceResult, error)
	planFn         func(typeName string, prior, proposed, cfg map[string]any, priv []byte) (*types.PlanResult, error)
	applyFn        func(typeName string, prior, planned, cfg map[string]any, priv []byte) (*types.ApplyResult, error)
	importFn       func(typeName, id string) (*types.ImportResult, error)
	readDataFn     func(typeName string, cfg map[string]any) (*types.ReadDataSourceResult, error)
}

var _ pool.ProviderIF = (*mockProvider)(nil)

func (m *mockProvider) ValidateResourceTypeConfig(_ context.Context, typeName string, config map[string]any) (
	*types.ValidateResult, error) {

	m.validateResCalls.Add(1)
	if m.validateResFn != nil {
		return m.validateResFn(typeName, config)
	}
	return &types.ValidateResult{}, nil
}

func (m *mockProvider) ValidateDataSourceConfig(_ context.Context, typeName string, config map[string]any) (
	*types.ValidateResult, error) {

	m.validateDataCalls.Add(1)
	if m.validateDataFn != nil {
		return m.validateDataFn(typeName, config)
	}
	return &types.ValidateResult{}, nil
}

func (m *mockProvider) UpgradeResourceState(_ context.Context, _ string, _ int64, _ *types.RawState) (
	*types.UpgradeResult, error) {

	return &types.UpgradeResult{}, nil
}

func (m *mockProvider) ReadResource(_ context.Context, typeName string, current map[string]any, private []byte) (
	*types.ReadResourceResult, error) {

	m.readResCalls.Add(1)
	m.mu.Lock()
	m.lastReadState = current
	m.lastReadPrivate = private
	m.mu.Unlock()
	if m.readResFn != nil {
		return m.readResFn(typeName, current, private)
	}
	return &types.ReadResourceResult{NewState: current, Private: private}, nil
}

func (m *mockProvider) PlanResourceChange(_ context.Context, typeName string,
	prior, proposed, config map[string]any, priorPrivate []byte) (*types.PlanResult, error) {

	m.planCalls.Add(1)
	m.mu.Lock()
	m.lastPlanPrior = prior
	m.lastPlanProposed = proposed
	m.lastPlanConfig = config
	m.lastPlanPrivate = priorPrivate
	m.mu.Unlock()
	if m.planFn != nil {
		return m.planFn(typeName, prior, proposed, config, priorPrivate)
	}
	return &types.PlanResult{PlannedState: proposed, PlannedPrivate: priorPrivate}, nil
}

func (m *mockProvider) ApplyResourceChange(_ context.Context, typeName string,
	prior, planned, config map[string]any, plannedPrivate []byte) (*types.ApplyResult, error) {

	m.applyCalls.Add(1)
	m.mu.Lock()
	m.lastApplyPrior = prior
	m.lastApplyPlanned = planned
	m.lastApplyPrivate = plannedPrivate
	m.mu.Unlock()
	if m.applyFn != nil {
		return m.applyFn(typeName, prior, planned, config, plannedPrivate)
	}
	return &types.ApplyResult{NewState: planned, Private: plannedPrivate}, nil
}

func (m *mockProvider) ImportResourceState(_ context.Context, typeName, id string) (*types.ImportResult, error) {
	m.importCalls.Add(1)
	if m.importFn != nil {
		return m.importFn(typeName, id)
	}
	return &types.ImportResult{
		Imported: []types.ImportedResource{{TypeName: typeName, State: map[string]any{"id": id}}},
	}, nil
}

func (m *mockProvider) ReadDataSource(_ context.Context, typeName string, config map[string]any) (
	*types.ReadDataSourceResult, error) {

	m.readDataCalls.Add(1)
	if m.readDataFn != nil {
		return m.readDataFn(typeName, config)
	}
	return &types.ReadDataSourceResult{State: map[string]any{"id": "ds-1"}}, nil
}

func (m *mockProvider) Release() {
	m.releaseCalls.Add(1)
}

func buildTestPluginSchema() *coreplugin.PluginSchema {
	resBlock := &tfjson.SchemaBlock{
		Attributes: map[string]*tfjson.SchemaAttribute{
			"id":            {AttributeType: cty.String, Computed: true},
			"instance_name": {AttributeType: cty.String, Required: true},
			"image_id":      {AttributeType: cty.String, Optional: true, Computed: true},
			"project_id":    {AttributeType: cty.Number, Optional: true, Computed: true},
			"password":      {AttributeType: cty.String, Optional: true, Sensitive: true},
		},
		NestedBlocks: map[string]*tfjson.SchemaBlockType{
			"clouds": {
				NestingMode: tfjson.SchemaNestingModeSingle,
				Block: &tfjson.SchemaBlock{
					NestedBlocks: map[string]*tfjson.SchemaBlockType{
						"names": {
							NestingMode: tfjson.SchemaNestingModeList,
							Block: &tfjson.SchemaBlock{
								Attributes: map[string]*tfjson.SchemaAttribute{
									"region": {AttributeType: cty.String, Required: true},
									"name":   {AttributeType: cty.String, Optional: true, Computed: true},
								},
							},
						},
					},
				},
			},
		},
	}
	dsBlock := &tfjson.SchemaBlock{
		Attributes: map[string]*tfjson.SchemaAttribute{
			"id":   {AttributeType: cty.String, Computed: true},
			"zone": {AttributeType: cty.String, Optional: true},
		},
	}
	return &coreplugin.PluginSchema{
		Plugin: coreplugin.PluginInfo{Name: "tencentcloud", Version: "1.81.5"},
		SchemaDetail: coreplugin.SchemaDetail{
			ProtocolVersion: "5.3",
			Capabilities: coreplugin.Capabilities{
				Apply: true, Plan: true, Refresh: true, ReadData: true, Validate: true, Import: true,
			},
			ProviderSchema: tfjson.ProviderSchema{
				ResourceSchemas: map[string]*tfjson.Schema{
					"tencentcloud_instance": {Block: resBlock},
					"tencentcloud_zone":     {Block: resBlock},
				},
				DataSourceSchemas: map[string]*tfjson.Schema{
					"tencentcloud_zone": {Block: dsBlock},
				},
			},
		},
	}
}

func newTestAdaptor(mp *mockProvider) *TFPluginAdaptor {
	return &TFPluginAdaptor{
		pluginName: "tencentcloud",
		instanceID: "inst-001",
		schema:     buildTestPluginSchema(),
		lease:      mp,
	}
}

func makeErrorDiag(summary, detail string, steps ...*tfplugin5.AttributePath_Step) *tfplugin5.Diagnostic {
	var attr *tfplugin5.AttributePath
	if len(steps) > 0 {
		attr = &tfplugin5.AttributePath{Steps: steps}
	}
	return &tfplugin5.Diagnostic{
		Severity:  tfplugin5.Diagnostic_ERROR,
		Summary:   summary,
		Detail:    detail,
		Attribute: attr,
	}
}

func TestTFPluginAdaptor_ExecuteAndPlan_RecomputeAndCloseIdempotent(t *testing.T) {
	mp := &mockProvider{
		readResFn: func(_ string, cur map[string]any, _ []byte) (*types.ReadResourceResult, error) {
			return &types.ReadResourceResult{
				NewState: cur,
				Private:  []byte("refreshed-priv"),
			}, nil
		},
		planFn: func(_ string, _, proposed, _ map[string]any, priv []byte) (*types.PlanResult, error) {
			planned := map[string]any{
				"id":            codec.Unknown,
				"instance_name": proposed["instance_name"],
				"image_id":      proposed["image_id"],
				"clouds":        proposed["clouds"],
			}
			return &types.PlanResult{
				PlannedState:   planned,
				PlannedPrivate: priv,
				RequiresReplace: []*tfplugin5.AttributePath{
					{
						Steps: []*tfplugin5.AttributePath_Step{
							{Selector: &tfplugin5.AttributePath_Step_AttributeName{AttributeName: "clouds"}},
							{Selector: &tfplugin5.AttributePath_Step_AttributeName{AttributeName: "names"}},
							{Selector: &tfplugin5.AttributePath_Step_ElementKeyInt{ElementKeyInt: 0}},
							{Selector: &tfplugin5.AttributePath_Step_AttributeName{AttributeName: "region"}},
						},
					},
				},
			}, nil
		},
		applyFn: func(_ string, _, planned, _ map[string]any, _ []byte) (*types.ApplyResult, error) {
			out := map[string]any{
				"id":            "ins-999",
				"instance_name": planned["instance_name"],
				"image_id":      "img-new",
				"clouds":        planned["clouds"],
			}
			return &types.ApplyResult{NewState: out, Private: []byte("applied-priv")}, nil
		},
	}

	adp := newTestAdaptor(mp)
	kt := kit.New()
	prior := map[string]any{
		"id":            "ins-100",
		"instance_name": "vm-old",
		"image_id":      "img-old",
		"clouds": map[string]any{
			"names": []any{
				map[string]any{"region": "ap-guangzhou", "name": "auto-gz"},
			},
		},
	}
	cfg := map[string]any{
		"instance_name": "vm-new",
		"clouds": map[string]any{
			"names": []any{
				map[string]any{"region": "ap-shanghai"},
			},
		},
	}
	priorPrivB64 := base64.StdEncoding.EncodeToString([]byte("prior-priv"))

	// 1. Plan 验证：嵌套列表下标 recompute ("clouds.names[0].name") 置 nil、codec.Unknown 清理为 nil、RequiresReplace 转换
	planRes, err := adp.Plan(kt, &proto.PlanReq{
		CommonInstanceReq: proto.CommonInstanceReq{
			Plugin: "tencentcloud", PluginInstanceID: "inst-001", TypeName: "tencentcloud_instance",
		},
		PriorState:   prior,
		Config:       cfg,
		Recompute:    []string{"clouds.names[0].name"},
		PriorPrivate: cvt.ValToPtr(priorPrivB64),
	})
	require.NoError(t, err)
	assert.Nil(t, planRes.State["id"])
	assert.Equal(t, []string{"clouds.names[0].region"}, planRes.ReplacePaths)
	cloudsMap := mp.lastPlanProposed["clouds"].(map[string]any)
	namesSlice := cloudsMap["names"].([]any)
	assert.Nil(t, namesSlice[0].(map[string]any)["name"])
	assert.Equal(t, "img-old", mp.lastPlanProposed["image_id"])

	// 2. Execute 验证：4 步 RPC 全部落在同一 Lease 上，且 Close 幂等调用 Release 1 次
	execRes, err := adp.Execute(kt, &proto.ExecuteReq{
		CommonInstanceReq: proto.CommonInstanceReq{
			Plugin: "tencentcloud", PluginInstanceID: "inst-001", TypeName: "tencentcloud_instance",
		},
		PriorState:   prior,
		Config:       cfg,
		Recompute:    []string{"clouds.names[0].name"},
		PriorPrivate: cvt.ValToPtr(priorPrivB64),
	})
	require.NoError(t, err)
	assert.Equal(t, enumor.IaCPluginTaskStatusSucceeded, execRes.Status)
	assert.Equal(t, "ins-999", execRes.State["id"])
	assert.Equal(t, base64.StdEncoding.EncodeToString([]byte("applied-priv")), cvt.PtrToVal(execRes.Private))

	adp.Close()
	adp.Close()
	assert.Equal(t, int32(1), mp.releaseCalls.Load())
}

func TestTFPluginAdaptor_ValidateInterceptAndDriftStateMachine(t *testing.T) {
	kt := kit.New()
	common := proto.CommonInstanceReq{
		Plugin: "tencentcloud", PluginInstanceID: "inst-001", TypeName: "tencentcloud_instance",
	}

	// 1. ValidateResourceTypeConfig 返回 Error 诊断时，立即终止，ReadResource/Plan/Apply 调用次数均为 0
	mpValErr := &mockProvider{
		validateResFn: func(_ string, _ map[string]any) (*types.ValidateResult, error) {
			diag := makeErrorDiag("Invalid instance_name", "must not be empty",
				&tfplugin5.AttributePath_Step{
					Selector: &tfplugin5.AttributePath_Step_AttributeName{AttributeName: "instance_name"},
				})
			return &types.ValidateResult{Diagnostics: []*tfplugin5.Diagnostic{diag}}, executor.ErrDiagnostic
		},
	}
	adpVal := newTestAdaptor(mpValErr)
	execRes, err := adpVal.Execute(kt, &proto.ExecuteReq{
		CommonInstanceReq: common,
		PriorState:        map[string]any{"id": "ins-1", "instance_name": "vm"},
		Config:            map[string]any{"instance_name": ""},
	})
	require.NoError(t, err)
	assert.Equal(t, enumor.IaCPluginTaskStatusFailed, execRes.Status)
	assert.Len(t, execRes.Diagnostics, 1)
	assert.Equal(t, int32(0), mpValErr.readResCalls.Load())
	assert.Equal(t, int32(0), mpValErr.planCalls.Load())
	assert.Equal(t, int32(0), mpValErr.applyCalls.Load())

	planRes, err := adpVal.Plan(kt, &proto.PlanReq{
		CommonInstanceReq: common,
		Config:            map[string]any{"instance_name": ""},
	})
	require.NoError(t, err)
	assert.Nil(t, planRes.State)
	assert.Len(t, planRes.Diagnostics, 1)

	// 2. 删除时 ReadResource 返回 NewState == nil -> 短路成功，Plan/Apply 调用次数为 0
	mpRemoteGone := &mockProvider{
		readResFn: func(_ string, _ map[string]any, _ []byte) (*types.ReadResourceResult, error) {
			return &types.ReadResourceResult{NewState: nil}, nil
		},
	}
	adpGone := newTestAdaptor(mpRemoteGone)
	delExec, err := adpGone.Execute(kt, &proto.ExecuteReq{
		CommonInstanceReq: common,
		PriorState:        map[string]any{"id": "ins-gone", "instance_name": "vm"},
		Config:            nil,
	})
	require.NoError(t, err)
	assert.Equal(t, enumor.IaCPluginTaskStatusSucceeded, delExec.Status)
	assert.Nil(t, delExec.State)
	assert.Equal(t, int32(0), mpRemoteGone.planCalls.Load())
	assert.Equal(t, int32(0), mpRemoteGone.applyCalls.Load())

	delPlan, err := adpGone.Plan(kt, &proto.PlanReq{
		CommonInstanceReq: common,
		PriorState:        map[string]any{"id": "ins-gone", "instance_name": "vm"},
		Config:            nil,
	})
	require.NoError(t, err)
	assert.Nil(t, delPlan.State)
	assert.Empty(t, delPlan.ReplacePaths)

	// 3. 更新时 ReadResource 返回 NewState == nil -> 自动转为创建（PlanResourceChange 的 PriorState == nil）
	updExec, err := adpGone.Execute(kt, &proto.ExecuteReq{
		CommonInstanceReq: common,
		PriorState:        map[string]any{"id": "ins-gone", "instance_name": "vm-old", "image_id": "img-old"},
		Config:            map[string]any{"instance_name": "vm-recreate"},
	})
	require.NoError(t, err)
	assert.Equal(t, enumor.IaCPluginTaskStatusSucceeded, updExec.Status)
	assert.Nil(t, mpRemoteGone.lastPlanPrior)
	assert.Nil(t, mpRemoteGone.lastPlanProposed["image_id"])
}

func TestTFPluginAdaptor_ExecuteThreeLevelStateSalvage(t *testing.T) {
	kt := kit.New()
	common := proto.CommonInstanceReq{
		Plugin: "tencentcloud", PluginInstanceID: "inst-001", TypeName: "tencentcloud_instance",
	}
	errDiag := makeErrorDiag("Apply failed", "quota exceeded")

	// Level 1: ApplyResourceChange 返回 Error 诊断但同时返回部分落地的非空 NewState -> 优先返回 applyResp.NewState
	partialState := map[string]any{"id": "ins-partial", "instance_name": "vm-partial"}
	mpLevel1 := &mockProvider{
		readResFn: func(_ string, _ map[string]any, _ []byte) (*types.ReadResourceResult, error) {
			return &types.ReadResourceResult{
				NewState: map[string]any{"id": "ins-1", "instance_name": "vm-drifted"},
			}, nil
		},
		applyFn: func(_ string, _, _, _ map[string]any, _ []byte) (*types.ApplyResult, error) {
			return &types.ApplyResult{
				NewState:    partialState,
				Private:     []byte("partial-priv"),
				Diagnostics: []*tfplugin5.Diagnostic{errDiag},
			}, executor.ErrDiagnostic
		},
	}
	res1, err := newTestAdaptor(mpLevel1).Execute(kt, &proto.ExecuteReq{
		CommonInstanceReq: common,
		PriorState:        map[string]any{"id": "ins-1", "instance_name": "vm-old"},
		Config:            map[string]any{"instance_name": "vm-new"},
	})
	require.NoError(t, err)
	assert.Equal(t, enumor.IaCPluginTaskStatusFailed, res1.Status)
	assert.Equal(t, partialState, res1.State)
	assert.Equal(t, base64.StdEncoding.EncodeToString([]byte("partial-priv")), cvt.PtrToVal(res1.Private))

	// Level 2: Apply 失败且 NewState == nil，但前置 ReadResource 检测到漂移 -> 挽回刷新后的 currentState
	driftedState := map[string]any{"id": "ins-1", "instance_name": "vm-drifted-in-cloud"}
	mpLevel2 := &mockProvider{
		readResFn: func(_ string, _ map[string]any, _ []byte) (*types.ReadResourceResult, error) {
			return &types.ReadResourceResult{
				NewState: driftedState,
				Private:  []byte("drifted-priv"),
			}, nil
		},
		applyFn: func(_ string, _, _, _ map[string]any, _ []byte) (*types.ApplyResult, error) {
			return &types.ApplyResult{
				NewState:    nil,
				Diagnostics: []*tfplugin5.Diagnostic{errDiag},
			}, executor.ErrDiagnostic
		},
	}
	res2, err := newTestAdaptor(mpLevel2).Execute(kt, &proto.ExecuteReq{
		CommonInstanceReq: common,
		PriorState:        map[string]any{"id": "ins-1", "instance_name": "vm-old"},
		Config:            map[string]any{"instance_name": "vm-new"},
	})
	require.NoError(t, err)
	assert.Equal(t, enumor.IaCPluginTaskStatusFailed, res2.Status)
	assert.Equal(t, driftedState, res2.State)
	assert.Equal(t, base64.StdEncoding.EncodeToString([]byte("drifted-priv")), cvt.PtrToVal(res2.Private))

	// Level 3: 首次创建失败且 NewState == nil -> 返回 nil
	mpLevel3 := &mockProvider{
		applyFn: func(_ string, _, _, _ map[string]any, _ []byte) (*types.ApplyResult, error) {
			return &types.ApplyResult{
				NewState:    nil,
				Diagnostics: []*tfplugin5.Diagnostic{errDiag},
			}, executor.ErrDiagnostic
		},
	}
	res3, err := newTestAdaptor(mpLevel3).Execute(kt, &proto.ExecuteReq{
		CommonInstanceReq: common,
		PriorState:        nil,
		Config:            map[string]any{"instance_name": "vm-create"},
	})
	require.NoError(t, err)
	assert.Equal(t, enumor.IaCPluginTaskStatusFailed, res3.Status)
	assert.Nil(t, res3.State)
}

func TestTFPluginAdaptor_ImportReadDataValidateRefreshAndErrors(t *testing.T) {
	kt := kit.New()
	commonRes := proto.CommonInstanceReq{
		Plugin: "tencentcloud", PluginInstanceID: "inst-001", TypeName: "tencentcloud_zone",
	}

	// 1. Import 正例与远端对象不存在（NewState == nil）主动封装 Error Diagnostic
	mpImp := &mockProvider{
		importFn: func(typeName, id string) (*types.ImportResult, error) {
			return &types.ImportResult{
				Imported: []types.ImportedResource{{
					TypeName: typeName,
					State:    map[string]any{"id": id},
					Private:  []byte("imp-priv"),
				}},
			}, nil
		},
		readResFn: func(_ string, cur map[string]any, priv []byte) (*types.ReadResourceResult, error) {
			if cur["id"] == "not-exist" {
				return &types.ReadResourceResult{NewState: nil}, nil
			}
			return &types.ReadResourceResult{
				NewState: map[string]any{"id": cur["id"], "instance_name": "imported-vm"},
				Private:  priv,
			}, nil
		},
	}
	adpImp := newTestAdaptor(mpImp)
	impOK, err := adpImp.Import(kt, &proto.ImportReq{CommonInstanceReq: commonRes, ImportID: "ins-777"})
	require.NoError(t, err)
	assert.Equal(t, "imported-vm", impOK.State["instance_name"])
	assert.Empty(t, impOK.Diagnostics)

	imp404, err := adpImp.Import(kt, &proto.ImportReq{CommonInstanceReq: commonRes, ImportID: "not-exist"})
	require.NoError(t, err)
	assert.Nil(t, imp404.State)
	require.Len(t, imp404.Diagnostics, 1)
	assert.Equal(t, constant.TFPluginDiagImportNonExistent, imp404.Diagnostics[0].Message)

	// 2. ReadData：前置校验拦截、空状态封装诊断、残留 codec.Unknown 清理并封装诊断
	mpDS := &mockProvider{
		readDataFn: func(_ string, cfg map[string]any) (*types.ReadDataSourceResult, error) {
			switch cfg["zone"] {
			case "null-state":
				return &types.ReadDataSourceResult{State: nil}, nil
			case "unknown-val":
				return &types.ReadDataSourceResult{
					State: map[string]any{"id": codec.Unknown, "zone": "ap-gz"},
				}, nil
			default:
				return &types.ReadDataSourceResult{
					State: map[string]any{"id": "zone-1", "zone": cfg["zone"]},
				}, nil
			}
		},
	}
	adpDS := newTestAdaptor(mpDS)
	dsNull, err := adpDS.ReadData(kt, &proto.ReadDataReq{
		CommonInstanceReq: commonRes,
		Config:            map[string]any{"zone": "null-state"},
	})
	require.NoError(t, err)
	assert.Nil(t, dsNull.State)
	require.Len(t, dsNull.Diagnostics, 1)
	assert.Equal(t, constant.TFPluginDiagNullDataSourceState, dsNull.Diagnostics[0].Message)

	dsUnk, err := adpDS.ReadData(kt, &proto.ReadDataReq{
		CommonInstanceReq: commonRes,
		Config:            map[string]any{"zone": "unknown-val"},
	})
	require.NoError(t, err)
	assert.Nil(t, dsUnk.State["id"])
	require.Len(t, dsUnk.Diagnostics, 1)
	assert.Equal(t, constant.TFPluginDiagStateNotWhollyKnown, dsUnk.Diagnostics[0].Message)

	// 3. Validate 按 config_type 分流：同名 tencentcloud_zone 分别触发 Resource 与 DataSource 校验
	beforeDataCalls := mpDS.validateDataCalls.Load()
	_, err = adpDS.Validate(kt, &proto.ValidateReq{
		CommonInstanceReq: commonRes,
		ConfigType:        proto.ConfigTypeResource,
		Config:            map[string]any{"instance_name": "vm"},
	})
	require.NoError(t, err)
	assert.Equal(t, int32(1), mpDS.validateResCalls.Load())
	assert.Equal(t, beforeDataCalls, mpDS.validateDataCalls.Load())

	_, err = adpDS.Validate(kt, &proto.ValidateReq{
		CommonInstanceReq: commonRes,
		ConfigType:        proto.ConfigTypeDataSource,
		Config:            map[string]any{"zone": "ap-gz"},
	})
	require.NoError(t, err)
	assert.Equal(t, beforeDataCalls+1, mpDS.validateDataCalls.Load())

	// 4. Refresh：远端已删除时返回 State == nil 且无 Error；底层非 ErrDiagnostic 故障转 errf.Aborted
	refGone, err := adpImp.Refresh(kt, &proto.RefreshReq{
		CommonInstanceReq: commonRes,
		State:             map[string]any{"id": "not-exist"},
	})
	require.NoError(t, err)
	assert.Nil(t, refGone.State)
	assert.Empty(t, refGone.Diagnostics)

	mpCrash := &mockProvider{
		readResFn: func(_ string, _ map[string]any, _ []byte) (*types.ReadResourceResult, error) {
			return nil, errors.New("rpc connection closed unexpectedly")
		},
	}
	_, err = newTestAdaptor(mpCrash).Refresh(kt, &proto.RefreshReq{
		CommonInstanceReq: commonRes,
		State:             map[string]any{"id": "ins-1"},
	})
	require.Error(t, err)
	assert.Equal(t, errf.Aborted, errf.Error(err).Code)
}

func TestTFPluginAdaptor_NewAdaptorPoolClosedAndSecretRedaction(t *testing.T) {
	const secretMarker = "SK_TF_SUPER_SECRET_999"
	logDir := t.TempDir()
	logs.InitLogger(logs.LogConfig{
		LogDir:         logDir,
		LogMaxSize:     500,
		LogLineMaxSize: 2,
		LogMaxNum:      5,
		ToStdErr:       false,
		AlsoToStdErr:   false,
		Verbosity:      3,
	})

	p, err := pool.NewPluginPool(pool.Options{
		PluginName:    "tencentcloud",
		PluginVersion: "1.81.5",
		ExecPath:      "/nonexistent/terraform-provider-tencentcloud",
		Config: pool.PoolConfig{
			MaxProcesses:   2,
			MaxQueueLength: 4,
			MaxQueueWait:   50 * time.Millisecond,
		},
	})
	require.NoError(t, err)
	require.NoError(t, p.Close(context.Background()))

	kt := kit.New()
	props := map[string]any{
		"secret_id":  "AKID123456",
		"secret_key": secretMarker,
	}
	_, acquireErr := NewTFPluginAdaptor(kt, p, buildTestPluginSchema(), "inst-sec-01", props)
	require.Error(t, acquireErr)
	assert.Equal(t, errf.Aborted, errf.Error(acquireErr).Code)
	assert.NotContains(t, acquireErr.Error(), secretMarker)

	// 再触发一次 RPC 通信失败，断言日志与错误中均不包含敏感凭证特征串
	mpFail := &mockProvider{
		validateResFn: func(_ string, _ map[string]any) (*types.ValidateResult, error) {
			return nil, errors.New("broken pipe")
		},
	}
	adp := newTestAdaptor(mpFail)
	_, rpcErr := adp.Execute(kt, &proto.ExecuteReq{
		CommonInstanceReq: proto.CommonInstanceReq{
			Plugin: "tencentcloud", PluginInstanceID: "inst-sec-01", TypeName: "tencentcloud_instance",
		},
		Config: map[string]any{"instance_name": "vm-1"},
	})
	require.Error(t, rpcErr)
	assert.NotContains(t, rpcErr.Error(), secretMarker)

	logs.CloseLogs()
	entries, _ := os.ReadDir(logDir)
	for _, entry := range entries {
		if entry.IsDir() {
			continue
		}
		raw, readErr := os.ReadFile(filepath.Join(logDir, entry.Name()))
		require.NoError(t, readErr)
		assert.NotContains(t, string(raw), secretMarker)
	}
}

func TestTFPluginAdaptor_50GoroutinesConcurrentAllEndpoints(t *testing.T) {
	const workers = 50
	mp := &mockProvider{
		readResFn: func(_ string, cur map[string]any, priv []byte) (*types.ReadResourceResult, error) {
			return &types.ReadResourceResult{
				NewState: map[string]any{"id": "ins-1", "instance_name": "vm-ok"},
				Private:  priv,
			}, nil
		},
	}
	schema := buildTestPluginSchema()
	commonRes := proto.CommonInstanceReq{
		Plugin: "tencentcloud", PluginInstanceID: "inst-001", TypeName: "tencentcloud_zone",
	}

	var wg sync.WaitGroup
	wg.Add(workers)
	for i := 0; i < workers; i++ {
		go func() {
			defer wg.Done()
			kt := kit.New()
			adp := &TFPluginAdaptor{
				pluginName: "tencentcloud",
				instanceID: "inst-001",
				schema:     schema,
				lease:      mp,
			}
			defer adp.Close()

			_, err := adp.Execute(kt, &proto.ExecuteReq{
				CommonInstanceReq: commonRes,
				PriorState:        map[string]any{"id": "ins-1", "instance_name": "vm-old"},
				Config:            map[string]any{"instance_name": "vm-ok"},
			})
			assert.NoError(t, err)

			_, err = adp.Plan(kt, &proto.PlanReq{
				CommonInstanceReq: commonRes,
				Config:            map[string]any{"instance_name": "vm-ok"},
			})
			assert.NoError(t, err)

			_, err = adp.Validate(kt, &proto.ValidateReq{
				CommonInstanceReq: commonRes,
				ConfigType:        proto.ConfigTypeResource,
				Config:            map[string]any{"instance_name": "vm-ok"},
			})
			assert.NoError(t, err)

			_, err = adp.Import(kt, &proto.ImportReq{
				CommonInstanceReq: commonRes,
				ImportID:          "ins-1",
			})
			assert.NoError(t, err)

			_, err = adp.Refresh(kt, &proto.RefreshReq{
				CommonInstanceReq: commonRes,
				State:             map[string]any{"id": "ins-1"},
			})
			assert.NoError(t, err)

			_, err = adp.ReadData(kt, &proto.ReadDataReq{
				CommonInstanceReq: commonRes,
				Config:            map[string]any{"zone": "ap-guangzhou"},
			})
			assert.NoError(t, err)
		}()
	}
	wg.Wait()
	assert.Equal(t, int32(workers), mp.releaseCalls.Load())
}

func TestTFPluginAdaptor_BigFloatNormalizationAndDriftEquality(t *testing.T) {
	kt := kit.New()
	common := proto.CommonInstanceReq{
		Plugin: "tencentcloud", PluginInstanceID: "inst-001", TypeName: "tencentcloud_instance",
	}

	mp := &mockProvider{
		readResFn: func(_ string, _ map[string]any, priv []byte) (*types.ReadResourceResult, error) {
			return &types.ReadResourceResult{
				NewState: map[string]any{
					"id":            "sg-100",
					"instance_name": "sg-test",
					"project_id":    big.NewFloat(0),
				},
				Private: priv,
			}, nil
		},
		applyFn: func(_ string, _, planned, _ map[string]any, priv []byte) (*types.ApplyResult, error) {
			return &types.ApplyResult{
				NewState: map[string]any{
					"id":            "sg-100",
					"instance_name": planned["instance_name"],
					"project_id":    big.NewFloat(0),
				},
				Private: priv,
			}, nil
		},
	}

	adp := newTestAdaptor(mp)

	// 1. Execute 返回的 *big.Float(0) 被归一化为 int64(0)，JSON 序列化后仍为数字 0 而非字符串 "0"
	execRes, err := adp.Execute(kt, &proto.ExecuteReq{
		CommonInstanceReq: common,
		Config:            map[string]any{"instance_name": "sg-test"},
	})
	require.NoError(t, err)
	assert.Equal(t, int64(0), execRes.State["project_id"])

	rawJSON, err := json.Marshal(execRes.State)
	require.NoError(t, err)
	assert.Contains(t, string(rawJSON), `"project_id":0`)

	var roundTripState map[string]any
	require.NoError(t, json.Unmarshal(rawJSON, &roundTripState))
	assert.Equal(t, float64(0), roundTripState["project_id"])

	// 2. 将 JSON 反序列化后的 roundTripState（float64(0)）作为 prior_state 传入：
	// 与 ReadResource 返回的 *big.Float(0) 数值相等，不应误判为 drifted
	errDiag := makeErrorDiag("Apply failed", "simulated failure")
	mp.applyFn = func(_ string, _, _, _ map[string]any, _ []byte) (*types.ApplyResult, error) {
		return &types.ApplyResult{
			NewState:    nil,
			Diagnostics: []*tfplugin5.Diagnostic{errDiag},
		}, executor.ErrDiagnostic
	}
	failedNoDrift, err := adp.Execute(kt, &proto.ExecuteReq{
		CommonInstanceReq: common,
		PriorState:        roundTripState,
		Config:            map[string]any{"instance_name": "sg-test"},
	})
	require.NoError(t, err)
	assert.Equal(t, enumor.IaCPluginTaskStatusFailed, failedNoDrift.Status)
	assert.Nil(t, failedNoDrift.State)
}
