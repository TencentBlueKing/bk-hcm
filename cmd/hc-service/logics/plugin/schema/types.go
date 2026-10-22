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

// Package schema 提供 hc-service 侧插件 Schema 获取、校验、持久化与缓存能力。
package schema

import (
	"time"

	coreplugin "hcm/pkg/api/core/plugin"
	"hcm/pkg/criteria/enumor"
	"hcm/pkg/kit"
)

const (
	// DefaultSchemaCacheTTL 缓存 Schema 的默认滑动 TTL（7 天）。
	DefaultSchemaCacheTTL = 7 * 24 * time.Hour
)

// SchemaGetter 抽象从 tf_provider / http 插件南向获取 Schema。
type SchemaGetter interface {
	GetSchema(kt *kit.Kit) (*coreplugin.PluginSchema, error)
}

// SchemaGetterFunc 把函数适配为 SchemaGetter。
type SchemaGetterFunc func(kt *kit.Kit) (*coreplugin.PluginSchema, error)

// GetSchema 实现 SchemaGetter。
func (f SchemaGetterFunc) GetSchema(kt *kit.Kit) (*coreplugin.PluginSchema, error) { return f(kt) }

// SyncSchemaResult SchemaManager.SyncSchema 操作结果。
// SyncResult 为 success 时 SyncMessage 为空且 Schema 有值；为 failed 时 SyncMessage 为原因且 Schema 为 nil。
type SyncSchemaResult struct {
	SyncResult  enumor.SchemaSyncResult  `json:"sync_result"`
	SyncMessage string                   `json:"sync_message"`
	Schema      *coreplugin.PluginSchema `json:"schema"`
}
