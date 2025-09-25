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

	logicsaction "hcm/cmd/task-server/logics/action"
	"hcm/cmd/task-server/service/capability"
	"hcm/cmd/task-server/service/controller"
	"hcm/cmd/task-server/service/producer"
	"hcm/cmd/task-server/service/viewer"
	"hcm/pkg/async"
	"hcm/pkg/async/backend"
	"hcm/pkg/async/consumer"
	"hcm/pkg/async/consumer/leader"
	"hcm/pkg/cc"
	"hcm/pkg/client"
	"hcm/pkg/client/data-service/global"
	"hcm/pkg/criteria/enumor"
	"hcm/pkg/criteria/errf"
	"hcm/pkg/dal/dao"
	"hcm/pkg/handler"
	"hcm/pkg/iam/auth"
	"hcm/pkg/logs"
	"hcm/pkg/metrics"
	"hcm/pkg/rest"
	restcli "hcm/pkg/rest/client"
	"hcm/pkg/runtime/shutdown"
	"hcm/pkg/serviced"
	"hcm/pkg/tools/ssl"

	"github.com/emicklei/go-restful/v3"
)

// Service do all the task server's work
type Service struct {
	client     *client.ClientSet
	dao        dao.Set
	serve      *http.Server
	async      async.Async
	authorizer auth.Authorizer
}

// NewService create a service instance.
func NewService(sd serviced.ServiceDiscover, shutdownWaitTimeSec int) (*Service, error) {
	tls := cc.TaskServer().Network.TLS

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

	// init db client
	dao, err := dao.NewDaoSet(cc.TaskServer().Database)
	if err != nil {
		return nil, err
	}

	logicsaction.Init(apiClientSet, dao)
	async, err := createAndStartAsync(sd, apiClientSet.DataService().Global.GlobalConfig, dao, shutdownWaitTimeSec)
	if err != nil {
		return nil, err
	}

	// 鉴权
	authorizer, err := auth.NewAuthorizer(sd, tls)
	svr := &Service{
		client:     apiClientSet,
		dao:        dao,
		async:      async,
		authorizer: authorizer,
	}

	return svr, nil
}

func createAndStartAsync(sd serviced.ServiceDiscover, globalCfgCli *global.GlobalConfigsClient, dao dao.Set, shutdownWaitTimeSec int) (async.Async, error) {
	// 创建async框架使用的backend
	bd, err := backend.Factory(enumor.BackendMysql, dao)
	if err != nil {
		return nil, err
	}

	leader := leader.NewLeader(sd)
	cfg := cc.TaskServer().Async
	opt := &async.Option{
		Register: metrics.Register(),
		ConsumerOption: &consumer.Option{
			Scheduler: &consumer.SchedulerOption{
				WatchIntervalSec:                cfg.Scheduler.WatchIntervalSec,
				WorkerNumber:                    cfg.Scheduler.WorkerNumber,
				ScheduledFlowFetcherConcurrency: cfg.Scheduler.ScheduledFlowFetcherConcurrency,
				CanceledFlowFetcherConcurrency:  cfg.Scheduler.CanceledFlowFetcherConcurrency,
			},
			Executor: &consumer.ExecutorOption{
				WorkerNumber:          cfg.Executor.WorkerNumber,
				TaskExecTimeoutSec:    cfg.Executor.TaskExecTimeoutSec,
				InitQueueCapacity:     cfg.Executor.InitQueueCapacity,
				FastTaskWorkerRatio:   cfg.Executor.FastTaskWorkerRatio,
				FastTaskThresholdSec:  cfg.Executor.FastTaskThresholdSec,
				TimeWindowCapacity:    cfg.Executor.TimeWindowCapacity,
				TimeWindowDurationMin: cfg.Executor.TimeWindowDurationMin,
				FastTaskQueueCapacity: cfg.Executor.FastTaskQueueCapacity,
				SlowTaskQueueCapacity: cfg.Executor.SlowTaskQueueCapacity,
			},
			Dispatcher: &consumer.DispatcherOption{
				WatchIntervalSec:              cfg.Dispatcher.WatchIntervalSec,
				PendingFlowFetcherConcurrency: cfg.Dispatcher.PendingFlowFetcherConcurrency,
			},
			WatchDog: &consumer.WatchDogOption{
				WatchIntervalSec:    cfg.WatchDog.WatchIntervalSec,
				TaskRunTimeoutSec:   cfg.WatchDog.TaskTimeoutSec,
				ShutdownWaitTimeSec: uint(shutdownWaitTimeSec),
				WorkerNumber:        cfg.WatchDog.WorkerNumber,
			},
		},
	}
	async, err := async.NewAsync(bd, leader, opt)
	if err != nil {
		return nil, err
	}

	go func() {
		notifier := shutdown.AddNotifier()
		select {
		case <-notifier.Signal:
			defer notifier.Done()
			logs.Infof("start shutdown async consumer gracefully...")
			async.GetConsumer().Close()
			logs.Infof("shutdown async consumer success...")
		}
	}()

	if err = async.GetConsumer().Start(globalCfgCli); err != nil {
		return nil, err
	}

	return async, nil
}

// ListenAndServeRest listen and serve the restful server
func (s *Service) ListenAndServeRest() error {
	root := http.NewServeMux()
	root.HandleFunc("/", s.apiSet().ServeHTTP)
	root.HandleFunc("/healthz", s.Healthz)
	handler.SetCommonHandler(root)

	network := cc.TaskServer().Network
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
	ws.Path("/api/v1/task")
	ws.Produces(restful.MIME_JSON)

	c := &capability.Capability{
		WebService: ws,
		ApiClient:  s.client,
		Async:      s.async,
		Dao:        s.dao,
		Authorizer: s.authorizer,
	}

	producer.Init(c)
	viewer.Init(c)
	controller.Init(c)

	return restful.NewContainer().Add(c.WebService)
}

// Healthz check whether the service is healthy.
func (s *Service) Healthz(w http.ResponseWriter, r *http.Request) {
	if shutdown.IsShuttingDown() {
		logs.Errorf("service healthz check failed, current service is shutting down")
		w.WriteHeader(http.StatusServiceUnavailable)
		rest.WriteResp(w, rest.NewBaseResp(errf.UnHealthy, "current service is shutting down"))
		return
	}

	if err := serviced.Healthz(r.Context(), cc.TaskServer().Service); err != nil {
		logs.Errorf("serviced healthz check failed, err: %v", err)
		rest.WriteResp(w, rest.NewBaseResp(errf.UnHealthy, "serviced healthz error, "+err.Error()))
		return
	}

	rest.WriteResp(w, rest.NewBaseResp(errf.OK, "healthy"))
	return
}
