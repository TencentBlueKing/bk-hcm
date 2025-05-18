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

package gcp

import (
	"fmt"

	"hcm/cmd/hc-service/logics/res-sync/common"
	adcore "hcm/pkg/adaptor/types/core"
	typesroutetable "hcm/pkg/adaptor/types/route-table"
	"hcm/pkg/api/core"
	cloudcoreroutetable "hcm/pkg/api/core/cloud/route-table"
	dataservice "hcm/pkg/api/data-service"
	routetable "hcm/pkg/api/data-service/cloud/route-table"
	"hcm/pkg/criteria/constant"
	"hcm/pkg/criteria/enumor"
	"hcm/pkg/criteria/errf"
	"hcm/pkg/criteria/validator"
	"hcm/pkg/dal/dao/tools"
	"hcm/pkg/kit"
	"hcm/pkg/logs"
	"hcm/pkg/runtime/filter"
	"hcm/pkg/tools/converter"
)

// SyncRouteOption ...
type SyncRouteOption struct {
}

// Validate ...
func (opt SyncRouteOption) Validate() error {
	return validator.Validate.Struct(opt)
}

// Route ...
func (cli *client) Route(kt *kit.Kit, params *SyncBaseParams, opt *SyncRouteOption) (*SyncResult, error) {
	if err := validator.ValidateTool(params, opt); err != nil {
		return nil, errf.NewFromErr(errf.InvalidParameter, err)
	}

	routeFromCloud, err := cli.listRouteFromCloud(kt, params)
	if err != nil {
		return nil, err
	}

	routeFromDB, err := cli.listRouteFromDB(kt, params, opt)
	if err != nil {
		return nil, err
	}

	if len(routeFromCloud) == 0 && len(routeFromDB) == 0 {
		return new(SyncResult), nil
	}

	addSlice, updateMap, delCloudIDs := common.Diff[typesroutetable.GcpRoute, cloudcoreroutetable.GcpRoute](
		routeFromCloud, routeFromDB, isRouteChange)

	if len(delCloudIDs) > 0 {
		if err := cli.deleteRoute(kt, params.AccountID, delCloudIDs, routeFromDB); err != nil {
			return nil, err
		}
	}

	if len(addSlice) > 0 {
		if err = cli.createRoute(kt, params.AccountID, addSlice); err != nil {
			return nil, err
		}
	}

	if len(updateMap) > 0 {
		if err = cli.updateRoute(kt, params.AccountID, updateMap); err != nil {
			return nil, err
		}
	}

	return new(SyncResult), nil
}

func (cli *client) createRoute(kt *kit.Kit, accountID string,
	addSlice []typesroutetable.GcpRoute) error {

	if len(addSlice) <= 0 {
		return fmt.Errorf("route addSlice is <= 0, not create")
	}

	createResources := make([]routetable.GcpRouteCreateReq, 0, len(addSlice))

	for _, one := range addSlice {
		tmpRes := routetable.GcpRouteCreateReq{
			CloudID:          one.CloudID,
			SelfLink:         one.SelfLink,
			Network:          one.Network,
			Name:             one.Name,
			DestRange:        one.DestRange,
			NextHopGateway:   one.NextHopGateway,
			NextHopIlb:       one.NextHopIlb,
			NextHopInstance:  one.NextHopInstance,
			NextHopIp:        one.NextHopIp,
			NextHopNetwork:   one.NextHopNetwork,
			NextHopPeering:   one.NextHopPeering,
			NextHopVpnTunnel: one.NextHopVpnTunnel,
			Priority:         one.Priority,
			RouteStatus:      one.RouteStatus,
			RouteType:        one.RouteType,
			Tags:             one.Tags,
			Memo:             one.Memo,
		}
		createResources = append(createResources, tmpRes)
	}

	createReq := &routetable.GcpRouteBatchCreateReq{
		GcpRoutes: createResources,
	}
	if _, err := cli.dbCli.Gcp.RouteTable.BatchCreateRoute(kt, createReq); err != nil {
		logs.Errorf("[%s] routetable batch compare db create failed. accountID: %s, err: %v",
			enumor.Gcp, accountID, err)
		return err
	}

	logs.Infof("[%s] sync route to create route success, accountID: %s, count: %d, rid: %s", enumor.Gcp,
		accountID, len(addSlice), kt.Rid)

	return nil
}

// gcp路由不支持更新，所以目前不适配更新操作
func (cli *client) updateRoute(kt *kit.Kit, accountID string,
	updateMap map[string]typesroutetable.GcpRoute) error {

	if len(updateMap) >= 1 {
		return fmt.Errorf("gcp route can not update, please check it")
	}

	logs.Infof("[%s] sync route to update route success, accountID: %s, count: %d, rid: %s", enumor.Gcp,
		accountID, len(updateMap), kt.Rid)

	return nil
}

func (cli *client) deleteRoute(kt *kit.Kit, accountID string, delCloudIDs []string,
	routeFromDB []cloudcoreroutetable.GcpRoute) error {

	if len(delCloudIDs) <= 0 {
		return fmt.Errorf("route delCloudIDs is <= 0, not delete")
	}

	checkParams := &SyncBaseParams{
		AccountID: accountID,
		CloudIDs:  delCloudIDs,
	}
	delRouteFromCloud, err := cli.listRouteFromCloud(kt, checkParams)
	if err != nil {
		return err
	}

	if len(delRouteFromCloud) > 0 {
		logs.Errorf("[%s] validate route not exist failed, before delete, opt: %v, failed_count: %d, rid: %s",
			enumor.Gcp, checkParams, len(delRouteFromCloud), kt.Rid)
		return fmt.Errorf("validate route not exist failed, before delete")
	}

	tableIDMap := cli.converterRouteSliceToMap(routeFromDB)

	for _, id := range delCloudIDs {
		if _, exsit := tableIDMap[id]; !exsit {
			return fmt.Errorf("delete route: %s not find in db", id)
		}

		deleteReq := &dataservice.BatchDeleteReq{
			Filter: tools.EqualExpression("cloud_id", id),
		}
		if err = cli.dbCli.Gcp.RouteTable.BatchDeleteRoute(kt.Ctx, kt.Header(), tableIDMap[id], deleteReq); err != nil {
			logs.Errorf("[%s] request dataservice to batch delete route failed, err: %v, rid: %s", enumor.Gcp,
				err, kt.Rid)
			return err
		}
	}

	logs.Infof("[%s] sync route to delete route success, accountID: %s, count: %d, rid: %s", enumor.Gcp,
		accountID, len(delCloudIDs), kt.Rid)

	return nil
}

func (cli *client) converterRouteSliceToMap(routeFromDB []cloudcoreroutetable.GcpRoute) map[string]string {
	m := make(map[string]string)
	for _, one := range routeFromDB {
		m[one.CloudID] = one.RouteTableID
	}

	return m
}

func (cli *client) listRouteFromCloud(kt *kit.Kit, params *SyncBaseParams) ([]typesroutetable.GcpRoute, error) {

	if err := params.Validate(); err != nil {
		return nil, errf.NewFromErr(errf.InvalidParameter, err)
	}

	opt := &typesroutetable.GcpListOption{
		CloudIDs: params.CloudIDs,
		Page: &adcore.GcpPage{
			PageSize: adcore.GcpQueryLimit,
		},
	}
	results, err := cli.cloudCli.ListRoute(kt, opt)
	if err != nil {
		logs.Errorf("[%s] list route from cloud failed, err: %v, account: %s, opt: %v, rid: %s", enumor.Gcp,
			err, params.AccountID, opt, kt.Rid)
		return nil, err
	}

	return results.Details, nil
}

func (cli *client) listRouteFromDB(kt *kit.Kit, params *SyncBaseParams, option *SyncRouteOption) (
	[]cloudcoreroutetable.GcpRoute, error) {

	if err := params.Validate(); err != nil {
		return nil, errf.NewFromErr(errf.InvalidParameter, err)
	}

	req := &routetable.GcpRouteListReq{
		ListReq: &core.ListReq{
			Filter: &filter.Expression{
				Op: filter.And,
				Rules: []filter.RuleFactory{
					&filter.AtomRule{Field: "cloud_id", Op: filter.In.Factory(), Value: params.CloudIDs},
				},
			},
			Page: core.NewDefaultBasePage(),
		},
	}
	result, err := cli.dbCli.Gcp.RouteTable.ListRoute(kt.Ctx, kt.Header(), req)
	if err != nil {
		logs.Errorf("[%s] list disk from db failed, err: %v, account: %s, req: %v, rid: %s", enumor.Gcp,
			err, params.AccountID, req, kt.Rid)
		return nil, err
	}

	return result.Details, nil
}

func (cli *client) RemoveRouteDeleteFromCloud(kt *kit.Kit, accountID string) error {
	req := &routetable.GcpRouteListReq{
		ListReq: &core.ListReq{
			Filter: tools.AllExpression(),
			Page: &core.BasePage{
				Start: 0,
				Limit: constant.BatchOperationMaxLimit,
			},
		},
	}
	for {
		resultFromDB, err := cli.dbCli.Gcp.RouteTable.ListRoute(kt.Ctx, kt.Header(), req)
		if err != nil {
			logs.Errorf("[%s] request dataservice to list route failed, err: %v, req: %v, rid: %s", enumor.Gcp,
				err, req, kt.Rid)
			return err
		}

		if len(resultFromDB.Details) == 0 {
			break
		}

		cloudIDs := make([]string, 0)
		for _, one := range resultFromDB.Details {
			cloudIDs = append(cloudIDs, one.CloudID)
		}

		params := &SyncBaseParams{
			AccountID: accountID,
			CloudIDs:  cloudIDs,
		}
		resultFromCloud, err := cli.listRouteFromCloud(kt, params)
		if err != nil {
			return err
		}

		// 如果有资源没有查询出来，说明数据被从云上删除
		if len(resultFromCloud) != len(cloudIDs) {
			cloudIDMap := converter.StringSliceToMap(cloudIDs)
			for _, one := range resultFromCloud {
				delete(cloudIDMap, one.CloudID)
			}

			delIDs := converter.MapKeyToStringSlice(cloudIDMap)
			if len(delIDs) > 0 {
				if err := cli.deleteRoute(kt, accountID, delIDs, resultFromDB.Details); err != nil {
					return err
				}
			}
		}

		if len(resultFromDB.Details) < constant.BatchOperationMaxLimit {
			break
		}

		req.Page.Start += constant.BatchOperationMaxLimit
	}

	return nil
}

func isRouteChange(cloud typesroutetable.GcpRoute, db cloudcoreroutetable.GcpRoute) bool {
	return false
}
