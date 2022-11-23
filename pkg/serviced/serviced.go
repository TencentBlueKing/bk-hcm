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
	"crypto/tls"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"sync"

	"hcm/pkg/cc"
	"hcm/pkg/tools/ssl"

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
}

// Healthz checks the service discovery middleware health state.
func Healthz(config cc.Service) error {
	etcdConf := config.Etcd

	var tlsConf *tls.Config
	var err error
	scheme := "http"
	if etcdConf.TLS.Enable() {
		if tlsConf, err = ssl.ClientTLSConfVerify(etcdConf.TLS.InsecureSkipVerify, etcdConf.TLS.CAFile,
			etcdConf.TLS.CertFile, etcdConf.TLS.KeyFile, etcdConf.TLS.Password); err != nil {
			return err
		}
		scheme = "https"
	}
	client := &http.Client{Transport: &http.Transport{TLSClientConfig: tlsConf}}

	for _, endpoint := range etcdConf.Endpoints {
		resp, err := client.Get(fmt.Sprintf("%s://%s/health", scheme, endpoint))
		if err != nil {
			return fmt.Errorf("get etcd health failed, err: %v", err)
		}

		if resp.StatusCode != http.StatusOK {
			return fmt.Errorf("response status: %d", resp.StatusCode)
		}

		body, err := ioutil.ReadAll(resp.Body)
		if err != nil {
			return fmt.Errorf("read etcd healthz body failed, err: %v", err)
		}

		info := &HealthInfo{}
		if err := json.Unmarshal(body, info); err != nil {
			return fmt.Errorf("unmarshal etcd healthz info failed, err: %v", err)
		}

		if info.Health != "true" {
			return fmt.Errorf("endpoint %s etcd not healthy", endpoint)
		}
		_ = resp.Body.Close()
	}

	return nil
}
