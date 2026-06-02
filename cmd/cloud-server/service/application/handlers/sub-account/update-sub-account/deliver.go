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

package updatesubaccount

import (
	"fmt"
	"strconv"

	"hcm/pkg/api/core"
	dssubaccount "hcm/pkg/api/data-service/cloud/sub-account"
	hssubaccount "hcm/pkg/api/hc-service/sub-account"
	"hcm/pkg/criteria/enumor"
	"hcm/pkg/dal/dao/tools"
	"hcm/pkg/logs"
	"hcm/pkg/tools/converter"
	"hcm/pkg/tools/slice"
)

// Deliver execute resource delivery after approval.
func (a *ApplicationOfUpdateSubAccount) Deliver() (enumor.ApplicationStatus, map[string]interface{}, error) {
	switch a.Vendor() {
	case enumor.TCloud:
		return a.deliverForTCloud()
	default:
		return enumor.DeliverError,
			map[string]interface{}{"error": fmt.Sprintf("vendor %s not supported", a.Vendor())},
			fmt.Errorf("vendor %s not supported for sub account update", a.Vendor())
	}
}

func (a *ApplicationOfUpdateSubAccount) deliverForTCloud() (enumor.ApplicationStatus, map[string]interface{}, error) {
	// 仅当有需要同步到云端的字段时，才调用云上接口更新账号基本信息（Email、PhoneNum、CountryCode、Memo）
	if a.hasCloudUpdatableFields() {
		if err := a.updateCloudSubAccount(); err != nil {
			return enumor.DeliverError,
				map[string]interface{}{"error": fmt.Sprintf("update cloud sub account failed, err: %v", err)}, err
		}
	}

	if err := a.updatePermissionTemplateOnCloud(); err != nil {
		logs.Errorf("update permission template on cloud failed, sub_account_id: %s, err: %v, rid: %s",
			a.req.ID, err, a.Cts.Kit.Rid)
		return enumor.DeliverError, map[string]interface{}{
			"error":          fmt.Sprintf("update permission template on cloud failed, err: %v", err),
			"sub_account_id": a.req.ID}, err
	}

	if err := a.updateLocalSubAccount(); err != nil {
		logs.Errorf("cloud sub account updated but local db update failed, id: %s, err: %v, rid: %s",
			a.req.ID, err, a.Cts.Kit.Rid,
		)
		return enumor.DeliverError,
			map[string]interface{}{"error": fmt.Sprintf("update local sub account failed, err: %v", err),
				"sub_account_id": a.req.ID}, err
	}

	return enumor.Completed, map[string]interface{}{
		"sub_account_id": a.req.ID,
	}, nil
}

// hasCloudUpdatableFields 判断请求中是否包含需要同步到云端的字段（邮箱、手机号、国际区号、备注）。
func (a *ApplicationOfUpdateSubAccount) hasCloudUpdatableFields() bool {
	return a.req.Email != nil || a.req.PhoneNum != nil ||
		a.req.CountryCode != nil || a.req.Memo != nil
}

func (a *ApplicationOfUpdateSubAccount) updateCloudSubAccount() error {
	req := &hssubaccount.TCloudUpdateSubAccountReq{
		AccountID: a.AccountID(),
		// Name 为云 API 必填项，取自 prepare 阶段获取的现有账号名，非用户变更字段。
		Name:        a.subAccountName,
		Email:       a.req.Email,
		PhoneNum:    a.req.PhoneNum,
		CountryCode: a.req.CountryCode,
		Remark:      a.req.Memo,
	}

	return a.Client.HCService().TCloud.Account.UpdateSubAccount(a.Cts.Kit, req)
}

func (a *ApplicationOfUpdateSubAccount) updatePermissionTemplateOnCloud() error {
	if a.req.PermissionTemplateIDs == nil {
		return nil
	}

	subAccounts, err := a.Client.DataService().Global.SubAccount.List(
		a.Cts.Kit,
		&core.ListReq{
			Filter: tools.ExpressionAnd(tools.RuleEqual("id", a.req.ID)),
			Page:   core.NewDefaultBasePage(),
		},
	)
	if err != nil {
		logs.Errorf("get sub account failed, id: %s, err: %v, rid: %s", a.req.ID, err, a.Cts.Kit.Rid)
		return fmt.Errorf("get sub account failed, id: %s, err: %v", a.req.ID, err)
	}
	if len(subAccounts.Details) == 0 {
		logs.Errorf("sub account not found, id: %s, rid: %s", a.req.ID, a.Cts.Kit.Rid)
		return fmt.Errorf("sub account not found, id: %s", a.req.ID)
	}

	subAccount := subAccounts.Details[0]
	uin, parseErr := strconv.ParseUint(subAccount.CloudID, 10, 64)
	if parseErr != nil {
		logs.Errorf("parse sub account cloud_id to uin failed, cloud_id: %s, err: %v, rid: %s",
			subAccount.CloudID, parseErr, a.Cts.Kit.Rid)
		return fmt.Errorf("parse sub account cloud_id to uin failed, cloud_id: %s, err: %w",
			subAccount.CloudID, parseErr)
	}

	toAttach := slice.NotIn(subAccount.PermissionTemplateIDs, a.req.PermissionTemplateIDs)
	toDetach := slice.NotIn(a.req.PermissionTemplateIDs, subAccount.PermissionTemplateIDs)

	if len(toAttach) > 0 {
		if err = a.AttachPolicies(uin, toAttach); err != nil {
			return err
		}
	}

	if len(toDetach) > 0 {
		if err = a.DetachPolicies(uin, toDetach); err != nil {
			return err
		}
	}

	return nil
}

func (a *ApplicationOfUpdateSubAccount) updateLocalSubAccount() error {
	field := dssubaccount.UpdateField{ID: a.req.ID}

	if a.req.Email != nil {
		field.Email = a.req.Email
	}
	if a.req.PhoneNum != nil {
		field.PhoneNum = a.req.PhoneNum
	}
	if a.req.CountryCode != nil {
		field.CountryCode = a.req.CountryCode
	}
	if a.req.Managers != nil {
		field.Managers = a.req.Managers
	}
	if a.req.Memo != nil {
		field.Memo = a.req.Memo
	}
	if a.req.BkBizID != nil {
		field.BkBizIDs = []int64{converter.PtrToVal(a.req.BkBizID)}
	}

	if len(a.req.PermissionTemplateIDs) > 0 {
		field.PermissionTemplateIDs = a.req.PermissionTemplateIDs
	}

	if err := a.Client.DataService().Global.SubAccount.BatchUpdate(a.Cts.Kit,
		&dssubaccount.UpdateReq{Items: []dssubaccount.UpdateField{field}}); err != nil {
		logs.Errorf("update sub account failed, err: %v, rid: %s", err, a.Cts.Kit.Rid)
		return err
	}

	updateFields, err := converter.StructToMap(field)
	if err != nil {
		logs.Errorf("convert update_sub_account field to map failed, err: %v, rid: %s", err, a.Cts.Kit.Rid)
		return err
	}
	err = a.CreateAudit(enumor.Update, enumor.SubAccountAuditResType, a.req.ID, a.subAccountName, updateFields)
	if err != nil {
		logs.Errorf("create update_sub_account audit failed, err: %v, rid: %s", err, a.Cts.Kit.Rid)
		return err
	}

	return nil
}
