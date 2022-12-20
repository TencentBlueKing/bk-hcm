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

package azure

import (
	"fmt"

	"hcm/pkg/adaptor/types"

	"github.com/Azure/azure-sdk-for-go/sdk/azidentity"
	"github.com/Azure/azure-sdk-for-go/sdk/resourcemanager/subscription/armsubscription"
)

type clientSet struct {
	credential *types.AzureCredential
}

func newClientSet(credential *types.AzureCredential) *clientSet {
	return &clientSet{credential}
}

func (c *clientSet) subscriptionClient() (*armsubscription.SubscriptionsClient, error) {
	credential, err := azidentity.NewClientSecretCredential(
		c.credential.CloudTenantID,
		c.credential.CloudClientID,
		c.credential.CloudClientSecret, nil)
	if err != nil {
		return nil, fmt.Errorf("init azure credential failed, err: %v", err)
	}

	client, err := armsubscription.NewSubscriptionsClient(credential, nil)
	if err != nil {
		return nil, fmt.Errorf("init azure vpn client failed, err: %v", err)
	}

	return client, nil
}
