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
	"net/http"

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
	// 兼容登记账号校验，过渡方案，后期去除
	h.Add("CheckAccount", http.MethodPost, "/accounts/check", svc.CheckAccount)

	h.Add("GetResCountBySecret", http.MethodPost, "/vendors/{vendor}/accounts/res_counts/by_secrets",
		svc.GetResCountBySecret)
	h.Add("GetAccountBySecret", http.MethodPost, "/vendors/{vendor}/accounts/secret", svc.GetAccountBySecret)
	h.Add("CheckByID", http.MethodPost, "/accounts/{account_id}/check", svc.CheckByID)
	h.Add("ListAccount", http.MethodPost, "/accounts/list", svc.ListAccount)
	h.Add("ResourceList", http.MethodPost, "/accounts/resources/accounts/list", svc.ResourceList)
	h.Add("GetAccount", http.MethodGet, "/accounts/{account_id}", svc.GetAccount)
	h.Add("GetSyncDetail", http.MethodGet, "/accounts/sync_details/{account_id}", svc.GetSyncDetail)
	h.Add("UpdateAccount", http.MethodPatch, "/accounts/{account_id}", svc.UpdateAccount)
	h.Add("UpdateBuiltInAccount", http.MethodPatch, "/account/builtIn", svc.UpdateBuiltInAccount)
	h.Add("SyncCloudResource", http.MethodPost, "/accounts/{account_id}/sync", svc.SyncCloudResource)
	h.Add("DeleteAccount", http.MethodDelete, "/accounts/{account_id}", svc.DeleteAccount)
	h.Add("DeleteValidate", http.MethodPost, "/accounts/{account_id}/delete/validate", svc.DeleteValidate)

	h.Add("SyncCloudResourceByCond", http.MethodPost,
		"/vendors/{vendor}/accounts/{account_id}/resources/{res}/sync_by_cond", svc.SyncCloudResourceByCond)
	h.Add("SyncBizCloudResourceByCond", http.MethodPost,
		"/bizs/{bk_biz_id}/vendors/{vendor}/accounts/{account_id}/resources/{res}/sync_by_cond",
		svc.SyncBizCloudResourceByCond)

	// 获取账号配额
	h.Add("GetBizTCloudZoneQuota", http.MethodPost,
		"/bizs/{bk_biz_id}/vendors/tcloud/accounts/{account_id}/zones/quotas",
		svc.GetBizTCloudZoneQuota)
	h.Add("GetBizHuaWeiRegionQuota", http.MethodPost,
		"/bizs/{bk_biz_id}/vendors/huawei/accounts/{account_id}/regions/quotas", svc.GetBizHuaWeiRegionQuota)
	h.Add("GetBizGcpRegionQuota", http.MethodPost, "/bizs/{bk_biz_id}/vendors/gcp/accounts/{account_id}/regions/quotas",
		svc.GetBizGcpRegionQuota)
	h.Add("GetResTCloudZoneQuota", http.MethodPost, "/vendors/tcloud/accounts/{account_id}/zones/quotas",
		svc.GetResTCloudZoneQuota)
	h.Add("GetResHuaWeiRegionQuota", http.MethodPost,
		"/vendors/huawei/accounts/{account_id}/regions/quotas", svc.GetResHuaWeiRegionQuota)
	h.Add("GetResGcpRegionQuota", http.MethodPost, "/vendors/gcp/accounts/{account_id}/regions/quotas",
		svc.GetResGcpRegionQuota)

	// Rel
	h.Add("ListByUsageBizID", http.MethodGet, "/accounts/bizs/{bk_biz_id}", svc.ListByUsageBizID)

	// 安全所需OpenAPI
	h.Add("ListWithExtension", http.MethodPost, "/accounts/extensions/list", svc.ListWithExtension)
	h.Add("ListSecretKey", http.MethodPost, "/accounts/secrets/list", svc.ListSecretKey)

	// 通过密钥获取账号权限策略
	h.Add("ListTCloudAuthPolicies", http.MethodPost, "/vendors/tcloud/accounts/auth_policies/list",
		svc.ListTCloudAuthPolicies)

	h.Add("GetTCloudNetworkAccountType", http.MethodGet, "/vendors/tcloud/accounts/{account_id}/network_type",
		svc.GetTCloudNetworkAccountType)

	h.Add("BizGetAccountUsageBizs", http.MethodGet, "/bizs/{bk_biz_id}/accounts/usage_bizs/{account_id}",
		svc.BizGetAccountUsageBizs)
	h.Add("GetAccountUsageBizs", http.MethodGet, "/accounts/usage_bizs/{account_id}",
		svc.GetAccountUsageBizs)

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

func (a *accountSvc) listAuthorized(cts *rest.Contexts, action meta.Action,
	typ meta.ResourceType) ([]string, bool, error) {
	resources, err := a.authorizer.ListAuthorizedInstances(cts.Kit, &meta.ListAuthResInput{Type: typ,
		Action: action})
	if err != nil {
		return []string{}, false, errf.NewFromErr(
			errf.PermissionDenied,
			fmt.Errorf("list account of %s permission failed, err: %v", action, err),
		)
	}

	return resources.IDs, resources.IsAny, err

}
