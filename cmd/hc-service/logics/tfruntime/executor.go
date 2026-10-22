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

import "hcm/pkg/kit"

// Executor 进程管理 Start 返回的执行器，是对接方案 6.2 中执行器的完整方法集。
//
// 实现持有 Provider，在同名方法上按 schema 校验入参、编解码 DynamicValue；
// 响应含 Error 级 Diagnostic 时返回 *DiagnosticsError，超时、连接断开、进程退出等返回普通 error。
type Executor interface {
	// GetProviderSchema 取装配好的统一 Schema。
	GetProviderSchema(kt *kit.Kit) (*ProviderSchema, error)
	// ConfigureProvider 把 Provider 配置写入插件进程，可重入，同一进程可被多次配置成不同的 config。
	// 实现内部先 PrepareProviderConfig 补默认值，再 ConfigureProvider，凭证只通过 Config 传递。
	// 成功之前不得调用资源或数据源方法。
	ConfigureProvider(kt *kit.Kit, req *ConfigureProviderRequest) (*ConfigureProviderResponse, error)

	ValidateResourceTypeConfig(kt *kit.Kit, req *ValidateResourceTypeConfigRequest) (
		*ValidateResourceTypeConfigResponse, error)
	ValidateDataSourceConfig(kt *kit.Kit, req *ValidateDataSourceConfigRequest) (
		*ValidateDataSourceConfigResponse, error)
	UpgradeResourceState(kt *kit.Kit, req *UpgradeResourceStateRequest) (*UpgradeResourceStateResponse, error)
	ReadResource(kt *kit.Kit, req *ReadResourceRequest) (*ReadResourceResponse, error)
	PlanResourceChange(kt *kit.Kit, req *PlanResourceChangeRequest) (*PlanResourceChangeResponse, error)
	ApplyResourceChange(kt *kit.Kit, req *ApplyResourceChangeRequest) (*ApplyResourceChangeResponse, error)
	ImportResourceState(kt *kit.Kit, req *ImportResourceStateRequest) (*ImportResourceStateResponse, error)
	ReadDataSource(kt *kit.Kit, req *ReadDataSourceRequest) (*ReadDataSourceResponse, error)
}
