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

	dssubaccount "hcm/pkg/api/data-service/cloud/sub-account"
	hssubaccount "hcm/pkg/api/hc-service/sub-account"
	"hcm/pkg/criteria/enumor"
	"hcm/pkg/logs"
	"hcm/pkg/tools/converter"
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
	if err := a.updateCloudSubAccount(); err != nil {
		return enumor.DeliverError,
			map[string]interface{}{"error": fmt.Sprintf("update cloud sub account failed, err: %v", err)}, err
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

func (a *ApplicationOfUpdateSubAccount) updateCloudSubAccount() error {
	req := &hssubaccount.TCloudUpdateSubAccountReq{
		AccountID:   a.AccountID(),
		Name:        a.subAccountName,
		Email:       a.req.Email,
		PhoneNum:    a.req.PhoneNum,
		CountryCode: a.req.CountryCode,
		Remark:      a.req.Memo,
	}

	return a.Client.HCService().TCloud.Account.UpdateSubAccount(a.Cts.Kit, req)
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

	updateFields, err := converter.StructToMap(field)
	if err != nil {
		logs.Errorf("convert update_sub_account field to map failed, err: %v, rid: %s", err, a.Cts.Kit.Rid)
		return err
	}
	err = a.Audit.ResUpdateAudit(a.Cts.Kit, enumor.SubAccountAuditResType, a.req.ID, updateFields)
	if err != nil {
		logs.Errorf("create update_sub_account audit failed, err: %v, rid: %s", err, a.Cts.Kit.Rid)
		return err
	}

	return a.Client.DataService().Global.SubAccount.BatchUpdate(
		a.Cts.Kit,
		&dssubaccount.UpdateReq{Items: []dssubaccount.UpdateField{field}},
	)
}
