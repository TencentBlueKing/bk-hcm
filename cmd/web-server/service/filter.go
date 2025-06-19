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

package service

import (
	"bytes"
	"fmt"
	"io/ioutil"
	"net/http"
	"regexp"
	"strings"

	"hcm/pkg/api/core"
	"hcm/pkg/criteria/constant"
	"hcm/pkg/criteria/errf"
	"hcm/pkg/kit"
	"hcm/pkg/logs"
	"hcm/pkg/rest"
	"hcm/pkg/thirdparty/api-gateway/login"
	"hcm/pkg/tools/uuid"

	"github.com/emicklei/go-restful/v3"
)

type loginVerifyRespData struct {
	UserName string `json:"username"`
	TenantID string `json:"tenant_id"`
}

func isITSMCallbackRequest(req *restful.Request) bool {
	if strings.HasSuffix(req.Request.RequestURI, "/api/v1/cloud/applications/approve") &&
		req.Request.Method == http.MethodPost {
		return true
	}
	return false
}

// 系统管理操作仅能从后台发起
func isSystemAdminRequest(req *restful.Request) bool {
	return strings.Contains(req.Request.RequestURI, "/api/v1/cloud/admin/system")
}

func newCheckLogin(loginCli login.Client, bkLoginUrl, bkLoginCookieName string) func(
	*restful.Request) (*rest.Response, error) {

	if bkLoginCookieName == constant.BKTicket {
		// 解析Login URL
		oaLoginClient, err := newOALoginClient(bkLoginUrl)
		if err != nil {
			// 登录有问题，则启动没意义
			panic(err)
		}

		return func(req *restful.Request) (*rest.Response, error) {
			// 获取cookie
			cookie, err := req.Request.Cookie(bkLoginCookieName)
			// Note: err只有一个ErrNoCookie可能，所以这里是无登录票据的情况
			if err != nil || cookie.Value == "" {
				return nil, fmt.Errorf("%s cookie don't exists", bkLoginCookieName)
			}
			// 校验bk_token是否有效
			ret, err := oaLoginClient.Verify(req.Request.Context(), cookie.Value)
			if err != nil {
				return nil, err
			}
			return ret, nil
		}
	}

	// 默认只能是bk_token,不支持其他的
	bkLoginCookieName = constant.BKToken

	return func(req *restful.Request) (*rest.Response, error) {
		// 获取cookie
		cookie, err := req.Request.Cookie(bkLoginCookieName)
		// Note: err只有一个ErrNoCookie可能，所以这里是无登录票据的情况
		if err != nil || cookie.Value == "" {
			return nil, fmt.Errorf("%s cookie don't exists", bkLoginCookieName)
		}
		// 校验bk_token是否有效
		kt := core.NewBackendKit()
		kt.SetBackendTenantID()
		resp, err := loginCli.VerifyToken(kt, cookie.Value)
		if err != nil {
			logs.Errorf("verify token failed, err: %v, cookie value: %s, rid: %s", err, cookie.Value, kt.Rid)
			return nil, err
		}
		return &rest.Response{
			Data: loginVerifyRespData{
				UserName: resp.Username,
				TenantID: resp.TenantID,
			},
		}, nil
	}
}

// NewUserAuthenticateFilter ...
func NewUserAuthenticateFilter(loginCli login.Client, bkLoginUrl, bkLoginCookieName string) restful.FilterFunction {

	checkLogin := newCheckLogin(loginCli, bkLoginUrl, bkLoginCookieName)

	return func(req *restful.Request, resp *restful.Response, chain *restful.FilterChain) {
		var err error
		username := ""
		// 系统管理操作仅能从后台发起，外部请求直接拒绝
		if isSystemAdminRequest(req) {
			resp.WriteErrorString(http.StatusForbidden, "system admin request can not be called outside hcm system")
			return
		}
		// 对于itsm 的回调请求，不能用户认证，而是处理请求时进行单独的Token认证，这里直接通过
		if isITSMCallbackRequest(req) {
			username = "itsm_callback"
		} else {
			ret, err := checkLogin(req)
			if err != nil {
				resp.WriteAsJson(rest.BaseResp{
					Code:    errf.UserNoAppAccess,
					Message: errf.Error(err).Message,
				})
				return
			}
			if ret != nil {
				dataContent, ok := ret.Data.(loginVerifyRespData)
				if ok {
					username = dataContent.UserName
					req.Request.Header.Set(constant.TenantIDKey, dataContent.TenantID)
				} else {
					logs.Errorf("change ret data to loginVerifyRespData failed")
				}
			}
		}

		// 这里直接修改请求的Header，后面需要用，可以直接从Header头里取
		req.Request.Header.Set(constant.UserKey, username)
		req.Request.Header.Set(constant.AppCodeKey, constant.WebSourceAppCode)

		// 使用Kit便于校验通用的Header是否满足
		kt, err := kit.FromHeader(req.Request.Context(), req.Request.Header)
		if err != nil {
			resp.WriteError(http.StatusForbidden, err)
			return
		}

		body, err := peekRequest(req.Request)
		if err != nil {
			resp.WriteError(http.StatusForbidden, err)
			logs.Errorf("peek request failed, err: %v, rid: %s", err, kt.Rid)
			return
		}
		// request and response details landing log for monitoring and troubleshooting problem.
		logs.Infof("uri: %s, method: %s, body: %s, appcode: %s, user: %s, remote addr: %s, "+
			"rid: %s", req.Request.RequestURI, req.Request.Method, body, kt.AppCode, kt.User,
			req.Request.RemoteAddr, kt.Rid)

		chain.ProcessFilter(req, resp)
	}
}

func peekRequest(req *http.Request) (string, error) {
	if req.Body != nil {
		byt, err := ioutil.ReadAll(req.Body)
		if err != nil {
			return "", err
		}

		req.Body = ioutil.NopCloser(bytes.NewBuffer(byt))

		reg := regexp.MustCompile("\\s+")
		str := reg.ReplaceAllString(string(byt), "")
		return str, nil
	}

	return "", nil
}

// NewCompleteRequestIDFilter creates a filter that adds a request ID if request ID is missing.
func NewCompleteRequestIDFilter() restful.FilterFunction {
	return func(req *restful.Request, resp *restful.Response, chain *restful.FilterChain) {
		rid := req.Request.Header.Get(constant.RidKey)
		if rid == "" {
			req.Request.Header.Set(constant.RidKey, uuid.UUID())
		}
		chain.ProcessFilter(req, resp)
	}
}
