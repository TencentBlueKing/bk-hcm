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
	dataprotocloud "hcm/pkg/api/data-service/cloud"
	"hcm/pkg/criteria/enumor"
)

// Deliver 执行资源交付
func (a *ApplicationOfAddAccount) Deliver() (enumor.ApplicationStatus, map[string]interface{}, error) {
	// 执行创建账号
	var (
		err       error
		accountID string
	)
	switch a.req.Vendor {
	case enumor.TCloud:
		accountID, err = a.createForTCloud()
	case enumor.Aws:
		accountID, err = a.createForAws()
	case enumor.HuaWei:
		accountID, err = a.createForHuaWei()
	case enumor.Gcp:
		accountID, err = a.createForGcp()
	case enumor.Azure:
		accountID, err = a.createForAzure()
	}
	// 交付失败
	if err != nil {
		return enumor.DeliverError, map[string]interface{}{"error": err.Error()}, err
	}

	// 交付成功，记录交付的账号ID
	return enumor.Completed, map[string]interface{}{"account_id": accountID}, nil
}

func (a *ApplicationOfAddAccount) createForTCloud() (string, error) {
	result, err := a.Client.DataService().TCloud.Account.Create(
		a.Cts.Kit.Ctx,
		a.Cts.Kit.Header(),
		&dataprotocloud.AccountCreateReq[dataprotocloud.TCloudAccountExtensionCreateReq]{
			Name:     a.req.Name,
			Managers: a.req.Managers,
			Type:     a.req.Type,
			Site:     a.req.Site,
			Memo:     a.req.Memo,
			BkBizIDs: a.req.BkBizIDs,
			Extension: &dataprotocloud.TCloudAccountExtensionCreateReq{
				CloudMainAccountID: a.req.Extension["cloud_main_account_id"],
				CloudSubAccountID:  a.req.Extension["cloud_sub_account_id"],
				CloudSecretID:      a.req.Extension["cloud_secret_id"],
				CloudSecretKey:     a.req.Extension["cloud_secret_key"],
			},
		},
	)
	if err != nil {
		return "", err
	}
	return result.ID, err
}

func (a *ApplicationOfAddAccount) createForAws() (string, error) {
	result, err := a.Client.DataService().Aws.Account.Create(
		a.Cts.Kit.Ctx,
		a.Cts.Kit.Header(),
		&dataprotocloud.AccountCreateReq[dataprotocloud.AwsAccountExtensionCreateReq]{
			Name:     a.req.Name,
			Managers: a.req.Managers,
			Type:     a.req.Type,
			Site:     a.req.Site,
			Memo:     a.req.Memo,
			BkBizIDs: a.req.BkBizIDs,
			Extension: &dataprotocloud.AwsAccountExtensionCreateReq{
				CloudAccountID:   a.req.Extension["cloud_account_id"],
				CloudIamUsername: a.req.Extension["cloud_iam_username"],
				CloudSecretID:    a.req.Extension["cloud_secret_id"],
				CloudSecretKey:   a.req.Extension["cloud_secret_key"],
			},
		},
	)
	if err != nil {
		return "", err
	}
	return result.ID, err
}

func (a *ApplicationOfAddAccount) createForHuaWei() (string, error) {
	result, err := a.Client.DataService().HuaWei.Account.Create(
		a.Cts.Kit.Ctx,
		a.Cts.Kit.Header(),
		&dataprotocloud.AccountCreateReq[dataprotocloud.HuaWeiAccountExtensionCreateReq]{
			Name:     a.req.Name,
			Managers: a.req.Managers,
			Type:     a.req.Type,
			Site:     a.req.Site,
			Memo:     a.req.Memo,
			BkBizIDs: a.req.BkBizIDs,
			Extension: &dataprotocloud.HuaWeiAccountExtensionCreateReq{
				CloudMainAccountName: a.req.Extension["cloud_main_account_name"],
				CloudSubAccountID:    a.req.Extension["cloud_sub_account_id"],
				CloudSubAccountName:  a.req.Extension["cloud_sub_account_name"],
				CloudSecretID:        a.req.Extension["cloud_secret_id"],
				CloudSecretKey:       a.req.Extension["cloud_secret_key"],
				CloudIamUserID:       a.req.Extension["cloud_iam_user_id"],
				CloudIamUsername:     a.req.Extension["cloud_iam_username"],
			},
		},
	)
	if err != nil {
		return "", err
	}
	return result.ID, err
}

func (a *ApplicationOfAddAccount) createForGcp() (string, error) {
	result, err := a.Client.DataService().Gcp.Account.Create(
		a.Cts.Kit.Ctx,
		a.Cts.Kit.Header(),
		&dataprotocloud.AccountCreateReq[dataprotocloud.GcpAccountExtensionCreateReq]{
			Name:     a.req.Name,
			Managers: a.req.Managers,
			Type:     a.req.Type,
			Site:     a.req.Site,
			Memo:     a.req.Memo,
			BkBizIDs: a.req.BkBizIDs,
			Extension: &dataprotocloud.GcpAccountExtensionCreateReq{
				CloudProjectID:          a.req.Extension["cloud_project_id"],
				CloudProjectName:        a.req.Extension["cloud_project_name"],
				CloudServiceAccountID:   a.req.Extension["cloud_service_account_id"],
				CloudServiceAccountName: a.req.Extension["cloud_service_account_name"],
				CloudServiceSecretID:    a.req.Extension["cloud_service_secret_id"],
				CloudServiceSecretKey:   a.req.Extension["cloud_service_secret_key"],
			},
		},
	)
	if err != nil {
		return "", err
	}
	return result.ID, err
}

func (a *ApplicationOfAddAccount) createForAzure() (string, error) {
	result, err := a.Client.DataService().Azure.Account.Create(
		a.Cts.Kit.Ctx,
		a.Cts.Kit.Header(),
		&dataprotocloud.AccountCreateReq[dataprotocloud.AzureAccountExtensionCreateReq]{
			Name:     a.req.Name,
			Managers: a.req.Managers,
			Type:     a.req.Type,
			Site:     a.req.Site,
			Memo:     a.req.Memo,
			BkBizIDs: a.req.BkBizIDs,
			Extension: &dataprotocloud.AzureAccountExtensionCreateReq{
				CloudTenantID:         a.req.Extension["cloud_tenant_id"],
				CloudSubscriptionID:   a.req.Extension["cloud_subscription_id"],
				CloudSubscriptionName: a.req.Extension["cloud_subscription_name"],
				CloudApplicationID:    a.req.Extension["cloud_application_id"],
				CloudApplicationName:  a.req.Extension["cloud_application_name"],
				CloudClientSecretID:   a.req.Extension["cloud_client_secret_id"],
				CloudClientSecretKey:  a.req.Extension["cloud_client_secret_key"],
			},
		},
	)
	if err != nil {
		return "", err
	}
	return result.ID, err
}
