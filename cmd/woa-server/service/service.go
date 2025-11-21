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

	"hcm/cmd/woa-server/logics/biz"
	configlogic "hcm/cmd/woa-server/logics/config"
	conflogics "hcm/cmd/woa-server/logics/config"
	cvmlogic "hcm/cmd/woa-server/logics/cvm"
	disLogics "hcm/cmd/woa-server/logics/dissolve"
	gclogics "hcm/cmd/woa-server/logics/green-channel"
	planctrl "hcm/cmd/woa-server/logics/plan"
	ressynclogics "hcm/cmd/woa-server/logics/res-sync"
	rslogics "hcm/cmd/woa-server/logics/rolling-server"
	srlogics "hcm/cmd/woa-server/logics/short-rental"
	taskLogics "hcm/cmd/woa-server/logics/task"
	"hcm/cmd/woa-server/logics/task/informer"
	"hcm/cmd/woa-server/logics/task/operation"
	"hcm/cmd/woa-server/logics/task/recoverer"
	"hcm/cmd/woa-server/logics/task/recycler"
	"hcm/cmd/woa-server/logics/task/scheduler"
	taskStatistics "hcm/cmd/woa-server/logics/task/statistics"
	"hcm/cmd/woa-server/service/capability"
	"hcm/cmd/woa-server/service/config"
	"hcm/cmd/woa-server/service/cvm"
	"hcm/cmd/woa-server/service/dissolve"
	greenchannel "hcm/cmd/woa-server/service/green-channel"
	"hcm/cmd/woa-server/service/meta"
	"hcm/cmd/woa-server/service/plan"
	"hcm/cmd/woa-server/service/pool"
	ressync "hcm/cmd/woa-server/service/res-sync"
	rollingserver "hcm/cmd/woa-server/service/rolling-server"
	"hcm/cmd/woa-server/service/task"
	"hcm/cmd/woa-server/storage/dal/mongo"
	"hcm/cmd/woa-server/storage/dal/mongo/local"
	"hcm/cmd/woa-server/storage/dal/redis"
	"hcm/cmd/woa-server/storage/driver/mongodb"
	redisCli "hcm/cmd/woa-server/storage/driver/redis"
	"hcm/cmd/woa-server/storage/stream"
	"hcm/pkg/api/core"
	"hcm/pkg/cc"
	"hcm/pkg/client"
	"hcm/pkg/criteria/constant"
	"hcm/pkg/criteria/errf"
	"hcm/pkg/dal/dao"
	"hcm/pkg/handler"
	"hcm/pkg/iam/auth"
	"hcm/pkg/kit"
	"hcm/pkg/logs"
	"hcm/pkg/metrics"
	"hcm/pkg/rest"
	restcli "hcm/pkg/rest/client"
	"hcm/pkg/runtime/shutdown"
	"hcm/pkg/serviced"
	"hcm/pkg/thirdparty"
	"hcm/pkg/thirdparty/api-gateway/cmdb"
	"hcm/pkg/thirdparty/api-gateway/cmsi"
	"hcm/pkg/thirdparty/api-gateway/itsm"
	"hcm/pkg/thirdparty/es"
	"hcm/pkg/tools/ssl"

	"github.com/emicklei/go-restful/v3"
)

// Service do all the woa server's work
type Service struct {
	client         *client.ClientSet
	dao            dao.Set
	planController planctrl.Logics
	cmdbCli        cmdb.Client
	itsmCli        itsm.Client
	// authorizer 鉴权所需接口集合
	authorizer     auth.Authorizer
	thirdCli       *thirdparty.Client
	clientConf     cc.WoaServerSetting
	schedulerIf    scheduler.Interface
	informerIf     informer.Interface
	recyclerIf     recycler.Interface
	operationIf    operation.Interface
	esCli          *es.EsCli
	rsLogic        rslogics.Logics
	srLogic        srlogics.Logics
	gcLogic        gclogics.Logics
	bizLogic       biz.Logics
	dissolveLogic  disLogics.Logics
	resSyncLogic   ressynclogics.Logics
	configLogics   configlogic.Logics
	taskLogic      taskLogics.Logics
	taskStatistics taskStatistics.Interface
	cvmLogic       cvmlogic.Logics
}

// NewService create a service instance.
func NewService(dis serviced.ServiceDiscover, sd serviced.State) (*Service, error) {
	tlsConfig, err := initTLSConfig()
	if err != nil {
		return nil, err
	}

	apiClientSet, err := initAPIClient(tlsConfig, dis)
	if err != nil {
		return nil, err
	}

	clients, err := initClients(apiClientSet, dis)
	if err != nil {
		return nil, err
	}

	logics, err := initLogics(sd, apiClientSet, clients)
	if err != nil {
		return nil, err
	}

	mongoComponents, err := initMongoComponents(dis, clients, logics)
	if err != nil {
		return nil, err
	}

	service := assembleService(apiClientSet, clients, logics, mongoComponents)
	return newOtherClient(core.NewBackendKit(), service, clients.itsmCli, sd)
}

// initTLSConfig 初始化TLS配置
func initTLSConfig() (*ssl.TLSConfig, error) {
	tls := cc.WoaServer().Network.TLS
	if !tls.Enable() {
		return nil, nil
	}

	return &ssl.TLSConfig{
		InsecureSkipVerify: tls.InsecureSkipVerify,
		CertFile:           tls.CertFile,
		KeyFile:            tls.KeyFile,
		CAFile:             tls.CAFile,
		Password:           tls.Password,
	}, nil
}

// initAPIClient 初始化API客户端
func initAPIClient(tlsConfig *ssl.TLSConfig, dis serviced.ServiceDiscover) (*client.ClientSet, error) {
	restCli, err := restcli.NewClient(tlsConfig)
	if err != nil {
		return nil, err
	}
	return client.NewClientSet(restCli, dis), nil
}

// clientSet 封装所有客户端
type clientSet struct {
	daoSet     dao.Set
	esCli      *es.EsCli
	cmdbCli    cmdb.Client
	itsmCli    itsm.Client
	cmsiCli    cmsi.Client
	authorizer auth.Authorizer
	thirdCli   *thirdparty.Client
}

// initClients 初始化所有客户端
func initClients(apiClientSet *client.ClientSet, dis serviced.ServiceDiscover) (*clientSet, error) {
	clients := &clientSet{}

	// init db client
	daoSet, err := dao.NewDaoSet(cc.WoaServer().Database)
	if err != nil {
		return nil, err
	}
	clients.daoSet = daoSet

	// init CMDB client
	cmdbConfig := cc.WoaServer().Cmdb
	if err = cmdb.InitCmdbClient(&cmdbConfig, metrics.Register()); err != nil {
		return nil, err
	}
	clients.cmdbCli = cmdb.CmdbClient()

	// init ITSM client
	itsmCfg := cc.WoaServer().ITSM
	itsmCli, err := itsm.NewClient(&itsmCfg, metrics.Register())
	if err != nil {
		return nil, err
	}
	clients.itsmCli = itsmCli

	// 创建调用第三方平台Client
	thirdCli, err := thirdparty.NewClient(cc.WoaServer().ClientConfig, metrics.Register())
	if err != nil {
		return nil, err
	}
	clients.thirdCli = thirdCli

	// init CMSI client
	cmsiCfg := cc.WoaServer().Cmsi
	cmsiCli, err := cmsi.NewClient(&cmsiCfg, metrics.Register())
	if err != nil {
		logs.Errorf("failed to create cmsi client, err: %v", err)
		return nil, err
	}
	clients.cmsiCli = cmsiCli

	// init elasticsearch client
	esCli, err := es.NewEsClient(cc.WoaServer().Es, cc.WoaServer().Blacklist)
	if err != nil {
		return nil, err
	}
	clients.esCli = esCli

	// create authorizer
	authorizer, err := auth.NewAuthorizer(dis, cc.WoaServer().Network.TLS)
	if err != nil {
		return nil, err
	}
	clients.authorizer = authorizer

	// init redis client
	if err := initRedisClient(); err != nil {
		return nil, err
	}

	return clients, nil
}

// initRedisClient 初始化Redis客户端
func initRedisClient() error {
	rConf := cc.WoaServer().Redis
	redisConf, err := redis.NewConf(&rConf)
	if err != nil {
		return err
	}
	return redisCli.InitClient("redis", redisConf)
}

// logicSet 封装所有逻辑组件
type logicSet struct {
	configLogics   configlogic.Logics
	gcLogics       gclogics.Logics
	bizLogic       biz.Logics
	rsLogics       rslogics.Logics
	srLogics       srlogics.Logics
	planCtrl       planctrl.Logics
	esCli          *es.EsCli
	dissolveLogics disLogics.Logics
}

// initLogics 初始化所有逻辑组件
func initLogics(sd serviced.State, apiClientSet *client.ClientSet, clients *clientSet) (*logicSet, error) {
	logics := &logicSet{}

	// new config logic
	logics.configLogics = conflogics.New(apiClientSet, clients.thirdCli, clients.cmdbCli, clients.daoSet)

	// new green channel logic
	gcLogics, err := gclogics.New(apiClientSet, logics.configLogics)
	if err != nil {
		logs.Errorf("new green channel logics failed, err: %v", err)
		return nil, err
	}
	logics.gcLogics = gcLogics

	// new business logic
	bizLogic, err := biz.New(clients.cmdbCli, clients.authorizer)
	if err != nil {
		logs.Errorf("new biz logic failed, err: %v", err)
		return nil, err
	}
	logics.bizLogic = bizLogic

	// new rolling server logic
	rsLogics, err := rslogics.New(sd, apiClientSet, clients.cmdbCli, clients.thirdCli, bizLogic, clients.cmsiCli,
		logics.configLogics)
	if err != nil {
		logs.Errorf("new rolling server logics failed, err: %v", err)
		return nil, err
	}
	logics.rsLogics = rsLogics

	// new short rental logic
	srLogics, err := srlogics.New(sd, apiClientSet, clients.thirdCli, bizLogic, clients.cmsiCli, logics.configLogics)
	if err != nil {
		logs.Errorf("new short rental logics failed, err: %v", err)
		return nil, err
	}
	logics.srLogics = srLogics

	// new resource plan controller
	planCtrl, err := planctrl.New(sd, apiClientSet, clients.daoSet, clients.cmsiCli, clients.itsmCli,
		clients.thirdCli.CVM, bizLogic)
	if err != nil {
		logs.Errorf("new plan controller failed, err: %v", err)
		return nil, err
	}
	logics.planCtrl = planCtrl

	// new dissolve logic
	logics.dissolveLogics = disLogics.New(clients.daoSet, clients.cmdbCli, clients.esCli, clients.thirdCli,
		cc.WoaServer(), logics.configLogics, apiClientSet)

	return logics, nil
}

// mongoComponentSet 封装MongoDB相关组件
type mongoComponentSet struct {
	informerIf  informer.Interface
	schedulerIf scheduler.Interface
}

// initMongoComponents 初始化涉及MongoDB的逻辑
func initMongoComponents(dis serviced.ServiceDiscover, clients *clientSet, logics *logicSet) (*mongoComponentSet,
	error) {
	if !cc.WoaServer().UseMongo {
		return &mongoComponentSet{}, nil
	}

	kt := core.NewBackendKit()
	loopW, watchDB, err := initMongoDB(kt, dis)
	if err != nil {
		return nil, err
	}

	informerIf, err := informer.New(loopW, watchDB)
	if err != nil {
		logs.Errorf("new informer failed, err: %v, rid: %s", err, kt.Rid)
		return nil, err
	}

	schedulerIf, err := scheduler.New(kt.Ctx, logics.rsLogics, logics.srLogics, logics.gcLogics,
		clients.thirdCli, clients.cmdbCli, informerIf, cc.WoaServer().ClientConfig,
		logics.planCtrl, logics.bizLogic, logics.configLogics)
	if err != nil {
		logs.Errorf("new scheduler failed, err: %v, rid: %s", err, kt.Rid)
		return nil, err
	}

	return &mongoComponentSet{
		informerIf:  informerIf,
		schedulerIf: schedulerIf,
	}, nil
}

// assembleService 组装Service结构体
func assembleService(apiClientSet *client.ClientSet, clients *clientSet, logics *logicSet,
	mongoComponents *mongoComponentSet) *Service {
	return &Service{
		client:         apiClientSet,
		dao:            clients.daoSet,
		cmdbCli:        clients.cmdbCli,
		authorizer:     clients.authorizer,
		thirdCli:       clients.thirdCli,
		clientConf:     cc.WoaServer(),
		informerIf:     mongoComponents.informerIf,
		schedulerIf:    mongoComponents.schedulerIf,
		esCli:          logics.esCli,
		rsLogic:        logics.rsLogics,
		srLogic:        logics.srLogics,
		gcLogic:        logics.gcLogics,
		planController: logics.planCtrl,
		bizLogic:       logics.bizLogic,
		dissolveLogic:  logics.dissolveLogics,
		configLogics:   logics.configLogics,
	}
}

// initMongoDB init mongodb client and watch client
func initMongoDB(kt *kit.Kit, dis serviced.ServiceDiscover) (stream.LoopInterface, *local.Mongo, error) {
	// init mongodb client
	mConf := cc.WoaServer().MongoDB
	mongoConf, err := mongo.NewConf(&mConf)
	if err != nil {
		return nil, nil, err
	}
	if err = mongodb.InitClient("", mongoConf); err != nil {
		return nil, nil, err
	}

	wConf := cc.WoaServer().Watch
	watchConf, err := mongo.NewConf(&wConf)
	if err != nil {
		return nil, nil, err
	}

	if err = mongodb.InitClient("", watchConf); err != nil {
		return nil, nil, err
	}

	// init task service logics
	loopW, err := stream.NewLoopStream(mongoConf.GetMongoConf(), dis)
	if err != nil {
		logs.Errorf("new loop stream failed, err: %v, rid: %s", err, kt.Rid)
		return nil, nil, err
	}

	watchDB, err := local.NewMgo(watchConf.GetMongoConf(), time.Minute)
	if err != nil {
		logs.Errorf("new watch mongo client failed, err: %v, rid: %s", err, kt.Rid)
		return nil, nil, err
	}
	return loopW, watchDB, err
}

func newOtherClient(kt *kit.Kit, service *Service, itsmCli itsm.Client, sd serviced.State) (*Service, error) {

	recyclerIf, err := recycler.New(kt.Ctx, service.thirdCli, service.bizLogic, service.cmdbCli, service.authorizer,
		service.rsLogic, service.srLogic, service.dissolveLogic, service.client, service.configLogics,
		service.planController)
	if err != nil {
		logs.Errorf("new recycler failed, err: %v, rid: %s", err, kt.Rid)
		return nil, err
	}

	go func() {
		defer func() {
			if err := recover(); err != nil {
				logs.Errorf("[%s] recycler stuck check loop exit unexpectedly, err: %v, rid: %s",
					constant.CvmRecycleStuck, err, kt.Rid)
			}
		}()

		recyclerIf.StartStuckCheckLoop(kt.NewSubKit())
	}()

	operationIf, err := operation.New(kt.Ctx, service.client)
	if err != nil {
		logs.Errorf("new operation failed, err: %v, rid: %s", err, kt.Rid)
		return nil, err
	}

	statisticsIf := taskStatistics.New(service.client)
	service.taskStatistics = statisticsIf

	taskLogic := taskLogics.New(service.schedulerIf, recyclerIf, service.informerIf, operationIf, statisticsIf)
	service.taskLogic = taskLogic

	cvmLogic := cvmlogic.New(service.thirdCli, service.clientConf.ClientConfig,
		service.configLogics, service.cmdbCli, service.rsLogic, service.taskLogic, service.schedulerIf)
	service.cvmLogic = cvmLogic

	// init recoverer client
	recoverConf := cc.WoaServer().Recover
	if err := recoverer.New(kt, &recoverConf, itsmCli, recyclerIf, service.schedulerIf, cvmLogic,
		service.cmdbCli, service.thirdCli.Sops, sd); err != nil {
		logs.Errorf("new recoverer failed, err: %v, rid: %s", err, kt.Rid)
		return nil, err
	}

	resSyncLogics, err := ressynclogics.New(sd, service.client, service.configLogics)
	if err != nil {
		logs.Errorf("new resource sync logics failed, err: %v", err)
		return nil, err
	}

	service.clientConf = cc.WoaServer()
	service.recyclerIf = recyclerIf
	service.operationIf = operationIf
	service.resSyncLogic = resSyncLogics
	return service, nil
}

// ListenAndServeRest listen and serve the restful server
func (s *Service) ListenAndServeRest() error {
	root := http.NewServeMux()
	root.HandleFunc("/", s.apiSet().ServeHTTP)
	root.HandleFunc("/healthz", s.Healthz)
	handler.SetCommonHandler(root)

	network := cc.WoaServer().Network
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

func (s *Service) apiSet() *restful.Container {
	ws := new(restful.WebService)
	ws.Path("/api/v1/woa")
	ws.Produces(restful.MIME_JSON)

	c := &capability.Capability{
		Dao:            s.dao,
		WebService:     ws,
		Authorizer:     s.authorizer,
		PlanController: s.planController,
		CmdbCli:        s.cmdbCli,
		ThirdCli:       s.thirdCli,
		Conf:           s.clientConf,
		SchedulerIf:    s.schedulerIf,
		InformerIf:     s.informerIf,
		RecyclerIf:     s.recyclerIf,
		OperationIf:    s.operationIf,
		EsCli:          s.esCli,
		RsLogic:        s.rsLogic,
		Client:         s.client,
		GcLogic:        s.gcLogic,
		BizLogic:       s.bizLogic,
		DissolveLogic:  s.dissolveLogic,
		ResSyncLogic:   s.resSyncLogic,
		ConfigLogics:   s.configLogics,
		TaskLogic:      s.taskLogic,
		TaskStatistics: s.taskStatistics,
		CvmLogic:       s.cvmLogic,
	}

	config.InitService(c)
	pool.InitService(c)
	cvm.InitService(c)
	task.InitService(c)
	meta.InitService(c)
	plan.InitService(c)
	dissolve.InitService(c)
	rollingserver.InitService(c)
	greenchannel.InitService(c)
	ressync.InitService(c)

	return restful.NewContainer().Add(c.WebService)
}

// Healthz service health check.
func (s *Service) Healthz(w http.ResponseWriter, r *http.Request) {
	if shutdown.IsShuttingDown() {
		logs.Errorf("service healthz check failed, current service is shutting down")
		w.WriteHeader(http.StatusServiceUnavailable)
		rest.WriteResp(w, rest.NewBaseResp(errf.UnHealthy, "current service is shutting down"))
		return
	}

	if err := serviced.Healthz(r.Context(), cc.WoaServer().Service); err != nil {
		logs.Errorf("serviced healthz check failed, err: %v", err)
		rest.WriteResp(w, rest.NewBaseResp(errf.UnHealthy, "serviced healthz error, "+err.Error()))
		return
	}

	rest.WriteResp(w, rest.NewBaseResp(errf.OK, "healthy"))
	return
}
