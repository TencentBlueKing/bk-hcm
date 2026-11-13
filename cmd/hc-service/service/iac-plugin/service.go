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

// Package iacplugin hc-service 的 http 插件内部路由，供 cloud-server 鉴权后转发。
package iacplugin

import (
	"net/http"

	httpplugin "hcm/cmd/hc-service/logics/iac-plugin/http-plugin"
	"hcm/cmd/hc-service/service/capability"
	proto "hcm/pkg/api/hc-service/iac-plugin"
	"hcm/pkg/criteria/errf"
	"hcm/pkg/logs"
	"hcm/pkg/rest"
)

// InitService initialize the iac plugin service.
func InitService(c *capability.Capability) {
	s := &svc{httpPlugin: c.HTTPPlugin}

	h := rest.NewHandler()

	h.Add("RegisterHTTPIaCPlugin", http.MethodPost, "/iac_plugins/http/register", s.RegisterHTTPIaCPlugin)
	h.Add("UpgradeHTTPIaCPlugin", http.MethodPost, "/iac_plugins/http/upgrade", s.UpgradeHTTPIaCPlugin)
	h.Add("UpdateHTTPIaCPlugin", http.MethodPost, "/iac_plugins/http/update", s.UpdateHTTPIaCPlugin)

	h.Load(c.WebService)
}

type svc struct {
	httpPlugin *httpplugin.Service
}

// RegisterHTTPIaCPlugin 注册 http 插件。
func (s *svc) RegisterHTTPIaCPlugin(cts *rest.Contexts) (any, error) {
	req := new(proto.HTTPRegisterReq)
	if err := cts.DecodeInto(req); err != nil {
		return nil, errf.NewFromErr(errf.DecodeRequestFailed, err)
	}

	if err := req.Validate(); err != nil {
		return nil, errf.NewFromErr(errf.InvalidParameter, err)
	}

	result, err := s.httpPlugin.Register(cts.Kit, req)
	if err != nil {
		logs.Errorf("register http iac plugin %s failed, err: %v, rid: %s", req.PluginName, err, cts.Kit.Rid)
		return nil, err
	}

	return result, nil
}

// UpgradeHTTPIaCPlugin 触发 http 插件的升级检查或异常恢复。
func (s *svc) UpgradeHTTPIaCPlugin(cts *rest.Contexts) (any, error) {
	req := new(proto.HTTPUpgradeReq)
	if err := cts.DecodeInto(req); err != nil {
		return nil, errf.NewFromErr(errf.DecodeRequestFailed, err)
	}

	if err := req.Validate(); err != nil {
		return nil, errf.NewFromErr(errf.InvalidParameter, err)
	}

	result, err := s.httpPlugin.Upgrade(cts.Kit, req.PluginName)
	if err != nil {
		logs.Errorf("upgrade http iac plugin %s failed, err: %v, rid: %s", req.PluginName, err, cts.Kit.Rid)
		return nil, err
	}

	return result, nil
}

// UpdateHTTPIaCPlugin 修改 http 插件。
func (s *svc) UpdateHTTPIaCPlugin(cts *rest.Contexts) (any, error) {
	req := new(proto.HTTPUpdateReq)
	if err := cts.DecodeInto(req); err != nil {
		return nil, errf.NewFromErr(errf.DecodeRequestFailed, err)
	}

	if err := req.Validate(); err != nil {
		return nil, errf.NewFromErr(errf.InvalidParameter, err)
	}

	result, err := s.httpPlugin.Update(cts.Kit, req)
	if err != nil {
		logs.Errorf("update http iac plugin %s failed, err: %v, rid: %s", req.PluginName, err, cts.Kit.Rid)
		return nil, err
	}

	return result, nil
}
