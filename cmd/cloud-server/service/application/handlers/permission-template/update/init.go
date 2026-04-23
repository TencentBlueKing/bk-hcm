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

package updatepermtemplate

import (
	"fmt"

	"hcm/cmd/cloud-server/service/application/handlers"
	"hcm/cmd/cloud-server/service/application/handlers/permission-template"
	proto "hcm/pkg/api/cloud-server/application"
	"hcm/pkg/criteria/enumor"
	"hcm/pkg/kit"
	"hcm/pkg/thirdparty/api-gateway/itsm"
	"hcm/pkg/tools/json"
)

var _ handlers.ApplicationHandler = (*ApplicationOfUpdatePermTemplate)(nil)

func init() {
	permissiontemplate.RegisterActionHandler(enumor.PermTemplateActionUpdate, newHandlerFromContent)
}

// updatePermTemplateContent is the full content stored in the application record for update action.
type updatePermTemplateContent struct {
	proto.BasePermTemplateContent `json:",inline"`

	// ID is the existing permission template ID to update.
	ID              string  `json:"id"`
	PolicyLibraryID string  `json:"policy_library_id"`
	Memo            *string `json:"memo"`
}

func newHandlerFromContent(opt *handlers.HandlerOption, base *proto.BasePermTemplateContent, content string,
) (handlers.ApplicationHandler, error) {

	ct := new(updatePermTemplateContent)
	if err := json.UnmarshalFromString(content, ct); err != nil {
		return nil, fmt.Errorf("unmarshal update permission template content failed, err: %w", err)
	}

	return newApplicationFromContent(opt, base, ct), nil
}

// ApplicationOfUpdatePermTemplate is the handler for operate_permission_template (update action).
type ApplicationOfUpdatePermTemplate struct {
	permissiontemplate.ApplicationBasePermissionTemplate

	content *updatePermTemplateContent
}

// NewApplicationOfUpdatePermTemplate creates a new handler from an HTTP request.
func NewApplicationOfUpdatePermTemplate(opt *handlers.HandlerOption, base *proto.BasePermTemplateContent,
	req *proto.BizUpdatePermissionTemplateReq) *ApplicationOfUpdatePermTemplate {

	ct := &updatePermTemplateContent{
		BasePermTemplateContent: *base,
		ID:                      req.ID,
		PolicyLibraryID:         req.PolicyLibraryID,
		Memo:                    req.Memo,
	}

	return newApplicationFromContent(opt, base, ct)
}

func newApplicationFromContent(opt *handlers.HandlerOption, base *proto.BasePermTemplateContent,
	ct *updatePermTemplateContent) *ApplicationOfUpdatePermTemplate {

	return &ApplicationOfUpdatePermTemplate{
		ApplicationBasePermissionTemplate: permissiontemplate.NewApplicationBasePermissionTemplate(opt, base),
		content:                           ct,
	}
}

// GetItsmApprover returns ITSM approver configuration.
func (a *ApplicationOfUpdatePermTemplate) GetItsmApprover(kt *kit.Kit, managers []string) (
	[]itsm.VariableApprover, error) {

	return a.GetItsmApproverByTemplateID(kt, a.content.ID)
}
