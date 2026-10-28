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

package constant

import "time"

const (
	// TFPluginCoreProtocolVersion 是 go-plugin 自身的握手版本，与 Terraform 插件协议版本无关
	TFPluginCoreProtocolVersion = 1
	// TFPluginProtocolVersion 是当前接受的 Terraform 插件协议主版本
	// 协议 6 接入时另加常量
	TFPluginProtocolVersion = 5
	// TFPluginSchemaProtocolVersion 是 GetProviderSchema 对外文档的 protocol_version，不是插件握手主版本
	TFPluginSchemaProtocolVersion = "1.0"
	// TFPluginProviderName 是插件进程内的 dispense 名称
	// 协议 5 和协议 6 使用同一个名字，版本由 VersionedPlugins 的键区分
	// 该值必须与插件二进制注册的名字一致，官方 Provider 固定为 provider
	TFPluginProviderName = "provider"

	// TFPluginMagicCookieKey 与 TFPluginMagicCookieValue 是插件握手常量
	TFPluginMagicCookieKey   = "TF_PLUGIN_MAGIC_COOKIE"
	TFPluginMagicCookieValue = "d602bf8f470bc67ca7faa0386276bbdd4330efaf76d1a219cb4d6991ca9872b2"

	// TFPluginMaxGRPCMsgSize 是 gRPC 收发消息上限。
	TFPluginMaxGRPCMsgSize = 256 << 20

	// TFPluginStopProviderTimeout 是 StopProvider 的等待时间。超时后仍 Kill 进程，避免卡住服务退出。
	TFPluginStopProviderTimeout = 5 * time.Second

	// TFPluginDiagNoImportedResource 导入资源时插件返回空 Imported 列表的错误诊断文案。
	TFPluginDiagNoImportedResource = "provider returned no imported resource"
	// TFPluginDiagImportNonExistent 导入资源后 ReadResource 返回空状态（远端对象不存在）的错误诊断文案。
	TFPluginDiagImportNonExistent = "cannot import non-existent remote object"
	// TFPluginDiagNullDataSourceState 读取数据源时插件返回空状态的错误诊断文案。
	TFPluginDiagNullDataSourceState = "provider produced a null state for data source"
	// TFPluginDiagStateNotWhollyKnown 读取数据源返回的状态中仍包含未知值时的错误诊断文案。
	TFPluginDiagStateNotWhollyKnown = "provider produced a state value that is not wholly known"
)
