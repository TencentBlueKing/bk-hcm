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
	"testing"

	"github.com/hashicorp/go-plugin"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"

	"hcm/pkg/plugin/tfplugin5"
)

func TestGRPCClientReturnsProviderClient(t *testing.T) {
	conn, err := grpc.NewClient("passthrough:///unused", grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		t.Fatalf("new grpc client: %v", err)
	}
	t.Cleanup(func() {
		if err := conn.Close(); err != nil {
			t.Errorf("close conn: %v", err)
		}
	})

	raw, err := (&providerPlugin{}).GRPCClient(context.Background(), nil, conn)
	if err != nil {
		t.Fatalf("GRPCClient: %v", err)
	}
	if _, ok := raw.(tfplugin5.ProviderClient); !ok {
		t.Fatalf("GRPCClient returned %T, want tfplugin5.ProviderClient", raw)
	}
}

func TestGRPCServerReturnsError(t *testing.T) {
	err := (&providerPlugin{}).GRPCServer(nil, nil)
	if !errors.Is(err, errPluginServerUnsupported) {
		t.Fatalf("GRPCServer error = %v, want %v", err, errPluginServerUnsupported)
	}
}

func TestPluginClientConfigHandshake(t *testing.T) {
	cfg := newPluginClientConfig(nil)

	if cfg.ProtocolVersion != coreProtocolVersion {
		t.Fatalf("ProtocolVersion = %d, want %d", cfg.ProtocolVersion, coreProtocolVersion)
	}
	if cfg.MagicCookieKey != magicCookieKey || cfg.MagicCookieValue != magicCookieValue {
		t.Fatalf("magic cookie = %s %s", cfg.MagicCookieKey, cfg.MagicCookieValue)
	}
	if len(cfg.VersionedPlugins) != 1 {
		t.Fatalf("VersionedPlugins = %#v, want only protocol %d", cfg.VersionedPlugins, pluginProtocolVersion)
	}
	set, ok := cfg.VersionedPlugins[pluginProtocolVersion]
	if !ok {
		t.Fatalf("VersionedPlugins missing protocol %d", pluginProtocolVersion)
	}
	impl, ok := set[providerPluginName]
	if !ok {
		t.Fatalf("plugin set = %#v, want %s", set, providerPluginName)
	}
	if _, ok := impl.(*providerPlugin); !ok {
		t.Fatalf("plugin %s type = %T", providerPluginName, impl)
	}
	if len(set) != 1 {
		t.Fatalf("plugin set = %#v, want only %s", set, providerPluginName)
	}

	if len(cfg.AllowedProtocols) != 1 || cfg.AllowedProtocols[0] != plugin.ProtocolGRPC {
		t.Fatalf("AllowedProtocols = %#v, want only grpc", cfg.AllowedProtocols)
	}
	if !cfg.AutoMTLS {
		t.Fatal("AutoMTLS is false")
	}
	if cfg.UnixSocketConfig == nil {
		t.Fatal("UnixSocketConfig is nil")
	}
	if maxGRPCMsgSize < 256<<20 {
		t.Fatalf("maxGRPCMsgSize = %d, want at least 256MiB", maxGRPCMsgSize)
	}
	if len(cfg.GRPCDialOptions) != 1 {
		t.Fatalf("GRPCDialOptions len = %d, want 1", len(cfg.GRPCDialOptions))
	}
}
