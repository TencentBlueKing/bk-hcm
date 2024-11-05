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
	"errors"
	"fmt"

	"hcm/cmd/cloud-server/service/common"
	proto "hcm/pkg/api/cloud-server/account"
	"hcm/pkg/api/core"
	"hcm/pkg/api/data-service/cloud"
	hcproto "hcm/pkg/api/hc-service/account"
	"hcm/pkg/client"
	"hcm/pkg/criteria/enumor"
	"hcm/pkg/criteria/errf"
	"hcm/pkg/iam/meta"
	"hcm/pkg/rest"
	"hcm/pkg/runtime/filter"
)

// CheckAccount account for RegistrationAccount，for backward compatibility
func (a *accountSvc) CheckAccount(cts *rest.Contexts) (interface{}, error) {
	req := new(proto.AccountCheckReq)
	if err := cts.DecodeInto(req); err != nil {
		return nil, errf.NewFromErr(errf.DecodeRequestFailed, err)
	}
	if err := req.Validate(); err != nil {
		return nil, errf.NewFromErr(errf.InvalidParameter, err)
	}
	if req.Type != enumor.RegistrationAccount && req.Type != enumor.SecurityAuditAccount {
		return nil, errors.New("only support check account type of registration or security_audit ")
	}

	var err error
	switch req.Vendor {
	case enumor.TCloud:
		_, err = ParseAndCheckTCloudExtension(cts, a.client, req.Type, req.Extension)
	case enumor.Aws:
		_, err = ParseAndCheckAwsExtension(cts, a.client, req.Type, req.Extension)
	case enumor.HuaWei:
		_, err = ParseAndCheckHuaWeiExtension(cts, a.client, req.Type, req.Extension)
	case enumor.Gcp:
		_, err = ParseAndCheckGcpExtension(cts, a.client, req.Type, req.Extension)
	case enumor.Azure:
		_, err = ParseAndCheckAzureExtension(cts, a.client, req.Type, req.Extension)
	default:
		err = fmt.Errorf("no support vendor: %s", req.Vendor)
	}

	return nil, errf.NewFromErr(errf.InvalidParameter, err)
}

// TODO: ParseAndCheckXXXXExtension 公开是为了与申请新增账号复用，但是这里只是复用，没有抽象，不应该复用XXXXAccountExtensionCreateReq数据结构

// ParseAndCheckTCloudExtension  联通性校验，并检查字段是否匹配
func ParseAndCheckTCloudExtension(
	cts *rest.Contexts, client *client.ClientSet, accountType enumor.AccountType, reqExtension json.RawMessage,
) (*proto.TCloudAccountExtensionCreateReq, error) {
	// 解析Extension
	extension := new(proto.TCloudAccountExtensionCreateReq)
	if err := common.DecodeExtension(cts.Kit, reqExtension, extension); err != nil {
		return nil, err
	}
	// 校验Extension
	if err := extension.Validate(accountType); err != nil {
		return nil, err
	}

	// 检查联通性，账号是否正确
	if accountType != enumor.RegistrationAccount || extension.IsFull() {
		err := client.HCService().TCloud.Account.Check(
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

// ParseAndCheckAwsExtension  联通性校验，并检查字段是否匹配
func ParseAndCheckAwsExtension(
	cts *rest.Contexts, client *client.ClientSet, accountType enumor.AccountType, reqExtension json.RawMessage,
) (*proto.AwsAccountExtensionCreateReq, error) {
	// 解析Extension
	extension := new(proto.AwsAccountExtensionCreateReq)
	if err := common.DecodeExtension(cts.Kit, reqExtension, extension); err != nil {
		return nil, err
	}
	// 校验Extension
	if err := extension.Validate(accountType); err != nil {
		return nil, err
	}

	// 检查联通性，账号是否正确
	if accountType != enumor.RegistrationAccount || extension.IsFull() {
		err := client.HCService().Aws.Account.Check(
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

// ParseAndCheckHuaWeiExtension  联通性校验，并检查字段是否匹配
func ParseAndCheckHuaWeiExtension(
	cts *rest.Contexts, client *client.ClientSet, accountType enumor.AccountType, reqExtension json.RawMessage,
) (*proto.HuaWeiAccountExtensionCreateReq, error) {
	// 解析Extension
	extension := new(proto.HuaWeiAccountExtensionCreateReq)
	if err := common.DecodeExtension(cts.Kit, reqExtension, extension); err != nil {
		return nil, err
	}
	// 校验Extension
	if err := extension.Validate(accountType); err != nil {
		return nil, err
	}

	// 检查联通性，账号是否正确
	if accountType != enumor.RegistrationAccount || extension.IsFull() {
		err := client.HCService().HuaWei.Account.Check(
			cts.Kit.Ctx,
			cts.Kit.Header(),
			&hcproto.HuaWeiAccountCheckReq{
				CloudSubAccountID:   extension.CloudSubAccountID,
				CloudSubAccountName: extension.CloudSubAccountName,
				CloudSecretID:       extension.CloudSecretID,
				CloudSecretKey:      extension.CloudSecretKey,
				CloudIamUserID:      extension.CloudIamUserID,
				CloudIamUsername:    extension.CloudIamUsername,
			},
		)
		if err != nil {
			return nil, err
		}
	}

	return extension, nil
}

// ParseAndCheckGcpExtension  联通性校验，并检查字段是否匹配
func ParseAndCheckGcpExtension(cts *rest.Contexts, client *client.ClientSet,
	accountType enumor.AccountType, reqExtension json.RawMessage) (*proto.GcpAccountExtensionCreateReq, error) {

	// 解析Extension
	extension := new(proto.GcpAccountExtensionCreateReq)
	if err := common.DecodeExtension(cts.Kit, reqExtension, extension); err != nil {
		return nil, err
	}
	// 校验Extension
	if err := extension.Validate(accountType); err != nil {
		return nil, err
	}

	// 检查联通性，账号是否正确
	if accountType != enumor.RegistrationAccount || extension.IsFull() {
		err := client.HCService().Gcp.Account.Check(
			cts.Kit.Ctx,
			cts.Kit.Header(),
			&hcproto.GcpAccountCheckReq{
				CloudServiceSecretKey:   extension.CloudServiceSecretKey,
				CloudProjectID:          extension.CloudProjectID,
				CloudProjectName:        extension.CloudProjectName,
				CloudServiceAccountID:   extension.CloudServiceAccountID,
				CloudServiceAccountName: extension.CloudServiceAccountName,
				CloudServiceSecretID:    extension.CloudServiceSecretID,
			},
		)
		if err != nil {
			return nil, err
		}
	}

	return extension, nil
}

// ParseAndCheckAzureExtension  联通性校验，并检查字段是否匹配
func ParseAndCheckAzureExtension(
	cts *rest.Contexts, client *client.ClientSet, accountType enumor.AccountType, reqExtension json.RawMessage,
) (*proto.AzureAccountExtensionCreateReq, error) {
	// 解析Extension
	extension := new(proto.AzureAccountExtensionCreateReq)
	if err := common.DecodeExtension(cts.Kit, reqExtension, extension); err != nil {
		return nil, err
	}
	// 校验Extension
	if err := extension.Validate(accountType); err != nil {
		return nil, err
	}

	// 检查联通性，账号是否正确
	if accountType != enumor.RegistrationAccount || extension.IsFull() {
		err := client.HCService().Azure.Account.Check(
			cts.Kit.Ctx,
			cts.Kit.Header(),
			&hcproto.AzureAccountCheckReq{
				CloudTenantID:         extension.CloudTenantID,
				CloudApplicationID:    extension.CloudApplicationID,
				CloudClientSecretKey:  extension.CloudClientSecretKey,
				CloudApplicationName:  extension.CloudApplicationName,
				CloudSubscriptionName: extension.CloudSubscriptionName,
				CloudSubscriptionID:   extension.CloudSubscriptionID,
			},
		)
		if err != nil {
			return nil, err
		}
	}

	return extension, nil
}

// CheckByID 更新秘钥信息的时候，重新获取一次信息覆盖并比较，和录入账号逻辑基本相同，但是判断账号唯一的id不能变
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
	baseInfo, err := a.client.DataService().Global.Cloud.GetResBasicInfo(cts.Kit,
		enumor.AccountCloudResType, accountID)
	if err != nil {
		return nil, errf.NewFromErr(errf.InvalidParameter, err)
	}

	switch baseInfo.Vendor {
	case enumor.TCloud:
		_, err = a.parseAndCheckTCloudExtensionByID(cts, accountID, req.Extension)
	case enumor.Aws:
		_, err = a.parseAndCheckAwsExtensionByID(cts, accountID, req.Extension)
	case enumor.HuaWei:
		_, err = a.parseAndCheckHuaWeiExtensionByID(cts, accountID, req.Extension)
	case enumor.Gcp:
		_, err = a.parseAndCheckGcpExtensionByID(cts, accountID, req.Extension)
	case enumor.Azure:
		_, err = a.parseAndCheckAzureExtensionByID(cts, accountID, req.Extension)
	default:
		err = fmt.Errorf("no support vendor: %s", baseInfo.Vendor)
	}

	return nil, errf.NewFromErr(errf.InvalidParameter, err)
}

func (a *accountSvc) parseAndCheckTCloudExtensionByID(
	cts *rest.Contexts, accountID string, reqExtension json.RawMessage,
) (*proto.TCloudAccountExtensionUpdateReq, error) {
	// 解析Extension
	extension := new(proto.TCloudAccountExtensionUpdateReq)
	if err := common.DecodeExtension(cts.Kit, reqExtension, extension); err != nil {
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
				// 传入数据库中的主账号信息，如果发生变更会报错
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
	cts *rest.Contexts, accountID string, reqExtension json.RawMessage,
) (*proto.AwsAccountExtensionUpdateReq, error) {
	// 解析Extension
	extension := new(proto.AwsAccountExtensionUpdateReq)
	if err := common.DecodeExtension(cts.Kit, reqExtension, extension); err != nil {
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
				// 传入数据库中的主账号信息，如果发生变更会报错
				CloudAccountID: account.Extension.CloudAccountID,

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
	cts *rest.Contexts, accountID string, reqExtension json.RawMessage,
) (*proto.HuaWeiAccountExtensionUpdateReq, error) {
	// 解析Extension
	extension := new(proto.HuaWeiAccountExtensionUpdateReq)
	if err := common.DecodeExtension(cts.Kit, reqExtension, extension); err != nil {
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
				// 传入数据库中的主账号信息，如果发生变更会报错
				CloudSubAccountID: account.Extension.CloudSubAccountID,

				CloudSubAccountName: extension.CloudSubAccountName,
				CloudIamUserID:      extension.CloudIamUserID,
				CloudIamUsername:    extension.CloudIamUsername,
				CloudSecretID:       extension.CloudSecretID,
				CloudSecretKey:      extension.CloudSecretKey,
			},
		)
		if err != nil {
			return nil, err
		}
	}

	return extension, nil
}

func (a *accountSvc) parseAndCheckGcpExtensionByID(
	cts *rest.Contexts, accountID string, reqExtension json.RawMessage,
) (*proto.GcpAccountExtensionUpdateReq, error) {
	// 解析Extension
	extension := new(proto.GcpAccountExtensionUpdateReq)
	if err := common.DecodeExtension(cts.Kit, reqExtension, extension); err != nil {
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
				CloudProjectID: account.Extension.CloudProjectID,

				CloudServiceSecretKey:   extension.CloudServiceSecretKey,
				CloudProjectName:        extension.CloudProjectName,
				CloudServiceAccountID:   extension.CloudServiceAccountID,
				CloudServiceAccountName: extension.CloudServiceAccountName,
				CloudServiceSecretID:    extension.CloudServiceSecretID,
			},
		)
		if err != nil {
			return nil, err
		}
	}

	return extension, nil
}

func (a *accountSvc) parseAndCheckAzureExtensionByID(
	cts *rest.Contexts, accountID string, reqExtension json.RawMessage,
) (*proto.AzureAccountExtensionUpdateReq, error) {
	// 解析Extension
	extension := new(proto.AzureAccountExtensionUpdateReq)
	if err := common.DecodeExtension(cts.Kit, reqExtension, extension); err != nil {
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
				CloudSubscriptionID:   account.Extension.CloudSubscriptionID,
				CloudTenantID:         extension.CloudTenantID,
				CloudApplicationID:    extension.CloudApplicationID,
				CloudApplicationName:  extension.CloudApplicationName,
				CloudSubscriptionName: extension.CloudSubscriptionName,
				CloudClientSecretKey:  extension.CloudClientSecretKey,
			},
		)
		if err != nil {
			return nil, err
		}
	}

	return extension, nil
}

// CheckDuplicateMainAccount 检查主账号是否重复
func CheckDuplicateMainAccount(cts *rest.Contexts, client *client.ClientSet, vendor enumor.Vendor,
	accountType enumor.AccountType, mainAccountIDFieldValue string) error {

	// 只校验资源账号的主账号是否重复，其他类型账号不检查
	if accountType != enumor.ResourceAccount {
		return nil
	}

	// TODO: 后续需要解决并发问题
	// 后台查询是否主账号重复
	mainAccountIDFieldName := vendor.GetMainAccountIDField()

	result, err := client.DataService().Global.Account.List(
		cts.Kit.Ctx,
		cts.Kit.Header(),
		&cloud.AccountListReq{
			Filter: &filter.Expression{
				Op: filter.And,
				Rules: []filter.RuleFactory{
					filter.AtomRule{
						Field: "vendor",
						Op:    filter.Equal.Factory(),
						Value: string(vendor),
					},
					filter.AtomRule{
						Field: "type",
						Op:    filter.Equal.Factory(),
						Value: string(accountType),
					},
					filter.AtomRule{
						Field: fmt.Sprintf("extension.%s", mainAccountIDFieldName),
						Op:    filter.JSONEqual.Factory(),
						Value: mainAccountIDFieldValue,
					},
				},
			},
			Page: &core.BasePage{
				Count: true,
			},
		},
	)
	if err != nil {
		return err
	}

	if result.Count > 0 {
		return fmt.Errorf("%s[%s] should be not duplicate", mainAccountIDFieldName, mainAccountIDFieldValue)
	}

	return nil
}
