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

package esb

import (
	"hcm/pkg/cc"
	"hcm/pkg/rest"
	"hcm/pkg/rest/client"
	"hcm/pkg/thirdparty/esb/cmdb"
	"hcm/pkg/thirdparty/esb/iam"
	"hcm/pkg/thirdparty/esb/itsm"
	"hcm/pkg/thirdparty/esb/login"
	"hcm/pkg/thirdparty/esb/usermgr"
	"hcm/pkg/tools/ssl"

	"github.com/prometheus/client_golang/prometheus"
)

// Client esb client
type Client interface {
	Cmdb() cmdb.Client
	Login() login.Client
	Usermgr() usermgr.Client
	Iam() iam.Client
	Itsm() itsm.Client
}

// NewClient new esb client.
func NewClient(cfg *cc.Esb, reg prometheus.Registerer) (Client, error) {
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
		Discover: &esbDiscovery{
			servers: cfg.Endpoints,
		},
		MetricOpts: client.MetricOption{Register: reg},
	}
	restCli := rest.NewClient(c, "/api/c/compapi/v2")
	return &esbCli{
		cc:      cmdb.NewClient(restCli, cfg),
		login:   login.NewClient(restCli, cfg),
		usermgr: usermgr.NewClient(restCli, cfg),
		iam:     iam.NewClient(restCli, cfg),
		itsm:    itsm.NewClient(restCli, cfg),
	}, nil
}

type esbCli struct {
	cc      cmdb.Client
	login   login.Client
	usermgr usermgr.Client
	iam     iam.Client
	itsm    itsm.Client
}

func (e *esbCli) Cmdb() cmdb.Client {
	return e.cc
}

func (e *esbCli) Login() login.Client {
	return e.login
}

func (e *esbCli) Usermgr() usermgr.Client {
	return e.usermgr
}

func (e *esbCli) Iam() iam.Client {
	return e.iam
}

func (e *esbCli) Itsm() itsm.Client {
	return e.itsm
}
