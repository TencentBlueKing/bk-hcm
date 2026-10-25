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
	"testing"
	"time"

	"hcm/pkg/plugin/tfplugin5"
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

func TestPluginExecutor_Encode(t *testing.T) {
	executor := &PluginExecutor{}
	_, err := executor.encode(&tfplugin5.Schema{}, map[string]any{"secret_id": "ak"})
	if !errors.Is(err, errValueCodecRequired) {
		t.Fatalf("err = %v", err)
	}
}

func TestPluginExecutor_Decode(t *testing.T) {
	executor := &PluginExecutor{}
	_, err := executor.decode(&tfplugin5.Schema{}, &tfplugin5.DynamicValue{Msgpack: []byte{1}})
	if !errors.Is(err, errValueCodecRequired) {
		t.Fatalf("err = %v", err)
	}
}

func TestPluginExecutor_CallRPC(t *testing.T) {
	t.Run("closing", func(t *testing.T) {
		executor := &PluginExecutor{}
		executor.mu.Lock()
		executor.closing = true
		executor.mu.Unlock()
		called := false
		_, err := callRPC(executor, func() (struct{}, error) {
			called = true
			return struct{}{}, nil
		})
		if !errors.Is(err, errExecutorClosed) || called {
			t.Fatalf("err = %v called = %v", err, called)
		}
	})

	t.Run("in flight", func(t *testing.T) {
		entered := make(chan struct{})
		release := make(chan struct{})
		executor := &PluginExecutor{}
		done := make(chan struct{})
		go func() {
			_, err := callRPC(executor, func() (struct{}, error) {
				close(entered)
				<-release
				return struct{}{}, nil
			})
			if err != nil {
				t.Errorf("callRPC: %v", err)
			}
			close(done)
		}()
		<-entered

		executor.mu.Lock()
		executor.closing = true
		executor.mu.Unlock()
		waited := make(chan struct{})
		go func() {
			executor.wg.Wait()
			close(waited)
		}()
		select {
		case <-waited:
			t.Fatal("wait finished while callRPC was in flight")
		case <-time.After(50 * time.Millisecond):
		}
		close(release)
		select {
		case <-waited:
		case <-time.After(time.Second):
			t.Fatal("wait did not finish after callRPC returned")
		}
		<-done
	})
}

func TestPluginExecutor_StopProvider(t *testing.T) {
	fake := &fakePlugin{}
	executor := &PluginExecutor{rpc: fake}
	executor.mu.Lock()
	executor.closing = true
	executor.mu.Unlock()
	if err := executor.stopProvider(context.Background()); err != nil {
		t.Fatal(err)
	}
	if fake.count("Stop") != 1 {
		t.Fatalf("stop calls = %d", fake.count("Stop"))
	}
}
