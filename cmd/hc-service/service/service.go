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
	"fmt"
	"net"
	"net/http"
	"runtime"
	"strconv"
	"sync"
	"time"

	"hcm/cmd/hc-service/service/account"
	"hcm/cmd/hc-service/service/capability"
	"hcm/cmd/hc-service/service/cloud-adaptor"
	"hcm/cmd/hc-service/service/subnet"
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

	cliSet := client.NewHCServiceClientSet(cli, dis)

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
	}

	account.InitAccountService(c)
	vpc.InitVpcService(c)
	subnet.InitSubnetService(c)

	return restful.NewContainer().Add(c.WebService)
}

// Healthz check whether the service is healthy.
func (s *Service) Healthz(w http.ResponseWriter, req *http.Request) {
	rest.WriteResp(w, rest.NewBaseResp(errf.OK, "healthy"))
	return
}

// GoAndWait provides safe concurrent handling. Per input handler, it starts a goroutine.
// Then it waits until all handlers are done and will recover if any handler panics.
// The returned error is the first non-nil error returned by one of the handlers.
// It can be set that non-nil error will be returned if the "key" handler fails while other handlers always
// return nil error.
func GoAndWait(handlers ...func() error) error {
	var (
		wg   sync.WaitGroup
		once sync.Once
		err  error
	)
	for _, f := range handlers {
		wg.Add(1)
		go func(handler func() error) {
			defer func() {
				if e := recover(); e != nil {
					buf := make([]byte, 1024)
					buf = buf[:runtime.Stack(buf, false)]
					logs.Errorf("[PANIC]%v\n%s\n", e, buf)
					once.Do(func() {
						err = errf.New(errf.Aborted, "panic found in call handlers")
					})
				}
				wg.Done()
			}()
			if e := handler(); e != nil {
				once.Do(func() {
					err = e
				})
			}
		}(f)
	}
	wg.Wait()
	return err
}
