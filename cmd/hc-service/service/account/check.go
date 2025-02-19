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
	"hcm/pkg/api/core/cloud"
	proto "hcm/pkg/api/hc-service/account"
	"hcm/pkg/criteria/errf"
	"hcm/pkg/rest"
)

// TCloudAccountCheck 根据传入秘钥去云上获取数据，并和传入其他数据对比，要求和云上获取数据一致
func (svc *service) TCloudAccountCheck(cts *rest.Contexts) (interface{}, error) {
	req := new(proto.TCloudAccountCheckReq)
	if err := cts.DecodeInto(req); err != nil {
		return nil, errf.NewFromErr(errf.DecodeRequestFailed, err)
	}
	if err := req.Validate(); err != nil {
		return nil, errf.NewFromErr(errf.InvalidParameter, err)
	}

	client, err := svc.ad.Adaptor().TCloud(
		&types.BaseSecret{
			CloudSecretID:  req.CloudSecretID,
			CloudSecretKey: req.CloudSecretKey,
		})
	if err != nil {
		return nil, err
	}

	infoBySecret, err := client.GetAccountInfoBySecret(cts.Kit)
	if err != nil {
		return nil, err
	}
	// check if cloud account info matches the hcm account detail.
	if infoBySecret.CloudSubAccountID != req.CloudSubAccountID {
		return nil, errf.New(errf.InvalidParameter,
			"CloudSubAccountID does not match the account to which the secret belongs")
	}

	if infoBySecret.CloudMainAccountID != req.CloudMainAccountID {
		return nil, errf.New(errf.InvalidParameter,
			"CloudMainAccountID does not match the account to which the secret belongs")
	}

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

	client, err := svc.ad.Adaptor().Aws(
		&types.BaseSecret{
			CloudSecretID:  req.CloudSecretID,
			CloudSecretKey: req.CloudSecretKey,
		}, req.CloudAccountID, req.Site)
	if err != nil {
		return nil, err
	}

	infoBySecret, err := client.GetAccountInfoBySecret(cts.Kit)
	if err != nil {
		return nil, err
	}

	if infoBySecret.CloudIamUsername != req.CloudIamUsername {
		return nil, errf.New(errf.InvalidParameter,
			"CloudIamUsername does not match the account to which the secret belongs")
	}
	if infoBySecret.CloudAccountID != req.CloudAccountID {
		return nil, errf.New(errf.InvalidParameter,
			"CloudAccountID does not match the account to which the secret belongs")
	}
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

	client, err := svc.ad.Adaptor().HuaWei(
		&types.BaseSecret{
			CloudSecretID:  req.CloudSecretID,
			CloudSecretKey: req.CloudSecretKey,
		})
	if err != nil {
		return nil, err
	}

	infoBySecret, err := client.GetAccountInfoBySecret(cts.Kit, req.CloudSecretID)
	if err != nil {
		return nil, err
	}

	// 强校验，要求和用户确认时一样
	if infoBySecret.CloudIamUsername != req.CloudIamUsername {
		return nil, errf.New(errf.InvalidParameter,
			"CloudIamUsername does not match the account to which the secret belongs")
	}
	if infoBySecret.CloudIamUserID != req.CloudIamUserID {
		return nil, errf.New(errf.InvalidParameter,
			"CloudIamUserID does not match the account to which the secret belongs")
	}
	if infoBySecret.CloudSubAccountName != req.CloudSubAccountName {
		return nil, errf.New(errf.InvalidParameter,
			"CloudSubAccountName does not match the account to which the secret belongs")
	}

	if infoBySecret.CloudSubAccountID != req.CloudSubAccountID {
		return nil, errf.New(errf.InvalidParameter,
			"CloudSubAccountID does not match the account to which the secret belongs")
	}
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

	client, err := svc.ad.Adaptor().Gcp(
		&types.GcpCredential{
			CloudProjectID: req.CloudProjectID,
			Json:           []byte(req.CloudServiceSecretKey),
		})
	if err != nil {
		return nil, err
	}

	infoBySecret, err := client.GetAccountInfoBySecret(cts.Kit, req.CloudServiceSecretKey)
	if err != nil {
		return nil, err
	}

	var projectInfo cloud.GcpProjectInfo
	for _, info := range infoBySecret.CloudProjectInfos {
		if info.CloudProjectID == req.CloudProjectID {
			projectInfo = info
			break
		}
	}
	if projectInfo.CloudProjectID != req.CloudProjectID {
		return nil, errf.New(errf.InvalidParameter,
			"CloudProjectID does not match the account to which the secret belongs")
	}
	if projectInfo.CloudProjectName != req.CloudProjectName {
		return nil, errf.New(errf.InvalidParameter,
			"CloudProjectName does not match the account to which the secret belongs")
	}
	if projectInfo.CloudServiceAccountID != req.CloudServiceAccountID {
		return nil, errf.New(errf.InvalidParameter,
			"CloudServiceAccountID does not match the account to which the secret belongs")
	}
	if projectInfo.CloudServiceAccountName != req.CloudServiceAccountName {
		return nil, errf.New(errf.InvalidParameter,
			"CloudServiceAccountName does not match the account to which the secret belongs")
	}
	if projectInfo.CloudServiceSecretID != req.CloudServiceSecretID {
		return nil, errf.New(errf.InvalidParameter,
			"CloudServiceSecretID does not match the account to which the secret belongs")
	}
	return nil, nil
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

	client, err := svc.ad.Adaptor().Azure(
		&types.AzureCredential{
			CloudTenantID:        req.CloudTenantID,
			CloudSubscriptionID:  req.CloudSubscriptionID,
			CloudApplicationID:   req.CloudApplicationID,
			CloudClientSecretKey: req.CloudClientSecretKey,
		})
	if err != nil {
		return nil, err
	}

	infoBySecret, err := client.GetAccountInfoBySecret(cts.Kit)
	if err != nil {
		return nil, err
	}

	var curSubscription cloud.AzureSubscriptionInfo
	for _, subscription := range infoBySecret.SubscriptionInfos {
		if subscription.CloudSubscriptionID == req.CloudSubscriptionID {
			curSubscription = subscription
			break
		}
	}
	if curSubscription.CloudSubscriptionID != req.CloudSubscriptionID {
		return nil, errf.New(errf.InvalidParameter,
			"CloudSubscriptionID does not match the account to which the secret belongs")
	}
	if curSubscription.CloudSubscriptionName != req.CloudSubscriptionName {
		return nil, errf.New(errf.InvalidParameter,
			"CloudSubscriptionName does not match the account to which the secret belongs")
	}

	var curApplication cloud.AzureApplicationInfo
	for _, application := range infoBySecret.ApplicationInfos {
		if application.CloudApplicationID == req.CloudApplicationID {
			curApplication = application
			break
		}
	}
	if curApplication.CloudApplicationName != req.CloudApplicationName {
		return nil, errf.New(errf.InvalidParameter,
			"CloudApplicationName does not match the account to which the secret belongs")
	}

	return nil, err
}
