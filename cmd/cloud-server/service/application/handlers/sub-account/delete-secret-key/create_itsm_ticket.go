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

package deletesecretkey

import (
	"fmt"
	"strings"

	"hcm/pkg/logs"
)

// RenderItsmTitle render ITSM ticket title.
func (a *ApplicationOfDeleteSecretKey) RenderItsmTitle() (string, error) {
	subAccountName, err := a.getSubAccountNameForDisplay()
	if err != nil {
		logs.Errorf("get sub account name for itsm title failed, err: %v, rid: %s", err, a.Cts.Kit.Rid)
		return "", fmt.Errorf("get sub account name for itsm title failed, err: %w", err)
	}

	return fmt.Sprintf("申请删除[%s]三级账号(%s)密钥", a.Vendor().GetNameZh(), subAccountName), nil
}

// RenderItsmForm render ITSM ticket form content.
func (a *ApplicationOfDeleteSecretKey) RenderItsmForm() (string, error) {
	accountData, err := a.GetAccount(a.AccountID())
	if err != nil {
		return "", fmt.Errorf("get account info failed, err: %w", err)
	}

	subAccountName, err := a.getSubAccountNameForDisplay()
	if err != nil {
		logs.Errorf("get sub account name for itsm form failed, err: %v, rid: %s", err, a.Cts.Kit.Rid)
		return "", fmt.Errorf("get sub account name for itsm form failed, err: %w", err)
	}

	cloudSecretID, err := a.getCloudSecretIDForDisplay()
	if err != nil {
		logs.Errorf("get cloud secret id for itsm form failed, err: %v, rid: %s", err, a.Cts.Kit.Rid)
		return "", fmt.Errorf("get cloud secret id for itsm form failed, err: %w", err)
	}

	items := []string{
		fmt.Sprintf("云厂商: %s", a.Vendor().GetNameZh()),
		fmt.Sprintf("所属二级账号: %s", accountData.Name),
		fmt.Sprintf("三级账号名称: %s", subAccountName),
		fmt.Sprintf("密钥ID: %s", cloudSecretID),
	}

	return strings.Join(items, "\n"), nil
}
