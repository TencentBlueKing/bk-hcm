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
	"context"
	"errors"
	"fmt"
	"net/http"
	"net/url"

	"hcm/pkg/criteria/constant"
	"hcm/pkg/metrics"
	"hcm/pkg/rest"
	"hcm/pkg/rest/client"
	"hcm/pkg/tools/uuid"
)

type oaLoginDiscovery struct {
	server string
}

// GetServers ...
func (s *oaLoginDiscovery) GetServers() ([]string, error) {
	if len(s.server) == 0 {
		return []string{}, errors.New("there is no oa login server can be used")
	}
	return []string{s.server}, nil
}

type oaLoginClient struct {
	// http client instance
	client rest.ClientInterface
}

func newOALoginClient(bkLoginUrl string) (*oaLoginClient, error) {
	// 解析Login URL
	u, err := url.Parse(bkLoginUrl)
	if err != nil {
		return nil, err
	}

	// 生成Client
	cli, err := client.NewClient(nil)
	if err != nil {
		return nil, err
	}
	c := &client.Capability{
		Client: cli,
		Discover: &oaLoginDiscovery{
			server: fmt.Sprintf("%s://%s/", u.Scheme, u.Host),
		},
		MetricOpts: client.MetricOption{Register: metrics.Register()},
	}
	restCli := rest.NewClient(c, "")
	return &oaLoginClient{client: restCli}, nil
}

// Verify ...
func (s *oaLoginClient) Verify(ctx context.Context, bkTicket string) (*rest.Response, error) {
	resp := &struct {
		Code    int    `json:"ret"`
		Message string `json:"msg"`
		Data    struct {
			Username string `json:"username"`
		} `json:"data"`
	}{}

	h := http.Header{}
	requestID := uuid.UUID()
	h.Set(constant.RidKey, requestID)

	err := s.client.Get().
		SubResourcef("/user/get_info").
		WithContext(ctx).
		WithHeaders(h).WithParam("bk_ticket", bkTicket).
		Do().Into(resp)

	if err != nil {
		return nil, err
	}

	ret := &rest.Response{
		Code:    int32(resp.Code),
		Message: resp.Message,
		Data: loginVerifyRespData{
			UserName: resp.Data.Username,
		},
	}

	if resp.Code != 0 {
		return nil, fmt.Errorf("%+v", ret)
	}

	return ret, nil
}
