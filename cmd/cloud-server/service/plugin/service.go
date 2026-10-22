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

// Package plugin defines iac plugin management service for cloud-server.
package plugin

import (
	"net/http"

	"hcm/cmd/cloud-server/service/capability"
	"hcm/pkg/client"
	"hcm/pkg/iam/auth"
	"hcm/pkg/rest"
)

// InitService initializes the iac plugin service routes.
func InitService(c *capability.Capability) {
	s := &svc{
		client:     c.ApiClient,
		authorizer: c.Authorizer,
	}

	h := rest.NewHandler()

	h.Add("ListIaCPlugin", http.MethodPost, "/iac_plugins/list", s.ListIaCPlugin)
	h.Add("SwitchIaCPlugin", http.MethodPost, "/iac_plugins/switch", s.SwitchIaCPlugin)
	h.Add("RegisterIaCPlugin", http.MethodPost, "/iac_plugins/register", s.RegisterIaCPlugin)

	h.Add("UpgradeIaCPlugin", http.MethodPost, "/iac_plugins/upgrade", s.UpgradeIaCPlugin)
	h.Add("UpdateIaCPlugin", http.MethodPost, "/iac_plugins/update", s.UpdateIaCPlugin)

	h.Load(c.WebService)
}

type svc struct {
	client     *client.ClientSet
	authorizer auth.Authorizer
}
