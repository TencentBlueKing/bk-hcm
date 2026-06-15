/*
 * TencentBlueKing is pleased to support the open source community by making
 * 蓝鲸智云 - 混合云管理平台 (BlueKing - Hybrid Cloud Management System) available.
 * Copyright (C) 2022 THL A29 Limited,
 * a Tencent company. All rights reserved.
 * Licensed under the MIT License (the "License");
 * you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at http://opensource.org/licenses/MIT
 * Unless required by applicable law or agreed to in writing,
 * software distributed under the MIT License is distributed on
 * an "AS IS" BASIS, WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND,
 * either express or implied. See the License for the
 * specific language governing permissions and limitations under the License.
 *
 * We undertake not to change the open source license (MIT license) applicable
 *
 * to the current version of the project delivered to anyone in the future.
 */

package updatepermtemplate

import (
	"fmt"
	"strings"

	"hcm/pkg/logs"
)

// RenderItsmTitle renders the ITSM ticket title.
func (a *ApplicationOfUpdatePermTemplate) RenderItsmTitle() (string, error) {
	return fmt.Sprintf("申请更新云权限模板(%s)", a.content.ID), nil
}

// RenderItsmForm renders the ITSM form content.
func (a *ApplicationOfUpdatePermTemplate) RenderItsmForm() (string, error) {
	bizName, err := a.GetBizName(a.content.BkBizID)
	if err != nil {
		logs.Errorf("get biz name for itsm form failed, bizID: %d, err: %v, rid: %s", a.content.BkBizID, err,
			a.Cts.Kit.Rid)
		return "", err
	}

	library, err := a.GetPolicyLibraryDetail(a.Cts.Kit, a.content.PolicyLibraryID)
	if err != nil {
		logs.Errorf("get policy library detail for itsm form failed, libraryID: %s, err: %v, rid: %s",
			a.content.PolicyLibraryID, err, a.Cts.Kit.Rid)
		return "", fmt.Errorf("get policy library detail failed, err: %w", err)
	}

	accountIDs, err := a.GetPermTmplAccountIDs(a.Cts.Kit, []string{a.content.ID})
	if err != nil {
		logs.Errorf("get permission template for itsm form failed, templateID: %s, err: %v, rid: %s",
			a.content.ID, err, a.Cts.Kit.Rid)
		return "", fmt.Errorf("get permission template failed, err: %w", err)
	}
	if len(accountIDs) != 1 {
		logs.Errorf("permission template id is invalid, templateID: %s, rid: %s", a.content.ID, a.Cts.Kit.Rid)
		return "", fmt.Errorf("permission template id is invalid")
	}

	accountID := accountIDs[0]
	accountInfo, err := a.GetAccount(accountID)
	if err != nil {
		logs.Errorf("get account for itsm form failed, accountID: %s, err: %v, rid: %s", accountID, err,
			a.Cts.Kit.Rid)
		return "", err
	}

	items := []string{
		fmt.Sprintf("业务: %s", bizName),
		fmt.Sprintf("云厂商: %s", a.content.Vendor.GetNameZh()),
		fmt.Sprintf("云账号: %s", accountInfo.Name),
		fmt.Sprintf("权限模版ID: %s", a.content.ID),
		fmt.Sprintf("权限策略库: %s", library.Name),
		fmt.Sprintf("策略库ID: %s", library.ID),
		fmt.Sprintf("策略内容: %s", library.PolicyDocument),
	}

	return strings.Join(items, "\n"), nil
}
