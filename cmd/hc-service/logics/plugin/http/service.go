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

// Package http 实现 http 插件的注册、升级、修改与巡检，插件经蓝鲸 API 网关访问。
package http

import (
	"encoding/json"
	"errors"
	"fmt"
	"strings"
	"sync"
	"time"

	"hcm/cmd/hc-service/logics/plugin/query"
	pluginschema "hcm/cmd/hc-service/logics/plugin/schema"
	coreplugin "hcm/pkg/api/core/plugin"
	proto "hcm/pkg/api/hc-service/plugin"
	"hcm/pkg/cc"
	dataservice "hcm/pkg/client/data-service"
	"hcm/pkg/criteria/enumor"
	"hcm/pkg/criteria/errf"
	"hcm/pkg/iacplugin"
	"hcm/pkg/kit"
	"hcm/pkg/lock"
	"hcm/pkg/logs"
	"hcm/pkg/metrics"
	"hcm/pkg/thirdparty/api-gateway/plugin"
	cvt "hcm/pkg/tools/converter"

	etcd3 "go.etcd.io/etcd/client/v3"
)

// Options http 插件流程参数。
type Options struct {
	// LockLeaseSeconds 插件名锁租约（秒），需大于一次注册 / 升级的最长耗时。
	LockLeaseSeconds int64
	// CheckInterval 巡检间隔。
	CheckInterval time.Duration
	// FailureThreshold enabled 插件连续探活失败多少次后置为 failed。
	FailureThreshold int
}

// Service 实现 http 插件的注册、升级、修改与巡检。
//
// 注册 / 升级 / 修改 / 巡检与 P3 的启停共用插件名锁 lock.IaCPluginKey；接口拿不到锁立即返回，巡检跳过本轮。
// 状态推进一律走 expect_state 条件更新，审计由 data-service 在同一事务内写入，操作人取自 kit。
type Service struct {
	dataCli      *dataservice.Client
	mutex        *lock.EtcdMutex
	gateway      *cc.ApiGateway
	schemaMgr    *pluginschema.SchemaManager
	opts         Options
	pluginCliMap sync.Map
}

// New 创建 Service。schemaMgr 与 tf_provider 注册共用同一个实例。
func New(dataCli *dataservice.Client, mutex *lock.EtcdMutex, gateway *cc.ApiGateway,
	schemaMgr *pluginschema.SchemaManager, opts Options) *Service {

	return &Service{dataCli: dataCli, mutex: mutex, gateway: gateway, schemaMgr: schemaMgr, opts: opts}
}

func (s *Service) getPluginClient(apiPath string) (plugin.Client, error) {
	if cli, ok := s.pluginCliMap.Load(apiPath); ok {
		return cli.(plugin.Client), nil
	}

	baseURL := strings.TrimRight(s.gateway.Endpoints[0], "/") + apiPath
	cli, err := plugin.NewClient(baseURL, s.gateway, metrics.Register())
	if err != nil {
		return nil, errf.NewFromErr(errf.InvalidParameter, err)
	}

	actual, _ := s.pluginCliMap.LoadOrStore(apiPath, cli)
	return actual.(plugin.Client), nil
}

// Register 注册 http 插件：持锁 → 插件名未被 tf 占用 → 探活取版本 → 按版本最高记录处理 → 同步 Schema → 收尾。
// 请求参数（地址、加密配置）在进入前已校验，校验或探活失败时不写注册表。
func (s *Service) Register(kt *kit.Kit, req *proto.HTTPRegisterReq) (*proto.HTTPPluginResult, error) {
	leaseID, err := s.tryLock(kt, req.PluginName)
	if err != nil {
		return nil, err
	}
	defer s.unlock(kt, req.PluginName, leaseID)

	records, err := query.List(kt, s.dataCli, req.PluginName)
	if err != nil {
		return nil, err
	}
	for _, rec := range records {
		if rec.Protocol != enumor.IaCPluginProtocolHTTP {
			return nil, errf.Newf(errf.InvalidParameter, "plugin name %s is already used by a %s plugin",
				req.PluginName, rec.Protocol)
		}
	}

	version, err := s.probe(kt, req.APIPath)
	if err != nil {
		return nil, err
	}

	latest, err := iacplugin.Latest(records)
	if err != nil {
		return nil, err
	}

	ext, err := json.Marshal(coreplugin.HTTPExtension{APIPath: req.APIPath, CryptoType: req.CryptoType,
		CryptoConfig: req.CryptoConfig})
	if err != nil {
		return nil, err
	}

	var id string
	if latest == nil {
		id, err = query.Insert(kt, s.dataCli, req.PluginName, version, enumor.IaCPluginProtocolHTTP, ext, req.Memo)
	} else {
		if latest.State == enumor.IaCPluginStateEnabled || latest.State == enumor.IaCPluginStateDisabled {
			return nil, errf.Newf(errf.Aborted, "iac plugin %s is already registered, current version: %s, "+
				"state: %s", req.PluginName, latest.PluginVersion, latest.State)
		}

		cmp, err := iacplugin.CompareVersion(version, latest.PluginVersion)
		if err != nil {
			return nil, err
		}
		switch {
		case cmp < 0:
			return nil, errf.Newf(errf.Aborted, "iac plugin %s probed version %s is lower than the "+
				"current version %s, rollback is not supported yet", req.PluginName, version, latest.PluginVersion)
		case cmp > 0:
			id, err = query.Insert(kt, s.dataCli, req.PluginName, version, enumor.IaCPluginProtocolHTTP, ext,
				req.Memo)
		default:
			id = latest.ID
			err = query.Update(kt, s.dataCli, latest.ID, latest.State, enumor.IaCPluginStateInit, ext, req.Memo)
		}
	}
	if err != nil {
		return nil, err
	}

	state, err := s.syncSchemaAndUpdate(kt, id, req.PluginName, version, req.APIPath, records)
	if err != nil {
		return nil, err
	}

	return &proto.HTTPPluginResult{ID: id, PluginName: req.PluginName, PluginVersion: version, State: state}, nil
}

// Upgrade 管理员在插件方发版后触发升级检查，或在插件 failed 时触发恢复。按插件名作用于版本最高的记录。
func (s *Service) Upgrade(kt *kit.Kit, pluginName string) (*proto.HTTPPluginResult, error) {
	leaseID, err := s.tryLock(kt, pluginName)
	if err != nil {
		return nil, err
	}
	defer s.unlock(kt, pluginName, leaseID)

	records, latest, err := s.listAndGetLatest(kt, pluginName)
	if err != nil {
		return nil, err
	}
	if latest.Protocol != enumor.IaCPluginProtocolHTTP {
		return nil, errf.Newf(errf.InvalidParameter, "iac plugin %s is a %s plugin, upgrade it by adding a new "+
			"version directory and rolling restart hc-service", latest.PluginName, latest.Protocol)
	}

	ext, err := parseExtension(latest)
	if err != nil {
		return nil, err
	}

	version, err := s.probe(kt, ext.APIPath)
	if err != nil {
		return nil, err
	}

	return s.upgradeTo(kt, records, latest, version)
}

// upgradeTo 按探活版本 V 与版本最高记录 L 判定升级，调用方必须已持插件名锁。巡检复用。
//   - V 高于 L：插入 V 的 init 记录（extension、memo 沿用 L）并同步，L 保持原状；
//   - V 等于 L 且 L 为 init / failed：L 重入并同步；
//   - V 等于 L 且 L 为 enabled / disabled：无需处理（disabled 且 Schema 缺失时重做同步，仍为 disabled）；
//   - V 低于 L：暂不支持回滚。
func (s *Service) upgradeTo(kt *kit.Kit, records []coreplugin.IaCPlugin, latest *coreplugin.IaCPlugin,
	version string) (*proto.HTTPPluginResult, error) {

	ext, err := parseExtension(latest)
	if err != nil {
		return nil, err
	}

	cmp, err := iacplugin.CompareVersion(version, latest.PluginVersion)
	if err != nil {
		return nil, err
	}

	if cmp < 0 {
		return nil, errf.Newf(errf.Aborted, "iac plugin %s probed version %s is lower than the "+
			"current version %s, rollback is not supported yet", latest.PluginName, version, latest.PluginVersion)
	}

	res := &proto.HTTPPluginResult{
		ID:            latest.ID,
		PluginName:    latest.PluginName,
		PluginVersion: version,
		State:         latest.State,
	}
	// 探活返回了更高版本
	if cmp > 0 {
		id, err := query.Insert(kt, s.dataCli, latest.PluginName, version, enumor.IaCPluginProtocolHTTP,
			latest.Extension, latest.Memo)
		if err != nil {
			return nil, err
		}

		state, err := s.syncSchemaAndUpdate(kt, id, latest.PluginName, version, ext.APIPath, records)
		if err != nil {
			return nil, err
		}
		res.ID = id
		res.State = state
		return res, nil
	}
	// 探活返回了相同版本
	switch latest.State {
	case enumor.IaCPluginStateInit, enumor.IaCPluginStateFailed:
		err = query.Update(kt, s.dataCli, latest.ID, latest.State, enumor.IaCPluginStateInit, nil, nil)
		if err != nil {
			return nil, err
		}

		res.State, err = s.syncSchemaAndUpdate(kt, latest.ID, latest.PluginName, version, ext.APIPath, records)
		if err != nil {
			return nil, err
		}

	case enumor.IaCPluginStateDisabled:
		// 停用插件的 Schema 可能缺失，需要同步。
		exists, err := query.SchemaExists(kt, s.dataCli, latest.PluginName, latest.PluginVersion)
		if err != nil {
			logs.Errorf("list iac plugin schema failed, err: %v, plugin_name: %s, plugin_version: %s, rid: %s",
				err, latest.PluginName, latest.PluginVersion, kt.Rid)
			return nil, err
		}
		if !exists {
			if err = s.syncSchema(kt, latest.PluginName, version, ext.APIPath); err != nil {
				return nil, err
			}
		}
	}

	return res, nil
}

// Update 修改 http 插件，作用于该插件名下版本最高的记录。
//   - 只改加密配置 / 备注：校验后直接条件更新，不改状态、不同步；
//   - 改 api_path：探活新地址，版本更高走升级（新记录用新地址），相同则重新同步成功后再更新，更低拒绝。
func (s *Service) Update(kt *kit.Kit, req *proto.HTTPUpdateReq) (*proto.HTTPPluginResult, error) {
	leaseID, err := s.tryLock(kt, req.PluginName)
	if err != nil {
		return nil, err
	}
	defer s.unlock(kt, req.PluginName, leaseID)

	records, latest, err := s.listAndGetLatest(kt, req.PluginName)
	if err != nil {
		return nil, err
	}
	if latest.Protocol != enumor.IaCPluginProtocolHTTP {
		return nil, errf.Newf(errf.InvalidParameter, "iac plugin %s is a %s plugin, its config is decided by "+
			"the plugin directory and the global config", latest.PluginName, latest.Protocol)
	}

	if latest.State != enumor.IaCPluginStateEnabled && latest.State != enumor.IaCPluginStateDisabled {
		return nil, errf.Newf(errf.Aborted, "iac plugin %s@%s is in %s state, only enabled or disabled plugin "+
			"can be updated", latest.PluginName, latest.PluginVersion, latest.State)
	}

	curExt, err := parseExtension(latest)
	if err != nil {
		return nil, err
	}

	newExt, err := buildExtension(curExt, req)
	if err != nil {
		return nil, err
	}
	newExtRaw, err := json.Marshal(newExt)
	if err != nil {
		return nil, err
	}

	// api_path 未变更：无需探活和同步 Schema，直接更新 extension 与 memo。
	if newExt.APIPath == curExt.APIPath {
		if err = query.Update(kt, s.dataCli, latest.ID, latest.State, latest.State, newExtRaw,
			req.Memo); err != nil {
			return nil, err
		}
		return &proto.HTTPPluginResult{ID: latest.ID, PluginName: latest.PluginName,
			PluginVersion: latest.PluginVersion, State: latest.State}, nil
	}

	return s.updateAPIPath(kt, records, latest, newExt.APIPath, newExtRaw, req.Memo)
}

// updateAPIPath 处理 api_path 变更：探活新地址，版本相同则先同步再改原记录，版本更高则插入新版本记录走升级。
func (s *Service) updateAPIPath(kt *kit.Kit, records []coreplugin.IaCPlugin, latest *coreplugin.IaCPlugin,
	apiPath string, extRaw json.RawMessage, reqMemo *string) (*proto.HTTPPluginResult, error) {

	version, err := s.probe(kt, apiPath)
	if err != nil {
		return nil, err
	}

	cmp, err := iacplugin.CompareVersion(version, latest.PluginVersion)
	if err != nil {
		return nil, err
	}

	if cmp < 0 {
		return nil, errf.Newf(errf.Aborted, "iac plugin %s probed version %s is lower than the "+
			"current version %s, rollback is not supported yet", latest.PluginName, version, latest.PluginVersion)
	}

	if cmp == 0 {
		// 版本不变但地址变了：先对新地址同步 Schema，成功才改记录；失败则记录原样保留。
		if err = s.syncSchema(kt, latest.PluginName, version, apiPath); err != nil {
			return nil, err
		}
		if err = query.Update(kt, s.dataCli, latest.ID, latest.State, latest.State, extRaw,
			reqMemo); err != nil {
			return nil, err
		}
		return &proto.HTTPPluginResult{ID: latest.ID, PluginName: latest.PluginName,
			PluginVersion: latest.PluginVersion, State: latest.State}, nil
	}

	// 新地址返回了更高版本：按升级处理，新版本记录用合并后的 extension 与 memo。
	memo := latest.Memo
	if reqMemo != nil {
		memo = reqMemo
	}
	newID, err := query.Insert(kt, s.dataCli, latest.PluginName, version, enumor.IaCPluginProtocolHTTP,
		extRaw, memo)
	if err != nil {
		return nil, err
	}

	state, err := s.syncSchemaAndUpdate(kt, newID, latest.PluginName, version, apiPath, records)
	if err != nil {
		return nil, err
	}

	return &proto.HTTPPluginResult{ID: newID, PluginName: latest.PluginName, PluginVersion: version, State: state}, nil
}

// buildExtension 把未缺省的字段合入当前 extension，并在改了加密字段时做完整加密校验。
func buildExtension(cur *coreplugin.HTTPExtension, req *proto.HTTPUpdateReq) (*coreplugin.HTTPExtension, error) {
	merged := *cur
	if req.APIPath != nil {
		merged.APIPath = cvt.PtrToVal(req.APIPath)
	}
	if req.CryptoType != nil {
		merged.CryptoType = cvt.PtrToVal(req.CryptoType)
	}
	if req.CryptoConfig != nil {
		merged.CryptoConfig = req.CryptoConfig
	}

	if req.CryptoType != nil || req.CryptoConfig != nil {
		if err := iacplugin.ValidateCrypto(merged.CryptoType, merged.CryptoConfig); err != nil {
			return nil, errf.NewFromErr(errf.InvalidParameter, err)
		}
	}

	return &merged, nil
}

// syncSchemaAndUpdate 同步 Schema 并把 init 记录收尾：成功按上一版本继承 enabled / disabled，失败置 failed 并返回错误。
// records 是落 init 之前读到的该插件名下全部记录，用于状态继承。
func (s *Service) syncSchemaAndUpdate(kt *kit.Kit, id, pluginName, version, apiPath string,
	records []coreplugin.IaCPlugin) (enumor.IaCPluginState, error) {

	if err := s.syncSchema(kt, pluginName, version, apiPath); err != nil {
		// 写不进 failed 时记录保持 init，下次注册 / 升级 / 巡检照样会重入。
		if merr := query.MarkFailed(kt, s.dataCli, id, enumor.IaCPluginStateInit, err); merr != nil {
			logs.Errorf("mark iac plugin %s@%s as failed failed, err: %v, cause: %v, rid: %s", pluginName, version,
				merr, err, kt.Rid)
		}
		return "", err
	}

	state := iacplugin.InheritState(records, version)
	err := query.Update(kt, s.dataCli, id, enumor.IaCPluginStateInit, state, nil, nil)
	if err != nil {
		logs.Errorf("mark iac plugin %s@%s as %s failed, err: %v, rid: %s", pluginName, version, state, err, kt.Rid)
		return "", err
	}

	logs.Infof("iac plugin %s@%s registered, state: %s, rid: %s", pluginName, version, state, kt.Rid)
	return state, nil
}

// syncSchema 经网关从插件取 Schema 并交给 SchemaManager 校验落库。
func (s *Service) syncSchema(kt *kit.Kit, pluginName, version, apiPath string) error {
	result, err := s.schemaMgr.SyncSchema(kt, s.schemaGetter(apiPath))
	if err != nil {
		logs.Errorf("sync iac plugin %s@%s schema failed, err: %v, rid: %s", pluginName, version, err, kt.Rid)
		return errf.Newf(errf.Aborted, "sync iac plugin %s@%s schema failed, err: %v", pluginName, version, err)
	}
	if result == nil || result.SyncResult != enumor.SchemaSyncSuccess {
		msg := "sync plugin schema failed"
		if result != nil && len(result.SyncMessage) != 0 {
			msg = result.SyncMessage
		}
		logs.Errorf("sync iac plugin %s@%s schema failed, msg: %s, rid: %s", pluginName, version, msg, kt.Rid)
		return errf.Newf(errf.Aborted, "sync iac plugin %s@%s schema failed, err: %s", pluginName, version, msg)
	}

	return nil
}

// schemaGetter 返回按 apiPath 取 Schema 的方法。
func (s *Service) schemaGetter(apiPath string) pluginschema.SchemaGetter {
	return pluginschema.SchemaGetterFunc(func(kt *kit.Kit) (*coreplugin.PluginSchema, error) {
		cli, err := s.getPluginClient(apiPath)
		if err != nil {
			return nil, err
		}

		schema, err := cli.GetSchema(kt)
		if err != nil {
			return nil, errf.Newf(errf.Aborted, "get http plugin schema failed, err: %v", err)
		}

		coreSchema, err := toCorePluginSchema(schema)
		if err != nil {
			return nil, errf.Newf(errf.Aborted, "convert http plugin schema failed, err: %v", err)
		}
		return coreSchema, nil
	})
}

func toCorePluginSchema(in *plugin.PluginSchema) (*coreplugin.PluginSchema, error) {
	capsRaw, err := json.Marshal(in.Capabilities)
	if err != nil {
		return nil, err
	}
	var caps map[string]bool
	if err = json.Unmarshal(capsRaw, &caps); err != nil {
		return nil, err
	}

	return &coreplugin.PluginSchema{
		PluginName:    in.Plugin.Name,
		PluginVersion: in.Plugin.Version,
		SchemaHash:    in.SchemaHash,
		RawSchema: coreplugin.RawSchema{
			ProtocolVersion:         in.ProtocolVersion,
			Capabilities:            caps,
			Provider:                in.ConfigSchema,
			ResourceSchemas:         in.ResourceSchemas,
			DataSourceSchemas:       in.DataSourceSchemas,
			ResourceIdentitySchemas: in.ResourceIdentitySchemas,
		},
	}, nil
}

// probe 探活取版本，并校验版本号格式。当网关未配置时打印告警并返回默认版本，便于本地联调。
func (s *Service) probe(kt *kit.Kit, apiPath string) (string, error) {
	if err := iacplugin.ValidateAPIPath(apiPath); err != nil {
		return "", errf.NewFromErr(errf.InvalidParameter, err)
	}

	version := "1.0.0"
	if s.gateway == nil || len(s.gateway.Endpoints) == 0 {
		logs.Warnf("iacGateway.endpoints is empty, return stub plugin version for api_path: %s, rid: %s",
			apiPath, kt.Rid)
	} else {
		cli, err := s.getPluginClient(apiPath)
		if err != nil {
			return "", err
		}

		health, err := cli.Health(kt)
		if err != nil {
			return "", errf.Newf(errf.Aborted, "probe http plugin failed, err: %v", err)
		}
		if health == nil || len(health.PluginVersion) == 0 {
			return "", errf.New(errf.Aborted, "plugin probe response has no plugin_version")
		}
		version = health.PluginVersion
	}

	if _, err := iacplugin.ParseVersion(version); err != nil {
		return "", errf.Newf(errf.Aborted, "plugin probe returned an invalid version: %v", err)
	}

	return version, nil
}

func (s *Service) listAndGetLatest(kt *kit.Kit, pluginName string) ([]coreplugin.IaCPlugin, *coreplugin.IaCPlugin, error) {
	records, err := query.List(kt, s.dataCli, pluginName)
	if err != nil {
		return nil, nil, err
	}

	latest, err := iacplugin.Latest(records)
	if err != nil {
		return nil, nil, err
	}
	if latest == nil {
		return nil, nil, errf.Newf(errf.RecordNotFound, "iac plugin %s not found", pluginName)
	}

	return records, latest, nil
}

func (s *Service) tryLock(kt *kit.Kit, pluginName string) (etcd3.LeaseID, error) {
	leaseID, err := s.mutex.TryLockWithTTL(lock.IaCPluginKey(pluginName), s.opts.LockLeaseSeconds)
	if err != nil {
		if errors.Is(err, lock.ErrLockFailed) {
			return 0, errf.Newf(errf.Aborted,
				"iac plugin %s is being registered, upgraded or modified, please retry later", pluginName)
		}
		logs.Errorf("try lock iac plugin %s failed, err: %v, rid: %s", pluginName, err, kt.Rid)
		return 0, err
	}

	return leaseID, nil
}

func (s *Service) unlock(kt *kit.Kit, pluginName string, leaseID etcd3.LeaseID) {
	if err := s.mutex.UnLock(leaseID); err != nil {
		// 租约已过期：锁已自动释放。
		if strings.Contains(err.Error(), "requested lease not found") {
			return
		}
		logs.Errorf("unlock iac plugin %s failed, err: %v, lease_id: %d, rid: %s", pluginName, err, leaseID, kt.Rid)
	}
}

func parseExtension(rec *coreplugin.IaCPlugin) (*coreplugin.HTTPExtension, error) {
	ext := new(coreplugin.HTTPExtension)
	if err := json.Unmarshal(rec.Extension, ext); err != nil {
		return nil, fmt.Errorf("parse iac plugin %s@%s extension failed: %w", rec.PluginName, rec.PluginVersion, err)
	}

	return ext, nil
}
