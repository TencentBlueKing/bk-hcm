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

	"hcm/cmd/cloud-server/logics"
	logicaudit "hcm/cmd/cloud-server/logics/audit"
	"hcm/cmd/cloud-server/service/account"
	"hcm/cmd/cloud-server/service/application"
	appcvm "hcm/cmd/cloud-server/service/application/handlers/cvm"
	approvalprocess "hcm/cmd/cloud-server/service/approval_process"
	argstpl "hcm/cmd/cloud-server/service/argument-template"
	"hcm/cmd/cloud-server/service/assign"
	"hcm/cmd/cloud-server/service/audit"
	"hcm/cmd/cloud-server/service/bill"
	"hcm/cmd/cloud-server/service/capability"
	cloudselection "hcm/cmd/cloud-server/service/cloud-selection"
	"hcm/cmd/cloud-server/service/cvm"
	"hcm/cmd/cloud-server/service/disk"
	"hcm/cmd/cloud-server/service/eip"
	"hcm/cmd/cloud-server/service/firewall"
	"hcm/cmd/cloud-server/service/image"
	instancetype "hcm/cmd/cloud-server/service/instance-type"
	networkinterface "hcm/cmd/cloud-server/service/network-interface"
	"hcm/cmd/cloud-server/service/recycle"
	"hcm/cmd/cloud-server/service/region"
	resourcegroup "hcm/cmd/cloud-server/service/resource-group"
	routetable "hcm/cmd/cloud-server/service/route-table"
	securitygroup "hcm/cmd/cloud-server/service/security-group"
	subaccount "hcm/cmd/cloud-server/service/sub-account"
	"hcm/cmd/cloud-server/service/subnet"
	"hcm/cmd/cloud-server/service/sync"
	"hcm/cmd/cloud-server/service/sync/lock"
	"hcm/cmd/cloud-server/service/user"
	"hcm/cmd/cloud-server/service/vpc"
	"hcm/cmd/cloud-server/service/zone"
	"hcm/pkg/cc"
	"hcm/pkg/client"
	"hcm/pkg/criteria/errf"
	"hcm/pkg/cryptography"
	"hcm/pkg/handler"
	"hcm/pkg/iam/auth"
	"hcm/pkg/logs"
	"hcm/pkg/metrics"
	"hcm/pkg/rest"
	restcli "hcm/pkg/rest/client"
	"hcm/pkg/runtime/shutdown"
	"hcm/pkg/serviced"
	"hcm/pkg/thirdparty/api-gateway/bkbase"
	"hcm/pkg/thirdparty/api-gateway/itsm"
	"hcm/pkg/thirdparty/esb"
	"hcm/pkg/tools/ssl"

	"github.com/emicklei/go-restful/v3"
)

// Service do all the cloud server's work
type Service struct {
	client     *client.ClientSet
	serve      *http.Server
	authorizer auth.Authorizer
	audit      logicaudit.Interface
	cipher     cryptography.Crypto
	// EsbClient 调用接入ESB的第三方系统API集合
	esbClient esb.Client
	// itsmCli itsm client.
	itsmCli   itsm.Client
	bkBaseCli bkbase.Client
}

// NewService create a service instance.
func NewService(sd serviced.ServiceDiscover) (*Service, error) {
	tls := cc.CloudServer().Network.TLS
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
	authorizer, err := auth.NewAuthorizer(sd, tls)
	if err != nil {
		return nil, err
	}

	// 加解密器
	cipher, err := newCipherFromConfig(cc.CloudServer().Crypto)
	if err != nil {
		return nil, err
	}

	// 创建ESB Client
	esbConfig := cc.CloudServer().Esb
	esbClient, err := esb.NewClient(&esbConfig, metrics.Register())
	if err != nil {
		return nil, err
	}

	itsmCfg := cc.CloudServer().Itsm
	itsmCli, err := itsm.NewClient(&itsmCfg, metrics.Register())
	if err != nil {
		logs.Errorf("failed to create itsm client, err: %v", err)
		return nil, err
	}

	bkbaseCfg := cc.CloudServer().CloudSelection.BkBase
	bkbaseCli, err := bkbase.NewClient(&bkbaseCfg.ApiGateway, metrics.Register())
	if err != nil {
		logs.Errorf("failed to create bkbase client, err: %v", err)
		return nil, err
	}

	svr := &Service{
		client:     apiClientSet,
		authorizer: authorizer,
		audit:      logicaudit.NewAudit(apiClientSet.DataService()),
		cipher:     cipher,
		esbClient:  esbClient,
		itsmCli:    itsmCli,
		bkBaseCli:  bkbaseCli,
	}

	etcdCfg, err := cc.CloudServer().Service.Etcd.ToConfig()
	if err != nil {
		return nil, err
	}
	err = lock.InitManger(etcdCfg, int64(cc.CloudServer().CloudResource.Sync.SyncFrequencyLimitingTimeMin)*60)
	if err != nil {
		return nil, err
	}
	if cc.CloudServer().CloudResource.Sync.Enable {
		interval := time.Duration(cc.CloudServer().CloudResource.Sync.SyncIntervalMin) * time.Minute
		go sync.CloudResourceSync(interval, sd, apiClientSet)
	}
	if cc.CloudServer().BillConfig.Enable {
		interval := time.Duration(cc.CloudServer().BillConfig.SyncIntervalMin) * time.Minute
		go bill.CloudBillConfigCreate(interval, sd, apiClientSet)
	}
	recycle.RecycleTiming(apiClientSet, sd, cc.CloudServer().Recycle, esbClient)

	go appcvm.TimingHandleDeliverApplication(svr.client, 2*time.Second)
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
	root.HandleFunc("/", s.apiSet(cc.CloudServer().BkHcmUrl).ServeHTTP)
	root.HandleFunc("/healthz", s.Healthz)
	handler.SetCommonHandler(root)

	network := cc.CloudServer().Network
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

func (s *Service) apiSet(bkHcmUrl string) *restful.Container {
	ws := new(restful.WebService)
	ws.Path("/api/v1/cloud")
	ws.Produces(restful.MIME_JSON)

	c := &capability.Capability{
		WebService: ws,
		ApiClient:  s.client,
		Authorizer: s.authorizer,
		Audit:      s.audit,
		Cipher:     s.cipher,
		EsbClient:  s.esbClient,
		Logics:     logics.NewLogics(s.client, s.esbClient),
		ItsmCli:    s.itsmCli,
		BKBaseCli:  s.bkBaseCli,
	}

	account.InitAccountService(c)
	securitygroup.InitSecurityGroupService(c)
	firewall.InitFirewallService(c)
	vpc.InitVpcService(c)
	disk.InitDiskService(c)
	subnet.InitSubnetService(c)
	image.InitImageService(c)
	routetable.InitRouteTableService(c)
	cvm.InitCvmService(c)
	resourcegroup.InitResourceGroupService(c)
	zone.InitZoneService(c)
	region.InitRegionService(c)
	eip.InitEipService(c)
	instancetype.InitInstanceTypeService(c)
	networkinterface.InitNetworkInterfaceService(c)
	subaccount.InitService(c)

	application.InitApplicationService(c, bkHcmUrl)
	audit.InitService(c)
	assign.InitService(c)
	recycle.InitService(c)
	bill.InitBillService(c)

	user.InitService(c)

	approvalprocess.InitService(c)
	cloudselection.InitService(c)
	argstpl.InitArgsTplService(c)

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

	if err := serviced.Healthz(cc.CloudServer().Service); err != nil {
		logs.Errorf("serviced healthz check failed, err: %v", err)
		rest.WriteResp(w, rest.NewBaseResp(errf.UnHealthy, "serviced healthz error, "+err.Error()))
		return
	}

	rest.WriteResp(w, rest.NewBaseResp(errf.OK, "healthy"))
	return
}
