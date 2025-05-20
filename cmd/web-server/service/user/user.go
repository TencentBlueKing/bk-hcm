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

// Package user ...
package user

import (
	"hcm/cmd/web-server/service/capability"
	"hcm/pkg/cc"
	"hcm/pkg/client"
	"hcm/pkg/criteria/constant"
	"hcm/pkg/rest"
	"hcm/pkg/thirdparty/api-gateway/login"
)

// InitUserService initial the userSvc service
func InitUserService(c *capability.Capability) {
	bkLoginCookieName := cc.WebServer().Web.BkLoginCookieName
	if bkLoginCookieName == "" {
		bkLoginCookieName = constant.BKToken
	}
	svr := &userSvc{
		client:            c.ApiClient,
		loginCli:          c.LoginCli,
		bkLoginCookieName: bkLoginCookieName,
	}

	h := rest.NewHandler()
	h.Add("GetUser", "GET", "/users", svr.GetUser)

	h.Load(c.WebService)
}

type userSvc struct {
	client            *client.ClientSet
	loginCli          login.Client
	bkLoginCookieName string
}

// GetUser get user info
func (u *userSvc) GetUser(cts *rest.Contexts) (interface{}, error) {
	cookie, err := cts.Request.Request.Cookie(u.bkLoginCookieName)
	if err != nil {
		return nil, err
	}
	cts.Kit.SetBackendTenantID()
	return u.loginCli.GetUserByToken(cts.Kit, cookie.Value)
}
