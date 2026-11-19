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

	"hcm/pkg/plugin/tfplugin5"

	"github.com/hashicorp/go-plugin"
	"google.golang.org/grpc"
)

// errPluginServerUnsupported is returned because hc-service only consumes plugins.
var errPluginServerUnsupported = errors.New("terraform plugin grpc server is not supported")

// Plugin is the protocol 5 client adapter. It does not serve the protocol.
type Plugin struct {
	// 暂时不需要支持 net/rpc 协议（版本4以下）的旧插件，但是需要实现 Plugin interface
	plugin.NetRPCUnsupportedPlugin
}

var (
	_ plugin.GRPCPlugin = (*Plugin)(nil)
)

// GRPCClient returns the executor for an established connection.
// The executor holds the generated provider client. It does not implement ProviderClient.
func (p *Plugin) GRPCClient(
	_ context.Context, _ *plugin.GRPCBroker, conn *grpc.ClientConn,
) (interface{}, error) {
	return &PluginExecutor{rpc: tfplugin5.NewProviderClient(conn)}, nil
}

// GRPCServer reports that hc-service does not host this protocol.
func (p *Plugin) GRPCServer(*plugin.GRPCBroker, *grpc.Server) error {
	return errPluginServerUnsupported
}
