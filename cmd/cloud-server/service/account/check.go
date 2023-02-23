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
	"encoding/json"
	"fmt"

	"hcm/cmd/cloud-server/service/common"
	proto "hcm/pkg/api/cloud-server/account"
	hcproto "hcm/pkg/api/hc-service"
	"hcm/pkg/criteria/enumor"
	"hcm/pkg/criteria/errf"
	"hcm/pkg/iam/meta"
	"hcm/pkg/rest"
)

// Check ...
func (a *accountSvc) Check(cts *rest.Contexts) (interface{}, error) {
	req := new(proto.AccountCheckReq)
	if err := cts.DecodeInto(req); err != nil {
		return nil, errf.NewFromErr(errf.DecodeRequestFailed, err)
	}
	if err := req.Validate(); err != nil {
		return nil, errf.NewFromErr(errf.InvalidParameter, err)
	}

	// 校验用户有该账号的录入权限
	if err := a.checkPermission(cts, meta.Import, ""); err != nil {
		return nil, err
	}

	var err error
	switch req.Vendor {
	case enumor.TCloud:
		_, err = a.parseAndCheckTCloudExtension(cts, req.Type, req.Extension)
	case enumor.Aws:
		_, err = a.parseAndCheckAwsExtension(cts, req.Type, req.Extension)
	case enumor.HuaWei:
		_, err = a.parseAndCheckHuaWeiExtension(cts, req.Type, req.Extension)
	case enumor.Gcp:
		_, err = a.parseAndCheckGcpExtension(cts, req.Type, req.Extension)
	case enumor.Azure:
		_, err = a.parseAndCheckAzureExtension(cts, req.Type, req.Extension)
	default:
		err = fmt.Errorf("no support vendor: %s", req.Vendor)
	}

	return nil, errf.NewFromErr(errf.InvalidParameter, err)
}

func (a *accountSvc) parseAndCheckTCloudExtension(
	cts *rest.Contexts, accountType enumor.AccountType, reqExtension json.RawMessage,
) (*proto.TCloudAccountExtensionCreateReq, error) {
	// 解析Extension
	extension := new(proto.TCloudAccountExtensionCreateReq)
	if err := common.DecodeExtension(cts, reqExtension, extension); err != nil {
		return nil, err
	}
	// 校验Extension
	if err := extension.Validate(accountType); err != nil {
		return nil, err
	}

	// 检查联通性，账号是否正确
	if accountType != enumor.RegistrationAccount || extension.IsFull() {
		err := a.client.HCService().TCloud.Account.Check(
			cts.Kit.Ctx,
			cts.Kit.Header(),
			&hcproto.TCloudAccountCheckReq{
				CloudMainAccountID: extension.CloudMainAccountID,
				CloudSubAccountID:  extension.CloudSubAccountID,
				CloudSecretID:      extension.CloudSecretID,
				CloudSecretKey:     extension.CloudSecretKey,
			},
		)
		if err != nil {
			return nil, err
		}
	}

	return extension, nil
}

func (a *accountSvc) parseAndCheckAwsExtension(
	cts *rest.Contexts, accountType enumor.AccountType, reqExtension json.RawMessage,
) (*proto.AwsAccountExtensionCreateReq, error) {
	// 解析Extension
	extension := new(proto.AwsAccountExtensionCreateReq)
	if err := common.DecodeExtension(cts, reqExtension, extension); err != nil {
		return nil, err
	}
	// 校验Extension
	if err := extension.Validate(accountType); err != nil {
		return nil, err
	}

	// 检查联通性，账号是否正确
	if accountType != enumor.RegistrationAccount || extension.IsFull() {
		err := a.client.HCService().Aws.Account.Check(
			cts.Kit.Ctx,
			cts.Kit.Header(),
			&hcproto.AwsAccountCheckReq{
				CloudAccountID:   extension.CloudAccountID,
				CloudIamUsername: extension.CloudIamUsername,
				CloudSecretID:    extension.CloudSecretID,
				CloudSecretKey:   extension.CloudSecretKey,
			},
		)
		if err != nil {
			return nil, err
		}
	}

	return extension, nil
}

func (a *accountSvc) parseAndCheckHuaWeiExtension(
	cts *rest.Contexts, accountType enumor.AccountType, reqExtension json.RawMessage,
) (*proto.HuaWeiAccountExtensionCreateReq, error) {
	// 解析Extension
	extension := new(proto.HuaWeiAccountExtensionCreateReq)
	if err := common.DecodeExtension(cts, reqExtension, extension); err != nil {
		return nil, err
	}
	// 校验Extension
	if err := extension.Validate(accountType); err != nil {
		return nil, err
	}

	// 检查联通性，账号是否正确
	if accountType != enumor.RegistrationAccount || extension.IsFull() {
		err := a.client.HCService().HuaWei.Account.Check(
			cts.Kit.Ctx,
			cts.Kit.Header(),
			&hcproto.HuaWeiAccountCheckReq{
				CloudMainAccountName: extension.CloudMainAccountName,
				CloudSubAccountID:    extension.CloudSubAccountID,
				CloudSubAccountName:  extension.CloudSubAccountName,
				CloudSecretID:        extension.CloudSecretID,
				CloudSecretKey:       extension.CloudSecretKey,
				CloudIamUserID:       extension.CloudIamUserID,
				CloudIamUsername:     extension.CloudIamUsername,
			},
		)
		if err != nil {
			return nil, err
		}
	}

	return extension, nil
}

func (a *accountSvc) parseAndCheckGcpExtension(
	cts *rest.Contexts, accountType enumor.AccountType, reqExtension json.RawMessage,
) (*proto.GcpAccountExtensionCreateReq, error) {
	// 解析Extension
	extension := new(proto.GcpAccountExtensionCreateReq)
	if err := common.DecodeExtension(cts, reqExtension, extension); err != nil {
		return nil, err
	}
	// 校验Extension
	if err := extension.Validate(accountType); err != nil {
		return nil, err
	}

	// 检查联通性，账号是否正确
	if accountType != enumor.RegistrationAccount || extension.IsFull() {
		err := a.client.HCService().Gcp.Account.Check(
			cts.Kit.Ctx,
			cts.Kit.Header(),
			&hcproto.GcpAccountCheckReq{
				CloudProjectID:        extension.CloudProjectID,
				CloudServiceSecretKey: extension.CloudServiceSecretKey,
			},
		)
		if err != nil {
			return nil, err
		}
	}

	return extension, nil
}

func (a *accountSvc) parseAndCheckAzureExtension(
	cts *rest.Contexts, accountType enumor.AccountType, reqExtension json.RawMessage,
) (*proto.AzureAccountExtensionCreateReq, error) {
	// 解析Extension
	extension := new(proto.AzureAccountExtensionCreateReq)
	if err := common.DecodeExtension(cts, reqExtension, extension); err != nil {
		return nil, err
	}
	// 校验Extension
	if err := extension.Validate(accountType); err != nil {
		return nil, err
	}

	// 检查联通性，账号是否正确
	if accountType != enumor.RegistrationAccount || extension.IsFull() {
		err := a.client.HCService().Azure.Account.Check(
			cts.Kit.Ctx,
			cts.Kit.Header(),
			&hcproto.AzureAccountCheckReq{
				CloudTenantID:        extension.CloudTenantID,
				CloudSubscriptionID:  extension.CloudSubscriptionID,
				CloudApplicationID:   extension.CloudApplicationID,
				CloudClientSecretKey: extension.CloudClientSecretKey,
			},
		)
		if err != nil {
			return nil, err
		}
	}

	return extension, nil
}

// CheckByID ...
func (a *accountSvc) CheckByID(cts *rest.Contexts) (interface{}, error) {
	req := new(proto.AccountCheckByIDReq)
	if err := cts.DecodeInto(req); err != nil {
		return nil, errf.NewFromErr(errf.DecodeRequestFailed, err)
	}
	if err := req.Validate(); err != nil {
		return nil, errf.NewFromErr(errf.InvalidParameter, err)
	}
	accountID := cts.PathParameter("account_id").String()

	// 校验用户有该账号的更新权限
	if err := a.checkPermission(cts, meta.Update, accountID); err != nil {
		return nil, err
	}

	// 查询该账号对应的Vendor
	baseInfo, err := a.client.DataService().Global.Cloud.GetResourceBasicInfo(
		cts.Kit.Ctx,
		cts.Kit.Header(),
		enumor.AccountCloudResType,
		accountID,
	)
	if err != nil {
		return nil, errf.NewFromErr(errf.InvalidParameter, err)
	}

	switch baseInfo.Vendor {
	case enumor.TCloud:
		_, err = a.parseAndCheckTCloudExtensionByID(cts, accountID, &req.Extension)
	case enumor.Aws:
		_, err = a.parseAndCheckAwsExtensionByID(cts, accountID, &req.Extension)
	case enumor.HuaWei:
		_, err = a.parseAndCheckHuaWeiExtensionByID(cts, accountID, &req.Extension)
	case enumor.Gcp:
		_, err = a.parseAndCheckGcpExtensionByID(cts, accountID, &req.Extension)
	case enumor.Azure:
		_, err = a.parseAndCheckAzureExtensionByID(cts, accountID, &req.Extension)
	default:
		err = fmt.Errorf("no support vendor: %s", baseInfo.Vendor)
	}

	return nil, errf.NewFromErr(errf.InvalidParameter, err)
}

func (a *accountSvc) parseAndCheckTCloudExtensionByID(
	cts *rest.Contexts, accountID string, reqExtension *json.RawMessage,
) (*proto.TCloudAccountExtensionUpdateReq, error) {
	// 解析Extension
	extension := new(proto.TCloudAccountExtensionUpdateReq)
	if err := common.DecodeExtension(cts, *reqExtension, extension); err != nil {
		return nil, err
	}

	// 查询账号其他信息
	account, err := a.client.DataService().TCloud.Account.Get(cts.Kit.Ctx, cts.Kit.Header(), accountID)
	if err != nil {
		return nil, err
	}

	// 校验Extension
	err = extension.Validate(account.Type)
	if err != nil {
		return nil, err
	}

	// 检查联通性，账号是否正确
	if account.Type != enumor.RegistrationAccount || extension.IsFull() {
		err = a.client.HCService().TCloud.Account.Check(
			cts.Kit.Ctx,
			cts.Kit.Header(),
			&hcproto.TCloudAccountCheckReq{
				CloudMainAccountID: account.Extension.CloudMainAccountID,
				CloudSubAccountID:  extension.CloudSubAccountID,
				CloudSecretID:      extension.CloudSecretID,
				CloudSecretKey:     extension.CloudSecretKey,
			},
		)
		if err != nil {
			return nil, err
		}
	}

	return extension, nil
}

func (a *accountSvc) parseAndCheckAwsExtensionByID(
	cts *rest.Contexts, accountID string, reqExtension *json.RawMessage,
) (*proto.AwsAccountExtensionUpdateReq, error) {
	// 解析Extension
	extension := new(proto.AwsAccountExtensionUpdateReq)
	if err := common.DecodeExtension(cts, *reqExtension, extension); err != nil {
		return nil, err
	}

	// 查询账号其他信息
	account, err := a.client.DataService().Aws.Account.Get(cts.Kit.Ctx, cts.Kit.Header(), accountID)
	if err != nil {
		return nil, err
	}

	// 校验Extension
	err = extension.Validate(account.Type)
	if err != nil {
		return nil, err
	}

	// 检查联通性，账号是否正确
	if account.Type != enumor.RegistrationAccount || extension.IsFull() {
		err = a.client.HCService().Aws.Account.Check(
			cts.Kit.Ctx,
			cts.Kit.Header(),
			&hcproto.AwsAccountCheckReq{
				CloudAccountID:   account.Extension.CloudAccountID,
				CloudIamUsername: extension.CloudIamUsername,
				CloudSecretID:    extension.CloudSecretID,
				CloudSecretKey:   extension.CloudSecretKey,
			},
		)
		if err != nil {
			return nil, err
		}
	}

	return extension, nil
}

func (a *accountSvc) parseAndCheckHuaWeiExtensionByID(
	cts *rest.Contexts, accountID string, reqExtension *json.RawMessage,
) (*proto.HuaWeiAccountExtensionUpdateReq, error) {
	// 解析Extension
	extension := new(proto.HuaWeiAccountExtensionUpdateReq)
	if err := common.DecodeExtension(cts, *reqExtension, extension); err != nil {
		return nil, err
	}

	// 查询账号其他信息
	account, err := a.client.DataService().HuaWei.Account.Get(cts.Kit.Ctx, cts.Kit.Header(), accountID)
	if err != nil {
		return nil, err
	}

	// 校验Extension
	err = extension.Validate(account.Type)
	if err != nil {
		return nil, err
	}

	// 检查联通性，账号是否正确
	if account.Type != enumor.RegistrationAccount || extension.IsFull() {
		err = a.client.HCService().HuaWei.Account.Check(
			cts.Kit.Ctx,
			cts.Kit.Header(),
			&hcproto.HuaWeiAccountCheckReq{
				CloudMainAccountName: account.Extension.CloudMainAccountName,
				CloudSubAccountID:    account.Extension.CloudSubAccountID,
				CloudSubAccountName:  account.Extension.CloudSubAccountName,
				CloudIamUserID:       extension.CloudIamUserID,
				CloudIamUsername:     extension.CloudIamUsername,
				CloudSecretID:        extension.CloudSecretID,
				CloudSecretKey:       extension.CloudSecretKey,
			},
		)
		if err != nil {
			return nil, err
		}
	}

	return extension, nil
}

func (a *accountSvc) parseAndCheckGcpExtensionByID(
	cts *rest.Contexts, accountID string, reqExtension *json.RawMessage,
) (*proto.GcpAccountExtensionUpdateReq, error) {
	// 解析Extension
	extension := new(proto.GcpAccountExtensionUpdateReq)
	if err := common.DecodeExtension(cts, *reqExtension, extension); err != nil {
		return nil, err
	}

	// 查询账号其他信息
	account, err := a.client.DataService().Gcp.Account.Get(cts.Kit.Ctx, cts.Kit.Header(), accountID)
	if err != nil {
		return nil, err
	}

	// 校验Extension
	err = extension.Validate(account.Type)
	if err != nil {
		return nil, err
	}

	// 检查联通性，账号是否正确
	if account.Type != enumor.RegistrationAccount || extension.IsFull() {
		err = a.client.HCService().Gcp.Account.Check(
			cts.Kit.Ctx,
			cts.Kit.Header(),
			&hcproto.GcpAccountCheckReq{
				CloudProjectID:        account.Extension.CloudProjectID,
				CloudServiceSecretKey: extension.CloudServiceSecretKey,
			},
		)
		if err != nil {
			return nil, err
		}
	}

	return extension, nil
}

func (a *accountSvc) parseAndCheckAzureExtensionByID(
	cts *rest.Contexts, accountID string, reqExtension *json.RawMessage,
) (*proto.AzureAccountExtensionUpdateReq, error) {
	// 解析Extension
	extension := new(proto.AzureAccountExtensionUpdateReq)
	if err := common.DecodeExtension(cts, *reqExtension, extension); err != nil {
		return nil, err
	}

	// 查询账号其他信息
	account, err := a.client.DataService().Azure.Account.Get(cts.Kit.Ctx, cts.Kit.Header(), accountID)
	if err != nil {
		return nil, err
	}

	// 校验Extension
	err = extension.Validate(account.Type)
	if err != nil {
		return nil, err
	}

	if account.Type != enumor.RegistrationAccount || extension.IsFull() {
		// 检查联通性，账号是否正确
		err = a.client.HCService().Azure.Account.Check(
			cts.Kit.Ctx,
			cts.Kit.Header(),
			&hcproto.AzureAccountCheckReq{
				CloudTenantID:        account.Extension.CloudTenantID,
				CloudSubscriptionID:  account.Extension.CloudSubscriptionID,
				CloudApplicationID:   extension.CloudApplicationID,
				CloudClientSecretKey: extension.CloudClientSecretKey,
			},
		)
		if err != nil {
			return nil, err
		}
	}

	return extension, nil
}
