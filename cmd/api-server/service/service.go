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

	"hcm/pkg/cc"
	"hcm/pkg/criteria/errf"
	"hcm/pkg/handler"
	"hcm/pkg/logs"
	"hcm/pkg/rest"
	"hcm/pkg/rest/client"
	"hcm/pkg/runtime/shutdown"
	"hcm/pkg/serviced"
	"hcm/pkg/tools/ssl"
)

// Service do all the api server's work
type Service struct {
	proxy *proxy
}

// NewService create a service instance.
func NewService(dis serviced.Discover) (*Service, error) {
	network := cc.ApiServer().Network

	var tlsConfig *ssl.TLSConfig
	if network.TLS.Enable() {
		tlsConfig = &ssl.TLSConfig{
			InsecureSkipVerify: network.TLS.InsecureSkipVerify,
			CertFile:           network.TLS.CertFile,
			KeyFile:            network.TLS.KeyFile,
			CAFile:             network.TLS.CAFile,
			Password:           network.TLS.Password,
		}
	}

	cli, err := client.NewClient(tlsConfig)
	if err != nil {
		return nil, err
	}

	p, err := newProxy(dis, cli)
	if err != nil {
		return nil, err
	}

	return &Service{
		proxy: p,
	}, nil
}

// ListenAndServeRest listen and serve the restful server
func (s *Service) ListenAndServeRest() error {

	root := http.NewServeMux()
	root.HandleFunc("/", s.proxy.apiSet().ServeHTTP)
	root.HandleFunc("/healthz", s.Healthz)
	handler.SetCommonHandler(root)

	network := cc.ApiServer().Network
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

	return nil
}

// Healthz service health check.
func (s *Service) Healthz(w http.ResponseWriter, r *http.Request) {
	if shutdown.IsShuttingDown() {
		logs.Errorf("service healthz check failed, current service is shutting down")
		w.WriteHeader(http.StatusServiceUnavailable)
		rest.WriteResp(w, rest.NewBaseResp(errf.UnHealthy, "current service is shutting down"))
		return
	}

	if err := serviced.Healthz(r.Context(), cc.ApiServer().Service); err != nil {
		logs.Errorf("serviced healthz check failed, err: %v", err)
		rest.WriteResp(w, rest.NewBaseResp(errf.UnHealthy, "serviced healthz error, "+err.Error()))
		return
	}

	rest.WriteResp(w, rest.NewBaseResp(errf.OK, "healthy"))
	return
}
