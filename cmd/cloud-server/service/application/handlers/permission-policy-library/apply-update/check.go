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

package applyupdate

import (
	"errors"
	"fmt"

	"hcm/pkg/logs"
)

// CheckReq validates the request fields and business logic.
func (a *ApplicationOfApplyPermPolicyLibUpdate) CheckReq() error {
	if a.Content.PolicyLibraryID == "" {
		return errors.New("policy_library_id is required")
	}

	if a.Content.PermissionTemplateID == "" {
		return errors.New("permission_template_id is required")
	}

	accountIDs, err := a.GetPermTmplAccountIDs(a.Cts.Kit, []string{a.Content.PermissionTemplateID})
	if err != nil {
		logs.Errorf("get permission template account ids failed, templateID: %s, err: %v, rid: %s",
			a.Content.PermissionTemplateID, err, a.Cts.Kit.Rid)
		return fmt.Errorf("get permission template account ids failed, err: %w", err)
	}
	if len(accountIDs) != 1 {
		logs.Errorf("permission template id is invalid, templateID: %s, rid: %s", a.Content.PermissionTemplateID,
			a.Cts.Kit.Rid)
		return errors.New("permission template id is invalid")
	}

	account, err := a.GetAccount(accountIDs[0])
	if err != nil {
		logs.Errorf("get policy library account failed, accountID: %s, err: %v, rid: %s",
			accountIDs[0], err, a.Cts.Kit.Rid)
		return fmt.Errorf("get policy library account failed, err: %v", err)
	}
	if a.GetBkBizIDs()[0] != account.BkBizID {
		logs.Errorf("get policy library account failed, current biz is %d ,account biz is %d, rid: %s",
			a.GetBkBizIDs()[0], account.BkBizID, a.Cts.Kit.Rid)
		return fmt.Errorf("get policy library account failed, biz mismatch current biz is %d ,account biz is %d",
			a.GetBkBizIDs()[0], account.BkBizID)
	}

	library, err := a.GetPolicyLibraryDetail(a.Cts.Kit, a.Content.PolicyLibraryID)
	if err != nil {
		logs.Errorf("get policy library detail failed, libraryID: %s, err: %v, rid: %s",
			a.Content.PolicyLibraryID, err, a.Cts.Kit.Rid)
		return fmt.Errorf("get policy library detail failed, err: %w", err)
	}

	if err = a.CheckAccountsBizInScope(a.Cts.Kit, library.BkBizIDs, accountIDs); err != nil {
		return err
	}

	return nil
}
