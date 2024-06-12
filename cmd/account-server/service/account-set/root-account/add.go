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

package rootaccount

import (
	"fmt"

	proto "hcm/pkg/api/account-server/account-set"
	"hcm/pkg/api/core"
	dataproto "hcm/pkg/api/data-service/account-set"
	"hcm/pkg/criteria/enumor"
	"hcm/pkg/criteria/errf"
	"hcm/pkg/dal/dao/tools"
	"hcm/pkg/iam/meta"
	"hcm/pkg/iam/sys"
	"hcm/pkg/logs"
	"hcm/pkg/rest"
)

// Add get main account with options
func (s *service) Add(cts *rest.Contexts) (interface{}, error) {
	req := new(proto.RootAccountAddReq)
	if err := cts.DecodeInto(req); err != nil {
		return nil, errf.NewFromErr(errf.DecodeRequestFailed, err)
	}
	if err := req.Validate(); err != nil {
		return nil, errf.NewFromErr(errf.InvalidParameter, err)
	}

	// 检查账号名是否重复
	if err := s.isDuplicateName(cts, req.Name); err != nil {
		return nil, errf.NewFromErr(errf.InvalidParameter, err)
	}

	// 检查资源账号的主账号是否重复
	mainAccountIDFieldName := req.Vendor.GetMainAccountIDField()
	mainAccountIDFieldValue := req.Extension[mainAccountIDFieldName]
	if err := CheckDuplicateRootAccount(cts, s.client, req.Vendor, mainAccountIDFieldValue); err != nil {
		return nil, errf.NewFromErr(errf.InvalidParameter, err)
	}

	// authorize
	authRes := meta.ResourceAttribute{Basic: &meta.Basic{Type: meta.RootAccount, Action: meta.Import}}
	err := s.authorizer.AuthorizeWithPerm(cts.Kit, authRes)
	if err != nil {
		return nil, err
	}

	var accountID string
	switch req.Vendor {
	case enumor.Aws:
		accountID, err = s.addForAws(cts, req)
	case enumor.Gcp:
		accountID, err = s.addForGcp(cts, req)
	case enumor.Azure:
		accountID, err = s.addForAzure(cts, req)
	case enumor.HuaWei:
		accountID, err = s.addForHuaWei(cts, req)
	case enumor.Zenlayer:
		accountID, err = s.addForZenlayer(cts, req)
	case enumor.Kaopu:
		accountID, err = s.addForKaopu(cts, req)
	}
	if err != nil {
		logs.Errorf("add root account for [%s] failed, err: %v, rid: %s", req.Vendor, err, cts.Kit.Rid)
		return nil, errf.NewFromErr(errf.InvalidParameter, err)
	}

	// 授予创建者创建资源默认附加权限
	iamReq := &meta.RegisterResCreatorActionInst{
		Type: string(sys.RootAccount),
		ID:   accountID,
		Name: req.Name,
	}

	if err = s.authorizer.RegisterResourceCreatorAction(cts.Kit, iamReq); err != nil {
		logs.Errorf("create account success, "+
			"but add create action associate permissions failed, err: %v, rid: %s", err, cts.Kit.Rid)
		return accountID, nil
	}

	return accountID, nil
}

func (s *service) isDuplicateName(cts *rest.Contexts, name string) error {
	// TODO: 后续需要解决并发问题
	// 后台查询是否主账号重复
	result, err := s.client.DataService().Global.RootAccount.List(
		cts.Kit,
		&core.ListWithoutFieldReq{
			Filter: tools.ExpressionAnd(tools.RuleEqual("name", name)),
			Page: &core.BasePage{
				Count: true,
			},
		},
	)
	if err != nil {
		return err
	}

	if result.Count > 0 {
		return fmt.Errorf("root account name [%s] has already exits, should be not duplicate", name)
	}

	return nil
}

func (s *service) addForAws(cts *rest.Contexts, req *proto.RootAccountAddReq) (string, error) {
	result, err := s.client.DataService().Aws.RootAccount.Create(
		cts.Kit,
		&dataproto.RootAccountCreateReq[dataproto.AwsRootAccountExtensionCreateReq]{
			Name:        req.Name,
			CloudID:     req.Extension["cloud_account_id"],
			Email:       req.Email,
			Managers:    req.Managers,
			BakManagers: req.BakManagers,
			Site:        req.Site,
			DeptID:      req.DeptID,
			Memo:        req.Memo,
			Extension: &dataproto.AwsRootAccountExtensionCreateReq{
				CloudAccountID:   req.Extension["cloud_account_id"],
				CloudIamUsername: req.Extension["cloud_iam_username"],
				CloudSecretID:    req.Extension["cloud_secret_id"],
				CloudSecretKey:   req.Extension["cloud_secret_key"],
			},
		},
	)
	if err != nil {
		return "", err
	}
	return result.ID, err
}

func (s *service) addForGcp(cts *rest.Contexts, req *proto.RootAccountAddReq) (string, error) {
	result, err := s.client.DataService().Gcp.RootAccount.Create(
		cts.Kit,
		&dataproto.RootAccountCreateReq[dataproto.GcpRootAccountExtensionCreateReq]{
			Name:        req.Name,
			CloudID:     req.Extension["cloud_project_id"],
			Email:       req.Email,
			Managers:    req.Managers,
			BakManagers: req.BakManagers,
			Site:        req.Site,
			DeptID:      req.DeptID,
			Memo:        req.Memo,
			Extension: &dataproto.GcpRootAccountExtensionCreateReq{
				CloudProjectID:          req.Extension["cloud_project_id"],
				CloudProjectName:        req.Extension["cloud_project_name"],
				CloudServiceAccountID:   req.Extension["cloud_service_account_id"],
				CloudServiceAccountName: req.Extension["cloud_service_account_name"],
				CloudServiceSecretID:    req.Extension["cloud_service_secret_id"],
				CloudServiceSecretKey:   req.Extension["cloud_service_secret_key"],
			},
		},
	)
	if err != nil {
		return "", err
	}
	return result.ID, err
}

func (s *service) addForAzure(cts *rest.Contexts, req *proto.RootAccountAddReq) (string, error) {
	result, err := s.client.DataService().Azure.RootAccount.Create(
		cts.Kit,
		&dataproto.RootAccountCreateReq[dataproto.AzureRootAccountExtensionCreateReq]{
			Name:        req.Name,
			CloudID:     req.Extension["cloud_subscription_id"],
			Email:       req.Email,
			Managers:    req.Managers,
			BakManagers: req.BakManagers,
			Site:        req.Site,
			DeptID:      req.DeptID,
			Memo:        req.Memo,
			Extension: &dataproto.AzureRootAccountExtensionCreateReq{
				CloudTenantID:         req.Extension["cloud_tenant_id"],
				CloudSubscriptionID:   req.Extension["cloud_subscription_id"],
				CloudSubscriptionName: req.Extension["cloud_subscription_name"],
				CloudApplicationID:    req.Extension["cloud_application_id"],
				CloudApplicationName:  req.Extension["cloud_application_name"],
				CloudClientSecretKey:  req.Extension["cloud_client_secret_key"],
			},
		},
	)
	if err != nil {
		return "", err
	}
	return result.ID, err
}

func (s *service) addForHuaWei(cts *rest.Contexts, req *proto.RootAccountAddReq) (string, error) {
	result, err := s.client.DataService().HuaWei.RootAccount.Create(
		cts.Kit,
		&dataproto.RootAccountCreateReq[dataproto.HuaWeiRootAccountExtensionCreateReq]{
			Name:        req.Name,
			CloudID:     req.Extension["cloud_sub_account_id"],
			Email:       req.Email,
			Managers:    req.Managers,
			BakManagers: req.BakManagers,
			Site:        req.Site,
			DeptID:      req.DeptID,
			Memo:        req.Memo,
			Extension: &dataproto.HuaWeiRootAccountExtensionCreateReq{
				CloudSubAccountID:   req.Extension["cloud_sub_account_id"],
				CloudSubAccountName: req.Extension["cloud_sub_account_name"],
				CloudSecretID:       req.Extension["cloud_secret_id"],
				CloudSecretKey:      req.Extension["cloud_secret_key"],
				CloudIamUserID:      req.Extension["cloud_iam_user_id"],
				CloudIamUsername:    req.Extension["cloud_iam_username"],
			},
		},
	)
	if err != nil {
		return "", err
	}
	return result.ID, err
}

func (s *service) addForZenlayer(cts *rest.Contexts, req *proto.RootAccountAddReq) (string, error) {
	result, err := s.client.DataService().Zenlayer.RootAccount.Create(
		cts.Kit,
		&dataproto.RootAccountCreateReq[dataproto.ZenlayerRootAccountExtensionCreateReq]{
			Name:        req.Name,
			CloudID:     req.Extension["cloud_account_id"],
			Email:       req.Email,
			Managers:    req.Managers,
			BakManagers: req.BakManagers,
			Site:        req.Site,
			DeptID:      req.DeptID,
			Memo:        req.Memo,
			Extension: &dataproto.ZenlayerRootAccountExtensionCreateReq{
				CloudAccountID: req.Extension["cloud_account_id"],
			},
		},
	)
	if err != nil {
		return "", err
	}
	return result.ID, err
}

func (s *service) addForKaopu(cts *rest.Contexts, req *proto.RootAccountAddReq) (string, error) {
	result, err := s.client.DataService().Kaopu.RootAccount.Create(
		cts.Kit,
		&dataproto.RootAccountCreateReq[dataproto.KaopuRootAccountExtensionCreateReq]{
			Name:        req.Name,
			CloudID:     req.Extension["cloud_account_id"],
			Email:       req.Email,
			Managers:    req.Managers,
			BakManagers: req.BakManagers,
			Site:        req.Site,
			DeptID:      req.DeptID,
			Memo:        req.Memo,
			Extension: &dataproto.KaopuRootAccountExtensionCreateReq{
				CloudAccountID: req.Extension["cloud_account_id"],
			},
		},
	)
	if err != nil {
		return "", err
	}
	return result.ID, err
}
