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

package permissionpolicylibrary

import (
	"fmt"

	"hcm/cmd/cloud-server/service/application/handlers"
	"hcm/cmd/cloud-server/service/permission-policy-library"
	proto "hcm/pkg/api/cloud-server/application"
	"hcm/pkg/criteria/enumor"
	"hcm/pkg/criteria/errf"
	"hcm/pkg/thirdparty/api-gateway/itsm"
	"hcm/pkg/tools/json"
)

// ActionHandlerFactory is a factory function that creates an ApplicationHandler for a given content.
type ActionHandlerFactory func(opt *handlers.HandlerOption,
	content *proto.ApplyPermPolicyLibContent) handlers.ApplicationHandler

var actionHandlerRegistry = map[enumor.PermPolicyLibAction]ActionHandlerFactory{}

// RegisterActionHandler registers a handler factory for the given action.
func RegisterActionHandler(action enumor.PermPolicyLibAction, factory ActionHandlerFactory) {
	actionHandlerRegistry[action] = factory
}

// NewHandlerFromApplication creates a handler from application content using the registry.
func NewHandlerFromApplication(opt *handlers.HandlerOption, appContent string) (
	handlers.ApplicationHandler, error) {

	content := new(proto.ApplyPermPolicyLibContent)
	if err := json.UnmarshalFromString(appContent, content); err != nil {
		return nil, fmt.Errorf("unmarshal apply permission policy library content failed, err: %w", err)
	}

	factory, ok := actionHandlerRegistry[content.Action]
	if !ok {
		return nil, errf.Newf(errf.InvalidParameter, "no handler registered for action: %s", content.Action)
	}

	return factory(opt, content), nil
}

// ApplicationBasePermissionPolicyLibrary is the base handler for permission policy library applications.
// It embeds BaseApplicationHandler and stores the application content.
type ApplicationBasePermissionPolicyLibrary struct {
	handlers.BaseApplicationHandler
	*permissionpolicylibrary.PolicyLibraryApplier

	Content *proto.ApplyPermPolicyLibContent
}

// NewApplicationBasePermPolicyLibrary creates a new base handler.
func NewApplicationBasePermPolicyLibrary(opt *handlers.HandlerOption,
	content *proto.ApplyPermPolicyLibContent) ApplicationBasePermissionPolicyLibrary {

	return ApplicationBasePermissionPolicyLibrary{
		BaseApplicationHandler: handlers.NewBaseApplicationHandler(
			opt, enumor.ApplyPermissionPolicyLibrary, content.Vendor,
		),
		PolicyLibraryApplier: permissionpolicylibrary.NewPolicyLibraryApplier(opt.Client, opt.Audit),
		Content:              content,
	}
}

// PrepareReq prepares the request data (no-op for this handler).
func (a *ApplicationBasePermissionPolicyLibrary) PrepareReq() error {
	return nil
}

// PrepareReqFromContent prepares the request from DB content (no-op for this handler).
func (a *ApplicationBasePermissionPolicyLibrary) PrepareReqFromContent() error {
	return nil
}

// GetBkBizIDs returns the business IDs for this application.
func (a *ApplicationBasePermissionPolicyLibrary) GetBkBizIDs() []int64 {
	return []int64{a.Content.BkBizID}
}

// GetItsmApprover returns ITSM approver configuration.
func (a *ApplicationBasePermissionPolicyLibrary) GetItsmApprover(managers []string) []itsm.VariableApprover {
	return []itsm.VariableApprover{
		{
			Variable:  "account_manager",
			Approvers: managers,
		},
	}
}
