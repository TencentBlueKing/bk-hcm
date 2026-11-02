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

package glog

import (
	"bufio"
	"bytes"
	"fmt"
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"sync"
	"time"
)

var pluginLogMu sync.Mutex

var (
	pluginOnce   sync.Once
	pluginKeeper *logKeeper
	pluginDirErr error
)

// PluginLogger writes one plugin to its own log file.
// File name, size limit, and retention use the existing glog settings.
type PluginLogger struct {
	plugin string
	file   *os.File
	w      *bufio.Writer
	nbytes uint32
	closed bool
}

// NewPluginLogger opens the plugin log file under {logDir}/plugin.
func NewPluginLogger(pluginName string) (*PluginLogger, error) {
	pluginLogMu.Lock()
	defer pluginLogMu.Unlock()
	if err := ensurePluginDir(); err != nil {
		return nil, err
	}
	logger := &PluginLogger{plugin: sanitizeLogName(pluginName)}
	if err := logger.reopen(time.Now()); err != nil {
		return nil, err
	}
	return logger, nil
}

// Write appends plugin output and rotates when the file reaches MaxSize.
func (p *PluginLogger) Write(data []byte) (int, error) {
	pluginLogMu.Lock()
	defer pluginLogMu.Unlock()
	if p.closed {
		return len(data), nil
	}
	if p.nbytes+uint32(len(data)) >= MaxSize() {
		if err := p.reopen(time.Now()); err != nil {
			return 0, err
		}
	}
	n, err := p.w.Write(data)
	p.nbytes += uint32(n)
	if err != nil {
		return n, err
	}
	if err = p.w.Flush(); err != nil {
		return n, err
	}
	return n, nil
}

// Close flushes and closes this file. The file stays on disk.
func (p *PluginLogger) Close() error {
	pluginLogMu.Lock()
	defer pluginLogMu.Unlock()
	if p.closed {
		return nil
	}
	p.closed = true
	if p.w != nil {
		_ = p.w.Flush()
	}
	if p.file != nil {
		err := p.file.Close()
		p.file = nil
		return err
	}
	return nil
}

func (p *PluginLogger) reopen(now time.Time) error {
	if p.w != nil {
		_ = p.w.Flush()
	}
	if p.file != nil {
		_ = p.file.Close()
		p.file = nil
	}
	appended := false
	pluginKeeper.onceLoad.Do(func() {
		pluginKeeper.load()
		appended = p.openTail()
	})
	if appended {
		return nil
	}
	return p.create(now)
}

func (p *PluginLogger) openTail() bool {
	if !IsRestartLogScrolling() || pluginKeeper.tail == nil {
		return false
	}
	filename := filepath.Join(pluginKeeper.dir, pluginKeeper.tail.name)
	info, err := os.Stat(filename)
	if err != nil || uint32(info.Size()) >= MaxSize() {
		return false
	}
	file, err := os.OpenFile(filename, os.O_WRONLY|os.O_APPEND, 0o666)
	if err != nil {
		return false
	}
	n, err := writePluginHeader(file, time.Now())
	if err != nil {
		_ = file.Close()
		return false
	}
	p.file = file
	p.w = bufio.NewWriterSize(file, bufferSize)
	p.nbytes = uint32(info.Size()) + n
	return true
}

func (p *PluginLogger) create(now time.Time) error {
	name, link := pluginLogName(p.plugin, now)
	file, err := os.Create(filepath.Join(pluginKeeper.dir, name))
	if err != nil {
		return err
	}
	n, err := writePluginHeader(file, now)
	if err != nil {
		_ = file.Close()
		return err
	}
	p.file = file
	p.w = bufio.NewWriterSize(file, bufferSize)
	p.nbytes = n
	pluginKeeper.add(&fileBlock{fileInfo: fileInfo{name: name, timestamp: ""}, next: nil})
	symlink := filepath.Join(pluginKeeper.dir, link)
	_ = os.Remove(symlink)
	_ = os.Symlink(name, symlink)
	return nil
}

func ensurePluginDir() error {
	pluginOnce.Do(func() {
		dir := filepath.Join(logDir, "plugin")
		pluginDirErr = os.MkdirAll(dir, 0o755)
		pluginKeeper = &logKeeper{dir: dir}
	})
	return pluginDirErr
}

func pluginLogName(plugin string, t time.Time) (name, link string) {
	name = fmt.Sprintf("%s.%s.%s.%s.log.%s.%04d%02d%02d-%02d%02d%02d.%d",
		program, plugin, host, userName, logFileTag,
		t.Year(), t.Month(), t.Day(), t.Hour(), t.Minute(), t.Second(), pid)
	return name, program + "." + logFileTag
}

func writePluginHeader(file *os.File, now time.Time) (uint32, error) {
	var buf bytes.Buffer
	fmt.Fprintf(&buf, "Log file created at: %s\n", now.Format("2006/01/02 15:04:05"))
	fmt.Fprintf(&buf, "Running on machine: %s\n", host)
	fmt.Fprintf(&buf, "Binary: Built with %s %s for %s/%s\n", runtime.Compiler, runtime.Version(),
		runtime.GOOS, runtime.GOARCH)
	fmt.Fprintf(&buf, "Log line format: [IWEF]mmdd hh:mm:ss.uuuuuu threadid file:line] msg\n")
	n, err := file.Write(buf.Bytes())
	return uint32(n), err
}

func sanitizeLogName(value string) string {
	if value == "" {
		return "unknown"
	}
	var b strings.Builder
	for _, r := range value {
		switch {
		case r >= 'a' && r <= 'z', r >= 'A' && r <= 'Z', r >= '0' && r <= '9', r == '.', r == '_', r == '-':
			b.WriteRune(r)
		default:
			b.WriteByte('_')
		}
	}
	if b.Len() == 0 {
		return "unknown"
	}
	return b.String()
}
