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

	"hcm/cmd/cloud-server/service/common"
	proto "hcm/pkg/api/cloud-server"
	dataproto "hcm/pkg/api/data-service/cloud"
	hcproto "hcm/pkg/api/hc-service"
	"hcm/pkg/criteria/enumor"
	"hcm/pkg/criteria/errf"
	"hcm/pkg/rest"
)

// Create ...
func (a *accountSvc) Create(cts *rest.Contexts) (interface{}, error) {
	req := new(proto.AccountCreateReq)
	if err := cts.DecodeInto(req); err != nil {
		return nil, errf.NewFromErr(errf.DecodeRequestFailed, err)
	}
	if err := req.Validate(); err != nil {
		return nil, errf.NewFromErr(errf.InvalidParameter, err)
	}

	// TODO: 校验用户是否有创建账号的权限

	switch req.Vendor {
	case enumor.TCloud:
		return a.createForTCloud(cts, req)
	case enumor.Aws:
		return a.createForAws(cts, req)
	case enumor.HuaWei:
		return a.createForHuaWei(cts, req)
	case enumor.Gcp:
		return a.createForGcp(cts, req)
	case enumor.Azure:
		return a.createForAzure(cts, req)
	default:
		return nil, errf.NewFromErr(errf.InvalidParameter, fmt.Errorf("no support vendor: %s", req.Vendor))
	}
}

func (a *accountSvc) createForTCloud(cts *rest.Contexts, req *proto.AccountCreateReq) (interface{}, error) {
	// 解析Extension
	extension := new(proto.TCloudAccountExtensionCreateReq)
	if err := common.DecodeExtension(cts, req.Extension, extension); err != nil {
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

	// 创建
	result, err := a.client.DataService().TCloud.Account.Create(
		cts.Kit.Ctx,
		cts.Kit.Header(),
		&dataproto.AccountCreateReq[dataproto.TCloudAccountExtensionCreateReq]{
			Spec: &dataproto.AccountSpecCreateReq{
				Name:         req.Spec.Name,
				Managers:     req.Spec.Managers,
				DepartmentID: req.Spec.DepartmentID,
				Type:         req.Spec.Type,
				Site:         req.Spec.Site,
				Memo:         req.Spec.Memo,
			},
			Attachment: &dataproto.AccountAttachmentCreateReq{BkBizIDs: req.Attachment.BkBizIDs},
			Extension: &dataproto.TCloudAccountExtensionCreateReq{
				CloudMainAccountID: extension.CloudMainAccountID,
				CloudSubAccountID:  extension.CloudSubAccountID,
				CloudSecretID:      extension.CloudSecretID,
				CloudSecretKey:     extension.CloudSecretKey,
			},
		},
	)

	return result, err
}

func (a *accountSvc) createForAws(cts *rest.Contexts, req *proto.AccountCreateReq) (interface{}, error) {
	// 解析Extension
	extension := new(proto.AwsAccountExtensionCreateReq)
	if err := common.DecodeExtension(cts, req.Extension, extension); err != nil {
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

	// 创建
	result, err := a.client.DataService().Aws.Account.Create(
		cts.Kit.Ctx,
		cts.Kit.Header(),
		&dataproto.AccountCreateReq[dataproto.AwsAccountExtensionCreateReq]{
			Spec: &dataproto.AccountSpecCreateReq{
				Name:         req.Spec.Name,
				Managers:     req.Spec.Managers,
				DepartmentID: req.Spec.DepartmentID,
				Type:         req.Spec.Type,
				Site:         req.Spec.Site,
				Memo:         req.Spec.Memo,
			},
			Attachment: &dataproto.AccountAttachmentCreateReq{BkBizIDs: req.Attachment.BkBizIDs},
			Extension: &dataproto.AwsAccountExtensionCreateReq{
				CloudAccountID:   extension.CloudAccountID,
				CloudIamUsername: extension.CloudIamUsername,
				CloudSecretID:    extension.CloudSecretID,
				CloudSecretKey:   extension.CloudSecretKey,
			},
		},
	)

	return result, err
}

func (a *accountSvc) createForHuaWei(cts *rest.Contexts, req *proto.AccountCreateReq) (interface{}, error) {
	// 解析Extension
	extension := new(proto.HuaWeiAccountExtensionCreateReq)
	if err := common.DecodeExtension(cts, req.Extension, extension); err != nil {
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

	// 创建
	result, err := a.client.DataService().HuaWei.Account.Create(
		cts.Kit.Ctx,
		cts.Kit.Header(),
		&dataproto.AccountCreateReq[dataproto.HuaWeiAccountExtensionCreateReq]{
			Spec: &dataproto.AccountSpecCreateReq{
				Name:         req.Spec.Name,
				Managers:     req.Spec.Managers,
				DepartmentID: req.Spec.DepartmentID,
				Type:         req.Spec.Type,
				Site:         req.Spec.Site,
				Memo:         req.Spec.Memo,
			},
			Attachment: &dataproto.AccountAttachmentCreateReq{BkBizIDs: req.Attachment.BkBizIDs},
			Extension: &dataproto.HuaWeiAccountExtensionCreateReq{
				CloudMainAccountName: extension.CloudMainAccountName,
				CloudSubAccountID:    extension.CloudSubAccountID,
				CloudSubAccountName:  extension.CloudSubAccountName,
				CloudSecretID:        extension.CloudSecretID,
				CloudSecretKey:       extension.CloudSecretKey,
				CloudIamUserID:       extension.CloudIamUserID,
				CloudIamUsername:     extension.CloudIamUsername,
			},
		},
	)

	return result, err
}

func (a *accountSvc) createForGcp(cts *rest.Contexts, req *proto.AccountCreateReq) (interface{}, error) {
	// 解析Extension
	extension := new(proto.GcpAccountExtensionCreateReq)
	if err := common.DecodeExtension(cts, req.Extension, extension); err != nil {
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

	// 创建
	result, err := a.client.DataService().Gcp.Account.Create(
		cts.Kit.Ctx,
		cts.Kit.Header(),
		&dataproto.AccountCreateReq[dataproto.GcpAccountExtensionCreateReq]{
			Spec: &dataproto.AccountSpecCreateReq{
				Name:         req.Spec.Name,
				Managers:     req.Spec.Managers,
				DepartmentID: req.Spec.DepartmentID,
				Type:         req.Spec.Type,
				Site:         req.Spec.Site,
				Memo:         req.Spec.Memo,
			},
			Attachment: &dataproto.AccountAttachmentCreateReq{BkBizIDs: req.Attachment.BkBizIDs},
			Extension: &dataproto.GcpAccountExtensionCreateReq{
				CloudProjectID:          extension.CloudProjectID,
				CloudProjectName:        extension.CloudProjectName,
				CloudServiceAccountID:   extension.CloudServiceAccountID,
				CloudServiceAccountName: extension.CloudServiceAccountName,
				CloudServiceSecretID:    extension.CloudServiceSecretID,
				CloudServiceSecretKey:   extension.CloudServiceSecretKey,
			},
		},
	)

	return result, err
}

func (a *accountSvc) createForAzure(cts *rest.Contexts, req *proto.AccountCreateReq) (interface{}, error) {
	// 解析Extension
	extension := new(proto.AzureAccountExtensionCreateReq)
	if err := common.DecodeExtension(cts, req.Extension, extension); err != nil {
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

	// 创建
	result, err := a.client.DataService().Azure.Account.Create(
		cts.Kit.Ctx,
		cts.Kit.Header(),
		&dataproto.AccountCreateReq[dataproto.AzureAccountExtensionCreateReq]{
			Spec: &dataproto.AccountSpecCreateReq{
				Name:         req.Spec.Name,
				Managers:     req.Spec.Managers,
				DepartmentID: req.Spec.DepartmentID,
				Type:         req.Spec.Type,
				Site:         req.Spec.Site,
				Memo:         req.Spec.Memo,
			},
			Attachment: &dataproto.AccountAttachmentCreateReq{BkBizIDs: req.Attachment.BkBizIDs},
			Extension: &dataproto.AzureAccountExtensionCreateReq{
				CloudTenantID:         extension.CloudTenantID,
				CloudSubscriptionID:   extension.CloudSubscriptionID,
				CloudSubscriptionName: extension.CloudSubscriptionName,
				CloudApplicationID:    extension.CloudApplicationID,
				CloudApplicationName:  extension.CloudApplicationName,
				CloudClientID:         extension.CloudClientID,
				CloudClientSecret:     extension.CloudClientSecret,
			},
		},
	)

	return result, err
}
