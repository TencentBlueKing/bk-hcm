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

	"hcm/cmd/cloud-server/service/application/handlers"
	"hcm/cmd/cloud-server/service/application/handlers/permission-template"
	proto "hcm/pkg/api/cloud-server/application"
	"hcm/pkg/criteria/enumor"
	"hcm/pkg/kit"
	"hcm/pkg/thirdparty/api-gateway/itsm"
	"hcm/pkg/tools/json"
)

var _ handlers.ApplicationHandler = (*ApplicationOfCreatePermTemplate)(nil)

func init() {
	permissiontemplate.RegisterOperationHandler(enumor.OpCreatePermTemplate, newHandlerFromContent)
}

// createPermTemplateContent is the full content stored in the application record for create action.
type createPermTemplateContent struct {
	proto.BasePermTemplateContent `json:",inline"`

	AccountID       string  `json:"account_id"`
	PolicyLibraryID string  `json:"policy_library_id"`
	Name            string  `json:"name"`
	Memo            *string `json:"memo"`
}

func newHandlerFromContent(opt *handlers.HandlerOption, base *proto.BasePermTemplateContent, content string,
) (handlers.ApplicationHandler, error) {

	ct := new(createPermTemplateContent)
	if err := json.UnmarshalFromString(content, ct); err != nil {
		return nil, fmt.Errorf("unmarshal create permission template content failed, err: %w", err)
	}

	return newApplicationFromContent(opt, base, ct), nil
}

// ApplicationOfCreatePermTemplate is the handler for operate_permission_template (create action).
type ApplicationOfCreatePermTemplate struct {
	permissiontemplate.ApplicationBasePermissionTemplate

	content *createPermTemplateContent
}

// NewApplicationOfCreatePermTemplate creates a new handler from an HTTP request.
func NewApplicationOfCreatePermTemplate(opt *handlers.HandlerOption, base *proto.BasePermTemplateContent,
	req *proto.BizCreatePermissionTemplateReq) *ApplicationOfCreatePermTemplate {

	ct := &createPermTemplateContent{
		BasePermTemplateContent: *base,
		AccountID:               req.AccountID,
		PolicyLibraryID:         req.PolicyLibraryID,
		Name:                    req.Name,
		Memo:                    req.Memo,
	}

	return newApplicationFromContent(opt, base, ct)
}

func newApplicationFromContent(opt *handlers.HandlerOption, base *proto.BasePermTemplateContent,
	ct *createPermTemplateContent) *ApplicationOfCreatePermTemplate {

	return &ApplicationOfCreatePermTemplate{
		ApplicationBasePermissionTemplate: permissiontemplate.NewApplicationBasePermissionTemplate(opt, base),
		content:                           ct,
	}
}

// GetItsmApprover returns ITSM approver configuration.
func (a *ApplicationOfCreatePermTemplate) GetItsmApprover(kt *kit.Kit, managers []string) (
	[]itsm.VariableApprover, error) {

	return a.GetAccountApprover(kt, a.content.AccountID)
}
