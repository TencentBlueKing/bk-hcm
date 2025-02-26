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

// Package discovery defines server discovery operation.
package discovery

import (
	"fmt"
	"sync"

	"hcm/pkg/cc"
	"hcm/pkg/serviced"
)

// APIDiscovery is service discovery for api call.
type APIDiscovery struct {
	discover serviced.Discover
	service  cc.Name
	index    int
	sync.Mutex
}

// NewAPIDiscovery create a new service discovery for api call.
func NewAPIDiscovery(service cc.Name, discover serviced.Discover) *APIDiscovery {
	return &APIDiscovery{
		discover: discover,
		service:  service,
		index:    0,
		Mutex:    sync.Mutex{},
	}
}

// GetServers get esb server host.
func (d *APIDiscovery) GetServers() ([]string, error) {
	return d.GetServersWithLabel()
}

// GetServersWithLabel get esb server host.
func (d *APIDiscovery) GetServersWithLabel(labels ...string) ([]string, error) {
	d.Lock()
	defer d.Unlock()

	servers, err := d.discover.ByLabels(labels).Discover(d.service)
	if err != nil {
		return nil, err
	}

	num := len(servers)
	if num == 0 {
		return []string{}, fmt.Errorf("there is no server can be used for %s with label %s", d.service, labels)
	}

	// move the servers in a round-robin way, e.g. last servers are [A, B, C], next server sequence is [B, C, A].
	if d.index < num-1 {
		d.index = d.index + 1
		return append(servers[d.index-1:], servers[:d.index-1]...), nil
	}

	// all servers are traversed, start from the beginning again.
	d.index = 0
	return append(servers[num-1:], servers[:num-1]...), nil
}
