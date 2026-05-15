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

package createsubaccount

import (
	"fmt"
	"strings"

	"hcm/pkg/criteria/enumor"
	"hcm/pkg/logs"
	"hcm/pkg/tools/converter"
)

// RenderItsmTitle render ITSM ticket title.
func (a *ApplicationOfCreateSubAccount) RenderItsmTitle() (string, error) {
	return fmt.Sprintf("申请新增[%s]三级账号(%s)", a.Vendor().GetNameZh(), a.req.Name), nil
}

// RenderItsmForm render ITSM ticket form content.
func (a *ApplicationOfCreateSubAccount) RenderItsmForm() (string, error) {
	accountData, err := a.GetAccount(a.req.AccountID)
	if err != nil {
		return "", fmt.Errorf("get account info failed, err: %w", err)
	}

	items := []string{
		fmt.Sprintf("云厂商: %s", a.Vendor().GetNameZh()),
		fmt.Sprintf("所属二级账号: %s", accountData.Name),
		fmt.Sprintf("三级账号名称: %s", a.req.Name),
		fmt.Sprintf("开通接收邮箱: %s", a.req.ReceiveEmail),
	}

	vendorItems, err := a.renderExtensionItems()
	if err != nil {
		return "", err
	}
	items = append(items, vendorItems...)

	if a.req.Email != "" {
		items = append(items, fmt.Sprintf("联系邮箱: %s", a.req.Email))
	}
	if a.req.PhoneNum != "" {
		phone := a.req.PhoneNum
		if a.req.CountryCode != "" {
			phone = "+" + a.req.CountryCode + " " + phone
		}
		items = append(items, fmt.Sprintf("手机号: %s", phone))
	}
	if len(a.req.Managers) > 0 {
		items = append(items, fmt.Sprintf("管理者: %s", strings.Join(a.req.Managers, ",")))
	}
	if a.req.Memo != nil && converter.PtrToVal(a.req.Memo) != "" {
		items = append(items, fmt.Sprintf("备注: %s", *a.req.Memo))
	}

	if len(a.req.PermissionTemplateIDs) == 0 {
		return "", fmt.Errorf("permission template ids is empty")
	}
	names, err := a.QueryPermissionTemplateNames(a.req.PermissionTemplateIDs)
	if err != nil {
		logs.Errorf("query permission template names failed, err: %v, rid: %s", err, a.Cts.Kit.Rid)
		return "", fmt.Errorf("query permission template names failed, err: %w", err)
	}

	items = append(items, fmt.Sprintf("绑定权限模版: %s", strings.Join(names, ",")))

	return strings.Join(items, "\n"), nil
}

// renderExtensionItems returns vendor-specific form items rendered from the extension field.
func (a *ApplicationOfCreateSubAccount) renderExtensionItems() ([]string, error) {
	switch a.Vendor() {
	case enumor.TCloud:
		ext, err := decodeTCloudExtension(a)
		if err != nil {
			return nil, fmt.Errorf("decode tcloud extension failed, err: %w", err)
		}
		return []string{fmt.Sprintf("账号类型: %s", ext.ConsoleLogin.GetNameZh())}, nil
	default:
		return nil, nil
	}
}
