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

package applycreate

import (
	"fmt"
	"strings"

	"hcm/pkg/logs"
)

// RenderItsmTitle renders the ITSM ticket title.
func (a *ApplicationOfApplyPermPolicyLibCreate) RenderItsmTitle() (string, error) {
	library, err := a.GetPolicyLibraryDetail(a.Cts.Kit, a.Content.PolicyLibraryID)
	if err != nil {
		logs.Errorf("get policy library detail for itsm title failed, libraryID: %s, err: %v, rid: %s",
			a.Content.PolicyLibraryID, err, a.Cts.Kit.Rid)
		return "", fmt.Errorf("get policy library detail failed, err: %w", err)
	}

	return fmt.Sprintf("申请应用权限策略库(%s)到账号(%s)", library.Name, a.Content.AccountID), nil
}

// RenderItsmForm renders the ITSM form content.
func (a *ApplicationOfApplyPermPolicyLibCreate) RenderItsmForm() (string, error) {
	bizName, err := a.GetBizName(a.Content.BkBizID)
	if err != nil {
		logs.Errorf("get biz name for itsm form failed, bizID: %s, err: %v, rid: %s", a.Content.BkBizID, err,
			a.Cts.Kit.Rid)
		return "", err
	}

	accountInfo, err := a.GetAccount(a.Content.AccountID)
	if err != nil {
		logs.Errorf("get account for itsm form failed, accountID: %s, err: %v, rid: %s", a.Content.AccountID, err,
			a.Cts.Kit.Rid)
		return "", err
	}

	library, err := a.GetPolicyLibraryDetail(a.Cts.Kit, a.Content.PolicyLibraryID)
	if err != nil {
		logs.Errorf("get policy library detail for itsm form failed, libraryID: %s, err: %v, rid: %s",
			a.Content.PolicyLibraryID, err, a.Cts.Kit.Rid)
		return "", fmt.Errorf("get policy library detail failed, err: %w", err)
	}

	items := []string{
		fmt.Sprintf("业务: %s", bizName),
		fmt.Sprintf("云厂商: %s", a.Content.Vendor.GetNameZh()),
		fmt.Sprintf("云账号: %s", accountInfo.Name),
		fmt.Sprintf("权限策略库: %s", library.Name),
		fmt.Sprintf("策略库ID: %s", library.ID),
		fmt.Sprintf("策略内容: %s", library.PolicyDocument),
	}

	return strings.Join(items, "\n"), nil
}
