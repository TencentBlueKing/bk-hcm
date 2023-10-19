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

package itsm

import (
	"net/http"

	"hcm/pkg/cc"
	"hcm/pkg/criteria/constant"
	"hcm/pkg/kit"
	"hcm/pkg/rest"
	"hcm/pkg/rest/client"
	"hcm/pkg/thirdparty"
	"hcm/pkg/tools/ssl"

	"github.com/prometheus/client_golang/prometheus"
)

// Client Itsm api.
type Client interface {
	// CreateTicket 创建单据。
	CreateTicket(kt *kit.Kit, params *CreateTicketParams) (string, error)
	// GetTicketResult 获取单据结果。
	GetTicketResult(kt *kit.Kit, sn string) (TicketResult, error)
	// WithdrawTicket 撤销单据。
	WithdrawTicket(kt *kit.Kit, sn string, operator string) error
	// VerifyToken 校验Token。
	VerifyToken(kt *kit.Kit, token string) (bool, error)
	// GetTicketsByUser 获取用户的单据。
	GetTicketsByUser(kt *kit.Kit, req *GetTicketsByUserReq) (*GetTicketsByUserRespData, error)
	// Approve 审批单据。
	Approve(kt *kit.Kit, req *ApproveReq) error
}

// NewClient initialize a new itsm client
func NewClient(cfg *cc.ApiGateway, reg prometheus.Registerer) (Client, error) {
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
		Discover: &thirdparty.Discovery{
			Name:    "itsm",
			Servers: cfg.Endpoints,
		},
		MetricOpts: client.MetricOption{Register: reg},
	}
	restCli := rest.NewClient(c, "/v2/itsm")
	return &itsm{
		client: restCli,
		config: cfg,
	}, nil
}

// itsm is an esb client to request itsm.
type itsm struct {
	config *cc.ApiGateway
	// http client instance
	client rest.ClientInterface
}

func (i *itsm) header(kt *kit.Kit) http.Header {
	header := http.Header{}
	header.Set(constant.RidKey, kt.Rid)
	header.Set(constant.BKGWAuthKey, i.config.GetAuthValue())
	return header
}
