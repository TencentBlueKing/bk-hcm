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
	"hcm/pkg/adaptor/types"
	proto "hcm/pkg/api/hc-service"
	"hcm/pkg/criteria/errf"
	"hcm/pkg/rest"
)

// TCloudAccountCheck authentication information and permissions.
func (svc *service) TCloudAccountCheck(cts *rest.Contexts) (interface{}, error) {
	req := new(proto.TCloudAccountCheckReq)
	if err := cts.DecodeInto(req); err != nil {
		return nil, errf.NewFromErr(errf.DecodeRequestFailed, err)
	}
	if err := req.Validate(); err != nil {
		return nil, errf.NewFromErr(errf.InvalidParameter, err)
	}

	client, err := svc.ad.Adaptor().TCloud(&types.BaseSecret{CloudSecretID: req.CloudSecretID,
		CloudSecretKey: req.CloudSecretKey})
	if err != nil {
		return nil, err
	}

	err = client.AccountCheck(
		cts.Kit,
		&types.TCloudAccountInfo{CloudMainAccountID: req.CloudMainAccountID, CloudSubAccountID: req.CloudSubAccountID},
	)

	return nil, err
}

// AwsAccountCheck authentication information and permissions.
func (svc *service) AwsAccountCheck(cts *rest.Contexts) (interface{}, error) {
	req := new(proto.AwsAccountCheckReq)
	if err := cts.DecodeInto(req); err != nil {
		return nil, errf.NewFromErr(errf.DecodeRequestFailed, err)
	}

	if err := req.Validate(); err != nil {
		return nil, errf.NewFromErr(errf.InvalidParameter, err)
	}

	client, err := svc.ad.Adaptor().Aws(&types.BaseSecret{CloudSecretID: req.CloudSecretID,
		CloudSecretKey: req.CloudSecretKey}, req.CloudAccountID)
	if err != nil {
		return nil, err
	}

	err = client.AccountCheck(
		cts.Kit,
		&types.AwsAccountInfo{CloudAccountID: req.CloudAccountID, CloudIamUsername: req.CloudIamUsername},
	)

	return nil, err
}

// HuaWeiAccountCheck authentication information and permissions.
func (svc *service) HuaWeiAccountCheck(cts *rest.Contexts) (interface{}, error) {
	req := new(proto.HuaWeiAccountCheckReq)
	if err := cts.DecodeInto(req); err != nil {
		return nil, errf.NewFromErr(errf.DecodeRequestFailed, err)
	}
	if err := req.Validate(); err != nil {
		return nil, errf.NewFromErr(errf.InvalidParameter, err)
	}

	client, err := svc.ad.Adaptor().HuaWei(&types.BaseSecret{CloudSecretID: req.CloudSecretID,
		CloudSecretKey: req.CloudSecretKey})
	if err != nil {
		return nil, err
	}

	err = client.AccountCheck(cts.Kit, &types.HuaWeiAccountInfo{
		CloudMainAccountName: req.CloudMainAccountName,
		CloudSubAccountID:    req.CloudSubAccountID,
		CloudSubAccountName:  req.CloudSubAccountName,
		CloudIamUserID:       req.CloudIamUserID,
		CloudIamUsername:     req.CloudIamUsername,
	})
	return nil, err
}

// GcpAccountCheck ...
func (svc *service) GcpAccountCheck(cts *rest.Contexts) (interface{}, error) {
	req := new(proto.GcpAccountCheckReq)
	if err := cts.DecodeInto(req); err != nil {
		return nil, errf.New(errf.DecodeRequestFailed, err.Error())
	}
	if err := req.Validate(); err != nil {
		return nil, errf.Newf(errf.InvalidParameter, err.Error())
	}

	client, err := svc.ad.Adaptor().Gcp(&types.GcpCredential{CloudProjectID: req.CloudProjectID,
		Json: []byte(req.CloudServiceSecretKey)})
	if err != nil {
		return nil, err
	}

	err = client.AccountCheck(cts.Kit)

	return nil, err
}

// AzureAccountCheck ...
func (svc *service) AzureAccountCheck(cts *rest.Contexts) (interface{}, error) {
	req := new(proto.AzureAccountCheckReq)
	if err := cts.DecodeInto(req); err != nil {
		return nil, errf.New(errf.DecodeRequestFailed, err.Error())
	}
	if err := req.Validate(); err != nil {
		return nil, errf.Newf(errf.InvalidParameter, err.Error())
	}

	client, err := svc.ad.Adaptor().Azure(&types.AzureCredential{
		CloudTenantID: req.CloudTenantID, CloudSubscriptionID: req.CloudSubscriptionID,
		CloudApplicationID: req.CloudApplicationID, CloudClientSecretKey: req.CloudClientSecretKey,
	})
	if err != nil {
		return nil, err
	}

	err = client.AccountCheck(cts.Kit)

	return nil, err
}
