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
	pluginOnce    sync.Once
	pluginRoot    string
	pluginRootErr error
	pluginSinks   map[string]*pluginSink
)

// pluginSink is the current log file of one plugin name.
// Processes of that name append to this file. Rotation follows glog.
type pluginSink struct {
	keeper *logKeeper
	file   *os.File
	w      *bufio.Writer
	nbytes uint32
	refs   int
}

// PluginLogger writes to the shared log file of one plugin name.
type PluginLogger struct {
	sink   *pluginSink
	closed bool
}

// NewPluginLogger opens the shared log file of pluginName under {logDir}/plugin/{name}.
func NewPluginLogger(pluginName string) (*PluginLogger, error) {
	pluginLogMu.Lock()
	defer pluginLogMu.Unlock()
	if err := ensurePluginRoot(); err != nil {
		return nil, err
	}
	name := sanitizeLogName(pluginName)
	sink := pluginSinks[name]
	if sink == nil {
		dir := filepath.Join(pluginRoot, name)
		if err := os.MkdirAll(dir, 0o755); err != nil {
			return nil, err
		}
		sink = &pluginSink{keeper: &logKeeper{dir: dir}}
		pluginSinks[name] = sink
	}
	if err := sink.open(time.Now()); err != nil {
		return nil, err
	}
	sink.refs++
	return &PluginLogger{sink: sink}, nil
}

// Write appends plugin output and rotates when the file reaches MaxSize.
func (p *PluginLogger) Write(data []byte) (int, error) {
	pluginLogMu.Lock()
	defer pluginLogMu.Unlock()
	if p.closed {
		return len(data), nil
	}
	sink := p.sink
	if sink.nbytes+uint32(len(data)) >= MaxSize() {
		if err := sink.rotate(time.Now()); err != nil {
			return 0, err
		}
	}
	n, err := sink.w.Write(data)
	sink.nbytes += uint32(n)
	if err != nil {
		return n, err
	}
	if err = sink.w.Flush(); err != nil {
		return n, err
	}
	return n, nil
}

// Close flushes this logger. The file stays open while another logger of the same plugin still uses it.
func (p *PluginLogger) Close() error {
	pluginLogMu.Lock()
	defer pluginLogMu.Unlock()
	if p.closed {
		return nil
	}
	p.closed = true
	sink := p.sink
	sink.refs--
	if sink.refs > 0 {
		return nil
	}
	if sink.w != nil {
		_ = sink.w.Flush()
		sink.w = nil
	}
	if sink.file != nil {
		err := sink.file.Close()
		sink.file = nil
		return err
	}
	return nil
}

func (s *pluginSink) open(now time.Time) error {
	if s.file != nil {
		return nil
	}
	first := false
	s.keeper.onceLoad.Do(func() {
		s.keeper.load()
		first = true
	})
	if first && !IsRestartLogScrolling() {
		return s.create(now)
	}
	if s.openTail() {
		return nil
	}
	return s.create(now)
}

func (s *pluginSink) rotate(now time.Time) error {
	if s.w != nil {
		_ = s.w.Flush()
		s.w = nil
	}
	if s.file != nil {
		_ = s.file.Close()
		s.file = nil
	}
	return s.create(now)
}

func (s *pluginSink) openTail() bool {
	if s.keeper.tail == nil {
		return false
	}
	filename := filepath.Join(s.keeper.dir, s.keeper.tail.name)
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
	s.file = file
	s.w = bufio.NewWriterSize(file, bufferSize)
	s.nbytes = uint32(info.Size()) + n
	return true
}

func (s *pluginSink) create(now time.Time) error {
	name, link := logName(logFileTag, now)
	file, err := os.Create(filepath.Join(s.keeper.dir, name))
	if err != nil {
		return err
	}
	n, err := writePluginHeader(file, now)
	if err != nil {
		_ = file.Close()
		return err
	}
	s.file = file
	s.w = bufio.NewWriterSize(file, bufferSize)
	s.nbytes = n
	s.keeper.add(&fileBlock{fileInfo: fileInfo{name: name, timestamp: ""}, next: nil})
	symlink := filepath.Join(s.keeper.dir, link)
	_ = os.Remove(symlink)
	_ = os.Symlink(name, symlink)
	return nil
}

func ensurePluginRoot() error {
	pluginOnce.Do(func() {
		pluginRoot = filepath.Join(logDir, "plugin")
		pluginRootErr = os.MkdirAll(pluginRoot, 0o755)
		pluginSinks = map[string]*pluginSink{}
	})
	return pluginRootErr
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
	name := b.String()
	if name == "." || name == ".." {
		return "unknown"
	}
	return name
}
