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

package mainaccount

import (
	"fmt"

	protocore "hcm/pkg/api/core/account-set"
	dataproto "hcm/pkg/api/data-service/account-set"
	"hcm/pkg/criteria/enumor"
	"hcm/pkg/iam/meta"
	"hcm/pkg/iam/sys"
	"hcm/pkg/logs"
)

// Complete complete the application by manual.
func (a *ApplicationOfCreateMainAccount) Complete() (enumor.ApplicationStatus, map[string]interface{}, error) {
	// Complete 的complete request 不能为空，其他情况可以为空
	if a.completeReq == nil {
		err := fmt.Errorf("complete request is nil cannot complete this application")
		return enumor.DeliverError, map[string]interface{}{"error": err.Error()}, err
	}

	// 验证complete request
	if err := a.completeReq.Validate(); err != nil {
		return enumor.DeliverError, map[string]interface{}{"error": err.Error()}, err
	}

	// 验证complete request的vendor和create request的vendor是否匹配
	if a.completeReq.Vendor != a.req.Vendor {
		err := fmt.Errorf("complete request's vendor and create request's vendor not match")
		return enumor.DeliverError, map[string]interface{}{"error": err.Error()}, err
	}

	// 验证一级账号是否有效
	rootAccount, err := a.Client.DataService().Global.RootAccount.GetBasicInfo(a.Cts.Kit, a.completeReq.RootAccountID)
	if err != nil {
		err := fmt.Errorf("cannot get root account info")
		return enumor.DeliverError, map[string]interface{}{"error": err.Error()}, err
	}

	var (
		accountID string
	)

	switch a.req.Vendor {
	case enumor.Aws:
		accountID, err = a.createForAws(&rootAccount.BaseRootAccount)
	case enumor.Gcp:
		accountID, err = a.createForGcp(&rootAccount.BaseRootAccount)
	case enumor.Azure:
		accountID, err = a.createForAzure(&rootAccount.BaseRootAccount)
	case enumor.HuaWei:
		accountID, err = a.createForHuaWei(&rootAccount.BaseRootAccount)
	case enumor.Zenlayer:
		accountID, err = a.createForZenlayer(&rootAccount.BaseRootAccount)
	case enumor.Kaopu:
		accountID, err = a.createForKaopu(&rootAccount.BaseRootAccount)
	}
	if err != nil {
		logs.Errorf("create main account for [%s] failed, err: %v, rid: %s", a.req.Vendor, err, a.Cts.Kit.Rid)
		return enumor.DeliverError, map[string]interface{}{"error": err.Error()}, err
	}

	req := &meta.RegisterResCreatorActionInst{
		Type: string(sys.MainAccount),
		ID:   accountID,
		Name: a.req.Extension[a.req.Vendor.GetMainAccountIDFieldName()],
	}
	if err = a.authorizer.RegisterResourceCreatorAction(a.Cts.Kit, req); err != nil {
		return enumor.DeliverError, map[string]interface{}{"error": fmt.Sprintf("create account success, "+
			"but add create action associate permissions failed, err: %v", err)}, err
	}

	// todo 异步发送邮件通知用户
	go a.sendEmail(accountID)

	// 交付成功，记录交付的账号ID
	return enumor.Completed, map[string]interface{}{"account_id": accountID}, nil
}

func (a *ApplicationOfCreateMainAccount) createForAws(rootAccount *protocore.BaseRootAccount) (string, error) {
	//todo 待添加自动化创建流程
	return "", fmt.Errorf("aws not implemented for create account auto")
}

func (a *ApplicationOfCreateMainAccount) createForGcp(rootAccount *protocore.BaseRootAccount) (string, error) {
	//todo 待添加自动化创建流程
	return "", fmt.Errorf("gcp not implemented for create account auto")
}

func (a *ApplicationOfCreateMainAccount) createForAzure(rootAccount *protocore.BaseRootAccount) (string, error) {
	req := a.req

	extension := &dataproto.AzureMainAccountExtensionCreateReq{
		CloudSubscriptionID:   req.Extension[a.Vendor().GetMainAccountIDFieldName()],
		CloudSubscriptionName: req.Extension[a.Vendor().GetMainAccountNameFieldName()],
		CloudInitPassword:     req.Extension[a.Vendor().GetMainAccountInitPasswordFieldName()],
	}
	extension.EncryptSecretKey(a.Cipher)

	result, err := a.Client.DataService().Azure.MainAccount.Create(
		a.Cts.Kit,
		&dataproto.MainAccountCreateReq[dataproto.AzureMainAccountExtensionCreateReq]{
			CloudID:           a.completeReq.Extension[a.Vendor().GetMainAccountIDFieldName()],
			Email:             req.Email,
			Managers:          req.Managers,
			BakManagers:       req.BakManagers,
			Site:              req.Site,
			BusinessType:      req.BusinessType,
			Status:            enumor.MainAccountStatusRUNNING,
			ParentAccountName: rootAccount.Name,
			ParentAccountID:   rootAccount.ID,
			DeptID:            req.DeptID,
			BkBizID:           req.BkBizID,
			OpProductID:       req.OpProductID,
			Memo:              req.Memo,
			Extension:         extension,
		},
	)
	if err != nil {
		return "", err
	}

	return result.ID, nil
}

func (a *ApplicationOfCreateMainAccount) createForHuaWei(rootAccount *protocore.BaseRootAccount) (string, error) {
	req := a.req

	extension := &dataproto.HuaWeiMainAccountExtensionCreateReq{
		CloudMainAccountID:   req.Extension[a.Vendor().GetMainAccountIDFieldName()],
		CloudMainAccountName: req.Extension[a.Vendor().GetMainAccountNameFieldName()],
		CloudInitPassword:    req.Extension[a.Vendor().GetMainAccountInitPasswordFieldName()],
	}
	extension.EncryptSecretKey(a.Cipher)

	result, err := a.Client.DataService().HuaWei.MainAccount.Create(
		a.Cts.Kit,
		&dataproto.MainAccountCreateReq[dataproto.HuaWeiMainAccountExtensionCreateReq]{
			CloudID:           a.completeReq.Extension[a.Vendor().GetMainAccountIDFieldName()],
			Email:             req.Email,
			Managers:          req.Managers,
			BakManagers:       req.BakManagers,
			Site:              req.Site,
			BusinessType:      req.BusinessType,
			Status:            enumor.MainAccountStatusRUNNING,
			ParentAccountName: rootAccount.Name,
			ParentAccountID:   rootAccount.ID,
			DeptID:            req.DeptID,
			BkBizID:           req.BkBizID,
			OpProductID:       req.OpProductID,
			Memo:              req.Memo,
			Extension:         extension,
		},
	)
	if err != nil {
		return "", err
	}

	return result.ID, nil
}

func (a *ApplicationOfCreateMainAccount) createForZenlayer(rootAccount *protocore.BaseRootAccount) (string, error) {
	req := a.req

	extension := &dataproto.ZenlayerMainAccountExtensionCreateReq{
		CloudMainAccountID:   req.Extension[a.Vendor().GetMainAccountIDFieldName()],
		CloudMainAccountName: req.Extension[a.Vendor().GetMainAccountNameFieldName()],
		CloudInitPassword:    req.Extension[a.Vendor().GetMainAccountInitPasswordFieldName()],
	}
	extension.EncryptSecretKey(a.Cipher)

	result, err := a.Client.DataService().Zenlayer.MainAccount.Create(
		a.Cts.Kit,
		&dataproto.MainAccountCreateReq[dataproto.ZenlayerMainAccountExtensionCreateReq]{
			CloudID:           a.completeReq.Extension[a.Vendor().GetMainAccountIDFieldName()],
			Email:             req.Email,
			Managers:          req.Managers,
			BakManagers:       req.BakManagers,
			Site:              req.Site,
			BusinessType:      req.BusinessType,
			Status:            enumor.MainAccountStatusRUNNING,
			ParentAccountName: rootAccount.Name,
			ParentAccountID:   rootAccount.ID,
			DeptID:            req.DeptID,
			BkBizID:           req.BkBizID,
			OpProductID:       req.OpProductID,
			Memo:              req.Memo,
			Extension:         extension,
		},
	)
	if err != nil {
		return "", err
	}

	return result.ID, nil
}

func (a *ApplicationOfCreateMainAccount) createForKaopu(rootAccount *protocore.BaseRootAccount) (string, error) {
	req := a.req

	extension := &dataproto.KaopuMainAccountExtensionCreateReq{
		CloudMainAccountID:   req.Extension[a.Vendor().GetMainAccountIDFieldName()],
		CloudMainAccountName: req.Extension[a.Vendor().GetMainAccountNameFieldName()],
		CloudInitPassword:    req.Extension[a.Vendor().GetMainAccountInitPasswordFieldName()],
	}
	extension.EncryptSecretKey(a.Cipher)

	result, err := a.Client.DataService().Kaopu.MainAccount.Create(
		a.Cts.Kit,
		&dataproto.MainAccountCreateReq[dataproto.KaopuMainAccountExtensionCreateReq]{
			CloudID:           a.completeReq.Extension[a.Vendor().GetMainAccountIDFieldName()],
			Email:             req.Email,
			Managers:          req.Managers,
			BakManagers:       req.BakManagers,
			Site:              req.Site,
			BusinessType:      req.BusinessType,
			Status:            enumor.MainAccountStatusRUNNING,
			ParentAccountName: rootAccount.Name,
			ParentAccountID:   rootAccount.ID,
			DeptID:            req.DeptID,
			BkBizID:           req.BkBizID,
			OpProductID:       req.OpProductID,
			Memo:              req.Memo,
			Extension:         extension,
		},
	)
	if err != nil {
		return "", err
	}

	return result.ID, nil
}

func (a *ApplicationOfCreateMainAccount) sendEmail(accountID string) {
	// todo
}
