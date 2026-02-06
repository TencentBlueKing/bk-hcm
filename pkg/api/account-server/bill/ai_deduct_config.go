/*
 * TencentBlueKing is pleased to support the open source community by making
 * 蓝鲸智云 - 混合云管理平台 (BlueKing - Hybrid Cloud Management System) available.
 * Copyright (C) 2024 THL A29 Limited,
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

package bill

import (
	"fmt"

	"hcm/pkg/criteria/errf"
	"hcm/pkg/criteria/validator"
)

// GetAIDeductConfigResp 获取AI账单抵扣配置响应
type GetAIDeductConfigResp struct {
	MainAccountIDs []string `json:"main_account_ids"`
}

// UpdateAIDeductConfigReq 更新AI账单抵扣配置请求
type UpdateAIDeductConfigReq struct {
	MainAccountIDs []string `json:"main_account_ids" validate:"required"`
}

// Validate UpdateAIDeductConfigReq
func (r *UpdateAIDeductConfigReq) Validate() error {
	if err := validator.Validate.Struct(r); err != nil {
		return err
	}
	// 验证主账号ID列表是否为有效的JSON数组格式
	if len(r.MainAccountIDs) == 0 {
		return nil
	}
	// 验证每个ID不为空
	for i, id := range r.MainAccountIDs {
		if len(id) == 0 {
			return errf.New(errf.InvalidParameter, fmt.Sprintf("main_account_ids[%d] cannot be empty", i))
		}
	}

	return nil
}
