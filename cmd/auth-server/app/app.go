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
	"sync"

	"hcm/cmd/auth-server/options"
	"hcm/cmd/auth-server/service"
	"hcm/pkg/cc"
	"hcm/pkg/logs"
	"hcm/pkg/metrics"
	"hcm/pkg/runtime/ctl"
	"hcm/pkg/runtime/ctl/cmd"
	"hcm/pkg/runtime/shutdown"
	"hcm/pkg/serviced"
)

// Run start the auth server
func Run(opt *options.Option) error {
	as := new(authService)
	if err := as.prepare(opt); err != nil {
		return err
	}

	svc, err := service.NewService(as.sd, cc.AuthServer().IAM, cc.AuthServer().Esb, as.disableAuth, as.disableWriteOpt)
	if err != nil {
		return fmt.Errorf("initialize service failed, err: %v", err)
	}
	as.svc = svc
	if err := as.svc.ListenAndServeRest(); err != nil {
		return err
	}

	if err := as.register(); err != nil {
		return err
	}

	shutdown.RegisterFirstShutdown(as.finalizer)
	shutdown.WaitShutdown(20)
	return nil
}

type authService struct {
	svc *service.Service
	sd  serviced.ServiceDiscover

	// disableAuth defines whether iam authorization is disabled
	disableAuth bool
	// disableWriteOpt defines which biz's write operation needs to be disabled
	disableWriteOpt *options.DisableWriteOption
}

// prepare do prepare jobs before run auth server.
func (as *authService) prepare(opt *options.Option) error {
	// load settings from config file.
	if err := cc.LoadSettings(opt.Sys); err != nil {
		return fmt.Errorf("load settings from config files failed, err: %v", err)
	}

	logs.InitLogger(cc.AuthServer().Log.Logs())

	logs.Infof("load settings from config file success.")

	// init metrics
	network := cc.AuthServer().Network
	metrics.InitMetrics(net.JoinHostPort(network.BindIP, strconv.Itoa(int(network.Port))))

	// init auth server's service discovery.
	svcOpt := serviced.NewServiceOption(cc.AuthServerName, cc.AuthServer().Network, opt.Sys)
	discOpt := serviced.DiscoveryOption{Services: []cc.Name{cc.DataServiceName}}
	sd, err := serviced.NewServiceD(cc.AuthServer().Service, svcOpt, discOpt)
	if err != nil {
		return fmt.Errorf("new service discovery faield, err: %v", err)
	}

	as.sd = sd
	logs.Infof("create service discovery success.")

	as.disableWriteOpt = &options.DisableWriteOption{
		IsDisabled: false,
		IsAll:      false,
		BizIDMap:   sync.Map{},
	}

	as.disableAuth = opt.DisableAuth
	if opt.DisableAuth {
		logs.Infof("authorize function is disabled.")
	}

	// init hcm control tool
	if err := ctl.LoadCtl(append(ctl.WithBasics(sd), cmd.WithAuth(as.disableWriteOpt)...)...); err != nil {
		return fmt.Errorf("load control tool failed, err: %v", err)
	}

	return nil
}

// register auth-server to etcd.
func (as *authService) register() error {
	if err := as.sd.Register(); err != nil {
		return fmt.Errorf("register auth server failed, err: %v", err)
	}

	logs.Infof("register auth server to etcd success.")
	return nil
}

func (as *authService) finalizer() {
	if err := as.sd.Deregister(); err != nil {
		logs.Errorf("process service shutdown, but deregister failed, err: %v", err)
		return
	}

	logs.Infof("shutting down service, deregister service success.")
}
