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
	"sync"

	"hcm/pkg/cc"

	etcd3 "go.etcd.io/etcd/client/v3"
)

// ServiceDiscover defines all the service and discovery
// related operations.
type ServiceDiscover interface {
	Service
	Discover
}

// NewServiceD create a service and discovery instance.
func NewServiceD(cli *etcd3.Client, svcOpt ServiceOption, discOpt DiscoveryOption) (ServiceDiscover, error) {
	if err := svcOpt.Validate(); err != nil {
		return nil, err
	}

	if err := discOpt.Validate(); err != nil {
		return nil, err
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
}
