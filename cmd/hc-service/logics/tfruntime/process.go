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

package tfruntime

import (
	"context"
	"fmt"
	"os"
	"os/exec"
	"strconv"
	"sync"

	"hcm/pkg/cc"
	"hcm/pkg/criteria/constant"
	"hcm/pkg/logs"
	"hcm/pkg/runtime/shutdown"

	"github.com/hashicorp/go-plugin"
)

var allowedTFLogs = map[string]struct{}{
	"TRACE": {},
	"DEBUG": {},
	"INFO":  {},
	"WARN":  {},
	"ERROR": {},
}

var errPluginStarted = fmt.Errorf("plugin process has already been started")

// ProcessManager owns one plugin process.
type ProcessManager struct {
	mu           sync.Mutex
	shutdownOnce sync.Once

	// name is written into log lines and the log file name.
	name string
	// path is the plugin executable.
	path string
	// args and dir are passed to the plugin command.
	args []string
	dir  string
	// tfLog is TF_LOG for the process started by this manager. Empty means ERROR.
	tfLog string

	executor *PluginExecutor
	starting bool
	// started stays set after a successful Start. Close does not clear it.
	started bool
}

// NewProcessManager returns a manager that launches the given plugin executable.
// The provider release version is not part of protocol 5, so log lines record it as unknown.
func NewProcessManager(name, path string, args []string, dir, tfLog string) *ProcessManager {
	return &ProcessManager{
		name:  name,
		path:  path,
		args:  args,
		dir:   dir,
		tfLog: tfLog,
	}
}

// Start launches this manager's plugin process once. Another call returns an
// error and does not create a plugin.Client. This manager stops the process
// in Close.
func (m *ProcessManager) Start(ctx context.Context) (*PluginExecutor, error) {
	if err := ctx.Err(); err != nil {
		return nil, err
	}
	if m.path == "" {
		return nil, fmt.Errorf("plugin executable path is empty")
	}
	if err := m.beginStart(); err != nil {
		return nil, err
	}
	succeeded := false
	defer func() {
		if succeeded {
			return
		}
		m.mu.Lock()
		m.starting = false
		m.mu.Unlock()
	}()

	executor, err := m.launch(ctx)
	if err != nil {
		return nil, err
	}
	m.mu.Lock()
	m.executor = executor
	m.starting = false
	m.started = true
	m.mu.Unlock()
	succeeded = true
	m.watchShutdown()
	return executor, nil
}

func (m *ProcessManager) beginStart() error {
	m.mu.Lock()
	defer m.mu.Unlock()
	if m.started || m.starting {
		return errPluginStarted
	}
	m.starting = true
	return nil
}

func (m *ProcessManager) launch(ctx context.Context) (*PluginExecutor, error) {
	env, err := pluginEnv(m.tfLog)
	if err != nil {
		return nil, err
	}
	cmd := exec.Command(m.path, m.args...)
	cmd.Dir = m.dir
	cmd.Env = env

	// TODO version 应该来自 GetSchema
	sink, err := newPluginLogs(m.name, "")
	if err != nil {
		return nil, err
	}
	minPort, maxPort := pluginPorts()
	cfg := newPluginClientConfig(cmd, minPort, maxPort)
	cfg.Stderr = sink.serviceWriter()
	cfg.SyncStdout = sink.stdout
	cfg.SyncStderr = sink.stderr
	cfg.Logger = sink.logger()

	client := plugin.NewClient(cfg)
	executor, err := connectPlugin(ctx, client, sink)
	if err != nil {
		client.Kill()
		sink.close()
		return nil, err
	}
	executor.attach(m, client, sink, m.name)
	return executor, nil
}

func connectPlugin(ctx context.Context, client *plugin.Client, sink *pluginLogs) (*PluginExecutor, error) {
	rpcClient, err := client.Client()
	if err != nil {
		return nil, err
	}
	if client.NegotiatedVersion() != constant.TFPluginProtocolVersion {
		version := client.NegotiatedVersion()
		return nil, fmt.Errorf("negotiated plugin protocol %d, want %d", version, constant.TFPluginProtocolVersion)
	}
	pid, err := pluginPID(client)
	if err != nil {
		return nil, err
	}
	if err = sink.setPID(pid); err != nil {
		return nil, err
	}
	raw, err := rpcClient.Dispense(constant.TFPluginProviderName)
	if err != nil {
		return nil, err
	}
	executor, ok := raw.(*PluginExecutor)
	if !ok {
		return nil, fmt.Errorf("plugin %s returned %T", constant.TFPluginProviderName, raw)
	}
	if err = ctx.Err(); err != nil {
		return nil, err
	}
	return executor, nil
}

func pluginPID(client *plugin.Client) (string, error) {
	reattach := client.ReattachConfig()
	if reattach == nil || reattach.Addr == nil || reattach.Addr.Network() != "unix" {
		return "", fmt.Errorf("plugin connection is not a local unix socket")
	}
	pid := client.ID()
	if pid == "" && reattach.Pid > 0 {
		pid = strconv.Itoa(reattach.Pid)
	}
	return pid, nil
}

// Close stops this manager's process. It returns immediately when none is running.
func (m *ProcessManager) Close() error {
	m.mu.Lock()
	executor := m.executor
	m.mu.Unlock()
	if executor == nil {
		return nil
	}
	return executor.Close()
}

func (m *ProcessManager) detach(executor *PluginExecutor) {
	m.mu.Lock()
	if m.executor == executor {
		m.executor = nil
	}
	m.mu.Unlock()
}

// pluginEnv builds the explicit environment for one plugin process.
// It does not copy os.Environ. go-plugin appends the magic cookie itself.
func pluginEnv(tfLog string) ([]string, error) {
	level := tfLog
	if level == "" {
		level = "ERROR"
	}
	if _, ok := allowedTFLogs[level]; !ok {
		return nil, fmt.Errorf("invalid TF_LOG %q", tfLog)
	}

	env := make([]string, 0, 4)
	for _, key := range []string{"PATH", "TMPDIR"} {
		if value, ok := os.LookupEnv(key); ok {
			env = append(env, key+"="+value)
		}
	}
	env = append(env,
		"TF_LOG="+level,
		"PLUGIN_PROTOCOL_VERSIONS="+strconv.Itoa(constant.TFPluginProtocolVersion),
	)
	return env, nil
}

// pluginPorts returns the configured PLUGIN_MIN_PORT / PLUGIN_MAX_PORT range.
func pluginPorts() (uint, uint) {
	cfg := cc.HCService().TFPlugin
	return cfg.MinPort, cfg.MaxPort
}

// watchShutdown closes this manager's process when hc-service shuts down.
// A manager that has not started does not register. After Close, the callback
// returns immediately and still calls Done.
func (m *ProcessManager) watchShutdown() {
	m.shutdownOnce.Do(func() {
		notifier := shutdown.AddNotifier()
		go func() {
			<-notifier.Signal
			if err := m.Close(); err != nil {
				logs.Errorf("close plugin on shutdown: plugin=%s err=%v", m.name, err)
			}
			notifier.Done()
		}()
	})
}
