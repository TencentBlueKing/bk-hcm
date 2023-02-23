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

	proto "hcm/pkg/api/cloud-server/account"
	"hcm/pkg/api/core"
	dataproto "hcm/pkg/api/data-service/cloud"
	"hcm/pkg/criteria/enumor"
	"hcm/pkg/criteria/errf"
	"hcm/pkg/iam/meta"
	"hcm/pkg/rest"
	"hcm/pkg/runtime/filter"
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

	// 校验用户是否有录入账号的权限
	if err := a.checkPermission(cts, meta.Import, ""); err != nil {
		return nil, err
	}

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

func (a *accountSvc) isDuplicateMainAccount(
	cts *rest.Contexts, req *proto.AccountCreateReq, mainAccountIDFieldName string, mainAccountIDFieldValue string,
) error {
	// 只需要检查资源账号或安全审计账号的主账号是否重复，其他类型账号不检查
	if req.Type != enumor.ResourceAccount && req.Type != enumor.SecurityAuditAccount {
		return nil
	}

	// TODO: 后续需要解决并发问题
	// 后台查询是否主账号重复
	result, err := a.client.DataService().Global.Account.List(
		cts.Kit.Ctx,
		cts.Kit.Header(),
		&dataproto.AccountListReq{
			Filter: &filter.Expression{
				Op: filter.And,
				Rules: []filter.RuleFactory{
					filter.AtomRule{
						Field: "vendor",
						Op:    filter.Equal.Factory(),
						Value: string(req.Vendor),
					},
					filter.AtomRule{
						Field: "type",
						Op:    filter.Equal.Factory(),
						Value: string(req.Type),
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

func (a *accountSvc) createForTCloud(cts *rest.Contexts, req *proto.AccountCreateReq) (interface{}, error) {
	extension, err := a.parseAndCheckTCloudExtension(cts, req.Type, req.Extension)
	if err != nil {
		return nil, errf.NewFromErr(errf.InvalidParameter, err)
	}

	// 检查资源账号的主账号是否重复
	err = a.isDuplicateMainAccount(cts, req, "cloud_main_account_id", extension.CloudMainAccountID)
	if err != nil {
		return nil, errf.NewFromErr(errf.InvalidParameter, err)
	}

	// 创建
	result, err := a.client.DataService().TCloud.Account.Create(
		cts.Kit.Ctx,
		cts.Kit.Header(),
		&dataproto.AccountCreateReq[dataproto.TCloudAccountExtensionCreateReq]{
			Name:          req.Name,
			Managers:      req.Managers,
			DepartmentIDs: req.DepartmentIDs,
			Type:          req.Type,
			Site:          req.Site,
			Memo:          req.Memo,
			BkBizIDs:      req.BkBizIDs,
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
	extension, err := a.parseAndCheckAwsExtension(cts, req.Type, req.Extension)
	if err != nil {
		return nil, errf.NewFromErr(errf.InvalidParameter, err)
	}

	// 检查资源账号的主账号是否重复
	err = a.isDuplicateMainAccount(cts, req, "cloud_account_id", extension.CloudAccountID)
	if err != nil {
		return nil, errf.NewFromErr(errf.InvalidParameter, err)
	}

	// 创建
	result, err := a.client.DataService().Aws.Account.Create(
		cts.Kit.Ctx,
		cts.Kit.Header(),
		&dataproto.AccountCreateReq[dataproto.AwsAccountExtensionCreateReq]{
			Name:          req.Name,
			Managers:      req.Managers,
			DepartmentIDs: req.DepartmentIDs,
			Type:          req.Type,
			Site:          req.Site,
			Memo:          req.Memo,
			BkBizIDs:      req.BkBizIDs,
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
	extension, err := a.parseAndCheckHuaWeiExtension(cts, req.Type, req.Extension)
	if err != nil {
		return nil, errf.NewFromErr(errf.InvalidParameter, err)
	}

	// 检查资源账号的主账号是否重复
	err = a.isDuplicateMainAccount(cts, req, "cloud_main_account_name", extension.CloudMainAccountName)
	if err != nil {
		return nil, errf.NewFromErr(errf.InvalidParameter, err)
	}

	// 创建
	result, err := a.client.DataService().HuaWei.Account.Create(
		cts.Kit.Ctx,
		cts.Kit.Header(),
		&dataproto.AccountCreateReq[dataproto.HuaWeiAccountExtensionCreateReq]{
			Name:          req.Name,
			Managers:      req.Managers,
			DepartmentIDs: req.DepartmentIDs,
			Type:          req.Type,
			Site:          req.Site,
			Memo:          req.Memo,
			BkBizIDs:      req.BkBizIDs,
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
	extension, err := a.parseAndCheckGcpExtension(cts, req.Type, req.Extension)
	if err != nil {
		return nil, errf.NewFromErr(errf.InvalidParameter, err)
	}

	// 检查资源账号的主账号是否重复
	err = a.isDuplicateMainAccount(cts, req, "cloud_project_id", extension.CloudProjectID)
	if err != nil {
		return nil, errf.NewFromErr(errf.InvalidParameter, err)
	}

	// 创建
	result, err := a.client.DataService().Gcp.Account.Create(
		cts.Kit.Ctx,
		cts.Kit.Header(),
		&dataproto.AccountCreateReq[dataproto.GcpAccountExtensionCreateReq]{
			Name:          req.Name,
			Managers:      req.Managers,
			DepartmentIDs: req.DepartmentIDs,
			Type:          req.Type,
			Site:          req.Site,
			Memo:          req.Memo,
			BkBizIDs:      req.BkBizIDs,
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
	extension, err := a.parseAndCheckAzureExtension(cts, req.Type, req.Extension)
	if err != nil {
		return nil, errf.NewFromErr(errf.InvalidParameter, err)
	}

	// 检查资源账号的主账号是否重复
	err = a.isDuplicateMainAccount(cts, req, "cloud_tenant_id", extension.CloudTenantID)
	if err != nil {
		return nil, errf.NewFromErr(errf.InvalidParameter, err)
	}

	// 创建
	result, err := a.client.DataService().Azure.Account.Create(
		cts.Kit.Ctx,
		cts.Kit.Header(),
		&dataproto.AccountCreateReq[dataproto.AzureAccountExtensionCreateReq]{
			Name:          req.Name,
			Managers:      req.Managers,
			DepartmentIDs: req.DepartmentIDs,
			Type:          req.Type,
			Site:          req.Site,
			Memo:          req.Memo,
			BkBizIDs:      req.BkBizIDs,
			Extension: &dataproto.AzureAccountExtensionCreateReq{
				CloudTenantID:         extension.CloudTenantID,
				CloudSubscriptionID:   extension.CloudSubscriptionID,
				CloudSubscriptionName: extension.CloudSubscriptionName,
				CloudApplicationID:    extension.CloudApplicationID,
				CloudApplicationName:  extension.CloudApplicationName,
				CloudClientSecretID:   extension.CloudClientSecretID,
				CloudClientSecretKey:  extension.CloudClientSecretKey,
			},
		},
	)

	return result, err
}
