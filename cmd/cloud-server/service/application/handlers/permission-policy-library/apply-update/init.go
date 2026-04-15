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

// Package applyupdate is the handler for apply_permission_policy_library (update action).
package applyupdate

import (
	"fmt"

	"hcm/cmd/cloud-server/service/application/handlers"
	"hcm/cmd/cloud-server/service/application/handlers/permission-policy-library"
	proto "hcm/pkg/api/cloud-server/application"
	"hcm/pkg/api/core"
	protocloud "hcm/pkg/api/data-service/cloud"
	"hcm/pkg/criteria/enumor"
	"hcm/pkg/dal/dao/tools"
	"hcm/pkg/kit"
	"hcm/pkg/logs"
	"hcm/pkg/thirdparty/api-gateway/itsm"
	"hcm/pkg/tools/json"
)

func init() {
	permissionpolicylibrary.RegisterActionHandler(
		enumor.PermPolicyLibActionApplyUpdate,
		func(opt *handlers.HandlerOption, base *proto.ApplyPermPolicyLibBaseContent,
			content string) (handlers.ApplicationHandler, error) {

			c := new(proto.ApplyPermPolicyLibUpdateContent)
			if err := json.UnmarshalFromString(content, c); err != nil {
				return nil, fmt.Errorf("unmarshal apply perm policy lib update content failed, err: %w", err)
			}
			return NewApplicationOfApplyPermPolicyLibUpdate(opt, c), nil
		},
	)
}

// ApplicationOfApplyPermPolicyLibUpdate is the handler for apply_permission_policy_library (update action).
type ApplicationOfApplyPermPolicyLibUpdate struct {
	permissionpolicylibrary.ApplicationBasePermissionPolicyLibrary

	Content *proto.ApplyPermPolicyLibUpdateContent
}

// NewApplicationOfApplyPermPolicyLibUpdate creates a new handler.
func NewApplicationOfApplyPermPolicyLibUpdate(opt *handlers.HandlerOption,
	content *proto.ApplyPermPolicyLibUpdateContent) *ApplicationOfApplyPermPolicyLibUpdate {

	return &ApplicationOfApplyPermPolicyLibUpdate{
		ApplicationBasePermissionPolicyLibrary: permissionpolicylibrary.NewApplicationBasePermPolicyLibrary(
			opt, &content.ApplyPermPolicyLibBaseContent,
		),
		Content: content,
	}
}

// GetItsmApprover returns ITSM approver configuration.
func (a *ApplicationOfApplyPermPolicyLibUpdate) GetItsmApprover(kt *kit.Kit, managers []string) (
	[]itsm.VariableApprover, error) {

	req := protocloud.PermissionTemplateListReq{
		Filter: tools.EqualExpression("id", a.Content.PermissionTemplateID),
		Page:   core.NewDefaultBasePage(),
	}
	template, err := a.Client.DataService().Global.PermissionTemplate.ListPermissionTemplate(kt, &req)
	if err != nil {
		logs.Errorf("list permission template failed, err: %v, id: %s, rid: %s", err,
			a.Content.PermissionTemplateID, kt.Rid)
		return nil, err
	}
	if template.Details == nil || len(template.Details) == 0 {
		logs.Errorf("permission template not found, id: %s, rid: %s", a.Content.PermissionTemplateID, kt.Rid)
		return nil, fmt.Errorf("permission template not found, id: %s", a.Content.PermissionTemplateID)
	}

	return a.GetAccountApprover(kt, template.Details[0].AccountID)
}

// BuildContent builds the application content for the given permission template.
func BuildContent(bkBizID int64, vendor enumor.Vendor, req *proto.BizApplyPermissionPolicyLibraryUpdateReq,
	permissionTemplateID string) *proto.ApplyPermPolicyLibUpdateContent {

	return &proto.ApplyPermPolicyLibUpdateContent{
		ApplyPermPolicyLibBaseContent: proto.ApplyPermPolicyLibBaseContent{
			Action:          enumor.PermPolicyLibActionApplyUpdate,
			Vendor:          vendor,
			BkBizID:         bkBizID,
			PolicyLibraryID: req.PolicyLibraryID,
		},
		PermissionTemplateID: permissionTemplateID,
	}
}
