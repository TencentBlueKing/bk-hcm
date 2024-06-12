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

package mainaccount

import (
	"fmt"
	"reflect"
	"strings"
)

type formItem struct {
	Label string
	Value string
}

// RenderItsmTitle 渲染ITSM单据标题
func (a *ApplicationOfUpdateMainAccount) RenderItsmTitle() (string, error) {
	return fmt.Sprintf("[%s]申请修改[%s]二级账号", a.Cts.Kit.User, a.Vendor().GetNameZh()), nil
}

// RenderItsmForm 渲染ITSM表单
func (a *ApplicationOfUpdateMainAccount) RenderItsmForm() (string, error) {
	req := a.req

	// 获取原来的账号
	oldAccount, err := a.Client.DataService().Global.MainAccount.GetBasicInfo(a.Cts.Kit, req.ID)
	if err != nil {
		return "", fmt.Errorf("get old account failed, err: %w", err)
	}

	formItems := []formItem{
		{Label: "云厂商", Value: req.Vendor.GetNameZh()},
		{Label: "站点类型", Value: string(oldAccount.Site)},
		{Label: "账号邮箱", Value: oldAccount.Email},
		{Label: "变更信息如下: ", Value: "        "},
	}

	// 管理员变更
	if !reflect.DeepEqual(req.Managers, oldAccount.Managers) {
		formItems = append(formItems,
			formItem{
				Label: "管理员变更前: ",
				Value: fmt.Sprintf(" %s", strings.Join(req.Managers, ",")),
			},
			formItem{
				Label: "管理员变更后: ",
				Value: fmt.Sprintf(" %s", strings.Join(oldAccount.Managers, ",")),
			},
		)

	}

	// 备份管理员变更
	if !reflect.DeepEqual(req.BakManagers, oldAccount.BakManagers) {
		formItems = append(formItems,
			formItem{
				Label: "备份管理员变更前: ",
				Value: fmt.Sprintf(" %s", strings.Join(oldAccount.BakManagers, ",")),
			},
			formItem{
				Label: "备份管理员变更后: ",
				Value: fmt.Sprintf(" %s", strings.Join(req.BakManagers, ",")),
			},
		)
	}

	// 组织架构变更
	if req.DeptID != oldAccount.DeptID {
		formItems = append(formItems,
			formItem{
				Label: "组织架构变更前: ",
				Value: fmt.Sprintf("%d", oldAccount.DeptID),
			},
			formItem{
				Label: "组织架构变更后: ",
				Value: fmt.Sprintf("%d", req.DeptID),
			},
		)
	}

	// 业务变更
	if req.BkBizID != oldAccount.BkBizID {
		// 获取业务名字
		bkName, err := a.GetBizName(req.BkBizID)
		if err != nil {
			return "", fmt.Errorf("list biz name failed, bk_biz_ids: %v, err: %w", req.BkBizID, err)
		}

		oldBkName, err := a.GetBizName(oldAccount.BkBizID)
		if err != nil {
			return "", fmt.Errorf("list biz name failed, bk_biz_ids: %v, err: %w", req.BkBizID, err)
		}

		formItems = append(formItems,
			formItem{
				Label: "组织架构变更前: ",
				Value: fmt.Sprintf("        %s", oldBkName),
			},
			formItem{
				Label: "组织架构变更后: ",
				Value: fmt.Sprintf("        %s,", bkName),
			},
		)
	}

	// 转换为ITSM表单内容数据
	content := make([]string, 0, len(formItems))
	for _, i := range formItems {
		content = append(content, fmt.Sprintf("%s: %s", i.Label, i.Value))
	}
	return strings.Join(content, "\n"), nil
}
