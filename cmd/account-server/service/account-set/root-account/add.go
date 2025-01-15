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
	"hcm/pkg/logs"
	"hcm/pkg/rest"
)

// AddRootAccount get root account with options
func (s *service) AddRootAccount(cts *rest.Contexts) (interface{}, error) {
	req := new(proto.RootAccountAddReq)
	if err := cts.DecodeInto(req); err != nil {
		return nil, errf.NewFromErr(errf.DecodeRequestFailed, err)
	}
	if err := req.Validate(); err != nil {
		return nil, errf.NewFromErr(errf.InvalidParameter, err)
	}

	// 校验用户有一级账号管理权限
	if err := s.checkPermission(cts, meta.RootAccount, meta.Find); err != nil {
		return nil, err
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

	var accountID string
	var err error
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

	return accountID, nil
}

func (s *service) isDuplicateName(cts *rest.Contexts, name string) error {
	// TODO: 后续需要解决并发问题
	// 后台查询是否主账号重复
	listReq := &core.ListReq{
		Filter: tools.ExpressionAnd(tools.RuleEqual("name", name)),
		Page:   core.NewCountPage(),
	}
	result, err := s.client.DataService().Global.RootAccount.List(cts.Kit, listReq)
	if err != nil {
		return err
	}

	if result.Count > 0 {
		return fmt.Errorf("root account name [%s] has already exits, should be not duplicate", name)
	}

	return nil
}

func (s *service) addForAws(cts *rest.Contexts, req *proto.RootAccountAddReq) (string, error) {
	extension := &dataproto.AwsRootAccountExtensionCreateReq{
		CloudAccountID:   req.Extension["cloud_account_id"],
		CloudIamUsername: req.Extension["cloud_iam_username"],
		CloudSecretID:    req.Extension["cloud_secret_id"],
		CloudSecretKey:   req.Extension["cloud_secret_key"],
	}
	if err := extension.Validate(); err != nil {
		return "", err
	}
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
			Extension:   extension,
		},
	)
	if err != nil {
		return "", err
	}
	return result.ID, err
}

func (s *service) addForGcp(cts *rest.Contexts, req *proto.RootAccountAddReq) (string, error) {
	// extension 的email如果没有填写则使用req的email，如果extension的email填写了则要求req.email和extension的email一致
	email, ok := req.Extension["email"]
	if !ok || email == "" {
		email = req.Email
	}
	if email != req.Email {
		return "", fmt.Errorf("request email [%s] and extension email [%s] should be same", req.Email, email)
	}

	extension := &dataproto.GcpRootAccountExtensionCreateReq{
		Email:                   email,
		CloudProjectID:          req.Extension["cloud_project_id"],
		CloudProjectName:        req.Extension["cloud_project_name"],
		CloudServiceAccountID:   req.Extension["cloud_service_account_id"],
		CloudServiceAccountName: req.Extension["cloud_service_account_name"],
		CloudServiceSecretID:    req.Extension["cloud_service_secret_id"],
		CloudServiceSecretKey:   req.Extension["cloud_service_secret_key"],
	}
	if err := extension.Validate(); err != nil {
		return "", err
	}

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
			Extension:   extension,
		},
	)
	if err != nil {
		return "", err
	}
	return result.ID, err
}

func (s *service) addForAzure(cts *rest.Contexts, req *proto.RootAccountAddReq) (string, error) {

	extension := &dataproto.AzureRootAccountExtensionCreateReq{
		DisplayNameName:       req.Extension["display_name_name"],
		CloudTenantID:         req.Extension["cloud_tenant_id"],
		CloudSubscriptionID:   req.Extension["cloud_subscription_id"],
		CloudSubscriptionName: req.Extension["cloud_subscription_name"],
		CloudApplicationID:    req.Extension["cloud_application_id"],
		CloudApplicationName:  req.Extension["cloud_application_name"],
		CloudClientSecretKey:  req.Extension["cloud_client_secret_key"],
	}
	if err := extension.Validate(); err != nil {
		return "", err
	}
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
			Extension:   extension,
		},
	)
	if err != nil {
		return "", err
	}
	return result.ID, err
}

func (s *service) addForHuaWei(cts *rest.Contexts, req *proto.RootAccountAddReq) (string, error) {
	extension := &dataproto.HuaWeiRootAccountExtensionCreateReq{
		CloudSubAccountID:   req.Extension["cloud_sub_account_id"],
		CloudSubAccountName: req.Extension["cloud_sub_account_name"],
		CloudSecretID:       req.Extension["cloud_secret_id"],
		CloudSecretKey:      req.Extension["cloud_secret_key"],
		CloudIamUserID:      req.Extension["cloud_iam_user_id"],
		CloudIamUsername:    req.Extension["cloud_iam_username"],
	}
	if err := extension.Validate(); err != nil {
		return "", err
	}

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
			Extension:   extension,
		},
	)
	if err != nil {
		return "", err
	}
	return result.ID, err
}

func (s *service) addForZenlayer(cts *rest.Contexts, req *proto.RootAccountAddReq) (string, error) {
	extension := &dataproto.ZenlayerRootAccountExtensionCreateReq{
		CloudAccountID: req.Extension["cloud_account_id"],
	}
	if err := extension.Validate(); err != nil {
		return "", err
	}
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
			Extension:   extension,
		},
	)
	if err != nil {
		return "", err
	}
	return result.ID, err
}

func (s *service) addForKaopu(cts *rest.Contexts, req *proto.RootAccountAddReq) (string, error) {
	extension := &dataproto.KaopuRootAccountExtensionCreateReq{
		CloudAccountID: req.Extension["cloud_account_id"],
	}
	if err := extension.Validate(); err != nil {
		return "", err
	}
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
			Extension:   extension,
		},
	)
	if err != nil {
		return "", err
	}
	return result.ID, err
}
