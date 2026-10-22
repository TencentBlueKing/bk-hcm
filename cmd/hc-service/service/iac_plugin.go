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
	"time"

	httpplugin "hcm/cmd/hc-service/logics/plugin/http"
	pluginschema "hcm/cmd/hc-service/logics/plugin/schema"
	"hcm/cmd/hc-service/logics/plugin/tf/discovery"
	"hcm/cmd/hc-service/logics/plugin/tf/register"
	"hcm/cmd/hc-service/logics/plugin/tf/watch"
	"hcm/pkg/api/core"
	"hcm/pkg/cc"
	"hcm/pkg/lock"
	"hcm/pkg/logs"
	"hcm/pkg/runtime/shutdown"
	cvt "hcm/pkg/tools/converter"
)

// scanPlugins 扫描本副本的 tf 插件目录。目录不合法时返回错误，进程不应启动服务。
func (s *Service) scanPlugins() error {
	pluginDir := cc.HCService().IaCPlugin.PluginDir

	entries, err := discovery.Scan(pluginDir)
	if err != nil {
		return fmt.Errorf("scan iac plugin dir %s failed, err: %v", pluginDir, err)
	}

	for _, entry := range entries {
		logs.Infof("found tf plugin %s@%s (%s), sha256: %s", entry.PluginName, entry.PluginVersion, entry.Arch,
			entry.SHA256)
	}
	s.plugins = entries

	return nil
}

// initIaCPlugin 初始化插件注册与升级共用的依赖：插件名锁、Schema 同步入口、http 插件流程。
// 网关 TLS 配置错误、etcd 配置错误时返回错误，进程不应启动服务。
func (s *Service) initIaCPlugin() error {
	cfg := cc.HCService().IaCPlugin

	etcdCfg, err := cc.HCService().Service.Etcd.ToConfig()
	if err != nil {
		return fmt.Errorf("get etcd config failed, err: %v", err)
	}
	s.iacMutex, err = lock.New(etcdCfg, cfg.LockLeaseSeconds)
	if err != nil {
		return fmt.Errorf("new etcd mutex failed, err: %v", err)
	}

	s.schemaMgr = pluginschema.NewSchemaManager(s.clientSet.DataService(), 0)
	s.httpPlugin = httpplugin.New(s.clientSet.DataService(), s.iacMutex, cvt.ValToPtr(cc.HCService().IaCGateway),
		s.schemaMgr, httpplugin.Options{
			LockLeaseSeconds: cfg.LockLeaseSeconds,
			CheckInterval:    time.Duration(cfg.HTTPCheckIntervalSeconds) * time.Second,
			FailureThreshold: int(cfg.HTTPCheckFailureThreshold),
		})

	return nil
}

// RegisterPlugins 逐个注册本副本扫描到的 tf 插件，再检查注册表里已启用的 tf 插件本副本是否都有。
// 任何一步失败都返回错误，由 Run 让进程非 0 退出。
func (s *Service) RegisterPlugins() error {
	cfg := cc.HCService().IaCPlugin

	// TODO: 插件进程管理模块就绪后注入 PluginProcessStarter 实现，替换此处的 nil。
	r := register.New(s.clientSet.DataService(), s.iacMutex, s.schemaMgr, nil,
		register.Options{
			LockLeaseSeconds: cfg.LockLeaseSeconds,
			LockWaitTimeout:  time.Duration(cfg.LockWaitTimeoutSeconds) * time.Second,
		})

	for _, entry := range s.plugins {
		if err := r.Register(core.NewBackendKit(), entry); err != nil {
			return fmt.Errorf("register tf plugin %s@%s failed, err: %v", entry.PluginName, entry.PluginVersion, err)
		}
	}

	if err := r.CheckRegistered(core.NewBackendKit(), s.plugins); err != nil {
		return fmt.Errorf("check registered tf plugins failed, err: %v", err)
	}

	return nil
}

// SetPluginReady 标记 IaC 插件已完成注册，之后 /healthz 返回 200。
func (s *Service) SetPluginReady() {
	s.pluginReady.Store(true)
}

// StartPluginVersionWatch 启动运行期插件版本检查：本副本二进制落后于注册表已完成最高版本并持续数个周期后，通知进程优雅退出。
// 检查随服务退出停止。没有本地 tf 插件时无需检查。
func (s *Service) StartPluginVersionWatch() {
	if len(s.plugins) == 0 {
		return
	}

	ctx, cancel := context.WithCancel(context.Background())
	s.stopPluginVersionWatch = cancel

	cfg := cc.HCService().IaCPlugin
	w := watch.New(s.clientSet.DataService().Global.IaCPlugin, s.plugins,
		shutdown.SignalShutdownGracefully,
		watch.Options{
			Interval:  time.Duration(cfg.VersionCheckIntervalSeconds) * time.Second,
			Threshold: int(cfg.VersionCheckThreshold),
		})
	go w.Run(ctx)
}

// StartHTTPPluginCheck 启动 http 插件巡检：仅 master 副本执行，随服务退出停止。
func (s *Service) StartHTTPPluginCheck() {
	ctx, cancel := context.WithCancel(context.Background())
	s.stopHTTPPluginCheck = cancel

	go s.httpPlugin.RunCheck(ctx, s.sd.IsMaster)
}

// StopPluginTasks 停止运行期插件版本检查与 http 插件巡检。
func (s *Service) StopPluginTasks() {
	if s.stopPluginVersionWatch != nil {
		s.stopPluginVersionWatch()
	}
	if s.stopHTTPPluginCheck != nil {
		s.stopHTTPPluginCheck()
	}
}
