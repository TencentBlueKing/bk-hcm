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

// Package service ...
package service

import (
	"context"
	"fmt"
	"net"
	"net/http"
	"strconv"
	"time"

	"hcm/cmd/hc-service/logics/cloud-adaptor"
	ressync "hcm/cmd/hc-service/logics/res-sync"
	"hcm/cmd/hc-service/service/account"
	argstpl "hcm/cmd/hc-service/service/argument-template"
	"hcm/cmd/hc-service/service/bill"
	"hcm/cmd/hc-service/service/capability"
	"hcm/cmd/hc-service/service/cvm"
	"hcm/cmd/hc-service/service/disk"
	"hcm/cmd/hc-service/service/eip"
	"hcm/cmd/hc-service/service/firewall"
	instancetype "hcm/cmd/hc-service/service/instance-type"
	routetable "hcm/cmd/hc-service/service/route-table"
	securitygroup "hcm/cmd/hc-service/service/security-group"
	"hcm/cmd/hc-service/service/subnet"
	"hcm/cmd/hc-service/service/sync"
	"hcm/cmd/hc-service/service/vpc"
	"hcm/pkg/cc"
	"hcm/pkg/client"
	"hcm/pkg/criteria/errf"
	"hcm/pkg/handler"
	"hcm/pkg/logs"
	"hcm/pkg/rest"
	restcli "hcm/pkg/rest/client"
	"hcm/pkg/runtime/shutdown"
	"hcm/pkg/serviced"
	"hcm/pkg/tools/ssl"

	"github.com/emicklei/go-restful/v3"
)

// Service do all the hc service's work
type Service struct {
	serve        *http.Server
	clientSet    *client.ClientSet
	cloudAdaptor *cloudadaptor.CloudAdaptorClient
}

// NewService create a service instance.
func NewService(dis serviced.Discover) (*Service, error) {
	cli, err := restcli.NewClient(nil)
	if err != nil {
		return nil, err
	}

	cliSet := client.NewClientSet(cli, dis)

	cloudAdaptor := cloudadaptor.NewCloudAdaptorClient(cliSet.DataService())

	svr := &Service{
		clientSet:    cliSet,
		cloudAdaptor: cloudAdaptor,
	}

	return svr, nil
}

// ListenAndServeRest listen and serve the restful server
func (s *Service) ListenAndServeRest() error {
	root := http.NewServeMux()
	root.HandleFunc("/", s.apiSet().ServeHTTP)
	root.HandleFunc("/healthz", s.Healthz)
	handler.SetCommonHandler(root)

	network := cc.HCService().Network
	server := &http.Server{
		Addr:    net.JoinHostPort(network.BindIP, strconv.FormatUint(uint64(network.Port), 10)),
		Handler: root,
	}

	if network.TLS.Enable() {
		tls := network.TLS
		tlsC, err := ssl.ClientTLSConfVerify(tls.InsecureSkipVerify, tls.CAFile, tls.CertFile, tls.KeyFile,
			tls.Password)
		if err != nil {
			return fmt.Errorf("init restful tls config failed, err: %v", err)
		}

		server.TLSConfig = tlsC
	}

	logs.Infof("listen restful server on %s with secure(%v) now.", server.Addr, network.TLS.Enable())

	go func() {
		notifier := shutdown.AddNotifier()
		select {
		case <-notifier.Signal:
			defer notifier.Done()

			logs.Infof("start shutdown restful server gracefully...")

			ctx, cancel := context.WithTimeout(context.TODO(), 20*time.Second)
			defer cancel()
			if err := server.Shutdown(ctx); err != nil {
				logs.Errorf("shutdown restful server failed, err: %v", err)
				return
			}

			logs.Infof("shutdown restful server success...")
		}
	}()

	go func() {
		if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			logs.Errorf("serve restful server failed, err: %v", err)
			shutdown.SignalShutdownGracefully()
		}
	}()

	s.serve = server

	return nil
}

func (s *Service) apiSet() *restful.Container {
	ws := new(restful.WebService)
	ws.Path("/api/v1/hc")
	ws.Produces(restful.MIME_JSON)

	c := &capability.Capability{
		WebService:   ws,
		ClientSet:    s.clientSet,
		CloudAdaptor: s.cloudAdaptor,
		ResSyncCli:   ressync.NewClient(s.cloudAdaptor, s.clientSet.DataService()),
	}

	account.InitAccountService(c)
	securitygroup.InitSecurityGroupService(c)
	firewall.InitFirewallService(c)
	vpc.InitVpcService(c)
	subnet.InitSubnetService(c)
	disk.InitDiskService(c)
	cvm.InitCvmService(c)
	routetable.InitRouteTableService(c)
	eip.InitEipService(c)
	instancetype.InitInstanceTypeService(c)
	sync.InitService(c)
	bill.InitBillService(c)
	argstpl.InitArgsTplService(c)

	return restful.NewContainer().Add(c.WebService)
}

// Healthz check whether the service is healthy.
func (s *Service) Healthz(w http.ResponseWriter, _ *http.Request) {

	if err := serviced.Healthz(cc.HCService().Service); err != nil {
		logs.Errorf("etcd healthz check failed, err: %v", err)
		rest.WriteResp(w, rest.NewBaseResp(errf.UnHealthy, "etcd healthz error, "+err.Error()))
		return
	}

	rest.WriteResp(w, rest.NewBaseResp(errf.OK, "healthy"))
	return
}
