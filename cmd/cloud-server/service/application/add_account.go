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

package application

import (
	"fmt"
	"strings"

	accountsvc "hcm/cmd/cloud-server/service/account"
	proto "hcm/pkg/api/cloud-server/application"
	"hcm/pkg/api/core"
	dataproto "hcm/pkg/api/data-service"
	dataprotocloud "hcm/pkg/api/data-service/cloud"
	"hcm/pkg/criteria/constant"
	"hcm/pkg/criteria/enumor"
	"hcm/pkg/criteria/errf"
	"hcm/pkg/rest"
	"hcm/pkg/runtime/filter"
	"hcm/pkg/thirdparty/esb/cmdb"
	"hcm/pkg/thirdparty/esb/itsm"
	"hcm/pkg/thirdparty/esb/usermgr"
	"hcm/pkg/tools/json"

	"github.com/TencentBlueKing/gopkg/conv"
)

var (
	vendorMainAccountIDFieldMap = map[enumor.Vendor]string{
		enumor.TCloud: "cloud_main_account_id",
		enumor.Aws:    "cloud_account_id",
		enumor.HuaWei: "cloud_main_account_name",
		enumor.Gcp:    "cloud_project_id",
		enumor.Azure:  "cloud_tenant_id",
	}

	vendorNameMap = map[enumor.Vendor]string{
		enumor.TCloud: "腾讯云",
		enumor.Aws:    "亚马逊云",
		enumor.HuaWei: "华为云",
		enumor.Gcp:    "谷歌云",
		enumor.Azure:  "微软云",
	}

	accountTypNameMap = map[enumor.AccountType]string{
		enumor.RegistrationAccount:  "登记账号",
		enumor.ResourceAccount:      "资源账号",
		enumor.SecurityAuditAccount: "安全审计账号",
	}

	accountSiteTypeNameMap = map[enumor.AccountSiteType]string{
		enumor.InternationalSite: "国际站",
		enumor.ChinaSite:         "中国站",
	}
)

// CreateForAddAccount ...
func (a *applicationSvc) CreateForAddAccount(cts *rest.Contexts) (interface{}, error) {
	req := new(proto.AccountAddReq)
	if err := cts.DecodeInto(req); err != nil {
		return nil, errf.NewFromErr(errf.DecodeRequestFailed, err)
	}

	// 检查数据正确性
	if err := a.checkForAddAccount(cts, req); err != nil {
		return nil, errf.NewFromErr(errf.InvalidParameter, err)
	}

	// 密钥加密
	secretKeyField := accountsvc.VendorSecretKeyFieldMap[req.Vendor]
	req.Extension[secretKeyField] = a.cipher.EncryptToBase64(conv.ToString(req.Extension[secretKeyField]))

	// 调用ITSM
	sn, err := a.createItsmTicketForAddAccount(cts, req)
	if err != nil {
		return nil, err
	}

	// 调用DB创建单据
	content, err := json.MarshalToString(req)
	if err != nil {
		return nil, errf.NewFromErr(errf.InvalidParameter, fmt.Errorf("json marshal request data failed, err: %w", err))
	}

	result, err := a.client.DataService().Global.Application.Create(
		cts.Kit.Ctx,
		cts.Kit.Header(),
		&dataproto.ApplicationCreateReq{
			SN:        sn,
			Type:      enumor.AddAccount,
			Status:    enumor.Pending,
			Applicant: cts.Kit.User,
			Content:   content,
		},
	)
	if err != nil {
		return nil, err
	}

	return result, nil
}

func (a *applicationSvc) checkForAddAccount(cts *rest.Contexts, req *proto.AccountAddReq) error {
	if err := req.Validate(); err != nil {
		return err
	}

	// 检查名称是否重复
	if err := a.isDuplicateName(cts, req.Name); err != nil {
		return err
	}

	// 检查账号是否有效
	extensionJson, err := json.Marshal(req.Extension)
	if err != nil {
		return fmt.Errorf("json marshal extension failed, err: %w", err)
	}
	switch req.Vendor {
	case enumor.TCloud:
		_, err = accountsvc.ParseAndCheckTCloudExtension(cts, a.client, req.Type, extensionJson)
	case enumor.Aws:
		_, err = accountsvc.ParseAndCheckAwsExtension(cts, a.client, req.Type, extensionJson)
	case enumor.HuaWei:
		_, err = accountsvc.ParseAndCheckHuaWeiExtension(cts, a.client, req.Type, extensionJson)
	case enumor.Gcp:
		_, err = accountsvc.ParseAndCheckGcpExtension(cts, a.client, req.Type, extensionJson)
	case enumor.Azure:
		_, err = accountsvc.ParseAndCheckAzureExtension(cts, a.client, req.Type, extensionJson)
	default:
		err = fmt.Errorf("no support vendor: %s", req.Vendor)
	}
	if err != nil {
		return err
	}

	// 检查资源账号的主账号是否重复
	mainAccountIDField := vendorMainAccountIDFieldMap[req.Vendor]
	err = accountsvc.IsDuplicateMainAccount(
		cts, a.client, req.Vendor, req.Type, mainAccountIDField, conv.ToString(req.Extension[mainAccountIDField]),
	)
	if err != nil {
		return err
	}

	return nil
}

func (a *applicationSvc) isDuplicateName(cts *rest.Contexts, name string) error {
	// TODO: 后续需要解决并发问题
	// 后台查询是否主账号重复
	result, err := a.client.DataService().Global.Account.List(
		cts.Kit.Ctx,
		cts.Kit.Header(),
		&dataprotocloud.AccountListReq{
			Filter: &filter.Expression{
				Op: filter.And,
				Rules: []filter.RuleFactory{
					filter.AtomRule{
						Field: "name",
						Op:    filter.Equal.Factory(),
						Value: name,
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
		return fmt.Errorf("account name [%s] has already exits, should be not duplicate", name)
	}

	return nil
}

func (a *applicationSvc) renderItsmApplicationForm(cts *rest.Contexts, req *proto.AccountAddReq) (string, error) {
	type formItem struct {
		Label string
		Value string
	}

	// 公共基础信息
	formItems := []formItem{
		{Label: "账号类型", Value: accountTypNameMap[req.Type]},
		{Label: "名称", Value: req.Name},
		{Label: "云厂商", Value: vendorNameMap[req.Vendor]},
		{Label: "站点类型", Value: accountSiteTypeNameMap[req.Site]},
	}

	// 云特性信息
	switch req.Vendor {
	case enumor.TCloud:
		formItems = append(formItems, []formItem{
			{Label: "主账号ID", Value: req.Extension["cloud_main_account_id"]},
			{Label: "子账号ID", Value: req.Extension["cloud_sub_account_id"]},
			{Label: "SecretId", Value: req.Extension["cloud_secret_id"]},
		}...)
	case enumor.Aws:
		formItems = append(formItems, []formItem{
			{Label: "账号ID", Value: req.Extension["cloud_account_id"]},
			{Label: "IAM用户名称", Value: req.Extension["cloud_iam_username"]},
			{Label: "SecretId/密钥ID", Value: req.Extension["cloud_secret_id"]},
		}...)
	case enumor.HuaWei:
		formItems = append(formItems, []formItem{
			{Label: "主账号名", Value: req.Extension["cloud_main_account_name"]},
			{Label: "账号ID", Value: req.Extension["cloud_sub_account_id"]},
			{Label: "账号名称", Value: req.Extension["cloud_sub_account_name"]},
			{Label: "IAM用户ID", Value: req.Extension["cloud_iam_user_id"]},
			{Label: "IAM用户名称", Value: req.Extension["cloud_iam_username"]},
			{Label: "SecretId/密钥ID", Value: req.Extension["cloud_secret_id"]},
		}...)
	case enumor.Gcp:
		formItems = append(formItems, []formItem{
			{Label: "项目 ID", Value: req.Extension["cloud_project_id"]},
			{Label: "项目名称", Value: req.Extension["cloud_project_name"]},
			{Label: "服务账号ID", Value: req.Extension["cloud_service_account_id"]},
			{Label: "服务账号名称", Value: req.Extension["cloud_service_account_name"]},
			{Label: "服务账号密钥ID", Value: req.Extension["cloud_service_secret_id"]},
		}...)
	case enumor.Azure:
		formItems = append(formItems, []formItem{
			{Label: "租户 ID", Value: req.Extension["cloud_tenant_id"]},
			{Label: "订阅 ID", Value: req.Extension["cloud_subscription_id"]},
			{Label: "订阅名称", Value: req.Extension["cloud_subscription_name"]},
			{Label: "应用程序(客户端) ID", Value: req.Extension["cloud_application_id"]},
			{Label: "应用程序名称", Value: req.Extension["cloud_application_name"]},
			{Label: "客户端密钥ID", Value: req.Extension["cloud_client_secret_id"]},
		}...)
	}

	// 负责人
	formItems = append(formItems, formItem{Label: "责任人", Value: strings.Join(req.Managers, ",")})

	// 查询部门名称
	departmentInfo, err := a.esbClient.Usermgr().RetrieveDepartment(cts.Kit.Ctx, &usermgr.RetrieveDepartmentReq{
		ID:            req.DepartmentIDs[0],
		Fields:        []string{},
		WithAncestors: true,
	})
	if err != nil {
		return "", fmt.Errorf("call usermgr retrieve department api failed, err: %v", err)
	}
	formItems = append(formItems, formItem{Label: "组织架构", Value: departmentInfo.FullName})

	// 查询业务名称
	if req.BkBizIDs[0] == constant.AttachedAllBiz {
		formItems = append(formItems, formItem{Label: "使用业务", Value: "全部"})
	} else {
		// 查询CC业务
		searchResp, err := a.esbClient.Cmdb().SearchBusiness(cts.Kit.Ctx, &cmdb.SearchBizParams{
			Fields: []string{"bk_biz_id", "bk_biz_name"},
		})
		if err != nil {
			return "", fmt.Errorf("call cmdb search business api failed, err: %v", err)
		}
		// 业务ID和Name映射关系
		bizNameMap := map[int64]string{}
		for _, biz := range searchResp.SearchBizResult.Info {
			bizNameMap[biz.BizID] = biz.BizName
		}
		// 匹配出业务名称列表
		bizNames := make([]string, 0, len(req.BkBizIDs))
		for _, bizID := range req.BkBizIDs {
			bizNames = append(bizNames, bizNameMap[bizID])
		}
		formItems = append(formItems, formItem{Label: "使用业务", Value: strings.Join(bizNames, ",")})
	}

	// 备注
	if req.Memo != nil && *req.Memo != "" {
		formItems = append(formItems, formItem{Label: "备注", Value: *req.Memo})
	}

	// 转换为ITSM表单内容数据
	content := make([]string, 0, len(formItems))
	for _, i := range formItems {
		content = append(content, fmt.Sprintf("%s: %s", i.Label, i.Value))
	}
	return strings.Join(content, "\n"), nil
}

func (a *applicationSvc) createItsmTicketForAddAccount(cts *rest.Contexts, req *proto.AccountAddReq) (string, error) {
	// 渲染ITSM表单内容
	contentDisplay, err := a.renderItsmApplicationForm(cts, req)
	if err != nil {
		return "", fmt.Errorf("render itsm application form error: %w", err)
	}

	params := &itsm.CreateTicketParams{
		// TODO: 从DB获取申请账号类型所需的流程服务ID，正在调整Helm Chart的初始化job,支持写入DB
		ServiceID:      104,
		Creator:        cts.Kit.User,
		CallbackURL:    a.getCallbackUrl(),
		Title:          fmt.Sprintf("申请新增账号[%s]", req.Name),
		ContentDisplay: contentDisplay,
		VariableApprovers: []itsm.VariableApprover{
			{
				Variable:  "superuser",
				Approvers: []string{"admin"},
			},
		},
	}

	sn, err := a.esbClient.Itsm().CreateTicket(cts.Kit.Ctx, params)
	if err != nil {
		return "", fmt.Errorf("call itsm create ticket api failed, err: %v", err)
	}

	return sn, nil
}
