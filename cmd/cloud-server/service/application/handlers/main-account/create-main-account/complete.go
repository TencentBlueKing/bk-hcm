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
	hsproto "hcm/pkg/api/hc-service/main-account"
	"hcm/pkg/criteria/enumor"
	"hcm/pkg/iam/meta"
	"hcm/pkg/iam/sys"
	"hcm/pkg/logs"
	"hcm/pkg/thirdparty/api-gateway/cmsi"
)

// Complete complete the application by manual.
func (a *ApplicationOfCreateMainAccount) Complete() (enumor.ApplicationStatus, map[string]interface{}, error) {
	// Complete 的complete request 不能为空，其他情况可以为空
	if a.completeReq == nil {
		err := fmt.Errorf("complete request is nil cannot complete this application")
		return enumor.Delivering, map[string]interface{}{"error": err.Error()}, err
	}
	// 验证complete request
	if err := a.completeReq.Validate(); err != nil {
		return enumor.Delivering, map[string]interface{}{"error": err.Error()}, err
	}
	// 验证complete request的vendor和create request的vendor是否匹配
	if a.completeReq.Vendor != a.req.Vendor {
		err := fmt.Errorf("complete request's vendor and create request's vendor not match")
		return enumor.Delivering, map[string]interface{}{"error": err.Error()}, err
	}

	// 验证一级账号是否有效
	rootAccount, err := a.Client.DataService().Global.RootAccount.GetBasicInfo(a.Cts.Kit, a.completeReq.RootAccountID)
	if err != nil {
		err := fmt.Errorf("cannot get root account info")
		return enumor.Delivering, map[string]interface{}{"error": err.Error()}, err
	}
	if rootAccount.Vendor != a.req.Vendor {
		err := fmt.Errorf("root account's vendor not match main account's vendor")
		return enumor.Delivering, map[string]interface{}{"error": err.Error()}, err
	}
	// 校验site是否匹配
	if enumor.MainAccountSiteType(rootAccount.Site) != a.req.Site {
		err := fmt.Errorf("root account's site(%s) not match main account's site (%s)",
			rootAccount.Site, a.req.Site)
		return enumor.Delivering, map[string]interface{}{"error": err.Error()}, err
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

	// 自动创建的账号请求里没有云上账号Id/CloudId，需要从数据库里查
	account, err := a.Client.DataService().Global.MainAccount.GetBasicInfo(a.Cts.Kit, accountID)
	if err != nil {
		err := fmt.Errorf("create account success, accountId: %s, "+
			"but get main account basic info failed, err: %v, rid: %s", accountID, err, a.Cts.Kit.Rid)
		return enumor.DeliverError, map[string]interface{}{"error": err.Error()}, err
	}
	req := &meta.RegisterResCreatorActionInst{
		Type: string(sys.MainAccount),
		ID:   accountID,
		Name: account.CloudID,
	}
	if err = a.authorizer.RegisterResourceCreatorAction(a.Cts.Kit, req); err != nil {
		err := fmt.Errorf("create account success, accountId: %s, "+
			"but add create action associate permissions failed, err: %v, rid: %s", accountID, err, a.Cts.Kit.Rid)
		logs.Errorf(err.Error())
		return enumor.DeliverError, map[string]interface{}{"error": err}, err
	}
	//  异步发送邮件通知用户
	go a.sendMail(account)

	// 交付成功，记录交付的账号ID
	return enumor.Completed, map[string]interface{}{"account_id": accountID, "cloud_account_name": account.Name,
		"cloud_account_id": account.CloudID}, nil
}

func (a *ApplicationOfCreateMainAccount) createForAws(rootAccount *protocore.BaseRootAccount) (string, error) {
	req := a.req

	accountResp, err := a.Client.HCService().Aws.MainAccount.Create(a.Cts.Kit, &hsproto.CreateAwsMainAccountReq{
		RootAccountID:    rootAccount.ID,
		Email:            req.Email,
		CloudAccountName: req.Extension[req.Vendor.GetMainAccountNameFieldName()],
	})
	if err != nil {
		return "", fmt.Errorf("create aws main account [%s] failed, err: %v, rid: %s",
			req.Extension[req.Vendor.GetMainAccountNameFieldName()], err, a.Cts.Kit.Rid)
	}

	// create aws main account in dbs
	extension := &dataproto.AwsMainAccountExtensionCreateReq{
		CloudMainAccountName: accountResp.AccountName,
		CloudMainAccountID:   accountResp.AccountID,
	}
	extension.EncryptSecretKey(a.Cipher)

	result, err := a.Client.DataService().Aws.MainAccount.Create(
		a.Cts.Kit,
		&dataproto.MainAccountCreateReq[dataproto.AwsMainAccountExtensionCreateReq]{
			Name:              accountResp.AccountName,
			CloudID:           accountResp.AccountID,
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
		return "", fmt.Errorf("create aws main account in db failed, cloud_id: %s, err: %v, rid: %s",
			accountResp.AccountID, err, a.Cts.Kit.Rid)
	}

	return result.ID, nil
}

func (a *ApplicationOfCreateMainAccount) createForGcp(rootAccount *protocore.BaseRootAccount) (string, error) {
	req := a.req

	// create gcp main account auto
	fullRootAccount, err := a.Client.DataService().Gcp.RootAccount.Get(a.Cts.Kit, rootAccount.ID)
	if err != nil {
		return "", err
	}

	billingAccount := fullRootAccount.Extension.CloudBillingAccount
	organization := fullRootAccount.Extension.CloudOrganization
	if billingAccount == "" || organization == "" {
		return "", fmt.Errorf("root account [%s] not have billing account or organization", rootAccount.ID)
	}

	accountResp, err := a.Client.HCService().Gcp.MainAccount.Create(a.Cts.Kit, &hsproto.CreateGcpMainAccountReq{
		RootAccountID:       a.completeReq.RootAccountID,
		Email:               req.Email,
		ProjectName:         req.Extension[req.Vendor.GetMainAccountNameFieldName()],
		CloudBillingAccount: billingAccount,
		CloudOrganization:   organization,
	})
	if err != nil {
		return "", fmt.Errorf("create gcp main account [%s] failed, err: %v, rid: %s",
			req.Extension[req.Vendor.GetMainAccountNameFieldName()], err, a.Cts.Kit.Rid)
	}

	// create gcp main account in dbs
	extension := &dataproto.GcpMainAccountExtensionCreateReq{
		CloudProjectID:   accountResp.ProjectID,
		CloudProjectName: accountResp.ProjectName,
	}
	extension.EncryptSecretKey(a.Cipher)

	result, err := a.Client.DataService().Gcp.MainAccount.Create(
		a.Cts.Kit,
		&dataproto.MainAccountCreateReq[dataproto.GcpMainAccountExtensionCreateReq]{
			Name:              accountResp.ProjectName,
			CloudID:           accountResp.ProjectID,
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
		return "", fmt.Errorf("create gcp main account in db failed, cloud_id: %s, err: %v, rid: %s",
			accountResp.ProjectID, err, a.Cts.Kit.Rid)
	}

	return result.ID, nil
}

func (a *ApplicationOfCreateMainAccount) createForAzure(rootAccount *protocore.BaseRootAccount) (string, error) {
	req := a.req
	comReq := a.completeReq

	extension := &dataproto.AzureMainAccountExtensionCreateReq{
		CloudSubscriptionID:   comReq.Extension[a.Vendor().GetMainAccountIDFieldName()],
		CloudSubscriptionName: comReq.Extension[a.Vendor().GetMainAccountNameFieldName()],
		CloudInitPassword:     comReq.Extension[a.Vendor().GetMainAccountInitPasswordFieldName()],
	}
	extension.EncryptSecretKey(a.Cipher)

	result, err := a.Client.DataService().Azure.MainAccount.Create(
		a.Cts.Kit,
		&dataproto.MainAccountCreateReq[dataproto.AzureMainAccountExtensionCreateReq]{
			Name:              a.completeReq.Extension[a.Vendor().GetMainAccountNameFieldName()],
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
	comReq := a.completeReq

	extension := &dataproto.HuaWeiMainAccountExtensionCreateReq{
		CloudMainAccountID:   comReq.Extension[a.Vendor().GetMainAccountIDFieldName()],
		CloudMainAccountName: comReq.Extension[a.Vendor().GetMainAccountNameFieldName()],
		CloudInitPassword:    comReq.Extension[a.Vendor().GetMainAccountInitPasswordFieldName()],
	}
	extension.EncryptSecretKey(a.Cipher)

	result, err := a.Client.DataService().HuaWei.MainAccount.Create(
		a.Cts.Kit,
		&dataproto.MainAccountCreateReq[dataproto.HuaWeiMainAccountExtensionCreateReq]{
			Name:              a.completeReq.Extension[a.Vendor().GetMainAccountNameFieldName()],
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
	comReq := a.completeReq

	extension := &dataproto.ZenlayerMainAccountExtensionCreateReq{
		CloudMainAccountID:   comReq.Extension[a.Vendor().GetMainAccountIDFieldName()],
		CloudMainAccountName: comReq.Extension[a.Vendor().GetMainAccountNameFieldName()],
		CloudInitPassword:    comReq.Extension[a.Vendor().GetMainAccountInitPasswordFieldName()],
	}
	extension.EncryptSecretKey(a.Cipher)

	result, err := a.Client.DataService().Zenlayer.MainAccount.Create(
		a.Cts.Kit,
		&dataproto.MainAccountCreateReq[dataproto.ZenlayerMainAccountExtensionCreateReq]{
			Name:              a.completeReq.Extension[a.Vendor().GetMainAccountNameFieldName()],
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
	comReq := a.completeReq

	extension := &dataproto.KaopuMainAccountExtensionCreateReq{
		CloudMainAccountID:   comReq.Extension[a.Vendor().GetMainAccountIDFieldName()],
		CloudMainAccountName: comReq.Extension[a.Vendor().GetMainAccountNameFieldName()],
		CloudInitPassword:    comReq.Extension[a.Vendor().GetMainAccountInitPasswordFieldName()],
	}
	extension.EncryptSecretKey(a.Cipher)

	result, err := a.Client.DataService().Kaopu.MainAccount.Create(
		a.Cts.Kit,
		&dataproto.MainAccountCreateReq[dataproto.KaopuMainAccountExtensionCreateReq]{
			Name:              a.completeReq.Extension[a.Vendor().GetMainAccountNameFieldName()],
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

func (a *ApplicationOfCreateMainAccount) sendMail(account *dataproto.MainAccountGetBaseResult) {
	if account == nil {
		logs.Errorf("send mail failed, account should not be nil when send email, rid: %s", a.Cts.Kit.Rid)
		return
	}

	var (
		loginUrl string
	)

	switch account.Vendor {
	case enumor.Aws:
		loginUrl = AwsLoginAddress
	case enumor.Gcp:
		loginUrl = fmt.Sprintf(GcpLoginAddress, account.CloudID)
	case enumor.HuaWei:
		loginUrl = HuaweiLoginAddress
	case enumor.Azure:
		loginUrl = AzureLoginAddress
	case enumor.Zenlayer:
		loginUrl = ZenlayerLoginAddress
	case enumor.Kaopu:
		loginUrl = KaopuLoginAddress
	default:
		logs.Errorf("send mail failed, unknown vendor: %s, rid: %s", account.Vendor, a.Cts.Kit.Rid)
		return
	}

	mail := &cmsi.CmsiMail{
		Receiver: a.req.Email,
		Title:    fmt.Sprintf(EmailTitleTemplate, account.Vendor.GetNameZh()),
		Content: fmt.Sprintf(EmailContentTemplate,
			account.Vendor.GetNameZh(),
			account.Name,
			account.CloudID,
			loginUrl,
			loginUrl,
		),
	}

	err := a.SendMail(mail)
	if err != nil {
		logs.Errorf("send email failed for main account, id: %s, cloud_id: %s, name: %s, err: %v, rid: %s",
			account.ID, account.CloudID, account.Name, err, a.Cts.Kit.Rid)
		return
	}

	logs.Infof("send email success for main account, id: %s, cloud_id: %s, name: %s, rid: %s", account.ID,
		account.CloudID, account.Name, a.Cts.Kit.Rid)
}
