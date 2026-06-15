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

package permissiontemplate

import (
	"fmt"

	"hcm/cmd/cloud-server/service/application/handlers"
	permissionpolicylibrary "hcm/cmd/cloud-server/service/permission-policy-library"
	proto "hcm/pkg/api/cloud-server/application"
	"hcm/pkg/api/core"
	protocloud "hcm/pkg/api/data-service/cloud"
	"hcm/pkg/criteria/enumor"
	"hcm/pkg/criteria/errf"
	"hcm/pkg/dal/dao/tools"
	"hcm/pkg/kit"
	"hcm/pkg/logs"
	"hcm/pkg/thirdparty/api-gateway/itsm"
	"hcm/pkg/tools/json"
)

// OperationHandlerFactory creates an ApplicationHandler for a specific permission template action.
type OperationHandlerFactory func(opt *handlers.HandlerOption, base *proto.BasePermTemplateContent, content string,
) (handlers.ApplicationHandler, error)

var OperationOperationHandlerRegistry = map[enumor.ApplicationOperation]OperationHandlerFactory{}

// RegisterOperationHandler registers a handler factory for the given action.
func RegisterOperationHandler(operation enumor.ApplicationOperation, factory OperationHandlerFactory) {
	OperationOperationHandlerRegistry[operation] = factory
}

// NewHandlerFromApplication dispatches to the registered action handler factory
// based on the action field in the application content.
func NewHandlerFromApplication(opt *handlers.HandlerOption, appContent string) (handlers.ApplicationHandler, error) {
	base := new(proto.BasePermTemplateContent)
	if err := json.UnmarshalFromString(appContent, base); err != nil {
		return nil, fmt.Errorf("unmarshal base permission template content failed, err: %w", err)
	}

	factory, ok := OperationOperationHandlerRegistry[base.Operation]
	if !ok {
		return nil, errf.Newf(errf.InvalidParameter, "no handler registered for action: %s", base.Operation)
	}

	return factory(opt, base, appContent)
}

// ApplicationBasePermissionTemplate is the shared base for all permission template operation handlers.
// Each action-specific handler embeds this base and only implements action-specific methods.
type ApplicationBasePermissionTemplate struct {
	handlers.BaseApplicationHandler
	*permissionpolicylibrary.PolicyLibraryApplier

	bkBizID int64
}

// NewApplicationBasePermissionTemplate creates a new base handler.
func NewApplicationBasePermissionTemplate(opt *handlers.HandlerOption,
	base *proto.BasePermTemplateContent) ApplicationBasePermissionTemplate {

	return ApplicationBasePermissionTemplate{
		BaseApplicationHandler: handlers.NewBaseApplicationHandler(
			opt, enumor.OperatePermissionTemplate, base.Operation, base.Vendor,
		),
		PolicyLibraryApplier: permissionpolicylibrary.NewPolicyLibraryApplier(opt.Client, opt.Audit),
		bkBizID:              base.BkBizID,
	}
}

// BkBizID returns the business ID.
func (a *ApplicationBasePermissionTemplate) BkBizID() int64 {
	return a.bkBizID
}

// PrepareReq no pre-processing needed.
func (a *ApplicationBasePermissionTemplate) PrepareReq() error {
	return nil
}

// PrepareReqFromContent no pre-processing needed when restoring from DB content.
func (a *ApplicationBasePermissionTemplate) PrepareReqFromContent() error {
	return nil
}

// GetBkBizIDs returns the business IDs for this application.
func (a *ApplicationBasePermissionTemplate) GetBkBizIDs() []int64 {
	return []int64{a.bkBizID}
}

// GetItsmApproverByTemplateID gets the itsm approver by template ID.
func (a *ApplicationBasePermissionTemplate) GetItsmApproverByTemplateID(kt *kit.Kit, id string) (
	[]itsm.VariableApprover, error) {

	req := protocloud.PermissionTemplateListReq{
		Filter: tools.EqualExpression("id", id),
		Page:   core.NewDefaultBasePage(),
	}
	template, err := a.Client.DataService().Global.PermissionTemplate.ListPermissionTemplate(kt, &req)
	if err != nil {
		logs.Errorf("list permission template failed, err: %v, id: %s, rid: %s", err, id, kt.Rid)
		return nil, err
	}
	if template.Details == nil || len(template.Details) == 0 {
		logs.Errorf("permission template not found, id: %s, rid: %s", id, kt.Rid)
		return nil, fmt.Errorf("permission template not found, id: %s", id)
	}

	return a.GetAccountApprover(kt, template.Details[0].AccountID)
}
