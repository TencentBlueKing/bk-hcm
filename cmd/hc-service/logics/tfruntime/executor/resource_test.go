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
	"testing"

	"hcm/cmd/hc-service/types"
	"hcm/pkg/plugin/tfplugin5"
)

func TestPluginExecutor_ReadResource(t *testing.T) {
	fake := &fakePlugin{getSchema: resourceSchema(t)}
	executor := newAPIExecutor(fake)
	executor.SetCodec(&jsonCodec{})
	if _, err := executor.GetProviderSchema(context.Background()); err != nil {
		t.Fatal(err)
	}
	ctx := context.Background()
	if _, err := executor.ReadResource(ctx, "tcloud_cvm", map[string]any{"id": "ins-1"}, nil); err == nil || err.Error() != "terraform plugin is not configured" {
		t.Fatalf("read before configure: %v", err)
	}
	if fake.count("ReadResource") != 0 {
		t.Fatal("read ran before configure")
	}
	loadEchoPrepare(fake)
	configureExecutor(t, executor)
	if _, err := executor.ReadResource(ctx, "tcloud_cvm", map[string]any{"id": "ins-1"}, []byte{1}); err != nil {
		t.Fatal(err)
	}
	if fake.count("ReadResource") != 1 {
		t.Fatalf("read calls = %d", fake.count("ReadResource"))
	}
	if _, err := executor.ReadResource(ctx, "missing", nil, nil); err == nil || err.Error() != `resource type "missing" is not in the schema` {
		t.Fatalf("unknown type: %v", err)
	}
	if fake.count("ReadResource") != 1 {
		t.Fatal("unknown type sent an RPC")
	}
}

func TestPluginExecutor_ValidateResourceTypeConfig(t *testing.T) {
	codec := &jsonCodec{}
	var got *tfplugin5.ValidateResourceTypeConfig_Request
	fake := &fakePlugin{
		getSchema: resourceSchema(t),
		validateResource: func(in *tfplugin5.ValidateResourceTypeConfig_Request) (*tfplugin5.ValidateResourceTypeConfig_Response, error) {
			got = in
			return &tfplugin5.ValidateResourceTypeConfig_Response{}, nil
		},
	}
	loadEchoPrepare(fake)
	executor := preparedExecutor(t, fake, codec)
	before := codec.calls
	if _, err := executor.ValidateResourceTypeConfig(context.Background(), "tcloud_cvm", nil); err != nil {
		t.Fatal(err)
	}
	if codec.calls != before+1 || !nullJSON(got.GetConfig()) {
		t.Fatalf("nil config = calls %d value %s", codec.calls-before, got.GetConfig().GetJson())
	}
}

func TestPluginExecutor_PlanResourceChange_AfterValidateError(t *testing.T) {
	fake := &fakePlugin{
		getSchema: resourceSchema(t),
		validateResource: func(*tfplugin5.ValidateResourceTypeConfig_Request) (*tfplugin5.ValidateResourceTypeConfig_Response, error) {
			return &tfplugin5.ValidateResourceTypeConfig_Response{Diagnostics: errorDiagnostic("invalid")}, nil
		},
	}
	loadEchoPrepare(fake)
	executor := preparedExecutor(t, fake, &jsonCodec{})
	ctx := context.Background()
	if _, err := executor.ValidateResourceTypeConfig(ctx, "tcloud_cvm", map[string]any{"id": "bad"}); !errors.Is(err, errDiagnostic) {
		t.Fatalf("validate err = %v", err)
	}
	if _, err := executor.PlanResourceChange(ctx, "tcloud_cvm", nil, map[string]any{"id": "ins-1"}, nil, nil); err != nil {
		t.Fatal(err)
	}
	if fake.count("PlanResourceChange") != 1 {
		t.Fatalf("plan calls = %d", fake.count("PlanResourceChange"))
	}
}

func TestPluginExecutor_UpgradeResourceState(t *testing.T) {
	upgradeErr := true
	fake := &fakePlugin{
		getSchema: resourceSchema(t),
		upgrade: func(in *tfplugin5.UpgradeResourceState_Request) (*tfplugin5.UpgradeResourceState_Response, error) {
			if upgradeErr {
				return &tfplugin5.UpgradeResourceState_Response{Diagnostics: errorDiagnostic("old")}, nil
			}
			if in.GetVersion() != 1 || string(in.GetRawState().GetJson()) != `{"id":"old"}` || len(in.GetRawState().GetFlatmap()) != 0 {
				return nil, errors.New("raw state was not forwarded as json")
			}
			return &tfplugin5.UpgradeResourceState_Response{
				UpgradedState: &tfplugin5.DynamicValue{Json: []byte(`{"id":"upgraded"}`)},
			}, nil
		},
	}
	loadEchoPrepare(fake)
	executor := preparedExecutor(t, fake, &jsonCodec{})
	ctx := context.Background()
	if _, err := executor.UpgradeResourceState(ctx, "tcloud_cvm", 1, &types.RawState{JSON: []byte(`{"id":"old"}`)}); !errors.Is(err, errDiagnostic) {
		t.Fatalf("upgrade err = %v", err)
	}
	upgradeErr = false
	upgraded, err := executor.UpgradeResourceState(ctx, "tcloud_cvm", 1, &types.RawState{JSON: []byte(`{"id":"old"}`)})
	if err != nil {
		t.Fatal(err)
	}
	if upgraded.UpgradedState["id"] != "upgraded" {
		t.Fatalf("upgraded = %#v", upgraded.UpgradedState)
	}
}

func TestPluginExecutor_UpgradeResourceState_RawState(t *testing.T) {
	fake := &fakePlugin{getSchema: resourceSchema(t)}
	loadEchoPrepare(fake)
	executor := preparedExecutor(t, fake, &jsonCodec{})
	ctx := context.Background()
	_, err := executor.UpgradeResourceState(ctx, "tcloud_cvm", 0, &types.RawState{
		JSON:    []byte(`{}`),
		Flatmap: map[string]string{"id": "x"},
	})
	if err == nil || err.Error() != "terraform plugin raw state must set json or flatmap, not both" || fake.count("UpgradeResourceState") != 0 {
		t.Fatalf("both encodings: err=%v calls=%d", err, fake.count("UpgradeResourceState"))
	}
	_, err = executor.UpgradeResourceState(ctx, "tcloud_cvm", 0, &types.RawState{})
	if err == nil || err.Error() != "terraform plugin raw state must set json or flatmap, not both" || fake.count("UpgradeResourceState") != 0 {
		t.Fatalf("neither encoding: err=%v calls=%d", err, fake.count("UpgradeResourceState"))
	}
}

func TestPluginExecutor_PlanResourceChange(t *testing.T) {
	var planned *tfplugin5.PlanResourceChange_Request
	fake := &fakePlugin{
		getSchema: resourceSchema(t),
		plan: func(in *tfplugin5.PlanResourceChange_Request) (*tfplugin5.PlanResourceChange_Response, error) {
			planned = in
			return &tfplugin5.PlanResourceChange_Response{
				PlannedState: &tfplugin5.DynamicValue{Json: []byte(`{"id":"planned"}`)},
				RequiresReplace: []*tfplugin5.AttributePath{{
					Steps: []*tfplugin5.AttributePath_Step{{
						Selector: &tfplugin5.AttributePath_Step_AttributeName{AttributeName: "name"},
					}},
				}},
			}, nil
		},
	}
	loadEchoPrepare(fake)
	executor := preparedExecutor(t, fake, &jsonCodec{})
	ctx := context.Background()
	if _, err := executor.PlanResourceChange(ctx, "tcloud_cvm", nil, nil, nil, nil); err == nil || err.Error() != "terraform plugin plan has no prior state and no proposed state" {
		t.Fatalf("empty plan: %v", err)
	}
	if fake.count("PlanResourceChange") != 0 {
		t.Fatal("empty plan sent an RPC")
	}

	created, err := executor.PlanResourceChange(ctx, "tcloud_cvm", nil, map[string]any{"id": "new"}, nil, nil)
	if err != nil {
		t.Fatal(err)
	}
	if !nullJSON(planned.GetPriorState()) || !nullJSON(planned.GetConfig()) || nullJSON(planned.GetProposedNewState()) {
		t.Fatal("create plan did not send a null prior state and config")
	}
	if created.PlannedState["id"] != "planned" {
		t.Fatalf("created = %#v", created.PlannedState)
	}

	if _, err = executor.PlanResourceChange(ctx, "tcloud_cvm", map[string]any{"id": "old"}, map[string]any{"id": "new"}, nil, []byte{4}); err != nil {
		t.Fatal(err)
	}
	if planned.GetPriorState() == nil || planned.GetProposedNewState() == nil || planned.GetProviderMeta() != nil ||
		planned.GetClientCapabilities() != nil || planned.GetPriorIdentity() != nil {
		t.Fatal("update plan lost a state or set an omitted protocol field")
	}
	if planned.GetPriorPrivate()[0] != 4 {
		t.Fatalf("prior private = %v", planned.GetPriorPrivate())
	}

	deleted, err := executor.PlanResourceChange(ctx, "tcloud_cvm", map[string]any{"id": "old"}, nil, nil, nil)
	if err != nil {
		t.Fatal(err)
	}
	if !nullJSON(planned.GetProposedNewState()) || !nullJSON(planned.GetConfig()) {
		t.Fatal("delete plan did not send a null proposed state and config")
	}
	if deleted.PlannedState != nil {
		t.Fatalf("delete planned state = %#v", deleted.PlannedState)
	}
	if len(deleted.RequiresReplace) != 1 || deleted.RequiresReplace[0].GetSteps()[0].GetAttributeName() != "name" {
		t.Fatalf("requires replace = %+v", deleted.RequiresReplace)
	}
	if fake.count("ApplyResourceChange") != 0 {
		t.Fatal("plan called apply")
	}
}

func TestPluginExecutor_ApplyResourceChange(t *testing.T) {
	fake := &fakePlugin{
		getSchema: resourceSchema(t),
		apply: func(in *tfplugin5.ApplyResourceChange_Request) (*tfplugin5.ApplyResourceChange_Response, error) {
			resp := &tfplugin5.ApplyResourceChange_Response{
				NewState: &tfplugin5.DynamicValue{Json: []byte(`{"id":"kept"}`)},
				Private:  []byte{8},
			}
			if !nullJSON(in.GetPlannedState()) {
				resp.Diagnostics = errorDiagnostic("apply failed")
			}
			return resp, nil
		},
	}
	loadEchoPrepare(fake)
	executor := preparedExecutor(t, fake, &jsonCodec{})
	ctx := context.Background()
	deleted, err := executor.ApplyResourceChange(ctx, "tcloud_cvm", map[string]any{"id": "old"}, nil, nil, nil)
	if err != nil {
		t.Fatal(err)
	}
	if deleted.NewState != nil || deleted.Private[0] != 8 {
		t.Fatalf("delete apply = %#v private %v", deleted.NewState, deleted.Private)
	}
	failed, err := executor.ApplyResourceChange(ctx, "tcloud_cvm", map[string]any{"id": "old"}, map[string]any{"id": "new"}, nil, nil)
	if !errors.Is(err, errDiagnostic) {
		t.Fatalf("apply err = %v", err)
	}
	if failed.NewState["id"] != "kept" {
		t.Fatalf("failed state = %#v", failed.NewState)
	}
}

func TestPluginExecutor_ReadResource_Private(t *testing.T) {
	var got []byte
	next := []byte{1, 2}
	fake := &fakePlugin{
		getSchema: resourceSchema(t),
		read: func(in *tfplugin5.ReadResource_Request) (*tfplugin5.ReadResource_Response, error) {
			got = append([]byte(nil), in.GetPrivate()...)
			return &tfplugin5.ReadResource_Response{Private: []byte{9, 9}}, nil
		},
	}
	loadEchoPrepare(fake)
	executor := preparedExecutor(t, fake, &jsonCodec{})
	ctx := context.Background()
	first, err := executor.ReadResource(ctx, "tcloud_cvm", nil, next)
	if err != nil {
		t.Fatal(err)
	}
	if string(got) != string(next) || string(first.Private) != string([]byte{9, 9}) {
		t.Fatalf("private got %v returned %v", got, first.Private)
	}
	if _, err = executor.ReadResource(ctx, "tcloud_cvm", nil, []byte{3}); err != nil {
		t.Fatal(err)
	}
	if string(got) != string([]byte{3}) {
		t.Fatalf("executor kept the previous private bytes: %v", got)
	}
}

func TestPluginExecutor_ReadResource_Warning(t *testing.T) {
	fake := &fakePlugin{
		getSchema: resourceSchema(t),
		read: func(*tfplugin5.ReadResource_Request) (*tfplugin5.ReadResource_Response, error) {
			return &tfplugin5.ReadResource_Response{Diagnostics: []*tfplugin5.Diagnostic{{
				Severity: tfplugin5.Diagnostic_WARNING,
				Summary:  "disk full",
				Detail:   "secret-state-body",
			}}}, nil
		},
	}
	loadEchoPrepare(fake)
	executor := preparedExecutor(t, fake, &jsonCodec{})
	read, err := executor.ReadResource(context.Background(), "tcloud_cvm", map[string]any{"secret_id": "ak-test-value"}, nil)
	if err != nil {
		t.Fatal(err)
	}
	if len(read.Diagnostics) != 1 || read.Diagnostics[0].GetSummary() != "disk full" || read.Diagnostics[0].GetDetail() != "secret-state-body" {
		t.Fatalf("diagnostics = %+v", read.Diagnostics)
	}
}

func TestPluginExecutor_PlanResourceChange_CodecError(t *testing.T) {
	codec := &jsonCodec{}
	fake := &fakePlugin{getSchema: resourceSchema(t)}
	loadEchoPrepare(fake)
	executor := preparedExecutor(t, fake, codec)
	codec.err = errors.New("bad value")
	_, err := executor.PlanResourceChange(context.Background(), "tcloud_cvm", nil, map[string]any{"id": "x"}, nil, nil)
	if err == nil || fake.count("PlanResourceChange") != 0 {
		t.Fatalf("plan calls = %d err = %v", fake.count("PlanResourceChange"), err)
	}
}

func TestPluginExecutor_ImportResourceState(t *testing.T) {
	fake := &fakePlugin{
		getSchema: resourceSchema(t),
		importState: func(*tfplugin5.ImportResourceState_Request) (*tfplugin5.ImportResourceState_Response, error) {
			return &tfplugin5.ImportResourceState_Response{
				ImportedResources: []*tfplugin5.ImportResourceState_ImportedResource{{
					TypeName: "tcloud_cvm",
					State:    &tfplugin5.DynamicValue{Json: []byte(`{"id":"ins-1"}`)},
				}},
			}, nil
		},
	}
	loadEchoPrepare(fake)
	executor := preparedExecutor(t, fake, &jsonCodec{})
	imported, err := executor.ImportResourceState(context.Background(), "tcloud_cvm", "ins-1")
	if err != nil {
		t.Fatal(err)
	}
	if imported.Imported[0].State["id"] != "ins-1" {
		t.Fatalf("imported = %#v", imported.Imported)
	}
	if fake.count("ReadResource") != 0 || fake.count("ApplyResourceChange") != 0 {
		t.Fatal("import read or applied the resource")
	}

	fake.importState = func(*tfplugin5.ImportResourceState_Request) (*tfplugin5.ImportResourceState_Response, error) {
		return &tfplugin5.ImportResourceState_Response{Diagnostics: errorDiagnostic("missing")}, nil
	}
	missing, err := executor.ImportResourceState(context.Background(), "tcloud_cvm", "gone")
	if !errors.Is(err, errDiagnostic) || missing.Diagnostics[0].Summary != "missing" {
		t.Fatalf("missing import = %+v err %v", missing, err)
	}
	if fake.count("ApplyResourceChange") != 0 {
		t.Fatal("missing import created a resource")
	}

	fake.importState = func(*tfplugin5.ImportResourceState_Request) (*tfplugin5.ImportResourceState_Response, error) {
		return &tfplugin5.ImportResourceState_Response{
			ImportedResources: []*tfplugin5.ImportResourceState_ImportedResource{{
				TypeName: "tcloud_disk",
				State:    &tfplugin5.DynamicValue{Json: []byte(`{"id":"disk-1"}`)},
			}},
		}, nil
	}
	unknown, err := executor.ImportResourceState(context.Background(), "tcloud_cvm", "ins-1")
	if err == nil || err.Error() != `resource type "tcloud_disk" is not in the schema` || len(unknown.Imported) != 0 {
		t.Fatalf("unknown import type = %+v err %v", unknown, err)
	}
}

func TestPluginExecutor_ReadDataSource(t *testing.T) {
	validateErr := true
	fake := &fakePlugin{
		getSchema: dataSourceSchema(t),
		validateData: func(*tfplugin5.ValidateDataSourceConfig_Request) (*tfplugin5.ValidateDataSourceConfig_Response, error) {
			if validateErr {
				return &tfplugin5.ValidateDataSourceConfig_Response{Diagnostics: errorDiagnostic("bad query")}, nil
			}
			return &tfplugin5.ValidateDataSourceConfig_Response{}, nil
		},
		readData: func(*tfplugin5.ReadDataSource_Request) (*tfplugin5.ReadDataSource_Response, error) {
			return &tfplugin5.ReadDataSource_Response{State: &tfplugin5.DynamicValue{Json: []byte(`{"id":"img-1"}`)}}, nil
		},
	}
	loadEchoPrepare(fake)
	executor := preparedExecutor(t, fake, &jsonCodec{})
	ctx := context.Background()
	if _, err := executor.ValidateDataSourceConfig(ctx, "tcloud_image", map[string]any{"id": "bad"}); !errors.Is(err, errDiagnostic) {
		t.Fatalf("validate err = %v", err)
	}
	if _, err := executor.ReadDataSource(ctx, "tcloud_image", map[string]any{"id": "img-1"}); err != nil {
		t.Fatal(err)
	}
	if fake.count("ReadDataSource") != 1 {
		t.Fatalf("read calls after a validate error = %d", fake.count("ReadDataSource"))
	}
	validateErr = false
	if _, err := executor.ValidateDataSourceConfig(ctx, "tcloud_image", map[string]any{"id": "img-1"}); err != nil {
		t.Fatal(err)
	}
	got, err := executor.ReadDataSource(ctx, "tcloud_image", map[string]any{"id": "img-1"})
	if err != nil {
		t.Fatal(err)
	}
	if got.State["id"] != "img-1" {
		t.Fatalf("state = %#v", got.State)
	}
	if fake.count("PlanResourceChange") != 0 || fake.count("ApplyResourceChange") != 0 {
		t.Fatal("data source read planned or applied")
	}
}
