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

package router

import (
	httpplugin "hcm/cmd/hc-service/logics/iac-plugin/http-plugin"
	"hcm/cmd/hc-service/logics/iac-plugin/store"
	"hcm/cmd/hc-service/logics/iac-plugin/tf-plugin/adaptor"
	"hcm/cmd/hc-service/logics/iac-plugin/tf-plugin/discovery"
	"hcm/cmd/hc-service/logics/iac-plugin/tf-plugin/pool"
	plugininst "hcm/cmd/hc-service/logics/plugin/instance"
	pluginschema "hcm/cmd/hc-service/logics/plugin/schema"
	coreplugin "hcm/pkg/api/core/plugin"
	proto "hcm/pkg/api/hc-service/iac-plugin"
	dataservice "hcm/pkg/client/data-service"
	"hcm/pkg/criteria/enumor"
	"hcm/pkg/criteria/errf"
	"hcm/pkg/iacplugin"
	"hcm/pkg/kit"
	"hcm/pkg/logs"
)

// Service implements the northbound plugin router for hc-service.
type Service struct {
	dataCli      *dataservice.Client
	schemaMgr    *pluginschema.SchemaManager
	instanceMgr  *plugininst.PluginInstanceManager
	httpPlugin   *httpplugin.Service
	poolRegistry *pool.PluginRegistry
	// plugins 启动时扫描到的本副本 tf 插件（以插件名为 key），之后只读。
	plugins map[string]discovery.Entry
}

// New creates a northbound Router Service instance.
func New(dataCli *dataservice.Client, schemaMgr *pluginschema.SchemaManager,
	httpPlugin *httpplugin.Service, poolRegistry *pool.PluginRegistry,
	plugins map[string]discovery.Entry) *Service {

	return &Service{
		dataCli:      dataCli,
		schemaMgr:    schemaMgr,
		instanceMgr:  plugininst.NewPluginInstanceManager(dataCli, schemaMgr),
		httpPlugin:   httpPlugin,
		poolRegistry: poolRegistry,
		plugins:      plugins,
	}
}

func (s *Service) acquireTFAdaptor(kt *kit.Kit, routeCtx *RouteContext,
	instanceID string) (*adaptor.TFPluginAdaptor, error) {

	if s.poolRegistry == nil {
		return nil, errf.Newf(errf.Aborted,
			"tf_provider plugin %s process pool registry is not initialized", routeCtx.Plugin.PluginName)
	}
	pluginPool, err := s.poolRegistry.GetPool(routeCtx.Plugin.PluginName)
	if err != nil {
		logs.Errorf("get process pool for plugin %s failed, err: %v, rid: %s",
			routeCtx.Plugin.PluginName, err, kt.Rid)
		return nil, errf.Newf(errf.Aborted,
			"get process pool for plugin %s failed: %v", routeCtx.Plugin.PluginName, err)
	}
	return adaptor.NewTFPluginAdaptor(kt, pluginPool, routeCtx.Schema, instanceID, routeCtx.Properties)
}

// Resolve validates plugin enabled state, active schema type_name and capabilities, and instance ownership
// to construct a RouteContext.
func (s *Service) Resolve(kt *kit.Kit, commonReq proto.CommonInstanceReq,
	op enumor.IaCPluginOperationType) (*RouteContext, error) {

	return s.resolveWithConfigType(kt, commonReq, op, "")
}

// ResolveValidate 针对 Validate 请求，额外按 req.ConfigType（"resource" 或 "data"）精确校验目标 Schema。
func (s *Service) ResolveValidate(kt *kit.Kit, req *proto.ValidateReq) (*RouteContext, error) {
	return s.resolveWithConfigType(kt, req.CommonInstanceReq, enumor.IaCPluginOperationValidate, req.ConfigType)
}

func (s *Service) resolveWithConfigType(kt *kit.Kit, commonReq proto.CommonInstanceReq,
	op enumor.IaCPluginOperationType, configType string) (*RouteContext, error) {

	targetPlugin, err := s.getEnabledTargetPlugin(kt, commonReq.Plugin)
	if err != nil {
		return nil, err
	}

	schema, err := s.schemaMgr.GetSchema(kt, coreplugin.PluginInfo{
		Name:    commonReq.Plugin,
		Version: targetPlugin.PluginVersion,
	})
	if err != nil {
		logs.Errorf("get active schema for plugin %s@%s failed, err: %v, rid: %s",
			commonReq.Plugin, targetPlugin.PluginVersion, err, kt.Rid)
		return nil, err
	}

	err = validateSchemaTypeAndCapability(
		schema, commonReq.Plugin, targetPlugin.PluginVersion, commonReq.TypeName, op, configType)
	if err != nil {
		logs.Errorf("validate schema type or capability failed, plugin: %s@%s, type_name: %s, op: %s, err: %v, rid: %s",
			commonReq.Plugin, targetPlugin.PluginVersion, commonReq.TypeName, op, err, kt.Rid)
		return nil, err
	}

	props, err := s.getInstanceProperties(kt, commonReq.Plugin, commonReq.PluginInstanceID)
	if err != nil {
		return nil, err
	}

	return &RouteContext{
		Plugin:     targetPlugin,
		Schema:     schema,
		Properties: props,
	}, nil
}

// getTargetPlugin 解析目标插件版本记录：
//   - 若为本地 tf_provider 插件（命中本副本 plugins 或数据库记录为 tf_provider），严格按本副本 PluginVersion 匹配记录，
//     防止滚动升级新旧副本并存期间用新版本 Schema 操作旧版本二进制；
//   - 若为三方 http 插件，选取最高版本记录。
func (s *Service) getTargetPlugin(kt *kit.Kit, pluginName string) (*coreplugin.IaCPlugin, error) {
	records, err := store.List(kt, s.dataCli, pluginName)
	if err != nil {
		logs.Errorf("list iac plugin %s failed, err: %v, rid: %s", pluginName, err, kt.Rid)
		return nil, err
	}
	if len(records) == 0 {
		return nil, errf.Newf(errf.RecordNotFound, "iac plugin %s not found", pluginName)
	}

	// 检查本地插件
	local, ok, err := s.getLocalPlugin(pluginName, records)
	if err != nil {
		return nil, err
	}
	if ok {
		return local, nil
	}

	// 非本地插件，查询最新的三方http插件记录
	latest, err := iacplugin.Latest(records)
	if err != nil {
		logs.Errorf("select latest iac plugin %s failed, err: %v, rid: %s", pluginName, err, kt.Rid)
		return nil, errf.NewFromErr(errf.Aborted, err)
	}
	if latest == nil {
		return nil, errf.Newf(errf.RecordNotFound, "iac plugin %s not found", pluginName)
	}
	return latest, nil
}

// getLocalPlugin 按本副本扫描到的插件版本，从 records 中取出本地 tf_provider 记录。
// 本副本 plugins 命中，或 records 中存在 tf_provider 协议，都按本地插件锁定版本：
// 数据库记录为 tf_provider 但本副本没有对应二进制时返回 Aborted；
// 本副本版本在 records 中不存在时返回 RecordNotFound。
// 两者都不满足时返回 nil, false, nil，调用方按三方插件选取最高版本。
func (s *Service) getLocalPlugin(pluginName string, records []coreplugin.IaCPlugin) (
	*coreplugin.IaCPlugin, bool, error) {

	entry, isLocal := s.plugins[pluginName]
	if isLocal {
		for i := range records {
			if records[i].PluginVersion == entry.PluginVersion {
				return &records[i], true, nil
			}
		}
		return nil, false, errf.Newf(errf.RecordNotFound,
			"iac plugin %s@%s not found", pluginName, entry.PluginVersion)
	}

	for i := range records {
		if records[i].Protocol == enumor.IaCPluginProtocolTFProvider {
			return nil, false, errf.Newf(errf.Aborted,
				"tf_provider plugin %s binary not found on this replica", pluginName)
		}
	}
	return nil, false, nil
}

func (s *Service) getEnabledTargetPlugin(kt *kit.Kit, pluginName string) (*coreplugin.IaCPlugin, error) {
	target, err := s.getTargetPlugin(kt, pluginName)
	if err != nil {
		return nil, err
	}
	if target.State != enumor.IaCPluginStateEnabled {
		return nil, errf.Newf(errf.Aborted, "iac plugin %s@%s is not enabled, current state: %s",
			pluginName, target.PluginVersion, target.State)
	}
	return target, nil
}

func validateSchemaTypeAndCapability(schema *coreplugin.PluginSchema, pluginName, version, typeName string,
	op enumor.IaCPluginOperationType, configType string) error {

	if schema == nil {
		return errf.Newf(errf.RecordNotFound, "plugin schema for %s@%s is nil", pluginName, version)
	}

	if err := checkTypeInSchema(schema, pluginName, version, typeName, op, configType); err != nil {
		return err
	}

	if !checkCapability(schema.Capabilities, op) {
		return errf.Newf(errf.Aborted, "plugin %s@%s does not support capability for operation %s",
			pluginName, version, op)
	}
	return nil
}

func checkTypeInSchema(schema *coreplugin.PluginSchema, pluginName, version, typeName string,
	op enumor.IaCPluginOperationType, configType string) error {

	inRes := schema.ResourceSchemas != nil && schema.ResourceSchemas[typeName] != nil
	inDS := schema.DataSourceSchemas != nil && schema.DataSourceSchemas[typeName] != nil

	switch op {
	case enumor.IaCPluginOperationReadData:
		if !inDS {
			return errf.Newf(errf.InvalidParameter, "data source type_name %s not found in plugin %s@%s schema",
				typeName, pluginName, version)
		}
	case enumor.IaCPluginOperationValidate:
		switch configType {
		case proto.ConfigTypeResource:
			if !inRes {
				return errf.Newf(errf.InvalidParameter, "resource type_name %s not found in plugin %s@%s schema",
					typeName, pluginName, version)
			}
		case proto.ConfigTypeDataSource:
			if !inDS {
				return errf.Newf(errf.InvalidParameter, "data source type_name %s not found in plugin %s@%s schema",
					typeName, pluginName, version)
			}
		case "":
			if !inRes && !inDS {
				return errf.Newf(errf.InvalidParameter, "type_name %s not found in plugin %s@%s schema",
					typeName, pluginName, version)
			}
		default:
			return errf.Newf(errf.InvalidParameter, "invalid config_type %q for validate", configType)
		}
	default:
		if !inRes {
			return errf.Newf(errf.InvalidParameter, "resource type_name %s not found in plugin %s@%s schema",
				typeName, pluginName, version)
		}
	}
	return nil
}

func checkCapability(cap coreplugin.Capabilities, op enumor.IaCPluginOperationType) bool {
	switch op {
	case enumor.IaCPluginOperationExecute:
		return cap.Apply
	case enumor.IaCPluginOperationGetResult:
		return cap.AsyncApply
	case enumor.IaCPluginOperationStop:
		return cap.Stop && cap.AsyncApply
	case enumor.IaCPluginOperationPlan:
		return cap.Plan
	case enumor.IaCPluginOperationValidate:
		return cap.Validate
	case enumor.IaCPluginOperationImport:
		return cap.Import
	case enumor.IaCPluginOperationRefresh:
		return cap.Refresh
	case enumor.IaCPluginOperationReadData:
		return cap.ReadData
	case enumor.IaCPluginOperationList:
		return cap.List
	default:
		return false
	}
}

func (s *Service) getInstanceProperties(kt *kit.Kit, pluginName, instanceID string) (map[string]any, error) {
	inst, err := s.instanceMgr.GetPluginInstance(kt, instanceID)
	if err != nil {
		return nil, err
	}

	if inst.PluginName != pluginName {
		logs.Errorf("plugin instance %s belongs to plugin %s, cannot be used by plugin %s, rid: %s",
			instanceID, inst.PluginName, pluginName, kt.Rid)
		return nil, errf.Newf(errf.InvalidParameter,
			"plugin instance %s belongs to plugin %s, cannot be used by plugin %s",
			instanceID, inst.PluginName, pluginName)
	}

	return inst.Properties, nil
}

// GetSchema reads plugin schema from SchemaManager (in-memory SchemaCache + DB) with zero southbound HTTP calls.
func (s *Service) GetSchema(kt *kit.Kit, req *proto.GetSchemaReq) (*coreplugin.PluginSchema, error) {
	version := req.PluginVersion
	if version == "" {
		target, err := s.getEnabledTargetPlugin(kt, req.Plugin)
		if err != nil {
			return nil, err
		}
		version = target.PluginVersion
	}

	return s.schemaMgr.GetSchema(kt, coreplugin.PluginInfo{
		Name:    req.Plugin,
		Version: version,
	})
}

// Health probes the plugin's southbound health endpoint and returns plugin_version.
func (s *Service) Health(kt *kit.Kit, req *proto.HealthReq) (*proto.HealthResult, error) {
	target, err := s.getTargetPlugin(kt, req.Plugin)
	if err != nil {
		return nil, err
	}

	switch target.Protocol {
	case enumor.IaCPluginProtocolHTTP:
		return s.httpPlugin.Health(kt, target)
	case enumor.IaCPluginProtocolTFProvider:
		if target.State != enumor.IaCPluginStateEnabled {
			return nil, errf.Newf(errf.Aborted, "iac plugin %s@%s is not enabled, current state: %s",
				req.Plugin, target.PluginVersion, target.State)
		}
		return &proto.HealthResult{PluginVersion: target.PluginVersion}, nil
	default:
		return nil, errf.Newf(errf.Aborted, "unsupported iac plugin protocol: %s", target.Protocol)
	}
}

// Execute resolves the route context and dispatches Execute by plugin protocol.
func (s *Service) Execute(kt *kit.Kit, req *proto.ExecuteReq) (*proto.ExecuteResult, error) {
	routeCtx, err := s.Resolve(kt, req.CommonInstanceReq, enumor.IaCPluginOperationExecute)
	if err != nil {
		return nil, err
	}

	switch routeCtx.Plugin.Protocol {
	case enumor.IaCPluginProtocolHTTP:
		return s.httpPlugin.Execute(kt, routeCtx.Plugin, routeCtx.Schema, routeCtx.Properties, req)
	case enumor.IaCPluginProtocolTFProvider:
		tfAdaptor, err := s.acquireTFAdaptor(kt, routeCtx, req.PluginInstanceID)
		if err != nil {
			return nil, err
		}
		defer tfAdaptor.Close()
		return tfAdaptor.Execute(kt, req)
	default:
		return nil, errf.Newf(errf.Aborted, "unsupported iac plugin protocol: %s", routeCtx.Plugin.Protocol)
	}
}

// Plan resolves the route context and dispatches Plan by plugin protocol.
func (s *Service) Plan(kt *kit.Kit, req *proto.PlanReq) (*proto.PlanResult, error) {
	routeCtx, err := s.Resolve(kt, req.CommonInstanceReq, enumor.IaCPluginOperationPlan)
	if err != nil {
		return nil, err
	}

	switch routeCtx.Plugin.Protocol {
	case enumor.IaCPluginProtocolHTTP:
		return s.httpPlugin.Plan(kt, routeCtx.Plugin, routeCtx.Schema, routeCtx.Properties, req)
	case enumor.IaCPluginProtocolTFProvider:
		tfAdaptor, err := s.acquireTFAdaptor(kt, routeCtx, req.PluginInstanceID)
		if err != nil {
			return nil, err
		}
		defer tfAdaptor.Close()
		return tfAdaptor.Plan(kt, req)
	default:
		return nil, errf.Newf(errf.Aborted, "unsupported iac plugin protocol: %s", routeCtx.Plugin.Protocol)
	}
}

// GetResult resolves the route context and dispatches GetResult by plugin protocol.
func (s *Service) GetResult(kt *kit.Kit, req *proto.GetResultReq) (*proto.ExecuteResult, error) {
	routeCtx, err := s.Resolve(kt, req.CommonInstanceReq, enumor.IaCPluginOperationGetResult)
	if err != nil {
		return nil, err
	}

	switch routeCtx.Plugin.Protocol {
	case enumor.IaCPluginProtocolHTTP:
		return s.httpPlugin.GetResult(kt, routeCtx.Plugin, routeCtx.Properties, req)
	case enumor.IaCPluginProtocolTFProvider:
		return nil, errf.Newf(errf.Aborted, "plugin %s@%s protocol %s does not support async result",
			routeCtx.Plugin.PluginName, routeCtx.Plugin.PluginVersion, routeCtx.Plugin.Protocol)
	default:
		return nil, errf.Newf(errf.Aborted, "unsupported iac plugin protocol: %s", routeCtx.Plugin.Protocol)
	}
}

// Stop resolves the route context and dispatches Stop by plugin protocol.
func (s *Service) Stop(kt *kit.Kit, req *proto.StopReq) (*proto.ExecuteResult, error) {
	routeCtx, err := s.Resolve(kt, req.CommonInstanceReq, enumor.IaCPluginOperationStop)
	if err != nil {
		return nil, err
	}

	switch routeCtx.Plugin.Protocol {
	case enumor.IaCPluginProtocolHTTP:
		return s.httpPlugin.Stop(kt, routeCtx.Plugin, routeCtx.Properties, req)
	case enumor.IaCPluginProtocolTFProvider:
		return nil, errf.Newf(errf.Aborted, "plugin %s@%s protocol %s does not support async stop",
			routeCtx.Plugin.PluginName, routeCtx.Plugin.PluginVersion, routeCtx.Plugin.Protocol)
	default:
		return nil, errf.Newf(errf.Aborted, "unsupported iac plugin protocol: %s", routeCtx.Plugin.Protocol)
	}
}

// Validate resolves the route context and dispatches Validate by plugin protocol.
func (s *Service) Validate(kt *kit.Kit, req *proto.ValidateReq) (*proto.ValidateResult, error) {
	routeCtx, err := s.ResolveValidate(kt, req)
	if err != nil {
		return nil, err
	}

	switch routeCtx.Plugin.Protocol {
	case enumor.IaCPluginProtocolHTTP:
		return s.httpPlugin.Validate(kt, routeCtx.Plugin, routeCtx.Properties, req)
	case enumor.IaCPluginProtocolTFProvider:
		tfAdaptor, err := s.acquireTFAdaptor(kt, routeCtx, req.PluginInstanceID)
		if err != nil {
			return nil, err
		}
		defer tfAdaptor.Close()
		return tfAdaptor.Validate(kt, req)
	default:
		return nil, errf.Newf(errf.Aborted, "unsupported iac plugin protocol: %s", routeCtx.Plugin.Protocol)
	}
}

// Import resolves the route context and dispatches Import by plugin protocol.
func (s *Service) Import(kt *kit.Kit, req *proto.ImportReq) (*proto.StateResult, error) {
	routeCtx, err := s.Resolve(kt, req.CommonInstanceReq, enumor.IaCPluginOperationImport)
	if err != nil {
		return nil, err
	}

	switch routeCtx.Plugin.Protocol {
	case enumor.IaCPluginProtocolHTTP:
		return s.httpPlugin.Import(kt, routeCtx.Plugin, routeCtx.Properties, req)
	case enumor.IaCPluginProtocolTFProvider:
		tfAdaptor, err := s.acquireTFAdaptor(kt, routeCtx, req.PluginInstanceID)
		if err != nil {
			return nil, err
		}
		defer tfAdaptor.Close()
		return tfAdaptor.Import(kt, req)
	default:
		return nil, errf.Newf(errf.Aborted, "unsupported iac plugin protocol: %s", routeCtx.Plugin.Protocol)
	}
}

// Refresh resolves the route context and dispatches Refresh by plugin protocol.
func (s *Service) Refresh(kt *kit.Kit, req *proto.RefreshReq) (*proto.StateResult, error) {
	routeCtx, err := s.Resolve(kt, req.CommonInstanceReq, enumor.IaCPluginOperationRefresh)
	if err != nil {
		return nil, err
	}

	switch routeCtx.Plugin.Protocol {
	case enumor.IaCPluginProtocolHTTP:
		return s.httpPlugin.Refresh(kt, routeCtx.Plugin, routeCtx.Properties, req)
	case enumor.IaCPluginProtocolTFProvider:
		tfAdaptor, err := s.acquireTFAdaptor(kt, routeCtx, req.PluginInstanceID)
		if err != nil {
			return nil, err
		}
		defer tfAdaptor.Close()
		return tfAdaptor.Refresh(kt, req)
	default:
		return nil, errf.Newf(errf.Aborted, "unsupported iac plugin protocol: %s", routeCtx.Plugin.Protocol)
	}
}

// ReadData resolves the route context and dispatches ReadData by plugin protocol.
func (s *Service) ReadData(kt *kit.Kit, req *proto.ReadDataReq) (*proto.ReadDataResult, error) {
	routeCtx, err := s.Resolve(kt, req.CommonInstanceReq, enumor.IaCPluginOperationReadData)
	if err != nil {
		return nil, err
	}

	switch routeCtx.Plugin.Protocol {
	case enumor.IaCPluginProtocolHTTP:
		return s.httpPlugin.ReadData(kt, routeCtx.Plugin, routeCtx.Properties, req)
	case enumor.IaCPluginProtocolTFProvider:
		tfAdaptor, err := s.acquireTFAdaptor(kt, routeCtx, req.PluginInstanceID)
		if err != nil {
			return nil, err
		}
		defer tfAdaptor.Close()
		return tfAdaptor.ReadData(kt, req)
	default:
		return nil, errf.Newf(errf.Aborted, "unsupported iac plugin protocol: %s", routeCtx.Plugin.Protocol)
	}
}

// List resolves the route context and dispatches List by plugin protocol.
func (s *Service) List(kt *kit.Kit, req *proto.ListReq) (*proto.ListResult, error) {
	routeCtx, err := s.Resolve(kt, req.CommonInstanceReq, enumor.IaCPluginOperationList)
	if err != nil {
		return nil, err
	}

	switch routeCtx.Plugin.Protocol {
	case enumor.IaCPluginProtocolHTTP:
		return s.httpPlugin.List(kt, routeCtx.Plugin, routeCtx.Properties, req)
	case enumor.IaCPluginProtocolTFProvider:
		return nil, errf.Newf(errf.Aborted, "plugin %s@%s protocol %s does not support list",
			routeCtx.Plugin.PluginName, routeCtx.Plugin.PluginVersion, routeCtx.Plugin.Protocol)
	default:
		return nil, errf.Newf(errf.Aborted, "unsupported iac plugin protocol: %s", routeCtx.Plugin.Protocol)
	}
}
