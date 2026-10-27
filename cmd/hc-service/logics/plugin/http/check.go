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

package http

import (
	"context"
	"errors"
	"fmt"
	"time"

	"hcm/cmd/hc-service/logics/plugin/query"
	"hcm/pkg/api/core"
	coreplugin "hcm/pkg/api/core/plugin"
	"hcm/pkg/criteria/enumor"
	"hcm/pkg/dal/dao/tools"
	"hcm/pkg/kit"
	"hcm/pkg/lock"
	"hcm/pkg/logs"
	"hcm/pkg/tools/slice"
)

// RunCheck 阻塞执行巡检，直到 ctx 取消。仅 isMaster 返回 true 的副本执行，避免多副本重复探活。
func (s *Service) RunCheck(ctx context.Context, isMaster func() bool) {
	failCounts := make(map[string]int)

	ticker := time.NewTicker(s.opts.CheckInterval)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			if !isMaster() {
				clear(failCounts)
				continue
			}
			s.checkRound(ctx, failCounts)
		}
	}
}

// checkRound 一轮巡检：对所有 http 插件逐个（串行）走升级流程。
func (s *Service) checkRound(ctx context.Context, failCounts map[string]int) {
	// 后端任务使用后端操作 kit：data-service 会校验请求头里的操作人与 AppCode，审计操作人即 hcm-backend-admin。
	kt := core.NewBackendKit()
	kt.Ctx = ctx

	names, err := s.listHTTPPluginNames(kt)
	if err != nil {
		logs.Warnf("list http iac plugins failed, skip this check round, err: %v, rid: %s", err, kt.Rid)
		return
	}

	for _, name := range names {
		if ctx.Err() != nil {
			return
		}
		s.checkPlugin(kt, name, failCounts)
	}
}

// checkPlugin 巡检一个插件。拿不到锁跳过本轮；探活失败按状态计数或告警；探活成功走升级判定（含 failed 恢复）。
func (s *Service) checkPlugin(kt *kit.Kit, pluginName string, failCounts map[string]int) {
	leaseID, err := s.mutex.TryLockWithTTL(lock.IaCPluginKey(pluginName), s.opts.LockLeaseSeconds)
	if err != nil {
		if !errors.Is(err, lock.ErrLockFailed) {
			logs.Warnf("lock http iac plugin %s for check failed, skip, err: %v, rid: %s", pluginName, err, kt.Rid)
		}
		return
	}
	defer s.unlock(kt, pluginName, leaseID)

	records, latest, err := s.listAndGetLatest(kt, pluginName)
	if err != nil {
		logs.Warnf("list http iac plugin %s for check failed, skip, err: %v, rid: %s", pluginName, err, kt.Rid)
		return
	}

	ext, err := parseExtension(latest)
	if err != nil {
		logs.Warnf("check http iac plugin %s failed, err: %v, rid: %s", pluginName, err, kt.Rid)
		return
	}

	version, err := s.probe(kt, pluginName, ext.APIPath)
	if err != nil {
		s.onProbeFailed(kt, latest, failCounts, err)
		return
	}
	delete(failCounts, pluginName)

	result, err := s.upgradeTo(kt, records, latest, version)
	if err != nil {
		logs.Warnf("check http iac plugin %s@%s, upgrade to probed version %s failed, err: %v, rid: %s",
			pluginName, latest.PluginVersion, version, err, kt.Rid)
		return
	}

	if result.PluginVersion != latest.PluginVersion || result.State != latest.State {
		logs.Infof("check http iac plugin %s: %s@%s -> %s@%s, rid: %s", pluginName, latest.State,
			latest.PluginVersion, result.State, result.PluginVersion, kt.Rid)
	}
}

// onProbeFailed enabled 插件连续失败达到阈值改为 failed 并写原因；disabled 等其他状态只告警，不改记录。
func (s *Service) onProbeFailed(kt *kit.Kit, latest *coreplugin.IaCPlugin, failCounts map[string]int, cause error) {
	if latest.State != enumor.IaCPluginStateEnabled {
		delete(failCounts, latest.PluginName)
		logs.Warnf("check http iac plugin %s@%s (%s) probe failed, err: %v, rid: %s", latest.PluginName,
			latest.PluginVersion, latest.State, cause, kt.Rid)
		return
	}

	failCounts[latest.PluginName]++
	n := failCounts[latest.PluginName]
	logs.Warnf("check http iac plugin %s@%s probe failed, consecutive: %d/%d, err: %v, rid: %s", latest.PluginName,
		latest.PluginVersion, n, s.opts.FailureThreshold, cause, kt.Rid)
	if n < s.opts.FailureThreshold {
		return
	}
	delete(failCounts, latest.PluginName)

	reason := fmt.Errorf("probe failed %d consecutive times: %v", n, cause)
	err := query.MarkFailed(kt, s.dataCli, latest.ID, enumor.IaCPluginStateEnabled, reason)
	if err != nil {
		logs.Errorf("mark http iac plugin %s@%s as failed failed, err: %v, rid: %s", latest.PluginName,
			latest.PluginVersion, err, kt.Rid)
		return
	}

	logs.Errorf("http iac plugin %s@%s is marked as failed, reason: %v, rid: %s", latest.PluginName,
		latest.PluginVersion, reason, kt.Rid)
}

// listHTTPPluginNames 列出所有 http 插件名（去重）。
func (s *Service) listHTTPPluginNames(kt *kit.Kit) ([]string, error) {
	names := make([]string, 0)
	req := &core.ListReq{
		Fields: []string{"plugin_name"},
		Filter: tools.EqualExpression("protocol", enumor.IaCPluginProtocolHTTP),
		Page:   core.NewDefaultBasePage(),
	}
	for {
		res, err := s.dataCli.Global.IaCPlugin.List(kt, req)
		if err != nil {
			return nil, err
		}
		if res == nil {
			break
		}

		names = append(names, slice.Map(res.Details, func(one coreplugin.IaCPlugin) string { return one.PluginName })...)
		if uint(len(res.Details)) < req.Page.Limit {
			break
		}
		req.Page.Start += uint32(req.Page.Limit)
	}

	return slice.Unique(names), nil
}
