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

package cloudadaptor

import (
	"errors"
	"fmt"

	"hcm/pkg/adaptor/types"
	dataservice "hcm/pkg/client/data-service"
	"hcm/pkg/criteria/enumor"
	"hcm/pkg/kit"
)

// SecretClient used to get secret by account id from data-service.
type SecretClient struct {
	data *dataservice.Client
}

// NewSecretClient new secret client that used to get secret info from data service.
func NewSecretClient(dataCli *dataservice.Client) *SecretClient {
	return &SecretClient{data: dataCli}
}

// TCloudSecret get tcloud secret and validate secret.
func (cli *SecretClient) TCloudSecret(kt *kit.Kit, accountID string) (*types.BaseSecret, error) {
	account, err := cli.data.TCloud.Account.Get(kt.Ctx, kt.Header(), accountID)
	if err != nil {
		return nil, fmt.Errorf("get tcloud account failed, err: %v", err)
	}

	if account.Extension == nil {
		return nil, errors.New("tcloud account extension is nil")
	}

	secret := &types.BaseSecret{
		CloudSecretID:  account.Extension.CloudSecretID,
		CloudSecretKey: account.Extension.CloudSecretKey,
	}

	if err := secret.Validate(); err != nil {
		return nil, err
	}

	return secret, nil
}

// AwsSecret get aws secret and validate secret.
func (cli *SecretClient) AwsSecret(kt *kit.Kit, accountID string) (
	*types.BaseSecret, string, enumor.AccountSiteType, error) {

	account, err := cli.data.Aws.Account.Get(kt.Ctx, kt.Header(), accountID)
	if err != nil {
		return nil, "", "", fmt.Errorf("get aws account failed, err: %v", err)
	}

	if account.Extension == nil {
		return nil, "", "", errors.New("aws account extension is nil")
	}

	secret := &types.BaseSecret{
		CloudSecretID:  account.Extension.CloudSecretID,
		CloudSecretKey: account.Extension.CloudSecretKey,
	}
	if err := secret.Validate(); err != nil {
		return nil, "", "", err
	}

	return secret, account.Extension.CloudAccountID, account.Site, nil
}

// HuaWeiSecret get huawei secret and validate secret.
func (cli *SecretClient) HuaWeiSecret(kt *kit.Kit, accountID string) (*types.BaseSecret, error) {
	account, err := cli.data.HuaWei.Account.Get(kt.Ctx, kt.Header(), accountID)
	if err != nil {
		return nil, fmt.Errorf("get huawei account failed, err: %v", err)
	}

	if account.Extension == nil {
		return nil, errors.New("huawei account extension is nil")
	}

	secret := &types.BaseSecret{
		CloudSecretID:  account.Extension.CloudSecretID,
		CloudSecretKey: account.Extension.CloudSecretKey,
	}

	if err := secret.Validate(); err != nil {
		return nil, err
	}

	return secret, nil
}

// AzureCredential get azure credential and validate credential.
func (cli *SecretClient) AzureCredential(kt *kit.Kit, accountID string) (*types.AzureCredential, error) {
	account, err := cli.data.Azure.Account.Get(kt.Ctx, kt.Header(), accountID)
	if err != nil {
		return nil, fmt.Errorf("get azure account failed, err: %v", err)
	}

	if account.Extension == nil {
		return nil, errors.New("azure account extension is nil")
	}

	cred := &types.AzureCredential{
		CloudTenantID:        account.Extension.CloudTenantID,
		CloudSubscriptionID:  account.Extension.CloudSubscriptionID,
		CloudApplicationID:   account.Extension.CloudApplicationID,
		CloudClientSecretKey: account.Extension.CloudClientSecretKey,
	}

	if err := cred.Validate(); err != nil {
		return nil, err
	}

	return cred, nil
}

// GcpCredential get gcp credential and validate credential.
func (cli *SecretClient) GcpCredential(kt *kit.Kit, accountID string) (*types.GcpCredential, error) {
	account, err := cli.data.Gcp.Account.Get(kt.Ctx, kt.Header(), accountID)
	if err != nil {
		return nil, fmt.Errorf("get gcp account failed, err: %v", err)
	}

	if account.Extension == nil {
		return nil, errors.New("gcp account extension is nil")
	}

	cred := &types.GcpCredential{
		CloudProjectID: account.Extension.CloudProjectID,
		Json:           []byte(account.Extension.CloudServiceSecretKey),
	}

	if err := cred.Validate(); err != nil {
		return nil, err
	}

	return cred, nil
}

// GcpRegisterCredential get gcp register credential and validate credential.
func (cli *SecretClient) GcpRegisterCredential(kt *kit.Kit, accountID string) (*types.GcpCredential, error) {
	account, err := cli.data.Gcp.Account.Get(kt.Ctx, kt.Header(), accountID)
	if err != nil {
		return nil, fmt.Errorf("get gcp register account failed, err: %v", err)
	}

	if account.Extension == nil {
		return nil, errors.New("gcp account extension is nil")
	}

	cred := &types.GcpCredential{
		CloudProjectID: account.Extension.CloudProjectID,
		Json:           []byte(account.Extension.CloudServiceSecretKey),
	}

	if err = cred.Validate(); err != nil {
		return nil, err
	}

	return cred, nil
}

// AwsRootSecret get aws secret and validate secret.
func (cli *SecretClient) AwsRootSecret(kt *kit.Kit, accountID string) (*types.BaseSecret, string,
	enumor.RootAccountSiteType, error) {

	account, err := cli.data.Aws.RootAccount.Get(kt, accountID)
	if err != nil {
		return nil, "", "", fmt.Errorf("get aws root account failed, err: %v", err)
	}

	if account.Extension == nil {
		return nil, "", "", errors.New("aws root account extension is nil")
	}

	secret := &types.BaseSecret{
		CloudSecretID:  account.Extension.CloudSecretID,
		CloudSecretKey: account.Extension.CloudSecretKey,
	}

	if err := secret.Validate(); err != nil {
		return nil, "", "", err
	}

	return secret, account.Extension.CloudAccountID, account.Site, nil
}

// GcpRootCredential get gcp credential and validate credential.
func (cli *SecretClient) GcpRootCredential(kt *kit.Kit, accountID string) (*types.GcpCredential, error) {
	account, err := cli.data.Gcp.RootAccount.Get(kt, accountID)
	if err != nil {
		return nil, fmt.Errorf("get gcp root account failed, err: %v", err)
	}

	if account.Extension == nil {
		return nil, errors.New("gcp root account extension is nil")
	}

	cred := &types.GcpCredential{
		CloudProjectID: account.Extension.CloudProjectID,
		Json:           []byte(account.Extension.CloudServiceSecretKey),
	}

	if err := cred.Validate(); err != nil {
		return nil, err
	}

	return cred, nil
}

// HuaWeiRootSecret get huawei secret and validate secret.
func (cli *SecretClient) HuaWeiRootSecret(kt *kit.Kit, accountID string) (*types.BaseSecret, error) {
	account, err := cli.data.HuaWei.RootAccount.Get(kt, accountID)
	if err != nil {
		return nil, fmt.Errorf("get huawei root account failed, err: %v", err)
	}

	if account.Extension == nil {
		return nil, errors.New("huawei root account extension is nil")
	}

	secret := &types.BaseSecret{
		CloudSecretID:  account.Extension.CloudSecretID,
		CloudSecretKey: account.Extension.CloudSecretKey,
	}

	if err := secret.Validate(); err != nil {
		return nil, err
	}

	return secret, nil
}

// AzureRootCredential get azure credential and validate credential.
func (cli *SecretClient) AzureRootCredential(kt *kit.Kit, accountID string) (*types.AzureCredential, error) {
	account, err := cli.data.Azure.RootAccount.Get(kt, accountID)
	if err != nil {
		return nil, fmt.Errorf("get azure root account failed, err: %v", err)
	}

	if account.Extension == nil {
		return nil, errors.New("azure root account extension is nil")
	}

	cred := &types.AzureCredential{
		CloudTenantID:        account.Extension.CloudTenantID,
		CloudSubscriptionID:  account.Extension.CloudSubscriptionID,
		CloudApplicationID:   account.Extension.CloudApplicationID,
		CloudClientSecretKey: account.Extension.CloudClientSecretKey,
	}

	if err := cred.Validate(); err != nil {
		return nil, err
	}

	return cred, nil
}
