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

package httpplugin

import (
	"crypto/rand"
	"crypto/rsa"
	"crypto/sha256"
	"crypto/x509"
	"encoding/base64"
	"encoding/json"
	"encoding/pem"
	"errors"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"sync/atomic"
	"testing"

	iacplugin "hcm/cmd/hc-service/logics/iac-plugin"
	coreplugin "hcm/pkg/api/core/plugin"
	proto "hcm/pkg/api/hc-service/iac-plugin"
	"hcm/pkg/criteria/enumor"
	"hcm/pkg/criteria/errf"
	"hcm/pkg/kit"
	"hcm/pkg/logs"
	"hcm/pkg/thirdparty/api-gateway/plugin"

	tfjson "github.com/hashicorp/terraform-json"
	"github.com/stretchr/testify/assert"
	"github.com/zclconf/go-cty/cty"
)

var (
	rsaKeyOnce    sync.Once
	sharedPrivKey *rsa.PrivateKey
	sharedPubPEM  string
)

func generateTestRSAKeyPair(t testing.TB) (*rsa.PrivateKey, string) {
	t.Helper()
	rsaKeyOnce.Do(func() {
		priv, err := rsa.GenerateKey(rand.Reader, 3072)
		if err != nil {
			t.Fatalf("generate rsa key failed: %v", err)
		}
		der, err := x509.MarshalPKIXPublicKey(&priv.PublicKey)
		if err != nil {
			t.Fatalf("marshal public key failed: %v", err)
		}
		sharedPrivKey = priv
		sharedPubPEM = string(pem.EncodeToMemory(&pem.Block{Type: "PUBLIC KEY", Bytes: der}))
	})
	return sharedPrivKey, sharedPubPEM
}

func decryptLaunchConfig(t testing.TB, priv *rsa.PrivateKey, launchConfig string) map[string]any {
	t.Helper()
	cipherBytes, err := base64.StdEncoding.DecodeString(launchConfig)
	assert.NoError(t, err)
	plainBytes, err := rsa.DecryptOAEP(sha256.New(), rand.Reader, priv, cipherBytes, nil)
	assert.NoError(t, err)
	var out map[string]any
	assert.NoError(t, json.Unmarshal(plainBytes, &out))
	return out
}

func buildCMDBHostBlock() *tfjson.SchemaBlock {
	return &tfjson.SchemaBlock{
		Attributes: map[string]*tfjson.SchemaAttribute{
			"bk_biz_id": {
				AttributeType: cty.Number,
				Required:      true,
			},
			"operator": {
				AttributeType: cty.String,
				Optional:      true,
			},
			"bk_cloud_id": {
				AttributeType: cty.Number,
				Optional:      true,
				Computed:      true,
			},
			"bk_module_id": {
				AttributeType: cty.Number,
				Optional:      true,
				Computed:      true,
			},
			"bk_host_id": {
				AttributeType: cty.Number,
				Computed:      true,
			},
		},
	}
}

type testPluginFixture struct {
	Plugin     *coreplugin.IaCPlugin
	Schema     *coreplugin.PluginSchema
	Properties map[string]any
}

func buildTestRouteContext(pubPEM string, props map[string]any) *testPluginFixture {
	extBytes, _ := json.Marshal(coreplugin.HTTPExtension{
		APIPath:      "/api/bkcmdb",
		CryptoType:   "RSA",
		CryptoConfig: &coreplugin.HTTPCryptoConfig{PublicKey: pubPEM},
	})
	detail := coreplugin.SchemaDetail{
		ProtocolVersion: "1.0",
		Capabilities: coreplugin.Capabilities{
			Apply: true, Plan: true, Refresh: true, ReadData: true,
			AsyncApply: true, Stop: true, Validate: true, Import: true, List: true,
		},
		ProviderSchema: tfjson.ProviderSchema{
			ConfigSchema: &tfjson.Schema{Block: &tfjson.SchemaBlock{}},
			ResourceSchemas: map[string]*tfjson.Schema{
				"bkcmdb_host": {Block: buildCMDBHostBlock()},
			},
			DataSourceSchemas: map[string]*tfjson.Schema{
				"bkcmdb_biz": {Block: &tfjson.SchemaBlock{}},
			},
		},
	}
	info := coreplugin.PluginInfo{Name: "bkcmdb", Version: "1.2.0"}
	hash, _ := coreplugin.ComputeSchemaHash(info, &detail)
	return &testPluginFixture{
		Plugin: &coreplugin.IaCPlugin{
			ID: "p1", PluginName: "bkcmdb", PluginVersion: "1.2.0",
			Protocol: enumor.IaCPluginProtocolHTTP, State: enumor.IaCPluginStateEnabled, Extension: extBytes,
		},
		Schema: &coreplugin.PluginSchema{
			Plugin: info, SchemaHash: hash, SchemaDetail: detail,
		},
		Properties: props,
	}
}

func TestComposeProposedState_PositiveAndRecompute(t *testing.T) {
	block := buildCMDBHostBlock()

	// Delete (config == nil) -> proposed_state == nil
	delState, err := iacplugin.ComposeProposedState(block, map[string]any{"bk_host_id": 10086}, nil, nil)
	assert.NoError(t, err)
	assert.Nil(t, delState)

	// AC-009: Update with 4 attribute categories + recompute=["bk_module_id"]
	prior := map[string]any{
		"bk_biz_id":    213,
		"bk_cloud_id":  0,
		"bk_module_id": 5001,
		"operator":     "old",
		"bk_host_id":   10086,
	}
	cfg := map[string]any{"bk_biz_id": 213}
	proposed, err := iacplugin.ComposeProposedState(block, prior, cfg, []string{"bk_module_id"})
	assert.NoError(t, err)
	assert.Equal(t, map[string]any{
		"bk_biz_id":    213,
		"bk_cloud_id":  0,
		"bk_module_id": nil,
		"operator":     nil,
		"bk_host_id":   10086,
	}, proposed)
}

func TestComposeProposedState_NegativeValidation(t *testing.T) {
	block := buildCMDBHostBlock()
	prior := map[string]any{"bk_biz_id": 213}

	cases := []struct {
		name      string
		config    map[string]any
		recompute []string
	}{
		{
			name:   "config contains pure computed attribute",
			config: map[string]any{"bk_biz_id": 213, "bk_host_id": 10086},
		},
		{
			name:   "config contains unknown attribute",
			config: map[string]any{"bk_biz_id": 213, "unknown": 1},
		},
		{
			name:      "recompute contains required attribute",
			config:    map[string]any{"bk_biz_id": 213},
			recompute: []string{"bk_biz_id"},
		},
		{
			name:      "recompute contains pure computed attribute",
			config:    map[string]any{"bk_biz_id": 213},
			recompute: []string{"bk_host_id"},
		},
		{
			name:      "recompute contains unknown attribute",
			config:    map[string]any{"bk_biz_id": 213},
			recompute: []string{"unknown_field"},
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			_, err := iacplugin.ComposeProposedState(block, prior, tc.config, tc.recompute)
			assert.Error(t, err)
			assert.Equal(t, errf.InvalidParameter, errf.Error(err).Code)
		})
	}
}

func buildNestedResourceBlock() *tfjson.SchemaBlock {
	return &tfjson.SchemaBlock{
		Attributes: map[string]*tfjson.SchemaAttribute{
			"id":            {AttributeType: cty.String, Computed: true},
			"instance_name": {AttributeType: cty.String, Required: true},
			"network": {
				Optional: true,
				Computed: true,
				AttributeNestedType: &tfjson.SchemaNestedAttributeType{
					NestingMode: tfjson.SchemaNestingModeSingle,
					Attributes: map[string]*tfjson.SchemaAttribute{
						"subnet_id":   {AttributeType: cty.String, Required: true},
						"private_ip":  {AttributeType: cty.String, Optional: true, Computed: true},
						"mac_address": {AttributeType: cty.String, Computed: true},
					},
				},
			},
		},
		NestedBlocks: map[string]*tfjson.SchemaBlockType{
			"system_disk": {
				NestingMode: tfjson.SchemaNestingModeSingle,
				Block: &tfjson.SchemaBlock{
					Attributes: map[string]*tfjson.SchemaAttribute{
						"disk_size": {AttributeType: cty.Number, Required: true},
						"disk_type": {AttributeType: cty.String, Optional: true, Computed: true},
						"disk_id":   {AttributeType: cty.String, Computed: true},
					},
				},
			},
			"forward_balancer_ids": {
				NestingMode: tfjson.SchemaNestingModeList,
				Block: &tfjson.SchemaBlock{
					Attributes: map[string]*tfjson.SchemaAttribute{
						"load_balancer_id": {AttributeType: cty.String, Required: true},
						"rule_id":          {AttributeType: cty.String, Optional: true, Computed: true},
						"kms_key_id":       {AttributeType: cty.String, Optional: true, Computed: true},
						"attachment_id":    {AttributeType: cty.String, Computed: true},
					},
					NestedBlocks: map[string]*tfjson.SchemaBlockType{
						"target_attribute": {
							NestingMode: tfjson.SchemaNestingModeList,
							Block: &tfjson.SchemaBlock{
								Attributes: map[string]*tfjson.SchemaAttribute{
									"port":      {AttributeType: cty.Number, Required: true},
									"weight":    {AttributeType: cty.Number, Optional: true, Computed: true},
									"target_id": {AttributeType: cty.String, Computed: true},
								},
							},
						},
					},
				},
			},
			"labels_map": {
				NestingMode: tfjson.SchemaNestingModeMap,
				Block: &tfjson.SchemaBlock{
					Attributes: map[string]*tfjson.SchemaAttribute{
						"value":  {AttributeType: cty.String, Required: true},
						"tag_id": {AttributeType: cty.String, Optional: true, Computed: true},
					},
				},
			},
		},
	}
}

func TestComposeProposedState_NestedBlocksAndRecompute(t *testing.T) {
	block := buildNestedResourceBlock()
	prior := map[string]any{
		"id":            "ins-10001",
		"instance_name": "web-prod-01",
		"network": map[string]any{
			"subnet_id":   "subnet-01",
			"private_ip":  "10.0.1.10",
			"mac_address": "52:54:00:12:34:56",
		},
		"system_disk": map[string]any{
			"disk_id":   "disk-sys-01",
			"disk_size": 50,
			"disk_type": "CLOUD_PREMIUM",
		},
		"forward_balancer_ids": []any{
			map[string]any{
				"load_balancer_id": "lb-01",
				"rule_id":          "loc-01",
				"kms_key_id":       "kms-old",
				"attachment_id":    "att-01",
				"target_attribute": []any{
					map[string]any{"port": 80, "weight": 10, "target_id": "tgt-01"},
				},
			},
		},
		"labels_map": map[string]any{
			"env": map[string]any{"value": "prod", "tag_id": "tag-01"},
		},
	}
	cfg := map[string]any{
		"instance_name": "web-prod-01-new",
		"network":       map[string]any{"subnet_id": "subnet-01"},
		"system_disk":   map[string]any{"disk_size": 100},
		"forward_balancer_ids": []any{
			map[string]any{
				"load_balancer_id": "lb-01",
				"target_attribute": []any{map[string]any{"port": 80}},
			},
			map[string]any{
				"load_balancer_id": "lb-02",
				"target_attribute": []any{map[string]any{"port": 8080, "weight": 20}},
			},
		},
		"labels_map": map[string]any{
			"env": map[string]any{"value": "staging"},
		},
	}

	recompute := []string{"forward_balancer_ids[0].kms_key_id", "labels_map.env.tag_id"}
	proposed, err := iacplugin.ComposeProposedState(block, prior, cfg, recompute)
	assert.NoError(t, err)
	assert.Equal(t, map[string]any{
		"id":            "ins-10001",
		"instance_name": "web-prod-01-new",
		"network": map[string]any{
			"subnet_id":   "subnet-01",
			"private_ip":  "10.0.1.10",
			"mac_address": "52:54:00:12:34:56",
		},
		"system_disk": map[string]any{
			"disk_id":   "disk-sys-01",
			"disk_size": 100,
			"disk_type": "CLOUD_PREMIUM",
		},
		"forward_balancer_ids": []any{
			map[string]any{
				"load_balancer_id": "lb-01",
				"rule_id":          "loc-01",
				"kms_key_id":       nil,
				"attachment_id":    "att-01",
				"target_attribute": []any{
					map[string]any{"port": 80, "weight": 10, "target_id": "tgt-01"},
				},
			},
			map[string]any{
				"load_balancer_id": "lb-02",
				"rule_id":          nil,
				"kms_key_id":       nil,
				"attachment_id":    nil,
				"target_attribute": []any{
					map[string]any{"port": 8080, "weight": 20, "target_id": nil},
				},
			},
		},
		"labels_map": map[string]any{
			"env": map[string]any{"value": "staging", "tag_id": nil},
		},
	}, proposed)
}

func TestComposeProposedState_NestedNegativeValidation(t *testing.T) {
	block := buildNestedResourceBlock()
	prior := map[string]any{"instance_name": "web-01"}

	cases := []struct {
		name      string
		config    map[string]any
		recompute []string
	}{
		{
			name: "unknown attribute in 2nd-level nested block",
			config: map[string]any{
				"instance_name": "web-01",
				"forward_balancer_ids": []any{
					map[string]any{
						"load_balancer_id": "lb-01",
						"target_attribute": []any{map[string]any{"port": 80, "bad_field": 1}},
					},
				},
			},
		},
		{
			name: "pure computed attribute in nested block",
			config: map[string]any{
				"instance_name": "web-01",
				"system_disk":   map[string]any{"disk_size": 50, "disk_id": "disk-manual"},
			},
		},
		{
			name: "pure computed attribute in AttributeNestedType",
			config: map[string]any{
				"instance_name": "web-01",
				"network":       map[string]any{"subnet_id": "s-1", "mac_address": "00:11"},
			},
		},
		{
			name: "nested list block given non-slice value",
			config: map[string]any{
				"instance_name":        "web-01",
				"forward_balancer_ids": "not-a-slice",
			},
		},
		{
			name:      "recompute references pure computed nested attribute",
			config:    map[string]any{"instance_name": "web-01"},
			recompute: []string{"system_disk.disk_id"},
		},
		{
			name:      "recompute uses slice index on single nested block",
			config:    map[string]any{"instance_name": "web-01"},
			recompute: []string{"system_disk[0].disk_type"},
		},
		{
			name:      "recompute has invalid slice index syntax",
			config:    map[string]any{"instance_name": "web-01"},
			recompute: []string{"forward_balancer_ids[abc].rule_id"},
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			_, err := iacplugin.ComposeProposedState(block, prior, tc.config, tc.recompute)
			assert.Error(t, err)
			assert.Equal(t, errf.InvalidParameter, errf.Error(err).Code)
		})
	}
}

type mockSouthbound struct {
	mu           sync.Mutex
	refreshCalls int32
	applyCalls   int32
	planCalls    int32
	lastLaunch   string
	lastApplyReq *plugin.ApplyReq
	lastPlanReq  *plugin.PlanReq
	refreshState map[string]any
	refreshErr   error
	applyData    *plugin.ApplyData
	applyErr     error
	planData     *plugin.PlanData
	planErr      error
}

func (m *mockSouthbound) Health(_ *kit.Kit) (*plugin.HealthData, error) {
	return &plugin.HealthData{PluginVersion: "1.2.0"}, nil
}

func (m *mockSouthbound) GetSchema(_ *kit.Kit) (*plugin.PluginSchema, error) {
	return nil, errors.New("unexpected GetSchema call")
}

func (m *mockSouthbound) Apply(_ *kit.Kit, req *plugin.ApplyReq) (*plugin.ApplyData, error) {
	atomic.AddInt32(&m.applyCalls, 1)
	m.mu.Lock()
	m.lastLaunch = req.LaunchConfig
	m.lastApplyReq = req
	data, err := m.applyData, m.applyErr
	m.mu.Unlock()
	return data, err
}

func (m *mockSouthbound) GetResult(_ *kit.Kit, req *plugin.GetResultReq) (*plugin.ApplyData, error) {
	m.mu.Lock()
	m.lastLaunch = req.LaunchConfig
	m.mu.Unlock()
	return &plugin.ApplyData{Status: enumor.IaCPluginTaskStatusSucceeded, ApplyID: req.ApplyID}, nil
}

func (m *mockSouthbound) Stop(_ *kit.Kit, req *plugin.StopReq) (*plugin.ApplyData, error) {
	m.mu.Lock()
	m.lastLaunch = req.LaunchConfig
	m.mu.Unlock()
	return &plugin.ApplyData{Status: enumor.IaCPluginTaskStatusFailed, Message: "stopped"}, nil
}

func (m *mockSouthbound) Plan(_ *kit.Kit, req *plugin.PlanReq) (*plugin.PlanData, error) {
	atomic.AddInt32(&m.planCalls, 1)
	m.mu.Lock()
	m.lastLaunch = req.LaunchConfig
	m.lastPlanReq = req
	data, err := m.planData, m.planErr
	m.mu.Unlock()
	return data, err
}

func (m *mockSouthbound) Validate(_ *kit.Kit, req *plugin.ValidateReq) (*plugin.ValidateData, error) {
	m.mu.Lock()
	m.lastLaunch = req.LaunchConfig
	m.mu.Unlock()
	return &plugin.ValidateData{
		Diagnostics: []plugin.Diagnostic{{Message: "invalid field", Path: []plugin.PathStep{plugin.NewAttributeStep("bk_biz_id")}}},
	}, nil
}

func (m *mockSouthbound) Import(_ *kit.Kit, req *plugin.ImportReq) (*plugin.StateData, error) {
	m.mu.Lock()
	m.lastLaunch = req.LaunchConfig
	m.mu.Unlock()
	return &plugin.StateData{State: map[string]any{"import_id": req.ImportID}}, nil
}

func (m *mockSouthbound) Refresh(_ *kit.Kit, req *plugin.RefreshReq) (*plugin.StateData, error) {
	atomic.AddInt32(&m.refreshCalls, 1)
	m.mu.Lock()
	m.lastLaunch = req.LaunchConfig
	state, err := m.refreshState, m.refreshErr
	m.mu.Unlock()
	if err != nil {
		return nil, err
	}
	return &plugin.StateData{State: state}, nil
}

func (m *mockSouthbound) ReadData(_ *kit.Kit, req *plugin.ReadDataReq) (*plugin.StateData, error) {
	m.mu.Lock()
	m.lastLaunch = req.LaunchConfig
	m.mu.Unlock()
	return &plugin.StateData{State: map[string]any{"biz_id": 213}}, nil
}

func (m *mockSouthbound) List(_ *kit.Kit, req *plugin.ListReq) (*plugin.ListData, error) {
	m.mu.Lock()
	m.lastLaunch = req.LaunchConfig
	m.mu.Unlock()
	return &plugin.ListData{Items: []plugin.ListItem{{ImportID: "1"}}}, nil
}

func newTestExecutor(sb plugin.Client) *Service {
	svc := New(nil, nil, nil, nil, Options{Enabled: true})
	svc.pluginCliMap.Store(pluginClientInfo{pluginName: "bkcmdb", apiPath: "/api/bkcmdb"}, sb)
	return svc
}

func TestExecutor_DriftDetectionShortCircuitAndConvertToCreate(t *testing.T) {
	priv, pubPEM := generateTestRSAKeyPair(t)
	routeCtx := buildTestRouteContext(pubPEM, map[string]any{"token": "tok_123"})
	sb := &mockSouthbound{
		refreshState: nil,
		applyData:    &plugin.ApplyData{Status: enumor.IaCPluginTaskStatusSucceeded, State: map[string]any{"bk_host_id": 1}},
		planData:     &plugin.PlanData{State: map[string]any{"bk_host_id": nil}, ReplacePaths: []string{}},
	}
	exec := newTestExecutor(sb)
	kt := kit.New()

	// AC-007: Delete when Refresh returns state=nil -> short-circuit without calling Apply or Plan
	delExecRes, err := exec.Execute(kt, routeCtx.Plugin, routeCtx.Schema, routeCtx.Properties, &proto.ExecuteReq{
		CommonInstanceReq: proto.CommonInstanceReq{Plugin: "bkcmdb", PluginInstanceID: "i1", TypeName: "bkcmdb_host"},
		PriorState:        map[string]any{"bk_host_id": 10086},
		Config:            nil,
	})
	assert.NoError(t, err)
	assert.Equal(t, enumor.IaCPluginTaskStatusSucceeded, delExecRes.Status)
	assert.Nil(t, delExecRes.State)
	assert.Equal(t, int32(0), atomic.LoadInt32(&sb.applyCalls))

	delPlanRes, err := exec.Plan(kt, routeCtx.Plugin, routeCtx.Schema, routeCtx.Properties, &proto.PlanReq{
		CommonInstanceReq: proto.CommonInstanceReq{Plugin: "bkcmdb", PluginInstanceID: "i1", TypeName: "bkcmdb_host"},
		PriorState:        map[string]any{"bk_host_id": 10086},
		Config:            nil,
	})
	assert.NoError(t, err)
	assert.Nil(t, delPlanRes.State)
	assert.Empty(t, delPlanRes.ReplacePaths)
	assert.Equal(t, int32(0), atomic.LoadInt32(&sb.planCalls))

	// AC-008: Update when Refresh returns state=nil -> converted to Create (PriorState=nil in southbound Apply)
	updExecRes, err := exec.Execute(kt, routeCtx.Plugin, routeCtx.Schema, routeCtx.Properties, &proto.ExecuteReq{
		CommonInstanceReq: proto.CommonInstanceReq{Plugin: "bkcmdb", PluginInstanceID: "i1", TypeName: "bkcmdb_host"},
		PriorState:        map[string]any{"bk_biz_id": 213, "bk_cloud_id": 0, "bk_host_id": 10086},
		Config:            map[string]any{"bk_biz_id": 213},
	})
	assert.NoError(t, err)
	assert.Equal(t, enumor.IaCPluginTaskStatusSucceeded, updExecRes.Status)
	assert.Equal(t, int32(1), atomic.LoadInt32(&sb.applyCalls))
	assert.Nil(t, sb.lastApplyReq.PriorState)
	assert.Nil(t, sb.lastApplyReq.ProposedState["bk_cloud_id"])
	assert.Nil(t, sb.lastApplyReq.ProposedState["bk_host_id"])
	decrypted := decryptLaunchConfig(t, priv, sb.lastLaunch)
	assert.Equal(t, "tok_123", decrypted["token"])
}

func TestExecutor_FailedApplyPreservesRefreshedStateAndAllEndpoints(t *testing.T) {
	priv, pubPEM := generateTestRSAKeyPair(t)
	routeCtx := buildTestRouteContext(pubPEM, map[string]any{"token": "tok_abc"})
	driftedState := map[string]any{"bk_biz_id": 213, "bk_cloud_id": 9, "bk_host_id": 10086}
	sb := &mockSouthbound{
		refreshState: driftedState,
		applyData: &plugin.ApplyData{
			Status:  enumor.IaCPluginTaskStatusFailed,
			Message: "apply err",
			State:   nil,
		},
	}
	exec := newTestExecutor(sb)
	kt := kit.New()
	commonReq := proto.CommonInstanceReq{Plugin: "bkcmdb", PluginInstanceID: "i1", TypeName: "bkcmdb_host"}

	// AC-011: Execute failed with nil state returns drifted state from pre-Refresh and outer error is nil
	execRes, err := exec.Execute(kt, routeCtx.Plugin, routeCtx.Schema, routeCtx.Properties, &proto.ExecuteReq{
		CommonInstanceReq: commonReq,
		PriorState:        map[string]any{"bk_biz_id": 213, "bk_cloud_id": 0, "bk_host_id": 10086},
		Config:            map[string]any{"bk_biz_id": 213},
	})
	assert.NoError(t, err)
	assert.Equal(t, enumor.IaCPluginTaskStatusFailed, execRes.Status)
	assert.Equal(t, "apply err", execRes.Message)
	assert.Equal(t, driftedState, execRes.State)

	// Health + Remaining 7 endpoints + RSA decrypt verification
	hRes, err := exec.Health(kt, routeCtx.Plugin)
	assert.NoError(t, err)
	assert.Equal(t, "1.2.0", hRes.PluginVersion)

	valRes, err := exec.Validate(kt, routeCtx.Plugin, routeCtx.Properties, &proto.ValidateReq{
		CommonInstanceReq: commonReq,
		ConfigType:        proto.ConfigTypeResource,
		Config:            map[string]any{},
	})
	assert.NoError(t, err)
	assert.Len(t, valRes.Diagnostics, 1)

	_, err = exec.GetResult(kt, routeCtx.Plugin, routeCtx.Properties, &proto.GetResultReq{CommonInstanceReq: commonReq, ApplyID: "ap-1"})
	assert.NoError(t, err)
	_, err = exec.Stop(kt, routeCtx.Plugin, routeCtx.Properties, &proto.StopReq{CommonInstanceReq: commonReq, ApplyID: "ap-1"})
	assert.NoError(t, err)
	_, err = exec.Import(kt, routeCtx.Plugin, routeCtx.Properties, &proto.ImportReq{CommonInstanceReq: commonReq, ImportID: "imp-1"})
	assert.NoError(t, err)
	_, err = exec.Refresh(kt, routeCtx.Plugin, routeCtx.Properties, &proto.RefreshReq{CommonInstanceReq: commonReq, State: map[string]any{"id": 1}})
	assert.NoError(t, err)
	_, err = exec.ReadData(kt, routeCtx.Plugin, routeCtx.Properties, &proto.ReadDataReq{CommonInstanceReq: commonReq, Config: map[string]any{}})
	assert.NoError(t, err)
	_, err = exec.List(kt, routeCtx.Plugin, routeCtx.Properties, &proto.ListReq{CommonInstanceReq: commonReq, Limit: 10})
	assert.NoError(t, err)
	assert.Equal(t, "tok_abc", decryptLaunchConfig(t, priv, sb.lastLaunch)["token"])
}

func TestExecutor_GatewayFailureAndRSAPayloadLimit(t *testing.T) {
	_, pubPEM := generateTestRSAKeyPair(t)
	sb := &mockSouthbound{
		applyErr: errors.New("gateway 503"),
	}
	exec := newTestExecutor(sb)
	kt := kit.New()
	commonReq := proto.CommonInstanceReq{Plugin: "bkcmdb", PluginInstanceID: "i1", TypeName: "bkcmdb_host"}

	// AC-012: Southbound gateway failure -> errf.Aborted (2000006)
	routeCtx := buildTestRouteContext(pubPEM, map[string]any{"token": "short"})
	_, err := exec.Execute(kt, routeCtx.Plugin, routeCtx.Schema, routeCtx.Properties, &proto.ExecuteReq{
		CommonInstanceReq: commonReq,
		Config:            map[string]any{"bk_biz_id": 213},
	})
	assert.Error(t, err)
	assert.Equal(t, errf.Aborted, errf.Error(err).Code)

	// RSA payload > keyBytes - 66 (318 bytes) -> errf.InvalidParameter (2000001) and 0 additional southbound calls
	callsBefore := atomic.LoadInt32(&sb.applyCalls)
	oversizedCtx := buildTestRouteContext(pubPEM, map[string]any{"big": strings.Repeat("x", 400)})
	_, err = exec.Execute(kt, oversizedCtx.Plugin, oversizedCtx.Schema, oversizedCtx.Properties, &proto.ExecuteReq{
		CommonInstanceReq: commonReq,
		Config:            map[string]any{"bk_biz_id": 213},
	})
	assert.Error(t, err)
	assert.Equal(t, errf.InvalidParameter, errf.Error(err).Code)
	assert.Equal(t, callsBefore, atomic.LoadInt32(&sb.applyCalls))
}

func TestZeroSecretLeakAnd50GoroutineConcurrency(t *testing.T) {
	const secretMarker = "SK_SUPER_SECRET_12345"
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

	_, pubPEM := generateTestRSAKeyPair(t)
	sb := &mockSouthbound{
		refreshState: map[string]any{"bk_biz_id": 213, "bk_host_id": 10086},
		applyData:    &plugin.ApplyData{Status: enumor.IaCPluginTaskStatusSucceeded, State: map[string]any{"bk_host_id": 10086}},
		planData:     &plugin.PlanData{State: map[string]any{"bk_host_id": 10086}, ReplacePaths: []string{}},
	}
	exec := newTestExecutor(sb)
	routeCtx := buildTestRouteContext(pubPEM, map[string]any{
		"host":       "https://cmdb.example.com",
		"secret_key": secretMarker,
	})

	// 1. 50 goroutines concurrently invoking all runtime endpoints + Health
	run50GoroutinesAllEndpoints(t, exec, routeCtx)

	// 2. Trigger failure and oversized paths to verify secret never leaks in errors or logs
	kt := kit.New()
	sb.mu.Lock()
	sb.applyErr = errors.New("gateway down")
	sb.mu.Unlock()
	_, failErr := exec.Execute(kt, routeCtx.Plugin, routeCtx.Schema, routeCtx.Properties, &proto.ExecuteReq{
		CommonInstanceReq: proto.CommonInstanceReq{Plugin: "bkcmdb", PluginInstanceID: "inst-1", TypeName: "bkcmdb_host"},
		Config:            map[string]any{"bk_biz_id": 213},
	})
	assert.Error(t, failErr)
	assert.NotContains(t, failErr.Error(), secretMarker)

	logs.CloseLogs()
	entries, _ := os.ReadDir(logDir)
	for _, entry := range entries {
		if entry.IsDir() {
			continue
		}
		content, readErr := os.ReadFile(filepath.Join(logDir, entry.Name()))
		assert.NoError(t, readErr)
		assert.NotContains(t, string(content), secretMarker)
	}
}

func run50GoroutinesAllEndpoints(t *testing.T, exec *Service, routeCtx *testPluginFixture) {
	t.Helper()
	const workers = 50
	var wg sync.WaitGroup
	wg.Add(workers)

	commonRes := proto.CommonInstanceReq{Plugin: "bkcmdb", PluginInstanceID: "inst-1", TypeName: "bkcmdb_host"}
	commonDS := proto.CommonInstanceReq{Plugin: "bkcmdb", PluginInstanceID: "inst-1", TypeName: "bkcmdb_biz"}

	for i := 0; i < workers; i++ {
		go func() {
			defer wg.Done()
			kt := kit.New()
			_, err := exec.Health(kt, routeCtx.Plugin)
			assert.NoError(t, err)
			_, err = exec.Execute(kt, routeCtx.Plugin, routeCtx.Schema, routeCtx.Properties, &proto.ExecuteReq{CommonInstanceReq: commonRes, Config: map[string]any{"bk_biz_id": 213}})
			assert.NoError(t, err)
			_, err = exec.Plan(kt, routeCtx.Plugin, routeCtx.Schema, routeCtx.Properties, &proto.PlanReq{CommonInstanceReq: commonRes, Config: map[string]any{"bk_biz_id": 213}})
			assert.NoError(t, err)
			_, err = exec.GetResult(kt, routeCtx.Plugin, routeCtx.Properties, &proto.GetResultReq{CommonInstanceReq: commonRes, ApplyID: "ap-1"})
			assert.NoError(t, err)
			_, err = exec.Stop(kt, routeCtx.Plugin, routeCtx.Properties, &proto.StopReq{CommonInstanceReq: commonRes, ApplyID: "ap-1"})
			assert.NoError(t, err)
			_, err = exec.Validate(kt, routeCtx.Plugin, routeCtx.Properties, &proto.ValidateReq{
				CommonInstanceReq: commonRes,
				ConfigType:        proto.ConfigTypeResource,
				Config:            map[string]any{"bk_biz_id": 213},
			})
			assert.NoError(t, err)
			_, err = exec.Import(kt, routeCtx.Plugin, routeCtx.Properties, &proto.ImportReq{CommonInstanceReq: commonRes, ImportID: "imp-1"})
			assert.NoError(t, err)
			_, err = exec.Refresh(kt, routeCtx.Plugin, routeCtx.Properties, &proto.RefreshReq{CommonInstanceReq: commonRes, State: map[string]any{"bk_host_id": 1}})
			assert.NoError(t, err)
			_, err = exec.ReadData(kt, routeCtx.Plugin, routeCtx.Properties, &proto.ReadDataReq{CommonInstanceReq: commonDS, Config: map[string]any{}})
			assert.NoError(t, err)
			_, err = exec.List(kt, routeCtx.Plugin, routeCtx.Properties, &proto.ListReq{CommonInstanceReq: commonRes, Limit: 10})
			assert.NoError(t, err)
		}()
	}
	wg.Wait()
}
