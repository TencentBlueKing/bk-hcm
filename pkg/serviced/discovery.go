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
	"encoding/json"
	"fmt"
	"strings"
	"sync"

	"hcm/pkg/cc"
	"hcm/pkg/logs"
	cvt "hcm/pkg/tools/converter"

	"go.etcd.io/etcd/api/v3/mvccpb"
	etcd3 "go.etcd.io/etcd/client/v3"
)

// ServerInfo server node info
type ServerInfo struct {
	// url endpoint of the server
	URL string `json:"url"`
	// labels of the server, will be used to match services, if not any labels match, use all others server
	Labels []string `json:"labels"`
	// DisableElection disable election, if true, will not election leader, always be follower
	DisableElection bool `json:"disable_election"`
	// Env filter of current server, only servers with given env will be used
	Env string `json:"env"`
}

func (s *ServerInfo) String() string {
	if s == nil {
		return "[nil]"
	}
	builder := strings.Builder{}
	builder.WriteString(s.URL)
	if len(s.Labels) != 0 || s.Env != "" || s.DisableElection {
		builder.WriteString("(")
		if len(s.Labels) != 0 {
			builder.WriteString("labels:")
			builder.WriteString(s.Labels[0])
			for _, label := range s.Labels[1:] {
				builder.WriteString(",")
				builder.WriteString(label)
			}
			builder.WriteString(";")
		}
		if s.Env != "" {
			builder.WriteString("env:")
			builder.WriteString(s.Env)
			builder.WriteString(";")
		}
		if s.DisableElection {
			builder.WriteString("disable_election")
		}
		builder.WriteString(")")
	}

	return builder.String()
}

type discoveryLabelWrapper struct {
	*discovery
	labels []string
}

// Discover ...
func (d *discoveryLabelWrapper) Discover(name cc.Name) ([]string, error) {
	return d.discoverByLabels(name, d.labels)
}

// ByLabels reset labels
func (d *discoveryLabelWrapper) ByLabels(labels []string) Discover {
	if labels != nil {
		return &discoveryLabelWrapper{discovery: d.discovery, labels: labels}
	}
	return d
}

// Discover defines service discovery related operations.
type Discover interface {
	Discover(cc.Name) ([]string, error)
	Services() []cc.Name
	// ByLabels return a discovery wrapper filtered by the given labels.
	// if labels is nil, return the original discovery instance (no modified)
	// if labels is not nil, the discovery labels will be overridden by given labels.
	ByLabels([]string) Discover
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

// ByLabels return a discovery wrapper by labels.
func (d *discovery) ByLabels(labels []string) Discover {
	if labels != nil {
		return &discoveryLabelWrapper{discovery: d, labels: labels}
	}
	return d
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
	return d.discoverByLabels(name, nil)
}

// Discover service addresses by service name and labels
func (d *discovery) discoverByLabels(name cc.Name, labels []string) ([]string, error) {
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

outside:
	for _, addressInfo := range addressInfos {
		// skip address with different env
		if addressInfo.env != d.discOpt.Env {
			continue
		}

		// if request has no label want, skip labeled service, avoid normal request match labeled service
		if len(labels) == 0 && len(addressInfo.labels) > 0 {
			continue
		}
		if len(labels) > len(addressInfo.labels) {
			continue
		}
		if len(labels) > 0 {
			for i := range labels {
				if _, ok := addressInfo.labels[labels[i]]; !ok {
					// skip address with any want label not match
					continue outside
				}
			}
		}

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
			serverInfo := new(ServerInfo)
			if err := json.Unmarshal(kv.Value, serverInfo); err != nil {
				logs.Errorf("unmarshal server info failed, service: %s, etcd key: %s, err: %v", service, kv.Key, err)
				continue
			}
			d.setAddress(service, string(kv.Key), serverInfo)
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
					serverInfo := new(ServerInfo)
					if err := json.Unmarshal(event.Kv.Value, serverInfo); err != nil {
						logs.Errorf("unmarshal server info on put envent failed, service: %s, etcd key: %s, err: %v",
							service, event.Kv, err)
						continue
					}
					d.setAddress(service, string(event.Kv.Key), serverInfo)
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
func (d *discovery) setAddress(service cc.Name, key string, server *ServerInfo) {
	d.addressesRwMux.Lock()
	defer d.addressesRwMux.Unlock()

	addresses := d.addresses[service]

	exists := false
	for idx := range addresses {
		if addresses[idx].key == key {
			exists = true
			addresses[idx].address = server.URL
			addresses[idx].labels = cvt.StringSliceToMap(server.Labels)
			addresses[idx].env = server.Env
			addresses[idx].disableElection = server.DisableElection
			break
		}
	}

	if !exists {
		addresses = append(addresses, serviceAddress{
			address:         server.URL,
			key:             key,
			labels:          cvt.StringSliceToMap(server.Labels),
			disableElection: server.DisableElection,
			env:             server.Env,
		})
	}

	logs.Infof("after set new address[%s:%s], service: %s has address: %+v", key, server.String(), service, addresses)
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
