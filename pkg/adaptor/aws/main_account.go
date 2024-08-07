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
	// types "hcm/pkg/adaptor/types/main-account"
	"fmt"
	"hcm/pkg/adaptor/poller"
	"hcm/pkg/adaptor/types"
	proto "hcm/pkg/api/hc-service/main-account"
	"hcm/pkg/kit"
	"hcm/pkg/logs"
	"hcm/pkg/logs/glog"
	"hcm/pkg/tools/converter"

	"github.com/aws/aws-sdk-go/service/organizations"
)

const (
	// AccountCreateStatusSucceeded	The account was created and confirmed.
	AccountCreateStatusSucceeded = "SUCCEEDED"
	// AccountCreateStatusFailed	The account could not be created.
	AccountCreateStatusFailed = "FAILED"
	// AccountCreateStatusInProgress	The account is currently being created.
	AccountCreateStatusInProgress = "IN_PROGRESS"

	// AccountCreateErrorMessageAccountAlreadyExists	The account already exists.
	AccountCreateErrorMessageAccountAlreadyExists = "EMAIL_ALREADY_EXISTS"
)

// CreateAccount
// reference: https://docs.aws.amazon.com/organizations/latest/APIReference/API_CreateAccount.html
func (a *Aws) CreateAccount(kt *kit.Kit, req *proto.CreateAwsMainAccountReq) (*proto.CreateAwsMainAccountResp, error) {
	// get aws client
	client, err := a.clientSet.organizations()
	if err != nil {
		logs.Errorf("init aws client failed, err: %v, rid: %s", err, kt.Rid)
		return nil, err
	}

	// create account
	output, err := client.CreateAccount(&organizations.CreateAccountInput{
		AccountName:            &req.CloudAccountName,
		Email:                  &req.Email,
		IamUserAccessToBilling: converter.ValToPtr("DENY"),
		// 不设置RoleName，使用默认的角色OrganizationAccountAccessRole
	})

	if err != nil {
		logs.Errorf("create aws account error, req: %+v, err: %v, rid: %s", req, err, kt.Rid)
		return nil, err
	}

	handler := &createMainAccountPollingHandler{}
	resPoller := poller.Poller[*Aws, *organizations.CreateAccountStatus, organizations.CreateAccountStatus]{Handler: handler}
	result, err := resPoller.PollUntilDone(a, kt, []*string{output.CreateAccountStatus.Id},
		types.NewCreateMainAccountPollerOption())
	if err != nil {
		return nil, err
	}
	if converter.PtrToVal(result.State) != AccountCreateStatusSucceeded {
		return nil, fmt.Errorf("create aws account failed, reason: %s, err: %v, rid: %s", *result.FailureReason, err, kt.Rid)
	}

	return &proto.CreateAwsMainAccountResp{
		AccountName: *output.CreateAccountStatus.AccountName,
		AccountID:   *output.CreateAccountStatus.AccountId,
	}, nil
}

type createMainAccountPollingHandler struct {
}

// Done ...
func (h *createMainAccountPollingHandler) Done(status *organizations.CreateAccountStatus) (bool, *organizations.CreateAccountStatus) {
	// Note: 没有error的情况分两种, 一种是在创建中，创建中不结束Poll，返回false，另一种是创建成功，返回true
	if converter.PtrToVal(status.State) == AccountCreateStatusInProgress {
		return false, status
	}
	return true, status
}

// Poll ...
func (h *createMainAccountPollingHandler) Poll(client *Aws, kt *kit.Kit, reqIds []*string) (*organizations.CreateAccountStatus, error) {
	if len(reqIds) == 0 {
		return nil, fmt.Errorf("operation group id is required")
	}

	reqId := reqIds[0]

	orgClient, err := client.clientSet.organizations()
	if err != nil {
		return nil, err
	}

	result, err := orgClient.DescribeCreateAccountStatus(&organizations.DescribeCreateAccountStatusInput{
		CreateAccountRequestId: reqId,
	})
	if err != nil {
		glog.Errorf("describe account create failed, err: %s, rid: %s", err.Error(), kt.Rid)
		return nil, err
	}

	// Note: 成功或者在创建中，都不返回error。如果失败则返回error
	switch converter.PtrToVal(result.CreateAccountStatus.State) {
	case AccountCreateStatusSucceeded, AccountCreateStatusInProgress:
		return result.CreateAccountStatus, nil
	case AccountCreateStatusFailed:
		if converter.PtrToVal(result.CreateAccountStatus.FailureReason) == AccountCreateErrorMessageAccountAlreadyExists {
			return nil, fmt.Errorf("create account failed, state: %s, reason: email already exists", *result.CreateAccountStatus.State)
		}
		return nil, fmt.Errorf("create account failed, state: %s, reason: %s", *result.CreateAccountStatus.State, *result.CreateAccountStatus.FailureReason)
	default:
		return nil, fmt.Errorf("create account unknown progress, state: %s", *result.CreateAccountStatus.State)
	}
}
