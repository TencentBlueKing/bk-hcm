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
	"hcm/cmd/cloud-server/service/application/handlers"
	"hcm/cmd/cloud-server/service/application/handlers/permission-policy-library"
	proto "hcm/pkg/api/cloud-server/application"
	"hcm/pkg/criteria/enumor"
)

func init() {
	permissionpolicylibrary.RegisterActionHandler(
		enumor.PermPolicyLibActionApplyUpdate,
		func(opt *handlers.HandlerOption, content *proto.ApplyPermPolicyLibContent) handlers.ApplicationHandler {
			return NewApplicationOfApplyPermPolicyLibUpdate(opt, content)
		},
	)
}

// ApplicationOfApplyPermPolicyLibUpdate is the handler for apply_permission_policy_library (update action).
type ApplicationOfApplyPermPolicyLibUpdate struct {
	permissionpolicylibrary.ApplicationBasePermissionPolicyLibrary
}

// NewApplicationOfApplyPermPolicyLibUpdate creates a new handler.
func NewApplicationOfApplyPermPolicyLibUpdate(opt *handlers.HandlerOption,
	content *proto.ApplyPermPolicyLibContent) *ApplicationOfApplyPermPolicyLibUpdate {

	return &ApplicationOfApplyPermPolicyLibUpdate{
		ApplicationBasePermissionPolicyLibrary: permissionpolicylibrary.NewApplicationBasePermPolicyLibrary(
			opt, content,
		),
	}
}

// BuildContent builds the application content for the given account.
func BuildContent(bkBizID int64, vendor enumor.Vendor, req *proto.BizApplyPermissionPolicyLibraryUpdateReq,
	accountID string) *proto.ApplyPermPolicyLibContent {

	return &proto.ApplyPermPolicyLibContent{
		Action:          enumor.PermPolicyLibActionApplyUpdate,
		Vendor:          vendor,
		BkBizID:         bkBizID,
		PolicyLibraryID: req.PolicyLibraryID,
		AccountID:       accountID,
	}
}
