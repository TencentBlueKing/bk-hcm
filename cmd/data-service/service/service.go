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

	mainaccount "hcm/cmd/data-service/service/account-set/main-account"
	rootaccount "hcm/cmd/data-service/service/account-set/root-account"
	"hcm/cmd/data-service/service/application"
	"hcm/cmd/data-service/service/audit"
	"hcm/cmd/data-service/service/auth"
	"hcm/cmd/data-service/service/bill/billadjustmentitem"
	"hcm/cmd/data-service/service/bill/billdailytask"
	"hcm/cmd/data-service/service/bill/billitem"
	"hcm/cmd/data-service/service/bill/billpuller"
	"hcm/cmd/data-service/service/bill/billsummary"
	"hcm/cmd/data-service/service/bill/billsummaryversion"
	"hcm/cmd/data-service/service/bill/rawbill"
	"hcm/cmd/data-service/service/capability"
	"hcm/cmd/data-service/service/cloud"
	cloudselection "hcm/cmd/data-service/service/cloud-selection"
	"hcm/cmd/data-service/service/cloud/account"
	accountbizrel "hcm/cmd/data-service/service/cloud/account-biz-rel"
	argstpl "hcm/cmd/data-service/service/cloud/argument-template"
	"hcm/cmd/data-service/service/cloud/bill"
	"hcm/cmd/data-service/service/cloud/cert"
	"hcm/cmd/data-service/service/cloud/cvm"
	"hcm/cmd/data-service/service/cloud/disk"
	diskcvmrel "hcm/cmd/data-service/service/cloud/disk-cvm-rel"
	"hcm/cmd/data-service/service/cloud/eip"
	eipcvmrel "hcm/cmd/data-service/service/cloud/eip-cvm-rel"
	"hcm/cmd/data-service/service/cloud/image"
	loadbalancer "hcm/cmd/data-service/service/cloud/load-balancer"
	networkinterface "hcm/cmd/data-service/service/cloud/network-interface"
	networkcvmrel "hcm/cmd/data-service/service/cloud/network-interface-cvm-rel"
	"hcm/cmd/data-service/service/cloud/region"
	resourcegroup "hcm/cmd/data-service/service/cloud/resource-group"
	routetable "hcm/cmd/data-service/service/cloud/route-table"
	securitygroup "hcm/cmd/data-service/service/cloud/security-group"
	sgcomrel "hcm/cmd/data-service/service/cloud/security-group-common-rel"
	sgcvmrel "hcm/cmd/data-service/service/cloud/security-group-cvm-rel"
	subaccount "hcm/cmd/data-service/service/cloud/sub-account"
	sync "hcm/cmd/data-service/service/cloud/sync"
	"hcm/cmd/data-service/service/cloud/zone"
	recyclerecord "hcm/cmd/data-service/service/recycle-record"
	"hcm/cmd/data-service/service/user"
	"hcm/pkg/cc"
	"hcm/pkg/criteria/errf"
	"hcm/pkg/cryptography"
	"hcm/pkg/dal/dao"
	"hcm/pkg/dal/objectstore"
	"hcm/pkg/handler"
	"hcm/pkg/logs"
	"hcm/pkg/metrics"
	"hcm/pkg/rest"
	"hcm/pkg/runtime/shutdown"
	"hcm/pkg/serviced"
	"hcm/pkg/thirdparty/esb"
	"hcm/pkg/tools/ssl"

	"github.com/emicklei/go-restful/v3"
)

// Service do all the data service's work
type Service struct {
	serve       *http.Server
	dao         dao.Set
	cipher      cryptography.Crypto
	esbClient   esb.Client
	objectStore objectstore.Storage
}

// NewService create a service instance.
func NewService() (*Service, error) {
	dao, err := dao.NewDaoSet(cc.DataService().Database)
	if err != nil {
		return nil, err
	}

	// 加解密器
	cipher, err := newCipherFromConfig(cc.DataService().Crypto)
	if err != nil {
		return nil, err
	}

	// esb client
	esbConfig := cc.DataService().Esb
	esbClient, err := esb.NewClient(&esbConfig, metrics.Register())
	if err != nil {
		return nil, err
	}

	// create object store
	oStore, err := objectstore.GetObjectStoreFromEnv()
	if err != nil {
		return nil, err
	}

	svr := &Service{
		dao:         dao,
		cipher:      cipher,
		esbClient:   esbClient,
		objectStore: oStore,
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

	network := cc.DataService().Network
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
	ws.Path("/api/v1/data")
	ws.Produces(restful.MIME_JSON)

	capability := &capability.Capability{
		WebService:  ws,
		Dao:         s.dao,
		Cipher:      s.cipher,
		EsbClient:   s.esbClient,
		ObjectStore: s.objectStore,
	}

	account.InitService(capability)
	accountbizrel.InitService(capability)
	securitygroup.InitSecurityGroupService(capability)
	securitygroup.InitGcpFirewallRuleService(capability)
	cloud.InitVpcService(capability)
	cloud.InitSubnetService(capability)
	cloud.InitCloudService(capability)
	auth.InitAuthService(capability)
	disk.InitService(capability)
	region.InitRegionService(capability)
	resourcegroup.InitAzureResourceGroupService(capability)
	audit.InitAuditService(capability)
	eip.InitEipService(capability)
	zone.InitZoneService(capability)
	image.InitService(capability)
	cvm.InitService(capability)
	sgcvmrel.InitService(capability)
	routetable.InitRouteTableService(capability)
	application.InitApplicationService(capability)
	application.InitApprovalProcessService(capability)
	diskcvmrel.InitService(capability)
	eipcvmrel.InitService(capability)
	networkinterface.InitNetInterfaceService(capability)
	networkcvmrel.InitService(capability)
	recyclerecord.InitRecycleRecordService(capability)
	bill.InitBillConfigService(capability)
	subaccount.InitService(capability)
	sync.InitService(capability)
	user.InitService(capability)
	cloudselection.InitService(capability)
	argstpl.InitService(capability)
	cert.InitService(capability)
	loadbalancer.InitService(capability)
	sgcomrel.InitService(capability)
	mainaccount.InitService(capability)
	rootaccount.InitService(capability)

	billpuller.InitService(capability)
	billsummary.InitService(capability)
	billsummaryversion.InitService(capability)
	billitem.InitService(capability)
	billdailytask.InitService(capability)
	billadjustmentitem.InitService(capability)
	if capability.ObjectStore != nil {
		rawbill.InitService(capability)
	}
	cert.InitService(capability)
	loadbalancer.InitService(capability)
	sgcomrel.InitService(capability)

	return restful.NewContainer().Add(capability.WebService)
}

// Healthz check whether the service is healthy.
func (s *Service) Healthz(w http.ResponseWriter, _ *http.Request) {

	if err := serviced.Healthz(cc.DataService().Service); err != nil {
		logs.Errorf("etcd healthz check failed, err: %v", err)
		rest.WriteResp(w, rest.NewBaseResp(errf.UnHealthy, "etcd healthz error, "+err.Error()))
		return
	}

	rest.WriteResp(w, rest.NewBaseResp(errf.OK, "healthy"))
	return
}
