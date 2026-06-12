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
	"errors"
	"fmt"
	"net/http"
	"strings"
	"time"

	"hcm/cmd/auth-server/types"
	"hcm/pkg/api/core"
	"hcm/pkg/criteria/constant"
	"hcm/pkg/criteria/errf"
	"hcm/pkg/iam/sys"
	"hcm/pkg/logs"
	"hcm/pkg/thirdparty/api-gateway/iam"
	"hcm/pkg/tools/uuid"

	"github.com/emicklei/go-restful/v3"
)

// moduleType auth logic module type.
type moduleType string

const (
	authModule    moduleType = "auth" // auth module.
	initialModule moduleType = "init" // initial hcm auth model in iam module.
	iamModule     moduleType = "iam"  // iam callback module.
)

// restFilter returns auth server restful request filter.
func (s *Service) restFilter() restful.FilterFunction {
	return func(req *restful.Request, resp *restful.Response, chain *restful.FilterChain) {
		r, w := req.Request, resp.ResponseWriter

		var module string
		// path format: /api/{api_version}/{service}/{module}/other
		paths := strings.Split(r.URL.Path, "/")
		if len(paths) > 4 {
			module = paths[4]
		} else {
			logs.Errorf("received url path length not conform to the regulations, path: %s", r.URL.Path)
			fmt.Fprintf(w, errf.New(http.StatusNotFound, "Not Found").Error())
			return
		}

		switch moduleType(module) {
		case iamModule:
			if err := iamRequestFilter(s.client.sys, w, r); err != nil {
				fmt.Fprintf(w, errf.Error(err).Error())
				return
			}

		case authModule:
			if err := authRequestFilter(w, r); err != nil {
				fmt.Fprintf(w, errf.Error(err).Error())
				return
			}

		case initialModule:

		default:
			logs.Errorf("received unkown module's request req: %v", r)
			fmt.Fprintf(w, errf.New(http.StatusNotFound, "Not Found").Error())
			return
		}

		chain.ProcessFilter(req, resp)
	}
}

// iamRequestFilter setups all api filters here. All request would cross here, and we filter request base on URL.
func iamRequestFilter(sysCli *sys.Sys, w http.ResponseWriter, req *http.Request) error {
	isAuthorized, err := checkRequestAuthorization(sysCli, req)
	if err != nil {
		return errf.NewFromErr(http.StatusInternalServerError, err)
	}
	if !isAuthorized {
		return errf.New(types.UnauthorizedErrorCode, "authorized failed")
	}

	rid := getRid(req.Header)
	req.Header.Set(constant.RidKey, rid)

	// set rid to response header, used to troubleshoot the problem.
	w.Header().Set(iam.RequestIDHeader, rid)

	// use sys language as hcm language
	req.Header.Set(constant.LanguageKey, req.Header.Get("Blueking-Language"))

	user := req.Header.Get(constant.UserKey)
	if len(user) == 0 {
		req.Header.Set(constant.UserKey, "auth")
	}

	appCode := req.Header.Get(constant.AppCodeKey)
	if len(appCode) == 0 {
		req.Header.Set(constant.AppCodeKey, iam.SystemIDIAM)
	}

	return nil
}

// getRid get request id from header. if rid is empty, generate a rid to return.
func getRid(h http.Header) string {
	if rid := h.Get(iam.RequestIDHeader); len(rid) != 0 {
		return rid
	}

	if rid := h.Get(constant.RidKey); len(rid) != 0 {
		return rid
	}

	return uuid.UUID()
}

// authRequestFilter set auth request filter.
func authRequestFilter(w http.ResponseWriter, req *http.Request) error {
	// TODO: set auth request filter.

	return nil
}

var iamToken = struct {
	token            string
	tokenRefreshTime time.Time
}{}

func checkRequestAuthorization(cli *sys.Sys, req *http.Request) (bool, error) {
	rid := req.Header.Get(iam.RequestIDHeader)
	name, pwd, ok := req.BasicAuth()
	if !ok || name != iam.SystemIDIAM {
		logs.Errorf("request have no basic authorization, rid: %s", rid)
		return false, nil
	}

	// if cached token is set within a minute, use it to check request authorization
	if iamToken.token != "" && time.Since(iamToken.tokenRefreshTime) <= time.Minute && pwd == iamToken.token {
		return true, nil
	}

	var err error
	iamToken.token, err = cli.GetSystemToken(core.NewBackendKit())
	if err != nil {
		logs.Errorf("check request authorization get system token failed, error: %s, rid: %s", err.Error(), rid)
		return false, err
	}

	iamToken.tokenRefreshTime = time.Now()
	if pwd != iamToken.token {
		return false, errors.New("request password not match system token")
	}

	return true, nil
}
