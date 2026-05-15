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
	typeaccount "hcm/pkg/adaptor/types/account"
	hssubaccount "hcm/pkg/api/hc-service/sub-account"
	"hcm/pkg/criteria/errf"
	"hcm/pkg/logs"
	"hcm/pkg/rest"
)

// TCloudListAccount list tcloud accounts (CAM ListUsers).
func (svc *service) TCloudListAccount(cts *rest.Contexts) (interface{}, error) {
	req := new(hssubaccount.TCloudListAccountReq)
	if err := cts.DecodeInto(req); err != nil {
		return nil, errf.NewFromErr(errf.DecodeRequestFailed, err)
	}
	if err := req.Validate(); err != nil {
		return nil, errf.NewFromErr(errf.InvalidParameter, err)
	}

	client, err := svc.ad.TCloud(cts.Kit, req.AccountID)
	if err != nil {
		return nil, err
	}

	result, err := client.ListAccount(cts.Kit)
	if err != nil {
		logs.Errorf("list tcloud account failed, err: %v, account: %s, rid: %s",
			err, req.AccountID, cts.Kit.Rid)
		return nil, err
	}

	return result, nil
}

// TCloudCreateSubAccount create tcloud subaccount (CAM AddUser).
func (svc *service) TCloudCreateSubAccount(cts *rest.Contexts) (interface{}, error) {
	req := new(hssubaccount.TCloudCreateSubAccountReq)
	if err := cts.DecodeInto(req); err != nil {
		return nil, errf.NewFromErr(errf.DecodeRequestFailed, err)
	}
	if err := req.Validate(); err != nil {
		return nil, errf.NewFromErr(errf.InvalidParameter, err)
	}

	client, err := svc.ad.TCloud(cts.Kit, req.AccountID)
	if err != nil {
		return nil, err
	}

	result, err := client.AddUser(cts.Kit, &typeaccount.AddUserOption{
		Name: req.Name,
		// 1代表生成子账户密钥,0代表不生成,创建子账户不生成密钥,用户需要到海垒上创建对应密钥。
		UseAPI:       0,
		Remark:       req.Remark,
		Email:        req.Email,
		PhoneNum:     req.PhoneNum,
		ConsoleLogin: uint64(*req.ConsoleLogin),
		CountryCode:  req.CountryCode,
	})
	if err != nil {
		return nil, err
	}

	return result, nil
}

// TCloudUpdateSubAccount update tcloud subaccount (CAM UpdateUser).
func (svc *service) TCloudUpdateSubAccount(cts *rest.Contexts) (interface{}, error) {
	req := new(hssubaccount.TCloudUpdateSubAccountReq)
	if err := cts.DecodeInto(req); err != nil {
		return nil, errf.NewFromErr(errf.DecodeRequestFailed, err)
	}
	if err := req.Validate(); err != nil {
		return nil, errf.NewFromErr(errf.InvalidParameter, err)
	}

	client, err := svc.ad.TCloud(cts.Kit, req.AccountID)
	if err != nil {
		return nil, err
	}

	opt := &typeaccount.UpdateUserOption{Name: req.Name}
	if req.Remark != nil {
		opt.Remark = req.Remark
	}
	if req.Email != nil {
		opt.Email = req.Email
	}
	if req.PhoneNum != nil {
		opt.PhoneNum = req.PhoneNum
	}
	if req.CountryCode != nil {
		opt.CountryCode = req.CountryCode
	}

	if err = client.UpdateUser(cts.Kit, opt); err != nil {
		return nil, err
	}

	return nil, nil
}

// TCloudDeleteSubAccount delete tcloud subaccount (CAM DeleteUser).
func (svc *service) TCloudDeleteSubAccount(cts *rest.Contexts) (interface{}, error) {
	req := new(hssubaccount.TCloudDeleteSubAccountReq)
	if err := cts.DecodeInto(req); err != nil {
		return nil, errf.NewFromErr(errf.DecodeRequestFailed, err)
	}
	if err := req.Validate(); err != nil {
		return nil, errf.NewFromErr(errf.InvalidParameter, err)
	}

	client, err := svc.ad.TCloud(cts.Kit, req.AccountID)
	if err != nil {
		return nil, err
	}

	if err = client.DeleteUser(cts.Kit, req.Name); err != nil {
		return nil, err
	}

	return nil, nil
}

// TCloudDescribeSubAccounts query sub accounts by UIN list (CAM DescribeSubAccounts).
func (svc *service) TCloudDescribeSubAccounts(cts *rest.Contexts) (interface{}, error) {
	req := new(hssubaccount.TCloudDescribeSubAccountsReq)
	if err := cts.DecodeInto(req); err != nil {
		return nil, errf.NewFromErr(errf.DecodeRequestFailed, err)
	}
	if err := req.Validate(); err != nil {
		return nil, errf.NewFromErr(errf.InvalidParameter, err)
	}

	client, err := svc.ad.TCloud(cts.Kit, req.AccountID)
	if err != nil {
		return nil, err
	}

	result, err := client.DescribeSubAccounts(cts.Kit, &typeaccount.DescribeSubAccountsOption{
		SubUin: req.SubUin,
	})
	if err != nil {
		logs.Errorf("describe sub accounts failed, err: %v, account: %s, rid: %s",
			err, req.AccountID, cts.Kit.Rid)
		return nil, err
	}

	return result, nil
}

// TCloudDescribeSafeAuthFlagColl get tcloud sub-account safe auth flag settings (CAM DescribeSafeAuthFlagColl).
func (svc *service) TCloudDescribeSafeAuthFlagColl(cts *rest.Contexts) (interface{}, error) {
	req := new(hssubaccount.TCloudDescribeSafeAuthFlagCollReq)
	if err := cts.DecodeInto(req); err != nil {
		return nil, errf.NewFromErr(errf.DecodeRequestFailed, err)
	}
	if err := req.Validate(); err != nil {
		return nil, errf.NewFromErr(errf.InvalidParameter, err)
	}

	client, err := svc.ad.TCloud(cts.Kit, req.AccountID)
	if err != nil {
		return nil, err
	}

	result, err := client.DescribeSafeAuthFlagColl(cts.Kit, &typeaccount.DescribeSafeAuthFlagCollOption{
		SubUins: req.SubUins,
	})
	if err != nil {
		return nil, err
	}

	return result, nil
}

// TCloudSetMfaFlag set tcloud sub-account login protection and sensitive operation protection (CAM SetMfaFlag).
func (svc *service) TCloudSetMfaFlag(cts *rest.Contexts) (interface{}, error) {
	req := new(hssubaccount.TCloudSetMfaFlagReq)
	if err := cts.DecodeInto(req); err != nil {
		return nil, errf.NewFromErr(errf.DecodeRequestFailed, err)
	}
	if err := req.Validate(); err != nil {
		return nil, errf.NewFromErr(errf.InvalidParameter, err)
	}

	client, err := svc.ad.TCloud(cts.Kit, req.AccountID)
	if err != nil {
		return nil, err
	}

	if err = client.SetMfaFlag(cts.Kit, &typeaccount.SetMfaFlagOption{
		OpUin:      req.OpUin,
		LoginFlag:  req.LoginFlag,
		ActionFlag: req.ActionFlag,
	}); err != nil {
		return nil, err
	}

	return nil, nil
}
