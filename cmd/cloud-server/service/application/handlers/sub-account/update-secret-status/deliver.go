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

package updatesecretstatus

import (
	"fmt"
	"strconv"
	"time"

	"hcm/pkg/api/core"
	coresass "hcm/pkg/api/core/cloud/sub-account-secret"
	protocloud "hcm/pkg/api/data-service/cloud"
	hssubaccount "hcm/pkg/api/hc-service/sub-account"
	"hcm/pkg/criteria/enumor"
	"hcm/pkg/dal/dao/tools"
	"hcm/pkg/logs"
	"hcm/pkg/tools/converter"
)

// Deliver execute resource delivery after approval.
func (a *ApplicationOfUpdateSecretKeyStatus) Deliver() (enumor.ApplicationStatus, map[string]interface{}, error) {
	switch a.Vendor() {
	case enumor.TCloud:
		return a.tcloudDeliver()
	default:
		return enumor.DeliverError,
			map[string]interface{}{
				"error": fmt.Sprintf("vendor %s not supported", a.Vendor()),
			},
			fmt.Errorf("vendor %s not supported", a.Vendor())
	}
}

func (a *ApplicationOfUpdateSecretKeyStatus) tcloudDeliver() (enumor.ApplicationStatus, map[string]interface{}, error) {
	if err := a.tcloudUpdateCloudSecretStatus(); err != nil {
		return enumor.DeliverError,
			map[string]interface{}{
				"error":     fmt.Sprintf("update cloud secret status failed, err: %v", err),
				"secret_id": a.req.ID,
			}, err
	}

	if err := a.tcloudUpdateLocalSecretStatus(); err != nil {
		logs.Errorf("cloud secret status updated but local db update failed, secret_id: %s, err: %v, rid: %s",
			a.req.ID, err, a.Cts.Kit.Rid)
		return enumor.DeliverError,
			map[string]interface{}{
				"error":     fmt.Sprintf("update local secret status failed, err: %v", err),
				"secret_id": a.req.ID,
			}, err
	}

	return enumor.Completed, map[string]interface{}{
		"secret_id": a.req.ID,
	}, nil
}

func (a *ApplicationOfUpdateSecretKeyStatus) getCloudSecretIDForDisplay() (string, error) {
	switch a.Vendor() {
	case enumor.TCloud:
		detail, err := a.getTCloudSecretDetail()
		if err != nil {
			return "", err
		}
		if detail.Extension != nil {
			return detail.Extension.CloudSecretID, nil
		}
		return "", nil
	default:
		return "", fmt.Errorf("unsupported vendor: %s", a.Vendor())
	}
}

func (a *ApplicationOfUpdateSecretKeyStatus) tcloudUpdateCloudSecretStatus() error {
	secretDetail, err := a.getTCloudSecretDetail()
	if err != nil {
		return err
	}

	targetUin, err := strconv.ParseUint(secretDetail.Extension.CloudSubAccountID, 10, 64)
	if err != nil {
		return fmt.Errorf("parse cloud_sub_account_id(%s) to uin failed, err: %w",
			secretDetail.Extension.CloudSubAccountID, err)
	}

	return a.Client.HCService().TCloud.Account.UpdateAccessKey(
		a.Cts.Kit,
		&hssubaccount.TCloudUpdateAccessKeyReq{
			AccountID:   a.AccountID(),
			TargetUin:   targetUin,
			AccessKeyID: secretDetail.Extension.CloudSecretID,
			Status:      hssubaccount.SecretStatusToTCloudAccessKeyStatus(a.req.Status),
		},
	)
}

func (a *ApplicationOfUpdateSecretKeyStatus) getTCloudSecretDetail() (
	*coresass.SubAccountSecret[coresass.TCloudSubAccountSecretExtension], error) {

	result, err := a.Client.DataService().TCloud.SubAccountSecret.
		ListSubAccountSecretWithExtension(
			a.Cts.Kit,
			&protocloud.SubAccountSecretExtListReq{
				Filter: tools.ExpressionAnd(tools.RuleEqual("id", a.req.ID)),
				Page:   &core.BasePage{Start: 0, Limit: 1},
			},
		)
	if err != nil {
		return nil, fmt.Errorf("query secret detail failed, err: %w", err)
	}

	if len(result.Details) == 0 {
		return nil, fmt.Errorf("secret(id=%s) not found", a.req.ID)
	}

	return &result.Details[0], nil
}

func (a *ApplicationOfUpdateSecretKeyStatus) tcloudUpdateLocalSecretStatus() error {
	updateSecret := protocloud.SubAccountSecretUpdate[coresass.TCloudSubAccountSecretExtension]{
		ID:     a.req.ID,
		Status: &a.req.Status,
	}
	if a.req.Status == enumor.DisabledSecretStatus {
		updateSecret.DisabledTime = converter.ValToPtr(time.Now().Local().Format(time.RFC3339))
	}

	err := a.Audit.ResUpdateAudit(a.Cts.Kit, enumor.SubAccountSecretAuditResType, a.req.ID,
		map[string]interface{}{"status": a.req.Status})
	if err != nil {
		logs.Errorf("create update_secret_status audit failed, err: %v, rid: %s", err, a.Cts.Kit.Rid)
		return err
	}

	return a.Client.DataService().TCloud.SubAccountSecret.BatchUpdateSubAccountSecret(
		a.Cts.Kit,
		&protocloud.SubAccountSecretBatchUpdateReq[coresass.TCloudSubAccountSecretExtension]{
			SubAccountSecrets: []protocloud.SubAccountSecretUpdate[coresass.TCloudSubAccountSecretExtension]{
				updateSecret},
		},
	)
}
