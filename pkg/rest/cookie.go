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

package rest

import (
	"net/url"

	"hcm/pkg/criteria/constant"

	"github.com/emicklei/go-restful/v3"
)

// GetLanguageByHTTPRequest ...
func GetLanguageByHTTPRequest(req *restful.Request) constant.Language {
	cookie, err := req.Request.Cookie(constant.BKHTTPCookieLanguageKey)
	if err != nil {
		return constant.Chinese
	}
	cookieLanguage, _ := url.QueryUnescape(cookie.Value)
	if cookieLanguage == "" {
		return constant.Chinese
	}

	return constant.Language(cookieLanguage)
}
