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

package serviced

import (
	"context"
	"fmt"
	"sync"

	"hcm/pkg/cc"
	"hcm/pkg/logs"

	"go.etcd.io/etcd/api/v3/mvccpb"
	etcd3 "go.etcd.io/etcd/client/v3"
)

// Discover defines service discovery related operations.
type Discover interface {
	Discover(cc.Name) ([]string, error)
	Services() []cc.Name
	Node
}

// Node defines the discovery's nodes related operations.
type Node interface {
	// GetServiceAllNodeKeys 获取当前服务全部节点Key
	GetServiceAllNodeKeys(name cc.Name) ([]string, error)
}

// NewDiscovery create a service discovery instance.
func NewDiscovery(config cc.Service, opt DiscoveryOption) (Discover, error) {
	if err := opt.Validate(); err != nil {
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
	d := &discovery{
		cli:       cli,
		discOpt:   opt,
		ctx:       ctx,
		cancel:    cancel,
		addresses: make(map[cc.Name][]serviceAddress),
	}

	// keep synchronizing server addresses for discovery
	d.syncAddresses()
	return d, nil
}

type discovery struct {
	cli     *etcd3.Client
	discOpt DiscoveryOption

	// addresses is service name to service address map for discovery.
	addresses      map[cc.Name][]serviceAddress
	addressesRwMux sync.RWMutex

	ctx    context.Context
	cancel context.CancelFunc
}

// GetServiceAllNodeKeys 获取当前服务全部节点Key
func (d *discovery) GetServiceAllNodeKeys(name cc.Name) ([]string, error) {

	resp, err := d.cli.Get(context.Background(), ServiceDiscoveryName(name), etcd3.WithPrefix())
	if err != nil {
		return nil, err
	}

	keys := make([]string, 0, len(resp.Kvs))
	for _, one := range resp.Kvs {
		keys = append(keys, string(one.Key))
	}

	return keys, nil
}

// Services returns the being discovered services
func (d *discovery) Services() []cc.Name {
	return d.discOpt.Services
}

// Discover service addresses by service name.
func (d *discovery) Discover(name cc.Name) ([]string, error) {
	nameExists := false
	for _, service := range d.discOpt.Services {
		if name == service {
			nameExists = true
			break
		}
	}

	if !nameExists {
		return nil, fmt.Errorf("service %s is not in supported discover servers", name)
	}

	d.addressesRwMux.RLock()
	addressInfos := d.addresses[name]
	d.addressesRwMux.RUnlock()

	addresses := make([]string, 0)
	for _, addressInfo := range addressInfos {
		addresses = append(addresses, addressInfo.address)
	}

	return addresses, nil
}

// syncAddresses synchronize server addresses for discovery.
func (d *discovery) syncAddresses() {
	d.addresses = make(map[cc.Name][]serviceAddress)

	for _, service := range d.discOpt.Services {
		d.watcher(service)
	}
}

// watcher watch etcd node event.
func (d *discovery) watcher(service cc.Name) {
	key := ServiceDiscoveryName(service)

	// Use serialized request so resolution still works if the target etcd
	// server is partitioned away from the quorum.
	resp, err := d.cli.Get(d.ctx, key, etcd3.WithPrefix(), etcd3.WithSerializable())
	if err != nil {
		logs.Infof("get %s key failed, err: %v", key, err)
	}

	opts := make([]etcd3.OpOption, 0)

	if err == nil {
		for _, kv := range resp.Kvs {
			d.setAddress(service, string(kv.Key), string(kv.Value))
		}
		opts = append(opts, etcd3.WithRev(resp.Header.Revision+1))
	}

	opts = append(opts, etcd3.WithPrefix(), etcd3.WithPrevKV())
	watch := d.cli.Watch(d.ctx, key, opts...)

	go func() {
		for response := range watch {
			for _, event := range response.Events {
				switch event.Type {
				case mvccpb.PUT:
					d.setAddress(service, string(event.Kv.Key), string(event.Kv.Value))
				case mvccpb.DELETE:
					d.delAddress(service, string(event.Kv.Key))
				default:
					logs.Infof("unknown event type, %d", event.Type)
					continue
				}
			}
		}
	}()
}

// setAddress set etcdResolver addressed.
func (d *discovery) setAddress(service cc.Name, key, address string) {
	d.addressesRwMux.Lock()
	defer d.addressesRwMux.Unlock()

	addresses := d.addresses[service]

	exists := false
	for idx := range addresses {
		if addresses[idx].key == key {
			exists = true
			addresses[idx].address = address
			break
		}
	}

	if !exists {
		addresses = append(addresses, serviceAddress{
			address: address,
			key:     key,
		})
	}

	logs.Infof("after set new address[%s:%s], service: %s has address: %+v", key, address, service, addresses)
	d.addresses[service] = addresses
}

// delAddress del etcdResolver addressed.
func (d *discovery) delAddress(service cc.Name, key string) {
	d.addressesRwMux.Lock()
	defer d.addressesRwMux.Unlock()

	addresses := d.addresses[service]
	for i := 0; i < len(addresses); i++ {
		if addresses[i].key == key {
			addresses = append(addresses[:i], addresses[i+1:]...)
		}
	}

	logs.Infof("after delete address[%s], service: %s has address: %+v", key, service, addresses)
	d.addresses[service] = addresses
}
