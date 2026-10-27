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
	"bytes"
	"fmt"
	"io"
	"sync"

	"hcm/pkg/logs"

	"github.com/hashicorp/go-hclog"
)

// pluginLogs writes plugin stdout and stderr to a plugin log file managed by glog.
// Framework logs stay in the hc-service log. Plugin name, version, and pid stay in the line.
type pluginLogs struct {
	mu      sync.Mutex
	name    string
	version string
	pid     string
	closed  bool
	file    io.WriteCloser

	stdout *streamBuf
	stderr *streamBuf
}

type streamBuf struct {
	parent *pluginLogs
	stream string
	part   []byte
}

func newPluginLogs(name, version string) (*pluginLogs, error) {
	if version == "" {
		version = "unknown"
	}
	file, err := logs.NewPluginLogger(name)
	if err != nil {
		return nil, err
	}
	pl := &pluginLogs{
		name:    name,
		version: version,
		file:    file,
	}
	pl.stdout = &streamBuf{parent: pl, stream: "stdout"}
	pl.stderr = &streamBuf{parent: pl, stream: "stderr"}
	return pl, nil
}

func (p *pluginLogs) logger() hclog.Logger {
	return hclog.New(&hclog.LoggerOptions{
		Name:   "tfplugin",
		Level:  hclog.Trace,
		Output: &serviceWriter{parent: p},
	})
}

func (p *pluginLogs) serviceWriter() *serviceWriter {
	return &serviceWriter{parent: p}
}

func (p *pluginLogs) setPID(pid string) error {
	p.mu.Lock()
	p.pid = pid
	p.mu.Unlock()
	return nil
}

func (p *pluginLogs) close() {
	p.mu.Lock()
	defer p.mu.Unlock()
	if p.closed {
		return
	}
	p.flushPartLocked(p.stdout)
	p.flushPartLocked(p.stderr)
	p.closed = true
	if p.file != nil {
		_ = p.file.Close()
		p.file = nil
	}
}

func (p *pluginLogs) flushPartLocked(stream *streamBuf) {
	if stream == nil || len(stream.part) == 0 {
		return
	}
	text := string(stream.part)
	stream.part = nil
	p.emitLocked(stream.stream, text)
}

func (p *pluginLogs) emitLocked(stream, text string) {
	if p.file == nil {
		return
	}
	line := fmt.Sprintf("plugin=%s plugin_version=%s pid=%s %s %s\n", p.name, p.version, p.pid, stream, text)
	if _, err := p.file.Write([]byte(line)); err != nil {
		logs.Errorf("write plugin log failed, plugin: %s, err: %v", p.name, err)
	}
}

// Write buffers plugin stdout or stderr and emits each complete line to the plugin log.
func (s *streamBuf) Write(data []byte) (int, error) {
	s.parent.mu.Lock()
	defer s.parent.mu.Unlock()
	if s.parent.closed {
		return len(data), nil
	}
	s.part = append(s.part, data...)
	for {
		idx := bytes.IndexByte(s.part, '\n')
		if idx < 0 {
			break
		}
		text := string(bytes.TrimRight(s.part[:idx], "\r"))
		s.part = append([]byte(nil), s.part[idx+1:]...)
		s.parent.emitLocked(s.stream, text)
	}
	return len(data), nil
}

// serviceWriter records go-plugin framework logs and pre-handshake stderr.
type serviceWriter struct {
	parent *pluginLogs
	part   []byte
	mu     sync.Mutex
}

// Write buffers framework logs and pre-handshake stderr, then emits each complete line to the hc-service log.
func (w *serviceWriter) Write(data []byte) (int, error) {
	w.mu.Lock()
	defer w.mu.Unlock()
	w.part = append(w.part, data...)
	for {
		idx := bytes.IndexByte(w.part, '\n')
		if idx < 0 {
			break
		}
		text := string(bytes.TrimRight(w.part[:idx], "\r"))
		w.part = append([]byte(nil), w.part[idx+1:]...)
		w.parent.writeService(text)
	}
	return len(data), nil
}

func (p *pluginLogs) writeService(text string) {
	p.mu.Lock()
	defer p.mu.Unlock()
	if p.closed {
		return
	}
	logs.Infof("plugin=%s plugin_version=%s pid=%s %s", p.name, p.version, p.pid, text)
}
