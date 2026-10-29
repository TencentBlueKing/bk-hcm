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

// SchemaHashAlgo 插件 Schema 哈希算法。
type SchemaHashAlgo string

const (
	// SchemaHashAlgoSHA256 SHA-256 算法标识。
	SchemaHashAlgoSHA256 SchemaHashAlgo = "sha256"
	// SchemaHashAlgoSHA384 SHA-384 算法标识。
	SchemaHashAlgoSHA384 SchemaHashAlgo = "sha384"
	// SchemaHashAlgoSHA512 SHA-512 算法标识。
	SchemaHashAlgoSHA512 SchemaHashAlgo = "sha512"
	// DefaultSchemaHashAlgo 默认 Schema 哈希算法。
	DefaultSchemaHashAlgo = SchemaHashAlgoSHA256
)

// Validate 校验 Schema 哈希算法是否为合法枚举值。
func (a SchemaHashAlgo) Validate() error {
	switch a {
	case SchemaHashAlgoSHA256, SchemaHashAlgoSHA384, SchemaHashAlgoSHA512:
		return nil
	default:
		return fmt.Errorf("unsupported schema hash algo: %s", a)
	}
}
