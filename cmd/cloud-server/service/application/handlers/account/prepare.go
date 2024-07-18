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
	"fmt"

	"hcm/pkg/thirdparty/api-gateway/itsm"

	"github.com/TencentBlueKing/gopkg/conv"
)

// PrepareReq 预处理请求参数，比如敏感数据加密，Note: 实际应该在DB层面去加密，但是DB层面加密又涉及到Dao需要理解各个云要加密的字段
func (a *ApplicationOfAddAccount) PrepareReq() error {
	// 密钥加密
	secretKeyField := a.req.Vendor.GetSecretField()
	a.req.Extension[secretKeyField] = a.Cipher.EncryptToBase64(conv.ToString(a.req.Extension[secretKeyField]))

	return nil
}

// GenerateApplicationContent 获取预处理过的数据，以interface格式
func (a *ApplicationOfAddAccount) GenerateApplicationContent() interface{} {
	return a.req
}

// PrepareReqFromContent 预处理请求参数，对于申请内容来着DB，其实入库前是加密了的
func (a *ApplicationOfAddAccount) PrepareReqFromContent() error {
	// 解密密钥
	secretKeyField := a.req.Vendor.GetSecretField()
	secretKey, err := a.Cipher.DecryptFromBase64(a.req.Extension[secretKeyField])
	if err != nil {
		return fmt.Errorf("decrypt secret key failed, err: %w", err)
	}
	a.req.Extension[secretKeyField] = secretKey

	return nil
}

// GetItsmApprover 获取itsm审批人
func (a *ApplicationOfAddAccount) GetItsmApprover(managers []string) []itsm.VariableApprover {
	return []itsm.VariableApprover{
		{
			Variable:  "platform_manager",
			Approvers: managers,
		},
	}
}

// GetBkBizIDs 获取当前的业务IDs
func (a *ApplicationOfAddAccount) GetBkBizIDs() []int64 {
	return a.req.BkBizIDs
}
