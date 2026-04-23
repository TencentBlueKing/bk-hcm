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

	"hcm/pkg/api/core"
	corecloud "hcm/pkg/api/core/cloud"
	"hcm/pkg/dal/dao/tools"
	"hcm/pkg/logs"
	"hcm/pkg/tools/converter"
)

// CheckReq validate the request and check that the sub account exists.
func (a *ApplicationOfUpdateSubAccount) CheckReq() error {
	if err := a.req.Validate(); err != nil {
		return err
	}

	if err := a.CheckSubAccountExists(a.req.ID); err != nil {
		return err
	}

	account, err := a.GetAccount(a.AccountID())
	if err != nil {
		return fmt.Errorf("found parent account(%s) failed, err: %w", a.AccountID(), err)
	}

	if a.BkBizID() != account.BkBizID {
		return fmt.Errorf("account(%s)'s biz_id is %d,biz_id(%d) no perssion to operate subaccount of it",
			a.AccountID(), account.BkBizID, a.BkBizID())
	}

	if err := a.checkPermissionTemplate(); err != nil {
		return err
	}

	return nil
}

func (a *ApplicationOfUpdateSubAccount) checkPermissionTemplate() error {
	// 不修改
	if a.req.PermissionTemplateIDs == nil {
		return nil
	}

	if len(a.req.PermissionTemplateIDs) == 0 {
		return fmt.Errorf("permission template ids is empty")
	}

	details, err := a.ListPermissionTemplate(a.req.PermissionTemplateIDs)
	if err != nil {
		return fmt.Errorf("list permission templates failed, err: %w", err)
	}

	if len(details) != len(a.req.PermissionTemplateIDs) {
		return fmt.Errorf("permission templates count mismatch, expected %d, got %d",
			len(a.req.PermissionTemplateIDs), len(details))
	}

	for _, tmpl := range details {
		if tmpl.AccountID != a.AccountID() {
			return fmt.Errorf("permission template(id=%s) account_id does not match", tmpl.ID)
		}
	}

	// 获取新增的权限模版
	addedTmpls, err := a.getAddedPermissionTemplates(details)
	if err != nil {
		return err
	}

	// 新增的权限模版必须有权限策略库ID
	for _, tmpl := range addedTmpls {
		if converter.PtrToVal(tmpl.PolicyLibraryID) == "" {
			return fmt.Errorf("permission template(id=%s) has empty policy_library_id", tmpl.ID)
		}
	}

	return nil
}

// getAddedPermissionTemplates returns the newly added permission templates as a map keyed by ID.
func (a *ApplicationOfUpdateSubAccount) getAddedPermissionTemplates(details []corecloud.BasePermissionTemplate) (
	[]corecloud.BasePermissionTemplate, error) {

	subAccounts, err := a.Client.DataService().Global.SubAccount.List(
		a.Cts.Kit,
		&core.ListReq{
			Filter: tools.ExpressionAnd(tools.RuleEqual("id", a.req.ID)),
			Page:   core.NewDefaultBasePage(),
		},
	)
	if err != nil {
		logs.Errorf("get sub account failed, id: %s, err: %v, rid: %s", a.req.ID, err, a.Cts.Kit.Rid)
		return nil, fmt.Errorf("get sub account failed, id: %s, err: %v", a.req.ID, err)
	}
	if len(subAccounts.Details) != 1 {
		logs.Errorf("sub account(id=%s) not found, rid: %s", a.req.ID, a.Cts.Kit.Rid)
		return nil, fmt.Errorf("sub account(id=%s) not found", a.req.ID)
	}

	existingIDs := make(map[string]struct{}, len(subAccounts.Details[0].PermissionTemplateIDs))
	for _, id := range subAccounts.Details[0].PermissionTemplateIDs {
		existingIDs[id] = struct{}{}
	}

	detailsMap := converter.SliceToMap(details, func(tmpl corecloud.BasePermissionTemplate) (
		string, corecloud.BasePermissionTemplate) {
		return tmpl.ID, tmpl
	})

	added := make(map[string]corecloud.BasePermissionTemplate)
	for _, id := range a.req.PermissionTemplateIDs {
		if _, exists := existingIDs[id]; exists {
			continue
		}

		added[id] = detailsMap[id]
	}

	logs.Infof("added permission templates: %v, rid: %s", converter.MapKeyToSlice(added), a.Cts.Kit.Rid)

	return converter.MapValueToSlice(added), nil
}
