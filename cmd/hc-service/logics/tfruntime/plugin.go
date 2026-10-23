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
	"errors"
	"sync"

	"hcm/pkg/criteria/constant"
	"hcm/pkg/logs"
	"hcm/pkg/plugin/tfplugin5"

	"github.com/hashicorp/go-plugin"
	"google.golang.org/grpc"
)

// errPluginServerUnsupported is returned because hc-service only consumes plugins.
var errPluginServerUnsupported = errors.New("terraform plugin grpc server is not supported")

// errRPCNotImplemented marks provider RPCs left for a later iteration.
var errRPCNotImplemented = errors.New("terraform plugin rpc is not implemented")

// providerPlugin is the protocol 5 client adapter. It does not serve the protocol.
type providerPlugin struct {
	// 暂时不需要支持 net/rpc 协议（版本4以下）的旧插件，但是需要实现 Plugin interface
	plugin.NetRPCUnsupportedPlugin
}

var (
	_ plugin.GRPCPlugin = (*providerPlugin)(nil)
)

// GRPCClient returns the executor for an established connection.
func (p *providerPlugin) GRPCClient(
	_ context.Context, _ *plugin.GRPCBroker, conn *grpc.ClientConn,
) (interface{}, error) {
	return &PluginExecutor{provider: tfplugin5.NewProviderClient(conn)}, nil
}

// GRPCServer reports that hc-service does not host this protocol.
func (p *providerPlugin) GRPCServer(*plugin.GRPCBroker, *grpc.Server) error {
	return errPluginServerUnsupported
}

// PluginExecutor is one running provider process and the protocol 5 client.
type PluginExecutor struct {
	manager  *ProcessManager
	client   *plugin.Client
	provider tfplugin5.ProviderClient
	logs     *pluginLogs
	name     string
	version  string

	mu       sync.Mutex
	closing  bool
	wg       sync.WaitGroup
	once     sync.Once
	closeErr error
}

var (
	_ tfplugin5.ProviderClient = (*PluginExecutor)(nil)
)

// attach binds the process that produced this executor. GRPCClient only has the
// connection, so the manager supplies the process after Dispense.
func (e *PluginExecutor) attach(manager *ProcessManager, client *plugin.Client, logs *pluginLogs, name string) {
	e.manager = manager
	e.client = client
	e.logs = logs
	e.name = name
	e.version = logs.version
}

// Close rejects new RPCs, waits for RPCs already inside a method, calls Stop, then Kill.
// Stop is the generated name of the StopProvider RPC. A Stop error is logged and
// does not skip Kill. A repeated Close runs this sequence once.
func (e *PluginExecutor) Close() error {
	e.once.Do(func() {
		e.mu.Lock()
		e.closing = true
		e.mu.Unlock()
		e.wg.Wait()

		defer func() {
			e.client.Kill()
			e.logs.close()
			e.manager.detach(e)
		}()

		ctx, cancel := context.WithTimeout(context.Background(), constant.TFPluginStopProviderTimeout)
		defer cancel()
		if _, err := e.Stop(ctx, &tfplugin5.Stop_Request{}); err != nil {
			logs.Errorf("plugin StopProvider failed, killing process: plugin=%s plugin_version=%s err=%v",
				e.name, e.version, err)
			e.closeErr = err
		}
	})
	return e.closeErr
}

// Stop 注意该方法和上游同名 Stop 方法行为不同，此处是优雅退出 provider
func (e *PluginExecutor) Stop(
	ctx context.Context, in *tfplugin5.Stop_Request, opts ...grpc.CallOption,
) (*tfplugin5.Stop_Response, error) {
	return e.provider.Stop(ctx, in, opts...)
}

// TODO: 以下 RPC 不在本迭代实现。实现每个方法时在内部加前置和后置门闩：
// 前置检查 closing，已关闭则直接返回；通过后 wg.Add(1)，方法返回时 wg.Done()。
// 调用方直接调用这些方法，不另设 Call 包装。Stop 由 Close 在门闩关闭后调用，不加这道门闩。

func (e *PluginExecutor) GetMetadata(
	context.Context, *tfplugin5.GetMetadata_Request, ...grpc.CallOption,
) (*tfplugin5.GetMetadata_Response, error) {
	return nil, errRPCNotImplemented
}

func (e *PluginExecutor) GetSchema(
	context.Context, *tfplugin5.GetProviderSchema_Request, ...grpc.CallOption,
) (*tfplugin5.GetProviderSchema_Response, error) {
	return nil, errRPCNotImplemented
}

func (e *PluginExecutor) GetResourceIdentitySchemas(
	context.Context, *tfplugin5.GetResourceIdentitySchemas_Request, ...grpc.CallOption,
) (*tfplugin5.GetResourceIdentitySchemas_Response, error) {
	return nil, errRPCNotImplemented
}

func (e *PluginExecutor) PrepareProviderConfig(
	context.Context, *tfplugin5.PrepareProviderConfig_Request, ...grpc.CallOption,
) (*tfplugin5.PrepareProviderConfig_Response, error) {
	return nil, errRPCNotImplemented
}

func (e *PluginExecutor) ValidateResourceTypeConfig(
	context.Context, *tfplugin5.ValidateResourceTypeConfig_Request, ...grpc.CallOption,
) (*tfplugin5.ValidateResourceTypeConfig_Response, error) {
	return nil, errRPCNotImplemented
}

func (e *PluginExecutor) ValidateDataSourceConfig(
	context.Context, *tfplugin5.ValidateDataSourceConfig_Request, ...grpc.CallOption,
) (*tfplugin5.ValidateDataSourceConfig_Response, error) {
	return nil, errRPCNotImplemented
}

func (e *PluginExecutor) UpgradeResourceState(
	context.Context, *tfplugin5.UpgradeResourceState_Request, ...grpc.CallOption,
) (*tfplugin5.UpgradeResourceState_Response, error) {
	return nil, errRPCNotImplemented
}

func (e *PluginExecutor) UpgradeResourceIdentity(
	context.Context, *tfplugin5.UpgradeResourceIdentity_Request, ...grpc.CallOption,
) (*tfplugin5.UpgradeResourceIdentity_Response, error) {
	return nil, errRPCNotImplemented
}

func (e *PluginExecutor) Configure(
	context.Context, *tfplugin5.Configure_Request, ...grpc.CallOption,
) (*tfplugin5.Configure_Response, error) {
	return nil, errRPCNotImplemented
}

func (e *PluginExecutor) ReadResource(
	context.Context, *tfplugin5.ReadResource_Request, ...grpc.CallOption,
) (*tfplugin5.ReadResource_Response, error) {
	return nil, errRPCNotImplemented
}

func (e *PluginExecutor) PlanResourceChange(
	context.Context, *tfplugin5.PlanResourceChange_Request, ...grpc.CallOption,
) (*tfplugin5.PlanResourceChange_Response, error) {
	return nil, errRPCNotImplemented
}

func (e *PluginExecutor) ApplyResourceChange(
	context.Context, *tfplugin5.ApplyResourceChange_Request, ...grpc.CallOption,
) (*tfplugin5.ApplyResourceChange_Response, error) {
	return nil, errRPCNotImplemented
}

func (e *PluginExecutor) ImportResourceState(
	context.Context, *tfplugin5.ImportResourceState_Request, ...grpc.CallOption,
) (*tfplugin5.ImportResourceState_Response, error) {
	return nil, errRPCNotImplemented
}

func (e *PluginExecutor) MoveResourceState(
	context.Context, *tfplugin5.MoveResourceState_Request, ...grpc.CallOption,
) (*tfplugin5.MoveResourceState_Response, error) {
	return nil, errRPCNotImplemented
}

func (e *PluginExecutor) ReadDataSource(
	context.Context, *tfplugin5.ReadDataSource_Request, ...grpc.CallOption,
) (*tfplugin5.ReadDataSource_Response, error) {
	return nil, errRPCNotImplemented
}

func (e *PluginExecutor) ValidateEphemeralResourceConfig(
	context.Context, *tfplugin5.ValidateEphemeralResourceConfig_Request, ...grpc.CallOption,
) (*tfplugin5.ValidateEphemeralResourceConfig_Response, error) {
	return nil, errRPCNotImplemented
}

func (e *PluginExecutor) OpenEphemeralResource(
	context.Context, *tfplugin5.OpenEphemeralResource_Request, ...grpc.CallOption,
) (*tfplugin5.OpenEphemeralResource_Response, error) {
	return nil, errRPCNotImplemented
}

func (e *PluginExecutor) RenewEphemeralResource(
	context.Context, *tfplugin5.RenewEphemeralResource_Request, ...grpc.CallOption,
) (*tfplugin5.RenewEphemeralResource_Response, error) {
	return nil, errRPCNotImplemented
}

func (e *PluginExecutor) CloseEphemeralResource(
	context.Context, *tfplugin5.CloseEphemeralResource_Request, ...grpc.CallOption,
) (*tfplugin5.CloseEphemeralResource_Response, error) {
	return nil, errRPCNotImplemented
}

func (e *PluginExecutor) GetFunctions(
	context.Context, *tfplugin5.GetFunctions_Request, ...grpc.CallOption,
) (*tfplugin5.GetFunctions_Response, error) {
	return nil, errRPCNotImplemented
}

func (e *PluginExecutor) CallFunction(
	context.Context, *tfplugin5.CallFunction_Request, ...grpc.CallOption,
) (*tfplugin5.CallFunction_Response, error) {
	return nil, errRPCNotImplemented
}
