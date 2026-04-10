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

package createsubaccount

import (
	"fmt"
	"strconv"

	typeaccount "hcm/pkg/adaptor/types/account"
	proto "hcm/pkg/api/cloud-server/application"
	"hcm/pkg/api/core"
	protocore "hcm/pkg/api/core/cloud"
	coresubaccount "hcm/pkg/api/core/cloud/sub-account"
	dataprotocloud "hcm/pkg/api/data-service/cloud"
	dssubaccount "hcm/pkg/api/data-service/cloud/sub-account"
	hssubaccount "hcm/pkg/api/hc-service/sub-account"
	"hcm/pkg/criteria/enumor"
	"hcm/pkg/dal/table/types"
	"hcm/pkg/logs"
	"hcm/pkg/thirdparty/api-gateway/cmsi"
	"hcm/pkg/tools/converter"
)

// Deliver execute resource delivery after approval.
func (a *ApplicationOfCreateSubAccount) Deliver() (enumor.ApplicationStatus, map[string]interface{}, error) {
	switch a.Vendor() {
	case enumor.TCloud:
		return a.deliverForTCloud()
	default:
		return enumor.DeliverError,
			map[string]interface{}{"error": fmt.Sprintf("vendor %s not supported", a.Vendor())},
			fmt.Errorf("vendor %s not supported for sub account creation", a.Vendor())
	}
}

func (a *ApplicationOfCreateSubAccount) deliverForTCloud() (enumor.ApplicationStatus, map[string]interface{}, error) {
	ext, err := decodeTCloudExtension(a)
	if err != nil {
		return enumor.DeliverError,
			map[string]interface{}{"error": fmt.Sprintf("decode tcloud extension failed, err: %v", err)}, err
	}

	cloudResult, err := a.createTCloudSubAccountInCloud(ext)
	if err != nil {
		return enumor.DeliverError,
			map[string]interface{}{"error": fmt.Sprintf("create cloud sub account failed, err: %v", err)}, err
	}

	parentAccount, err := a.Client.DataService().TCloud.Account.Get(a.Cts.Kit.Ctx, a.Cts.Kit.Header(), a.req.AccountID)
	if err != nil {
		return enumor.DeliverError,
			map[string]interface{}{"error": fmt.Sprintf("get parent account failed, err: %v", err)}, err
	}

	cloudID := strconv.FormatUint(converter.PtrToVal(cloudResult.Uin), 10)
	subAccountIDs, accountID, err := a.saveLocalSubAccount(cloudResult, ext, parentAccount)
	if err != nil {
		logs.Errorf("cloud sub account created (uin=%s) but local persistence failed, err: %v, rid: %s", cloudID,
			err, a.Cts.Kit.Rid)
		return enumor.DeliverError,
			map[string]interface{}{"error": fmt.Sprintf("save sub account/account to db failed, err: %v", err),
				"cloud_id": converter.PtrToVal(cloudResult.Uin)}, err
	}

	if err := a.sendSubAccountMail(cloudResult); err != nil {
		logs.Errorf("cloud sub account created (uin=%s) but send secret mail failed, err: %v, rid: %s",
			cloudID, err, a.Cts.Kit.Rid)
		return enumor.DeliverError,
			map[string]interface{}{"error": fmt.Sprintf("send secret mail failed, err: %v", err),
				"cloud_id": cloudID}, err
	}

	return enumor.Completed, map[string]interface{}{"sub_account_ids": subAccountIDs, "account_id": accountID,
		"cloud_id": cloudID,
	}, nil
}

// tcloudCreateCloudResult aggregates cloud API results during sub account creation.
type tcloudCreateCloudResult struct {
	hssubaccount.TCloudCreateSubAccountResult
	SafeAuth   *typeaccount.SafeAuthFlagResult
	CreateTime *string
}

// createTCloudSubAccountInCloud creates the subaccount on Tencent Cloud, queries its detail
// and best-effort loads safe auth flags.
func (a *ApplicationOfCreateSubAccount) createTCloudSubAccountInCloud(ext *proto.TCloudSubAccountAddExtension,
) (*tcloudCreateCloudResult, error) {

	cloudResult, err := a.Client.HCService().TCloud.Account.CreateSubAccount(
		a.Cts.Kit,
		&hssubaccount.TCloudCreateSubAccountReq{
			AccountID:    a.req.AccountID,
			Name:         a.req.Name,
			Email:        a.req.Email,
			PhoneNum:     a.req.PhoneNum,
			CountryCode:  a.req.CountryCode,
			ConsoleLogin: ext.ConsoleLogin,
		},
	)
	if err != nil {
		logs.Errorf("create tcloud sub account (%s) failed, err: %v, rid: %s", a.req.Name, err, a.Cts.Kit.Rid)
		return nil, fmt.Errorf("create tcloud sub account (%s) failed, err: %v", a.req.Name, err)
	}

	uin := converter.PtrToVal(cloudResult.Uin)
	err = a.Client.HCService().TCloud.Account.SetMfaFlag(a.Cts.Kit, &hssubaccount.TCloudSetMfaFlagReq{
		AccountID:  a.req.AccountID,
		OpUin:      uin,
		LoginFlag:  &typeaccount.LoginActionFlag{Stoken: converter.ValToPtr(uint64(1))},
		ActionFlag: &typeaccount.LoginActionFlag{Stoken: converter.ValToPtr(uint64(1))},
	})
	if err != nil {
		logs.Errorf("set mfa flag for sub account (%s) failed, err: %v, rid: %s", a.req.Name, err, a.Cts.Kit.Rid)
		return nil, fmt.Errorf("set mfa flag for sub account (%s) failed, err: %v", a.req.Name, err)
	}

	subAccounts, err := a.Client.HCService().TCloud.Account.DescribeSubAccounts(
		a.Cts.Kit,
		&hssubaccount.TCloudDescribeSubAccountsReq{AccountID: a.req.AccountID, SubUin: []uint64{uin}},
	)
	if err != nil {
		logs.Errorf("describe sub accounts for sub account (%s) failed, err: %v, rid: %s",
			a.req.Name, err, a.Cts.Kit.Rid)
		return nil, fmt.Errorf("describe sub accounts for sub account (%s) failed, err: %v", a.req.Name, err)
	}
	if len(subAccounts) != 1 {
		logs.Errorf("sub account count is not 1, uin=%d, name=%s, count=%d, rid: %s",
			uin, a.req.Name, len(subAccounts), a.Cts.Kit.Rid)
		return nil, fmt.Errorf("sub account count is not 1, uin=%d, name=%s, count=%d",
			uin, a.req.Name, len(subAccounts))
	}

	safeAuthFlag, err := a.Client.HCService().TCloud.Account.DescribeSafeAuthFlag(
		a.Cts.Kit, &hssubaccount.TCloudDescribeSafeAuthFlagReq{AccountID: a.req.AccountID, SubUin: uin},
	)
	if err != nil {
		logs.Errorf("sub account created (uin=%d, name=%s) but get safe auth flag failed, err: %v, rid: %s",
			uin, a.req.Name, err, a.Cts.Kit.Rid)
		return nil, fmt.Errorf("get safe auth flag failed, err: %v", err)
	}

	result := &tcloudCreateCloudResult{
		TCloudCreateSubAccountResult: converter.PtrToVal(cloudResult),
		SafeAuth:                     safeAuthFlag,
		CreateTime:                   subAccounts[0].CreateTime,
	}

	return result, nil
}

func (a *ApplicationOfCreateSubAccount) registerAccountForTCloud(cloudID string, createResult *tcloudCreateCloudResult,
	parentAccount *dataprotocloud.AccountGetResult[protocore.TCloudAccountExtension]) (string, error) {

	result, err := a.Client.DataService().TCloud.Account.Create(
		a.Cts.Kit.Ctx,
		a.Cts.Kit.Header(),
		&dataprotocloud.AccountCreateReq[dataprotocloud.TCloudAccountExtensionCreateReq]{
			Name:           a.req.Name,
			Managers:       a.req.Managers,
			Type:           enumor.RegistrationAccount,
			Site:           parentAccount.Site,
			Memo:           a.req.Memo,
			BkBizID:        a.BkBizID(),
			CloudCreatedAt: createResult.CreateTime,
			UsageBizIDs:    []int64{a.BkBizID()},
			Extension: &dataprotocloud.TCloudAccountExtensionCreateReq{
				CloudMainAccountID: parentAccount.Extension.CloudMainAccountID,
				CloudSubAccountID:  cloudID,
				CloudSecretID:      createResult.SecretID,
				CloudSecretKey:     createResult.SecretKey,
			},
		},
	)
	if err != nil {
		logs.Errorf("register account for tcloud failed, err: %v, rid: %s", err, a.Cts.Kit.Rid)
		return "", fmt.Errorf("register account for tcloud failed, err: %v", err)
	}

	return result.ID, nil
}

func (a *ApplicationOfCreateSubAccount) saveLocalSubAccount(cloudResult *tcloudCreateCloudResult,
	ext *proto.TCloudSubAccountAddExtension,
	parentAccount *dataprotocloud.AccountGetResult[protocore.TCloudAccountExtension],
) ([]string, string, error) {

	cloudID := strconv.FormatUint(converter.PtrToVal(cloudResult.Uin), 10)

	if ext == nil {
		return nil, "", fmt.Errorf("extension is required")
	}

	var loginProt, actionProt enumor.AccountProtectionFlag
	if cloudResult.SafeAuth != nil {
		if cloudResult.SafeAuth.LoginFlag != nil {
			loginProt = cloudResult.SafeAuth.LoginFlag.ToProtectionFlag()
		}
		if cloudResult.SafeAuth.ActionFlag != nil {
			actionProt = cloudResult.SafeAuth.ActionFlag.ToProtectionFlag()
		}
	}

	tCloudExt := &coresubaccount.TCloudExtension{
		CloudMainAccountID: parentAccount.Extension.CloudMainAccountID,
		Uin:                cloudResult.Uin,
		NickName:           cloudResult.Name,
		CreateTime:         cloudResult.CreateTime,
		LoginFlag:          converter.ValToPtr(loginProt),
		ActionFlag:         converter.ValToPtr(actionProt),
		ConsoleLogin:       ext.ConsoleLogin,
	}
	extBytes, err := core.MarshalStruct(tCloudExt)
	if err != nil {
		return nil, "", fmt.Errorf("marshal extension failed, err: %v", err)
	}

	createResult, err := a.Client.DataService().Global.SubAccount.BatchCreate(
		a.Cts.Kit,
		&dssubaccount.CreateReq{
			Items: []dssubaccount.CreateField{
				{
					CloudID:   cloudID,
					Name:      a.req.Name,
					Vendor:    a.Vendor(),
					Site:      parentAccount.Site,
					AccountID: a.req.AccountID,
					Managers:  a.req.Managers,
					BkBizIDs:  types.Int64Array{a.BkBizID()},
					// 创建的三级账号为CurrentAccount类型
					AccountType: string(enumor.CurrentAccount),
					Email:       converter.ValToPtr(a.req.Email),
					PhoneNum:    converter.ValToPtr(a.req.PhoneNum),
					Memo:        a.req.Memo,
					Extension:   extBytes,
				},
			},
		},
	)
	if err != nil {
		return nil, "", err
	}

	// registerAccountForTCloud 将用户创建的三级账号登记到account表，防止触发HCM未纳管该账号的安全工单
	accountID, err := a.registerAccountForTCloud(cloudID, cloudResult, parentAccount)
	if err != nil {
		return nil, "", err
	}

	return createResult.IDs, accountID, nil
}

func (a *ApplicationOfCreateSubAccount) sendSubAccountMail(result *tcloudCreateCloudResult) error {
	if a.req.ReceiveEmail == "" {
		logs.Errorf("send secret mail failed, receive email is empty, rid: %s", a.Cts.Kit.Rid)
		return fmt.Errorf("send secret mail failed, receive email is empty")
	}

	content := fmt.Sprintf("您的三级账号已创建成功.\n\n账号名称: %s", converter.PtrToVal(result.Name))

	if result.SecretID != "" {
		content += fmt.Sprintf("\nSecretId: %s", result.SecretID)
	}
	if result.SecretKey != "" {
		content += fmt.Sprintf("\nSecretKey: %s", result.SecretKey)
	}
	if result.Password != "" {
		content += fmt.Sprintf("\n密码: %s", result.Password)
	}

	err := a.SendMail(&cmsi.CmsiMail{
		Receiver:   a.req.ReceiveEmail,
		Title:      fmt.Sprintf("三级账号(%s)开通通知", converter.PtrToVal(result.Name)),
		Content:    content,
		BodyFormat: "Text",
	})
	if err != nil {
		logs.Errorf("send secret mail to %s failed, err: %v, rid: %s", a.req.ReceiveEmail, err, a.Cts.Kit.Rid)
		return fmt.Errorf("send secret mail to %s failed, err: %v", a.req.ReceiveEmail, err)
	}

	return nil
}
