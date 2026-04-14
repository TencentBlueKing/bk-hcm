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

package createpermtemplate

import (
	"fmt"

	"hcm/cmd/cloud-server/service/permission-policy-library"
	"hcm/pkg/api/cloud-server"
	"hcm/pkg/criteria/enumor"
	"hcm/pkg/logs"
)

// GenerateApplicationContent generates the content to be stored in DB for this application.
func (a *ApplicationOfCreatePermTemplate) GenerateApplicationContent() interface{} {
	return a.content
}

// Deliver executes the actual resource creation after approval.
func (a *ApplicationOfCreatePermTemplate) Deliver() (enumor.ApplicationStatus, map[string]interface{}, error) {
	tmplInfo := permissionpolicylibrary.CreateTmplBaseInfo{Name: a.content.Name, Memo: a.content.Memo}
	accountIDs := []string{a.content.AccountID}
	resp, err := a.ApplyCreateWithTmplInfo(a.Cts.Kit, a.content.Vendor, a.content.PolicyLibraryID, accountIDs, tmplInfo)
	if err != nil {
		logs.Errorf("deliver: apply create failed, err: %v, accountID: %s, tmplInfo: %v, rid: %s", err,
			a.content.AccountID, tmplInfo, a.Cts.Kit.Rid)
		return enumor.DeliverError, map[string]interface{}{"error": fmt.Sprintf("apply create failed, err: %v",
			err)}, err
	}

	if len(resp.Results) != 1 {
		logs.Errorf("deliver: apply create resp invalid, accountID: %s, resp: %v, rid: %s", a.content.AccountID,
			resp, a.Cts.Kit.Rid)
		return enumor.DeliverError, map[string]interface{}{"error": fmt.Sprintf("apply create resp invalid, "+
			"resp: %v", resp)}, fmt.Errorf("apply create resp invalid, resp: %v", resp)
	}

	result := resp.Results[0]
	if result.Status == cloudserver.ApplyStatusFailed {
		logs.Errorf("deliver: apply create failed, accountID: %s, result: %v, rid: %s", a.content.AccountID, result,
			a.Cts.Kit.Rid)
		return enumor.DeliverError, map[string]interface{}{"error": result.Reason},
			fmt.Errorf("apply create failed, %v", result.Reason)
	}

	return enumor.Completed, map[string]interface{}{"policy_library_id": a.content.PolicyLibraryID,
		"account_id": a.content.AccountID}, nil
}
