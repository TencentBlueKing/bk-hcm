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

// Package register 实现 tf_provider 插件的启动注册与升级。
package register

import (
	"encoding/json"
	"errors"
	"fmt"
	"sort"
	"strings"
	"time"

	"hcm/cmd/hc-service/logics/iac-plugin/tf-plugin/discovery"
	pluginschema "hcm/cmd/hc-service/logics/plugin/schema"
	"hcm/cmd/hc-service/logics/tfruntime"
	"hcm/cmd/hc-service/logics/tfruntime/executor"
	"hcm/cmd/hc-service/types"
	coreplugin "hcm/pkg/api/core/plugin"
	dataplugin "hcm/pkg/api/data-service/plugin"
	dataservice "hcm/pkg/client/data-service"
	"hcm/pkg/criteria/enumor"
	"hcm/pkg/dal/dao/tools"
	tabletypes "hcm/pkg/dal/table/types"
	"hcm/pkg/iacplugin"
	"hcm/pkg/kit"
	"hcm/pkg/lock"
	"hcm/pkg/logs"
	cvt "hcm/pkg/tools/converter"
	"hcm/pkg/tools/slice"
)

// errPluginVersionConflict 本地插件状态与注册表冲突（版本回退、sha256 不一致），进程应退出且不修改记录。
var errPluginVersionConflict = errors.New("conflicts with the registered plugin")

// errMissingPlugin 注册表里有已启用的 tf 插件，但本地插件目录里不存在，进程应当退出。
var errMissingPlugin = errors.New("enabled plugin is missing in the local plugin directory")

type tfSchemaGetter struct {
	executor      executor.Executor
	pluginName    string
	pluginVersion string
}

func (g *tfSchemaGetter) GetSchema(kt *kit.Kit) (*coreplugin.PluginSchema, error) {
	schema, err := g.executor.GetProviderSchema(kt.Ctx)
	if err != nil {
		return nil, fmt.Errorf("get provider schema failed: %w", err)
	}
	coreSchema, err := toCorePluginSchema(schema, g.pluginName, g.pluginVersion)
	if err != nil {
		return nil, fmt.Errorf("convert provider schema failed: %w", err)
	}
	return coreSchema, nil
}

// Options 注册参数。
type Options struct {
	// LockLeaseSeconds 插件名锁的租约时长。
	LockLeaseSeconds int64
	// LockWaitTimeout 等待插件名锁的最长时间，等不到即按注册失败处理。
	LockWaitTimeout time.Duration
}

// Service 实现 tf_provider 插件的启动注册与升级。
//
// 失败一律返回错误，由 app.Run 让进程非 0 退出、交给编排重启，进程内不做重试。
type Service struct {
	dataCli   *dataservice.Client
	etcdMutex *lock.EtcdMutex
	schemaMgr *pluginschema.SchemaManager
	opts      Options
}

// New 创建 Service。
func New(dataCli *dataservice.Client, etcdMutex *lock.EtcdMutex, schemaMgr *pluginschema.SchemaManager,
	opts Options) *Service {

	return &Service{dataCli: dataCli, etcdMutex: etcdMutex, schemaMgr: schemaMgr, opts: opts}
}

// Register 注册一个插件，成功返回 nil；任何错误都意味着 hc-service 应当退出。
// 注册、升级、启停共用同一把插件名锁
func (s *Service) Register(kt *kit.Kit, entry discovery.PluginEntry) error {
	leaseID, err := s.etcdMutex.Lock(kt.Ctx, lock.IaCPluginKey(entry.PluginName), s.opts.LockLeaseSeconds,
		s.opts.LockWaitTimeout)
	if err != nil {
		return fmt.Errorf("lock plugin %s failed: %w", entry.PluginName, err)
	}
	defer s.etcdMutex.UnlockIaCPlugin(kt, entry.PluginName, leaseID)

	records, err := s.dataCli.Global.IaCPlugin.ListAll(kt, tools.EqualExpression("plugin_name", entry.PluginName), nil)
	if err != nil {
		return err
	}

	return s.registerWithRecords(kt, entry, records)
}

// registerWithRecords 按同名插件的已有记录执行本版本的校验或注册。
func (s *Service) registerWithRecords(kt *kit.Kit, entry discovery.PluginEntry,
	records []coreplugin.IaCPlugin) error {

	if _, err := iacplugin.ParseVersion(entry.PluginVersion); err != nil {
		return err
	}
	for _, rec := range records {
		if rec.Protocol != enumor.IaCPluginProtocolTFProvider {
			return fmt.Errorf("plugin %s is registered with protocol %s, but the local directory is a %s",
				rec.PluginName, rec.Protocol, enumor.IaCPluginProtocolTFProvider)
		}
	}

	match, err := matchPluginVersionRecord(records, entry.PluginVersion)
	if err != nil {
		return err
	}

	if match != nil {
		switch match.State {
		case enumor.IaCPluginStateEnabled, enumor.IaCPluginStateDisabled:
			return s.verifyExistingRecord(kt, entry, match, records)
		case enumor.IaCPluginStateInit, enumor.IaCPluginStateFailed:
			// 上次注册中断或失败，继续走注册重入。
		default:
			return fmt.Errorf("plugin record %s has an unknown state %q", match.ID, match.State)
		}
	}

	return s.registerVersion(kt, entry, match, records)
}

// verifyExistingRecord 校验已注册完成的同版本记录；仅当已完成最高版本为 enabled 且高于本版本时拒绝。
func (s *Service) verifyExistingRecord(kt *kit.Kit, entry discovery.PluginEntry, match *coreplugin.IaCPlugin,
	records []coreplugin.IaCPlugin) error {

	completedHighest, err := iacplugin.Latest(records, func(rec coreplugin.IaCPlugin) bool {
		return rec.State == enumor.IaCPluginStateEnabled || rec.State == enumor.IaCPluginStateDisabled
	})
	if err != nil {
		return err
	}
	if completedHighest != nil && completedHighest.State == enumor.IaCPluginStateEnabled {
		if err = rejectVersionLowerThan(entry.PluginVersion, completedHighest); err != nil {
			return err
		}
	}

	logs.Infof("plugin %s@%s is already registered as %s, verify sha256, rid: %s",
		entry.PluginName, entry.PluginVersion, match.State, kt.Rid)
	return s.verifyRecord(kt, entry, match)
}

// registerVersion 执行新版本注册或 init / failed 重入，写入前校验不得低于全量最高版本。
func (s *Service) registerVersion(kt *kit.Kit, entry discovery.PluginEntry, match *coreplugin.IaCPlugin,
	records []coreplugin.IaCPlugin) error {

	highest, err := iacplugin.Latest(records)
	if err != nil {
		return err
	}
	if err = rejectVersionLowerThan(entry.PluginVersion, highest); err != nil {
		return err
	}

	if match != nil {
		logs.Infof("plugin %s@%s previous registration state is %s, re-register, rid: %s",
			entry.PluginName, entry.PluginVersion, match.State, kt.Rid)
	} else {
		logs.Infof("plugin %s@%s is not registered, register now, rid: %s",
			entry.PluginName, entry.PluginVersion, kt.Rid)
	}

	id, err := s.saveInitRecord(kt, entry, match)
	if err != nil {
		return err
	}
	return s.syncSchemaAndUpdate(kt, entry, id, records)
}

func rejectVersionLowerThan(localVersion string, highest *coreplugin.IaCPlugin) error {
	if highest == nil {
		return nil
	}
	cmp, err := iacplugin.CompareVersion(localVersion, highest.PluginVersion)
	if err != nil {
		return err
	}
	if cmp < 0 {
		return fmt.Errorf("%w: local plugin version %s is lower than the highest registered version %s",
			errPluginVersionConflict, localVersion, highest.PluginVersion)
	}
	return nil
}

// matchPluginVersionRecord 返回版本号等于 version 的记录。没有这条记录时返回 nil。
func matchPluginVersionRecord(records []coreplugin.IaCPlugin, version string) (*coreplugin.IaCPlugin, error) {
	for i := range records {
		cmp, err := iacplugin.CompareVersion(records[i].PluginVersion, version)
		if err != nil {
			return nil, fmt.Errorf("plugin record %s has an invalid version: %w", records[i].ID, err)
		}
		if cmp == 0 {
			return cvt.ValToPtr(records[i]), nil
		}
	}
	return nil, nil
}

// saveInitRecord 把该版本的记录落库为 init，返回记录 ID 供同步流程收尾。
// 没有记录则新建；处于 init（上次中断）或 failed（上次失败）的记录则重置为 init 后重入。
func (s *Service) saveInitRecord(kt *kit.Kit, entry discovery.PluginEntry, rec *coreplugin.IaCPlugin) (string, error) {
	if rec == nil {
		return s.dataCli.Global.IaCPlugin.CreateInit(kt, &dataplugin.IaCPluginCreateReq{
			PluginName:    entry.PluginName,
			PluginVersion: entry.PluginVersion,
			Protocol:      enumor.IaCPluginProtocolTFProvider,
			State:         enumor.IaCPluginStateInit,
			Extension:     `{}`,
		})
	}

	err := s.dataCli.Global.IaCPlugin.Update(kt, &dataplugin.IaCPluginUpdateReq{
		ID:           rec.ID,
		CurrentState: rec.State,
		State:        enumor.IaCPluginStateInit,
		Message:      cvt.ValToPtr(""),
	})
	if err != nil {
		return "", fmt.Errorf("reset plugin %s@%s to init failed: %w", rec.PluginName, rec.PluginVersion, err)
	}

	return rec.ID, nil
}

// verifyRecord 已注册完成的版本只校验本架构 sha256，不拉起进程、不同步 Schema。
// 本架构一致直接通过，不一致返回冲突退出，缺失则补写本架构 hash（其它架构保持不变）。
func (s *Service) verifyRecord(kt *kit.Kit, entry discovery.PluginEntry, rec *coreplugin.IaCPlugin) error {
	if rec == nil {
		return fmt.Errorf("plugin record is nil")
	}

	var ext coreplugin.TFProviderExtension
	if len(rec.Extension) != 0 {
		if err := json.Unmarshal([]byte(rec.Extension), &ext); err != nil {
			return fmt.Errorf("plugin record %s has an invalid extension: %w", rec.ID, err)
		}
	}

	if recorded, ok := ext.Architectures[entry.Arch]; ok && recorded != "" {
		if recorded == entry.SHA256 {
			return nil
		}
		// 同一版本号下本架构的文件内容变了：不修改任何记录，进程退出。
		return fmt.Errorf("%w: plugin %s@%s sha256 of %s differs from the registered one, "+
			"bump the plugin version instead of replacing the binary", errPluginVersionConflict, entry.PluginName,
			entry.PluginVersion, entry.Arch)
	}

	// 例如首个 arm64 副本第一次启动：补写本架构的 hash，其它架构的值保持不变。
	updatedExt, err := withArchHash(rec.Extension, entry.Arch, entry.SHA256)
	if err != nil {
		return err
	}
	err = s.dataCli.Global.IaCPlugin.Update(kt, &dataplugin.IaCPluginUpdateReq{
		ID:           rec.ID,
		CurrentState: rec.State,
		State:        rec.State,
		Message:      cvt.ValToPtr(""),
		Extension:    updatedExt,
	})
	if err != nil {
		return fmt.Errorf("write sha256 of %s for plugin %s@%s failed: %w", entry.Arch,
			entry.PluginName, entry.PluginVersion, err)
	}
	return nil
}

// withArchHash 返回在 extension 基础上补写（或设置）arch 对应 sha256 后的新 extension。
// 其它架构已有的值与 extension 中的其它字段原样保留，不会被覆盖。
func withArchHash(extension tabletypes.JsonField, arch, sha256 string) (tabletypes.JsonField, error) {
	all := make(map[string]json.RawMessage)
	if len(extension) != 0 {
		if err := json.Unmarshal([]byte(extension), &all); err != nil {
			return "", fmt.Errorf("invalid extension: %w", err)
		}
	}

	archs := make(map[string]string)
	if raw, ok := all["architectures"]; ok && len(raw) != 0 && string(raw) != "null" {
		if err := json.Unmarshal(raw, &archs); err != nil {
			return "", fmt.Errorf("invalid extension.architectures: %w", err)
		}
	}
	archs[arch] = sha256

	encoded, err := json.Marshal(archs)
	if err != nil {
		return "", err
	}
	all["architectures"] = encoded

	out, err := json.Marshal(all)
	if err != nil {
		return "", err
	}

	return tabletypes.JsonField(out), nil
}

// syncSchemaAndUpdate 同步 Schema，并把记录从 init 收尾为 enabled / disabled / failed。
func (s *Service) syncSchemaAndUpdate(kt *kit.Kit, entry discovery.PluginEntry, id string,
	records []coreplugin.IaCPlugin) error {

	if err := s.syncSchema(kt, entry); err != nil {
		// 确定性失败：置 failed 并写原因。写不进去时记录保持 init，下一个副本照样会重入。
		s.markFailed(kt, entry, id, err)
		return fmt.Errorf("register plugin %s@%s failed: %w", entry.PluginName, entry.PluginVersion, err)
	}

	target := iacplugin.CalcTargetState(records, entry.PluginVersion)
	ext, err := withArchHash(`{}`, entry.Arch, entry.SHA256)
	if err != nil {
		return err
	}

	err = s.dataCli.Global.IaCPlugin.Update(kt, &dataplugin.IaCPluginUpdateReq{
		ID:           id,
		CurrentState: enumor.IaCPluginStateInit,
		State:        target,
		Message:      cvt.ValToPtr(""),
		Extension:    ext,
	})
	if err != nil {
		return fmt.Errorf("mark plugin %s@%s as %s failed: %w", entry.PluginName, entry.PluginVersion, target, err)
	}

	logs.Infof("plugin %s@%s registered, state: %s, rid: %s", entry.PluginName, entry.PluginVersion, target, kt.Rid)
	return nil
}

// syncSchema 拉起一次性进程并同步插件的 Schema，结束时无论成败均关闭进程。
func (s *Service) syncSchema(kt *kit.Kit, entry discovery.PluginEntry) error {
	manager := tfruntime.NewProcessManager(entry.PluginName, entry.ExecPath, nil, "", "")
	defer func() {
		if cerr := manager.Close(); cerr != nil {
			logs.Errorf("close plugin %s@%s process failed, err: %v, rid: %s", entry.PluginName,
				entry.PluginVersion, cerr, kt.Rid)
		}
	}()

	exec, err := manager.Start(kt.Ctx)
	if err != nil {
		return fmt.Errorf("start plugin executor failed: %w", err)
	}

	getter := &tfSchemaGetter{
		executor:      exec,
		pluginName:    entry.PluginName,
		pluginVersion: entry.PluginVersion,
	}

	result, err := s.schemaMgr.SyncSchema(kt, getter)
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

func toCorePluginSchema(in *types.ProviderSchema, pluginName, pluginVersion string) (*coreplugin.PluginSchema, error) {
	if in == nil {
		return nil, fmt.Errorf("provider schema is nil")
	}
	data, err := json.Marshal(in)
	if err != nil {
		return nil, err
	}
	var schema coreplugin.PluginSchema
	if err = json.Unmarshal(data, &schema); err != nil {
		return nil, err
	}
	schema.Plugin = coreplugin.PluginInfo{Name: pluginName, Version: pluginVersion}
	hash, err := coreplugin.ComputeSchemaHashByAlgo(enumor.DefaultSchemaHashAlgo, schema.Plugin,
		cvt.ValToPtr(schema.SchemaDetail))
	if err != nil {
		return nil, err
	}
	schema.SchemaHash = hash
	return cvt.ValToPtr(schema), nil
}

// markFailed 把记录置 failed 并写入失败原因。写库失败只记日志：记录保持 init，不影响重入。
func (s *Service) markFailed(kt *kit.Kit, entry discovery.PluginEntry, id string, cause error) {
	if err := s.dataCli.Global.IaCPlugin.MarkFailed(kt, id, enumor.IaCPluginStateInit, cause); err != nil {
		logs.Errorf("mark plugin %s@%s as failed failed, err: %v, cause: %v, rid: %s", entry.PluginName,
			entry.PluginVersion, err, cause, kt.Rid)
	}
}

// CheckRegistered 检查注册表中已启用的 tf 插件是否均存在于本地目录。
func (s *Service) CheckRegistered(kt *kit.Kit, entries map[string]discovery.PluginEntry) error {
	records, err := s.dataCli.Global.IaCPlugin.ListAll(kt,
		tools.EqualExpression("protocol", enumor.IaCPluginProtocolTFProvider), nil)
	if err != nil {
		return fmt.Errorf("list tf_provider plugins failed: %w", err)
	}

	missing, err := FindMissing(records, entries)
	if err != nil {
		return err
	}
	if len(missing) > 0 {
		return fmt.Errorf("%w: %s", errMissingPlugin, strings.Join(missing, ", "))
	}

	return nil
}

// FindMissing 返回注册表中最高版本为 enabled 但本地目录缺失的 tf 插件名列表。
func FindMissing(records []coreplugin.IaCPlugin, entries map[string]discovery.PluginEntry) ([]string, error) {
	// 按插件名分组：同一插件名可能有多条版本记录。
	byName := make(map[string][]coreplugin.IaCPlugin, len(records))
	for _, rec := range slice.Filter(records, func(rec coreplugin.IaCPlugin) bool {
		return rec.Protocol == enumor.IaCPluginProtocolTFProvider
	}) {
		byName[rec.PluginName] = append(byName[rec.PluginName], rec)
	}

	missing := make([]string, 0)
	for name, recs := range byName {
		if _, ok := entries[name]; ok {
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
