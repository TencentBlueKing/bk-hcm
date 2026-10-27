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
	"os/exec"

	"hcm/pkg/criteria/constant"

	"github.com/hashicorp/go-plugin"
	"google.golang.org/grpc"
)

// newPluginClientConfig returns the fixed handshake settings for one plugin process.
// cmd is the unstarted plugin command. The caller sets its environment and the
// stdio writers before plugin.NewClient.
//
// SkipHostEnv stops go-plugin from copying os.Environ into the plugin.
// On non-Windows, go-plugin v1.6.3 makes the plugin process listen on a unix
// socket and return that address in the handshake line. An empty UnixSocketConfig
// keeps that default temporary directory and does not switch the transport to TCP.
// minPort and maxPort are still read by this go-plugin version for the TCP fallback.
func newPluginClientConfig(cmd *exec.Cmd, minPort, maxPort uint) *plugin.ClientConfig {
	return &plugin.ClientConfig{
		HandshakeConfig: plugin.HandshakeConfig{
			ProtocolVersion:  constant.TFPluginCoreProtocolVersion,
			MagicCookieKey:   constant.TFPluginMagicCookieKey,
			MagicCookieValue: constant.TFPluginMagicCookieValue,
		},
		VersionedPlugins: map[int]plugin.PluginSet{
			constant.TFPluginProtocolVersion: {
				constant.TFPluginProviderName: &providerPlugin{},
			},
		},
		Cmd:              cmd,
		AllowedProtocols: []plugin.Protocol{plugin.ProtocolGRPC},
		AutoMTLS:         true,
		SkipHostEnv:      true,
		MinPort:          minPort,
		MaxPort:          maxPort,
		GRPCDialOptions: []grpc.DialOption{
			grpc.WithDefaultCallOptions(
				grpc.MaxCallRecvMsgSize(constant.TFPluginMaxGRPCMsgSize),
				grpc.MaxCallSendMsgSize(constant.TFPluginMaxGRPCMsgSize),
			),
		},
		UnixSocketConfig: &plugin.UnixSocketConfig{},
	}
}
