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
	"hcm/pkg/adaptor"
	"hcm/pkg/adaptor/aws"
	"hcm/pkg/adaptor/azure"
	"hcm/pkg/adaptor/gcp"
	"hcm/pkg/adaptor/huawei"
	"hcm/pkg/adaptor/tcloud"
	dataservice "hcm/pkg/client/data-service"
	"hcm/pkg/criteria/enumor"
	"hcm/pkg/kit"
)

// NewCloudAdaptorClient new cloud adaptor client.
func NewCloudAdaptorClient(dataCli *dataservice.Client) *CloudAdaptorClient {
	return &CloudAdaptorClient{
		adaptor:   adaptor.New(),
		secretCli: NewSecretClient(dataCli),
	}
}

// CloudAdaptorClient define cloud adaptor client used to request cloud api.
type CloudAdaptorClient struct {
	adaptor   *adaptor.Adaptor
	secretCli *SecretClient
}

// Adaptor return adaptor.
func (cli *CloudAdaptorClient) Adaptor() *adaptor.Adaptor {
	return cli.adaptor
}

// TCloud return tcloud client.
func (cli *CloudAdaptorClient) TCloud(kt *kit.Kit, accountID string) (tcloud.TCloud, error) {
	secret, err := cli.secretCli.TCloudSecret(kt, accountID)
	if err != nil {
		return nil, err
	}

	client, err := cli.adaptor.TCloud(secret)
	if err != nil {
		return nil, err
	}
	client.SetRateLimitRetryWithRandomInterval(kt.RequestSource == enumor.AsynchronousTasks)

	return client, nil
}

// Aws return aws client.
func (cli *CloudAdaptorClient) Aws(kt *kit.Kit, accountID string) (*aws.Aws, error) {
	secret, cloudAccountID, err := cli.secretCli.AwsSecret(kt, accountID)
	if err != nil {
		return nil, err
	}

	return cli.adaptor.Aws(secret, cloudAccountID)
}

// HuaWei return huawei client.
func (cli *CloudAdaptorClient) HuaWei(kt *kit.Kit, accountID string) (*huawei.HuaWei, error) {
	secret, err := cli.secretCli.HuaWeiSecret(kt, accountID)
	if err != nil {
		return nil, err
	}

	return cli.adaptor.HuaWei(secret)
}

// Gcp return gcp client.
func (cli *CloudAdaptorClient) Gcp(kt *kit.Kit, accountID string) (*gcp.Gcp, error) {
	cred, err := cli.secretCli.GcpCredential(kt, accountID)
	if err != nil {
		return nil, err
	}

	return cli.adaptor.Gcp(cred)
}

// GcpProxy return gcp proxy client.
func (cli *CloudAdaptorClient) GcpProxy(kt *kit.Kit, accountID string) (*gcp.Gcp, error) {
	cred, err := cli.secretCli.GcpRegisterCredential(kt, accountID)
	if err != nil {
		return nil, err
	}

	return cli.adaptor.Gcp(cred)
}

// Azure return azure client.
func (cli *CloudAdaptorClient) Azure(kt *kit.Kit, accountID string) (*azure.Azure, error) {
	cred, err := cli.secretCli.AzureCredential(kt, accountID)
	if err != nil {
		return nil, err
	}

	return cli.adaptor.Azure(cred)
}

// AwsRoot return aws root client.
func (cli *CloudAdaptorClient) AwsRoot(kt *kit.Kit, accountID string) (*aws.Aws, error) {
	secret, cloudAccountID, err := cli.secretCli.AwsRootSecret(kt, accountID)
	if err != nil {
		return nil, err
	}

	return cli.adaptor.Aws(secret, cloudAccountID)
}

// GcpRoot return gcp client.
func (cli *CloudAdaptorClient) GcpRoot(kt *kit.Kit, accountID string) (*gcp.Gcp, error) {
	cred, err := cli.secretCli.GcpRootCredential(kt, accountID)
	if err != nil {
		return nil, err
	}

	return cli.adaptor.Gcp(cred)
}

// HuaWeiRoot return huawei client.
func (cli *CloudAdaptorClient) HuaWeiRoot(kt *kit.Kit, accountID string) (*huawei.HuaWei, error) {
	secret, err := cli.secretCli.HuaWeiRootSecret(kt, accountID)
	if err != nil {
		return nil, err
	}

	return cli.adaptor.HuaWei(secret)
}

// AzureRoot return azure client.
func (cli *CloudAdaptorClient) AzureRoot(kt *kit.Kit, accountID string) (*azure.Azure, error) {
	cred, err := cli.secretCli.AzureRootCredential(kt, accountID)
	if err != nil {
		return nil, err
	}

	return cli.adaptor.Azure(cred)
}
