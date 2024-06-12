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

package handlers

import (
	"fmt"

	"hcm/cmd/cloud-server/logics/audit"
	"hcm/pkg/api/core"
	"hcm/pkg/client"
	"hcm/pkg/criteria/enumor"
	"hcm/pkg/cryptography"
	"hcm/pkg/rest"
	itsm2 "hcm/pkg/thirdparty/api-gateway/itsm"
	"hcm/pkg/thirdparty/esb"
)

// HandlerOption 这里是为了方便调用传参构造Handler,避免参数太多
type HandlerOption struct {
	Cts       *rest.Contexts
	Client    *client.ClientSet
	EsbClient esb.Client
	Cipher    cryptography.Crypto
	Audit     audit.Interface
	ItsmCli   itsm2.Client
}

// BaseApplicationHandler 基础的Handler 一些公共函数和属性处理，可以给到其他具体Handler组合
type BaseApplicationHandler struct {
	applicationType enumor.ApplicationType
	vendor          enumor.Vendor

	Cts       *rest.Contexts
	Client    *client.ClientSet
	EsbClient esb.Client
	Cipher    cryptography.Crypto
	Audit     audit.Interface
}

// NewBaseApplicationHandler ...
func NewBaseApplicationHandler(
	opt *HandlerOption, applicationType enumor.ApplicationType, vendor enumor.Vendor,
) BaseApplicationHandler {
	return BaseApplicationHandler{
		applicationType: applicationType,
		vendor:          vendor,
		Cts:             opt.Cts,
		Client:          opt.Client,
		EsbClient:       opt.EsbClient,
		Cipher:          opt.Cipher,
		Audit:           opt.Audit,
	}
}

// GetType 申请单类型
func (a *BaseApplicationHandler) GetType() enumor.ApplicationType {
	return a.applicationType
}

// Vendor ...
func (a *BaseApplicationHandler) Vendor() enumor.Vendor {
	return a.vendor
}

// ConvertMemoryMBToGB 将内存的MB转换为可用于展示的GB, 特殊展示，不适合其他通用的转换
func (a *BaseApplicationHandler) ConvertMemoryMBToGB(m int64) string {
	if m%1024 == 0 {
		return fmt.Sprintf("%d", m/1024)
	}

	return fmt.Sprintf("%.1f", float64(m/1024))
}

func (a *BaseApplicationHandler) getPageOfOneLimit() *core.BasePage {
	return &core.BasePage{Count: false, Start: 0, Limit: 1}
}

// GetItsmPlatformAndAccountApprover get itsm platform and account approver.
func (a *BaseApplicationHandler) GetItsmPlatformAndAccountApprover(managers []string,
	accountID string) []itsm2.VariableApprover {

	allManagers := []itsm2.VariableApprover{
		{
			Variable:  "platform_manager",
			Approvers: managers,
		},
	}

	accountData, err := a.GetAccount(accountID)
	if err != nil {
		return allManagers
	}

	allManagers = append(allManagers, itsm2.VariableApprover{
		Variable:  "account_manager",
		Approvers: accountData.Managers,
	})

	return allManagers
}

// Complete complete the application by manual.
func (a *BaseApplicationHandler) Complete() (status enumor.ApplicationStatus, deliverDetail map[string]interface{}, err error) {
	return enumor.DeliverError, map[string]interface{}{}, fmt.Errorf("not implemented")
}
