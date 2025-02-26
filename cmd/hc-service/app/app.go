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

	"hcm/cmd/hc-service/options"
	"hcm/cmd/hc-service/service"
	adptmetric "hcm/pkg/adaptor/metric"
	mocktcloud "hcm/pkg/adaptor/mock/tcloud"
	"hcm/pkg/cc"
	"hcm/pkg/logs"
	"hcm/pkg/metrics"
	"hcm/pkg/runtime/ctl"
	"hcm/pkg/runtime/shutdown"
	"hcm/pkg/serviced"
)

// Run start the hc service.
func Run(opt *options.Option) error {
	hs := new(hcService)
	if err := hs.prepare(opt); err != nil {
		return err
	}

	if err := hs.svc.ListenAndServeRest(); err != nil {
		return err
	}

	if err := hs.register(); err != nil {
		return err
	}

	shutdown.RegisterFirstShutdown(hs.finalizer)
	shutdown.WaitShutdown(20)
	return nil
}

type hcService struct {
	svc *service.Service
	sd  serviced.Service
}

// prepare do prepare jobs before run api discover.
func (ds *hcService) prepare(opt *options.Option) error {
	// load settings from config file.
	if err := cc.LoadSettings(opt.Sys); err != nil {
		return fmt.Errorf("load settings from config files failed, err: %v", err)
	}

	logs.InitLogger(cc.HCService().Log.Logs())

	logs.Infof("load settings from config file success.")

	// init metrics
	network := cc.HCService().Network
	metrics.InitMetrics(net.JoinHostPort(network.BindIP, strconv.Itoa(int(network.Port))))
	adptmetric.InitCloudApiMetrics(metrics.Register())

	// register hc service.
	svcOpt := serviced.NewServiceOption(cc.HCServiceName, cc.HCService().Network, opt.Sys)
	disOpt := serviced.DiscoveryOption{
		Services: []cc.Name{cc.DataServiceName},
	}

	sd, err := serviced.NewServiceD(cc.HCService().Service, svcOpt, disOpt)
	if err != nil {
		return fmt.Errorf("new service discovery failed, err: %v", err)
	}

	ds.sd = sd

	svc, err := service.NewService(sd)
	if err != nil {
		return fmt.Errorf("initialize service failed, err: %v", err)
	}
	ds.svc = svc

	// init hcm control tool
	if err := ctl.LoadCtl(ctl.WithBasics(sd)...); err != nil {
		return fmt.Errorf("load control tool failed, err: %v", err)
	}

	return nil
}

// register hc-service to etcd.
func (ds *hcService) register() error {
	if err := ds.sd.Register(); err != nil {
		return fmt.Errorf("register hc service failed, err: %v", err)
	}

	logs.Infof("register hc service to etcd success.")
	return nil
}

func (ds *hcService) finalizer() {
	if err := ds.sd.Deregister(); err != nil {
		logs.Errorf("process service shutdown, but deregister failed, err: %v", err)
		return
	}

	mocktcloud.Finish()
	logs.Infof("shutting down service, deregister service success.")
	return
}
