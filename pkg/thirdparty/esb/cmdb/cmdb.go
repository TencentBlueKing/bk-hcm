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
	"hcm/pkg/thirdparty/esb/types"
)

// Client is an esb client to request cmdb.
type Client interface {
	SearchBusiness(kt *kit.Kit, params *SearchBizParams) (*SearchBizResult, error)
	SearchCloudArea(kt *kit.Kit, params *SearchCloudAreaParams) (*SearchCloudAreaResult, error)
	AddCloudHostToBiz(kt *kit.Kit, params *AddCloudHostToBizParams) (*BatchCreateResult, error)
	DeleteCloudHostFromBiz(kt *kit.Kit, params *DeleteCloudHostFromBizParams) error
	ListBizHost(kt *kit.Kit, params *ListBizHostParams) (*ListBizHostResult, error)
	GetBizBriefCacheTopo(kt *kit.Kit, params *GetBizBriefCacheTopoParams) (*GetBizBriefCacheTopoResult, error)
	FindHostTopoRelation(kt *kit.Kit, params *FindHostTopoRelationParams) (*HostTopoRelationResult, error)
	SearchModule(kt *kit.Kit, params *SearchModuleParams) (*ModuleInfoResult, error)
}

// NewClient initialize a new cmdb client
func NewClient(client rest.ClientInterface, config *cc.Esb) Client {
	return &cmdb{
		client: client,
		config: config,
	}
}

var _ Client = new(cmdb)

// cmdb is an esb client to request cmdb.
type cmdb struct {
	config *cc.Esb
	// http client instance
	client rest.ClientInterface
}

// SearchBusiness search business
func (c *cmdb) SearchBusiness(kt *kit.Kit, params *SearchBizParams) (*SearchBizResult, error) {

	return types.EsbCall[SearchBizParams, SearchBizResult](c.client, c.config, rest.POST, kt, params,
		"/cc/search_business/")
}

// SearchCloudArea search cmdb cloud area
func (c *cmdb) SearchCloudArea(kt *kit.Kit, params *SearchCloudAreaParams) (*SearchCloudAreaResult, error) {

	return types.EsbCall[SearchCloudAreaParams, SearchCloudAreaResult](c.client, c.config, rest.POST, kt, params,
		"/cc/search_cloud_area/")
}

// AddCloudHostToBiz add cmdb cloud host to biz.
func (c *cmdb) AddCloudHostToBiz(kt *kit.Kit, params *AddCloudHostToBizParams) (*BatchCreateResult, error) {

	return types.EsbCall[AddCloudHostToBizParams, BatchCreateResult](c.client, c.config, rest.POST, kt, params,
		"/cc/add_cloud_host_to_biz/")
}

// DeleteCloudHostFromBiz delete cmdb cloud host from biz.
func (c *cmdb) DeleteCloudHostFromBiz(kt *kit.Kit, params *DeleteCloudHostFromBizParams) error {
	_, err := types.EsbCall[DeleteCloudHostFromBizParams, struct{}](c.client, c.config, rest.POST, kt, params,
		"/cc/delete_cloud_host_from_biz/")
	return err
}

// ListBizHost list cmdb host in biz.
func (c *cmdb) ListBizHost(kt *kit.Kit, params *ListBizHostParams) (*ListBizHostResult, error) {

	return types.EsbCall[ListBizHostParams, ListBizHostResult](c.client, c.config, rest.POST, kt, params,
		"/cc/list_biz_hosts/")
}

// FindHostTopoRelation 获取主机拓扑
func (c *cmdb) FindHostTopoRelation(kt *kit.Kit, params *FindHostTopoRelationParams) (
	*HostTopoRelationResult, error) {

	return types.EsbCall[FindHostTopoRelationParams, HostTopoRelationResult](c.client, c.config, rest.POST, kt, params,
		"/cc/find_host_topo_relation/")
}

// SearchModule 查询模块信息
func (c *cmdb) SearchModule(kt *kit.Kit, params *SearchModuleParams) (*ModuleInfoResult, error) {

	return types.EsbCall[SearchModuleParams, ModuleInfoResult](c.client, c.config, rest.POST, kt, params,
		"/cc/search_module/")
}
