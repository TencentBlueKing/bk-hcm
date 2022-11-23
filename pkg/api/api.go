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

// Package api defines all server's api client.
package api

import (
	"hcm/pkg/api/cloud-server"
	dataservice "hcm/pkg/api/data-service"
	"hcm/pkg/api/discovery"
	"hcm/pkg/api/healthz"
	"hcm/pkg/cc"
	"hcm/pkg/rest/client"
	"hcm/pkg/serviced"
)

// ClientSet defines all server's api client set.
type ClientSet struct {
	version      string
	client       client.HTTPClient
	apiDiscovery map[cc.Name]*discovery.APIDiscovery
	// TODO add flow control option
}

// newClientSet create a new empty client set.
func newClientSet(client client.HTTPClient, discover serviced.Discover, discoverServices []cc.Name) *ClientSet {
	cs := &ClientSet{
		version:      "v1",
		client:       client,
		apiDiscovery: make(map[cc.Name]*discovery.APIDiscovery),
	}

	for _, service := range discoverServices {
		cs.apiDiscovery[service] = discovery.NewAPIDiscovery(service, discover)
	}
	return cs
}

// NewAPIServerClientSet create a new api-server used client set.
func NewAPIServerClientSet(client client.HTTPClient, discover serviced.Discover) *ClientSet {
	discoverServices := []cc.Name{cc.CloudServerName}
	return newClientSet(client, discover, discoverServices)
}

// CloudServer get cloud-server client.
func (cs *ClientSet) CloudServer() *cloudserver.Client {
	c := &client.Capability{
		Client:   cs.client,
		Discover: cs.apiDiscovery[cc.CloudServerName],
	}
	return cloudserver.NewClient(c, cs.version)
}

// DataService get data-service client.
func (cs *ClientSet) DataService() *dataservice.Client {
	c := &client.Capability{
		Client:   cs.client,
		Discover: cs.apiDiscovery[cc.DataServiceName],
	}
	return dataservice.NewClient(c, cs.version)
}

// Healthz get service health check client.
func (cs *ClientSet) Healthz(service cc.Name) *healthz.Client {
	c := &client.Capability{
		Client:   cs.client,
		Discover: cs.apiDiscovery[service],
	}
	return healthz.NewClient(c)
}
