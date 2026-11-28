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

package check

import (
	"encoding/json"
	"errors"
	"fmt"
	"time"

	httpplugin "hcm/cmd/hc-service/logics/iac-plugin/http-plugin"
	coreplugin "hcm/pkg/api/core/plugin"
	"hcm/pkg/client/data-service/global"
	"hcm/pkg/criteria/enumor"
	"hcm/pkg/dal/dao/tools"
	"hcm/pkg/iacplugin"
	"hcm/pkg/kit"
	"hcm/pkg/lock"
	"hcm/pkg/logs"
	"hcm/pkg/tools/slice"
)

var _ Checker = (*HTTPChecker)(nil)

// HTTPOptions http 插件巡检器的运行参数。
type HTTPOptions struct {
	// Interval 巡检周期。
	Interval time.Duration
	// FailureThreshold enabled 插件连续探活失败多少次后置为 failed。
	FailureThreshold int
	// LockLeaseSeconds 插件名锁租约（秒）。
	LockLeaseSeconds int64
}

// HTTPChecker 周期性巡检 http 插件，仅 master 副本执行，避免多副本重复探活。
type HTTPChecker struct {
	cli        *global.IaCPluginClient
	mutex      *lock.EtcdMutex
	httpSvc    *httpplugin.Service
	isMaster   func() bool
	opts       HTTPOptions
	failCounts map[string]int
}

// NewHTTPChecker 创建 http 插件巡检器。
func NewHTTPChecker(cli *global.IaCPluginClient, mutex *lock.EtcdMutex, httpSvc *httpplugin.Service,
	isMaster func() bool, opts HTTPOptions) *HTTPChecker {

	return &HTTPChecker{
		cli:        cli,
		mutex:      mutex,
		httpSvc:    httpSvc,
		isMaster:   isMaster,
		opts:       opts,
		failCounts: make(map[string]int),
	}
}

// Name 返回检查器标识。
func (c *HTTPChecker) Name() string {
	return string(enumor.IaCPluginProtocolHTTP)
}

// Interval 返回巡检周期。
func (c *HTTPChecker) Interval() time.Duration {
	return c.opts.Interval
}

// Check 执行一轮巡检，返回 false 以保持周期调度。
func (c *HTTPChecker) Check(kt *kit.Kit) bool {
	if !c.isMaster() {
		clear(c.failCounts)
		return false
	}

	names, err := c.listHTTPPluginNames(kt)
	if err != nil {
		logs.Warnf("list http iac plugins failed, skip this check round, err: %v, rid: %s", err, kt.Rid)
		return false
	}

	for _, name := range names {
		if kt.Ctx.Err() != nil {
			return false
		}
		c.checkPlugin(kt, name)
	}
	return false
}

// checkPlugin 巡检一个插件。拿不到锁跳过本轮；探活失败按状态计数或告警；探活成功走升级判定（含 failed 恢复）。
func (c *HTTPChecker) checkPlugin(kt *kit.Kit, pluginName string) {
	leaseID, err := c.mutex.TryLockWithTTL(lock.IaCPluginKey(pluginName), c.opts.LockLeaseSeconds)
	if err != nil {
		if !errors.Is(err, lock.ErrLockFailed) {
			logs.Warnf("lock http iac plugin %s for check failed, skip, err: %v, rid: %s", pluginName, err, kt.Rid)
		}
		return
	}
	defer c.mutex.UnlockIaCPlugin(kt, pluginName, leaseID)

	records, err := c.cli.ListAll(kt, tools.EqualExpression("plugin_name", pluginName), nil)
	if err != nil {
		logs.Warnf("list http iac plugin %s for check failed, skip, err: %v, rid: %s", pluginName, err, kt.Rid)
		return
	}

	latest, err := iacplugin.Latest(records)
	if err != nil || latest == nil {
		logs.Warnf("get latest http iac plugin %s for check failed, skip, err: %v, rid: %s", pluginName, err, kt.Rid)
		return
	}

	ext := new(coreplugin.HTTPExtension)
	if err := json.Unmarshal([]byte(latest.Extension), ext); err != nil {
		logs.Warnf("parse http iac plugin %s@%s extension failed, err: %v, rid: %s", latest.PluginName,
			latest.PluginVersion, err, kt.Rid)
		return
	}

	version, err := c.httpSvc.Probe(kt, pluginName, ext.APIPath)
	if err != nil {
		c.onProbeFailed(kt, latest, err)
		return
	}
	delete(c.failCounts, pluginName)

	result, err := c.httpSvc.UpgradeTo(kt, records, latest, version)
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
func (c *HTTPChecker) onProbeFailed(kt *kit.Kit, latest *coreplugin.IaCPlugin, cause error) {
	if latest.State != enumor.IaCPluginStateEnabled {
		delete(c.failCounts, latest.PluginName)
		logs.Warnf("check http iac plugin %s@%s (%s) probe failed, err: %v, rid: %s", latest.PluginName,
			latest.PluginVersion, latest.State, cause, kt.Rid)
		return
	}

	c.failCounts[latest.PluginName]++
	n := c.failCounts[latest.PluginName]
	logs.Warnf("check http iac plugin %s@%s probe failed, consecutive: %d/%d, err: %v, rid: %s", latest.PluginName,
		latest.PluginVersion, n, c.opts.FailureThreshold, cause, kt.Rid)
	if n < c.opts.FailureThreshold {
		return
	}
	delete(c.failCounts, latest.PluginName)

	reason := fmt.Errorf("probe failed %d consecutive times: %v", n, cause)
	err := c.cli.MarkFailed(kt, latest.ID, enumor.IaCPluginStateEnabled, reason)
	if err != nil {
		logs.Errorf("mark http iac plugin %s@%s as failed failed, err: %v, rid: %s", latest.PluginName,
			latest.PluginVersion, err, kt.Rid)
		return
	}

	logs.Errorf("http iac plugin %s@%s is marked as failed, reason: %v, rid: %s", latest.PluginName,
		latest.PluginVersion, reason, kt.Rid)
}

// listHTTPPluginNames 列出所有 http 插件名（去重）。
func (c *HTTPChecker) listHTTPPluginNames(kt *kit.Kit) ([]string, error) {
	details, err := c.cli.ListAll(kt,
		tools.EqualExpression("protocol", enumor.IaCPluginProtocolHTTP), []string{"plugin_name"})
	if err != nil {
		return nil, err
	}

	names := slice.Map(details, func(one coreplugin.IaCPlugin) string { return one.PluginName })
	return slice.Unique(names), nil
}
