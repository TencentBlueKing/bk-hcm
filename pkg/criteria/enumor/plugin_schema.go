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

package enumor

import "fmt"

// SchemaSyncResult 插件 Schema 同步结果。
type SchemaSyncResult string

// Validate 校验插件 Schema 同步结果是否为合法枚举值。
func (s SchemaSyncResult) Validate() error {
	switch s {
	case SchemaSyncSuccess, SchemaSyncFailed:
	default:
		return fmt.Errorf("unsupported schema sync result: %s", s)
	}

	return nil
}

const (
	// SchemaSyncSuccess 同步成功，包含首次落库与同版本哈希一致的幂等跳过。
	SchemaSyncSuccess SchemaSyncResult = "success"
	// SchemaSyncFailed 同步失败，包含拉取失败、基础校验不通过与同版本 schema_hash 冲突。
	SchemaSyncFailed SchemaSyncResult = "failed"
)
