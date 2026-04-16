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

package subaccountsecret

import (
	"fmt"
	"strconv"

	proto "hcm/pkg/api/cloud-server/sub-account-secret"
	"hcm/pkg/api/core"
	coresubaccount "hcm/pkg/api/core/cloud/sub-account"
	coresass "hcm/pkg/api/core/cloud/sub-account-secret"
	protocloud "hcm/pkg/api/data-service/cloud"
	hssubaccount "hcm/pkg/api/hc-service/sub-account"
	"hcm/pkg/criteria/enumor"
	"hcm/pkg/criteria/errf"
	"hcm/pkg/dal/dao/tools"
	"hcm/pkg/iam/meta"
	"hcm/pkg/kit"
	"hcm/pkg/logs"
	"hcm/pkg/rest"
)

// CreateBizSubAccountSecret creates sub account secret under a business.
func (svc *service) CreateBizSubAccountSecret(cts *rest.Contexts) (interface{}, error) {
	bizID, err := cts.PathParameter("bk_biz_id").Int64()
	if err != nil {
		return nil, errf.NewFromErr(errf.InvalidParameter, err)
	}
	if bizID <= 0 {
		return nil, errf.NewFromErr(errf.InvalidParameter, fmt.Errorf("invalid business id: %d", bizID))
	}

	vendor := enumor.Vendor(cts.PathParameter("vendor").String())
	if err := vendor.Validate(); err != nil {
		return nil, errf.NewFromErr(errf.InvalidParameter, err)
	}

	req := new(proto.CreateReq)
	if err := cts.DecodeInto(req); err != nil {
		return nil, errf.NewFromErr(errf.DecodeRequestFailed, err)
	}
	if err := req.Validate(); err != nil {
		return nil, errf.NewFromErr(errf.InvalidParameter, err)
	}

	if err := svc.authorizeSubAccountSecret(cts.Kit, bizID); err != nil {
		return nil, err
	}

	switch vendor {
	case enumor.TCloud:
		return svc.createTCloudSubAccountSecret(cts.Kit, bizID, req.SubAccountID)
	default:
		return nil, fmt.Errorf("unsupported vendor: %s", vendor)
	}
}

// authorizeSubAccountSecret checks IAM permission for sub account secret creation.
func (svc *service) authorizeSubAccountSecret(kt *kit.Kit, bizID int64) error {
	authRes := meta.ResourceAttribute{
		Basic: &meta.Basic{Type: meta.SubAccountSecret, Action: meta.Create},
		BizID: bizID,
	}
	if err := svc.authorizer.AuthorizeWithPerm(kt, authRes); err != nil {
		return err
	}

	return nil
}

// createTCloudSubAccountSecret creates access key on TCloud and persists to DB.
func (svc *service) createTCloudSubAccountSecret(kt *kit.Kit, bizID int64, subAccountID string) (
	*proto.CreateResult, error) {

	subAccount, err := svc.getTCloudSubAccount(kt, subAccountID)
	if err != nil {
		return nil, err
	}

	if err = svc.validateAccountBiz(kt, bizID, subAccount.AccountID); err != nil {
		return nil, err
	}

	targetUin, err := strconv.ParseUint(subAccount.CloudID, 10, 64)
	if err != nil {
		return nil, errf.Newf(errf.InvalidParameter,
			"invalid cloud_id %s for sub account %s", subAccount.CloudID, subAccountID)
	}

	akResult, err := svc.client.HCService().TCloud.Account.CreateAccessKey(kt,
		&hssubaccount.TCloudCreateAccessKeyReq{
			AccountID: subAccount.AccountID,
			TargetUin: targetUin,
		})
	if err != nil {
		logs.Errorf("create access key failed, sub_account_id: %s, err: %v, rid: %s", subAccountID, err, kt.Rid)
		return nil, err
	}

	dbID, err := svc.saveTCloudSubAccountSecret(kt, subAccount, akResult)
	if err != nil {
		return nil, err
	}

	return &proto.CreateResult{
		ID: dbID,
		Extension: &proto.TCloudCreateExtension{
			CloudSecretID:  akResult.AccessKeyID,
			CloudSecretKey: akResult.SecretAccessKey,
		},
	}, nil
}

// getTCloudSubAccount retrieves a TCloud sub-account by HCM ID.
func (svc *service) getTCloudSubAccount(kt *kit.Kit, subAccountID string) (
	*coresubaccount.SubAccount[coresubaccount.TCloudExtension], error) {

	subAccount, err := svc.client.DataService().TCloud.SubAccount.Get(kt, subAccountID)
	if err != nil {
		logs.Errorf("get sub account failed, id: %s, err: %v, rid: %s",
			subAccountID, err, kt.Rid)
		return nil, err
	}

	if subAccount.AccountType == string(enumor.MainAccount) {
		return nil, fmt.Errorf("sub account %s is main account", subAccountID)
	}

	return subAccount, nil
}

// validateAccountBiz checks business ownership based on accountID resolved from sub-account ID.
func (svc *service) validateAccountBiz(kt *kit.Kit, bizID int64, accountID string) error {
	if accountID == "" {
		return errf.Newf(errf.InvalidParameter, "account id is empty")
	}

	accountListReq := &protocloud.AccountListReq{
		Filter: tools.ExpressionAnd(tools.RuleEqual("id", accountID)),
		Page:   core.NewDefaultBasePage(),
		Fields: []string{"id", "bk_biz_id"},
	}

	accountResult, err := svc.client.DataService().Global.Account.List(kt.Ctx, kt.Header(), accountListReq)
	if err != nil {
		return err
	}

	if len(accountResult.Details) == 0 {
		return errf.Newf(errf.RecordNotFound, "account %s not found", accountID)
	}

	if accountResult.Details[0].BkBizID != bizID {
		return errf.Newf(errf.PermissionDenied, "account %s does not belong to business %d", accountID, bizID)
	}

	return nil
}

// saveTCloudSubAccountSecret saves the access key info to data-service.
func (svc *service) saveTCloudSubAccountSecret(kt *kit.Kit,
	subAccount *coresubaccount.SubAccount[coresubaccount.TCloudExtension],
	akResult *hssubaccount.TCloudCreateAccessKeyResult) (string, error) {

	cloudMainAccountID := ""
	if subAccount.Extension != nil {
		cloudMainAccountID = subAccount.Extension.CloudMainAccountID
	}

	createReq := &protocloud.SubAccountSecretBatchCreateReq[coresass.TCloudSubAccountSecretExtension]{
		SubAccountSecrets: []protocloud.SubAccountSecretCreate[coresass.TCloudSubAccountSecretExtension]{
			{
				AccountID:      subAccount.AccountID,
				SubAccountID:   subAccount.ID,
				Status:         hssubaccount.TCloudAccessKeyStatusToSecretStatus(akResult.Status),
				CloudCreatedAt: akResult.CreateTime,
				Extension: &coresass.TCloudSubAccountSecretExtension{
					CloudSecretID:      akResult.AccessKeyID,
					CloudMainAccountID: cloudMainAccountID,
					CloudSubAccountID:  subAccount.CloudID,
				},
			},
		},
	}

	result, err := svc.client.DataService().TCloud.SubAccountSecret.BatchCreateSubAccountSecret(
		kt, createReq)
	if err != nil {
		logs.Errorf("persist sub account secret failed, sub_account_id: %s, err: %v, rid: %s", subAccount.ID, err, kt.Rid)
		return "", err
	}

	if len(result.IDs) != 1 {
		return "", errf.New(errf.Aborted, "create sub account secret returned length of ids not equal to 1")
	}

	return result.IDs[0], nil
}
