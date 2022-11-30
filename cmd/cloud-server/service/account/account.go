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

	"hcm/cmd/cloud-server/service/capability"
	"hcm/pkg/api"
	cloudserver "hcm/pkg/api/protocol/cloud-server"
	dataservice "hcm/pkg/api/protocol/data-service"
	"hcm/pkg/criteria/errf"
	"hcm/pkg/dal/table"
	"hcm/pkg/iam/auth"
	"hcm/pkg/iam/meta"
	"hcm/pkg/rest"
)

// InitAccountService initial the account service
func InitAccountService(c *capability.Capability) {
	svr := &account{
		client:     c.ApiClient,
		authorizer: c.Authorizer,
	}

	h := rest.NewHandler()
	h.Add("CreateAccount", "POST", "/create/account/account", svr.CreateAccount)

	h.Load(c.WebService)
}

type account struct {
	client     *api.ClientSet
	authorizer auth.Authorizer
}

// CreateAccount create account with options
func (a *account) CreateAccount(cts *rest.Contexts) (interface{}, error) {
	req := new(cloudserver.CreateAccountReq)
	if err := cts.DecodeInto(req); err != nil {
		return nil, errf.New(errf.DecodeRequestFailed, err.Error())
	}

	authRes := meta.ResourceAttribute{Basic: &meta.Basic{Type: meta.Account, Action: meta.Create}}
	err := a.authorizer.AuthorizeWithPerm(cts.Kit, authRes)
	if err != nil {
		return nil, err
	}

	if err := req.Validate(); err != nil {
		return nil, errf.Newf(errf.InvalidParameter, err.Error())
	}

	createReq := &dataservice.CreateAccountReq{
		Spec: &table.AccountSpec{
			Name: req.Name,
			Memo: req.Memo,
		},
	}

	res, err := a.client.DataService().Account().Create(cts.Kit.Ctx, cts.Kit.Header(), createReq)
	if err != nil {
		return nil, fmt.Errorf("create account failed, err: %v", err)
	}

	return res, nil
}
