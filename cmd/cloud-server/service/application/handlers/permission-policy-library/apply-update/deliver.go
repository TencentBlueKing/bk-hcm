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

package applyupdate

import (
	"fmt"

	"hcm/pkg/api/cloud-server"
	"hcm/pkg/criteria/enumor"
	"hcm/pkg/logs"
)

// GenerateApplicationContent generates the content to be stored in DB for this application.
func (a *ApplicationOfApplyPermPolicyLibUpdate) GenerateApplicationContent() interface{} {
	return a.Content
}

// Deliver executes the actual resource update after approval.
func (a *ApplicationOfApplyPermPolicyLibUpdate) Deliver() (enumor.ApplicationStatus, map[string]interface{}, error) {
	resp, err := a.ApplyUpdate(a.Cts.Kit, a.Content.Vendor, a.Content.PolicyLibraryID,
		[]string{a.Content.PermissionTemplateID})
	if err != nil {
		logs.Errorf("deliver: apply update failed, templateID: %s, err: %v, rid: %s", a.Content.PermissionTemplateID,
			err, a.Cts.Kit.Rid)
		return enumor.DeliverError, map[string]interface{}{"error": fmt.Sprintf("apply update failed, err: %v",
			err)}, err
	}

	if len(resp.Results) != 1 {
		logs.Errorf("deliver: apply update resp invalid, templateID: %s, resp: %v, rid: %s",
			a.Content.PermissionTemplateID, resp, a.Cts.Kit.Rid)
		return enumor.DeliverError, map[string]interface{}{"error": fmt.Sprintf("apply update resp invalid, "+
			"resp: %v", resp)}, fmt.Errorf("apply update resp invalid, resp: %v", resp)
	}

	result := resp.Results[0]
	if result.Status == cloudserver.ApplyStatusFailed {
		logs.Errorf("deliver: apply update failed, templateID: %s, result: %v, rid: %s",
			a.Content.PermissionTemplateID, result, a.Cts.Kit.Rid)
		return enumor.DeliverError, map[string]interface{}{"error": result.Reason},
			fmt.Errorf("apply update failed, %v", result.Reason)
	}

	return enumor.Completed, map[string]interface{}{
		"policy_library_id":      a.Content.PolicyLibraryID,
		"permission_template_id": a.Content.PermissionTemplateID,
	}, nil
}
