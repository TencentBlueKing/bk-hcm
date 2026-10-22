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

package plugin

import (
	hciac "hcm/pkg/api/hc-service/plugin"
	"hcm/pkg/criteria/errf"
	"hcm/pkg/iam/meta"
	"hcm/pkg/kit"
	"hcm/pkg/logs"
	"hcm/pkg/rest"
)

// 注册 / 升级 / 修改 http 插件：cloud-server 只做参数校验与鉴权，加锁、探活、写记录、同步与收尾都在 hc-service。

// RegisterIaCPlugin 注册 http 插件，同步执行，完成后才响应。
func (s *svc) RegisterIaCPlugin(cts *rest.Contexts) (interface{}, error) {
	req := new(hciac.HTTPRegisterReq)
	if err := cts.DecodeInto(req); err != nil {
		logs.Errorf("decode register iac plugin request failed, err: %v, rid: %s", err, cts.Kit.Rid)
		return nil, errf.NewFromErr(errf.DecodeRequestFailed, err)
	}

	if err := req.Validate(); err != nil {
		return nil, errf.NewFromErr(errf.InvalidParameter, err)
	}

	if err := s.authorizeIaCPlugin(cts.Kit, meta.Create); err != nil {
		return nil, err
	}

	result, err := s.client.HCService().IaCPlugin.RegisterHTTP(cts.Kit, req)
	if err != nil {
		logs.Errorf("register iac plugin %s failed, err: %v, rid: %s", req.PluginName, err, cts.Kit.Rid)
		return nil, err
	}

	return result, nil
}

// UpgradeIaCPlugin 触发 http 插件的升级检查（探活发现新版本则注册并同步 Schema，failed 时重做同步以恢复）。同步执行，完成后才响应。
func (s *svc) UpgradeIaCPlugin(cts *rest.Contexts) (interface{}, error) {
	req := new(hciac.HTTPUpgradeReq)
	if err := cts.DecodeInto(req); err != nil {
		logs.Errorf("decode upgrade iac plugin request failed, err: %v, rid: %s", err, cts.Kit.Rid)
		return nil, errf.NewFromErr(errf.DecodeRequestFailed, err)
	}

	if err := req.Validate(); err != nil {
		return nil, errf.NewFromErr(errf.InvalidParameter, err)
	}

	if err := s.authorizeIaCPlugin(cts.Kit, meta.Update); err != nil {
		return nil, err
	}

	result, err := s.client.HCService().IaCPlugin.UpgradeHTTP(cts.Kit, req)
	if err != nil {
		logs.Errorf("upgrade iac plugin %s failed, err: %v, rid: %s", req.PluginName, err, cts.Kit.Rid)
		return nil, err
	}

	return result, nil
}

// UpdateIaCPlugin 修改 http 插件的地址、加密配置或备注。同步执行，完成后才响应。
func (s *svc) UpdateIaCPlugin(cts *rest.Contexts) (interface{}, error) {
	req := new(hciac.HTTPUpdateReq)
	if err := cts.DecodeInto(req); err != nil {
		logs.Errorf("decode update iac plugin request failed, err: %v, rid: %s", err, cts.Kit.Rid)
		return nil, errf.NewFromErr(errf.DecodeRequestFailed, err)
	}

	if err := req.Validate(); err != nil {
		return nil, errf.NewFromErr(errf.InvalidParameter, err)
	}

	if err := s.authorizeIaCPlugin(cts.Kit, meta.Update); err != nil {
		return nil, err
	}

	result, err := s.client.HCService().IaCPlugin.UpdateHTTP(cts.Kit, req)
	if err != nil {
		logs.Errorf("update iac plugin %s failed, err: %v, rid: %s", req.PluginName, err, cts.Kit.Rid)
		return nil, err
	}

	return result, nil
}

func (s *svc) authorizeIaCPlugin(kt *kit.Kit, action meta.Action) error {
	authRes := meta.ResourceAttribute{
		Basic: &meta.Basic{
			Type:   meta.IaCPlugin,
			Action: action,
		},
	}
	if err := s.authorizer.AuthorizeWithPerm(kt, authRes); err != nil {
		logs.Errorf("iac plugin %s auth failed, err: %v, rid: %s", action, err, kt.Rid)
		return err
	}

	return nil
}
