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

	// 获取三级账号的二级账号
	account, err := a.Client.DataService().TCloud.Account.Get(a.Cts.Kit.Ctx, a.Cts.Kit.Header(), a.req.AccountID)
	if err != nil {
		return enumor.DeliverError,
			map[string]interface{}{"error": fmt.Sprintf("get parent account failed, err: %v", err)}, err
	}

	// Step 1: 先在云上创建账号，并保存云上的base信息
	cloudResult, err := a.createTCloudSubAccountInCloud(ext)
	if err != nil {
		return enumor.DeliverError,
			map[string]interface{}{"error": fmt.Sprintf("create cloud sub account failed, err: %v", err)}, err
	}
	cloudID := strconv.FormatUint(converter.PtrToVal(cloudResult.Uin), 10)

	// Step 2: 本地保存云上的base信息
	subAccountIDs, accountID, err := a.saveSubAccountBasicInfo(cloudResult, ext, account)
	if err != nil {
		logs.Errorf("cloud sub account created (uin=%s) but local persistence failed, err: %v, rid: %s", cloudID,
			err, a.Cts.Kit.Rid)
		return enumor.DeliverError,
			map[string]interface{}{"error": fmt.Sprintf("save sub account/account to db failed, err: %v", err),
				"cloud_id": converter.PtrToVal(cloudResult.Uin)}, err
	}

	// Step 3: 配置子账号（同步详情、设置MFA标志、附加权限模板）
	var subAccountID string
	if len(subAccountIDs) > 0 {
		subAccountID = subAccountIDs[0]
	}
	err = a.configureSubAccount(subAccountID, cloudResult, account)
	if err != nil {
		logs.Errorf("sub account created (uin=%s) but configuration failed, err: %v, rid: %s",
			cloudID, err, a.Cts.Kit.Rid)
		return enumor.DeliverError,
			map[string]interface{}{"error": fmt.Sprintf("configure sub account failed, err: %v", err),
				"cloud_id": cloudID}, err
	}

	if err := a.sendSubAccountMail(&cloudResult.TCloudCreateSubAccountResult); err != nil {
		logs.Errorf("cloud sub account created (uin=%s) but send sub account mail failed, err: %v, rid: %s",
			cloudID, err, a.Cts.Kit.Rid)
		return enumor.DeliverError, map[string]interface{}{
			"error":    fmt.Sprintf("send secret mail failed, err: %v", err),
			"cloud_id": cloudID}, err
	}

	return enumor.Completed, map[string]interface{}{"sub_account_ids": subAccountIDs, "account_id": accountID,
		"cloud_id": cloudID,
	}, nil
}

// createTCloudSubAccountInCloud creates the subaccount on Tencent Cloud, queries its detail
// and best-effort loads safe auth flags.
func (a *ApplicationOfCreateSubAccount) createTCloudSubAccountInCloud(ext *proto.TCloudSubAccountAddExtension) (
	*hssubaccount.TCloudCreateAggregateResult, error) {

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
		logs.Errorf("create tcloud subaccount (%s) failed, err: %v, rid: %s", a.req.Name, err, a.Cts.Kit.Rid)
		return nil, fmt.Errorf("create tcloud sub account (%s) failed, err: %v", a.req.Name, err)
	}
	if cloudResult == nil {
		logs.Errorf("create tcloud subaccount(%s) result failed, result is nil, rid: %s",
			a.req.Name, a.Cts.Kit.Rid)
		return nil, fmt.Errorf("create tcloud subaccount(%s) result is nil", a.req.Name)
	}

	return &hssubaccount.TCloudCreateAggregateResult{
		TCloudCreateSubAccountResult: *cloudResult,
	}, nil
}

// configureSubAccount configures sub account with all necessary settings including MFA flags,
// permission templates, and syncs detail info from cloud to local DB.
func (a *ApplicationOfCreateSubAccount) configureSubAccount(subAccountID string,
	cloudResult *hssubaccount.TCloudCreateAggregateResult,
	account *dataprotocloud.AccountGetResult[protocore.TCloudAccountExtension]) error {

	if subAccountID == "" {
		logs.Errorf("sub account id is empty, name=%s, rid: %s", a.req.Name, a.Cts.Kit.Rid)
		return fmt.Errorf("sub account id is empty")
	}

	uin := converter.PtrToVal(cloudResult.Uin)

	// Step 1: Set MFA flags for the sub account
	err := a.Client.HCService().TCloud.Account.SetMfaFlag(a.Cts.Kit, &hssubaccount.TCloudSetMfaFlagReq{
		AccountID:  a.req.AccountID,
		OpUin:      uin,
		LoginFlag:  &typeaccount.LoginActionFlag{Stoken: converter.ValToPtr(uint64(1))},
		ActionFlag: &typeaccount.LoginActionFlag{Stoken: converter.ValToPtr(uint64(1))},
	})
	if err != nil {
		logs.Errorf("set mfa flag for sub account (%s) failed, err: %v, rid: %s", a.req.Name, err, a.Cts.Kit.Rid)
		return fmt.Errorf("set mfa flag for sub account (%s) failed, err: %v", a.req.Name, err)
	}

	// Step 2: Attach permission templates to the subaccount
	err = a.AttachPolicies(uin, a.req.PermissionTemplateIDs)
	if err != nil {
		logs.Errorf("attach permission to cloud failed for sub account (%s), uin=%d, err: %v, rid: %s",
			a.req.Name, uin, err, a.Cts.Kit.Rid)
		return fmt.Errorf("attach permission to cloud failed, err: %v", err)
	}

	// Step 3: Query subaccount detail info from cloud
	cloudSubAccounts, err := a.Client.HCService().TCloud.Account.DescribeSubAccounts(
		a.Cts.Kit, &hssubaccount.TCloudDescribeSubAccountsReq{AccountID: a.req.AccountID, SubUin: []uint64{uin}},
	)
	if err != nil {
		logs.Errorf("describe sub accounts for sub account (%s) failed, err: %v, rid: %s",
			a.req.Name, err, a.Cts.Kit.Rid)
		return fmt.Errorf("describe sub accounts failed, err: %v", err)
	}
	if len(cloudSubAccounts) != 1 {
		logs.Errorf("sub account count is not 1, uin=%d, name=%s, count=%d, rid: %s",
			uin, a.req.Name, len(cloudSubAccounts), a.Cts.Kit.Rid)
		return fmt.Errorf("sub account count is not 1, got %d", len(cloudSubAccounts))
	}
	cloudResult.CreateTime = cloudSubAccounts[0].CreateTime

	// Step 4: Get safe auth flags from cloud
	safeAuthFlags, err := a.Client.HCService().TCloud.Account.DescribeSafeAuthFlagColl(
		a.Cts.Kit, &hssubaccount.TCloudDescribeSafeAuthFlagCollReq{AccountID: a.req.AccountID, SubUins: []uint64{uin}},
	)
	if err != nil {
		logs.Errorf("get safe auth flag failed for sub account (uin=%d, name=%s), err: %v, rid: %s",
			uin, a.req.Name, err, a.Cts.Kit.Rid)
		return fmt.Errorf("get safe auth flag failed, err: %v", err)
	}
	if len(safeAuthFlags) != 1 {
		logs.Errorf("safe auth flag result count is not 1, uin=%d, name=%s, count=%d, rid: %s",
			uin, a.req.Name, len(safeAuthFlags), a.Cts.Kit.Rid)
		return fmt.Errorf("safe auth flag result count is not 1, got %d", len(safeAuthFlags))
	}

	// Step 5: Update subaccount with all detail info including permission template IDs
	if err = a.updateSubAccountConfigureDetail(subAccountID, cloudResult, &safeAuthFlags[0], account,
		a.req.PermissionTemplateIDs); err != nil {
		return fmt.Errorf("update sub account with detail failed, err: %v", err)
	}

	return nil
}

func (a *ApplicationOfCreateSubAccount) registerAccountForTCloud(cloudID string,
	createResult *hssubaccount.TCloudCreateAggregateResult,
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

// saveSubAccountBasicInfo saves the basic subaccount info to local DB.
// It creates the sub_account record and registers the account. This should succeed,
// otherwise the cloud account becomes orphaned.
func (a *ApplicationOfCreateSubAccount) saveSubAccountBasicInfo(cloudResult *hssubaccount.TCloudCreateAggregateResult,
	ext *proto.TCloudSubAccountAddExtension,
	parentAccount *dataprotocloud.AccountGetResult[protocore.TCloudAccountExtension]) ([]string, string, error) {

	cloudID := strconv.FormatUint(converter.PtrToVal(cloudResult.Uin), 10)

	if ext == nil {
		return nil, "", fmt.Errorf("extension is required")
	}

	tCloudExt := &coresubaccount.TCloudExtension{
		CloudMainAccountID: parentAccount.Extension.CloudMainAccountID,
		Uin:                cloudResult.Uin,
		NickName:           cloudResult.Name,
		ConsoleLogin:       ext.ConsoleLogin,
	}
	extBytes, err := core.MarshalStruct(tCloudExt)
	if err != nil {
		return nil, "", fmt.Errorf("marshal extension failed, err: %v", err)
	}

	detail := dssubaccount.CreateField{
		CloudID:               cloudID,
		Name:                  a.req.Name,
		Vendor:                a.Vendor(),
		Site:                  parentAccount.Site,
		AccountID:             a.req.AccountID,
		Managers:              a.req.Managers,
		BkBizIDs:              types.Int64Array{a.BkBizID()},
		CountryCode:           converter.ValToPtr(a.req.CountryCode),
		PermissionTemplateIDs: a.req.PermissionTemplateIDs,
		Email:                 converter.ValToPtr(a.req.Email),
		PhoneNum:              converter.ValToPtr(a.req.PhoneNum),
		Memo:                  a.req.Memo,
		Extension:             extBytes,
	}

	createResult, err := a.Client.DataService().Global.SubAccount.BatchCreate(
		a.Cts.Kit, &dssubaccount.CreateReq{Items: []dssubaccount.CreateField{detail}})
	if err != nil {
		return nil, "", err
	}

	// registerAccountForTCloud 将用户创建的三级账号登记到account表，防止触发HCM未纳管该账号的安全工单
	accountID, err := a.registerAccountForTCloud(cloudID, cloudResult, parentAccount)
	if err != nil {
		return nil, "", err
	}

	if len(createResult.IDs) > 0 {
		if err = a.CreateAudit(enumor.Create, enumor.SubAccountAuditResType,
			createResult.IDs[0], a.req.Name, detail); err != nil {
			logs.Errorf("create sub account audit failed, sub_account_id: %s, err: %v, rid: %s",
				createResult.IDs[0], err, a.Cts.Kit.Rid)
			return nil, "", err
		}
	}

	return createResult.IDs, accountID, nil
}

// updateSubAccountConfigureDetail updates the subaccount with detail info, safe auth flags and permission template IDs.
func (a *ApplicationOfCreateSubAccount) updateSubAccountConfigureDetail(subAccountID string,
	cloudResult *hssubaccount.TCloudCreateAggregateResult, safeAuth *typeaccount.SafeAuthFlagCollResult,
	account *dataprotocloud.AccountGetResult[protocore.TCloudAccountExtension], permissionTemplateIDs []string) error {

	if subAccountID == "" {
		return fmt.Errorf("subAccountID is required for update")
	}

	var loginProt, actionProt *enumor.AccountProtectionFlag
	if safeAuth != nil {
		if safeAuth.LoginFlag != nil {
			loginProt = safeAuth.LoginFlag.ToProtectionFlag()
		}
		if safeAuth.ActionFlag != nil {
			actionProt = safeAuth.ActionFlag.ToProtectionFlag()
		}
	}

	tCloudExt := &coresubaccount.TCloudExtension{
		CloudMainAccountID: account.Extension.CloudMainAccountID,
		Uin:                cloudResult.Uin,
		NickName:           cloudResult.Name,
		CreateTime:         cloudResult.CreateTime,
		LoginFlag:          loginProt,
		ActionFlag:         actionProt,
	}
	extBytes, err := core.MarshalStruct(tCloudExt)
	if err != nil {
		return fmt.Errorf("marshal extension failed, err: %v", err)
	}

	updateField := dssubaccount.UpdateField{
		ID: subAccountID, Extension: &extBytes, PermissionTemplateIDs: permissionTemplateIDs,
	}
	err = a.Client.DataService().Global.SubAccount.BatchUpdate(
		a.Cts.Kit, &dssubaccount.UpdateReq{Items: []dssubaccount.UpdateField{updateField}},
	)
	if err != nil {
		return fmt.Errorf("update sub account failed, err: %v", err)
	}

	updateMap, err := converter.StructToMap(updateField)
	if err != nil {
		logs.Errorf("convert update field to map failed, err: %v, rid: %s", err, a.Cts.Kit.Rid)
		return fmt.Errorf("convert update field to map failed, err: %v", err)
	}

	if err = a.CreateAudit(enumor.Update, enumor.SubAccountAuditResType,
		subAccountID, a.req.Name, updateMap); err != nil {
		logs.Errorf("update sub account audit failed, sub_account_id: %s, err: %v, rid: %s",
			subAccountID, err, a.Cts.Kit.Rid)
		return err
	}
	return nil
}

func (a *ApplicationOfCreateSubAccount) sendSubAccountMail(result *hssubaccount.TCloudCreateSubAccountResult) error {
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

	err := a.SendMail(&cmsi.CmsiMailParams{
		Receiver:   []string{a.req.ReceiveEmail},
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
