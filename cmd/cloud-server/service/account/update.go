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
	dataproto "hcm/pkg/api/data-service/cloud"
	hcproto "hcm/pkg/api/hc-service"
	"hcm/pkg/criteria/enumor"
	"hcm/pkg/criteria/errf"
	"hcm/pkg/rest"
)

func (a *accountSvc) Update(cts *rest.Contexts) (interface{}, error) {
	req := new(proto.AccountUpdateReq)
	if err := cts.DecodeInto(req); err != nil {
		return nil, errf.NewFromErr(errf.DecodeRequestFailed, err)
	}
	if err := req.Validate(); err != nil {
		return nil, errf.NewFromErr(errf.InvalidParameter, err)
	}
	accountID := cts.PathParameter("account_id").String()

	// TODO: 校验用户有该账号的更新权限

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

	// 更新账号与业务关系
	if req.Attachment != nil && len(req.Attachment.BkBizIDs) > 0 {
		_, err = a.client.DataService().Global.Account.UpdateBizRel(
			cts.Kit.Ctx,
			cts.Kit.Header(),
			accountID,
			&dataproto.AccountBizRelUpdateReq{
				BkBizIDs: req.Attachment.BkBizIDs,
			},
		)
		if err != nil {
			return nil, err
		}
	}

	// 基本信息
	var spec *dataproto.AccountSpecUpdateReq = nil
	if req.Spec != nil {
		spec = &dataproto.AccountSpecUpdateReq{
			Name:         req.Spec.Name,
			Managers:     req.Spec.Managers,
			DepartmentID: req.Spec.DepartmentID,
			Memo:         req.Spec.Memo,
		}
	}

	switch vendor {
	case enumor.TCloud:
		return a.updateForTCloud(cts, req, accountID, spec)
	case enumor.Aws:
		return a.updateForAws(cts, req, accountID, spec)
	case enumor.HuaWei:
		return a.updateForHuaWei(cts, req, accountID, spec)
	case enumor.Gcp:
		return a.updateForGcp(cts, req, accountID, spec)
	case enumor.Azure:
		return a.updateForAzure(cts, req, accountID, spec)
	default:
		return nil, errf.NewFromErr(errf.InvalidParameter, fmt.Errorf("no support vendor: %s", vendor))
	}
}

func (a *accountSvc) updateForTCloud(
	cts *rest.Contexts, req *proto.AccountUpdateReq, accountID string, spec *dataproto.AccountSpecUpdateReq,
) (
	interface{}, error,
) {
	// 解析Extension
	extension := new(proto.TCloudAccountExtensionUpdateReq)
	if req.Extension != nil {
		if err := a.decodeExtension(cts, *req.Extension, extension); err != nil {
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
	}

	var shouldUpdatedExtension *dataproto.TCloudAccountExtensionUpdateReq = nil
	if req.Extension != nil {
		shouldUpdatedExtension = &dataproto.TCloudAccountExtensionUpdateReq{
			CloudSecretID:  extension.CloudSecretID,
			CloudSecretKey: extension.CloudSecretKey,
		}
	}

	// 更新
	if spec != nil || shouldUpdatedExtension != nil {
		_, err := a.client.DataService().TCloud.Account.Update(
			cts.Kit.Ctx,
			cts.Kit.Header(),
			accountID,
			&dataproto.AccountUpdateReq[dataproto.TCloudAccountExtensionUpdateReq]{
				Spec:      spec,
				Extension: shouldUpdatedExtension,
			},
		)
		if err != nil {
			return nil, errf.NewFromErr(errf.InvalidParameter, err)
		}
	}

	return nil, nil

}

func (a *accountSvc) updateForAws(
	cts *rest.Contexts, req *proto.AccountUpdateReq, accountID string, spec *dataproto.AccountSpecUpdateReq,
) (
	interface{}, error,
) {
	// 解析Extension
	extension := new(proto.AwsAccountExtensionUpdateReq)
	if req.Extension != nil {
		if err := a.decodeExtension(cts, *req.Extension, extension); err != nil {
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
	}

	var shouldUpdatedExtension *dataproto.AwsAccountExtensionUpdateReq = nil
	if req.Extension != nil {
		shouldUpdatedExtension = &dataproto.AwsAccountExtensionUpdateReq{
			CloudSecretID:  extension.CloudSecretID,
			CloudSecretKey: extension.CloudSecretKey,
		}
	}

	// 更新
	if spec != nil || shouldUpdatedExtension != nil {
		_, err := a.client.DataService().Aws.Account.Update(
			cts.Kit.Ctx,
			cts.Kit.Header(),
			accountID,
			&dataproto.AccountUpdateReq[dataproto.AwsAccountExtensionUpdateReq]{
				Spec:      spec,
				Extension: shouldUpdatedExtension,
			},
		)
		if err != nil {
			return nil, errf.NewFromErr(errf.InvalidParameter, err)
		}
	}

	return nil, nil

}

func (a *accountSvc) updateForHuaWei(
	cts *rest.Contexts, req *proto.AccountUpdateReq, accountID string, spec *dataproto.AccountSpecUpdateReq,
) (
	interface{}, error,
) {
	// 解析Extension
	extension := new(proto.HuaWeiAccountExtensionUpdateReq)
	if req.Extension != nil {
		if err := a.decodeExtension(cts, *req.Extension, extension); err != nil {
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
	}

	var shouldUpdatedExtension *dataproto.HuaWeiAccountExtensionUpdateReq = nil
	if req.Extension != nil {
		shouldUpdatedExtension = &dataproto.HuaWeiAccountExtensionUpdateReq{
			CloudSecretID:  extension.CloudSecretID,
			CloudSecretKey: extension.CloudSecretKey,
		}
	}

	// 更新
	if spec != nil || shouldUpdatedExtension != nil {
		_, err := a.client.DataService().HuaWei.Account.Update(
			cts.Kit.Ctx,
			cts.Kit.Header(),
			accountID,
			&dataproto.AccountUpdateReq[dataproto.HuaWeiAccountExtensionUpdateReq]{
				Spec:      spec,
				Extension: shouldUpdatedExtension,
			},
		)
		if err != nil {
			return nil, errf.NewFromErr(errf.InvalidParameter, err)
		}
	}

	return nil, nil

}

func (a *accountSvc) updateForGcp(
	cts *rest.Contexts, req *proto.AccountUpdateReq, accountID string, spec *dataproto.AccountSpecUpdateReq,
) (
	interface{}, error,
) {
	// 解析Extension
	extension := new(proto.GcpAccountExtensionUpdateReq)
	if req.Extension != nil {
		if err := a.decodeExtension(cts, *req.Extension, extension); err != nil {
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
	}

	var shouldUpdatedExtension *dataproto.GcpAccountExtensionUpdateReq = nil
	if req.Extension != nil {
		shouldUpdatedExtension = &dataproto.GcpAccountExtensionUpdateReq{
			CloudServiceSecretID:  extension.CloudServiceSecretID,
			CloudServiceSecretKey: extension.CloudServiceSecretKey,
		}
	}

	// 更新
	if spec != nil || shouldUpdatedExtension != nil {
		_, err := a.client.DataService().Gcp.Account.Update(
			cts.Kit.Ctx,
			cts.Kit.Header(),
			accountID,
			&dataproto.AccountUpdateReq[dataproto.GcpAccountExtensionUpdateReq]{
				Spec:      spec,
				Extension: shouldUpdatedExtension,
			},
		)
		if err != nil {
			return nil, errf.NewFromErr(errf.InvalidParameter, err)
		}
	}

	return nil, nil

}

func (a *accountSvc) updateForAzure(
	cts *rest.Contexts, req *proto.AccountUpdateReq, accountID string, spec *dataproto.AccountSpecUpdateReq,
) (
	interface{}, error,
) {
	// 解析Extension
	extension := new(proto.AzureAccountExtensionUpdateReq)
	if req.Extension != nil {
		if err := a.decodeExtension(cts, *req.Extension, extension); err != nil {
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
	}

	var shouldUpdatedExtension *dataproto.AzureAccountExtensionUpdateReq = nil
	if req.Extension != nil {
		shouldUpdatedExtension = &dataproto.AzureAccountExtensionUpdateReq{
			CloudClientID:     extension.CloudClientID,
			CloudClientSecret: extension.CloudClientSecret,
		}
	}

	// 更新
	if spec != nil || shouldUpdatedExtension != nil {
		_, err := a.client.DataService().Azure.Account.Update(
			cts.Kit.Ctx,
			cts.Kit.Header(),
			accountID,
			&dataproto.AccountUpdateReq[dataproto.AzureAccountExtensionUpdateReq]{
				Spec:      spec,
				Extension: shouldUpdatedExtension,
			},
		)
		if err != nil {
			return nil, errf.NewFromErr(errf.InvalidParameter, err)
		}
	}

	return nil, nil

}
