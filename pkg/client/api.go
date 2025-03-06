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

// Package client defines all server's api client.
package client

import (
	"hcm/pkg/cc"
	accountserver "hcm/pkg/client/account-server"
	authserver "hcm/pkg/client/auth-server"
	cloudserver "hcm/pkg/client/cloud-server"
	dataservice "hcm/pkg/client/data-service"
	"hcm/pkg/client/discovery"
	hcservice "hcm/pkg/client/hc-service"
	"hcm/pkg/client/healthz"
	taskserver "hcm/pkg/client/task-server"
	"hcm/pkg/rest/client"
	rdisc "hcm/pkg/rest/discovery"
	"hcm/pkg/serviced"
)

// ClientSet defines all server's api client set.
type ClientSet struct {
	version      string
	client       client.HTTPClient
	apiDiscovery map[cc.Name]*discovery.APIDiscovery
	// TODO add flow control option
}

// NewClientSet create a new empty client set.
func NewClientSet(client client.HTTPClient, discover serviced.Discover) *ClientSet {
	cs := &ClientSet{
		version:      "v1",
		client:       client,
		apiDiscovery: make(map[cc.Name]*discovery.APIDiscovery),
	}

	for _, service := range discover.Services() {
		cs.apiDiscovery[service] = discovery.NewAPIDiscovery(service, discover)
	}
	return cs
}

type discoveryLabelWrapper struct {
	*discovery.APIDiscovery
	labels []string
}

// GetServers ...
func (w *discoveryLabelWrapper) GetServers() ([]string, error) {
	return w.APIDiscovery.GetServersWithLabel(w.labels...)
}

func (cs *ClientSet) discovery(nm cc.Name, labels ...string) rdisc.Interface {
	sd, exist := cs.apiDiscovery[nm]
	if exist {
		if labels != nil {
			return &discoveryLabelWrapper{
				APIDiscovery: sd,
				labels:       labels,
			}
		}
		return sd
	} else {
		return rdisc.DeniedServers(nm)
	}
}

// CloudServer get cloud-server client.
func (cs *ClientSet) CloudServer() *cloudserver.Client {
	c := &client.Capability{
		Client:   cs.client,
		Discover: cs.discovery(cc.CloudServerName),
	}
	return cloudserver.NewClient(c, cs.version)
}

// DataService get data-service client.
func (cs *ClientSet) DataService() *dataservice.Client {
	c := &client.Capability{
		Client:   cs.client,
		Discover: cs.discovery(cc.DataServiceName),
	}

	return dataservice.NewClient(c, cs.version)
}

// HCService get hc-service client.
func (cs *ClientSet) HCService(labels ...string) *hcservice.Client {
	c := &client.Capability{
		Client:   cs.client,
		Discover: cs.discovery(cc.HCServiceName, labels...),
	}
	return hcservice.NewClient(c, cs.version)
}

// AuthServer get auth-server client.
func (cs *ClientSet) AuthServer() *authserver.Client {
	c := &client.Capability{
		Client:   cs.client,
		Discover: cs.discovery(cc.AuthServerName),
	}
	return authserver.NewClient(c, cs.version)
}

// TaskServer get task-server client.
func (cs *ClientSet) TaskServer(labels ...string) *taskserver.Client {
	c := &client.Capability{
		Client:   cs.client,
		Discover: cs.discovery(cc.TaskServerName),
	}
	return taskserver.NewClient(c, cs.version)
}

// AccountServer get task-server client.
func (cs *ClientSet) AccountServer() *accountserver.Client {
	c := &client.Capability{
		Client:   cs.client,
		Discover: cs.discovery(cc.AccountServerName),
	}
	return accountserver.NewClient(c, cs.version)
}

// Healthz get service health check client.
func (cs *ClientSet) Healthz(service cc.Name) *healthz.Client {
	c := &client.Capability{
		Client:   cs.client,
		Discover: cs.apiDiscovery[service],
	}
	return healthz.NewClient(c)
}
