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

// Package auth ...
package auth

import (
	"hcm/cmd/web-server/service/capability"
	webserver "hcm/pkg/api/web-server"
	"hcm/pkg/client"
	"hcm/pkg/iam/auth"
	"hcm/pkg/iam/meta"
	"hcm/pkg/logs"
	"hcm/pkg/rest"
)

// InitAuthService initial the authSvc service
func InitAuthService(c *capability.Capability) {
	svr := &authSvc{
		client:     c.ApiClient,
		authorizer: c.Authorizer,
	}

	h := rest.NewHandler()
	h.Add("AuthVerify", "POST", "/auth/verify", svr.AuthVerify)
	h.Add("GetApplyPermUrl", "POST", "/auth/find/apply_perm_url", svr.GetApplyPermUrl)

	h.Load(c.WebService)
}

type authSvc struct {
	client     *client.ClientSet
	authorizer auth.Authorizer
}

// AuthVerify verify if auth is authorized to specified resources.
func (a *authSvc) AuthVerify(cts *rest.Contexts) (interface{}, error) {
	input := new(webserver.AuthVerifyReq)
	if err := cts.DecodeInto(input); err != nil {
		return nil, err
	}

	anyAuthAttrs := make([]meta.ResourceAttribute, 0)
	exactAuthAttrs := make([]meta.ResourceAttribute, 0)

	exactAuthMap := make(map[int]struct{})
	for idx, res := range input.Resources {
		attr := meta.ResourceAttribute{
			Basic: &meta.Basic{
				Type:       meta.ResourceType(res.ResourceType),
				Action:     meta.Action(res.Action),
				ResourceID: res.ResourceID,
			},
			BizID: res.BizID,
		}

		// check whether resource needs exact authorization or not
		if len(res.ResourceID) > 0 || res.BizID > 0 {
			exactAuthAttrs = append(exactAuthAttrs, attr)
			exactAuthMap[idx] = struct{}{}
		} else {
			anyAuthAttrs = append(anyAuthAttrs, attr)
		}
	}

	resourceLen := len(input.Resources)
	resources := make([]webserver.AuthVerifyRes, resourceLen)
	unauthorizedRes := make([]meta.ResourceAttribute, 0)

	if len(exactAuthAttrs) > 0 {
		verifyResults, _, err := a.authorizer.Authorize(cts.Kit, exactAuthAttrs...)
		if err != nil {
			logs.Errorf("authorize failed, err: %v, attrs: %#v, rid: %s", err, exactAuthAttrs, cts.Kit.Rid)
			return nil, err
		}

		index := 0
		for i := 0; i < resourceLen; i++ {
			if _, exists := exactAuthMap[i]; exists {
				resources[i].Authorized = verifyResults[index].Authorized
				if !verifyResults[index].Authorized {
					unauthorizedRes = append(unauthorizedRes, exactAuthAttrs[index])
				}
				index++
			}
		}
	}

	if len(anyAuthAttrs) > 0 {
		verifyResults, err := a.authorizer.AuthorizeAny(cts.Kit, anyAuthAttrs...)
		if err != nil {
			logs.Errorf("authorize any failed, err: %v, attrs: %#v, rid: %s", err, anyAuthAttrs, cts.Kit.Rid)
			return nil, err
		}

		index := 0
		for i := 0; i < resourceLen; i++ {
			if _, exists := exactAuthMap[i]; !exists {
				resources[i].Authorized = verifyResults[index].Authorized
				if !verifyResults[index].Authorized {
					unauthorizedRes = append(unauthorizedRes, anyAuthAttrs[index])
				}
				index++
			}
		}
	}

	if len(unauthorizedRes) == 0 {
		return &webserver.AuthVerifyResp{Results: resources}, nil
	}

	permission, err := a.authorizer.GetPermissionToApply(cts.Kit, unauthorizedRes...)
	if err != nil {
		return nil, err
	}

	return &webserver.AuthVerifyResp{Results: resources, Permission: permission}, nil
}

// GetApplyPermUrl get iam apply permission url for front end to redirect auth to it.
func (a *authSvc) GetApplyPermUrl(cts *rest.Contexts) (interface{}, error) {
	perm := new(meta.IamPermission)
	if err := cts.DecodeInto(perm); err != nil {
		return nil, err
	}

	url, err := a.authorizer.GetApplyPermUrl(cts.Kit, perm)
	if err != nil {
		logs.Errorf("get iam apply permission url failed, err: %v, rid: %s", err, cts.Kit.Rid)
		return nil, err
	}

	return url, nil
}
