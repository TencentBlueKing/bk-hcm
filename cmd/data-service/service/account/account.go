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

	"hcm/cmd/data-service/service/capability"
	"hcm/pkg/api/protocol/base"
	dataservice "hcm/pkg/api/protocol/data-service"
	"hcm/pkg/criteria/errf"
	"hcm/pkg/dal/dao"
	"hcm/pkg/dal/table"
	"hcm/pkg/rest"
)

// InitAccountService initial the account service
func InitAccountService(cap *capability.Capability) {
	svr := &account{
		dao: cap.Dao,
	}

	h := rest.NewHandler()
	h.Add("CreateAccount", "POST", "/api/v1/data/create/account/account", svr.CreateAccount)

	h.Load(cap.WebService)
}

type account struct {
	dao dao.Set
}

// CreateAccount create account with options
func (a *account) CreateAccount(cts *rest.Contexts) (interface{}, error) {
	req := new(dataservice.CreateAccountReq)
	if err := cts.DecodeInto(req); err != nil {
		return nil, errf.New(errf.DecodeRequestFailed, err.Error())
	}

	if err := req.Validate(); err != nil {
		return nil, errf.Newf(errf.InvalidParameter, err.Error())
	}

	account := &table.Account{
		Spec: req.Spec,
		Revision: &table.Revision{
			Creator: cts.Kit.User,
			Reviser: cts.Kit.User,
		},
	}

	id, err := a.dao.Account().Create(cts.Kit, account)
	if err != nil {
		return nil, fmt.Errorf("create account failed, err: %v", err)
	}

	return &base.CreateResult{ID: id}, nil
}
