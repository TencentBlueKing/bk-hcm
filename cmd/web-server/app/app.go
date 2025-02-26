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

// Package app ...
package app

import (
	"fmt"
	"net"
	"strconv"

	"hcm/cmd/web-server/options"
	"hcm/cmd/web-server/service"
	"hcm/pkg/cc"
	"hcm/pkg/logs"
	"hcm/pkg/metrics"
	"hcm/pkg/runtime/ctl"
	"hcm/pkg/runtime/ctl/cmd"
	"hcm/pkg/runtime/shutdown"
	"hcm/pkg/serviced"
)

// Run start the web server
func Run(opt *options.Option) error {
	s := new(webService)
	if err := s.prepare(opt); err != nil {
		return err
	}

	if err := s.svc.ListenAndServeRest(); err != nil {
		return err
	}

	shutdown.RegisterFirstShutdown(s.finalizer)
	shutdown.WaitShutdown(20)
	return nil
}

type webService struct {
	svc *service.Service
	dis serviced.Discover
}

// prepare do prepare jobs before run api server.
func (s *webService) prepare(opt *options.Option) error {
	// load settings from config file.
	if err := cc.LoadSettings(opt.Sys); err != nil {
		return fmt.Errorf("load settings from config files failed, err: %v", err)
	}

	logs.InitLogger(cc.WebServer().Log.Logs())

	logs.Infof("load settings from config file success.")

	logs.Infof("start service %s with option: env: %s, labels: %v, disable election: %v \n",
		cc.WebServerName, opt.Sys.Environment, opt.Sys.Labels, opt.Sys.DisableElection)

	// init metrics
	network := cc.WebServer().Network
	metrics.InitMetrics(net.JoinHostPort(network.BindIP, strconv.Itoa(int(network.Port))))

	// new api server discovery client.
	discOpt := serviced.DiscoveryOption{Services: []cc.Name{cc.CloudServerName, cc.AuthServerName, cc.AccountServerName}}
	dis, err := serviced.NewDiscovery(cc.WebServer().Service, discOpt)
	if err != nil {
		return fmt.Errorf("new service discovery faield, err: %v", err)
	}

	s.dis = dis
	logs.Infof("create discovery success.")

	svc, err := service.NewService(s.dis)
	if err != nil {
		return fmt.Errorf("initialize service failed, err: %v", err)
	}
	s.svc = svc

	// init hcm control tool
	if err := ctl.LoadCtl(cmd.WithLog()); err != nil {
		return fmt.Errorf("load control tool failed, err: %v", err)
	}

	return nil
}

func (s *webService) finalizer() {
	return
}
