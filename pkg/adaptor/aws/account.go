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

package aws

import (
	"errors"
	"fmt"
	"strings"

	"hcm/pkg/adaptor/types/account"
	"hcm/pkg/api/core/cloud"
	"hcm/pkg/kit"
	"hcm/pkg/logs"
	"hcm/pkg/tools/converter"

	"github.com/aws/aws-sdk-go/service/organizations"
	"github.com/aws/aws-sdk-go/service/sts"
)

// ListAccount 查询账号列表，因为账号列表的数量不会很多，且其他云也是全量返回，所以，这里将aws的账号列表进行了全量查询。
// reference: https://docs.amazonaws.cn/organizations/latest/APIReference/API_ListAccounts.html
func (a *Aws) ListAccount(kt *kit.Kit) ([]account.AwsAccount, error) {
	client, err := a.clientSet.organizations()
	if err != nil {
		return nil, err
	}

	list := make([]account.AwsAccount, 0)
	req := new(organizations.ListAccountsInput)
	req.MaxResults = converter.ValToPtr(int64(20))
	for {
		resp, err := client.ListAccountsWithContext(kt.Ctx, req)
		if err != nil {
			logs.Errorf("list accounts failed, err: %v, rid: %s", err, kt.Rid)
			return nil, err
		}

		for _, one := range resp.Accounts {
			list = append(list, account.AwsAccount{
				Arn:             one.Arn,
				Email:           one.Email,
				ID:              one.Id,
				JoinedMethod:    one.JoinedMethod,
				JoinedTimestamp: one.JoinedTimestamp,
				Name:            one.Name,
				Status:          one.Status,
			})
		}

		if resp.NextToken == nil || *resp.NextToken == "" {
			break
		}

		req.NextToken = resp.NextToken
	}

	return list, nil
}

// CountAccount 返回账号下子账号数量，基于 ListAccountsWithContext 接口
// reference: https://docs.amazonaws.cn/organizations/latest/APIReference/API_ListAccounts.html
func (a *Aws) CountAccount(kt *kit.Kit) (int32, error) {
	client, err := a.clientSet.organizations()
	if err != nil {
		return 0, err
	}

	req := new(organizations.ListAccountsInput)
	req.MaxResults = converter.ValToPtr(int64(20))

	total := 0

	for {
		resp, err := client.ListAccountsWithContext(kt.Ctx, req)
		if err != nil {
			logs.Errorf("count accounts failed, err: %v, rid: %s", err, kt.Rid)
			return 0, err
		}
		total += len(resp.Accounts)

		if resp.NextToken != nil {
			req.NextToken = resp.NextToken
			continue
		}
		break
	}

	return int32(total), nil
}

// GetAccountInfoBySecret 根据秘钥获取账号信息
// reference: https://docs.aws.amazon.com/STS/latest/APIReference/API_GetCallerIdentity.html
func (a *Aws) GetAccountInfoBySecret(kt *kit.Kit) (*cloud.AwsInfoBySecret, error) {
	var defaultRegion *string = nil
	// else use nil to indicate sdk default region
	if a.IsChinaSite() {
		defaultRegion = converter.ValToPtr(a.DefaultRegion())
	}
	client, err := a.clientSet.stsClient(defaultRegion)

	if err != nil {
		return nil, fmt.Errorf("init aws client failed, err: %v", err)
	}

	req := new(sts.GetCallerIdentityInput)
	resp, err := client.GetCallerIdentityWithContext(kt.Ctx, req)
	if err != nil {
		logs.Errorf("describe regions failed, err: %v, rid: %s", err, kt.Rid)
		return nil, err
	}

	if resp.Account == nil {
		return nil, errors.New("get caller identity return account is nil")
	}

	if resp.Arn == nil {
		return nil, errors.New("get caller identity return arn is nil")
	}

	// arn最后一部分是用户名
	parts := strings.Split(converter.PtrToVal(resp.Arn), "/")
	return &cloud.AwsInfoBySecret{
		CloudAccountID:   converter.PtrToVal(resp.Account),
		CloudIamUsername: parts[len(parts)-1],
	}, nil
}
