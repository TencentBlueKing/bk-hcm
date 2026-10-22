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

// Package plugin 提供 hc-service 侧插件实例的 HTTP 服务路由与处理器。
package plugin

import (
	"net/http"

	"hcm/cmd/hc-service/logics/plugin/instance"
	"hcm/cmd/hc-service/logics/plugin/schema"
	"hcm/cmd/hc-service/service/capability"
	hcplugin "hcm/pkg/api/hc-service/plugin"
	"hcm/pkg/criteria/errf"
	"hcm/pkg/logs"
	"hcm/pkg/rest"
)

// InitService 初始化 hc-service 插件实例服务路由。
func InitService(cap *capability.Capability) {
	schemaMgr := schema.NewSchemaManager(cap.ClientSet.DataService(), schema.DefaultSchemaCacheTTL)
	svc := &service{
		schemaMgr:   schemaMgr,
		instanceMgr: instance.NewPluginInstanceManager(cap.ClientSet.DataService(), schemaMgr),
	}

	h := rest.NewHandler()

	svc.initPluginInstanceRoutes(h)

	h.Load(cap.WebService)
}

type service struct {
	schemaMgr   *schema.SchemaManager
	instanceMgr *instance.PluginInstanceManager
}

func (svc *service) initPluginInstanceRoutes(h *rest.Handler) {
	h.Add("CreatePluginInstance", http.MethodPost, "/plugin_instances/create", svc.CreatePluginInstance)
	h.Add("UpdatePluginInstance", http.MethodPatch, "/plugin_instances/{id}", svc.UpdatePluginInstance)
	h.Add("DeletePluginInstance", http.MethodDelete, "/plugin_instances/{id}", svc.DeletePluginInstance)
	h.Add("GetPluginInstance", http.MethodGet, "/plugin_instances/{id}", svc.GetPluginInstance)
	h.Add("ListPluginInstances", http.MethodPost, "/plugin_instances/list", svc.ListPluginInstances)
}

// CreatePluginInstance 创建插件实例。
func (svc *service) CreatePluginInstance(cts *rest.Contexts) (any, error) {
	req := new(hcplugin.PluginInstanceCreateReq)
	if err := cts.DecodeInto(req); err != nil {
		logs.Errorf("create plugin instance decode request failed, err: %v, rid: %s", err, cts.Kit.Rid)
		return nil, errf.NewFromErr(errf.InvalidParameter, err)
	}

	if err := req.Validate(); err != nil {
		logs.Errorf("create plugin instance validate request failed, err: %v, rid: %s", err, cts.Kit.Rid)
		return nil, errf.NewFromErr(errf.InvalidParameter, err)
	}

	res, err := svc.instanceMgr.CreatePluginInstance(cts.Kit, req)
	if err != nil {
		logs.Errorf("create plugin instance failed, plugin: %s, name: %s, err: %v, rid: %s",
			req.PluginName, req.Name, err, cts.Kit.Rid)
		return nil, err
	}

	return res, nil
}

// UpdatePluginInstance 更新插件实例。
func (svc *service) UpdatePluginInstance(cts *rest.Contexts) (any, error) {
	id := cts.PathParameter("id").String()
	if len(id) == 0 {
		return nil, errf.New(errf.InvalidParameter, "id is required")
	}

	req := new(hcplugin.PluginInstanceUpdateReq)
	if err := cts.DecodeInto(req); err != nil {
		logs.Errorf("update plugin instance decode request failed, id: %s, err: %v, rid: %s", id, err, cts.Kit.Rid)
		return nil, errf.NewFromErr(errf.InvalidParameter, err)
	}

	if err := req.Validate(); err != nil {
		logs.Errorf("update plugin instance validate request failed, id: %s, err: %v, rid: %s", id, err, cts.Kit.Rid)
		return nil, errf.NewFromErr(errf.InvalidParameter, err)
	}

	if err := svc.instanceMgr.UpdatePluginInstance(cts.Kit, id, req); err != nil {
		logs.Errorf("update plugin instance failed, id: %s, err: %v, rid: %s", id, err, cts.Kit.Rid)
		return nil, err
	}

	return nil, nil
}

// DeletePluginInstance 删除插件实例。
func (svc *service) DeletePluginInstance(cts *rest.Contexts) (any, error) {
	id := cts.PathParameter("id").String()
	if len(id) == 0 {
		return nil, errf.New(errf.InvalidParameter, "id is required")
	}

	if err := svc.instanceMgr.DeletePluginInstance(cts.Kit, id); err != nil {
		logs.Errorf("delete plugin instance failed, id: %s, err: %v, rid: %s", id, err, cts.Kit.Rid)
		return nil, err
	}

	return nil, nil
}

// GetPluginInstance 获取单条插件实例详情。
func (svc *service) GetPluginInstance(cts *rest.Contexts) (any, error) {
	id := cts.PathParameter("id").String()
	if len(id) == 0 {
		return nil, errf.New(errf.InvalidParameter, "id is required")
	}

	res, err := svc.instanceMgr.GetPluginInstance(cts.Kit, id)
	if err != nil {
		logs.Errorf("get plugin instance failed, id: %s, err: %v, rid: %s", id, err, cts.Kit.Rid)
		return nil, err
	}

	return res, nil
}

// ListPluginInstances 列表查询插件实例。
func (svc *service) ListPluginInstances(cts *rest.Contexts) (any, error) {
	req := new(hcplugin.PluginInstanceListReq)
	if err := cts.DecodeInto(req); err != nil {
		logs.Errorf("list plugin instances decode request failed, err: %v, rid: %s", err, cts.Kit.Rid)
		return nil, errf.NewFromErr(errf.InvalidParameter, err)
	}

	if err := req.Validate(); err != nil {
		logs.Errorf("list plugin instances validate request failed, err: %v, rid: %s", err, cts.Kit.Rid)
		return nil, errf.NewFromErr(errf.InvalidParameter, err)
	}

	res, err := svc.instanceMgr.ListPluginInstance(cts.Kit, req)
	if err != nil {
		logs.Errorf("list plugin instances failed, err: %v, rid: %s", err, cts.Kit.Rid)
		return nil, err
	}

	return res, nil
}
