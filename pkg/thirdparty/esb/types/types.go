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

package types

import (
	"net/http"

	"hcm/pkg/cc"
	"hcm/pkg/criteria/constant"
	"hcm/pkg/tools/json"
	"hcm/pkg/tools/uuid"
)

// CommParams defines esb request common parameter
type CommParams struct {
	AppCode   string `json:"bk_app_code"`
	AppSecret string `json:"bk_app_secret"`
	UserName  string `json:"bk_username"`
}

// GetCommParams generate esb request common parameter from esb config and request user
func GetCommParams(config *cc.Esb) *CommParams {
	return &CommParams{
		AppCode:   config.AppCode,
		AppSecret: config.AppSecret,
		UserName:  config.User,
	}
}

// GetCommonHeader 通用Header包括调用ESB所需用户和应用认证、RequestID
func GetCommonHeader(config *cc.Esb) *http.Header {
	return GetCommonHeaderByUser(config, "")
}

// GetCommonHeaderByUser 通用Header包括调用ESB所需用户和应用认证、RequestID，支持传入自定义的username
func GetCommonHeaderByUser(config *cc.Esb, username string) *http.Header {
	h := http.Header{}
	// RequestID
	h.Set(constant.RidKey, uuid.UUID())

	if username == "" {
		username = config.User
	}
	// ESB所需用户和应用认证, Note: json可以确保100%成功的，所以忽略error返回值
	bkApiAuthorization, _ := json.MarshalToString(map[string]string{
		"bk_app_code":   config.AppCode,
		"bk_app_secret": config.AppSecret,
		"bk_username":   username,
	})
	h.Set("X-Bkapi-Authorization", bkApiAuthorization)

	return &h
}

// BaseResponse is esb http base response.
type BaseResponse struct {
	Result  bool   `json:"result"`
	Code    int    `json:"code"`
	Message string `json:"message"`
	Rid     string `json:"request_id"`
}
