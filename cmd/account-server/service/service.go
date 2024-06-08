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

	logicaudit "hcm/cmd/account-server/logics/audit"
	mainaccount "hcm/cmd/account-server/service/account-set/main-account"
	rootaccount "hcm/cmd/account-server/service/account-set/root-account"
	"hcm/cmd/account-server/service/capability"
	"hcm/pkg/cc"
	"hcm/pkg/client"
	"hcm/pkg/criteria/errf"
	"hcm/pkg/cryptography"
	"hcm/pkg/handler"
	"hcm/pkg/iam/auth"
	"hcm/pkg/logs"
	"hcm/pkg/rest"
	restcli "hcm/pkg/rest/client"
	"hcm/pkg/runtime/shutdown"
	"hcm/pkg/serviced"
	"hcm/pkg/tools/ssl"

	"github.com/emicklei/go-restful/v3"
)

// Service do all the account server's work
type Service struct {
	clientSet  *client.ClientSet
	serve      *http.Server
	authorizer auth.Authorizer
	audit      logicaudit.Interface
}

// NewService create a service instance.
func NewService(sd serviced.ServiceDiscover) (*Service, error) {
	tls := cc.AccountServer().Network.TLS

	var tlsConfig *ssl.TLSConfig
	if tls.Enable() {
		tlsConfig = &ssl.TLSConfig{
			InsecureSkipVerify: tls.InsecureSkipVerify,
			CertFile:           tls.CertFile,
			KeyFile:            tls.KeyFile,
			CAFile:             tls.CAFile,
			Password:           tls.Password,
		}
	}

	// initiate system api client set.
	restCli, err := restcli.NewClient(tlsConfig)
	if err != nil {
		return nil, err
	}
	apiClientSet := client.NewClientSet(restCli, sd)

	// 鉴权
	authorizer, err := auth.NewAuthorizer(sd, tls)
	if err != nil {
		return nil, err
	}

	svr := &Service{
		clientSet:  apiClientSet,
		authorizer: authorizer,
		audit:      logicaudit.NewAudit(apiClientSet.DataService()),
	}

	return svr, nil
}

// newCipherFromConfig 根据配置文件里的加密配置，选择配置的算法并生成对应的加解密器
func newCipherFromConfig(cryptoConfig cc.Crypto) (cryptography.Crypto, error) {
	// TODO: 目前只支持国际加密，还未支持中国国家商业加密，待后续支持再调整
	cfg := cryptoConfig.AesGcm
	return cryptography.NewAESGcm([]byte(cfg.Key), []byte(cfg.Nonce))
}

// ListenAndServeRest listen and serve the restful server
func (s *Service) ListenAndServeRest() error {
	root := http.NewServeMux()
	root.HandleFunc("/", s.apiSet().ServeHTTP)
	root.HandleFunc("/healthz", s.Healthz)
	handler.SetCommonHandler(root)

	network := cc.AccountServer().Network
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
	ws.Path("/api/v1/account")
	ws.Produces(restful.MIME_JSON)

	c := &capability.Capability{
		WebService: ws,
		ApiClient:  s.clientSet,
		Authorizer: s.authorizer,
		Audit:      s.audit,
	}

	mainaccount.InitService(c)
	rootaccount.InitService(c)

	return restful.NewContainer().Add(c.WebService)
}

// Healthz check whether the service is healthy.
func (s *Service) Healthz(w http.ResponseWriter, _ *http.Request) {
	if shutdown.IsShuttingDown() {
		logs.Errorf("service healthz check failed, current service is shutting down")
		w.WriteHeader(http.StatusServiceUnavailable)
		rest.WriteResp(w, rest.NewBaseResp(errf.UnHealthy, "current service is shutting down"))
		return
	}

	if err := serviced.Healthz(cc.AccountServer().Service); err != nil {
		logs.Errorf("serviced healthz check failed, err: %v", err)
		rest.WriteResp(w, rest.NewBaseResp(errf.UnHealthy, "serviced healthz error, "+err.Error()))
		return
	}

	rest.WriteResp(w, rest.NewBaseResp(errf.OK, "healthy"))
	return
}
