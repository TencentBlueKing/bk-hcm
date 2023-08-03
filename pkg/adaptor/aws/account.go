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

	"github.com/aws/aws-sdk-go/service/organizations"
	"github.com/aws/aws-sdk-go/service/sts"

	"hcm/pkg/adaptor/types"
	"hcm/pkg/adaptor/types/account"
	"hcm/pkg/criteria/errf"
	"hcm/pkg/kit"
	"hcm/pkg/logs"
	"hcm/pkg/tools/converter"
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

// AccountCheck check account authentication information(account id and iam user name) and permissions.
// GetCallerIdentity: https://docs.aws.amazon.com/STS/latest/APIReference/API_GetCallerIdentity.html
func (a *Aws) AccountCheck(kt *kit.Kit, opt *types.AwsAccountInfo) error {
	if opt == nil {
		return errf.New(errf.InvalidParameter, "account check option is required")
	}

	if err := opt.Validate(); err != nil {
		return err
	}

	client, err := a.clientSet.stsClient()
	if err != nil {
		return fmt.Errorf("init aws client failed, err: %v", err)
	}

	req := new(sts.GetCallerIdentityInput)
	resp, err := client.GetCallerIdentityWithContext(kt.Ctx, req)
	if err != nil {
		logs.Errorf("describe regions failed, err: %v, rid: %s", err, kt.Rid)
		return err
	}

	if resp.Account == nil {
		return errors.New("get caller identity return account is nil")
	}

	// check account info: account id、user name
	if *resp.Account != opt.CloudAccountID {
		return fmt.Errorf("account id does not match the account to which the secret belongs")
	}

	if resp.Arn == nil {
		return errors.New("get caller identity return arn is nil")
	}

	split := strings.Split(*resp.Arn, "/")
	if split[len(split)-1] != opt.CloudIamUsername {
		return fmt.Errorf("iam user name does not match the account to which the secret belongs")
	}

	return nil
}
