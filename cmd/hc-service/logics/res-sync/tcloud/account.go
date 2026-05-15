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
 * We undertake not to change the open source license (MIT license) available
 *
 * to the current version of the project delivered to anyone in the future.
 */

package tcloud

import (
	"fmt"
	"strconv"

	typeaccount "hcm/pkg/adaptor/types/account"
	"hcm/pkg/api/core"
	protocore "hcm/pkg/api/core/cloud"
	coreas "hcm/pkg/api/core/cloud/account-secret"
	protocloud "hcm/pkg/api/data-service/cloud"
	"hcm/pkg/criteria/enumor"
	"hcm/pkg/criteria/errf"
	"hcm/pkg/criteria/validator"
	"hcm/pkg/dal/dao/tools"
	"hcm/pkg/kit"
	"hcm/pkg/logs"
	"hcm/pkg/tools/converter"
)

// SyncAccountOption defines sync account option.
type SyncAccountOption struct {
	AccountID string `json:"account_id" validate:"required"`
}

// Validate SyncAccountOption.
func (opt SyncAccountOption) Validate() error {
	return validator.Validate.Struct(opt)
}

// Account syncs a single TCloud second-level account (资源账号) cloud info to DB.
// It updates LoginFlag, ActionFlag, and account secret status.
func (cli *client) Account(kt *kit.Kit, opt *SyncAccountOption) (*SyncResult, error) {
	if err := opt.Validate(); err != nil {
		return nil, errf.NewFromErr(errf.InvalidParameter, err)
	}

	account, err := cli.dbCli.TCloud.Account.Get(kt.Ctx, kt.Header(), opt.AccountID)
	if err != nil {
		logs.Errorf("[%s] get account from db failed, err: %v, accountID: %s, rid: %s",
			enumor.TCloud, err, opt.AccountID, kt.Rid)
		return nil, err
	}

	if err = cli.syncAccountAuthFlag(kt, account); err != nil {
		return nil, err
	}

	// 二级账号使用的是高权限的三级账号的密钥，应该更新CloudSubAccountID的密钥信息
	subUin, err := strconv.ParseUint(account.Extension.CloudSubAccountID, 10, 64)
	if err != nil {
		return nil, fmt.Errorf("parse cloud_sub_account_id failed, accountID: %s, err: %v", opt.AccountID, err)
	}
	if err = cli.syncAccountSecretStatus(kt, account, subUin); err != nil {
		return nil, err
	}

	return new(SyncResult), nil
}

// syncAccountAuthFlag fetches LoginFlag and ActionFlag from TCloud DescribeSafeAuthFlag
// and patches the account extension. Skips if subaccount UIN equals main account UIN.
func (cli *client) syncAccountAuthFlag(kt *kit.Kit,
	account *protocloud.AccountGetResult[protocore.TCloudAccountExtension]) error {

	uin, err := strconv.ParseUint(account.Extension.CloudMainAccountID, 10, 64)
	if err != nil {
		logs.Errorf("[%s] parse cloud_main_account_id failed, accountID: %s, err: %v,rid: %s", enumor.TCloud,
			account.ID, err, kt.Rid)
		return errf.NewFromErr(errf.Aborted, err)
	}

	authFlags, err := cli.cloudCli.DescribeSafeAuthFlagColl(kt, &typeaccount.DescribeSafeAuthFlagCollOption{
		SubUins: []uint64{uin},
	})
	if err != nil {
		logs.Errorf("[%s] describe safe auth flag failed, err: %v, accountID: %s, rid: %s",
			enumor.TCloud, err, account.ID, kt.Rid)
		return err
	}
	if len(authFlags) != 1 {
		logs.Errorf("[%s] describe safe auth flag result length != 1, accountID: %s, rid: %s",
			enumor.TCloud, account.ID, kt.Rid)
		return errf.NewFromErr(errf.Aborted,
			fmt.Errorf("describe safe auth flag result length != 1, accountID: %s", account.ID))
	}
	authFlag := authFlags[0]

	var cloudLoginFlag, cloudActionFlag *enumor.AccountProtectionFlag
	if authFlag.LoginFlag != nil {
		cloudLoginFlag = authFlag.LoginFlag.ToProtectionFlag()
	}
	if authFlag.ActionFlag != nil {
		cloudActionFlag = authFlag.ActionFlag.ToProtectionFlag()
	}

	updateReq := &protocloud.AccountUpdateReq[protocloud.TCloudAccountExtensionUpdateReq]{
		Extension: &protocloud.TCloudAccountExtensionUpdateReq{
			LoginFlag:  cloudLoginFlag,
			ActionFlag: cloudActionFlag,
		},
	}
	if _, err = cli.dbCli.TCloud.Account.Update(kt.Ctx, kt.Header(), account.ID, updateReq); err != nil {
		logs.Errorf("[%s] update account auth flag failed, err: %v, accountID: %s, rid: %s",
			enumor.TCloud, err, account.ID, kt.Rid)
		return err
	}

	logs.Infof("[%s] sync account auth flag success, accountID: %s, rid: %s", enumor.TCloud, account.ID, kt.Rid)

	return nil
}

// syncAccountSecretStatus fetches access key status from TCloud ListAccessKeys and
// updates account_secret.status to match. Logs and skips on permission errors.
func (cli *client) syncAccountSecretStatus(kt *kit.Kit,
	account *protocloud.AccountGetResult[protocore.TCloudAccountExtension], subUin uint64) error {

	secrets, err := cli.dbCli.TCloud.AccountSecret.ListAccountSecretWithExtension(
		kt,
		&protocloud.AccountSecretExtListReq{
			Filter: tools.EqualExpression("account_id", account.ID),
			Page:   core.NewDefaultBasePage(),
		},
	)
	if err != nil {
		logs.Errorf("[%s] list account secret failed, err: %v, accountID: %s, rid: %s",
			enumor.TCloud, err, account.ID, kt.Rid)
		return err
	}

	if len(secrets.Details) == 0 {
		return nil
	}

	cloudKeys, err := cli.cloudCli.ListAccessKeys(kt, &typeaccount.ListAccessKeysOption{TargetUin: subUin})
	if err != nil {
		logs.Errorf("[%s] list access keys failed, err: %v, accountID: %s, rid: %s, skip secret status sync",
			enumor.TCloud, err, account.ID, kt.Rid)
		return nil
	}

	cloudKeyStatusMap := make(map[string]string, len(cloudKeys))
	for _, k := range cloudKeys {
		cloudKeyStatusMap[k.AccessKeyID] = k.Status
	}

	updateItems := make([]protocloud.AccountSecretUpdate[coreas.TCloudAccountSecretExtension], 0)
	for _, secret := range secrets.Details {
		if secret.Extension == nil {
			return errf.NewFromErr(errf.InvalidParameter, fmt.Errorf("secret(id=%s) extension is nil, rid: %s",
				secret.ID, kt.Rid))
		}

		cloudStatus, ok := cloudKeyStatusMap[secret.Extension.CloudSecretID]
		if !ok {
			logs.Errorf("[%s] cloud key not found for secret, id: %s, rid: %s", enumor.TCloud, secret.ID, kt.Rid)
			return errf.NewFromErr(errf.InvalidParameter,
				fmt.Errorf("cloud key not found for secret, id: %s, rid: %s", secret.ID, kt.Rid))
		}

		newStatus := enumor.NewAccountSecretStatusFromTCloud(cloudStatus)
		if newStatus == secret.Status {
			continue
		}

		updateItems = append(updateItems, protocloud.AccountSecretUpdate[coreas.TCloudAccountSecretExtension]{
			ID:     secret.ID,
			Status: converter.ValToPtr(newStatus),
		})
	}

	if len(updateItems) == 0 {
		return nil
	}

	updateReq := &protocloud.AccountSecretBatchUpdateReq[coreas.TCloudAccountSecretExtension]{
		AccountSecrets: updateItems,
	}
	if err = cli.dbCli.TCloud.AccountSecret.BatchUpdateAccountSecret(kt, updateReq); err != nil {
		logs.Errorf("[%s] batch update account secret status failed, err: %v, accountID: %s, rid: %s",
			enumor.TCloud, err, account.ID, kt.Rid)
		return err
	}

	logs.Infof("[%s] sync account secret status success, accountID: %s, updated: %d, rid: %s",
		enumor.TCloud, account.ID, len(updateItems), kt.Rid)

	return nil
}
