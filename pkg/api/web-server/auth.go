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

// Package webserver defines api-server api call protocols.
package webserver

import "hcm/pkg/iam/meta"

// AuthVerifyReq auth verify request.
type AuthVerifyReq struct {
	Resources []AuthVerifyResource `json:"resources"`
}

// AuthVerifyResource auth verify resource.
type AuthVerifyResource struct {
	BizID        int64  `json:"bk_biz_id"`
	ResourceType string `json:"resource_type"`
	ResourceID   string `json:"resource_id"`
	Action       string `json:"action"`
}

// AuthVerifyRes auth verify result for one resource.
type AuthVerifyRes struct {
	Authorized bool `json:"authorized"`
}

// AuthVerifyResp auth verify response.
type AuthVerifyResp struct {
	Results    []AuthVerifyRes     `json:"results"`
	Permission *meta.IamPermission `json:"permission"`
}
