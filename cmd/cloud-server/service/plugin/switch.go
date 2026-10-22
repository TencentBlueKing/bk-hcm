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
	"errors"
	"strings"

	proto "hcm/pkg/api/cloud-server"
	"hcm/pkg/api/core"
	coreplugin "hcm/pkg/api/core/plugin"
	dataplugin "hcm/pkg/api/data-service/plugin"
	"hcm/pkg/criteria/enumor"
	"hcm/pkg/criteria/errf"
	"hcm/pkg/dal/dao/tools"
	"hcm/pkg/iacplugin"
	"hcm/pkg/iam/meta"
	"hcm/pkg/kit"
	"hcm/pkg/lock"
	"hcm/pkg/logs"
	"hcm/pkg/rest"

	etcd3 "go.etcd.io/etcd/client/v3"
)

// switchLockLeaseSeconds 启停操作的 etcd 锁租约时长（秒）。启停只涉及数次数据库读写，60 秒充足。
const switchLockLeaseSeconds int64 = 60

// SwitchIaCPlugin 启用或停用 IaC 插件。
func (s *svc) SwitchIaCPlugin(cts *rest.Contexts) (interface{}, error) {
	req := new(proto.IaCPluginSwitchReq)
	if err := cts.DecodeInto(req); err != nil {
		logs.Errorf("decode switch iac plugin request failed, err: %v, rid: %s", err, cts.Kit.Rid)
		return nil, errf.NewFromErr(errf.DecodeRequestFailed, err)
	}

	if err := req.Validate(); err != nil {
		return nil, errf.NewFromErr(errf.InvalidParameter, err)
	}

	authRes := meta.ResourceAttribute{
		Basic: &meta.Basic{
			Type:   meta.IaCPlugin,
			Action: meta.Update,
		},
	}
	if err := s.authorizer.AuthorizeWithPerm(cts.Kit, authRes); err != nil {
		logs.Errorf("switch iac plugin auth failed, err: %v, plugin_name: %s, rid: %s",
			err, req.PluginName, cts.Kit.Rid)
		return nil, err
	}

	if err := s.switchIaCPlugin(cts.Kit, req.PluginName, req.State); err != nil {
		return nil, err
	}

	return nil, nil
}

// switchIaCPlugin 持插件名锁，按该插件名下版本最高的记录判定状态规则并批量更新相关版本状态。
func (s *svc) switchIaCPlugin(kt *kit.Kit, pluginName string, targetState enumor.IaCPluginState) error {
	leaseID, err := s.tryLockIaCPlugin(kt, pluginName)
	if err != nil {
		return err
	}
	defer s.unlockIaCPlugin(kt, pluginName, leaseID)

	current, err := s.getLatestIaCPlugin(kt, pluginName)
	if err != nil {
		return err
	}

	schemaExists, err := s.checkSchemaForEnable(kt, current, targetState)
	if err != nil {
		return err
	}

	fromStates, err := decideSwitch(current, targetState, schemaExists)
	if err != nil || len(fromStates) == 0 {
		return err
	}

	return s.batchUpdateIaCPluginState(kt, pluginName, fromStates, targetState)
}

// decideSwitch 按当前版本（最高版本）记录判定启停动作，返回需批量改为 to 的源状态列表；无需更新时返回 nil, nil。
func decideSwitch(current *coreplugin.IaCPlugin, to enumor.IaCPluginState,
	schemaExists bool) ([]enumor.IaCPluginState, error) {

	if current.State == to {
		return nil, nil
	}
	if current.State == enumor.IaCPluginStateInit {
		return nil, errf.Newf(errf.Aborted, "iac plugin %s@%s registration or upgrade is not finished",
			current.PluginName, current.PluginVersion)
	}

	switch to {
	case enumor.IaCPluginStateDisabled:
		return []enumor.IaCPluginState{enumor.IaCPluginStateEnabled, enumor.IaCPluginStateFailed}, nil
	case enumor.IaCPluginStateEnabled:
		if current.State == enumor.IaCPluginStateFailed {
			return nil, errf.Newf(errf.Aborted, "iac plugin %s@%s is in failed state (%s), please recover first",
				current.PluginName, current.PluginVersion, current.Message)
		}
		if !schemaExists {
			return nil, errf.Newf(errf.Aborted, "iac plugin %s@%s schema does not exist, please recover first",
				current.PluginName, current.PluginVersion)
		}
		return []enumor.IaCPluginState{enumor.IaCPluginStateDisabled}, nil
	default:
		return nil, errf.Newf(errf.InvalidParameter, "unsupported switch target state: %s", to)
	}
}

func (s *svc) getLatestIaCPlugin(kt *kit.Kit, pluginName string) (*coreplugin.IaCPlugin, error) {
	records, err := s.listIaCPluginsByName(kt, pluginName)
	if err != nil {
		return nil, err
	}

	current, err := iacplugin.Latest(records)
	if err != nil {
		logs.Errorf("select latest iac plugin version failed, err: %v, plugin_name: %s, rid: %s",
			err, pluginName, kt.Rid)
		return nil, err
	}
	if current == nil {
		return nil, errf.Newf(errf.RecordNotFound, "iac plugin %s not found", pluginName)
	}

	return current, nil
}

func (s *svc) listIaCPluginsByName(kt *kit.Kit, pluginName string) ([]coreplugin.IaCPlugin, error) {
	resp, err := s.client.DataService().Global.IaCPlugin.List(kt, &core.ListReq{
		Filter: tools.EqualExpression("plugin_name", pluginName),
		Page:   &core.BasePage{Limit: core.DefaultMaxPageLimit},
	})
	if err != nil {
		logs.Errorf("list iac plugin by name failed, err: %v, plugin_name: %s, rid: %s", err, pluginName, kt.Rid)
		return nil, err
	}

	return resp.Details, nil
}

// checkSchemaForEnable 仅在将 disabled 插件启用为 enabled 时查询当前版本 Schema 是否存在；其余场景无需查询。
func (s *svc) checkSchemaForEnable(kt *kit.Kit, current *coreplugin.IaCPlugin,
	targetState enumor.IaCPluginState) (bool, error) {

	if current.State != enumor.IaCPluginStateDisabled || targetState != enumor.IaCPluginStateEnabled {
		return false, nil
	}

	return s.isIaCPluginSchemaExists(kt, current.PluginName, current.PluginVersion)
}

// isIaCPluginSchemaExists 按插件名和版本查询 Schema 是否已落库。
func (s *svc) isIaCPluginSchemaExists(kt *kit.Kit, pluginName, pluginVersion string) (bool, error) {
	resp, err := s.client.DataService().Global.IacPluginSchema.List(kt, &core.ListReq{
		Fields: []string{"id"},
		Filter: tools.ExpressionAnd(
			tools.RuleEqual("plugin_name", pluginName),
			tools.RuleEqual("plugin_version", pluginVersion),
		),
		Page: &core.BasePage{Limit: 1},
	})
	if err != nil {
		logs.Errorf("list iac plugin schema failed, err: %v, plugin_name: %s, plugin_version: %s, rid: %s",
			err, pluginName, pluginVersion, kt.Rid)
		return false, err
	}

	return resp != nil && len(resp.Details) > 0, nil
}

func (s *svc) batchUpdateIaCPluginState(kt *kit.Kit, pluginName string,
	fromStates []enumor.IaCPluginState, toState enumor.IaCPluginState) error {

	err := s.client.DataService().Global.IaCPlugin.BatchUpdateState(kt,
		&dataplugin.IaCPluginBatchUpdateStateReq{
			PluginName: pluginName,
			FromStates: fromStates,
			ToState:    toState,
		})
	if err != nil {
		logs.Errorf("batch update iac plugin state failed, err: %v, plugin_name: %s, to_state: %s, rid: %s",
			err, pluginName, toState, kt.Rid)
		return err
	}

	return nil
}

func (s *svc) tryLockIaCPlugin(kt *kit.Kit, pluginName string) (etcd3.LeaseID, error) {
	leaseID, err := lock.Manager.TryLockWithTTL(lock.IaCPluginKey(pluginName), switchLockLeaseSeconds)
	if err != nil {
		if errors.Is(err, lock.ErrLockFailed) {
			return 0, errf.Newf(errf.Aborted,
				"iac plugin %s is being registered, upgraded or modified, please retry later", pluginName)
		}
		logs.Errorf("try lock iac plugin failed, err: %v, plugin_name: %s, rid: %s", err, pluginName, kt.Rid)
		return 0, err
	}

	return leaseID, nil
}

func (s *svc) unlockIaCPlugin(kt *kit.Kit, pluginName string, leaseID etcd3.LeaseID) {
	if err := lock.Manager.UnLock(leaseID); err != nil {
		if strings.Contains(err.Error(), "requested lease not found") {
			return
		}
		logs.Errorf("unlock iac plugin lock failed, err: %v, plugin_name: %s, lease_id: %d, rid: %s",
			err, pluginName, leaseID, kt.Rid)
	}
}
