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
	"hcm/pkg/api/core"
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

// deliverError is a helper to build a DeliverError result consistently.
func deliverError(msg string, err error) (enumor.ApplicationStatus, map[string]interface{}, error) {
	return enumor.DeliverError, map[string]interface{}{"error": msg}, err
}

// Deliver execute resource delivery after approval.
func (a *ApplicationOfCreateSubAccount) Deliver() (enumor.ApplicationStatus, map[string]interface{}, error) {
	switch a.Vendor() {
	case enumor.TCloud:
		return a.deliverForTCloud()
	default:
		return deliverError(
			fmt.Sprintf("vendor %s not supported", a.Vendor()),
			fmt.Errorf("vendor %s not supported for sub account creation", a.Vendor()),
		)
	}
}

func (a *ApplicationOfCreateSubAccount) deliverForTCloud() (enumor.ApplicationStatus, map[string]interface{}, error) {
	ext, err := decodeTCloudExtension(a)
	if err != nil {
		return deliverError(fmt.Sprintf("decode tcloud extension failed, err: %v", err), err)
	}

	cloudResult, err := a.Client.HCService().TCloud.Account.CreateSubAccount(
		a.Cts.Kit,
		&hssubaccount.CreateSubAccountReq{
			AccountID:    a.req.AccountID,
			Name:         a.req.Name,
			Email:        a.req.Email,
			PhoneNum:     a.req.PhoneNum,
			CountryCode:  a.req.CountryCode,
			ConsoleLogin: ext.ConsoleLogin,
		},
	)
	if err != nil {
		return deliverError(fmt.Sprintf("create cloud sub account failed, err: %v", err), err)
	}

	cloudID := strconv.FormatUint(converter.PtrToVal(cloudResult.Uin), 10)

	safeAuthFlag, err := a.Client.HCService().TCloud.Account.DescribeSafeAuthFlag(
		a.Cts.Kit,
		&hssubaccount.DescribeSafeAuthFlagReq{
			AccountID: a.req.AccountID,
			SubUin:    converter.PtrToVal(cloudResult.Uin),
		},
	)
	if err != nil {
		// 获取安全配置失败，不应该影响创建子账号的流程
		logs.Warnf(
			"sub account created (uin=%s) but get safe auth flag failed, err: %v, rid: %s",
			cloudID, err, a.Cts.Kit.Rid,
		)
	}

	tCloudExt := &coresubaccount.TCloudExtension{
		CloudMainAccountID: a.req.AccountID,
		Uin:                cloudResult.Uin,
		NickName:           cloudResult.Name,
		LoginFlag:          loginActionFlagToProtectionFlag(safeAuthFlag.LoginFlag),
		ActionFlag:         loginActionFlagToProtectionFlag(safeAuthFlag.ActionFlag),
		ConsoleLogin:       ext.ConsoleLogin,
	}
	// JETTTODO: 开发密钥相关功能后，密钥需要保存到DB中
	extBytes, err := core.MarshalStruct(tCloudExt)
	if err != nil {
		return deliverError(fmt.Sprintf("marshal extension failed, err: %v", err), err)
	}

	subAccountIDs, err := a.saveSubAccountToDB(cloudID, extBytes)
	if err != nil {
		logs.Errorf(
			"cloud sub account created (uin=%s) but local db write failed, err: %v, rid: %s",
			cloudID, err, a.Cts.Kit.Rid,
		)
		return enumor.DeliverError,
			map[string]interface{}{
				"error":    fmt.Sprintf("save sub account to db failed, err: %v", err),
				"cloud_id": cloudID,
			}, err
	}

	// 三级账号创建后需要作为登记账号到Account表中，否则会触发账号未纳管的安全检查
	accountID, err := a.registerAccountForTCloud(cloudID, cloudResult)
	if err != nil {
		logs.Errorf(
			"cloud sub account created (uin=%s) but register account failed, err: %v, rid: %s",
			cloudID, err, a.Cts.Kit.Rid,
		)
		return enumor.DeliverError,
			map[string]interface{}{
				"error":    fmt.Sprintf("register sub account as account failed, err: %v", err),
				"cloud_id": cloudID,
			}, err
	}

	a.sendSecretMail(cloudResult)

	return enumor.Completed, map[string]interface{}{
		"sub_account_ids": subAccountIDs,
		"account_id":      accountID,
		"cloud_id":        cloudID,
	}, nil
}

func (a *ApplicationOfCreateSubAccount) registerAccountForTCloud(cloudID string,
	subAccountResult *hssubaccount.CreateSubAccountResult,
) (string, error) {
	parentAccount, err := a.Client.DataService().TCloud.Account.Get(
		a.Cts.Kit.Ctx, a.Cts.Kit.Header(), a.req.AccountID,
	)
	if err != nil {
		return "", fmt.Errorf("get parent account failed, err: %w", err)
	}

	result, err := a.Client.DataService().TCloud.Account.Create(
		a.Cts.Kit.Ctx,
		a.Cts.Kit.Header(),
		&dataprotocloud.AccountCreateReq[dataprotocloud.TCloudAccountExtensionCreateReq]{
			Name:        a.req.Name,
			Managers:    a.req.Managers,
			Type:        enumor.RegistrationAccount,
			Site:        parentAccount.Site,
			Memo:        a.req.Memo,
			BkBizID:     a.BkBizID(),
			UsageBizIDs: []int64{a.BkBizID()},
			Extension: &dataprotocloud.TCloudAccountExtensionCreateReq{
				CloudMainAccountID: parentAccount.Extension.CloudMainAccountID,
				CloudSubAccountID:  cloudID,
				CloudSecretID:      subAccountResult.SecretID,
				CloudSecretKey:     subAccountResult.SecretKey,
			},
		},
	)
	if err != nil {
		logs.Errorf("register account for tcloud failed, err: %v, rid: %s", err, a.Cts.Kit.Rid)
		return "", fmt.Errorf("register account for tcloud failed, err: %v", err)
	}

	return result.ID, nil
}

func (a *ApplicationOfCreateSubAccount) saveSubAccountToDB(cloudID string, ext []byte) ([]string, error) {
	createResult, err := a.Client.DataService().Global.SubAccount.BatchCreate(
		a.Cts.Kit,
		&dssubaccount.CreateReq{
			Items: []dssubaccount.CreateField{
				{
					CloudID:   cloudID,
					Name:      a.req.Name,
					Vendor:    a.Vendor(),
					Site:      enumor.InternationalSite,
					AccountID: a.req.AccountID,
					Managers:  a.req.Managers,
					BkBizIDs:  types.Int64Array{a.BkBizID()},
					Email:     converter.ValToPtr(a.req.Email),
					PhoneNum:  converter.ValToPtr(a.req.PhoneNum),
					Memo:      a.req.Memo,
					Extension: ext,
				},
			},
		},
	)
	if err != nil {
		return nil, err
	}

	return createResult.IDs, nil
}

// loginActionFlagToProtectionFlag converts a LoginActionFlag to an AccountProtectionFlag by returning
// the first enabled flag (value == 1) in priority order: Phone > Token > Stoken > Wechat > Custom > Mail > U2FToken.
// Returns nil if the flag is nil or no protection is enabled.
func loginActionFlagToProtectionFlag(flag *typeaccount.LoginActionFlag) *enumor.AccountProtectionFlag {
	if flag == nil {
		return nil
	}

	type entry struct {
		val  *uint64
		name enumor.AccountProtectionFlag
	}

	checks := []entry{
		{flag.Phone, enumor.PhoneProtection},
		{flag.Token, enumor.TokenProtection},
		{flag.Stoken, enumor.StokenProtection},
		{flag.Wechat, enumor.WechatProtection},
		{flag.Custom, enumor.CustomProtection},
		{flag.Mail, enumor.MailProtection},
		{flag.U2FToken, enumor.U2FTokenProtection},
	}

	for _, c := range checks {
		if converter.PtrToVal(c.val) == 1 {
			return &c.name
		}
	}

	return nil
}

func (a *ApplicationOfCreateSubAccount) sendSecretMail(result *hssubaccount.CreateSubAccountResult) {
	if a.req.ReceiveEmail == "" {
		logs.Warnf("send secret mail failed, receive email is empty, rid: %s", a.Cts.Kit.Rid)
		return
	}

	content := fmt.Sprintf(
		"您的三级账号已创建成功。\n\n"+
			"账号名称: %s\nSecretId: %s\nSecretKey: %s",
		converter.PtrToVal(result.Name), result.SecretID, result.SecretKey,
	)
	if result.Password != "" {
		content += fmt.Sprintf("\n密码: %s", result.Password)
	}

	err := a.SendMail(&cmsi.CmsiMail{
		Receiver: a.req.ReceiveEmail,
		Title:    fmt.Sprintf("三级账号(%s)开通通知", converter.PtrToVal(result.Name)),
		Content:  content,
	})
	if err != nil {
		logs.Errorf(
			"send secret mail to %s failed, err: %v, rid: %s",
			a.req.ReceiveEmail, err, a.Cts.Kit.Rid,
		)
	}
}
