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

package deletesubaccount

import (
	"fmt"

	"hcm/pkg/api/core"
	"hcm/pkg/runtime/filter"
)

// CheckReq validate the request and check that the sub account exists.
func (a *ApplicationOfDeleteSubAccount) CheckReq() error {
	if a.req.ID == "" {
		return fmt.Errorf("sub account id is required")
	}

	if err := a.checkSubAccountExists(); err != nil {
		return err
	}

	if _, err := a.GetAccount(a.AccountID()); err != nil {
		return fmt.Errorf("parent account(%s) not found, err: %w", a.AccountID(), err)
	}

	// TODO: 密钥管理功能实现后，需要校验三级账号关联的密钥是否已全部删除，
	// 如果存在未删除的密钥，应阻止删除流程并返回错误提示。

	return nil
}

func (a *ApplicationOfDeleteSubAccount) checkSubAccountExists() error {
	result, err := a.Client.DataService().Global.SubAccount.List(
		a.Cts.Kit,
		&core.ListReq{
			Filter: &filter.Expression{
				Op: filter.And,
				Rules: []filter.RuleFactory{
					filter.AtomRule{
						Field: "id",
						Op:    filter.Equal.Factory(),
						Value: a.req.ID,
					},
				},
			},
			Page: &core.BasePage{Count: true},
		},
	)
	if err != nil {
		return fmt.Errorf("query sub account failed, err: %w", err)
	}

	if result.Count == 0 {
		return fmt.Errorf("sub account(id=%s) not found", a.req.ID)
	}

	return nil
}
