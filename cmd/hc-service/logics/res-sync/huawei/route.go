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

package huawei

import (
	"fmt"

	"hcm/cmd/hc-service/logics/res-sync/common"
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
				enumor.HuaWei, params.AccountID, param, err, kt.Rid)
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

	addSlice, updateMap, delCloudIDs := common.Diff[typesroutetable.HuaWeiRoute,
		routetable.HuaWeiRoute](routeFromCloud, routeFromDB, isRouteChange)

	if len(delCloudIDs) > 0 {
		if err = cli.deleteRoute(kt, opt.AccountID, opt.Region, opt.CloudRouteTableID, routeTable.ID, delCloudIDs,
			routeFromDB); err != nil {

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

// createRoute create route in db
func (cli *client) createRoute(kt *kit.Kit, accountID string, region string, routeTableID string,
	addSlice []typesroutetable.HuaWeiRoute) error {

	if len(addSlice) <= 0 {
		return fmt.Errorf("route addSlice is <= 0, not create")
	}

	createResources := make([]dataproto.HuaWeiRouteCreateReq, 0, len(addSlice))
	for _, one := range addSlice {
		tmpRes := dataproto.HuaWeiRouteCreateReq{
			CloudRouteTableID: one.CloudRouteTableID,
			Type:              one.Type,
			Destination:       one.Destination,
			NextHop:           one.NextHop,
			Memo:              one.Memo,
		}
		createResources = append(createResources, tmpRes)
	}

	createReq := &dataproto.HuaWeiRouteBatchCreateReq{
		HuaWeiRoutes: createResources,
	}
	if _, err := cli.dbCli.HuaWei.RouteTable.BatchCreateRoute(kt.Ctx, kt.Header(),
		routeTableID, createReq); err != nil {
		logs.Errorf("[%s] batch create route failed. err: %v, accountID: %s, region: %s, "+
			"routeTableID: %s, rid: %s", enumor.HuaWei, err, accountID, region, routeTableID, kt.Rid)
		return err
	}

	logs.Infof("[%s] sync route to create route success, accountID: %s, count: %d, rid: %s", enumor.HuaWei,
		accountID, len(addSlice), kt.Rid)

	return nil
}

// updateRoute update route in db
func (cli *client) updateRoute(kt *kit.Kit, accountID, region, routeTableID string,
	updateMap map[string]typesroutetable.HuaWeiRoute) error {

	if len(updateMap) <= 0 {
		return fmt.Errorf("route updateMap is <= 0, not update")
	}

	updateResources := make([]dataproto.HuaWeiRouteUpdateReq, 0, len(updateMap))
	for id, one := range updateMap {
		tmpRes := dataproto.HuaWeiRouteUpdateReq{
			ID: id,
		}
		tmpRes.Type = one.Type
		tmpRes.Destination = one.Destination
		tmpRes.NextHop = one.NextHop
		tmpRes.Memo = one.Memo

		updateResources = append(updateResources, tmpRes)
	}

	updateReq := &dataproto.HuaWeiRouteBatchUpdateReq{
		HuaWeiRoutes: updateResources,
	}
	if err := cli.dbCli.HuaWei.RouteTable.BatchUpdateRoute(kt.Ctx, kt.Header(), routeTableID,
		updateReq); err != nil {
		logs.Errorf("[%s] batch update route failed. err: %v, accountID: %s, region: %s, "+
			"routeTableID: %s, rid: %s", enumor.HuaWei, err, accountID, region, routeTableID, kt.Rid)
		return err
	}
	logs.Infof("[%s] sync route to update route success, accountID: %s, count: %d, rid: %s", enumor.HuaWei,
		accountID, len(updateMap), kt.Rid)

	return nil
}

// deleteRoute delete route in db
func (cli *client) deleteRoute(kt *kit.Kit, accountID, region, cloudRTID, rtID string,
	delCloudIDs []string, routeFromDB []routetable.HuaWeiRoute) error {

	if len(delCloudIDs) <= 0 {
		return fmt.Errorf("route delCloudIDs is <= 0, not delete")
	}

	// check route not exist in cloud, before delete
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
		if _, exist := cloudIDMap[route.GetCloudID()]; exist {
			logs.Errorf("[%s] validate route not exist failed, before delete, opt: %v, routeTableID: %s, "+
				"destination: %s, rid: %s", enumor.HuaWei, checkParams, cloudRTID, route.Destination, kt.Rid)
			return fmt.Errorf("validate route not exist failed, before delete")
		}
	}

	deleteIDs := make([]string, 0)
	for _, one := range routeFromDB {
		if _, exist := cloudIDMap[one.GetCloudID()]; exist {
			deleteIDs = append(deleteIDs, one.ID)
		}
	}

	deleteReq := &dataservice.BatchDeleteReq{
		Filter: tools.ContainersExpression("id", deleteIDs),
	}
	if err = cli.dbCli.HuaWei.RouteTable.BatchDeleteRoute(kt.Ctx, kt.Header(), rtID, deleteReq); err != nil {
		logs.Errorf("[%s] request dataservice to batch delete route failed, err: %v, rid: %s",
			enumor.HuaWei, err, kt.Rid)
		return err
	}

	logs.Infof("[%s] sync route to delete route success, accountID: %s, count: %d, rid: %s", enumor.HuaWei,
		accountID, len(delCloudIDs), kt.Rid)

	return nil
}

// listRouteFromCloud list route table from cloud
func (cli *client) listRouteFromCloud(kt *kit.Kit, opt *syncRouteOption) ([]typesroutetable.HuaWeiRoute, error) {

	if err := opt.Validate(); err != nil {
		return nil, errf.NewFromErr(errf.InvalidParameter, err)
	}

	routeOpt := &typesroutetable.HuaWeiRouteTableGetOption{
		Region: opt.Region,
		ID:     opt.CloudRouteTableID,
	}
	routeTable, err := cli.cloudCli.GetRouteTable(kt, routeOpt)
	if err != nil {
		logs.Errorf("[%s] list routeTable from cloud failed, err: %v, account: %s, opt: %v, rid: %s", enumor.HuaWei, err,
			opt.AccountID, opt, kt.Rid)
		return nil, err
	}

	results := make([]typesroutetable.HuaWeiRoute, 0)
	if len(routeTable.Extension.Routes) > 0 {
		results = append(results, routeTable.Extension.Routes...)
	}

	return results, nil
}

// listRouteTableFromDB list route table from db
func (cli *client) listRouteFromDB(kt *kit.Kit, opt *syncRouteOption, rtID string) ([]routetable.HuaWeiRoute, error) {
	if err := opt.Validate(); err != nil {
		return nil, errf.NewFromErr(errf.InvalidParameter, err)
	}

	req := &core.ListReq{
		Filter: tools.EqualExpression("cloud_route_table_id", opt.CloudRouteTableID),
		Page: &core.BasePage{
			Limit: core.DefaultMaxPageLimit,
		},
	}
	results, err := cli.dbCli.HuaWei.RouteTable.ListRoute(kt.Ctx, kt.Header(), rtID, req)
	if err != nil {
		logs.Errorf("[%s] batch list route failed. err: %v, accountID: %s, region: %s, "+
			"routeTableID: %s, rid: %s", enumor.HuaWei, err, opt.AccountID, opt.Region, rtID, kt.Rid)
		return nil, err
	}

	routes := make([]routetable.HuaWeiRoute, 0)
	for _, item := range results.Details {
		routes = append(routes, item)
	}

	return routes, nil
}

// isRouteChange checks if the route has changed
func isRouteChange(cloud typesroutetable.HuaWeiRoute,
	db routetable.HuaWeiRoute) bool {

	if cloud.Type == db.Type {
		return true
	}
	if cloud.Destination == db.Destination {
		return true
	}
	if cloud.NextHop == db.NextHop {
		return true
	}
	if converter.PtrToVal(cloud.Memo) == converter.PtrToVal(db.Memo) {
		return true
	}

	return false
}
