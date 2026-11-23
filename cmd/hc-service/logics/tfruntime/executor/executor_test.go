/*
 * TencentBlueKing is pleased to support the open source community by making
 * 蓝鲸智云 - 混合云管理平台 (BlueKing - Hybrid Cloud Management System) available.
 * Copyright (C) 2022 THL A29 Limited,
 * a Tencent company. All rights reserved.
 * Licensed under the MIT License (the "License");
 * you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at http://opensource.org/licenses/MIT
 * Unless required by applicable law or agreed to in writing,
 * software distributed under the License is distributed on an "AS IS" BASIS,
 * WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
 * See the License for the specific language governing permissions and
 * limitations under the License.
 *
 * We undertake not to change the open source license (MIT license) applicable
 *
 * to the current version of the project delivered to anyone in the future.
 */

package executor

import (
	"context"
	"errors"
	"slices"
	"sync"
	"testing"
	"time"

	"hcm/pkg/plugin/tfplugin5"

	"github.com/hashicorp/go-plugin"
)

func TestPluginExecutor_SetCodec(t *testing.T) {
	executor := &PluginExecutor{}
	_, err := executor.encode(&tfplugin5.Schema{}, "ak")
	if !errors.Is(err, errValueCodecRequired) {
		t.Fatalf("before set err = %v", err)
	}
	want := &tfplugin5.DynamicValue{Msgpack: []byte("packed")}
	executor.SetCodec(staticCodec{value: want})
	got, err := executor.encode(&tfplugin5.Schema{}, "ak")
	if err != nil {
		t.Fatal(err)
	}
	if string(got.GetMsgpack()) != "packed" {
		t.Fatalf("encoded = %v", got.GetMsgpack())
	}
}

type staticCodec struct {
	value *tfplugin5.DynamicValue
}

func (c staticCodec) Encode(*tfplugin5.Schema, any) (*tfplugin5.DynamicValue, error) {
	return c.value, nil
}

func (c staticCodec) Decode(*tfplugin5.Schema, *tfplugin5.DynamicValue) (any, error) {
	return "decoded", nil
}

func TestPluginExecutor_Decode(t *testing.T) {
	executor := &PluginExecutor{}
	_, err := executor.decode(&tfplugin5.Schema{}, &tfplugin5.DynamicValue{Msgpack: []byte{1}})
	if !errors.Is(err, errValueCodecRequired) {
		t.Fatalf("err = %v", err)
	}
}

type trackingCloser struct {
	calls  int
	onCall func()
}

func (c *trackingCloser) Close() error {
	c.calls++
	if c.onCall != nil {
		c.onCall()
	}
	return nil
}

func TestPluginExecutor_CloseDrainsInFlightAndRejectsNew(t *testing.T) {
	entered := make(chan struct{})
	release := make(chan struct{})
	var eventsMu sync.Mutex
	var events []string
	recordEvent := func(ev string) {
		eventsMu.Lock()
		events = append(events, ev)
		eventsMu.Unlock()
	}
	fake := &fakePlugin{
		getSchema: resourceSchema(t),
		read: func(*tfplugin5.ReadResource_Request) (*tfplugin5.ReadResource_Response, error) {
			close(entered)
			<-release
			recordEvent("rpc_done")
			return &tfplugin5.ReadResource_Response{
				NewState: &tfplugin5.DynamicValue{Json: []byte(`{"id":"ins-1"}`)},
			}, nil
		},
		stop: func(*tfplugin5.Stop_Request) (*tfplugin5.Stop_Response, error) {
			recordEvent("stop")
			return &tfplugin5.Stop_Response{}, nil
		},
	}
	loadEchoPrepare(fake)
	logsCloser := &trackingCloser{onCall: func() {
		recordEvent("logs_close")
	}}
	exec := preparedExecutor(t, fake, &jsonCodec{})
	exec.Attach(&plugin.Client{}, logsCloser, "tencentcloud", "1.0.0")

	rpcDone := make(chan error, 1)
	go func() {
		_, err := exec.ReadResource(context.Background(), "tcloud_cvm", map[string]any{"id": "ins-1"}, nil)
		rpcDone <- err
	}()
	<-entered

	closeDone := make(chan error, 1)
	go func() {
		closeDone <- exec.Close()
	}()

	// Wait until Close enters the closing phase and blocks on wg.Wait().
	deadline := time.Now().Add(time.Second)
	for {
		exec.mu.Lock()
		closing := exec.closing
		exec.mu.Unlock()
		if closing {
			break
		}
		if time.Now().After(deadline) {
			t.Fatal("Close did not mark executor closing in time")
		}
		time.Sleep(time.Millisecond)
	}

	// New RPCs must be rejected immediately and never reach the plugin while draining.
	if _, err := exec.ReadResource(context.Background(), "tcloud_cvm", map[string]any{"id": "ins-2"}, nil); !errors.Is(err, errExecutorClosed) {
		t.Fatalf("new RPC while closing err = %v, want %v", err, errExecutorClosed)
	}
	if fake.count("ReadResource") != 1 {
		t.Fatalf("rejected RPC should not reach plugin, calls=%d", fake.count("ReadResource"))
	}
	select {
	case err := <-closeDone:
		t.Fatalf("Close returned before in-flight RPC completed: %v", err)
	case <-time.After(50 * time.Millisecond):
	}
	if fake.count("Stop") != 0 || logsCloser.calls != 0 {
		t.Fatalf("Close ran Stop/logs.Close before drain: stop=%d logs=%d", fake.count("Stop"), logsCloser.calls)
	}

	close(release)
	if err := <-rpcDone; err != nil {
		t.Fatalf("in-flight ReadResource: %v", err)
	}
	select {
	case err := <-closeDone:
		if err != nil {
			t.Fatalf("Close: %v", err)
		}
	case <-time.After(time.Second):
		t.Fatal("Close did not finish after in-flight RPC returned")
	}

	wantOrder := []string{"rpc_done", "stop", "logs_close"}
	if !slices.Equal(events, wantOrder) {
		t.Fatalf("close order = %v, want %v", events, wantOrder)
	}

	// Repeated Close is idempotent and does not call Stop or logs.Close again.
	if err := exec.Close(); err != nil {
		t.Fatalf("second Close: %v", err)
	}
	if fake.count("Stop") != 1 || logsCloser.calls != 1 {
		t.Fatalf("repeated Close ran again: stop=%d logs=%d", fake.count("Stop"), logsCloser.calls)
	}
}

func TestPluginExecutor_CloseStopErrorStillClosesLogs(t *testing.T) {
	stopErr := errors.New("stop rpc failed")
	fake := &fakePlugin{
		stop: func(*tfplugin5.Stop_Request) (*tfplugin5.Stop_Response, error) {
			return nil, stopErr
		},
	}
	logsCloser := &trackingCloser{}
	exec := &PluginExecutor{rpc: fake}
	exec.Attach(&plugin.Client{}, logsCloser, "tencentcloud", "1.0.0")

	if err := exec.Close(); !errors.Is(err, stopErr) {
		t.Fatalf("Close err = %v, want %v", err, stopErr)
	}
	if fake.count("Stop") != 1 || logsCloser.calls != 1 {
		t.Fatalf("stop=%d logs=%d, want 1 and 1", fake.count("Stop"), logsCloser.calls)
	}
	if err := exec.Close(); !errors.Is(err, stopErr) {
		t.Fatalf("second Close err = %v, want %v", err, stopErr)
	}
	if fake.count("Stop") != 1 || logsCloser.calls != 1 {
		t.Fatalf("repeated Close re-executed: stop=%d logs=%d", fake.count("Stop"), logsCloser.calls)
	}
}
