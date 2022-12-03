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

package cloud

import (
	"fmt"

	"hcm/cmd/data-service/service/capability"
	"hcm/pkg/api/protocol/base"
	"hcm/pkg/api/protocol/data-service/cloud"
	"hcm/pkg/criteria/errf"
	"hcm/pkg/rest"

	"hcm/pkg/dal/dao"
)

// InitAccountService initial the account service
func InitAccountService(cap *capability.Capability) {
	svr := &account{
		dao: cap.Dao,
	}

	h := rest.NewHandler()

	// 采用类似 iac 接口的结构简化处理, 不遵循 RESTful 风格
	h.Add("CreateAccount", "POST", "/cloud/accounts/create/", svr.CreateAccount)
	h.Add("UpdateAccount", "POST", "/cloud/accounts/update/", svr.UpdateAccount)
	h.Add("ListAccounts", "POST", "/cloud/accounts/list/", svr.ListAccounts)

	h.Load(cap.WebService)
}

type account struct {
	dao dao.Set
}

// CreateAccount create account with options
func (a *account) CreateAccount(cts *rest.Contexts) (interface{}, error) {
	req := new(cloud.CreateAccountReq)

	if err := cts.DecodeInto(req); err != nil {
		return nil, errf.New(errf.DecodeRequestFailed, err.Error())
	}

	if err := req.Validate(); err != nil {
		return nil, errf.Newf(errf.InvalidParameter, err.Error())
	}

	id, err := a.dao.CloudAccount().Create(cts.Kit, req.ToModel())
	if err != nil {
		return nil, fmt.Errorf("create cloud account failed, err: %v", err)
	}

	return &base.CreateResult{ID: id}, nil
}

// UpdateAccount create account with options
func (a *account) UpdateAccount(cts *rest.Contexts) (interface{}, error) {
	req := new(cloud.UpdateAccountReq)

	if err := cts.DecodeInto(req); err != nil {
		return nil, errf.New(errf.DecodeRequestFailed, err.Error())
	}

	if err := req.Validate(); err != nil {
		return nil, errf.Newf(errf.InvalidParameter, err.Error())
	}

	err := a.dao.CloudAccount().Update(cts.Kit, &req.FilterExpr, req.ToModel())

	return nil, err
}

// ListAccounts create account with options
func (a *account) ListAccounts(cts *rest.Contexts) (interface{}, error) {
	req := new(cloud.ListAccountsReq)
	if err := cts.DecodeInto(req); err != nil {
		return nil, err
	}

	if err := req.Validate(); err != nil {
		return nil, errf.Newf(errf.InvalidParameter, err.Error())
	}
	mData, err := a.dao.CloudAccount().List(cts.Kit, req.ToListOption())
	if err != nil {
		return nil, err
	}

	var details []cloud.AccountData
	for _, m := range mData {
		details = append(details, *cloud.NewAccountData(m))
	}

	return &cloud.ListAccountsResult{Details: details}, nil
}
