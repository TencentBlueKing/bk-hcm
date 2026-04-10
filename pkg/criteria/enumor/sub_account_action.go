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

import (
	"fmt"
)

// SubAccountAction 三级账号子操作类型
type SubAccountAction string

// Validate the SubAccountAction is valid or not.
func (a SubAccountAction) Validate() error {
	switch a {
	case SubAccountActionCreate:
	case SubAccountActionUpdate:
	case SubAccountActionDelete:
	case SubAccountActionCreateSecretKey:
	case SubAccountActionDisableSecretKey:
	case SubAccountActionEnableSecretKey:
	case SubAccountActionDeleteSecretKey:
	default:
		return fmt.Errorf("unsupported sub account action: %s", a)
	}

	return nil
}

const (
	// SubAccountActionCreate 创建三级账号
	SubAccountActionCreate SubAccountAction = "create"
	// SubAccountActionUpdate 更新三级账号
	SubAccountActionUpdate SubAccountAction = "update"
	// SubAccountActionDelete 删除三级账号
	SubAccountActionDelete SubAccountAction = "delete"
	// SubAccountActionCreateSecretKey 新增三级账号密钥
	SubAccountActionCreateSecretKey SubAccountAction = "create_secret_key"
	// SubAccountActionDisableSecretKey 禁用三级账号密钥
	SubAccountActionDisableSecretKey SubAccountAction = "disable_secret_key"
	// SubAccountActionEnableSecretKey 开启三级账号密钥
	SubAccountActionEnableSecretKey SubAccountAction = "enable_secret_key"
	// SubAccountActionDeleteSecretKey 删除三级账号密钥
	SubAccountActionDeleteSecretKey SubAccountAction = "delete_secret_key"
)
