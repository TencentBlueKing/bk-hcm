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

package instance

import (
	"hcm/cmd/hc-service/logics/plugin/schema"
	"hcm/pkg/api/core"
	coreplugin "hcm/pkg/api/core/plugin"
	dataplugin "hcm/pkg/api/data-service/plugin"
	hcplugin "hcm/pkg/api/hc-service/plugin"
	dataservice "hcm/pkg/client/data-service"
	"hcm/pkg/criteria/enumor"
	"hcm/pkg/criteria/errf"
	"hcm/pkg/dal/dao/tools"
	"hcm/pkg/kit"
	"hcm/pkg/logs"
)

// PluginInstanceManager hc-service 侧插件实例业务管理器。
type PluginInstanceManager struct {
	dataCli   *dataservice.Client
	schemaMgr *schema.SchemaManager
}

// NewPluginInstanceManager 创建 PluginInstanceManager。
func NewPluginInstanceManager(dataCli *dataservice.Client, schemaMgr *schema.SchemaManager) *PluginInstanceManager {
	return &PluginInstanceManager{
		dataCli:   dataCli,
		schemaMgr: schemaMgr,
	}
}

// CreatePluginInstance 校验活跃插件版本 Schema、拆分 properties 后在 data-service 创建插件实例。
func (m *PluginInstanceManager) CreatePluginInstance(kt *kit.Kit, req *hcplugin.PluginInstanceCreateReq) (
	*core.CreateResult, error) {

	if req == nil {
		return nil, errf.New(errf.InvalidParameter, "create plugin instance request is nil")
	}
	if err := req.Validate(); err != nil {
		logs.Errorf("validate create plugin instance request failed, err: %v, rid: %s", err, kt.Rid)
		return nil, errf.NewFromErr(errf.InvalidParameter, err)
	}
	if err := m.checkDependencies(); err != nil {
		logs.Errorf("create plugin instance check dependencies failed, err: %v, rid: %s", err, kt.Rid)
		return nil, err
	}

	pluginSchema, err := m.getActivePluginSchema(kt, req.PluginName)
	if err != nil {
		return nil, err
	}

	attrs, secretAttrs, err := splitAndValidateProperties(pluginSchema.ConfigSchema.Block, req.Properties)
	if err != nil {
		logs.Errorf("split and validate plugin instance properties failed, plugin: %s, err: %v, rid: %s",
			req.PluginName, err, kt.Rid)
		return nil, err
	}

	dsReq := &dataplugin.IacPluginInstanceBatchCreateReq{
		Items: []dataplugin.IacPluginInstanceCreateField{
			{
				PluginName:  req.PluginName,
				Name:        req.Name,
				Attrs:       attrs,
				SecretAttrs: secretAttrs,
				Extension:   req.Extension,
				Memo:        req.Memo,
			},
		},
	}

	res, err := m.dataCli.Global.IacPluginInstance.BatchCreate(kt, dsReq)
	if err != nil {
		logs.Errorf("create iac_plugin_instance in data-service failed, plugin: %s, name: %s, err: %v, rid: %s",
			req.PluginName, req.Name, err, kt.Rid)
		return nil, err
	}
	if res == nil || len(res.IDs) == 0 {
		return nil, errf.New(errf.Aborted, "create iac_plugin_instance returned empty ids")
	}

	return &core.CreateResult{ID: res.IDs[0]}, nil
}

// UpdatePluginInstance 更新插件实例；若携带 properties 则按全量覆盖语义校验并更新，否则仅更新元数据字段。
func (m *PluginInstanceManager) UpdatePluginInstance(kt *kit.Kit, id string, req *hcplugin.PluginInstanceUpdateReq) error {
	if len(id) == 0 {
		return errf.New(errf.InvalidParameter, "id is required")
	}
	if req == nil {
		return errf.New(errf.InvalidParameter, "update plugin instance request is nil")
	}
	if err := req.Validate(); err != nil {
		logs.Errorf("validate update plugin instance request failed, id: %s, err: %v, rid: %s", id, err, kt.Rid)
		return errf.NewFromErr(errf.InvalidParameter, err)
	}
	if err := m.checkDependencies(); err != nil {
		logs.Errorf("update plugin instance check dependencies failed, id: %s, err: %v, rid: %s", id, err, kt.Rid)
		return err
	}

	updateField := dataplugin.IacPluginInstanceUpdateField{
		ID:        id,
		Name:      req.Name,
		Extension: req.Extension,
		Memo:      req.Memo,
	}

	if req.Properties != nil {
		if err := m.populateUpdatedProperties(kt, id, req.Properties, &updateField); err != nil {
			return err
		}
	}

	dsReq := &dataplugin.IacPluginInstanceBatchUpdateReq{
		Items: []dataplugin.IacPluginInstanceUpdateField{updateField},
	}

	if err := m.dataCli.Global.IacPluginInstance.BatchUpdate(kt, dsReq); err != nil {
		logs.Errorf("update iac_plugin_instance in data-service failed, id: %s, err: %v, rid: %s", id, err, kt.Rid)
		return err
	}

	return nil
}

func (m *PluginInstanceManager) populateUpdatedProperties(kt *kit.Kit, id string, properties map[string]any,
	updateField *dataplugin.IacPluginInstanceUpdateField) error {

	existing, err := m.getRawPluginInstanceByID(kt, id)
	if err != nil {
		logs.Errorf("get existing plugin instance before update failed, id: %s, err: %v, rid: %s", id, err, kt.Rid)
		return err
	}

	pluginSchema, err := m.getActivePluginSchema(kt, existing.PluginName)
	if err != nil {
		return err
	}

	attrs, secretAttrs, err := splitAndValidateProperties(pluginSchema.ConfigSchema.Block, properties)
	if err != nil {
		logs.Errorf("split and validate updated properties failed, id: %s, plugin: %s, err: %v, rid: %s",
			id, existing.PluginName, err, kt.Rid)
		return err
	}

	updateField.Attrs = attrs
	updateField.SecretAttrs = secretAttrs
	return nil
}

// DeletePluginInstance 按 ID 删除插件实例。
func (m *PluginInstanceManager) DeletePluginInstance(kt *kit.Kit, id string) error {
	if len(id) == 0 {
		return errf.New(errf.InvalidParameter, "id is required")
	}
	if err := m.checkDependencies(); err != nil {
		logs.Errorf("delete plugin instance check dependencies failed, id: %s, err: %v, rid: %s", id, err, kt.Rid)
		return err
	}

	dsReq := &dataplugin.IacPluginInstanceBatchDeleteReq{
		Filter: tools.EqualExpression("id", id),
	}

	if err := m.dataCli.Global.IacPluginInstance.BatchDelete(kt, dsReq); err != nil {
		logs.Errorf("delete iac_plugin_instance in data-service failed, id: %s, err: %v, rid: %s", id, err, kt.Rid)
		return err
	}

	return nil
}

// GetPluginInstance 获取单条插件实例并将 attrs 与 secret_attrs 合并为 properties。
func (m *PluginInstanceManager) GetPluginInstance(kt *kit.Kit, id string) (*coreplugin.PluginInstance, error) {
	if len(id) == 0 {
		return nil, errf.New(errf.InvalidParameter, "id is required")
	}
	if err := m.checkDependencies(); err != nil {
		logs.Errorf("get plugin instance check dependencies failed, id: %s, err: %v, rid: %s", id, err, kt.Rid)
		return nil, err
	}

	raw, err := m.getRawPluginInstanceByID(kt, id)
	if err != nil {
		logs.Errorf("get iac_plugin_instance from data-service failed, id: %s, err: %v, rid: %s", id, err, kt.Rid)
		return nil, err
	}

	return toCorePluginInstance(raw), nil
}

func (m *PluginInstanceManager) getRawPluginInstanceByID(kt *kit.Kit, id string) (
	*coreplugin.IacPluginInstance, error) {

	listReq := &dataplugin.IacPluginInstanceListReq{
		Filter: tools.EqualExpression("id", id),
		Page: &core.BasePage{
			Start: 0,
			Limit: 1,
		},
	}

	res, err := m.dataCli.Global.IacPluginInstance.List(kt, listReq)
	if err != nil {
		return nil, err
	}
	if res == nil || len(res.Details) == 0 {
		return nil, errf.Newf(errf.RecordNotFound, "iac_plugin_instance %s not found", id)
	}

	return &res.Details[0], nil
}

// ListPluginInstance 列表查询插件实例并将每条记录的 attrs 与 secret_attrs 合并为 properties。
func (m *PluginInstanceManager) ListPluginInstance(kt *kit.Kit, req *hcplugin.PluginInstanceListReq) (
	*hcplugin.PluginInstanceListResult, error) {

	if req == nil {
		return nil, errf.New(errf.InvalidParameter, "list plugin instance request is nil")
	}
	if err := req.Validate(); err != nil {
		logs.Errorf("validate list plugin instance request failed, err: %v, rid: %s", err, kt.Rid)
		return nil, errf.NewFromErr(errf.InvalidParameter, err)
	}
	if err := m.checkDependencies(); err != nil {
		logs.Errorf("list plugin instance check dependencies failed, err: %v, rid: %s", err, kt.Rid)
		return nil, err
	}

	dsReq := &dataplugin.IacPluginInstanceListReq{
		Fields: normalizeInstanceListFields(req.Fields),
		Filter: req.Filter,
		Page:   req.Page,
	}

	rawRes, err := m.dataCli.Global.IacPluginInstance.List(kt, dsReq)
	if err != nil {
		logs.Errorf("list iac_plugin_instance from data-service failed, err: %v, rid: %s", err, kt.Rid)
		return nil, err
	}

	details := make([]coreplugin.PluginInstance, 0, len(rawRes.Details))
	for i := range rawRes.Details {
		details = append(details, *toCorePluginInstance(&rawRes.Details[i]))
	}

	return &hcplugin.PluginInstanceListResult{
		Count:   rawRes.Count,
		Details: details,
	}, nil
}

// getActivePluginSchema 查询当前处于 enabled 状态的活跃插件版本并加载其 Provider Schema。
func (m *PluginInstanceManager) getActivePluginSchema(kt *kit.Kit, pluginName string) (
	*coreplugin.PluginSchema, error) {

	listReq := &core.ListReq{
		Filter: tools.ExpressionAnd(
			tools.RuleEqual("plugin_name", pluginName),
			tools.RuleEqual("state", enumor.IaCPluginStateEnabled),
		),
		Page: &core.BasePage{
			Start: 0,
			Limit: 1,
		},
	}

	res, err := m.dataCli.Global.IaCPlugin.List(kt, listReq)
	if err != nil {
		logs.Errorf("query active plugin from iac_plugin failed, plugin: %s, err: %v, rid: %s",
			pluginName, err, kt.Rid)
		return nil, err
	}
	if res == nil || len(res.Details) == 0 {
		logs.Errorf("active plugin with state=%s not found, plugin: %s, rid: %s",
			enumor.IaCPluginStateEnabled, pluginName, kt.Rid)
		return nil, errf.Newf(errf.InvalidParameter,
			"active plugin %s with state=%s not found", pluginName, enumor.IaCPluginStateEnabled)
	}

	activePlugin := res.Details[0]
	pluginInfo := coreplugin.PluginInfo{
		Name:    activePlugin.PluginName,
		Version: activePlugin.PluginVersion,
	}
	pluginSchema, err := m.schemaMgr.GetSchema(kt, pluginInfo)
	if err != nil {
		logs.Errorf("get plugin schema for active plugin failed, plugin: %s@%s, err: %v, rid: %s",
			activePlugin.PluginName, activePlugin.PluginVersion, err, kt.Rid)
		return nil, err
	}
	if pluginSchema == nil || pluginSchema.ConfigSchema == nil || pluginSchema.ConfigSchema.Block == nil {
		logs.Errorf("provider schema block is empty for active plugin, plugin: %s@%s, rid: %s",
			activePlugin.PluginName, activePlugin.PluginVersion, kt.Rid)
		return nil, errf.Newf(errf.InvalidParameter,
			"provider schema block is empty for plugin %s@%s", activePlugin.PluginName, activePlugin.PluginVersion)
	}

	return pluginSchema, nil
}

func (m *PluginInstanceManager) checkDependencies() error {
	if m == nil || m.dataCli == nil || m.dataCli.Global == nil ||
		m.dataCli.Global.IacPluginInstance == nil || m.dataCli.Global.IaCPlugin == nil {
		return errf.New(errf.InvalidParameter, "data-service plugin instance client is not initialized")
	}
	if m.schemaMgr == nil {
		return errf.New(errf.InvalidParameter, "plugin schema manager is not initialized")
	}
	return nil
}

func toCorePluginInstance(raw *coreplugin.IacPluginInstance) *coreplugin.PluginInstance {
	if raw == nil {
		return nil
	}
	return &coreplugin.PluginInstance{
		ID:         raw.ID,
		PluginName: raw.PluginName,
		Name:       raw.Name,
		Properties: mergeInstanceProperties(raw.Attrs, raw.SecretAttrs),
		Extension:  raw.Extension,
		Memo:       raw.Memo,
		TenantID:   raw.TenantID,
		Revision:   raw.Revision,
	}
}

func normalizeInstanceListFields(fields []string) []string {
	if len(fields) == 0 {
		return nil
	}
	seen := make(map[string]struct{}, len(fields)+2)
	result := make([]string, 0, len(fields)+1)
	for _, f := range fields {
		if f == "properties" {
			for _, dbCol := range []string{"attrs", "secret_attrs"} {
				if _, ok := seen[dbCol]; !ok {
					seen[dbCol] = struct{}{}
					result = append(result, dbCol)
				}
			}
			continue
		}
		if _, ok := seen[f]; !ok {
			seen[f] = struct{}{}
			result = append(result, f)
		}
	}
	return result
}
