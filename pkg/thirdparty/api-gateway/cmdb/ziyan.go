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
	"errors"
	"fmt"

	"hcm/pkg"
	"hcm/pkg/kit"
	"hcm/pkg/logs"
	"hcm/pkg/rest"
	apigateway "hcm/pkg/thirdparty/api-gateway"
	"hcm/pkg/tools/converter"
	"hcm/pkg/tools/querybuilder"
	"hcm/pkg/tools/slice"
)

// ZiyanCmdbClient 内部版使用的cmdb接口
type ZiyanCmdbClient interface {
	UpdateCvmOSAndSvrStatus(kt *kit.Kit, req *UpdateCvmOSReq) error

	// SearchBizCompanyCmdbInfo 返回cc业务在公司cmdb的信息
	SearchBizCompanyCmdbInfo(kt *kit.Kit, params *SearchBizCompanyCmdbInfoParams) (*[]CompanyCmdbInfo, error)
	GetHostBizIds(kt *kit.Kit, hostIds []int64) (map[int64]int64, error)
	// GetBizInternalModule get business's internal module
	GetBizInternalModule(kt *kit.Kit, req *GetBizInternalModuleReq) (*BizInternalModuleRespRst, error)
	// GetBizRecycleModuleID get business's recycle module id
	GetBizRecycleModuleID(kt *kit.Kit, bizID int64) (int64, error)

	// ListHost same as ListHostWithoutBiz
	ListHost(kt *kit.Kit, req *ListHostReq) (*ListHostResult, error)
	// 下面为 ListHost封装好查询条件的接口

	// GetHostId gets host id by ip in cc 3.0
	GetHostId(kt *kit.Kit, ip string) (int64, error)
	// GetHostInfoByIP get host info by ip in CMDB
	GetHostInfoByIP(kt *kit.Kit, ip string, bkCloudID int) (*Host, error)
	// GetHostInfoByHostID get host info by host id in CMDB
	GetHostInfoByHostID(kt *kit.Kit, bkHostID int64) (*Host, error)
	// GetHostIDByAssetID gets host id by asset id in cc 3.0
	GetHostIDByAssetID(kt *kit.Kit, assetID string) (int64, error)
	// FindManyHostsByAssetID 通过固资号批量查询主机
	FindManyHostsByAssetID(kt *kit.Kit, assetIDs []string) ([]HostByAssetID, error)

	// write operation

	// AddHost adds host to cc 3.0, once 10 hosts at most
	AddHost(kt *kit.Kit, req *AddHostReq) error
	// Hosts2CrTransit transfer hosts to given business's CR transit module in CMDB
	Hosts2CrTransit(kt *kit.Kit, req *CrTransitReq) (*CrTransitRst, error)
	// TransferHost transfer host to another business in cc 3.0
	TransferHost(kt *kit.Kit, req *TransferHostReq) error
	// HostsCrTransit2Idle transfer hosts to given business's idle module in CMDB
	HostsCrTransit2Idle(kt *kit.Kit, req *CrTransitIdleReq) error
	// UpdateHosts update host info in cc 3.0
	UpdateHosts(kt *kit.Kit, req *UpdateHostsReq) (*[]ModuleHost, error)
}

// AddHost adds host to cc 3.0, once 10 hosts at most
func (c *cmdbApiGateWay) AddHost(kt *kit.Kit, req *AddHostReq) error {

	if err := req.Validate(); err != nil {
		return err
	}
	_, err := apigateway.ApiGatewayCall[AddHostReq, interface{}](c.client, c.config,
		rest.POST, kt, req, "/shipper/sync/cmdb/add_host_from_cmpy")
	if err != nil {
		return err
	}
	return nil
}

// SearchBizCompanyCmdbInfo 查询cc 业务和公司cmdb运营产品、一二级业务的关系
func (c *cmdbApiGateWay) SearchBizCompanyCmdbInfo(kt *kit.Kit, params *SearchBizCompanyCmdbInfoParams) (
	*[]CompanyCmdbInfo, error) {

	if err := params.Validate(); err != nil {
		return nil, err
	}
	return apigateway.ApiGatewayCall[SearchBizCompanyCmdbInfoParams, []CompanyCmdbInfo](c.client, c.config,
		rest.POST, kt, params, "/sidecar/findmany/business/cost_info_relation")
}

// ListHost gets hosts info in cc 3.0, limit 500
func (c *cmdbApiGateWay) ListHost(kt *kit.Kit, req *ListHostReq) (*ListHostResult, error) {

	if err := req.Validate(); err != nil {
		return nil, err
	}
	return apigateway.ApiGatewayCall[ListHostReq, ListHostResult](c.client, c.config,
		rest.POST, kt, req, "/hosts/list_hosts_without_app")
}

// GetHostBizIds 实际调用的是 FindHostBizRelations 方法, 返回值为 map[hostID]bkBizID
func (c *cmdbApiGateWay) GetHostBizIds(kt *kit.Kit, hostIds []int64) (map[int64]int64, error) {

	result := make(map[int64]int64, len(hostIds))
	for _, ids := range slice.Split(hostIds, pkg.BKMaxInstanceLimit) {
		req := &HostModuleRelationParams{HostID: ids}

		relations, err := c.FindHostBizRelations(kt, req)
		if err != nil {
			logs.Errorf("find host biz relations failed, err: %v, rid: %s", err, kt.Rid)
			return nil, err
		}
		for _, relation := range converter.PtrToVal(relations) {
			result[relation.HostID] = relation.BizID
		}
	}
	return result, nil
}

// Hosts2CrTransit transfer hosts to given business's CR transit module in CMDB
func (c *cmdbApiGateWay) Hosts2CrTransit(kt *kit.Kit, req *CrTransitReq) (*CrTransitRst, error) {
	if err := req.Validate(); err != nil {
		return nil, err
	}
	return apigateway.ApiGatewayCall[CrTransitReq, CrTransitRst](c.client, c.config,
		rest.POST, kt, req, "/shipper/transfer/cmdb/hosts_to_cr_transit")
}

// TransferHost transfer host to another business in cc 3.0
func (c *cmdbApiGateWay) TransferHost(kt *kit.Kit, req *TransferHostReq) error {
	err := req.Validate()
	if err != nil {
		return err
	}
	_, err = apigateway.ApiGatewayCall[TransferHostReq, interface{}](c.client, c.config,
		rest.POST, kt, req, "/sidecar/host/transfer_host_to_another_biz")
	if err != nil {
		return err
	}
	return nil
}

// HostsCrTransit2Idle transfer hosts to given business's idle module in CMDB
func (c *cmdbApiGateWay) HostsCrTransit2Idle(kt *kit.Kit, req *CrTransitIdleReq) error {

	err := req.Validate()
	if err != nil {
		return err
	}
	_, err = apigateway.ApiGatewayCall[CrTransitIdleReq, interface{}](c.client, c.config,
		rest.POST, kt, req, "/shipper/transfer/cmdb/hosts_cr_transit_to_idle")
	if err != nil {
		return err
	}
	return nil
}

// GetHostId gets host id by ip in cc 3.0
func (c *cmdbApiGateWay) GetHostId(kt *kit.Kit, ip string) (int64, error) {

	req := &ListHostReq{
		HostPropertyFilter: &QueryFilter{
			Rule: querybuilder.CombinedRule{
				Condition: querybuilder.ConditionAnd,
				Rules: []querybuilder.Rule{
					querybuilder.AtomRule{
						Field:    pkg.BKHostInnerIPField,
						Operator: querybuilder.OperatorEqual,
						Value:    ip,
					},
				},
			},
		},
		Fields: []string{
			pkg.BKHostIDField,
		},
		Page: BasePage{
			Start: 0,
			Limit: 1,
		},
	}

	resp, err := c.ListHost(kt, req)
	if err != nil {
		logs.Errorf("failed to get host id by ip, ip: %s, err: %v", ip, err)
		return 0, err
	}
	if len(resp.Info) != 1 {
		return 0, fmt.Errorf("failed to get host id by ip, ip: %s, for return data size %d not equal 1", ip,
			len(resp.Info))
	}

	return resp.Info[0].BkHostID, nil
}

// GetHostInfoByIP get host info by ip in cc 3.0(bkCloudID是管控区ID)
func (c *cmdbApiGateWay) GetHostInfoByIP(kt *kit.Kit, ip string, bkCloudID int) (*Host, error) {

	req := &ListHostReq{
		HostPropertyFilter: &QueryFilter{
			Rule: CombinedRule{
				Condition: ConditionAnd,
				Rules: []Rule{
					AtomRule{
						Field:    pkg.BKHostInnerIPField,
						Operator: OperatorEqual,
						Value:    ip,
					},
					AtomRule{
						Field:    pkg.BKCloudIDField,
						Operator: OperatorEqual,
						Value:    bkCloudID,
					},
				},
			},
		},
		Page: BasePage{Start: 0, Limit: 1},
	}

	resp, err := c.ListHost(kt, req)
	if err != nil {
		logs.Errorf("failed to get host id by ip, ip: %s, err: %v", ip, err)
		return nil, err
	}
	if len(resp.Info) != 1 {
		return nil, fmt.Errorf("failed to get host id by ip, ip: %s, for return data size %d not equal 1", ip,
			len(resp.Info))
	}

	return &resp.Info[0], nil
}

// GetHostInfoByHostID get hosts info by host id in cc 3.0
func (c *cmdbApiGateWay) GetHostInfoByHostID(kt *kit.Kit, bkHostID int64) (*Host, error) {

	req := &ListHostReq{
		HostPropertyFilter: &QueryFilter{
			Rule: querybuilder.CombinedRule{
				Condition: querybuilder.ConditionAnd,
				Rules: []querybuilder.Rule{
					querybuilder.AtomRule{
						Field:    pkg.BKHostIDField,
						Operator: querybuilder.OperatorEqual,
						Value:    bkHostID,
					},
				},
			},
		},
		Page: BasePage{Start: 0, Limit: 1},
	}

	resp, err := c.ListHost(kt, req)
	if err != nil {
		logs.Errorf("failed to get host id by hostID, hostID: %d, err: %v", bkHostID, err)
		return nil, err
	}
	if len(resp.Info) != 1 {
		return nil, fmt.Errorf("failed to get host id by hostID, hostID: %d, for return data size %d not equal 1",
			bkHostID, len(resp.Info))
	}

	return &resp.Info[0], nil
}

// UpdateHosts update host info in cc 3.0
func (c *cmdbApiGateWay) UpdateHosts(kt *kit.Kit, req *UpdateHostsReq) (*[]ModuleHost, error) {
	err := req.Validate()
	if err != nil {
		return nil, err
	}
	return apigateway.ApiGatewayCall[UpdateHostsReq, []ModuleHost](c.client, c.config,
		rest.PUT, kt, req, "/hosts/property/batch")
}

// GetBizInternalModule get business's internal module
func (c *cmdbApiGateWay) GetBizInternalModule(kt *kit.Kit, req *GetBizInternalModuleReq) (*BizInternalModuleRespRst,
	error) {

	err := req.Validate()
	if err != nil {
		return nil, err
	}
	// url: /topo/internal/{bk_supplier_account}/{bk_biz_id}
	// 内部版 bk_supplier_account 指定为 tencent
	return apigateway.ApiGatewayCall[GetBizInternalModuleReq, BizInternalModuleRespRst](c.client, c.config,
		rest.GET, kt, req, "/topo/internal/tencent/%d", req.BkBizID)
}

// GetBizRecycleModuleID get business's recycle module ID
func (c *cmdbApiGateWay) GetBizRecycleModuleID(kt *kit.Kit, bizID int64) (int64, error) {
	req := &GetBizInternalModuleReq{
		BkBizID: bizID,
	}
	resp, err := c.GetBizInternalModule(kt, req)
	if err != nil {
		return 0, err
	}

	moduleID := int64(0)
	for _, module := range resp.Module {
		if module.Default == DftModuleRecycle {
			moduleID = module.BkModuleId
			break
		}
	}
	if moduleID <= 0 {
		return 0, errors.New("get no biz recycle module ID")
	}

	return moduleID, nil
}

// GetHostIDByAssetID gets host id by ip in cc 3.0
func (c *cmdbApiGateWay) GetHostIDByAssetID(kt *kit.Kit, assetID string) (int64, error) {

	req := &ListHostReq{
		HostPropertyFilter: &QueryFilter{
			Rule: querybuilder.CombinedRule{
				Condition: querybuilder.ConditionAnd,
				Rules: []querybuilder.Rule{
					querybuilder.AtomRule{
						Field:    pkg.BKAssetIDField,
						Operator: querybuilder.OperatorEqual,
						Value:    assetID,
					},
				},
			},
		},
		Fields: []string{
			pkg.BKHostIDField,
		},
		Page: BasePage{
			Start: 0,
			Limit: 1,
		},
	}

	result, err := c.ListHost(kt, req)
	if err != nil {
		return -1, err
	}
	if len(result.Info) != 1 {
		return -1, fmt.Errorf("failed to get host id by assetID, assetID: %s, for return data size %d not equal 1",
			assetID, len(result.Info))
	}
	return result.Info[0].BkHostID, nil
}

// FindManyHostsByAssetID gets parent machine IP by asset id from bkcc
func (c *cmdbApiGateWay) FindManyHostsByAssetID(kt *kit.Kit, assetIDs []string) ([]HostByAssetID, error) {
	req := &FindManyHostsByAssetIDReq{BkAssetIDs: assetIDs, Fields: []string{"bk_host_innerip"}}
	if err := req.Validate(); err != nil {
		return nil, err
	}
	resp, err := apigateway.ApiGatewayCall[FindManyHostsByAssetIDReq, []HostByAssetID](c.client, c.config,
		rest.POST, kt, req, "/shipper/findmany/cmdb/hosts/by_asset_id")
	if err != nil {
		logs.Errorf("failed to find hosts by asset id from bkcc,assetIDs: %v, err: %v,rid:%s", assetIDs, err, kt.Rid)
		return nil, err
	}
	if resp == nil {
		return []HostByAssetID{}, nil
	}
	return *resp, nil
}
