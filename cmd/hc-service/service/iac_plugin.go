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

package service

import (
	"context"
	"fmt"
	"sync/atomic"
	"time"

	"hcm/cmd/hc-service/logics/iac-plugin/check"
	httpplugin "hcm/cmd/hc-service/logics/iac-plugin/http-plugin"
	"hcm/cmd/hc-service/logics/iac-plugin/tf-plugin/discovery"
	"hcm/cmd/hc-service/logics/iac-plugin/tf-plugin/register"
	pluginschema "hcm/cmd/hc-service/logics/plugin/schema"
	"hcm/pkg/api/core"
	"hcm/pkg/cc"
	"hcm/pkg/lock"
	"hcm/pkg/logs"
	"hcm/pkg/runtime/shutdown"
	cvt "hcm/pkg/tools/converter"
)

// iacPluginManager 汇总 IaC 插件在 hc-service 内的组件与运行期生命周期。
type iacPluginManager struct {
	// plugins 启动时从本地目录扫描到的 tf 插件（以插件名为 key），之后只读。
	plugins map[string]discovery.PluginEntry
	// ready 为 true 表示 IaC 插件已注册完成，可以接收流量；在此之前 /healthz 返回 503，/alivez 不受影响。
	ready atomic.Bool

	iacMutex   *lock.EtcdMutex
	tfRegister *register.Service
	httpPlugin *httpplugin.Service

	// checkers 运行期周期检查实现列表（tf 插件版本检查、http 插件巡检）。
	checkers []check.Checker
	// stopCheckers 停止所有运行期检查，未启动时为 nil。
	stopCheckers context.CancelFunc
}

// initIaCPlugin 扫描本地 tf 插件目录，并初始化插件注册、Schema 同步与 http 插件。
// 目录不合法、网关 TLS 配置错误或 etcd 配置错误时返回错误，进程不应启动服务。
func (s *Service) initIaCPlugin() error {
	cfg := cc.HCService().IaCPlugin

	entries, err := discovery.Scan(cfg.TF.PluginDir)
	if err != nil {
		return fmt.Errorf("scan iac plugin dir %s failed, err: %v", cfg.TF.PluginDir, err)
	}
	for _, entry := range entries {
		logs.Infof("found tf plugin %s@%s (%s), sha256: %s", entry.PluginName, entry.PluginVersion, entry.Arch,
			entry.SHA256)
	}

	etcdCfg, err := cc.HCService().Service.Etcd.ToConfig()
	if err != nil {
		return fmt.Errorf("get etcd config failed, err: %v", err)
	}
	iacMutex, err := lock.New(etcdCfg, cfg.LockLeaseSeconds)
	if err != nil {
		return fmt.Errorf("new etcd mutex failed, err: %v", err)
	}

	dataCli := s.clientSet.DataService()
	schemaMgr := pluginschema.NewSchemaManager(dataCli, 0)
	tfRegister := register.New(dataCli, iacMutex, schemaMgr, register.Options{
		LockLeaseSeconds: cfg.LockLeaseSeconds,
		LockWaitTimeout:  time.Duration(cfg.LockWaitTimeoutSeconds) * time.Second,
	})
	httpPlugin := httpplugin.New(dataCli, iacMutex, cvt.ValToPtr(cfg.HTTP.Gateway), schemaMgr, httpplugin.Options{
		Enabled:          cfg.HTTP.Enabled,
		LockLeaseSeconds: cfg.LockLeaseSeconds,
	})

	checkers := make([]check.Checker, 0, 2)
	if len(entries) > 0 {
		checkers = append(checkers, check.NewTFChecker(dataCli.Global.IaCPlugin, entries,
			shutdown.SignalShutdownGracefully, check.TFOptions{
				Interval:  time.Duration(cfg.TF.VersionCheckIntervalSeconds) * time.Second,
				Threshold: int(cfg.TF.VersionCheckThreshold),
			}))
	}
	if cfg.HTTP.Enabled {
		checkers = append(checkers, check.NewHTTPChecker(dataCli.Global.IaCPlugin, iacMutex, httpPlugin,
			s.sd.IsMaster, check.HTTPOptions{
				Interval:         time.Duration(cfg.HTTP.CheckIntervalSeconds) * time.Second,
				FailureThreshold: int(cfg.HTTP.CheckFailureThreshold),
				LockLeaseSeconds: cfg.LockLeaseSeconds,
			}))
	}

	s.iacPlugin = &iacPluginManager{
		plugins:    entries,
		iacMutex:   iacMutex,
		tfRegister: tfRegister,
		httpPlugin: httpPlugin,
		checkers:   checkers,
	}
	return nil
}

// RegisterPlugins 注册本地目录扫描到的 tf 插件，并校验注册表中已启用的 tf 插件本地是否齐全。
// 任何一步失败都返回错误，由 Run 让进程非 0 退出。
func (s *Service) RegisterPlugins() error {
	for _, entry := range s.iacPlugin.plugins {
		if err := s.iacPlugin.tfRegister.Register(core.NewBackendKit(), entry); err != nil {
			return fmt.Errorf("register tf plugin %s@%s failed, err: %v", entry.PluginName, entry.PluginVersion, err)
		}
	}

	if err := s.iacPlugin.tfRegister.CheckRegistered(core.NewBackendKit(), s.iacPlugin.plugins); err != nil {
		return fmt.Errorf("check registered tf plugins failed, err: %v", err)
	}

	return nil
}

// SetPluginReady 标记 IaC 插件已完成注册，之后 /healthz 返回 200。
func (s *Service) SetPluginReady() {
	s.iacPlugin.ready.Store(true)
}

// StartPluginTasks 启动所有插件运行期检查任务（tf 插件版本检查、http 插件巡检），随服务退出停止。
func (s *Service) StartPluginTasks() {
	if len(s.iacPlugin.checkers) == 0 {
		return
	}

	ctx, cancel := context.WithCancel(context.Background())
	s.iacPlugin.stopCheckers = cancel

	for _, c := range s.iacPlugin.checkers {
		go check.Run(ctx, c)
	}
}

// StopPluginTasks 停止运行期插件检查任务并关闭插件锁客户端。
func (s *Service) StopPluginTasks() {
	if s.iacPlugin.stopCheckers != nil {
		s.iacPlugin.stopCheckers()
	}
	if s.iacPlugin.iacMutex != nil {
		s.iacPlugin.iacMutex.Close()
	}
}
