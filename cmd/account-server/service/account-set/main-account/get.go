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

	"hcm/pkg/criteria/enumor"
	"hcm/pkg/criteria/errf"
	"hcm/pkg/iam/meta"
	"hcm/pkg/rest"
)

// Get get main account with options
func (s *service) Get(cts *rest.Contexts) (interface{}, error) {
	accountID := cts.PathParameter("account_id").String()

	// 校验用户有该账号的查看权限
	if err := s.checkMainAccountPermission(cts, meta.Find, accountID); err != nil {
		return nil, err
	}

	// 查询该账号对应的Vendor
	baseInfo, err := s.client.DataService().Global.MainAccount.GetBasicInfo(cts.Kit, accountID)
	if err != nil {
		return nil, errf.NewFromErr(errf.InvalidParameter, err)
	}

	switch baseInfo.Vendor {
	case enumor.Aws:
		account, err := s.client.DataService().Aws.MainAccount.Get(cts.Kit, accountID)
		if account != nil {
			account.Extension.CloudInitPassword = ""
		}
		return account, err
	case enumor.Gcp:
		account, err := s.client.DataService().Gcp.MainAccount.Get(cts.Kit, accountID)
		// 	 nothing to set null
		// if account != nil {
		// }
		return account, err
	case enumor.Azure:
		account, err := s.client.DataService().Azure.MainAccount.Get(cts.Kit, accountID)
		// 	 nothing to set null
		if account != nil {
			account.Extension.CloudInitPassword = ""
		}
		return account, err
	case enumor.HuaWei:
		account, err := s.client.DataService().HuaWei.MainAccount.Get(cts.Kit, accountID)
		if account != nil {
			account.Extension.CloudInitPassword = ""
		}
		return account, err
	case enumor.Zenlayer:
		account, err := s.client.DataService().Zenlayer.MainAccount.Get(cts.Kit, accountID)
		if account != nil {
			account.Extension.CloudInitPassword = ""
		}
		return account, err
	case enumor.Kaopu:
		account, err := s.client.DataService().Kaopu.MainAccount.Get(cts.Kit, accountID)
		if account != nil {
			account.Extension.CloudInitPassword = ""
		}
		return account, err
	default:
		return nil, errf.NewFromErr(errf.InvalidParameter, fmt.Errorf("no support vendor: %s", baseInfo.Vendor))
	}
}

func (s *service) checkMainAccountPermission(cts *rest.Contexts, action meta.Action, accountID string) error {
	return s.checkPermissions(cts, action, []string{accountID})
}

func (s *service) checkPermissions(cts *rest.Contexts, action meta.Action, accountIDs []string) error {
	resources := make([]meta.ResourceAttribute, 0, len(accountIDs))
	for _, accountID := range accountIDs {
		resources = append(resources, meta.ResourceAttribute{
			Basic: &meta.Basic{
				Type:       meta.MainAccount,
				Action:     action,
				ResourceID: accountID,
			},
		})
	}

	decisions, authorized, err := s.authorizer.Authorize(cts.Kit, resources...)
	if err != nil {
		return errf.NewFromErr(
			errf.PermissionDenied,
			fmt.Errorf("check %s account permissions failed, err: %v", action, err),
		)
	}

	if !authorized {
		// 查询无权限的ID列表，用于提示
		unauthorizedIDs := make([]string, 0, len(accountIDs))
		for index, d := range decisions {
			if !d.Authorized && index < len(accountIDs) {
				unauthorizedIDs = append(unauthorizedIDs, accountIDs[index])
			}
		}

		return errf.NewFromErr(
			errf.PermissionDenied,
			fmt.Errorf("you have not permission of %s accounts(ids=%v)", action, unauthorizedIDs),
		)
	}

	return nil
}
