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

package register

import (
	"encoding/json"
	"errors"
	"fmt"
	"sort"
	"strings"

	"hcm/cmd/hc-service/logics/plugin/query"
	pluginschema "hcm/cmd/hc-service/logics/plugin/schema"
	"hcm/cmd/hc-service/logics/plugin/tf/discovery"
	"hcm/pkg/api/core"
	coreplugin "hcm/pkg/api/core/plugin"
	dataservice "hcm/pkg/client/data-service"
	"hcm/pkg/criteria/enumor"
	"hcm/pkg/dal/dao/tools"
	"hcm/pkg/iacplugin"
	"hcm/pkg/kit"
	"hcm/pkg/lock"
	"hcm/pkg/logs"
	cvt "hcm/pkg/tools/converter"
	"hcm/pkg/tools/slice"
)

// New 创建 Service。
func New(dataCli *dataservice.Client, etcdMutex *lock.EtcdMutex, schemaMgr *pluginschema.SchemaManager,
	starter PluginProcessStarter, opts Options) *Service {

	return &Service{dataCli: dataCli, etcdMutex: etcdMutex, schemaMgr: schemaMgr, starter: starter, opts: opts}
}

// Register 注册一个插件，成功返回 nil；任何错误都意味着 hc-service 应当退出。
// 注册、升级、启停共用同一把插件名锁
func (s *Service) Register(kt *kit.Kit, entry discovery.Entry) error {
	leaseID, err := s.etcdMutex.Lock(kt.Ctx, lock.IaCPluginKey(entry.PluginName), s.opts.LockLeaseSeconds,
		s.opts.LockWaitTimeout)
	if err != nil {
		return fmt.Errorf("lock plugin %s failed: %w", entry.PluginName, err)
	}
	defer func() {
		if uerr := s.etcdMutex.UnLock(leaseID); uerr != nil {
			logs.Errorf("unlock plugin %s failed, err: %v, rid: %s", entry.PluginName, uerr, kt.Rid)
		}
	}()

	records, err := query.List(kt, s.dataCli, entry.PluginName)
	if err != nil {
		return err
	}

	decision, err := Decide(records, entry.PluginVersion)
	if err != nil {
		return fmt.Errorf("decide plugin %s failed: %w", entry.PluginName, err)
	}
	logs.Infof("plugin %s@%s decision: %s (%s), rid: %s", entry.PluginName, entry.PluginVersion,
		decision.Action, decision.Reason, kt.Rid)

	switch decision.Action {
	case IaCPluginActionExit:
		return fmt.Errorf("%w: %s", errPluginVersionConflict, decision.Reason)

	case IaCPluginActionVerify:
		return s.verifyRecord(kt, entry, decision.Record)

	case IaCPluginActionRegister:
		id, err := s.saveInitRecord(kt, entry, decision.Record)
		if err != nil {
			return err
		}
		return s.syncSchemaAndUpdate(kt, entry, id, records)

	default:
		return fmt.Errorf("unexpected decision action %q", decision.Action)
	}
}

// saveInitRecord 把该版本的记录落库为 init，返回记录 ID 供同步流程收尾。
// 没有记录则新建；处于 init（上次中断）或 failed（上次失败）的记录则重置为 init 后重入。
func (s *Service) saveInitRecord(kt *kit.Kit, entry discovery.Entry, rec *coreplugin.IaCPlugin) (string, error) {
	if rec == nil {
		return query.Insert(kt, s.dataCli, entry.PluginName, entry.PluginVersion,
			enumor.IaCPluginProtocolTFProvider, json.RawMessage(`{}`), nil)
	}

	err := query.Update(kt, s.dataCli, rec.ID, rec.State, enumor.IaCPluginStateInit, nil, nil)
	if err != nil {
		return "", fmt.Errorf("reset plugin %s@%s to init failed: %w", rec.PluginName, rec.PluginVersion, err)
	}

	return rec.ID, nil
}

// verifyRecord 已注册完成的版本只校验本架构 sha256，不拉起进程、不同步 Schema。
func (s *Service) verifyRecord(kt *kit.Kit, entry discovery.Entry, rec *coreplugin.IaCPlugin) error {
	result, err := VerifyHash(rec, entry.Arch, entry.SHA256)
	if err != nil {
		return err
	}

	switch result {
	case IaCPluginHashResultMatch:
		return nil

	case IaCPluginHashResultMismatch:
		// 同一版本号下本架构的文件内容变了：不修改任何记录，进程退出。
		return fmt.Errorf("%w: plugin %s@%s sha256 of %s differs from the registered one, "+
			"bump the plugin version instead of replacing the binary", errPluginVersionConflict, entry.PluginName,
			entry.PluginVersion, entry.Arch)

	case IaCPluginHashResultMissing:
		// 例如首个 arm64 副本第一次启动：补写本架构的 hash，其它架构的值保持不变。
		ext, err := WithArchHash(rec.Extension, entry.Arch, entry.SHA256)
		if err != nil {
			return err
		}
		if err = query.Update(kt, s.dataCli, rec.ID, rec.State, rec.State, ext, nil); err != nil {
			return fmt.Errorf("write sha256 of %s for plugin %s@%s failed: %w", entry.Arch,
				entry.PluginName, entry.PluginVersion, err)
		}
		return nil

	default:
		return fmt.Errorf("unexpected hash result %q", result)
	}
}

// syncSchemaAndUpdate 同步 Schema，并把记录从 init 收尾为 enabled / disabled / failed。
func (s *Service) syncSchemaAndUpdate(kt *kit.Kit, entry discovery.Entry, id string,
	records []coreplugin.IaCPlugin) error {

	if err := s.syncSchema(kt, entry); err != nil {
		// 确定性失败：置 failed 并写原因。写不进去时记录保持 init，下一个副本照样会重入。
		s.markFailed(kt, entry, id, err)
		return fmt.Errorf("register plugin %s@%s failed: %w", entry.PluginName, entry.PluginVersion, err)
	}

	target := iacplugin.InheritState(records, entry.PluginVersion)
	ext, err := WithArchHash(json.RawMessage(`{}`), entry.Arch, entry.SHA256)
	if err != nil {
		return err
	}

	err = query.Update(kt, s.dataCli, id, enumor.IaCPluginStateInit, target, ext, nil)
	if err != nil {
		return fmt.Errorf("mark plugin %s@%s as %s failed: %w", entry.PluginName, entry.PluginVersion, target, err)
	}

	logs.Infof("plugin %s@%s registered, state: %s, rid: %s", entry.PluginName, entry.PluginVersion, target, kt.Rid)
	return nil
}

// syncSchema 拉起一次性进程并同步插件的 Schema。Sync 只通过 SchemaGetter 取 Schema，不感知插件是本地进程还是 http 后端；
// 这里的一次性进程是我们拉起的，所以由我们负责关闭：无论成败，只要拉起成功就一定 Close，拉起失败时不调用 Sync。
func (s *Service) syncSchema(kt *kit.Kit, entry discovery.Entry) error {
	// TODO: 接入插件进程管理实现后，由 Service.RegisterPlugins 注入 PluginProcessStarter。
	if s.starter == nil {
		return errors.New("plugin process starter is not configured")
	}

	manager, err := s.starter.NewProcessManager(kt, entry.PluginName, entry.PluginVersion, entry.ExecPath)
	if err != nil {
		return fmt.Errorf("start plugin process failed: %w", err)
	}
	defer func() {
		if cerr := manager.Close(); cerr != nil {
			logs.Errorf("close plugin %s@%s process failed, err: %v, rid: %s", entry.PluginName,
				entry.PluginVersion, cerr, kt.Rid)
		}
	}()

	executor, err := manager.Start(kt)
	if err != nil {
		return fmt.Errorf("start plugin executor failed: %w", err)
	}

	// 只把执行器"取 Schema 的方法"交给 SyncSchema，执行器与进程管理对象都不进入同步逻辑。
	result, err := s.schemaMgr.SyncSchema(kt, pluginschema.SchemaGetterFunc(executor.GetProviderSchema))
	if err != nil {
		return fmt.Errorf("sync schema failed: %w", err)
	}
	if result == nil || result.SyncResult != enumor.SchemaSyncSuccess {
		msg := "sync plugin schema failed"
		if result != nil && len(result.SyncMessage) != 0 {
			msg = result.SyncMessage
		}
		return fmt.Errorf("sync schema failed: %s", msg)
	}

	return nil
}

// markFailed 把记录置 failed 并写入失败原因。写库失败只记日志：记录保持 init，不影响重入。
func (s *Service) markFailed(kt *kit.Kit, entry discovery.Entry, id string, cause error) {
	if err := query.MarkFailed(kt, s.dataCli, id, enumor.IaCPluginStateInit, cause); err != nil {
		logs.Errorf("mark plugin %s@%s as failed failed, err: %v, cause: %v, rid: %s", entry.PluginName,
			entry.PluginVersion, err, cause, kt.Rid)
	}
}

// CheckRegistered 读取注册表并检查是否有已启用的 tf 插件在本副本目录里缺失；缺失时返回 ErrMissingPlugin。
//
// 它只读，不加锁。要在所有插件注册完成之后调用：本副本刚注册的插件此时已经在注册表里，不会被误判。
func (s *Service) CheckRegistered(kt *kit.Kit, entries []discovery.Entry) error {
	res, err := s.dataCli.Global.IaCPlugin.List(kt, &core.ListReq{
		Filter: tools.EqualExpression("protocol", enumor.IaCPluginProtocolTFProvider),
		Page:   core.NewDefaultBasePage(),
	})
	if err != nil {
		return fmt.Errorf("list tf_provider plugins failed: %w", err)
	}

	missing, err := FindMissing(res.Details, entries)
	if err != nil {
		return err
	}
	if len(missing) > 0 {
		return fmt.Errorf("%w: %s", errMissingPlugin, strings.Join(missing, ", "))
	}

	return nil
}

// FindMissing 找出"注册表里已启用、本副本目录里却没有"的插件名。
//
// 每个插件名只看版本最高的那条记录（不论状态）的 state：enabled 才算缺失；disabled / init / failed 都跳过。
//   - disabled：管理员主动停用，本来就不该有流量，目录里没有无妨；
//   - init / failed：注册没成功，没有可用版本可言，缺失不是新问题，由注册流程自己收敛。
//
// 本副本目录里有该插件（不论版本）就不算缺失：版本是否匹配由 Decide 负责。
func FindMissing(records []coreplugin.IaCPlugin, entries []discovery.Entry) ([]string, error) {
	local := cvt.StringSliceToMap(slice.Map(entries, func(e discovery.Entry) string { return e.PluginName }))

	// 按插件名分组：同一插件名可能有多条版本记录。
	byName := make(map[string][]coreplugin.IaCPlugin, len(records))
	for _, rec := range slice.Filter(records, func(rec coreplugin.IaCPlugin) bool {
		return rec.Protocol == enumor.IaCPluginProtocolTFProvider
	}) {
		byName[rec.PluginName] = append(byName[rec.PluginName], rec)
	}

	missing := make([]string, 0)
	for name, recs := range byName {
		if _, ok := local[name]; ok {
			continue
		}

		highest, err := iacplugin.Latest(recs)
		if err != nil {
			return nil, fmt.Errorf("plugin %s: %w", name, err)
		}
		if highest != nil && highest.State == enumor.IaCPluginStateEnabled {
			missing = append(missing, name)
		}
	}

	sort.Strings(missing)
	return missing, nil
}
