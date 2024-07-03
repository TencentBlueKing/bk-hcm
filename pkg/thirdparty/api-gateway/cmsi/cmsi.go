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

// Package cmsi ...
package cmsi

import (
	"hcm/pkg/cc"
	"hcm/pkg/criteria/constant"
	"hcm/pkg/kit"
	"hcm/pkg/rest"
	"hcm/pkg/rest/client"
	apigateway "hcm/pkg/thirdparty/api-gateway"
	"hcm/pkg/tools/ssl"
	"net/http"

	"github.com/prometheus/client_golang/prometheus"
)

// Client cmsi client
type Client interface {
	SendMail(kt *kit.Kit, m *CmsiMail) (err error)
}

// NewClient return a new cmsi client
func NewClient(cfg *cc.CMSI, reg prometheus.Registerer) (Client, error) {
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
		Discover: &apigateway.Discovery{
			Name:    "cmsi",
			Servers: cfg.Endpoints,
		},
		MetricOpts: client.MetricOption{Register: reg},
	}
	restCli := rest.NewClient(c, "/v2/cmsi")
	return &cmsi{
		client: restCli,
		config: &cfg.ApiGateway,
		sender: cfg.Sender,
		cc:     cfg.CC,
	}, nil
}

type cmsi struct {
	config *cc.ApiGateway
	// http client instance
	client rest.ClientInterface
	// email sender 需要加入白名单
	sender string
	// cc 抄送人
	cc []string
}

func (i *cmsi) header(kt *kit.Kit) http.Header {
	header := http.Header{}
	header.Set(constant.RidKey, kt.Rid)
	header.Set(constant.BKGWAuthKey, i.config.GetAuthValue())
	return header
}
