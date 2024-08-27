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

package account

import (
	"strconv"

	"hcm/pkg/criteria/validator"
	"hcm/pkg/tools/converter"
)

// TCloudAccount define tcloud sub account.
type TCloudAccount struct {
	// 子用户用户 ID
	Uin *uint64 `json:"uin"`

	// 子用户用户名
	Name *string `json:"name"`

	// 子用户 UID
	Uid *uint64 `json:"uid"`

	// 子用户备注
	Remark *string `json:"remark"`

	// 子用户能否登录控制台
	ConsoleLogin *uint64 `json:"console_login"`

	// 手机号
	PhoneNum *string `json:"phone_num"`

	// 区号
	CountryCode *string `json:"country_code"`

	// 邮箱
	Email *string `json:"email"`

	// 创建时间
	// 注意：此字段可能返回 null，表示取不到有效值。
	CreateTime *string `json:"create_time"`

	// 昵称
	// 注意：此字段可能返回 null，表示取不到有效值。
	NickName *string `json:"nick_name"`
}

// GetCloudID ...
func (account TCloudAccount) GetCloudID() string {
	return strconv.FormatUint(converter.PtrToVal(account.Uin), 10)
}

// TCloudListPolicyOption define tcloud list policy option.
type TCloudListPolicyOption struct {
	Uin         uint64  `json:"uin" validate:"required"`
	ServiceType *string `json:"service_type" validate:"omitempty"`
}

// Validate define tcloud list policy option.
func (opt TCloudListPolicyOption) Validate() error {
	return validator.Validate.Struct(opt)
}
