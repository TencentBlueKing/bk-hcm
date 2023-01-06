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
	rawjson "encoding/json"

	"hcm/pkg/tools/json"

	"hcm/cmd/cloud-server/service/capability"
	"hcm/pkg/client"
	"hcm/pkg/criteria/errf"
	"hcm/pkg/iam/auth"
	"hcm/pkg/logs"
	"hcm/pkg/rest"
)

// InitAccountService initial the account service
func InitAccountService(c *capability.Capability) {
	svc := &accountSvc{
		client:     c.ApiClient,
		authorizer: c.Authorizer,
	}

	h := rest.NewHandler()
	h.Add("Create", "POST", "/accounts/create", svc.Create)
	h.Add("Check", "POST", "/accounts/check", svc.Check)
	h.Add("CheckByID", "POST", "/accounts/{account_id}/check", svc.CheckByID)
	h.Add("List", "POST", "/accounts/list", svc.List)
	h.Add("Get", "GET", "/accounts/{account_id}", svc.Get)
	h.Add("Update", "PATCH", "/accounts/{account_id}", svc.Update)

	h.Load(c.WebService)
}

type accountSvc struct {
	client     *client.ClientSet
	authorizer auth.Authorizer
}

func (a *accountSvc) decodeExtension(cts *rest.Contexts, rawExtension rawjson.RawMessage, extension interface{}) error {
	err := json.Unmarshal(rawExtension, &extension)
	if err != nil {
		logs.ErrorDepthf(1, "decode extension from request body failed, err: %s, rid: %s", err.Error(), cts.Kit.Rid)
		return errf.NewFromErr(errf.InvalidParameter, err)
	}
	return nil
}
