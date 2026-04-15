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

// Package applycreate is the handler for apply_permission_policy_library (create action).
package applycreate

import (
	"fmt"

	"hcm/cmd/cloud-server/service/application/handlers"
	"hcm/cmd/cloud-server/service/application/handlers/permission-policy-library"
	proto "hcm/pkg/api/cloud-server/application"
	"hcm/pkg/criteria/enumor"
	"hcm/pkg/kit"
	"hcm/pkg/thirdparty/api-gateway/itsm"
	"hcm/pkg/tools/json"
)

func init() {
	permissionpolicylibrary.RegisterActionHandler(
		enumor.PermPolicyLibActionApplyCreate,
		func(opt *handlers.HandlerOption, base *proto.ApplyPermPolicyLibBaseContent, content string) (
			handlers.ApplicationHandler, error) {

			c := new(proto.ApplyPermPolicyLibCreateContent)
			if err := json.UnmarshalFromString(content, c); err != nil {
				return nil, fmt.Errorf("unmarshal apply perm policy lib create content failed, err: %w", err)
			}
			return NewApplicationOfApplyPermPolicyLibCreate(opt, c), nil
		},
	)
}

// ApplicationOfApplyPermPolicyLibCreate is the handler for apply_permission_policy_library (create action).
type ApplicationOfApplyPermPolicyLibCreate struct {
	permissionpolicylibrary.ApplicationBasePermissionPolicyLibrary

	Content *proto.ApplyPermPolicyLibCreateContent
}

// NewApplicationOfApplyPermPolicyLibCreate creates a new handler.
func NewApplicationOfApplyPermPolicyLibCreate(opt *handlers.HandlerOption,
	content *proto.ApplyPermPolicyLibCreateContent) *ApplicationOfApplyPermPolicyLibCreate {

	return &ApplicationOfApplyPermPolicyLibCreate{
		ApplicationBasePermissionPolicyLibrary: permissionpolicylibrary.NewApplicationBasePermPolicyLibrary(
			opt, &content.ApplyPermPolicyLibBaseContent,
		),
		Content: content,
	}
}

// GetItsmApprover returns ITSM approver configuration.
func (a *ApplicationOfApplyPermPolicyLibCreate) GetItsmApprover(kt *kit.Kit, managers []string) (
	[]itsm.VariableApprover, error) {

	return a.GetAccountApprover(kt, a.Content.AccountID)
}

// BuildContent builds the application content for the given account.
func BuildContent(bkBizID int64, vendor enumor.Vendor, req *proto.BizApplyPermissionPolicyLibraryCreateReq,
	accountID string) *proto.ApplyPermPolicyLibCreateContent {

	return &proto.ApplyPermPolicyLibCreateContent{
		ApplyPermPolicyLibBaseContent: proto.ApplyPermPolicyLibBaseContent{
			Action:          enumor.PermPolicyLibActionApplyCreate,
			Vendor:          vendor,
			BkBizID:         bkBizID,
			PolicyLibraryID: req.PolicyLibraryID,
		},
		AccountID: accountID,
	}
}
