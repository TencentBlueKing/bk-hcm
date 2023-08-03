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

	"hcm/pkg/adaptor/types/account"
	"hcm/pkg/kit"
	"hcm/pkg/logs"
)

// ListAccount list account.
// reference: https://learn.microsoft.com/en-us/graph/api/user-list?view=graph-rest-1.0&tabs=http
// 接口需要特殊权限，文档：https://learn.microsoft.com/en-us/graph/api/user-list?view=graph-rest-1.0&tabs=http
func (az *Azure) ListAccount(kt *kit.Kit) ([]account.AzureAccount, error) {

	graphClient, err := az.clientSet.graphServiceClient()
	if err != nil {
		logs.Errorf("new graph service client failed, err: %v, rid: %s", err, kt.Rid)
		return nil, fmt.Errorf("new graph service client failed, err: %v", err)
	}

	resp, err := graphClient.Users().Get(kt.Ctx, nil)
	if err != nil {
		logs.Errorf("list users failed, err: %v, rid: %s", err, kt.Rid)
		return nil, err
	}

	users := resp.GetValue()
	list := make([]account.AzureAccount, 0, len(users))
	for _, one := range users {
		list = append(list, account.AzureAccount{
			DisplayNameName:   one.GetDisplayName(),
			GivenName:         one.GetGivenName(),
			SurName:           one.GetSurname(),
			UserPrincipalName: one.GetUserPrincipalName(),
			ID:                one.GetId(),
		})
	}

	return list, nil
}

// AccountCheck ...
// 接口参考 "https://pkg.go.dev/github.com/Azure/azure-sdk-for-go/services/preview/subscription/mgmt/
// 2018-03-01-preview/subscription#SubscriptionsClient.Get"
func (az *Azure) AccountCheck(kt *kit.Kit) error {
	client, err := az.clientSet.subscriptionClient()
	if err != nil {
		return fmt.Errorf("init azure client failed, err: %v", err)
	}

	cloudSubscriptionID := az.clientSet.credential.CloudSubscriptionID
	_, err = client.Get(kt.Ctx, cloudSubscriptionID, nil)
	if err != nil {
		logs.Errorf(
			"gets details about subscription(%s), err: %v, rid: %s",
			cloudSubscriptionID,
			err,
			kt.Rid,
		)
		return err
	}

	return nil
}
