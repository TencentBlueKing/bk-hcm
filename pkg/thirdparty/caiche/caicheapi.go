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

// Package caiche is a client for Third Party API
package caiche

import (
	"errors"
	"fmt"

	"hcm/pkg/cc"
	"hcm/pkg/kit"
	"hcm/pkg/logs"
	"hcm/pkg/rest"
	"hcm/pkg/rest/client"
	cvt "hcm/pkg/tools/converter"

	"github.com/prometheus/client_golang/prometheus"
)

// CaiCheClientInterface caiche api interface
type CaiCheClientInterface interface {
	ListDevice(kt *kit.Kit, req *ListDeviceReq) (*DeviceListData, error)
	ListDeviceV2(kt *kit.Kit, req *ListDeviceV2Req) (*DeviceListV2Result, error)
}

// NewCaiCheClientInterface creates a caiche api instance
func NewCaiCheClientInterface(opts cc.CaiCheCli, reg prometheus.Registerer) (CaiCheClientInterface, error) {
	cli, err := client.NewClient(nil)
	if err != nil {
		return nil, err
	}

	c := &client.Capability{
		Client: cli,
		Discover: &ServerDiscovery{
			name:    "caiche api",
			servers: []string{opts.Host},
		},
		MetricOpts: client.MetricOption{Register: reg},
	}

	return &caicheApi{
		client: rest.NewClient(c, "/"),
		opts:   opts,
	}, nil
}

// caicheApi Alarm api interface implementation
type caicheApi struct {
	client rest.ClientInterface
	opts   cc.CaiCheCli
}

// getToken https://iwiki.woa.com/p/4008182364
func (c *caicheApi) getToken(kt *kit.Kit) (string, error) {
	req := &GetTokenReq{
		ID:      kt.Rid,
		JsonRPC: "2.0",
		Params: GrantParams{
			AppKey:    c.opts.AppKey,
			AppSecret: c.opts.AppSecret,
			GrantType: ClientCredentials,
		},
		Reason:   "hcm",
		XTraceID: kt.Rid,
	}
	subPath := "/trpc.teg_devops.open_api.OpenApiService/GetToken"

	resp := new(GetTokenResp)
	err := c.client.Post().
		WithContext(kt.Ctx).
		Body(req).
		SubResourcef(subPath).
		WithHeaders(kt.Header()).
		Do().
		Into(resp)

	if err != nil {
		logs.Errorf("failed to get token, err: %v, req: %+v, rid: %s", err, cvt.PtrToVal(req), kt.Rid)
		return "", err
	}

	if resp.Code != 0 {
		logs.Errorf("get token is invalid, code: %d, msg: %s, req: %+v, rid: %s", resp.Code, resp.Msg,
			cvt.PtrToVal(req), kt.Rid)
		return "", fmt.Errorf("get token is invalid, code: %d, msg: %s", resp.Code, resp.Msg)
	}

	if resp.Result == nil {
		logs.Errorf("resp result is nil, req: %+v, rid: %s", cvt.PtrToVal(req), kt.Rid)
		return "", errors.New("get token failed, result is nil")
	}

	return resp.Result.AccessToken, nil
}

// ListDevice list device https://iwiki.woa.com/p/4008334375
func (c *caicheApi) ListDevice(kt *kit.Kit, req *ListDeviceReq) (*DeviceListData, error) {
	token, err := c.getToken(kt)
	if err != nil {
		logs.Errorf("get token failed, err: %v, rid: %s", err, kt.Rid)
		return nil, err
	}

	subPath := "/openapi_gateway/trpc.teg_devops.web_api.WebApiService/DeviceList"
	header := kt.Header()
	header.Set(authorizationHeader, token)

	resp := new(ListDeviceResp)
	err = c.client.Post().
		WithContext(kt.Ctx).
		Body(req).
		SubResourcef(subPath).
		WithHeaders(header).
		Do().
		Into(resp)
	if err != nil {
		logs.Errorf("send request failed, err: %v, req: %+v, rid: %s", err, cvt.PtrToVal(req), kt.Rid)
		return nil, err
	}

	if resp.Code != 0 {
		logs.Errorf("list device resp is invalid, code: %d, msg: %s, req: %+v, rid: %s", resp.Code, resp.Msg,
			cvt.PtrToVal(req), kt.Rid)
		return nil, fmt.Errorf("list device resp is invalid, code: %d, msg: %s", resp.Code, resp.Msg)
	}

	if resp.Data == nil {
		logs.Errorf("device data is nil, req: %+v, rid: %s", cvt.PtrToVal(req), kt.Rid)
		return nil, errors.New("device data is nil")
	}

	return resp.Data, nil
}

// ListDeviceV2 list device v2 https://iwiki.woa.com/p/4015994172
func (c *caicheApi) ListDeviceV2(kt *kit.Kit, req *ListDeviceV2Req) (*DeviceListV2Result, error) {
	token, err := c.getToken(kt)
	if err != nil {
		logs.Errorf("get token failed, err: %v, rid: %s", err, kt.Rid)
		return nil, err
	}

	subPath := "/openapi_gateway/abolish-backend/device/listDeviceOpenapi"
	header := kt.Header()
	header.Set(authorizationHeader, token)

	resp := new(ListDeviceV2Resp)
	err = c.client.Post().
		WithContext(kt.Ctx).
		Body(req).
		SubResourcef(subPath).
		WithHeaders(header).
		Do().
		Into(resp)
	if err != nil {
		logs.Errorf("send request failed, err: %v, req: %+v, rid: %s", err, cvt.PtrToVal(req), kt.Rid)
		return nil, err
	}

	if resp.Result == nil {
		logs.Errorf("device v2 data is nil, req: %+v, XTraceID: %s, rid: %s", cvt.PtrToVal(req), resp.XTraceID, kt.Rid)
		return nil, errors.New("device v2 data is nil")
	}

	return resp.Result, nil
}
