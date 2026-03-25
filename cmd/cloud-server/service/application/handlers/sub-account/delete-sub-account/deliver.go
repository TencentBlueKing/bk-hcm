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

package deletesubaccount

import (
	"fmt"

	dataservice "hcm/pkg/api/data-service"
	dataprotocloud "hcm/pkg/api/data-service/cloud"
	hssubaccount "hcm/pkg/api/hc-service/sub-account"
	"hcm/pkg/criteria/enumor"
	"hcm/pkg/logs"
	"hcm/pkg/runtime/filter"
)

// Deliver execute resource delivery after approval.
func (a *ApplicationOfDeleteSubAccount) Deliver() (enumor.ApplicationStatus, map[string]interface{}, error) {
	switch a.Vendor() {
	case enumor.TCloud:
		return a.deliverForTCloud()
	default:
		return enumor.DeliverError,
			map[string]interface{}{"error": fmt.Sprintf("vendor %s not supported", a.Vendor())},
			fmt.Errorf("vendor %s not supported for sub account deletion", a.Vendor())
	}
}

func (a *ApplicationOfDeleteSubAccount) deliverForTCloud() (enumor.ApplicationStatus, map[string]interface{}, error) {
	if err := a.deleteCloudSubAccount(); err != nil {
		return enumor.DeliverError,
			map[string]interface{}{
				"error":    fmt.Sprintf("delete cloud sub account failed, err: %v", err),
				"cloud_id": a.req.CloudID,
			}, err
	}

	if err := a.deleteLocalSubAccount(); err != nil {
		logs.Errorf(
			"cloud sub account deleted but local db delete failed, id: %s, cloud_id: %s, err: %v, rid: %s",
			a.req.ID, a.req.CloudID, err, a.Cts.Kit.Rid,
		)
		return enumor.DeliverError,
			map[string]interface{}{
				"error":    fmt.Sprintf("delete local sub account failed, err: %v", err),
				"cloud_id": a.req.CloudID,
			}, err
	}

	a.deleteRegistrationAccount()

	return enumor.Completed, map[string]interface{}{
		"deleted_id":       a.req.ID,
		"deleted_cloud_id": a.req.CloudID,
	}, nil
}

func (a *ApplicationOfDeleteSubAccount) deleteCloudSubAccount() error {
	return a.Client.HCService().TCloud.Account.DeleteSubAccount(
		a.Cts.Kit,
		&hssubaccount.DeleteSubAccountReq{
			AccountID: a.AccountID(),
			Name:      a.req.Name,
		},
	)
}

func (a *ApplicationOfDeleteSubAccount) deleteLocalSubAccount() error {
	return a.Client.DataService().Global.SubAccount.BatchDelete(
		a.Cts.Kit,
		&dataservice.BatchDeleteReq{
			Filter: &filter.Expression{
				Op: filter.And,
				Rules: []filter.RuleFactory{
					filter.AtomRule{
						Field: "id",
						Op:    filter.Equal.Factory(),
						Value: a.req.ID,
					},
				},
			},
		},
	)
}

// deleteRegistrationAccount deletes the registration account record in the account table
// by matching cloud_sub_account_id. Non-fatal if not found.
func (a *ApplicationOfDeleteSubAccount) deleteRegistrationAccount() {
	if err := a.deleteRegistrationAccountByCloudID(); err != nil {
		logs.Warnf(
			"delete registration account for cloud_id(%s) failed, err: %v, rid: %s",
			a.req.CloudID, err, a.Cts.Kit.Rid,
		)
	}
}

func (a *ApplicationOfDeleteSubAccount) deleteRegistrationAccountByCloudID() error {
	_, err := a.Client.DataService().Global.Account.Delete(
		a.Cts.Kit.Ctx,
		a.Cts.Kit.Header(),
		&dataprotocloud.AccountDeleteReq{
			Filter: &filter.Expression{
				Op: filter.And,
				Rules: []filter.RuleFactory{
					filter.AtomRule{
						Field: "type",
						Op:    filter.Equal.Factory(),
						Value: string(enumor.RegistrationAccount),
					},
					filter.AtomRule{
						Field: "extension.cloud_sub_account_id",
						Op:    filter.Equal.Factory(),
						Value: a.req.CloudID,
					},
				},
			},
		},
	)
	return err
}
