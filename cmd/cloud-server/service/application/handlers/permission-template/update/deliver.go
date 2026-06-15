/*
 * TencentBlueKing is pleased to support the open source community by making
 * 蓝鲸智云 - 混合云管理平台 (BlueKing - Hybrid Cloud Management System) available.
 * Copyright (C) 2022 THL A29 Limited,
 * a Tencent company. All rights reserved.
 * Licensed under the MIT License (the "License");
 * you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at http://opensource.org/licenses/MIT
 * Unless required by applicable law or agreed to in writing,
 * software distributed under the MIT License is distributed on
 * an "AS IS" BASIS, WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND,
 * either express or implied. See the License for the
 * specific language governing permissions and limitations under the License.
 *
 * We undertake not to change the open source license (MIT license) applicable
 *
 * to the current version of the project delivered to anyone in the future.
 */

package updatepermtemplate

import (
	"fmt"

	"hcm/cmd/cloud-server/service/permission-policy-library"
	"hcm/pkg/api/cloud-server"
	"hcm/pkg/criteria/enumor"
	"hcm/pkg/logs"
)

// GenerateApplicationContent generates the content to be stored in DB for this application.
func (a *ApplicationOfUpdatePermTemplate) GenerateApplicationContent() interface{} {
	return a.content
}

// Deliver executes the actual resource update after approval.
func (a *ApplicationOfUpdatePermTemplate) Deliver() (enumor.ApplicationStatus, map[string]interface{}, error) {
	policyLibraryID := a.content.PolicyLibraryID
	templateIDs := []string{a.content.ID}
	tmplInfo := permissionpolicylibrary.UpdateTmplBaseInfo{Memo: a.content.Memo}
	resp, err := a.ApplyUpdateWithTmplInfo(a.Cts.Kit, a.content.Vendor, policyLibraryID, templateIDs, tmplInfo)
	if err != nil {
		logs.Errorf("deliver: apply update template failed, templateID: %s, libraryID: %s, err: %v, rid: %s",
			a.content.ID, a.content.PolicyLibraryID, err, a.Cts.Kit.Rid)
		return enumor.DeliverError, map[string]interface{}{
			"error": fmt.Sprintf("apply update template failed, err: %v", err),
		}, err
	}

	if len(resp.Results) != 1 {
		logs.Errorf("deliver: apply update template resp invalid, templateID: %s, libraryID: %s, resp: %v, rid: %s",
			a.content.ID, a.content.PolicyLibraryID, resp, a.Cts.Kit.Rid)
		return enumor.DeliverError, map[string]interface{}{
			"error": fmt.Sprintf("apply update template resp invalid, resp: %v",
				resp)}, fmt.Errorf("apply update template resp invalid, resp: %v", resp)
	}

	result := resp.Results[0]
	if result.Status == cloudserver.ApplyStatusFailed {
		logs.Errorf("deliver: apply update template failed, templateID: %s, libraryID: %s, result: %v, rid: %s",
			a.content.ID, a.content.PolicyLibraryID, result, a.Cts.Kit.Rid)
		return enumor.DeliverError, map[string]interface{}{"error": result.Reason},
			fmt.Errorf("apply update template failed, %v", result.Reason)
	}

	return enumor.Completed, map[string]interface{}{
		"policy_library_id": a.content.PolicyLibraryID,
		"id":                a.content.ID,
	}, nil
}
