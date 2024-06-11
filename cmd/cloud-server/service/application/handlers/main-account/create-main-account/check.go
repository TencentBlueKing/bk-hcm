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
	"encoding/json"
	"fmt"

	"hcm/pkg/api/core"
	"hcm/pkg/criteria/enumor"
	"hcm/pkg/dal/dao/tools"
	"hcm/pkg/logs"
	"hcm/pkg/runtime/filter"
)

type fields struct {
	key   string
	value string
}

// CheckReq 申请单的表单校验
func (a *ApplicationOfCreateMainAccount) CheckReq() error {
	if err := a.req.Validate(); err != nil {
		return err
	}

	// 检查vendor
	switch a.req.Vendor {
	case enumor.Aws:
	case enumor.Gcp:
	case enumor.HuaWei:
	case enumor.Azure:
	case enumor.Zenlayer:
	case enumor.Kaopu:
	default:
		return fmt.Errorf("vendor [%s] is not supported", a.req.Vendor)
	}

	// 检查参数
	if len(a.req.Managers) < 1 || len(a.req.Managers) > 5 || len(a.req.BakManagers) < 1 || len(a.req.BakManagers) > 5 {
		return fmt.Errorf("managers and backup managers length should be 1~5")
	}

	// 检查扩展参数
	if a.req.Extension == nil {
		return fmt.Errorf("extension is nil")
	}
	if _, err := json.Marshal(a.req.Extension); err != nil {
		return fmt.Errorf("json marshal extension failed, err: %w", err)
	}

	// 检查扩展参数
	account_name := a.req.Extension[a.req.Vendor.GetMainAccountNameFieldName()]
	if account_name == "" {
		return fmt.Errorf("extension[%s] is empty", a.req.Vendor.GetMainAccountNameFieldName())
	}

	// 检查邮箱是否重复
	if err := a.isDuplicateEmail(a.req.Vendor, a.req.Email); err != nil {
		logs.Errorf("check email duplicate failed, err: %s, rid: %s", err, a.Cts.Kit.Rid)
		return err
	}

	// 检查名称是否重复
	if err := a.isDuplicateName(a.req.Vendor.GetMainAccountNameFieldName(), account_name); err != nil {
		logs.Errorf("check name duplicate failed, err: %s, rid: %s", err, a.Cts.Kit.Rid)
		return err
	}

	return nil
}

func (a *ApplicationOfCreateMainAccount) isDuplicateEmail(vendor enumor.Vendor, email string) error {
	rules := []*filter.AtomRule{
		tools.RuleEqual("email", email),
		tools.RuleEqual("vendor", string(vendor)),
	}

	return a.isDuplicateField(rules...)
}

func (a *ApplicationOfCreateMainAccount) isDuplicateName(field, name string) error {
	rule := tools.RuleJSONEqual(fmt.Sprintf("extension.%s", field), name)

	if err := a.isDuplicateField(rule); err != nil {
		return fmt.Errorf("main account [%s] duplicate checking err, err: %s", name, err.Error())
	}
	return nil
}

func (a *ApplicationOfCreateMainAccount) isDuplicateField(rules ...*filter.AtomRule) error {
	if len(rules) == 0 {
		return nil
	}

	// TODO: 后续需要解决并发问题
	// 后台查询是否主账号重复
	result, err := a.Client.DataService().Global.MainAccount.List(
		a.Cts.Kit,
		&core.ListWithoutFieldReq{
			Filter: tools.ExpressionAnd(rules...),
			Page: &core.BasePage{
				Count: true,
			},
		},
	)
	if err != nil {
		return err
	}

	if result.Count > 0 {
		return fmt.Errorf("apply info account has already exits, should be not duplicate, duplicate: %v", rules)
	}

	return nil
}
