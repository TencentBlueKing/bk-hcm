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
	cvt "hcm/pkg/tools/converter"

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

	accountInput := &organizations.CreateAccountInput{
		AccountName:            cvt.ValToPtr(req.CloudAccountName),
		Email:                  cvt.ValToPtr(req.Email),
		IamUserAccessToBilling: cvt.ValToPtr("DENY"),
		// 不设置RoleName，使用默认的角色OrganizationAccountAccessRole
	}
	// create account
	output, err := client.CreateAccount(accountInput)
	if err != nil {
		logs.Errorf("create aws account error, err: %v, req: %+v, rid: %s", err, req, kt.Rid)
		return nil, err
	}

	handler := &createMainAccountPollingHandler{}
	resPoller := poller.Poller[*Aws, *organizations.CreateAccountStatus, organizations.CreateAccountStatus]{Handler: handler}
	result, err := resPoller.PollUntilDone(a, kt, []*string{output.CreateAccountStatus.Id},
		types.NewCreateMainAccountPollerOption())
	if err != nil {
		logs.Errorf("fail to poll aws account create state, err: %v, status id: %+v, rid: %s",
			err, output.CreateAccountStatus.Id, kt.Rid)
		return nil, err
	}
	switch state := cvt.PtrToVal(result.State); state {
	case AccountCreateStatusFailed:
		logs.Errorf("create aws account failed, status: %s, rid: %s", result.String(), kt.Rid)
		if cvt.PtrToVal(result.FailureReason) == AccountCreateErrorMessageAccountAlreadyExists {
			return nil, fmt.Errorf("create account failed, reason: email already exists")
		}
		return nil, fmt.Errorf("create aws account failed, state: %s, reason: %s",
			cvt.PtrToVal(result.State), cvt.PtrToVal(result.FailureReason))
	case AccountCreateStatusSucceeded:
		return &proto.CreateAwsMainAccountResp{
			AccountName: cvt.PtrToVal(result.AccountName),
			AccountID:   cvt.PtrToVal(result.AccountId),
		}, nil
	default:
		logs.Errorf("create aws account failed, unknown status state: %s, status: %s, rid: %s ",
			state, result.String(), kt.Rid)
		return nil, fmt.Errorf("create aws account failed, unknown state: %s, rid: %s", state, kt.Rid)
	}

}

type createMainAccountPollingHandler struct {
}

// Done ...
func (h *createMainAccountPollingHandler) Done(status *organizations.CreateAccountStatus) (
	done bool, ret *organizations.CreateAccountStatus) {

	// 只有还在进行中需要继续轮询
	if cvt.PtrToVal(status.State) == AccountCreateStatusInProgress {
		return false, status
	}
	// 其他情况均为已完成，交由上层处理
	return true, status
}

// Poll ...
func (h *createMainAccountPollingHandler) Poll(client *Aws, kt *kit.Kit, reqIds []*string) (
	*organizations.CreateAccountStatus, error) {

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
		logs.Errorf("describe aws account create status failed, err: %v, result: %+v, rid: %s",
			err, result, kt.Rid)
		return nil, err
	}

	// 只有获取失败，或者状态未知返回error
	switch cvt.PtrToVal(result.CreateAccountStatus.State) {
	case AccountCreateStatusSucceeded, AccountCreateStatusInProgress, AccountCreateStatusFailed:
		return result.CreateAccountStatus, nil
	default:
		logs.Errorf("create aws account got unknown status: %s, rid: %s", result.CreateAccountStatus.String(), kt.Rid)
		return nil, fmt.Errorf("create  aws account unknown progress, state: %s",
			cvt.PtrToVal(result.CreateAccountStatus.State))
	}
}
