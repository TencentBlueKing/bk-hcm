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

// BlockKind IaC 解析出的节点类型，enumeration values such as: resource/data/import.
type BlockKind string

const (
	// BlockKindResource 受管资源节点，比如 resource "tencentcloud_vpc" "main" {}
	BlockKindResource BlockKind = "resource"
	// BlockKindData 数据源查询节点，比如 data "tencentcloud_images" "ubuntu" {}
	BlockKindData BlockKind = "data"
	// BlockKindImport 存量资源导入节点，比如 import { to = tencentcloud_vpc.main, id = "vpc-xxx" }
	BlockKindImport BlockKind = "import"
)

// Validate 检查节点类型是否为合法枚举值。
func (k BlockKind) Validate() error {
	switch k {
	case BlockKindResource, BlockKindData, BlockKindImport:
		return nil
	default:
		return fmt.Errorf("unsupported block kind: %s", k)
	}
}
