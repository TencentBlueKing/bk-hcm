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

	"hcm/pkg/logs"
)

// CheckReq validates the request fields and business logic.
func (a *ApplicationOfUpdatePermTemplate) CheckReq() error {
	accountIDs, err := a.GetPermTmplAccountIDs(a.Cts.Kit, []string{a.content.ID})
	if err != nil {
		logs.Errorf("get permission template for itsm form failed, templateID: %s, err: %v, rid: %s",
			a.content.ID, err, a.Cts.Kit.Rid)
		return fmt.Errorf("get permission template failed, err: %w", err)
	}
	if len(accountIDs) != 1 {
		logs.Errorf("permission template id is invalid, templateID: %s, rid: %s", a.content.ID, a.Cts.Kit.Rid)
		return fmt.Errorf("permission template id is invalid")
	}

	accountID := accountIDs[0]
	account, err := a.GetAccount(accountID)
	if err != nil {
		logs.Errorf("get account for itsm form failed, accountID: %s, err: %v, rid: %s", accountID, err, a.Cts.Kit.Rid)
		return fmt.Errorf("get account failed, err: %w", err)
	}

	if a.BkBizID() != account.BkBizID {
		return fmt.Errorf("account(%s)'s biz_id is %d, biz_id(%d) has no permission to operate account of it",
			accountID, account.BkBizID, a.BkBizID())
	}

	library, err := a.GetPolicyLibraryDetail(a.Cts.Kit, a.content.PolicyLibraryID)
	if err != nil {
		logs.Errorf("get policy library detail failed, libraryID: %s, err: %v, rid: %s", a.content.PolicyLibraryID, err,
			a.Cts.Kit.Rid)
		return fmt.Errorf("get policy library detail failed, err: %w", err)
	}

	if err = a.CheckAccountsBizInScope(a.Cts.Kit, library.BkBizIDs, []string{accountID}); err != nil {
		return err
	}

	err = a.CheckPermTmplUpdatability(a.Cts.Kit, a.Vendor(), []string{a.content.ID}, a.content.PolicyLibraryID)
	if err != nil {
		logs.Errorf("check permission template updatability failed, libraryID: %s, templateID: %s, err: %v, rid: %s",
			library.ID, a.content.ID, err, a.Cts.Kit.Rid)
		return err
	}

	return nil
}
