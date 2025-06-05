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

package login

import (
	"hcm/pkg/cc"
	"hcm/pkg/kit"
	"hcm/pkg/rest"
	"hcm/pkg/rest/client"
	apigateway "hcm/pkg/thirdparty/api-gateway"
	"hcm/pkg/thirdparty/api-gateway/bkuser"
	"hcm/pkg/thirdparty/api-gateway/discovery"
	"hcm/pkg/tools/ssl"

	"github.com/prometheus/client_golang/prometheus"
)

// Client ...
type Client interface {
	VerifyToken(kt *kit.Kit, token string) (*VerifyTokenRes, error)
	GetUserByToken(kt *kit.Kit, token string) (*UserInfo, error)
}

type login struct {
	config    *cc.ApiGateway
	client    rest.ClientInterface
	bkUserCli bkuser.Client
}

// NewClient ...
func NewClient(cfg *cc.ApiGateway, bkUserCli bkuser.Client, reg prometheus.Registerer) (Client, error) {
	tls := &ssl.TLSConfig{
		InsecureSkipVerify: cfg.TLS.InsecureSkipVerify,
		CertFile:           cfg.TLS.CertFile,
		KeyFile:            cfg.TLS.KeyFile,
		CAFile:             cfg.TLS.CAFile,
		Password:           cfg.TLS.Password,
	}
	cli, err := client.NewClient(tls)
	if err != nil {
		return nil, err
	}

	c := &client.Capability{
		Client: cli,
		Discover: &discovery.Discovery{
			Name:    "login",
			Servers: cfg.Endpoints,
		},
		MetricOpts: client.MetricOption{Register: reg},
	}
	return &login{
		config:    cfg,
		client:    rest.NewClient(c, "/"),
		bkUserCli: bkUserCli,
	}, nil
}

// VerifyToken verify user token
func (l *login) VerifyToken(kt *kit.Kit, token string) (*VerifyTokenRes, error) {
	params := map[string]string{"bk_token": token}
	return apigateway.ApiGatewayCallWithRichErrorWithoutReq[VerifyTokenRes](
		l.client, l.bkUserCli, l.config, rest.GET, kt, params, "/login/api/v3/open/bk-tokens/verify/")
}

// GetUserByToken get user info by token
func (l *login) GetUserByToken(kt *kit.Kit, token string) (*UserInfo, error) {
	params := map[string]string{"bk_token": token}
	return apigateway.ApiGatewayCallWithRichErrorWithoutReq[UserInfo](
		l.client, l.bkUserCli, l.config, rest.GET, kt, params, "/login/api/v3/open/bk-tokens/userinfo/")
}
