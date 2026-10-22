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

	"github.com/hashicorp/go-plugin"
	"google.golang.org/grpc"

	"hcm/pkg/plugin/tfplugin5"
)

// errPluginServerUnsupported is returned because hc-service only consumes plugins.
var errPluginServerUnsupported = errors.New("terraform plugin grpc server is not supported")

// providerPlugin is the protocol 5 client adapter. It does not serve the protocol.
type providerPlugin struct {
	plugin.NetRPCUnsupportedPlugin
}

var (
	_ plugin.Plugin     = (*providerPlugin)(nil)
	_ plugin.GRPCPlugin = (*providerPlugin)(nil)
)

// GRPCClient returns the generated provider client for an established connection.
func (p *providerPlugin) GRPCClient(
	_ context.Context, _ *plugin.GRPCBroker, conn *grpc.ClientConn,
) (interface{}, error) {
	return tfplugin5.NewProviderClient(conn), nil
}

// GRPCServer reports that hc-service does not host this protocol.
func (p *providerPlugin) GRPCServer(*plugin.GRPCBroker, *grpc.Server) error {
	return errPluginServerUnsupported
}
