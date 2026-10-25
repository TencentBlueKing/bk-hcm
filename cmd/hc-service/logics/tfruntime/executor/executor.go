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
	"io"
	"sync"

	"hcm/cmd/hc-service/types"
	"hcm/pkg/criteria/constant"
	"hcm/pkg/logs"
	"hcm/pkg/plugin/tfplugin5"

	"github.com/hashicorp/go-plugin"
)

// errExecutorClosed is returned when a business RPC starts after Close.
var errExecutorClosed = errors.New("terraform plugin executor is closing")

// errValueCodecRequired is returned when a method needs to encode or decode and no codec is set.
var errValueCodecRequired = errors.New("terraform plugin value codec is not set")

// Executor is the provider API a caller uses after a plugin process is running.
// Close stops that process. Attach and SetCodec stay on the concrete type.
type Executor interface {
	Close() error
	GetProviderSchema(ctx context.Context) (*types.ProviderSchema, error)
	ConfigureProvider(ctx context.Context, config any) (*types.ConfigureResult, error)
	ValidateResourceTypeConfig(ctx context.Context, typeName string, config any) (*types.ValidateResult, error)
	UpgradeResourceState(ctx context.Context, typeName string, stateVersion int64, raw *types.RawState) (*types.UpgradeResult, error)
	ReadResource(ctx context.Context, typeName string, current any, private []byte) (*types.ReadResourceResult, error)
	PlanResourceChange(ctx context.Context, typeName string, prior, proposed, config any, priorPrivate []byte) (*types.PlanResult, error)
	ApplyResourceChange(ctx context.Context, typeName string, prior, planned, config any, plannedPrivate []byte) (*types.ApplyResult, error)
	ImportResourceState(ctx context.Context, typeName, id string) (*types.ImportResult, error)
	ValidateDataSourceConfig(ctx context.Context, typeName string, config any) (*types.ValidateResult, error)
	ReadDataSource(ctx context.Context, typeName string, config any) (*types.ReadDataSourceResult, error)
}

// PluginExecutor is a running provider process.
// It implements Executor. Business RPCs call the generated provider client.
// StopProvider does not enter the close latch.
// Attach and SetCodec wire the process and are not part of Executor.
type PluginExecutor struct {
	client  *plugin.Client
	rpc     tfplugin5.ProviderClient
	logs    io.Closer
	name    string
	version string
	codec   ValueCodec

	configured        bool
	providerSchema    *tfplugin5.Schema
	resourceSchemas   map[string]*tfplugin5.Schema
	dataSourceSchemas map[string]*tfplugin5.Schema

	mu       sync.Mutex
	closing  bool
	wg       sync.WaitGroup
	once     sync.Once
	closeErr error
}

var _ Executor = (*PluginExecutor)(nil)

// Attach binds the process that produced this executor. GRPCClient only has the
// connection, so the manager supplies the process, log, and registered version after Dispense.
// version is the registered plugin version. An empty version stays empty.
func (e *PluginExecutor) Attach(client *plugin.Client, logs io.Closer, name, version string) {
	e.client = client
	e.logs = logs
	e.name = name
	e.version = version
}

// SetCodec stores the codec used before config and state RPCs.
func (e *PluginExecutor) SetCodec(codec ValueCodec) {
	e.mu.Lock()
	e.codec = codec
	e.mu.Unlock()
}

func (e *PluginExecutor) valueCodec() ValueCodec {
	e.mu.Lock()
	defer e.mu.Unlock()
	return e.codec
}

// encode turns a caller value into a protocol DynamicValue.
// A nil value stays nil and is not encoded. A missing codec returns an error and does not touch the plugin.
func (e *PluginExecutor) encode(schema *tfplugin5.Schema, value any) (*tfplugin5.DynamicValue, error) {
	if value == nil {
		return nil, nil
	}
	codec := e.valueCodec()
	if codec == nil {
		return nil, errValueCodecRequired
	}
	return codec.Encode(schema, value)
}

// decode turns a protocol DynamicValue back into a caller value.
// An empty value stays nil. A missing codec returns an error and does not touch the plugin.
func (e *PluginExecutor) decode(schema *tfplugin5.Schema, value *tfplugin5.DynamicValue) (any, error) {
	if value == nil || (len(value.GetMsgpack()) == 0 && len(value.GetJson()) == 0) {
		return nil, nil
	}
	codec := e.valueCodec()
	if codec == nil {
		return nil, errValueCodecRequired
	}
	return codec.Decode(schema, value)
}

// Close rejects new RPCs, waits for RPCs already inside a method, calls StopProvider, then Kill.
// A StopProvider error is logged and does not skip Kill. A repeated Close runs this sequence once.
// StopProvider does not enter the business-RPC latch.
func (e *PluginExecutor) Close() error {
	e.once.Do(func() {
		e.mu.Lock()
		e.closing = true
		e.mu.Unlock()
		e.wg.Wait()

		defer func() {
			e.client.Kill()
			if e.logs != nil {
				_ = e.logs.Close()
			}
		}()

		ctx, cancel := context.WithTimeout(context.Background(), constant.TFPluginStopProviderTimeout)
		defer cancel()
		if err := e.stopProvider(ctx); err != nil {
			logs.Errorf("plugin StopProvider failed, killing process: plugin=%s plugin_version=%s err=%v",
				e.name, e.version, err)
			e.closeErr = err
		}
	})
	return e.closeErr
}

// stopProvider calls StopProvider without the business-RPC latch.
func (e *PluginExecutor) stopProvider(ctx context.Context) error {
	_, err := e.rpc.Stop(ctx, &tfplugin5.Stop_Request{})
	return err
}

// enter records one in-flight business RPC. Close rejects a new entry.
func (e *PluginExecutor) enter() error {
	e.mu.Lock()
	defer e.mu.Unlock()
	if e.closing {
		return errExecutorClosed
	}
	e.wg.Add(1)
	return nil
}

func (e *PluginExecutor) leave() {
	e.wg.Done()
}

// callRPC runs one business RPC inside the close latch.
func callRPC[T any](e *PluginExecutor, fn func() (T, error)) (T, error) {
	var zero T
	if err := e.enter(); err != nil {
		return zero, err
	}
	defer e.leave()
	return fn()
}
