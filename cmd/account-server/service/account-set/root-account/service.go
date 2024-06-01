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

package rootaccount

import (
	"fmt"
	"net/http"

	"hcm/cmd/account-server/logics/audit"
	"hcm/cmd/account-server/service/capability"
	"hcm/pkg/client"
	"hcm/pkg/criteria/errf"
	"hcm/pkg/iam/auth"
	"hcm/pkg/iam/meta"
	"hcm/pkg/rest"
)

// InitService initial the root account service
func InitService(c *capability.Capability) {
	svc := &service{
		client:     c.ApiClient,
		authorizer: c.Authorizer,
		audit:      c.Audit,
	}

	h := rest.NewHandler()

	// register handler
	h.Add("GetRootAccount", http.MethodGet, "/root_accounts/{account_id}", svc.Get)
	h.Add("ListRootAccount", http.MethodPost, "/root_accounts/list", svc.List)
	h.Add("UpdateRootAccount", http.MethodPatch, "/root_accounts/{account_id}", svc.Update)
	h.Add("AddRootAccount", http.MethodPost, "/root_accounts/add", svc.Add)

	h.Load(c.WebService)
}

type service struct {
	client     *client.ClientSet
	authorizer auth.Authorizer
	audit      audit.Interface
}

func (s *service) checkPermission(cts *rest.Contexts, action meta.Action, accountID string) error {
	return s.checkPermissions(cts, action, []string{accountID})
}

func (s *service) checkPermissions(cts *rest.Contexts, action meta.Action, accountIDs []string) error {
	resources := make([]meta.ResourceAttribute, 0, len(accountIDs))
	for _, accountID := range accountIDs {
		resources = append(resources, meta.ResourceAttribute{
			Basic: &meta.Basic{
				Type:       meta.RootAccount,
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
