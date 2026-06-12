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

// Package authserver defines auth-server api call protocols.
package authserver

import (
	"hcm/pkg/criteria/errf"
	"hcm/pkg/iam/meta"
	"hcm/pkg/rest"
	"hcm/pkg/thirdparty/api-gateway/iam"
)

// InitAuthCenterReq initialize auth center request.
type InitAuthCenterReq struct {
	Host string `json:"host"`
}

// Validate InitAuthCenterReq.
func (r *InitAuthCenterReq) Validate() error {
	if len(r.Host) == 0 {
		return errf.New(errf.InvalidParameter, "host is required")
	}

	return nil
}

// PullResourceResp iam pull resource response.
type PullResourceResp struct {
	rest.BaseResp `json:",inline"`
	Data          interface{} `json:"data"`
}

// AuthorizeBatchReq authorize batch request.
type AuthorizeBatchReq struct {
	User      *meta.UserInfo           `json:"user"`
	Resources []meta.ResourceAttribute `json:"data"`
}

// AuthorizeBatchResp authorize batch response.
type AuthorizeBatchResp struct {
	rest.BaseResp `json:",inline"`
	Data          []meta.Decision `json:"data"`
}

// GetPermissionToApplyReq get permission to apply request.
type GetPermissionToApplyReq struct {
	Resources []meta.ResourceAttribute `json:"resources"`
}

// GetPermissionToApplyResp get permission to apply response.
type GetPermissionToApplyResp struct {
	rest.BaseResp `json:",inline"`
	Data          *meta.IamPermission `json:"data"`
}

// ListAuthorizedInstancesReq list authorized instances request.
type ListAuthorizedInstancesReq struct {
	User   *meta.UserInfo    `json:"user"`
	Type   meta.ResourceType `json:"type"`
	Action meta.Action       `json:"action"`
}

// ListAuthorizedInstancesResp list authorized instances response.
type ListAuthorizedInstancesResp struct {
	rest.BaseResp `json:",inline"`
	Data          *meta.AuthorizedInstances `json:"data"`
}

// RegisterResourceCreatorActionReq register resource creator action request.
type RegisterResourceCreatorActionReq struct {
	Creator  string                             `json:"creator"`
	Instance *meta.RegisterResCreatorActionInst `json:"instance"`
}

// RegisterResourceCreatorActionResp register resource creator action response.
type RegisterResourceCreatorActionResp struct {
	rest.BaseResp `json:",inline"`
	Data          []iam.CreatorActionPolicy `json:"data"`
}

// GetNoAuthSkipUrlResp get iam apply permission url response.
type GetNoAuthSkipUrlResp struct {
	rest.BaseResp `json:",inline"`
	Data          string `json:"data"`
}
