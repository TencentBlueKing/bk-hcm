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

package createpermtemplate

import (
	"fmt"

	"hcm/pkg/logs"
)

// CheckReq validates the request fields and business logic.
func (a *ApplicationOfCreatePermTemplate) CheckReq() error {
	account, err := a.GetAccount(a.content.AccountID)
	if err != nil {
		return fmt.Errorf("get account(%s) failed, err: %w", a.content.AccountID, err)
	}

	if a.BkBizID() != account.BkBizID {
		return fmt.Errorf("account(%s)'s biz_id is %d, biz_id(%d) has no permission to operate account of it",
			a.content.AccountID, account.BkBizID, a.BkBizID())
	}

	library, err := a.GetPolicyLibraryDetail(a.Cts.Kit, a.content.PolicyLibraryID)
	if err != nil {
		logs.Errorf("get policy library detail failed, libraryID: %s, err: %v, rid: %s", a.content.PolicyLibraryID, err,
			a.Cts.Kit.Rid)
		return fmt.Errorf("get policy library detail failed, err: %w", err)
	}

	if err = a.CheckAccountsBizInScope(a.Cts.Kit, library.BkBizIDs, []string{a.content.AccountID}); err != nil {
		return err
	}

	applied, err := a.CheckAccountApplied(a.Cts.Kit, a.content.PolicyLibraryID, a.content.AccountID)
	if err != nil {
		logs.Errorf("check account applied failed, libraryID: %s, accountID: %s, err: %v, rid: %s",
			a.content.PolicyLibraryID, a.content.AccountID, err, a.Cts.Kit.Rid)
		return fmt.Errorf("check account applied failed, err: %w", err)
	}

	if applied {
		return fmt.Errorf("account %s has already created a permission template from this policy library",
			a.content.AccountID)
	}

	return nil
}
