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

package updatesubaccount

import (
	"fmt"
	"strings"

	"hcm/pkg/logs"
	"hcm/pkg/tools/converter"
)

// RenderItsmTitle render ITSM ticket title.
func (a *ApplicationOfUpdateSubAccount) RenderItsmTitle() (string, error) {
	return fmt.Sprintf("申请修改[%s]三级账号(%s)", a.Vendor().GetNameZh(), a.subAccountName), nil
}

// RenderItsmForm render ITSM ticket form content.
func (a *ApplicationOfUpdateSubAccount) RenderItsmForm() (string, error) {
	accountData, err := a.GetAccount(a.AccountID())
	if err != nil {
		return "", fmt.Errorf("get account info failed, err: %w", err)
	}

	items := []string{
		fmt.Sprintf("云厂商: %s", a.Vendor().GetNameZh()),
		fmt.Sprintf("所属二级账号: %s", accountData.Name),
		fmt.Sprintf("三级账号名称: %s", a.subAccountName),
	}

	if a.req.Email != nil {
		items = append(items, fmt.Sprintf("修改邮箱: %s", converter.PtrToVal(a.req.Email)))
	}
	if a.req.PhoneNum != nil {
		phone := converter.PtrToVal(a.req.PhoneNum)
		if a.req.CountryCode != nil {
			phone = "+" + converter.PtrToVal(a.req.CountryCode) + " " + phone
		}
		items = append(items, fmt.Sprintf("修改手机号: %s", phone))
	}
	if a.req.Managers != nil {
		items = append(items, fmt.Sprintf("修改管理者: %s", strings.Join(a.req.Managers, ",")))
	}
	if a.req.Memo != nil {
		items = append(items, fmt.Sprintf("修改备注: %s", converter.PtrToVal(a.req.Memo)))
	}

	if a.req.PermissionTemplateIDs != nil {
		if len(a.req.PermissionTemplateIDs) == 0 {
			logs.Errorf("permission template ids is empty,rid:%s", a.Cts.Kit.Rid)
			return "", fmt.Errorf("permission template ids is empty")
		}

		names, err := a.QueryPermissionTemplateNames(a.req.PermissionTemplateIDs)
		if err != nil {
			return "", fmt.Errorf("query permission template names failed, err: %v", err)
		}

		items = append(items, fmt.Sprintf("修改权限模版为: %s", strings.Join(names, ",")))
	}

	return strings.Join(items, "\n"), nil
}
