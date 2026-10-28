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

	httpplugin "hcm/cmd/hc-service/logics/plugin/http"
	pluginschema "hcm/cmd/hc-service/logics/plugin/schema"
	tfdiscovery "hcm/cmd/hc-service/logics/plugin/tf/discovery"
	"hcm/cmd/hc-service/logics/plugin/tf/register"
	"hcm/cmd/hc-service/logics/plugin/tf/watch"
	"hcm/pkg/api/core"
	"hcm/pkg/cc"
	"hcm/pkg/lock"
	"hcm/pkg/runtime/shutdown"
	cvt "hcm/pkg/tools/converter"
)

// iacPluginManager 汇总 IaC 插件在 hc-service 内的组件与运行期生命周期。
type iacPluginManager struct {
	// plugins 启动时扫描到的本副本 tf 插件（以插件名为 key），之后只读。
	plugins map[string]tfdiscovery.PluginEntry
	// ready 为 true 表示 IaC 插件已注册完成，可以接收流量；在此之前 /healthz 返回 503，/alivez 不受影响。
	ready atomic.Bool

	tfRegister *register.Service
	httpPlugin *httpplugin.Service

	// stopVersionWatch 停止运行期插件版本检查，没有本地 tf 插件时为 nil。
	stopVersionWatch context.CancelFunc
	// stopHTTPCheck 停止 http 插件巡检，未启动时为 nil。
	stopHTTPCheck context.CancelFunc
}

// initIaCPlugin 扫描本地 tf 插件目录，并初始化插件注册、Schema 同步与 http 插件。
// 目录不合法、网关 TLS 配置错误或 etcd 配置错误时返回错误，进程不应启动服务。
func (s *Service) initIaCPlugin() error {
	cfg := cc.HCService().IaCPlugin

	entries, err := tfdiscovery.Scan(cfg.TF.PluginDir)
	if err != nil {
		return fmt.Errorf("scan iac plugin dir %s failed, err: %v", cfg.TF.PluginDir, err)
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
	// TODO: 插件进程管理模块就绪后注入 PluginProcessStarter 实现，替换此处的 nil。
	tfRegister := register.New(dataCli, iacMutex, schemaMgr, nil, register.Options{
		LockLeaseSeconds: cfg.LockLeaseSeconds,
		LockWaitTimeout:  time.Duration(cfg.LockWaitTimeoutSeconds) * time.Second,
	})
	httpPlugin := httpplugin.New(dataCli, iacMutex, cvt.ValToPtr(cfg.HTTP.Gateway), schemaMgr,
		httpplugin.Options{
			Enabled:          cfg.HTTP.Enabled,
			LockLeaseSeconds: cfg.LockLeaseSeconds,
			CheckInterval:    time.Duration(cfg.HTTP.CheckIntervalSeconds) * time.Second,
			FailureThreshold: int(cfg.HTTP.CheckFailureThreshold),
		})

	s.iacPlugin = &iacPluginManager{
		plugins:    entries,
		tfRegister: tfRegister,
		httpPlugin: httpPlugin,
	}
	return nil
}

// RegisterPlugins 逐个注册本副本扫描到的 tf 插件，再检查注册表里已启用的 tf 插件本副本是否都有。
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

// StartPluginVersionWatch 启动运行期插件版本检查：本副本二进制落后于注册表已完成最高版本并持续数个周期后，通知进程优雅退出。
// 检查随服务退出停止。没有本地 tf 插件时无需检查。
func (s *Service) StartPluginVersionWatch() {
	if len(s.iacPlugin.plugins) == 0 {
		return
	}

	ctx, cancel := context.WithCancel(context.Background())
	s.iacPlugin.stopVersionWatch = cancel

	cfg := cc.HCService().IaCPlugin.TF
	w := watch.New(s.clientSet.DataService().Global.IaCPlugin, s.iacPlugin.plugins,
		shutdown.SignalShutdownGracefully,
		watch.Options{
			Interval:  time.Duration(cfg.VersionCheckIntervalSeconds) * time.Second,
			Threshold: int(cfg.VersionCheckThreshold),
		})
	go w.Run(ctx)
}

// StartHTTPPluginCheck 启动 http 插件巡检：仅 master 副本执行，随服务退出停止。未启用 http 插件时不启动。
func (s *Service) StartHTTPPluginCheck() {
	if !cc.HCService().IaCPlugin.HTTP.Enabled {
		return
	}

	ctx, cancel := context.WithCancel(context.Background())
	s.iacPlugin.stopHTTPCheck = cancel

	go s.iacPlugin.httpPlugin.RunCheck(ctx, s.sd.IsMaster)
}

// StopPluginTasks 停止运行期插件版本检查与 http 插件巡检。
func (s *Service) StopPluginTasks() {
	if s.iacPlugin.stopVersionWatch != nil {
		s.iacPlugin.stopVersionWatch()
	}
	if s.iacPlugin.stopHTTPCheck != nil {
		s.iacPlugin.stopHTTPCheck()
	}
}
