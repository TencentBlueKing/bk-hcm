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

// Package serviced defines service discovery related operations.
package serviced

import (
	"context"
	"fmt"
	"sync"

	"hcm/pkg/cc"
	"hcm/pkg/logs"

	etcd3 "go.etcd.io/etcd/client/v3"
)

// ServiceDiscover defines all the service and discovery
// related operations.
type ServiceDiscover interface {
	Service
	Discover
}

// NewServiceD create a service and discovery instance.
func NewServiceD(config cc.Service, svcOpt ServiceOption, discOpt DiscoveryOption) (ServiceDiscover, error) {
	if err := svcOpt.Validate(); err != nil {
		return nil, err
	}

	if err := discOpt.Validate(); err != nil {
		return nil, err
	}

	etcdOpt, err := config.Etcd.ToConfig()
	if err != nil {
		return nil, fmt.Errorf("get etcd config failed, err: %v", err)
	}

	cli, err := etcd3.New(etcdOpt)
	if err != nil {
		return nil, fmt.Errorf("new etcd client failed, err: %v", err)
	}

	ctx, cancel := context.WithCancel(context.Background())
	s := &serviced{
		service: &service{
			cli:    cli,
			svcOpt: svcOpt,
			ctx:    ctx,
			cancel: cancel,
		},
		discovery: &discovery{
			cli:            cli,
			discOpt:        discOpt,
			addresses:      make(map[cc.Name][]serviceAddress),
			addressesRwMux: sync.RWMutex{},
			ctx:            ctx,
			cancel:         cancel,
		},
	}

	// keep synchronizing current node's master state.
	s.syncMasterState()

	// keep synchronizing server addresses for discovery
	s.syncAddresses()
	return s, nil
}

type serviced struct {
	*service
	*discovery
}

// serviceAddress is service address info.
type serviceAddress struct {
	address string
	key     string
	// labels of the server, will be used to match services, if not any labels match, use all others server
	labels map[string]struct{}
	// disableElection disable election, if true, will not election leader, always be follower
	disableElection bool
	// env filter of server, only servers with given env will be used
	env string
}

// etcdCli 用于etcd healthz 检查
var etcdCli *etcd3.Client

var initOnce sync.Once

// Healthz checks the service discovery middleware health state.
func Healthz(ctx context.Context, config cc.Service) error {
	var err error
	initOnce.Do(func() {
		var cfg etcd3.Config
		cfg, err = config.Etcd.ToConfig()
		if err != nil {
			return
		}

		etcdCli, err = etcd3.New(cfg)
		if err != nil {
			return
		}
	})
	if err != nil {
		logs.Errorf("init healthz etcd client failed, err: %v", err)
		return err
	}

	for _, endpoint := range etcdCli.Endpoints() {
		if _, err = etcdCli.Status(ctx, endpoint); err != nil {
			logs.Errorf("etcd healthz check status failed, endpoint: %s, err: %v", endpoint, err)
			return err
		}
	}

	return nil
}
