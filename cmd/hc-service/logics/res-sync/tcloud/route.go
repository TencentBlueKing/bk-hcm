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

package tcloud

import (
	"fmt"
	"strings"

	"hcm/cmd/hc-service/logics/res-sync/common"
	"hcm/pkg/adaptor/tcloud"
	adcore "hcm/pkg/adaptor/types/core"
	typesroutetable "hcm/pkg/adaptor/types/route-table"
	"hcm/pkg/api/core"
	routetable "hcm/pkg/api/core/cloud/route-table"
	dataservice "hcm/pkg/api/data-service"
	dataproto "hcm/pkg/api/data-service/cloud/route-table"
	"hcm/pkg/criteria/constant"
	"hcm/pkg/criteria/enumor"
	"hcm/pkg/criteria/errf"
	"hcm/pkg/criteria/validator"
	"hcm/pkg/dal/dao/tools"
	"hcm/pkg/kit"
	"hcm/pkg/logs"
	"hcm/pkg/tools/concurrence"
	"hcm/pkg/tools/converter"
	"hcm/pkg/tools/slice"
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

	err := concurrence.BaseExec(constant.SyncConcurrencyDefaultMaxLimit, params.CloudIDs, func(param string) error {
		syncOpt := &syncRouteOption{
			AccountID:         params.AccountID,
			Region:            params.Region,
			CloudRouteTableID: param,
		}
		if _, err := cli.route(kt, syncOpt); err != nil {
			logs.ErrorDepthf(1, "[%s] account: %s route_table: %s sync route failed, err: %v, rid: %s",
				enumor.TCloud, params.AccountID, param, err, kt.Rid)
			return err
		}

		return nil
	})
	if err != nil {
		return nil, err
	}

	return new(SyncResult), nil
}

type syncRouteOption struct {
	AccountID         string `json:"account_id" validate:"required"`
	Region            string `json:"region" validate:"required"`
	CloudRouteTableID string `json:"cloud_route_table_id" validate:"required"`
}

// Validate ...
func (opt syncRouteOption) Validate() error {
	return validator.Validate.Struct(opt)
}

func (cli *client) route(kt *kit.Kit, opt *syncRouteOption) (*SyncResult, error) {
	if err := opt.Validate(); err != nil {
		return nil, errf.NewFromErr(errf.InvalidParameter, err)
	}

	params := &SyncBaseParams{
		AccountID: opt.AccountID,
		Region:    opt.Region,
		CloudIDs:  []string{opt.CloudRouteTableID},
	}
	rts, err := cli.listRouteTableFromDB(kt, params)
	if err != nil {
		return nil, err
	}

	if len(rts) == 0 {
		return nil, fmt.Errorf("route table: %s not found from db", opt.CloudRouteTableID)
	}
	routeTable := rts[0]

	routeFromDB, err := cli.listRouteFromDB(kt, opt, routeTable.ID)
	if err != nil {
		return nil, err
	}

	routeFromCloud, err := cli.listRouteFromCloud(kt, opt)
	if err != nil {
		return nil, err
	}

	if len(routeFromCloud) == 0 && len(routeFromDB) == 0 {
		return new(SyncResult), nil
	}

	addSlice, updateMap, delCloudIDs := common.Diff[typesroutetable.TCloudRoute,
		routetable.TCloudRoute](routeFromCloud, routeFromDB, isRouteChange)

	if len(delCloudIDs) > 0 {
		if err = cli.deleteRoute(kt, opt.AccountID, opt.Region, opt.CloudRouteTableID, routeTable.ID,
			delCloudIDs); err != nil {

			return nil, err
		}
	}

	if len(addSlice) > 0 {
		err := cli.createRoute(kt, opt.AccountID, opt.Region, routeTable.ID, addSlice)
		if err != nil {
			return nil, err
		}
	}

	if len(updateMap) > 0 {
		err := cli.updateRoute(kt, opt.AccountID, opt.Region, routeTable.ID, updateMap)
		if err != nil {
			return nil, err
		}
	}

	return new(SyncResult), nil
}

func (cli *client) createRoute(kt *kit.Kit, accountID string, region string, routeTableID string,
	addSlice []typesroutetable.TCloudRoute) error {

	if len(addSlice) <= 0 {
		return fmt.Errorf("route addSlice is <= 0, not create")
	}

	// split all routes into batches to avoid reaching batch operation limit
	routeBatches := slice.Split(addSlice, constant.BatchOperationMaxLimit)

	for batchIdx, routeBatch := range routeBatches {
		createResources := slice.Map(routeBatch,
			func(route typesroutetable.TCloudRoute) dataproto.TCloudRouteCreateReq {
				return dataproto.TCloudRouteCreateReq{
					CloudID:                  route.CloudID,
					CloudRouteTableID:        route.CloudRouteTableID,
					DestinationCidrBlock:     route.DestinationCidrBlock,
					DestinationIpv6CidrBlock: route.DestinationIpv6CidrBlock,
					GatewayType:              route.GatewayType,
					CloudGatewayID:           route.CloudGatewayID,
					Enabled:                  route.Enabled,
					RouteType:                route.RouteType,
					PublishedToVbc:           route.PublishedToVbc,
					Memo:                     route.Memo,
				}
			})
		createReq := &dataproto.TCloudRouteBatchCreateReq{
			TCloudRoutes: createResources,
		}
		if _, err := cli.dbCli.TCloud.RouteTable.BatchCreateRoute(kt, routeTableID, createReq); err != nil {
			logs.Errorf("[%s] batch create route failed. err: %v, accountID: %s, region: %s, "+
				"routeTableID: %s, batchIdx: %d, rid: %s", enumor.TCloud, err, accountID, region, routeTableID,
				batchIdx, kt.Rid)
			return err
		}

	}
	logs.Infof("[%s] sync route to create route success, accountID: %s, count: %d, rid: %s", enumor.TCloud,
		accountID, len(addSlice), kt.Rid)
	return nil
}

func (cli *client) updateRoute(kt *kit.Kit, accountID, region, routeTableID string,
	updateMap map[string]typesroutetable.TCloudRoute) error {

	if len(updateMap) <= 0 {
		return fmt.Errorf("route updateMap is <= 0, not update")
	}

	updateResources := make([]dataproto.TCloudRouteUpdateReq, 0, len(updateMap))
	for id, one := range updateMap {
		tmpRes := dataproto.TCloudRouteUpdateReq{
			ID: id,
		}
		tmpRes.DestinationCidrBlock = one.DestinationCidrBlock
		tmpRes.DestinationIpv6CidrBlock = one.DestinationIpv6CidrBlock
		tmpRes.GatewayType = one.GatewayType
		tmpRes.CloudGatewayID = one.CloudGatewayID
		tmpRes.Enabled = converter.ValToPtr(one.Enabled)
		tmpRes.RouteType = one.RouteType
		tmpRes.PublishedToVbc = converter.ValToPtr(one.PublishedToVbc)
		tmpRes.Memo = one.Memo

		updateResources = append(updateResources, tmpRes)
	}

	updateReqBatches := slice.Split(updateResources, constant.BatchOperationMaxLimit)
	for batchIdx, updateBatch := range updateReqBatches {
		updateReq := &dataproto.TCloudRouteBatchUpdateReq{
			TCloudRoutes: updateBatch,
		}
		if err := cli.dbCli.TCloud.RouteTable.BatchUpdateRoute(kt, routeTableID, updateReq); err != nil {
			logs.Errorf("[%s] batch update route failed. err: %v, accountID: %s, region: %s, "+
				"routeTableID: %s, batchIdx: %d, rid: %s",
				enumor.TCloud, err, accountID, region, routeTableID, batchIdx, kt.Rid)
			return err
		}
	}

	logs.Infof("[%s] sync route to update route success, accountID: %s, count: %d, rid: %s", enumor.TCloud,
		accountID, len(updateMap), kt.Rid)

	return nil
}

func (cli *client) deleteRoute(kt *kit.Kit, accountID, region, cloudRTID, rtID string,
	delCloudIDs []string) error {

	if len(delCloudIDs) <= 0 {
		return fmt.Errorf("route delCloudIDs is <= 0, not delete")
	}

	checkParams := &syncRouteOption{
		AccountID:         accountID,
		Region:            region,
		CloudRouteTableID: cloudRTID,
	}
	delRouteFromCloud, err := cli.listRouteFromCloud(kt, checkParams)
	if err != nil {
		return err
	}

	cloudIDMap := converter.StringSliceToMap(delCloudIDs)
	for _, route := range delRouteFromCloud {
		if _, exist := cloudIDMap[route.CloudID]; exist {
			logs.Errorf("[%s] validate route not exist failed, before delete, opt: %v, cloud_id: %s, rid: %s",
				enumor.TCloud, checkParams, route.CloudID, kt.Rid)
			return fmt.Errorf("validate route not exist failed, before delete")
		}
	}

	cloudIDBatches := slice.Split(delCloudIDs, constant.BatchOperationMaxLimit)
	for batchIdx, cloudIDBatch := range cloudIDBatches {

		deleteReq := &dataservice.BatchDeleteReq{
			Filter: tools.ContainersExpression("cloud_id", cloudIDBatch),
		}
		if err = cli.dbCli.TCloud.RouteTable.BatchDeleteRoute(kt, rtID, deleteReq); err != nil {

			logs.Errorf("[%s] request dataservice to batch delete route failed, err: %v, batchIdx: %d,rid: %s",
				enumor.TCloud, err, batchIdx, kt.Rid)
			return err
		}
	}

	logs.Infof("[%s] sync route to delete route success, accountID: %s, count: %d, rid: %s", enumor.TCloud,
		accountID, len(delCloudIDs), kt.Rid)

	return nil
}

func (cli *client) listRouteFromCloud(kt *kit.Kit, opt *syncRouteOption) ([]typesroutetable.TCloudRoute, error) {

	if err := opt.Validate(); err != nil {
		return nil, errf.NewFromErr(errf.InvalidParameter, err)
	}

	routeOpt := &adcore.TCloudListOption{
		Region:   opt.Region,
		CloudIDs: []string{opt.CloudRouteTableID},
		Page: &adcore.TCloudPage{
			Offset: 0,
			Limit:  adcore.TCloudQueryLimit,
		},
	}
	routeTables, err := cli.cloudCli.ListRouteTable(kt, routeOpt)
	if err != nil {
		if strings.Contains(err.Error(), tcloud.ErrNotFound) {
			return nil, nil
		}

		logs.Errorf("[%s] list route from cloud failed, err: %v, account: %s, opt: %v, rid: %s", enumor.TCloud, err,
			opt.AccountID, opt, kt.Rid)
		return nil, err
	}

	results := make([]typesroutetable.TCloudRoute, 0)
	for _, routeTable := range routeTables.Details {
		results = append(results, routeTable.Extension.Routes...)
	}

	return results, nil
}

func (cli *client) listRouteFromDB(kt *kit.Kit, opt *syncRouteOption, routeTableID string) (
	[]routetable.TCloudRoute, error) {

	if err := opt.Validate(); err != nil {
		return nil, errf.NewFromErr(errf.InvalidParameter, err)
	}

	req := &core.ListReq{
		Filter: tools.EqualExpression("cloud_route_table_id", opt.CloudRouteTableID),
		Page:   core.NewDefaultBasePage(),
	}
	results, err := cli.dbCli.TCloud.RouteTable.ListRoute(kt.Ctx, kt.Header(), routeTableID, req)
	if err != nil {
		logs.Errorf("[%s] batch list route failed. err: %v, accountID: %s, region: %s, "+
			"routeTableID: %s, rid: %s", enumor.TCloud, err, opt.AccountID, opt.Region, routeTableID, kt.Rid)
		return nil, err
	}

	routes := make([]routetable.TCloudRoute, 0)
	for _, item := range results.Details {
		routes = append(routes, item)
	}

	return routes, nil
}

func isRouteChange(cloud typesroutetable.TCloudRoute,
	db routetable.TCloudRoute) bool {

	if cloud.DestinationCidrBlock != db.DestinationCidrBlock {
		return true
	}

	if converter.PtrToVal(cloud.DestinationIpv6CidrBlock) !=
		converter.PtrToVal(db.DestinationIpv6CidrBlock) {
		return true
	}

	if cloud.GatewayType != db.GatewayType {
		return true
	}

	if cloud.CloudGatewayID != db.CloudGatewayID {
		return true
	}

	if cloud.Enabled != db.Enabled {
		return true
	}

	if cloud.RouteType != db.RouteType {
		return true
	}

	if cloud.PublishedToVbc != db.PublishedToVbc {
		return true
	}

	if converter.PtrToVal(cloud.Memo) != converter.PtrToVal(db.Memo) {
		return true
	}

	return false
}
