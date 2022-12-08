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

// Package account defines account service.
package account

import (
	"hcm/cmd/hc-service/service/capability"
	"hcm/pkg/adaptor"
	hcservice "hcm/pkg/api/protocol/hc-service"
	"hcm/pkg/criteria/errf"
	"hcm/pkg/rest"
)

// InitAccountService initial the account service
func InitAccountService(cap *capability.Capability) {
	a := &account{
		ad: cap.Adaptor,
	}

	h := rest.NewHandler()
	h.Add("AccountCheck", "POST", "/account/check", a.AccountCheck)

	h.Load(cap.WebService)
}

type account struct {
	ad adaptor.Adaptor
}

// AccountCheck authentication information and permissions.
func (a account) AccountCheck(cts *rest.Contexts) (interface{}, error) {
	req := new(hcservice.AccountCheckReq)
	if err := cts.DecodeInto(req); err != nil {
		return nil, errf.New(errf.DecodeRequestFailed, err.Error())
	}

	if err := req.Validate(); err != nil {
		return nil, errf.Newf(errf.InvalidParameter, err.Error())
	}

	if err := a.ad.Vendor(req.Vendor).AccountCheck(cts.Kit, req.Secret, req.AccountInfo); err != nil {
		return nil, err
	}

	return nil, nil
}
