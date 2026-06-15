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
	"hcm/pkg/rest"
)

// TCloudCreateAccessKey create access key for TCloud CAM sub-user.
func (svc *service) TCloudCreateAccessKey(cts *rest.Contexts) (interface{}, error) {
	req := new(hssubaccount.TCloudCreateAccessKeyReq)
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

	result, err := client.CreateAccessKey(cts.Kit, &typeaccount.CreateAccessKeyOption{
		TargetUin:   req.TargetUin,
		Description: req.Description,
	})
	if err != nil {
		return nil, err
	}

	return result, nil
}

// TCloudDeleteAccessKey delete access key for TCloud CAM sub-user.
func (svc *service) TCloudDeleteAccessKey(cts *rest.Contexts) (interface{}, error) {
	req := new(hssubaccount.TCloudDeleteAccessKeyReq)
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

	if err = client.DeleteAccessKey(cts.Kit, &typeaccount.DeleteAccessKeyOption{
		AccessKeyID: req.AccessKeyID,
		TargetUin:   req.TargetUin,
	}); err != nil {
		return nil, err
	}

	return nil, nil
}

// TCloudUpdateAccessKey update access key status (Active/Inactive) for TCloud CAM sub-user.
func (svc *service) TCloudUpdateAccessKey(cts *rest.Contexts) (interface{}, error) {
	req := new(hssubaccount.TCloudUpdateAccessKeyReq)
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

	if err = client.UpdateAccessKey(cts.Kit, &typeaccount.UpdateAccessKeyOption{
		AccessKeyID: req.AccessKeyID,
		Status:      req.Status,
		TargetUin:   req.TargetUin,
	}); err != nil {
		return nil, err
	}

	return nil, nil
}

// TCloudListAccessKeys list access keys for TCloud CAM user.
func (svc *service) TCloudListAccessKeys(cts *rest.Contexts) (interface{}, error) {
	req := new(hssubaccount.TCloudListAccessKeysReq)
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

	result, err := client.ListAccessKeys(cts.Kit, &typeaccount.ListAccessKeysOption{
		TargetUin: req.TargetUin,
	})
	if err != nil {
		return nil, err
	}

	return result, nil
}

// TCloudGetSecurityLastUsed get access key last usage info for TCloud CAM.
func (svc *service) TCloudGetSecurityLastUsed(cts *rest.Contexts) (interface{}, error) {
	req := new(hssubaccount.TCloudGetSecurityLastUsedReq)
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

	result, err := client.GetSecurityLastUsed(cts.Kit, &typeaccount.GetSecurityLastUsedOption{
		SecretIdList: req.SecretIdList,
	})
	if err != nil {
		return nil, err
	}

	return result, nil
}
