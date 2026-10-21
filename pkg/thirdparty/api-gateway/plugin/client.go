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

package plugin

import (
	"errors"
	"fmt"
	"net/url"
	"strings"

	"hcm/pkg/cc"
	"hcm/pkg/kit"
	"hcm/pkg/rest"
	"hcm/pkg/rest/client"
	apigateway "hcm/pkg/thirdparty/api-gateway"
	"hcm/pkg/tools/ssl"

	"github.com/prometheus/client_golang/prometheus"
)

// Client defines the southbound HTTP client interface for HCM plugins.
type Client interface {
	Health(kt *kit.Kit) (*HealthData, error)
	GetSchema(kt *kit.Kit) (*PluginSchema, error)
	Apply(kt *kit.Kit, req *ApplyReq) (*ApplyData, error)
	GetResult(kt *kit.Kit, req *GetResultReq) (*ApplyData, error)
	Stop(kt *kit.Kit, req *StopReq) (*ApplyData, error)
	Plan(kt *kit.Kit, req *PlanReq) (*PlanData, error)
	Validate(kt *kit.Kit, req *ValidateReq) (*ValidateData, error)
	Import(kt *kit.Kit, req *ImportReq) (*StateData, error)
	Refresh(kt *kit.Kit, req *RefreshReq) (*StateData, error)
	ReadData(kt *kit.Kit, req *ReadDataReq) (*StateData, error)
	List(kt *kit.Kit, req *ListReq) (*ListData, error)
}

type pluginApiGateway struct {
	config *cc.ApiGateway
	client rest.ClientInterface
}

var _ Client = (*pluginApiGateway)(nil)

// NewClient initializes a new southbound HTTP plugin client for the given gateway baseURL.
func NewClient(baseURL string, cfg *cc.ApiGateway, reg prometheus.Registerer) (Client, error) {
	if cfg == nil {
		return nil, errors.New("api gateway config is nil")
	}

	serverHost, basePath, err := parseBaseURL(baseURL)
	if err != nil {
		return nil, err
	}

	tls := &ssl.TLSConfig{
		InsecureSkipVerify: cfg.TLS.InsecureSkipVerify,
		CertFile:           cfg.TLS.CertFile,
		KeyFile:            cfg.TLS.KeyFile,
		CAFile:             cfg.TLS.CAFile,
		Password:           cfg.TLS.Password,
	}
	cli, err := client.NewClient(tls)
	if err != nil {
		return nil, err
	}

	c := &client.Capability{
		Client: cli,
		Discover: &apigateway.Discovery{
			Name:    "pluginApiGateWay",
			Servers: []string{serverHost},
		},
		MetricOpts: client.MetricOption{Register: reg},
	}
	restCli := rest.NewClient(c, basePath)

	return &pluginApiGateway{
		config: cfg,
		client: restCli,
	}, nil
}

func parseBaseURL(rawURL string) (string, string, error) {
	trimmed := strings.TrimSpace(rawURL)
	if trimmed == "" {
		return "", "", errors.New("base_url is empty")
	}

	u, err := url.Parse(trimmed)
	if err != nil {
		return "", "", fmt.Errorf("parse base_url failed, err: %v", err)
	}
	if (u.Scheme != "http" && u.Scheme != "https") || u.Host == "" {
		return "", "", fmt.Errorf("invalid base_url %q: scheme must be http/https and host is required", rawURL)
	}

	basePath := u.Path
	if basePath == "" {
		basePath = "/"
	}
	return fmt.Sprintf("%s://%s", u.Scheme, u.Host), basePath, nil
}

// Health calls GET /health to check plugin liveness and get plugin_version.
func (p *pluginApiGateway) Health(kt *kit.Kit) (*HealthData, error) {
	return apigateway.ApiGatewayCallWithoutReq[HealthData](p.client, p.config, rest.GET, kt, nil, "/health")
}

// GetSchema calls GET /schema to retrieve the complete plugin schema and capabilities.
func (p *pluginApiGateway) GetSchema(kt *kit.Kit) (*PluginSchema, error) {
	return apigateway.ApiGatewayCallWithoutReq[PluginSchema](p.client, p.config, rest.GET, kt, nil, "/schema")
}

// Apply calls POST /apply to create, update, or delete a resource in the external system.
func (p *pluginApiGateway) Apply(kt *kit.Kit, req *ApplyReq) (*ApplyData, error) {
	if err := req.Validate(); err != nil {
		return nil, err
	}
	return apigateway.ApiGatewayCall[ApplyReq, ApplyData](p.client, p.config, rest.POST, kt, req, "/apply")
}

// GetResult calls POST /result to query the current execution result of an async Apply.
func (p *pluginApiGateway) GetResult(kt *kit.Kit, req *GetResultReq) (*ApplyData, error) {
	if err := req.Validate(); err != nil {
		return nil, err
	}
	return apigateway.ApiGatewayCall[GetResultReq, ApplyData](p.client, p.config, rest.POST, kt, req, "/result")
}

// Stop calls POST /stop to request stopping an in-progress async Apply.
func (p *pluginApiGateway) Stop(kt *kit.Kit, req *StopReq) (*ApplyData, error) {
	if err := req.Validate(); err != nil {
		return nil, err
	}
	return apigateway.ApiGatewayCall[StopReq, ApplyData](p.client, p.config, rest.POST, kt, req, "/stop")
}

// Plan calls POST /plan to preview resource changes without modifying the external system.
func (p *pluginApiGateway) Plan(kt *kit.Kit, req *PlanReq) (*PlanData, error) {
	if err := req.Validate(); err != nil {
		return nil, err
	}
	return apigateway.ApiGatewayCall[PlanReq, PlanData](p.client, p.config, rest.POST, kt, req, "/plan")
}

// Validate calls POST /validate to validate user configuration against plugin business rules.
func (p *pluginApiGateway) Validate(kt *kit.Kit, req *ValidateReq) (*ValidateData, error) {
	if err := req.Validate(); err != nil {
		return nil, err
	}
	return apigateway.ApiGatewayCall[ValidateReq, ValidateData](p.client, p.config, rest.POST, kt, req, "/validate")
}

// Import calls POST /import to read the complete state of an existing external resource by import_id.
func (p *pluginApiGateway) Import(kt *kit.Kit, req *ImportReq) (*StateData, error) {
	if err := req.Validate(); err != nil {
		return nil, err
	}
	return apigateway.ApiGatewayCall[ImportReq, StateData](p.client, p.config, rest.POST, kt, req, "/import")
}

// Refresh calls POST /refresh to refresh the latest external state using the saved state.
func (p *pluginApiGateway) Refresh(kt *kit.Kit, req *RefreshReq) (*StateData, error) {
	if err := req.Validate(); err != nil {
		return nil, err
	}
	return apigateway.ApiGatewayCall[RefreshReq, StateData](p.client, p.config, rest.POST, kt, req, "/refresh")
}

// ReadData calls POST /read_data to query read-only data source state by filter conditions.
func (p *pluginApiGateway) ReadData(kt *kit.Kit, req *ReadDataReq) (*StateData, error) {
	if err := req.Validate(); err != nil {
		return nil, err
	}
	return apigateway.ApiGatewayCall[ReadDataReq, StateData](p.client, p.config, rest.POST, kt, req, "/read_data")
}

// List calls POST /list to paginate all resource items under the current plugin instance scope.
func (p *pluginApiGateway) List(kt *kit.Kit, req *ListReq) (*ListData, error) {
	if err := req.Validate(); err != nil {
		return nil, err
	}
	return apigateway.ApiGatewayCall[ListReq, ListData](p.client, p.config, rest.POST, kt, req, "/list")
}
