/*
 * TencentBlueKing is pleased to support the open source community by making
 * 成本服务中心 (Cost Optimization Service Center) available.
 * Copyright (C) 2022 THL A29 Limited, a Tencent company. All rights reserved.
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

package apigateway

import (
	"encoding/json"
	"fmt"
	"net/http"

	"hcm/pkg/cc"
	"hcm/pkg/criteria/constant"
	"hcm/pkg/kit"
	"hcm/pkg/logs"
	"hcm/pkg/rest"
	"hcm/pkg/thirdparty/api-gateway/bkuser"
	"hcm/pkg/tools/util"
)

// BaseResponse is esb http base response.
type BaseResponse struct {
	Result  bool   `json:"result"`
	Code    int    `json:"code"`
	Message string `json:"message"`
}

// ApiGatewayResp ...
type ApiGatewayResp[T any] struct {
	Result         bool     `json:"result"`
	Code           any      `json:"code"`
	BKErrorCode    int      `json:"bk_error_code"`
	Message        string   `json:"message"`
	BKErrorMessage string   `json:"bk_error_msg"`
	Data           T        `json:"data"`
	Error          ApiError `json:"error"`
}

// ApiError api错误响应且状态码>=500时返回的error结构体的完整结构
// 目前login接口和cmsi接口使用该结构体
type ApiError struct {
	Code    string                 `json:"code"`
	Message string                 `json:"message"`
	ErrData map[string]interface{} `json:"data"`
	Details []interface{}          `json:"details"`
}

// ApiGatewayCall general call helper function for api gateway
func ApiGatewayCall[IT any, OT any](cli rest.ClientInterface, bkUserCli bkuser.Client, cfg *cc.ApiGateway,
	method rest.VerbType, kt *kit.Kit, req *IT, url string, urlParams ...any) (*OT, error) {

	header := GetCommonHeader(kt, bkUserCli, cfg)
	resp := new(ApiGatewayResp[*OT])
	err := cli.Verb(method).
		SubResourcef(url, urlParams...).
		WithContext(kt.Ctx).
		WithHeaders(header).
		Body(req).
		Do().Into(resp)

	if err != nil {
		logs.Errorf("fail to call api gateway api, err: %v, url: %s, rid: %s", err, url, kt.Rid)
		return nil, err
	}

	// 兼容code为string类型的情况（itsm_v4）
	var codeInt int
	if resp.Code != nil {
		codeInt, err = util.GetIntByInterface(resp.Code)
		if err != nil {
			logs.Errorf("failed to convert code to int, code: %v, url: %s, rid: %s", resp.Code, url, kt.Rid)
			return nil, err
		}
	}

	if !resp.Result || codeInt != 0 {
		err := fmt.Errorf("failed to call api gateway, code: %d, msg: %s, bk_error_code: %d, bk_error_msg: %s",
			resp.Code, resp.Message, resp.BKErrorCode, resp.BKErrorMessage)
		logs.Errorf("api gateway returns error, url: %s, err: %v, rid: %s", url, err, kt.Rid)
		return nil, err
	}
	return resp.Data, nil
}

// ApiGatewayCallWithRichError 该apigw辅助调用函数的适用场景如下：
// 成功响应时只返回 data 结构体，失败响应且状态码≥500时只返回带有更完整错误信息的 error 结构体
func ApiGatewayCallWithRichError[IT any, OT any](cli rest.ClientInterface, bkUserCli bkuser.Client, cfg *cc.ApiGateway,
	method rest.VerbType, kt *kit.Kit, req *IT, url string, urlParams ...any) (*OT, error) {

	header := GetCommonHeader(kt, bkUserCli, cfg)
	resp := new(ApiGatewayResp[*OT])
	resp.Result = true

	// Into函数本身会将基本网络错误打印出日志
	r := cli.Verb(method).
		SubResourcef(url, urlParams...).
		WithContext(kt.Ctx).
		WithHeaders(header).
		Body(req).
		Do()

	err := r.Into(resp)
	if err != nil {
		logs.Errorf("fail to call api gateway api, err: %v, url: %s, rid: %s", err, url, kt.Rid)
		return nil, err
	}

	// 兼容code为string类型的情况（itsm_v4）
	var codeInt int
	if resp.Code != nil {
		codeInt, err = util.GetIntByInterface(resp.Code)
		if err != nil {
			logs.Errorf("failed to convert code to int, code: %v, url: %s, rid: %s", resp.Code, url, kt.Rid)
			return nil, err
		}
	}

	if !resp.Result || codeInt != 0 {
		err := fmt.Errorf("failed to call api gateway, code: %d, msg: %s, bk_error_code: %d, bk_error_msg: %s",
			resp.Code, resp.Message, resp.BKErrorCode, resp.BKErrorMessage)
		logs.Errorf("api gateway returns error, url: %s, err: %v, rid: %s", url, err, kt.Rid)
		return nil, err
	}

	if r.StatusCode >= 500 { // api执行错误
		errData, err := json.Marshal(resp.Error.ErrData)
		if err != nil {
			logs.Errorf("failed to marshal error data, err: %v, url: %s, errData: %s, rid: %s", err, url,
				resp.Error.ErrData, kt.Rid)
			return nil, fmt.Errorf("failed to marshal error data: %v", err)
		}

		errDetails, err := json.Marshal(resp.Error.Details)
		if err != nil {
			logs.Errorf("failed to marshal error details, err: %v, url: %s, errDetails: %s, rid: %s", err, url,
				resp.Error.Details, kt.Rid)
			return nil, fmt.Errorf("failed to marshal error details: %v", err)
		}

		err = fmt.Errorf("failed to call api, code: %d, msg: %s, data: %s, details: %s",
			resp.Error.Code, resp.Error.Message, errData, errDetails)
		logs.Errorf("api returns error, url: %s, err: %v, rid: %s", url, err, kt.Rid)
		return nil, err
	}

	return resp.Data, nil
}

// ApiGatewayCallOriginal general call helper function for api gateway
// 该方法不会处理接口返回的code，适用于code非标准的场景，在上层自行处理
func ApiGatewayCallOriginal[IT any, OT any](cli rest.ClientInterface, bkUserCli bkuser.Client, cfg *cc.ApiGateway,
	method rest.VerbType, kt *kit.Kit, req *IT, url string, urlParams ...any) (int, string, *OT, error) {

	header := GetCommonHeader(kt, bkUserCli, cfg)
	resp := new(ApiGatewayResp[*OT])
	err := cli.Verb(method).
		SubResourcef(url, urlParams...).
		WithContext(kt.Ctx).
		WithHeaders(header).
		Body(req).
		Do().Into(resp)

	if err != nil {
		logs.Errorf("fail to call api gateway api, err: %v, url: %s, rid: %s", err, url, kt.Rid)
		return 0, "", nil, err
	}

	// 兼容code为string类型的情况（itsm_v4）
	var codeInt int
	if resp.Code != nil {
		codeInt, err = util.GetIntByInterface(resp.Code)
		if err != nil {
			logs.Errorf("failed to convert code to int, code: %v, url: %s, rid: %s", resp.Code, url, kt.Rid)
			return 0, "", nil, err
		}
	}

	if !resp.Result {
		err := fmt.Errorf("failed to call api gateway, code: %d, msg: %s, bk_error_code: %d, bk_error_msg: %s",
			resp.Code, resp.Message, resp.BKErrorCode, resp.BKErrorMessage)
		logs.Errorf("api gateway returns error, url: %s, err: %v, rid: %s", url, err, kt.Rid)
		return 0, "", nil, err
	}
	return codeInt, resp.Message, resp.Data, nil
}

// ApiGatewayCallWithoutReq general call helper function for api gateway
func ApiGatewayCallWithoutReq[OT any](cli rest.ClientInterface, bkUserCli bkuser.Client, cfg *cc.ApiGateway,
	method rest.VerbType, kt *kit.Kit, params map[string]string, url string, urlParams ...any) (*OT, error) {

	header := GetCommonHeader(kt, bkUserCli, cfg)
	resp := new(ApiGatewayResp[*OT])
	err := cli.Verb(method).
		SubResourcef(url, urlParams...).
		WithContext(kt.Ctx).
		WithHeaders(header).
		WithParams(params).
		Do().Into(resp)

	if err != nil {
		logs.Errorf("fail to call api gateway api, err: %v, url: %s, rid: %s", err, url, kt.Rid)
		return nil, err
	}

	// 兼容code为string类型的情况（itsm_v4）
	var codeInt int
	if resp.Code != nil {
		codeInt, err = util.GetIntByInterface(resp.Code)
		if err != nil {
			logs.Errorf("failed to convert code to int, code: %v, url: %s, rid: %s", resp.Code, url, kt.Rid)
			return nil, err
		}
	}

	if !resp.Result || codeInt != 0 {
		err := fmt.Errorf("failed to call api gateway, code: %d, msg: %s, bk_error_code: %d, bk_error_msg: %s",
			resp.Code, resp.Message, resp.BKErrorCode, resp.BKErrorMessage)
		logs.Errorf("api gateway returns error, url: %s, err: %v, rid: %s", url, err, kt.Rid)
		return nil, err
	}
	return resp.Data, nil
}

// ApiGatewayCallWithRichErrorWithoutReq 该apigw辅助调用函数的适用场景如下：
// 成功响应时只返回 data 结构体，失败响应且状态码≥500时只返回带有更完整错误信息的 error 结构体
func ApiGatewayCallWithRichErrorWithoutReq[OT any](cli rest.ClientInterface, bkUserCli bkuser.Client,
	cfg *cc.ApiGateway,
	method rest.VerbType, kt *kit.Kit, params map[string]string, url string, urlParams ...any) (*OT, error) {

	header := GetCommonHeader(kt, bkUserCli, cfg)
	resp := new(ApiGatewayResp[*OT])
	resp.Result = true

	// Into函数本身会将基本网络错误打印出日志
	r := cli.Verb(method).
		SubResourcef(url, urlParams...).
		WithContext(kt.Ctx).
		WithHeaders(header).
		WithParams(params).
		Do()

	err := r.Into(resp)
	if err != nil {
		logs.Errorf("fail to call api gateway api, err: %v, url: %s, rid: %s", err, url, kt.Rid)
		return nil, err
	}

	// 兼容code为string类型的情况（itsm_v4）
	var codeInt int
	if resp.Code != nil {
		codeInt, err = util.GetIntByInterface(resp.Code)
		if err != nil {
			logs.Errorf("failed to convert code to int, code: %v, url: %s, rid: %s", resp.Code, url, kt.Rid)
			return nil, err
		}
	}

	if !resp.Result || codeInt != 0 {
		err := fmt.Errorf("failed to call api gateway, code: %d, msg: %s, bk_error_code: %d, bk_error_msg: %s",
			resp.Code, resp.Message, resp.BKErrorCode, resp.BKErrorMessage)
		logs.Errorf("api gateway returns error, url: %s, err: %v, rid: %s", url, err, kt.Rid)
		return nil, err
	}

	if r.StatusCode >= 500 { // api执行错误
		errData, err := json.Marshal(resp.Error.ErrData)
		if err != nil {
			logs.Errorf("failed to marshal error data, err: %v, url: %s, errData: %s, rid: %s", err, url,
				resp.Error.ErrData, kt.Rid)
			return nil, fmt.Errorf("failed to marshal error data: %v", err)
		}

		errDetails, err := json.Marshal(resp.Error.Details)
		if err != nil {
			logs.Errorf("failed to marshal error details, err: %v, url: %s, errDetails: %s, rid: %s", err, url,
				resp.Error.Details, kt.Rid)
			return nil, fmt.Errorf("failed to marshal error details: %v", err)
		}

		err = fmt.Errorf("failed to call api, code: %d, msg: %s, data: %s, details: %s",
			resp.Error.Code, resp.Error.Message, errData, errDetails)
		logs.Errorf("api returns error, url: %s, err: %v, rid: %s", url, err, kt.Rid)
		return nil, err
	}

	return resp.Data, nil
}

// ApiGatewayCallOriginalWithoutReq general call helper function for api gateway
// 该方法不会处理接口返回的code，适用于code非标准的场景，在上层自行处理
func ApiGatewayCallOriginalWithoutReq[OT any](cli rest.ClientInterface, bkUserCli bkuser.Client, cfg *cc.ApiGateway,
	method rest.VerbType, kt *kit.Kit, params map[string]string, url string, urlParams ...any) (
	int, string, *OT, error) {

	header := GetCommonHeader(kt, bkUserCli, cfg)
	resp := new(ApiGatewayResp[*OT])
	err := cli.Verb(method).
		SubResourcef(url, urlParams...).
		WithContext(kt.Ctx).
		WithHeaders(header).
		WithParams(params).
		Do().Into(resp)

	if err != nil {
		logs.Errorf("fail to call api gateway api, err: %v, url: %s, rid: %s", err, url, kt.Rid)
		return 0, "", nil, err
	}

	// 兼容code为string类型的情况（itsm_v4）
	var codeInt int
	if resp.Code != nil {
		codeInt, err = util.GetIntByInterface(resp.Code)
		if err != nil {
			logs.Errorf("failed to convert code to int, code: %v, url: %s, rid: %s", resp.Code, url, kt.Rid)
			return 0, "", nil, err
		}
	}

	if !resp.Result {
		err := fmt.Errorf("failed to call api gateway, code: %d, msg: %s, bk_error_code: %d, bk_error_msg: %s",
			resp.Code, resp.Message, resp.BKErrorCode, resp.BKErrorMessage)
		logs.Errorf("api gateway returns error, url: %s, err: %v, rid: %s", url, err, kt.Rid)
		return 0, "", nil, err
	}
	return codeInt, resp.Message, resp.Data, nil
}

// GetCommonHeader get common header
func GetCommonHeader(kt *kit.Kit, bkUserCli bkuser.Client, cfg *cc.ApiGateway) http.Header {
	header := kt.Header()
	// 如果配置了指定用户，使用指定用户调用
	user := kt.User
	if len(cfg.User) > 0 {
		// 通过用户管理获取指定用户的bk_username
		username, err := getBkUsername(kt, bkUserCli, cfg.User)
		if err != nil {
			logs.Warnf("fail to get bk_username by user, err: %v, user: %s, rid: %s", err, cfg.User, kt.Rid)
			return header
		}
		user = username
	}
	// TODO: 目前调用方式和itsm 不同，后期改成统一的ApiGateWay 客户端
	bkAuth := fmt.Sprintf(`{"bk_app_code": "%s", "bk_app_secret": "%s","bk_username":"%s"}`,
		cfg.AppCode, cfg.AppSecret, user)
	header.Set(constant.BKGWAuthKey, bkAuth)
	header.Set(constant.RidKey, kt.Rid)
	return header
}

// getCommonHeaderWithoutUser get common header without bk_username
func getCommonHeaderWithoutUser(kt *kit.Kit, cfg *cc.ApiGateway) http.Header {
	return bkuser.GetCommonHeaderWithoutUser(kt, cfg)
}

// getBkUsername get bk_username by login_name
func getBkUsername(kt *kit.Kit, bkUserCli bkuser.Client, loginName string) (string, error) {
	resp, err := bkUserCli.BatchLookupVirtualUser(kt, []string{loginName})
	if err != nil {
		logs.Errorf("fail to get bk_username by login_name, err: %v, login_name: %s, rid: %s", err, loginName,
			kt.Rid)
		return "", err
	}

	var bkUsername string
	for _, item := range resp.Data {
		if item.LoginName == loginName {
			bkUsername = item.BkUsername
			break
		}
	}

	if bkUsername == "" {
		logs.Errorf("login_name not found bk_username, login_name: %s, rid: %s", loginName, kt.Rid)
		return "", fmt.Errorf("login_name not found bk_username, login_name: %s", loginName)
	}
	return bkUsername, nil
}
