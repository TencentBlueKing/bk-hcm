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

package cmdb

import (
	"hcm/pkg/cc"
	"hcm/pkg/kit"
	"hcm/pkg/rest"
	"hcm/pkg/rest/client"
	apigateway "hcm/pkg/thirdparty/api-gateway"
	"hcm/pkg/tools/ssl"

	"github.com/prometheus/client_golang/prometheus"
)

// Client is an api-gateway client to request cmdb.
type Client interface {
	SearchBusiness(kt *kit.Kit, params *SearchBizParams) (*SearchBizResult, error)
	SearchCloudArea(kt *kit.Kit, params *SearchCloudAreaParams) (*SearchCloudAreaResult, error)
	ListBizHost(kt *kit.Kit, params *ListBizHostParams) (*ListBizHostResult, error)
	SearchModule(kt *kit.Kit, params *SearchModuleParams) (*ModuleInfoResult, error)
	FindHostTopoRelation(kt *kit.Kit, params *FindHostTopoRelationParams) (*HostTopoRelationResult, error)
	FindHostBizRelations(kt *kit.Kit, params *HostModuleRelationParams) (*[]HostTopoRelation, error)
	ResourceWatch(kt *kit.Kit, params *WatchEventParams) (*WatchEventResult, error)
	GetBizBriefCacheTopo(kt *kit.Kit, params *GetBizBriefCacheTopoParams) (*GetBizBriefCacheTopoResult, error)
	DeleteCloudHostFromBiz(kt *kit.Kit, params *DeleteCloudHostFromBizParams) error
	AddCloudHostToBiz(kt *kit.Kit, params *AddCloudHostToBizParams) (*BatchCreateResult, error)
	ListHostWithoutBiz(kt *kit.Kit, req *ListHostWithoutBizParams) (*ListHostWithoutBizResult, error)
}

// NewClient initialize a new cmdbApiGateWay client
func NewClient(cfg *cc.ApiGateway, reg prometheus.Registerer) (Client, error) {
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
			Name:    "cmdbApiGateWay",
			Servers: cfg.Endpoints,
		},
		MetricOpts: client.MetricOption{Register: reg},
	}
	restCli := rest.NewClient(c, "/api/v3")

	agw := &cmdbApiGateWay{
		config: cfg,
		client: restCli,
	}
	return agw, nil
}

var _ Client = (*cmdbApiGateWay)(nil)

// cmdbApiGateWay is an esb client to request cmdbApiGateWay.
type cmdbApiGateWay struct {
	config *cc.ApiGateway
	// http client instance
	client rest.ClientInterface
}

// DeleteCloudHostFromBiz ...
func (c *cmdbApiGateWay) DeleteCloudHostFromBiz(kt *kit.Kit, params *DeleteCloudHostFromBizParams) error {
	err := params.Validate()
	if err != nil {
		return err
	}
	_, err = apigateway.ApiGatewayCall[DeleteCloudHostFromBizParams, interface{}](c.client, c.config,
		rest.DELETE, kt, params, "/deletemany/cloud_hosts")
	if err != nil {
		return err
	}
	return nil
}

// AddCloudHostToBiz ...
func (c *cmdbApiGateWay) AddCloudHostToBiz(kt *kit.Kit, params *AddCloudHostToBizParams) (*BatchCreateResult, error) {
	err := params.Validate()
	if err != nil {
		return nil, err
	}
	return apigateway.ApiGatewayCall[AddCloudHostToBizParams, BatchCreateResult](c.client, c.config,
		rest.POST, kt, params, "/createmany/cloud_hosts")
}

// GetBizBriefCacheTopo 根据业务ID,查询该业务的全量简明拓扑树信息。
// 该业务拓扑的全量信息，包含了从业务这个根节点开始，到自定义层级实例(如果主线的拓扑层级中包含)，到集群、模块等中间的所有拓扑层级树数据。
func (c *cmdbApiGateWay) GetBizBriefCacheTopo(kt *kit.Kit, params *GetBizBriefCacheTopoParams) (
	*GetBizBriefCacheTopoResult, error) {

	err := params.Validate()
	if err != nil {
		return nil, err
	}
	return apigateway.ApiGatewayCall[GetBizBriefCacheTopoParams, GetBizBriefCacheTopoResult](c.client, c.config,
		rest.GET, kt, params, "/cache/find/cache/topo/brief/biz/%d", params.BkBizID)
}

// ResourceWatch ...
func (c *cmdbApiGateWay) ResourceWatch(kt *kit.Kit, params *WatchEventParams) (*WatchEventResult, error) {
	err := params.Validate()
	if err != nil {
		return nil, err
	}
	return apigateway.ApiGatewayCall[WatchEventParams, WatchEventResult](c.client, c.config, rest.POST, kt, params,
		"/event/watch/resource/%s", params.Resource)
}

// FindHostBizRelations ...
func (c *cmdbApiGateWay) FindHostBizRelations(kt *kit.Kit, params *HostModuleRelationParams) (*[]HostTopoRelation, error) {
	err := params.Validate()
	if err != nil {
		return nil, err
	}
	return apigateway.ApiGatewayCall[HostModuleRelationParams, []HostTopoRelation](c.client, c.config, rest.POST, kt, params,
		"/hosts/modules/read")
}

// FindHostTopoRelation ...
func (c *cmdbApiGateWay) FindHostTopoRelation(kt *kit.Kit, params *FindHostTopoRelationParams) (
	*HostTopoRelationResult, error) {

	if err := params.Validate(); err != nil {
		return nil, err
	}

	return apigateway.ApiGatewayCall[FindHostTopoRelationParams, HostTopoRelationResult](c.client, c.config, rest.POST,
		kt, params, "/host/topo/relation/read")
}

// SearchModule ...
func (c *cmdbApiGateWay) SearchModule(kt *kit.Kit, params *SearchModuleParams) (*ModuleInfoResult, error) {
	err := params.Validate()
	if err != nil {
		return nil, err
	}

	// 0 代表了bk_supplier_account
	return apigateway.ApiGatewayCall[SearchModuleParams, ModuleInfoResult](c.client, c.config, rest.POST, kt, params,
		"/module/search/0/%d/%d", params.BizID, params.BkSetID)
}

// ListBizHost ...
func (c *cmdbApiGateWay) ListBizHost(kt *kit.Kit, req *ListBizHostParams) (*ListBizHostResult, error) {
	err := req.Validate()
	if err != nil {
		return nil, err
	}
	return apigateway.ApiGatewayCall[ListBizHostParams, ListBizHostResult](c.client, c.config, rest.POST, kt, req,
		"/hosts/app/%d/list_hosts", req.BizID)
}

// SearchBusiness ...
func (c *cmdbApiGateWay) SearchBusiness(kt *kit.Kit, req *SearchBizParams) (*SearchBizResult, error) {
	// 0 代表了bk_supplier_account
	return apigateway.ApiGatewayCall[SearchBizParams, SearchBizResult](c.client, c.config, rest.POST,
		kt, req, "/biz/search/0")
}

// SearchCloudArea search cmdb cloud area
func (c *cmdbApiGateWay) SearchCloudArea(kt *kit.Kit, params *SearchCloudAreaParams) (*SearchCloudAreaResult, error) {

	return apigateway.ApiGatewayCall[SearchCloudAreaParams, SearchCloudAreaResult](c.client, c.config, rest.POST, kt, params,
		"/findmany/cloudarea")
}

// ListHostWithoutBiz list cmdb host without biz.
func (c *cmdbApiGateWay) ListHostWithoutBiz(kt *kit.Kit, req *ListHostWithoutBizParams) (
	*ListHostWithoutBizResult, error) {

	if err := req.Validate(); err != nil {
		return nil, err
	}
	return apigateway.ApiGatewayCall[ListHostWithoutBizParams, ListHostWithoutBizResult](c.client, c.config,
		rest.POST, kt, req, "/hosts/list_hosts_without_app")
}
