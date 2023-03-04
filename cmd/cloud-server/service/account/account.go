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

package account

import (
	"fmt"

	"hcm/cmd/cloud-server/logics/audit"
	"hcm/cmd/cloud-server/service/capability"
	"hcm/pkg/client"
	"hcm/pkg/criteria/errf"
	"hcm/pkg/iam/auth"
	"hcm/pkg/iam/meta"
	"hcm/pkg/rest"
)

// InitAccountService initial the account service
func InitAccountService(c *capability.Capability) {
	svc := &accountSvc{
		client:     c.ApiClient,
		authorizer: c.Authorizer,
		audit:      c.Audit,
	}

	h := rest.NewHandler()
	h.Add("Create", "POST", "/accounts/create", svc.Create)
	h.Add("Check", "POST", "/accounts/check", svc.Check)
	h.Add("CheckByID", "POST", "/accounts/{account_id}/check", svc.CheckByID)
	h.Add("List", "POST", "/accounts/list", svc.List)
	h.Add("Get", "GET", "/accounts/{account_id}", svc.Get)
	h.Add("Update", "PATCH", "/accounts/{account_id}", svc.Update)

	// Rel
	h.Add("ListByBkBizID", "GET", "/accounts/bizs/{bk_biz_id}", svc.ListByBkBizID)

	// 安全所需OpenAPI
	h.Add("ListWithExtension", "POST", "/accounts/extensions/list", svc.ListWithExtension)
	h.Add("ListSecretKey", "POST", "/accounts/secrets/list", svc.ListSecretKey)

	h.Load(c.WebService)
}

type accountSvc struct {
	client     *client.ClientSet
	authorizer auth.Authorizer
	audit      audit.Interface
}

func (a *accountSvc) checkPermission(cts *rest.Contexts, action meta.Action, accountID string) error {
	return a.checkPermissions(cts, action, []string{accountID})
}

func (a *accountSvc) checkPermissions(cts *rest.Contexts, action meta.Action, accountIDs []string) error {
	resources := make([]meta.ResourceAttribute, 0, len(accountIDs))
	for _, accountID := range accountIDs {
		resources = append(resources, meta.ResourceAttribute{
			Basic: &meta.Basic{
				Type:       meta.Account,
				Action:     action,
				ResourceID: accountID,
			},
		})
	}

	decisions, authorized, err := a.authorizer.Authorize(cts.Kit, resources...)
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

func (a *accountSvc) listAuthorized(cts *rest.Contexts, action meta.Action) ([]string, bool, error) {
	resources, err := a.authorizer.ListAuthorizedInstances(cts.Kit, &meta.ListAuthResInput{Type: meta.Account,
		Action: action})
	if err != nil {
		return []string{}, false, errf.NewFromErr(
			errf.PermissionDenied,
			fmt.Errorf("list account of %s permission failed, err: %v", action, err),
		)
	}

	return resources.IDs, resources.IsAny, err

}
