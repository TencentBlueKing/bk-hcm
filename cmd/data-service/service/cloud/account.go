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
	"hcm/pkg/criteria/validator"
	"hcm/pkg/rest"

	"hcm/pkg/dal/dao"
)

// InitAccountService initial the account service
func InitAccountService(cap *capability.Capability) {
	svr := &account{
		dao: cap.Dao,
	}

	h := rest.NewHandler()
	// RESTful API
	h.Add("CreateAccount", "POST", "/cloud/accounts/", svr.CreateAccount)
	h.Add("UpdateAccount", "PUT", "/cloud/accounts/", svr.UpdateAccount)

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

	if err := validator.Validate.Struct(req); err != nil {
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

	if err := validator.Validate.Struct(req); err != nil {
		return nil, errf.Newf(errf.InvalidParameter, err.Error())
	}

	err := a.dao.CloudAccount().Update(cts.Kit, &req.FilterExpr, req.ToModel())

	return nil, err
}
