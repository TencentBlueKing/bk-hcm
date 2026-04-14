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

package deletepermtemplate

import (
	"fmt"
	"strings"

	"hcm/pkg/api/core"
	protocloud "hcm/pkg/api/data-service/cloud"
	"hcm/pkg/dal/dao/tools"
	"hcm/pkg/logs"
)

// RenderItsmTitle renders the ITSM ticket title.
func (a *ApplicationOfDeletePermTemplate) RenderItsmTitle() (string, error) {
	kt := a.Cts.Kit

	req := &protocloud.PermissionTemplateListReq{
		Filter: tools.ExpressionAnd(tools.RuleEqual("id", a.content.ID)),
		Page:   core.NewDefaultBasePage(),
	}
	result, err := a.Client.DataService().Global.PermissionTemplate.ListPermissionTemplate(kt, req)
	if err != nil {
		logs.Errorf("get permission template for itsm title failed, id: %s, err: %v, rid: %s",
			a.content.ID, err, kt.Rid)
		return "", fmt.Errorf("get permission template failed, err: %w", err)
	}

	if len(result.Details) == 0 {
		return "", fmt.Errorf("permission template(%s) not found", a.content.ID)
	}

	return fmt.Sprintf("申请删除云权限模板(%s)", result.Details[0].Name), nil
}

// RenderItsmForm renders the ITSM form content.
func (a *ApplicationOfDeletePermTemplate) RenderItsmForm() (string, error) {
	kt := a.Cts.Kit

	bizName, err := a.GetBizName(a.content.BkBizID)
	if err != nil {
		logs.Errorf("get biz name for itsm form failed, bizID: %d, err: %v, rid: %s",
			a.content.BkBizID, err, kt.Rid)
		return "", err
	}

	req := &protocloud.PermissionTemplateListReq{
		Filter: tools.ExpressionAnd(tools.RuleEqual("id", a.content.ID)),
		Page:   core.NewDefaultBasePage(),
	}
	result, err := a.Client.DataService().Global.PermissionTemplate.ListPermissionTemplate(kt, req)
	if err != nil {
		logs.Errorf("get permission template for itsm form failed, id: %s, err: %v, rid: %s",
			a.content.ID, err, kt.Rid)
		return "", fmt.Errorf("get permission template failed, err: %w", err)
	}

	if len(result.Details) == 0 {
		return "", fmt.Errorf("permission template(%s) not found", a.content.ID)
	}

	tmpl := result.Details[0]

	accountInfo, err := a.GetAccount(tmpl.AccountID)
	if err != nil {
		logs.Errorf("get account for itsm form failed, accountID: %s, err: %v, rid: %s", tmpl.AccountID, err, kt.Rid)
		return "", err
	}

	items := []string{
		fmt.Sprintf("业务: %s", bizName),
		fmt.Sprintf("云厂商: %s", a.content.Vendor.GetNameZh()),
		fmt.Sprintf("云账号: %s", accountInfo.Name),
		fmt.Sprintf("权限模板ID: %s", a.content.ID),
		fmt.Sprintf("权限模板名称: %s", tmpl.Name),
	}

	return strings.Join(items, "\n"), nil
}
