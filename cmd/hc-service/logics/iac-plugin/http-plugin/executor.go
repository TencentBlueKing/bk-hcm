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
	"encoding/json"
	"strings"

	logicsplugin "hcm/cmd/hc-service/logics/iac-plugin"
	coreplugin "hcm/pkg/api/core/plugin"
	proto "hcm/pkg/api/hc-service/iac-plugin"
	"hcm/pkg/criteria/enumor"
	"hcm/pkg/criteria/errf"
	"hcm/pkg/iacplugin"
	"hcm/pkg/kit"
	"hcm/pkg/logs"
	"hcm/pkg/thirdparty/api-gateway/plugin"
)

// Health probes the HTTP plugin's southbound health endpoint and returns plugin_version.
func (s *Service) Health(kt *kit.Kit, pluginRec *coreplugin.IaCPlugin) (*proto.HealthResult, error) {
	if pluginRec == nil {
		return nil, errf.New(errf.InvalidParameter, "plugin record is nil")
	}

	ext, err := parseExtension(pluginRec)
	if err != nil {
		logs.Errorf("unmarshal http plugin %s@%s extension failed, err: %v, rid: %s",
			pluginRec.PluginName, pluginRec.PluginVersion, err, kt.Rid)
		return nil, errf.Newf(errf.Aborted, "parse http plugin %s extension failed", pluginRec.PluginName)
	}

	cli, err := s.getPluginClient(pluginRec.PluginName, ext.APIPath)
	if err != nil {
		logs.Errorf("create southbound client for plugin %s failed, err: %v, rid: %s",
			pluginRec.PluginName, err, kt.Rid)
		return nil, errf.NewFromErr(errf.Aborted, err)
	}

	health, err := cli.Health(kt)
	if err != nil {
		logs.Errorf("probe http plugin %s health failed, err: %v, rid: %s", pluginRec.PluginName, err, kt.Rid)
		return nil, errf.Newf(errf.Aborted, "probe http plugin %s health failed, err: %v",
			pluginRec.PluginName, err)
	}
	if health == nil || strings.TrimSpace(health.PluginVersion) == "" {
		return nil, errf.Newf(errf.Aborted, "http plugin %s health response has empty plugin_version",
			pluginRec.PluginName)
	}

	return &proto.HealthResult{PluginVersion: health.PluginVersion}, nil
}

// EncryptLaunchConfig parses HTTPExtension and encrypts properties into a Base64 launch_config string.
func (s *Service) EncryptLaunchConfig(kt *kit.Kit, pluginRec *coreplugin.IaCPlugin, properties map[string]any) (
	string, string, error) {

	if pluginRec == nil {
		return "", "", errf.New(errf.InvalidParameter, "plugin record is nil")
	}

	ext, err := parseExtension(pluginRec)
	if err != nil {
		logs.Errorf("unmarshal http plugin %s@%s extension failed, err: %v, rid: %s",
			pluginRec.PluginName, pluginRec.PluginVersion, err, kt.Rid)
		return "", "", errf.Newf(errf.InvalidParameter, "invalid http plugin extension for %s",
			pluginRec.PluginName)
	}

	if err := iacplugin.ValidateAPIPath(ext.APIPath); err != nil {
		return "", "", errf.NewFromErr(errf.InvalidParameter, err)
	}

	encryptor, err := s.getOrCreateEncryptor(ext)
	if err != nil {
		logs.Errorf("init launch_config encryptor for plugin %s@%s failed, err: %v, rid: %s",
			pluginRec.PluginName, pluginRec.PluginVersion, err, kt.Rid)
		return "", "", errf.NewFromErr(errf.InvalidParameter, err)
	}

	launchConfig, err := encryptor.Encrypt(properties)
	if err != nil {
		logs.Errorf("encrypt launch_config for plugin %s@%s failed, err: %v, rid: %s",
			pluginRec.PluginName, pluginRec.PluginVersion, err, kt.Rid)
		return "", "", errf.NewFromErr(errf.InvalidParameter, err)
	}

	return ext.APIPath, launchConfig, nil
}

func (s *Service) getOrCreateEncryptor(ext *coreplugin.HTTPExtension) (plugin.LaunchConfigEncryptor, error) {
	if ext.CryptoConfig == nil {
		return nil, errf.New(errf.InvalidParameter, "crypto_config is nil")
	}

	cacheKey := ext.CryptoType + ":" + ext.CryptoConfig.PublicKey
	if cached, ok := s.encryptorMap.Load(cacheKey); ok {
		return cached.(plugin.LaunchConfigEncryptor), nil
	}

	rawCfg, err := json.Marshal(ext.CryptoConfig)
	if err != nil {
		return nil, err
	}

	encryptor, err := plugin.NewLaunchConfigEncryptor(plugin.CryptoType(ext.CryptoType), rawCfg)
	if err != nil {
		return nil, err
	}

	actual, _ := s.encryptorMap.LoadOrStore(cacheKey, encryptor)
	return actual.(plugin.LaunchConfigEncryptor), nil
}

func (s *Service) prepareCall(kt *kit.Kit, pluginRec *coreplugin.IaCPlugin, properties map[string]any) (
	plugin.Client, string, error) {

	apiPath, launchConfig, err := s.EncryptLaunchConfig(kt, pluginRec, properties)
	if err != nil {
		return nil, "", err
	}

	cli, err := s.getPluginClient(pluginRec.PluginName, apiPath)
	if err != nil {
		logs.Errorf("get southbound client for plugin %s failed, err: %v, rid: %s",
			pluginRec.PluginName, err, kt.Rid)
		return nil, "", errf.NewFromErr(errf.Aborted, err)
	}

	return cli, launchConfig, nil
}

type stateChangePlan struct {
	shortCircuit   bool
	baselinePrior  map[string]any
	refreshedState map[string]any
	proposedState  map[string]any
}

func (s *Service) prepareStateChange(kt *kit.Kit, pluginRec *coreplugin.IaCPlugin, schema *coreplugin.PluginSchema,
	properties map[string]any, typeName string, priorState, config map[string]any, recompute []string) (
	plugin.Client, string, *stateChangePlan, error) {

	if schema == nil || schema.ResourceSchemas == nil {
		return nil, "", nil, errf.Newf(errf.InvalidParameter, "resource schema not found for %s", typeName)
	}
	resSchema := schema.ResourceSchemas[typeName]
	if resSchema == nil || resSchema.Block == nil {
		return nil, "", nil, errf.Newf(errf.InvalidParameter, "resource schema block is nil for %s", typeName)
	}

	if err := logicsplugin.ValidateConfigAndRecompute(resSchema.Block, config, recompute); err != nil {
		return nil, "", nil, err
	}

	cli, launchConfig, err := s.prepareCall(kt, pluginRec, properties)
	if err != nil {
		return nil, "", nil, err
	}

	baselinePrior := priorState
	var refreshedState map[string]any
	if priorState != nil && schema.Capabilities.Refresh {
		refreshData, refErr := cli.Refresh(kt, &plugin.RefreshReq{
			TypeName:     typeName,
			LaunchConfig: launchConfig,
			State:        priorState,
		})
		if refErr != nil {
			logs.Errorf("refresh resource %s before change failed, plugin: %s, err: %v, rid: %s",
				typeName, pluginRec.PluginName, refErr, kt.Rid)
			return nil, "", nil, errf.Newf(errf.Aborted,
				"refresh resource %s before change failed, err: %v", typeName, refErr)
		}

		if refreshData == nil || refreshData.State == nil {
			if config == nil {
				return cli, launchConfig, &stateChangePlan{shortCircuit: true}, nil
			}
			baselinePrior = nil
		}
		if refreshData != nil && refreshData.State != nil {
			baselinePrior = refreshData.State
			refreshedState = refreshData.State
		}
	}

	proposedState, err := logicsplugin.ComposeProposedState(resSchema.Block, baselinePrior, config, recompute)
	if err != nil {
		return nil, "", nil, err
	}

	return cli, launchConfig, &stateChangePlan{
		shortCircuit:   false,
		baselinePrior:  baselinePrior,
		refreshedState: refreshedState,
		proposedState:  proposedState,
	}, nil
}

// Execute runs pre-Refresh drift detection, composes proposed_state, and calls southbound Apply.
func (s *Service) Execute(kt *kit.Kit, pluginRec *coreplugin.IaCPlugin, schema *coreplugin.PluginSchema,
	properties map[string]any, req *proto.ExecuteReq) (*proto.ExecuteResult, error) {

	cli, launchConfig, plan, err := s.prepareStateChange(
		kt, pluginRec, schema, properties, req.TypeName, req.PriorState, req.Config, req.Recompute)
	if err != nil {
		return nil, err
	}
	if plan.shortCircuit {
		return &proto.ExecuteResult{
			Status: enumor.IaCPluginTaskStatusSucceeded,
			State:  nil,
		}, nil
	}

	applyData, err := cli.Apply(kt, &plugin.ApplyReq{
		TypeName:      req.TypeName,
		LaunchConfig:  launchConfig,
		PriorState:    plan.baselinePrior,
		ProposedState: plan.proposedState,
	})
	if err != nil {
		logs.Errorf("apply resource %s failed, plugin: %s, err: %v, rid: %s",
			req.TypeName, pluginRec.PluginName, err, kt.Rid)
		return nil, errf.Newf(errf.Aborted, "apply resource %s failed, err: %v", req.TypeName, err)
	}

	state := applyData.State
	if applyData.Status == enumor.IaCPluginTaskStatusFailed && state == nil && plan.refreshedState != nil {
		state = plan.refreshedState
	}

	return &proto.ExecuteResult{
		Status:   applyData.Status,
		Message:  applyData.Message,
		ApplyID:  applyData.ApplyID,
		State:    state,
		Private:  nil,
		Identity: nil,
		Progress: applyData.Progress,
	}, nil
}

// Plan runs pre-Refresh drift detection, composes proposed_state, and calls southbound Plan.
func (s *Service) Plan(kt *kit.Kit, pluginRec *coreplugin.IaCPlugin, schema *coreplugin.PluginSchema,
	properties map[string]any, req *proto.PlanReq) (*proto.PlanResult, error) {

	cli, launchConfig, plan, err := s.prepareStateChange(
		kt, pluginRec, schema, properties, req.TypeName, req.PriorState, req.Config, req.Recompute)
	if err != nil {
		return nil, err
	}
	if plan.shortCircuit {
		return &proto.PlanResult{
			State:        nil,
			ReplacePaths: []string{},
		}, nil
	}

	planData, err := cli.Plan(kt, &plugin.PlanReq{
		TypeName:      req.TypeName,
		LaunchConfig:  launchConfig,
		PriorState:    plan.baselinePrior,
		ProposedState: plan.proposedState,
	})
	if err != nil {
		logs.Errorf("plan resource %s failed, plugin: %s, err: %v, rid: %s",
			req.TypeName, pluginRec.PluginName, err, kt.Rid)
		return nil, errf.Newf(errf.Aborted, "plan resource %s failed, err: %v", req.TypeName, err)
	}

	return &proto.PlanResult{
		State:        planData.State,
		ReplacePaths: planData.ReplacePaths,
	}, nil
}

// GetResult encrypts launch_config and queries async task result from the southbound plugin.
func (s *Service) GetResult(kt *kit.Kit, pluginRec *coreplugin.IaCPlugin, properties map[string]any,
	req *proto.GetResultReq) (*proto.ExecuteResult, error) {

	cli, launchConfig, err := s.prepareCall(kt, pluginRec, properties)
	if err != nil {
		return nil, err
	}

	data, err := cli.GetResult(kt, &plugin.GetResultReq{
		TypeName:     req.TypeName,
		LaunchConfig: launchConfig,
		ApplyID:      req.ApplyID,
	})
	if err != nil {
		logs.Errorf("get result for resource %s apply_id %s failed, plugin: %s, err: %v, rid: %s",
			req.TypeName, req.ApplyID, pluginRec.PluginName, err, kt.Rid)
		return nil, errf.Newf(errf.Aborted, "get result for resource %s failed, err: %v", req.TypeName, err)
	}

	return &proto.ExecuteResult{
		Status:   data.Status,
		Message:  data.Message,
		ApplyID:  data.ApplyID,
		State:    data.State,
		Private:  nil,
		Identity: nil,
		Progress: data.Progress,
	}, nil
}

// Stop encrypts launch_config and requests stopping an async task on the southbound plugin.
func (s *Service) Stop(kt *kit.Kit, pluginRec *coreplugin.IaCPlugin, properties map[string]any,
	req *proto.StopReq) (*proto.ExecuteResult, error) {

	cli, launchConfig, err := s.prepareCall(kt, pluginRec, properties)
	if err != nil {
		return nil, err
	}

	data, err := cli.Stop(kt, &plugin.StopReq{
		TypeName:     req.TypeName,
		LaunchConfig: launchConfig,
		ApplyID:      req.ApplyID,
	})
	if err != nil {
		logs.Errorf("stop resource %s apply_id %s failed, plugin: %s, err: %v, rid: %s",
			req.TypeName, req.ApplyID, pluginRec.PluginName, err, kt.Rid)
		return nil, errf.Newf(errf.Aborted, "stop resource %s failed, err: %v", req.TypeName, err)
	}

	return &proto.ExecuteResult{
		Status:   data.Status,
		Message:  data.Message,
		ApplyID:  data.ApplyID,
		State:    data.State,
		Private:  nil,
		Identity: nil,
		Progress: data.Progress,
	}, nil
}

// Validate encrypts launch_config and calls southbound Validate.
func (s *Service) Validate(kt *kit.Kit, pluginRec *coreplugin.IaCPlugin, properties map[string]any,
	req *proto.ValidateReq) (*proto.ValidateResult, error) {

	cli, launchConfig, err := s.prepareCall(kt, pluginRec, properties)
	if err != nil {
		return nil, err
	}

	data, err := cli.Validate(kt, &plugin.ValidateReq{
		TypeName:     req.TypeName,
		LaunchConfig: launchConfig,
		Config:       req.Config,
	})
	if err != nil {
		logs.Errorf("validate type %s failed, plugin: %s, err: %v, rid: %s",
			req.TypeName, pluginRec.PluginName, err, kt.Rid)
		return nil, errf.Newf(errf.Aborted, "validate type %s failed, err: %v", req.TypeName, err)
	}

	return &proto.ValidateResult{
		Diagnostics: data.Diagnostics,
	}, nil
}

// Import encrypts launch_config and calls southbound Import.
func (s *Service) Import(kt *kit.Kit, pluginRec *coreplugin.IaCPlugin, properties map[string]any,
	req *proto.ImportReq) (*proto.StateResult, error) {

	cli, launchConfig, err := s.prepareCall(kt, pluginRec, properties)
	if err != nil {
		return nil, err
	}

	data, err := cli.Import(kt, &plugin.ImportReq{
		TypeName:     req.TypeName,
		LaunchConfig: launchConfig,
		ImportID:     req.ImportID,
	})
	if err != nil {
		logs.Errorf("import resource %s import_id %s failed, plugin: %s, err: %v, rid: %s",
			req.TypeName, req.ImportID, pluginRec.PluginName, err, kt.Rid)
		return nil, errf.Newf(errf.Aborted, "import resource %s failed, err: %v", req.TypeName, err)
	}

	return &proto.StateResult{
		State:    data.State,
		Private:  nil,
		Identity: nil,
	}, nil
}

// Refresh encrypts launch_config and calls southbound Refresh.
func (s *Service) Refresh(kt *kit.Kit, pluginRec *coreplugin.IaCPlugin, properties map[string]any,
	req *proto.RefreshReq) (*proto.StateResult, error) {

	cli, launchConfig, err := s.prepareCall(kt, pluginRec, properties)
	if err != nil {
		return nil, err
	}

	data, err := cli.Refresh(kt, &plugin.RefreshReq{
		TypeName:     req.TypeName,
		LaunchConfig: launchConfig,
		State:        req.State,
	})
	if err != nil {
		logs.Errorf("refresh resource %s failed, plugin: %s, err: %v, rid: %s",
			req.TypeName, pluginRec.PluginName, err, kt.Rid)
		return nil, errf.Newf(errf.Aborted, "refresh resource %s failed, err: %v", req.TypeName, err)
	}

	return &proto.StateResult{
		State:    data.State,
		Private:  nil,
		Identity: nil,
	}, nil
}

// ReadData encrypts launch_config and calls southbound ReadData.
func (s *Service) ReadData(kt *kit.Kit, pluginRec *coreplugin.IaCPlugin, properties map[string]any,
	req *proto.ReadDataReq) (*proto.ReadDataResult, error) {

	cli, launchConfig, err := s.prepareCall(kt, pluginRec, properties)
	if err != nil {
		return nil, err
	}

	data, err := cli.ReadData(kt, &plugin.ReadDataReq{
		TypeName:     req.TypeName,
		LaunchConfig: launchConfig,
		Config:       req.Config,
	})
	if err != nil {
		logs.Errorf("read_data %s failed, plugin: %s, err: %v, rid: %s",
			req.TypeName, pluginRec.PluginName, err, kt.Rid)
		return nil, errf.Newf(errf.Aborted, "read_data %s failed, err: %v", req.TypeName, err)
	}

	return &proto.ReadDataResult{
		State: data.State,
	}, nil
}

// List encrypts launch_config and calls southbound List.
func (s *Service) List(kt *kit.Kit, pluginRec *coreplugin.IaCPlugin, properties map[string]any,
	req *proto.ListReq) (*proto.ListResult, error) {

	cli, launchConfig, err := s.prepareCall(kt, pluginRec, properties)
	if err != nil {
		return nil, err
	}

	data, err := cli.List(kt, &plugin.ListReq{
		TypeName:     req.TypeName,
		LaunchConfig: launchConfig,
		Config:       req.Config,
		Limit:        req.Limit,
		PageToken:    req.PageToken,
	})
	if err != nil {
		logs.Errorf("list resource %s failed, plugin: %s, err: %v, rid: %s",
			req.TypeName, pluginRec.PluginName, err, kt.Rid)
		return nil, errf.Newf(errf.Aborted, "list resource %s failed, err: %v", req.TypeName, err)
	}

	return data, nil
}
