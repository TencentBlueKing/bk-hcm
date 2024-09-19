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

package notice

import (
	"errors"
	"net/http"

	"hcm/cmd/web-server/service/capability"
	"hcm/pkg/cc"
	"hcm/pkg/rest"
	pkgnotice "hcm/pkg/thirdparty/api-gateway/notice"
)

// InitService initialize the load balancer service.
func InitService(c *capability.Capability) {
	svc := &service{
		client: c.NoticeCli,
	}

	h := rest.NewHandler()

	h.Add("GetCurrentAnnouncements", http.MethodGet,
		"/notice/current_announcements", svc.GetCurrentAnnouncements)

	h.Load(c.WebService)
}

type service struct {
	client pkgnotice.Client
}

// GetCurrentAnnouncements ...
func (s service) GetCurrentAnnouncements(cts *rest.Contexts) (interface{}, error) {
	if !cc.WebServer().Notice.Enable {
		return nil, errors.New("notification is not enabled")
	}
	params := make(map[string]string)
	for key, val := range cts.Request.Request.URL.Query() {
		params[key] = val[0]
	}
	params["platform"] = cc.WebServer().Notice.AppCode
	language := rest.GetLanguageByHTTPRequest(cts.Request)
	params["language"] = string(language)
	return s.client.GetCurAnn(cts.Kit, params)
}
