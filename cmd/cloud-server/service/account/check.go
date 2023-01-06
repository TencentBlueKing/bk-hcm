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
	"fmt"

	proto "hcm/pkg/api/cloud-server"
	hcproto "hcm/pkg/api/hc-service"
	"hcm/pkg/criteria/enumor"
	"hcm/pkg/criteria/errf"
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

	// TODO: 校验用户是否有创建账号的权限

	switch req.Vendor {
	case enumor.TCloud:
		return a.checkForTCloud(cts, req)
	case enumor.Aws:
		return a.checkForAws(cts, req)
	case enumor.HuaWei:
		return a.checkForHuaWei(cts, req)
	case enumor.Gcp:
		return a.checkForGcp(cts, req)
	case enumor.Azure:
		return a.checkForAzure(cts, req)
	default:
		return nil, errf.NewFromErr(errf.InvalidParameter, fmt.Errorf("no support vendor: %s", req.Vendor))
	}
}

func (a *accountSvc) checkForTCloud(cts *rest.Contexts, req *proto.AccountCheckReq) (interface{}, error) {
	// 解析Extension
	extension := new(proto.TCloudAccountExtensionCreateReq)
	if err := a.decodeExtension(cts, req.Extension, extension); err != nil {
		return nil, err
	}
	// 校验Extension
	if err := extension.Validate(); err != nil {
		return nil, err
	}

	// 检查联通性，账号是否正确
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
		return nil, errf.NewFromErr(errf.InvalidParameter, err)
	}

	return nil, nil
}

func (a *accountSvc) checkForAws(cts *rest.Contexts, req *proto.AccountCheckReq) (interface{}, error) {
	// 解析Extension
	extension := new(proto.AwsAccountExtensionCreateReq)
	if err := a.decodeExtension(cts, req.Extension, extension); err != nil {
		return nil, err
	}
	// 校验Extension
	if err := extension.Validate(); err != nil {
		return nil, err
	}

	// 检查联通性，账号是否正确
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
		return nil, errf.NewFromErr(errf.InvalidParameter, err)
	}

	return nil, nil
}

func (a *accountSvc) checkForHuaWei(cts *rest.Contexts, req *proto.AccountCheckReq) (interface{}, error) {
	// 解析Extension
	extension := new(proto.HuaWeiAccountExtensionCreateReq)
	if err := a.decodeExtension(cts, req.Extension, extension); err != nil {
		return nil, err
	}
	// 校验Extension
	if err := extension.Validate(); err != nil {
		return nil, err
	}

	// 检查联通性，账号是否正确
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
		return nil, errf.NewFromErr(errf.InvalidParameter, err)
	}

	return nil, nil
}

func (a *accountSvc) checkForGcp(cts *rest.Contexts, req *proto.AccountCheckReq) (interface{}, error) {
	// 解析Extension
	extension := new(proto.GcpAccountExtensionCreateReq)
	if err := a.decodeExtension(cts, req.Extension, extension); err != nil {
		return nil, err
	}
	// 校验Extension
	if err := extension.Validate(); err != nil {
		return nil, err
	}

	// 检查联通性，账号是否正确
	err := a.client.HCService().Gcp.Account.Check(
		cts.Kit.Ctx,
		cts.Kit.Header(),
		&hcproto.GcpAccountCheckReq{
			CloudProjectID:        extension.CloudProjectID,
			CloudServiceSecretKey: extension.CloudServiceSecretKey,
		},
	)
	if err != nil {
		return nil, errf.NewFromErr(errf.InvalidParameter, err)
	}

	return nil, nil
}

func (a *accountSvc) checkForAzure(cts *rest.Contexts, req *proto.AccountCheckReq) (interface{}, error) {
	// 解析Extension
	extension := new(proto.AzureAccountExtensionCreateReq)
	if err := a.decodeExtension(cts, req.Extension, extension); err != nil {
		return nil, err
	}
	// 校验Extension
	if err := extension.Validate(); err != nil {
		return nil, err
	}

	// 检查联通性，账号是否正确
	err := a.client.HCService().Azure.Account.Check(
		cts.Kit.Ctx,
		cts.Kit.Header(),
		&hcproto.AzureAccountCheckReq{
			CloudTenantID:         extension.CloudTenantID,
			CloudSubscriptionID:   extension.CloudSubscriptionID,
			CloudSubscriptionName: extension.CloudSubscriptionName,
			CloudClientID:         extension.CloudClientID,
			CloudClientSecret:     extension.CloudClientSecret,
		},
	)
	if err != nil {
		return nil, errf.NewFromErr(errf.InvalidParameter, err)
	}

	return nil, nil
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

	// TODO: 校验用户有该账号的权限

	// 查询该账号对应的Vendor
	vendor, err := a.client.DataService().Global.Cloud.GetResourceVendor(
		cts.Kit.Ctx,
		cts.Kit.Header(),
		enumor.AccountCloudResType,
		accountID,
	)
	if err != nil {
		return nil, errf.NewFromErr(errf.InvalidParameter, err)
	}

	switch vendor {
	case enumor.TCloud:
		return a.checkByIDForTCloud(cts, req, accountID)
	case enumor.Aws:
		return a.checkByIDForAws(cts, req, accountID)
	case enumor.HuaWei:
		return a.checkByIDForHuaWei(cts, req, accountID)
	case enumor.Gcp:
		return a.checkByIDForGcp(cts, req, accountID)
	case enumor.Azure:
		return a.checkByIDForAzure(cts, req, accountID)
	default:
		return nil, errf.NewFromErr(errf.InvalidParameter, fmt.Errorf("no support vendor: %s", vendor))
	}
}

func (a *accountSvc) checkByIDForTCloud(cts *rest.Contexts, req *proto.AccountCheckByIDReq, accountID string) (interface{}, error) {
	// 解析Extension
	extension := new(proto.TCloudAccountExtensionCheckByIDReq)
	if err := a.decodeExtension(cts, req.Extension, extension); err != nil {
		return nil, err
	}
	// 校验Extension
	if err := extension.Validate(); err != nil {
		return nil, err
	}

	// 查询账号其他信息
	account, err := a.client.DataService().TCloud.Account.Get(cts.Kit.Ctx, cts.Kit.Header(), accountID)
	if err != nil {
		return nil, errf.NewFromErr(errf.InvalidParameter, err)
	}

	// 检查联通性，账号是否正确
	err = a.client.HCService().TCloud.Account.Check(
		cts.Kit.Ctx,
		cts.Kit.Header(),
		&hcproto.TCloudAccountCheckReq{
			CloudMainAccountID: account.Extension.CloudMainAccountID,
			CloudSubAccountID:  account.Extension.CloudSubAccountID,
			CloudSecretID:      extension.CloudSecretID,
			CloudSecretKey:     extension.CloudSecretKey,
		},
	)
	if err != nil {
		return nil, errf.NewFromErr(errf.InvalidParameter, err)
	}

	return nil, nil
}

func (a *accountSvc) checkByIDForAws(cts *rest.Contexts, req *proto.AccountCheckByIDReq, accountID string) (interface{}, error) {
	// 解析Extension
	extension := new(proto.AwsAccountExtensionCheckByIDReq)
	if err := a.decodeExtension(cts, req.Extension, extension); err != nil {
		return nil, err
	}
	// 校验Extension
	if err := extension.Validate(); err != nil {
		return nil, err
	}

	// 查询账号其他信息
	account, err := a.client.DataService().Aws.Account.Get(cts.Kit.Ctx, cts.Kit.Header(), accountID)
	if err != nil {
		return nil, errf.NewFromErr(errf.InvalidParameter, err)
	}

	// 检查联通性，账号是否正确
	err = a.client.HCService().Aws.Account.Check(
		cts.Kit.Ctx,
		cts.Kit.Header(),
		&hcproto.AwsAccountCheckReq{
			CloudAccountID:   account.Extension.CloudAccountID,
			CloudIamUsername: account.Extension.CloudIamUsername,
			CloudSecretID:    extension.CloudSecretID,
			CloudSecretKey:   extension.CloudSecretKey,
		},
	)
	if err != nil {
		return nil, errf.NewFromErr(errf.InvalidParameter, err)
	}

	return nil, nil
}

func (a *accountSvc) checkByIDForHuaWei(cts *rest.Contexts, req *proto.AccountCheckByIDReq, accountID string) (interface{}, error) {
	// 解析Extension
	extension := new(proto.HuaWeiAccountExtensionCheckByIDReq)
	if err := a.decodeExtension(cts, req.Extension, extension); err != nil {
		return nil, err
	}
	// 校验Extension
	if err := extension.Validate(); err != nil {
		return nil, err
	}

	// 查询账号其他信息
	account, err := a.client.DataService().HuaWei.Account.Get(cts.Kit.Ctx, cts.Kit.Header(), accountID)
	if err != nil {
		return nil, errf.NewFromErr(errf.InvalidParameter, err)
	}

	// 检查联通性，账号是否正确
	err = a.client.HCService().HuaWei.Account.Check(
		cts.Kit.Ctx,
		cts.Kit.Header(),
		&hcproto.HuaWeiAccountCheckReq{
			CloudMainAccountName: account.Extension.CloudMainAccountName,
			CloudSubAccountID:    account.Extension.CloudSubAccountID,
			CloudSubAccountName:  account.Extension.CloudSubAccountName,
			CloudSecretID:        extension.CloudSecretID,
			CloudSecretKey:       extension.CloudSecretKey,
			CloudIamUserID:       account.Extension.CloudIamUserID,
			CloudIamUsername:     account.Extension.CloudIamUsername,
		},
	)
	if err != nil {
		return nil, errf.NewFromErr(errf.InvalidParameter, err)
	}

	return nil, nil
}

func (a *accountSvc) checkByIDForGcp(cts *rest.Contexts, req *proto.AccountCheckByIDReq, accountID string) (interface{}, error) {
	// 解析Extension
	extension := new(proto.GcpAccountExtensionCheckByIDReq)
	if err := a.decodeExtension(cts, req.Extension, extension); err != nil {
		return nil, err
	}
	// 校验Extension
	if err := extension.Validate(); err != nil {
		return nil, err
	}

	// 查询账号其他信息
	account, err := a.client.DataService().Gcp.Account.Get(cts.Kit.Ctx, cts.Kit.Header(), accountID)
	if err != nil {
		return nil, errf.NewFromErr(errf.InvalidParameter, err)
	}

	// 检查联通性，账号是否正确
	err = a.client.HCService().Gcp.Account.Check(
		cts.Kit.Ctx,
		cts.Kit.Header(),
		&hcproto.GcpAccountCheckReq{
			CloudProjectID:        account.Extension.CloudProjectID,
			CloudServiceSecretKey: extension.CloudServiceSecretKey,
		},
	)
	if err != nil {
		return nil, errf.NewFromErr(errf.InvalidParameter, err)
	}

	return nil, nil
}

func (a *accountSvc) checkByIDForAzure(cts *rest.Contexts, req *proto.AccountCheckByIDReq, accountID string) (interface{}, error) {
	// 解析Extension
	extension := new(proto.AzureAccountExtensionCheckByIDReq)
	if err := a.decodeExtension(cts, req.Extension, extension); err != nil {
		return nil, err
	}
	// 校验Extension
	if err := extension.Validate(); err != nil {
		return nil, err
	}

	// 查询账号其他信息
	account, err := a.client.DataService().Azure.Account.Get(cts.Kit.Ctx, cts.Kit.Header(), accountID)
	if err != nil {
		return nil, errf.NewFromErr(errf.InvalidParameter, err)
	}

	// 检查联通性，账号是否正确
	err = a.client.HCService().Azure.Account.Check(
		cts.Kit.Ctx,
		cts.Kit.Header(),
		&hcproto.AzureAccountCheckReq{
			CloudTenantID:         account.Extension.CloudTenantID,
			CloudSubscriptionID:   account.Extension.CloudSubscriptionID,
			CloudSubscriptionName: account.Extension.CloudSubscriptionName,
			CloudClientID:         extension.CloudClientID,
			CloudClientSecret:     extension.CloudClientSecret,
		},
	)
	if err != nil {
		return nil, errf.NewFromErr(errf.InvalidParameter, err)
	}

	return nil, nil
}
