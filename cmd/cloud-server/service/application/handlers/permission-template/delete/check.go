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

	"hcm/pkg/api/core"
	protocloud "hcm/pkg/api/data-service/cloud"
	"hcm/pkg/criteria/enumor"
	"hcm/pkg/criteria/errf"
	"hcm/pkg/dal/dao/tools"
	"hcm/pkg/logs"
)

// CheckReq validates the request fields and business logic before creating the application.
func (a *ApplicationOfDeletePermTemplate) CheckReq() error {
	switch a.Vendor() {
	case enumor.TCloud:
		return a.checkTCloud()
	default:
		return errf.Newf(errf.InvalidParameter, "unsupported vendor: %s", a.Vendor())
	}
}

func (a *ApplicationOfDeletePermTemplate) checkTCloud() error {
	kt := a.Cts.Kit

	req := &protocloud.PermissionTemplateExtListReq{
		Filter: tools.ExpressionAnd(tools.RuleEqual("id", a.content.ID)),
		Page:   core.NewDefaultBasePage(),
	}
	result, err := a.Client.DataService().TCloud.PermissionTemplate.ListPermissionTemplateExt(kt, req)
	if err != nil {
		logs.Errorf("get tcloud permission template failed, id: %s, err: %v, rid: %s", a.content.ID, err, kt.Rid)
		return fmt.Errorf("get permission template failed, err: %w", err)
	}

	if len(result.Details) == 0 {
		return fmt.Errorf("permission template(%s) not found", a.content.ID)
	}

	tmpl := result.Details[0]

	if tmpl.Extension == nil || tmpl.Extension.CloudType != enumor.TCloudCustomPolicy {
		return fmt.Errorf("只有自定义策略模板才允许删除")
	}

	account, err := a.GetAccount(tmpl.AccountID)
	if err != nil {
		logs.Errorf("get account failed, accountID: %s, err: %v, rid: %s", tmpl.AccountID, err, kt.Rid)
		return fmt.Errorf("get account failed, err: %w", err)
	}

	if a.BkBizID() != account.BkBizID {
		return fmt.Errorf("account(%s)'s biz_id is %d, biz_id(%d) has no permission to operate account of it",
			tmpl.AccountID, account.BkBizID, a.BkBizID())
	}

	return a.checkSubAccountCount(tmpl.ID)
}

func (a *ApplicationOfDeletePermTemplate) checkSubAccountCount(templateID string) error {
	kt := a.Cts.Kit

	countResult, err := a.Client.DataService().Global.SubAccount.List(kt, &core.ListReq{
		Filter: tools.ExpressionAnd(tools.RuleJSONContains("permission_template_ids", templateID)),
		Page:   core.NewCountPage(),
	})
	if err != nil {
		logs.Errorf("count sub accounts linked to template failed, templateID: %s, err: %v, rid: %s", templateID, err,
			kt.Rid)
		return fmt.Errorf("count linked sub accounts failed, err: %w", err)
	}

	if countResult.Count > 0 {
		return errf.Newf(errf.InvalidParameter, "权限模板(%s)已关联 %d 个三级账号，无法删除", templateID,
			countResult.Count)
	}

	return nil
}
