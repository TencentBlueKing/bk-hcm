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

// UserCollectionResType 用户收藏的资源类型
type UserCollectionResType string

// Validate UserCollectionResType.
func (typ UserCollectionResType) Validate() error {
	switch typ {
	case BizCollResType:
	case CloudSelectionSchemeCollResType:
	default:
		return fmt.Errorf("res type: %s not support", typ)
	}

	return nil
}

const (
	// BizCollResType 业务资源类型
	BizCollResType UserCollectionResType = "biz"
	// CloudSelectionSchemeCollResType 云选型资源类型
	CloudSelectionSchemeCollResType UserCollectionResType = "cloud_selection_scheme"
)
