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

	"hcm/cmd/task-server/options"
	"hcm/cmd/task-server/service"
	"hcm/pkg/cc"
	"hcm/pkg/logs"
	"hcm/pkg/metrics"
	"hcm/pkg/runtime/ctl"
	"hcm/pkg/runtime/shutdown"
	"hcm/pkg/serviced"
)

const shutdownWaitTimeSec = 60

// Run start the task server.
func Run(opt *options.Option) error {
	ds := new(taskServer)
	if err := ds.prepare(opt); err != nil {
		return err
	}

	if err := ds.svc.ListenAndServeRest(); err != nil {
		return err
	}

	if err := ds.register(); err != nil {
		return err
	}

	shutdown.RegisterFirstShutdown(ds.finalizer)
	shutdown.WaitShutdown(shutdownWaitTimeSec)
	return nil
}

type taskServer struct {
	svc *service.Service
	sd  serviced.Service
}

// prepare do prepare jobs before run api discover.
func (ds *taskServer) prepare(opt *options.Option) error {
	// load settings from config file.
	if err := cc.LoadSettings(opt.Sys); err != nil {
		return fmt.Errorf("load settings from config files failed, err: %v", err)
	}

	logs.InitLogger(cc.TaskServer().Log.Logs())

	logs.Infof("load settings from config file success.")
	logs.Infof("use label: %+v", cc.TaskServer().UseLabel)

	// init metrics
	network := cc.TaskServer().Network
	metrics.InitMetrics(net.JoinHostPort(network.BindIP, strconv.Itoa(int(network.Port))))

	// init service discovery.
	svcOpt := serviced.NewServiceOption(cc.TaskServerName, cc.TaskServer().Network, opt.Sys)
	discOpt := serviced.DiscoveryOption{
		Services: []cc.Name{cc.DataServiceName, cc.HCServiceName},
	}
	sd, err := serviced.NewServiceD(cc.TaskServer().Service, svcOpt, discOpt)
	if err != nil {
		return fmt.Errorf("new service discovery failed, err: %v", err)
	}
	ds.sd = sd

	// init service.
	svc, err := service.NewService(sd, shutdownWaitTimeSec)
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

// register task-server to etcd.
func (ds *taskServer) register() error {
	if err := ds.sd.Register(); err != nil {
		return fmt.Errorf("register task server failed, err: %v", err)
	}

	logs.Infof("register task server to etcd success.")
	return nil
}

// finalizer ...
func (ds *taskServer) finalizer() {
	if err := ds.sd.Deregister(); err != nil {
		logs.Errorf("process service shutdown, but deregister failed, err: %v", err)
		return
	}

	logs.Infof("shutting down service, deregister service success.")
	return
}
