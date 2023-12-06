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
	"strings"

	"hcm/pkg/adaptor/types/account"
	"hcm/pkg/api/core/cloud"
	"hcm/pkg/kit"
	"hcm/pkg/logs"
	"hcm/pkg/tools/converter"
)

// CountAccount count account.
// reference: https://learn.microsoft.com/en-us/graph/api/user-list?view=graph-rest-1.0&tabs=http
func (az *Azure) CountAccount(kt *kit.Kit) (int32, error) {

	graphClient, err := az.clientSet.graphServiceClient()
	if err != nil {
		logs.Errorf("new graph service client failed, err: %v, rid: %s", err, kt.Rid)
		return 0, fmt.Errorf("new graph service client failed, err: %v", err)
	}

	resp, err := graphClient.Users().Get(kt.Ctx, nil)
	if err != nil {
		logs.Errorf("list users failed, err: %v, rid: %s", err, kt.Rid)
		return 0, err
	}

	return int32(len(resp.GetValue())), nil
}

// ListAccount list account.
// reference: https://learn.microsoft.com/en-us/graph/api/user-list?view=graph-rest-1.0&tabs=http
// 接口需要特殊权限，文档：https://learn.microsoft.com/en-us/graph/auth-v2-service?tabs=http
func (az *Azure) ListAccount(kt *kit.Kit) ([]account.AzureAccount, error) {

	graphClient, err := az.clientSet.graphServiceClient()
	if err != nil {
		logs.Errorf("new graph service client failed, err: %v, rid: %s", err, kt.Rid)
		return nil, fmt.Errorf("new graph service client failed, err: %v", err)
	}

	resp, err := graphClient.Users().Get(kt.Ctx, nil)
	if err != nil {
		extracted := extractGraphError(err)
		logs.Errorf("list users failed, err: %v, rid: %s", extracted, kt.Rid)
		return nil, extracted
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

// GetAccountInfoBySecret 根据秘钥获取账号信息
// 1. https://learn.microsoft.com/en-us/rest/api/resources/subscriptions/list
// 2. https://learn.microsoft.com/en-us/graph/api/application-list
func (az *Azure) GetAccountInfoBySecret(kt *kit.Kit) (*cloud.AzureInfoBySecret, error) {
	graphClient, err := az.clientSet.graphServiceClient()
	if err != nil {
		return nil, err
	}

	subClient, err := az.clientSet.subscriptionClient()
	if err != nil {
		return nil, err
	}
	azInfo := new(cloud.AzureInfoBySecret)

	// 1. 获取该账号可以访问的订阅，要求订阅数量刚好一个
	// https://learn.microsoft.com/en-us/rest/api/resources/subscriptions/list
	pager := subClient.NewListPager(nil)
	if !pager.More() {
		return nil, fmt.Errorf("no subscription found")
	}
	subscriptionListResp, err := pager.NextPage(kt.Ctx)
	if err != nil {
		return nil, err
	}

	subscriptions := subscriptionListResp.Value
	if len(subscriptions) == 0 {
		return nil, fmt.Errorf("no subscription found")
	}
	if len(subscriptions) > 1 {
		subs := make([]string, len(subscriptions))
		for i, sub := range subscriptions {
			subs[i] = "(" + converter.PtrToVal(sub.SubscriptionID) + ")" + converter.PtrToVal(sub.DisplayName)
		}
		return nil, fmt.Errorf("more than one subscription found: " + strings.Join(subs, ","))
	}

	azInfo.CloudSubscriptionName = converter.PtrToVal(subscriptions[0].DisplayName)
	azInfo.CloudSubscriptionID = converter.PtrToVal(subscriptions[0].SubscriptionID)

	// 2. 获取应用信息 https://learn.microsoft.com/en-us/graph/api/application-list
	resp, err := graphClient.Applications().Get(kt.Ctx, nil)
	if err != nil {
		extracted := extractGraphError(err)
		logs.Errorf("fail to get azure applications, err: %v, rid: %s", extracted, kt.Rid)
		return nil, extracted
	}

	for _, one := range resp.GetValue() {
		//	 过滤id
		if converter.PtrToVal(one.GetAppId()) == az.clientSet.credential.CloudApplicationID {
			azInfo.CloudApplicationName = converter.PtrToVal(one.GetDisplayName())
			break
		}

	}
	// 没有拿到应用id的情况
	if len(azInfo.CloudApplicationName) == 0 {
		return nil, fmt.Errorf("failed to get application name")
	}

	return azInfo, nil
}
