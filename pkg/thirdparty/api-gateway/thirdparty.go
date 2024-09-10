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
	"fmt"
	"net/http"
	"sync"

	"hcm/pkg/cc"
	"hcm/pkg/criteria/constant"
	"hcm/pkg/kit"
	"hcm/pkg/logs"
	"hcm/pkg/rest"
)

// Discovery used to third-party service discovery.
type Discovery struct {
	Name    string
	Servers []string
	index   int
	sync.Mutex
}

// BaseResponse is esb http base response.
type BaseResponse struct {
	Result  bool   `json:"result"`
	Code    int    `json:"code"`
	Message string `json:"message"`
}

// GetServers get third-party service server host.
func (d *Discovery) GetServers() ([]string, error) {
	d.Lock()
	defer d.Unlock()
	num := len(d.Servers)
	if num == 0 {
		return []string{}, fmt.Errorf("there is no %s server can be used", d.Name)
	}
	if d.index < num-1 {
		d.index = d.index + 1
		return append(d.Servers[d.index-1:], d.Servers[:d.index-1]...), nil
	}
	d.index = 0
	return append(d.Servers[num-1:], d.Servers[:num-1]...), nil
}

// ApiGatewayResp ...
type ApiGatewayResp[T any] struct {
	Result         bool   `json:"result"`
	Code           int    `json:"code"`
	BKErrorCode    int    `json:"bk_error_code"`
	Message        string `json:"message"`
	BKErrorMessage string `json:"bk_error_msg"`
	Data           T      `json:"data"`
}

// ApiGatewayCall general call helper function for api gateway
func ApiGatewayCall[IT any, OT any](cli rest.ClientInterface, cfg *cc.ApiGateway,
	method rest.VerbType, kt *kit.Kit, req *IT, url string, urlParams ...any) (*OT, error) {

	header := getCommonHeader(kt, cfg)
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

	if !resp.Result || resp.Code != 0 {
		err := fmt.Errorf("failed to call api gateway, code: %d, msg: %s, bk_error_code: %d, bk_error_msg: %s",
			resp.Code, resp.Message, resp.BKErrorCode, resp.BKErrorMessage)
		logs.Errorf("api gateway returns error, url: %s, err: %v, rid: %s", url, err, kt.Rid)
		return nil, err
	}
	return resp.Data, nil
}

// ApiGatewayCallWithoutReq general call helper function for api gateway
func ApiGatewayCallWithoutReq[OT any](cli rest.ClientInterface, cfg *cc.ApiGateway,
	method rest.VerbType, kt *kit.Kit, params map[string]string, url string, urlParams ...any) (*OT, error) {

	header := getCommonHeader(kt, cfg)
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

	if !resp.Result || resp.Code != 0 {
		err := fmt.Errorf("failed to call api gateway, code: %d, msg: %s, bk_error_code: %d, bk_error_msg: %s",
			resp.Code, resp.Message, resp.BKErrorCode, resp.BKErrorMessage)
		logs.Errorf("api gateway returns error, url: %s, err: %v, rid: %s", url, err, kt.Rid)
		return nil, err
	}
	return resp.Data, nil
}

func getCommonHeader(kt *kit.Kit, cfg *cc.ApiGateway) http.Header {
	header := kt.Header()
	// 如果配置了指定用户，使用指定用户调用
	user := kt.User
	if len(cfg.User) > 0 {
		user = cfg.User
	}
	// TODO: 目前调用方式和itsm 不同，后期改成统一的ApiGateWay 客户端
	bkAuth := fmt.Sprintf(`{"bk_app_code": "%s", "bk_app_secret": "%s","bk_username":"%s"}`,
		cfg.AppCode, cfg.AppSecret, user)
	header.Set(constant.BKGWAuthKey, bkAuth)
	header.Set(constant.RidKey, kt.Rid)
	return header
}
